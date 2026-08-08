# StepLauncher-Error-18: al instalar un modloader se creaba una carpeta "shared" con librerías

- **Fecha**: 2026-08-07
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (en desarrollo)

---

## 1. Síntoma

Al instalar un modloader (Forge o NeoForge) en el launcher se creaba dentro del espacio de trabajo una carpeta `shared` que no debería existir: el modloader debe instalarse a nivel de versión ({WorkDir}/versions y {WorkDir}/libraries), nunca en una carpeta compartida.

## 2. Causa raíz

El `Orchestrator` de `internal/Core/ModLoader` tenía un campo `sharedDir` (`filepath.Join(workDir, "shared")`) que se propagaba como librerías de la instancia: `NewOrchestrator(workDir, sharedDir, ...)` recibía una ruta `.../shared/libraries` como `librariesPath`. Esa ruta compartida sobraba: las librerías del modloader se resuelven dentro de `{WorkDir}/libraries`, así que la carpeta `shared` terminaba creándose por los `os.MkdirAll` de la descarga de librerías.

## 3. Diagnóstico y evidencia

- `internal/Core/ModLoader/Orchestrator.go`: el campo `sharedDir` era un parámetro del `Orchestrator` (`NewOrchestrator(workDir, sharedDir, cacheDir, ...)`) y `LibrariesPath` devolvía `filepath.Join(o.sharedDir, "libraries")` cuando `sharedDir` no estaba vacío.
- `internal/Handlers/Engine/Engine.go`: la llamada pasaba `filepath.Join(cfg.WorkDir, cfg.SharedDir)` como `sharedDir` (la carpeta `{WorkDir}/shared`), así que la descarga de librerías del instalador y de los profile libraries creaba `{WorkDir}/shared/libraries` en disco.

## 4. Solución aplicada

- `internal/Core/ModLoader/Orchestrator.go`: eliminado el campo `sharedDir` y la construcción de cualquier ruta `shared`.
- Firmas actualizadas a `NewOrchestrator(workDir, cacheDir, client, reg, broadcast, logFn)` (se quita `sharedDir`; `LibrariesPath` ahora devuelve siempre `filepath.Join(instancePath, "libraries")` = `{WorkDir}/libraries`, y el cache de modloaders se mantiene en `{WorkDir}/cache/modloader`).
- Actualizada la llamada en `internal/Handlers/Engine/Engine.go` (se pasa `cfg.CacheDir` en lugar de la ruta compartida).
- Actualizado el comentario de cabecera de `Orchestrator.go` y `Modloader.go` para reflejar que NO se usa carpeta `shared`.

## 5. Verificación

- `go build ./...` OK en Windows.
- El flujo de instalación de un modloader ya solo toca `{WorkDir}/versions`, `{WorkDir}/libraries` y `{WorkDir}/cache/modloader`.