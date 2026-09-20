# Cambios de StepLauncher 2.5.0 (Change-9) — Infraestructura de build Wails v3: Taskfile, build/* y bindings

- **Fecha**: 2026-08-31
- **Versión**: 2.5.0
- **Estado**: implementado y verificado
- **Release**: en desarrollo — aún no mencionado en ninguna release.

## Qué cambió

Migración completa de la infraestructura de build de Wails v2 (`wails.json` + `frontend/package-lock.json` + `frontend/wailsjs`) a Wails v3 (`build/config.yml` + `Taskfile.yml` + `frontend/bindings`). Esta entrada traza `build/config.yml` nuevo, `wails.json` eliminado, `Taskfile.yml` nuevo y `frontend/bindings/*` nuevo.

### 1. `wails.json` eliminado → `build/config.yml` (Wails v3)

- **Antes** (`C:\Users\Stepnicka012\Desktop\Workflow\Go-Projects\StepLauncher-Wails3\wails.json:1` en `HEAD`): `"$schema": "https://wails.io/schemas/config.v2.json"`, `frontend:install: "npm install"`, `version: "2.3.1"`, sin `info.productIdentifier`.
- **Ahora**: `build/config.yml:1` con `version: '3'`, `info.productName: StepLauncher`, `productIdentifier: dev.github.novastepstudio.steplauncher`, `version: "2.5.0"`, `dev_mode.root_path: .` y `info.comments`. El icono se embebe desde `build/appicon.png` (`//go:embed build/appicon.png` en `internal/Handlers/App.go:18`) y `main.go:12` usa `application.New` + `AssetFileServerFS`.
- `wails.json` pasa a `D` en `git status` y `build/config.yml` a `??`, ambos reflejados en `git ls-files --others`.

### 2. Pipeline `Taskfile.yml` por plataforma

**Archivos nuevos**:

- `Taskfile.yml:1` raíz (includes `build/windows/Taskfile.yml`, `build/darwin/Taskfile.yml`, `build/linux/Taskfile.yml`, `build/Taskfile.yml`).
- `build/Taskfile.yml:1`, `build/windows/Taskfile.yml:1` (nsis `project.nsi` + `wails_tools.nsh` + `msix/app_manifest.xml`), `build/darwin/Taskfile.yml:1` (`Icons.icns`, `Info.plist`, `Assets.car`, `dmg-background.png`), `build/linux/Taskfile.yml:1` (`appimage/build.sh`, `desktop`, `nfpm/nfpm.yaml` + `scripts/*.sh`), `build/docker/Dockerfile.cross` y `Dockerfile.server`.
- `build/appicon.icon/icon.json` + `Assets/wails_icon_vector.svg` + `build/apptray.png` (tray fallback) + `build/config.yml`.

Decisión trazada en `Changelogs/Errors/StepLauncher-2.5.0/StepLauncher-Error-1.md:1`: `build/ios` eliminado definitivo (desktop-only Windows/macOS/Linux); su include no existe en `Taskfile.yml`.

### 3. Gestor `bun` y bindings Wails v3

- `frontend/.npmrc:1` nuevo (`registry` con `bun`), `frontend/bun.lock:1` (28 líneas diff vs `HEAD`), `frontend/package.json:8` cambia `frontend:install` de `npm install` a `bun install` y `frontend:build` a `bun run build`, elimina `package-lock.json` (`D` en git).
- `frontend/bindings/*` nuevo (287 paquetes): `frontend/bindings/StepLauncher/internal/{Config,Core/*,Handlers,Services/*,Music/*}` + `frontend/bindings/github.com/wailsapp/wails/v3/internal/eventcreate.ts`. Antes `frontend/wailsjs/go/main/App.d.ts` + `models.ts` + `runtime/*` (Wails v2) — ahora `D` en git, reemplazado por `frontend/bindings` generado con `wails3 generate bindings -ts -i` (ver `Changelogs/Changes/StepLauncher-2.5.0/StepLauncher-Change-1.md:12` y `Change-5.md:73` para `json:"-"` que elimina warnings).
- `frontend/tsconfig.app.json:5` y `vite.config.ts:8` con `alias @wailsjs -> ./bindings` y `wails('./bindings')` con ruta absoluta (`fileURLToPath(new URL('./bindings', import.meta.url))`), fix del bug `Event bindings module not found at import specifier './bindings/...'`.
- `frontend/web/public/wails/custom.js:1` nuevo para `AssetFileServerFS`.

### 4. `go.mod` / `go.sum` a Wails v3

- `go.mod:4` diff: elimina `wails/v2 v2.13.0`, `labstack/echo`, `go-toast`, `debounce`, etc.; añade `wails/v3 v3.0.0-beta.9`, `imaging v1.6.2`, `audiometa v0.10.0`, `x/mod v0.37.0`, `x/image v0.41.0`, `x/sync v0.21.0`, `x/sys v0.46.0`, `xdg v0.5.3`, `coder/websocket v1.8.14`.

## Por qué

- `wails.json` v2 no es compatible con `wails3 dev/build` (requiere `build/config.yml` v3). Mantener ambos rompe `wails3 generate bindings` y el `Taskfile` raíz (includes fallan si falta `build/<plataforma>/Taskfile.yml`, ver Error-1).
- `npm` + `package-lock.json` ralentiza CI y no coincide con el requisito del proyecto (`AGENTS.md:5` exige `bun install`/`bun run build`). `bun.lock` es determinista y `wails3` espera `bun`.
- `frontend/wailsjs` v2 no genera `Services` (9 servicios) y deja `App.d.ts` con 304 líneas obsoletas; `frontend/bindings` v3 expone 9 servicios + 161 métodos + 93 modelos sin warnings.

## API afectada

- Build: `build/config.yml`, `Taskfile.yml`, `build/*`, `frontend/.npmrc`, `frontend/bindings/*`.
- Go: `go.mod`/`go.sum` (v3).
- Frontend: `tsconfig.app.json`, `vite.config.ts`, `frontend/web/public/wails/custom.js`.

## Comportamiento anterior/nuevo

- `wails dev`/`wails build` con `wails.json` v2 → falla `GetFileAttributesEx ... build/ios/Taskfile.yml` y `frontend/wailsjs` desactualizado.
- Ahora `wails3 dev` levanta `frontend/web` vía `bun`, `wails3 generate bindings -dry` reporta `0 WARNINGS` y `wails3 build` produce `StepLauncher.exe` con `build/windows/info.json` (`comments` actualizado).

## Cómo verificar

- `go build ./...` (raíz) — EXIT 0 sin `build/ios`.
- `wails3 generate bindings -dry` — `Processed: 287 Packages, 9 Services, 161 Methods, 5 Enums, 93 Models, 0 WARNINGS`.
- `bun run build` en `frontend/` — `✓ built in ~49s` (6582 módulos, sin `vue-tsc` errors).
