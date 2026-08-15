# Changes/StepLauncher-2.3.1/StepLauncher-Change-27.md

- **Fecha**: 2026-08-13
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado (build OK).

## Qué cambió

### 1. Directorio de trabajo configurable (Normal / Minecraft / Portable / Custom)

Nuevo `internal/Handlers/Engine/engineconfig/Directory.go`: sistema de **bootstrap de directorio** que persiste la preferencia FUERA del WorkDir (para poder leerla antes de resolver la ruta y que sobreviva a cambios/borrados):

- `Bootstrap{ Mode, CustomPath, Configured }` guardado en `%APPDATA%\StepLauncher\directory.json` (Windows), `~/Library/Application Support/StepLauncher/directory.json` (macOS) o `~/.config/StepLauncher/directory.json` (Linux).
- Modos: `normal` (`%APPDATA%\.StepLauncher`), `minecraft` (`%APPDATA%\.minecraft` / `~/Library/Application Support/minecraft` / `~/.minecraft`), `portable` (`.StepLauncher` junto al ejecutable) y `custom` (ruta elegida por el usuario).
- `DetectMinecraftDir()` detecta una instalación `.minecraft` existente en el SO actual.
- `engineconfig/Config.go` (Load/LoadFile) y `app.go` (`defaultConfigDir`, ahora unificado) resuelven el WorkDir a través del bootstrap.
- `internal/Handlers/Engine/Directory.go` (nuevo): `DirectoryInfo()` para la UI, `SetDirectory(mode, customPath)` que valida (bloquea con juegos/descargas activos), persiste el bootstrap y **copia `launcher_config.json` al destino solo si el destino no tiene uno** (nunca sobrescribe; el resto de datos no se migra, como pidió el usuario) y `copyLauncherConfigIfMissing`.
- Bindings nuevos en `app.go`: `GetDirectorySettings`, `SetDirectoryMode`, `PickDirectory` (`runtime.OpenDirectoryDialog`) y `RestartApp` (relanza el exe y `runtime.Quit`).

### 2. GameDir desactivable (carpeta de juego separada)

- Nueva opción `SeparateGameDir *bool` (nil = true) en `internal/Config.MinecraftConfig` (persistida en `launcher_config.json`), `engine.MinecraftConfig` y `engineconfig.Config`.
- `launcher.ManagerConfig.SeparateGameDir` + `LaunchManager.SetSeparateGameDir`: en `Manager.go` el default de `adv.GameDir` pasa a ser `<workDir>/game` (separado) o el propio `<workDir>` (desactivado).
- `InstanceManager.SetSeparateGameDir`: en `Instance/Launch.go` el default pasa a ser `<instancia>/game` o la propia carpeta de la instancia.
- **En modo Minecraft se fuerza a `false` siempre** (el gameDir es el propio `.minecraft`, igual que el launcher oficial): se aplica en `NewEngine` y en `UpdateMinecraftConfig` (`internal/Handlers/Engine/Java.go`).
- Bindings `GetSeparateGameDir` / `SetSeparateGameDir`; toggle en ajustes deshabilitado en modo Minecraft.

### 3. Lista de archivos protegidos (no sobrescribir)

Nuevo `internal/Core/Utils/Protected.go`: `IsProtectedFile` + `SafeWriteFile`. Lista de archivos que el launcher NUNCA sobrescribe si ya existen (p. ej. los del launcher oficial de Minecraft): `launcher_profiles.json`, `launcher_settings.json`, `launcher.properties`, `options.txt`, `servers.dat`, `usercache.json`, `usernamecache.json`, `splashes.txt`.

Aplicado en:
- `internal/Config/Config.go` `Save()` (escribe la config con `SafeWriteFile`).
- `internal/Core/Launcher/Launcher.go` `ensureLauncherProperties` (no pisa `launcher.properties` existente; log si se omitió).
- `Downloader/Manager.go`: el `version.json` global ya no se sobrescribe si existe (log "Version JSON already exists"). El resto de descargas ya saltaba archivos existentes vía `processTask`/`FileExists` (contados como `existing`).

### 4. Frontend: paso "Carpeta del launcher" en la bienvenida

