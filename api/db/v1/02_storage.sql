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
