# StepLauncher-Error-12: El panel de Fotos no cargaba la galería y el botón no se actualizaba al cerrar Minecraft

- **Fecha**: 2026-08-06
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

El panel de Fotos abría un modal vacío ("Aún no hay capturas…") aunque en
`game/screenshots` existieran capturas: en DevTools, `ListScreenshots()`
resolvía con el `Array(6)` correcto, pero la UI no mostraba ni un solo
elemento. Además, el botón "Fotos" del sidebar no aparecía ni desaparecía al
cerrar Minecraft tras una sesión nueva con capturas.

## 2. Causa raíz

Dos bugs independientes en el frontend (`frontend/web/src/Modals/ScreenshotsModal.vue`):

### a) La galería nunca pedía datos al abrirse

`refresh()` comienza con `if (!props.visible) return;` y solo se llamaba en
`onMounted`. El modal está SIEMPRE montado en `App.vue`
(`<ScreenshotsModal v-model:visible="showShots" />`), así que el `onMounted`
se ejecutaba al arrancar la app, cuando `visible` era `false`, y el `return`
temprano descartaba la única llamada. No existía ningún `watch` sobre
`props.visible`, por lo que al abrir el modal la lista se quedaba vacía.

### b) Las miniaturas no se renderizaban aunque la lista cargara

`thumbCache` era un `Map` plano: `thumbOf()` devolvía `''` en el primer
llamado y guardaba la blob URL en el `Map` al resolver `ReadLocalFile`, pero
Vue no reacciona a mutaciones de un `Map` no reactivo → el `v-if="thumbOf(...)"`
de la cuadrícula nunca se re-evaluaba y solo se veían los placeholders.

### c) El botón no se refrescaba al cerrar el juego

`checkShots()` (App.vue) solo se llamaba al arrancar y al abrir el modal; no
había ninguna suscripción a los eventos de cierre de Minecraft
(`game_exited`/`game_crashed`/`game_stopped`).

## 3. Solución aplicada

### `ScreenshotsModal.vue`

- `watch(() => props.visible)` → `refresh()` al abrir (y resetea la vista
  ampliada al cerrar).
- `thumbCache` pasa a `reactive(new Map())` para que guardar la blob URL
  re-renderice la cuadrícula y aparezca la miniatura.
- Nuevo listener del evento de ventana `sl:shots-refresh` (que despacha
  App.vue al cerrarse el juego) para recargar la galería si el modal está
  abierto durante el cierre.
- Texto del estado vacío corregido (decía "Cierra el juego y no aparecerá
  este botón", que era confuso).

### `App.vue`

- Suscripción a `game_exited`, `game_crashed` y `game_stopped`: al cerrarse
  Minecraft se vuelve a comprobar si la sesión dejó capturas
  (`checkShots()` → actualiza el botón Fotos) y se despacha
  `sl:shots-refresh` para la galería abierta. Suscripciones dadas de baja en
  `onUnmounted`.

## 4. Verificación

- `bun run build` (incluye `vue-tsc`) → OK.
- Con capturas reales en `game/screenshots`, al abrir Fotos la cuadrícula
  muestra los 6 elementos con sus miniaturas; al cerrar Minecraft el botón
  Fotos aparece/desaparece según la sesión haya dejado capturas.

## 5. Regla aprendida

Un modal que se monta una vez en el árbol principal y se muestra/oculta con
`v-model:visible` NO puede depender de `onMounted` para cargar datos: el
disparador correcto es un `watch` sobre `props.visible`. Y para caches de UI
(blob URLs, etc.) que se mutan en segundo plano, usar `reactive(Map)` en vez
de un `Map` plano.
