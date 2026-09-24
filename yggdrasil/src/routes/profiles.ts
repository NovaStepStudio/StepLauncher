// Dominio `api` (compatibilidad Mojang): búsqueda de perfiles por nombres.
// POST /api/profiles/minecraft — lote de 1–10 nombres → [{ id, name }].
// Sin `properties` aquí por spec (las texturas van en sessionserver).

import { Hono } from "hono";
import type { YggEnv } from "../types/app";
import { getEnv } from "../env";
import { serializeProfile, yggError } from "../lib/ygg";
import { fetchProfilesByNames } from "../lib/store";
import { rateLimit, rateLimitPresets } from "../middleware/rate-limit";
import { validateJson } from "../middleware/validate";
import { profilesBatchSchema } from "../schemas/ygg";

export const profilesApi = new Hono<YggEnv>();

profilesApi.post(
  "/profiles/minecraft",
  rateLimit(rateLimitPresets.account),
  validateJson(profilesBatchSchema),
  async (c) => {
    const names = c.req.valid("json");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return yggError(c, 500, "InternalError", "Servicio no disponible.");
    }
    const found = await fetchProfilesByNames(env, names);
    return c.json(found.map((p) => serializeProfile(p.uuid, p.name)));
  },
);
