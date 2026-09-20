# StepLauncher-Change-12: Rueda y flechas del teclado solo previsualizan (guardado al soltar/cerrar)

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

Tras Change-11 la rueda del ratón y las flechas del teclado hacían `commit()`
(guardar en disco vía `UpdatePersonalization`) en cada notcha/pulsación,
persistiendo "a lo loco". Ahora:

**1) La rueda y las flechas solo `preview`:** `onWheelHue` y `onKeydown`
actualizan el HUE en vivo (redibuja el canvas y emite `preview`, que aplica el
color sin persistir) y **no** llaman a `commit()`.

**2) Guardado diferido al "soltar":** el color se persiste solo cuando se
cierra el selector (`closeNow` o el auto-cierre de 5s de `scheduleClose`),
con `commitIfDirty()`: persiste únicamente si `outputText() !== lastEmitted`
(es decir, si el usuario cambió algo con la rueda/las flechas y no lo guardó
antes). Abrir y cerrar sin tocar nada no genera escrituras.

**3) `syncFromValue` marca el valor como guardado:** al sincronizar desde
fuera (prop, historial o campo de código), `lastEmitted` se actualiza al valor
resultante, de modo que el cierre no hace un commit falso de algo sin tocar.

**Archivos tocados:**
- `Layouts/Sections/Settings/ColorField.vue`: `commitIfDirty()` (nuevo),
  usado en `scheduleClose` y `closeNow`; `commit()` solo en los ups de
  arrastre y `onWindowBlur`; `onWheelHue`/`onKeydown` sin `commit`;
  `syncFromValue` actualiza `lastEmitted`.

## Por qué

- El guardado debe ocurrir "al soltar": ni el arrastre ni la rueda ni las
  flechas deben persistir a cada evento. La rueda y las flechas, por no tener
  gesto de soltado, persisten al cerrar el selector y solo si hubo cambios.

## API afectada

- Ninguna: solo frontend; `ColorField` no cambia props ni eventos y no hay
  cambios en bindings Go.

## Cómo verificar

- `bun run build` (dentro de `frontend/`) → OK (type-check incluido).
- Con el ColorField abierto: girar la rueda o pulsar ↑/↓ cambia el HUE en vivo
  sin escrituras a disco hasta que el selector se cierra (auto-cierre o
  toggling); si solo se abre y cierra sin tocar nada, no se persiste nada.