# Cambios de StepLauncher 2.5.0 (Change-1) — Migración completa de Wails v2 a Wails v3

- **Fecha**:   2026-08-16
- **Versión**: 2.5.0 (en desarrollo)
- **Estado**:  implementado y verificado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

### 1. Backend Go migrado a Wails v3

- `go.mod` reescrito: módulo `StepLauncher`, `go 1.26.4`, dependencias de v3 (`github.com/wailsapp/wails/v3 v3.0.0-beta.9`, `go-winio v0.6.2`, `golang.org/x/sys v0.46.0`) y `go mod tidy` sin errores.
- Eliminados los artefactos de la plantilla v2 (`greetservice.go`, `frontend/bindings` viejo).
- `main.go` raíz reescrito: `application.New` (Name, AssetOptions con `AssetFileServerFS` del frontend embebido), `app.RegisterService(application.NewService(NewApp(app)))`, ventana 1024×600 (mín. 950×600, fondo negro, URL `/`) y `app.Run()`.
- `App.go` raíz: puerto completo de `OLD/app.go` como servicio v3 (≈150 bindings) con `ServiceStartup`/`ServiceShutdown`, `runtimeBridge` no exportado (no se filtra como binding) y tipos de resultado v3 (`CreateInstanceResult`, `GetInstanceResult`, `AddInstanceVersionResult`, `HistoryStatsResult`, `InstanceDownloadStatusResult`, `RecommendedRAMResult`).
- `internal/Handlers/Runtime.go`: nueva interfaz `RuntimeBridge` (`OpenFileDialog`, `OpenDirectoryDialog`, `BrowserOpenURL`, `Quit`) + `FileFilter`. Migrados `Handlers/App.go`, `Music.go`, `Background.go` y `Assets.go` fuera de la API de wails v2 (el contexto `runtime` de v2 ya no existe).
- `wails3 generate bindings -ts -i` genera los bindings en `frontend/bindings/StepLauncher/` (1 servicio, 153 métodos, 5 enums, 92 modelos). Advertencia conocida: `package internal/Core/Launcher: function types are not supported by encoding/json` (tipos con campos func no se serializan; no bloquea la generación).
- `build/ios` fue eliminado por completo (carpeta + include del `Taskfile.yml`) porque el proyecto es **desktop-only (Windows/macOS/Linux)** por decisión del owner (ver `Changelogs/Errors/StepLauncher-2.5.0/StepLauncher-Error-1.md`). Inicialmente se restauró con un stub (`build_stub.go`) para que `go build ./...` pasara, pero la decisión final eliminó el soporte iOS.

### 2. Frontend migrado a Wails v3

- `bun add @wailsio/runtime` (`@wailsio/runtime@3.0.0-beta.9`, coincide con la versión del backend).
- `vite.config.ts` + `tsconfig.app.json`: plugin `wails('./bindings')` y alias `@wailsjs` → `./bindings` (en v3 no existe `@wailsjs/go` ni `@wailsjs/runtime`). El plugin recibe la ruta **absoluta** a `frontend/bindings` (`fileURLToPath(new URL('./bindings', import.meta.url))`): con la ruta relativa, `vite build` la resolvía contra el cwd (`frontend`) pero el dev server la resolvía contra el root del proyecto (`frontend/web`) y fallaba con `Event bindings module not found at import specifier './bindings/github.com/wailsapp/wails/v3/internal/eventcreate'`.
- Todos los stores migrados: `Launcher/Store.ts`, `Accounts/Store.ts`, `News/Store.ts`, `Downloads/Store.ts`, `Updates/Store.ts`, `Instances/Store.ts`, `Common/Stores/Ui.ts`, `Common/Stores/Music.ts`, `Common/Stores/Idle.ts`.
- Todos los componentes migrados: `App.vue`, `News/News.vue`, `Screenshots/Screenshots.vue`, `Instances/{Form,Detail,Download}.vue`, `Welcome/Welcome.vue`, `Downloads/Installation.vue`, `Accounts/Content.vue`, `Settings/Settings.vue`, `Settings/Sections/{General,Personalization,Minecraft,About}.vue`, `Settings/FontManager.vue`, `Crash/Crash.vue`, `Versions/ProfileForm.vue`.
- Mapeo v2 → v3: `EventsOn(nombre, cb)` → `Events.On(nombre, (ev) => cb(ev.data))`; `WindowHide/WindowShow` → `Window.Hide()/Show()`; `BrowserOpenURL` → `Browser.OpenURL(url)`; `(window as any).go.main.App...` → imports estáticos de bindings.
- Casts en las fronteras por tipos generados más estrictos: `[]byte` → `string` (base64), slices → `T[] | null`, slices de punteros → `(T | null)[] | null`, enums Go → enums TS (no string unions), `fonts: FontSlot[] | null`.

