# Changes/StepLauncher-2.3.1/StepLauncher-Change-1.md

- **Fecha**: 2026-08-08
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.

## Qué cambió

Sesión de mejoras del menú principal y del flujo de instalación/lanzamiento. Se documenta todo en esta única entrada porque es una sola tanda de trabajo sobre el mismo área (botón Jugar, widgets y configuración de launcher):

### 1. Widgets flotantes como componentes independientes

- Nuevo `frontend/web/src/Widgets/DownloadWidget/DownloadWidget.vue`: cada widget del menú principal vive en `src/Widgets/${elemento}/${elemento}.vue`, con su markup, estilos scoped y consumo directo de stores (`Stores/Downloads` calcula `percent` y visibilidad).
- `App.vue` quedó más delgado: solo lo renderiza en un `Transition` con `v-if` y maneja el evento `@open` para abrir el modal de instalación. Se eliminaron del `App.vue` los estilos `.DownloadWidget_*` y los computeds `widgetPercent`.
- `frontend/vite.config.ts`: nuevo chunk `widgets` en `manualChunks`.
- **Corrección de espaciado fantasma**: el widget tenía `bottom: 5rem` fijo (pensado para quedar sobre la barra Jugar), así que cuando `hasVersions` era falso (botón Jugar y selector de versión ocultos) el widget quedaba con un hueco de 5rem abajo. Ahora el componente lee `hasVersions` del store y aplica `bottom: 5rem` solo cuando la barra de control existe; sin ella queda pegado abajo a `0.75rem`, sin espaciado fantasma.

### 2. Botón Jugar y globo de mensaje con fases reales

- `internal/Core/Launcher/Helpers/Natives.go`: `ExtractNatives` acepta un callback `onProgress(cur, total, name)` y lo llama antes (0/total), por jar (i/total) y al terminar.
- `internal/Core/Launcher/Launcher.go`: la extracción de nativos ahora emite eventos `game_prepare` con `phase: "natives"` y progreso (antes solo emitía la fase `libraries` y la extracción era silenciosa: el botón decía "Descargando…" durante la extracción — ver Errors/StepLauncher-2.3.1/StepLauncher-Error-1.md).
- `frontend/web/src/Stores/Launcher.ts`: `launchPrepare` incluye `phase`; textos por fase ("Descargando archivos faltantes (x/y)…" / "Extrayendo archivos nativos (x/y)…", con nombre de archivo cuando hay label); nuevo computed `launchingPhaseLabel` (Descargando… / Extrayendo… / Lanzando…).
- `App.vue`: el botón muestra la fase real; el globo `LaunchMsg` muestra el texto de la fase con progreso, se limpia con `EventsOn('game_started', hideLaunchMessage)` (antes quedaba "Lanzando <versión>…" pegado para siempre) y sus estilos permiten texto en varias líneas en vez de truncar con ellipsis.

### 3. Nueva opción "Lanzar Minecraft al terminar una instalación"

- Backend: `internal/Config/Config.go` con el campo `LaunchAfterInstall bool` en `LauncherConfig` (JSON `launchAfterInstall`, defecto `false`) y `SetLaunchAfterInstall(v bool)`; `internal/Handlers/App.go` con la delegación; `app.go` con el binding Wails `App.SetLaunchAfterInstall(v bool)`.
- `frontend/web/src/Layouts/Sections/Settings/GeneralSettings.vue`: nuevo `SsRow` en Ajustes → General → Comportamiento; carga `cfg.launcher?.launchAfterInstall` y guarda con `SetLaunchAfterInstall`.
- `frontend/web/src/Modals/InstallationModal.vue`: al pasar a `phase === 'done'` (descarga completada, y en su caso ModLoader instalado) y con la opción activada: `refreshAfterDownload()` → `selectVersion()` de la versión recién instalada → cierra el modal → `launchGame()`. Guard `autoLaunchHandled` evita lanzamientos dobles por ciclo (se resetea en `resetRun()`). Funciona también si se cerró el modal durante la descarga (los eventos siguen llegando al componente, que está montado siempre).

## Por qué

Los widgets ganan visibilidad/espaciado según el estado real de la UI y son reutilizables; el botón Jugar y su globo engañaban (decía "Descargando…" aunque estaba extrayendo nativos, y el globo quedaba con un texto pegado sin caducar); y el flujo instalar → Jugar manual era un paso innecesario que ahora puede automatizarse.

## API afectada

- Binding Wails nuevo: `App.SetLaunchAfterInstall(v bool)`; se regeneran los bindings `frontend/wailsjs` con `wails build`/`wails dev`.
- `Config.launcher.launchAfterInstall` (campo nuevo; defecto `false`).
- Evento interno `game_prepare` (`GamePrepareData`): nuevo uso de `phase` con valor `"natives"` y callback de progreso en `helpers.ExtractNatives` (solo lo usa `Launcher.go`).
- Store de TS: `launchPrepare` agrega `phase`; nuevo `launchingPhaseLabel`.

## Comportamiento anterior/nuevo

- Widget de descarga: flotaba siempre a 5rem del borde inferior aunque no hubiera barra de control → se ajusta a la barra solo cuando existe.
- Botón Jugar: decía "Descargando…" durante toda la preparación → muestra Descargando…/Extrayendo…/Lanzando… según la fase real, y el globo muestra contador por archivo y desaparece al arrancar el juego.
- Instalación: al terminar quedaba parada en el modal → con la opción activa se cierra y lanza el juego con la versión recién instalada; desactivada, igual que antes.

## Cómo verificar

- `go build ./...`.
- `bun run build` en `frontend/` (type-check + build).
- Sin versiones instaladas, el widget de descarga queda pegado abajo; con la barra Jugar, encima de ella.
- Lanzar una versión sin librerías: el botón pasa por Descargando… → Extrayendo… con contadores en el globo; al abrir Minecraft el globo se limpia.
- Marcar la opción de Comportamiento e instalar una versión: al llegar al estado "Listo" el modal se cierra y Minecraft se lanza con esa versión.