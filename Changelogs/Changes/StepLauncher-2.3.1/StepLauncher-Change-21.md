# Changes/StepLauncher-2.3.1/StepLauncher-Change-21.md

- **Fecha**: 2026-08-12
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado (build, vet y type-check verificados; verificación en runtime pendiente).

## Qué cambió

El **modal de crash** mejora su estética y muestra **más datos de error** (los datos en **grilla CSS**, no en flex), y la **interfaz ahora detecta los crashes siempre**: antes, con la opción "ocultar launcher al lanzar" desactivada, el juego podía crashear sin que la UI se enterara.

## Detalle de los cambios

### 1. Más datos de error en el evento de crash

- `internal/Core/Launcher/Events.go`: `GameEventData` gana `javaExec` (ruta del Java usado), `maxRam` (MB configurados) y `vanillaVersion` (versión base real, p. ej. `1.21.1`), rellenados desde `inst.PreInfo` (`JavaExec`, `MaxRAM`, `VanillaVersionID`) en `NewGameEventData`.
- `frontend/web/src/Stores/Launcher.ts`: `CrashInfo` incorpora `javaExec?`, `maxRam?` y `vanillaVersion?`; `onGameCrash` los mapea del payload.

### 2. Modal de crash rediseñado (grilla, badge y copiado)

- `frontend/web/src/Modals/CrashModal.vue`: cabecera con **badge de categoría** (`CrashModal_CatBadge`); la sección de datos pasa a una **grilla de 3 columnas** con 9 celdas (Versión, Fecha y hora, Instancia, Jugador, PID, Tiempo en juego, Estado, RAM máxima, Java) y el **Motivo a ancho completo**; el recuadro de explicación usa icono; el pie gana el botón **«Copiar log»** (con estado "copiado"). Helpers: `categoryName()` (etiquetas cortas), `fmtTimestamp()` y `javaName()` (nombre del archivo de Java).
- `frontend/web/src/Styles/Modals/CrashModal.scss`: diálogo con glow de error, cabecera con gradiente y animación de pulso del icono, grilla `repeat(3, 1fr)` con celdas hover, recuadro de explicación tipo alerta y estilos del botón copiar; se elimina la clase muerta `CrashModal_CodesRow`.

### 3. Detección de crash incondicional en la UI

- `frontend/web/src/App.vue`: el listener de `game_crashed` ya no depende de `subscribeWindowHideRestore()` (que solo se registraba desde `hideOnLaunchIfEnabled()` y **retornaba temprano** cuando `cfg.launcher.hideLauncherOnLaunch === false`, por lo que la UI nunca se enteraba del crash). Ahora `EventsOn('game_crashed', ...)` se registra siempre en `onMounted` y ejecuta `onGameCrash(data)` (abre el modal), `maybeShowWindow()` (restaura la ventana si estaba oculta) y `onGameClosed()` (refresca capturas y estado).

## Comportamiento anterior/nuevo

- **Anterior**: con "ocultar launcher al lanzar" desactivado, un crash del juego no abría el modal ni restauraba la ventana (el listener solo existía dentro del flujo de ocultado); el modal mostraba pocos datos y en flex.
- **Nuevo**: el crash abre el modal con más datos (Java, RAM máxima, versión base), organizados en grilla de 3 columnas, con badge de categoría, explicación con icono y botón para copiar el log.

## Cómo verificar

- `go build ./...` OK; `go vet ./...` OK; `bun run type-check` OK (frontend).
- Con `hideLauncherOnLaunch = false`: lanzar una versión y forzar el crash (cerrar el juego de golpe) → el modal debe aparecer con la grilla de datos (Java, RAM máxima, versión base incluidos), el badge de categoría y el botón «Copiar log».