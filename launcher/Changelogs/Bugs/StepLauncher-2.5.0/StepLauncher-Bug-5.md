# Bugs de StepLauncher 2.5.0 (Bug-5) — Botones secundarios pegados al fondo del modal, hover imperceptible

- **Fecha**:   2026-09-25
- **Versión**: 2.5.0
- **Estado**:  en corrección
- **Release**: StepLauncher-2.5.0 — arreglo incluido en esta release (entrada en corrección al cierre).

## El bug en cuestión

Reporte del owner con capturas: hay botones que se pegan muchísimo al color del modal/fondo y el cambio de color al hover es imperceptible. Lugares: Acerca de (ahora mismo), Añadir cuenta / Configurar, Nuevo perfil de Minecraft y Descargar un mod.

## Qué afectaba y qué hacía

- Todos esos botones usan las clases globales `.SsBtn` (secundario) y `.SsBtnPrimary` de `Common/Styles/Components/_buttons.scss`.
- Causa raíz: el modal es `#0b0b0b` (`--background-modal-primary`), el secundario en reposo es blanco 4% (`--control-bg` ≈ `#121212`) con borde `#131313` (`--border-style`): bordes y relleno casi idénticos al fondo. El hover solo subía a blanco 8% sin tocar borde ni posición, así que el cambio era imperceptible (y con personalizaciones de fondo podía desaparecer del todo).
- El primario en reposo (`--background-button-primary` al 12%) también quedaba demasiado pegado al fondo.

## Solución final

- **`Common/Styles/Components/_buttons.scss`** (arreglo único global, cubre los 4 lugares reportados y todos los que usan estas clases):
  - `.SsBtn` en reposo: borde `rgba(255, 255, 255, 0.14)` para recortar el botón del fondo sin cambiar su relleno personalizable.
  - `.SsBtn:hover`: fondo blanco 12% + borde blanco 26% + `translateY(-1px)` — triple señal imposible de no ver.
  - `.SsBtn:active`: fondo blanco 16% y vuelta a su sitio (respuesta al clic).
  - `.SsBtn:focus-visible`: anillo de foco blanco 50% para navegación por teclado.
  - `.SsBtn:disabled:hover`: sin desplazamiento (no promete clic).
  - `.SsBtnPrimary`: reposo 12%→16% de tinta y borde 25%→32%; hover 28% de tinta, borde 50% y `translateY(-1px)`; active 34% sin desplazamiento.
- Se revisaron los overrides de dominio (`.SsBtn` en `ProfileForm.scss`, `DownloadDialog.scss`, etc.): solo tocan padding/margen, sin choques con el cambio global.

## Comportamiento anterior/nuevo

- Antes: secundario ≈ fondo del modal; hover apenas perceptible.
- Ahora: botón recortado en reposo y hover inconfundible (más claro + borde marcado + se eleva 1px).

## Verificación

- Pendiente: el owner pidió no compilar en esta tanda, así que `bun run build` en `launcher/frontend` y `go build ./...` en `launcher/` quedan para la próxima sesión o para el owner antes de la release. Verificación visual: abrir los 4 lugares reportados y comprobar reposo + hover + clic.
