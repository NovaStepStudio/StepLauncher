// Identificador único por petición para correlacionar logs, errores y soporte.
// Se expone como `X-Request-Id` y viaja en `c.get("requestId")`.

import type { MiddlewareHandler } from "hono";
import type { AppEnv } from "../types/app";

export const requestId = (): MiddlewareHandler<AppEnv> => {
  return async (c, next) => {
    const id = crypto.randomUUID();
    c.set("requestId", id);
    await next();
    c.header("X-Request-Id", id);
  };
};
