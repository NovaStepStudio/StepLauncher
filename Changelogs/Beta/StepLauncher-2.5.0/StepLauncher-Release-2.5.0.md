# Beta 2.5.0 — StepLauncher (Wails v3 + Música real + Servicios)

- **Fecha**: 2026-08-31
- **Versión**: 2.5.0-beta
- **Estado**: beta

## Resumen

Beta de la actualización más grande desde la 2.3.1: **migración completa a Wails v3** (`go 1.26.4`, `build/config.yml`, `Taskfile` multiplataforma, `frontend/bindings` con 9 servicios), **panel de Mods** (Modrinth solo cliente), **sistema Bootstrap** con splash trazable y 8 fases, **Ajustes reorganizados** para usuario promedio (Red/Descargas/Integridad/Almacenamiento), **panel de Música reconstruido** con biblioteca real, selector visual y `PlayerStore`, **capa de Servicios desacoplada**, **índice persistente `index_v2.gob.gz`** para bibliotecas de 5000 pistas, **infra de build** y **configuración extendida** (`VerifyBeforeLaunch`, galería dedup por URL).

> Esta beta consolida **11 cambios**, **4 errores** y **2 bugs** de la rama 2.5.0. Es para pruebas internas: puede tener regresiones menores y el índice musical se valida con bibliotecas grandes antes de la final.

---

## Funcionalidades nuevas

### 1. Migración completa a Wails v3
- Backend `wails/v3 v3.0.0-beta.9` con `application.New` + `RegisterService`, `App` como servicio (`ServiceStartup`/`Shutdown`, `RuntimeBridge` para diálogos), frontend `@wailsio/runtime` con alias `@wailsjs -> ./bindings` en ruta absoluta.
- `build/ios` eliminado definitivo (desktop-only Windows/macOS/Linux, ver `Errors/2.5.0/Error-1`). Ventana 1024×600 (mín. 950×600) con `AssetFileServerFS`.
- `wails3 generate bindings` ahora reporta `0 WARNINGS` (vs 4m con warnings) tras añadir `json:"-"` a campos `func`.

### 2. Música de fondo fiable
- Vuelve a `<audio>` nativo + data URI base64 (los blob URLs en `wails://` quedaban mudos), caché de 2 pistas, reintento `canplay` 800 ms y estado `loadError` visible.
- Volumen en Ajustes (`Personalization.backgroundMusic.volume` default 0.8) con `watch` en `Music.ts`, `coverStyle: background` con carátula doble (fondo + widget) y validación con `<audio>` temporal (`measureAudioDuration`).

### 3. Tray renovado — un solo icono
- Un solo icono en bandeja (fix de `Run()` doble), clic izquierdo/derecho/doble, menús “Últimas Sesiones” e “Instancias” clicables, cuenta con radio items y eventos `tray_open_settings`/`instances`/`downloads`/`tray_launch_version`.

### 4. Panel de Mods (maqueta Modrinth)
- Dominio `Mods/` con `Store.ts` (Labrinth v2, `client_only`/`client_and_server` etc., `GET /search` con facets AND/OR), `Mods.vue` overlay a pantalla completa, tabs Todos/Mods/Modpacks/Shaders/Texturas y card de ejemplo con chips de entorno.

### 5. Sistema Bootstrap + SplashScreen
- `Common/Bootstrap/` con `types.ts`/`state.ts`/`runner.ts`/`logger.ts` + 8 tareas ponderadas (`welcome`, `config`, `appearance`, `system`, `accounts`, `versions`, `events`, `background`) con progreso 0-100% y logs.
- `SplashScreen.vue` con logo glow, barra shimmer, 8 `StepDot` y panel de logs expandible. `Main.ts` orquesta `runBootstrap({minSplashMs: 950, maxSplashMs: 6500})`, `App.vue` aligerado -140 líneas y `useBackground` extraído.

### 6. Ajustes reorganizados para usuario promedio
- `General.vue` de 964 → 398 líneas, 4 secciones nuevas: `Network.vue` (proxy), `Downloads.vue` (hilos/límite), `Integrity.vue` (auto vs manual, `VerifyBeforeLaunch`), `Storage.vue` (caché). Lenguaje sin jerga, `About.vue` con 5 librerías de terceros + 10 internas con `Abrir sitio`.

