# Changes/StepLauncher-2.3.1/StepLauncher-Change-14.md

- **Fecha**: 2026-08-09
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado (build verificado; verificación en runtime pendiente).

## Qué cambió

El registro de versiones de una instancia se hace ahora leyendo la carpeta `versions/` sin condicionarse al `<version>.json`, la descarga/verificación en curso se ve en la interfaz aunque se cierre el modal de instalación (banner como visualizador temporal + botones deshabilitados), y los modales de instalación recuperan su estado al abrirlos: retoman la descarga activa o se restablecen al menú, con el reset automático gobernado por la opción "Minutos de Inactividad" del usuario.

### 1. `syncVersionsFromDisk` lee la carpeta `versions/` (no el json)

- `internal/Core/Launcher/Instance/Helpers.go`: se elimina el filtro que exigía `<version>.json` en disco. Ahora se registra **todo** subdirectorio presente en `versions/`: la carpeta es la fuente de verdad y no se condiciona al json (corrección de la petición anterior, que solo registraba "versiones completas").
- Comportamiento idéntico en el flujo: se sigue invocando solo al terminar la instalación COMPLETA (descarga base + modloader) desde la goroutine de `InstallModLoader` (`Modloader.go`).

### 2. La instancia muestra la descarga/verificación en curso aunque se cierre el modal

- `frontend/web/src/Modals/InstanceDetailView.vue`: nuevo **visualizador temporal sobre el banner del hero** (`InstDet_DlOverlay`) alimentado por el store (`downloads[name]`, que se actualiza con los eventos `download_*` aunque el modal esté cerrado): estado ("Descargando", "Verificando archivos", "En pausa"…), versión, porcentaje, barra de progreso y botón Cancelar.
- Botones deshabilitados mientras hay descarga activa: "Jugar", "Descargar" (hero, con texto "Descargando…") y "Añadir versión" (pestaña Versiones).
- `frontend/web/src/Modals/InstancesView.vue` (grid): barra fina de progreso sobre el banner de la tarjeta (`InstCard_BannerDl`) y botón "Jugar" deshabilitado mientras descarga activa.
- `frontend/web/src/Stores/Instances.ts`: helper compartido `dlStateText(state)` para los textos de estado de descarga.

### 3. Los modales de instalación recuperan/reinician su estado al abrirlos

- `frontend/web/src/Modals/InstallationModal.vue` (descarga global):
  - `syncFromBackend` (se ejecuta en cada apertura): si hay una descarga activa que **no pertenece a una instancia** (las de instancia las gestiona su propio modal, se excluyen por su `dlId`), se retoma su visualización con la versión que estaba descargando; si el estado previo era `done`/`error` (instalación ya terminada), se restablece al menú de instalación sin tener que pulsar "Nueva instalación".
- `frontend/web/src/Modals/InstanceDownloadModal.vue` (instalación en instancia):
  - `syncFromBackend`: si la instancia tiene descarga activa (estado del store), se adopta y se retoma la vista de instalación; si el estado previo era `done`/`error`, se restablece al menú.
- Ambos modales: el resumen de instalación terminada (`scheduleDoneReset`) ahora se mantiene durante los "Minutos de Inactividad" configurados por el usuario (`idleOptions.idleMinutes`, opción de Ajustes) en lugar de los 5 minutos fijos; al cumplirse, vuelve solo al menú de instalación.

## Por qué

- El registro anterior exigía `<version>.json` en disco ("leer lo que dice el json") y la petición fue explícita: leer la carpeta `versions/` tal cual.
- Al cerrar el modal de descarga, el detalle de la instancia no mostraba ningún indicio de la descarga en curso (el estado vivía solo en el store y en la tarjeta del grid): parecía que la instancia no hacía nada.
- Los modales mantienen el estado de su última sección al reabrirse; si la instalación había terminado o fallado, el usuario debía pulsar "Nueva instalación" a mano para volver al menú.

## API afectada

- Backend Go: `Helpers.go` (Instance) — únicamente interna a `InstallModLoader`; sin cambios en bindings ni config.
- Frontend: `Stores/Instances.ts`, `Modals/InstanceDetailView.vue`, `Modals/InstancesView.vue`, `Modals/InstallationModal.vue`, `Modals/InstanceDownloadModal.vue` y sus hojas SCSS en `Styles/Modals/`.

## Comportamiento anterior/nuevo

- Anterior: `syncVersionsFromDisk` solo registraba versiones con su `<version>.json` en disco; al cerrar el modal de instalación la UI de la instancia no indicaba descarga/verificación; al reabrir un modal en `done`/`error` había que pulsar "Nueva instalación"; el reset del resumen era fijo a 5 minutos.
- Nuevo: se registra todo directorio de `versions/`; el banner (detalle) y la tarjeta (grid) muestran la descarga/verificación en vivo con cancelación y los botones de jugar/descargar se deshabilitan hasta completar; al abrir un modal se retoma la descarga activa o se restablece el menú; el reset usa los "Minutos de Inactividad" configurados.

## Cómo verificar

- `go build ./...` sin errores y `bun run build` (frontend, incluye `vue-tsc`) sin errores.
- Runtime (pendiente):
  1. Instalar un modloader con procesadores (Forge/NeoForge) en una instancia → al completar, `instance.metadata.json` incluye la versión del loader aunque esta no tenga su `<version>.json` exacto en disco.
  2. Cerrar el modal de instalación a mitad de la descarga → el detalle muestra el visualizador en el banner (estado/%, cancelar), "Jugar"/"Descargar" deshabilitados y la tarjeta del grid con la barra; al terminar/cancelar desaparece.
  3. Reabrir el modal de instalación durante una descarga activa → retoma la vista de instalación; reabrir con la instalación terminada/fallida → vuelve directo al menú.
  4. Cambiar "Minutos de Inactividad" en Ajustes → el resumen de instalación vuelve solo al menú tras ese tiempo.