# Changes/StepLauncher-2.3.1/StepLauncher-Change-11.md

- **Fecha**: 2026-08-09
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado en runtime (instalación completa de Forge y NeoForge con los jars parcheados generados).

## Qué cambió

Instalación de Forge/NeoForge modernos nueva vía: cuando el `install_profile.json` declara `processors`, el launcher ahora ejecuta el **instalador oficial** de Forge/NeoForge con `--installClient` en lugar de intentar replicar la instalación solo con extracción + descargas (los jars parcheados solo los generan los procesadores del instalador). Además se resuelve **qué Java usar** con prioridad (runtime oficial de Mojang → Java del launcher → Java del sistema) y el estado de la instancia usa el **id real del version.json** escrito por el instalador.

### 1. Ejecución del instalador oficial (backend)

- Nuevo `internal/Core/ModLoader/Installer/JavaInstaller.go`:
  - `RunInstallerJar(javaPath, installerJar, installDir, logFn)`: ejecuta `java -jar <instalador> --installClient <dir>` reenviando stdout/stderr línea a línea a `logFn` (visible en el modal vía eventos `modloader_*`); error si el exit code != 0. Crea `launcher_profiles.json` (`{}`) si no existe (ClientInstall de Forge/NeoForge lo valida).
  - `LookupInstalledJsonID(instancePath, loaderName, loaderVersion)`: busca en `versions/` el `id` real escrito por el instalador (Contiene loaderName + loaderVersion) para usarlo en el estado.
- `Installer/Executor.go`: `InstallProfile` ahora parsea `processors` y `ExecuteInstaller` devuelve además el `jsonID` real y `hasProcessors` (perfil con procesadores).
- `Provider/Forge.go` (AbstractForgeProvider, compartido con NeoForge): si el perfil tiene procesadores → `runOfficialInstaller`: resuelve Java (`JavaResolver`) indicando en el log qué Java usa, ejecuta el instalador oficial y hace el gap-fill de librerías del perfil que ya existía. Si no hay Java → error claro. **Sin check fijo de versión de Java**: cada instalador exige la suya (Forge 1.12.2 Java 8, 1.20.x Java 17, 1.21+ Java 21) y el propio instalador reporta si el Java no sirve.
- `Provider/Neoforge.go`: elimina el `RunInstaller` duplicado (hereda el del Abstract) y ambos constructores reciben `javaResolver`.
- `Orchestrator.Install`: tras la instalación, `LookupInstalledJsonID` ajusta `VersionJsonID` del estado al id real de disco (fix para `BuildExecution` y `ProfileVersion`).

### 2. Resolución de Java por prioridad

- `Engine.go` + `Helpers/Java.go`: nuevo `javaResolver(mcVersion, instancePath)` con orden: (1) `ResolveMinecraftJava` — el **Java oficial que el launcher ya descargó para la versión base de MC** (`javaVersion.component` de `versions/<mc>/<mc>.json` → `runtime/<component>/<osKey>/bin/java`), el Java exacto que Mojang eligió para esa versión y compatible con el instalador del modloader; (2) **Java configurado en el launcher** (`JavaCustomPath`); (3) **Java del sistema** (`javaw`/`java` de PATH). El runtime oficial se descarga al primer lanzamiento (flujo `ensureOfficialJava`), así que el launcher no exige Java en el sistema para instalar modloaders ya jugados.

## Por qué

- Los jars parcheados (`net/minecraftforge/forge/...-client.jar` en Forge, `net/neoforged/minecraft-client-patched` en NeoForge) NO existen descargables en ningún maven (verificado con HEAD requests reales: 404 en `maven.minecraftforge.net` y `maven.neoforged.net`); solo los generan los procesadores (`binarypatcher` con `data/client.lzma` en Forge; `PROCESS_MINECRAFT_JAR` de installertools en NeoForge), así que la instalación previa quedaba incompleta.
- El id del `version.json` lo decide el instalador (`26.2-forge-65.1.0`, no `forge-26.2-65.1.0`): sin leerlo de disco, `BuildExecution`/`ProfileVersion` apuntaban a un version.json inexistente.
- Ver `Investigacion/ModLoaders/` (instaladores decompilados con CFR) + `Changelogs/Errors` del subsistema (Error-13/14/16/18 y Change-34/35) para el historial previo.

## API afectada

- Backend Go: `JavaInstaller.go` (nuevo); `Executor.go`, `Forge.go`, `Neoforge.go`, `Orchestrator.go`, `Engine.go`, `Helpers/Java.go` (modificados).
- Frontend: sin cambios (los logs del instalador llegan por los eventos `modloader_*` existentes).

## Comportamiento anterior/nuevo

- Anterior: instalación incompleta (faltaban los jars parcheados, el juego no arrancaba) y sin requisito de Java resuelto.
- Nuevo: Forge y NeoForge modernos se instalan completos ejecutando el instalador oficial con el mejor Java disponible; el log del modal muestra la salida del instalador y qué Java se usó.

## Cómo verificar

- `go build ./...` sin errores.
- Verificado en runtime (2026-08-09): instalar Forge 26.2-65.1.0 y NeoForge 26.2.0.57 en una instancia con MC 26.2 descargado genera los jars parcheados y el juego arranca con el modloader.

## Pendientes (opcional futuro)

- Si no hay Java en disco ni en el sistema y la MC base nunca se lanzó (runtime no descargado), el launcher podría descargar el runtime oficial (reutilizando `downloader.BuildJavaRuntimeTasks` + `ensureOfficialJava`) durante la instalación; hoy falla con un error claro en el log.