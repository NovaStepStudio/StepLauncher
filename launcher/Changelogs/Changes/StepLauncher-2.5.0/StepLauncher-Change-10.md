# Cambios de StepLauncher 2.5.0 (Change-10) — Configuración extendida: VerifyBeforeLaunch y metadatos de galería

- **Fecha**: 2026-08-31
- **Versión**: 2.5.0
- **Estado**: implementado y verificado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

Extensión de `internal/Config/Config.go` con nuevo flag `VerifyBeforeLaunch` y metadatos `ImageAuthor/ImageModName/ImageUrl` para la galería, vinculados a `internal/Core/Assets/Assets.go` (`PruneOrphanGallery`/`RegisterGalleryBackground` con dedup por URL) y a `internal/Handlers/Gallery.go`.

### 1. `LauncherConfig.VerifyBeforeLaunch` (`internal/Config/Config.go:61`)

- **Antes**: solo `VerifyIntegrity *bool` + `IntegritySector string`. No había forma de distinguir “verificar todo” vs “verificar antes de lanzar”.
- **Ahora**:
  ```go
  VerifyBeforeLaunch *bool `json:"verifyBeforeLaunch"`
  func (l LauncherConfig) VerifyBeforeLaunchEnabled() bool {
      if l.VerifyBeforeLaunch == nil { return true } // default true
      return *l.VerifyBeforeLaunch
  }
  func floatPtr(v float64) *float64 { return &v } // helper añadido
  ```
- Default `true` si `nil` (instalaciones antiguas sin campo mantienen verificación). `sanitize()` y `Default()` registran el extra `FileVerify` si procede.
- Consumidores: `internal/Core/Launcher/Manager.go:50` (`if cfg.VerifyBeforeLaunchEnabled() { Verify() }`), `internal/Handlers/Engine/Integrity.go:1` (`loadVerifyBeforeLaunch`), `internal/Handlers/App.go:320` (`registerLoaderSession`).

### 2. `BackgroundConfig` con autor/mod/url (`internal/Config/Config.go:94`)

- **Antes**: `BackgroundConfig` con `ImagePath`, `VideoPath`, `DynamicImages`, `DynamicOrder`, `DynamicInterval`.
- **Ahora** añade:
  ```go
  ImageAuthor  string `json:"imageAuthor,omitempty"`
  ImageModName string `json:"imageModName,omitempty"`
  ImageUrl     string `json:"imageUrl,omitempty"`
  ```
- Persistido en `launcher_config.json` vía `UpdatePersonalization` y usado por `referencedBackgrounds()` (`internal/Handlers/App.go:792`) para `PruneOrphanGallery`. Antes `referencedBackgrounds()` solo guardaba `Base`, ahora guarda `clean` + `basename` + clave original para comparar tanto `cache/backgrounds/foo.png` como `foo.png`.

### 3. `Assets` con ciclo de vida de galería (`internal/Core/Assets/Assets.go:413`)

- `RegisterGalleryBackground(path, author, modName, url, title)` ahora normaliza `TrimSpace` y **deduplica por `Url`** además de por `Path`: si existe entrada con misma `Url` pero distinto `Path`, borra el archivo viejo con `removeWithRetry` y elimina la entrada vieja antes de insertar (evita duplicados `UnixNano` del reporte `Bugs/2.5.0/Bug-1.md:15`).
- `PruneOrphanGallery(referenced map[string]bool) (int, error)` nuevo: mantiene solo entradas cuyo `Path` limpio o `Base` esté en `referenced`; para cada huérfana borra `filepath.Join(m.rootDir, filepath.FromSlash(clean))` y elimina del slice; `Save` solo si `n>0`.
- `RemoveGalleryBackground(relPath) error` helper para rollback si `UpdatePersonalization` falla tras descargar.
- `internal/Handlers/App.go:496` `updatePersonalizationInternal(p Personalization) error` envuelve `config.UpdatePersonalization(p)` (que hace `cleanupBackgrounds` de archivos) + `pruneOrphanGallery()` (JSON). `Startup():146` llama `pruneOrphanGallery()` para migrar instalaciones ya afectadas (22 entradas → 1). `ClearAllCache():831` también purga.

