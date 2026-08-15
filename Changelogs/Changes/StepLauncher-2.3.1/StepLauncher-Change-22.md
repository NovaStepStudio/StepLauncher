# Changes/StepLauncher-2.3.1/StepLauncher-Change-22.md

- **Fecha**: 2026-08-12
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado (build OK; fix reproducido en runtime con `java.exe`).

## Qué cambió

**NeoForge 21.1.x no arrancaba con `java.lang.module.ResolutionException: Modules minecraft and _1._21._1 export package com.mojang.blaze3d.systems to module mixin_synthetic`** (exit code 1 ~14 s después de lanzar). La causa: el client jar de la versión base (`versions/1.21.1/1.21.1.jar`) se incluye en el `-cp` con su nombre de archivo original, y el bootstraplauncher de Forge/NeoForge 1.17+ convierte **cada jar del classpath en un módulo** cuyo nombre se deriva del nombre del archivo. Como `1.21.1.jar` no trae `module-info.class`, el módulo resultante se llama `_1._21._1`, y choca con el módulo `minecraft` que modlauncher crea para el jar del juego: dos módulos exportando el mismo paquete → `ResolutionException`.

El ignoreList del template (`-DignoreList=client-extra,${version_name}.jar`) excluye al client jar del procesado de módulos **solo si el archivo se llama como el id de la versión lanzada** (`neoforge-21.1.248.jar`), igual que en el launcher oficial de Mojang, que guarda/copia el client jar como `versions/<version_id>/<version_id>.jar`.

## Detalle de los cambios

### `internal/Core/Launcher/Launcher.go` (`launchGame` y nueva función `ensureClientJarCopy`)

- Cuando el plan de ejecución declara `mainClass` que contiene `bootstraplauncher` (Forge 1.17+ / NeoForge), se copia el client jar de la versión base a `versions/<id_lanzada>/<id_lanzada>.jar` (si no existe o difiere en tamaño) y ese es el jar que entra al classpath, en lugar del de la versión base.
- Si la copia falla (p. ej. aún no está descargado el jar base), se hace fallback al comportamiento anterior (jar de la versión base) para no romper el flujo.
- Vanilla, Fabric y Forge antiguo no se ven afectados (no usan bootstraplauncher; el nombre del archivo del client jar es irrelevante).

## Verificación en runtime (reproducción con java.exe)

Comando exacto del log de lanzamiento (`StepLauncher-neoforge-21.1.248-2026-08-12.log`) reconstruido con `java.exe`:

- Con `-cp` conteniendo `versions/1.21.1/1.21.1.jar` → `ResolutionException: Modules minecraft and _1._21._1 export package com.mojang.blaze3d.systems to module mixin_synthetic` (reproducción del bug).
- Con `-cp` conteniendo `minecraft.jar` → `Module minecraft reads another module named minecraft` (confirma que modlauncher nombra el módulo del juego "minecraft" por defecto).
- Con `-cp` conteniendo `versions/neoforge-21.1.248/neoforge-21.1.248.jar` (copia del client jar, nombre = id de versión, cubierto por el ignoreList) → el juego pasa la resolución de módulos, arranca NeoForge y continúa el arranque ("Launching target 'forgeclient'", MixinExtras inicializado). **Fix confirmado.**

## Pruebas

- `go build ./...` en la raíz: OK.
- `wails dev`: lanzar neoforge-21.1.248 → el log del lanzamiento debe mostrar en el classpath `versions/neoforge-21.1.248/neoforge-21.1.248.jar` y el juego debe entrar a la pantalla de carga (antes moría con exit code 1 a los ~14 s). Lanzar vanilla 1.21.1 y una versión Fabric → sin cambios de comportamiento.

## Ampliación: lanzamiento sin instancia (misma causa raíz)

El fix original solo se activaba con `adv.ExecutionPlan != nil` (flujo de instancias). Al lanzar `neoforge-21.1.248` desde la lista de versiones (sin instancia) no existe ExecutionPlan, el json del modloader ya declara el main class del bootstraplauncher y la copia del client jar nunca se ejecutaba: el classpath volvía a incluir `versions/1.21.1/1.21.1.jar` → crash `java.lang.module.ResolutionException: Module minecraft contains package com.mojang.blaze3d.platform, module _1._21._1 exports package com.mojang.blaze3d.platform to minecraft` (mismo split package; en el primer reporte era `com.mojang.blaze3d.systems`, ahora `com.mojang.blaze3d.platform`).

- `internal/Core/Launcher/Launcher.go` (`Launch`): la decisión de copiar el client jar ahora usa el **main class efectivo** (`adv.ExecutionPlan.MainClass` si existe, si no `l.ver.MainClass` del json mergeado): si contiene `bootstraplauncher` se copia el jar de la versión base a `versions/<id_lanzada>/<id_lanzada>.jar` y ese es el jar que entra al classpath. El `mainClass` final del lanzamiento reutiliza ese mismo cálculo (ya no hay doble resolución).
- `internal/Core/Launcher/Launcher.go` (`downloadMissingLibraries`): ahora recibe `clientJarVer` como parámetro; si el jar destino (`versions/<id_lanzada>/<id_lanzada>.jar`) falta y la copia temprana falló porque el jar base aún no estaba descargado, el client jar se descarga **directamente al destino con el nombre correcto** (cubierto por el ignoreList) en lugar de al jar de la versión base.

## Verificación (ampliación)

- `go build ./...`: OK.
- `wails dev`: lanzar `neoforge-21.1.248` desde la lista de versiones (sin instancia) → el log del lanzamiento debe mostrar `versions/neoforge-21.1.248/neoforge-21.1.248.jar` en el classpath (antes salía `versions/1.21.1/1.21.1.jar`) y el juego debe pasar la resolución de módulos y llegar a la carga.