# StepLauncher-Error-4: La barra de tono (HUE) no se podía mover en el ColorField

- **Fecha**: 2026-08-04
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

Tras añadir Pointer Lock + arrastre por deltas (Change-9), al arrastrar la
barra de tono del selector de color el HUE no cambiaba (quedaba congelado o
saltaba a un extremo). El canvas SV (color) sí respondía.

## 2. Causa raíz

En `ColorField.vue` el arrastre por deltas dependía de tres cosas frágiles:

- **`e.buttons & 1` como gate**: durante el Pointer Lock de WebView2 la
  propiedad `buttons` puede reportar `0`, con lo que `onHueDrag` quedaba
  bloqueado (salía antes de acumular `movementY`).
- **Re-anclaje de deltas al concederse el lock** (`onPointerLockChange`
  reseteaba `moveY`/`moveX` y re-anclaba `anchorHue`): WebView2 revoca y
  reconcede el lock cuando el elemento se mueve (la animación de centrado del
  preview desplaza el selector bajo el cursor) → los deltas se ponían a cero
  constantemente y el valor apenas avanzaba.
- **`requestPointerLock` sobre las barras de tono/transparencia**: elementos
  pequeños que se mueven al centrarse en preview → el lock falla/molesta justo
  donde no aporta nada (los bordes de ventana nunca cortaban esas barras).

## 3. Solución aplicada

- El **Pointer Lock solo se pide sobre el canvas SV** (superficie principal).
  Las barras de tono y transparencia usan solo `setPointerCapture` + deltas.
- Se introduce **`dragKind` ('sv' | 'hue' | 'alpha' | null)** como estado real
  del arrastre; se elimina el gate de `e.buttons` en el camino normal. Solo si
  `buttons === 0` sin estar en lock se cierra el arrastre (red de seguridad).
- **Sin re-anclaje de deltas** al concederse el lock: solo se descarta UN
  pointermove (posible salto fantasma del primer evento) con `skipMove`.
- Red de seguridad con `window blur`: si se pierde el foco a mitad de un
  arrastre (Alt+Tab), se cierra el arrastre y se guarda lo previsualizado.

## 4. Verificación

- `bun run build` (dentro de `frontend/`) → OK (type-check incluido).
- Arrastrar la barra de tono mueve el HUE; arrastrar el canvas SV sigue
  funcionando (incluido en preview con el centrado); la barra de transparencia
  también.

## 5. Regla aprendida

No usar `e.buttons` como condición de arrastre sobre un control que pide
Pointer Lock (WebView2 lo reporta a 0 mientras el lock está activo). Y no
poner lock en elementos que se mueven durante el arrastre ni re-anclar los
acumuladores de deltas en `pointerlockchange`: los deltas solo se acumulan
desde el `pointerdown`.