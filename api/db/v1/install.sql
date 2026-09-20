-- ============================================================================
-- StepLauncher API — Paquete v1 (instalador único, GENERADO, no editar a mano).
-- Generado con: bun run db:build  (fuente: db/v1/01_*.sql … 07_*.sql)
--
-- Uso: pegar este archivo COMPLETO en Supabase → SQL Editor → Run.
-- Sirve para instalación nueva Y para BD existentes (todo es idempotente:
-- IF NOT EXISTS / OR REPLACE / DROP IF EXISTS / ON CONFLICT DO NOTHING).
-- Se aplica en UNA transacción: o entra todo o no entra nada.
-- Después, ejecutar db/v1/verify.sql para comprobar el resultado.
-- v2 será db/v2/ con su propio install.sql (v1 nunca se reescribe a mano).
-- ============================================================================

begin;

-- Ventana suficiente para instalaciones grandes sin bloquear en exceso.
set local statement_timeout = '120s';
set local lock_timeout = '30s';

-- gen_random_uuid() de file_uploads/notifications/cosmetics lo exige.
create extension if not exists "pgcrypto";

-- Registro de versión aplicada (trazabilidad: qué paquete vio esta BD).
create table if not exists public.schema_version (
  version text primary key,
  applied_at timestamptz not null default now()
);

-- ----------------------------------------------------------------------------
-- 01_profiles.sql
-- ----------------------------------------------------------------------------
-- db/v1/01_profiles.sql — Perfil de StepLauncher + RLS + funciones base.
-- Parte 1 de 7 del paquete v1. NO ejecutar suelta: usar db/v1/install.sql
-- (un solo pegar-y-ejecutar transaccional). Idempotente y reejecutable.
-- La identidad (email + contraseña) vive en auth.users; aquí solo el perfil.

