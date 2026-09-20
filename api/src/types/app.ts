// Tipos compartidos de la API. Sin secretos ni lógica, solo contratos.

/** Variables de entorno (Bindings) que expone el Worker. */
export interface AppBindings {
  /** URL pública del proyecto Supabase. Ej: https://xyzcompany.supabase.co */
  SUPABASE_URL: string;
  /** Clave anónima de Supabase (pública, limitada por RLS). */
  SUPABASE_ANON_KEY: string;
  /**
   * Clave de servicio (ULTRA SECRETA). Solo se usa en servidor y nunca
   * se envía al cliente ni se registra en logs.
   */
  SUPABASE_SERVICE_ROLE_KEY: string;
  /** Entorno lógico: local | preview | production. */
  API_ENV?: string;
  /** Orígenes CORS permitidos, separados por coma. Vacío = denegar todo. */
  ALLOWED_ORIGINS?: string;
}

/** Usuario autenticado mínimo que guardamos en el contexto. */
export interface AuthUser {
  /** UUID de auth.users en Supabase. Único dueño del recurso. */
  id: string;
  email?: string;
}

/** Variables por petición que viajan en el contexto de Hono. */
export interface AppVariables {
  /** Identificador único por petición para trazabilidad en logs y errores. */
  requestId: string;
  /** Presente solo cuando `auth()` ha validado el Bearer. */
  user?: AuthUser;
}

/** Entorno Hono de toda la API. */
export interface AppEnv {
  Bindings: AppBindings;
  Variables: AppVariables;
}
