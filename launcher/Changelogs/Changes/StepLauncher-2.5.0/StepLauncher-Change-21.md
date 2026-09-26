# Cambios de StepLauncher 2.5.0 (Step-21) — Descarga e instalación de contenido de Modrinth (mods, shaders, texturas y modpacks .mrpack)

- **Fecha**:   2026-09-25
- **Versión**: 2.5.0
- **Estado**:  implementado y verificado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Ronda 1 — Instalación base (diálogo + backend + modpacks con instancia nueva)

(Contenido original de la primera implementación; la ronda 2 está abajo.)

## Qué cambió

### 1. Backend: contenido de Modrinth instalable (nuevo dominio)

- **`internal/Core/Mods/` (nuevo paquete)**:
  - `Content.go`: mapeo `project_type` → subcarpeta del gameDir (`mod` → `mods/`, `resourcepack` → `resourcepacks/`, `shader` → `shaderpacks/`), `ResolveGameDir` (`<workDir>/game` con la opción separada activa o el propio `workDir` si está deshabilitada) y descarga atómica (temporal + rename) con omisión si el archivo ya existe con igual tamaño y hash.
  - `Hash.go`: verificación sha1/sha512 de lo descargado.
  - `Modpack.go`: manifiesto `modrinth.index.json` (archivos + `dependencies.minecraft` + loader `fabric-loader`/`quilt-loader`/`forge`/`neoforge`), extracción segura del `.mrpack` (zip con otra extensión, rechaza `..` y rutas absolutas) y copia de `overrides/` fusionando al destino. Verificado contra el ejemplo real `Fabulously.Optimized-v14.1.0` (50 mods + 2 resourcepacks en el manifiesto, `config/` y `resourcepacks/` en overrides).
  - `Events.go`: eventos `modcontent_*` (contenido suelto) y `modpack_*` (fases resolving/downloading/installing/installed/error) con `sessionId`, mismo patrón que `modloader_*`.
  - `Mods_test.go`: 6 tests (mapeo, gameDir, sanitización, descarga con hash contra `httptest`, extracción + overrides con zip maligno que intenta escapar).
- **`internal/Core/Launcher/Instance/GameDir.go` (nuevo)**: `GameDirFor(name)` resuelve el gameDir de la instancia (`<instancia>/game` con separada activa o la propia carpeta si no), falla si no existe o está en verificación; `HasInstance` para existencia.
- **`internal/Core/Launcher/Manager.go`**: getter `GetSeparateGameDir()` (valor efectivo con `RLock`, sin I/O bajo lock) para que el destino global coincida con el gameDir real del lanzamiento también en modo Minecraft (donde el manager fuerza `false`).
- **`internal/Handlers/Engine/Mods.go` (nuevo)**: `InstallModContent` (descarga al juego global o a una instancia con verificación de hash), `InstallModpack` (orquesta: descarga `.mrpack` → crea instancia con nombre único → descarga su Minecraft → instala su loader → copia overrides → baja los archivos del manifiesto con progreso), `CancelModContent`, `GlobalGameDir`, `InstanceGameDir`, `ContentSubdirFor`. Todo asíncrono por sesión con `context` cancelable (los bindings nunca se bloquean).
- **`internal/Services/Mods/Service.go` (nuevo) + registro en `main.go`**: décimo servicio de dominio (`ModsService`), con bindings generados.

### 2. Frontend: el diálogo de descarga ya instala de verdad

- **`Mods/DownloadDialog.vue`**: el botón Descargar (antes deshabilitado con tooltip) ahora instala. Selector de destino **Juego global / Instancia** (con lista de instancias y ruta de destino visible `.../mods/<archivo>`), aviso no bloqueante si el destino no tiene modloader (los `.jar` no cargan sin loader), y para modpacks campo de nombre + explicación (crea la instancia, baja su MC, instala su loader, copia overrides y sus mods/texturas). Progreso en vivo por eventos del backend (con contador de archivos en modpacks), cancelación y cierre sin matar la instalación (sigue en segundo plano, como las versiones).
- **`Mods/Styles/DownloadDialog.scss`**: bloques `.ModsDl_Dest`, radios, aviso de loader, barra de progreso y estado de éxito, con `@use` y `var(--color-*)` del sistema existente.
- **Bindings a mano** (`Services/Mods/modsservice.ts` + `index.ts`, tipos en `Handlers/Engine/models.ts` e `index.ts`): se regeneró con `wails3 generate bindings`, pero el wails3 local (beta.25) reescribe los 99 bindings al formato nuevo e introduce un diff masivo, así que se revirtieron y se añadieron solo los del servicio nuevo en el formato del repo (beta.9). Regla: no regenerar bindings hasta alinear la versión de wails3.

## Por qué

El panel de Mods solo navegaba y el diálogo solo elegía versión sin instalar nada. Ahora el launcher resuelve el destino según la opción de carpeta `game` (separada → `<workDir>/game` o `<instancia>/game`; deshabilitada → directorio base o raíz de la instancia) y los modpacks `.mrpack` se instalan creando su instancia desde el manifiesto, sin `.jar` empaquetados (vienen por URL en el manifiesto).

