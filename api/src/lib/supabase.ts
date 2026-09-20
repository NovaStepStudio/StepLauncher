// Factoría de clientes Supabase. Sin singletons con secretos fuera de petición.
// Cada cliente se crea por petición a partir de `c.env` (edge-safe).

import { createClient, type SupabaseClient } from "@supabase/supabase-js";
import type { SafeEnv } from "../env";

/**
 * Cliente con la identidad del usuario (RLS activo).
 * El token Bearer del usuario viaja como Authorization y Supabase aplica sus políticas.
 */
export function supabaseForUser(env: SafeEnv, accessToken: string): SupabaseClient {
  return createClient(env.supabaseUrl, env.supabaseAnonKey, {
    global: { headers: { Authorization: `Bearer ${accessToken}` } },
    auth: { persistSession: false, autoRefreshToken: false },
  });
}

/**
 * Cliente administrativo (bypass RLS). USO RESTRINGIDO:
 * - Solo en servidor, solo cuando sea imprescindible (p. ej. alta de cuenta).
 * - Nunca loguear la service key ni devolverla al cliente.
 */
export function supabaseAdmin(env: SafeEnv): SupabaseClient {
  return createClient(env.supabaseUrl, env.supabaseServiceRoleKey, {
    auth: { persistSession: false, autoRefreshToken: false },
  });
}
