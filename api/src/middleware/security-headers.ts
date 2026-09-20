// Cabeceras de seguridad para respuestas JSON de API.
// Se aplican globalmente en `src/index.ts`.

import type { MiddlewareHandler } from "hono";
import type { AppEnv } from "../types/app";

export const securityHeaders = (): MiddlewareHandler<AppEnv> => {
  return async (c, next) => {
    await next();
    c.header("X-Content-Type-Options", "nosniff");
    c.header("X-Frame-Options", "DENY");
    c.header("Referrer-Policy", "no-referrer");
    c.header("Permissions-Policy", "camera=(), microphone=(), geolocation=()");
    // API JSON sin HTML: cerrar cualquier ejecución en navegador.
    c.header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'");
    // HSTS solo tiene sentido detrás de HTTPS (producción); en local se omite.
    if ((c.env.API_ENV ?? "production") === "production") {
      c.header("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload");
    }
  };
};
