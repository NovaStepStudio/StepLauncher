# StepLauncher-Error-13: Un perfil importado con `lastVersionId` (p. ej. el instalador de Forge) no se podía ejecutar: lanzaba la versión base en vez de la del modloader

- **Fecha**: 2026-08-06
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

Tras instalar Forge 1.12.2 (instalador oficial, que escribe `launcher_profiles.json`
a su manera), el perfil `forge` no se comportaba como una versión normal:

- El perfil importado aparecía sin versión en la UI ("Cualquier versión") pese a
  que su versión `1.12.2-forge-14.23.5.2859` sí estaba descargada.
- Al pulsar Jugar con ese perfil, el log del juego
  (`logs/game/StepLauncher-1.12.2-2026-08-06.log`) mostraba `=== Launching
  Minecraft 1.12.2 ===` con `Main Class: net.minecraft.client.main.Main` y
  natives de la base: se lanzaba la vanilla, no Forge (se sospechaba
  erróneamente que el problema era la resolución de `inheritsFrom`).
- Adicionalmente, la pantalla completa (fullscreen) de un perfil nunca se
  aplicaba al juego.

## 2. Causa raíz

El formato oficial de `launcher_profiles.json` usa el campo **`lastVersionId`**
para la versión del perfil, pero el struct `Profile`
(`internal/Core/Launcher/Profile/Profile.go`) solo leía `json:"version"`.
`json.Unmarshal` descartaba silenciosamente `lastVersionId` → el perfil de Forge
quedaba con `Version == ""`.

Ese vacío se propagaba en el lanzamiento: `mergeProfileIntoConfig`
(`internal/Handlers/Engine/Launch.go`) solo sobreescribe la versión con
`if p.Version != ""`, así que se lanzaba el `selectedVersion` del menú
(la vanilla 1.12.2). La resolución de `inheritsFrom` en el motor era correcta:
el log no mentía, la versión enviada al motor ya era la base.

Además `mergeProfileIntoConfig` fusionaba GameDir, JavaExec, JavaArgs y
resolución, pero olvidaba `Fullscreen` → la ventana del perfil ignoraba la
opción.

## 3. Solución aplicada

### `internal/Core/Launcher/Profile/Profile.go`

- Nuevo campo `LastVersionID string json:"lastVersionId,omitempty"` en `Profile`.
- En `Load()` se normaliza por perfil: si `Version` está vacío y
  `LastVersionID` no, `Version = LastVersionID` (compatibilidad con perfiles
  creados por el launcher oficial y los instaladores de Forge/Fabric).
- En `save()` se mantiene `lastVersionId` sincronizado con `version` para que
  el fichero siga siendo compatible con la estructura estándar.

### `internal/Handlers/Engine/Launch.go`

- `mergeProfileIntoConfig` ahora aplica `adv.Fullscreen = p.Fullscreen` (el
  perfil tiene prioridad sobre la config del launcher, igual que el resto de
  campos).

### `frontend/web/src/stores/launcher.ts`

- `launchGame()` calcula la **versión efectiva**: si el perfil activo define
  `version`, se lanza esa (el perfil funciona como una versión normal con su
  propia configuración) y el mensaje de lanzamiento lo refleja. Si no, se usa
  la versión del menú.

## 4. Verificación

- Test local con el `launcher_profiles.json` real del usuario: el perfil `forge`
  ahora carga con `version="1.12.2-forge-14.23.5.2859"` (antes `""`), sin
  modificar el archivo original (se trabaja sobre una copia).
- `go build ./...` → OK. `bun run build` (incluye `vue-tsc`) → OK.
- Las librerías de Forge ya presentes (`libraries/net/minecraftforge/...`,
  `launchwrapper`) y el jar del cliente base (`versions/1.12.2/1.12.2.jar`)
  cubren el classpath: el lanzamiento debe arrancar con
  `net.minecraft.launchwrapper.Launch` y `--versionType Forge`.

## 5. Regla aprendida

Antes de culpar a la resolución de versiones (`inheritsFrom`), verificar el
valor exacto que llega a `LaunchConfig.Version` (el log lo imprime en
`=== Launching Minecraft %s ===`). El dato real de entrada de un perfil
importado viene del campo oficial `lastVersionId`, no de `version`: el struct
Go debe leer ambos y normalizar al cargar.
