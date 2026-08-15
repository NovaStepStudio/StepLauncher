# Changes/StepLauncher-2.3.1/StepLauncher-Change-7.md

- **Fecha**: 2026-08-09
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.

## Qué cambió

Simplificación del registro de transiciones introducido en el Change-6 (59 variables casi todas redundantes, "la misma mierda con distinto nombre") y mejora del estado activo de los botones de Instancias que quedaban casi invisibles (mezcla sobre negro `--background-button-primary`).

### 1. Registro de transiciones: de 59 variables a 13

- **Idea**: la mayoría de variables repetían el mismo conjunto de propiedades (background, border-color, color, box-shadow, opacity, transform, filter) cambiando solo unos milisegundos (120/130/140/150, 160/180/200, .45s/400...). No aportan nada como nombres semánticos: se colapsan en **4 variables genéricas por duración** que animan ese conjunto de propiedades (si una propiedad no cambia, no anima, así que el comportamiento visual es idéntico):
  - `--transition` (150ms) — el caso general.
  - `--transition-fast` (100ms) — micro hovers (antes 80–100ms).
  - `--transition-slow` (250ms) — paneles/menús (antes 160–250ms).
  - `--transition-modal` (400ms) — fundidos de overlays (antes 400–450ms).
- Se conservan solo las que **sí difieren de verdad** (9): `--transition-width` (250ms), `--transition-stroke` (120ms) y `--transition-stroke-dashoffset` (220ms, barras de progreso), `--transition-rotate` (rebote de iconos), `--transition-grid-template-rows` (acordeones), `--transition-inset-opacity` (labels de la sidebar), `--transition-backdrop-filter-background` (blur de overlays), `--transition-panel` (opacity 200ms + transform 450ms con spring) y `--transition-spring` (border-color/box-shadow/transform 200ms con spring).
- Se mantienen las 4 preexistentes de la sidebar/botón Jugar (`--transition-play-button`, `--transition-sidebar-item`, `--transition-icon-item`, `--transition-label-item`).
- **Total: 13 variables nuevas + 4 preexistentes** (antes 59 + 4). Los 143 puntos de uso se reescribieron con un mapeo literal por nombre (`var(--transition-opacity-transform-130ms-2)` → `var(--transition)`, etc.); cero sobras (`LEFTOVERS: 0`) y verificación programática: 16 usadas = 16 definidas.
- Efectos secundarios deseados: `--transition` (150ms) incluye `filter` y `box-shadow`, así que hovers que antes no animaban blur/sombra ahora lo hacen de forma sutil y consistente; `data-anim="off"` y la personalización siguen intactos.

### 2. Botones activos de Instancias con acento real

- **Causa**: el estado activo usaba `--background-button-primary` (que es `#111`, negro) mezclado al 12–18% sobre fondo oscuro: prácticamente invisible.
- **Nuevo estado activo con `--progress-color`** (el acento verde del launcher, con fallback `--color-success`), siguiendo el lenguaje ya usado por las acciones primarias (Change-5):
  - Filtros `InstView_FilterChip.on`: fondo teñido al 18%, borde al 55%, texto en el acento y un halo suave con `var(--shadow-settings-normal)`.
  - Pestañas `InstanceDetailView` `.active`: fondo al 14%, borde al 45%, texto/em en el acento e icono a opacidad plena (`svg { opacity: 1 }`).
- Los colores siguen siendo variables: si el usuario cambia la personalización, el estado activo se adapta.

## Por qué

59 variables para 6 propiedades × unos pocos milisegundos era insostenible de mantener y de leer (nombres como `--transition-opacity-transform-130ms-2`); 4 duraciones + un puñado de casos especiales cubren todo el launcher. Y los botones activos de la sección de Instancias no comunicaban su estado.

## API afectada

Ninguna. Solo estilo y variables CSS de `frontend/web/index.css`; los puntos de uso en `frontend/web/src` se actualizaron con el mismo número de líneas (143 reemplazos en 25 archivos).

## Comportamiento anterior/nuevo

- Anterior: 59 variables de transición casi idénticas; chips de filtro y pestañas activas negros sobre negro.
- Nuevo: 13 variables (4 genéricas por duración + 9 especiales); activos verdes con halo, texto en acento e icono a plena opacidad.

## Cómo verificar

- `bun run build` en `frontend/` (vue-tsc + sass + vite): pasa.
- En el launcher: pasar de "Todas" a "Jugadas/Descargando" en la grilla (el chip activo se pinta de verde), abrir el detalle de una instancia y cambiar de pestaña (la activa se ve con el acento), y cambiar la personalización para comprobar que el acento sigue al tema.
- Devtools: `:root` solo define 16 `--transition-*`.

## Pendientes (verificables al ejecutar)

- `go build ./...`: sin cambios en backend, no requiere repetición (pasa igual que en Change-6).
- Revisar visualmente los hovers que ahora animan también `filter`/`box-shadow` (efecto añadido de la genérica `--transition`).