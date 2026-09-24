// GET / — metadatos authlib-injector + indicación de ubicación (ALI).
// Respuesta: `{ meta, skinDomains, signaturePublickey }`.
// Además expone `X-Authlib-Injector-API-Location` para descubrimiento.

import type { Context } from "hono";
import type { YggEnv } from "../types/app";
import { getEnv } from "../env";
import { derivePublicPem } from "../lib/crypto";

export async function metaHandler(c: Context<YggEnv>) {
  let env;
  try {
    env = getEnv(c);
  } catch {
    return c.json({ error: "InternalError", errorMessage: "Servicio no disponible." }, 500);
  }

  const signaturePublickey = await derivePublicPem(env.signPrivateKeyPem);
  // ALI: apunta a la raíz del propio servicio (descubrimiento por launchers).
  const here = new URL(c.req.url);
  const apiRoot = `${here.protocol}//${here.host}/`;
  c.header("X-Authlib-Injector-API-Location", apiRoot);

  // Las texturas se sirven desde este mismo Worker: su host siempre vale en
  // la whitelist (más lo configurado en SKIN_DOMAINS), si no el juego las
  // rechaza con "non-whitelisted domain".
  const skinDomains = [...new Set([...env.skinDomains, here.hostname])];

  return c.json({
    meta: {
      serverName: env.serverName,
      implementationName: "steplauncher-yggdrasil",
      implementationVersion: "0.1.0",
      links: { homepage: env.homepageUrl, register: env.registerUrl },
      "feature.non_email_login": true,
    },
    skinDomains,
    signaturePublickey,
  });
}
