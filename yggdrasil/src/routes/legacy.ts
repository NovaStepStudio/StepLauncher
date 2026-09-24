// Skins y capas por nombre, estilo Ely.by + API legacy de Mojang.
// - `GET /skins/MinecraftSkins/<usuario>.png`: forma legacy que redirige el
//   injector (`skins.minecraft.net/...`) para juegos viejos (pre-1.8).
// - `GET /skins/:nombre` y `GET /cloaks/:nombre` (con o sin `.png`): acceso
//   directo por nombre para launchers, plugins y depuración.
// Todo devuelve el PNG tal cual o 404, como hacía Mojang. La subida sigue
// viviendo en la web/`api/` (aquí solo lectura).

import { Hono, type Context } from "hono";
import type { YggEnv } from "../types/app";
import { getEnv, type YggEnvSafe } from "../env";
import { downloadPng, latestUploads } from "../lib/textures";
import { fetchProfileByName } from "../lib/store";
import { rateLimit, rateLimitPresets } from "../middleware/rate-limit";
import { validateParam } from "../middleware/validate";
import { cloakNameSchema, legacySkinParamSchema, skinNameSchema } from "../schemas/ygg";

type TextureKind = "skin" | "cape";

/** PNG actual de un jugador por nombre (case-insensitive) o null. */
async function texturePngByName(
  env: YggEnvSafe,
  rawName: string,
  kind: TextureKind,
): Promise<Uint8Array | null> {
  const name = rawName.replace(/\.png$/i, "");
  if (!/^[A-Za-z0-9_]{3,16}$/.test(name)) return null;
  const profile = await fetchProfileByName(env, name);
  if (!profile) return null;
  const uploads = await latestUploads(env, profile.userId);
  const file = kind === "skin" ? uploads.skin : uploads.cape;
  if (!file) return null;
  return downloadPng(env, file.bucket, file.path);
}

function pngOr404(c: Context<YggEnv>, bytes: Uint8Array | null) {
  if (!bytes) {
    return c.json({ error: "Not Found", errorMessage: "Textura no encontrada." }, 404);
  }
  c.header("Content-Type", "image/png");
  c.header("Cache-Control", "public, max-age=3600");
  return c.body(new Uint8Array(bytes), 200);
}

export const legacyApi = new Hono<YggEnv>();

// Forma legacy Mojang (la redirige el injector para juegos pre-1.8).
legacyApi.get(
  "/MinecraftSkins/:file",
  rateLimit(rateLimitPresets.public),
  validateParam(legacySkinParamSchema),
  async (c) => {
    const param = c.req.valid("param");
    const m = /^([A-Za-z0-9_]{3,16})\.png$/.exec(param.file);
    if (!m || !m[1]) {
      return c.json({ error: "Not Found", errorMessage: "Skin no encontrada." }, 404);
    }
    let env;
    try {
      env = getEnv(c);
    } catch {
      return c.json({ error: "InternalError", errorMessage: "Servicio no disponible." }, 500);
    }
    return pngOr404(c, await texturePngByName(env, m[1], "skin"));
  },
);

// Acceso directo por nombre, estilo Ely.by (con o sin `.png`).
legacyApi.get(
  "/:name",
  rateLimit(rateLimitPresets.public),
  validateParam(skinNameSchema),
  async (c) => {
    const param = c.req.valid("param");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return c.json({ error: "InternalError", errorMessage: "Servicio no disponible." }, 500);
    }
    return pngOr404(c, await texturePngByName(env, param.name, "skin"));
  },
);

export const cloaksApi = new Hono<YggEnv>();

// Capa por nombre, estilo Ely.by (con o sin `.png`).
cloaksApi.get(
  "/:name",
  rateLimit(rateLimitPresets.public),
  validateParam(cloakNameSchema),
  async (c) => {
    const param = c.req.valid("param");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return c.json({ error: "InternalError", errorMessage: "Servicio no disponible." }, 500);
    }
    return pngOr404(c, await texturePngByName(env, param.name, "cape"));
  },
);