-- Tabla base (instalación nueva). En BD existentes no hace nada.
create table if not exists public.profiles (
  user_id uuid primary key references auth.users (id) on delete cascade,
  username text not null,
  display_name text not null,
  avatar_url text,
  bio text check (bio is null or char_length(bio) <= 5000),
  last_mc_version text,
  mc_uuid uuid,
  banner_url text,
  is_online boolean not null default false,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

-- Convergencia para BD que ya existen con la tabla antigua: cada columna
-- nueva se añade solo si falta. El NOT NULL de mc_uuid se aplica en
-- 07_consolidate.sql, DESPUÉS del relleno (si se hiciera aquí fallaría
-- con "ya existe" o dejaría nulos en carreras de registro).
alter table public.profiles
  add column if not exists bio text
  check (bio is null or char_length(bio) <= 5000);

alter table public.profiles
  add column if not exists last_mc_version text;

alter table public.profiles
  add column if not exists mc_uuid uuid;

alter table public.profiles
  add column if not exists banner_url text;

alter table public.profiles
  add column if not exists is_online boolean not null default false;

comment on column public.profiles.bio is
  'Descripción pública del jugador (máx 5000 caracteres).';
comment on column public.profiles.last_mc_version is
  'Última versión de Minecraft Java jugada (ej: 1.21). La actualiza el launcher al jugar.';
comment on column public.profiles.mc_uuid is
  'UUID de Minecraft: auto-generado (offline) al crear la cuenta; el jugador puede enlazar su premium después. Nunca null tras 07_consolidate.';
comment on column public.profiles.is_online is
  'Estado en línea del jugador (true = en línea). Lo cambia el dueño desde la API; visible en su perfil, búsquedas, amigos y preview.';

create unique index if not exists profiles_username_lower_uq
  on public.profiles (lower(username));
create unique index if not exists profiles_mc_uuid_uq
  on public.profiles (mc_uuid);
create index if not exists profiles_is_online_idx
  on public.profiles (is_online) where (is_online = true);

alter table public.profiles enable row level security;

-- Cada jugador solo ve y toca SU fila. Sin política de delete (el borrado
-- de cuenta es un flujo aparte, sensible y futuro).
drop policy if exists "profiles_select_own" on public.profiles;
create policy "profiles_select_own" on public.profiles
  for select using (auth.uid() = user_id);

drop policy if exists "profiles_insert_own" on public.profiles;
create policy "profiles_insert_own" on public.profiles
  for insert with check (auth.uid() = user_id);

drop policy if exists "profiles_update_own" on public.profiles;
create policy "profiles_update_own" on public.profiles
  for update using (auth.uid() = user_id) with check (auth.uid() = user_id);

-- updated_at automático en cada cambio.
create or replace function public.set_updated_at()
returns trigger language plpgsql set search_path = public as $$
begin
  new.updated_at = now();
  return new;
end $$;

drop trigger if exists profiles_set_updated_at on public.profiles;
create trigger profiles_set_updated_at before update on public.profiles
  for each row execute function public.set_updated_at();

-- ¿Username ocupado? La llama SOLO la API con service_role (login/registro).
create or replace function public.username_taken(p_username text)
returns boolean language sql stable security definer set search_path = public as $$
  select exists (select 1 from public.profiles where lower(username) = lower(p_username));
$$;
revoke all on function public.username_taken(text) from public, anon, authenticated;

-- Resolver email para login con username. SENSIBLE: revocada a todo salvo
-- service_role, porque expone qué email corresponde a cada usuario.
create or replace function public.email_for_username(p_username text)
returns text language sql stable security definer set search_path = public, auth as $$
  select u.email from auth.users u
  join public.profiles p on p.user_id = u.id
  where lower(p.username) = lower(p_username) limit 1;
$$;
revoke all on function public.email_for_username(text) from public, anon, authenticated;

-- Auto-crear perfil al registrarse: el username llega en user_metadata.
-- Si choca con otro (carrera entre dos registros), aborta con 'username_taken'
-- y la API lo traduce a 409. Si no trae username, genera uno temporal único.
-- El mc_uuid lo rellena el trigger profiles_fill_mc_uuid (más abajo).
create or replace function public.handle_new_user()
returns trigger language plpgsql security definer set search_path = public as $$
declare
  v_username text := nullif(trim(coalesce(new.raw_user_meta_data ->> 'username', '')), '');
begin
  if v_username is not null
     and exists (select 1 from public.profiles where lower(username) = lower(v_username)) then
    raise exception 'username_taken';
  end if;
  if v_username is null then
    v_username := 'jugador-' || substr(new.id::text, 1, 8);
  end if;
  insert into public.profiles (user_id, username, display_name)
  values (new.id, v_username, v_username)
  on conflict (user_id) do nothing;
  return new;
end $$;

drop trigger if exists on_auth_user_created on auth.users;
create trigger on_auth_user_created after insert on auth.users
  for each row execute function public.handle_new_user();

-- UUID v3 offline de Minecraft a partir del username (inmutable, determinista:
-- MD5 de "OfflinePlayer:<username>" con versión 3 y variante RFC 4122).
-- Declarada AQUÍ (antes de cualquier INSERT) para que ningún registro
-- concurrente pueda colarse con mc_uuid en null entre el relleno y el trigger.
create or replace function public.offline_mc_uuid(p_username text)
returns uuid language plpgsql immutable set search_path = public as $$
declare
  h text := md5('OfflinePlayer:' || p_username);
  v integer;
begin
  h := overlay(h placing '3' from 13 for 1);
  v := (get_byte(decode(substr(h, 17, 2), 'hex'), 0) & 63) | 128;
  h := overlay(h placing lpad(to_hex(v), 2, '0') from 17 for 2);
  return (
    substr(h, 1, 8) || '-' || substr(h, 9, 4) || '-' ||
    substr(h, 13, 4) || '-' || substr(h, 17, 4) || '-' || substr(h, 21, 12)
  )::uuid;
end $$;

-- Auto-relleno al crear (INSERT) y si se intenta dejar en null (UPDATE).
-- Nunca pisa un premium ya enlazado: solo actúa cuando el nuevo valor es null.
create or replace function public.fill_mc_uuid()
returns trigger language plpgsql set search_path = public as $$
begin
  if new.mc_uuid is null then
    new.mc_uuid := public.offline_mc_uuid(new.username);
  end if;
  return new;
end $$;

drop trigger if exists profiles_fill_mc_uuid on public.profiles;
create trigger profiles_fill_mc_uuid before insert or update on public.profiles
  for each row execute function public.fill_mc_uuid();

-- ----------------------------------------------------------------------------
-- 02_storage.sql
-- ----------------------------------------------------------------------------
-- db/v1/02_storage.sql — Buckets y políticas de Storage (estado final v1).
-- Parte 2 de 7 del paquete v1. NO ejecutar suelta: usar db/v1/install.sql.
-- Rutas obligatorias: "<user_id>/..." (primera carpeta = dueño, ver API).
--
-- Estado final: 4 buckets (avatars y banners públicos; skins y capes privados)
-- SIN ninguna política SELECT amplia. Lección aprendida: los buckets públicos
-- NO necesitan política SELECT (la descarga por URL pública funciona igual);
-- una política `bucket_id = 'X'` permite LISTAR todo el bucket (aviso Supabase).
-- Aquí solo hay lectura restringida al dueño: sin listado global.

insert into storage.buckets (id, name, public)
values
  ('avatars', 'avatars', true),
  ('skins', 'skins', false),
  ('capes', 'capes', false),
  ('banners', 'banners', true)
on conflict (id) do nothing;

-- Limpieza de la política amplia heredada (migración 0002): si existe, fuera.
drop policy if exists "avatars_public_read" on storage.objects;

-- AVATARS: lectura/inserción/cambio/borrado solo en carpeta propia.
drop policy if exists "avatars_owner_read" on storage.objects;
create policy "avatars_owner_read" on storage.objects
  for select using (
    bucket_id = 'avatars'
    and (storage.foldername(name))[1] = auth.uid()::text
  );

drop policy if exists "avatars_owner_insert" on storage.objects;
create policy "avatars_owner_insert" on storage.objects
  for insert with check (
    bucket_id = 'avatars'
    and (storage.foldername(name))[1] = auth.uid()::text
  );

drop policy if exists "avatars_owner_update" on storage.objects;
create policy "avatars_owner_update" on storage.objects
  for update using (
    bucket_id = 'avatars'
    and (storage.foldername(name))[1] = auth.uid()::text
  ) with check (
    bucket_id = 'avatars'
    and (storage.foldername(name))[1] = auth.uid()::text
  );

drop policy if exists "avatars_owner_delete" on storage.objects;
create policy "avatars_owner_delete" on storage.objects
  for delete using (
    bucket_id = 'avatars'
    and (storage.foldername(name))[1] = auth.uid()::text
  );

-- SKINS y CAPAS: privados; el dueño accede y la API genera URLs firmadas (1h).
drop policy if exists "skins_owner_all" on storage.objects;
create policy "skins_owner_all" on storage.objects
  for all using (
    bucket_id = 'skins'
    and (storage.foldername(name))[1] = auth.uid()::text
  ) with check (
    bucket_id = 'skins'
    and (storage.foldername(name))[1] = auth.uid()::text
  );

drop policy if exists "capes_owner_all" on storage.objects;
create policy "capes_owner_all" on storage.objects
  for all using (
    bucket_id = 'capes'
    and (storage.foldername(name))[1] = auth.uid()::text
  ) with check (
    bucket_id = 'capes'
    and (storage.foldername(name))[1] = auth.uid()::text
  );

-- BANNERS: mismo patrón que avatars (público sin listado global).
drop policy if exists "banners_owner_read" on storage.objects;
create policy "banners_owner_read" on storage.objects
  for select using (
    bucket_id = 'banners'
    and (storage.foldername(name))[1] = auth.uid()::text
  );

drop policy if exists "banners_owner_insert" on storage.objects;
create policy "banners_owner_insert" on storage.objects
  for insert with check (
    bucket_id = 'banners'
    and (storage.foldername(name))[1] = auth.uid()::text
  );

drop policy if exists "banners_owner_update" on storage.objects;
create policy "banners_owner_update" on storage.objects
  for update using (
    bucket_id = 'banners'
    and (storage.foldername(name))[1] = auth.uid()::text
  ) with check (
    bucket_id = 'banners'
    and (storage.foldername(name))[1] = auth.uid()::text
  );

drop policy if exists "banners_owner_delete" on storage.objects;
create policy "banners_owner_delete" on storage.objects
  for delete using (
    bucket_id = 'banners'
    and (storage.foldername(name))[1] = auth.uid()::text
  );

-- ----------------------------------------------------------------------------
-- 03_uploads.sql
-- ----------------------------------------------------------------------------
-- db/v1/03_uploads.sql — Libro de subidas para la quota diaria.
-- Parte 3 de 7 del paquete v1. NO ejecutar suelta: usar db/v1/install.sql.
-- La quota (10 skins+capas/día) se calcula en la API contando filas propias
-- de las últimas 24h (ver src/middleware/upload-quota.ts). El libro también
-- registra avatares y banners con fin de auditoría (no cuentan para la quota).

create table if not exists public.file_uploads (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references auth.users (id) on delete cascade,
  kind text not null check (kind in ('skin', 'cape', 'avatar', 'banner')),
  bucket text not null,
  path text not null,
  size_bytes integer not null check (size_bytes > 0),
  mime text not null,
  created_at timestamptz not null default now()
);

-- Convergencia: en BD creadas con la migración 0003 el CHECK solo admitía
-- ('skin','cape','avatar'). Se sustituye por el definitivo (con 'banner').
alter table public.file_uploads drop constraint if exists file_uploads_kind_check;
alter table public.file_uploads
  add constraint file_uploads_kind_check
  check (kind in ('skin', 'cape', 'avatar', 'banner'));

create index if not exists file_uploads_owner_time_idx
  on public.file_uploads (user_id, created_at desc);
create index if not exists file_uploads_owner_kind_time_idx
  on public.file_uploads (user_id, kind, created_at desc);

alter table public.file_uploads enable row level security;

-- El jugador lista e inserta solo lo suyo; puede borrar lo suyo (libera quota).
-- Sin update: una subida es inmutable (se sube otro archivo en su lugar).
drop policy if exists "uploads_select_own" on public.file_uploads;
create policy "uploads_select_own" on public.file_uploads
  for select using (auth.uid() = user_id);

drop policy if exists "uploads_insert_own" on public.file_uploads;
create policy "uploads_insert_own" on public.file_uploads
  for insert with check (auth.uid() = user_id);

drop policy if exists "uploads_delete_own" on public.file_uploads;
create policy "uploads_delete_own" on public.file_uploads
  for delete using (auth.uid() = user_id);

-- ----------------------------------------------------------------------------
-- 04_privacy_notifications.sql
-- ----------------------------------------------------------------------------
-- db/v1/04_privacy_notifications.sql — Privacidad y notificaciones de la cuenta.
-- Parte 4 de 7 del paquete v1. NO ejecutar suelta: usar db/v1/install.sql.
-- El relleno de filas por defecto para cuentas existentes vive en
-- 07_consolidate.sql (orden garantizado: primero existe la tabla y el trigger).

-- 1) Privacidad de la cuenta (una fila por usuario, todo true por defecto).
create table if not exists public.privacy_settings (
  user_id uuid primary key references auth.users (id) on delete cascade,
  searchable boolean not null default true,
  allow_email_search boolean not null default true,
  receive_friend_requests boolean not null default true,
  receive_notifications boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

alter table public.privacy_settings enable row level security;

drop policy if exists "privacy_select_own" on public.privacy_settings;
create policy "privacy_select_own" on public.privacy_settings
  for select using (auth.uid() = user_id);

drop policy if exists "privacy_insert_own" on public.privacy_settings;
create policy "privacy_insert_own" on public.privacy_settings
  for insert with check (auth.uid() = user_id);

drop policy if exists "privacy_update_own" on public.privacy_settings;
create policy "privacy_update_own" on public.privacy_settings
  for update using (auth.uid() = user_id) with check (auth.uid() = user_id);

drop trigger if exists privacy_set_updated_at on public.privacy_settings;
create trigger privacy_set_updated_at before update on public.privacy_settings
  for each row execute function public.set_updated_at();

-- Fila por defecto al nacer el perfil (el trigger de perfiles ya existe:
-- 01_profiles.sql se aplica antes dentro de install.sql).
create or replace function public.handle_new_profile()
returns trigger language plpgsql security definer set search_path = public as $$
begin
  insert into public.privacy_settings (user_id) values (new.user_id)
  on conflict (user_id) do nothing;
  return new;
end $$;

drop trigger if exists on_profile_created on public.profiles;
create trigger on_profile_created after insert on public.profiles
  for each row execute function public.handle_new_profile();

-- 2) Notificaciones de la cuenta. Las crea el servidor (service_role);
-- el jugador solo las lee, las marca como leídas o las descarta.
-- Sin INSERT de cliente: si el cliente pudiera insertar, podría llenar
-- su propia bandeja y falsear avisos del sistema.
create table if not exists public.notifications (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references auth.users (id) on delete cascade,
  title text not null check (char_length(title) between 1 and 120),
  body text not null check (char_length(body) between 1 and 1000),
  read boolean not null default false,
  created_at timestamptz not null default now()
);

