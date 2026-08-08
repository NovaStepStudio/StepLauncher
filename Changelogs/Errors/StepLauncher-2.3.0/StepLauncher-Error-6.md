# StepLauncher-Error-6: Guardado continuo del color durante el arrastre ("a lo loco")

- **Fecha**: 2026-08-04
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

El ColorField guardaba el color en cada movimiento del ratón durante el
arrastre (Change-11 llamaba a `commit()` dentro de `onSVDrag`, `onHueDrag` y
`onAlphaDrag`). Eso disparaba `save()` → `UpdatePersonalization` (persistencia
en disco vía Go) a ráfaga de eventos, con tirones y escrituras continuas. El
comportamiento pedido es: **guardar SOLO al soltar** el arrastre.

## 2. Causa raíz

En Change-11 se añadió `commit()` (emit `update:model-value` → `save()` del
padre) en cada `pointermove` para "no perder el color". Resultaba en guardados
continuos durante el arrastre con escrituras a disco cada frame.

## 3. Solución aplicada

- Se elimina `commit()` de `onSVDrag`, `onHueDrag` y `onAlphaDrag`.
- Durante el arrastre solo se emite `preview` (aplicado en vivo en el launcher
  SIN persistir), y el valor se persiste (`commit()`) únicamente al soltar
  el botón (`onSVUp`/`onHueUp`/`onAlphaUp`), al perder el foco la ventana
  (`onWindowBlur`) o en los gestos discretos: rueda del ratón
  (`onWheelHue`), flechas del teclado (`onKeydown`), historial recents y
  campo de código.
- Se mantienen `lastEmitted` y su declaración previa al watcher (ver
  Error-5), comentarios actualizados.

## 4. Verificación

- `bun run build` (dentro de `frontend/`) → OK (type-check incluido).
- Arrastrar un color: el selector previsualiza en vivo sin escribir en disco y
  guarda al soltar; la rueda y las flechas guardan al usarlas.

## 5. Regla aprendida

No persistir por cada `pointermove`: el arrastre debe emitir solamente
`preview` (vista previa efímera) y diferir el guardado al `pointerup`. Antes
de volver a añadir un commit en un drag, confirmar con el usuario si quiere
guardado continuo o al soltar.