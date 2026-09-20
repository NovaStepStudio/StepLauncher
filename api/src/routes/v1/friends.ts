// Amistades v1: solicitudes por email/usuario/UUID, aceptación y lista.
// La resolución del destinatario respeta SU privacidad a nivel SQL
// (ver `resolve_friend_target`): bloqueado o inexistente = 404 genérico.

import { Hono, type Context } from "hono";
import type { SupabaseClient } from "@supabase/supabase-js";
import type { AppEnv } from "../../types/app";
import { getEnv } from "../../env";
import { ok, fail } from "../../lib/respond";
import { supabaseAdmin } from "../../lib/supabase";
import { requireAuth } from "../../middleware/auth";
import { rateLimit, rateLimitPresets } from "../../middleware/rate-limit";
import { validateJson, validateQuery, validateParam } from "../../middleware/validate";
import {
  sendFriendRequestSchema,
  listRequestsSchema,
  requestIdSchema,
  friendIdSchema,
  blockUserSchema,
  blockUserIdSchema,
} from "../../schemas/v1/friends";

export const friendRoutes = new Hono<AppEnv>();

/** Usuario + cliente admin (las resoluciones sensibles exigen service_role). */
function authedAdmin(c: Context<AppEnv>) {
  const user = c.get("user");
  if (!user) return null;
  try {
    return { user, admin: supabaseAdmin(getEnv(c)) };
  } catch {
    return null;
  }
}

interface MiniProfileRow {
  user_id: string;
  username: string;
  display_name: string;
  avatar_url: string | null;
}

function miniShape(p: MiniProfileRow) {
  return {
    userId: p.user_id,
    username: p.username,
    displayName: p.display_name,
    avatarUrl: p.avatar_url,
  };
}

async function miniProfile(admin: SupabaseClient, userId: string) {
  const { data } = await admin
    .from("profiles")
    .select("user_id, username, display_name, avatar_url")
    .eq("user_id", userId)
    .maybeSingle();
  return data ? miniShape(data as MiniProfileRow) : null;
}

/** Notificación best-effort respetando `receive_notifications` (defecto: sí). */
async function maybeNotify(admin: SupabaseClient, userId: string, title: string, body: string) {
  try {
    const { data: priv } = await admin
      .from("privacy_settings")
      .select("receive_notifications")
      .eq("user_id", userId)
      .maybeSingle();
    if (priv && (priv as { receive_notifications: boolean }).receive_notifications === false) {
      return;
    }
    await admin.from("notifications").insert({ user_id: userId, title, body });
  } catch {
    // Mejor esfuerzo: la acción principal ya se completó.
  }
}

const unauthorized = (c: Context<AppEnv>) =>
  fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });

