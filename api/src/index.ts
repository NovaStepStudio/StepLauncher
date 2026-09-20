// Entrada del Worker: tubería global + router versionado.
// Orden: request-id → security-headers → CORS → rutas → 404 → errores.
// Ningún endpoint fuera de `/vN` salvo `/` (índice) y `/v1/health` (sonda).

import { Hono } from "hono";
import type { AppEnv } from "./types/app";
import { requestId } from "./middleware/request-id";
import { securityHeaders } from "./middleware/security-headers";
import { strictCors } from "./middleware/cors";
import { ok, fail } from "./lib/respond";
import { v1 } from "./routes/v1";

const app = new Hono<AppEnv>();

app.use("*", requestId());
app.use("*", securityHeaders());
app.use("*", strictCors());

// Índice raíz: descubrimiento, sin datos sensibles.
app.get("/", (c) => {
  return ok(c, {
    service: "steplauncher-api",
    versions: ["v1"],
    docs: "ver carpeta docs/ del repo",
    health: "/v1/health",
  });
});

app.route("/v1", v1);

// 404 con el mismo sobre que el resto (el launcher lo gestiona por `code`).
app.notFound((c) => {
  return fail(c, { code: "not_found", message: "Recurso no encontrado.", status: 404 });
});

// Errores no controlados: mensaje genérico, sin stack ni secretos al cliente.
app.onError((err, c) => {
  // Log mínimo correlacionado por requestId (en Workers va a `wrangler tail`).
  console.error(`[${c.get("requestId") ?? "sin-id"}] error no controlado:`, err?.message ?? err);
  if ((err as Error)?.message === "server_misconfigured") {
    return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
  }
  return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
});

export default app;