create index if not exists notifications_owner_time_idx
  on public.notifications (user_id, created_at desc);
create index if not exists notifications_owner_unread_idx
  on public.notifications (user_id, created_at desc) where (read = false);

alter table public.notifications enable row level security;

drop policy if exists "notifications_select_own" on public.notifications;
create policy "notifications_select_own" on public.notifications
  for select using (auth.uid() = user_id);

drop policy if exists "notifications_update_own" on public.notifications;
create policy "notifications_update_own" on public.notifications
  for update using (auth.uid() = user_id) with check (auth.uid() = user_id);

drop policy if exists "notifications_delete_own" on public.notifications;
create policy "notifications_delete_own" on public.notifications
  for delete using (auth.uid() = user_id);

-- ----------------------------------------------------------------------------
-- 05_social.sql
-- ----------------------------------------------------------------------------
-- db/v1/05_social.sql — Amistades v1: solicitudes, lista y búsqueda.
-- Parte 5 de 7 del paquete v1. NO ejecutar suelta: usar db/v1/install.sql.
-- La escritura (solicitar/aceptar/bloquear) vive SOLO en funciones
-- SECURITY DEFINER revocadas: ningún cliente toca estas tablas directamente.
-- `search_users` y `my_friends` se crean directamente en su forma FINAL
-- (con is_online). En BD antiguas se borra la firma previa primero porque
-- Postgres no permite cambiar el tipo de retorno con OR REPLACE.

