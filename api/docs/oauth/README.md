# Sesiones y Auth (OAuth y modelo actual)

Todo lo de autenticación de StepLauncher: cómo entra el jugador, qué guarda el
launcher y hacia dónde evoluciona. El contrato HTTP exacto vive en
`docs/api/v1/auth.md`; aquí el modelo y el porqué.

## Modelo actual v1 (implementado, ver `src/routes/v1/auth.ts`)

Las **sesiones las emite Supabase Auth** (`access_token` + `refresh_token`); la API
**no inventa criptografía** (ADR-007). Flujo del launcher (desktop):

1. `POST /v1/auth/register` (email + usuario + contraseña) o
   `POST /v1/auth/login` (email **o** usuario + contraseña).
2. La API responde `{ user: { id, email, username }, session: { accessToken,
   refreshToken, expiresAt, tokenType: "bearer" } }` (registro: 201).
3. El launcher guarda ambos tokens en el dispositivo y envía
   `Authorization: Bearer <accessToken>` en cada petición de cuenta.
4. Antes de caducar, `POST /v1/auth/refresh` con el `refreshToken` renueva la sesión.
5. Al cerrar sesión, `POST /v1/auth/logout` revoca en servidor **y el cliente borra
   sus tokens** (si solo borra sin llamar, el access sigue válido hasta caducar).

Detalles honestos del MVP:

- `email_confirm: true` al registrar: sin fricción de correo (revisar en ADR-007
  cuando se quiera verificación real).
- Login con usuario resuelve username→email vía función `email_for_username`
  (`SECURITY DEFINER`, revocada): el 401 es siempre genérico
  (`invalid_credentials`, sin revelar si existe el usuario).
- Tras el registro, la API inicia sesión inmediata; si eso falla, devuelve 500 con
  `Cuenta creada, pero inicia sesión manualmente` (la cuenta SÍ quedó creada).

## Tokens y PKCE (qué hay y qué no)

- **Hoy**: Bearer opaco de Supabase validado con `auth.getUser()` en cada petición
  (`requireAuth()`). Sin refresh rotation visible al cliente más allá del
  `refreshSession` de Supabase; sin scopes: un token da acceso a **su** cuenta y
  solo a la suya (RLS `auth.uid()`).
- **PKCE con navegador (futuro, NO implementado)**: si algún día el registro pasa
  por web con `code_challenge`, esta carpeta documentará el flujo (authorize →
  callback con `code` → intercambio → `access`+`refresh`), la custodia del
  `code_verifier` y la sesión del launcher desktop. Hoy no existe ningún endpoint
  OAuth: no inventarlo ni exponerlo.

## Reglas para el launcher

- Guardar tokens en almacenamiento seguro del SO (nunca en logs ni en capturas).
- Renovar proactivamente antes de `expiresAt`; ante 401, reintentar una vez con
  refresh y si falla, pedir login (no bucles infinitos).
- `refreshToken` comprometido = llamar a `logout` y pedir login (rota la sesión).
