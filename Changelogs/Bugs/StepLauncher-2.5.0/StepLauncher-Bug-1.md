# Bugs de StepLauncher 2.5.0 (Bug-1) — Acumulación infinita de fondos de galería en launcher_assets.json

- **Fecha**: 2026-08-31
- **Versión**: 2.5.0
- **Estado**: corregido y verificado
- **Release**: en desarrollo — aún no mencionado en ninguna release.

## El bug en cuestión

El sistema de fondos de mods (“Aplicar fondo” en la galería de `Mods/Detail.vue:566`) descargaba correctamente la imagen a `cache/backgrounds/` y la registraba en `launcher_assets.json` (`gallery: [{path, author, modName, url, title}]` via `internal/Core/Assets/Assets.go:413` y `internal/Handlers/Gallery.go:13`), y la establecía como `Personalization.Background.ImagePath`.

Al **quitar** el fondo (tipo `none` en `Settings/Sections/Personalization.vue:326` → `clearImage()`) o **cambiarlo** por otro (otro “Aplicar fondo”, importar imagen/video, pasar a `video`/`dynamic`), el archivo físico huérfano sí se borraba — `Config.cleanupBackgrounds()` en `internal/Config/Config.go:1194` elimina de `cache/backgrounds/` y `backgrounds/` todo lo no referenciado en `ImagePath/VideoPath/DynamicImages` — pero **la entrada JSON en `launcher_assets.json.gallery` no se eliminaba nunca**.

Evidencia aportada por el usuario (22 entradas en `gallery`, solo 1 archivo existente `Solas Shader_Septonious_1787986530182608100.png`):
```json
"gallery": [
  {"path":"cache/backgrounds/Complementary Shaders - Reimagined_Desconocido_1787970607299158700.png", "url":"...17a9de..."},
  {"path":"cache/backgrounds/Complementary Shaders - Reimagined_EminGT_1787971462959453800.png", "url":"...35b1b4eb..."},
  {"path":"cache/backgrounds/Complementary Shaders - Reimagined_EminGT_1787973401232295500.png", "url":"...35b1b4eb..."}, // mismo URL, distinto timestamp -> duplicado
  // ... 19 más, todas huérfanas
  {"path":"cache/backgrounds/Solas Shader_Septonious_1787986530182608100.png", "url":"...2cf04a..."} // único vivo
]
```
Incluso la **misma URL** de Modrinth generaba entradas duplicadas con distinto `UnixNano` en el filename (`DownloadGalleryImageAsBackground` hacía `fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixNano(), ext)` en `Gallery.go:58` y `RegisterGalleryBackground` solo deduplicaba por `Path`, nunca por `Url`), por eso `35b1b4eb6a...` y `52b5e927b7...` aparecen 2-3 veces.

## Qué afectaba y qué hacía

- **Crecimiento infinito de `launcher_assets.json`**: cada “Aplicar fondo” añadía 1 entrada que nunca se purgaba. Con uso intensivo de la galería, el JSON crece sin límite.
- **Inconsistencia disco ↔ JSON**: `ClearAllCache` y `cleanupBackgrounds` borraban el `.png/.webp` huérfano, pero el JSON seguía apuntando a paths inexistentes. El conteo de caché (`App.launcherCacheCount:811`) y el pruning de archivos estaban desacoplados del pruning JSON.
- **No hay ciclo de vida**: ni `App.UpdatePersonalization:496` (wrapper de `Config.UpdatePersonalization`) ni `DownloadGalleryImageAsBackground:109` purgaban `gallery`; tampoco había purga al arrancar para instalaciones ya afectadas ni al hacer “Limpiar caché”.
- **Desperdicio de red/disco**: aplicar dos veces la misma imagen (misma URL) descargaba y guardaba dos archivos distintos en vez de reutilizar el ya existente.

## Solución final

Se implementa **ciclo de vida completo** para `gallery`, simétrico al de los archivos: *solo permanece en `launcher_assets.json` lo que está referenciado por `Personalization.Background`*.

### 1. Backend: `internal/Core/Assets/Assets.go`

- **`RegisterGalleryBackground` (línea 413)**: ahora deduplica **por URL** además de por `Path`. Si la misma `Url` ya existe con otro `Path`, borra el archivo viejo con `removeWithRetry` y elimina la entrada antigua antes de insertar la nueva. Evita los duplicados del reporte (`35b1b4eb`, `52b5e9...`). Normaliza `author/modName/url/title` con `TrimSpace`.
- **`PruneOrphanGallery(referenced map[string]bool)` (nuevo)**: mantiene solo entradas cuyo `Path` limpio o `Base` esté en `referenced` (set construido desde `Background.ImagePath/VideoPath/DynamicImages`). Para cada entrada huérfana borra el archivo físico si existe (`filepath.Join(m.rootDir, filepath.FromSlash(clean))`) y elimina la entrada. Retorna `n` eliminadas y hace `Save` solo si hubo cambios. Es la contraparte JSON de `Config.cleanupBackgrounds`.
- **`RemoveGalleryBackground(relPath)` (nuevo)**: helper para rollback si `UpdatePersonalization` falla tras descargar.

### 2. Backend: `internal/Handlers/App.go`

