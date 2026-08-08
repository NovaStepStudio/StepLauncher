# StepLauncher-Change-27: Clasificación de crashes con los códigos de salida oficiales de Minecraft (net.minecraft.ExitCodes)

- **Fecha**: 2026-08-06
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

### 1. Mapeo de códigos de salida oficiales del cliente

El usuario facilitó la clase decompilada `net.minecraft.ExitCodes` del cliente
actual. Se incorpora a `internal/Core/Launcher/Helpers/System.go`:

| Código | Constante del cliente            | Motivo (`CrashReasonLabel`) | Categoría (`CrashCategory`) |
|--------|----------------------------------|------------------------------|------------------------------|
| -1     | `EXIT_CODE_CRASH_DEFAULT`        | `crash`                      | `game_crash`                |
| -2     | `EXIT_CODE_CRASH_WITHOUT_REPORT` | `crash_without_report`       | `game_crash`                |
| -3     | `EXIT_CODE_CRASH_EARLY_NO_MODULES` | `early_crash_no_modules`   | `early_crash`               |
| -4     | `EXIT_CODE_CRASH_EARLY_ARGUMENT_LIBRARY_LOAD` | `early_crash_library_load` | `early_crash` |
| -5     | `EXIT_CODE_CRASH_EARLY_ARGUMENT_PARSE` | `early_crash_argument_parse` | `early_crash`          |
| -6     | `EXIT_CODE_CRASH_SHUTDOWN`       | `crash_on_shutdown`          | `shutdown_crash`           |
| -7     | `EXIT_CODE_VERSION_PARSING_FAIL` | `version_parsing_failed`     | `version_error`           |
| -8     | `EXIT_CODE_CLIENT_WATCHDOG`      | `client_watchdog`            | `watchdog`                |

Antes, -2..-8 caían al genérico `exit_N`/`unknown`, y -1 se trataba como
`killed_or_oom` (ambigüedad): ahora -1 es un crash con reporte de la propia app
(si la detección por patrón de log detecta OOM, `scanLogForCrashPatterns` sigue
ganando y devuelve `oom`).

### 2. Forma unsigned de Windows

En Windows los procesos que terminan con exit code negativo se reportan como
**unsigned 32 bits** (0xFFFFFFFF para -1, 0xFFFFFFF8 para -8). Cada constante
tiene su par (firmada + unsigned) para que la clasificación funcione igual en
Windows y en UNIX (`mcExitCrashDefaultU` etc., calculadas con `int(^uint32(N))`).

### 3. Frontend

`frontend/web/src/Modals/CrashModal.vue` (`categoryLabel`): se añaden
explicaciones en español para las nuevas categorías (`crash`, `early_crash`,
`shutdown_crash`, `version_error`, `watchdog`) y para las que ya emitían los
patrones de log (`oom`, `jvm_launch`).

## Por qué

El usuario estaba descompilando Minecraft y aportó la lista real de exit codes:
sin ella el launcher clasificaba los crashes del cliente moderno como
"código inesperado". Con el mapeo, el modal de crash da un motivo concreto
(p. ej. `-7` = no se pudo interpretar la versión → cliente corrupto, `-3/4/5` =
fallo de arranque temprano, `-8` = watchdog/cuelgue).

## API afectada

- Sin cambios de bindings: `CrashReasonLabel`/`CrashCategory` mantienen firma y
  se usan internamente en `Launcher.go`/`recordCrash`.
- Se añaden motivos nuevas a `CrashModal.vue` (solo texto de UI).

## Cómo verificar

- `go build ./...` → OK.
- `bun run build` (frontend) → OK.
- Con un crash real del cliente (p. ej. killer `-1`/0xFFFFFFFF) el modal debe
  mostrar categoría `crash` y explicación de crash del juego; `-7` → `version_error`.