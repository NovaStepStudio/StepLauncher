# Auth v1 — registro con confirmación, login, sesiones y recupero

Sesiones **Supabase** (`access_token` + `refresh_token`); la API no inventa
criptografía (ADR-007, ADR-015). Modelo completo: `docs/oauth/README.md`.
Rate-limit `sensitive` en register/login/refresh/resend/recover/confirm/reset-password
(20/10 min), `account` en logout.

Config Supabase requerida (panel → Authentication): activar `Confirm email`,
`Secure email change` y las plantillas **Confirm sign up**, **Change email address**,
**Reset password**, **Password changed** y **Email address changed**.
`Site URL` = base de la web y `Redirect URLs` debe incluir
`<SITE_URL>/auth/callback` (+ `http://localhost:5173/auth/callback` en local).
La API envía `emailRedirectTo`/`redirectTo` a ese callback con `SITE_URL`
(`vars` en `wrangler.jsonc`, `.dev.vars` en local).

## `POST /v1/auth/register` — alta con email + usuario + contraseña (CON confirmación)

Cuerpo (`application/json`, estricto):

```json
{ "email": "jugador@ejemplo.com", "username": "Steve_01", "password": "secreta123" }
```

Reglas: `email` válido ≤254 (se normaliza a minúsculas); `username` 3–20 con
`A-Za-z0-9._-` (único insensible a mayúsculas); `password` 8–72 caracteres.

```bash
curl -X POST http://localhost:8787/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"jugador@ejemplo.com","username":"Steve_01","password":"secreta123"}'
```

`201` pendiente (caso normal con `Confirm email` activo):

```json
{
  "success": true,
  "data": {
    "user": { "id": "uuid", "email": "jugador@ejemplo.com", "username": "Steve_01" },
    "session": null,
    "emailConfirmationRequired": true,
    "message": "Revisa tu correo para confirmar tu cuenta."
  }
}
```

El perfil se crea por trigger (`handle_new_user`) con ese username, pero **sin
sesión**: hay que confirmar en `POST /v1/auth/confirm` (type `signup`) o desde
el enlace que llega al correo (plantilla **Confirm sign up** → `<SITE_URL>/auth/callback`).
Si el proyecto aún tiene la confirmación desactivada, devuelve `{ user, session }`
como antes (compatibilidad transitoria).

Errores: `validation_error` 400 · `username_taken` 409 · `email_taken` 409 ·
`weak_password` 400.

## `POST /v1/auth/resend` — reenviar confirmación (`sensitive`)

```json
{ "email": "jugador@ejemplo.com", "type": "signup" }
```

`type`: `signup` (defecto) o `email_change`. Respuesta siempre genérica para no
enumerar correos: `200` → `{ resent: true, message: "Si ese correo está pendiente…" }`
(aunque no exista o ya esté confirmado). El enlace usa la plantilla
**Confirm sign up** o **Change email address** según `type`.

## `POST /v1/auth/confirm` — confirmar con el código del correo (`sensitive`)

```json
{ "email": "jugador@ejemplo.com", "token": "123456", "type": "signup" }
```

`type`: `signup` | `email_change` | `recovery`. `token` es el código/OTP o el
`token_hash` del enlace (la web lo saca de `/auth/callback`). Acepta ambos.
`signup`/`email_change` → `200` con `{ user: { id, email, username }, session }`
(sesión lista para guardar); `recovery` → `200` con `{ recoveryVerified: true }`
(el cambio real va en `/reset-password`). Error de código → `400 invalid_code`
(`Código inválido o expirado. Pide un nuevo enlace.`).

## `POST /v1/auth/recover` — pedir "Reset password" (`sensitive`)

```json
{ "email": "jugador@ejemplo.com" }
```

Envía la plantilla **Reset password** con vuelta a `<SITE_URL>/auth/callback`.
Respuesta siempre genérica (sin enumerar): `200` →
`{ recoverySent: true, message: "Si ese correo está registrado…" }`.

## `POST /v1/auth/reset-password` — fijar contraseña nueva (`sensitive`)

```json
{ "email": "jugador@ejemplo.com", "token": "123456", "newPassword": "nueva-secreta123" }
```

Verifica el OTP de `recovery` y actualiza la contraseña (el código prueba el
correo, por eso usa `admin` solo aquí). `200` → `{ passwordReset: true }`
(después hay que hacer login). Supabase envía además el aviso **Password changed**.
Errores: `invalid_code` 400 · `weak_password` 400.

## `POST /v1/auth/recovery-password` — fijar contraseña con la sesión del enlace (`sensitive`, auth)

Es el camino normal del enlace: Supabase llega a `<SITE_URL>/auth/callback` con
`#access_token=…&type=recovery` (la sesión ya prueba el correo, nada que pegar).
Se llama con `Authorization: Bearer <access_token del enlace>`:

```json
{ "newPassword": "nueva-secreta123" }
```

Actualiza con la identidad del token (sin `admin`, sin email del cliente).
`200` → `{ passwordReset: true }` (después hay que hacer login, sin guardar la
sesión de recupero). Errores: `unauthorized` 401 · `weak_password` 400 ·
`recovery_expired` 400 (enlace vencido o ya usado: vale 1 h y un solo uso).

## `POST /v1/auth/login` — email O usuario + contraseña

```json
{ "identifier": "Steve_01", "password": "secreta123" }
```

`identifier` 3–254 (con `@` se trata como email; si no, se resuelve username→email
en servidor). `200` devuelve `{ user: { id, email, username }, session }`
(`username` puede ser `null` si el trigger aún no creó el perfil).

```bash
curl -X POST http://localhost:8787/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"Steve_01","password":"secreta123"}'
```

Error de credenciales → `401 invalid_credentials` genérico (no revela si existe
el usuario). Sin confirmar → `403 email_not_confirmed`
(`Confirma tu correo antes de entrar.`): la web debe ofrecer reenviar
(`POST /v1/auth/resend`).

## `POST /v1/auth/refresh` — renovar sesión

```json
{ "refreshToken": "…" }
```

(`refreshToken` 10–4096 caracteres.) `200` → `{ session }` nueva. Si falla →
`401 invalid_refresh` (`Sesión expirada. Inicia sesión de nuevo.`): el launcher
debe pedir login, no reintentar en bucle.

## `POST /v1/auth/logout` — cerrar sesión (auth)

Requiere `Authorization: Bearer <access_token>`. Revoca la sesión en servidor;
**el cliente debe borrar sus tokens** (si no, el access sigue válido hasta caducar).
`200` → `{ loggedOut: true }`. Aunque la revocación falle en servidor, responde
`loggedOut: true` (el cliente ya limpió lo suyo).
