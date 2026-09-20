// Límite de peticiones por IP y ventana fija (en memoria por instancia).
// Protege login/registro de fuerza bruta sin añadir latencia ni dependencias.
// Límite: para protección distribuida real usar KV/Rate Limit API (ver docs).

import type { MiddlewareHandler } from "hono";
import type { AppEnv } from "../types/app";
import { fail } from "../lib/respond";

interface Bucket {
  count: number;
  resetAt: number;
}

const buckets = new Map<string, Bucket>();

function clientIp(c: { req: { header: (n: string) => string | undefined } }): string {
  // Cloudflare envía la IP real aquí; en local cae a "unknown" (misma cubeta).
  return c.req.header("CF-Connecting-IP")?.trim() || "unknown";
}

export interface RateLimitOptions {
  /** Ventana en milisegundos. */
  windowMs: number;
  /** Máximo de peticiones por ventana e IP. */
  max: number;
}

/** Presets por sensibilidad del endpoint. */
export const rateLimitPresets = {
  // Lecturas públicas poco sensibles.
  public: { windowMs: 60_000, max: 120 },
  // Operaciones de cuenta: estrictas por defecto.
  account: { windowMs: 60_000, max: 30 },
  // Login/registro/recuperación: muy estrictas.
  sensitive: { windowMs: 10 * 60_000, max: 20 },
} satisfies Record<string, RateLimitOptions>;

export const rateLimit = (opts: RateLimitOptions): MiddlewareHandler<AppEnv> => {
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
      return fail(c, {
        code: "rate_limited",
        message: "Demasiadas peticiones. Inténtalo de nuevo más tarde.",
        status: 429,
      });
    }

    current.count += 1;
    return next();
  };
};
