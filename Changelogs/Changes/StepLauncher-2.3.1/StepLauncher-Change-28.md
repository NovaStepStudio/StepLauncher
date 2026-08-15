# Changes/StepLauncher-2.3.1/StepLauncher-Change-28.md

- **Fecha**: 2026-08-13
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado (build OK).

## Qué cambió

### 1. Descargas resilientes (conexiones lentas/inestables)

`internal/Core/Downloader/Download.go` y `Manager.go`:

- Timeout global del `DefaultHTTPClient` de **30s → 90s** (el error real de los usuarios era `context deadline exceeded while awaiting headers` con conexiones malas; además 2 managers con hasta 64 workers cada uno saturaban el enlace con 128+ conexiones simultáneas).
- Retry base de 500ms → **1000ms con jitter aleatorio** (`rand.Int63n`), para no martillear los servidores en caídas.
- Reintentos por defecto de 3 → **5** (`MaxRetries`).
- **Concurrencia por defecto 24 → 8 workers** (`MaxConcurrency`) y **downloadAll adaptativo**: los workers se calculan con `min(MaxConcurrency, len(tasks))` (tope 64); si fallan la mitad o más de las tareas y hay más de 2 workers, se reintenta con la mitad de workers (recuperación progresiva en conexiones malas).
- Tras el pase de descarga, **reintento final de las tareas fallidas** (las que quedaron en estado `FileError`) antes de la verificación, usando el mismo backoff lento.

### 2. Pre-escaneo: no se descarga lo que ya existe

`internal/Core/Downloader/Manager.go`:

- Nuevo `preScan()` al arrancar cada descarga: marca como `existing` todo archivo que exista **con el mismo tamaño** (si el tamaño objetivo se conoce y coincide; si no se conoce el tamaño, basta con que exista), y borra los `.tmp` huérfanos de descargas abortadas.
- `mbTotal` y `FilesTotal` ahora cuentan **solo los bytes/archivos pendientes reales**: una instancia que ya tiene el 90% de sus assets muestra 100 MB a descargar, no los 900 MB de antes (el usuario veía que el launcher "descargaba de nuevo" lo ya descargado).
- Nuevo campo `Download.pendingTotal` para que el porcentaje se calcule sobre el total pendiente y no sobre el total absoluto.

### 3. IDs únicos: descargas simultáneas sin colisión

- `Types.go`: `Config.IDPrefix` (nuevo) para prefijar los IDs de descarga.
- `Engine.go`: el manager de versiones usa `IDPrefix: "ver-"` y el manager de instancias/modloaders `"inst-"` (antes ambos arrancaban en `dl-1`, así que dos descargas activas a la vez compartían ID y se pisaban en el frontend; `Integrity.go`, `Launcher.go`, `Fabric.go` y `Instance/Verify.go` no crean managers, solo usan `Config` para BuildTasks/FetchJSON, por lo que no necesitan prefijo).
- Default de `NewManager`: `IDPrefix = "dl-"`.

### 4. Frontend: store central de descargas + widget multi-descarga

`frontend/web/src/Stores/Downloads.ts` (reescrito):

- `downloads: Record<string, ActiveDownload>` con `label`, `kind` ('version' | 'instance'), `mbDownloaded`, `mbTotal`, `filesDownloaded`, `filesTotal` y `error` (se eliminaron `loader`/`mb`).
- **Un solo registro de eventos** (`download_progress`, `download_state`, `download_error`) idempotente: `ensureDownloadEvents()`.
- `registerDownload(dlId, kind, label)`, `updateDownload`, `clearDownload(id)` (limpieza automática: error → 8s/12s, resto → 800ms), computeds `activeDownloads`, `anyActive`, `latest`, `downloadOf`.

`frontend/web/src/Stores/Instances.ts`: proyección por instancia vía `watch(dlStore, deep)` + `dlToInstance` (las descargas se siguen mostrando en su instancia); `createInstance` y `addVersion` registran la descarga en el store central; `clearDownload(instName)` delega en `clearCentralDownload(dl.dlId)`; refresco de instancias al pasar a estado terminal (`ACTIVE_DL`/`TERMINAL_DL`).

`InstallationModal.vue` y `InstanceDownloadModal.vue`: `syncDownloadWidget` usa `registerDownload` con label (`Minecraft <versión>` / nombre de instancia) y kind; los clears (cancelar/error/resetRun) usan `clearCentralDownload`.

