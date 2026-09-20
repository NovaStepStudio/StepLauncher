// Cuenta propia v1: perfil + personalización (email, usuario, contraseña, avatar).
// TODOS los endpoints exigen sesión: el dueño sale del Bearer, nunca del cliente.
// El token de una cuenta NO puede tocar datos de otra (RLS auth.uid() + carpetas propias).

import { Hono, type Context } from "hono";
import { createClient } from "@supabase/supabase-js";
import type { AppEnv } from "../../types/app";
import { getEnv, type SafeEnv } from "../../env";
import { ok, fail } from "../../lib/respond";
import { supabaseAdmin, supabaseForUser } from "../../lib/supabase";
import { requireAuth, getBearerToken } from "../../middleware/auth";
import { rateLimit, rateLimitPresets } from "../../middleware/rate-limit";
import { validateJson } from "../../middleware/validate";
import {
  updateProfileSchema,
  emailChangeSchema,
  passwordChangeSchema,
  presenceUpdateSchema,
} from "../../schemas/v1/account";
import { privacyUpdateSchema } from "../../schemas/v1/privacy";
import { cosmeticEquipSchema } from "../../schemas/v1/cosmetics";
import { imageInfo } from "../../lib/images";

export const accountRoutes = new Hono<AppEnv>();

interface ProfileRow {
  username: string;
  display_name: string;
  avatar_url: string | null;
  bio: string | null;
  last_mc_version: string | null;
  mc_uuid: string | null;
  banner_url: string | null;
  is_online: boolean;
}

/** Fechas de Auth (fuente autoritativa): creación de la cuenta y última sesión. */
interface AuthDates {
  createdAt: string | null;
  lastSessionAt: string | null;
}

function profileShape(
  userId: string,
  email: string | null,
  profile: ProfileRow | null,
  dates?: AuthDates,
) {
  return {
    id: userId,
    email,
    username: profile?.username ?? null,
    displayName: profile?.display_name ?? null,
    avatarUrl: profile?.avatar_url ?? null,
    bio: profile?.bio ?? null,
    // Última versión de Minecraft Java jugada (la actualiza el launcher al jugar).
    lastMcVersion: profile?.last_mc_version ?? null,
    // UUID de Minecraft Java enlazado (obtenible aquí y en búsquedas visibles).
    mcUuid: profile?.mc_uuid ?? null,
    bannerUrl: profile?.banner_url ?? null,
    // Booleano interno de presencia: true = en línea (lo cambia el dueño).
    isOnline: profile?.is_online ?? false,
    createdAt: dates?.createdAt ?? null,
    lastSessionAt: dates?.lastSessionAt ?? null,
  };
}

/** Lee el perfil propio con RLS (null si aún no existe la fila). */
async function ownProfile(env: SafeEnv, token: string, userId: string) {
  const { data, error } = await supabaseForUser(env, token)
    .from("profiles")
    .select("username, display_name, avatar_url, bio, last_mc_version, mc_uuid, banner_url, is_online")
    .eq("user_id", userId)
    .maybeSingle();
  if (error) throw new Error(`profile_read: ${error.message}`);
  return (data as ProfileRow | null) ?? null;
}

/** Avatar: PNG/JPEG/WebP ≤ 2 MB, verificado por firma mágica (no por MIME declarado). */
const AVATAR_MAX_BYTES = 2 * 1024 * 1024;

async function sniffAvatar(file: File): Promise<"png" | "jpg" | "webp" | null> {
  const bytes = new Uint8Array(await file.slice(0, 12).arrayBuffer());
  if (
    bytes.length >= 8 &&
    bytes[0] === 0x89 && bytes[1] === 0x50 && bytes[2] === 0x4e && bytes[3] === 0x47 &&
    bytes[4] === 0x0d && bytes[5] === 0x0a && bytes[6] === 0x1a && bytes[7] === 0x0a
  ) {
    return "png";
  }
  if (bytes.length >= 3 && bytes[0] === 0xff && bytes[1] === 0xd8 && bytes[2] === 0xff) return "jpg";
  if (
    bytes.length >= 12 &&
    bytes[0] === 0x52 && bytes[1] === 0x49 && bytes[2] === 0x46 && bytes[3] === 0x46 &&
    bytes[8] === 0x57 && bytes[9] === 0x45 && bytes[10] === 0x42 && bytes[11] === 0x50
  ) {
    return "webp";
  }
  return null;
}

