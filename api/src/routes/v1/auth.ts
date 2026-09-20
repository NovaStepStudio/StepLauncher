// Autenticación v1: registro, login, renovación y cierre de sesión.
// Las SESIONES las emite Supabase Auth (access + refresh tokens); la API no
// inventa criptografía (ver ADR-003). Cada token solo da acceso a SU cuenta
// porque todo endpoint posterior usa `requireAuth()` + RLS (auth.uid()).

import { Hono } from "hono";
import { createClient, type Session } from "@supabase/supabase-js";
import type { AppEnv } from "../../types/app";
import { getEnv, type SafeEnv } from "../../env";
import { ok, fail } from "../../lib/respond";
import { supabaseAdmin } from "../../lib/supabase";
import { requireAuth, getBearerToken } from "../../middleware/auth";
import { rateLimit, rateLimitPresets } from "../../middleware/rate-limit";
import { validateJson } from "../../middleware/validate";
import { registerSchema, loginSchema, refreshSchema } from "../../schemas/v1/auth";

export const authRoutes = new Hono<AppEnv>();

/** Cliente anónimo (sin persistencia) para signIn/refresh en el edge. */
function anonClient(env: SafeEnv) {
  return createClient(env.supabaseUrl, env.supabaseAnonKey, {
    auth: { persistSession: false, autoRefreshToken: false },
  });
}

/** Forma pública de una sesión: lo que el launcher guarda en el dispositivo. */
function sessionShape(session: Session) {
  return {
    accessToken: session.access_token,
    refreshToken: session.refresh_token,
    expiresAt: session.expires_at ?? null,
    tokenType: "bearer" as const,
  };
}

/** Perfil mínimo para acompañar la sesión (null si el trigger aún no lo creó). */
async function fetchUsername(admin: ReturnType<typeof supabaseAdmin>, userId: string) {
  const { data } = await admin
    .from("profiles")
    .select("username")
    .eq("user_id", userId)
    .maybeSingle();
  return (data?.username as string | undefined) ?? null;
}

// POST /v1/auth/register — alta con email + usuario + contraseña.
authRoutes.post(
  "/register",
  rateLimit(rateLimitPresets.sensitive),
  validateJson(registerSchema),
  async (c) => {
    const input = c.req.valid("json");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }
    const admin = supabaseAdmin(env);

    // Username único (insensible a mayúsculas) antes de crear nada.
    const { data: taken, error: takenError } = await admin.rpc("username_taken", {
      p_username: input.username,
    });
    if (takenError) {
      console.error(`[${c.get("requestId")}] register: error rpc username_taken:`, takenError.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    if (taken === true) {
      return fail(c, { code: "username_taken", message: "Ese usuario ya está en uso.", status: 409 });
    }

    // Alta en Auth. El trigger `handle_new_user` crea el perfil con el username.
    const { data: created, error: createError } = await admin.auth.admin.createUser({
      email: input.email,
      password: input.password,
      email_confirm: true, // MVP: sin fricción de correo (revisar en ADR-007).
      user_metadata: { username: input.username },
    });
    if (createError || !created.user) {
      const msg = createError?.message ?? "";
      if (/already/i.test(msg)) {
        return fail(c, { code: "email_taken", message: "Ese email ya está registrado.", status: 409 });
      }
      if (/username_taken/.test(msg)) {
        return fail(c, { code: "username_taken", message: "Ese usuario ya está en uso.", status: 409 });
      }
      if (/password/i.test(msg)) {
        return fail(c, { code: "weak_password", message: "Esa contraseña no es válida.", status: 400 });
      }
      console.error(`[${c.get("requestId")}] register: error createUser:`, msg);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }

    // Sesión inmediata para el launcher (access + refresh tokens).
    const { data: signed, error: signError } = await anonClient(env).auth.signInWithPassword({
      email: input.email,
      password: input.password,
    });
    if (signError || !signed.session) {
      console.error(`[${c.get("requestId")}] register: sin sesión tras alta:`, signError?.message);
      return fail(c, { code: "internal_error", message: "Cuenta creada, pero inicia sesión manualmente.", status: 500 });
    }

    const username = (await fetchUsername(admin, created.user.id)) ?? input.username;
    return ok(
      c,
      {
        user: { id: created.user.id, email: input.email, username },
        session: sessionShape(signed.session),
      },
      201,
    );
  },
);

// POST /v1/auth/login — email O usuario + contraseña. 401 genérico siempre.
authRoutes.post(
  "/login",
  rateLimit(rateLimitPresets.sensitive),
  validateJson(loginSchema),
  async (c) => {
    const input = c.req.valid("json");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }
    const admin = supabaseAdmin(env);
    const invalid = () =>
      fail(c, { code: "invalid_credentials", message: "Credenciales inválidas.", status: 401 });

    let email = input.identifier.trim();
    if (!email.includes("@")) {
      const { data: resolved, error: resolveError } = await admin.rpc("email_for_username", {
        p_username: email,
      });
      if (resolveError) {
        console.error(`[${c.get("requestId")}] login: error rpc email_for_username:`, resolveError.message);
        return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
      }
      if (!resolved) return invalid();
      email = resolved as string;
    }

    const { data: signed, error: signError } = await anonClient(env).auth.signInWithPassword({
      email: email.toLowerCase(),
      password: input.password,
    });
    if (signError || !signed.session || !signed.user) return invalid();

    const username = await fetchUsername(admin, signed.user.id);
    return ok(c, {
      user: { id: signed.user.id, email: signed.user.email ?? null, username },
      session: sessionShape(signed.session),
    });
  },
);

// POST /v1/auth/refresh — renueva la sesión con el refresh_token del dispositivo.
authRoutes.post(
  "/refresh",
  rateLimit(rateLimitPresets.sensitive),
  validateJson(refreshSchema),
  async (c) => {
    const input = c.req.valid("json");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }
    const { data, error } = await anonClient(env).auth.refreshSession({
      refresh_token: input.refreshToken,
    });
    if (error || !data.session) {
      return fail(c, {
        code: "invalid_refresh",
        message: "Sesión expirada. Inicia sesión de nuevo.",
        status: 401,
      });
    }
    return ok(c, { session: sessionShape(data.session) });
  },
);

// POST /v1/auth/logout — revoca la sesión actual. El cliente debe borrar sus tokens.
authRoutes.post(
  "/logout",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  async (c) => {
    const token = getBearerToken(c);
    if (!token) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }
    const { error } = await supabaseAdmin(env).auth.admin.signOut(token);
    if (error) {
      console.error(`[${c.get("requestId")}] logout: error signOut:`, error.message);
    }
    return ok(c, { loggedOut: true });
  },
);
