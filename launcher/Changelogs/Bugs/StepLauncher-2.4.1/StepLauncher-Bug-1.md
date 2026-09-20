# Bugs de StepLauncher 2.4.1 (Bug-1) — Java del launcher e instancias: selector, herencia de config, refresco de modloader y loader pendiente

- **Fecha**: 2026-08-16
- **Versión**: 2.4.1 (en desarrollo)
- **Estado**: corregido y verificado
- **Release**: en desarrollo — aún no mencionado en ninguna release.

## El bug en cuestión

Tanda de 6 bugs relacionados con la elección de Java y el flujo de instancias, reportados por el usuario en una misma sesión:

### 1. El selector de Java muestra entradas de las variables del sistema

`DetectJavaInstallations()` (`internal/Handlers/Engine/Java.go`) escaneaba `JAVA_HOME` y `PATH` además de las carpetas conocidas, así que el selector de Java de Ajustes podía ofrecer "instalaciones" que no son instalaciones reales (p. ej. una ruta de un JDK embebido de otra aplicación, o `javapath` de Windows).

### 2. El modo "system" ignora el Java elegido y usa `javapath`

En `buildBaseLaunchConfig` (`internal/Handlers/Engine/Launch.go`) el caso `"system"` ponía la bandera muerta `adv.UseSystemJava = true`, que nadie leía (el campo ni siquiera se serializaba en `AdvancedConfig`). El `javaCustomPath` elegido por el usuario en Minecraft.vue se ignoraba: `resolveCustomJava("")` resolvía por `LookPath("javaw")` y caía en el `javapath` de Windows, no en el Java elegido.

### 3. La instancia con "utilizar el java del launcher" no consulta la config global

`LaunchInstance` (`internal/Core/Launcher/Instance/Launch.go`) con `useOfficialJava=true` (el valor por defecto de la UI) lanzaba con el runtime oficial de Mojang, sin consultar la config global del launcher. Si el usuario tenía elegido un Java custom o el modo "system", la instancia lo ignoraba.

### 4. El modloader instalado no aparece hasta reiniciar

Al recibir `modloader_installed`, `Instances/Store.ts` solo refrescaba el loader instalado (`loadInstalledLoader`), nunca el detalle de la instancia ni la lista: las versiones nuevas que crea el instalador (p. ej. Forge/NeoForge) no aparecían en el selector de versión ni en las tarjetas hasta recargar la web.

### 5. La lista de versiones de una instancia se leía de metadata frágil en memoria

`List()`, `Get()` y `Versions()` (`Instance/Manager.go`) devolvían `meta.Versions` del `instance.metadata.json`, que podía quedar desincronizado respecto a la carpeta `versions/` real en disco (p. ej. tras una instalación de modloader o un borrado manual).

### 6. Cerrar el modal durante la descarga cancelaba silenciosamente el modloader

El modloader pendiente vivía solo en el modal (`Download.vue`, `pendingLoaderInstall` + `startLoaderInstall` en `handleGameCompleted`). Si el usuario cerraba el modal mientras descargaba la versión base, la instalación del loader nunca arrancaba: se perdía en silencio.

## Qué afectaba y qué hacía

- **Selector de Java**: opciones irrelevantes y confusas (entradas de `PATH`/`JAVA_HOME`) en Ajustes → Minecraft.
- **Modo system**: el usuario elegía un Java y la instancia/el launcher lanzaba con otro (el `javapath` del sistema), sin aviso.
- **Instancias con "utilizar el java del launcher"**: ignoraban la elección del usuario; solo las instancias con Java propio explícito respetaban algo.
- **Modloader instalado**: la instalación terminaba pero la UI no reflejaba las nuevas versiones; obligaba a reiniciar el launcher.
- **Versiones**: la fuente de verdad estaba en memoria; cualquier desincronización con disco daba listas incorrectas.
- **Loader pendiente**: la instalación de un modloader tras la descarga era frágil y se perdía al cerrar el modal.

## Solución final

### 1. Selector de Java limpio

Se eliminaron los bloques de escaneo de `JAVA_HOME` y `PATH` de `DetectJavaInstallations()`: ahora solo detecta instalaciones reales en las carpetas conocidas (`scanWindowsJava`/`scanUnixJava`).

### 2. Modo "system" usa el Java elegido

En `buildBaseLaunchConfig`, `case "system"` y `case "custom"` copian ahora `ec.JavaCustomPath` a `adv.JavaExec`. Se eliminó el campo muerto `UseSystemJava` de `AdvancedConfig` (verificado con grep: cero referencias). Además, el log de lanzamiento muestra la versión real del Java usado: `l.log("Java: %s (%s)", javaPath, helpers.JavaVersionLabel(javaPath))` con la nueva helper `JavaVersionLabel` (`Helpers/Java.go`), que extrae la versión limpia de `java -version` (o "desconocida" si falla).

### 3. La instancia hereda la config global en vivo

`InstanceManager` ya no depende de `engineconfig` (Core no importa Handlers): se le inyecta un resolver con `SetGlobalJavaConfig(func() (mode, customPath string))` desde `Engine.go`, que lee la config global en el momento de lanzar. En `LaunchInstance`, cuando la instancia tiene activo "utilizar el java del launcher" (`UseOfficialJava == nil || *UseOfficialJava`), el switch aplica: `official` → runtime de Mojang; `system`/`custom` → `adv.JavaExec = customPath`; `auto`/vacío → `LookPath` del sistema. Con `UseOfficialJava = false` se mantiene el `JavaExec` propio de la instancia. Así la elección del usuario se hereda siempre y la auto-conmutación a `jre-legacy` para versiones pre-1.17 sigue intacta.

### 4. Refresco tras instalar modloader

En `Instances/Store.ts`, el caso `modloader_installed` de `updateModLoaderEvent` ahora también refresca `loadDetails(inst)` y `loadInstances()` (además de `loadInstalledLoader`). El watch de `dlStore` también llama `loadDetails` para la instancia que pasa a estado terminal, cubriendo cualquier descarga terminada (vanilla o con loader).

### 5. Versiones leídas del disco

Nueva `scanVersionsFromDisk(name)` en `Instance/Manager.go` que lee la carpeta real `instances/<name>/versions`; `Versions()`, `List()` y `Get()` la usan como fuente de verdad (el metadata se conserva para el resto de campos). Cualquier cambio en disco (instalador de loader, borrado manual) se refleja sin reiniciar.

### 6. Loader pendiente gestionado por el Store

El modloader pendiente ahora vive en el Store (`pendingLoaderByInst` + `scheduleLoaderInstall`/`cancelPendingLoader`), no en el modal. Al pasar la descarga a `completed`, `runPendingLoader` consume la cola y arranca `InstallInstanceModLoader` + `registerLoaderSession`, aunque el modal esté cerrado. El modal deja de arrancar la instalación (`startLoaderInstall` eliminado): solo observa los eventos `modloader_*` filtrados por instancia con el nuevo helper `isLoaderSessionOf` (evita interferencias entre instalaciones de distintas instancias). `onCancel`/`onDownloadError` limpian la cola.

## Verificación

- `go build ./...` (raíz): compila sin errores.
- `bun run build` (frontend): type-check + build de producción correctos.
- Grep verificado: sin referencias a `UseSystemJava` ni a `startLoaderInstall` en el modal de instancias.