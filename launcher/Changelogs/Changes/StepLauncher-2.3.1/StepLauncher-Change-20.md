# Changes/StepLauncher-2.3.1/StepLauncher-Change-20.md

- **Fecha**: 2026-08-12
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado (build, vet y type-check OK; lanzamiento de NeoForge reproducido y resuelto en runtime).

## Qué cambió

Cuatro correcciones: (1) el **classpath/module path** del lanzamiento se arma correctamente, arreglando que **vanilla 1.21.1 y NeoForge 21.1 no arrancaran** (salían con exit code 1 en menos de 2 segundos y sin mensaje visible, porque `javaw.exe` se traga el stderr); (2) el classpath ahora se **deduplica** (el merge parent+child duplicaba jars y el bootstraplauncher moría con "Duplicate key"); (3) la **memoria RAM mínima es siempre 512 MB fijos** (nunca la mitad de la máxima); (4) los **logs reportan la versión 2.3.1** (decían 2.3.0).

## Detalle de los cambios

### 1. Lanzamiento: classpath/module path corregido (vanilla + NeoForge no arrancaban)

Análisis con `versions/neoforge-21.1.248/neoforge-21.1.248.json` + `logs/game/StepLauncher-neoforge-21.1.248-2026-08-12.log`:

- El **JSON de NeoForge** declara `mainClass: cpw.mods.bootstraplauncher.BootstrapLauncher` y en `arguments.jvm` el flag corto `-p` seguido de la lista de jars del bootstrap (bootstraplauncher, securejarhandler, asm, JarJarFileSystems) + `--add-modules ALL-MODULE-PATH`. Sin `-p` (y sin `-cp`), Java toma la lista de jars como nombre de clase principal → "Could not find or load main class" → exit 1.
- El **JSON de vanilla 1.21.1** declara `-cp` y `${classpath}` en `arguments.jvm`: el classpath se retiraba del template pero **nunca se re-emitía**.
- `internal/Core/Launcher/Launcher.go` (`buildJVMArgs`): el antiguo filtro marcaba `tplHasModulePath` cuando veía `-cp` (¡y `-cp` es classpath, no module path!), por lo que el `switch` final tomaba la rama vacía y **no se agregaba ningún classpath**: vanilla moría sin classpath y NeoForge quedaba a merced de su `-p` (que además no se reconocía como flag de module path).
- **Fix**: el filtro ahora reconoce `--module-path` y `-p` como flags de module path (con su valor), retira `-cp`/`--class-path`/`${classpath}` del template y decide al final:
  - template con module path propio (`-p <jarList>` de NeoForge) → se re-emite `--module-path <jarList>` **y además `-cp <classpath>`**.
  - template con `-p ${classpath}` (Forge 1.17+) → se re-emite `--module-path <classpath>` **y además `-cp <classpath>`**.
  - `ExecutionPlan.UseModulePath` (instancias) sin module path en template → `--module-path <classpath>` **y además `-cp <classpath>`**.
  - resto (vanilla, Fabric) → `-cp <classpath>` al final.
- **Por qué también `-cp` en los casos de module path**: el `BootstrapLauncher` de Forge/NeoForge 1.17+ construye su capa `MC-BOOTSTRAP` leyendo el classpath legacy (`java.class.path`/`legacyClassPath`); sin él muere al instante con `java.util.NoSuchElementException: No value present` (`BootstrapLauncher.run(BootstrapLauncher.java:210)`). Verificado reproduciendo el comando con `java.exe`: sin `-cp` → "No value present"; con `-cp` → el juego arranca y carga los 50+ mods.
- **Deduplicación del classpath** (`internal/Core/Launcher/Helpers/Classpath.go`, `BuildClasspath`): al mergear librerías de parent+child, jars como gson/guava/log4j aparecían 2 veces (94 entradas) y el bootstraplauncher moría con `java.lang.IllegalStateException: Duplicate key ... gson-2.10.1.jar`. Ahora se deduplican las rutas (normalizadas y case-insensitive en Windows); quedan 83 únicas.

### 2. RAM mínima siempre 512 MB (nunca la mitad de la máxima)

