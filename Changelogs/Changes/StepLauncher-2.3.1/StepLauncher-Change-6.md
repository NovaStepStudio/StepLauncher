# Changes/StepLauncher-2.3.1/StepLauncher-Change-6.md

- **Fecha**: 2026-08-09
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.

## Qué cambió

Tanda de pulido y consistencia del frontend sobre la sección de Instancias ya documentada (Change-4/5): se registra todo en una única entrada porque es una sola sesión de trabajo de cohesión visual y de comportamiento.

### 1. Registro de transiciones en variables CSS (elimina el hardcodeo)

- **Inventariado**: se auditaron las ~143 declaraciones `transition:` hardcodeadas en 25 archivos SCSS/Vue de `frontend/web/src` (solo quedan fuera las 4 variables preexistentes de la sidebar y botón Jugar).
- **Nuevas variables en `frontend/web/index.css`** (`:root`): 59 variable `--transition-*` claves por valor canónico (props ordenadas + duración/función), p. ej. `--transition-opacity: opacity .45s ease;`, `--transition-background-border-color: background 130ms, border-color 130ms;`, `--transition-border-color-box-shadow-transform`, `--transition-inset-opacity`, `--transition-width-250ms: width 0.25s ease;`.
- **Desambiguación de nombres**: cuando dos valores distintos tienen las mismas props y solo difieren en duración, el nombre lleva sufijo con la duración (`opacity 120ms` → `--transition-opacity-120ms`); si aún colisionan, sufijo numérico (`--transition-opacity-transform-200ms-3`).
- Las 143 declaraciones pasan a `transition: var(--transition-*);`. Verificado programáticamente: 63 variable usadas = 63 definidas (sin referencias huérfanas ni variables muertas).
- Consecuencia colateral: los toggles de animaciones (`:root[data-anim="off"]`) y la personalización de duraciones por tema siguen funcionando porque el `transition: none !important` y las variables sobreescribibles ya eran parte del sistema.

### 2. Exclusividad de modales pesados (Instancias ↔ Capturas)

- **Nuevo `frontend/web/src/Stores/Modals.ts`**: `heavyPanel` (`'instances' | 'shots' | null`) con `openHeavyPanel`/`closeHeavyPanel`. Un panel pesado abierto cierra el otro: a la vez nunca están montados render ni `InstancesModal` ni `ScreenshotsModal`.
- **`App.vue`**: funciones `openInstances()`/`openShots()` (marcan el heavy, cierran el opuesto con `closeHeavyPanel` y abren su propio v-if) y un `watch(heavyPanel)` que mantiene el estado coherente al cerrar desde fuera.
- **`InstancesModal.vue`**: emite `@open` cuando se abre; desde el detalle, `openShots(name)` guarda `shotsReturn` y el modal de capturas se abre sobre él; al cerrarse, si venía del detalle se reemite `visible=true` (vuelve al detalle sin perder selección) y en caso contrario cierra una capa.
- **`ScreenshotsModal.vue`**: el visor pasa a ser una capa interna (`Shots_Viewer` dentro de `Shots_Overlay` con botón "Volver a la galería") en vez de un segundo overlay a nivel de `App.vue`; el close con Esc/overlay-x usa el mismo canal (`CLOSE_OVERLAYS_EVENT`) y `handleIdle` de `App.vue` cierra también cualquier heavy panel abierto.
- Los down eventos del visor (Esc) y el cierre por idle quedan coherentes: cerrar el visor devuelve a la galería o cierra todo según el camino de entrada.

### 3. Paleta "activado" (filtros y pestañas de Instancias)

- Pestañas activas de `InstanceDetailView.vue` (Versiones/ModLoader/etc.) y el filtro de estado de `InstancesView.vue` (Todas/Jugadas/Descargando) ya no fuerzan tonos propios: usan `--background-button-primary` (mezcla al 12%/25% con `color-mix`) + `--border-style` + `--text-primary`, coherentes con el sistema de variables y visibles sobre fondo oscuro (antes el estado activo quedaba casi invisible).
- El hover y la relación activo/inactivo de las pestañas del detalle también se recalibran con las mismas variables de acento.

### 4. "Abrir carpeta" de una instancia

- **Backend**: nuevo `internal/Core/Launcher/Instance/Folder.go` (`OpenInstanceFolder` en el gestor: valida nombre seguro y abre `instances/<nombre>/` con el explorador del SO).
- **Binding**: `Engine.OpenInstanceFolder` en `internal/Handlers/Engine/Instance.go` y `App.OpenInstanceFolder` en `app.go` (ya regenerado en `frontend/wailsjs`).
- **Frontend**: `Stores/Instances.ts` (`openInstanceFolder`) y botón `.InstDet_Tool` "Abrir carpeta de la instancia" junto a la cabecera del detalle (`InstanceDetailView.vue`).

### 5. Ajustes menores (configuración de instancia y widget de descarga)

- `InstanceSettingsModal.vue`: la rama de Java pasa a un switch con texto claro ("Utilizar el Java que tengas configurado" + campo de ruta solo si está desactivado), el preset de GC se muestra solo con su switch activado, y se elimina la nota explicativa bajo la RAM máxima.
- `DownloadWidget.vue`: el widget muestra ahora la instancia en curso ("Descargando `<instancia>`", `instversion · Progreso X%`, o "Instalando en N instancias" cuando hay varias) y su ancho máximo se acota para que el texto largo no desborde (se mantiene pegado abajo según `hasVersions`).

## Por qué

La sección de Instancias había quedado con estilos aislados (transiciones literales fuera del sistema de variables, estados activos con colores propios, widget sin identificar la instancia) y los modales pesados de Instancias/Capturas podían montarse a la vez (dos overlays de scroll simultáneos). Esto unifica el sistema y evita estados extraños de apertura/cierre.

## API afectada

- Nuevos bindings (sin romper ninguno): `App.OpenInstanceFolder` (backend + `frontend/wailsjs` regenerado por `wails build` previo).
- Store nuevo: `Stores/Modals.ts`. Sin cambios en la API Go previa.

## Comportamiento anterior/nuevo

- Anterior: transiciones hardcodeadas por archivo; Instancias y Capturas montables a la vez; pestañas/filtros activos casi invisibles; sin "abrir carpeta"; widget de descarga genérico.
- Nuevo: 59 variables `--transition-*` en `index.css` usadas en 143 puntos; exclusividad de paneles pesados con vuelta al origen; acento de activado coherente; botón "Abrir carpeta" por instancia; widget identifica la instancia.

## Cómo verificar

- `go build ./...` en la raíz: pasa.
- `bun run build` en `frontend/` (vue-tsc + vite + sass): pasa; el build completo de `wails build` queda pendiente de ejecutar (bindings ya regenerados).
- Abrir una instancia desde la lista y, desde su detalle, abrir Capturas: solo un overlay montado; al cerrar el visor se vuelve al detalle y al pulsar Esc se cierra todo.
- Inspeccionar con devtools que las pestañas activas y el filtro de estado se ven con el acento del launcher y que las transiciones aplican las variables (cambiar `--transition-opacity` en `:root` y ver el efecto global).

## Pendientes (verificables al ejecutar)

- Compilar el binario completo con `wails build` (no ejecutado en esta sesión).
- Verificación visual de la mezcla de acento de pestañas/filtros con fondos claros personalizados (los fondos de `personalization` deben leerse bien en ambos estados).