# StepLauncher-Change-15: Modal de instalación reescrito (manifest de Mojang siempre vía Go, progreso global + por elementos y control de descarga)

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

Se rehace por completo el modal de instalación (`frontend/web/src/Modals/InstallationModal.vue`):

**1) Versiones SIEMPRE desde el manifest de Mojang:** el selector de versiones se carga
obligatoriamente vía el binding Go `FetchVersionManifest()` (prohibido el fetch JSON en el
frontend). Se agrupan por tipo (ÚLTIMA / SNAPSHOT / ANTIGUA), con buscador y selección
automática de la release más reciente.

**2) Selector de loader con iconos:** grid de 6 loaders (Vanilla, Fabric, Forge, NeoForged,
Quilt, Legacy Fabric) usando los iconos pixel-art de `frontend/web/assets/icons/*.png` con
`image-rendering: pixelated`. Al elegir loader no-vanilla se cargan sus versiones vía
`GetModLoaderVersions` y se elige una específica. Si la obtención de versiones del loader
falla, se puede reintentar. Cambiar la versión de MC recarga la lista del loader.

**3) Panel de progreso completo:** barra global con porcentaje superpuesto, estadísticas
(descargado, velocidad, tamaño total, archivos), barra del archivo actual, secciones en 3
columnas (tarjetas con contador y barra), archivos activos con barra individual por archivo,
chips de archivos en cola con "+N más" y chips de secciones completadas bajo la barra.

**4) Control de la descarga:** botones Pausar / Reanudar / Cancelar (bindings
`PauseDownload` / `ResumeDownload` / `CancelDownload`). Fases `setup → installing → done`,
con pantalla `error` y resumen final (archivos, existentes, MB, tiempo).

**5) Instalación del loader en segundo plano:** tras completarse la descarga del juego, si el
loader no es vanilla se dispara `InstallModLoader` automáticamente; se siguen los eventos
`modloader_resolving/downloading/installing/installed/error`. Durante esa fase el footer avisa
que el loader se está instalando y no se permite cancelar.

**6) Recuperación de una descarga activa:** al abrir el modal, si hay una descarga en curso
(`ListDownloads` + `GetDownloadStatus`) se resincroniza y se vuelve a mostrar su progreso en
vivo.

## Archivos tocados

- `frontend/web/src/Modals/InstallationModal.vue` (reescrito completo)

## Backend implicado (ya completado en el trabajo previo, sin cambios nuevos)

- `internal/Core/Downloader/Types.go` y `internal/Core/Downloader/Manager.go`: snapshot con
  `sections`, `activeFiles`, `queuedCount`, `queuedPreview`, `speedMbps` (EWMA) y seguimiento
  por archivo; `ManifestVersion` con `type`/`releaseTime`.
- `frontend/wailsjs/go/*`: bindings/modelos regenerados.

## Por qué

- El modal antiguo no mostraba progreso real por elemento y no ofrecía control sobre la
  descarga; ahora se consume el manifest de Mojang por el binding Go (no por fetch en el
  frontend) y la interfaz refleja el estado completo del backend en tiempo real.

## Cómo verificar

- `bun run type-check` (dentro de `frontend/`) → OK (exit 0). No se ha ejecutado
  `wails build` (pendiente de confirmación visual con `wails dev`).