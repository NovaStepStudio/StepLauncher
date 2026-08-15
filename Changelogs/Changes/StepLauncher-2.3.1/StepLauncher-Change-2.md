# Changes/StepLauncher-2.3.1/StepLauncher-Change-2.md

- **Fecha**: 2026-08-09
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.

## Qué cambió

Separación de todas las hojas de estilo SCSS de los componentes Vue: los bloques `<style scoped lang="scss">` de los 22 `.vue` dejan de llevar CSS inline y pasan a archivos `.scss` independientes organizados en subcarpetas dentro de `frontend/web/src/Styles/`. Es una refactorización mecánica 1:1: el CSS emitido es idéntico al anterior (mismos selectores, mismo scoping).

### 1. Nueva estructura de `frontend/web/src/Styles/`

```
Styles/
├── App/        → App.scss            (movido y fusionado con el CSS inline que quedaba en App.vue)
├── Settings/   → AboutSettings, AccountsSettings, ColorField, GeneralSettings, MinecraftSettings, PersonalizationSettings .scss
├── Modals/     → AccountFormModal, AccountsModal, AccountsView, CrashModal, FontManagerModal, InstallationModal (movido), LoginProgressModal, NewsModal, ProfileFormModal, ScreenshotsModal, SettingsModal, UpdateModal, VersionsModal, VersionsView .scss
├── Widgets/    → DownloadWidget.scss
└── Shared/     → Settings.scss        (hoja compartida de Ajustes, movida desde la raíz)
```

- Cada `.scss` se llama igual que su componente (`NewsModal.vue` → `Styles/Modals/NewsModal.scss`).
- Los 3 SCSS que ya existían en la raíz de `Styles/` (`App.scss`, `InstallationModal.scss`, `Settings.scss`) se movieron a su subcarpeta; `Settings.scss` pasa a `Shared/` porque lo comparten ~16 componentes.
- Los `.vue` de Ajustes quedan con `@use '../../../Styles/Settings/X.scss';`, los modales con `@use '../Styles/Modals/X.scss';`, el widget con `@use '../../Styles/Widgets/DownloadWidget.scss';` (la ruta `../../` es porque vive en `src/Widgets/DownloadWidget/`) y `App.vue` con `@use 'Styles/App/App.scss';`.
- Los archivos extraídos que usaban la hoja compartida la referencian al inicio con `@use '../Shared/Settings.scss';` (conservando `as *` donde ya existía); Sass exige `@use` como primera sentencia.
- El bloque `<style>` de cada `.vue` queda reducido a 3 líneas (`<style scoped lang="scss">` + `@use ...` + `</style>`). El atributo `scoped` se conserva: el `@use` dentro del bloque importa el contenido y el compilador de Vue aplica el scope a todo lo importado, igual que ya hacían `App.scss`/`Settings.scss` antes.

### 2. Ganancias de tamaño

- `NewsModal.vue`: 1690 → 564 líneas; `ColorField.vue`: 1120 → 675; `InstallationModal.vue` ya estaba extraído (1212 líneas); `AccountsView.vue`: 697 → 267; `ScreenshotsModal.vue`: 786 → 418; `App.vue`: 949 → 648. En total se movieron ~5.300 líneas de CSS a archivos dedicados.

## Por qué

Organización: cada componente tiene su propia hoja de estilos localizable al instante, la carpeta `Styles/` pasa a ser la única fuente de estilos de componentes, y se elimina el problema de archivos `.vue` gigantes de más de 1k líneas.

## API afectada

Ninguna. No cambia bindings, stores, config ni el CSS resultante; solo la ubicación física de los estilos. Si algún día se toca el CSS, el archivo a editar es el `.scss` del componente (no el `.vue`).

## Comportamiento anterior/nuevo

- Anterior: CSS inline en los 22 componentes (bloques `<style>` de 3 a 1.129 líneas).
- Nuevo: cada componente importa su `.scss` dedicado; visualmente idéntico.

## Cómo verificar

- `bun run build` en `frontend/` (type-check + build): pasa.
- El CSS de dist mantiene los selectores esperados por chunk (`News_Overlay`, `UserCardWrap`, `SsGroupHead`, `InstallationModal_Overlay`...).
- Comprobación visual con `wails dev` (menú principal, modales y Ajustes).
