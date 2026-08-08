# StepLauncher-Error-10: LWJGL no encuentra lwjgl.dll en el arranque de versiones con natives en subcarpetas (26.2)

- **Fecha**: 2026-08-05
- **Estado**: corregido (pendiente verificación de ejecución real)
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

Al lanzar **Minecraft 26.2** (template moderno con natives en subcarpetas) el
juego moría a los pocos segundos sin abrir ventana, con crash-report repetido:

```
java.lang.UnsatisfiedLinkError: Failed to locate library: lwjgl.dll
    at org.lwjgl.system.Library.loadSystem(Library.java:177)
    at com.mojang.blaze3d.platform.NativeLibrariesBootstrap.loadLWJGLSystem(...)
Contents of java.library.path :
    jtracy-jni-windows.dll: application/x-msdownload
```

La primera tanda de crash-reports (20:27–21:55) mostraba
`Contents of java.library.path : <not a directory>` y la segunda (22:34)
mostraba la carpeta como existencia pero sin `lwjgl.dll`. Las versiones legacy
(1.12.2, 1.16.5) lanzaban limpio.

## 2. Causa raíz

El template de 26.2 referencia los natives **con subcarpetas** dentro de
`natives_directory`:

- `-Djava.library.path=${natives_directory}/java` → dir donde el JVM busca
  las DLL (incl. `lwjgl.dll`).
- `-Djna.tmpdir=${natives_directory}/jna`, etc.

`helpers.ExtractNatives` extraía **todas las DLL a la raíz de `natives/`**
(en una primera versión) o **agrupadas por librería en subcarpetas** `java/`,
`jna/`, `lwjgl/`, `netty/` (en una segunda versión). Ninguna de las dos
coincidía con el contrato del template:

- En la raíz → `natives/java` no existía → `Contents of java.library.path :
  <not a directory>`.
- Agrupando por librería → `lwjgl.dll` caía en `natives/lwjgl/`, pero el JVM la
  busca en `natives/java/` → `Contents of java.library.path :` vacío y
  `Failed to locate library: lwjgl.dll`.

El core del error: **el launcher imponía su propio layout de natives en vez de
leer el que indica el propio template de la versión**.

## 3. Solución aplicada (propuesta del usuario: leer la ruta del template)

En `internal/Core/Launcher/Helpers/Natives.go`:

- `ExtractNatives(libraries, librariesDir, nativesDir, jvmArgs)` ahora **lee
  las `jvmArgs` del template** y detecta la subcarpeta que el propio template
  referencia con `${natives_directory}/<sub>`.
- Prioriza la subcarpeta de `-Djava.library.path` (el dir real de carga de
  DLLs: en 26.2 es `java`). Si el template usa `${natives_directory}` plano
  (formato legacy de 1.12.2/1.16.5), extrae a la raíz sin cambiar nada.
- `nativeSubDir`-classificación por librería se **eliminó**: ya no se reparte
  `lwjgl.dll` a `natives/lwjgl`; todas las DLL se extraen a `natives/java`
  (el dir exacto que el JVM busca).
- No hay **ningún mapeo de versiones hardcodeado** (con >1000 versiones sería
  insostenible): la ruta sale del template en cada lanzamiento.
- `Launcher.go` pasa `l.ver.Arguments.JVM` a `helpers.ExtractNatives` (con
  salvaguarda de `Arguments == nil`).

También se refuerza la selección de binarios Java: **siempre `javaw`**
(Windows) en `Helpers/Java.go` (`resolveOfficialJava` y `resolveCustomJava`),
nunca `java`/`java.exe`, incluso en la búsqueda de Java oficial y del sistema.

## 4. Verificación

- `go build ./...` → OK (antes/después del cambio).
- Análisis mental del layout resultante para 26.2: `natives/java` + todas las
  DLL extraídas (incl. `lwjgl.dll`).
- La validación real (lanzar Minecraft 26.2) queda pendiente de ejecución en
  equipo con el binario regenerado.

## 5. Regla aprendida

Los natives de versions modernas de Mojang ya NO van todos a la raíz: el
**propio template** define con `${natives_directory}/<sub>` dónde deben quedar
(ej. `java`, `jna`, `lwjgl`, `netty`). La política correcta es **leer la ruta
que el template indica y concatenarla con el `nativesDir` base**, nunca decidir
el layout por heurística de librería ni mapeo de versiones.

## 6. Refinamiento posterior (feedback del usuario)

La primera corrección se rehizo al revisar el código de extracción (el usuario
reportó que era "un desastre" y que "borraba y re-extraía todo en cada inicio"):

- **Se eliminó `os.RemoveAll(nativesDir)` del camino de lanzamiento.** Antes
  cada inicio borraba `natives/` entero y lo volvía a extraer. Ahora la
  extracción es **incremental** (`utils.ExtractJarNatives` no reescribe
  destino existente) y el directorio solo se limpia en la descarga de la
  versión (`utils.ExtractNatives` del Downloader, que sí limpia porque ahí se
  instalan/actualizan natives de cero).
- **El resolutor de jars nativos podía coger la arquitectura equivocada.** En
  26.2 los natives son entradas separadas con el clasificador en el nombre
  Maven (`org.lwjgl:lwjgl:3.4.1:natives-windows`), y en disco coexisten los
  jars `-arm64` y `-x86`. El barrido con `strings.HasPrefix(key,
  "natives-windows")` los incluía todos (DLL con el mismo nombre, ganaba la
  lexitura no determinista del mapa) → podrías acabar con `lwjgl.dll` de la
  arquitectura incorrecta. Ahora `nativeClassifier(osName)` calcula el sufijo
  exacto del SO+arquitectura actuales y solo se acepta ese.
- **`utils.ExtractNatives` extrae plano** (todo a la raíz) eliminando la
  clasificación `nativeSubDir` por librería (`lwjgl/`, `jna/`, `netty/`), que
  era lo que producía el layout roto en primer lugar. El Downloader ya no
  ensucia con subcarpetas por librería.
- El launcher converge el layout en cada lanzamiento sin borrar: crea el dir de
  carga (`natives/<sub>`) y las subcarpetas de trabajo del template, y elimina
  solo los binarios nativos *residuos* de layouts previos (no la carpeta).
- Se eliminó la obligación de flags `-Dsun.java2d.*` forzados
  (`helpers.HWAccelDisableFlags`): cuando `hardwareAcceleration`/`gcPreset`/
  `gpuPreference` están deshabilitados, el comando final no debe llevar NINGÚN
  flag derivado de ellos (regla del usuario: configuración deshabilitada → no
  inyectar, ni variables ni argumentos).