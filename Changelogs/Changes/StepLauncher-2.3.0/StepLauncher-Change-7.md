# StepLauncher-Change-7: Animación en TopOptions + preview de colores + fuentes en Sidebar/TopOptions

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

**1) Animación de entrada SOLO en la TopOptions (la sidebar NO se toca):**
los 4 botones de la barra superior (Notificaciones, Noticias, Descargas,
Configuracion) entran con un pequeño deslizamiento vertical + escala,
escalonados (delay 50ms, 100ms, 150ms, 200ms). El toggle global de
animaciones (`data-anim="off"` en `index.css`) la desactiva automáticamente.

- `Styles/App.scss`: `@keyframes TopOptionIn` aplicada a `.TopOptions .Others`
  y a cada `.OptionOther` con `nth-child` delays.

**2) Preview de color: al cambiar un color de la sección "Colores" el panel
de Configuración se oculta por completo salvo el propio selector de color**
(overlay, fondo, header, sidebar y contenido quedan `visibility: hidden`;
solo el `ColorField` abierto sigue visible). Así el usuario ve el color
aplicado en el launcher en tiempo real mientras lo arrastra. La sección
"Tipografías" (colores de letra principal/secundaria) NO oculta el panel.

- `stores/colorfield.ts`: nuevo `previewColorFieldId` compartido.
- `Layouts/Sections/Settings/ColorField.vue`: nueva prop `preview`; al abrirse
  con `preview` publica su id en el store y añade la clase `previewing`; al
  cerrarse (o al desmontarse) lo limpia.
- `Modals/SettingsModal.vue`: computada `previewMode`; el overlay recibe la
  clase `preview` (overlay transparente, `.SettingsModal` oculto y solo
  `:deep(.Cf.previewing)` visible). En preview el click en el overlay no
  cierra el modal (para no perder el color que se está ajustando).
- `PersonalizationSettings.vue`: los 5 ColorField del grupo "Colores"
  (sidebar, modal, botones, bordes de modales, bordes generales) llevan
  `preview`; los de Tipografías no.

**3) Fuentes, colores y tamaños en Sidebar y TopOptions:** los labels de la
sidebar (`.Item_Label`) y de la TopOptions (`.OptionLabel`) usaban la fuente
hardcodeada 'Fredoka' con tamaño fijo; ahora consumen las variables de
personalización que no llegaban a aplicarse ahí:

- `font-family: var(--font-primary)` (con 'Fredoka' como fallback).
- `font-size: calc(.8rem * var(--font-size-primary, 1))`.
- `color: var(--text-primary)` y `text-shadow: var(--text-shadow-primary, none)`.
- `.Others` y los iconos de `.OptionOther` también aplican `--text-primary`.

## Por qué

- Dar feedback visual a los botones nuevos de la TopOptions sin tocar la
  sidebar (que ya tiene su propio comportamiento).
- Poder previsualizar el color elegido sobre el launcher real sin que el
  modal tape la vista (solo para colores de UI, no para tipografías).
- Los parámetros `--text-primary`, `--text-secondary`, `--font-primary`,
  `--font-secundary`, `--font-size-primary` y `--font-size-secundary` no se
  respetaban en sidebar ni TopOptions; ahora sí.

## API afectada

- Ninguna (solo frontend; ninguna prop rompe el contrato existente de
  `ColorField`, `preview` es opcional con default `false`).

## Cómo verificar

- `bun run build` (dentro de `frontend/`) → OK (type-check incluido).
- Al abrir la app, los 4 botones de la TopOptions entran escalonados; la
  sidebar aparece sin animación nueva.
- En Configuración → Personalización → Colores: al fijar cualquier selector
  de color, todo el panel desaparece y solo queda el selector visible sobre
  el launcher; al soltar el color se ve aplicado en vivo. Al cerrar el
  selector vuelve el panel. En Tipografías el panel NO se oculta.
- Cambiando tipografía/color de letra en Ajustes, los tooltips de la sidebar
  y de la TopOptions reflejan la fuente, el tamaño y el color configurados.
