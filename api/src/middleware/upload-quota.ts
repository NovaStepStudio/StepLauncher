// Quota diaria de subidas (skins + capas): máximo 10 archivos por día y usuario.
// Se calcula contando el libro `file_uploads` (RLS: solo filas propias) de las
// últimas 24h. Es aproximada bajo carreras simultáneas, suficiente para quota.
// Usar DESPUÉS de `requireAuth()` (necesita el Bearer y `c.get("user")`).

import type { MiddlewareHandler } from "hono";
import type { AppEnv } from "../types/app";
import { getEnv } from "../env";
import { fail } from "../lib/respond";
import { supabaseForUser } from "../lib/supabase";
import { getBearerToken } from "./auth";

/** Máximo de skins+capas por ventana de 24h. Cambiarlo es decisión con ADR. */
export const MAX_UPLOADS_PER_DAY = 10;
/** Ventana deslizante de quota en milisegundos. */
const QUOTA_WINDOW_MS = 24 * 60 * 60 * 1000;

export const uploadQuota = (): MiddlewareHandler<AppEnv> => {
  return async (c, next) => {
    const user = c.get("user");
    const token = getBearerToken(c);
    if (!user || !token) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }

    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }

    const since = new Date(Date.now() - QUOTA_WINDOW_MS).toISOString();
    const supabase = supabaseForUser(env, token);
    const { count, error } = await supabase
      .from("file_uploads")
      .select("id", { count: "exact", head: true })
      .eq("user_id", user.id)
      .in("kind", ["skin", "cape"])
      .gte("created_at", since);

    if (error) {
      console.error(`[${c.get("requestId")}] quota: error contando subidas:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }

    if ((count ?? 0) >= MAX_UPLOADS_PER_DAY) {
      return fail(c, {
        code: "upload_quota_exceeded",
        message: `Límite diario alcanzado (${MAX_UPLOADS_PER_DAY} archivos/día). Inténtalo mañana.`,
        status: 429,
      });
    }

    await next();
  };
};
