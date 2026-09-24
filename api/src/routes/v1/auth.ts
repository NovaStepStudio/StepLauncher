// Autenticación v1: registro con confirmación, login, renovación y cierre.
// Las SESIONES las emite Supabase Auth (access + refresh tokens); la API no
// inventa criptografía (ver ADR-003). Cada token solo da acceso a SU cuenta
// porque todo endpoint posterior usa `requireAuth()` + RLS (auth.uid()).
// Correos Supabase (plantillas del panel): Confirm signup, Change email,
// Reset password + avisos Password/Email changed (ver ADR-015 y docs/oauth).

import { Hono } from "hono";
import { createClient, type Session } from "@supabase/supabase-js";
import type { AppEnv } from "../../types/app";
import { emailCallbackUrl, getEnv, type SafeEnv } from "../../env";
import { ok, fail } from "../../lib/respond";
import { supabaseAdmin, supabaseForUser } from "../../lib/supabase";
import { requireAuth, getBearerToken } from "../../middleware/auth";
import { rateLimit, rateLimitPresets } from "../../middleware/rate-limit";
import { validateJson } from "../../middleware/validate";
import {
  confirmSchema,
  loginSchema,
  recoverSchema,
  recoveryPasswordSchema,
  refreshSchema,
  registerSchema,
  resendSchema,
  resetPasswordSchema,
} from "../../schemas/v1/auth";

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

// POST /v1/auth/register — alta con email + usuario + contraseña (CON confirmación).
// Exige confirmar el correo (plantilla Supabase "Confirm sign up"): NO devuelve
// sesión hasta verificar. El trigger `handle_new_user` crea el perfil igual.
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

    // Alta vía signUp para que Supabase envíe "Confirm sign up" al correo.
    const emailRedirectTo = emailCallbackUrl(env.siteUrl);
    const { data: created, error: createError } = await anonClient(env).auth.signUp({
      email: input.email,
      password: input.password,
      options: {
        data: { username: input.username },
        ...(emailRedirectTo ? { emailRedirectTo } : {}),
      },
    });
    if (createError || !created.user) {
      const msg = createError?.message ?? "";
      if (/already|registered|exists/i.test(msg)) {
        return fail(c, { code: "email_taken", message: "Ese email ya está registrado.", status: 409 });
      }
      if (/username_taken/.test(msg)) {
        return fail(c, { code: "username_taken", message: "Ese usuario ya está en uso.", status: 409 });
      }
      if (/password/i.test(msg)) {
        return fail(c, { code: "weak_password", message: "Esa contraseña no es válida.", status: 400 });
      }
      console.error(`[${c.get("requestId")}] register: error signUp:`, msg);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }

    // Supabase no revela si el email ya existe: sin identidades = ya registrado.
    const identities = (created.user as { identities?: unknown[] }).identities;
    if (Array.isArray(identities) && identities.length === 0) {
      return fail(c, { code: "email_taken", message: "Ese email ya está registrado.", status: 409 });
    }

    // Si el proyecto aún tiene confirmación desactivada, llega sesión inmediata
    // (compatibilidad transitoria): se devuelve como antes para no romper.
    if (created.session) {
      const username = (await fetchUsername(admin, created.user.id)) ?? input.username;
      return ok(
        c,
        {
          user: { id: created.user.id, email: input.email, username },
          session: sessionShape(created.session),
        },
        201,
      );
    }

    // Caso normal con "Confirm email" activo: pendiente de verificación.
    const username = (await fetchUsername(admin, created.user.id)) ?? input.username;
    return ok(
      c,
      {
        user: { id: created.user.id, email: input.email, username },
        session: null,
        emailConfirmationRequired: true,
        message: "Revisa tu correo para confirmar tu cuenta.",
      },
      201,
    );
  },
);

// POST /v1/auth/login — email O usuario + contraseña. 401 genérico salvo no confirmado.
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
    if (signError || !signed.session || !signed.user) {
      // Sin confirmar: código propio para que la web ofrezca reenviar.
      if (signError && /confirm/i.test(signError.message)) {
        return fail(c, {
          code: "email_not_confirmed",
          message: "Confirma tu correo antes de entrar. Revisa tu bandeja.",
          status: 403,
        });
      }
      return invalid();
    }

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

// POST /v1/auth/resend — reenviar "Confirm sign up" o "Change email address".
// Siempre éxito genérico para no enumerar correos (el fallo real queda en logs).
authRoutes.post(
  "/resend",
  rateLimit(rateLimitPresets.sensitive),
  validateJson(resendSchema),
  async (c) => {
    const input = c.req.valid("json");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }
    const emailRedirectTo = emailCallbackUrl(env.siteUrl);
    const { error } = await anonClient(env).auth.resend({
      type: input.type,
      email: input.email,
      options: emailRedirectTo ? { emailRedirectTo } : undefined,
    });
    if (error) {
      console.error(`[${c.get("requestId")}] resend (${input.type}):`, error.message);
    }
    return ok(c, {
      resent: true,
      message: "Si ese correo está pendiente, recibirás un nuevo enlace.",
    });
  },
);

