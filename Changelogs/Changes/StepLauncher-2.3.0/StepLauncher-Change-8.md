# StepLauncher-Change-8: Animación y preview en tiempo real del ColorField

- **Fecha**: 2026-08-05
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

Refinamiento de la funcionalidad de preview de colores de Change-7.

**1) Animación pura (fade) en la TopOptions:** la entrada de los botones de la
barra superior ahora es un fade simple sin desplazamiento ni escala
(`@keyframes TopOptionFadeIn`). Se conservan los delays escalonados y sigue
honrando `data-anim="off"`. La sidebar no tiene animación.

**2) El preview se activa SOLO al arrastrar el cursor (no al expandir):**
antes se ocultaba el panel al expandir el selector; ahora solo mientras el
usuario está moviendo el cursor/mantiene pulsado sobre el canvas SV, la barra
de tono o la de transparencia (pointerdown/pointermove capturados). Al soltar
el clic el panel vuelve.

**3) SOLO algunos colores ocultan la interfaz:** los ColorFields de "Modal" y
"Borde de Modales" han dejado de ocultar el panel (no llevan `preview`). Los
que sí ocultan: Barra lateral, Botones y Bordes generales. Tipografías nunca
ocultan.

**4) Animación de centrado del selector:** al activarse el preview, el
ColorField se mueve al centro de la pantalla con una animación
(fade + deslizamiento + escala) vía Web Animations API; al soltar el cursor
vuelve animado a su posición original mientras el panel reaparece. La entrada
usa `fill: both` y la salida `fill: none` para no dejar transform residual.

**5) Actualización en tiempo real sin guardar hasta soltar:**
- `ColorField.vue` emite ahora el evento `preview` (con el color actual) en
  cada movimiento del arrastre.
- `PersonalizationSettings.vue` escucha `@preview` y aplica el color con
  `applyPersonalization` en vivo (solo CSS/sitio, NO persiste).
- El guardado real sigue ocurriendo en `update:model-value` → `save()`, que se
  dispara al soltar el cursor.

**Archivos tocados:**
- `Styles/App.scss`: `TopOptionFadeIn` (fade puro).
- `stores/colorfield.ts`: `previewColorFieldId` (de Change-7).
- `Layouts/Sections/Settings/ColorField.vue`: removes que se descartan de
  Modal/Borde; `previewing` + `rootEl` + `previewAnim`; `setPreviewing`
  animado; emisión de `preview` en drag; cleanup en unmount.
- `Layouts/Sections/Settings/PersonalizationSettings.vue`: `onPreviewColor`
  (aplicar en vivo) y `@preview` en los tres campos que ocultan.
- `Modals/SettingsModal.vue`: overlay con transición de fondo
  (`transition: background 200ms ease`) para el fade del preview.

## Por qué

- Flujo del preview: solo mientras se manipula el color, solo en algunos
  campos, con animación hacia el centro y de regreso, y con vista previa en
  vivo sin persistir hasta soltar el cursor.

## API afectada

- `ColorField` gana el evento opcional `preview`; la prop `preview` sigue
  siendo opcional (default false). Ningún binding Go cambia.

## Cómo verificar

- `bun run build` (en `frontend/`) → OK (type-check incluido).
- En Personalización → Colores: al arrastrar sobre "Barra lateral", "Botones"
  o "Bordes generales" el panel se oculta, el selector se centra con la
  animación y el color se ve en vivo sobre el launcher sin guardar; al soltar
  vuelve a su sitio y se guarda. En "Modal" y "Borde de Modales" no se oculta
  nada.
- Los botones de la TopOptions aparecen con un fade escalonado.