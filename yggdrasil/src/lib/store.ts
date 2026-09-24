// Acceso a datos Yggdrasil sobre Supabase (la misma base que usa `api/`).
// - Usuarios: `auth.users` vía `signInWithPassword` (verifica contraseña real).
// - Perfiles: `public.profiles` (`username` + `mc_uuid`, únicos).
// - Tokens/sesiones: tablas propias `ygg_tokens` / `ygg_joins` (ver `db/`).
// Todo acceso usa `service_role` en servidor y nunca expone filas crudas.

import { createClient, type SupabaseClient } from "@supabase/supabase-js";
import type { YggEnvSafe } from "../env";
import type { McProfile, YggToken } from "../types/app";
import { hyphenateUuid, joinExpiryIso, newAccessToken, tokenExpiryIso } from "./crypto";

export interface LoginOk {
  userId: string;
  email: string | null;
}

/** Cliente administrativo (bypass RLS). Solo en servidor, jamás al cliente. */
export function adminClient(env: YggEnvSafe): SupabaseClient {
  return createClient(env.supabaseUrl, env.supabaseServiceRoleKey, {
    auth: { persistSession: false, autoRefreshToken: false },
  });
}

/**
 * Verifica email+contraseña contra Supabase Auth.
 * Devuelve `{ userId, email }` o `null` (credenciales inválidas o sin confirmar).
 * No revela el motivo: el llamador responde 403 genérico.
 */
export async function verifyCredentials(
  env: YggEnvSafe,
  email: string,
  password: string,
): Promise<LoginOk | null> {
  const anon = createClient(env.supabaseUrl, env.supabaseAnonKey, {
    auth: { persistSession: false, autoRefreshToken: false },
  });
  const { data, error } = await anon.auth.signInWithPassword({ email, password });
  if (error || !data.user || !data.session) return null;
  // Sin email confirmado no se juega (igual que `api/` exige confirmación).
  if (!data.user.email_confirmed_at && !data.user.confirmed_at) return null;
  return { userId: data.user.id, email: data.user.email ?? null };
}

/**
 * Resuelve un identificador (email o nombre de jugador) a email.
 * Reutiliza la función `email_for_username` creada por `api/` (revocada salvo
 * `service_role`, por eso solo funciona con el cliente admin).
 */
export async function resolveEmail(env: YggEnvSafe, identifier: string): Promise<string | null> {
  const id = identifier.trim();
  if (id.includes("@")) return id.toLowerCase();
  const admin = adminClient(env);
  const { data, error } = await admin.rpc("email_for_username", { p_username: id });
  if (error || !data) return null;
  return String(data).toLowerCase();
}

/** Lee el perfil StepLauncher (`username` + `mc_uuid`) de un usuario. */
export async function fetchMcProfile(env: YggEnvSafe, userId: string): Promise<McProfile | null> {
  const admin = adminClient(env);
  const { data, error } = await admin
    .from("profiles")
    .select("user_id, username, mc_uuid")
    .eq("user_id", userId)
    .maybeSingle();
  if (error || !data) return null;
  const row = data as { user_id?: unknown; username?: unknown; mc_uuid?: unknown };
  const username = row.username;
  const mcUuid = row.mc_uuid;
  const owner = row.user_id;
  if (
    typeof username !== "string" ||
    typeof mcUuid !== "string" ||
    typeof owner !== "string" ||
    !username ||
    !mcUuid
  ) {
    return null;
  }
  return { uuid: mcUuid, name: username, userId: owner };
}

/** Busca perfiles por nombre exacto (para login con nombre de perfil). */
export async function fetchProfileByName(env: YggEnvSafe, name: string): Promise<McProfile | null> {
  const admin = adminClient(env);
  const { data, error } = await admin
    .from("profiles")
    .select("user_id, username, mc_uuid")
    .ilike("username", name)
    .limit(2);
  if (error || !data || !Array.isArray(data) || data.length !== 1) return null;
  const row = data[0] as { user_id?: unknown; username?: unknown; mc_uuid?: unknown };
  if (typeof row.username !== "string" || typeof row.mc_uuid !== "string") return null;
  if (typeof row.user_id !== "string") return null;
  if (row.username.toLowerCase() !== name.toLowerCase()) return null;
  return { uuid: row.mc_uuid, name: row.username, userId: row.user_id };
}

