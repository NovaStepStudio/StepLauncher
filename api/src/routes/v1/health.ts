// Sonda pública de salud y versión. Sin auth, sin secretos, sin Supabase.
// Sirve para uptime, despliegues y para verificar el versionado.

import { Hono } from "hono";
import type { AppEnv } from "../../types/app";
import { ok } from "../../lib/respond";
import { rateLimit, rateLimitPresets } from "../../middleware/rate-limit";

export const healthRoutes = new Hono<AppEnv>();

healthRoutes.get("/", rateLimit(rateLimitPresets.public), (c) => {
  return ok(c, {
    status: "ok",
    version: "v1",
    env: c.env.API_ENV ?? "production",
    time: new Date().toISOString(),
  });
});