- **`referencedBackgrounds():792`**: ampliado para poblar el map con **tanto `clean` completo (`cache/backgrounds/foo.png`) como `basename` y clave original**. Antes solo guardaba `Base`, ahora permite que `PruneOrphanGallery` compare por ambas formas (compatibilidad con paths guardados con o sin normalización).
- **`pruneOrphanGallery()` (nuevo, línea ~503)**: wrapper que llama `assets.PruneOrphanGallery(referencedBackgrounds())`, loguea `WARN` o `Gallery huérfana purgada: %d entradas`. Centraliza la política.
- **`updatePersonalizationInternal(p Personalization) error` (nuevo)**: hace `config.UpdatePersonalization(p)` (que ya hace `cleanupBackgrounds` de archivos) y luego `pruneOrphanGallery()` para JSON. `UpdatePersonalization(p)` ahora delega ahí, preservando firma void para bindings.
- **`Startup():146`**: tras aplicar `MinecraftConfig` y `RichPresence`, invoca `pruneOrphanGallery()` para **migrar instalaciones ya afectadas** (las 21 huérfanas del reporte se purgan al siguiente arranque sin esperar cambio manual).
- **`ClearAllCache():831`**: tras borrar archivos huérfanos de `launcherBackgroundDirs()` y `ClearCoverCache()`, también purga `gallery` con `PruneOrphanGallery(ref)` y loguea.

### 3. Backend: `internal/Handlers/Gallery.go`

- **Deduplicación antes de descargar (línea 24-52)**: antes de crear `destDir/fileName`, busca en `assets.Load()` una entrada con `Url == urlNorm` y verifica `os.Stat(full)` existe. Si existe, **reutiliza** esa `Path` sin descargar: registra metadata actualizada, hace `updatePersonalizationInternal` con `ImagePath = g.Path`, logea `[Gallery] Fondo reutilizado (deduplicado por URL)` y retorna. Si el archivo no existe, sigue flujo normal y `RegisterGalleryBackground` reemplazará la entrada vieja (borrando su archivo).
- **Post-guardado con prune y rollback (línea 94-115)**: tras `RegisterGalleryBackground`, usa `updatePersonalizationInternal` en lugar de `config.UpdatePersonalization` directo, de modo que el cambio dispara tanto `cleanupBackgrounds` (archivos) como `pruneOrphanGallery` (JSON) en un paso. Si falla, hace rollback: `os.Remove(destPath)` + `assets.RemoveGalleryBackground(rel)`.

**Comportamiento nuevo vs anterior:**

| Acción | Antes | Ahora |
|---|---|---|
| Aplicar fondo (URL nueva) | Añade 1 entrada, nunca borra la anterior | Añade 1, `prune` borra la anterior no referenciada (JSON + archivo) |
| Aplicar fondo (misma URL) | Añade duplicado con nuevo `UnixNano` | Reutiliza archivo existente, solo actualiza metadata |
| Quitar fondo (`none`) | Borra archivo, JSON queda huérfano | Borra archivo *y* purga `gallery` → queda `[]` |
| Cambiar a video/dynamic/import | Borra archivo imagen, JSON queda | Purga `gallery` completa si no queda `image` referenciado |
| Arrancar launcher con JSON huérfano | Mantiene 22 entradas | `Startup` purga al arrancar |
| Limpiar caché | Borra archivos huérfanos, deja JSON | También purga `gallery` |

### Archivos y APIs afectadas

- `internal/Core/Assets/Assets.go:413-554` — `RegisterGalleryBackground` (dedup URL), `PruneOrphanGallery`, `RemoveGalleryBackground`
- `internal/Handlers/App.go:496-530, 146, 792-830, 831` — `UpdatePersonalization`/`updatePersonalizationInternal`/`pruneOrphanGallery`/`referencedBackgrounds`/`Startup`/`ClearAllCache`
- `internal/Handlers/Gallery.go:13-115` — `DownloadGalleryImageAsBackground` (dedup + prune + rollback)
- Sin cambios de bindings expuestos (`DownloadGalleryImageAsBackground`, `UpdatePersonalization`, `ClearAllCache` mantienen firma); el frontend `Mods/Detail.vue:266` y `Settings/Sections/Personalization.vue` se benefician sin modificar.

### Cómo verificar

1. **Caso del reporte**: poblar `launcher_assets.json` con las 22 entradas del ejemplo, dejar solo `Solas Shader_Septonious_1787986530182608100.png` en `cache/backgrounds/`, setear `personalization.background.imagePath` a esa última. Al arrancar o al cambiar/quitar fondo, verificar que `gallery` queda con 1 entrada (o 0 si se quita) y que los 21 archivos inexistentes no dejaron rastro JSON.
   ```bash
   go test ./internal/Core/Assets -run TestPruneOrphanGallery_ScenarioUsuario -v
   go test ./internal/Core/Assets -run TestRegisterGalleryBackground_DedupUrl -v
   go test ./internal/Core/Assets -run TestPruneQuitarFondo -v
   ```
   Los tres tests (añadidos temporalmente para validar) pasan.

2. **Dedup URL**: aplicar dos veces “Aplicar fondo” sobre la misma imagen (misma `url`). Verificar que solo se descarga una vez, `gallery` sigue con 1 entrada y el log muestra `Fondo reutilizado (deduplicado por URL)`.

3. **Quitar fondo**: `Personalization` → Tipo `Ninguno` → Guardar. Verificar `launcher_assets.json.gallery` vacío o sin la entrada previa.

4. **Compilación**:
   ```bash
   go vet ./...
   go build -o NUL ./internal/Core/Assets
   go build -o NUL ./internal/Handlers
   ```
   Sin errores. El frontend no requiere cambios (`vue-tsc` preexistente sin regresiones en este flujo).
