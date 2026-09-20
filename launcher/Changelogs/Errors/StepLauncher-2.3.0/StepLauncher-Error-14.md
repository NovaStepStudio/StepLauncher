# StepLauncher-Error-14: Instalar un modloader creaba estructura de instancias y dejaba la versión fuera de `versions/` (no se veía ni se podía ejecutar)

- **Fecha**: 2026-08-06
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

Al instalar un modloader desde el instalador (pestaña de instalación de
versiones), se creaba toda la estructura de **instancias**:
`instances/<version>-<loader>/versions/...`, `libraries/`, etc. La versión del
modloader quedaba fuera del directorio `versions/` de la raíz, por lo que no
aparecía en el panel ("Versiones") ni se podía lanzar como una versión normal.

## 2. Causa raíz

`startLoaderInstall()` de `frontend/web/src/Modals/InstallationModal.vue`
construía a mano una ruta de instancia con
`const dir = '${root}/instances/${selectedVersion}-${selectedLoader}'` y la
pasaba como `instancePath` a `InstallModLoader`. El orquestador de modloaders
(`internal/Core/ModLoader`) usa ese `instancePath` como base para escribir
`{instancePath}/versions/<id>/<id>.json`, `{instancePath}/versions/<mc>/...` y
`loader-state.json`. Resultado: carpeta `instances/` completa en vez de una
instalación a nivel de versión como hace el instalador oficial de Forge.

El sistema de instancias es **opcional** en este launcher (ver AGENTS.md); el
flujo principal de perfil/versión no debe depender de él.

## 3. Solución aplicada

### `internal/Handlers/Engine/Modloader.go`

- `InstallModLoader` ignora el `instancePath` del binding (resto del flujo de
  instancias) y fija como destino `targetDir = cfg.WorkDir`. Los providers ya
  construyen sus rutas sobre esa base (`{WorkDir}/versions/<id>`), de modo que
  la versión del modloader queda en la carpeta estándar de versiones y las
  librerías en `shared/libraries`.

### `frontend/web/src/Modals/InstallationModal.vue`

- Se eliminó el cálculo `root/instances/...` y la importación de
  `LocalAssetsDir`: ahora se llamar `InstallModLoader(loader, version, mc, '')`.

## 4. Verificación

- `go build ./...` → OK. `bun run build` (incluye `vue-tsc`) → OK.
- Con el cambio, instalar Forge/Fabric desde el instalador escribe la versión
  en `versions/<id>/<id>.json` (visible en `ListDownloadedVersions` y el
  selector de versión de un perfil) sin tocar `instances/`.

## 5. Regla aprendida

El modelo de instancias no comparte destinatario con el modelo de versiones:
una tarea de instalación de mods debe volver a la estructura de versión
(WorkDir base + `versions/` + `shared/libraries`), y las rutas de instancia no
deben filtrarse desde la UI.