-- 1) Solicitudes de amistad.
create table if not exists public.friend_requests (
  id uuid primary key default gen_random_uuid(),
  from_user_id uuid not null references auth.users (id) on delete cascade,
  to_user_id uuid not null references auth.users (id) on delete cascade,
  status text not null default 'pending'
    check (status in ('pending', 'accepted', 'declined', 'cancelled')),
  created_at timestamptz not null default now(),
  responded_at timestamptz,
  check (from_user_id <> to_user_id)
);

create index if not exists friend_requests_to_pending_idx
  on public.friend_requests (to_user_id, created_at desc) where (status = 'pending');
create index if not exists friend_requests_from_idx
  on public.friend_requests (from_user_id, created_at desc);
create index if not exists friend_requests_pair_idx
  on public.friend_requests (from_user_id, to_user_id, status);

alter table public.friend_requests enable row level security;

drop policy if exists "requests_select_involved" on public.friend_requests;
create policy "requests_select_involved" on public.friend_requests
  for select using (auth.uid() = from_user_id or auth.uid() = to_user_id);

-- 2) Amistades (doble fila por amistad; la crea aceptar, la borra romper).
create table if not exists public.friendships (
  user_id uuid not null references auth.users (id) on delete cascade,
  friend_id uuid not null references auth.users (id) on delete cascade,
  created_at timestamptz not null default now(),
  primary key (user_id, friend_id),
  check (user_id <> friend_id)
);

