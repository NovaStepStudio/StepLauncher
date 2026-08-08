# StepLauncher-Change-21: Toggle "Ocultar launcher al abrir Minecraft"

- **Fecha**: 2026-08-05
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

Nueva opción en Ajustes → GENERAL que oculta la ventana del launcher en cuanto
se inicia Minecraft y la vuelve a mostrar cuando el juego termina, se cierra o
se cae.

- **Backend**: campo `HideLauncherOnLaunch` en la config de launcher
  (`LauncherConfig` de `internal/Config/Config.go`), con valor **por defecto
  `true`** y `<json:"hideLauncherOnLaunch">`. Nuevo setter `SetHideLauncher(v)`
  en el `Manager` de Config, expuesto como `Handlers.App.SetHideLauncher` y
  binding `App.SetHideLauncher` en `app.go`. Al activar cualquier configuración
  se persiste el `launcher_config.json` (`save-launcher`).
- **Frontend** en `frontend/web/src/stores/launcher.ts`:
  - `launchGame()` oculta la ventana al recibir el `resp.id` del lanzamiento si
    el toggle está activado (`hideOnLaunchIfEnabled()`): lee
    `cfg.launcher?.hideLauncherOnLaunch` (se activa si no es exactamente
    `false`), comprueba con `ListGames()` que exista un juego en
    `starting|running` y llama a `WindowHide()` de `@wailsjs/runtime/runtime`.
  - Suscripción **perezosa y una sola vez** a `game_exited`, `game_crashed` y
    `game_stopped`: al recibir cualquiera de ellos se verifica con `ListGames()`
    que ya no queden juegos corriendo y, si la ventana sigue oculta, se restaura
    con `WindowShow()`.
  - Entorno de pruebas: si `window.runtime` no existe no lanza ningún error.
- **UI** en `frontend/web/src/Layouts/Sections/Settings/GeneralSettings.vue`:
  nuevo grupo **Comportamiento** (icono de rejilla) justo antes del grupo
  Internet, con el toggle `.SsTg` "Ocultar ventana al abrir Minecraft"
  (ref `hideLauncher`, valor inicial `cfg.launcher?.hideLauncherOnLaunch ?? true`)
  y `saveHideLauncher()` que invoca el binding.

## Por qué

Al abrir Minecraft la ventana del launcher estorbaba en pantalla. Se pidió que
el launcher se oculte automáticamente durante el juego y reaparezca al salir.

## API afectada

- Binding nuevo: `App.SetHideLauncher(hidden bool) (err error)` (se regenera con
  `wails dev`/`wails build`).
- Modelo/config: `LauncherConfig.hideLauncherOnLaunch` en
  `%APPDATA%\.StepLauncher\launcher_config.json`.

## Cómo verificar

- `go build ./...` → OK; `bun run build` (dentro de `frontend/`) → OK.
- `wails dev`: Jugar relanza la app → la ventana se oculta al arrancar Minecraft
  y vuelve a aparecer al cerrar/cerrar a la fuerza (eventos `game_exited` /
  `game_crashed` / `game_stopped`), sin suscripciones duplicadas.
- Ajustes → GENERAL → Comportamiento: activar/desactivar el toggle y comprobar
  que persiste (reiniciar la app).