// CORS permisivo para el protocolo Yggdrasil.
// Lo consumen launchers dedicados y servidores de Minecraft (no navegadores),
// así que se permite cualquier origen sin credenciales. Sin secretos en CORS.

import type { MiddlewareHandler } from "hono";
import { cors } from "hono/cors";
import type { YggEnv } from "../types/app";

export const openCors = (): MiddlewareHandler<YggEnv> => {
  return cors({
    origin: "*",
    allowMethods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
    allowHeaders: ["Content-Type", "Authorization", "X-Request-Id", "X-Authlib-Injector-API-Location"],
    exposeHeaders: ["X-Request-Id", "Retry-After", "X-Authlib-Injector-API-Location"],
    maxAge: 600,
    credentials: false,
  });
};
