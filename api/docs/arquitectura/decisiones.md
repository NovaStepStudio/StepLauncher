# Decisiones de arquitectura (ADRs)

Registro de **qué se decidió y por qué**. Cada entrada tiene fecha, contexto,
decisión y consecuencias. Antes de proponer un cambio grande, leer esto
y añadir un ADR nuevo en vez de reescribir historia.

## ADR-001 — Cloudflare Workers como runtime (2026-09-20)

- **Contexto**: API para el launcher, con picos en lanzamientos y usuarios globales.
- **Decisión**: Workers en el edge (V8 isolates), sin servidor propio.
- **Consecuencias**: arranque ~0ms, escala automática, sin estado en disco;
  todo estado compartido va a Supabase. Código edge-safe (sin APIs de Node).
  El rate-limit en memoria es por instancia (ver ADR-006).

## ADR-002 — Hono como framework HTTP (2026-09-20)

- **Contexto**: router rápido y mantenible en Workers.
- **Decisión**: Hono (diseñado para edge, tipado, middlewares estándar).
- **Consecuencias**: estructura `src/index.ts` → `src/routes/vN/`; middlewares propios
  finos (`request-id`, `security-headers`, `cors`, `auth`, `rate-limit`, `validate`,
  `upload-quota`) en vez de librerías pesadas.

## ADR-003 — Supabase como datos + Auth (2026-09-20)

- **Contexto**: cuentas StepLauncher con registro, login y perfil.
- **Decisión**: Supabase Auth como fuente de identidad + Postgres con RLS como
  fuente de datos. El Worker valida el Bearer con `auth.getUser()` y usa clientes
  por petición (`supabaseForUser` con RLS, `supabaseAdmin` solo cuando sea
  imprescindible y nunca expuesto).
- **Consecuencias**: la seguridad real vive en políticas RLS, no solo en el Worker.
  Ver `docs/arquitectura/seguridad.md` y `docs/cuentas/README.md`.

## ADR-004 — Versionado por carpetas `/v1`, `/v2` (2026-09-20)

- **Contexto**: el launcher distribuido no se actualiza a la vez; romper v1 es inaceptable.
- **Decisión**: cada versión es una carpeta (`src/routes/v1/`, `src/schemas/v1/`,
  `docs/api/v1/`) montada en `/vN`. v2 se crea **copiando** v1, nunca editando v1
  para romper. Lo mismo para el SQL: `db/v1/`, `db/v2/…`.
- **Consecuencias**: duplicación asumida a cambio de compatibilidad.
  Ver `docs/arquitectura/versionado.md`.

## ADR-005 — Zod como contratos de entrada (2026-09-20)

- **Contexto**: evitar inyecciones y que un endpoint contamine a otro por tipos rotos.
- **Decisión**: todo JSON/query/param de cuenta se valida con Zod (`validateJson`,
  `validateQuery`, `validateParam`, `.strict()` donde aplique). Esquemas por versión.
- **Consecuencias**: 400 con `code: "validation_error"` y detalles por campo.

## ADR-006 — Rate-limit en memoria + camino a KV (2026-09-20)

- **Contexto**: frenar fuerza bruta en login/registro sin añadir latencia.
- **Decisión**: ventana fija por IP en memoria por instancia (rápido, no distribuido),
  con presets `public` (120/min), `account` (30/min) y `sensitive` (20/10 min).
  Cuando haya KV, migrar a conteo distribuido sin cambiar la interfaz del middleware.
- **Consecuencias**: límite conocido y documentado en `docs/api/v1/README.md`;
  bajo N instancias el límite real se multiplica (aceptado en v1).

## ADR-007 — Registro/login por API con `service_role` (2026-09-20)

- **Contexto**: el launcher pide email + usuario + contraseña en un solo formulario;
  Supabase Auth no conoce el `username` único de StepLauncher.
- **Decisión**: endpoints `POST /v1/auth/register|login|refresh|logout` en el Worker
  con `supabaseAdmin()` SOLO en servidor (alta, resolución username→email vía función
  `SECURITY DEFINER` revocada a `anon`/`authenticated`, revocación). `email_confirm: true`
  en MVP para no friccionar el registro. Contraseñas solo en tránsito HTTPS, jamás en logs.
- **Consecuencias**: la API toca la llave más sensible; este dominio exige preset
  `sensitive` + revisión de secretos. Reversible a flujo PKCE con navegador
  (ver `docs/oauth/README.md`).

## ADR-008 — Storage privado + libro de subidas para quota (2026-09-20)

- **Contexto**: skins/capas son de cada jugador y la quota es 10 archivos/día.
- **Decisión**: buckets `skins`/`capes` privados con políticas por carpeta `<user_id>/`
  (+ `avatars`/`banners` públicos sin listado global), tabla `file_uploads` como libro
  y quota calculada en la API contando 24h. Descargas por URL firmada de 1h.
- **Consecuencias**: quota aproximada bajo carreras (suficiente); sin transacción entre
  storage y libro → limpieza best-effort de huérfanos en el handler.

