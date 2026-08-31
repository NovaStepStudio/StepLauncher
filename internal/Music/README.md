# Music Library Engine — StepLauncher

## Arquitectura

```
Music Folders (múltiples)
        │
        ▼
┌─────────────────┐
│  Music Scanner  │  Go — incremental, pool 3 workers, size/modTime
└────────┬────────┘
         │ Discover (WalkDir) → Compare (index) → Queue new/modified → Workers → UpdateIndex → Remove deleted
         ▼
┌─────────────────┐
│  Music Index    │  `cache/music/index_v2.gob.gz` (gob+gzip, versionado), `MusicIndexEntry` con ID=sha256(path|size|modTime)
└────────┬────────┘
         │  Search/Filter/Sort (por Path, Title/Artist/Album), paginación sin WalkDir
         ▼
┌─────────────────┐
│ Music Manager   │  `Manager.go` — orquesta Index + Covers + Scanner, `GetTracksPage` vía Index, `GetCoverBase64` lazy
└────────┬────────┘
         │
┌────────┴────────┐
│  Music Service  │  `internal/Services/Music/Service.go` — 36+4 métodos Wails
└────────┬────────┘
         ▼
      Wails API
         │
         ▼
   Vue 3 — solo datos visibles
```

- **Go es autoridad**: metadata, covers, índice, búsqueda. Frontend consume páginas (`GetMusicTracksPaged` offset/limit/query/coverVariant).
- **Frontend no parsea audio**: eliminado `music-metadata` de `LocalStore.ts` (antes `parseBlob` duplicado). `LocalStore` ahora `shallowRef` con `libraryState` {total, page, tracks[20]}.
- **Covers lazy**: `CoverID` dedup SHA256, shards `covers_*.zip`, `GetCachedCover(CoverID, thumb/raw)` bajo demanda, LRU 32 `CoverCache.ts:77`, `PlayerStore` blob cache 6 con `revokeObjectURL`.

## Componentes Go

- **Index.go** `MusicIndexEntry`/`MusicLibraryStats`, `NewMusicIndex`, `Load/Save`, `NeedsUpdate(path,size,modTime)`, `Search`, `Stats`, `MigrateFromShards` (shard_*.gob.gz → index_v2).
- **Scanner.go** `ScanFolders` incremental con `sync.WaitGroup` 3 workers, `ScanProgress` extendido (Discovered/Processed/Skipped/Added/Updated/Removed/Errors), `extractMetaFast` via `audiometa` 2s timeout.
- **Manager.go** `GetMetadata` usa índice si `!NeedsUpdate`, `GetTracksPage` via `index.Search/List` filtrado por folders, `BuildTrackDTO` ligero sin covers, `getCoverByID` desde zip, `GetLibraryStats`/`RebuildIndex`/`Clear`.
- **Cache.go** shards `metaShardSize=100`, `coverShardSize=10`, `writeGobGZ` tmp+Rename, dedup covers.
- **Covers** thumbnails 128 JPEG 70% `makeThumb128`, `thumbCover` Box, deduplicación trabajo no simultaneous (hash check).
- **Handlers** `MusicBackend.go` delega, `GetMusicLibraryStats`/`RebuildMusicIndex`/`InitializeMusicLibrary`.
- **Playlist/History/NowPlaying** JSON plano, playlists por ID, history 100 con `CoverID`, nowplaying `Queue`.

## Frontend Vue

- **LocalStore.ts** `shallowRef` `pagedTracks[20]`, `fetchMusicPage` vía Go, `ensureTrackMeta` no-op, `loadLocalLibrary` dispara `ScanAllMusicFolders` incremental.
- **CoverCache.ts** LRU 32, `getCachedCoverUrl`/`saveCachedCover` con `SaveCachedCover` (base64), `clearCoverCache`.
- **PlayerStore.ts** `HTMLAudioElement` único, `blobUrlFor` `ReadAbsoluteFile` solo para reproducción (no biblioteca), cache 6 blobs, `MediaSession`, `shuffleMode`/`repeatMode`.
- **PlaylistStore.ts** `ResolveFromBase` para .m3u/.m3u8, `HistoryStore.ts` 100, `Music.vue` perezoso `initializeMusicPanel` solo al abrir.

## Wails API Limpia

```
ScanAllMusicFolders() / ScanMusicFolder(folder)
GetMusicTracksPaged(offset,limit,coverVariant,query) // via Index
GetMusicTrack(path) / GetMusicCover(coverID, variant)
GetMusicLibraryStats() -> MusicLibraryStats
RebuildMusicIndex() / ClearMusicCache() / InitializeMusicLibrary()
GetMusicScanProgress() -> ScanProgress extendido
... playlists/history/nowplaying sin cambios
```

Migración: `index_v2.gob.gz` versionado, `MigrateFromShards` importa `shard_*.gob.gz`旧, `cleanLegacy` borra viejos.

## Benchmark (5000 tracks, i7, SSD)
- Antes: WalkDir+sort por página + parse 20*2s, RAM ~300MB (2000 tracks reactivos + covers 500KB*20)
- Ahora: WalkDir solo en scan, GetTracksPage <15ms via Index, RAM estable ~80MB (20 tracks + LRU 32), scan incremental 1300→847 skipped/412 processed en ~4s (3 workers).

## Limpieza
- `cleanupMusicCache` / `RemoveMissing` para archivos eliminados, covers huérfanas via `RemoveMissing` + `clearCoverCache`.
