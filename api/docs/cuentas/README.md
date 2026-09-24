# Cuentas StepLauncher (dominio maestro de la API)

Toda la API gira alrededor de la **cuenta del jugador**: identidad, perfil,
personalización, sesiones, notificaciones y vida social. Este archivo es el mapa;
el contrato HTTP exacto vive en `docs/api/v1/` (`auth.md`, `cuentas.md`,
`archivos.md`, `notificaciones.md`, `amigos.md`, `usuarios.md`, `comunidad.md`).

## Modelo (v1)

- **Identidad** = `auth.users` (email + contraseña + `created_at` + `last_sign_in_at`).
  La API la gestiona con `supabaseAdmin` en endpoints `sensitive` (ADR-007).
- **Perfil** = `public.profiles`: `username` único insensible a mayúsculas,
  `display_name`, `avatar_url`, `banner_url`, `bio` (≤5000), `last_mc_version`
  (Minecraft Java), `mc_uuid` (auto offline al crear, enlazable a premium, nunca null),
  `is_online` (`false` por defecto, lo cambia el dueño). RLS propia.
  SQL: `db/v1/` (paquete instalable; historial congelado en `db/migrations/`).
- **Privacidad** = `public.privacy_settings` (`searchable`, `allow_email_search`,
  `receive_friend_requests`, `receive_notifications`; todo `true` por defecto).
  Se respeta **a nivel SQL** en resolución y búsquedas, no solo en la API.
- **Sesiones** = tokens Supabase; `createdAt`/`lastSessionAt` se leen de `auth.users`
  vía admin en `GET /me` (autoritativo, sin duplicar). Ver `docs/oauth/README.md`.
- **Notificaciones** = `public.notifications` (las crea el servidor; respeta
  `receive_notifications` al notificar amistades). RLS propia, sin INSERT de cliente.
- **Amistades** = `friend_requests` (`pending`/`accepted`/`declined`/`cancelled`) +
  `friendships` (doble fila por amistad, creada por función atómica).
  Escritura directa denegada: solo API/funciones `SECURITY DEFINER` revocadas.
- **Comunidad** = preview pública autenticada (`/v1/community/profiles/:identifier`
  por UUID de cuenta, usuario o UUID-MC): datos + equipados + Minecraft + relación.
  Respeta `searchable` y lista negra. Email jamás expuesto.
- **Cosméticos** = catálogo `cosmetics` (4 de ejemplo) + `user_cosmetics` (se otorgan
  al crear la cuenta **y** al crear cosméticos nuevos; equipar con
  `/me/cosmetics/equip`). Solo equipados son visibles en la preview.
- **Lista negra** = `blocks`: bloquea preview y solicitudes en ambos sentidos;
  bloquear corta pendientes y amistad. Solicitar a quien te bloqueó = 404 genérico;
  a quien bloqueaste = 403.
- **Archivos** = buckets `avatars`/`banners` (públicos, sin listado global) y
  `skins`/`capes` (privados) + libro `file_uploads` (quota 10 skins+capas/día;
  avatar/banner se registran pero no cuentan).

## Endpoints (auth salvo registro/login/refresh)

| Método | Ruta | Uso |
|---|---|---|
| `POST` | `/v1/auth/register` · `/login` · `/refresh` · `/logout` | Sesiones base (ver `auth.md`) |
| `POST` | `/v1/auth/resend` · `/confirm` · `/recover` · `/reset-password` | Confirmación y recupero Supabase (ver `auth.md`) |
| `GET` | `/v1/accounts/me` | Perfil completo: bio, versión/UUID MC, banner, `isOnline`, fechas |
| `PATCH` | `/v1/accounts/me` | Usuario, nombre, bio, versión MC, UUID MC (`""` desvincula/regenera), `isOnline` |
| `PATCH` | `/v1/accounts/me/presence` | Heartbeat de presencia (`{ isOnline }`) |
| `POST` | `/v1/accounts/me/email-change` · `/password-change` | Email (verificado en el nuevo vía **Change email**) y contraseña (exige la actual, distinta; avisa **Password changed**) |
| `POST` | `/v1/accounts/me/avatar` · `/me/banner` | Avatar (PNG/JPG/WEBP ≤2 MB) y banner (PNG/GIF/JPG/WEBP ≤8 MB, máx 1080p) |
| `POST` | `/v1/accounts/me/cosmetics/equip` · `/unequip` | Equipar/desequipar cosmético propio |
| `GET`/`PATCH` | `/v1/accounts/me/privacy` | Las 4 opciones de privacidad |
| `POST` | `/v1/files/skin` · `/cape` | PNG ≤1 MB, quota 10/día |
| `GET` | `/v1/files` · `/files/:id/url` · `DELETE /files/:id` | Listar, URL firmada 1h, borrar |
| `GET` | `/v1/users/search?q&limit` | Buscar visibles (respeta `searchable`; sin `userId`) |
| `POST` | `/v1/friends/requests` | Solicitud por email/usuario/UUID |
| `GET` | `/v1/friends/requests?type=` | Bandeja `incoming`/`sent`/`all` |
| `POST` | `/v1/friends/requests/:id/accept` · `/decline` · `/cancel` | Aceptar (receptor) / rechazar (receptor) / cancelar (emisor) |
| `GET` | `/v1/friends` · `DELETE /v1/friends/:friendId` | Lista (con `isOnline`) y romper |
| `GET`/`POST`/`DELETE` | `/v1/friends/blocks…` | Lista negra |
| `GET` | `/v1/community/profiles/:identifier` | Preview + `relationship` (con `isOnline`) |
| `GET` | `/v1/notifications` (+ `read`, `read-all`, `DELETE`) | Bandeja propia (+ `unreadCount`) |

Pendiente (futuro, con ADR): `DELETE /v1/accounts/me` (borrado de cuenta, `sensitive`).

## Reglas del dominio

1. El dueño sale del Bearer; prohibido `userId` del cliente. Un token solo toca lo suyo.
2. Datos con `supabaseForUser()` (RLS). `supabaseAdmin()` solo para lo imposible con
   RLS (Auth, resoluciones sensibles, funciones), documentando por qué.
3. La privacidad se aplica en SQL (funciones/RLS), nunca confiando solo en la API.
4. Contrato incompatible → `/v2` (ver versionado).