create index if not exists friendships_user_idx on public.friendships (user_id);

alter table public.friendships enable row level security;

drop policy if exists "friendships_select_own" on public.friendships;
create policy "friendships_select_own" on public.friendships
  for select using (auth.uid() = user_id);

-- 3) Resolver destinatario respetando privacidad. Devuelve NULL si está
-- bloqueado o no existe (genérico a propósito, sin filtrar el motivo).
--  - UUID (user_id o mc_uuid): exige que acepte solicitudes.
--  - email: exige allow_email_search + que acepte solicitudes.
--  - username: exige searchable + que acepte solicitudes.
create or replace function public.resolve_friend_target(p_identifier text)
returns uuid language plpgsql stable security definer
set search_path = public, auth as $$
declare
  v_id uuid;
  v_trim text := trim(p_identifier);
begin
  if v_trim ~ '^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$' then
    select p.user_id into v_id from public.profiles p
      left join public.privacy_settings s on s.user_id = p.user_id
      where (p.user_id = v_trim::uuid or p.mc_uuid = v_trim::uuid)
        and coalesce(s.receive_friend_requests, true) = true
      limit 1;
    return v_id;
  end if;
  if position('@' in v_trim) > 0 then
    select u.id into v_id from auth.users u
      left join public.privacy_settings s on s.user_id = u.id
      where lower(u.email) = lower(v_trim)
        and coalesce(s.allow_email_search, true) = true
        and coalesce(s.receive_friend_requests, true) = true;
    return v_id;
  end if;
  select p.user_id into v_id from public.profiles p
    left join public.privacy_settings s on s.user_id = p.user_id
    where lower(p.username) = lower(v_trim)
      and coalesce(s.searchable, true) = true
      and coalesce(s.receive_friend_requests, true) = true;
  return v_id;
end $$;
revoke all on function public.resolve_friend_target(text) from public, anon, authenticated;

-- 4) Aceptar solicitud (atómico: valida receptor + crea doble amistad).
create or replace function public.accept_friend_request(p_request_id uuid, p_user_id uuid)
returns boolean language plpgsql security definer set search_path = public as $$
declare
  v_from uuid;
