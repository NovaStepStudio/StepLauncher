// `X-Request-Id` por petición (trazabilidad en logs).
// Copia adaptada del patrón de `api/`, sin importar nada de fuera.

import type { MiddlewareHandler } from "hono";
import type { YggEnv } from "../types/app";

export const requestId = (): MiddlewareHandler<YggEnv> => {
  return async (c, next) => {
    const incoming = c.req.header("X-Request-Id")?.trim();
    const id =
      incoming && incoming.length >= 8 && incoming.length <= 128
        ? incoming
        : crypto.randomUUID().replace(/-/g, "");
    c.set("requestId", id);
    c.header("X-Request-Id", id);
    await next();
  };
};
