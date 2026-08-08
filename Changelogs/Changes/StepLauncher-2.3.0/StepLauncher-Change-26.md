# StepLauncher-Change-26: Extracción incremental de natives + resolución por arquitectura + sin flags sun.java2d forzados

- **Fecha**: 2026-08-05
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

### 1. Los natives ya no se borran y re-extraen en cada lanzamiento

`internal/Core/Launcher/Helpers/Natives.go` (`ExtractNatives`):

- **Eliminado `os.RemoveAll(nativesDir)`** del camino de lanzamiento. Antes,
  cada inicio borraba `natives/` completo y volvía a extraer los jars
  (reestablecimiento total por partida).
- La extracción es ahora **incremental**: `utils.ExtractJarNatives` no
  reescribe un destino que ya existe. Llamarla en cada lanzamiento no rehace
  nada; solo escribe lo que falta.
- El launcher **converge** el layout sin borrar: crea el dir de carga que el
  template indica (`natives/<sub>`, p. ej. `java`) y las subcarpetas de trabajo
  del template (`jna`, `lwjgl`, `netty`), y elimina únicamente los **binarios
  nativos residuos** de layouts previos (los `*.dll`/`*.so` sueltos fuera del
  dir de carga), nunca la carpeta ni el resto de archivos.
- La limpieza total de natives queda solo en la descarga/actualización de la
  versión (`utils.ExtractNatives`, usado por `Downloader/Manager.go`), que es
  donde toca instalar natives de cero.

### 2. `utils.ExtractNatives` extrae plano (sin clasificar por librería)

`internal/Core/Utils/Extract.go`:

- Eliminada la clasificación `nativeSubDir` (que repartía por `lwjgl/`, `jna/`,
  `netty/`, `java/` según el path del jar). Ahora **todas** las DLL van a la
  raíz de `natives/`.
- `ExtractJarNatives` es idempotente (skip si el destino existe).
- `isNativeFile` se exporta como `IsNativeFile` para reutilizarla en la
  limpieza de residuos del launcher.

### 3. Resolución de jars nativos por SO + arquitectura exacta

`resolveNativeJar`/nueva `nativeClassifier(osName)`:

- Formato actual (26.2+): cada librería nativa es una entrada separada con el
  clasificador en el **nombre Maven** (`org.lwjgl:lwjgl:3.4.1:natives-windows`).
  Se acepta solo el **sufijo exacto** del SO+arquitectura actuales
  (windows x64 → `natives-windows`; arm64 → `natives-windows-arm64`; x86 →
  `natives-windows-x86`; linux/osx análogos). El barrido anterior con
  `strings.HasPrefix(..., "natives-windows")` incluía también los jars
  `-arm64` y `-x86` (mismas DLL, ganaba un orden no determinista del mapa).
- Formato antiguo (mapa `natives` + `downloads.classifiers`): se mantiene
  `resolveClassifier` con el reemplazo de `${arch}` y fallback a la clave
  exacta del clasificador actual.

### 4. Sin flags `-Dsun.java2d.*` forzados en el comando final

- Eliminada la inyección de `helpers.HWAccelDisableFlags()` en
  `Launcher.go/buildJVMArgs`: cuando `hardwareAcceleration` está deshabilitado
  (o sin configurar), **no se añade ningún argumento** al comando del juego.
- Eliminada la función `helpers.HWAccelDisableFlags` (muerta).
- Misma política que pedía el usuario para los tres campos relacionados:
  `gcPreset` vacío → sin flags de GC (ya era así), `gpuPreference` vacío → sin
  env vars de GPU (ya era así), `hardwareAcceleration` deshabilitado → sin
  flags de Java2D (este cambio).

## Por qué

- Re-extract total en cada inicio era desperdicio (I/O y borrado de carpetas
  de trabajo que otros procesos usan en runtime) y el usuario lo marcó como
  inaceptable.
- El resolutor con `HasPrefix` podía instalar DLL de la arquitectura
  equivocada al coexistir `-arm64`/`-x86` en disco.
- Los flags `sun.java2d.*` aparecían en el comando con solo no configurar la
  aceleración (nil se trataba como "deshabilitada" → inyección), y la regla es
  no meter NADA en el comando cuando la configuración está deshabilitada.

## API afectada

- `utils.isNativeFile` → `utils.IsNativeFile` (exportado).
- Eliminadas: `utils.nativeSubDir`, `helpers.HWAccelDisableFlags`.
- Sin cambios de firma en `helpers.ExtractNatives` ni `utils.ExtractNatives`.

## Cómo verificar

- `go build ./...` → OK.
- Lanzar 26.2 dos veces seguidas: la segunda no debe re-extraer nada (nada de
  "Extracted N native files" repetido; `natives/java` completo, `jna`/
  `lwjgl`/`netty` sin DLL).
- Comprobar que el comando final ya no contiene `-Dsun.java2d.*`.
- Lanzar 1.16.5/1.12.2: siguen saliendo limpias (layout legacy plano intacto).
