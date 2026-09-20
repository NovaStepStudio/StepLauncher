# Changes/StepLauncher-2.3.1/StepLauncher-Change-18.md

- **Fecha**: 2026-08-12
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado (build verificado; verificación en runtime pendiente).

## Qué cambió

Dos mejoras: (1) el **sistema de caché** almacena, carga y limpia de forma más eficiente (se corrige que la categoría `default` —donde viven los JSON de meta de Fabric/Quilt/Forge— quedara fuera de `Info`/`Clear`/`cleanup`, se limpian los instaladores JAR y sus logs por TTL, se escribe atómicamente y el reporte incluye tamaños), y (2) el **sistema de crasheo** deja de mostrar códigos de error: el modal muestra las **rutas de los logs** (launcher, Minecraft y JVM) con un botón para **abrirlos directamente**, y expone el **resumen de lanzamiento** completo (todo lo que arma el `game_log`: version, main class, directorios, memoria, HW accel, classpath `[OK]/[MISSING]` y argumentos numerados con `[REDACTED]`) en texto plano sin separadores ASCII.

## Reporte general de la carpeta de caché (antes del cambio)

`%appdata%/.StepLauncher/cache` (datos leídos del disco el 2026-08-12):

| Carpeta | Archivos | Tamaño | Uso |
|---|---|---|---|
| `default/` | 5 | 3.64 MB | Meta de modloaders (fabric-versions-*, quilt, forge-meta) |
| `modloader/` | 2 | 13.15 MB | Instaladores NeoForge (JAR) |
| `modloader-logs/` | 2 | 1.21 MB | Logs de instaladores |
| `manifest/` | 2 | 0.51 MB | Manifiesto de versiones |
| `java/` | 2 | 0.21 MB | Productos/manifiestos de Java |
| `assets/` | 1 | 0.57 MB | Índices de assets |
| `versions/` | 1 | 0.05 MB | JSON de versiones |
| `fabric/ forge/ quilt/ legacyfabric/ neoforge/` | 0 | 0 | Vacías (creadas por `ensureDirs`, sin uso real) |

Problemas encontrados: la categoría `default` concentraba los JSON más grandes pero **no estaba en `subdirs()`**, así que `Info()`, `Clear()` y `cleanup()` la ignoraban (nunca se limpiaban ni se contaban); `modloader/` (13 MB en JARs) y `modloader-logs/` tampoco se limpiaban jamás; `Set` escribía directo (riesgo de corrupción si se cortaba a mitad); y `Info` no reportaba tamaños.

## Detalle de los cambios

### 1. Cache: categoría `default` cubierta y limpieza por TTL de instaladores

- `internal/Core/Cache/Cache.go`: `subdirs()` ahora incluye `"default"` (es el destino del fallback de `Downloader/Fetch.go:cacheCategory`, donde `fetchJSONWithCache` guarda fabric/quilt/forge). `Clear()` y `DeleteCategory()` también barren las carpetas de artefactos (`artifactDirs()` = `modloader` + `modloader-logs`) sin filtrar por `.json`; `cleanup()` ahora elimina los JARs de instaladores y sus logs por **antigüedad** usando `ttlFor(category)` (modloader: 24 h) con `ModTime`, además de los JSON expirados por `expires_at`.
- `internal/Core/Cache/Cache.go`: `Set` escribe a `path.tmp` y hace `os.Rename` (escritura atómica; borra el tmp si falla).
- `internal/Core/Cache/Cache.go`: `cleanupLoop` ejecuta un `cleanup()` inmediato al arrancar (antes esperaba la primera hora del ticker).
- `internal/Core/Cache/Cache_info.go`: `Info` ahora reporta `totalBytes`, `sizes` por categoría y cuenta/pondera también `default` y los artefactos.

### 2. Crash: logsPath en el evento y resumen de lanzamiento expuesto

