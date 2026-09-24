# Contrato Yggdrasil implementado (MVP)

Base: `https://<tu-worker-yggdrasil>/` (local: `http://localhost:8787/`).
Todo JSON con `Content-Type: application/json`. UUID en cable **sin guiones**.
Errores con forma `{ error, errorMessage, cause? }`.

## Compatibilidad con StepLauncher

- **Guiones**: el launcher guarda y reenvía `accessToken`/`clientToken`/UUID
  con guiones; aquí se aceptan con o sin guiones y se comparan sin ellos.
- **`refresh` con perfil**: el launcher siempre manda `selectedProfile` en el
  refresh; se acepta si es el MISMO perfil atado (solo es `400` si pide otro
  distinto) y el nombre se refresca desde la base.
- **`GET /authserver`**: devuelve los mismos metadatos que `GET /`, para que
  el pre-verify pase aunque configuren la URL con ese sufijo.
- **Texturas**: `hasJoined` y `profile` devuelven `properties` con `textures`
  en base64 + firma `SHA1withRSA` cuando el jugador tiene skin/capa subida
  (vía web/`api/`); si no tiene, van sin `properties` (skin por defecto).
  No se expone `uploadableTextures`: la subida de skins vive en la web/`api/`,
  no en este servicio.

## `GET /` — metadatos + ALI

`200`:

```json
{
  "meta": {
    "serverName": "StepLauncher Yggdrasil",
    "implementationName": "steplauncher-yggdrasil",
    "implementationVersion": "0.1.0",
    "links": { "homepage": "https://…", "register": "https://…" },
    "feature.non_email_login": true
  },
  "skinDomains": ["steplauncher.pages.dev"],
  "signaturePublickey": "-----BEGIN PUBLIC KEY-----\n…\n-----END PUBLIC KEY-----\n"
}
```

Cabecera: `X-Authlib-Injector-API-Location: <raíz del servicio>`.

## `POST /authserver/authenticate` (`sensitive` 20/10 min)

```json
{ "username": "Steve_01", "password": "secreta123", "clientToken": "opcional", "requestUser": true, "agent": { "name": "Minecraft", "version": 1 } }
```

`200` → `{ accessToken, clientToken, availableProfiles: [{ id, name }],
selectedProfile: { id, name }, user?: { id, properties: [] } }`.
Errores: `403 ForbiddenOperationException` (credenciales o sin confirmar o sin
perfil) · `429` con `Retry-After`.

```bash
curl -X POST http://localhost:8787/authserver/authenticate \
  -H 'Content-Type: application/json' \
  -d '{"username":"Steve_01","password":"secreta123","requestUser":true,"agent":{"name":"Minecraft","version":1}}'
```

## `POST /authserver/refresh` (`account` 30/min)

```json
{ "accessToken": "…", "clientToken": "…", "requestUser": false }
```

Con `selectedProfile: { id, name }` en dos casos: para elegir perfil cuando el
token viejo no tenía ninguno, o repitiendo el MISMO perfil atado (lo que hace
el launcher StepLauncher siempre). `200` → `{ accessToken, clientToken,
selectedProfile?, user? }`.
Errores: `403` token inválido · `400 IllegalArgumentException` solo si ya
tenía perfil y se intenta atar otro DISTINTO.

## `POST /authserver/validate` y `/authserver/invalidate` (`account`)

```json
{ "accessToken": "…", "clientToken": "opcional" }
```

`validate`: `204` si vale, `403 Invalid token` si no.
`invalidate`: `204` siempre (revoca aunque el token no exista).

## `POST /authserver/signout` (`sensitive`)

```json
{ "username": "Steve_01", "password": "secreta123" }
```

Revoca **todos** los tokens del usuario → `204`. Credenciales mal →
`403` genérico.

## `POST /sessionserver/session/minecraft/join` (`join` 6/30 s)

```json
{ "accessToken": "…", "selectedProfile": "uuidsinguiones", "serverId": "hash-del-handshake" }
```

`204` si el token está vivo y el perfil coincide; `403 Invalid token` si no.
El registro vive 30 s.

## `GET /sessionserver/session/minecraft/hasJoined` (`account`)

`?username=Steve_01&serverId=…&ip?=` → `200` perfil o `204` vacío:

```json
{
  "id": "uuidsinguiones",
  "name": "Steve_01",
  "properties": [{ "name": "textures", "value": "base64…", "signature": "base64…" }]
}
```

`properties` SIEMPRE es un arreglo (vacío `[]` si no hay skin ni capa →
skin por defecto): así lo exige el parser del injector. Solo vale el ÚLTIMO
`join` del jugador. El `username` se compara case-insensitive (como Mojang).
Al verificarse, el join queda auditado (`verified_at` + `server_ip` del
servidor que verificó); los nunca verificados quedan en `null`.

## `GET /sessionserver/session/minecraft/profile/:uuid` (`account`)

`/session/minecraft/profile/853c80ef…?unsigned=true` → `200` perfil o
`204` si no existe. Con `unsigned=false` incluye la firma (el launcher pide
siempre `unsigned=false` para el avatar). Sin skin ni capa, sin `properties`.

## `GET /textures/:hash` (`public` 120/min)

Sirve el PNG de una skin/capa (`:hash` = SHA-256 en hex). Sin auth: el hash
es impredecible y solo se revela por `hasJoined`/`profile`. Responde
`image/png` con caché inmutable (la URL es eterna: si cambia el contenido,
cambia el hash). `404` si el hash es desconocido; `500` si falta la tabla
`ygg_textures` (re-pegar `db/01_ygg.sql` y mirar `wrangler tail`).

## `GET /sessionserver/blockedservers` (`public`)

El juego la pide al abrir el multijugador. Responde `[]` (nada bloqueado).

## `GET /skins/MinecraftSkins/:usuario.png` (`public`)

Respaldo de la API legacy para juegos viejos (pre-1.8): devuelve el PNG de
la skin actual o `404`. Normalmente el polyfill local del injector lo
resuelve sin llegar aquí.

## `GET /skins/:nombre` y `GET /cloaks/:nombre` (`public`, estilo Ely.by)

Skin/capa por nombre (con o sin `.png`, case-insensitive): PNG o `404`.
Para launchers, plugins de servidor y depuración rápida en el navegador.

## `GET /signature-verification-key.pem` (público)

La clave pública en PEM (la misma de `GET /`): sirve para verificar firmas
fuera del juego y para depurar rotaciones.

## `POST /api/profiles/minecraft` (`account`)

```json
["Steve_01", "Alex"]
```

`200` → `[{ "id": "…", "name": "Steve_01" }]` (solo existentes, sin orden
garantizado, máx. 10 por llamada).

## Límites y códigos

| HTTP | `error` | Cuándo |
|---|---|---|
| 400 | `IllegalArgumentException` | Cuerpo/query inválido o perfil ya asignado |
| 403 | `ForbiddenOperationException` | Credenciales o token inválidos (genérico) |
| 404 | `Not Found` | Ruta inexistente |
| 429 | `TooManyRequestsException` | Límite por IP (trae `Retry-After`) |
| 500 | `InternalError` | Error interno o sin entorno |

## Fuera del MVP (futuro)

- Subida Yggdrasil (`PUT/DELETE /api/user/profile/:uuid/:skin|cape`): no se
  implementa a propósito (ver ADR-Y07); la subida vive en la web/`api/`.
- Modelo `slim` en texturas: `api/` no guarda el modelo (clásico por defecto).
- `feature.enable_profile_key` (firmas de chat 1.19+) y fallback a Mojang.
