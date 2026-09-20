// Comunidad v1: previsualización de perfiles por UUID, usuario o UUID de Minecraft.
// Solo perfiles con "poder buscarte" activado (amigos siempre pueden verse).
// Bloqueo en cualquier sentido = solo aviso de bloqueo, sin datos.
// El email NUNCA se expone aquí.

import { Hono, type Context } from "hono";
import type { SupabaseClient } from "@supabase/supabase-js";
import type { AppEnv } from "../../types/app";
import { getEnv } from "../../env";
import { ok, fail } from "../../lib/respond";
import { supabaseAdmin } from "../../lib/supabase";
import { requireAuth } from "../../middleware/auth";
import { rateLimit, rateLimitPresets } from "../../middleware/rate-limit";
import { validateParam } from "../../middleware/validate";
import { profileIdentifierSchema } from "../../schemas/v1/community";

export const communityRoutes = new Hono<AppEnv>();

function authedAdmin(c: Context<AppEnv>) {
  const user = c.get("user");
  if (!user) return null;
  try {
    return { user, admin: supabaseAdmin(getEnv(c)) };
  } catch {
    return null;
  }
}

/** Bloqueo en algún sentido entre ambos (blocker_id del que bloqueó). */
async function blockDirection(admin: SupabaseClient, me: string, target: string) {
  const { data } = await admin
    .from("blocks")
    .select("blocker_id")
    .or(`and(blocker_id.eq.${me},blocked_id.eq.${target}),and(blocker_id.eq.${target},blocked_id.eq.${me})`);
  const rows = (data ?? []) as Array<{ blocker_id: string }>;
  if (rows.some((r) => r.blocker_id === target)) return "blocked_you" as const;
  if (rows.some((r) => r.blocker_id === me)) return "blocked_by_you" as const;
  return null;
}

// GET /v1/community/profiles/:identifier — previsualizar perfil.
communityRoutes.get(
  "/profiles/:identifier",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateParam(profileIdentifierSchema),
  async (c) => {
    const a = authedAdmin(c);
    if (!a) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }
    const identifier = c.req.valid("param").identifier;
    const notFound = () =>
      fail(c, { code: "profile_not_found", message: "Perfil no encontrado.", status: 404 });

    const { data: targetId, error: resolveError } = await a.admin.rpc("resolve_user_identity", {
      p_identifier: identifier,
    });
    if (resolveError) {
      console.error(`[${c.get("requestId")}] preview resolve:`, resolveError.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    if (!targetId) return notFound();
    const target = targetId as string;
    const self = target === a.user.id;

    // Lista negra primero: en cualquier sentido solo se devuelve el aviso.
    if (!self) {
      const blocked = await blockDirection(a.admin, a.user.id, target);
      if (blocked === "blocked_you") {
        return ok(c, {
          blocked: true,
          reason: "blocked_you",
          message: "Ese jugador te tiene en su lista negra y no puedes ver su perfil.",
        });
      }
      if (blocked === "blocked_by_you") {
        return ok(c, {
          blocked: true,
          reason: "blocked_by_you",
          message: "Tienes a ese jugador en tu lista negra.",
        });
      }
    }

    // Amigos y uno mismo ven siempre; el resto exige `searchable`.
    let friendsSince: string | null = null;
    if (!self) {
      const { data: fr } = await a.admin
        .from("friendships")
        .select("created_at")
        .eq("user_id", a.user.id)
        .eq("friend_id", target)
        .maybeSingle();
      friendsSince = ((fr as { created_at: string } | null)?.created_at) ?? null;
      if (!friendsSince) {
        const { data: priv } = await a.admin
          .from("privacy_settings")
          .select("searchable")
          .eq("user_id", target)
          .maybeSingle();
        const searchable = (priv as { searchable: boolean } | null)?.searchable ?? true;
        if (!searchable) return notFound();
      }
    }

    const { data: profile } = await a.admin
      .from("profiles")
      .select("username, display_name, avatar_url, banner_url, bio, mc_uuid, last_mc_version, created_at, is_online")
      .eq("user_id", target)
      .maybeSingle();
    if (!profile) return notFound();
    const p = profile as {
      username: string;
      display_name: string;
      avatar_url: string | null;
      banner_url: string | null;
      bio: string | null;
      mc_uuid: string | null;
      last_mc_version: string | null;
      created_at: string;
      is_online: boolean;
    };

    const { data: equipped } = await a.admin.rpc("equipped_cosmetics", { p_user_id: target });
    const cosmetics = ((equipped ?? []) as Array<{
      id: string;
      slug: string;
      name: string;
      kind: string;
      image_url: string | null;
    }>).map((e) => ({
      id: e.id,
      slug: e.slug,
      name: e.name,
      kind: e.kind,
      imageUrl: e.image_url,
    }));

    // Relación para actuar desde la preview (requestId permite retirar la enviada).
    let relationship: { status: string; requestId: string | null; friendsSince: string | null } = {
      status: self ? "self" : "none",
      requestId: null,
      friendsSince,
    };
    if (!self && !friendsSince) {
      const { data: pending } = await a.admin
        .from("friend_requests")
        .select("id, from_user_id")
        .eq("status", "pending")
        .in("from_user_id", [a.user.id, target])
        .in("to_user_id", [a.user.id, target])
        .limit(1);
      const row = (pending ?? [])[0] as { id: string; from_user_id: string } | undefined;
      if (row) {
        relationship = {
          status: row.from_user_id === a.user.id ? "pending_sent" : "pending_received",
          requestId: row.id,
          friendsSince: null,
        };
      }
    } else if (!self) {
      relationship = { status: "friends", requestId: null, friendsSince };
    }

    return ok(c, {
      blocked: false,
      profile: {
        userId: target,
        username: p.username,
        displayName: p.display_name,
        avatarUrl: p.avatar_url,
        bannerUrl: p.banner_url,
        bio: p.bio,
        memberSince: p.created_at,
        isOnline: p.is_online ?? false,
      },
      minecraft: {
        uuid: p.mc_uuid,
        lastVersion: p.last_mc_version,
      },
      equippedCosmetics: cosmetics,
      relationship,
    });
  },
);
