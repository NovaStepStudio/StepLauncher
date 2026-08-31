# Plan Refactorización Music Library Engine

## Auditoría Actual (2026-08-30)

### Backend Go
- `Manager.go:22` tiene `metaMem` en RAM pero `GetAllMetadata` retorna vacío, `SaveMetadata` no-op. No hay índice persistente real; cada `GetTracksPage` hace `WalkDir+sort` y `BuildTrackDTO` re-parsea `audiometa` por página (20 archivos) -> OK para página, pero `ScanFolder` solo lista paths sin metadata, y `extractMetaFast` se llama por cada DTO en `GetTracksPage` (20*2s timeout). Para 5000 archivos, paginación 20 es OK, pero `ScanAllMusicFolders` lista todos y luego frontend hace `localTracks` 2000 objetos completos (LocalStore.ts:111).
- `Scanner.go:36` `scanFolderPaths` WalkDir + audioExts, sort, límite 5000, sin incremental.
- `Cache.go:25` shards gob.gz para meta y covers zip, pero `Manager.SaveMetadata` es no-op, `GetAllMetadata` vacío -> cache no se usa para metadata, solo covers (pero covers también no-op para Refresh etc.). `coverShardSize=10`, `metaShardSize=100`.
- `Playlist`, `History`, `NowPlaying` usan JSON plano, bien.
- `Handlers/MusicBackend.go:215` `GetMusicTracksPaged` hace `GetTracksPage` que hace WalkDir cada vez (no usa índice).
- Frontend `LocalStore.ts:91` `loadLocalLibrary` guarda 2000 `LocalTrack` completos en `localTracks` ref reactivo, `ensureTrackMeta:183` hace `ReadAbsoluteFile` + `parseBlob` con `music-metadata` para 20 pistas -> duplica Go. `CoverCache.ts:13` LRU 80 Blob URLs sin revoke controlado.

### Problemas identificados (41 requisitos)
1. No índice persistente -> re-parse
2. Scanner no incremental (size/modTime)
3. WalkDir+sort en cada página
4. Frontend parsea audio (music-metadata)
5. DTO incluye cover raw/base64 en lista (500KB*20=10MB)
6. Covers sin lazy (thumb/raw ambos)
7. LocalStore 2000 objetos reactivos
8. Duplicación Go/Vue
9. Sin límite concurrencia (sequencial pero sin pool)
10. Sin stats, sin cleanup, sin invalidación

## Decisión Arquitectónica
- **No SQLite**: 5000 tracks * ~200 bytes = 1MB, gob+gzip shards ya funciona, evita CGO, mantiene `cache/music` existente, migración compatible. Mejorar `Index.go` sobre shards existentes.
- **Índice**: `MusicIndexEntry` con `ID=sha256(path+size+modTime)`, persistido `cache/music/index_v2.gob.gz` (versionado), mapa `path->entry`, LRU no necesario (todo en RAM ~1MB). Alternativa shards 100 como antes pero simplificar a single file para búsqueda/orden O(n log n) 5000 es trivial (<5ms).
- **Scanner incremental**: `Discover` WalkDir -> `Compare` index (size+modTime) -> `Queue` new/modified -> workers 3 -> `Process` -> `UpdateIndex` -> `Deleted` cleanup.
- **Progress extendido**: `Discovered/Processed/Skipped/Added/Updated/Removed/Errors`.

## Fases Implementación
1. **Index.go**: tipos, Load/Save, Get/Set, List, Search, Stats, version 2, migración old shards -> index.
2. **Scanner.go**: refactor `ScanFolders` incremental + worker pool + progress.
3. **Manager.go**: integrar Index, `GetTracksPage` via Index (no WalkDir), `GetMetadata` via Index, `SaveMetadata` via Index, `RebuildIndex`, `GetStats`.
4. **Covers.go**: separar `Covers.go` de Cache.go, LRU 32, dedup por CoverID, `GetCover(CoverID, variant)`.
5. **Handlers**: `GetMusicTracksPaged` via Index, nuevo `InitializeMusicLibrary`, `GetMusicLibraryStats`, `RebuildMusicIndex`, `ClearMusicCache`.
6. **Frontend**: `LocalStore` solo `libraryState` {total, page, tracks[20], query}, eliminar `music-metadata`, `ReadAbsoluteFile` solo para `blobUrlFor` reproducción, `ensureTrackMeta` eliminado, `PlayerStore` blob cache 6, `CoverCache` LRU 32 + revoke.

## Validación
- `go build ./...`, `wails3 generate bindings` 9 Services, `bun run type-check`/`build-only`
- Tests: 100/500/1300/3000 tracks simulados, medir RAM (runtime.ReadMemStats) antes/después scan, page request <50ms
- Compatibilidad: migrar `shard_*.gob.gz` old -> `index_v2.gob.gz`, playlists .m3u, history, nowplaying intactos
