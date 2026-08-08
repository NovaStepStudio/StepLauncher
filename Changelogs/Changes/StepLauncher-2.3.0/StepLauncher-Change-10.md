# StepLauncher-Change-10: Tono (HUE) accesible: rueda del ratón + flechas del teclado

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

El HUE ahora se puede ajustar de dos formas adicionales a la barra de tono,
para que el selector sea usable también sin rueda del ratón:

**1) Rueda del ratón sobre el canvas SV o la barra de tono:** con
`@wheel.prevent` se suma `deltaY * 0.2` al HUE (redibuja el canvas, emite
`preview` en vivo y hace `commit`). Permite cambiar color (sat/val arrastrando
con el ratón) y tono (rueda) a la vez, sin soltar el arrastre. El
`preventDefault` evita que la rueda scrollee el panel de ajustes.

**2) Flechas ↑/↓ del teclado:** con el selector abierto, las flechas ajustan
el HUE en pasos de 5° (15° con Shift), con `preview` + `commit`. Funciona
también mientras se arrastra el color (no interfiere con el arrastre del
ratón), y se ignora si el foco está en el campo de código del color (para no
estorbar al escribir).

**3) Aviso visible:** el panel muestra un texto discreto
(`.CfHint`: "Rueda o flechas ↑/↓ (Shift: 15°): tono") usando las variables CSS
existentes (`--font-secundary`, `--text-secondary`, `--text-shadow-secundary`,
`--font-size-secundary`).

**Archivos tocados:**
- `Layouts/Sections/Settings/ColorField.vue`: `onWheelHue` (rueda),
  `onKeydown` (flechas), listener de `keydown` en mount/unmount, hint
  `.CfHint` en el panel y su estilo.

## Por qué

- Muchos usuarios no tienen rueda del ratón; el ajuste de tono no debía
  depender de ella. Las flechas del teclado son la alternativa universal y
  permiten seguir ajustando sat/val con el ratón a la vez.

## API afectada

- Ninguna: solo frontend; `ColorField` no cambia props ni eventos y no hay
  cambios en bindings Go.

## Cómo verificar

- `bun run build` (dentro de `frontend/`) → OK (type-check incluido).
- En Personalización → Colores, con el selector abierto: girar la rueda sobre
  el canvas SV (o la barra de tono) cambia el HUE en vivo y se guarda; pulsar
  ↑/↓ lo cambia en pasos de 5° (15° con Shift), incluso mientras se arrastra
  el color. Con el foco en el campo del código, las flechas no interfieren.