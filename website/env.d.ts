/// <reference types="vite/client" />

// Variables propias de la web (ver `.env.example`).
// La base de la API se lee en `web/src/Auth/Api.ts` con fallback local.
interface ImportMetaEnv {
    readonly VITE_API_BASE_URL?: string;
}
