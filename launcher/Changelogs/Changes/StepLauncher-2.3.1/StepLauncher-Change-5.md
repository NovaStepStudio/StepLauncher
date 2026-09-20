# Changes/StepLauncher-2.3.1/StepLauncher-Change-5.md

- **Fecha**: 2026-08-09
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.

## Qué cambió

Rediseño de la sección **Instancias** según feedback de pruebas: la descarga se integra en la propia instancia (vista de detalle al estilo de la sección de Noticias), cada instancia puede instalar **modloaders** y **varias versiones** de Minecraft, los recursos compartidos quedan decididos en el backend (sin opción de usuario) y el menú de 3 puntos cierra al hacer clic fuera. También se arreglan los botones/selects/inputs que se veían sin estilos.

### 1. Navegación lista ⇄ detalle (cambio de sección)

- Al hacer clic en cualquier parte de una tarjeta se abre la **vista de detalle** de la instancia: `InstancesModal.vue` tiene ahora `view` (`'list' | 'detail'`) y `selected`; la cabecera cambia de título/subtítulo y añade un botón de volver (`.InstancesModal_Back`).
- `InstancesDetailView.vue` (nuevo) muestra: hero con banner/icono/nombre, chips, acciones (Jugar, favorito, editar, configurar, eliminar), sección "Versiones instaladas" (versión activa, cambiar, eliminar), "Modloader" (quitar), "Integridad" (verificar) y "Capturas" (miniaturas que abren el visor).
- `InstancesModal.vue` centraliza los modales secundarios (crear/editar, configurar, capturas) y la confirmación de borrado (`.InstDelete_*`); si borras desde el detalle vuelve a la grilla.
- `InstancesView.vue` deja de tener sheets internos y solo emite `open/new/edit/settings/delete`.

### 2. Panel de descarga integrado (réplica del instalador)

- `InstanceDetailView.vue` contiene el panel con buscador, pestañas Releases/Snapshots/Antiguas (desde `FetchVersionManifest`), lista de versiones con badges INSTALADA/ÚLTIMA/SNAPSHOT, selector de ModLoader en el lateral con compatibilidad (`GetModLoaderVersions`), versión recomendada automática o selector manual, y flujo de progreso con cancelar + fase de modloader con logs.
- Cada instancia admite **varias versiones** descargadas (tantas como quieras) y elegir con cuál jugar.
- El formulario de creación (`InstanceFormModal.vue`) ya **no descarga versiones**: crea solo el metadato; la descarga vive en el panel del detalle.

### 3. ModLoaders por instancia (nuevo backend + API)

- `internal/Core/Launcher/Instance/Modloader.go` (nuevo): `InstallModLoader(name, loader, loaderVersion, mcVersion)` valida que ya exista `instances/<nombre>/versions/<mc>`, lanza `orchestrator.Install(sessionID, ...)` con target el directorio de la instancia y devuelve `ml-inst-<ts>`; más `InstalledModLoader(name)` y `RemoveModLoaderState(name)`.
- `internal/Handlers/Engine/Instance.go` + `app.go`: bindings `InstallInstanceModLoader`, `GetInstalledInstanceModLoader`, `RemoveInstanceModLoaderState`.
- Store `Instances.ts`: `installInstanceModLoader` / `getInstalledInstanceModLoader` / `removeInstanceModLoaderState`; el detalle escucha los eventos `modloader_*` (`EventsOn`) y los desregistra al desmontarse.

### 4. Recursos compartidos fijados en el backend

- La reutilización de librerías/assets/cache globales es decisión del motor (ya existía: `instanceDir/sharedDir/cacheDir` en `internal/Core/Launcher/Instance`). No se expone ninguna opción en la UI; la tarjeta lo menciona como texto informativo.

### 5. Fix de estilos en botones/selects/inputs (feedback de pruebas)

- **Causa raíz**: `InstanceFormModal.scss` y `InstanceSettingsModal.scss` no hacían `@use '../Shared/Settings.scss'`, así que las clases compartidas (`SsIn`, `SsSel`, `SsBtn`, `SsBtnPrimary`, `SsBtnDanger`) no se aplicaban en esos dos modales.
- Se añade el `@use` a ambos y también a `InstancesModal.scss` (necesario para el confirm de borrado) y `InstanceDetailView.scss` (nuevo, todo con variables CSS existentes).
- `InstancesView.scss` reescrito para el marcado actual (banner, chips, menú desplegable, descarga mini) y `InstanceDetailView.scss` creado desde cero (hero, secciones, panel de descarga, loaders, progreso, logs, miniaturas).

### 6. Bindings regenerados y limpieza del store

- El `wails build` previo regeneró `frontend/wailsjs` (models `InstanceInfo`, `GetInstanceResult`, `InstalledLoader`, `ModLoaderInstallResult`, etc.).
- `Stores/Instances.ts` pasa a usar los bindings tipados `@wailsjs/go/main/App` en vez de `window.go`; `CreateInstanceBinding` y `UpdateInstanceConfig` reciben cast a los modelos generados (`instance.CreateInstanceReq`, etc.) porque sus clases exigen `convertValues`.
- Iconos inexistentes corregidos: `IconPhotoPlus` / `IconPhoto` (en vez de los que no exportea `@tabler/icons-vue`).
- `v-model` con expresiones no permitidas → `:visible` + `@update:visible`.