### 7. Panel de Música reconstruido — música REAL
- `PlayerStore.ts` con Blob URLs LRU 6, `TrackSelector.vue` visual (búsqueda, paginación 20, carátulas), `PlaylistForm` con `TrackSelector` embebido y validación, `Music.vue` orquesta `LocalStore+PlaylistStore+PlayerStore`.
- 5 vistas autocontenidas: `MusicView` (búsqueda + 20/50/100 + paginación), `LibraryView` (hero + grid 2×2 de carátulas), `MenuView` (stats sin “Offline”), `NowPlaying` (cover huge + cola paginada), `Cola` (20 por página). SCSS completados (antes vacíos) con `backdrop-filter` y paletas `useCoverPalette`.
- Historial `launcher_music_history.json` (100, `PlayCount`/`PlayedAt`), `NowPlaying` `launcher_music_nowplaying.json` con cola persistida, widget dual Fondo/Biblioteca con mini cola.

### 8. Capa de Servicios desacoplada
- 9 servicios `internal/Services/*` (`Account`, `Appearance`, `Config`, `Download`, `Game`, `Instance`, `ModLoader`, `Music`, `System`) registrados con `application.NewService`, cada uno con su `frontend/bindings/StepLauncher/internal/Services/*`. `App.go` ya no concentra 150 métodos; `RuntimeBridge` inyectado.

### 9. Motor de biblioteca musical con índice persistente
- `internal/Music/Index.go` con `index_v2.gob.gz` (~1 MB para 5000 pistas, <5 ms), migración de shards antiguos. `Scanner.go` con `Discover` → `Compare` (`size+modTime`) → pool 3 workers → `Progress`.
- `Manager.go` con `GetTracksPage` vía índice (no `WalkDir`), `Cache.go` → `Covers.go` LRU 32 con dedup `CoverID` y `URL.revokeObjectURL`. Frontend `LocalStore` carga ligera (solo 20 visibles) + `pendingMeta` 12 concurrentes, CPU <5% vs >20% antes.

### 10. Infra de build Wails v3
- `wails.json` (v2, `npm`) → `build/config.yml` v3 (`version: '3'`, `productIdentifier`), `Taskfile.yml` por plataforma (`darwin` `Icons.icns`/`Info.plist`, `linux` `appimage`/`nfpm`, `docker` `Dockerfile.cross`, `windows` `msix`/`nsis`), `frontend/.npmrc` + `bun.lock` + `frontend/bindings/*` (287 paquetes).

### 11. Configuración extendida y stores transversales
- `VerifyBeforeLaunch *bool` (default true) + `BackgroundConfig.ImageAuthor/ModName/Url` + `PruneOrphanGallery`/`RemoveGalleryBackground` con dedup por URL y `Gallery.go` reutilización.
- `Connectivity.ts` (`isOffline`), `OfflineBadge.vue`, `useBackground.ts` (extraído de `App.vue`), `useCoverPalette.ts` (canvas 1×1), `useAppVersion.ts` (`v2.5.0` desde `build/config.yml`), `BackgroundMusic.vue` dual con `Bgm_Queue`.

---

## Errores de la auditoría resueltos en esta versión

- [StepLauncher-Error-1: `wails3 dev/build` rotos por eliminación de `build/ios`](../../Errors/StepLauncher-2.5.0/StepLauncher-Error-1.md) — corregido.
- [StepLauncher-Error-2: Tray nunca aparecía — `Setup` nunca cableado y `Run()` no-op](../../Errors/StepLauncher-2.5.0/StepLauncher-Error-2.md) — corregido.
- [StepLauncher-Error-3: Dos iconos de tray — `Run()` dos veces](../../Errors/StepLauncher-2.5.0/StepLauncher-Error-3.md) — corregido.
- [StepLauncher-Error-4: Verificación antes de lanzar ignorada y fondos duplicados por `UnixNano`](../../Errors/StepLauncher-2.5.0/StepLauncher-Error-4.md) — corregido.

---

## Correcciones y bugs resueltos

