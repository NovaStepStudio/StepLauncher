# Cambios de StepLauncher 2.5.0 (Change-14) — Eliminación completa del sistema Login/Register/Dashboard

- **Fecha**:   2026-09-20
- **Versión**: 2.5.0
- **Estado**:  implementado y verificado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

### 1. Dominios eliminados por completo (`website/web/src/`)

- `Auth/` (5 archivos): `Login.vue`, `Register.vue`, `Store.ts` (store Pinia `auth` con login/register/logout/refresh/perfil), `Components/AuthLayout.vue` y `Components/TurnstileWidget.vue`.
- `Dashboard/` (23 archivos): `Index.vue`, `Store.ts` (store Pinia `dashboard-demo`), `Api.ts` (mocks de actividad/cosméticos/amigos/admin) y sus 20 componentes (`AccountPanel`, `ActivityFeed`, `ActivityStats`, `AdminPanel`, `CommunityPanel`, `CosmeticsPanel`, `FriendsPanel`, `InstancesList`, `KitsPanel`, `NotificationsBell`, `OwnerPanel`, `PrivacyPanel`, `ReportButton`, `ResourcesPanel`, `ResumenPanel`, `SessionsPanel`, `SettingsPanel`, `Sidebar`, `SocialPanel`, `StatsGrid`, `VersionsList`).
- `Common/Api/` (2 archivos): `client.ts` (sesión, `Authorization: Bearer`, refresh, Turnstile, `VITE_ACCOUNTS_API_BASE`) y `accounts.ts` (los 8 módulos: auth, user, admin, owner, content, explore, libraries, updates). Nadie más los importaba.
- Huérfanos del sistema: `Common/BlobPet.vue` (mascota de la cuenta, solo usada por `Register` y el dashboard) y `web/.env` (solo contenía `VITE_TURNSTILE_SITEKEY` y `VITE_ACCOUNTS_API_BASE`).

### 2. Referencias limpiadas

- `Router.ts`: solo quedan las 4 rutas públicas (`/`, `/download`, `/about`, `/changelog`). Fuera las rutas `/login`, `/register`, `/dashboard`, el guard `beforeEach` y el import de `leerSesionApi`.
- `Main.ts`: fuera `blobatar/motion.css`, fuera `createPinia` y `app.use(createPinia())` (Pinia ya no la usa nadie).
- `Common/Header.vue`: fuera `useAuthStore`, fuera los enlaces condicionales `Acceder`/`Dashboard` y sus iconos (`IconLogin2`, `IconLayoutDashboard`), fuera el CSS muerto `.Login`/`.Dash`. Quedan Inicio, Descarga, Historial, Acerca De y Github.
- `Common/Footer.vue`: eliminada la columna `Cuenta demo` (`/login`, `/register`, `/dashboard`).
- `App.vue`: actualizado el comentario que mencionaba el dashboard.
- `package.json` (+ `bun.lock` regenerado): eliminadas `pinia`, `@blobatar/vue` y `blobatar` (`bun install` quitó 3 paquetes).

## Por qué

A pedido del owner: el sistema de cuentas web (login/register/dashboard contra la Accounts API) se elimina por completo del sitio. El sitio queda como web pública e informativa sin sesiones ni panel.

## API afectada

Solo frontend web (`website/`). Sin cambios en el backend Go ni bindings. La Accounts API remota no se toca; simplemente el sitio deja de consumirla.

## Comportamiento anterior/nuevo

- Antes: rutas `/login`, `/register` y `/dashboard` con guard de sesión, header/footer con accesos a la cuenta y dependencias `pinia`/`blobatar`.
- Ahora: esas rutas devuelven 404 del router (solo existen `/`, `/download`, `/about`, `/changelog`), no hay sesión, ni guard, ni stores, ni Turnstile. Verificado con grep: cero menciones a `login`, `pinia`, `Auth` o `Dashboard` en `website/web/src`.

## Cómo verificar

- `bun install` en `website/` (quita 3 paquetes, regenera el lockfile).
- `bun run build` en `website/` (pasa `vue-tsc --build` y `vite build`: 6318 módulos, `✓ built in ~15s`).
