// Formato único de respuestas JSON para toda la API.
// Éxito: { "success": true, "data": ... }
// Error: { "success": false, "error": { "code": "...", "message": "...", "requestId": "..." } }

import type { Context } from "hono";
import type { ContentfulStatusCode } from "hono/utils/http-status";
import type { AppEnv } from "../types/app";

export function ok<T>(c: Context<AppEnv>, data: T, status: ContentfulStatusCode = 200) {
  return c.json({ success: true, data }, status);
}

interface FailOptions {
  /** Código estable en snake_case para que el launcher lo gestione (ej: "unauthorized"). */
  code: string;
  /** Mensaje genérico para el usuario. Nunca incluir secretos ni SQL ni stack. */
  message: string;
  status?: ContentfulStatusCode;
  /** Detalles de validación por campo. Nunca datos sensibles. */
  details?: unknown;
}

export function fail(c: Context<AppEnv>, opts: FailOptions) {
  const status: ContentfulStatusCode = opts.status ?? 500;
  const requestId = c.get("requestId") as string | undefined;
  return c.json(
    {
      success: false,
      error: {
        code: opts.code,
        message: opts.message,
        ...(opts.details !== undefined ? { details: opts.details } : {}),
        ...(requestId ? { requestId } : {}),
      },
    },
    status,
  );
}
