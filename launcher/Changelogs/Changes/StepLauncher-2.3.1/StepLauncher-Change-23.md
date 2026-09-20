# Changes/StepLauncher-2.3.1/StepLauncher-Change-23.md

- **Fecha**: 2026-08-12
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado (build OK).

## Qué cambió

El merge de librerías parent+child (`mergeVersions`) hacía `append(parent.Libraries, child.Libraries...)`: las librerías del modloader (child) solo ganaban sobre las de la versión base (parent) cuando coincidía la ruta exacta del jar. Si ambas declaraban el mismo artefacto (grupo:artefacto) con **versiones distintas** (p. ej. `guava:33.x` de la vanilla vs `guava:32.1.2-jre` del modloader), las dos entradas llegaban al classpath. El bootstraplauncher de Forge/NeoForge convierte cada jar del classpath en un módulo, y dos versiones del mismo artefacto exportan los mismos paquetes → `java.lang.module.ResolutionException` (split package).

## Detalle de los cambios

### `internal/Core/Launcher/Launcher.go` (nueva función `mergeLibraries`)

- `mergeVersions` ahora usa `mergeLibraries(parent.Libraries, child.Libraries)` en lugar del `append` directo.
- La clave de deduplicación es `group:artifact` (+ classifier cuando existe, para conservar las variantes natives de cada lado).
- **Prioridad al modloader**: si el child declara el mismo artefacto que el parent, la entrada del parent se reemplaza por la del child (se preserva la posición en el orden del classpath).
- Caso actual verificado (NeoForge 21.1.248 + vanilla 1.21.1): ambos json declaran las mismas versiones (guava 32.1.2-jre, gson 2.10.1, log4j 2.22.1), por lo que el classpath resultante no cambia (83 entradas únicas); la mejora cubre futuros casos donde las versiones difieran.

## Pruebas

- `go build ./...` en la raíz: OK.
- Verificado del log de lanzamiento real (sesión 21:57): el `-cp` del proceso tiene 83 entradas sin strings duplicados ni versiones distintas del mismo artefacto (la dedup por ruta de `BuildClasspath` ya operaba; esta función evita el duplicado en la fuente del merge).