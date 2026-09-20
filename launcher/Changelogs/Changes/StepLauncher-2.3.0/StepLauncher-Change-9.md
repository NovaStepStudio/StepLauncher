# StepLauncher-Change-9: Pointer Lock + arrastre por deltas en el ColorField

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

Dos partes distintas sobre el arrastre del selector de color, ambas solo en el
frontend (navegador/WebView2), sin tocar Go:

**1) Captura del cursor (Pointer Lock del navegador):** al hacer `pointerdown`
sobre el canvas SV, la barra de tono o la de transparencia, el selector pide
`requestPointerLock()` sobre ese elemento. El cursor queda capturado (inmóvil
y centrado sobre el selector) mientras se arrastra, de modo que:

- El arrastre no se corta al llegar al borde de la ventana (se puede seguir
  arrastrando indefinidamente).
- El selector centrado en preview "se lleva" el control con el ratón.

Si el navegador no concede el lock (o no lo soporta) se degrada sin romperse
al arrastre por deltas con `setPointerCapture`. Se escucha
`pointerlockchange` para saber si el lock está activo y se reancla el punto de
partida al concederse; `exitPointerLock()` se llama al soltar, al cerrar el
selector y al desmontar.

**2) Arrastre por deltas anclado al pointerdown (`movementX/movementY`):** el
valor ya no se mapea con coordenadas absolutas `clientX/clientY` en cada
`pointermove`. En el `pointerdown` se ancla la posición del thumb y el punto
de partida; en cada movimiento se acumula `movementX/movementY` y se desplaza
el valor desde el ancla con escala 1:1 (misma sensibilidad, sin reducirla).
Esto elimina el "disparo" del thumb que aparecía al centrarse el selector en
preview: la animación de centrado mueve el elemento bajo el cursor mientras se
arrastra y el mapeado absoluto hacía saltar el valor aunque el ratón no se
moviera. Con el anclaje, el thumb solo cambia con el movimiento real del ratón.

**Archivos tocados:**
- `Layouts/Sections/Settings/ColorField.vue`: `tryLock`/`unlockPointer`/
  `onPointerLockChange`; anclas y acumuladores de deltas
  (`moveX`/`moveY`/`dragRect`); reescritura de los pares down/drag/up de SV,
  hue y alpha; `unlockPointer()` en `closeNow`/montaje/desmontaje; listener
  `pointerlockchange` en mount/unmount.

## Por qué

- El cursor debía quedar capturado y centrado sobre el ColorField al
  arrastrar (con APIs del navegador) para que el arrastre no se "fuera" de
  los límites del selector.
- El thumb se desplazaba muy rápido / de forma errática al centrarse el
  selector en preview por mapear coordenadas absolutas contra un elemento en
  movimiento; se arregla anclando el valor y usando deltas, manteniendo 1:1
  la sensibilidad (no se bajó la velocidad).

## API afectada

- Ninguna: solo frontend; `ColorField` no cambia props ni eventos y no hay
  cambios en bindings Go.

## Cómo verificar

- `bun run build` (dentro de `frontend/`) → OK (type-check incluido).
- En Personalización → Colores: al arrastrar sobre un campo con `preview`
  (Barra lateral, Botones, Bordes) el cursor queda capturado y el selector
  centrado sigue 1:1 al ratón sin que el thumb se dispare mientras se centra;
  se puede arrastrar sin cortarse aunque el cursor pase el borde de la
  ventana. Al soltar, el cursor reaparece y el color se guarda. En campos sin
  preview el cursor también se captura mientras se arrastra. Si el lock no se
  concede, el arrastre por deltas funciona igual.