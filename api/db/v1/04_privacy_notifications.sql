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