- `internal/Core/Launcher/Types.go`: `GameInstance` gana `PreInfo *gamelog.PreLaunchInfo` (el resumen que arma el launcher) y `LauncherLogPath string`.
- `internal/Core/Launcher/Launcher.go`: tras construir el `preInfo` (que ya se escribía al log), se conserva una copia en `instance.PreInfo`.
- `internal/Core/Launcher/Events.go`: `GameEventData` ahora incluye `launcherLogPath`, `minecraftLogPath`, `jvmLogPath` y `launchInfo` (texto formateado del resumen). `NewGameEventData` los rellena desde la instancia.
- `internal/Core/Launcher/Log/Game_log.go`: nuevo `FormatPreLaunchInfo(info)` que produce el resumen en **texto plano sin separadores ASCII** (mismo contenido que el log: config de versión/dirs/memoria/HW accel, `Total entries` con `[OK]/[MISSING]` por librería, y comandos completos numerados `[  0] ...` con `[REDACTED]` en accessToken/session/token).
- `internal/Core/Launcher/Utils/Open.go` (nuevo): `OpenInExplorer(path)` abre una ruta (archivo o carpeta) en el explorador según plataforma (explorer/open/xdg-open). `Instance/Folder.go` ahora delega en esta utilidad.
- `internal/Handlers/Engine/Launch.go`: `LaunchMinecraft` setea `inst.LauncherLogPath = e.log.GetLogPath()`; nuevo `GetGameLaunchInfo(id)` que devuelve el resumen formateado de un juego en curso; nuevo `OpenPath(path)`.
- `internal/Handlers/Engine/Instance.go`: `LaunchInstance` setea también `LauncherLogPath` en la instancia lanzada.
- `app.go`: wrappers `GetGameLaunchInfo(id)` y `OpenPath(path)`.

### 3. Frontend: modal de crash sin códigos de error

- `frontend/web/src/Stores/Launcher.ts`: `CrashInfo` y `onGameCrash` incorporan `launcherLogPath`, `minecraftLogPath`, `jvmLogPath` y `launchInfo`.
- `frontend/web/src/Modals/CrashModal.vue`: se elimina la pestaña «Códigos de errores» (y los dos lugares que mostraban el exit code). Nueva pestaña «Logs y lanzamiento» con: cada ruta de log con su etiqueta y botón **Abrir** (llama `App.OpenPath`), el **Resumen de lanzamiento** en `<pre>` (texto plano sin `====`/`----`), y el log de error con su botón de copiar existente.
- `frontend/web/src/Styles/Modals/CrashModal.scss`: estilos nuevos (`CrashModal_LogsList`, `CrashModal_LogRow`, `CrashModal_LogPath`, `CrashModal_OpenBtn`, `CrashModal_LaunchBox`, `CrashModal_LaunchPre`).
- `frontend/web/src/Layouts/Sections/Settings/GeneralSettings.vue`: la fila de caché muestra ahora tamaño total (`X archivos · Y MB`) usando `totalBytes`/`sizes`.

## API afectada

- Bindings nuevos: `App.GetGameLaunchInfo(id) string` y `App.OpenPath(path) error` (requieren regenerar `frontend/wailsjs` con `wails build`).
- `cache.Info` gana `totalBytes` y `sizes` (campos nuevos, compatibles con consumidores actuales).
- `GameEventData` gana 4 campos opcionales; el modal anterior de "códigos de error" se elimina de la UI.

## Comportamiento anterior/nuevo

- **Cache**: anterior: `default/`, `modloader/` y `modloader-logs/` nunca se limpiaban ni aparecían en la UI; escritura no atómica. nuevo: todo se limpia por TTL (JSON por `expires_at`, artefactos por antigüedad), escritura atómica con tmp+rename, cleanup al arrancar y reporte con tamaños.
- **Crash**: anterior: el modal mostraba el código de salida y el texto del log, sin rutas ni botón para abrirlos, y no había forma de ver el resumen de lanzamiento. nuevo: sin códigos, con los 3 `logsPath` y botón «Abrir», resumen de lanzamiento completo con `[REDACTED]` en datos sensibles y sin separadores ASCII.

## Cómo verificar

- `go build ./...` OK; `go vet ./...` OK; `bun run type-check` OK (frontend).
- `wails dev`: lanzar y forzar un crash (p. ej. cerrar la ventana del juego con el launcher oculto) → el modal debe mostrar «Logs y lanzamiento» con las rutas, el botón Abrir debe abrir el explorador en el archivo, y el resumen debe verse sin `====`/`----` y con `[REDACTED]` en accessToken. En Ajustes → Cache debe verse el tamaño total y las categorías `default`/`modloader`/`modloader-logs` si tienen archivos.