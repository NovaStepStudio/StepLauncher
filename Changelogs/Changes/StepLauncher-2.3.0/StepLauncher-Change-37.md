# StepLauncher-Change-37.md

## Fecha
2026-08-07

## Release
StepLauncher-2.3.0 — se mencionó por primera vez en esta release.

## Cambio
**Sistema de actualizaciones desde GitHub**. El launcher consulta la última release del repositorio `https://github.com/NovaStepStudio/StepLauncher` (GitHub API), compara su versión con la instalada (`2.3.0`) y, si hay una más nueva, muestra un modal con los detalles de la release y el botón "Actualizar". En Windows la actualización se aplica de forma automática descargando el `StepLauncher-Updater.exe` de los assets de la release, lanzándolo como proceso independiente y cerrando el launcher; en Linux/macOS NO se descarga nada: se abre la última release en el navegador para descarga manual. La actualización NUNCA es obligatoria ("Más tarde" siempre disponible). Nueva opción en Ajustes → Comportamiento: "Buscar actualizaciones al iniciar" (toggle) + botón "Buscar actualización".

## Por qué
Permitir que los usuarios tengan siempre la última versión del launcher de forma cómoda y automática en Windows, sin depender de descargas manuales, y avisar de nuevas versiones en el resto de plataformas.

## Qué se hizo
1. **`internal/Handlers/Engine/Update.go`** (nuevo):
   - `UpdateInfo` (hasUpdate, latestVersion, currentVersion, releaseUrl, releaseName, releaseDate, notes, hasUpdater, updaterUrl, platform, error) — resultado de la comprobación que viaja al frontend por el evento `update_check` (JSON) y se guarda para aplicar sin repetir la consulta.
   - `CheckForUpdates`: lanza en segundo plano (la red no bloquea el binding) la consulta a la API `releases/latest`, emite `update_check` y guarda el resultado en `lastUpdate` (protegido con `updateMu`). Devuelve inmediatamente.
   - `checkUpdate`: GET con timeout de 20s a GitHub API, parser de la release, `compareVersions` (X.Y.Z, sin prefijo "v"), búsqueda del asset `StepLauncher-Updater.exe` solo en Windows (no se filtra por arquitectura: ya no se publican builds de 32 bits). Nunca devuelve error: las fallas van a `UpdateInfo.Error`.
   - `DownloadUpdater`: descarga el updater a `%temp%/StepLauncher-Updater/StepLauncher-Updater.exe` (carpeta separada para que pueda reemplazar el binario sin colisiones; retry por rename cuando el destino ya existe).
   - `LaunchUpdater`: ejecuta el updater como proceso desacoplado.
   - Constantes: `updateRepo`, `updateAPILatest`, `updaterAssetName = "StepLauncher-Updater.exe"`, `updaterTempDir = "StepLauncher-Updater"`.
2. **`internal/Handlers/Engine/updater_windows.go`** (nuevo): lanza el updater con `CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS`: proceso independiente y sin consola, no se cierra cuando el launcher termina. **`updater_unix.go`** (nuevo): `Setsid` (aunque no exista updater en esas plataformas, deja preparado el mecanismo).
3. **`internal/Handlers/Engine/Engine.go`**: campos `updateMu sync.Mutex` y `lastUpdate *UpdateInfo`; helper `LastUpdateInfo`.
4. **`internal/Handlers/App.go`**: bindings `CheckForUpdates` (delegado), `ApplyUpdate` (en Windows descarga el updater con `DownloadUpdater`, lo lanza con `LaunchUpdater` y `runtime.Quit`; si la release no trae updater o en Linux/macOS abre la release con `runtime.BrowserOpenURL`), `GetCheckForUpdatesOnStart`/`SetCheckForUpdatesOnStart`.
5. **`internal/Config/Config.go`**: nueva sección `launcher.checkForUpdatesOnStart` (bool, default `false`) y setter `SetCheckForUpdatesOnStart`.
6. **`app.go` (raíz)**: bindings Wails de `ApplyUpdate`, `CheckForUpdates`, `GetCheckForUpdatesOnStart`, `SetCheckForUpdatesOnStart` (se regeneran con `wails build`/`wails dev`).
7. **Frontend**: nueva store `frontend/web/src/stores/update.ts` (escucha el evento `update_check`; distingue comprobación manual `checkForUpdates()` vs automática que solo muestra modal si hay update; `installUpdate`, `closeUpdateModal`); nuevo modal `frontend/web/src/Modals/UpdateModal.vue` (checker con spinner, estado "actualizado" con check verde, estado de error, botones "Actualizar ahora" (abre release en Windows sin updater) / "Más tarde"); `App.vue` lo registra (`bindUpdateEvents()` al montar, `<UpdateModal/>` y `checkForUpdates(true)` al arrancar si `checkForUpdatesOnStart` está activo) y `GeneralSettings.vue` añade el toggle y el botón "Buscar actualización" (con feedback mientras el check está en curso).

## API afectada
- Bindings Wails nuevos: `App.CheckForUpdates()`, `App.ApplyUpdate()`, `App.GetCheckForUpdatesOnStart()`, `App.SetCheckForUpdatesOnStart(bool)`.
- Evento nuevo: `update_check` (JSON de `UpdateInfo`).
- `launcher_config.json`: nueva prop `launcher.checkForUpdatesOnStart` (ausente → `false`).

## Cómo verificar
- `go build ./...` OK en Windows; `go build ./internal/Handlers/Engine/` OK cruzado para Linux; `bun run build` (type-check + vite) OK en `frontend/`; `wails build` OK (regenera bindings y embeber el dist).
- Con una release más nueva en `NovaStepStudio/StepLauncher`: al pulsar "Buscar actualización" (o al iniciar con la opción activa) aparece el modal con la nueva versión; "Actualizar ahora" en Windows descarga el updater, lo lanza y el launcher se cierra; con una release sin updater (o en Linux/macOS) abre la release en el navegador.
- Con la versión al día: el modal indica "Estás al día"; si la comprobación fue automática al iniciar no se muestra el modal (no molesta).

## Notas
- `compareVersions` compara segmentos numéricamente si es posible y como texto en caso contrario (soporta tags con prefijo `v`).
- El updater se lanza DESACOPLADO del launcher (nuevo grupo de procesos): el launcher puede cerrarse justo después y el reemplazo de archivos sigue en pie.
- La integración del modal respeta el sistema de variables CSS existente (spinner/gifs de assets, iconos de lucide) y los paneles registrados en Ajustes → Comportamiento (prefijo `Ss*`).