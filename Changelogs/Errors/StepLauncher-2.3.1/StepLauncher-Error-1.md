# Errors/StepLauncher-2.3.1/StepLauncher-Error-1.md — El botón Jugar decía "Descargando…" durante la extracción de nativos

- **Fecha**: 2026-08-08
- **Versión**: 2.3.1
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release StepLauncher-2.3.1.

## Síntoma

Al pulsar **Jugar** con librerías faltantes, el botón mostraba "Descargando…" y el globo de mensaje "Descargando archivos faltantes…" incluso cuando el motor ya había terminado de descargar y estaba **extrayendo los nativos**. Además, el globo se quedaba con "Lanzando <versión>…" pegado en pantalla después de iniciar el juego.

## Causa raíz

- En `internal/Core/Launcher/Launcher.go`, `downloadMissingLibraries` emitía eventos `game_prepare` con `phase: "libraries"` (+ `finished: true` al terminar), pero la extracción de nativos (`helpers.ExtractNatives`) no emitía ningún evento: la fase de gráficos quedaba vacía en el frontend y el texto del botón derivaba a "Lanzando…"/"Descargando…" según el estado de un prepare ya finalizado.
- `frontend/web/src/Stores/Launcher.ts` solo conocía una fase implícita "descargando" y no manejaba ni el fin de la extracción ni `game_started`: el mensaje "Lanzando 1.21.4…" se configuraba con `persist=true` (sin timeout) y nunca se limpiaba.

## Diagnóstico y evidencia

- Lectura del flujo `Launch()`: tras `downloadMissingLibraries(...)` (con sus `prepareEmit`), la llamada `helpers.ExtractNatives(...)` recorre cada jar sin emitir progreso.
- Frontend: `LaunchPrepareData` no tenía campo `phase` y `launchPrepareText` solo sabía la frase "Descargando…".

## Solución aplicada

- `helpers.ExtractNatives` ahora acepta `onProgress(cur, total, name)` y notifica 0/total, cada jar y total/total; `Launcher.go` lo traduce a eventos `game_prepare` con `phase: "natives"` (y `finished` al terminar).
- `Stores/Launcher.ts` guarda `phase` y produce textos por fase tanto para el globo (`launchPrepareText`) como para el botón (`launchingPhaseLabel`).
- `App.vue` suscribe `game_started` → `hideLaunchMessage()` para limpiar el globo al arrancar el juego.

## Regla aprendida

Todo paso lento de la fase de lanzamiento (descargas, extracción, verificación) debe emitir su propia fase en `game_prepare`; el frontend nunca debe mostrar una etapa por el estado "default" de otra. Los mensajes de lanzamiento con persistencia infinita deben asociarse a un evento de cierre (p. ej. `game_started`).

## Verificación

- `go build ./...` OK.
- `bun run build` (frontend) OK.
- Comportamiento: al lanzar una versión sin librerías el botón pasa por Descargando… → Extrayendo…, el globo muestra contadores por fase y desaparece al abrir el juego.