### Música y descargas
- **Fondos de galería con crecimiento infinito**: 22 entradas en `launcher_assets.json.gallery` con misma URL pero distinto `UnixNano` → ahora dedup por URL, `PruneOrphanGallery` al arrancar/actualizar/limpiar caché y reutilización sin descargar.
- **Widget de descargas con textos largos**: `Widget.scss` con `max-width` + `ellipsis` y `overflow: hidden` para slugs largos.
- **Carátulas sin `revoke` con pico RAM >500 MB**: LRU 32 (antes 80) + `revoke` en `CoverCache`/`PlayerStore`, `LocalStore` solo 20 visibles, `GetTracksPage` <50 ms para 5000 pistas.
- **Música de fondo muda en `wails://`**: `Howler` + blob → `<audio>` + data URI base64 con caché 2 pistas.

### Tray y Bootstrap
- **Tray no aparecía / dos iconos**: `Setup(icon)` no se llamaba + `Run()` no-op en `ServiceStartup` (guard `running==false`) + doble `Run()` vía `pendingRun` y `ApplicationStarted`. Ahora `New()` difiere solo, `Setup` solo configura.

### Configuración
- **VerifyBeforeLaunch ignorado**: `Manager.go` leía `VerifyIntegrity` genérico; ahora `VerifyBeforeLaunchEnabled()` default true y sectores `IntegritySector`.

---

## Cambios técnicos (para curiosos y desarrolladores)

- **9 servicios** desacoplados: `AccountService`, `AppearanceService`, `ConfigService`, `DownloadService`, `GameService`, `InstanceService`, `ModLoaderService`, `MusicService`, `SystemService` (`frontend/bindings/StepLauncher/internal/Services/*`).
- **Índice musical** `cache/music/index_v2.gob.gz` (`ID=sha256(path+size+modTime)`) con `Search`/`Stats` y scanner pool 3, `go.mod` añade `imaging v1.6.2` + `audiometa v0.10.0` + `x/image`/`x/sync`/`x/mod`.
- **Build** `build/config.yml` v3 + `Taskfile.yml` por plataforma + `frontend/bindings` (287 paquetes, 9 servicios, 161 métodos, 93 modelos, 0 warnings).
- **Bootstrap** `Common/Bootstrap/` con `runner` ponderado (`weight` suma 10, `minSplashMs:950`/`maxSplashMs:6500`) y `SplashScreen` con `StepDot` y `BootstrapSplash_Logs`.
- **Frontend** por dominios: `Mods/`, `Music/`, `Common/Bootstrap/`, `Common/Stores/Connectivity`, `Common/Composables/useCoverPalette` etc., estilos con `@use` y `var(--color-*)`.

---

## Qué significa para el usuario

- **No necesitas hacer nada manual**: tu `launcher_config.json`, `launcher_assets.json` (`gallery` purgada al arrancar), `launcher_music_history.json` y `index_v2.gob.gz` se migran solos. Los shards antiguos `shard_*.gob.gz` se migran a `index_v2`.
- **Música**: al abrir Música se escanea `musicFolder` (mp3/wav/ogg/m4a/flac, máx. 2000) sin pico de RAM; las carátulas solo de las 20 visibles. Usa “Escanear” solo si añades música nueva.
- **Mods**: el botón “Mods” ya abre el panel a pantalla completa (por ahora maqueta con card de ejemplo, la carga real llegará en la siguiente beta).
- **Tray**: un solo icono, clics y menús funcionales; la música se controla desde el tray con “Pausar/Reproducir”.

---

## Cómo actualizar a esta beta

- **Desde 2.3.1**: descarga `StepLauncher-2.5.0-beta` desde `Changelogs/Beta/StepLauncher-2.5.0/` y reemplaza el ejecutable; `directory.json` fuera del workdir conserva tu directorio.
- **Validación beta**: probar con 500+ pistas (`CPU <5%`, `RAM estable`), aplicar dos veces “Aplicar fondo” misma URL (debe reutilizar), quitar fondo (debe quedar `gallery: []`), verificar `VerifyBeforeLaunch` en Ajustes > Integridad.
- La final 2.5.0 se publicará en `Changelogs/Releases/StepLauncher-2.5.0/` con `latest: 2.5.0` tras validar el índice.

## Notas beta

- Esta beta puede tener regresiones menores (p. ej. `coverBgStyle` con imágenes muy pequeñas). Reportar con `launcher_music_history.json` y `cache/music/index_v2.gob.gz` si `GetTracksPage` tarda >50 ms.