begin
  select from_user_id into v_from from public.friend_requests
    where id = p_request_id and to_user_id = p_user_id and status = 'pending';
  if v_from is null then
    return false;
  end if;
  update public.friend_requests
    set status = 'accepted', responded_at = now()
    where id = p_request_id;
  insert into public.friendships (user_id, friend_id)
  values (p_user_id, v_from), (v_from, p_user_id)
  on conflict do nothing;
  return true;
end $$;
revoke all on function public.accept_friend_request(uuid, uuid) from public, anon, authenticated;

-- 5) Romper amistad (cualquiera de los dos lados; borra ambas filas).
create or replace function public.remove_friendship(p_user_id uuid, p_friend_id uuid)
returns boolean language plpgsql security definer set search_path = public as $$
declare
  v_n integer;
begin
  delete from public.friendships
    where (user_id = p_user_id and friend_id = p_friend_id)
       or (user_id = p_friend_id and friend_id = p_user_id);
  get diagnostics v_n = row_count;
  return v_n > 0;
end $$;
revoke all on function public.remove_friendship(uuid, uuid) from public, anon, authenticated;

-- 6) Buscar jugadores (solo `searchable`; columnas públicas mínimas + presencia).
-- DROP previo: en BD antiguas existe la firma sin is_online y OR REPLACE fallaría.
drop function if exists public.search_users(text, integer);

create or replace function public.search_users(p_query text, p_limit integer default 20)
returns table (
  user_id uuid, username text, display_name text, avatar_url text, mc_uuid uuid,
  is_online boolean
)
language plpgsql stable security definer set search_path = public as $$
declare
  v_q text := replace(replace(replace(trim(p_query), '\', '\\'), '%', '\%'), '_', '\_');
  v_limit integer := least(greatest(coalesce(p_limit, 20), 1), 50);
begin
  return query
  select p.user_id, p.username, p.display_name, p.avatar_url, p.mc_uuid, p.is_online
    from public.profiles p
    join public.privacy_settings s on s.user_id = p.user_id
    where s.searchable = true
      and (p.username ilike '%' || v_q || '%' escape '\'
        or p.display_name ilike '%' || v_q || '%' escape '\')
    order by p.username
    limit v_limit;
end $$;
revoke all on function public.search_users(text, integer) from public, anon, authenticated;

-- 7) Lista de amigos propios (perfil mínimo + desde cuándo + presencia).
-- Mismo caso que search_users: DROP previo obligatorio por cambio de retorno.
drop function if exists public.my_friends(uuid);

create or replace function public.my_friends(p_user_id uuid)
returns table (
  friend_id uuid, username text, display_name text, avatar_url text,
  mc_uuid uuid, friends_since timestamptz, is_online boolean
)
language sql stable security definer set search_path = public as $$
  select f.friend_id, p.username, p.display_name, p.avatar_url, p.mc_uuid, f.created_at, p.is_online
    from public.friendships f
    join public.profiles p on p.user_id = f.friend_id
    where f.user_id = p_user_id
    order by f.created_at desc;
$$;
revoke all on function public.my_friends(uuid) from public, anon, authenticated;

-- ----------------------------------------------------------------------------
-- 06_community.sql
-- ----------------------------------------------------------------------------
-- db/v1/06_community.sql — Comunidad v1: cosméticos y lista negra.
-- Parte 6 de 7 del paquete v1. NO ejecutar suelta: usar db/v1/install.sql.
-- El relleno para cuentas anteriores al catálogo vive en 07_consolidate.sql.
-- Novedad de estabilidad v1: el catálogo también se otorga cuando se crea
-- un cosmético NUEVO (trigger on_cosmetic_grant_all), no solo al nacer el
-- perfil. Sin esto, añadir un 5º cosmético dejaría a los usuarios viejos sin él.

-- 1) Catálogo de cosméticos (lo gestiona el servidor; lectura autenticada).
create table if not exists public.cosmetics (
  id uuid primary key default gen_random_uuid(),
  slug text not null unique,
  name text not null check (char_length(name) between 1 and 80),
  kind text not null check (kind in ('cape', 'hat', 'pet', 'effect', 'other')),
  image_url text,
  created_at timestamptz not null default now()
);

alter table public.cosmetics enable row level security;

drop policy if exists "cosmetics_select_all" on public.cosmetics;
create policy "cosmetics_select_all" on public.cosmetics
  for select using (auth.role() = 'authenticated');

-- Contenido inicial de ejemplo (revisable; no se duplica al reejecutar).
insert into public.cosmetics (slug, name, kind)
values
  ('capa-fundador', 'Capa del Fundador', 'cape'),
  ('gorro-aventurero', 'Gorro de Aventurero', 'hat'),
  ('lobo-companero', 'Lobo Compañero', 'pet'),
  ('aura-ender', 'Aura Ender', 'effect')
on conflict (slug) do nothing;

-- 2) Cosméticos por jugador (solo se muestra lo EQUIPADO en la preview).
create table if not exists public.user_cosmetics (
  user_id uuid not null references auth.users (id) on delete cascade,
  cosmetic_id uuid not null references public.cosmetics (id) on delete cascade,
  equipped boolean not null default false,
  acquired_at timestamptz not null default now(),
  primary key (user_id, cosmetic_id)
);

create index if not exists user_cosmetics_equipped_idx
  on public.user_cosmetics (user_id) where (equipped = true);

alter table public.user_cosmetics enable row level security;

drop policy if exists "user_cosmetics_select_own" on public.user_cosmetics;
create policy "user_cosmetics_select_own" on public.user_cosmetics
  for select using (auth.uid() = user_id);

drop policy if exists "user_cosmetics_update_own" on public.user_cosmetics;
create policy "user_cosmetics_update_own" on public.user_cosmetics
  for update using (auth.uid() = user_id) with check (auth.uid() = user_id);
-- Sin INSERT/DELETE de cliente: los otorga el servidor (trigger/admin).

-- Al nacer el perfil se otorga el catálogo (equipado = false).
create or replace function public.grant_catalog_cosmetics()
returns trigger language plpgsql security definer set search_path = public as $$
begin
  insert into public.user_cosmetics (user_id, cosmetic_id)
  select new.user_id, c.id from public.cosmetics c
  on conflict do nothing;
  return new;
end $$;

drop trigger if exists on_profile_grant_cosmetics on public.profiles;
create trigger on_profile_grant_cosmetics after insert on public.profiles
  for each row execute function public.grant_catalog_cosmetics();

-- Al crear un cosmético nuevo se otorga a TODOS los perfiles existentes.
-- Cierra el hueco del diseño anterior (solo cubría perfiles futuros).
create or replace function public.grant_new_cosmetic_to_all()
returns trigger language plpgsql security definer set search_path = public as $$
begin
  insert into public.user_cosmetics (user_id, cosmetic_id)
  select p.user_id, new.id from public.profiles p
  on conflict do nothing;
  return new;
end $$;

drop trigger if exists on_cosmetic_grant_all on public.cosmetics;
create trigger on_cosmetic_grant_all after insert on public.cosmetics
  for each row execute function public.grant_new_cosmetic_to_all();

-- 3) Lista negra (bloqueos). Espejo: bloqueado en algún sentido = sin preview.
create table if not exists public.blocks (
  blocker_id uuid not null references auth.users (id) on delete cascade,
  blocked_id uuid not null references auth.users (id) on delete cascade,
  created_at timestamptz not null default now(),
  primary key (blocker_id, blocked_id),
  check (blocker_id <> blocked_id)
);

