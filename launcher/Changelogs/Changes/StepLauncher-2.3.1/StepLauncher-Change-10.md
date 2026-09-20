# Changes/StepLauncher-2.3.1/StepLauncher-Change-10.md

- **Fecha**: 2026-08-09
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.

## Qué cambió

Al abrir cualquiera de los cinco paneles principales (Instancias, Fotos, Cuentas, Versiones y perfiles, Noticias) el contenido del menú principal se oculta ahora con una animación de zoom + desenfoque, y los paneles (y sus sub-paneles y avisos) se cierran con ESC respetando el orden de profundidad: primero el sub-panel/aviso abierto y luego el panel raíz.

### 1. Sistema de capas para ESC (frontend)

- Nuevo composable `Composables/useModalEscape.ts`: cada componente registra una capa `{ handler, isActive, depth }` (se da de alta en `onMounted` y de baja en `onUnmounted`). Hay UN solo listener global de `keydown`, activado perezosamente; al pulsar ESC se ejecuta la capa activa de mayor profundidad (LIFO por profundidad, no deja que dos modales cierren a la vez con una sola pulsación).
- **Profundidades**: 0 = paneles raíz, 1 = sub-paneles y vistas dentro de paneles, 2 = avisos y modales transitorios (confirmaciones, descarga).
- Retirado el listener propio de cada modal que ya lo llevaba para no duplicar: `NewsModal`, `ScreenshotsModal`, `InstanceDetailView` e `InstanceDownloadModal` (sus atajos de flechas/zoom se conservan en su keydown propio; solo Escape pasa al sistema de capas).

### 2. ESC por panel

- `AccountsModal`, `VersionsModal` e `InstancesModal` (profundidad 0): ESC emite `update:visible, false`.
- Sub-paneles (profundidad 1): `AccountFormModal`, `ProfileFormModal`, `InstanceFormModal`, `InstanceSettingsModal`, el visor de capturas (`ScreenshotsModal`: si hay preview abierto, ESC vuelve a la galería; si no, cierra) y la vista de detalle de instancia (`InstanceDetailView`: ESC vuelve a la cuadrícula, igual que antes).
- Avisos (profundidad 2): confirmaciones de Eliminar y Clonar en `InstancesModal` (ESC las cancela sin borrar nada) y `InstanceDownloadModal` (solo si está visible; antes el listener no comprobaba `visible`).
- Comportamiento combinado: con el detalle de una instancia abierto y el aviso de eliminar encima, ESC cierra primero el aviso y el siguiente ESC vuelve a la cuadrícula (dos pulsaciones, dos capas).

### 3. Animación al ocultar el menú principal

- `App.scss`: `.MainContent` ahora tiene `transition` base (opacity/transform/filter 400ms) y `.menuHidden` aplica `opacity: 0`, `visibility: hidden`, `transform: scale(.92) translateY(-12px)` y `filter: blur(8px)`, con `pointer-events: none` y retardo de `visibility` (400ms lineales) para que el fade-out se vea completo antes de quedar oculto; al volver, el fade-in es inmediato en `visibility` y animado en el resto.
- El temporizado (400ms) coincide con `var(--transition-modal)` pero listado explícito (sin background/border) para no interferir con otros estilos del contenedor.

## Por qué

Los cinco paneles se abren sobre el menú pero hasta ahora el contenido de fondo seguía visible (lo pedido era ocultarlo del todo con estética); además, cada modal que cerraba con ESC lo hacía con su propio listener sin tener en cuenta a los demás, de modo que con más de una capa abierta se cerraban varios a la vez o, peor, se cerraba el panel raíz mientras su formulario flotaba abierto. El sistema de capas centraliza la resolución y ordena la profundidad de forma natural.

## API afectada

- Frontend únicamente. Nuevo composable `Composables/useModalEscape.ts` (3 parámetros: handler, `isActive` y `depth`, todos opcionales salvo handler).
- Sin cambios en bindings Wails, backend, config ni `launcher_config.json`.

## Comportamiento anterior/nuevo

- Anterior: al abrir un panel el menú quedaba visible detrás (solo se ocultaba con un fade de opacidad inmediato y sin animación); ESC solo funcionaba en Noticias, capturas, detalle de instancia y descarga, cerrando capas de forma descoordinada; el resto de paneles no se cerraban con teclado.
- Nuevo: el menú se desvanece con zoom (0.92) y desenfoque (8px) al abrir cualquiera de los 5 paneles y reaparece suavemente al cerrarlo; ESC cierra el panel, y con sub-paneles/avisos abiertos cierra primero el de mayor profundidad.

## Cómo verificar

- `bun run build` en `frontend/` (vue-tsc + sass + vite): pasa.
- Abrir cada uno de los 5 paneles: el menú desaparece con la animación (zoom + blur) y vuelve al cerrar.
- ESC en cada panel: lo cierra; con el aviso de eliminar/clonar abierto sobre el Resumen de instancia: 1er ESC cancela el aviso, 2º ESC vuelve a la cuadrícula, 3º ESC cierra el panel.
- En el visor de fotos con una captura en preview: ESC vuelve a la galería; otro ESC cierra el visor.
- En Noticias con un documento abierto en el lector: las flechas siguen navegando y ESC cierra (mismo comportamiento que antes).

## Pendientes (verificables al ejecutar)

- Confirmar visualmente la sensación de la animación (escala 0.92 + blur 8px + 400ms) en monitores grandes y con personalización de fondos claros; ajustar valores si se ve demasiado marcada.