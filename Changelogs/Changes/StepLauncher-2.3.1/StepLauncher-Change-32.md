# Changes/StepLauncher-2.3.1/StepLauncher-Change-32.md

- **Fecha**: 2026-08-14
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado (bun run build OK + go build ./... OK).

## Qué cambió

Refactor integral del frontend (`frontend/web/src`): reorganización por **dominios** (feature-first), sistema global de overlays con Host, paneles reales y visibilidad centralizada. Es una tarea única de arquitectura con tres bloques:

### 1. Reestructuración por dominios (F1)

- Todos los componentes, stores y estilos se movieron de `Modals/`, `Stores/`, `Widgets/`, `Layouts/`, `Composables/` y `Styles/` a **carpetas por dominio**: `Accounts/`, `Instances/`, `Versions/`, `Settings/` (con `Settings/Sections/`), `Downloads/`, `News/`, `Welcome/`, `Updates/`, `Crash/`, `Login/`, `Screenshots/`, `Launcher/`.
- Dentro de cada dominio los archivos son **planos con roles**: el archivo raíz lleva el nombre del dominio (`Instances.vue`, `Settings.vue`) y los subcomponentes usan roles (`Form.vue`, `Detail.vue`, `List.vue`, `Content.vue`, `Manager.vue`, `Widget.vue`, `Installation.vue`...).
- Lo transversal vive en `Common/`: `Common/Components/` (ColorField), `Common/Composables/` (useOverlayEscape, SkinPlayer — sin consumidores), `Common/Stores/` (Ui, Idle, Fonts), `Common/Overlays/` (Store, Host, Confirm) y `Common/Styles/`.
- **Estilos**: cada dominio tiene su `Styles/` colocalizado. El antiguo `Styles/Shared/Settings.scss` (con `SsBtnDanger` duplicado) se dividió en el módulo compartido `Common/Styles/Components/` (barrel `_index.scss` + `_buttons`, `_inputs`, `_selects`, `_toggles`, `_steps`, `_tips`, `_groups` + `ColorField.scss` y `Confirm.scss`), consumido con `@use '../../Common/Styles/Components' as *;`. El `index.css` global se movió a `Common/Styles/base/` (`_fonts`, `_reset`, `_variables`, `_index`), importado desde `Main.ts`.
- `useModalEscape` se renombró a `useOverlayEscape` (archivo, export y todos los call sites; conserva el parámetro `depth`).
- `vite.config.ts`: `manualChunks` reescrito a chunks por dominio.
- `AGENTS.md`: actualizada la regla de estilos (ahora residen en `Styles/` del dominio o en `Common/Styles/`).

### 2. Sistema global de overlays: Host + Confirm + ask() (F2)

- `Common/Overlays/Store.ts` se amplió con una **pila de diálogos** (`dialogs`, `openDialog(name, props?, listeners?)`, `closeDialog(id)`, `closeAllDialogs()`) y un **confirmador global por Promise** (`ask(options) → Promise<{confirmed, value}>`, `confirmState`, `resolveConfirm`).
- Nuevo `Common/Overlays/Host.vue`: único punto de acoplamiento cruzado; monta los diálogos registrados (`account-form`, `profile-form`, `font-manager`, `instances-form`, `instances-settings`, `instances-download`) pasándoles `:visible="true"` y traduciendo `update:visible` a `closeDialog`. Se monta desde `App.vue`.
- Nuevo `Common/Overlays/Confirm.vue` + `Common/Styles/Components/Confirm.scss` (estilo heredado de los overlays `InstOverlay_*`): confirmación/aviso con icono, mensaje, botón de peligro opcional y campo de texto opcional (para clonar). Cierra todo al recibir `CLOSE_OVERLAYS_EVENT`.
- **Se eliminaron todos los anidamientos de modales**: `Accounts/Content.vue` (formulario de cuenta + confirmación de borrado), `Versions/Content.vue` (formulario de perfil + confirmación de borrado), `Settings/Sections/Personalization.vue` (FontManager), `Instances/Instances.vue` (formulario, ajustes, descarga y los overlays inline de **clonar/eliminar**). Los diálogos ya no se importan dentro de otros overlays: se abren vía store.
- Los 6 diálogos migrados ahora inicializan su estado con `watch(visible, …, { immediate: true })`, ya que se montan frescos desde el Host con `visible = true`.
- Limpieza: eliminados los estilos muertos `InstOverlay_*` de `Instances/Styles/Instances.scss` (el bloque del verificador ya no se usaba) y el listener `CLOSE_OVERLAYS_EVENT` de `Personalization.vue` (ahora lo gestiona el Host).

