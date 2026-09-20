# Auth v1 — registro, login, sesiones

Sesiones **Supabase** (`access_token` + `refresh_token`); la API no inventa
criptografía (ADR-007). Modelo completo: `docs/oauth/README.md`.
Rate-limit `sensitive` en register/login/refresh (20/10 min), `account` en logout.

## `POST /v1/auth/register` — alta con email + usuario + contraseña

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

`201`:

```json
{
  "success": true,
  "data": {
    "user": { "id": "uuid", "email": "jugador@ejemplo.com", "username": "Steve_01" },
    "session": { "accessToken": "…", "refreshToken": "…", "expiresAt": 123, "tokenType": "bearer" }
  }
}
```

El perfil se crea por trigger (`handle_new_user`) con ese username; el launcher
guarda la sesión y ya puede llamar a `/v1/accounts/me`. `email_confirm: true`
en MVP (sin fricción de correo).

Errores: `validation_error` 400 · `username_taken` 409 · `email_taken` 409 ·
`weak_password` 400. Caso raro: cuenta creada pero sin sesión → 500
(`Cuenta creada, pero inicia sesión manualmente`: la cuenta SÍ existe, pedir login).

## `POST /v1/auth/login` — email O usuario + contraseña

```json
{ "identifier": "Steve_01", "password": "secreta123" }
```

`identifier` 3–254 (con `@` se trata como email; si no, se resuelve username→email
en servidor). `200` devuelve `{ user: { id, email, username }, session }` igual
que registro (`username` puede ser `null` si el trigger aún no creó el perfil).

```bash
curl -X POST http://localhost:8787/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"Steve_01","password":"secreta123"}'
```

Error de credenciales → `401 invalid_credentials` genérico (no revela si existe
el usuario).

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