/** Busca un perfil por UUID-MC (con o sin guiones). */
export async function fetchProfileByUuid(env: YggEnvSafe, uuid: string): Promise<McProfile | null> {
  const admin = adminClient(env);
  const { data, error } = await admin
    .from("profiles")
    .select("user_id, username, mc_uuid")
    .eq("mc_uuid", hyphenateUuid(uuid))
    .maybeSingle();
  if (error || !data) return null;
  const row = data as { user_id?: unknown; username?: unknown; mc_uuid?: unknown };
  if (typeof row.username !== "string" || typeof row.mc_uuid !== "string") return null;
  if (typeof row.user_id !== "string") return null;
  return { uuid: row.mc_uuid, name: row.username, userId: row.user_id };
}

/** Búsqueda por lotes para `POST /api/profiles/minecraft` (máx. 10). */
export async function fetchProfilesByNames(env: YggEnvSafe, names: string[]): Promise<McProfile[]> {
  const clean = [...new Set(names.map((n) => n.trim()).filter((n) => n.length > 0))].slice(0, 10);
  if (clean.length === 0) return [];
  const lowered = clean.map((n) => n.toLowerCase());
  const admin = adminClient(env);
  const { data, error } = await admin
    .from("profiles")
    .select("user_id, username, mc_uuid")
    .in("username", clean);
  if (error || !data || !Array.isArray(data)) return [];
  const out: McProfile[] = [];
  for (const row of data as Array<{ user_id?: unknown; username?: unknown; mc_uuid?: unknown }>) {
    if (typeof row.username !== "string" || typeof row.mc_uuid !== "string") continue;
    if (typeof row.user_id !== "string") continue;
    if (!lowered.includes(row.username.toLowerCase())) continue;
    out.push({ uuid: row.mc_uuid, name: row.username, userId: row.user_id });
  }
  return out;
}

/** Cuenta tokens vivos de un usuario (para aplicar el límite de 10). */
export async function countLiveTokens(env: YggEnvSafe, userId: string): Promise<number> {
  const admin = adminClient(env);
  const { count } = await admin
    .from("ygg_tokens")
    .select("access_token", { count: "exact", head: true })
    .eq("user_id", userId)
    .eq("revoked", false)
    .gt("expires_at", new Date().toISOString());
  return count ?? 0;
}

/** Revoca los tokens más viejos si el usuario supera el límite. */
export async function enforceTokenLimit(env: YggEnvSafe, userId: string, max = 10): Promise<void> {
  const live = await countLiveTokens(env, userId);
  if (live < max) return;
  const admin = adminClient(env);
  const { data } = await admin
    .from("ygg_tokens")
    .select("access_token")
    .eq("user_id", userId)
    .eq("revoked", false)
    .gt("expires_at", new Date().toISOString())
    .order("issued_at", { ascending: true })
    .limit(live - max + 1);
  if (!data || !Array.isArray(data) || data.length === 0) return;
  const tokens = (data as Array<{ access_token?: unknown }>)
    .map((r) => r.access_token)
    .filter((t): t is string => typeof t === "string");
  if (tokens.length === 0) return;
  await admin.from("ygg_tokens").update({ revoked: true }).in("access_token", tokens);
}

/** Emite un token nuevo (revoca viejos si hace falta). */
export async function issueToken(
  env: YggEnvSafe,
  userId: string,
  clientToken: string,
  profile: McProfile | null,
): Promise<YggToken> {
  await enforceTokenLimit(env, userId);
  const admin = adminClient(env);
  const now = new Date().toISOString();
  const row = {
    access_token: newAccessToken(),
    client_token: clientToken,
    user_id: userId,
    profile_uuid: profile ? hyphenateUuid(profile.uuid) : null,
    profile_name: profile ? profile.name : null,
    issued_at: now,
    expires_at: tokenExpiryIso(),
    revoked: false,
  };
  const { error } = await admin.from("ygg_tokens").insert(row);
  if (error) throw new Error(`no se pudo emitir el token: ${error.message}`);
  return {
    accessToken: row.access_token,
    clientToken: row.client_token,
    userId,
    profileUuid: row.profile_uuid,
    profileName: row.profile_name,
    issuedAt: row.issued_at,
    expiresAt: row.expires_at,
    revoked: false,
  };
}

