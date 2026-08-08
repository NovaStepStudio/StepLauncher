# StepLauncher-Change-32: Formulario de perfil: sin iconos predefinidos y con defaults de ventana del launcher

- **Fecha**: 2026-08-06
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
---

## 1. Qué cambia

### `frontend/web/src/Modals/ProfileFormModal.vue`

- **Eliminada la galería de iconos predefinidos** (Steve, Creeper, Pico,
  Espada, Diamante, Manzana, Libro, Hierba, Baúl): se quitaron los SVG inline
  (`iconGallery`, `svgDataUri`, `pickIcon`), su cuadrícula del template y sus
  estilos. El icono del perfil ahora solo admite imagen propia (PNG/JPG/WEBP/GIF
  importada como data URI, el formato estándar de `launcher_profiles.json`).
- **Defaults de ventana**: al crear un perfil nuevo, resolución (ancho/alto) y
  pantalla completa se precargan con la configuración actual del launcher
  (`GetConfig()` → `minecraftConfig.windowWidth/Height/fullscreen`), de modo
  que el perfil nace listo para jugar con la ventana del launcher. Si la
  config no está disponible, se deja vacío (el perfil hereda al lanzar).
  Al editar un perfil existente se siguen mostrando sus propios valores.

## 2. Verificación

- `bun run build` (incluye `vue-tsc`) → OK.
- `go build ./...` → OK (no hay cambios de bindings: `lastVersionId` se
  normaliza en Go y el frontend sigue usando `version`).
