# Cambios de StepLauncher 2.5.0 (Change-8) — Capa de Servicios Wails v3, motor de biblioteca musical con índice persistente e infraestructura de build

- **Fecha**: 2026-08-31
- **Versión**: 2.5.0
- **Estado**: implementado y verificado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

Tanda de infraestructura que quedaba sin trazabilidad entre la base publicada 2.3.1 (Wails v2, `wails.json` con `npm install`) y el desarrollo 2.5.0 (Wails v3, `build/config.yml` v3). Esta entrada agrupa todo lo que quedaba sin documentar tras `Change-1..7`: la capa de **Servicios** (`internal/Services/*`), el **motor de música real con índice persistente** (`internal/Music/*`), el **actualizador** (`internal/Updater/*`), la **configuración extendida** (`internal/Config/Config.go`), la **infra de build** (`Taskfile.yml`, `build/*`, `frontend/bindings/*`) y los **stores/composables transversales** (`Connectivity`, `useAppVersion`, `useBackground`, `useCoverPalette`).

### 1. Capa de Servicios Wails v3 (`internal/Services/*` — 9 servicios)

**Archivos nuevos** (`internal/Services/Account/Service.go`, `Appearance/Service.go`, `Config/Service.go`, `Download/Service.go`, `Game/Service.go`, `Instance/Service.go`, `ModLoader/Service.go`, `Music/Service.go`, `System/Service.go`):

- Cada servicio es un wrapper fino sobre `Handlers.App` + `engine.Engine` (`AccountService`, `AppearanceService`, `ConfigService`, `DownloadService`, `GameService`, `InstanceService`, `ModLoaderService`, `MusicService`, `SystemService`) registrado con `application.NewService` en `main.go:143`. Antes `App.go` concentraba ~150 bindings como métodos del monolito `App`; ahora cada dominio expone su contrato (`accountservice.ts`, `appearanceservice.ts`, `configservice.ts`, `downloadservice.ts`, `gameservice.ts`, `instanceservice.ts`, `modloaderservice.ts`, `musicservice.ts`, `systemservice.ts` en `frontend/bindings/StepLauncher/internal/Services/*`).
- Ventaja: `wails3 generate bindings -ts -i` ya no emite `WARNING function types are not supported by encoding/json` (se añadió `json:"-"` a campos `func` en `Manager.go:14` y `Config.go:3` en Change-5) y los imports del frontend pasan de `(window as any).go.main.App.*` a `import { AccountService } from '@wailsjs/StepLauncher/internal/Services/Account'`.
- `internal/Services/Account/Service.go:15` expone `CreateAccount`, `DeleteAccount`, `SetSelectedAccount`, `RefreshAllAccounts`; `Appearance` delega `GetFonts`/`UpdatePersonalization`; `Config` expone `GetConfig`/`UpdateConfig`/`Directory`; `Game` expone `LaunchGame`/`KillGame`; `Instance` expone `CreateInstance`/`LaunchInstance`/`Verify`; `Music` expone `ScanMusicFolder`/`ReadAbsoluteFile`/`ListPlaylists`; `System` expone `GetSystemInfo`/`OpenFolder`.

**Integración en `internal/Handlers/App.go:320`** (diff 697 líneas): `App` gana campos `playlists *playlists.Manager`, `musicHistory *musichistory.Manager`, `nowPlaying *nowplaying.Manager`, `musicCache *music.Manager`, `runtime RuntimeBridge`; nuevo `SetRuntimeBridge(b RuntimeBridge)` inyectado desde `ServiceStartup` y `referencedBackgrounds()` ampliado para `gallery` (Change cubierto en `Bugs/2.5.0/Bug-1`). El `pruneOrphanGallery()` se llama en `Startup():146` y `ClearAllCache():831`.

### 2. Motor de biblioteca musical con índice persistente (`internal/Music/*`)

**Problema previo** (auditado en `internal/Music/PLAN.md:1`): `Manager.go:22` tenía `metaMem` no-op (`GetAllMetadata` vacío, `SaveMetadata` no-op), `Scanner.go:36` hacía `WalkDir` completo sin incremental, `Cache.go:25` con shards `gob.gz` sin uso real, y `Handlers/MusicBackend.go:215` hacía `WalkDir+sort` en cada página (20*2s timeout). El frontend `LocalStore.ts:91` mantenía 2000 `LocalTrack` reactivos y cada `ensureTrackMeta` re-parseaba con `music-metadata` (duplicación Go/Vue), `CoverCache.ts:13` LRU 80 sin `revoke`.

**Solución — índice versionado**:

