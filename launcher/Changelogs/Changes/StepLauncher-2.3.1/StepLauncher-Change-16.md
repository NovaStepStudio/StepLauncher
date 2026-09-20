# Changes/StepLauncher-2.3.1/StepLauncher-Change-16.md

- **Fecha**: 2026-08-10
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado (build verificado; verificación en runtime pendiente).

## Qué cambió

Los modales de instalación permiten reinstalar una versión ya descargada (avisando de que se redescargará), los logs de los instaladores oficiales de Forge/NeoForge se conservan en la carpeta cache del launcher (ya no se borran), y el RichPresence deja de reintentar al cliente de Discord tras 5 fallos consecutivos hasta que el usuario vuelve a activar la presencia.

### 1. Reinstalación de versiones ya instaladas en ambos modales

- `frontend/web/src/Modals/InstanceDownloadModal.vue`: se elimina el bloqueo de versiones ya instaladas en la instancia (`isInstalled` ya no deshabilita la fila ni frena `installReady`). La fila pasa a la clase `reinstalling` (borde ámbar, cursor normal) y el badge cambia de "INSTALADA" a "REINSTALAR" con tooltip "Ya está instalada: se volverá a descargar (reinstalación)". Si la versión seleccionada ya está instalada, aparece un aviso `.InstDl_Notice.info`: "Esta versión ya está instalada en la instancia: se volverá a descargar (reinstalación)".
- `frontend/web/src/Modals/InstallationModal.vue` (flujo global): badge "REINSTALAR" + clase `reinstalling` cuando la versión ya figura en `installedVersions` (store Launcher) y aviso `.InstallationModal_Notice.info` equivalente al seleccionarla.
- `pickDefaultVersion` (modal de instancia): si todas las versiones están instaladas, cae a la primera del manifest en lugar de quedar vacío.
- Backend: no bloquea la re-descarga — `AddInstanceVersion` (`internal/Core/Launcher/Instance/Download.go`) arranca siempre un nuevo flujo y `addVersionToMetadata` es idempotente (no duplica versiones en el metadata). El verificador ya salta archivos válidos por SHA1 y redescarga los corruptos/ausentes.
- SCSS: `.InstDl_Version.reinstalling`/`.InstallationModal_Version.reinstalling`, badges `.installed` en ámbar (warning) y variantes `.info` de `.InstDl_Notice`/`.InstallationModal_Notice`.

### 2. Logs de los instaladores oficiales → carpeta cache (antes se borraban)

- `internal/Core/ModLoader/Installer/JavaInstaller.go`: `cleanInstallerLogs` se sustituye por `moveInstallerLogs(installDir, logsCacheDir, installerJar)`, que **mueve** `installer.log`, `<jar>.log` y `<jar>.jarlog` a la carpeta cache del launcher (crea `cache/modloader-logs`). `moveWithRetry` tolera el retardo de Windows en liberar el handle (5 intentos, 100 ms) y usa rename con fallback copia+borrado (p. ej. otro volumen). Si `logsCacheDir` está vacío, borra como antes (best-effort; nunca condiciona la instalación).
- `internal/Core/ModLoader/Provider/Forge.go`: `runOfficialInstaller` pasa `filepath.Join(p.CacheDir, "modloader-logs")` a `RunInstallerJar` (cubre Forge y NeoForge: ambos comparten `AbstractForgeProvider`).
- Sustituye a la limpieza documentada en `StepLauncher-Change-13.md` (sección 2): los logs ya no se eliminan, se conservan en cache.

### 3. RichPresence: tope de reintentos y suspensión hasta re-activación

- `internal/RichPresence/RichPresence.go`: nueva constante `maxDialAttempts = 5` y flag `suspended`. En `loop()`, tras 5 fallos de conexión consecutivos se loguea "se detienen los reintentos hasta volver a activar la presencia" y el hilo entra en espera (sin volver a tocar el pipe de Discord). `SetEnabled(true)` (arranque, toggle del usuario o reaplicación de config) limpia `suspended` y avisa por `wake` → se reanuda el ciclo. `SetEnabled(false)` también despierta el ciclo para que vuelva al estado "desactivado" normal.
- Efecto: con Discord cerrado, el launcher ya no acumula cientos de intentos de conexión ni spam de logs.

## Verificación

- `go build ./...` OK.
- `bun run build` (frontend) OK.
- Comportamiento: al seleccionar una versión ya instalada se puede pulsar instalar y aparece el aviso de re-descarga; tras instalar Forge/NeoForge los `*.log` quedan en `cache/modloader-logs/`; con Discord cerrado se detienen los intentos tras 5 fallos y se reanudan al activar/desactivar la presencia.