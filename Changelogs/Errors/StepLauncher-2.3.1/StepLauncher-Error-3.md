# Errors/StepLauncher-2.3.1/StepLauncher-Error-3.md — Instalación de modloaders legacy: perfil incompleto, instalador no siempre ejecutado y logs no guardados

- **Fecha**: 2026-08-10
- **Versión**: 2.3.1
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release StepLauncher-2.3.1.

## Síntoma

En versiones viejas de Forge (p. ej. 1.8.9) la instalación quedaba incompleta: **algunas versiones arrancan y otras no**. Además, el instalador oficial nunca llegaba a ejecutarse en esos casos y **no se guardaba ningún log** de la instalación. El sistema trataba "el jar del instalador ya está descargado" como "instalación completa".

## Causa raíz (3 fallos)

1. **El flujo daba por buena la instalación solo por tener el jar del instalador en cache** (`internal/Core/ModLoader/Orchestrator.go`, `downloadEntries`: los planos sin SHA1 que ya existen se saltan) y la **ejecución del instalador oficial con Java estaba condicionada a `hasProcessors`** (`Forge.go`): si el `install_profile.json` no declara procesadores (el de Forge 1.8.9 **no los tiene**), el jar nunca se ejecutaba.
2. **El protocolo legacy estaba mal replicado**: el instalador clásico (era 1.7.x-1.8.x, ver descompilación en `Investigacion/ModLoaders/forge-1.8.9-...-installer/`) no trae `maven/` (el de 1.8.9 trae el universal jar en la **raíz** del zip) y el `versionInfo.libraries` no tiene librerías top-level en `install_profile.json`. `ExecuteInstaller` solo copiaba `maven/` + librerías del `libraries` top-level → **nunca se extraía `install.filePath` (forge-...-universal.jar) a su coordenada maven** (`libraries/net/minecraftforge/forge/<ver>/forge-<ver>.jar`, exactamente `VersionInfo.extractFile`/`getLibraryPath` del instalador) → la librería principal del perfil no existía en disco ("algunas no arrancan").
3. **Logs**: como el jar oficial no se ejecutaba en esos casos, nunca se generaba su log (el `<jar>.log` lo crea el propio `SimpleInstaller.setupLogger` al arrancar el proceso); y fuera de la ejecución oficial no se guardaba nada.

## Diagnóstico y evidencia

- Descompilación con CFR del instalador 1.8.9 (`Investigacion/ModLoaders`): `SimpleInstaller.main`/`handleOptions` solo soporta `--installServer`, `--extract`, `--offline` y help — **no existe `--installClient`**; con un flag desconocido jopt-simple lanza excepción → el proceso termina con código ≠ 0 tras imprimir ayuda (y crea el log al inicio de `main`).
- `install_profile.json` de 1.8.9: sin `processors` sin `libraries`, con `install.filePath: forge-1.8.9-11.15.1.2318-1.8.9-universal.jar` y `install.path: net.minecraftforge:forge:1.8.9-11.15.1.2318-1.8.9`; `versionInfo.libraries` con urls maven legacy.
- `ClientInstall.run` (descompilado): copia/extrae `filePath` a `getLibraryPath(librariesDir)` + escribe el version.json desde `versionInfo` + deja el resto de librerías para el lanzador.

## Solución aplicada

- **Protocolo de ejecución obligatorio ("sí o sí")**: `AbstractForgeProvider.RunInstaller` (`Forge.go`) ejecuta ahora SIEMPRE el instalador oficial con Java (`--installClient`), tenga procesadores o no y aunque el jar ya esté en cache. Si el perfil declara procesadores y la ejecución falla → error (la versión no sería jugable); si no declara procesadores (legacy como 1.8.9, cuyo jar no soporta headless de cliente) → se tolera con un aviso en el flujo: la extracción ya dejó la versión lista.
- **Protocolo legacy completo en `ExecuteInstaller`** (`Installer/Executor.go`): nueva sección `install` en `InstallProfile` y extracción del archivo contenido (`install.filePath`) a `libraries/<MavenPath(install.path)>` (mismo destino que el instalador oficial: `net/minecraftforge/forge/<ver>/forge-<ver>.jar`). Extracción refactorizada a `extractZipEntryTo`.
- **Logs**: al ejecutarse el jar oficial siempre, su log (p. ej. `forge-1.8.9-...-installer.jar.log`) se genera siempre y `moveInstallerLogs` (Change-16) lo conserva en `cache/modloader-logs` tanto en éxito como en error (el proceso legacy fallido también deja su log).
- Se mantiene la descarga por cache del jar (SHA1 vacío + archivo existente → no se re-descarga): **guardar/descargar el jar nunca sustituye a ejecutarlo**.

## Regla aprendida

Para Forge/NeoForge, "el jar del instalador está en cache" ≠ "instalado": el protocolo exige ejecutar el instalador oficial. Los instaladores legacy (1.7.x-1.8.x) pueden no soportar `--installClient` y traen el universal jar en `install.filePath` (no en `maven/`): hay que replicar extractFile/getLibraryPath, tolerar el no-op headless cuando no hay procesadores y conservar siempre el log del proceso.

## Verificación

- `go build ./...` OK.
- Comportamiento (pendiente de runtime): con el jar de 1.8.9 en cache, la instalación ejecuta Java con `--installClient` (proceso que termina en no-op), extrae el universal jar a `libraries/net/minecraftforge/forge/.../forge-....jar`, escribe el version.json y guarda el log en `cache/modloader-logs/`.