/** Lee un token vivo por `accessToken` (con o sin `clientToken`). */
export async function findLiveToken(
  env: YggEnvSafe,
  accessToken: string,
  clientToken?: string,
): Promise<YggToken | null> {
  // Los launchers pueden reenviar tokens con guiones: se comparan sin ellos.
  const cleanAccess = accessToken.replace(/-/g, "");
  const cleanClient = clientToken?.replace(/-/g, "");
  const admin = adminClient(env);
  let q = admin.from("ygg_tokens").select("*").eq("access_token", cleanAccess).limit(1);
  if (cleanClient) q = q.eq("client_token", cleanClient);
  const { data, error } = await q.maybeSingle();
  if (error || !data) return null;
  const r = data as Record<string, unknown>;
  if (r["revoked"] === true) return null;
  const expiresAt = r["expires_at"];
  if (typeof expiresAt !== "string" || Date.parse(expiresAt) <= Date.now()) return null;
  if (typeof r["access_token"] !== "string" || typeof r["client_token"] !== "string") return null;
  if (typeof r["user_id"] !== "string") return null;
  return {
    accessToken: r["access_token"] as string,
    clientToken: r["client_token"] as string,
    userId: r["user_id"] as string,
    profileUuid: typeof r["profile_uuid"] === "string" ? (r["profile_uuid"] as string) : null,
    profileName: typeof r["profile_name"] === "string" ? (r["profile_name"] as string) : null,
    issuedAt: typeof r["issued_at"] === "string" ? (r["issued_at"] as string) : "",
    expiresAt: expiresAt as string,
    revoked: false,
  };
}

/** Revoca un token concreto (siempre éxito de cara al protocolo). */
export async function revokeToken(env: YggEnvSafe, accessToken: string): Promise<void> {
  const admin = adminClient(env);
  await admin.from("ygg_tokens").update({ revoked: true }).eq("access_token", accessToken.replace(/-/g, ""));
}

/** Revoca todos los tokens vivos de un usuario (`signout`). */
export async function revokeAllForUser(env: YggEnvSafe, userId: string): Promise<void> {
  const admin = adminClient(env);
  await admin.from("ygg_tokens").update({ revoked: true }).eq("user_id", userId).eq("revoked", false);
}

/** Rota un token: revoca el viejo y emite uno nuevo con el mismo `clientToken`. */
export async function rotateToken(
  env: YggEnvSafe,
  old: YggToken,
  profile: McProfile | null,
): Promise<YggToken> {
  await revokeToken(env, old.accessToken);
  return issueToken(env, old.userId, old.clientToken, profile);
}

/** Registra un `join` del cliente (vive 30 s, clave = `serverId`). */
export async function recordJoin(
  env: YggEnvSafe,
  serverId: string,
  token: YggToken,
  clientIp: string,
): Promise<void> {
  const admin = adminClient(env);
  await admin.from("ygg_joins").upsert(
    {
      server_id: serverId,
      access_token: token.accessToken,
      profile_uuid: token.profileUuid,
      profile_name: token.profileName,
      client_ip: clientIp || "unknown",
      created_at: new Date().toISOString(),
      expires_at: joinExpiryIso(),
    },
    { onConflict: "server_id" },
  );
  // El spec manda: solo el ÚLTIMO join del jugador vale. Se borran los anteriores
  // (best-effort: si falla la limpieza, el hasJoined igual filtra por serverId).
  if (token.profileName) {
    await admin
      .from("ygg_joins")
      .delete()
      .eq("profile_name", token.profileName)
      .neq("server_id", serverId);
  }
}

/** Verifica un `hasJoined` del servidor: devuelve el perfil si la sesión vale. */
export async function checkJoin(
  env: YggEnvSafe,
  username: string,
  serverId: string,
  serverIp: string,
): Promise<McProfile | null> {
  const admin = adminClient(env);
  const { data, error } = await admin
    .from("ygg_joins")
    .select("access_token, profile_uuid, profile_name")
    .eq("server_id", serverId)
    .gt("expires_at", new Date().toISOString())
    .maybeSingle();
  if (error || !data) return null;
  const r = data as { profile_name?: unknown; profile_uuid?: unknown; access_token?: unknown };
  // El spec manda username case-insensitive (así lo compara Mojang).
  if (typeof r.profile_name !== "string" || r.profile_name.toLowerCase() !== username.toLowerCase()) {
    return null;
  }
  if (typeof r.profile_uuid !== "string" || typeof r.access_token !== "string") return null;
  // El token debe seguir vivo (si se revocó tras el join, no entra).
  const live = await findLiveToken(env, r.access_token);
  if (!live || !live.profileUuid || !live.profileName) return null;
  if (live.profileName.toLowerCase() !== username.toLowerCase()) return null;
  // Auditoría (best-effort): marca el join como verificado y desde qué servidor.
  // Si falla (ej: faltan las columnas por no re-pegar el SQL), se registra y
  // la verificación igual vale: el juego entra igual.
  const { error: auditError } = await admin
    .from("ygg_joins")
    .update({ verified_at: new Date().toISOString(), server_ip: serverIp || "unknown" })
    .eq("server_id", serverId);
  if (auditError) {
    console.error("checkJoin: no se pudo auditar la verificación:", auditError.message);
  }
  return { uuid: r.profile_uuid, name: r.profile_name, userId: live.userId };
}
