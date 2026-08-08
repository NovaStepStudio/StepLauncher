# StepLauncher-Change-6: Nueva navegación (TopOptions + Sidebar)

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

**TopOptions (parte nueva):** la barra superior derecha (`App.vue`) ahora
contiene 4 botones de icono con tooltip al pasar el cursor:

- Notificaciones (`IconBell`)
- Noticias (`IconNews`)
- Descargas (`IconDownload`)
- Configuracion (`IconSettings`, abre el modal de ajustes)

**Sidebar:** se reemplazan los botones anteriores ("Configuracion" y
"Descargar una version") por 5 botones sin funcionalidad por el momento:

- Inicio (`IconHome`)
- Fotos (`IconPhoto`)
- Instancias (`IconBox`)
- Mods (`IconPuzzle`)
- Skins (`IconShirt`)

Todos los iconos provienen de `@tabler/icons-vue`.

**Frontend (Vue 3 + TS):**
- `App.vue`: importa `IconNews`, `IconPhoto`, `IconPuzzle`, `IconBox` e
  `IconShirt`; nuevo markup de `TopOptions .Others` (4 `.OptionOther`) y de
  `.Sidebar` (5 `.Item`).
- `Styles/App.scss`: `.Others` ahora usa `gap: .5rem`; `.OptionOther` pasa a
  ser un botón con `cursor: pointer`, `position: relative`, hover y tooltip
  `.OptionLabel` (mismo estilo que los labels de la sidebar) que aparece bajo
  el icono al pasar el cursor.

## Por qué

- Separar la navegación principal (Inicio, Fotos, Instancias, Mods, Skins) en
  la sidebar de las acciones de cabecera (Notificaciones, Noticias, Descargas,
  Configuracion) en la barra superior.
- El botón "Configuracion" pasa a la TopOptions; los botones de la sidebar
  quedan como puntos de entrada para futuras secciones (sin utilidad aún).

## API afectada

- Ninguna (solo frontend, sin bindings nuevos).

## Cómo verificar

- `bun run build` (dentro de `frontend/`) → OK (type-check incluido).
- La sidebar muestra los 5 iconos nuevos con tooltip al hacer hover.
- La TopOptions muestra los 4 botones; "Configuracion" abre el modal de
  ajustes; los demás no hacen nada por ahora.
