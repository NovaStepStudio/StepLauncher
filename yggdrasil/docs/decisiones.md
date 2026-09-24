# Decisiones de Yggdrasil (ADRs)

Registro de **qué se decidió y por qué**. Antes de un cambio grande, leer esto
y añadir un ADR nuevo en vez de reescribir historia.

## ADR-Y01 — Servicio separado fuera de `api/` (2026-09-23)

- **Contexto**: el usuario exige carpeta propia para Yggdrasil, sin tocar la
  API principal. El protocolo exige rutas en raíz (`/authserver/*`) y errores
  propios, incompatibles con `/v1` y su sobre `ok()`/`fail()`.
- **Decisión**: subproyecto `yggdrasil/` con Worker propio
  (`steplauncher-yggdrasil`), sin imports compartidos con `api/`. Lee la misma
  base Supabase pero con tablas propias (`ygg_tokens`, `ygg_joins`).
- **Consecuencias**: dos despliegues (api + yggdrasil); la raíz debe exceptuar
  `yggdrasil/wrangler.jsonc` en `.gitignore` (cambio cruzado pendiente de
  permiso explícito). `api/` queda intacta.

## ADR-Y02 — Sin registro aquí: Supabase como fuente única (2026-09-23)

- **Contexto**: `api/` ya es dueña de cuentas (registro, confirmación,
  recupero, `profiles` con `mc_uuid` garantizado).
- **Decisión**: Yggdrasil verifica con `signInWithPassword` + lee `profiles`;
  exige email confirmado. Nada de altas, nada de doble fuente.
- **Consecuencias**: sin cuenta confirmada no hay juego; el registro/recupero
  se hace en la web/`api/`.

## ADR-Y03 — Tokens Yggdrasil propios, no JWT Supabase (2026-09-23)

- **Contexto**: el protocolo pide `accessToken`/`clientToken` opacos con
  rotación y perfil atado, ajenos a Supabase Auth.
- **Decisión**: tabla `ygg_tokens`, expiración 15 días, máx. 10 vivos por
  usuario (se revoca la más vieja), rotación en `refresh`, revocación total
  en `signout`, `join` de 30 s en `ygg_joins`.
- **Consecuencias**: dos familias de tokens conviven (Supabase para la cuenta,
  Yggdrasil para el juego) y nunca se mezclan.

## ADR-Y04 — Formato de errores del protocolo, no el de `api/` (2026-09-23)

- **Contexto**: `api/` obliga a `ok()`/`fail()`; authlib-injector y el juego
  solo entienden `{ error, errorMessage, cause? }` con sus HTTP.
- **Decisión**: todo este servicio responde en forma Yggdrasil, incluidos
  rate-limit (429) y validación (400). El 404 y el 500 usan la misma forma.
- **Consecuencias**: excepción documentada y consciente a la regla de `api/`;
  no se copia `respond.ts` a propósito.

## ADR-Y05 — MVP sin skins: perfiles sin `textures` (2026-09-23)

- **Contexto**: firmar texturas exige clave RSA, subida con re-codificación
  PNG, hash como nombre, whitelist y quota (riesgo de RCE y PNG-bomb).
- **Decisión**: MVP devuelve perfiles sin `properties`; la infraestructura de
  firma (`YGG_SIGN_PRIVATE_KEY`, `signaturePublickey`) ya queda cableada para
  la fase 2.
- **Consecuencias**: skins por defecto hasta la fase 2; `unsigned=false` aún
  no firma (documentado en `contrato.md`).

## ADR-Y06 — Compatibilidad real con StepLauncher (2026-09-23)

- **Contexto**: el launcher ya habla Yggdrasil (`internal/Core/Auth/Authlib.go`,
  `internal/Core/Accounts/Manager.go`, `internal/Core/Launcher/Launcher.go`),
  pero con dos hábitos que el spec estricto rechaza: guarda y reenvía
  `accessToken` con guiones (`normalizeAccessToken`) y manda `selectedProfile`
  en TODOS los refresh. Sin adaptarlo, cada validate/refresh/join daría 403/400.
