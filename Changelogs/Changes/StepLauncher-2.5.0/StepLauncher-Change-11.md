# Cambios de StepLauncher 2.5.0 (Change-11) — Stores transversales y composables: Connectivity, OfflineBadge, useBackground y useCoverPalette

- **Fecha**: 2026-08-31
- **Versión**: 2.5.0
- **Estado**: implementado y verificado
- **Release**: en desarrollo — aún no mencionado en ninguna release.

## Qué cambió

Dominio transversal `Common/` que `git ls-files --others` mostraba como `?? frontend/web/src/Common/Stores/Connectivity.ts`, `?? Common/Components/OfflineBadge.vue`, `?? Common/Composables/useBackground.ts`, `?? useCoverPalette.ts`, `?? useAppVersion.ts` y `?? Common/Widgets/BackgroundMusic.vue` sin trazar (Change-8 los agrupó, aquí se detallan con su contrato real).

### 1. `Common/Stores/Connectivity.ts` — estado offline

**Archivo nuevo** `frontend/web/src/Common/Stores/Connectivity.ts:1` (12 líneas):
- `ref<boolean> isOffline` (default `!navigator.onLine`), listeners `window.addEventListener('online'/'offline', ...)`, `onMounted`/`onUnmounted`. Expone `isOffline` para `OfflineBadge` y `Music/MenuView.vue` (stats `Offline`).
- Antes `App.vue` y `Music` no distinguían offline; ahora `MenuView.vue:33` muestra `En biblioteca` vs `Offline` según `isOffline`.

### 2. `Common/Components/OfflineBadge.vue` — badge offline

**Archivo nuevo** `frontend/web/src/Common/Components/OfflineBadge.vue:1` (34 líneas, `<script setup>`):
- Props `offline: boolean`, renderiza `<span v-if="offline" class="OfflineBadge">Sin conexión</span>` con `var(--warning)` y `IconWifiOff`. Usado en `App.vue:120` y `Music/MenuView.vue:33`.

### 3. `Common/Composables/useBackground.ts` — extracción de fondos

**Archivo nuevo** `frontend/web/src/Common/Composables/useBackground.ts:1` (extraído de `App.vue:241-332` en `Change-5` pero no detallado allí):
- `export function useBackground()` con `bg, bgImageUrl, bgVideoUrl, dynamicImage, dynamicIndex, videoReady, videoRef, refreshBackground(), startDynamicTimer(), onVideoReady(), onVideoError()`.
- `App.vue:60` ahora hace `const { bg, bgImageUrl, bgVideoUrl, dynamicImage, dynamicIndex, videoReady, videoRef, refreshBackground } = useBackground()` en lugar de 90 líneas inline. `Main.ts:56` ya no toca fondos; `useBackground` es testeable y `refreshBackground` se llama desde `Bootstrap/tasks/appearance.ts:12`.

### 4. `Common/Composables/useCoverPalette.ts` — paleta de carátula

**Archivo nuevo** `frontend/web/src/Common/Composables/useCoverPalette.ts:1` (47 líneas):
- `export function useCoverPalette(coverUrl: Ref<string>)` con `palette: Ref<{bg, fg, accent}>`, `coverBgStyle` (`background: linear-gradient(...)`), `paletteToCss(p)` helper (`--cover-bg`, `--cover-fg`).
- Usa `createImage` + `canvas 1x1` + `getImageData` para extraer color dominante sin `color-thief`. Usado en `Music/Components/LibraryView.vue:22` (`bibPalette`), `Music/MenuView.vue:18` (`coverBgStyle`), `Mods/Detail.vue:566` (fondo de galería).
- Antes cada vista hacía `new Image()` inline; ahora centralizado y con `watch(coverUrl)`.

### 5. `Common/Composables/useAppVersion.ts` — versión desde `build/config.yml`

