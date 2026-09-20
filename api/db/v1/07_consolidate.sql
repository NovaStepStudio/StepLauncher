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
