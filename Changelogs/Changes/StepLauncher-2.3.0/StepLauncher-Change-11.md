# StepLauncher-Change-11: Cursor capturado en tono/opacidad + guardado continuo del color

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

**1) Pointer Lock también en la barra de tono (HUE) y la de opacidad
(Alpha):** antes el cursor solo se capturaba/centraba al arrastrar el canvas
SV (Color); ahora `tryLock` se pide también en `onHueDown` y `onAlphaDown`,
con lo que al arrastrar cualquiera de los tres controles el cursor queda
capturado de la misma manera (inmóvil y centrado sobre el control).
`onPointerLockChange` reconoce los tres controles y `onHueUp`/`onAlphaUp`
liberan el lock. El `skipMove` (descarte de un pointermove fantasma tras
concederse el lock) se consume en los tres drags. Si WebView2 revoca el lock
a mitad del arrastre (el centrado de preview mueve el elemento), el arrastre
sigue con `setPointerCapture` + deltas sin cortarse.

**2) Guardado continuo del color ("a lo loco"):** `commit()` (emite
`update:model-value` → `save()` del padre, que persiste con
`UpdatePersonalization`) se llama ahora en CADA movimiento del arrastre
(SV, hue y alpha), además de en la rueda, las flechas, el historial y el
campo de código. El color ya no se pierde si el selector se cierra o la
ventana pierde el foco sin soltar el botón.

**3) Guard del feedback del v-model:** como ahora se emite en cada
movimiento, el watcher de `props.modelValue` ignora los valores emitidos por
el propio campo (`lastEmitted` se fija en `commit()`); si no, el feedback del
padre dispararía `syncFromValue` y tironearía del thumb durante el arrastre.
Los cambios externos (padre, recents, código) siguen sincronizándose igual.

**Archivos tocados:**
- `Layouts/Sections/Settings/ColorField.vue`: `tryLock` en hue/alpha,
  `onPointerLockChange` con los tres controles, `unlockPointer` en los ups,
  `commit()` en los tres drags, `lastEmitted` en `commit()` y en el watcher.

## Por qué

- La captura del cursor debe ser consistente entre Color, tono y opacidad:
  al seleccionar opacidad o hue el ratón no se centraba con el control como
  sí hacía con el Color.
- El color ajustado debe quedar guardado siempre, sin depender de soltar el
  botón justo en el momento correcto.

## API afectada

- Ninguna: solo frontend; `ColorField` no cambia props ni eventos y no hay
  cambios en bindings Go.

## Cómo verificar

- `bun run build` (dentro de `frontend/`) → OK (type-check incluido).
- En Personalización → Colores: arrastrar la barra de tono o la de opacidad
  captura el cursor igual que el canvas SV; el color (HEX/RGBA) se actualiza
  y se guarda en cada movimiento del arrastre (recargando la app o el panel,
  el color queda persistido); el thumb no "tirona" durante el arrastre.