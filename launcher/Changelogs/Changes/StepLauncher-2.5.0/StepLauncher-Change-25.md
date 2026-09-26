# Cambios de StepLauncher 2.5.0 (Step-25) — Actualizador directo a GitHub: installer en Windows e instalación manual en Linux/macOS

- **Fecha**:   2026-09-26
- **Versión**: 2.5.0
- **Estado**:  implementado y verificado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

### 1. Adiós al Worker de Cloudflare: GitHub directo

- El provider de Wails (`internal/Updater/worker_provider.go`) y el check legacy (`internal/Handlers/Engine/Update.go`) consultaban `https://steplauncher.stepnicka012.workers.dev/updates/steplauncher/*`. Ahora ambos van directo a `https://api.github.com/repos/NovaStepStudio/StepLauncher/releases`, que devuelve el mismo JSON sin intermediarios.
- Wails usa el provider oficial `providers/github` de Wails v3 (beta.25): semver, canal estable (`/releases/latest`, sin prereleases ni drafts) y verificación con el sidecar `CHECKSUMS-SHA256.txt` que ya publica el workflow de release.
- El transporte sigue siendo el directo del launcher (sin proxy del sistema ni HTTP/2): el proxy de Ajustes > Red es SOLO para Minecraft (Bug-4 de la 2.5.0).
- El check legacy filtra prereleases para seguir el canal estable, igual que antes.

### 2. Windows: descarga el -installer.exe, lo ejecuta y cierra el launcher

- El match del asset cambió de `StepLauncher-Updater.exe` (que ya no se publica) a `steplauncher-v*-windows-<arq>-installer.exe`, con alias de arquitectura (`amd64/x86_64/x64`, `arm64/aarch64`) en backend (`isWindowsInstallerAsset`) y provider (`IsWindowsInstaller`).
- `DownloadUpdater` guarda el instalador con su nombre real de la release en `%TEMP%/StepLauncher-Updater/`; `App.ApplyUpdate` lo lanza desacoplado y cierra el launcher para que el NSIS instale limpio (flujo ya existente, ahora con el asset correcto).
- `Store.installUpdate` prioriza este flujo legacy cuando hay actualización: el flujo Wails nunca instala en Windows (pondría el instalador como binario con el swap en caliente).

### 3. Linux/macOS: diálogo de instalación manual

- No hay instalador ni actualizador en estas plataformas: `HasUpdater` siempre es falso y `ApplyUpdate` abre la release en el navegador, como antes.
- El diálogo ahora lo dice claro: "Hay una nueva actualización disponible y en tu sistema debes instalarla manualmente" (.deb, .rpm, .AppImage, .dmg o .app); el botón en ese caso es "Abrir GitHub".

## Por qué

- A pedido del owner: sin dependencias externas para actualizar y con el instalador NSIS como vía oficial en Windows, que es lo que publican los workflows desde la mejora de releases.

## API afectada

- Sin cambios de bindings: `CheckForUpdates`, `ApplyUpdate`, `WailsCheckForUpdate` mantienen firma. `UpdateInfo` sin campos nuevos (`updaterUrl` ahora apunta al `-installer.exe`).

## Comportamiento anterior/nuevo

- Antes: checks contra el Worker; en Windows se buscaba un `StepLauncher-Updater.exe` inexistente en las releases nuevas (caía al navegador); en Linux/macOS el diálogo no aclaraba que la instalación es manual.
- Ahora: checks contra GitHub; en Windows se descarga/ejecuta el instalador y se cierra el launcher; en Linux/macOS el diálogo pide instalación manual.

## Cómo verificar

- `go build ./...` en `launcher/` — pasa.
- `bun run build` en `launcher/frontend` — limpio.
- Manual: con una release publicada, "Buscar actualización" en Windows → botón "Actualizar ahora" descarga el `-installer.exe` y cierra el launcher; en Linux/macOS → botón "Abrir GitHub" con el texto de instalación manual.
