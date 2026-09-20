# StepLauncher-Change-30: Colores del botón de jugar y botones primarios configurables

- **Fecha**: 2026-08-06
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

Las variables CSS `--background-play-button` y `--background-button-primary`
dejaron de estar fijas a `#111` en `frontend/web/index.css` y pasaron a ser
**configurables y persistidas** en la personalización, editables desde
Ajustes → Colores.

### 1. Backend (`internal/Config/Config.go`)

- `ThemeColors` gana dos campos: `PlayButton` (`json:"playButton"`) y
  `ButtonPrimary` (`json:"buttonPrimary"`), con valor por defecto `#111`.
- `sanitize()` los valida con `sanitizeColor` (configs viejas en disco se
  auto-reparan con el default al cargar).
- `UpdatePersonalization` los incluye en el historial de colores recientes.

### 2. Frontend

- `stores/ui.ts`: `ThemeColors` gana `playButton`/`buttonPrimary`, con default
  `#111` en `normalizePersonalization`, y `applyPersonalization` los aplica
  con `applyRootVar('--background-play-button', ...)` y
  `applyRootVar('--background-button-primary', ...)`. Se eliminó el comentario
  obsoleto que aseguraba que eran alias de `--color-selected` (no lo eran).
- `PersonalizationSettings.vue`: dos nuevos selectores de color ("Botón de
  jugar" y "Botones principales") que cargan su valor desde config, se
  guardan en `UpdatePersonalization` y entran al historial de recientes.
- `frontend/wailsjs/go/models.ts`: registrados `playButton`/`buttonPrimary` en
  la clase `ThemeColors` (el próximo `wails build` los regenerará igual).

## Por qué

- Esos dos colores eran los únicos del frontend que no se podían personalizar:
  estaban hardcodeados a `#111` en `index.css` y `applyPersonalization` nunca
  los tocaba, así que sospechaban (con razón) que "no se guardaban": nunca
  hubo un campo en la config ni en el store para ellos.

## API afectada

- Binding `UpdatePersonalization` y modelo `Config.Personalization.Colors`:
  dos campos nuevos con default, retrocompatible (campos opcionales).

## Cómo verificar

- `go build ./...` → OK.
- `bun run build` (frontend, incluye type-check) → OK.
- Abrir Ajustes -> Colores, cambiar "Botón de jugar" y "Botones principales",
  cerrar y reabrir: se mantienen y se reflejan en `launcher_config.json`.