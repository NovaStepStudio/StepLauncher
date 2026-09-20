# SQL versionado de StepLauncher API — Paquete `v1`

Todo lo que necesita la API **v1** (`src/routes/v1/`, 8 dominios) en **un solo
archivo**: `install.sql`. Pegar una vez en Supabase → SQL Editor → Run.
Después, pegar `verify.sql` para comprobar.

## Uso (una sola vez por BD, reejecutable sin riesgo)

1. Supabase → SQL Editor → New query.
2. Pegar `db/v1/install.sql` **completo** → Run.
   - Instalación nueva: crea todo.
   - BD existente (migraciones `0001–0008` ya aplicadas): converge al mismo
     estado final sin duplicar ni borrar datos (todo es `IF NOT EXISTS` /
     `OR REPLACE` / `DROP IF EXISTS` / `ON CONFLICT DO NOTHING`).
   - Transaccional: o entra todo o no entra nada.
3. Pegar `db/v1/verify.sql` → Run y comparar con los esperados de cada consulta.

## Qué contiene

| Archivo | Crea |
|---|---|
| `01_profiles.sql` | `profiles` final (bio, versión MC, `mc_uuid`, banner, `is_online`) + RLS + `username_taken`, `email_for_username`, `handle_new_user`, `offline_mc_uuid`, `fill_mc_uuid` |
| `02_storage.sql` | Buckets `avatars`/`banners` (públicos, sin listado global) + `skins`/`capes` (privados) |
| `03_uploads.sql` | `file_uploads` (con `banner` incluido desde el inicio) |
| `04_privacy_notifications.sql` | `privacy_settings` + `notifications` |
| `05_social.sql` | `friend_requests`, `friendships`, `resolve_friend_target`, `accept_friend_request`, `remove_friendship`, `search_users`, `my_friends` (finales con `is_online`) |
| `06_community.sql` | `cosmetics` (+4 de ejemplo), `user_cosmetics`, `blocks`, `resolve_user_identity`, `block_user`, `unblock_user`, `equipped_cosmetics` + auto-otorga de cosméticos futuros |
| `07_consolidate.sql` | Rellenos (privacidad, UUID, cosméticos) + `mc_uuid NOT NULL` + `ANALYZE` |
| `install.sql` | **GENERADO** (`bun run db:build`): transacción + todo lo anterior + registro en `schema_version` |
| `verify.sql` | 9 comprobaciones con su esperado |

## Mantenimiento (cómo evoluciona v1 sin romperla)

- **Editar solo las partes** `01–07`. **Jamás** `install.sql` a mano.
- Regenerar: `bun run db:build` (dentro de `api/`).
- v1 solo admite cambios **compatibles** (columna opcional nueva, índice nuevo,
  función nueva). Un cambio **incompatible** (quitar/renombrar, cambiar semántica)
  exige `db/v2/` nuevo copiando `v1/` (ver `docs/arquitectura/versionado.md`).
- `db/migrations/0001–0008.sql` queda **congelado** como historial: no se toca.
  El camino soportado es este paquete.

## Reglas

- Sin secretos en el SQL. Sin datos reales en el repo.
- Probar RLS con dos usuarios (A no lee/escribe lo de B) antes de mergear.
- Tras aplicar en producción: guardar el resultado de `verify.sql`.
