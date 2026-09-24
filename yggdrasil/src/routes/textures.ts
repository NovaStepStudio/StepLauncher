// Texturas públicas del juego: `GET /textures/:hash`.
// El `:hash` es el SHA-256 del PNG (el cliente lo usa como clave de caché).
// Sin auth: el hash es impredecible y solo se revela por endpoints
// autenticados (`hasJoined`, `profile`), igual que en Mojang.

import { Hono } from "hono";
import type { YggEnv } from "../types/app";
import { getEnv } from "../env";
import { downloadPng, lookupTexture } from "../lib/textures";
import { rateLimit, rateLimitPresets } from "../middleware/rate-limit";
import { validateParam } from "../middleware/validate";
import { textureHashSchema } from "../schemas/ygg";

export const texturesApi = new Hono<YggEnv>();

texturesApi.get(
  "/:hash",
  rateLimit(rateLimitPresets.public),
  validateParam(textureHashSchema),
  async (c) => {
    const param = c.req.valid("param");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return c.json({ error: "InternalError", errorMessage: "Servicio no disponible." }, 500);
    }
    let found;
    try {
      found = await lookupTexture(env, param.hash);
    } catch (err) {
      // Configuración incompleta (falta el SQL): visible en `wrangler tail`.
      console.error("GET /textures:", (err as Error)?.message ?? err);
      return c.json({ error: "InternalError", errorMessage: "Texturas no configuradas." }, 500);
    }
    if (!found) return c.json({ error: "Not Found", errorMessage: "Textura no encontrada." }, 404);
    const bytes = await downloadPng(env, found.bucket, found.path);
    if (!bytes) return c.json({ error: "Not Found", errorMessage: "Textura no encontrada." }, 404);
    // `image/png` obligatorio por spec (anti MIME-sniffing) + caché inmutable:
    // el hash cambia con el contenido, así que la URL es eterna.
    c.header("Content-Type", "image/png");
    c.header("Cache-Control", "public, max-age=31536000, immutable");
    return c.body(new Uint8Array(bytes), 200);
  },
);
