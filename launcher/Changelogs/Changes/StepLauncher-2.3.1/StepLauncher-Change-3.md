# Changes/StepLauncher-2.3.1/StepLauncher-Change-3.md

- **Fecha**: 2026-08-09
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.

## Qué cambió

El caché de descarga/verificación del sistema de instancias deja de vivir dentro del directorio compartido: ya no se crea `shared/cache`. El caché es **global y único** en `<WorkDir>/cache` (`%APPDATA%\.StepLauncher\cache`), igual que el del resto del launcher, y nunca se comparte ni se duplica por instancia.

### 1. `internal/Core/Launcher/Instance/Manager.go`

- Nuevo campo `cacheDir string` en `InstanceManager` + setter `SetCacheDir(dir)`. Documentado: el caché nunca se guarda bajo `shared/`.

### 2. `internal/Core/Launcher/Instance/Verify.go`

- `VerifyInstance` usaba `CacheDir: filepath.Join(m.sharedDir, "cache")` al construir el `downloader.Config` con el que verifica SHA-1 de librerías/assets → ahora usa `m.cacheDir` (el global). El fallback a `shared/cache` queda solo si el setter no se llamó (compatibilidad con usos aislados del paquete).

### 3. `internal/Handlers/Engine/Engine.go`

- El engine conecta el caché en la construcción: `instMgr.SetCacheDir(cfg.CacheDir)`.

## Por qué

El directorio `shared/` debe contener únicamente recursos que se comparten entre instancias para no volver a descargarlos (libraries, assets, runtime). El caché no es un recurso que se comparta: es un almacén temporal del launcher y debe vivir en la ubicación global `.StepLauncher/cache` junto al resto del caché (manifiestos, JSONs, backgrounds). Ya no se duplica en disco ni se mezcla con los recursos compartidos.

## API afectada

Ninguna binding cambia. Solo se añade el setter interno `InstanceManager.SetCacheDir` (Go, privado al engine): el frontend y los bindings de Wails quedan intactos.

## Comportamiento anterior/nuevo

- Anterior: la verificación de instancias descargaba/cacheaba JSONs en `shared/cache` (se creaba `%APPDATA%\.StepLauncher\shared\cache` en el primer uso).
- Nuevo: todo el caché (descargas normales, instancias, verificación, modloaders, backgrounds) vive en `%APPDATA%\.StepLauncher\cache`. La carpeta `shared/cache` existente en discos viejos se ignora y no se vuelve a crear (la app funciona con cualquier estado en disco).

## Cómo verificar

- `go build ./...` en la raíz: pasa.
- Crear una instancia, descargar una versión y ejecutar `VerifyInstance`: `%APPDATA%\.StepLauncher\shared\` no contiene ninguna carpeta `cache`.