- `internal/Core/Launcher/Manager.go`: la normalización pasa de `if adv.MinRAM <= 0` a `if adv.MinRAM < 512`, forzando el piso de 512 MB.
- `internal/Core/Launcher/Launcher.go` (`buildJVMArgs`): **se eliminó el fallback `minMem = maxMem / 2`** (generaba `-Xms2048M` con RAM máxima de 4096 MB, lo que el usuario rechazó: "quiero 512 nomas"). Ahora `minMem <= 0 → 512` y `minMem < 512 → 512`; además la garantía `minMem ≤ maxMem` (si el usuario pidió una mínima mayor que la máxima, se recorta a la máxima para que la JVM no rechace `-Xms > -Xmx`).

### Ampliación: el flujo Engine seguía poniendo la mitad de la máxima

El fix de `buildJVMArgs` no cubría a quien calculaba el valor: `internal/Handlers/Engine/Launch.go` (`buildBaseLaunchConfig`) seguía haciendo `minRAM := maxRAM / 2`, así que con RAM máxima de 4096 MB la mínima quedaba en 2048 MB (`RAM : 4096 MB (2048 min / 4096 max)` en el encabezado del log de juego). Ahora:

- `internal/Handlers/Engine/Launch.go`: `adv.MinRAM = 512` fijo (se elimina `maxRAM / 2`).
- `internal/Core/Launcher/Manager.go`: la normalización pasa de piso a **valor fijo** `adv.MinRAM = helpers.MinRAM` (512), y `RecommendedRAM()` devuelve siempre 512 como mínima (solo recorta si la máxima fuera menor).
- `internal/Core/Launcher/Launcher.go` (`buildJVMArgs`): `minMem := helpers.MinRAM` (512) sin derivar nada de `adv.MinRAM`; se conserva la garantía `minMem ≤ maxMem`.

### 3. Versión del launcher 2.3.1 en los logs

- `internal/Handlers/Engine/engineconfig/Config.go`: `AppVersion = "2.3.1"` (era "2.3.0"; los logs de lanzamiento y del sistema decían v2.3.0 mientras el launcher ya era 2.3.1 según `wails.json`).
- `internal/Core/Auth/Authlib.go`: User-Agent de descarga de skins → `StepLauncher/2.3.1`.
- `frontend/package.json`: `version: "2.3.1"`.

## Comportamiento anterior/nuevo

- **Anterior**: vanilla 1.21.1 lanzaba `net.minecraft.client.main.Main` **sin classpath** → crash inmediato (exit 1, sin output); NeoForge 21.1.248 lanzaba el `BootstrapLauncher` **sin `-p` ni `-cp`** → "No value present"; con `-cp` duplicado → "Duplicate key"; la RAM mínima podía quedar en la mitad de la máxima (o 0/256 MB); los logs decían "StepLauncher v2.3.0".
- **Nuevo**: vanilla lanza con `-cp <classpath>`; NeoForge/Forge moderno lanza con `--module-path` (su lista propia o el classpath) **+ `-cp <classpath>` deduplicado**; la RAM mínima es siempre 512 MB; los logs dicen v2.3.1.

## Cómo verificar

- `go build ./...` OK; `go vet ./...` OK; `bun run type-check` OK (frontend).
- **Verificación en runtime (reproducción con java.exe)**: el comando de neoforge-21.1.248 sin `-cp` muere con `NoSuchElementException: No value present`; con `--module-path <jarList> -cp <classpath deduplicado>` el juego arranca, lista los 50+ mods (NeoForge 21.1.248, Sodium, JEI, Xaero, etc.) y continúa el arranque.
- `wails dev`: lanzar vanilla 1.21.1 → debe entrar al juego (en el log del lanzamiento debe verse `-cp` con las librerías al final de los JVM args). Lanzar neoforge-21.1.248 → el comando debe incluir `--module-path` con la lista de jars del bootstrap **y** `-cp` con el classpath sin duplicados, y el juego debe arrancar. En Ajustes, poner RAM mínima por debajo de 512 MB → el log debe mostrar `Min Memory : 512 MB` (y NUNCA `-Xms` con la mitad de la máxima). El encabezado de los logs debe decir `StepLauncher v2.3.1`.