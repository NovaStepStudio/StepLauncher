# StepLauncher-Change-16: Rediseño del modal de instalación (filtros de versión, compatibilidad de modloaders, modo auto/manual y layout de paneles)

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

Rediseño total de la fase `setup` de `frontend/web/src/Modals/InstallationModal.vue`:

**1) Filtrado por tipo de versión:** además del buscador, se añaden pestañas
`Todas / Releases / Snapshots / Antiguas` (`versionFilter`) que filtran la lista
del manifest de Mojang. El contador de la cabecera del panel muestra el número de
versiones visibles, y el estado vacío se dibuja con icono y centrado.

**2) Panel de dos columnas (estilo "panel", no bloques apilados):**
- Panel principal (3fr): buscador + filtros + lista de versiones (scroll interno).
- Panel lateral (2fr): lista vertical de modloaders con icono pixel-art, toggle
  del selector manual, y el selector/información de versiones del loader.

**3) Compatibilidad automática de modloaders con la versión de MC elegida:**
- Al cambiar la versión de MC se consulta vía Go (`GetModLoaderVersions`) la lista
  de versiones de **cada** modloader para esa versión (con `Promise.allSettled` y
  token anti-race).
- Los botones de modloaders sin soporte se **deshabilitan** y muestran
  "No disponible para {versión}" (o "No se pudo verificar") bajo el nombre; los
  que están comprobándose muestran un spinner; el loaders seleccionado que falla
  muestra un aviso destacado.
- Vanilla siempre está disponible.

**4) Dos modos de elección de versión del modloader (toggle "Elegir versión del
modloader"):**
- **Desactivado (por defecto):** el usuario solo marca el modloader; se usa la
  versión recomendada automáticamente. Se muestra una tarjeta verde con la
  versión que se instalará.
- **Activado:** aparece un desplegable con **todas** las versiones del modloader
  para la versión de MC seleccionada, sin obligar a revisarlas en el modo auto.

**5) Estados de carga con estilo:** los textos "Cargando versiones del manifest…"
/ "Buscando versiones de X…" / "Comprobando X…" usan un spinner de anillo CSS
(animado con `@keyframes InstallationModal_Spin`) en lugar del GIF pixelado, con
tarjeta contorneada.

**6) Instalación:** `installReady` reemplaza la condición suelta del botón
Instalar (requiere manifest cargado, sin error, y versión efectiva del loader cuando
no es vanilla). La versión efectiva usa `effectiveLoaderVersion` (manual o auto) en
el panel de progreso, el estado de loader y la pantalla de "listo".

## Archivos tocados

- `frontend/web/src/Modals/InstallationModal.vue` (fase setup reescrita + estilos)

## Por qué

- La selección de versión manifest sin filtros dificultaba encontrar versiones
  entre cientos de entradas.
- Los modloaders tienen muchísimas versiones: obligar a elegir una versión era una
  fricción innecesaria. Ahora el modo por defecto resuelve la recomendada y el
  selector manual con todas las versiones es opcional.
- Los loaders incompatibles con la versión de MC elegida se identifican y se
  inhabilitan con su nota, evitando instalaciones que fallarían.
- El layout en paneles (buscador/filtros | botones laterales) separa mejor el
  contenido principal de las acciones.

## Cómo verificar

- `bun run type-check` (dentro de `frontend/`) → OK (exit 0).
- Pendiente de comprobación visual con `wails dev`.