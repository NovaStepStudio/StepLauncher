# Changes/StepLauncher-2.3.1/StepLauncher-Change-15.md

- **Fecha**: 2026-08-10
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado (build verificado; verificación en runtime pendiente).

## Qué cambió

Soporte de instaladores legacy de Forge, progreso en vivo de la verificación de archivos y presencia visual del modloader en instalación en banner, tarjetas, detalle y modales (el anillo desaparece cuando arranca el instalador y su contenido toma el protagonismo, con advertencia de consumo de recursos y logs en vivo).

### 1. Progreso en vivo de la verificación (X/Y archivos)

- `internal/Core/Downloader/Verify.go`: nueva `VerifyBatchWithProgress(tasks, maxWorkers, onProgress(done, total))` que notifica por archivo (incluidos los sin SHA1); `VerifyBatch` delega en ella sin callback.
- `internal/Core/Downloader/Manager.go` (`verifyAll`): en la fase `StateVerifying` y en la re-verificación tras `StateReDownload` se setean `FilesTotal`/`FilesDownloaded` y se emite progreso con throttling (`done%25==0 || done==total`).
- Frontend: contador "X/Y archivos verificados" en `InstanceDownloadModal.vue` y `InstallationModal.vue` (clases `InstDl_VerifyCount` / `InstallationModal_VerifyCount`) durante las fases de verificación y re-descarga.

### 2. Presencia del modloader en instalación (instancias)

- `frontend/web/src/Stores/Instances.ts`: nuevo estado `loaderDls` por instancia (fase resolving/downloading/installing/done/error + progreso) alimentado por los eventos `modloader_*`; `registerLoaderSession(name, sessionId, loader, loaderVersion, mcVersion)` mapea la sesión del backend a la instancia (llama `ensureEvents()` y siembra el estado con `phase: resolving`); limpieza 8 s tras `installed`/12 s tras `error` (+ `loadInstalledLoader` al instalar); `isInstanceBusy(name)` (lanzando, descarga activa o loader activo) y `loaderDlStateText(ld)`.
- `InstanceDownloadModal.vue`: al recibir `sessionId` de `installInstanceModLoader` se registra la sesión → la instalación del loader se ve aunque se cierre el modal.
- `InstanceDetailView.vue`: el overlay sobre el banner soporta ahora el bloque de descarga (con contador de verificación y Cancelar) y el bloque del modloader (fase, %, barra indeterminada, error, meta loader+MC); los botones del hero usan `opBusy` (renombrado de `busy` por conflicto con el `ref` de capturas, línea ~158) y cambian de texto (Jugar/Descargando…/Instalando…).
- `InstancesView.vue` (grid): chip `InstCard_Ldr` con pulso + texto + mini barra cuando el loader está en curso, "Verificando · X/Y" en `InstCard_DlMeta` y botón Jugar bloqueado con "Ocupada…".

### 3. Fase del instalador en los modales (anillo oculto + advertencia + logs)

- En `InstallationModal.vue` e `InstanceDownloadModal.vue` el anillo de progreso queda envuelto en `<template v-if="loaderPhase !== 'running'">`: cuando arranca el proceso del instalador, este desaparece y la vista dedicada toma todo el ancho.
- Nueva vista con `.LoaderWarn` (aviso: "Instalador independiente y de consumo alto. Forge/NeoForge pueden pasar de los 600 MB. Ten paciencia y no cierres la aplicación."), `.LoaderStatus.running` a ancho completo, `.LoaderBody` (width 100%, min-height 9rem) con barra + línea "Descargando parte X de Y" (`.LoaderParts`) y `.LoaderLog` (max-height 13rem, siempre renderizado) con las líneas en vivo del proceso.
- SCSS: `.InstallationModal_LoaderWarn`/`.InstDl_LoaderWarn`, `.LoaderStatus.running`, `.LoaderBody`/`.LoaderLog` ampliados, `.LoaderParts` y `.VerifyCount` en ambos modales.

## Verificación

- `go build ./...` OK.
- `bun run build` (frontend) OK (el único fallo de vite, conflicto de identificador `busy`, se resolvió renombrando el computed a `opBusy`).
- Comportamiento: durante el instalador de Forge/NeoForge el anillo desaparece, aparece la advertencia y se ven las líneas del proceso; la verificación muestra X/Y archivos.

## Relacionado

- `Changelogs/Errors/StepLauncher-2.3.1/StepLauncher-Error-2.md` (soporte legacy: `install_profile.versionInfo` como fallback de `version.json` en `Executor.go`).