// POST /v1/friends/requests — enviar solicitud por email, usuario o UUID.
friendRoutes.post(
  "/requests",
  rateLimit(rateLimitPresets.sensitive),
  requireAuth(),
  validateJson(sendFriendRequestSchema),
  async (c) => {
    const a = authedAdmin(c);
    if (!a) return unauthorized(c);
    const input = c.req.valid("json");

    const { data: targetId, error: resolveError } = await a.admin.rpc("resolve_friend_target", {
      p_identifier: input.identifier,
    });
    if (resolveError) {
      console.error(`[${c.get("requestId")}] friend resolve:`, resolveError.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    if (!targetId) {
      return fail(c, { code: "user_not_found", message: "Jugador no encontrado o no disponible.", status: 404 });
    }
    const target = targetId as string;
    if (target === a.user.id) {
      return fail(c, { code: "cannot_add_self", message: "No puedes añadirte a ti mismo.", status: 400 });
    }

    // Lista negra: si te bloquearon, genérico (sin revelar); si bloqueaste tú, 403.
    const { data: blockRows } = await a.admin
      .from("blocks")
      .select("blocker_id")
      .or(`and(blocker_id.eq.${a.user.id},blocked_id.eq.${target}),and(blocker_id.eq.${target},blocked_id.eq.${a.user.id})`);
    const rows = (blockRows ?? []) as Array<{ blocker_id: string }>;
    if (rows.some((b) => b.blocker_id === target)) {
      return fail(c, { code: "user_not_found", message: "Jugador no encontrado o no disponible.", status: 404 });
    }
    if (rows.some((b) => b.blocker_id === a.user.id)) {
      return fail(c, { code: "user_blocked", message: "Desbloquea a ese jugador primero.", status: 403 });
    }

    const { data: existing } = await a.admin
      .from("friendships")
      .select("friend_id")
      .eq("user_id", a.user.id)
      .eq("friend_id", target)
      .maybeSingle();
    if (existing) {
      return fail(c, { code: "already_friends", message: "Ya sois amigos.", status: 409 });
    }

    const { data: pending } = await a.admin
      .from("friend_requests")
      .select("id")
      .eq("status", "pending")
      .in("from_user_id", [a.user.id, target])
      .in("to_user_id", [a.user.id, target])
      .limit(1);
    if (pending && pending.length > 0) {
      return fail(c, { code: "already_pending", message: "Ya hay una solicitud pendiente.", status: 409 });
    }

    const { data: created, error: createError } = await a.admin
      .from("friend_requests")
      .insert({ from_user_id: a.user.id, to_user_id: target })
      .select("id, created_at")
      .single();
    if (createError || !created) {
      console.error(`[${c.get("requestId")}] friend create:`, createError?.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }

    const me = await miniProfile(a.admin, a.user.id);
    await maybeNotify(
      a.admin,
      target,
      "Nueva solicitud de amistad",
      `${me?.username ?? "Un jugador"} quiere ser tu amigo.`,
    );
    const toUser = await miniProfile(a.admin, target);
    return ok(
      c,
      {
        id: (created as { id: string }).id,
        status: "pending",
        toUser,
      },
      201,
    );
  },
);

// GET /v1/friends/requests?type=incoming|sent|all — listar solicitudes.
friendRoutes.get(
  "/requests",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateQuery(listRequestsSchema),
  async (c) => {
    const a = authedAdmin(c);
    if (!a) return unauthorized(c);
    const q = c.req.valid("query");

    let query = a.admin
      .from("friend_requests")
      .select("id, from_user_id, to_user_id, status, created_at, responded_at")
      .order("created_at", { ascending: false });
    if (q.type === "incoming") query = query.eq("to_user_id", a.user.id);
    else if (q.type === "sent") query = query.eq("from_user_id", a.user.id);
    else query = query.or(`from_user_id.eq.${a.user.id},to_user_id.eq.${a.user.id}`);

    const { data, error } = await query;
    if (error) {
      console.error(`[${c.get("requestId")}] friend list:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    const rows = (data ?? []) as Array<{
      id: string;
      from_user_id: string;
      to_user_id: string;
      status: string;
      created_at: string;
      responded_at: string | null;
    }>;
    const counterpartIds = [...new Set(rows.map((r) => (r.from_user_id === a.user.id ? r.to_user_id : r.from_user_id)))];
    const profiles = new Map<string, ReturnType<typeof miniShape>>();
    if (counterpartIds.length > 0) {
      const { data: profs } = await a.admin
        .from("profiles")
        .select("user_id, username, display_name, avatar_url")
        .in("user_id", counterpartIds);
      for (const p of (profs ?? []) as MiniProfileRow[]) profiles.set(p.user_id, miniShape(p));
    }
    return ok(c, {
      requests: rows.map((r) => ({
        id: r.id,
        direction: r.from_user_id === a.user.id ? "sent" : "incoming",
        status: r.status,
        createdAt: r.created_at,
        respondedAt: r.responded_at,
        user: profiles.get(r.from_user_id === a.user.id ? r.to_user_id : r.from_user_id) ?? null,
      })),
    });
  },
);

// POST /v1/friends/requests/:id/accept — aceptar (receptor). Crea la amistad doble.
friendRoutes.post(
  "/requests/:id/accept",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateParam(requestIdSchema),
  async (c) => {
    const a = authedAdmin(c);
    if (!a) return unauthorized(c);
    const id = c.req.valid("param").id;

    const { data: accepted, error } = await a.admin.rpc("accept_friend_request", {
      p_request_id: id,
      p_user_id: a.user.id,
    });
    if (error) {
      console.error(`[${c.get("requestId")}] friend accept:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    if (accepted !== true) {
      return fail(c, { code: "not_found", message: "Recurso no encontrado.", status: 404 });
    }
    const { data: req } = await a.admin
      .from("friend_requests")
      .select("from_user_id")
      .eq("id", id)
      .maybeSingle();
    const fromId = (req as { from_user_id: string } | null)?.from_user_id;
    const me = await miniProfile(a.admin, a.user.id);
    if (fromId) {
      await maybeNotify(a.admin, fromId, "Solicitud aceptada", `${me?.username ?? "Un jugador"} aceptó tu solicitud.`);
    }
    return ok(c, {
      id,
      status: "accepted",
      friend: fromId ? await miniProfile(a.admin, fromId) : null,
    });
  },
);

// POST /v1/friends/requests/:id/decline — rechazar (receptor).
friendRoutes.post(
  "/requests/:id/decline",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateParam(requestIdSchema),
  async (c) => {
    const a = authedAdmin(c);
    if (!a) return unauthorized(c);
    const { data, error } = await a.admin
      .from("friend_requests")
      .update({ status: "declined", responded_at: new Date().toISOString() })
      .eq("id", c.req.valid("param").id)
      .eq("to_user_id", a.user.id)
      .eq("status", "pending")
      .select("id")
      .maybeSingle();
    if (error) {
      console.error(`[${c.get("requestId")}] friend decline:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    if (!data) {
      return fail(c, { code: "not_found", message: "Recurso no encontrado.", status: 404 });
    }
    return ok(c, { id: (data as { id: string }).id, status: "declined" });
  },
);

// POST /v1/friends/requests/:id/cancel — cancelar una enviada (emisor).
friendRoutes.post(
  "/requests/:id/cancel",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateParam(requestIdSchema),
  async (c) => {
    const a = authedAdmin(c);
    if (!a) return unauthorized(c);
    const { data, error } = await a.admin
      .from("friend_requests")
      .update({ status: "cancelled", responded_at: new Date().toISOString() })
      .eq("id", c.req.valid("param").id)
      .eq("from_user_id", a.user.id)
      .eq("status", "pending")
      .select("id")
      .maybeSingle();
    if (error) {
      console.error(`[${c.get("requestId")}] friend cancel:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    if (!data) {
      return fail(c, { code: "not_found", message: "Recurso no encontrado.", status: 404 });
    }
    return ok(c, { id: (data as { id: string }).id, status: "cancelled" });
  },
);

// GET /v1/friends — lista de amigos propios.
friendRoutes.get("/", rateLimit(rateLimitPresets.account), requireAuth(), async (c) => {
  const a = authedAdmin(c);
  if (!a) return unauthorized(c);
  const { data, error } = await a.admin.rpc("my_friends", { p_user_id: a.user.id });
  if (error) {
    console.error(`[${c.get("requestId")}] friends list:`, error.message);
    return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
  }
  const rows = (data ?? []) as Array<{
    friend_id: string;
    username: string;
    display_name: string;
    avatar_url: string | null;
    mc_uuid: string | null;
    friends_since: string;
    is_online: boolean;
  }>;
  return ok(c, {
    friends: rows.map((r) => ({
      userId: r.friend_id,
      username: r.username,
      displayName: r.display_name,
      avatarUrl: r.avatar_url,
      mcUuid: r.mc_uuid,
      friendsSince: r.friends_since,
      isOnline: r.is_online ?? false,
    })),
    count: rows.length,
  });
});

// DELETE /v1/friends/:friendId — romper amistad (cualquier lado).
friendRoutes.delete(
  "/:friendId",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateParam(friendIdSchema),
  async (c) => {
    const a = authedAdmin(c);
    if (!a) return unauthorized(c);
    const { data: removed, error } = await a.admin.rpc("remove_friendship", {
      p_user_id: a.user.id,
      p_friend_id: c.req.valid("param").friendId,
    });
    if (error) {
      console.error(`[${c.get("requestId")}] friend remove:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    if (removed !== true) {
      return fail(c, { code: "not_found", message: "Recurso no encontrado.", status: 404 });
    }
    return ok(c, { removed: true });
  },
);

// GET /v1/friends/blocks — mi lista negra.
friendRoutes.get("/blocks", rateLimit(rateLimitPresets.account), requireAuth(), async (c) => {
  const a = authedAdmin(c);
  if (!a) return unauthorized(c);
  const { data, error } = await a.admin
    .from("blocks")
    .select("blocked_id, created_at")
    .eq("blocker_id", a.user.id)
    .order("created_at", { ascending: false });
  if (error) {
    console.error(`[${c.get("requestId")}] blocks list:`, error.message);
    return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
  }
  const rows = (data ?? []) as Array<{ blocked_id: string; created_at: string }>;
  const ids = rows.map((r) => r.blocked_id);
  const users = new Map<string, ReturnType<typeof miniShape>>();
  if (ids.length > 0) {
    const { data: profs } = await a.admin
      .from("profiles")
      .select("user_id, username, display_name, avatar_url")
      .in("user_id", ids);
    for (const p of (profs ?? []) as MiniProfileRow[]) users.set(p.user_id, miniShape(p));
  }
  return ok(c, {
    blocked: rows.map((r) => ({
      user: users.get(r.blocked_id) ?? { userId: r.blocked_id },
      blockedSince: r.created_at,
    })),
  });
});

// POST /v1/friends/blocks — bloquear por email, usuario o UUID.
// Corta la relación (pendientes cancelados, amistad borrada).
friendRoutes.post(
  "/blocks",
  rateLimit(rateLimitPresets.sensitive),
  requireAuth(),
  validateJson(blockUserSchema),
  async (c) => {
    const a = authedAdmin(c);
    if (!a) return unauthorized(c);
    const input = c.req.valid("json");

    const { data: targetId, error: resolveError } = await a.admin.rpc("resolve_user_identity", {
      p_identifier: input.identifier,
    });
    if (resolveError) {
      console.error(`[${c.get("requestId")}] block resolve:`, resolveError.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    if (!targetId) {
      return fail(c, { code: "user_not_found", message: "Jugador no encontrado o no disponible.", status: 404 });
    }
    const { data: blocked, error } = await a.admin.rpc("block_user", {
      p_user_id: a.user.id,
      p_target_id: targetId as string,
    });
    if (error) {
      console.error(`[${c.get("requestId")}] block:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    if (blocked !== true) {
      return fail(c, { code: "cannot_block", message: "No se puede bloquear a ese jugador.", status: 400 });
    }
    return ok(c, { blocked: true }, 201);
  },
);

// DELETE /v1/friends/blocks/:userId — desbloquear.
friendRoutes.delete(
  "/blocks/:userId",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateParam(blockUserIdSchema),
  async (c) => {
    const a = authedAdmin(c);
    if (!a) return unauthorized(c);
    const { data: removed, error } = await a.admin.rpc("unblock_user", {
      p_user_id: a.user.id,
      p_target_id: c.req.valid("param").userId,
    });
    if (error) {
      console.error(`[${c.get("requestId")}] unblock:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    if (removed !== true) {
      return fail(c, { code: "not_found", message: "Recurso no encontrado.", status: 404 });
    }
    return ok(c, { unblocked: true });
  },
);
