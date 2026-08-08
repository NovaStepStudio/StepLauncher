# StepLauncher-Change-24: Modal de crash de Minecraft (datos completos del evento game_crashed)

- **Fecha**: 2026-08-05
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

Se integra en `App.vue` el modal de crash (`frontend/web/src/Modals/CrashModal.vue`)
que se abre automáticamente cuando el backend emite el evento `game_crashed`.
El modal muestra todos los datos recibidos y una explicación breve por categoría
de crash.

- **Modal reescrito**: `CrashModal.vue` quedó con código huérfano (imports a
  funciones inexistentes, referencias a `crashLog`/`loadingLog` nunca definidos
  y un icono equivocado). Se reescribe limpio:
  - usa `crashInfo`/`clearCrash`/`onGameCrash` del store `stores/launcher.ts`;
  - muestra versión, código de salida, categoría, motivo, tiempo en juego
    (`fmtUptime`), PID, jugador y la ruta del crash report/log si existe;
  - `categoryLabel(cat)` traduce cada categoría (`oom_or_killed`, `java_vm_crash`,
    `game_error`, `killed`, `interrupted`, `signal`) a una explicación en español;
  - se corrige el icono del header (`IconAlertOctagon` en vez de `IconX`) y se
    elimina el estilo `display: none` que ocultaba el svg del icono.
- **Conexión en el store**: la suscripción perezosa a `game_crashed` en
  `stores/launcher.ts` ahora hace `onGameCrash(data)` (rellena `crashInfo`) y
  además `maybeShowWindow()` para que la ventana del launcher vuelva a verse y
  el jugador pueda leer el modal.
- **Integración en App.vue**: se importa `CrashModal`, se declara `showCrash` y
  un `watch(crashInfo)` que abre/cierra el modal según haya datos de crash, y se
  renderiza junto al resto de modales.

## Por qué

Antes un crash de Minecraft solo restauraba la ventana del launcher, sin explicar
al jugador qué había pasado. Ahora se informa de la causa (memory, JVM, mods, etc.)
con los datos que el backend ya enviaba por `game_crashed` pero que nadie
consumía en la UI.

## API afectada

- Ninguna (consumo de eventos del backend ya existentes; sin cambios Go).

## Cómo verificar

- `bun run build` (dentro de `frontend/`) → OK (`vue-tsc --build` + `vite build`).
- Lanzar una versión que crashee (p. ej. RAM insuficiente u OOM): el modal aparece
  con versión, código de salida, categoría, motivo, uptime, PID y jugador, y el
  botón `Cerrar` lo cierra y limpia `crashInfo`.
- Si está activo "Ocultar launcher al abrir Minecraft", la ventana vuelve a
  mostrarse al crashear el juego.