### 3. Gestor de paquetes bun

- `Taskfile.yml`: `PACKAGE_MANAGER` por defecto `"bun"` (antes `npm`). Todos los task de build/dev del frontend usan bun.
- `frontend/package.json`: añadido el script `build:dev` (`vite build`) requerido por el Taskfile de Wails cuando `DEV=true`.
- `wails3 build` (producción y `DEV=true`) compilan correctamente: `go mod tidy` → `bun install` → bindings → frontend → syso → binario.

### 4. Metadatos de build

- `build/config.yml`: `productName: StepLauncher`, `productIdentifier: com.novastepstudio.steplauncher`, versión `2.5.0`, empresa/copyright/comentarios de NovaStepStudio.

### 5. Eliminado `wails.json`

- `wails.json` era el archivo de configuración de Wails v2 y no existe en v3 (la configuración vive en `build/config.yml`). `About.vue` y `Settings.vue` lo importaban para mostrar nombre/versión; ahora usan constantes (`StepLauncher` / `2.5.0`). El archivo se eliminó de la raíz.
- Nota: el `Changelogs/README.md` menciona `productVersion de wails.json` como fuente de la versión; en v3 la fuente es `build/config.yml`.

## Por qué

StepLauncher usaba Wails v2 (>= 80 archivos acoplados). La migración a Wails v3 (beta.9) era necesaria para mantener el framework al día: v3 cambia por completo la API del backend (servicios en vez de `runtime` global), el sistema de bindings y el runtime del frontend (`@wailsio/runtime`), y el CLI (`wails3` + Taskfile). Además el usuario pidió explícitamente usar **bun** en lugar de npm por seguridad.

## API afectada

- Backend: todas las llamadas `runtime.*` de v2 eliminadas; bindings expuestos por el servicio `StepLauncher.App` (mismo conjunto de métodos que v2).
- Frontend: `@wailsjs/go` y `@wailsjs/runtime` ya no existen; todo importa de `@wailsjs/StepLauncher/...` (estático) y de `@wailsio/runtime` (Events, Window, Browser).
- Eventos del backend a frontend: se conservan los mismos nombres (`game_prepare`, `game_exited`, `account_assets`, `account_refresh`, `download_*`, `install_*`, etc.) con el nuevo suscriptor `Events.On`.

## Comportamiento anterior/nuevo

- **Antes**: `wails.json` en la raíz (config v2), bindings en `frontend/wailsjs`, frontend con `(window as any).go` / `(window as any).runtime`, npm como gestor.
- **Ahora**: `build/config.yml` (config v3), bindings en `frontend/bindings`, frontend con imports estáticos de `@wailsjs/StepLauncher` + `@wailsio/runtime`, bun como gestor.

## Cómo verificar

- `go build ./...` en la raíz — pasa (incluye `build/ios` con el stub).
- `bun run build` (y `bun run type-check`) en `frontend/` — pasan sin errores TS.
- `wails3 build DEV=true` — pipeline completo OK (bun install, bindings, frontend dev, syso, binario).
- `wails3 build` (producción) — OK.
- `wails3 dev` — el error de `build/ios/Taskfile.yml` desapareció; la app arranca.
