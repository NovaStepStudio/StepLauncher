# Estructura del código (dónde va cada cosa)

Mapa de `api/src/` para saber dónde poner cada archivo sin preguntar.
Isla del monorepo: un cambio aquí **solo** toca `api/` (ver `api/AGENTS.md`).

```text
api/
  src/
    index.ts              # Entrada: middlewares globales + GET / + route("/v1") + 404 + onError
    env.ts                # getEnv(): única vía a variables (lanza "server_misconfigured")
    lib/
      respond.ts          # ok() / fail(): el único sobre de respuesta JSON
      supabase.ts         # supabaseForUser() (RLS) / supabaseAdmin() (bypass, restringido)
      images.ts           # imageInfo(): formato + dimensiones reales por cabecera
    middleware/
      request-id.ts       # X-Request-Id por petición (trazabilidad)
      security-headers.ts # Cabeceras duras (CSP, HSTS solo en producción…)
      cors.ts             # CORS por lista blanca ALLOWED_ORIGINS (sin wildcard)
      auth.ts             # requireAuth() + getBearerToken()
      rate-limit.ts       # Ventana fija por IP: presets public/account/sensitive
      validate.ts         # validateJson / validateQuery / validateParam (Zod → 400)
      upload-quota.ts     # Quota 10 skins+capas/día contando file_uploads 24h
    routes/v1/            # Un dominio = un archivo (health, auth, accounts, files,
                          # notifications, friends, users, community) + index.ts que los cuelga
      index.ts            # SOLO route(): sin lógica
    schemas/v1/           # Un archivo por dominio; jamás se reutilizan en v2
    types/app.ts          # AppBindings, AuthUser, AppVariables, AppEnv (solo tipos)
  docs/                   # Esta documentación (obligatoria)
  db/v1/                  # SQL versionado: partes + install.sql + verify.sql
  scripts/build-db-bundle.ts  # Genera db/v1/install.sql (bun run db:build)
```

## Tubería de una petición (orden real en `src/index.ts`)

1. `requestId()` — genera UUID, lo guarda en contexto y lo devuelve en `X-Request-Id`.
2. `securityHeaders()` — `nosniff`, `DENY`, `no-referrer`, `Permissions-Policy`,
   `CSP default-src 'none'` (+ `HSTS` solo si `API_ENV=production`).
3. `strictCors()` — refleja el origen solo si está en `ALLOWED_ORIGINS`; si no,
   sin cabecera `Access-Control-Allow-Origin`. Métodos `GET,POST,PUT,PATCH,DELETE,
   OPTIONS`; cabeceras `Content-Type, Authorization, X-Request-Id`; expone
   `X-Request-Id, Retry-After`; `maxAge 600`; sin credenciales.
4. Rutas `/v1/*` (cada endpoint: `rateLimit` → `requireAuth` si es cuenta →
   `validate*` → handler con `ok`/`fail`).
5. `GET /` — descubrimiento: `{ service, versions: ["v1"], docs, health }`.
6. `notFound` → `not_found` 404 con el mismo sobre. `onError` → `internal_error`
   500 genérico (`server_misconfigured` si falta entorno).

## Añadir un endpoint (chuleta, ver `api/AGENTS.md` §7)

1. Esquema Zod en `src/schemas/v1/<dominio>.ts` (si hay entrada). `.strict()`
   donde aplique; límites pegados al dominio (ej: bio ≤ 5000).
2. Ruta en `src/routes/v1/<dominio>.ts`: `rateLimit(preset)` → `requireAuth()`
   (si toca datos del usuario) → `validate*` → handler con `ok`/`fail`.
   El dueño sale de `c.get("user").id`; prohibido `userId` del cliente.
3. Si es dominio nuevo, una línea `route()` en `src/routes/v1/index.ts`.
4. Docs en `docs/api/v1/<dominio>.md` (+ `docs/cuentas/` u `docs/oauth/` si aplica).
5. Si toca SQL, evoluciona `db/v1/` (compatible) o crea `db/v2/` (incompatible).
6. `bun run typecheck` + revisión de secretos antes del PR.

## Reglas que no se negocian

- Respuestas solo con `ok()` / `fail()` (ver `docs/api/v1/README.md`).
- Variables solo vía `getEnv()`; prohibido `c.env` directo en rutas.
- `supabaseAdmin()` solo si es imprescindible, documentando el porqué, y jamás
  devolviendo su resultado crudo con campos sensibles.
- Errores genéricos con `requestId`; 401 sin distinguir expirado de inválido.
- Gestor **siempre Bun**; **prohibido npm**. Todo se ejecuta dentro de `api/`.
