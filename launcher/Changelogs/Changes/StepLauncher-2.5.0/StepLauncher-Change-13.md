# Cambios de StepLauncher 2.5.0 (Change-13) — Dashboard conectado a la Accounts API

- **Fecha**:   2026-09-19
- **Versión**: 2.5.0
- **Estado**:  implementado y verificado
- **Release**: en desarrollo — aún no mencionado en ninguna release.

## Qué cambió

### 1. Cliente real de la Accounts API (`Common/Api/`)

- `client.ts`: base `https://steplauncher.stepnicka012.workers.dev/v1` (o `VITE_ACCOUNTS_API_BASE`), interceptor `Authorization: Bearer`, refresh proactivo (5 min de margen) y reactivo en `401 SESSION_NOT_FOUND`/`TOKEN_EXPIRED`, backoff con `Retry-After` para `RATE_LIMITED`, errores tipados por `code` (nunca por `message`), autodetección de dispositivo (`device_name`, `os`, `os_version`) y loader de Turnstile.
- `accounts.ts`: endpoints tipados en 8 módulos — `auth`, `user`, `admin` (6), `owner` (7), `content`, `explore` (8), `libraries` (18), `updates` (2).

### 2. Login simple para usuarios (sin demo, sin campos técnicos)

- `Auth/Login.vue`: solo email + contraseña + "Recuérdame". Sin tipo de sesión (la web usa siempre `web`), sin nombre de dispositivo (se autodetecta), sin botón de cuenta demo, sin credenciales precargadas.
- "Recuérdame" activado guarda en `localStorage`; desactivado usa `sessionStorage` (se cierra al cerrar el navegador).
- Turnstile visible: se renderiza con la sitekey de `web/.env` (`VITE_TURNSTILE_SITEKEY`) y el login/register exigen completarlo antes de enviar.
- Pantalla "Cuenta suspendida" con motivo y expiración cuando la API devuelve `ban_info`, y mensaje "intentá en Xs" con `Retry-After` en rate-limit.

### 3. Register real (5 campos, 201 sin token)

- `Auth/Register.vue`: email + contraseña (mín. 8) + nombre (máx. 50) + sexo + Turnstile invisible. Al crear (201) redirige al login porque la API no devuelve token.

### 4. Dashboard ampliado (11 secciones)

- Nuevos: `SessionsPanel` (gestionar dispositivos, revocar una/todas), `PrivacyPanel` (12 controles), `ResourcesPanel` (upload drag&drop + preview + gestión + barra de límites del rango), `KitsPanel` (CRUD + aplicar), `SocialPanel` (posts/feed, amigos con buscador, follows, likes, presencia, historial con gráfica), `CommunityPanel` (buscador, trending, colecciones, notificaciones, explore equipable, libraries Modrinth, canal de updates), `NotificationsBell` (badge `unread-count` + inbox), `OwnerPanel` (deep-user, roles, verificado, unban, cleanup, reset-ids, sync).
- `AccountPanel`: perfil real, edición con `current_password` si cambia email/clave, export GDPR (descarga JSON), zona peligro con doble confirmación, timeline de baneos, badge verificado y límites del rango.
- `AdminPanel`: stats reales con gráfica `usersByDay`, tabla de usuarios, modales ban/set-rank (tope 24h y sin legend para admin), monitor realtime por SSE (`EventSource`), cola de moderación, reportes y categorías.

### 5. Sin modo demo

- Se eliminó la cuenta demo local (`demo@steplauncher.dev`), los botones "usar demo", las cajas de credenciales y las etiquetas "Demo local"/"Panel demo". El router solo acepta sesión real.

## Por qué

El dashboard era una demo local con mocks. Ahora usa la API real de punta a punta, con login pensado para usuarios (sin exponer detalles de sesión/dispositivo) y cobertura de los 8 módulos.

## API afectada

Solo frontend web (`website/web/src`): `Common/Api/*`, `Auth/*`, `Dashboard/*`, `Router.ts`. Sin cambios en el backend Go ni bindings.

## Comportamiento anterior/nuevo

- Antes: login demo local con cualquier cuenta inventada; dashboard con mocks.
- Ahora: login real contra la API; sin sesión no se entra; cada panel consulta su endpoint.

## Cómo verificar

- `bun run type-check` en `website/` (pasa).
- `bun run build-only` en `website/` (compila).
- Captcha configurado en `website/web/.env` con la sitekey pública.