`frontend/web/src/Widgets/DownloadWidget/DownloadWidget.vue`: muestra todas las descargas activas (título "N descargas activas", subtítulo con estado o "Verificando..."); al pulsarlo con varias descargas abre el panel.

`frontend/web/src/Modals/DownloadsPanel.vue` + `Styles/Modals/DownloadsPanel.scss` (nuevos): mini-panel con la lista de descargas activas, barra de progreso con MB y % reales, botones pausa/reanudar/cancelar (`PauseDownload`/`ResumeDownload`/`CancelDownload`) y cierre automático al terminar.

`frontend/web/src/App.vue`: `widgetVisible = anyActive && !showInstall`; `openWidget()` decide destino: 1 sola descarga de versión → modal de instalación, de instancia → instancias, varias → panel.

### 5. Frontend: selector de carpeta con `<select>`

- `WelcomeModal.vue`: el grid 2x2 de modos se sustituye por un `<select>` con las 4 opciones (nombre + descripción), se muestra siempre la ruta destino ("Se guardará en: ..."), se mantienen el aviso de `.minecraft` con "Usar", el campo de ruta personalizada con "Examinar" y la lógica de guardado/reinicio; se eliminan `pickDirMode` e imports de iconos sin uso.
- `GeneralSettings.vue`: el grid `.SsDirGrid`/`.SsDirMode` se sustituye por `<select class="SsSel">` (estilo compartido de ajustes); `onDirModeChange` fuerza `separateGameDir = false` al elegir modo Minecraft.
- Limpieza de SCSS muerto: `.WelcomeModal_DirModes`, `.WelcomeModal_DirMode`, `.WelcomeModal_DirModeCheck`, `.SsDirGrid`, `.SsDirMode`; nuevo `.WelcomeModal_Select`.

## Por qué

Los usuarios con conexiones malas ("ilimitadas" pero inestables) no podían descargar ni 1 MB de assets: el timeout de 30s y la avalancha de 128 conexiones abrían y abortaban conexiones constantemente. Además el launcher contaba como pendiente todo lo que ya existía (mostraba 900 MB cuando casi todo estaba descargado) y dos descargas simultáneas (versión + instancia) compartían IDs y se pisaban. Se pidió además un selector de carpeta más cómodo (select en vez de botones) y, tras confirmar con el usuario, concurrencia solo auto-adaptativa y mini-panel con la lista de descargas.

## API afectada

- `downloader.Config.IDPrefix` (nuevo); defaults de `NewManager`: `MaxRetries=5`, `MaxConcurrency=8`, `IDPrefix="dl-"`.
- `Download.pendingTotal` (nuevo campo); `Manager.preScan`, `downloadAll` adaptativo, `runWorkers`, `failedTasks` (internos).
- Sin cambios en bindings de Wails (se reutilizan `PauseDownload`, `ResumeDownload`, `CancelDownload` y los eventos de progreso existentes).

## Comportamiento anterior/nuevo

- **Antes**: timeout 30s global; 64+ workers por manager y 2 managers con IDs duplicados (`dl-1`); reintentos cada 500ms sin jitter; el total mostrado incluía archivos existentes; dos descargas simultáneas se pisaban en el frontend; selector de carpeta con grid de botones.
- **Ahora**: timeout 90s; 8 workers por descarga con reducción adaptativa ante fallos masivos y retry final; backoff 1s + jitter con 5 reintentos; pre-escaneo por existencia+tamaño que borra `.tmp` y calcula los totales sobre lo pendiente; IDs prefijados (`ver-`/`inst-`); store central con widget multi-descarga y panel de descargas; selector de carpeta con `<select>` en bienvenida y ajustes.

## Cómo verificar

- `go build ./...` en la raíz: OK.
- `bun run build` en `frontend/`: type-check y vite build OK.
- Manual: instalar una versión con assets ya descargados (se muestra el total pendiente real y arranca con "Verificando..."); lanzar la descarga de una versión y de una instancia a la vez (ambas progresan en el panel; el widget muestra "2 descargas activas"); con conexión mala o proxy lento las descargas se reintentan y no fallan por timeout; bienvenida y ajustes → el modo de carpeta se elige con el desplegable y el hint muestra la ruta destino.
