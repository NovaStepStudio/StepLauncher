# StepLauncher-Change-29: Paleta unificada de borde/fondo para botones de acción

- **Fecha**: 2026-08-06
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

### 1. Nuevas variables de control en `frontend/web/index.css` (`:root`)

Paleta unificada de **fondo y borde** para todos los botones de acción,
seleccionables, icon-tools, toggles y links del frontend. Solo colores base
(sin hover, transiciones ni tamaños):

```css
--control-bg: rgba(255, 255, 255, 0.04);         /* base: .SsBtn, tabs, items, cards */
--control-bg-soft: rgba(255, 255, 255, 0.06);    /* toggles, toolbars, preview */
--control-bg-dark: rgba(0, 0, 0, 0.2);           /* seleccionables sobre modal */
--control-option-bg: #1a1a2e;                    /* fondos de <option> desplegado */
--control-toggle-thumb: #fff;                    /* pastilla del toggle */
--control-border: rgba(255, 255, 255, 0.08);     /* borde base: items, tools, links */
--control-border-strong: rgba(255, 255, 255, 0.1); /* borde fuerte: toggles, nav preview */
```

Los micro-valores sueltos (0.03 / 0.05 / 0.07 / 0.09 / 0.12 / 0.14 / 0.15) se
colapsaron a estos alphas canónicos para que todos los controles compartan
estilo y no haya valores literales repartidos por la app.

### 2. Reemplazo de colores duros por las variables en componentes

- `Styles/Settings.scss`: `.SsStep`, `.SsBtn`, `.SsTgS` (track + thumb),
  `option` del `.SsSel`.
- `Styles/InstallationModal.scss`: `.InstallationModal_FilterTab`,
  `.InstallationModal_Loader`, `.InstallationModal_ToggleTrack`,
  `.InstallationModal_DetailsToggle`.
- `Modals/VersionsView.vue`: `.Vers_Tab`, `.Vers_SearchBox`, `.Vers_Item`,
  `.Vers_Tool`.
- `Modals/AccountsView.vue`: `.Acct_Item`, `.Acct_Tool`.
- `Modals/ScreenshotsModal.vue`: `.Shots_Close`, `.Shots_Card`,
  `.Shots_PreviewTools button`, `.Shots_PreviewZoomBar`, `.Shots_Nav`.
- `Modals/FontManagerModal.vue`, `Modals/CrashModal.vue`,
  `Modals/ProfileFormModal.vue`, `Layouts/Sections/Settings/AboutSettings.vue`
  (`.SsLink`).

### 3. No se tocaron (decisión de alcance)

- Estados `:hover`, activos y `:disabled` siguen con sus valores propios.
- Badges semánticos (`.Vers_TypeHue_*`, `.InstallationModal_VersionBadge`),
  divisores, strokes, paneles de contenedor y colores de texto.

## Por qué

- Cada modal y cada sección repetía los mismos blancos translúcidos con alphas
  distintos y a veces colores fijos (rojos `rgba(255,90,90,…)`, `#1111`, etc.):
  los controles no compartían estilo entre sí.
- Se centraliza la paleta en `index.css` para que un ajuste futuro (o una
  futura personalización en Ajustes) cierre todo el sistema de controles.

## API afectada

- Sin cambios de bindings ni backend: solo `index.css` y SCSS de los
  componentes del frontend.

## Cómo verificar

- `bun run build` (frontend) → OK.
- Revisar visualmente: botones de Ajustes, pills de instalación, tabs de
  versiones, items de cuentas, tools (lápiz/papelera), preview de capturas y
  toggles: todos con el mismo fondo/borde para el mismo rol.