## ADR-009 — Extras de cuenta + cierre de listado global (2026-09-20)

- **Contexto**: la cuenta pedía bio (≤5000), última versión de Minecraft Java y
  notificaciones; Supabase avisaba de listado global en buckets públicos.
- **Decisión**: columnas `bio` + `last_mc_version` en `profiles`; `createdAt`/
  `lastSessionAt` desde `auth.users` vía admin en `GET /me` (autoritativo, sin
  duplicar); tabla `notifications` con RLS propia sin INSERT de cliente; buckets
  públicos sin política SELECT amplia (la descarga por URL pública sigue igual,
  lo que se cierra es el listado).
- **Consecuencias**: regla SQL — todo reejecutable; ver `db/v1/README.md`.

## ADR-010 — UUID Minecraft, amistades y banners 1080p (2026-09-20)

- **Contexto**: UUID de Minecraft obtenible, solicitudes por email/usuario/UUID,
  4 opciones de privacidad y banners 1080p.
- **Decisión**: `mc_uuid` fijado por el jugador (UUID válido y único, `""` desvincula;
  sin dependencia de Mojang) y expuesto en `/me` y búsquedas visibles. Amistades con
  `friend_requests` + `friendships` (doble fila) y privacidad aplicada en SQL
  (`resolve_friend_target`, `search_users`, RLS). Banner PNG/GIF/JPG/WEBP hasta
  1920×1080 verificado por cabecera (`src/lib/images.ts`), bucket `banners` sin listado.
- **Consecuencias**: aceptar es atómico en SQL; `receive_notifications` se consulta al
  notificar; la quota 10/día sigue siendo solo skins+capas.

## ADR-011 — UUID de Minecraft garantizado para todos (2026-09-20)

- **Contexto**: todo jugador debe tener UUID obtenible desde el minuto uno.
- **Decisión**: `offline_mc_uuid()` en SQL (MD5 de `OfflinePlayer:<username>` con
  versión 3 y variante RFC 4122); backfill a existentes, trigger que auto-rellena al
  crear (y si se deja en null), columna `NOT NULL`. El auto nunca pisa un premium
  enlazado ni rota al cambiar el username. `""` en `PATCH` regenera el offline.
- **Consecuencias**: `mcUuid` siempre presente en `/me`; invariante a nivel BD.

## ADR-012 — Comunidad: preview, cosméticos y lista negra (2026-09-20)

- **Contexto**: previsualizar perfiles por UUID/usuario/UUID-MC con equipados,
  pudiendo actuar (solicitar/retirar) y con bloqueos que cierran el perfil.
- **Decisión**: `GET /v1/community/profiles/:identifier` (auth) con visibilidad por
  capas — lista negra (aviso 200 sin datos) → amigos/uno mismo → `searchable` — y
  `relationship` con `requestId` para retirar la enviada sin buscarla. Email jamás
  expuesto. Cosméticos: catálogo + posesión con `equipped`; 4 de ejemplo auto-otorgados;
  la preview solo muestra equipados. `blocks` corta pendientes y amistad al bloquear;
  solicitar a quien te bloqueó es 404 genérico, a quien bloqueaste es 403.
- **Consecuencias**: lo social sensible vive en funciones `SECURITY DEFINER` revocadas.

## ADR-013 — SQL versionado por paquete v1 instalable de una vez (2026-09-20)

- **Contexto**: 8 migraciones sueltas que había que pegar y ejecutar una por una,
  con estados intermedios frágiles y dos huecos (cosmético nuevo sin reparto a
  usuarios viejos; `NOT NULL` de `mc_uuid` con posible carrera).
- **Decisión**: paquete `db/v1/` (partes `01–07` como fuente + `install.sql`
  generado con `bun run db:build` para pegar UNA vez en una transacción +
  `verify.sql` con 9 comprobaciones + `schema_version` como trazabilidad).
  `db/migrations/0001–0008` queda congelado como historial.
- **Consecuencias**: una ejecución cubre instalación nueva y BD existente;
  v1 solo admite cambios compatibles; lo incompatible exige `db/v2/`.

## ADR-014 — Reconstrucción de `docs/` desde el código (2026-09-20)

- **Contexto**: `docs/` desapareció del disco junto con `db/migrations/` por una
  modificación externa al repo; los contratos solo vivían en el código.
- **Decisión**: recrear `docs/` entera desde `src/` (contratos verificados contra
  rutas, esquemas y middlewares reales, sin inventar endpoints) y `db/v1/`.
  `docs/oauth/` documenta el modelo real de sesiones + el camino PKCE como futuro,
  no como existente.
- **Consecuencias**: la docs vuelve a ser la memoria del proyecto; cualquier
  divergencia futura se detecta comparando con el código.

## Plantilla para el próximo ADR

```md
## ADR-00X — Título (AAAA-MM-DD)

- **Contexto**: ...
- **Decisión**: ...
- **Consecuencias**: ...
```
