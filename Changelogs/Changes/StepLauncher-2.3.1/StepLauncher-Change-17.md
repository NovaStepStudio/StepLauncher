# Changes/StepLauncher-2.3.1/StepLauncher-Change-17.md

- **Fecha**: 2026-08-12
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado (build verificado; verificación en runtime pendiente).

## Qué cambió

La búsqueda de versiones de modloaders ahora devuelve la lista **siempre de mayor a menor** (orden numérico real, no lexicográfico), el algoritmo de recomendación elige siempre la versión **más reciente y estable**, y el `loader-state.json` deja de aparecer en cualquier carpeta según el path de instalación: las instancias lo guardan sí o sí dentro de `.StepLauncher` (carpeta fija `loader-states/`) y el flujo clásico lo **elimina al terminar la instalación**.

### 1. Orden numérico descendente en `GetModLoaderVersions` (todos los providers)

- `internal/Core/ModLoader/Version.go` (nuevo): `CompareLoaderVersions(a, b)` compara versiones segmento a segmento (separa por `.` y `-`; compara numéricos por valor y `alpha`/`beta`/... como texto) y `SortLoaderVersionsDesc` ordena de mayor a menor. Antes el orden dependía de cada provider: Forge/NeoForge ordenaban **lexicográficamente** (`"21.1.99" > "21.1.248"`, `"52.1.9" > "52.1.16"` → colocaba versiones antiguas arriba) y Fabric/Quilt/LegacyFabric no ordenaban nada (la API de Quilt devuelve p. ej. `0.20.0-beta.9, beta.7, beta.8` desordenadas).
- `internal/Core/ModLoader/Orchestrator.go`: `GetVersions` aplica `SortLoaderVersionsDesc` de forma centralizada, como **única garantía de orden** para la UI y el algoritmo de recomendación.
- `internal/Core/ModLoader/Provider/Forge.go` y `Provider/Neoforge.go`: se eliminan los `sort.Slice` lexicográficos (y sus imports) que quedaban en los providers.

### 2. Recomendación siempre la más reciente y estable

- `internal/Core/ModLoader/Orchestrator.go`: `ResolveVersion` con estrategia `recommended` devuelve la **primera estable** de la lista (con el orden descendente corregido es ahora la más reciente estable); si ninguna es estable, cae a la más reciente. Antes, con el orden roto, podía recomendar versiones viejas (`21.1.99` en vez de `21.1.248`).
- Frontend (`InstallationModal.vue` / `InstanceDownloadModal.vue`, `recommendFrom`): sin cambios de código — el primer "stable" de una lista ya ordenada desc por el backend es la versión correcta.

### 3. `loader-state.json` en una ubicación fija dentro de `.StepLauncher`

- `internal/Core/ModLoader/Orchestrator.go`: nuevo campo `stateDir` = `<workDir>/loader-states` y método `statePath(instancePath)` que deriva la clave por **hash del path absoluto** del destino de instalación (`loader-state-<hash16>.json`). `saveState` crea la carpeta, escribe ahí y migra (borra best-effort) el `loader-state.json` legacy que quedaba en `<instancePath>/`; `LoadState` y `RemoveState` usan la misma ruta canónica y `RemoveState` también limpia el legacy.
- `internal/Core/ModLoader/Types.go`: se elimina `LoaderStatePath` (calculaba `<instancePath>/loader-state.json`, la fuente del "aparece en todas partes al cambiar el path").
- Efecto: aunque la instancia o el directorio de juego cambie de ubicación, el estado del loader vive siempre en `.StepLauncher/loader-states/` y nunca en la carpeta de la instancia/juego.

### 4. Eliminación de `loader-state.json` al terminar la instalación clásica

- `internal/Handlers/Engine/Modloader.go`: `InstallModLoader` (flujo clásico sobre `WorkDir`) ahora, tras instalar correctamente, llama a `RemoveState(targetDir)`. El estado clásico no lo lee nadie (la versión queda en `versions/` y el lanzamiento la usa directamente), así que se elimina al terminar la instalación; en el flujo de instancias el estado se conserva (lo leen `Instance/Launch.go` y el store `Instances.ts`), pero en su ubicación fija.

## API afectada

- Bindings: ninguno cambia de firma (`GetModLoaderVersions`, `InstallModLoader`, `GetInstalledModLoader`, `RemoveModLoaderState`, instancias incluidas).
- Interno: `LoaderStatePath` (exportada, uso solo interno) eliminada; `LoadState`/`RemoveState`/`saveState` siguen recibiendo `instancePath` pero escriben/leen en `loader-states/`.

## Comportamiento anterior/nuevo

- **Listas de versiones**: anterior: Forge/NeoForge con orden incorrecto (viejas arriba) y Quilt desordenada. nuevo: siempre de mayor a menor numérica (Fabric: `0.19.3 > 0.19.2 > ...`; NeoForge: `21.1.248 > 21.1.99`; Forge: `52.1.16 > 52.1.15`; Quilt: `0.20.0-beta.9 > beta.8 > beta.7`).
- **Recomendación**: anterior: podía recomendar una versión vieja; nuevo: siempre la más reciente estable.
- **loader-state.json**: anterior: en `<destino>/loader-state.json` (se movía con el path y dejaba archivos en raíz del WorkDir tras instalar clásico); nuevo: solo en `.StepLauncher/loader-states/` y eliminado al terminar la instalación clásica.

## Cómo verificar

- `go build ./...` OK.
- `wails dev`: abrir el instalador, elegir una MC y un loader → la lista de versiones debe verse de mayor a menor y la recomendada debe ser la más reciente estable; tras instalar un loader en el flujo clásico no debe quedar `loader-state.json` en la raíz de `.StepLauncher`; en instancias queda dentro de `.StepLauncher/loader-states/`.