create index if not exists blocks_blocker_idx on public.blocks (blocker_id);
create index if not exists blocks_blocked_idx on public.blocks (blocked_id);

alter table public.blocks enable row level security;

drop policy if exists "blocks_select_own" on public.blocks;
create policy "blocks_select_own" on public.blocks
  for select using (auth.uid() = blocker_id);
-- Sin escritura directa: solo vía API/funciones.

-- 4) Resolver identidad SIN filtro de privacidad (para preview y bloqueos;
-- la privacidad/bloqueo se evalúa después en la API, con mensajes genéricos).
create or replace function public.resolve_user_identity(p_identifier text)
returns uuid language plpgsql stable security definer
set search_path = public, auth as $$
declare
  v_id uuid;
  v_trim text := trim(p_identifier);
begin
  if v_trim ~ '^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$' then
    select p.user_id into v_id from public.profiles p
      where p.user_id = v_trim::uuid or p.mc_uuid = v_trim::uuid
      limit 1;
    return v_id;
  end if;
  if position('@' in v_trim) > 0 then
    select u.id into v_id from auth.users u
      where lower(u.email) = lower(v_trim);
    return v_id;
  end if;
  select p.user_id into v_id from public.profiles p
    where lower(p.username) = lower(v_trim);
  return v_id;
end $$;
revoke all on function public.resolve_user_identity(text) from public, anon, authenticated;