### 4. `Gallery.go` con dedup antes de descargar (`internal/Handlers/Gallery.go:24`)

- Antes de crear `destDir/fileName` (`fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixNano(), ext)`), busca en `assets.Load()` entrada con `Url == urlNorm` y `os.Stat(full)` existe: si existe, **reutiliza** `Path` sin descargar, registra metadata, hace `updatePersonalizationInternal` con `ImagePath = g.Path`, log `[Gallery] Fondo reutilizado (deduplicado por URL)` y retorna. Evita descargar dos veces la misma URL de Modrinth.
- Post-guardado con prune y rollback (`Gallery.go:94-115`): tras `RegisterGalleryBackground`, usa `updatePersonalizationInternal` y si falla hace `os.Remove(destPath)` + `assets.RemoveGalleryBackground(rel)`.

## Por qué

- `VerifyBeforeLaunch` faltaba para el flujo “Verificar integridad antes de lanzar” pedido en `Settings/Sections/Integrity.vue`; sin él, la verificación era global o nada. Con `*bool` + default `true`, instalaciones antiguas no pierden protección y el usuario puede desactivarlo por instancias.
- La galería crecía infinito (`Bugs/2.5.0/Bug-1.md:22` con 22 entradas huérfanas, solo 1 archivo vivo) porque `cleanupBackgrounds` borraba archivos pero no `launcher_assets.json.gallery`. Faltaba el par JSON (`PruneOrphanGallery`) y el dedup por URL (misma URL → dos `UnixNano` distintos).

## API afectada

- `internal/Config/Config.go:61-103` — `LauncherConfig.VerifyBeforeLaunch *bool`, `VerifyBeforeLaunchEnabled()`, `BackgroundConfig.ImageAuthor/ModName/Url`, `floatPtr`, `ExtraKeyPlaylists`/`FilePlaylists`/`FileMusicHistory` ya en Change-8 pero aquí se vinculan.
- `internal/Core/Assets/Assets.go:413-554` — `RegisterGalleryBackground` (dedup URL), `PruneOrphanGallery`, `RemoveGalleryBackground`.
- `internal/Handlers/App.go:496-530,146,792,831` — `updatePersonalizationInternal`, `pruneOrphanGallery`, `referencedBackgrounds`, `Startup`, `ClearAllCache`.
- `internal/Handlers/Gallery.go:13-115` — `DownloadGalleryImageAsBackground` (dedup + prune + rollback).

## Comportamiento anterior/nuevo

| Acción | Antes | Ahora |
|---|---|---|
| `VerifyBeforeLaunch` nil | No existía | `true` por defecto, `Integrity` distingue “antes de lanzar” |
| Aplicar fondo (URL nueva) | Añade 1 entrada, nunca borra anterior | Añade 1, `prune` borra no referenciada (JSON+archivo) |
| Aplicar fondo (misma URL) | Duplicado `UnixNano` | Reutiliza archivo, log dedup |
| Quitar fondo (`none`) | Borra archivo, JSON huérfano | Purga `gallery` → `[]` |
| Arrancar con JSON huérfano | Mantiene 22 entradas | `Startup` purga al arrancar |

## Cómo verificar

- `go vet ./internal/Config ./internal/Core/Assets ./internal/Handlers` — 0 errores.
- Setear `launcher_config.json` sin `verifyBeforeLaunch` → `VerifyBeforeLaunchEnabled()` == true; setear `false` → false.
- Aplicar dos veces “Aplicar fondo” misma URL → `gallery` sigue con 1 entrada, log `Fondo reutilizado`.
- Quitar fondo → `launcher_assets.json.gallery == []` y `cache/backgrounds/` sin huérfanos.
