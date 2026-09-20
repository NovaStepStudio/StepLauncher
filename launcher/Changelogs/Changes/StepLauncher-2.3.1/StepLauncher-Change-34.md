# Changes/StepLauncher-2.3.1/StepLauncher-Change-34.md

- **Fecha**: 2026-08-15
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado.

## Qué cambió

Fix de lanzamiento para versiones legacy (Forge 1.7.10): el `Main.class` de Minecraft 1.7.10 exige la opción `--userProperties` como obligatoria (vía joptsimple) y abortaba con `Missing required option(s) ['userProperties']` cuando no llegaba en la línea de lanzamiento.

En `internal/Core/Launcher/Launcher.go` (`buildGameArgs`) se cambió la prioridad de lectura de los argumentos de juego del version.json:

- **Antes**: se usaba `arguments.game` siempre que el campo `arguments` existiera (`l.ver.Arguments != nil`), ignorando `minecraftArguments` incluso cuando `arguments.game` estaba vacío.
- **Ahora**: si `minecraftArguments` existe (no vacío), se lee y sustituye directamente (`--username ${auth_player_name} --version ... --userProperties ${user_properties} --userType ${user_type} ...`); solo si no existe se cae a `arguments.game`.

El `minecraftArguments` de la versión vanilla 1.7.10 (y del perfil de Forge 1.7.10) ya incluye `--userProperties ${user_properties}` (sustituido a `{}`), por lo que el juego recibe el parámetro obligatorio sin necesidad de hardcodearlo. No se añade ningún argumento manual: se respeta fielmente el template del version.json.

## Por qué

El juego (Forge 1.7.10) esperaba `--userProperties {}` junto a `--username`, `--uuid`, `--accessToken` y `--userType`, y sin él abortaba con `joptsimple.MissingRequiredOptionException`. La solución correcta es leer el `minecraftArguments` del version.json (que ya lo declara), no inyectar el argumento a mano.

## API afectada

- Sin cambios en bindings de Wails ni en el frontend (solo backend, dentro del lanzador).

## Comportamiento anterior/nuevo

- **Antes**: si `arguments` existía con `game` vacío, se ignoraba el `minecraftArguments` y el juego podía arrancar sin `--userProperties` (y sin el resto del template), muriendo con `Missing required option(s) ['userProperties']`.
- **Ahora**: `minecraftArguments` tiene prioridad cuando existe; Forge 1.7.10 recibe `--userProperties {}` y arranca correctamente.

## Cómo verificar

- Lanzar la versión `1.7.10-Forge10.13.4.1614-1.7.10` (flujo directo, sin instancias) y confirmar que el juego entra sin el error `Missing required option(s) ['userProperties']`.
- Revisar en el log del lanzador que la línea de juego incluye `--userProperties {}` (sustituido desde `${user_properties}`).