-- 5) Bloquear: registra + corta relación (pendientes cancelados, amistad borrada).
create or replace function public.block_user(p_user_id uuid, p_target_id uuid)
returns boolean language plpgsql security definer set search_path = public as $$
begin
  if p_user_id = p_target_id then
    return false;
  end if;
  if not exists (select 1 from public.profiles where user_id = p_target_id) then
    return false;
  end if;
  insert into public.blocks (blocker_id, blocked_id)
  values (p_user_id, p_target_id)
  on conflict do nothing;
  update public.friend_requests
    set status = 'cancelled', responded_at = now()
    where status = 'pending'
      and ((from_user_id = p_user_id and to_user_id = p_target_id)
        or (from_user_id = p_target_id and to_user_id = p_user_id));
  delete from public.friendships
    where (user_id = p_user_id and friend_id = p_target_id)
       or (user_id = p_target_id and friend_id = p_user_id);
  return true;
end $$;
revoke all on function public.block_user(uuid, uuid) from public, anon, authenticated;

-- 6) Desbloquear.
create or replace function public.unblock_user(p_user_id uuid, p_target_id uuid)
returns boolean language plpgsql security definer set search_path = public as $$
declare
  v_n integer;
begin
  delete from public.blocks
    where blocker_id = p_user_id and blocked_id = p_target_id;
  get diagnostics v_n = row_count;
  return v_n > 0;
end $$;
revoke all on function public.unblock_user(uuid, uuid) from public, anon, authenticated;

-- 7) Cosméticos EQUIPADOS de un jugador (solo eso muestra la preview).
create or replace function public.equipped_cosmetics(p_user_id uuid)
returns table (id uuid, slug text, name text, kind text, image_url text)
language sql stable security definer set search_path = public as $$
  select c.id, c.slug, c.name, c.kind, c.image_url
    from public.user_cosmetics uc
    join public.cosmetics c on c.id = uc.cosmetic_id
    where uc.user_id = p_user_id and uc.equipped = true
    order by c.name;
$$;
revoke all on function public.equipped_cosmetics(uuid) from public, anon, authenticated;

-- ----------------------------------------------------------------------------
-- 07_consolidate.sql
-- ----------------------------------------------------------------------------
-- db/v1/07_consolidate.sql — Rellenos e invariantes finales de v1.
-- Parte 7 de 7 del paquete v1. NO ejecutar suelta: usar db/v1/install.sql.
-- Orden importa: va la ÚLTIMA porque exige que existan perfiles, privacidad,
-- cosméticos y los triggers de 01/04/06. Todo es reejecutable sin duplicar
-- (ON CONFLICT DO NOTHING / WHERE ... IS NULL).
-- El invariante NOT NULL de mc_uuid se aplica AQUÍ, después del relleno:
-- hacerlo antes (o en 01) rompería instalaciones con cuentas antiguas sin UUID.

-- 1) Privacidad por defecto para cuentas anteriores al trigger.
insert into public.privacy_settings (user_id)
select user_id from public.profiles
on conflict (user_id) do nothing;

-- 2) UUID de Minecraft para cuentas sin él (no toca premiums enlazados).
-- El trigger profiles_fill_mc_uuid (01) cubre los registros futuros.
update public.profiles
  set mc_uuid = public.offline_mc_uuid(username)
  where mc_uuid is null;

-- Invariante a nivel BD: todo perfil tiene UUID (repetible sin error).
-- Si quedara algún null (carrera extrema durante la instalación), el UPDATE
-- de arriba ya lo cubrió dentro de la misma transacción de install.sql.
alter table public.profiles alter column mc_uuid set not null;

-- 3) Catálogo de cosméticos para cuentas anteriores a su creación.
-- Los perfiles futuros lo reciben por trigger; los cosméticos futuros llegan
-- a todos por el trigger on_cosmetic_grant_all (06).
insert into public.user_cosmetics (user_id, cosmetic_id)
select p.user_id, c.id from public.profiles p cross join public.cosmetics c
on conflict do nothing;

-- 4) Estadísticas para el planificador (tablas pequeñas; coste despreciable).
analyze public.profiles;
analyze public.privacy_settings;
analyze public.friend_requests;
analyze public.friendships;
analyze public.cosmetics;
analyze public.user_cosmetics;
analyze public.blocks;
analyze public.notifications;
analyze public.file_uploads;

-- Marca v1 como aplicada (reejecutable: no duplica).
insert into public.schema_version (version) values ('v1')
on conflict (version) do nothing;

commit;
