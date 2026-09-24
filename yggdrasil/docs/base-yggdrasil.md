# Base Yggdrasil — qué es al completo

Yggdrasil es el **sistema de autenticación de Minecraft: Java Edition**
introducido por Mojang en la 1.6.1 (sustituyó al login legacy). Hoy el juego
oficial usa cuentas Microsoft, pero Yggdrasil sigue vivo como **protocolo
abierto** gracias a **authlib-injector**: cualquier servidor y launcher puede
implementarlo para tener login propio sin Mojang/Microsoft.

## Idea en una frase

El juego nunca ve la contraseña: el launcher consigue un `accessToken`, el
cliente lo anuncia al entrar a un servidor (`join`) y el servidor lo verifica
(`hasJoined`). Si ambos hablan con el **mismo** servidor Yggdrasil, la sesión
vale.

## Actores y piezas

- **Usuario**: cuenta con email + contraseña (en nuestro caso, Supabase Auth).
- **Perfil**: el jugador dentro del juego (`uuid` + `name` + texturas). Un
  usuario puede tener varios perfiles; Mojang oficial solo permite uno.
- **Token**: credencial opaca (`accessToken` + `clientToken`) con perfil atado,
  vigencia (aquí 15 días) y estados válida → revocada. Hay límite por usuario
  (aquí 10, se revoca la más vieja).
- **Servidor Yggdrasil** (este servicio): emite y verifica tokens y perfiles.
- **Cliente Minecraft + authlib-injector**: `-javaagent` que redirige el login
  del juego hacia nuestro servidor.
- **Servidor Minecraft en `online-mode=true` + authlib-injector**: verifica
  cada entrada contra nuestro `hasJoined`.

## Los tres dominios históricos de Mojang

| Dominio | Base original | Qué hace |
|---|---|---|
| Auth | `https://authserver.mojang.com` | `authenticate`, `refresh`, `validate`, `invalidate`, `signout` |
| Session | `https://sessionserver.mojang.com` | `join` (cliente), `hasJoined` + `profile` (servidor) |
| API | `https://api.mojang.com` | `profiles/minecraft` (nombres → UUID) |

authlib-injector permite servir los tres bajo **una sola raíz** (este Worker):
`POST /authserver/authenticate`, `POST /sessionserver/session/minecraft/join`,
etc. La raíz `GET /` devuelve **metadatos** para autoconfiguración.

## Flujo completo (entrada a un servidor)

1. El jugador escribe email/usuario + contraseña en el launcher.
2. El launcher llama `POST /authserver/authenticate` → recibe `accessToken`,
   `clientToken` y `selectedProfile` (`uuid` sin guiones + `name`).
3. El launcher guarda los tokens y lanza el juego con:
   `-javaagent:authlib-injector.jar=<URL de este servicio>` más el supuesto
   `--uuid <uuid> --accessToken <accessToken> --userType mojang`.
4. Al entrar a un servidor, cliente y servidor negocian un `serverId` (hash del
   handshake de cifrado; para nosotros es opaco).
5. El **cliente** llama `POST /sessionserver/session/minecraft/join` con
   `{ accessToken, selectedProfile, serverId }`. El servicio lo guarda 30 s
   junto a la IP.
6. El **servidor** llama
   `GET /sessionserver/session/minecraft/hasJoined?username=&serverId=`:
   si hay un `join` vivo con ese `serverId` y el mismo `username`, responde
   `200` con el perfil completo; si no, `204` y el juego expulsa al jugador.

## Endpoints del spec (resumen fiel)

- `POST /authserver/authenticate` — `{ username, password, clientToken?,
  requestUser?, agent? }` → `{ accessToken, clientToken, availableProfiles,
  selectedProfile?, user? }`. Sin `clientToken` el servidor genera uno.
  Sin perfil (cuenta sin Minecraft) no hay `selectedProfile`.
- `POST /authserver/refresh` — revoca el viejo y emite uno nuevo con el mismo
  `clientToken`. Acepta `selectedProfile` solo para elegir perfil cuando el
  token aún no tiene ninguno; si ya tiene, es `400`.
