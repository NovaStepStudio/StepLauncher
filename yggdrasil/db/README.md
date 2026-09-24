# SQL propio de Yggdrasil — tablas de juego

Solo lo que `api/` **no** tiene: tokens Yggdrasil y sesiones de entrada.
Los usuarios y perfiles viven en `auth.users` + `public.profiles` (fuente de
`api/`): aquí **no** se crean ni se migran.

## Uso (una vez por BD, reejecutable sin riesgo)

1. Supabase → SQL Editor → New query.
2. Pegar `db/01_ygg.sql` **completo** → Run (transaccional, idempotente).
3. Comprobar con `select * from public.schema_version where version = 'ygg-v1';`
   y con las consultas de `verify` al final del archivo (comentadas).

## Qué contiene

| Objeto | Para qué |
|---|---|
| `ygg_tokens` | `accessToken`/`clientToken` + usuario + perfil + vigencia 15 días |
| `ygg_joins` | `serverId` → sesión anunciada por el cliente (30 s de vida) + auditoría (`verified_at`, `server_ip`) |
| `ygg_textures` | `hash` SHA-256 → PNG en storage (caché para `GET /textures/:hash`) |
| Índices | Limpieza por expiración y límite de 10 por usuario |
| `schema_version` | Marca `ygg-v1` (misma tabla que usa `api/`, sin romperla) |

## Reglas

- Sin secretos en el SQL. Sin RLS de cliente: estas tablas solo las toca el
  Worker con `service_role` (el juego nunca habla SQL directo).
- Limpieza: los `join` caducan por `expires_at` (el Worker los ignora); un
  cron o `delete` ocasional puede borrar expirados/revocados viejos.
- Si el protocolo pide un cambio incompatible, crear `db/02_*.sql` nuevo en
  vez de reescribir este archivo con otra semántica.
