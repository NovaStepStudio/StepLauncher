# Changes/StepLauncher-2.3.1/StepLauncher-Change-35.md

- **Fecha**: 2026-08-15
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado.

## Qué cambió

Mejora del sistema de cache para que las carpetas se creen solo cuando realmente se usan, eliminando la creación masiva de carpetas vacías en el primer arranque.

### 1. Creación bajo demanda (`internal/Core/Cache/Cache.go`)

- Se eliminó `ensureDirs()` y su llamada dentro de `Set()`. Antes, cada vez que se escribía un JSON en el cache se creaban TODAS las categorías (incluidas las que nunca se usan). Ahora `Set()` crea únicamente la carpeta de la categoría concreta que se escribe (ya lo hacía vía `os.MkdirAll(filepath.Dir(path))`).
- `subdirs()` se redujo a las categorías reales que almacenan datos: `default`, `manifest`, `versions`, `assets`, `java`.
- Se añadió `obsoleteDirs()`: `forge`, `neoforge`, `fabric`, `quilt`, `legacyfabric`, `assets/indexes` y `assets/manifests`. Estas carpetas las creaban versiones antiguas del launcher pero nunca guardan nada. El `cleanup()` (que corre al arrancar y cada hora) las elimina del disco si quedaron vacías.

### 2. Assets del juego (`internal/Core/Downloader/Tasks.go`)

- Se eliminó el `os.MkdirAll(... assets/indexes)` incondicional de `BuildTasks()`. La carpeta `assets/indexes` se crea ahora automáticamente al descargar un asset index real (`DownloadFile` ya crea el directorio destino), es decir, solo cuando una versión declara assets.

## Por qué

El launcher creaba en el primer arranque (primer `Set` del cache) un montón de carpetas que no tienen sentido porque nunca se escribe nada en ellas: `fabric`, `forge`, `neoforge`, `quilt`, `legacyfabric`, `assets/indexes` y `assets/manifests`.

## API afectada

- Sin cambios en bindings de Wails ni en el frontend (solo backend, sistema de cache y descargas).

## Comportamiento anterior/nuevo

- **Antes**: al primer arranque se creaban 12 carpetas en `cache/` (todas las de `subdirs()`), la mayoría vacías para siempre.
- **Ahora**: solo se crea la carpeta de la categoría cuando se escribe algo real en ella; las carpetas obsoletas que quedaron de versiones antiguas se eliminan automáticamente si están vacías.

## Cómo verificar

- Borrar o limpiar el cache y arrancar el launcher: comprobar que ya no aparecen `cache/fabric`, `cache/forge`, `cache/neoforge`, `cache/quilt`, `cache/legacyfabric`, `cache/assets/indexes` ni `cache/assets/manifests`.
- Usar un modloader (p. ej. abrir el selector de versiones de Fabric) y comprobar que su JSON se guarda en `cache/default/` sin crear carpetas nuevas innecesarias.
- Descargar una versión con assets y comprobar que `assets/indexes/<id>.json` (del workdir del juego) se crea al descargarse.