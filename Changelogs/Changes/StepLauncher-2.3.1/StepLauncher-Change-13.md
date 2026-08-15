# Changes/StepLauncher-2.3.1/StepLauncher-Change-13.md

- **Fecha**: 2026-08-09
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado (build verificado; verificación en runtime pendiente).

## Qué cambió

Al instalar un modloader en una instancia ya se registra realmente en el metadata de la instancia, y el directorio de trabajo queda limpio de los logs que dejan los instaladores oficiales de Forge/NeoForge.

### 1. Registro del modloader en el metadata de la instancia

- Nuevo `syncVersionsFromDisk(name)` en `internal/Core/Launcher/Instance/Helpers.go`: relee el directorio `versions/` de la instancia y registra en `instance.metadata.json` (campo `versions`) todas las carpetas presentes que todavía no constan y que tienen su `<version>.json` en disco.
- `internal/Core/Launcher/Instance/Modloader.go` (`InstallModLoader`): al terminar **exitosamente** la instalación del modloader (goroutine), se invoca el sync y se loguea el resultado. Hasta ahora la goroutine solo logueaba el error y el directorio de versión creado por el instalador (p. ej. `26.2-forge-65.1.0`) nunca se registraba: el metadata solo tenía la versión base, porque el registro ocurría al completar la **descarga** de la versión (mucho antes de que el instalador del modloader terminara sus pasos).
- La lectura de `versions/` se hace SOLO al terminar la instalación completa (descarga base + modloader), que es cuando la carpeta ya no está en escritura continua y el instalador ya escribió su `version.json` real.

### 2. Limpieza de logs de los instaladores oficiales

- `internal/Core/ModLoader/Installer/JavaInstaller.go`: nueva `cleanInstallerLogs(installDir, installerJar)` llamada después de `cmd.Wait()` (éxito o error). Elimina los restos que los instaladores oficiales dejan en el directorio de trabajo (`SimpleInstaller.getLog`): `<instalador>.jar.log` (p. ej. `forge-26.2-65.1.0-installer.jar.log`), `<instalador>.jarlog` (p. ej. `neoforge-21.11.45-installer.jarlog`) e `installer.log` (fallback del instalador).
- `removeWithRetry`: reintenta el borrado (5 intentos, 100 ms) por el retardo de Windows en liberar el handle del proceso java recién salido. Best-effort: nunca condiciona el resultado de la instalación.
- Cubre ambos flujos porque `RunInstallerJar` ejecuta con `cmd.Dir = installDir`: flujo de instancia (`<WorkDir>/instances/<nombre>/`) y flujo del modal (`WorkDir`, donde quedaba contaminado el `.StepLauncher`).

### 3. Interfaz: el modloader instalado se ve en las instancias

- `frontend/web/src/Stores/Instances.ts`: nuevo estado `loaders` por instancia + `loaderOf(name)`, `loadInstalledLoader(name)` y `loaderLabel()` (mapa `LOADER_LABELS` tipo → nombre). `loadDetails` refresca también el modloader del backend (`GetInstalledInstanceModLoader` → `loader-state.json`) en un paso independiente con su propio try/catch.
- `frontend/web/src/Modals/InstancesView.vue` (grid): cada tarjeta muestra el chip del modloader instalado (icono + nombre, tooltip con versión del loader y MC).
- `frontend/web/src/Modals/InstanceDetailView.vue`:
  - Hero: chip con icono y nombre del modloader.
  - Resumen: tile "Modloader" con icono, nombre y versión (o "Vanilla" si no hay).
  - Pestaña Versiones: la fila de la versión del loader (coincide con `versionJsonId` o contiene el tipo del loader) muestra su icono y el tag "Modloader"; el subtítulo indica "Versión de <loader>".
- `frontend/web/src/Modals/InstancesModal.vue`: el modal de descarga/instalación ahora emite `done` → `loadDetails(instancia)` para que el detalle se refresque al terminar la instalación del modloader (antes el evento no estaba conectado y la vista quedaba con datos viejos).
- Estilos en `Styles/Modals/InstancesView.scss` (`.InstCard_LoaderChip img`) y `Styles/Modals/InstanceDetailView.scss` (`.InstDet_LoaderChip img`, `.InstDet_TileLoaderImg`, `.InstDet_VersionLoaderIcon`, `.InstDet_VersionTag.loader` usando `--color-tag`), respetando las variables CSS existentes.

## Por qué

- La descarga base registra la versión nada más completarse, pero la instalación de un modloader tiene más pasos (extracción del instalador, descargas de librerías, procesadores con el instalador oficial): el directorio `versions/` del loader se crea al final, cuando el metadata ya se escribió. La instancia "tomaba como referencia solo la versión" porque nadie releía la carpeta tras el cierre de la instalación completa.
- Forge/NeoForge (decompilados en `Investigacion/ModLoaders/`, `SimpleInstaller.getLog()` en ambos) escriben `<nombre-del-jar>.log` en el cwd del proceso java (o `installer.log` si no se puede resolver el jar): esos logs quedaban huérfanos en la raíz de la instancia / `.StepLauncher` tras cada instalación.

## API afectada

- Backend Go: `Helpers.go` y `Modloader.go` (Instance), `JavaInstaller.go` (ModLoader/Installer). Sin cambios en bindings ni config.
- Frontend: `Stores/Instances.ts`, `Modals/InstancesView.vue`, `Modals/InstanceDetailView.vue`, `Modals/InstancesModal.vue` y sus hojas SCSS en `Styles/Modals/`.

## Comportamiento anterior/nuevo

- Anterior: tras instalar Forge/NeoForge en una instancia, `instance.metadata.json` solo listaba la MC base (`26.2`); la versión del loader (`26.2-forge-65.1.0`) existía en disco pero no en la UI, y quedaban `*.jar.log`/`*.jarlog`/`installer.log` en el directorio de trabajo.
- Nuevo: al terminar la instalación del modloader, el metadata registra todas las versiones de `versions/` (base + loader) y los logs del instalador se eliminan automáticamente del directorio donde se ejecutó.

## Cómo verificar

- `go build ./...` sin errores y `bun run build` (frontend, incluye `vue-tsc`) sin errores.
- Runtime (pendiente): instalar un modloader con procesadores (Forge/NeoForge) en una instancia → al completar, la instancia lista la versión del loader, el detalle muestra el modloader (chip, tile y tag) y no quedan logs `*.jar.log`/`installer.log` ni en la instancia ni en la raíz del WorkDir.