- **Decisión**: aceptar tokens/UUID con o sin guiones (se comparan sin ellos),
  aceptar refresh con el MISMO perfil atado (400 solo ante perfil distinto,
  con nombre refrescado desde la base) y responder metadatos también en
  `GET /authserver` (el pre-verify del launcher espera 200 en la URL configurada).
- **Consecuencias**: cero cambios en `launcher/`; el servidor es tolerante en
  lo que recibe y estricto en lo que emite (responde sin guiones, como el spec).

## ADR-Y07 — Skins en lectura, subida en `api/` (2026-09-24)

- **Contexto**: el launcher pide `profile?unsigned=false` para el avatar y el
  juego necesita `properties.textures` firmadas; sin eso todo carga sin skin
  (ADR-Y05 lo había diferido). `api/` ya guarda PNGs verificados (≤1 MB) en
  buckets `skins`/`capes` con libro `file_uploads` (rutas únicas por subida).
- **Decisión**: Yggdrasil LEE la última subida por tipo, sirve los bytes por
  `GET /textures/:hash` (SHA-256, `image/png`, caché inmutable) y firma el
  payload con `YGG_SIGN_PRIVATE_KEY`. Mapa hash→ubicación en `ygg_textures`.
  La SUBIDA Yggdrasil no se implementa: vive en la web/`api/` (por eso no hay
  `uploadableTextures`). Sin skin/capa, perfil sin `properties` (defecto).
- **Consecuencias**: el avatar del launcher y la skin del juego funcionan sin
  tocar `api/`; `/textures/:hash` es público por diseño (hash impredecible,
  igual que Mojang). Limitación honesta: no se re-codifica el PNG al servir
  (se sirve tal cual lo validó `api/`); el modelo es siempre clásico (`api/`
  no guarda slim).

## ADR-Y08 — Auditoría de joins verificados (2026-09-24)

- **Contexto**: `ygg_joins` solo se llena con handshake online; en servidores
  offline o con auth por plugins el cliente nunca anuncia `join` y verla
  vacía es normal (no un bug). Pero cuando sí hay joins, interesa saber
  cuáles llegaron a verificarse y desde qué servidor.
- **Decisión**: columnas `verified_at` + `server_ip` (IP del servidor que
  llamó a `hasJoined`), escritas best-effort al verificar. La tabla sigue
  siendo interna del Worker (`service_role`, sin RLS de cliente).
- **Consecuencias**: re-pegar `db/01_ygg.sql` (idempotente, agrega columnas);
  dashboard de uso real por servidor sin tabla nueva.

## ADR-Y09 — Paridad Ely.by: clave pública y endpoints por nombre (2026-09-24)

- **Contexto**: Ely.by (referencia madura) expone su pública para verificar
  firmas, sirve skins/capas por nombre y documenta montaje sin JVM-args y
  versiones <1.7.2. Nuestro servicio ya firmaba y servía por hash, pero
  faltaba esa capa de acceso y de docs.
- **Decisión**: `GET /signature-verification-key.pem` (misma clave de
  `GET /`), `GET /skins/:nombre` + `GET /cloaks/:nombre` (además del legacy
  `/skins/MinecraftSkins/:usuario.png`), y docs de BungeeCord-sin-args y de
  versiones soportadas (1.7.2+ vía injector; <1.7.2 fuera de alcance).
  El `refresh` acepta repetir el mismo perfil y rechaza uno distinto (más
  cercano al spec que Ely, que ignora el parámetro).
- **Consecuencias**: paridad de acceso con Ely sin copiar su modelo de
  cuentas; la subida sigue en la web/`api/` (sin `uploadableTextures`).

## Plantilla para el próximo ADR

```md
## ADR-Y0X — Título (AAAA-MM-DD)

- **Contexto**: ...
- **Decisión**: ...
- **Consecuencias**: ...
```
