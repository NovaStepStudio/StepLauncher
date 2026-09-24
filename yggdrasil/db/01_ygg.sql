-- db/01_ygg.sql — Tablas propias del servidor Yggdrasil (tokens + joins).
-- Servicio separado `yggdrasil/`: NO toca tablas de `api/`.
-- Usuarios y perfiles: auth.users + public.profiles (fuente única, ya existe).
-- Idempotente y reejecutable: IF NOT EXISTS / OR REPLACE / ON CONFLICT.
-- Uso: pegar COMPLETO en Supabase → SQL Editor → Run (una transacción).

begin;

-- Límite de ventana suficiente sin bloquear en exceso.
set local statement_timeout = '120s';
set local lock_timeout = '30s';

-- Tokens Yggdrasil: credencial opaca del juego (no es JWT de Supabase).
-- Un usuario puede tener varios; el Worker aplica máx. 10 vivos (revoca viejos).
create table if not exists public.ygg_tokens (
  access_token text primary key,
  client_token text not null,
  user_id uuid not null references auth.users (id) on delete cascade,
  profile_uuid uuid,
  profile_name text,
  issued_at timestamptz not null default now(),
  expires_at timestamptz not null default (now() + interval '15 days'),
  revoked boolean not null default false
);

create index if not exists ygg_tokens_user_live_idx
  on public.ygg_tokens (user_id, issued_at)
  where (revoked = false);

create index if not exists ygg_tokens_expires_idx
  on public.ygg_tokens (expires_at);

comment on table public.ygg_tokens is
  'Tokens Yggdrasil del juego (accessToken opaco + clientToken + perfil atado). Solo los toca el Worker con service_role.';
comment on column public.ygg_tokens.access_token is
  'Opaco de 32 hex (128 bits). Clave primaria por su aleatoriedad.';
comment on column public.ygg_tokens.profile_uuid is
  'UUID-MC con guiones (copia de profiles.mc_uuid al emitir; no FK para no acoplar el borrado de perfil).';

-- Sesiones de entrada: el cliente anuncia serverId (join) y el servidor lo
-- verifica (hasJoined). Viven 30 s; la aleatoriedad de serverId es la PK.
create table if not exists public.ygg_joins (
  server_id text primary key,
  access_token text not null,
  profile_uuid uuid,
  profile_name text,
  client_ip text not null default 'unknown',
  created_at timestamptz not null default now(),
  expires_at timestamptz not null default (now() + interval '30 seconds')
);

create index if not exists ygg_joins_expires_idx
  on public.ygg_joins (expires_at);

comment on table public.ygg_joins is
  'Anuncios join del cliente para verificación hasJoined (30 s de vida). Solo los toca el Worker con service_role.';

-- Auditoría de verificación: qué joins llegaron a usarse y desde qué servidor.
-- Solo el Worker escribe estas columnas al resolver `hasJoined` (ver `lib/store.ts`).
-- Nota: la tabla solo se llena si el servidor hace handshake online (cifrado):
-- en servidores offline o con auth por plugins el cliente nunca anuncia `join`
-- y es normal verla vacía.
alter table public.ygg_joins
  add column if not exists verified_at timestamptz;
alter table public.ygg_joins
  add column if not exists server_ip text;

comment on column public.ygg_joins.verified_at is
  'Cuándo un servidor verificó esta sesión con hasJoined (null = anunciada pero nunca verificada).';
comment on column public.ygg_joins.server_ip is
  'IP del servidor Minecraft que verificó la sesión (sirve para ver dónde juegan tus usuarios).';

-- Mapa hash → textura para `GET /textures/:hash` (caché de escritura).
-- El hash es el SHA-256 del PNG: el cliente lo usa como clave de caché y la
-- URL es estable. Se rellena solo al servir perfiles (write-through); las
-- filas viejas de skins borradas se pueden purgar por `created_at`.
create table if not exists public.ygg_textures (
  hash text primary key check (hash ~ '^[0-9a-f]{64}$'),
  bucket text not null check (bucket in ('skins', 'capes')),
  path text not null,
  kind text not null check (kind in ('skin', 'cape')),
  user_id uuid not null references auth.users (id) on delete cascade,
  created_at timestamptz not null default now()
);

create index if not exists ygg_textures_owner_idx
  on public.ygg_textures (user_id, created_at desc);

comment on table public.ygg_textures is
  'Caché hash → PNG en storage para servir texturas del juego. Solo la escribe/lee el Worker con service_role.';

-- Sin RLS de cliente a propósito: el juego nunca habla SQL directo.
-- Se deja RLS desactivada (tablas internas del Worker) para no dar falsa
-- sensación de acceso por usuario. El acceso real lo limita service_role.
alter table public.ygg_tokens disable row level security;
alter table public.ygg_joins disable row level security;
alter table public.ygg_textures disable row level security;

-- Trazabilidad: misma tabla que usa `api/` (marca propia, sin romper v1).
create table if not exists public.schema_version (
  version text primary key,
  applied_at timestamptz not null default now()
);

insert into public.schema_version (version) values ('ygg-v1')
on conflict (version) do nothing;

commit;

-- Verificación (pegar aparte tras el Run y comparar):
-- select * from public.schema_version where version = 'ygg-v1';
-- select count(*) from public.ygg_tokens;
-- select count(*) from public.ygg_joins;
-- select count(*) from public.ygg_textures;
