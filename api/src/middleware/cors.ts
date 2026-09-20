// CORS estricto por lista blanca (ALLOWED_ORIGINS).
// Sin wildcard en producción. Sin origen permitido = sin cabecera ACAO.

import { cors } from "hono/cors";
import type { MiddlewareHandler } from "hono";
import type { AppEnv } from "../types/app";
import { parseAllowedOrigins } from "../env";

export const strictCors = (): MiddlewareHandler<AppEnv> => {
  return cors({
    origin: (origin, c) => {
      const allowed = parseAllowedOrigins((c.env as AppEnv["Bindings"]).ALLOWED_ORIGINS);
      if (!origin) return null;
      const clean = origin.trim().replace(/\/+$/, "");
      return allowed.includes(clean) ? clean : null;
    },
    allowMethods: ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"],
    // Solo cabeceras necesarias. Authorization para Bearer de Supabase.
    allowHeaders: ["Content-Type", "Authorization", "X-Request-Id"],
    exposeHeaders: ["X-Request-Id", "Retry-After"],
    maxAge: 600,
    credentials: false,
  });
};
