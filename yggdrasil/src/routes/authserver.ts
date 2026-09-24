// Dominio authserver: login, refresco, validación y revocación de tokens.
// Formatos exactos del spec Yggdrasil (nada de sobre de `api/`).
// Los usuarios se verifican en Supabase (misma base que `api/`).

import { Hono } from "hono";
import type { YggEnv } from "../types/app";
import { getEnv } from "../env";
import { newClientToken, stripUuid } from "../lib/crypto";
import {
  invalidCredentials,
  invalidToken,
  serializeProfile,
  serializeUser,
  yggError,
} from "../lib/ygg";
import {
  fetchMcProfile,
  fetchProfileByName,
  findLiveToken,
  issueToken,
  resolveEmail,
  revokeAllForUser,
  revokeToken,
  rotateToken,
  verifyCredentials,
} from "../lib/store";
import { rateLimit, rateLimitPresets } from "../middleware/rate-limit";
import { validateJson } from "../middleware/validate";
import {
  authenticateSchema,
  refreshSchema,
  signoutSchema,
  tokenPairSchema,
} from "../schemas/ygg";

export const authserver = new Hono<YggEnv>();

// POST /authserver/authenticate — email o usuario + contraseña → tokens + perfil.
authserver.post(
  "/authenticate",
  rateLimit(rateLimitPresets.sensitive),
  validateJson(authenticateSchema),
  async (c) => {
    const input = c.req.valid("json");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return yggError(c, 500, "InternalError", "Servicio no disponible.");
    }

    // El identificador puede ser email o nombre de jugador (non_email_login).
    const email = await resolveEmail(env, input.username);
    if (!email) return invalidCredentials(c);

    const login = await verifyCredentials(env, email, input.password);
    if (!login) return invalidCredentials(c);

    // Perfil Minecraft del usuario (username + mc_uuid de `profiles`).
    let profile = await fetchMcProfile(env, login.userId);
    // Si entró con nombre de perfil distinto al username actual, revalidar.
    if (!profile && !email.includes("@")) {
      void email;
      profile = await fetchProfileByName(env, input.username);
    }
    if (!profile) {
      return yggError(c, 403, "ForbiddenOperationException", "La cuenta no tiene perfil de juego.");
    }

    const clientToken =
      typeof input.clientToken === "string" && input.clientToken.length > 0
        ? input.clientToken
        : newClientToken();
    const token = await issueToken(env, login.userId, clientToken, profile);

    return c.json({
      accessToken: token.accessToken,
      clientToken: token.clientToken,
      availableProfiles: [serializeProfile(profile.uuid, profile.name)],
      selectedProfile: serializeProfile(profile.uuid, profile.name),
      ...(input.requestUser ? { user: serializeUser(login.userId) } : {}),
    });
  },
);

// POST /authserver/refresh — revoca el viejo y emite uno nuevo.
authserver.post(
  "/refresh",
  rateLimit(rateLimitPresets.account),
  validateJson(refreshSchema),
  async (c) => {
    const input = c.req.valid("json");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return yggError(c, 500, "InternalError", "Servicio no disponible.");
    }

    const old = await findLiveToken(env, input.accessToken, input.clientToken);
    if (!old) {
      return yggError(c, 403, "ForbiddenOperationException", "Invalid token.", "Token does not exist");
    }

    // Selección de perfil: el launcher StepLauncher SIEMPRE manda `selectedProfile`
    // en el refresh, así que si el token ya tiene perfil atado y piden el MISMO,
    // se acepta como refresco normal. Solo es error si intentan atar uno DISTINTO.
    if (input.selectedProfile && old.profileUuid) {
      if (stripUuid(old.profileUuid) !== input.selectedProfile.id) {
        return yggError(c, 400, "IllegalArgumentException", "Access token already has a profile assigned.");
      }
    } else if (input.selectedProfile) {
      const current = await fetchMcProfile(env, old.userId);
      if (
        !current ||
        stripUuid(current.uuid) !== input.selectedProfile.id ||
        current.name !== input.selectedProfile.name
      ) {
        return yggError(c, 403, "ForbiddenOperationException", "Invalid token.");
      }
      const rotated = await rotateToken(env, old, current);
      return c.json({
        accessToken: rotated.accessToken,
        clientToken: rotated.clientToken,
        selectedProfile: serializeProfile(current.uuid, current.name),
        ...(input.requestUser ? { user: serializeUser(old.userId) } : {}),
      });
    }

    // Refresco normal: conserva el perfil atado (puede ser null si la cuenta no tenía).
    // El nombre se refresca desde la base por si cambió en la web/`api/`.
    const fresh = await fetchMcProfile(env, old.userId);
    const keep =
      old.profileUuid && old.profileName
        ? { uuid: old.profileUuid, name: fresh?.name ?? old.profileName, userId: old.userId }
        : fresh;
    const rotated = await rotateToken(env, old, keep);
    const body: Record<string, unknown> = {
      accessToken: rotated.accessToken,
      clientToken: rotated.clientToken,
    };
    if (keep) body["selectedProfile"] = serializeProfile(keep.uuid, keep.name);
    if (input.requestUser) body["user"] = serializeUser(old.userId);
    return c.json(body);
  },
);

// POST /authserver/validate — 204 si vale, 403 si no.
authserver.post(
  "/validate",
  rateLimit(rateLimitPresets.account),
  validateJson(tokenPairSchema),
  async (c) => {
    const input = c.req.valid("json");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return yggError(c, 500, "InternalError", "Servicio no disponible.");
    }
    const live = await findLiveToken(env, input.accessToken, input.clientToken);
    if (!live) return invalidToken(c);
    return c.body(null, 204);
  },
);

// POST /authserver/invalidate — revoca el token (204 siempre).
authserver.post(
  "/invalidate",
  rateLimit(rateLimitPresets.account),
  validateJson(tokenPairSchema),
  async (c) => {
    const input = c.req.valid("json");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return yggError(c, 500, "InternalError", "Servicio no disponible.");
    }
    await revokeToken(env, input.accessToken);
    return c.body(null, 204);
  },
);

// POST /authserver/signout — revoca TODOS los tokens del usuario (204).
authserver.post(
  "/signout",
  rateLimit(rateLimitPresets.sensitive),
  validateJson(signoutSchema),
  async (c) => {
    const input = c.req.valid("json");
    let env;
    try {
      env = getEnv(c);
    } catch {
      return yggError(c, 500, "InternalError", "Servicio no disponible.");
    }
    const email = await resolveEmail(env, input.username);
    if (!email) return invalidCredentials(c);
    const login = await verifyCredentials(env, email, input.password);
    if (!login) return invalidCredentials(c);
    await revokeAllForUser(env, login.userId);
    return c.body(null, 204);
  },
);
