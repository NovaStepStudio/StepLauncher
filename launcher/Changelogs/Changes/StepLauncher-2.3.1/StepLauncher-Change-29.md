# Changes/StepLauncher-2.3.1/StepLauncher-Change-29.md

- **Fecha**: 2026-08-14
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado (go build + darwin build + bun run build OK).

## Qué cambió

### 1. Progreso por elemento sin "MB" duplicado

`frontend/web/src/Modals/InstallationModal.vue` y `InstanceDownloadModal.vue`:

- `fmtMb()` ya devuelve el sufijo de unidad ("X MB" / "X GB"), pero los templates añadían otro " MB" (`25,6MB/26,6 MB MB`). Se elimina el sufijo extra: ahora se muestra `25,6MB/26,6MB`.

### 2. El widget de descarga permanece visible al instalar un modloader

- `frontend/web/src/Stores/Instances.ts`: nuevo `MergedActiveDownload` y computed **`allActiveDownloads`** que fusiona las descargas centrales (`download_*`) con las instalaciones de modloaders (`loaderDls`, eventos `modloader_*`) en fase `resolving|downloading|installing`, con `kind: 'loader'`, `state = phase`, `percent = progress/total` (en archivos) y campos `phase`/`message`/`loader`; `cancellable: false` para loaders. Nuevo `anyAllActive`.
- `frontend/web/src/App.vue`: `widgetVisible` y `openWidget()` usan la lista combinada; kind `'loader'` → abre instancias (igual que `'instance'`).
- `frontend/web/src/Widgets/DownloadWidget/DownloadWidget.vue`: estados del loader en título/subtítulo ("Instalando Forge en <instancia>", "Resolviendo modloader…", "Descargando…", "Instalando…"); verificando/re-descargando sigue mostrando archivos.
- `frontend/web/src/Modals/DownloadsPanel.vue`: soporta items `kind: 'loader'` (textos de estado `resolving`/`installing`, subtítulo con versión + mensaje, sin botones de pausa/cancelar, icono `IconPuzzle`), mantiene MB para descargas normales.

### 3. Widget de instancia: reabrir el panel + información de MB

- `frontend/web/src/Stores/Downloads.ts`: nuevo estado global `panelVisible` con `showDownloadsPanel()`/`hideDownloadsPanel()` (cualquier vista puede abrir el panel sin duplicar estado).
- `frontend/web/src/App.vue`: usa el ref del store en el `v-model:visible` del `DownloadsPanel` (se elimina el ref local).
- `frontend/web/src/Modals/InstanceDetailView.vue`: el overlay de descarga activa muestra **MB descargados/totales + archivos** (nuevo `fmtMb` local) y un botón **"Panel"** que abre el panel global de descargas; nuevo `.InstDet_DlOverlayActions`/`.InstDet_DlOverlayPanel` en `InstanceDetailView.scss`.

### 4. Java 8 oficial (jre-legacy) para versiones antiguas

- `internal/Core/Launcher/Helpers/Java.go`: `resolveOfficialJava` ya no falla con componente vacío; usa **`jre-legacy`** (Java 8, el runtime oficial del launcher para pre-1.17).
- `internal/Core/Downloader/Tasks.go`: `BuildJavaRuntimeTasks` hace lo mismo (componente vacío → `jre-legacy`); `addJavaRuntime` se mantiene conservador (solo descarga Java si el version.json declara componente; el runtime antiguo se baja de forma perezosa al lanzar).
- `internal/Core/Launcher/Launcher.go` (`Launch`): si la versión no declara componente (antigua) y **no** hay `JavaExec` personalizado ni `UseOfficialJava`, se detecta el Java del sistema (`DetectJavaMajorVersion`); si es **>= 17**, se auto-conmuta a Java 8 oficial con log informativo. Si el Java oficial no puede prepararse (p. ej. sin red), se avisa por log y se continúa con el Java del sistema. `ensureOfficialJava` usa también el fallback `jre-legacy`.
- La herencia de `javaVersion` desde la versión base (via `inheritsFrom`) ya existía (`mergeVersions`): Forge 1.12.2 no declara componente → cae en jre-legacy; perfiles modernos heredan `java-runtime-delta` del vanilla.

### 5. Fallback `-universal.jar` para Forge antiguo

- `internal/Core/Launcher/Launcher.go`: en `downloadMissingLibraries`, si la descarga de una librería `net.minecraftforge:forge:*` falla y la URL apunta al jar "plain" (`forge-X.jar`, que devuelve 404 en maven), se reintenta con **`forge-X-universal.jar`** (nuevo helper `universalForgeURL`). Esto arregla el 404 en el lanzamiento de instancias con Forge 1.12.2.

### 6. Sin carpeta ni archivos de estado del modloader

`internal/Core/ModLoader/Orchestrator.go`:

- Eliminada la persistencia en disco: ya no existe la carpeta `loader-states/` ni `loader-state-*.json`.
- El estado vive en una **caché en memoria** (`stateCache`) con tombstones (`stateRemoved`) tras `RemoveState`; `LoadState` consulta la caché y, si no hay, **deriva el estado del disco** (`deriveFromDisk`): escanea `instancePath/versions/*.json`, toma el primero con `inheritsFrom` y clasifica el loader con `DetectLoaderFromID` (fabric/legacyfabric/quilt/neoforge/forge por el id del json) y extrae la versión con `extractLoaderVersion` (quita el prefijo `<mc>-<loader>-`).
- `NewOrchestrator` **borra la carpeta `loader-states/`** que pudiera quedar de versiones anteriores; `saveState` solo cachea (y limpia el `loader-state.json` legacy); `RemoveState` limpia caché + tombstone + archivo legacy.

