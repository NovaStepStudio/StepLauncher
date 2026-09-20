// Autenticación obligatoria para endpoints de cuenta.
// Valida el Bearer contra Supabase Auth y guarda el usuario en contexto.
// Mensajes 401 genéricos para no filtrar si el token existe, expiró o es inválido.

import { createClient } from "@supabase/supabase-js";
import type { MiddlewareHandler } from "hono";
import type { AppEnv } from "../types/app";
import { getEnv } from "../env";
import { fail } from "../lib/respond";

/** Extrae el Bearer de la cabecera Authorization (sin validarlo). */
export function getBearerToken(c: {
  req: { header: (n: string) => string | undefined };
}): string | null {
  const header = c.req.header("Authorization")?.trim();
  if (!header || !header.toLowerCase().startsWith("bearer ")) return null;
  const token = header.slice(7).trim();
  return token.length > 0 ? token : null;
}

/** Exige sesión válida. Usar en TODO endpoint que toque datos del usuario. */
export const requireAuth = (): MiddlewareHandler<AppEnv> => {
  return async (c, next) => {
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }

    const token = getBearerToken(c);
    if (!token) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }

    // Cliente mínimo sin persistencia para validar el token en el edge.
    const supabase = createClient(env.supabaseUrl, env.supabaseAnonKey, {
      global: { headers: { Authorization: `Bearer ${token}` } },
      auth: { persistSession: false, autoRefreshToken: false },
    });

    const { data, error } = await supabase.auth.getUser(token);
    if (error || !data.user) {
      return fail(c, { code: "unauthorized", message: "Sesión inválida o expirada.", status: 401 });
    }

    c.set("user", { id: data.user.id, email: data.user.email ?? undefined });
    await next();
  };
};
