# Integración con la API real — de dónde salen los usuarios

Este servicio **no crea cuentas**: la fuente única de usuarios es la **misma
base Supabase** que usa la API principal (`api/`): `auth.users` + tabla
`public.profiles`. La API principal sigue siendo la dueña del registro, la
confirmación por correo y el perfil; Yggdrasil solo **verifica y juega**.

## Mapa de fuentes

| Dato Yggdrasil | Origen real | Cómo se lee |
|---|---|---|
| Email + contraseña | `auth.users` (Supabase Auth) | `signInWithPassword` en servidor (como `api/src/routes/v1/auth.ts:176`) |
| Email desde usuario | `public.email_for_username` (RPC de `api/`, `db/v1/01_profiles.sql:94`) | Solo con `service_role` (revocada a resto) |
| Nombre de jugador | `public.profiles.username` (único ci) | Lectura admin por `user_id` |
| UUID del perfil | `public.profiles.mc_uuid` (offline auto u premium enlazado) | Lectura admin por `user_id` / `mc_uuid` |
| Confirmación de correo | `auth.users.email_confirmed_at` | Sin confirmar → sin juego (igual que `api/` exige) |
| Skins/capas (lectura) | Buckets `skins`/`capes` + libro `file_uploads` | Última subida por tipo; bytes por `service_role`, hash SHA-256 como nombre |
| Mapa de texturas | Propio `ygg_textures` (hash → bucket/path) | Caché de escritura al servir perfiles; reejecutable |
| Skins/capas (fase 2) | Buckets `skins`/`capes` de `api/` + libro `file_uploads` | URL firmada + hash como nombre de archivo |

## Por qué así y no con usuarios propios

- Una sola contraseña: el jugador usa la **misma** en la web, el launcher y el
  juego. Sin sincronías ni migraciones.
- Un solo nombre: `profiles.username` es el `name` del perfil Yggdrasil; el
  `mc_uuid` ya existe para todos (offline garantizado por `api/`, ver ADR-011).
- La API principal ya resolvió lo difícil: unicidad, confirmación, recupero,
  RLS y quota. Duplicarlo aquí sería una segunda fuente de verdad (prohibido).

## Qué tablas son propias (solo juego)

`db/01_ygg.sql` crea **solo** lo que `api/` no tiene:

- `ygg_tokens` (`access_token` PK, `client_token`, `user_id` → `auth.users`,
  `profile_uuid`, `profile_name`, `issued_at`, `expires_at` +15 días,
  `revoked`). Límite 10 vivos por usuario (se revoca la más vieja).
- `ygg_joins` (`server_id` PK, `access_token`, `profile_uuid`, `profile_name`,
  `client_ip`, `expires_at` +30 s). Sin RLS de cliente: solo `service_role`.

Nada de estas tablas se expone por `api/` ni por RLS de usuario: son internas
del Worker Yggdrasil (igual que `api/` restringe `supabaseAdmin()`).

## Flujo de login (puente real)

1. Llega `POST /authserver/authenticate` con `{ username, password }`.
2. Si trae `@` es email; si no, se resuelve con `email_for_username`
   (nombre de StepLauncher → email). Así vale email **o** usuario.
3. `signInWithPassword(email, password)` en servidor. Si falla o el email no
   está confirmado → `403` genérico (no se revela el motivo).
4. Se lee `profiles` por `user_id` → `{ username, mc_uuid }`.
5. Se emite `ygg_tokens` con ese perfil atado y se responde en formato
   Yggdrasil (`availableProfiles`, `selectedProfile`, `user?`).

Registro, confirmación, recupero y cambio de email/contraseña **no existen
aquí**: se hacen en `api/` (`POST /v1/auth/*`) y en la web. Este servicio
solo refleja su resultado (poder entrar o no).

## Límites y secretos compartidos

- Mismo proyecto Supabase que `api/`: mismos `SUPABASE_URL`,
  `SUPABASE_ANON_KEY`, `SUPABASE_SERVICE_ROLE_KEY` (secretos por entorno,
  jamás en código).
- Clave nueva y propia: `YGG_SIGN_PRIVATE_KEY` (RSA 2048 PKCS8) para firmar
  `textures`. La pública se publica en `GET /`. Rotarla invalida firmas
  viejas: los clientes re-descargan metadatos (documentar la rotación).
- Rate-limit propio (más duro en `authenticate`/`signout`: 20/10 min por IP).
