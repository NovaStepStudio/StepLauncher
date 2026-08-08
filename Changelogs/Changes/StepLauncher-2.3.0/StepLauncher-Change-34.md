# StepLauncher-Change-34.md

## Fecha
2026-08-06

## Release
StepLauncher-2.3.0 — se mencionó por primera vez en esta release.

## Cambio
Soporte correcto de lanzamiento para Forge/NeoForge modernos (1.13+) y NeoForge 21.5+ (FancyModLoader/JPMS): gestión del token `${launcher_properties}` y del classpath por `--module-path` en lugar del `-cp` clásico.

## Qué pasaba
- El `version.json` de Forge/NeoForge modernos referencia un argumento JVM `${launcher_properties}` que el launcher no resolvía: quedaba literal y ModLauncher lo recibía como argumento desconocido.
- NeoForge 21.5+ (MC 1.21.4+) arranca con módulos JPMS (FancyModLoader): el classpath se entrega con `--module-path`, no con `-cp`. El launcher forzaba siempre `-cp <classpath>` al final, por lo que estas versiones no podían arrancar. `ExecutionPlan.UseModulePath` existía pero no se usaba.

## Qué se hizo
1. **`internal/Core/Launcher/Helpers/Args.go`**:
   - `VarConfig.LauncherProperties` nuevo y `BuildVarsMap` registra la variable `launcher_properties` (path del fichero) para la sustitución `${launcher_properties}`.
2. **`internal/Core/Launcher/Launcher.go`**:
   - Nuevo `ensureLauncherProperties(gameDir)`: genera `launcher.properties` en el game dir con `fml.client.secret=<hex aleatorio>` (mismo formato que el launcher oficial de Mojang para FML 1.13+) y devuelve su ruta. Solo se genera cuando hay `ExecutionPlan` (modloader instalado); si falla se devuelve vacío y el arg se omite.
   - `buildJVMArgs`: detecta si el template ya declara `--module-path`/`-p` (no añade el `-cp` final) y, si `ExecutionPlan.UseModulePath` está marcado pero el template no lo trae, añade `--module-path <classpath>`. El `-cp` clásico se mantiene para el resto (vanilla, Forge ≤1.12, Fabric, Forge 1.13–1.21).
3. **`internal/Core/ModLoader/Provider/Forge.go`**:
   - `BuildExecution` marca `UseModulePath = true` cuando el profile del loader declara `--module-path` o `-p` en sus `arguments.jvm` (cubre Forge/NeoForge).

## API afectada
- Sin cambios de bindings Wails ni de tipos expuestos al frontend (los campos ya existían o son internos).
- Comportamiento del lanzamiento: las versiones cuyo profile declare módulos ya no reciben `-cp`.

## Cómo verificar
- `go build ./...` OK.
- Lanzar una versión Forge 1.16.5: los logs deben mostrar `-Dlegacy.classpath=<cp>` y el arg `launcher.properties` con ruta válida, sin `${launcher_properties}` literal.
- Lanzar NeoForge 21.5+/1.21.4+: el comando Java debe usar `--module-path` y no contener `-cp`.
- `go fmt ./...` sin cambios pendientes.

## Notas
- El fichero `launcher.properties` se regenera en cada lanzamiento con un secret nuevo (idempotente, sin acumulación).
- El `-Dlegacy.classpath=${classpath}` de Forge 1.13–1.21 ya se resolvía correctamente; este cambio no lo altera.
