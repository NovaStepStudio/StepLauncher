# StepLauncher-Error-11: El descargador instalaba natives de 32 bits (y de todas las plataformas) en sistemas de 64 bits

- **Fecha**: 2026-08-06
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

Al instalar versiones modernas (26.2 y similares) en Windows, el launcher
descargaba **todas** las variantes de natives y podía acabar extrayendo la DLL
de **32 bits** (o de ARM64) en un sistema de **64 bits**. Como la extracción es
incremental y "el primero que escribe gana", un `lwjgl.dll` de la arquitectura
equivocada dejaba inservible el arranque del juego.

## 2. Causa raíz

En el JSON de versiones modernas cada variante de natives es una **entrada
separada** con el clasificador en el nombre Maven:

```
org.lwjgl:lwjgl:3.4.1:natives-windows        (x64)
org.lwjgl:lwjgl:3.4.1:natives-windows-arm64
org.lwjgl:lwjgl:3.4.1:natives-windows-x86    (32 bits)
org.lwjgl:lwjgl:3.4.1:natives-linux / -osx / ...
```

Y las **reglas `rules` NO llevan `arch`**: todas las variantes de Windows
tienen `allow os windows`. El código del descargador (`internal/Core/Downloader/Tasks.go`)
filtra y resuelve **por prefijo**:

- `MatchRules` solo miraba `os.name` (sin `arch`) → dejaba pasar x64, arm64 y
  x86 a la vez.
- `addLibraryTasks` añadía **cualquier** entrada con `Downloads.Artifact`, así
  que los clasificadores con natives de **todas** las plataformas/x86/arm64
  entraban como "librerías" normales.
- `ResolveNativeArtifact` barría con `strings.HasPrefix(key, "natives-windows")`
  → aceptaba `natives-windows-arm64` y `natives-windows-x86` además del correcto.
- `addNativeTasks` hacía lo mismo → el jar equivocado se añadía a
  `dl.nativeJars` y se extraía.

Resultado: 72 entradas de clasificador (x86, arm64, linux, macOS...) se
descargaban en Windows/x64 y la extracción no determinista podía dejar DLLs de
la arquitectura incorrecta. Diagnóstico reforzado con el JSON real de **26.2**
(descargado de piston-meta): 131 librerías, 72 de ellas clasificadores de
natives, y solo 12 corresponden a Windows/x64.

## 3. Solución aplicada

### Fuente única de clasificador de plataforma

- `internal/Core/Platform/Platform.go`: nuevas `NativeClassifierFor(os, arch)`
  y `NativeClassifier()`, expuestas en `internal/Core/Utils/Platform.go`.

### Descargador (`internal/Core/Downloader/Tasks.go`)

- **`ResolveNativeArtifact` reescrito** con coincidencia **exacta**:
  - Formato moderno: solo se acepta la librería cuyo nombre termina en
    `:<clasificador>` EXACTO del SO+arq actuales (`natives-windows`, etc.).
    Los de otras arquitecturas devuelven `nil`.
  - Formato antiguo: se resuelve la plantilla `${arch}` y la clave exacta en
    `downloads.classifiers`.
- **`addLibraryTasks`**: se salta cualquier entrada `IsNativeClassifierEntry`
  (todas las variantes de natives) → ya no se descargan como librerías.
- Eliminada `archMatches` (muerta tras el cambio).

### Classpath (`internal/Core/Launcher/Helpers/Classpath.go`)

- `BuildClasspath` excluye las entradas `IsNativeClassifierEntry`: así el
  classpath no referencia las variantes de otras plataformas que ya no se
  descargan, y no aparecen como "missing".

### `helpers.Natives.go`

- Eliminada la copia local `nativeClassifier(osName)` y se usa
  `utils.NativeClassifier()` (una sola fuente de verdad para resolver el jar).

## 4. Verificación

- `go build ./...` → OK.
- `bun run build` (frontend) → OK.
- Simulación contra el JSON real de **26.2**: se seleccionan exactamente las
  12 entradas `natives-windows` (x64) y se descartan las 60 restantes
  (x86/arm64/los demás SO) en la sección de librerías.

## 5. Regla aprendida

Los clasificadores de natives del formato moderno NO llevan la arquitectura en
las reglas: el launcher debe elegirlos por el **sufijo exacto** del SO+arq
actuales (windows→`natives-windows`, windows-x86→`natives-windows-x86`, etc.)
y **nunca** por prefijo ni por `os.name` solo. Cualquier resolución de natives
(descarga, classpath, extracción) debe usar la misma función compartida
(`utils.NativeClassifier`), nunca lógica local duplicada que pueda divergir.