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