### 7. Librerías de instancia movidas a shared

- `internal/Core/Launcher/Instance/Manager.go`: nuevo **`mergeInstanceLibraries(instPath)`** que traslada `<instancia>/libraries` → `shared/libraries` (salta archivos ya presentes con el mismo tamaño; si el destino difiere, lo reemplaza; fallback a copia si el rename cruza volúmenes) y, solo si no hubo fallos, **elimina la carpeta** de la instancia.
- Se llama en `LaunchInstance` (antes de lanzar) y al terminar `InstallModLoader` (`Instance/Modloader.go`). Las libs que dejaba el instalador del modloader quedaban como peso muerto en la instancia.

### 8. Toggle "Carpeta de juego separada": feedback visual de deshabilitado

- `frontend/web/src/Styles/Shared/Settings.scss`: nuevo estado `input:disabled + .SsTgS` (opacidad + `cursor: not-allowed`). El toggle se deshabilita en modo Minecraft pero parecía inerte porque no había diferencia visual; ahora se ve apagado.
- `GeneralSettings.vue`: `title="No disponible en modo Minecraft"` en el toggle.

## Por qué

El usuario reportó que la instancia Criztx (Forge 1.12.2) no arrancaba: el launcher usaba Java 25 del sistema (Forge 1.12.2 solo funciona con Java 8), el version.json de Forge apunta al jar `forge-X.jar` que ya no existe en maven (404) y el widget/panel desaparecía al instalar modloaders porque esos eventos no pasaban por el store central. Además pidió explícitamente: sin archivos ni carpetas de estado para el modloader (estado en memoria o derivado), mover las librerías de la instancia a shared y eliminar la carpeta, un botón para reabrir el panel de descargas desde la instancia, y arreglar el "MB MB" duplicado. El error de compilación darwin (`syscall.SysctlUint64` indefinido) se corrigió en el change anterior y se mantiene verificado con `GOOS=darwin`.

## API afectada

- `internal/Core/ModLoader/Orchestrator.go`: eliminados `stateDir`/`statePath` y la escritura/lectura de `loader-state-*.json`; nueva caché en memoria + `deriveFromDisk`/`DetectLoaderFromID`/`extractLoaderVersion` (exportado `DetectLoaderFromID`). Firmas públicas (`Install`, `LoadState`, `GetInstalledLoader`, `RemoveState`, `BuildExecution`) sin cambios.
- `internal/Core/Launcher/Helpers/Java.go` y `Downloader.Tasks.BuildJavaRuntimeTasks`: componente vacío → `jre-legacy` (sin errores).
- `internal/Core/Launcher/Instance/Manager.go`: `mergeInstanceLibraries` (nuevo, interno).
- Frontend: `Stores/Downloads.panelVisible` (nuevo), `Stores/Instances.allActiveDownloads/anyAllActive` (nuevos); sin cambios en bindings de Wails.

## Comportamiento anterior/nuevo

- **Antes**: "25,6MB/26,6 MB MB"; el widget desaparecía al instalar un modloader; el widget de instancia no mostraba MB ni permitía abrir el panel; Forge 1.12.2 crasheaba al instante con Java 25 del sistema; 404 en el jar de forge al lanzar; carpeta `loader-states/` + `loader-state-*.json` en el workDir; `libraries/` duplicada dentro de cada instancia con modloader; toggle deshabilitado sin feedback visual.
- **Ahora**: MB correctos; el widget y el panel cubren descargas Y modloaders (resolviendo/descargando/instalando) con sus mensajes; botón "Panel" + MB/archivos en la instancia; versiones antiguas usan automáticamente Java 8 oficial (jre-legacy) si el sistema tiene Java >= 17; descarga de librerías con fallback `-universal.jar`; estado del modloader solo en memoria/derivado del disco (sin archivos sueltos; la carpeta legacy se borra al arrancar); librerías de instancia movidas a shared y carpeta eliminada; toggle deshabilitado visible con tooltip.

## Cómo verificar

- `go build ./...` en la raíz: OK; `GOOS=darwin go build ./internal/Core/Platform/`: OK.
- `bun run build` en `frontend/`: type-check y vite build OK.
- Manual: instalar un modloader en una instancia (el widget y el panel siguen visibles durante resolving/downloading/installing con sus estados; `libraries/` de la instancia se vacía y desaparece); lanzar Criztx (Forge 1.12.2): descarga `runtime/jre-legacy/windows-x64` (Java 8) en el primer arranque y el juego inicia; si el jar de forge falta en shared, se descarga con `-universal.jar`; desde el detalle de una instancia con descarga activa, pulsar "Panel" abre el panel global; verificar que no existen `loader-states/` ni `loader-state.json` en el workDir ni en las instancias; en ajustes con modo Minecraft el toggle "Carpeta de juego separada" se ve atenuado con tooltip.
