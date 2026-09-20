// Lectura y validación de variables de entorno.
// Toda variable sale de `c.env` (wrangler / .dev.vars / secrets).
// Prohibido hardcodear URLs o claves en código.

import type { Context } from "hono";
import type { AppBindings, AppEnv } from "./types/app";

export interface SafeEnv {
  supabaseUrl: string;
  supabaseAnonKey: string;
  supabaseServiceRoleKey: string;
  apiEnv: string;
  allowedOrigins: string[];
}

/** Falla con mensaje genérico (sin filtrar qué variable falta). */
export function getEnv(c: Context<AppEnv>): SafeEnv {
  const raw = c.env as AppBindings;

  const supabaseUrl = raw.SUPABASE_URL?.trim();
  const supabaseAnonKey = raw.SUPABASE_ANON_KEY?.trim();
  const supabaseServiceRoleKey = raw.SUPABASE_SERVICE_ROLE_KEY?.trim();

  // No detallar en la respuesta qué falta: sería enumeración para un atacante.
  if (!supabaseUrl || !supabaseAnonKey || !supabaseServiceRoleKey) {
    throw new Error("server_misconfigured");
  }

  return {
    supabaseUrl,
    supabaseAnonKey,
    supabaseServiceRoleKey,
    apiEnv: (raw.API_ENV ?? "production").trim() || "production",
    allowedOrigins: parseAllowedOrigins(raw.ALLOWED_ORIGINS),
  };
}

/** Convierte "https://a, https://b" en lista limpia y sin duplicados. */
export function parseAllowedOrigins(value: string | undefined): string[] {
  if (!value) return [];
  const list = value
    .split(",")
    .map((o) => o.trim().replace(/\/+$/, ""))
    .filter((o) => o.length > 0 && (o.startsWith("https://") || o.startsWith("http://")));
  return [...new Set(list)];
}
