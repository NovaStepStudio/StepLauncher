# StepLauncher-Change-22: Iconos de perfil (galería + importar imagen)

- **Fecha**: 2026-08-05
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

Los perfiles ya muestran un **icono/avatar** propio, configurable desde el
formulario de perfil, y se renderiza en la lista de perfiles del panel de
versiones.

- **Galería**: en `frontend/web/src/Modals/ProfileFormModal.vue`, nuevo bloque
  "Icono del perfil" arriba del formulario con 9 iconos de Minecraft como SVG
  inline embebidos con data-URI (`svgDataUri()` + `encodeURIComponent`): Steve,
  Creeper, Pico, Espada, Diamante, Manzana, Libro, Hierba y Baúl. Se selecciona
  con la celda en cuadrícula (`ProfileForm_IconCell.tile`), y la activa marca
  `ProfileForm_IconCell.active`.
- **Importar imagen propia**: botón de importación con
  `<input type="file" accept="image/png,image/jpeg,image/webp,image/gif">`,
  límite de **3 MB**, lectura como `FileReader` y **recorte a un canvas de
  96×96 PNG** (mantiene la proporción y rellena de fondo gris oscuro) vía
  `fileToDataUrl()`. Botón "Quitar" (`clearIcon`) para volver a "Sin icono".
- **Estado del perfil**: `form.icon` guarda el data URI resultante; un watcher
  lo carga desde el perfil existente (`e?.icon`), y el payload al crear/editar
  incluye el campo `icon`.
- **Renderizado**: en `frontend/web/src/Modals/VersionsView.vue`, cada ítem de
  la lista de perfiles muestra `<span class="Vers_ProfileAvatar">` con
  `<img :src="p.icon">` cuando hay icono o la inicial del nombre en caso
  contrario. Estilos `.Vers_ProfileAvatar` (2.2rem, `image-rendering: pixelated`).
- El campo `icon` ya existía en backend (`Profile.Icon`) y en el modelo
  `LauncherProfile` del store; no ha habido cambios en API.

## Por qué

Faltaba una forma visual de distinguir perfiles en la lista; los iconos de
Minecraft con opción de importar imagen propia ayudan a identificar cada perfil
de un vistazo.

## API afectada

- Ninguna (el campo `icon` ya viajaba en el modelo de perfil).

## Cómo verificar

- `bun run build` (dentro de `frontend/`) → OK (`vue-tsc --build` + `vite build`).
- Abrir el modal de crear/editar perfil: elegir un icono de la galería o
  importar una imagen (rechaza >3 MB y formatos no permitidos), guardar, y ver
  el avatar en la lista de perfiles del panel de versiones (imagen o inicial).