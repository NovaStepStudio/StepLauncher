# Cuentas v1 — perfil propio y personalización

Todo exige `Authorization: Bearer` (el dueño sale del token, nunca del cliente) y
rate-limit `account`, salvo email/contraseña que usan `sensitive`.
Modelo: `docs/cuentas/README.md`. SQL: `db/v1/`.

## `GET /v1/accounts/me` — perfil propio completo

```bash
curl http://localhost:8787/v1/accounts/me -H 'Authorization: Bearer …'
```

`200`:

```json
{
  "success": true,
  "data": {
    "id": "uuid", "email": "jugador@ejemplo.com",
    "username": "Steve_01", "displayName": "Steve",
    "avatarUrl": "https://…/avatar.png", "bio": "…",
    "lastMcVersion": "1.21", "mcUuid": "uuid-mc",
    "bannerUrl": "https://…/banner.png", "isOnline": true,
    "createdAt": "2026-…", "lastSessionAt": "2026-…"
  }
}
```

`createdAt`/`lastSessionAt` vienen de `auth.users` vía admin (autoritativo).

## `PATCH /v1/accounts/me` — cambiar perfil

Todo opcional, al menos un campo (estricto, campos desconocidos = 400):

```json
{ "username": "Nuevo_01", "displayName": "Nuevo", "bio": "", "lastMcVersion": "1.21", "mcUuid": "", "isOnline": false }
```

Reglas: `username` 3–20 `A-Za-z0-9._-` (único ci → `username_taken` 409);
`displayName` 3–32 (letras, números, espacios, guiones); `bio` ≤5000 (`""` la
borra, se guarda `null`); `lastMcVersion` formato MC (ej `1.21`); `mcUuid` UUID
válido o `""` (desvincula y **regenera el offline**); duplicado → `mc_uuid_taken`
409. `200` devuelve el perfil con la misma forma del `GET`.

## `PATCH /v1/accounts/me/presence` — presencia propia

Heartbeat del launcher (encender al abrir, apagar al cerrar):

```json
{ "isOnline": true }
```

`200` → `{ isOnline: true }`. También editable vía `PATCH /me`.

## `POST /v1/accounts/me/email-change` — pedir cambio de email (`sensitive`)

```json
{ "newEmail": "nuevo@ejemplo.com" }
```

La confirmación viaja al correo **nuevo**. `200` →
`{ emailChangeRequested: true, message: "Revisa tu nuevo correo…" }`.
Errores: `same_email` 400 (ya usas ese) · `email_taken` 409.

## `POST /v1/accounts/me/password-change` — cambiar contraseña (`sensitive`)

```json
{ "currentPassword": "…", "newPassword": "…distinta…" }
```

Verifica la actual (si falla: `current_password_incorrect` 401, sin más detalle);
la nueva 8–72 y distinta. `200` → `{ passwordChanged: true }`.

## `POST /v1/accounts/me/avatar` — subir avatar (`multipart`)

Campo `avatar`, imagen ≤2 MB, firma real PNG/JPEG/WebP (no vale la extensión).
Se guarda en `avatars/<user_id>/avatar.<ext>` (público, upsert) y se registra en
el libro. `200` → `{ avatarUrl }`. Errores: `validation_error` (falta campo) ·
`file_too_large` · `invalid_file_type` · `upload_failed`.

```bash
curl -X POST http://localhost:8787/v1/accounts/me/avatar \
  -H 'Authorization: Bearer …' -F 'avatar=@yo.png'
```

## `POST /v1/accounts/me/banner` — subir banner (`multipart`, máx 1080p)

Campo `banner`, ≤8 MB, formato real PNG/GIF/JPG/WEBP verificado por cabecera
(`src/lib/images.ts`) y dimensiones **≤1920×1080**. `200` → `{ bannerUrl }`.
Si supera: `invalid_dimensions` 400 (`El banner no puede superar los 1920×1080`).

## Cosméticos propios — equipar / desequipar

```bash
curl -X POST http://localhost:8787/v1/accounts/me/cosmetics/equip \
  -H 'Authorization: Bearer …' -H 'Content-Type: application/json' \
  -d '{"cosmeticId":"uuid-del-cosmetico"}'
```

`cosmeticId` UUID estricto; solo lo que **posees** (se otorgan al crear la cuenta).
`200` → `{ cosmeticId, equipped: true|false }`. Si no lo posees:
`cosmetic_not_owned` 404. Rutas: `/me/cosmetics/equip` y `/me/cosmetics/unequip`.

## Privacidad propia — ver y cambiar

`GET /v1/accounts/me/privacy` → `200`:

```json
{ "success": true, "data": { "searchable": true, "allowEmailSearch": true, "receiveFriendRequests": true, "receiveNotifications": true } }
```

`PATCH /v1/accounts/me/privacy` con al menos una de las 4 (estricto) devuelve la
forma completa. Qué hace cada una (se aplica **en SQL**, ver `amigos.md`):

| Opción | Efecto en `false` |
|---|---|
| `searchable` | No apareces en `/v1/users/search` ni por usuario en solicitudes |
| `allowEmailSearch` | No te encuentran por email |
| `receiveFriendRequests` | Nadie puede solicitarte (UUID/email/usuario) |
| `receiveNotifications` | No recibes avisos (p. ej. solicitudes) |
