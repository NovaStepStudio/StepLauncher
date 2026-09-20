# StepLauncher-Change-1: Sistema de Cuentas del launcher

- **Fecha**: 2026-08-03
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

Se añadió el **sistema de Cuentas** del launcher, siguiendo el patrón del
sistema de perfiles de `internal/Core/Launcher/Profile` (Manager con
`sync.RWMutex`, persistencia JSON, CRUD + selección y `sanitize`):

**Backend (Go):**
- Nuevo paquete `internal/Core/Accounts/`:
  - `Types.go`: `Account`, `AccountType` (`offline | authlib | microsoft`),
    `AccountInfo` (vista segura sin tokens), `AccountCredentials` (credenciales
    resueltas para lanzar), `CreateAccountReq`, `AccountsFile`.
  - `Auth.go`: validación de nombre de jugador, `OfflineUUID` (UUID v3 por MD5
    de `OfflinePlayer:<nombre>`, igual que Minecraft vanilla) y validación de
    peticiones por tipo de cuenta.
  - `Manager.go`: persistencia en `workDir/launcher_accounts.json`, CRUD,
    cuenta seleccionada, `ResolveCredentials` (usa la seleccionada o la primera)
    y `TouchLastUsed`. Sigue la disciplina anti-deadlock de `internal/Config`
    (mutar bajo `Lock()`, loguear fuera; `logf` nunca se llama con el lock).
- `internal/Handlers/Engine/Engine.go`: el motor NovaCore registra un
  `accounts.Manager` (carga al arrancar, loguea fallos sin romper la app).
- `internal/Handlers/Engine/Accounts.go`: handlers del engine y helper
  `fillAccountCredentials`, que rellena las credenciales de `LaunchConfig` con
  la cuenta seleccionada cuando el frontend no envía `username` (nunca pisa
  credenciales explícitas).
- `app.go`: bindings `ListAccounts`, `GetAccount`, `CreateAccount`,
  `UpdateAccount`, `DeleteAccount`, `GetSelectedAccount`, `SetSelectedAccount`,
  `ResolveAccountCredentials`.
- Integración de lanzamiento: `LaunchMinecraft` y `LaunchInstance` usan la
  cuenta seleccionada si no se pasan credenciales; el `userType` se ajusta
  (`authlib` para cuentas AuthLib).
- `internal/Handlers/App.go`: `launcher_accounts.json` se registra en
  `extraData` de `launcher_config.json` junto a `launcher_assets.json`.

**Frontend (Vue 3 + TS):**
- `stores/accounts.ts`: store reactivo (lista, seleccionada, CRUD, etiquetas).
- `components/AccountsView.vue`: vista reutilizable (lista + formulario por
  tipo + edición + borrado + selección), usa las variables CSS existentes y
  las clases `Ss*` de `Styles/Settings.scss`. Nuevo alias `@components` en
  `vite.config.ts` y `tsconfig.app.json`.
- `Modals/AccountsModal.vue`: modal de cuentas (abre desde la tarjeta de
  usuario; se cierra con el reposo vía `CLOSE_OVERLAYS_EVENT`).
- `Layouts/Sections/Settings/AccountsSettings.vue`: sección **Cuentas** en
  Ajustes (embebe la vista compartida, recarga con `onActivated`).
- `App.vue`: nueva sección en `settingsSections`, la `UserCard` muestra la
  cuenta activa (`username` + tipo) y abre el modal al pulsarla; carga las
  cuentas al arrancar.

## Por qué

El launcher no tenía forma de gestionar con qué cuenta se juega: el flujo de
lanzamiento dependía de credenciales pasadas a mano. Con este sistema, las
cuentas quedan guardadas, se elige una activa y el motor la inyecta
automáticamente al lanzar.

## API afectada

- Nuevos bindings en `window.go.main.App` (regenerados por `wails build`):
  los 8 métodos de cuentas listados arriba.
- Nuevos modelos en `frontend/wailsjs/go/models.ts` (`accounts.AccountInfo`,
  `accounts.AccountCredentials`, `accounts.CreateAccountReq`).
- `LaunchMinecraft` / `LaunchInstance`: comportamiento ampliado (sin
  credenciales explícitas usan la cuenta seleccionada); firma sin cambios.
- Nuevo alias de import `@components` (vite.config.ts + tsconfig.app.json).

## Cómo verificar

- `go build ./...` requiere el recurso de Wails: usar `wails build` (regenera
  bindings, compila el frontend con `vue-tsc` y produce
  `build/bin/StepLauncher.exe`).
- En la app: abrir Ajustes → **Cuentas** (o pulsar la tarjeta de usuario),
  crear una cuenta Offline, seleccionarla (la tarjeta muestra el username),
  editar y eliminar. El archivo `%APPDATA%\.StepLauncher\launcher_accounts.json`
  se crea/actualiza con cada operación.
- Verificar `bun run build` en `frontend/` (type-check + build de Vite).
