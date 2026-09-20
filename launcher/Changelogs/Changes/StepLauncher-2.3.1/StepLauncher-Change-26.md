# Changes/StepLauncher-2.3.1/StepLauncher-Change-26.md

- **Fecha**: 2026-08-13
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado (build OK).

## Qué cambió

### 1. Vista previa completa del launcher en la configuración

`frontend/web/src/Modals/PersonalizationPreviewModal.vue` (nuevo) + `frontend/web/src/Styles/Modals/PersonalizationPreviewModal.scss` (nuevo):

- Vista previa a **pantalla completa** (`.PvRoot`, `position: fixed; inset: 0; z-index: 250`, sin scroll) que muestra el launcher **con los colores que el usuario ya configuró** (NO temas/paletas predefinidas). Se construye exclusivamente con variables CSS (`var(--background-sidebar)`, `var(--background-button-primary)`, `var(--background-play-button)`, `var(--background-modal-primary)`, `var(--border-style)`, `var(--progress-color)`, `var(--color-tag)`, `var(--color-success)`, `var(--color-warning)`, `var(--color-error)`, `var(--text-primary)`, `var(--text-secondary)`, `var(--font-primary)`, `var(--font-secundary)`), por lo que se repinta **en vivo** mientras el usuario mueve los selectores de color detrás.
- El mockup central (`.PvMain`, `height: min(58vh, 34rem)`, `aspect-ratio: 16/9`) replica la interfaz real importando `Styles/App/App.scss` con las mismas clases e imágenes del launcher (`Sidebar`, `Item`, `Content`, `BottomControlVersion`, `VersionSelected`, `ImageVersion`, `InfoVersion`, `PlayBlock`, `PlayButton`, `Decoration`, `TopOptions`, `Others`, `OptionOther`, `OptionLabel`, `UserCardWrap`, `UserCard`, `Avatar`, `Username`, `ExpandButtonProfiles`; `not_found_version.png`, `avatar_not_found.png`, `chicken.png`, `steve_and_alex.png`).
- **Barra lateral de fondos** (`.PvSide`): items cliqueables "Sin fondo", "Imagen", "Video animado" y "Dinámico", cada uno con thumbnail (fondo real del usuario vía `loadLocal` o ejemplo `bg-welcome*.webp`), badge "tu fondo"/"ejemplo" y check sobre el item activo (`.on`).
- Todos los fondos se muestran como **imagen estática** (incluso video/dinámico: solo se usa la primera imagen o el video como imagen) y sin pixelado (`image-rendering: auto` en `.PvMain .BackgroundLayer img`).
- Fila compacta de ejemplos de colores (`.PvExamples`) bajo el mockup: badge (`--color-tag`), barra de progreso (`--progress-color`), mensajes de éxito/aviso/error (`--color-success`/`--color-warning`/`--color-error`) y botones primario/ghost/peligro.
- Cierre por botón X y evento `CLOSE_OVERLAYS_EVENT` (inactividad).

### 2. Apertura: cierra Configuración y oculta el menú principal

- `frontend/web/src/Stores/Modals.ts`: nuevo evento `PERSONALIZATION_PREVIEW_EVENT = 'personalization-preview-open'`.
- `frontend/web/src/App.vue`: `showPreview` añadido a `mainMenuHidden` (el menú principal se oculta igual que con Instancias/Versiones) y a `handleIdle()`; el listener del evento ejecuta `openPersonalizationPreview()`, que cierra el modal de configuración (`showSettings = false`) y abre la vista.
- `frontend/web/src/Layouts/Sections/Settings/PersonalizationSettings.vue` + `frontend/web/src/Styles/Settings/PersonalizationSettings.scss`: botón `SsBtnPrimary` "Abrir vista previa" al final del grupo Colores (fila `.PsPreviewRow`) que despacha el evento (ya no monta el modal directamente).

## Por qué

El usuario pidió llevar la experiencia de preview de la bienvenida a la configuración: un botón que abra un panel completo mostrando cómo se verá el launcher con los colores personalizados que él mismo eligió (no paletas/temas), con el mismo estilo, layout, fondo e imágenes del launcher real, fondo intercambiable y ejemplos visibles.

## API afectada

- Frontend: componente nuevo, sin bindings nuevos. Reutiliza `applyPersonalization`/`personalization`/`loadLocal` de `Stores/Ui.ts` y `CLOSE_OVERLAYS_EVENT` de `Stores/Idle.ts`. Nuevo evento de window `PERSONALIZATION_PREVIEW_EVENT` en `Stores/Modals.ts`.
- Assets: reutiliza `not_found_version.png`, `avatar_not_found.png`, `chicken.png`, `steve_and_alex.png` y `bg-welcome*.webp`.

## Comportamiento anterior/nuevo

- **Antes**: en la configuración solo existían los ejemplos semánticos pequeños (`PsExamples`); no había forma de ver el launcher completo con los colores configurados.
- **Ahora**: botón "Abrir vista previa" en el grupo Colores → cierra Configuración, oculta el menú principal y abre una vista a pantalla completa con el launcher mockup usando los colores actuales, fondo intercambiable por la barra lateral (tu fondo o ejemplos, siempre como imagen estática y sin pixelado) y fila de ejemplos de colores.

## Cómo verificar

- `bun run build` en `frontend/`: type-check y vite build OK.
- `go build ./...` en la raíz: OK.
- Manual: Configuración → Personalización → Colores → "Abrir vista previa": Configuración se cierra, el menú principal se oculta y se ve la vista completa; pulsando cada item de la barra lateral cambia el fondo del mockup; cambiando un color (p. ej. `--background-button-primary` o `--progress-color`) el mockup se repinta al instante; cierra con X o al dejar el launcher inactivo.