**Archivo nuevo** `frontend/web/src/Common/Composables/useAppVersion.ts:1` (28 líneas):
- `export function useAppVersion()` con `version: Ref<string>` (`2.5.0` desde `build/config.yml:12`), `isDev: Ref<boolean>` (`import.meta.env.DEV`), `appVersionLabel` (`v2.5.0` o `v2.5.0-dev`). Usado en `Settings/Sections/About.vue:188` (`thirdPartyLibs`).

### 6. `Common/Stores/Music.ts` — música de fondo con dedup por URL

**Archivo `??` existente** `frontend/web/src/Common/Stores/Music.ts:1` (180 líneas, reescrito en `Change-2` pero aquí se vincula a `useBackground` y `Connectivity`):
- Estado `musicList`, `currentSrc` (`data URI base64` vía `musicDataUri()`), `playing`, `volume` (`localStorage stl_bgm_volume` + `watch` a `Personalization.backgroundMusic.volume`), `loadError`.
- `validateAndReadMusic` con `<audio>` temporal (`measureAudioDuration`) en lugar de `music-metadata` (ver `Change-2.md:18`).

### 7. `Common/Widgets/BackgroundMusic.vue` + `Styles/BackgroundMusic.scss` — widget dual

**Archivos `??`** `frontend/web/src/Common/Widgets/BackgroundMusic.vue:1` (240 líneas) + `Styles/BackgroundMusic.scss:1` (180 líneas):
- Fuente dual `launcherSource` (`localStorage stl_bgm_source`, default `background`), `isLauncher`, `setSource`, datos unificados `displayTitle/Artist/CoverUrl`, `playing/currentTime/duration/volume/playMode` computados según fuente, `show` (`launcherHasMusic` vs `backgroundMusic.enabled`), `onTogglePlay/onNext/onPrev/onMode/onVolume` bifurcados, `queueOpen` + `Bgm_Queue` (10 pistas, `Bgm_QueueList`).
- `BackgroundMusic.scss` con `.Bgm_Source`, `.Bgm_SourceBtn`, `.Bgm_Queue`, `.Bgm_QueueRow`, `.Bgm_VolMenu` con `backdrop-filter` y `scrollbar` estilizado. Antes `Widget.vue` era 65 líneas sin fuente dual.

## Por qué

- `App.vue` tenía 732 líneas con fondos, offline y paleta inline; `useBackground`/`useCoverPalette` permiten `App.vue` ligero (ver `Change-5.md:61` que elimina 140 líneas).
- `Connectivity` faltaba para el modo offline de `Music` (stats `Offline` vs `En biblioteca` en `MenuView.vue:33`).
- `useAppVersion` evita hardcodear `2.5.0` en `About.vue`.

## API afectada

- `Common/Stores/Connectivity.ts` — `isOffline: Ref<boolean>`.
- `Common/Components/OfflineBadge.vue` — prop `offline`.
- `Common/Composables/useBackground.ts` — `useBackground()`.
- `Common/Composables/useCoverPalette.ts` — `useCoverPalette(coverUrl)`, `paletteToCss`.
- `Common/Composables/useAppVersion.ts` — `useAppVersion()`.
- `Common/Widgets/BackgroundMusic.vue` — `launcherSource`, `isLauncher`, `displayTitle` etc.

## Comportamiento anterior/nuevo

- Antes: fondos inline en `App.vue`, paleta duplicada en cada vista, sin badge offline, versión hardcodeada.
- Ahora: `App.vue` usa `useBackground()` + `useCoverPalette(bgImageUrl)` + `OfflineBadge(:offline="isOffline")`; `MenuView` muestra `En biblioteca`/`Offline` dinámico; `About` muestra `v2.5.0` desde `build/config.yml`.

## Cómo verificar

- `bun run build` en `frontend/` — `vue-tsc` pasa con `useBackground`/`useCoverPalette` (imports `@/...`).
- Abrir launcher sin conexión → `OfflineBadge` visible en `App.vue` y `MenuView` muestra `Offline`.
- Cambiar `build/config.yml:12` version → `About.vue` refleja nuevo `appVersionLabel` sin hardcode.
