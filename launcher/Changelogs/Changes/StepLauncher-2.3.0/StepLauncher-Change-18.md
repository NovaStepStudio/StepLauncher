# StepLauncher-Change-18: Descarga rediseñada — círculo de progreso, detalles técnicos con datos del motor, widget en segundo plano y `--progress-color`

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

Rediseño completo de la experiencia de descarga/instalación en
`frontend/web/src/Modals/InstallationModal.vue` + backend de descarga.

### 1. Círculo de progreso grande (sin brillo animado)

La barra horizontal con brillo desplazándose (`::after` + keyframes
`InstallationModal_Shine`) se eliminó por completo. En su lugar hay un **aro SVG**
de 9.5rem (r=52 en `viewBox 0 0 120 120`) con `stroke-dasharray/offset`, el
porcentaje centrado dentro (`RingPct`) y la estela del fill animada vía
`transition: stroke-dashoffset`. En estado `completed` el aro pasa a verde
`#5ed89a`.

### 2. Detalles técnicos colapsables y con TODOS los datos del motor

El botón "Detalles técnicos" (`.InstallationModal_DetailsToggle`) expande una
tarjeta compacta que usa la información completa de `DownloadProgress`:

- **Stats globales** (4 tarjetas): archivos descargados/total, MB
  descargados/total, archivos en cola y velocidad MB/s.
- **Archivo actual** (`currentFile`/`currentSection`/`currentProgress`): nombre,
  barra de progreso, sección y %.
- **Progreso por sección** (`sections` + `sectionsCompleted`): filas con label
  (`sectionLabel`), barra, % y desglose `archivos · MB descargados/MB total`
  (los campos `mbDownloaded`/`mbTotal` de `SectionProgress`, que antes no se
  mostraban); las secciones completadas se marcan en verde.
- **Descargando ahora** (`activeFiles`, hasta 4): nombre + barra + %.
- **En cola** (`queuedCount` + `queuedPreview`): chips con "+N" pendientes.
- **Registro del motor** (`log`, últimas 10 líneas): bloque monoespaciado con
  scroll.

### 3. Widget de descarga en segundo plano

Nuevo store `frontend/web/src/stores/downloads.ts` (`download`, `setDownload`,
`clearDownload`, `isDownloading`) y componente `DownloadWidget` en `App.vue`:
si hay descarga activa y el modal está cerrado, aparece un botón flotante
"Descargando {versión}" con barra y % (`--progress-color`) que al hacer clic
reabre el modal. Cerrar el modal NO detiene la descarga (el modal sigue montado
y escuchando eventos Wails). `z-index: 5` (antes 95) para no flotar sobre todo.

### 4. Nueva variable de color `--progress-color`

- Definida en `frontend/web/index.css` con fallback
  `var(--background-button-primary)`.
- `stores/ui.ts` la aplica en runtime desde `c.buttons`.
- La consumen el aro, el widget, las barras de secciones y los stats (vía `$pc`
  en el SCSS). El color verde `#5ed89a` (completado/ok) se mantiene aparte.

### 5. Un único SCSS

Los estilos viven en `frontend/web/src/Styles/InstallationModal.scss`
(importado con `@use '../Styles/InstallationModal.scss';` dentro del
`<style scoped>`); se **eliminó** `frontend/web/src/Modals/InstallationModal.scss`
(el duplicado viejo).

### 6. Imagen del modloader y del icono limpias

`.InstallationModal_Icon` y las imágenes de los loaders ya no tienen fondo, ni
borde ni `image-rendering: pixelated` (antes salían encajonadas y pixeladas).

### 7. Vuelta al inicio tras completar

Al completar la descarga el modal vuelve solo al estado de configuración (`setup`)
transcurridos 5 minutos aunque esté cerrado (`scheduleDoneReset`, limpiado en
`onUnmounted`). El resumen final muestra solo archivos descargados y MB (nada de
"archivos existentes").

## Backend (corrección relacionada)

- `internal/Core/Downloader/Manager.go`: fix del crash al cancelar
  (doble `Unlock()` en `runDownload`) — ver `Changelogs/Errors/StepLauncher-Error-8.md`.

## Archivos tocados

- `frontend/web/src/Modals/InstallationModal.vue`
- `frontend/web/src/Styles/InstallationModal.scss`
- `frontend/web/src/Modals/InstallationModal.scss` (eliminado)
- `frontend/web/src/App.vue` (widget + `z-index: 5`)
- `frontend/web/src/stores/downloads.ts` (nuevo)
- `frontend/web/index.css` (`--progress-color`)
- `frontend/web/src/stores/ui.ts` (`applyRootVar('--progress-color', ...)`)
- `internal/Core/Downloader/Manager.go` (fix double unlock)

## Por qué

- El brillo animado en la barra era molesto; el aro grande es más limpio y la
  vuelta a `setup` a los 5 min evita deja el modal bloqueado en "completado".
- Los detalles técnicos servían de poco porque solo pintaban %; ahora muestran
  todo lo que emite el motor (MB por sección, archivo actual, activos, cola y log).
- El widget en segundo plano con `z-index` propio permite seguir la descarga al
  cerrar el modal sin tapar la UI.
- El color del progreso se controla con una variable CSS propia en lugar de un
  azul fijo.

## Cómo verificar

- `bun run build` (dentro de `frontend/`) → OK (`vue-tsc --build` + `vite build`).
- `go build ./...` → OK.
- CSS compilado contiene `InstallationModal_Ring`, `InstallationModal_RingPct`,
  `InstallationModal_DetailsToggle`, `InstallationModal_DetailStats/DetailRow` etc.;
  ya no contiene `InstallationModal_GlobalBar` ni `InstallationModal_Shine`;
  `.DownloadWidget{...z-index:5...}`.
- Flujo de cancelación: cancelar la descarga ya no crashea la app.