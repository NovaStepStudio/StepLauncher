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
