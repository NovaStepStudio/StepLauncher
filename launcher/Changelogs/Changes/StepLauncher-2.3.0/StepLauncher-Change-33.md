# StepLauncher-Change-33.md

## Fecha
2026-08-06

## Release
StepLauncher-2.3.0 — se mencionó por primera vez en esta release.

## Cambio
El launcher recuerda entre reinicios la última versión/perfil elegido, y los modos "versión" y "perfil" se vuelven excluyentes para que el lanzamiento quede siempre determinado.

## Qué pasaba
Al seleccionar una versión en el panel, el launcher no guardaba esa elección en disco: al reabrir la app volvía a la versión más reciente instalada, perdiendo la última selección. Además, podía darse el caso de que el menú mostrara una versión y el lanzamiento usara la del perfil activo (incoherencia visible en el selector).

## Qué se hizo
1. **Persistencia en backend** (`internal/Core/Launcher/Profile/Profile.go`):
   - Nuevos métodos `Manager.SelectedVersion()` y `Manager.SetSelectedVersion(version)` (protegidos por `m.mu`, con mutar bajo `Lock()` y persistir vía `save()`).
   - `save()` ya escribe `raw["selectedVersion"]` y `Load()` restaura `ProfilesFile.SelectedVersion` desde `launcher_profiles.json`.
2. **Nuevos bindings** (`internal/Handlers/Engine/Profiles.go` y `app.go`):
   - `GetSelectedVersion() string` y `SetSelectedVersion(version string) error` en el `Engine` (delegando al Manager) y en `App` para Wails.
   - Se regeneraron los bindings (`frontend/wailsjs/go/main/App.js` y `App.d.ts`) con ambos métodos.
3. **Store del frontend** (`frontend/web/src/stores/launcher.ts`):
   - `loadVersions()` restaura la última versión guardada si sigue instalada; si no, `ensureVersionSelected()` elige la más reciente (comportamiento anterior).
   - `selectVersion(id)` persiste la elección con `persistSelectedVersion(id)` y desactiva el perfil activo (modos excluyentes).
   - `setSelectedProfile(name)` refleja y persiste la versión del perfil si este la define (`syncProfileVersion()`), de modo que `loadProfiles()` y la selección de perfil siempre muestran la versión que se lanzará.
   - Nuevo computed `canLaunch`: habilita Jugar si hay versión de menú o si el perfil activo fija su versión.
4. **UI** (`frontend/web/src/App.vue` y `frontend/web/src/Styles/App.scss`):
   - El botón Jugar ahora usa `canLaunch` en vez de `selectedVersion` pelado.
   - El selector muestra "Perfil • <nombre>" cuando hay perfil activo y la versión efectiva (la del perfil si la fija).
   - `.InfoVersion` (p y h5): `white-space: nowrap` + `overflow: hidden` + `text-overflow: ellipsis` para que versiones largas como `1.12.2-forge-14.23.5.2859` no rompan ni se corten en dos líneas.

## Verificación
- `go build ./...` OK.
- Regeneración de bindings Wails: `GetSelectedVersion`/`SetSelectedVersion` presentes en `frontend/wailsjs/go/main/App.js` y `App.d.ts`.
- `bun run build` en `frontend/` OK (incluye `vue-tsc --build` + `vite build`).
- `go fmt ./...` OK.

## Notas
- La selección se guarda en `launcher_profiles.json` junto a `selectedProfile` (se reutiliza el mismo archivo; NO se usa SQLite).
- Al elegir una versión se limpia el perfil activo; al elegir un perfil se muestra su versión. Ambos estados se recuerdan al reabrir la app.
- `persistSelectedVersion`, `dismissPersistedProfile` y la restauración son `try/catch` aislados: un fallo de red/backend nunca rompe la selección en la UI.