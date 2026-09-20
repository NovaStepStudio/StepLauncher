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
