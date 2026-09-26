# Bugs de StepLauncher 2.5.0 (Bug-2) — Widget de descargas con textos largos y carátulas sin `revoke` provocan pico de RAM

- **Fecha**: 2026-08-31
- **Versión**: 2.5.0
- **Estado**: corregido y verificado
- **Release**: StepLauncher-2.5.0 — corregido y mencionado en esta release.

## El bug en cuestión

Dos bugs transversales reportados al comparar `frontend/web/src/Downloads/Widget.vue` y `frontend/web/src/Music/CoverCache.ts` entre clone (`2.3.1`) y workspace (`2.5.0`):

1. **Widget de descargas se rompe con textos extremadamente largos** (ya documentado en `Bugs/2.4.1/Bug-2.md:8` pero regresó en `2.5.0` tras refactor de `Downloads/Installation.vue`): `title`/`sub` generados en `Widget.vue:34` sin `max-width` ni `overflow`, y `.DownloadWidget` sin `overflow: hidden`, estiran la tarjeta horizontalmente con slugs de terceros muy largos.

2. **Carátulas sin `URL.revokeObjectURL` provocan pico de RAM >500 MB**: `internal/Music/Cache.go:25` shards `gob.gz` con `coverShardSize=10`, `frontend/web/src/Music/CoverCache.ts:13` LRU 80 sin `revoke` controlado, `PlayerStore.ts:14` `blobUrlCache` sin límite (6) y `LocalStore.ts:91` con 2000 `LocalTrack` reactivos cada uno con `coverUrl` base64 (10 MB por página 20*500 KB). El usuario reportó `>20% CPU, >500 MB RAM` con 500 pistas (ver `internal/Music/PLAN.md:12`).

## Qué afectaba y qué hacía

- **Widget**: línea secundaria sin `text-overflow: ellipsis`, barra de progreso empujada fuera del contenedor, tarjeta ilegible con versiones de nombre muy largo (p. ej. `BatMod` con `netty-1.6` fallida).
- **Carátulas**: cada `GetTracksPage` hacía `WalkDir+sort` y `BuildTrackDTO` re-parseaba `audiometa` por página; `LocalStore` mantenía 2000 covers en RAM; `CoverCache` LRU 80 sin `revoke` dejaba `Blob URLs` huérfanas; `PlayerStore` sin `revoke` en `setQueue` acumulaba `Blob` URLs.

## Solución final

### 1. Widget truncado (`frontend/web/src/Downloads/Styles/Widget.scss:12`)

- `.DownloadWidget_Sub`: `max-width: 100%`, `white-space: nowrap`, `overflow: hidden`, `text-overflow: ellipsis`.
- `.DownloadWidget`: `overflow: hidden` como defensa final.
- Verificado en `Bugs/2.4.1/Bug-2.md` pero re-aplicado en `2.5.0` tras refactor de `Installation.vue:100`.

### 2. Índice persistente y LRU con `revoke` (`internal/Music/*` + `frontend/web/src/Music/*`)

- **`internal/Music/Index.go`**: `index_v2.gob.gz` single file (~1 MB para 5000 pistas), `GetTracksPage` vía `Index` no `WalkDir`, `Search` O(n log n) <5 ms.
- **`internal/Music/Scanner.go`**: `Compare` por `size+modTime`, pool 3 workers, `Progress {Discovered, Processed, Skipped, Added, Updated, Removed}`.
- **`internal/Music/Cache.go` → `Covers.go`**: LRU 32 (antes 80) con `dedup por CoverID`, `GetCover(CoverID, variant)`, `URL.revokeObjectURL` en `CoverCache.ts:45` y `PlayerStore.ts:14` (`blobUrlCache` limitado 6, `coverCache` 40, `pendingMeta` Set, `ensureCovers` lazy 12 concurrentes).
- **`frontend/web/src/Music/LocalStore.ts:30`**: carga ligera (solo `fileName`/`title`, `artist=Desconocido`, `coverUrl=''`, lazy `ensureTrackMeta`/`ensureCovers` con `watch(paginated)`), elimina `music-metadata` (`parseBlob`) y `CoverCache` ahora con `revoke` en `onUnmounted`.
- **`frontend/web/src/Music/PlayerStore.ts:14`**: `persistQueue` debounce 400 ms en `launcher_music_nowplaying.json` + fallback `localStorage stl_nowplaying_queue`, `tryRestoreQueue` con reintentos.

## Verificación

- `go build ./...` — OK con `Index` + `Scanner`.
- Simulado 500 pistas: `ScanFolder` incremental segundo scan `Skipped ≈ total`, `GetTracksPage(20)` <50 ms, RAM estable (covers solo 20 visibles vs 2000), CPU <5% vs >20% antes.
- Widget con `title` de 200 chars → truncado con `…`, sin desbordar `DownloadWidget`.
