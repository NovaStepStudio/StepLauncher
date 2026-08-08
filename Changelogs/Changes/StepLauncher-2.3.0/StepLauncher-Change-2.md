# StepLauncher-Change-2: Login AuthLib (Yggdrasil), renovación de sesiones y panel de Cuentas rediseñado

- **Fecha**: 2026-08-03
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

El sistema de cuentas pasa a ser el flujo real de autenticación del launcher:
login contra servidores Yggdrasil (AuthLib) sin OAuth de Microsoft, renovación
de sesiones (manual, automática al arrancar y al lanzar) y un gestor rediseñado
como panel dedicado.

**Backend (Go):**
- Nuevo paquete `internal/Core/Auth/` (`Authlib.go`): cliente Yggdrasil puro
  (`Authenticate`, `Validate`, `Refresh`, `Invalidate`) con normalización de
  URL (`New`/`NormalizeServerURL`, siempre acaba en `<base>/authserver`),
  `AuthResult`/`Profile`, `AuthError` con el código del servidor y
  `readableStatus`.
- `internal/Core/Accounts/`:
  - `Types.go`: solo `offline | authlib` (se elimina `microsoft`); `Account`
    con `Email`, `SessionValid` y `AuthServerURL`; `AuthlibLoginReq`.
  - `Auth.go`: `NormalizeAccountType` sin Microsoft, `newClientToken` (UUID v4).
  - `Manager.go`: `LoginAuthLib` (goroutine + evento `account_login`),
    `RefreshAuthLib` (evento `account_refresh`), `RefreshAll` (evento
    `account_refresh_all`), `CancelLogin` (cancela el login en curso vía
    `context.CancelFunc` guardada en el Manager, sin locks durante la llamada),
    `SetAutoRefresh`/`GetAutoRefresh`, y `ResolveForLaunch` con política de
    sesión (validar → si la red no responde lanzar con el token guardado → si
    el token expiró: autoRefresh renueva o aborta pidiendo reiniciar sesión).
  - `Load()` ahora lee directamente `launcher_accounts.json` (ya no depende de
    `launcher_profiles.json` ni de su sección `accounts`); `persist()` escribe
    el fichero completo.
- `internal/Handlers/Engine/Accounts.go`: handlers `LoginAuthlib`,
  `RefreshAccount`, `RefreshAllAccounts`, `SetAccountsAutoRefresh`,
  `GetAccountsAutoRefresh`, `ResolveAccountCredentials` (vía
  `ResolveForLaunch`) y `CancelLogin`.
- `app.go`: bindings `LoginAuthlib`, `CancelAuthlibLogin`, `RefreshAccount`,
  `RefreshAllAccounts`, `SetAccountsAutoRefresh`, `GetAccountsAutoRefresh`.
- `Engine.go`: `SetEventFn` conectado al gestor de cuentas para emitir eventos
  al frontend.

**Frontend (Vue 3 + TS):**
- `stores/accounts.ts`: tipos sin Microsoft, `loginAuthlib`, `cancelLogin`,
  `refreshAccount`, `refreshAllAccounts`, `setAutoRefresh`/`autoRefresh` y las
  constantes de eventos de ventana `ACCOUNT_LOGIN_START_EVENT` y
  `ACCOUNT_OPEN_SETTINGS_EVENT`.
- `Modals/AccountsModal.vue`: el modal pasa a **panel completo** (56rem × 86vh):
  contenido central con ligera transparencia (fondo `rgba(0,0,0,0.38)`) y se
  elimina la franja `::before` con gradiente de acento.
- `Modals/AccountsView.vue`: **barra lateral derecha sólida** (fondo
  `var(--background-modal-primary)`, sin transparencia) con las acciones
  principales (Añadir Offline, Iniciar sesión AuthLib, Renovar todas,
  Actualizar, Configuración); el formulario y la lista viven en el contenido
  central. Se quita de aquí el toggle de renovación automática.
- `Modals/LoginProgressModal.vue` (nuevo): al pulsar "Iniciar sesión" el panel
  se cierra y este modal muestra el progreso con spinner y botón **Cancelar**
  (llama a `CancelAuthlibLogin`); estados de éxito y error con el mensaje del
  backend. Sin franja de gradiente.
- `Layouts/Sections/Settings/AccountsSettings.vue` (nuevo): sección **Cuentas**
  de Configuración con el toggle "Renovar sesiones al iniciar el launcher".
- `Modals/SettingsModal.vue`: nueva prop `initialSection` para abrir
  Configuración aterrizando en una sección concreta.
- `App.vue`: sección "Cuentas" en `settingsSections`; `handleIdle` cierra
  ahora también el panel de cuentas y el modal de login; suscripciones a los
  eventos `account_login`/`account_refresh`/`account_refresh_all` y llamada a
  `refreshAllAccounts()` al arrancar si el autoRefresh está activo.

## Por qué

Se rehizo el sistema de cuentas con soporte AuthLib (Yggdrasil) sin OAuth
Microsoft, con un panel dedicado (sidebar derecha sólida, contenido central
translúcido) y un login con modal de progreso cancelable; el toggle de
renovación se movió a una sección nueva de Configuración.

## API afectada

- Nuevos bindings (regenerados por `wails build`): `LoginAuthlib`,
  `CancelAuthlibLogin`, `RefreshAccount`, `RefreshAllAccounts` (int),
  `SetAccountsAutoRefresh`, `GetAccountsAutoRefresh`.
- Nuevos eventos de runtime: `account_login`, `account_refresh`,
  `account_refresh_all` (emitidos por el engine hacia el frontend).
- `launcher_accounts.json` pasa a ser el archivo real de cuentas (antes era la
  sección `accounts` de `launcher_profiles.json`).

## Cómo verificar

- `wails build` en la raíz (regenera bindings + embebe el frontend) y
  `bun run build` en `frontend/`.
- En la app: abrir Cuentas desde la tarjeta de usuario → panel con sidebar
  derecha; "Iniciar sesión AuthLib" con un servidor Yggdrasil real → se cierra
  el panel y aparece el modal de progreso; Cancelar aborta la petición.
- Comprobar la renovación: con autoRefresh activo, al arrancar se emiten
  `account_refresh_all`; en el panel, "Renovar todas" y el botón por cuenta.
- El toggle de renovación vive en Configuración → Cuentas.
