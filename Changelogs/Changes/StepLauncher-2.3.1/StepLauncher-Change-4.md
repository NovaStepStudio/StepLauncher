# Changes/StepLauncher-2.3.1/StepLauncher-Change-4.md

- **Fecha**: 2026-08-09
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.

## Qué cambió

Se implementa por completo la **sección "Instancias"** del launcher: un panel en modal con la grilla de instancias (crear, editar, favorito, versión activa, jugar, descarga con progreso, cancelar, clonar, verificar, eliminar, capturas por instancia e importación de icono/banner). El sistema de instancias ya existía en el backend (`internal/Core/Launcher/Instance`); este cambio añade los bindings que faltaban y todo el frontend.

### 1. Nuevos bindings backend (`internal/Handlers/App.go` + `app.go`)

- `ListInstanceScreenshots(instanceName string) []ScreenshotsEntry`: lista las capturas guardadas en `instances/<nombre>/screenshots/` usando el helper refactorizado `listScreenshots` (compartido con `ListScreenshots`, que ahora es un wrapper del mismo con el root global).
- `validateInstanceName(name string) error` (helper): nombre no vacío, sin `..` ni separadores de ruta.
- `ImportInstanceAsset(name, kind, src string) ImportAssetResult`: importa un archivo propio como asset de una instancia (`icon`, `banner` o `background`), lo copia a `instances/<nombre>/assets/<kind><ext>` y devuelve la ruta relativa al root del launcher para guardarla en metadatos.
- `PickInstanceAssetFile() string`: abre un dialog de selección de archivo (OpenFileDialog con filtros png/jpg/webp) y devuelve la ruta elegida.

### 2. Store frontend (`frontend/web/src/Stores/Instances.ts`)

- Estado: `instances`, `details` (Record por nombre), `downloads` (Record por nombre), `launching`, `loadingList`; computed `sortedInstances` (favoritas primero, luego por `lastPlayed`).
- CRUD: `loadInstances`, `loadDetails`/`detailOf`, `loadAllDetails`, `createInstance`, `updateMetadata`, `toggleFavorite`, `deleteInstance`, `cloneInstance`, `updateConfig`, `addVersion`, `cancelDownload`, `verifyInstance`, `launchInstance`, `setLaunching`, `clearDownload`, `formatPlayTime`.
- Eventos de Wails: `download_progress`, `download_state`, `download_error` (se rastrean por instancia mediante `dlToInstance` que mapea `downloadId` → nombre), y `game_exited`/`game_crashed`/`game_stopped` (recarga estado al cerrar el juego) con `bindGameEvents`/`unbindGameEvents`.
- Los bindings nuevos no existen aún en `frontend/wailsjs` (se regeneran con `wails build`): todos los llamados usan el patrón `window.go.main.App` ya empleado en el proyecto, con fallback seguro.

### 3. Componentes frontend (`frontend/web/src/Modals/`)

- `InstancesModal.vue`: overlay de modais con cabecera y `InstancesView` dentro. Es el punto de entrada desde la sidebar (se abre con `showInstances` en `App.vue`).
- `InstancesView.vue`: la grilla. Tarjetas con banner, icono (fallback), favorito (estrella), título/descripción, versión activa, capturas, config, jugar (con estado "lanzando"), barra de progreso de descarga con cancelar, menú de contexto (editar, añadir versión, clonar, verificar, eliminar con confirmación de 4s).
- `InstanceFormModal.vue`: crear/editar — nombre, título, descripción, favorito, selector de versión a descargar (release/snapshot desde `GetVersions`), e importación de icono/banner con preview (`PickInstanceAssetFile` + `ImportInstanceAsset`).
- `InstanceSettingsModal.vue`: configuración individual — versión activa, RAM (min/max), Java oficial o ruta personalizada, fullscreen, resolución custom, preset de GC, preferencia de GPU.
- `ScreenshotsModal.vue`: ahora acepta `instance?: string | null`; si se pasa, lista las capturas de esa instancia (`ListInstanceScreenshots`) y titula "Capturas de X".

### 4. Estilos (`frontend/web/src/Styles/Modals/`)

- Nuevos `InstancesModal.scss`, `InstancesView.scss`, `InstanceFormModal.scss`, `InstanceSettingsModal.scss`: tarjetas, banner, estrellas, barra de progreso, dropdown de menú, overlays de sheets (añadir versión/clonar/verificar) y formularios. Cumplen el sistema de variables CSS existente (`--color-*`, `--background-*`, `--border-style`, `--progress-color`, etc.).

## Por qué

El backend de instancias estaba completo pero sin interfaz de usuario. La petición era "la sección de Instancias": panel completo con descarga, ejecución, config individual, assets (icono/banner), nombre/descripción, favorito y visor de capturas propio. Se eligió modal sobre la interfaz (patrón `AccountsModal`) y config por instancia en versión básica.

## API afectada

- `app.go` gana 3 bindings nuevos (`ListInstanceScreenshots`, `PickInstanceAssetFile`, `ImportInstanceAsset`). Ninguna binding anterior cambia.
- `wailsjs` se regenerará en el próximo `wails build`; mientras, el frontend usa `window.go` directamente.

## Comportamiento anterior/nuevo

- Anterior: el sistema de instancias solo accesible por bindings crudos (sin UI).
- Nuevo: panel completo en la UI; las instancias se crean con su carpeta propia (`instances/<nombre>/`), comparten librerías/assets globales y cache en `%APPDATA%\.StepLauncher\cache` (ver Change-3).

## Cómo verificar

- `go build ./...` en la raíz: pasa.
- `bun run build` en `frontend/` (vue-tsc + vite): pasa.
- Crear instancia → se descarga con progreso en su tarjeta → Jugar → juegos del sistema de instancias.
- Importar icono/banner en el formulario → aparecen en la tarjeta (preview con `loadLocal`).
- Capturas de una instancia: `InstancesModal` abierto desde la card.

## Pendientes (verificables al ejecutar)

- El progreso de archivos (`filesDownloaded/filesTotal`) depende del payload de eventos del `Downloader` (verificable con `wails dev`).
- `ListInstanceScreenshots`/`ImportInstanceAsset` solo disponibles tras regenerar bindings con `wails build`.