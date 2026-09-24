// Cabeceras duras de seguridad para el Worker Yggdrasil.
// El protocolo lo consumen launchers y servidores (no navegadores),
// pero los metadatos sí pueden leerse desde la web: CSP restrictiva igual.

import type { MiddlewareHandler } from "hono";
import type { YggEnv } from "../types/app";
import { getEnv } from "../env";

export const securityHeaders = (): MiddlewareHandler<YggEnv> => {
  return async (c, next) => {
    await next();
    c.header("X-Content-Type-Options", "nosniff");
    c.header("X-Frame-Options", "DENY");
    c.header("Referrer-Policy", "no-referrer");
    c.header("Permissions-Policy", "camera=(), microphone=(), geolocation=()");
    c.header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'");
    try {
      if (getEnv(c).apiEnv === "production") {
        c.header("Strict-Transport-Security", "max-age=31536000; includeSubDomains");
      }
    } catch {
      // Sin entorno no se puede decidir HSTS: no se añade y listo.
    }
  };
};
