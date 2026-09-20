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