// POST /v1/auth/recover — pedir "Reset password" (enlace o código al correo).
// Respuesta genérica siempre: no revela si el email existe.
authRoutes.post(
  "/recover",
  rateLimit(rateLimitPresets.sensitive),
  validateJson(recoverSchema),
  async (c) => {
    const input = c.req.valid("json");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }
    const redirectTo = emailCallbackUrl(env.siteUrl);
    const { error } = await anonClient(env).auth.resetPasswordForEmail(input.email, {
      ...(redirectTo ? { redirectTo } : {}),
    });
    if (error) {
      console.error(`[${c.get("requestId")}] recover:`, error.message);
    }
    return ok(c, {
      recoverySent: true,
      message: "Si ese correo está registrado, recibirás un enlace para restablecer.",
    });
  },
);

/** Verifica OTP por código (email+token) o por hash del enlace (token_hash). */
async function verifyEmailOtp(
  env: SafeEnv,
  input: { email: string; token: string; type: "signup" | "email_change" | "recovery" },
) {
  const anon = anonClient(env);
  const byCode = await anon.auth.verifyOtp({
    email: input.email,
    token: input.token,
    type: input.type,
  });
  if (!byCode.error && byCode.data.session && byCode.data.user) return byCode;
  // Enlace con token_hash largo: reintentar como hash (los enlaces de
  // Supabase llegan a /auth/callback con token/hash según plantilla).
  const byHash = await anon.auth.verifyOtp({
    token_hash: input.token,
    type: input.type,
  } as { token_hash: string; type: "signup" | "email_change" | "recovery" });
  return byHash;
}

// POST /v1/auth/confirm — confirmar registro/cambio-email/recupero con el código.
// Para signup/email_change devuelve sesión inmediata; recovery solo verifica.
authRoutes.post(
  "/confirm",
  rateLimit(rateLimitPresets.sensitive),
  validateJson(confirmSchema),
  async (c) => {
    const input = c.req.valid("json");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }
    const { data, error } = await verifyEmailOtp(env, input);
    if (error || !data.user) {
      return fail(c, {
        code: "invalid_code",
        message: "Código inválido o expirado. Pide un nuevo enlace.",
        status: 400,
      });
    }
    if (input.type === "recovery") {
      // El recupero se completa en /reset-password (no se inicia sesión aquí).
      return ok(c, { recoveryVerified: true });
    }
    const admin = supabaseAdmin(env);
    const username = (await fetchUsername(admin, data.user.id)) ?? null;
    if (!data.session) {
      return ok(c, {
        user: { id: data.user.id, email: data.user.email ?? null, username },
        session: null,
        message: "Correo confirmado. Ya puedes entrar.",
      });
    }
    return ok(c, {
      user: { id: data.user.id, email: data.user.email ?? null, username },
      session: sessionShape(data.session),
    });
  },
);

// POST /v1/auth/reset-password — fijar contraseña nueva con el código de recupero.
// Verifica el OTP de recovery y actualiza vía admin (el código prueba el correo).
authRoutes.post(
  "/reset-password",
  rateLimit(rateLimitPresets.sensitive),
  validateJson(resetPasswordSchema),
  async (c) => {
    const input = c.req.valid("json");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }
    const { data, error } = await verifyEmailOtp(env, {
      email: input.email,
      token: input.token,
      type: "recovery",
    });
    if (error || !data.user) {
      return fail(c, {
        code: "invalid_code",
        message: "Código inválido o expirado. Pide un nuevo enlace.",
        status: 400,
      });
    }
    const { error: updateError } = await supabaseAdmin(env).auth.admin.updateUserById(data.user.id, {
      password: input.newPassword,
    });
    if (updateError) {
      if (/password/i.test(updateError.message)) {
        return fail(c, { code: "weak_password", message: "Esa contraseña no es válida.", status: 400 });
      }
      console.error(`[${c.get("requestId")}] reset-password:`, updateError.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    return ok(c, { passwordReset: true });
  },
);

// POST /v1/auth/recovery-password — fijar contraseña con la sesión del enlace.
// El Bearer ES el access_token que Supabase pone en el hash (`#access_token=…&type=recovery`):
// esa sesión prueba el correo, así que no se pide email ni código (nada que pegar).
// Actualiza con la identidad del token (sin admin, sin email del cliente).
authRoutes.post(
  "/recovery-password",
  rateLimit(rateLimitPresets.sensitive),
  requireAuth(),
  validateJson(recoveryPasswordSchema),
  async (c) => {
    const token = getBearerToken(c);
    if (!token || !c.get("user")) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }
    const input = c.req.valid("json");
    const { error } = await supabaseForUser(env, token).auth.updateUser({
      password: input.newPassword,
    });
    if (error) {
      // Clave débil o reutilizada (Supabase la rechaza con varios mensajes).
      if (/password|weak|breach|pwned|leak|compromised|commonly|same|identical|short/i.test(error.message)) {
        return fail(c, { code: "weak_password", message: "Esa contraseña no es válida.", status: 400 });
      }
      // Sesión de recupero vencida o ya usada: pedir un enlace nuevo.
      if (/expir|invalid|jwt|session|not found/i.test(error.message)) {
        return fail(c, {
          code: "recovery_expired",
          message: "Enlace vencido o ya usado. Pedí uno nuevo.",
          status: 400,
        });
      }
      console.error(`[${c.get("requestId")}] recovery-password:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    return ok(c, { passwordReset: true });
  },
);
