# Errores de StepLauncher 2.5.0 (Error-1) — `wails3 dev`/`wails3 build` rotos por la eliminación de build/ios

- **Fecha**:     2026-08-16
- **Versión**:   2.5.0 (en desarrollo)
- **Estado**:    corregido
- **Release**:   en desarrollo — aún no mencionado en ninguna release.

## Síntoma

Durante la migración a Wails v3, `go build ./...` fallaba porque `build/ios/app_options_default.go` (`//go:build !ios`) compilaba en Windows como paquete `main` sin `func main`. Para "arreglarlo" se eliminó la carpeta `build/ios` completa. Resultado:

- `wails3 dev` y `wails3 build` fallaban con: `ERROR GetFileAttributesEx ...\build\ios\Taskfile.yml: The system cannot find the file specified.`
- `wails3 update build-assets` NO restauró la carpeta: solo regenera los archivos "updatable" (`Info.plist`, `Info.dev.plist`, `entitlements.plist`, `LaunchScreen.storyboard`, `project.pbxproj`, `Assets.xcassets`), pero `Taskfile.yml`, `main.m`, `main_ios.go`, `app_options_*.go`, `icon.png` y `scripts/deps/install_deps.go` solo los extrae `wails3 init` (embed `internal/commands/build_assets`, no `updatable_build_assets`).

## Causa raíz

- La carpeta `build/ios` es parte integral del pipeline del Taskfile raíz (`Taskfile.yml` incluye `ios: ./build/ios/Taskfile.yml`); al faltar el archivo, **todas** las invocaciones de `wails3` (build, dev, package) fallaban al resolver los includes, no solo las de iOS.
- El conflicto real era único y localizado: `app_options_default.go` convierte `build/ios` en paquete `main` en escritorio sin declarar `func main`, lo que rompe `go build ./...` (comando de verificación del proyecto, AGENTS.md).

## Diagnóstico y evidencia

- Lectura del fuente del CLI: `internal/commands/build-assets.go` — `GenerateBuildAssets` (init) extrae `build_assets` + `updatable_build_assets`; `UpdateBuildAssets` (update) solo extrae `updatable_build_assets`. Por eso el update no pudo recrear la carpeta.
- Reproducción: `wails3 build` fallaba con el error de `GetFileAttributesEx` antes de tocar el código de la app.
- Inspección del embed del módulo (`internal/commands/build_assets/ios/`): contenido original de la carpeta.

## Solución aplicada

1. Restaurar `build/ios` completo copiando desde el embed del módulo (`build_assets/ios/*` → `build/ios/`): `Taskfile.yml`, `main.m`, `main_ios.go`, `app_options_ios.go`, `app_options_default.go`, `icon.png`, `scripts/deps/install_deps.go` (sin tocar los archivos que el update ya había regenerado).
2. Añadir `build/ios/build_stub.go` (`//go:build !ios`, `func main() {}`) — declara el `func main` que exige el paquete `main` de escritorio sin interferir con iOS (donde `main_ios.go` llama al `main()` definido por el usuario).
3. `go build ./...` pasó, y `wails3 build DEV=true` completó el pipeline hasta generar el binario.

### Resolución definitiva (decisión del owner)

El owner decidió que el proyecto sea **desktop-only (Windows/macOS/Linux)**: el soporte iOS no es necesario. Con su autorización explícita se eliminó `build/ios` completo y su include `ios: ./build/ios/Taskfile.yml` del `Taskfile.yml` raíz (el include faltante era lo que rompía `wails3 dev`/`build`). `build/ios/build_stub.go` dejó de existir con la carpeta y `go build ./...` ya no lo necesita.

## Regla aprendida

- **El alcance del proyecto es desktop-only (Windows/macOS/Linux)**: no existe `build/ios` ni include de iOS en el `Taskfile.yml`; si en el futuro se quisiera iOS, habría que regenerar la plantilla con `wails3 init` y añadir el include de nuevo.
- **No borrar carpetas de plataforma existentes del pipeline de Wails v3**: `build/<plataforma>/` (windows, darwin, linux, android) son requeridas por los includes del `Taskfile.yml` raíz aunque no se compilen.
- **`wails3 update build-assets` no restaura la carpeta completa de iOS**: solo regenera los archivos "updatable"; el resto solo lo extrae `wails3 init`.
- **Si un archivo de la plantilla rompe `go build ./...`** (p. ej. paquete `main` sin `func main`), se resuelve con un stub condicional por build tag en un archivo propio, nunca eliminando archivos o carpetas de la plantilla.

## Verificación

- `go build ./...` — EXIT 0.
- `wails3 build DEV=true` — EXIT 0 (bun install, bindings, frontend, syso, `bin/steplauncher.exe`).
- `bun run type-check` en `frontend/` — EXIT 0.