- `POST /authserver/validate` — `204` si el token vale, `403` si no. Acepta
  `clientToken` opcional (si se envía, debe coincidir).
- `POST /authserver/invalidate` — revoca (`204` siempre, ignore `clientToken`).
- `POST /authserver/signout` — `{ username, password }` revoca **todos** los
  tokens del usuario (`204`).
- `POST /sessionserver/session/minecraft/join` — `{ accessToken,
  selectedProfile, serverId }` → `204`. Solo vale con token vivo y perfil
  coincidente.
- `GET /sessionserver/session/minecraft/hasJoined` — `?username&serverId&ip?`
  → `200` perfil completo o `204` si falla.
- `GET /sessionserver/session/minecraft/profile/:uuid?unsigned=` — perfil
  completo con `properties` (y firma si `unsigned=false`).
- `POST /api/profiles/minecraft` — `["Notch", …]` (máx. 10) → `[{ id, name }]`.
- Texturas `PUT/DELETE /api/user/profile/:uuid/:skin|cape` — subida con
  `multipart` + `Authorization: Bearer <accessToken>` (fase 2, aún no en MVP).

## Modelos del protocolo

- **UUID sin guiones** en todo el cable (`853c80ef3c3749fdaa49938b674adae6`).
- **Perfil**: `{ id, name, properties? }`; cada propiedad
  `{ name, value, signature? }`. La propiedad `textures` es un base64 de
  `{ timestamp, profileId, profileName, textures: { SKIN: { url, metadata? },
  CAPE: { url } } }`, firmada con **SHA1withRSA**.
- **Firma**: el servidor firma con su privada; los clientes verifican con la
  pública de `GET /` (`signaturePublickey`, PEM). Sin la pública correcta, el
  juego rechaza las texturas si exige firma.
- **URL de textura**: el nombre del archivo tras el último `/` es el hash que
  el cliente usa como caché; debe servirse como `image/png`. Se recomienda
  SHA-256 del PNG. El PNG debe re-codificarse al subir (quitar metadatos,
  validar tamaño antes de leerlo entero: PNG-bomb y código oculto).
- **Whitelist `skinDomains`**: el juego solo descarga texturas de esos
  dominios (más `.minecraft.net`/`.mojang.com` por defecto). Regla: si empieza
  por `.` es sufijo; si no, igualdad exacta.

## Metadatos y descubrimiento (extensión authlib-injector)

`GET /` → `{ meta, skinDomains, signaturePublickey }`, con `meta` libre:

- `serverName`, `implementationName/Version`, `links: { homepage, register }`.
- Flags `feature.*`: `non_email_login` (aquí `true`: vale usuario además de
  email), `legacy_skin_api`, `no_mojang_namespace`, `enable_mojang_anti_features`,
  `enable_profile_key`, `username_check` (aquí `false` en MVP).
- Cabecera **ALI** `X-Authlib-Injector-API-Location`: apunta a la raíz real
  del servicio para que el launcher la descubra desde una URL corta.

## Errores del protocolo

Forma `{ "error": "…", "errorMessage": "…", "cause?": "…" }`:

| HTTP | `error` | Cuándo |
|---|---|---|
| 400 | `IllegalArgumentException` | Token ya con perfil, cuerpo inválido |
| 403 | `ForbiddenOperationException` | Credenciales o token inválidos (genérico a propósito) |
| 403 | `TooManyRequestsException` | Token inexistente tras refresh |
| 429 | `TooManyRequestsException` | Límite excedido (con `Retry-After`) |
| 204 | — (vacío) | `validate`/`invalidate`/`signout`/`join` OK; `hasJoined`/`profile` fallido |

Fuentes: wiki de Mojang (Yggdrasil), spec técnica de authlib-injector
(`yggdrasil-server-technical-specification`) y comportamiento observado de
`sessionserver`/`authserver` oficiales.