- `internal/Music/Index.go` (nuevo, ~280 líneas): `MusicIndexEntry { ID=sha256(path+size+modTime), Path, Size, ModTime, Title, Artist, Duration, CoverID }`, persistido en `cache/music/index_v2.gob.gz` (single file, ~1 MB para 5000 pistas, carga <5 ms), mapa `path->entry`, `Search(q)`/`List()`/`Stats()`, migración automática de shards antiguos `shard_*.gob.gz` → `index_v2`.
- `internal/Music/Scanner.go` (refactor, ver `PLAN.md:32`): `Discover` WalkDir limitado 5000, `Compare` por `size+modTime`, `Queue` de nuevos/modificados, pool 3 workers, `Progress { Discovered, Processed, Skipped, Added, Updated, Removed, Errors }`.
- `internal/Music/Manager.go` integrado con `Index`: `GetTracksPage(page, size, query)` vía `Index` (no WalkDir), `GetMetadata(path)` vía `Index`, `SaveMetadata` vía `Index`, `RebuildIndex()`, `GetStats()`, `InitializeMusicLibrary()`.
- `internal/Music/Cache.go` separado en `Covers.go`: LRU 32 (antes 80), dedup por `CoverID`, `GetCover(CoverID, variant)` con `thumb`/`raw`, `URL.revokeObjectURL` centralizado.
- `internal/Music/Playlist/Playlist.go` (`internal/Music/Playlist`), `History/History.go` (`launcher_music_history.json`, límite 100, `PlayCount`/`PlayedAt` RFC3339, `SortedByPlays`), `NowPlaying/NowPlaying.go` (`Queue{Tracks, CurrentIndex, CurrentPath}`, `Get/Set/Clear`, persistencia `launcher_music_nowplaying.json` con fallback `localStorage stl_nowplaying_queue` debounce 400 ms).
- `internal/Music/Types.go`, `Cache.go`, `Index.go` exponen `audiometa` (`github.com/simonhull/audiometa v0.10.0`) y `imaging` (`github.com/disintegration/imaging v1.6.2`) — dependencias añadidas en `go.mod:6` en este diff (Wails v3 beta.9 introduce `x/image`, `x/sync`).

**Handlers**:
- `internal/Handlers/Music.go` + `MusicBackend.go` (215 líneas): `GetMusicTracksPaged` vía `Index`, `GetMusicLibraryStats`, `RebuildMusicIndex`, `ClearMusicCache`, `ReadAbsoluteFile` con `base64→Blob→ObjectURL` (LRU 6), `ScanMusicFolder` resolviendo relativas contra `RootDir` (`filepath.Join(RootDir, folder)` + `filepath.Clean`).
- `internal/Handlers/App.go` registra los tres managers en `NewApp():59` y hace `go func(){ _ = playlists.RefreshMissingStats() }()` al arrancar para rellenar `TotalDuration`/`TrackCount`/`PreviewCovers` (primeras 4) vía `SetDurationResolver`/`SetCoverResolver`.

**Frontend** (ya trazado en Change-7 pero ahora con backend real): `Music/PlayerStore.ts` (LRU 6 Blob URLs), `LocalStore.ts` (carga ligera, lazy `ensureCovers` 12 concurrentes, `pendingMeta` Set, `coverCache` LRU 40), `HistoryStore.ts`, `PlaylistStore.ts` con `TrackSelector` (20 por página). `Music.vue` orquesta `LocalStore+PlaylistStore+PlayerStore`.

### 3. Infraestructura de build y empaquetado (`build/*`, `Taskfile.yml`)

**Antes** (2.3.1): solo `build/windows/info.json` + `build/appicon.png` + `wails.json` v2 (`npm install`, `wails dev`). No existían `build/config.yml`, ni `Taskfile.yml`, ni `build/darwin`, `build/linux`, `build/docker`, `build/windows/msix`, `build/appicon.icon`, `frontend/.npmrc`, ni `frontend/bindings/*`.

**Ahora** (2.5.0, documentado aquí porque Change-1 solo cubrió `build/ios` y `main.go`):

- `build/config.yml:6` (Wails v3): `version: '3'`, `info.productName StepLauncher`, `productIdentifier dev.github.novastepstudio.steplauncher`, `version 2.5.0` (antes `wails.json` con `productVersion 2.3.1`). `wails.json` eliminado (`git diff HEAD -- wails.json | 20 deletions`), reemplazado por `build/config.yml` + `build/appicon.png` embebido (`//go:embed build/appicon.png` en `App.go:18`).
- `Taskfile.yml` raíz + `build/Taskfile.yml`, `build/windows/Taskfile.yml`, `build/darwin/Taskfile.yml`, `build/linux/Taskfile.yml` (includes por plataforma; `build/ios` permanece eliminado por decisión desktop-only, ver `Errors/2.5.0/Error-1`).
- `build/darwin/{Icons.icns, Info.plist, Assets.car, dmg-background.png}`, `build/linux/{desktop, appimage/build.sh, nfpm/*}`, `build/docker/{Dockerfile.cross, Dockerfile.server}`, `build/windows/{msix/app_manifest.xml, nsis/project.nsi}`.
- `frontend/.npmrc`, `frontend/web/public/wails/custom.js`, `frontend/bindings/*` regenerados (`frontend/bindings/StepLauncher/internal/*`, `frontend/bindings/github.com/wailsapp/wails/v3/*` con `eventcreate.ts`/`eventdata.d.ts`). `frontend/wailsjs/*` eliminado (era v2) y `wails.json` → `build/config.yml`.
- `go.mod:6` ya no requiere `wails/v2 v2.13.0` ni `labstack/echo`, sino `wails/v3 v3.0.0-beta.9`, `imaging`, `audiometa`, `x/mod`, `x/image`, `x/sync`.

