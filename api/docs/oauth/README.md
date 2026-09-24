# Sesiones y Auth (OAuth y modelo actual)

Todo lo de autenticación de StepLauncher: cómo entra el jugador, qué guarda el
launcher y hacia dónde evoluciona. El contrato HTTP exacto vive en
`docs/api/v1/auth.md`; aquí el modelo y el porqué.

## Modelo actual v1 (implementado, ver `src/routes/v1/auth.ts`)

Las **sesiones las emite Supabase Auth** (`access_token` + `refresh_token`); la API
**no inventa criptografía** (ADR-007, ADR-015). Flujo con confirmación (web y launcher):

1. `POST /v1/auth/register` (email + usuario + contraseña) crea la cuenta vía
   `signUp` y envía la plantilla **Confirm sign up** a `<SITE_URL>/auth/callback`.
   Responde `201` pendiente (`session: null`, `emailConfirmationRequired: true`).
2. El jugador confirma en `POST /v1/auth/confirm` (`type: signup`, con el código o
   `token_hash` del enlace) o desde el propio enlace; ahí recibe
   `{ user, session }`. Reenvío en `POST /v1/auth/resend` (siempre genérico).
3. `POST /v1/auth/login` (email **o** usuario + contraseña). Sin confirmar →
   `403 email_not_confirmed` (la web ofrece reenviar); si no, `{ user, session }`.
4. El cliente guarda ambos tokens y envía `Authorization: Bearer <accessToken>`.
5. Antes de caducar, `POST /v1/auth/refresh` renueva la sesión.
6. Recupero: `POST /v1/auth/recover` (plantilla **Reset password**, genérico) →
   `POST /v1/auth/reset-password` (`email` + `token` + `newPassword`) → login.
   Supabase avisa además con **Password changed**.
7. Cambio de email: `POST /v1/accounts/me/email-change` (plantilla
   **Change email address** al NUEVO, aviso **Email address changed** al viejo) →
   confirmar en `POST /v1/auth/confirm` (`type: email_change`).
8. Al cerrar sesión, `POST /v1/auth/logout` revoca en servidor **y el cliente borra
   sus tokens** (si solo borra sin llamar, el access sigue válido hasta caducar).

Detalles honestos:

- `register` ya no usa `email_confirm: true` (fin del MVP sin fricción, ver ADR-015).
  Si el proyecto aún tiene la confirmación desactivada, puede llegar sesión
  inmediata (compatibilidad transitoria).
- Login con usuario resuelve username→email vía función `email_for_username`
  (`SECURITY DEFINER`, revocada). `invalid_credentials` sigue genérico, pero
  `email_not_confirmed` es propio para guiar a la bandeja.
- `resend`/`recover` responden éxito genérico siempre para no enumerar correos;
  el error real queda en logs con `requestId`.

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