## API afectada

- Nuevo servicio `ModsService`: `InstallModContent`, `InstallModpack`, `CancelModContent`, `GlobalGameDir`, `InstanceGameDir`, `ContentSubdirFor`.
- Nuevos eventos frontend: `modcontent_resolving/downloading/installed/error`, `modpack_resolving/downloading/installing/installed/error` (payload con `sessionId`, `message`, `progress/total`, `instance`).
- Sin cambios en servicios existentes ni en el store central de descargas (el progreso vive en el diálogo; el widget sigue con versiones/instancias/loaders).

## Comportamiento anterior/nuevo

- Antes: botón Descargar deshabilitado ("la instalación se habilitará con el sistema de descargas").
- Ahora: mods/shaders/texturas se guardan en su carpeta del destino elegido; modpacks crean su instancia lista para jugar. Si el destino no tiene loader, avisa pero igual guarda el archivo.

## Cómo verificar

- `go build ./...` en `launcher/` — pasa.
- `go test ./internal/Core/Mods/...` en `launcher/` — 6 tests en verde (incluye zip maligno con `../`).
- `go vet` de los paquetes tocados — limpio.
- `bun run build` en `launcher/frontend` (type-check + Vite) — pasa.
- Manual: Mods → cualquier mod → Descargar → elegir destino → comprobar el `.jar` en `mods/`; modpack → Descargar → comprobar la instancia nueva con sus mods y overrides.

---

## Ronda 2 — Destino libre para modpacks, icono, widget, gestión y UI (feedback del owner)

### Qué cambió

### 1. Modpacks sin imposiciones: global, existente o nueva

- `InstallModpack` acepta `destination`: `global` (verifica/descarga el MC con `StartFullDownload` e instala el loader en el juego global de forma síncrona en su goroutine), `instance` (verifica el MC, instala el loader del manifiesto **solo si la instancia no tiene ninguno** — si ya tiene uno se reutiliza — con mensaje de error accionable si falta la versión base) o `new` (flujo anterior).
- En las tres ramas se crean los directorios (`EnsureContentDirs`) **antes** de copiar nada, y el diálogo ofrece las tres opciones con explicaciones (nada se impone).
- La instancia nueva nace con el **icono del modpack**: se descarga (`DownloadIcon`, extensión detectada del contenido porque las URLs de Modrinth no traen), se guarda en `<instancia>/assets/` y se registra en el metadata (best-effort, no rompe la instalación si falla).

### 2. Adiós al `no modloader installed` en el log

- Causa: `LoadState` devolvía error cuando no hay loader (estado normal en instancias vanilla), y cada consulta desde el frontend quedaba registrada como `ERR Binding call failed`.
- `ErrNoLoaderState` ahora es sentinela exportado y `InstalledModLoader` (instancias) + `GetInstalledModLoader` (global) devuelven `(nil, nil)` en ese caso. El frontend ya trataba `null` como "sin loader".

### 3. Progreso visible en el menú principal

- `Instances/Store.ts`: sesiones `contentDls` alimentadas por `modcontent_*/modpack_*` (viven en el store para sobrevivir al cierre del diálogo), mezcladas en `allActiveDownloads` como `kind: 'mod'` cancelables (`cancelContentDownload` → `CancelModContent`).
- `Downloads/Widget.vue` muestra `Instalando <pack> en <instancia>` con su fase; al pulsarlo abre el panel de Mods (`App.vue`).

### 4. Sección Instalado para administrar el contenido

- Backend: `ListInstalledContent`, `SetInstalledContentEnabled` (rename `.disabled`: Minecraft ignora lo que no es `.jar`/`.zip`), `DeleteInstalledContent`.
- Frontend: `Mods/Installed.vue` (nueva sección **Instalado** en la cabecera del panel, junto a Explorar): selector global/instancias, filtros con conteos, activar/desactivar, borrar con confirmación global, sesiones en curso cancelables y refresco tras instalar.

### 5. Explorador menos saturado

- `Mods/Content.vue`: sidebar plegable (botón en la barra), tarjetas con 2 chips de loader + descargas (la etiqueta de entorno vive solo en el detalle).

### API afectada (ronda 2)

- `ModsService`: `+ ListInstalledContent`, `+ SetInstalledContentEnabled`, `+ DeleteInstalledContent`; `ModpackInstallRequest` gana `destination/instance/iconUrl`.
- Bindings a mano como en la ronda 1 (mismo motivo: el wails3 local reescribe los 99 archivos a otro formato).

### Cómo verificar (ronda 2)

- `go build ./...` en `launcher/` — pasa. `go test -count=1 ./internal/Core/Mods/...` — 9 tests en verde (gestión + icono + dirs).
- `bun run build` (type-check + Vite) en `launcher/frontend` — pasa; `bun run type-check` final — limpio.
- Manual: instalar el modpack en global/existente/nueva; comprobar icono, loader y carpetas; ver el widget durante la instalación; gestionar desde Instalado.
