# SQL de StepLauncher API (`api/db/`)

Toda la base de datos versionada en el repo: lo que no está aquí, no existe en
producción. Supabase Auth (`auth.users`) es la identidad; este SQL crea
el perfil, el almacenamiento y el libro de subidas con RLS.

## Camino soportado: paquete versionado (un solo pegar-y-ejecutar)

**`v1/install.sql`** — todo lo que necesita la API v1 en UNA transacción.
Supabase → SQL Editor → pegar completo → Run. Después, `v1/verify.sql`.
Detalle: `v1/README.md`.

- Instalación nueva: crea todo.
- BD existente (con `migrations/0001–0008` aplicadas): converge al mismo
  estado final sin duplicar ni borrar datos (idempotente).
- Trazabilidad: registra `v1` en `public.schema_version`.
- v2 será `db/v2/` con su propio `install.sql` (v1 no se reescribe a mano).

## Historial congelado (no usar para instalaciones nuevas)

`migrations/0001–0008.sql` es el historial de cómo se llegó a v1: **no se
modifica ni se aplica por partes**. Queda como referencia. El paquete `v1/`
es la consolidación limpia y transaccional de esas 8 migraciones más dos
mejoras de estabilidad (auto-otorga de cosméticos futuros y orden
trigger-antes-que-relleno en `mc_uuid`).

## Orden antiguo de aplicación (solo referencia histórica)

1. `migrations/0001_profiles.sql` — tabla `profiles` + RLS + funciones + trigger.
2. `migrations/0002_storage.sql` — buckets `avatars` (público), `skins`/`capes` (privados).
3. `migrations/0003_uploads.sql` — tabla `file_uploads` (quota diaria).
4. `migrations/0004_account_extras.sql` — `bio` + `last_mc_version`, tabla
   `notifications` y fix del aviso de listado en `avatars`.
5. `migrations/0005_social.sql` — `mc_uuid` + `banner_url`, `privacy_settings`,
   `friend_requests` + `friendships` + 5 funciones, bucket `banners` y `banner`
   en el libro de subidas.
6. `migrations/0006_mc_uuid_auto.sql` — UUID offline auto al crear + relleno a los
   que no tienen + `NOT NULL` (el auto nunca pisa un premium enlazado).
7. `migrations/0007_community.sql` — `cosmetics` (+4 de ejemplo y auto-otorga),
   `user_cosmetics`, `blocks` y 4 funciones (identidad, bloquear, desbloquear, equipados).
8. `migrations/0008_presence.sql` — `profiles.is_online` (booleano interno, `false`
   por defecto) + `search_users`/`my_friends` devolviendo `is_online`.
   OJO: recrea ambas funciones con `DROP ... IF EXISTS` previo (Postgres no permite
   cambiar el tipo de retorno con `OR REPLACE`).

Pegar cada archivo completo y ejecutar. Son idempotentes: se pueden reejecutar
sin errores de "ya existe" (todo es `IF NOT EXISTS` / `OR REPLACE` / `DROP IF EXISTS`).
(Este párrafo describe el método antiguo por partes; el soportado hoy es `v1/install.sql`.)

## Qué crea cada una

| Archivo | Crea | Seguridad |
|---|---|---|
| `0001_profiles.sql` | `profiles(user_id, username, display_name, avatar_url, ...)` + índice único `lower(username)` + trigger `handle_new_user` + funciones `username_taken` / `email_for_username` | RLS: cada uno solo su fila. Funciones `SECURITY DEFINER` **revocadas** a `anon`/`authenticated` (solo `service_role`, o sea la API) |
| `0002_storage.sql` | Buckets `avatars`, `skins`, `capes` | `avatars` lectura pública; escritura siempre en carpeta propia `<user_id>/`. `skins`/`capes` privados (URLs firmadas de 1h desde la API) |
| `0003_uploads.sql` | `file_uploads(id, user_id, kind, bucket, path, size_bytes, mime, created_at)` | RLS propia; sin update (las subidas son inmutables) |
| `0004_account_extras.sql` | `profiles.bio` (≤5000) + `profiles.last_mc_version` (Minecraft Java) + `notifications(...)` + reemplazo de la política amplia de `avatars` | Avatars: sin listado global (fix del warning); notificaciones RLS propias sin INSERT de cliente |
| `0005_social.sql` | `profiles.mc_uuid` (único) + `banner_url` + `privacy_settings` + `friend_requests`/`friendships` + 5 funciones + bucket `banners` | Privacidad y amistad a nivel SQL; funciones revocadas a `anon`/`authenticated`; banners sin listado global |
| `0006_mc_uuid_auto.sql` | `offline_mc_uuid()` + backfill + trigger `fill_mc_uuid` + `NOT NULL` en `mc_uuid` | Todo perfil con UUID; el auto respeta premium enlazado |
| `0007_community.sql` | `cosmetics` + `user_cosmetics` + `blocks` + `resolve_user_identity`/`block_user`/`unblock_user`/`equipped_cosmetics` | Escritura social solo vía funciones revocadas; preview muestra solo equipados |
| `0008_presence.sql` | `profiles.is_online` (no null, `false`) + índice parcial + `search_users`/`my_friends` con `is_online` | Sin RLS nueva (cubre `profiles_update_own`); funciones revocadas a `anon`/`authenticated` |

## Verificación tras aplicar

```sql
-- Tablas y RLS activos
select tablename, rowsecurity from pg_tables
where schemaname = 'public' and tablename in ('profiles', 'file_uploads', 'notifications');
-- Columnas nuevas del perfil
select column_name from information_schema.columns
where table_schema = 'public' and table_name = 'profiles'
  and column_name in ('bio', 'last_mc_version', 'mc_uuid', 'banner_url', 'is_online');
-- Funciones sociales NO accesibles a anon/autenticados
select routine_name from information_schema.role_routine_grants
where routine_schema = 'public'
  and routine_name in ('resolve_friend_target', 'accept_friend_request', 'remove_friendship', 'search_users', 'my_friends')
  and grantee in ('anon', 'authenticated');
-- Debe devolver CERO filas.
-- Nadie sin UUID (cero tras 0006) + columna obligatoria
select count(*) as sin_uuid from public.profiles where mc_uuid is null;
select is_nullable from information_schema.columns
where table_schema = 'public' and table_name = 'profiles' and column_name = 'mc_uuid';
-- Esperado: sin_uuid = 0, is_nullable = NO.
-- Políticas de avatars SIN lectura global (vacío = warning resuelto)
select policyname from pg_policies
where schemaname = 'storage' and tablename = 'objects'
  and policyname = 'avatars_public_read';
-- Buckets
select id, public from storage.buckets where id in ('avatars', 'skins', 'capes');
-- Funciones sensibles NO accesibles a anon/autenticados
select grantee, privilege_type from information_schema.role_routine_grants
where routine_schema = 'public'
  and routine_name in ('username_taken', 'email_for_username');
-- Debe devolver filas SOLO para postgres/service_role, nunca anon/authenticated.
```

## Reglas

- Cambiar el SQL = nueva migración `0004_*.sql`, nunca editar una ya aplicada.
- Probar RLS con dos usuarios (A no lee/escribe lo de B) antes de mergear.
- Sin secretos en el SQL. Sin datos reales en el repo.
