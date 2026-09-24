// Límite de peticiones por IP y ventana fija (en memoria por instancia).
// Protege `authenticate`/`signout` de fuerza bruta sin añadir dependencias.
// Los errores usan forma Yggdrasil (no el sobre de `api/`).

import type { MiddlewareHandler } from "hono";
import type { YggEnv } from "../types/app";
import { yggError } from "../lib/ygg";

interface Bucket {
  count: number;
  resetAt: number;
}

const buckets = new Map<string, Bucket>();

function clientIp(c: { req: { header: (n: string) => string | undefined } }): string {
  return c.req.header("CF-Connecting-IP")?.trim() || "unknown";
}

export interface RateLimitOptions {
  /** Ventana en milisegundos. */
  windowMs: number;
  /** Máximo de peticiones por ventana e IP. */
  max: number;
}

/** Presets por sensibilidad del endpoint Yggdrasil. */
export const rateLimitPresets = {
  // Lecturas pesadas de juego/clientes (texturas, blockedservers).
  public: { windowMs: 60_000, max: 120 },
  // Login y cierre global: muy estrictos (fuerza bruta de contraseñas).
  sensitive: { windowMs: 10 * 60_000, max: 20 },
  // Operaciones con token: estrictas por defecto.
  account: { windowMs: 60_000, max: 30 },
  // `join` del cliente: el juego lo llama poco; 6/30s como Mojang.
  join: { windowMs: 30_000, max: 6 },
} satisfies Record<string, RateLimitOptions>;

export const rateLimit = (opts: RateLimitOptions): MiddlewareHandler<YggEnv> => {
  return async (c, next) => {
    const now = Date.now();
    const key = `${clientIp(c)}:${c.req.path}`;
    const current = buckets.get(key);

    if (!current || current.resetAt <= now) {
      buckets.set(key, { count: 1, resetAt: now + opts.windowMs });
      return next();
    }

    if (current.count >= opts.max) {
      const retryAfter = Math.max(1, Math.ceil((current.resetAt - now) / 1000));
      c.header("Retry-After", String(retryAfter));
      return yggError(
        c,
        429,
        "TooManyRequestsException",
        "Demasiadas peticiones. Inténtalo de nuevo más tarde.",
      );
    }

    current.count += 1;
    return next();
  };
};
