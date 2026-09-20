-- db/v1/verify.sql — Comprobación tras aplicar db/v1/install.sql.
-- Pegar en Supabase → SQL Editor → Run. Cada consulta indica el esperado.
-- Si alguna falla, NO desplegar la API contra esa BD: revisar el mensaje.

-- 1) Versión registrada (esperado: una fila con version = 'v1').
select version, applied_at from public.schema_version where version = 'v1';

-- 2) Tablas v1 con RLS activo (esperado: 9 filas, rowsecurity = true).
select tablename, rowsecurity from pg_tables
where schemaname = 'public' and tablename in (
  'profiles', 'privacy_settings', 'file_uploads', 'notifications',
  'friend_requests', 'friendships', 'cosmetics', 'user_cosmetics', 'blocks'
)
order by tablename;

-- 3) Columnas finales del perfil (esperado: 7 filas).
select column_name from information_schema.columns
where table_schema = 'public' and table_name = 'profiles'
  and column_name in ('bio', 'last_mc_version', 'mc_uuid', 'banner_url', 'is_online',
                      'created_at', 'updated_at')
order by column_name;

-- 4) Nadie sin UUID + columna obligatoria (esperado: sin_uuid = 0, is_nullable = NO).
select count(*) as sin_uuid from public.profiles where mc_uuid is null;
select is_nullable from information_schema.columns
where table_schema = 'public' and table_name = 'profiles' and column_name = 'mc_uuid';

-- 5) Buckets v1 (esperado: 4 filas; avatars/banners públicos, skins/capes no).
select id, public from storage.buckets
where id in ('avatars', 'skins', 'capes', 'banners')
order by id;

-- 6) Sin política amplia de lectura global (esperado: 0 filas = warning resuelto).
select policyname from pg_policies
where schemaname = 'storage' and tablename = 'objects'
  and policyname in ('avatars_public_read', 'banners_public_read');

-- 7) Funciones sensibles NO accesibles a anon/authenticated
-- (esperado: 0 filas; solo postgres/service_role pueden ejecutarlas).
select routine_name, grantee from information_schema.role_routine_grants
where routine_schema = 'public'
  and routine_name in (
    'username_taken', 'email_for_username', 'resolve_friend_target',
    'accept_friend_request', 'remove_friendship', 'search_users', 'my_friends',
    'resolve_user_identity', 'block_user', 'unblock_user', 'equipped_cosmetics'
  )
  and grantee in ('anon', 'authenticated')
order by routine_name, grantee;

-- 8) Catálogo de cosméticos (esperado: >= 4 filas).
select slug, kind from public.cosmetics order by slug;

-- 9) file_uploads acepta banner (esperado: 1 fila con 'banner' en el CHECK).
select conname, pg_get_constraintdef(oid) as definicion
from pg_constraint
where conrelid = 'public.file_uploads'::regclass and conname = 'file_uploads_kind_check';
