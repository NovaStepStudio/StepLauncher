# StepLauncher-Change-35.md

## Fecha
2026-08-06

## Release
StepLauncher-2.3.0 — se mencionó por primera vez en esta release.

## Cambio
Mejoras de la descarga y el arranque: (1) descarga de modloaders con el repositorio Maven correcto para Forge/NeoForge, (2) el botón "Jugar" muestra el progreso de descarga de librerías faltantes en vez de un "Lanzando..." fijo, (3) el modal de instalación ya no pasa a "done" al terminar de verificar la descarga del juego: sigue centrado en la instalación del ModLoader con barra de progreso y log en vivo, y (4) nueva opción "Verificación de integridad de archivos" (SHA1) en Ajustes → Comportamiento.

## Qué pasaba
- Forge/NeoForge modernos dejaban 1-2 librerías sin URL (`artifact.url=""`); el fallback fijo a `libraries.minecraft.net` daba 404 y el jar patcheado faltaba (crash `Could not find .forge_patched_minecraft`, ver Error-16).
- Al pulsar Jugar con librerías faltantes el botón mostraba "Lanzando…" estático durante la descarga en segundo plano, sin feedback.
- En el modal de instalación, al terminar la descarga del juego se pasaba a instalar el loader sin apenas información ("se está instalando en segundo plano…").

## Qué se hizo
1. **Repositorio de descarga por grupo Maven** (`internal/Core/Downloader/Helpers.go`, `Tasks.go` y `internal/Core/Launcher/Helpers/Classpath.go`):
   - `LibraryRepositoryBase` elige `maven.minecraftforge.net` para los grupos de Forge/NeoForge y `libraries.minecraft.net` para el resto.
   - `addLibraryTasks` y `ResolveLibraryDownload` resuelven artifacts con `url==""` pero `path`/`sha1`/`size` presentes (se conserva la verificación SHA1).
2. **Progreso de preparación al pulsar Jugar**:
   - Backend: nuevo evento `game_prepare` (`internal/Core/Launcher/Events.go`) emitido desde `downloadMissingLibraries` (fase, current/total, label, finished).
   - Frontend: `stores/launcher.ts` (estado `launchPrepare` + suscripción perezosa) y `App.vue` (el botón alterna "Descargando… / Lanzando…" y bajo el botón se muestra "Descargando archivos faltantes (n/total)…").
3. **Modal de instalación enfocado en el ModLoader** (`InstallationModal.vue`):
   - Al terminar la descarga del juego con loader pendiente ya no pasa a "done": sigue en `installing` mostrando el ring en busy.
   - Nueva barra de progreso del loader (`loaderProgress`/`loaderTotal`) y log en vivo (`loaderLogs`, sin duplicados) con las líneas de los eventos `modloader_*`.
   - Footer del loader: "Instalando {loader} {version} · {estado}" en lugar del aviso genérico.
4. **Verificación de integridad configurable**:
   - `internal/Config/Config.go`: `verifyIntegrity` como `*bool` (nil → activado; `VerifyEnabled()`), `SetVerifyIntegrity`, sanitización y helper `boolPtr`.
   - `internal/Handlers/Engine/engineconfig/Config.go`: campo `VerifyIntegrity` (default true), `Engine.SetVerifyIntegrity`, y `StartFullDownload` usa `skipVerify = !VerifyIntegrity`.
   - `internal/Handlers/App.go` y `app.go`: binding `SetVerifyIntegrity` + sincronización del engine en `Startup` y `ResetConfig`.
   - Frontend: toggle "Verificación de integridad de archivos" en `GeneralSettings.vue` (Comportamiento).

## API afectada
- Binding Wails nuevo: `App.SetVerifyIntegrity(v bool)` (se regenera con `wails build`/`wails dev`).
- `launcher_config.json`: nuevo campo `verifyIntegrity` (ausente → activado con un valor por defecto true).
- Evento Wails nuevo: `game_prepare`.

## Cómo verificar
- `go build ./...` OK. `bun run build` (con type-check) OK.
- Borrar `libraries/net/minecraftforge/forge/26.2-65.1.0/forge-26.2-65.1.0-client.jar` y pulsar Jugar en una versión Forge/NeoForge moderna: debe descargar desde `maven.minecraftforge.net` y lanzar sin el error de `.forge_patched_minecraft`.
- Pulsar Jugar con librerías faltantes: el botón muestra "Descargando…" y la progresión (n/total) bajo él.
- Instalar MC + Forge: al terminar verificar, se sigue viendo la barra y el log de instalación del loader hasta "instalado".
- Ajustes → Comportamiento: desactivar "Verificación de integridad de archivos" → `StartFullDownload` pasa `skipVerify=true`; reactivarla → vuelve a verificar SHA1.

## Notas
- El `*bool` evita que una config guardada sin el campo se interprete como "desactivado": por defecto se trata como true hasta que el usuario lo toque.
- Los eventos `game_prepare` solo se reenvían al frontend mientras `launching` (suscripción perezosa mediante `window.runtime`).