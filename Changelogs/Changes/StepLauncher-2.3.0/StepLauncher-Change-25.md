# StepLauncher-Change-25: Extracción de natives guiada por el template + preferencia forzada de javaw

- **Fecha**: 2026-08-05
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

### 1. Native extraction: layout dictado por el template, no mapeo de versiones

`internal/Core/Launcher/Helpers/Natives.go` se reescribe:

- `UsesNativesSubDirs` (firma booleana) desaparece; ahora
  `ExtractNatives(libraries, librariesDir, nativesDir, jvmArgs)` recibe las
  **jvmArgs del template** y decide la subcarpeta de carga leyendo la
  referencia `${natives_directory}/<sub>` del propio JSON de la versión
  (prioridad a `-Djava.library.path`).
- **Nuevos formatos modernos** (p. ej. 26.2: `-Djava.library.path=${natives_directory}/java`)
  → todas las DLL de natives se extraen a `natives/java` (limpia antes el dir
  para quitar restos de layouts previos). Así `lwjgl.dll` queda donde el JVM la
  busca, eliminando el crash `UnsatisfiedLinkError: Failed to locate library:
  lwjgl.dll`.
- **Formatos legacy** (1.12.2, 1.16.5: `${natives_directory}` plano) → se
  extrae a la raíz de `natives/`, sin cambios de comportamiento.
- `Launcher.go` pasa `l.ver.Arguments.JVM` (con guard de `Arguments == nil`).

### 2. Binario Java: SIEMPRE `javaw`, nunca `java`

`internal/Core/Launcher/Helpers/Java.go`:

- `resolveOfficialJava` → en Windows devuelve `bin\javaw.exe` (con fallback
  defensivo a `java.exe` solo si el runtime no trae `javaw`, que en los
  runtimes oficiales de Mojang siempre viene).
- `resolveCustomJava` → si el usuario no fija ruta, busca primero `javaw` por
  PATH y recién después `java`. Si pasa una ruta custom con `java`/`java.exe`
  en una carpeta `bin`, se sustituye por el `javaw.exe` hermano.

## Por qué

El crash de 26.2 y la aparición de windows consolas con `java.exe` eran dos
deudas de implementación: el launcher imponía su propio layout de natives
ignorando el contrato del template, y elegía `java.exe` en lugar de `javaw.exe`
en el path oficial.

## API afectada

- `helpers.ExtractNatives` cambia de firma: `(libraries, librariesDir,
  nativesDir, jvmArgs []interface{})` (antes booleano `split`).
- Eliminada `helpers.UsesNativesSubDirs` (reemplazada por la lectura directa
  del template).

## Cómo verificar

- `go build ./...` → OK.
- Lanzar 26.2: nacimiento de `natives/java/` con `lwjgl.dll` y resto de DLL, y
  el juego abre (pendiente de ejecución real).
- Lanzar 1.16.5/1.12.2: siguen saliendo limpias (layouts legacy intactos).