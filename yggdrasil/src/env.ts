// Lectura y validación de variables de entorno del Worker Yggdrasil.
// Toda variable sale de `c.env` (wrangler / .dev.vars / secrets).
// Prohibido hardcodear URLs o claves en código.

import type { Context } from "hono";
import type { YggBindings, YggEnv } from "./types/app";

export interface YggEnvSafe {
  supabaseUrl: string;
  supabaseAnonKey: string;
  supabaseServiceRoleKey: string;
  signPrivateKeyPem: string;
  apiEnv: string;
  serverName: string;
  skinDomains: string[];
  homepageUrl: string;
  registerUrl: string;
}

/** Falla con mensaje genérico (sin filtrar qué variable falta). */
export function getEnv(c: Context<YggEnv>): YggEnvSafe {
  const raw = c.env as YggBindings;

  const supabaseUrl = raw.SUPABASE_URL?.trim();
  const supabaseAnonKey = raw.SUPABASE_ANON_KEY?.trim();
  const supabaseServiceRoleKey = raw.SUPABASE_SERVICE_ROLE_KEY?.trim();
  const signPrivateKeyPem = raw.YGG_SIGN_PRIVATE_KEY?.trim();

  // Sin estos cuatro no hay login ni firmas: servicio no disponible.
  if (!supabaseUrl || !supabaseAnonKey || !supabaseServiceRoleKey || !signPrivateKeyPem) {
    throw new Error("server_misconfigured");
  }

  return {
    supabaseUrl,
    supabaseAnonKey,
    supabaseServiceRoleKey,
    signPrivateKeyPem,
    apiEnv: (raw.API_ENV ?? "production").trim() || "production",
    serverName: (raw.SERVER_NAME ?? "StepLauncher Yggdrasil").trim() || "StepLauncher Yggdrasil",
    skinDomains: parseList(raw.SKIN_DOMAINS),
    homepageUrl: (raw.HOMEPAGE_URL ?? "https://steplauncher.pages.dev").trim(),
    registerUrl: (raw.REGISTER_URL ?? "https://steplauncher.pages.dev").trim(),
  };
}

/** Convierte "a, b, c" en lista limpia sin duplicados. */
export function parseList(value: string | undefined): string[] {
  if (!value) return [];
  const list = value
    .split(",")
    .map((s) => s.trim())
    .filter((s) => s.length > 0);
  return [...new Set(list)];
}