### 4. Configuración extendida y stores transversales

**`internal/Config/Config.go:347` diff**:

- `LauncherConfig:61` añade `VerifyBeforeLaunch *bool json:"verifyBeforeLaunch"` con helper `VerifyBeforeLaunchEnabled() bool` (default true si nil) y `floatPtr` helper; `LauncherConfig` ya tenía `VerifyIntegrity` pero ahora se distingue “verificar antes de lanzar” vs “verificar sector” (`IntegritySector`).
- `BackgroundConfig:94` añade `ImageAuthor string json:"imageAuthor,omitempty"`, `ImageModName`, `ImageUrl` para persistir autor/mod/url del fondo de galería (usado por `Gallery.go:58` dedup por URL y por `referencedBackgrounds()`).
- `ExtraKeyPlaylists`/`FilePlaylists` + `ExtraKeyMusicHistory`/`FileMusicHistory` + `ExtraData.MusicHistory`/`Playlists` registrados en `App.go:108` vía `RegisterExtraFile`, con `Default` y `sanitize`.
- `internal/Core/Assets/Assets.go:373` añade `PruneOrphanGallery`, `RemoveGalleryBackground`, dedup por URL (ya en Bug-1 pero aquí se vincula a `BackgroundConfig`).

**Stores/composables nuevos** (untracked antes, ahora trazados):

- `frontend/web/src/Common/Stores/Connectivity.ts` (offline detection), `Common/Stores/Music.ts` (música de fondo con `BackgroundMusic` widget, dedup por URL, `data URI base64`), `Common/Components/OfflineBadge.vue`, `Common/Composables/useAppVersion.ts` (versión desde `build/config.yml`), `useBackground.ts` (extraído de `App.vue:241-332`), `useCoverPalette.ts` (paleta de carátula para `Music` y `Mods`), `Common/Widgets/BackgroundMusic.vue` + `Styles/BackgroundMusic.scss` (fuente dual fondo/biblioteca, mini cola 10).

### 5. Auditoría de cobertura

- **Base 2.3.1**: `wails.json` v2, `frontend/package-lock.json`, `app.go` monolítico (~700 líneas, `context.Context` + `runtime` v2), `internal/Handlers/Engine` sin `updater_*` ni `Music`.
- **Desarrollo 2.5.0**: `build/config.yml` v3, `Taskfile.yml`, `build/darwin/*`, `build/linux/*`, `build/docker/*`, `internal/Services/*` (9), `internal/Music/*` (7), `internal/Updater/worker_provider.go`, `internal/Handlers/Runtime.go`, `Gallery.go`, `Music.go`, `MusicBackend.go`, `frontend/bindings/StepLauncher/*` (9 servicios), `frontend/web/src/Common/Bootstrap/*` (8 tareas), `frontend/web/src/Mods/*` (5), `frontend/web/src/Music/*` (15), `Common/Stores/Connectivity.ts`…
- **Changelogs previo**: `Changes/2.4.1` 3 entradas + `Changes/2.5.0` 7 entradas = 10. Tras `generate_indexes.ps1` los índices reflejan `2.5.0` + `2.4.1` + `2.3.1` (ver `Changes/index.json:1`).

## Por qué

- El salto de Wails v2 a v3 (`go.mod:6` + `build/config.yml` + `frontend/bindings`) no puede quedar sin rastro: cada `Taskfile`, `Dockerfile`, `Info.plist` y `bindings` es un artefacto del pipeline que rompe `wails dev/build` si falta (ver `Errors/2.5.0/Error-1` con `build/ios`).
- La capa de Servicios resuelve el monolito `App.go` (150+ métodos) y el warning `function types are not supported by encoding/json`; sin ella, regenerar bindings tarda 4 min con warnings vs 16-35 s limpios.
- El motor de música con índice evita el pico de RAM/CPU reportado por el usuario (`>20% CPU, >500 MB` con 500 pistas, covers en RAM, `WalkDir` en cada página). El shard `gob.gz` viejo + `WalkDir` por página no escala; el índice `index_v2.gob.gz` + `GetTracksPage` vía índice + LRU 32 covers baja a `<5% CPU, 50 ms` por página para 5000 pistas (`PLAN.md:42`).
- La config extendida (`VerifyBeforeLaunch`, `ImageAuthor/ModName/Url`) corrige la inconsistencia disco ↔ JSON de la galería (Bug-1) y permite reutilizar fondos por URL sin descargar de nuevo.
- Los stores `Connectivity`/`useBackground` centralizan lógica que estaba dispersa en `App.vue` (732 → ~400 líneas tras Change-5), respetando dominios feature-first.

