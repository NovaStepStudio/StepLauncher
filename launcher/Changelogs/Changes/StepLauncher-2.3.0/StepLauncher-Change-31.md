# StepLauncher-Change-31: extraData como objeto y nuevo historial de crasheos

- **Fecha**: 2026-08-06
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

### 1. `extraData` pasa de array de strings a objeto con claves

La sección `extraData` de `launcher_config.json` ya no es una lista de
archivos sino un objeto que referencia **todos** los `launcher_*.json` que
genera el launcher (los JSON de descarga de Minecraft no se incluyen):

```json
"extraData": {
  "assets": "launcher_assets.json",
  "accounts": "launcher_accounts.json",
  "history": "launcher_history.json",
  "profiles": "launcher_profiles.json",
  "crashHistory": "launcher_history_crashes.json"
}
```

- `internal/Config/Config.go`: nuevo tipo `ExtraData` con las 5 claves y un
  `UnmarshalJSON` propio que acepta **tanto el objeto nuevo como el array de
  strings legacy** (las configs viejas se migran solas al cargar, sin perder
  el resto de secciones). `sanitize` rellena/restaura cada clave con su
  nombre por defecto si falta o es inválido.
- `RegisterExtraFile(key, name)` ahora recibe la clave del campo
  (`ExtraKeyAssets`, `ExtraKeyAccounts`, ...) además del nombre del archivo.
- `internal/Handlers/App.go` (`initAssets`) registra los 5 archivos en el
  arranque (antes solo `assets` y `accounts`).

### 2. Historial de crasheos (`launcher_history_crashes.json`)

Nuevo paquete `internal/Core/Launcher/History/CrashHistory.go`: `CrashManager`
(mismo patrón que `Manager`) con `Load`/`EnsureFile`/`AddEntry`/`GetEntries`/
`Clear` persistido en `launcher_history_crashes.json`.

Se registra una entrada por cada crash con:

- `version`, `exit_code`, `crash_reason`, `crash_category`, `timestamp`.
- **Rutas de los logs implicados, relativas al workdir** (`/` como separador):
  - `launcherLogPath`: log del launcher, p. ej. `logs/StepLauncher-2.3.0-2026-08-06.log`
  - `minecraftLogPath`: log del juego, p. ej. `logs/game/StepLauncher-2.3.0-2026-08-06.log`
  - `jvmLogPath`: crash-report del juego o log JVM (`hs_err`) si existe, p. ej.
    `game/crash-reports/crash-2026-08-05_23.49.30-client.log`
- Se **eliminaron** `InstanceID` y `PlayerName` del registro: cada lanzamiento
  usa un ID único (no es una instancia persistente) y el jugador no es dato
  relevante del crash.

Dónde se dispara: `internal/Handlers/Engine/Engine.go` en `OnGameExitFn`
cuando `gi.Status == GameCrashed`. `Logger.GetLogPath()` (nuevo en
`internal/Core/Logger`) expone la ruta del log del launcher. Helper
`relPathToWorkDir` normaliza a rutas relativas.

### 3. Binding nuevo

- `Engine.GetCrashHistory()` → `[]history.CrashEntry` y el binding
  `App.GetCrashHistory()` (tipo `CrashEntry` en `internal/Handlers/Engine/History.go`).

## Por qué

- `extraData` como lista plana no permitía saber qué era cada archivo; como
  objeto cada clave identifica un tipo de dato del launcher.
- No existía un historial persistente de crashes con referencias a los logs
  (el usuario solo veía el modal en vivo y el histórico general no guardaba
  rutas de logs).

## API afectada

- `Config.Config.extraData`: de `string[]` a objeto (retrocompatible: las
  configs con el array viejo se migran solas).
- `Config.Manager.RegisterExtraFile(key, name)`: firma cambiada.
- Nuevo binding `App.GetCrashHistory()`.
- Bindings regenerados con `wails build` (models.ts: `ExtraData`, `history.CrashEntry`).

## Cómo verificar

- `go build ./internal/...` → OK (el `go build ./...` completo requiere el
  `StepLauncher-res.syso` que genera `wails build`).
- Lanzar una versión que crashee, cerrar, revisar `launcher_history_crashes.json`
  (con las 3 rutas de logs) y `extraData` en `launcher_config.json`.