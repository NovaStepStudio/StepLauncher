# API v1 — Contrato público (lo que consume el launcher)

Base URL: `https://<tu-worker>/v1` (local: `http://localhost:8787/v1`).
Todo lo que no sea `/` (índice) cuelga de `/v1`. v1 nunca se rompe: lo
incompatible va a `/v2` (ver `docs/arquitectura/versionado.md`).

## Sobre único de respuesta (`src/lib/respond.ts`)

Éxito (`200`, o `201` en creaciones):

```json
{ "success": true, "data": { "...": "..." } }
```

Error (el launcher gestiona por `code`, nunca por `message`):

```json
{
  "success": false,
  "error": { "code": "unauthorized", "message": "Autenticación requerida.", "requestId": "…" }
}
```

Los 400 de validación añaden `details` por campo (`{ "username": ["Mínimo 3 caracteres."] }`).
El 404 global y los errores no controlados usan el mismo sobre.

## Auth

Cabecera en todo endpoint privado: `Authorization: Bearer <access_token>`.
Sesiones: `POST /v1/auth/register|login` → guardar `accessToken` + `refreshToken`;
renovar con `POST /v1/auth/refresh`; revocar con `POST /v1/auth/logout` y borrar
tokens en el dispositivo. Detalle: `auth.md` + `docs/oauth/README.md`.

## Cabeceras

- Petición: `Content-Type: application/json` (o `multipart/form-data` en subidas),
  `Authorization: Bearer …`, opcional `X-Request-Id` propio (el servidor siempre
  devuelve el suyo en `X-Request-Id` y en `error.requestId`).
- Límites: el 429 trae `Retry-After` en segundos.

## Rate-limit por IP (`src/middleware/rate-limit.ts`, memoria por instancia)

| Preset | Ventana | Máx | Dónde |
|---|---|---|---|
| `public` | 1 min | 120 | `GET /v1/health` |
| `account` | 1 min | 30 | Todo lo de cuenta por defecto |
| `sensitive` | 10 min | 20 | `register`, `login`, `refresh`, `email-change`, `password-change`, `friends/requests` (POST), `friends/blocks` (POST) |

Excedido → `429 rate_limited` (`Demasiadas peticiones…`).

## Códigos de error (estables)

| HTTP | `code` | Cuándo |
|---|---|---|
| 400 | `validation_error` | Cuerpo/query/param no cumple el contrato (+ `details`) |
| 400 | `weak_password` | Contraseña rechazada por Auth al registrar |
| 400 | `same_email` | Pedir cambio al email que ya usas |
| 400 | `cannot_add_self` | Solicitud de amistad a uno mismo |
| 400 | `cannot_block` | Bloqueo imposible (a uno mismo o sin perfil) |
| 400 | `file_too_large` | Supera el máximo (skins/capas 1 MB, avatar 2 MB, banner 8 MB) |
| 400 | `invalid_file_type` | Firma mágica no válida (no fiarse de la extensión) |
| 400 | `invalid_dimensions` | Banner mayor de 1920×1080 |
| 401 | `unauthorized` | Falta Bearer o sesión inválida/expirada (genérico a propósito) |
| 401 | `invalid_credentials` | Login fallido (genérico: no revela si existe el usuario) |
| 401 | `invalid_refresh` | Refresh inválido/expirado (volver a login) |
| 401 | `current_password_incorrect` | La actual no coincide al cambiar contraseña |
| 403 | `user_blocked` | Solicitar a quien bloqueaste (desbloquea primero) |
| 404 | `not_found` | Recurso ajeno/inexistente (genérico) |
| 404 | `user_not_found` | Destinatario inexistente o que te bloqueó (genérico a propósito) |
| 404 | `profile_not_found` | Preview sin perfil visible |
| 404 | `cosmetic_not_owned` | Equipar lo que no posees |
| 409 | `username_taken` | Usuario en uso (registro o cambio) |
| 409 | `email_taken` | Email ya registrado |
| 409 | `mc_uuid_taken` | Ese UUID de Minecraft ya está enlazado |
| 409 | `already_friends` | Ya sois amigos |
| 409 | `already_pending` | Ya hay solicitud pendiente entre ambos |
| 429 | `rate_limited` | Límite por IP excedido |
| 429 | `upload_quota_exceeded` | 10 skins+capas/día alcanzados |
| 500 | `internal_error` | Error interno (genérico, con `requestId` para soporte) |
| 500 | `server_misconfigured` | Servicio no disponible (falta entorno) |
| 500 | `upload_failed` | Storage rechazó el guardado |

## Paginación

`?limit=` (defecto 20; máx 50 en archivos/notificaciones, 20 en búsqueda) y
`?offset=` (defecto 0). Las listas devuelven `limit` y `offset` de eco.

## Índice de endpoints

| Dominio | Archivo | Rutas |
|---|---|---|
| Salud | `health.md` | `GET /v1/health` |
| Auth | `auth.md` | `POST /v1/auth/register|login|refresh|logout` |
| Cuenta | `cuentas.md` | `GET|PATCH /v1/accounts/me`, presencia, email, contraseña, avatar, banner, cosméticos, privacidad |
| Archivos | `archivos.md` | `POST /v1/files/skin|cape`, `GET /v1/files`, `GET /v1/files/:id/url`, `DELETE /v1/files/:id` |
| Notificaciones | `notificaciones.md` | `GET /v1/notifications`, `PATCH …/:id/read`, `POST …/read-all`, `DELETE …/:id` |
| Amigos | `amigos.md` | Solicitudes, amistades y `blocks` bajo `/v1/friends` |
| Usuarios | `usuarios.md` | `GET /v1/users/search` |
| Comunidad | `comunidad.md` | `GET /v1/community/profiles/:identifier` |