### 3. Paneles reales y visibilidad centralizada (F3)

- `Instances/Instances.vue` y `Screenshots/Screenshots.vue` dejaron de ser overlays con fondo oscuro: son **paneles a pantalla completa** con `v-show` desde `App.vue`, controlados por el estado exclusivo `heavyPanel` (`'instances' | 'shots' | null`). `App.vue` ya no usa `showInstances`/`showShots`.
- Las capturas se abren desde el panel de instancias (con `shotsInstance` + `shotsReturn` en el store: al cerrar capturas se vuelve al detalle de la instancia) o desde el menú (Fotos → panel global).
- Los 9 modales de nivel superior de `App.vue` pasaron a refs del store: `settingsOpen`, `accountsOpen`, `loginOpen`, `installOpen`, `versionsOpen`, `crashOpen`, `newsOpen`, `welcomeOpen`, `previewOpen` (se eliminan los `showXxx` locales). `mainMenuHidden` ahora deriva de `heavyPanel` + esas refs.
- Los overlays raíz (Versions, Accounts/Manager, News, Welcome, Login, Crash, Settings, Preview) mantienen su estilo modal con `v-model:visible`; los paneles usan transiciones propias (nombres `InstancesModal`/`ScreenshotsModal`) con `v-show` para conservar el estado del detalle al volver de capturas.

## Por qué

El usuario pidió una restructuración completa del frontend para eliminar la dependencia del árbol `Modals/` (que obligaba a importar modales dentro de otros modales), facilitar la navegación y preparar el terreno para la división de monolitos y la extracción de `App.vue` en una sesión posterior.

## API afectada

- Sin cambios en bindings de Wails ni en el backend (ningún archivo Go tocado).
- Frontend: nueva API interna `Common/Overlays/Store.ts` (`openDialog`, `closeDialog`, `closeAllDialogs`, `ask`, `resolveConfirm`, `heavyPanel`, `shotsInstance`, `shotsReturn`, `settingsOpen`, `accountsOpen`, `loginOpen`, `installOpen`, `versionsOpen`, `crashOpen`, `newsOpen`, `welcomeOpen`, `previewOpen`). `useModalEscape` → `useOverlayEscape`. `manualChunks` de Vite por dominio.

## Comportamiento anterior/nuevo

- **Antes**: carpetas genéricas (`Modals/`, `Stores/`, `Styles/`) con nombres tipo `XxxModal.vue`; los diálogos pequeños se importaban directamente dentro de otros overlays (anidamiento); Instancias y Capturas eran overlays con fondo oscurecido; `App.vue` tenía 12 refs `showXxx` locales.
- **Ahora**: cada dominio es autocontenido; los diálogos se abren por nombre a través del `Host`; Instancias/Capturas son paneles reales a pantalla completa; la visibilidad de todos los modales vive en el store de overlays. Flujo de clonar/eliminar ahora pasa por el Confirm global (con campo de texto para el nombre de la copia).

## Pendiente (sesiones posteriores)

- F4: división de monolitos (`Downloads/Installation.vue` 1267 líneas, `Instances/Download.vue` 1262, `Welcome/Welcome.vue` 647, `Settings/Sections/Personalization.vue` 673, `News/News.vue` 563) — pospuesto por decisión del usuario (checkpoint).
- F5: extraer `App.vue` en composables (`useBootstrap`, `useBackground`, `useSplash`, `useZoom`, `useUserMenu`) + `Launcher/Home.vue`.

## Cómo verificar

- `bun run build` en `frontend/`: OK (type-check + vite build).
- `go build ./...` en la raíz: OK.
- Manual (`wails dev`): abrir/cerrar Instancias y Capturas (menú y desde el detalle de instancia; volver de capturas debe restaurar el detalle); clonar una instancia (Confirm con campo "Nombre de la copia"); eliminar instancia/cuenta/perfil (Confirm con botón de peligro); crear/editar instancia, cuenta y perfil (diálogos vía Host); gestionar tipografías desde Personalización; ESC cierra en el orden correcto (diálogo → panel); ESC con la app en reposo cierra todo; inactividad cierra todos los overlays.