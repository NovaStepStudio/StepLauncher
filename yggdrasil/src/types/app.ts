// Tipos compartidos de yggdrasil/. Sin secretos ni lógica, solo contratos.
// Servicio independiente de `api/`: no se importa nada de fuera de esta carpeta.

/** Variables de entorno (Bindings) que expone el Worker Yggdrasil. */
export interface YggBindings {
  /** URL pública del proyecto Supabase (la misma base que usa `api/`). */
  SUPABASE_URL: string;
  /** Clave anónima de Supabase (pública, limitada por RLS). */
  SUPABASE_ANON_KEY: string;
  /** Clave de servicio (ULTRA SECRETA). Solo en servidor, nunca al cliente. */
  SUPABASE_SERVICE_ROLE_KEY: string;
  /** Clave privada RSA en PEM (PKCS8) para firmar `textures` (SHA1withRSA). */
  YGG_SIGN_PRIVATE_KEY: string;
  /** Entorno lógico: local | preview | production. */
  API_ENV?: string;
  /** Nombre visible del servidor en GET / (meta.serverName). */
  SERVER_NAME?: string;
  /** Dominios de texturas separados por coma. */
  SKIN_DOMAINS?: string;
  /** Enlaces visibles en meta.links. */
  HOMEPAGE_URL?: string;
  REGISTER_URL?: string;
}

/** Variables por petición que viajan en el contexto de Hono. */
export interface YggVariables {
  /** Identificador único por petición para trazabilidad en logs. */
  requestId: string;
}

/** Entorno Hono de todo el servicio Yggdrasil. */
export interface YggEnv {
  Bindings: YggBindings;
  Variables: YggVariables;
}

/** Perfil Minecraft mínimo leído desde `profiles` (fuente: Supabase). */
export interface McProfile {
  /** UUID con guiones (como vive en `profiles.mc_uuid`). */
  uuid: string;
  /** Nombre de jugador (`profiles.username`). */
  name: string;
  /** Dueño (`profiles.user_id`): clave para buscar sus skins/capas. */
  userId: string;
}

/** Token Yggdrasil guardado en `ygg_tokens`. */
export interface YggToken {
  accessToken: string;
  clientToken: string;
  userId: string;
  profileUuid: string | null;
  profileName: string | null;
  issuedAt: string;
  expiresAt: string;
  revoked: boolean;
}