## Por qué

El feedback de las pruebas indicaba: botones/selects/inputs sin estilos, la experiencia de descarga al crear era mala y debía integrarse en la instancia, recursos compartidos no configurable por el usuario, la carga de la lista se veía fea, tocar la tarjeta debía abrir el detalle (como el apartado de Noticias), el menú de 3 puntos no cerraba al hacer clic fuera, y cada instancia debía poder instalar **modloaders** y **varias versiones**.

## API afectada

- Nuevos métodos en `app.go`/`internal/Handlers/Engine/Instance.go` (modloader por instancia). Ningún binding previo se ha eliminado o cambiarle la firma.
- Backend: `internal/Core/Launcher/Instance/Modloader.go` nuevo; `internal/Core/ModLoader/` sin cambios (el Orchestrator ya recibía el directorio destino).
- Frontend: `InstanceDetailView.vue` + `.scss` nuevos; `InstancesView.vue`, `InstancesModal.vue`, `InstanceFormModal.vue` reestructurados; `Stores/Instances.ts` con bindings `@wailsjs` tipados.

## Comportamiento anterior/nuevo

- Anterior: tarjetas con sheets internos (añadir versión/clonar/verificar), descarga solo al crear la instancia, sin modloaders, menú que no cerraba, controles sin estilos compartidos.
- Nuevo: tarjeta → detalle de instancia (sección), descargas múltiples con modloaders dentro de la instancia, recursos compartidos propios del motor, menú con click-outside, todo con estilos del sistema de variables.

## Cómo verificar

- `go build ./...` en la raíz: pasa.
- `bun run build` en `frontend/` (vue-tsc + vite): pasa.
- Crear una instancia (sin versión) → abrir su detalle → descargar 2 versiones → instalar un modloader → cambiar la versión activa → jugar → eliminar una versión.
- Clic fuera del menú de 3 puntos → se cierra. Campos del formulario y la configuración usan `SsIn/SsSel/SsBtn`.

## Pendientes (verificables al ejecutar)

- Falta compilar el binario completo con `wails build` (bindings ya regenerados; `bun run build` y `go build` pasan).
- Progreso de archivos (`filesDownloaded/filesTotal`) depende del payload de eventos del `Downloader` (si se ve 0, verificar con `wails dev`).
- Si `GetModLoaderVersions` devuelve lista vacía para una versión vieja, el loader se muestra "No disponible".

### 7. Segunda ronda de feedback (2026-08-09)

Ajustes según nuevas pruebas, manteniendo el diseño de detalle:

- **Versión al crear restaurada**: `InstanceFormModal.vue` vuelve a tener selector de versiones (Releases/Snapshots) y la descarga opcional al crear; tras crear la instancia, `InstancesModal` abre automáticamente el detalle de la nueva instancia (evento `created`) para seguir descargando más versiones o instalar un modloader.
- **Switches en vez de checkboxes** (petición explícita): se usan los toggles ya existentes del sistema compartido (`SsTg` + `SsTgS`) en favorito y "Descargar ya" del formulario, en los tres switches de configuración (Java oficial, pantalla completa, resolución custom) y en "Elegir versión del modloader" del detalle.
- **Menú de 3 puntos**: ahora se despliega inmediatamente bajo el botón (wrapper `.InstCard_MenuWrap` + `top: calc(100% + 6px)`), la tarjeta ya no recorta (`overflow` movido al banner) y contiene **Clonar** y **Verificar** y **Capturas** además de Editar/Configurar/Eliminar (feedback: faltaban clonar y capturas; el menú "aparecía arriba del todo").
- **Overlays centralizados en InstancesModal**: borrar (confirmación), **clonar** (con input de nombre, usa `CloneInstance`) y **verificar** (resultado por versión con `verifyInstance`) con el patrón `InstOverlay_*`.
- **Botón primario con color real**: `SsBtnPrimary` global usa una mezcla al 12% (casi invisible, "no cambia de color"): los botones de crear/guardar/clonar/descargar y los botones Jugar de tarjeta y detalle usan `var(--color-progress)` (acento del launcher) con texto oscuro, hover con brillo y estado disabled gris.
- El formulario curif no descarga, pero el detalle sí; se mantiene el flujo completo de descarga/multiple versiones.

#### Cómo verificar (ronda 2)

- Crear instancia eligiendo versión → se descarga y se abre el detalle; añadir más versiones desde el panel.
- Los switches de formulario/configuración son toggles y responden al clic en toda la fila.
- Menú de puntos: se cierra al hacer clic fuera, se abre bajo el botón, y permite Clonar (sheet con nombre), Verificar (resultados), Capturas, Editar, Configurar y Eliminar.