/** Contexto autenticado común: user + token + env, o respuesta de error. */
function authed(c: Context<AppEnv>) {
  const user = c.get("user");
  const token = getBearerToken(c);
  if (!user || !token) return null;
  try {
    return { user, token, env: getEnv(c) };
  } catch {
    return null;
  }
}

// GET /v1/accounts/me — perfil propio completo.
accountRoutes.get("/me", rateLimit(rateLimitPresets.account), requireAuth(), async (c) => {
  const a = authed(c);
  if (!a) {
    return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
  }
  try {
    const profile = await ownProfile(a.env, a.token, a.user.id);
    // createdAt = fecha de creación de la cuenta; lastSessionAt = última sesión.
    const { data: authUser } = await supabaseAdmin(a.env).auth.admin.getUserById(a.user.id);
    return ok(
      c,
      profileShape(a.user.id, a.user.email ?? null, profile, {
        createdAt: authUser?.user?.created_at ?? null,
        lastSessionAt: authUser?.user?.last_sign_in_at ?? null,
      }),
    );
  } catch (e) {
    console.error(`[${c.get("requestId")}] me:`, (e as Error).message);
    return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
  }
});

// PATCH /v1/accounts/me — cambiar usuario y/o nombre visible.
accountRoutes.patch(
  "/me",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateJson(updateProfileSchema),
  async (c) => {
    const a = authed(c);
    if (!a) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }
    const input = c.req.valid("json");
    const admin = supabaseAdmin(a.env);
    const userClient = supabaseForUser(a.env, a.token);

    try {
      const current = await ownProfile(a.env, a.token, a.user.id);

      if (input.username && (!current || current.username.toLowerCase() !== input.username.toLowerCase())) {
        const { data: taken, error: takenError } = await admin.rpc("username_taken", {
          p_username: input.username,
        });
        if (takenError) throw new Error(`username_check: ${takenError.message}`);
        if (taken === true) {
          return fail(c, { code: "username_taken", message: "Ese usuario ya está en uso.", status: 409 });
        }
      }

      const payload: {
        username?: string;
        display_name?: string;
        bio?: string | null;
        last_mc_version?: string | null;
        mc_uuid?: string | null;
        is_online?: boolean;
      } = {};
      if (input.username) payload.username = input.username;
      if (input.displayName) payload.display_name = input.displayName;
      // Vacío = sin descripción (se guarda null).
      if (input.bio !== undefined) payload.bio = input.bio.trim() === "" ? null : input.bio;
      if (input.lastMcVersion !== undefined) payload.last_mc_version = input.lastMcVersion;
      // UUID de Minecraft: "" desvincula; duplicado → 409.
      if (input.mcUuid !== undefined) payload.mc_uuid = input.mcUuid === "" ? null : input.mcUuid;
      // Booleano interno de presencia (lo cambia el dueño).
      if (input.isOnline !== undefined) payload.is_online = input.isOnline;

      let row: ProfileRow | null;
      if (current) {
        const { data, error } = await userClient
          .from("profiles")
          .update(payload)
          .eq("user_id", a.user.id)
          .select("username, display_name, avatar_url, bio, last_mc_version, mc_uuid, banner_url, is_online")
          .single();
        if (error) {
          if (error.code === "23505") {
            if (/mc_uuid/i.test(error.message)) {
              return fail(c, { code: "mc_uuid_taken", message: "Ese UUID de Minecraft ya está enlazado.", status: 409 });
            }
            return fail(c, { code: "username_taken", message: "Ese usuario ya está en uso.", status: 409 });
          }
          throw new Error(`profile_update: ${error.message}`);
        }
        row = data as ProfileRow;
      } else {
        // Sin fila (cuenta anterior al trigger): se crea con lo recibido.
        const username = payload.username ?? `jugador-${a.user.id.slice(0, 8)}`;
        const { data, error } = await userClient
          .from("profiles")
          .insert({
            user_id: a.user.id,
            username,
            display_name: payload.display_name ?? username,
            bio: payload.bio ?? null,
            last_mc_version: payload.last_mc_version ?? null,
            mc_uuid: payload.mc_uuid ?? null,
            is_online: payload.is_online ?? false,
          })
          .select("username, display_name, avatar_url, bio, last_mc_version, mc_uuid, banner_url, is_online")
          .single();
        if (error) {
          if (error.code === "23505") {
            if (/mc_uuid/i.test(error.message)) {
              return fail(c, { code: "mc_uuid_taken", message: "Ese UUID de Minecraft ya está enlazado.", status: 409 });
            }
            return fail(c, { code: "username_taken", message: "Ese usuario ya está en uso.", status: 409 });
          }
          throw new Error(`profile_create: ${error.message}`);
        }
        row = data as ProfileRow;
      }

      return ok(c, profileShape(a.user.id, a.user.email ?? null, row));
    } catch (e) {
      console.error(`[${c.get("requestId")}] patch me:`, (e as Error).message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
  },
);

// PATCH /v1/accounts/me/presence — cambiar estado en línea propio (booleano interno).
accountRoutes.patch(
  "/me/presence",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateJson(presenceUpdateSchema),
  async (c) => {
    const a = authed(c);
    if (!a) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }
    const input = c.req.valid("json");
    try {
      const { data, error } = await supabaseForUser(a.env, a.token)
        .from("profiles")
        .update({ is_online: input.isOnline })
        .eq("user_id", a.user.id)
        .select("is_online")
        .single();
      if (error) throw new Error(`presence_update: ${error.message}`);
      return ok(c, { isOnline: (data as { is_online: boolean }).is_online });
    } catch (e) {
      console.error(`[${c.get("requestId")}] presence:`, (e as Error).message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
  },
);

// POST /v1/accounts/me/email-change — pedir cambio de email (se confirma en el NUEVO correo).
accountRoutes.post(
  "/me/email-change",
  rateLimit(rateLimitPresets.sensitive),
  requireAuth(),
  validateJson(emailChangeSchema),
  async (c) => {
    const a = authed(c);
    if (!a) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }
    const input = c.req.valid("json");
    if (a.user.email && a.user.email.toLowerCase() === input.newEmail) {
      return fail(c, { code: "same_email", message: "Ya usas ese email.", status: 400 });
    }
    const { error } = await supabaseAdmin(a.env).auth.admin.updateUserById(a.user.id, {
      email: input.newEmail,
    });
    if (error) {
      if (/already/i.test(error.message)) {
        return fail(c, { code: "email_taken", message: "Ese email ya está registrado.", status: 409 });
      }
      console.error(`[${c.get("requestId")}] email-change:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    return ok(c, {
      emailChangeRequested: true,
      message: "Revisa tu nuevo correo para confirmar el cambio.",
    });
  },
);

// POST /v1/accounts/me/password-change — cambiar contraseña verificando la actual.
accountRoutes.post(
  "/me/password-change",
  rateLimit(rateLimitPresets.sensitive),
  requireAuth(),
  validateJson(passwordChangeSchema),
  async (c) => {
    const a = authed(c);
    if (!a) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }
    const input = c.req.valid("json");
    if (!a.user.email) {
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    // Verificar la actual: si no coincide, 401 sin más detalle.
    const anon = createClient(a.env.supabaseUrl, a.env.supabaseAnonKey, {
      auth: { persistSession: false, autoRefreshToken: false },
    });
    const { error: verifyError } = await anon.auth.signInWithPassword({
      email: a.user.email,
      password: input.currentPassword,
    });
    if (verifyError) {
      return fail(c, { code: "current_password_incorrect", message: "La contraseña actual no es correcta.", status: 401 });
    }
    const { error: updateError } = await supabaseAdmin(a.env).auth.admin.updateUserById(a.user.id, {
      password: input.newPassword,
    });
    if (updateError) {
      console.error(`[${c.get("requestId")}] password-change:`, updateError.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    return ok(c, { passwordChanged: true });
  },
);

// POST /v1/accounts/me/avatar — subir avatar (imagen ≤ 2 MB, bucket público).
accountRoutes.post("/me/avatar", rateLimit(rateLimitPresets.account), requireAuth(), async (c) => {
  const a = authed(c);
  if (!a) {
    return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
  }
  const body = await c.req.parseBody();
  const raw = body["avatar"];
  const file = Array.isArray(raw) ? raw[0] : raw;
  if (!(file instanceof File) || file.size === 0) {
    return fail(c, { code: "validation_error", message: "Falta la imagen (campo `avatar`).", status: 400 });
  }
  if (file.size > AVATAR_MAX_BYTES) {
    return fail(c, { code: "file_too_large", message: "Máximo 2 MB.", status: 400 });
  }
  const kind = await sniffAvatar(file);
  if (!kind) {
    return fail(c, { code: "invalid_file_type", message: "Solo PNG, JPEG o WebP.", status: 400 });
  }
  const mime = kind === "png" ? "image/png" : kind === "jpg" ? "image/jpeg" : "image/webp";
  const path = `${a.user.id}/avatar.${kind}`;
  const userClient = supabaseForUser(a.env, a.token);

  const { error: upError } = await userClient.storage
    .from("avatars")
    .upload(path, file, { contentType: mime, upsert: true });
  if (upError) {
    console.error(`[${c.get("requestId")}] avatar: error storage:`, upError.message);
    return fail(c, { code: "upload_failed", message: "No se pudo guardar el avatar.", status: 500 });
  }

  const publicUrl = userClient.storage.from("avatars").getPublicUrl(path).data.publicUrl;

  // Registrar en el libro y enlazar al perfil (best-effort: el avatar ya quedó guardado).
  await userClient.from("file_uploads").insert({
    user_id: a.user.id,
    kind: "avatar",
    bucket: "avatars",
    path,
    size_bytes: file.size,
    mime,
  });
  await userClient.from("profiles").update({ avatar_url: publicUrl }).eq("user_id", a.user.id);

  return ok(c, { avatarUrl: publicUrl });
});

interface PrivacyRow {
  searchable: boolean;
  allow_email_search: boolean;
  receive_friend_requests: boolean;
  receive_notifications: boolean;
}

const DEFAULT_PRIVACY = {
  searchable: true,
  allowEmailSearch: true,
  receiveFriendRequests: true,
  receiveNotifications: true,
};

function privacyShape(row: PrivacyRow | null) {
  if (!row) return { ...DEFAULT_PRIVACY };
  return {
    searchable: row.searchable,
    allowEmailSearch: row.allow_email_search,
    receiveFriendRequests: row.receive_friend_requests,
    receiveNotifications: row.receive_notifications,
  };
}

// GET /v1/accounts/me/privacy — ver opciones de privacidad propias.
accountRoutes.get("/me/privacy", rateLimit(rateLimitPresets.account), requireAuth(), async (c) => {
  const a = authed(c);
  if (!a) {
    return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
  }
  const { data, error } = await supabaseForUser(a.env, a.token)
    .from("privacy_settings")
    .select("searchable, allow_email_search, receive_friend_requests, receive_notifications")
    .eq("user_id", a.user.id)
    .maybeSingle();
  if (error) {
    console.error(`[${c.get("requestId")}] privacy get:`, error.message);
    return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
  }
  return ok(c, privacyShape((data as PrivacyRow | null) ?? null));
});

// PATCH /v1/accounts/me/privacy — cambiar opciones de privacidad propias.
accountRoutes.patch(
  "/me/privacy",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateJson(privacyUpdateSchema),
  async (c) => {
    const a = authed(c);
    if (!a) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }
    const input = c.req.valid("json");
    const payload: Partial<Omit<PrivacyRow, never>> & { user_id: string } = { user_id: a.user.id };
    if (input.searchable !== undefined) payload.searchable = input.searchable;
    if (input.allowEmailSearch !== undefined) payload.allow_email_search = input.allowEmailSearch;
    if (input.receiveFriendRequests !== undefined) payload.receive_friend_requests = input.receiveFriendRequests;
    if (input.receiveNotifications !== undefined) payload.receive_notifications = input.receiveNotifications;

    const { data, error } = await supabaseForUser(a.env, a.token)
      .from("privacy_settings")
      .upsert(payload, { onConflict: "user_id" })
      .select("searchable, allow_email_search, receive_friend_requests, receive_notifications")
      .single();
    if (error || !data) {
      console.error(`[${c.get("requestId")}] privacy patch:`, error?.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    return ok(c, privacyShape(data as PrivacyRow));
  },
);

/** Banner del perfil: PNG/GIF/JPG/WEBP de hasta 1920×1080 (máximo 1080p). */
const BANNER_WIDTH = 1920;
const BANNER_HEIGHT = 1080;
const BANNER_MAX_BYTES = 8 * 1024 * 1024;

const BANNER_MIME = {
  png: "image/png",
  gif: "image/gif",
  jpg: "image/jpeg",
  webp: "image/webp",
} as const;

// POST /v1/accounts/me/banner — subir banner (hasta 1080p, bucket público).
accountRoutes.post("/me/banner", rateLimit(rateLimitPresets.account), requireAuth(), async (c) => {
  const a = authed(c);
  if (!a) {
    return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
  }
  const body = await c.req.parseBody();
  const raw = body["banner"];
  const file = Array.isArray(raw) ? raw[0] : raw;
  if (!(file instanceof File) || file.size === 0) {
    return fail(c, { code: "validation_error", message: "Falta la imagen (campo `banner`).", status: 400 });
  }
  if (file.size > BANNER_MAX_BYTES) {
    return fail(c, { code: "file_too_large", message: "Máximo 8 MB.", status: 400 });
  }
  const info = await imageInfo(file);
  if (!info) {
    return fail(c, { code: "invalid_file_type", message: "Solo PNG, GIF, JPG o WEBP.", status: 400 });
  }
  if (info.width > BANNER_WIDTH || info.height > BANNER_HEIGHT) {
    return fail(c, {
      code: "invalid_dimensions",
      message: `El banner no puede superar los ${BANNER_WIDTH}×${BANNER_HEIGHT}.`,
      status: 400,
    });
  }

  const path = `${a.user.id}/banner.${info.format}`;
  const mime = BANNER_MIME[info.format];
  const userClient = supabaseForUser(a.env, a.token);

  const { error: upError } = await userClient.storage
    .from("banners")
    .upload(path, file, { contentType: mime, upsert: true });
  if (upError) {
    console.error(`[${c.get("requestId")}] banner: error storage:`, upError.message);
    return fail(c, { code: "upload_failed", message: "No se pudo guardar el banner.", status: 500 });
  }

  const publicUrl = userClient.storage.from("banners").getPublicUrl(path).data.publicUrl;

  await userClient.from("file_uploads").insert({
    user_id: a.user.id,
    kind: "banner",
    bucket: "banners",
    path,
    size_bytes: file.size,
    mime,
  });
  await userClient.from("profiles").update({ banner_url: publicUrl }).eq("user_id", a.user.id);

  return ok(c, { bannerUrl: publicUrl });
});

/** Marca equipado/desequipado un cosmético propio (debe poseerlo). */
async function setEquipped(c: Context<AppEnv>, cosmeticId: string, equipped: boolean) {
  const a = authed(c);
  if (!a) {
    return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
  }
  const userClient = supabaseForUser(a.env, a.token);
  const { data, error } = await userClient
    .from("user_cosmetics")
    .update({ equipped })
    .eq("user_id", a.user.id)
    .eq("cosmetic_id", cosmeticId)
    .select("cosmetic_id, equipped")
    .maybeSingle();
  if (error) {
    console.error(`[${c.get("requestId")}] cosmetics:`, error.message);
    return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
  }
  if (!data) {
    return fail(c, { code: "cosmetic_not_owned", message: "No posees ese cosmético.", status: 404 });
  }
  const row = data as { cosmetic_id: string; equipped: boolean };
  return ok(c, { cosmeticId: row.cosmetic_id, equipped: row.equipped });
}

// POST /v1/accounts/me/cosmetics/equip — equipar un cosmético propio.
accountRoutes.post(
  "/me/cosmetics/equip",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateJson(cosmeticEquipSchema),
  (c) => setEquipped(c, c.req.valid("json").cosmeticId, true),
);

// POST /v1/accounts/me/cosmetics/unequip — desequipar un cosmético propio.
accountRoutes.post(
  "/me/cosmetics/unequip",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateJson(cosmeticEquipSchema),
  (c) => setEquipped(c, c.req.valid("json").cosmeticId, false),
);
