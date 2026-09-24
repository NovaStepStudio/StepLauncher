// Entrada del Worker Yggdrasil: tubería global + routers del protocolo.
// Orden: request-id → security-headers → CORS → rutas → 404 → errores.
// Formatos Yggdrasil en todo (nada del sobre `ok()`/`fail()` de `api/`).

import { Hono } from "hono";
import type { YggEnv } from "./types/app";
import { requestId } from "./middleware/request-id";
import { securityHeaders } from "./middleware/security-headers";
import { openCors } from "./middleware/cors";
import { authserver } from "./routes/authserver";
import { sessionserver } from "./routes/sessionserver";
import { profilesApi } from "./routes/profiles";
import { texturesApi } from "./routes/textures";
import { cloaksApi, legacyApi } from "./routes/legacy";
import { metaHandler } from "./routes/meta";
import { getEnv } from "./env";
import { derivePublicPem } from "./lib/crypto";

const app = new Hono<YggEnv>();

app.use("*", requestId());
app.use("*", securityHeaders());
app.use("*", openCors());

// Metadatos authlib-injector en la raíz (descubrimiento + ALI).
app.get("/", metaHandler);

// El pre-verify del launcher hace GET a la URL configurada esperando 200:
// si configuran la URL con sufijo `/authserver`, también responde metadatos.
app.get("/authserver", metaHandler);

// Dominios del protocolo con prefijos Mojang.
app.route("/authserver", authserver);
app.route("/sessionserver", sessionserver);
app.route("/api", profilesApi);

// Texturas del juego (URLs estables por hash, con caché inmutable).
app.route("/textures", texturesApi);

// API legacy de skins para juegos viejos (`skins.minecraft.net/...`).
app.route("/skins", legacyApi);

// Capas por nombre, estilo Ely.by (para launchers, plugins y depuración).
app.route("/cloaks", cloaksApi);

// Clave pública para verificar firmas (formato PEM, como Ely.by).
// Útil para plugins de servidor y para depurar (`diff` contra `GET /`).
app.get("/signature-verification-key.pem", async (c) => {
  let env;
  try {
    env = getEnv(c);
  } catch {
    return c.json({ error: "InternalError", errorMessage: "Servicio no disponible." }, 500);
  }
  const pub = await derivePublicPem(env.signPrivateKeyPem);
  if (!pub) {
    return c.json({ error: "InternalError", errorMessage: "Sin clave configurada." }, 500);
  }
  c.header("Content-Type", "application/x-pem-file");
  return c.body(pub, 200);
});

// 404 con forma Yggdrasil (el launcher lo distingue por `error`).
app.notFound((c) => {
  return c.json({ error: "Not Found", errorMessage: "Recurso no encontrado." }, 404);
});

// Errores no controlados: genéricos, sin stack ni secretos al cliente.
app.onError((err, c) => {
  const id = (c.get("requestId") as string | undefined) ?? "sin-id";
  console.error(`[${id}] error no controlado:`, err?.message ?? err);
  if ((err as Error)?.message === "server_misconfigured") {
    return c.json({ error: "InternalError", errorMessage: "Servicio no disponible." }, 500);
  }
  return c.json({ error: "InternalError", errorMessage: "Error interno." }, 500);
});

export default app;
