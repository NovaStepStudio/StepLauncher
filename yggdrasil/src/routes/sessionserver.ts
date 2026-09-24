// Dominio sessionserver: entrada a servidores (`join`/`hasJoined`), perfil con
// texturas firmadas y lista de servidores bloqueados.
// La skin/capa se lee de los buckets de `api/` y se sirve por `/textures/:hash`.

import { Hono } from "hono";
import type { YggEnv } from "../types/app";
import { getEnv } from "../env";
import { stripUuid } from "../lib/crypto";
import { invalidToken, serializeProfileFull, yggError } from "../lib/ygg";
import { signedTexturesFor } from "../lib/textures";
import { checkJoin, fetchProfileByUuid, findLiveToken, recordJoin } from "../lib/store";
import { rateLimit, rateLimitPresets } from "../middleware/rate-limit";
import { validateParam, validateQuery, validateJson } from "../middleware/validate";
import {
  hasJoinedSchema,
  joinSchema,
  profileParamSchema,
  profileQuerySchema,
} from "../schemas/ygg";

export const sessionserver = new Hono<YggEnv>();

function clientIp(c: { req: { header: (n: string) => string | undefined } }): string {
  return c.req.header("CF-Connecting-IP")?.trim() || c.req.header("X-Forwarded-For")?.trim() || "unknown";
}

/** Origen público del Worker (`https://host`) para construir URLs de texturas. */
function originOf(c: { req: { url: string } }): string {
  return new URL(c.req.url).origin;
}

// POST /sessionserver/session/minecraft/join — el cliente anuncia la sesión.
sessionserver.post(
  "/session/minecraft/join",
  rateLimit(rateLimitPresets.join),
  validateJson(joinSchema),
  async (c) => {
    const input = c.req.valid("json");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return yggError(c, 500, "InternalError", "Servicio no disponible.");
    }
    const token = await findLiveToken(env, input.accessToken);
    if (!token || !token.profileUuid) return invalidToken(c);
    if (stripUuid(token.profileUuid) !== input.selectedProfile) return invalidToken(c);
    await recordJoin(env, input.serverId, token, clientIp(c));
    return c.body(null, 204);
  },
);

// GET /sessionserver/session/minecraft/hasJoined — el servidor verifica.
sessionserver.get(
  "/session/minecraft/hasJoined",
  rateLimit(rateLimitPresets.account),
  validateQuery(hasJoinedSchema),
  async (c) => {
    const input = c.req.valid("query");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return yggError(c, 500, "InternalError", "Servicio no disponible.");
    }
    const profile = await checkJoin(env, input.username, input.serverId, clientIp(c));
    if (!profile) return c.body(null, 204);
    const signed = await signedTexturesFor(env, {
      userId: profile.userId,
      profileUuid: profile.uuid,
      profileName: profile.name,
      origin: originOf(c),
    });
    return c.json(serializeProfileFull(profile.uuid, profile.name, signed?.value, signed?.signature));
  },
);

// GET /sessionserver/session/minecraft/profile/:uuid — perfil (+textures si hay).
sessionserver.get(
  "/session/minecraft/profile/:uuid",
  rateLimit(rateLimitPresets.account),
  validateParam(profileParamSchema),
  validateQuery(profileQuerySchema),
  async (c) => {
    const param = c.req.valid("param");
    const query = c.req.valid("query");
    void query;
    let env;
    try {
      env = getEnv(c);
    } catch {
      return yggError(c, 500, "InternalError", "Servicio no disponible.");
    }
    const profile = await fetchProfileByUuid(env, param.uuid);
    if (!profile) return c.body(null, 204);
    const signed = await signedTexturesFor(env, {
      userId: profile.userId,
      profileUuid: profile.uuid,
      profileName: profile.name,
      origin: originOf(c),
    });
    return c.json(serializeProfileFull(profile.uuid, profile.name, signed?.value, signed?.signature));
  },
);

// GET /sessionserver/blockedservers — lista de hashes bloqueados por Mojang.
// El juego la pide al abrir el multijugador; vacía = nada bloqueado.
sessionserver.get(
  "/blockedservers",
  rateLimit(rateLimitPresets.public),
  (c) => c.json([]),
);