`frontend/web/src/Modals/WelcomeModal.vue` + `Styles/Modals/WelcomeModal.scss`:

- Nuevo paso `directory` entre `customize` y `account` (solo cuando `configured === false`, además del firstLaunch): selector de los 4 modos (grid 2x2), aviso "Detectamos una instalación de Minecraft" con botón "Usar" (`.WelcomeModal_DirNotice`), campo de ruta personalizada con botón "Examinar" (`.WelcomeModal_DirPickRow`) e indicador de pasos dinámico (3 o 4).
- Al guardar con cambio de directorio → `SetDirectoryMode` + `RestartApp` (el launcher se reinicia; la bienvenida continúa tras el reinicio sin repetir el paso porque `configured` ya quedó en true).
- Iconos `IconFolder`, `IconSettings`, `IconUsb` añadidos a los imports.

### 5. Frontend: sección "Directorio del launcher" en ajustes

`frontend/web/src/Layouts/Sections/Settings/GeneralSettings.vue` + `Styles/Settings/GeneralSettings.scss`:

- Grupo nuevo con: selector de modo (grid `.SsDirMode`), ruta personalizada + "Examinar", aviso de instalación `.minecraft` detectada con botón "Usar .minecraft", carpeta actual (WorkDir), toggle "Carpeta de juego separada" (deshabilitado y forzado a off en modo Minecraft) y botón "Guardar y reiniciar" (con mensaje de estado `.SsDirMsg`).
- `loadDirectorySettings()` en `onActivated`.

## Por qué

El usuario pidió poder cambiar el directorio donde el launcher guarda todo, con 4 modos (incluido usar la carpeta `.minecraft` oficial sin sobrescribir sus archivos), detección de una instalación `.minecraft` existente con aviso al usuario, sin migrar datos al cambiar (solo `launcher_config.json`), y además que el gameDir pueda desactivarse para usar el workDir directamente (forzado en modo Minecraft).

## API afectada

- Bindings nuevos: `GetDirectorySettings`, `SetDirectoryMode(mode, customPath)`, `PickDirectory`, `RestartApp`, `GetSeparateGameDir`, `SetSeparateGameDir`.
- `engine.DirectoryInfo` (nuevo tipo expuesto); `engineconfig.Bootstrap`, `engineconfig.DirMode`, `engineconfig.LoadBootstrap/SaveBootstrap/ResolveWorkDir/DefaultMinecraftDir/DetectMinecraftDir/DefaultNormalDir/PortableDir`, `engineconfig.Config.SeparateGameDir` (nuevo).
- `Config.MinecraftConfig.SeparateGameDir` (nuevo campo opcional; nil = true, sin migración necesaria).
- `launcher.LaunchManager.SetSeparateGameDir`, `instance.InstanceManager.SetSeparateGameDir` (nuevos).
- `globalutils.SafeWriteFile`/`IsProtectedFile` (nuevos en `internal/Core/Utils`).

## Comportamiento anterior/nuevo

- **Antes**: el WorkDir era siempre `%APPDATA%\.StepLauncher` (hardcodeado en dos sitios); el gameDir siempre `<workDir>/game`; los archivos de juego/descarga se podían sobrescribir.
- **Ahora**: el WorkDir se resuelve según el modo elegido (bootstrap fuera del WorkDir); el gameDir puede ser el propio workDir (forzado en `.minecraft`); existe una lista de archivos que nunca se sobrescriben; al cambiar de carpeta solo se copia `launcher_config.json` (si el destino no tiene uno) y el launcher se reinicia solo.

## Cómo verificar

- `go build ./...` en la raíz: OK.
- `bun run build` en `frontend/`: type-check y vite build OK.
- Manual: primer arranque → bienvenida → Personalizar → paso "Carpeta del launcher" (solo la primera vez); con `.minecraft` presente se muestra el aviso. Ajustes → Directorio del launcher: cambiar de modo → "Guardar y reiniciar" → el launcher se reinicia y la bienvenida no repite el paso; con `.minecraft` el toggle de gameDir queda bloqueado y los saves/mods se escriben en la raíz de `.minecraft`.