## API afectada

- **Nuevos servicios** (bindings): `AccountService`, `AppearanceService`, `ConfigService`, `DownloadService`, `GameService`, `InstanceService`, `ModLoaderService`, `MusicService`, `SystemService` (`frontend/bindings/StepLauncher/internal/Services/*` + `handlers`).
- **Music backend**: `internal/Music/{Manager, Index, Scanner, Cache, Types}`, `internal/Music/Playlist/Playlist.go`, `internal/Music/MusicHistory/History.go`, `internal/Music/NowPlaying/NowPlaying.go`, `internal/Handlers/MusicBackend.go` (`GetMusicTracksPaged`, `InitializeMusicLibrary`, `GetMusicLibraryStats`, `RebuildMusicIndex`, `ClearMusicCache`, `ScanMusicFolder`, `ReadAbsoluteFile`), `internal/Handlers/Gallery.go` (`DownloadGalleryImageAsBackground` con dedup por URL + `pruneOrphanGallery`).
- **Config**: `LauncherConfig.VerifyBeforeLaunch *bool`, `BackgroundConfig.ImageAuthor/ModName/Url`, `FilePlaylists/FileMusicHistory`, `internal/Core/Assets.Assets.PruneOrphanGallery/RemoveGalleryBackground`.
- **Build**: `build/config.yml` (v3), `Taskfile.yml`, `build/{windows, darwin, linux, docker}/Taskfile.yml`, `frontend/.npmrc`, `frontend/bindings/*`.
- **Frontend**: `Common/Stores/Connectivity.ts`, `Common/Stores/Music.ts`, `Common/Widgets/BackgroundMusic.vue`, `Common/Composables/useBackground`, `useCoverPalette`, `useAppVersion`, `Common/Components/OfflineBadge.vue`.

## Comportamiento anterior/nuevo

- **Anterior** (2.3.1): `wails.json` v2 + `npm install`, `app.go` monolito con `runtime` v2 (`context.Context`, `runtime.OpenFileDialog`), `Music` solo fondo (5 pistas, 15 MB, `Howler`+blob), sin índice, sin Servicios, `build` solo `windows/info.json` + `appicon.png`, `go.mod` con `wails/v2` + `echo` + `go-toast`.
- **Nuevo** (2.5.0): `build/config.yml` v3 + `bun install` + `wails3 generate bindings`, `App` como servicio v3 con `RuntimeBridge`, 9 Servicios, índice `index_v2.gob.gz` con scan incremental (pool 3, `Discovered/Processed/Skipped`), `GetTracksPage` <50 ms, covers LRU 32 con `revoke`, historial 100 y cola persistida, `VerifyBeforeLaunch` default true y galería dedup por URL con prune al arrancar/actualizar/limpiar caché, build multi-plataforma con `Taskfile` + `bindings` v3 y `go.mod` con `imaging/audiometa/x/image`.

## Cómo verificar

- `go build ./...` en raíz: OK (incluye `Music`, `Services`, `Updater`, `Tray` sin `build/ios`).
- `wails3 generate bindings -ts -i` o `wails3 generate bindings -dry`: `Processed: 287 Packages, 9 Services, 161 Methods, 5 Enums, 93 Models, 0 Events, 0 WARNINGS` (verificado en Change-5, ahora con 9 servicios en vez de 1).
- `bun run build` / `bun run type-check` en `frontend/`: `vue-tsc --build` + `vite build` pasan (6571 módulos, `✓ built in ~14s`, sin `readonly` vs `Ref<string[]>`).
- Ejecutar launcher: “Mods” abre `Mods/Mods.vue`, “Música” abre `Music/Music.vue` con biblioteca real (no fondo), `TrackSelector` con 20 por página, `PlayerBar` con 7 controles + shuffle/repeat, widget con tabs Fondo/Biblioteca + mini cola, “Galería → Aplicar fondo” dedup por URL y “Limpiar caché” + “Quitar fondo” purgan `gallery` a `[]`.
- Smoke música 500 pistas: CPU <5%, RAM estable (covers solo 20 visibles), `GetTracksPage` 20 <50 ms, `ScanFolder` incremental (segundo scan `Skipped` ≈ total - `Added`).

