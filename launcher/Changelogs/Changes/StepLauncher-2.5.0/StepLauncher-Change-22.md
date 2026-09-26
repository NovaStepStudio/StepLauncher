# Cambios de StepLauncher 2.5.0 (Step-22) — Iconos del contenido, provisioning de modpacks, Mods por instancia y banner colapsable

- **Fecha**:   2026-09-25
- **Versión**: 2.5.0
- **Estado**:  implementado y verificado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

### 1. Iconos reales del contenido instalado

- `internal/Core/Mods/Icon.go` (nuevo): `ExtractContentIcon` extrae la imagen que ya trae cada archivo a `cache/content-icons/` y devuelve su ruta relativa (lista para `loadLocal`):
  - texturas/shaders (`.zip`): `pack.png` de la raíz (verificado contra `FreshAnimations_v1.10.5.zip` real).
  - mods Fabric (`.jar`): campo `icon` de `fabric.mod.json` (string o mapa por tamaños), verificado contra `fabric-api-0.161.0+26.3.jar` real (`assets/fabric/icon.png`).
  - mods Quilt: `metadata.icon` de `quilt.mod.json`.
  - mods Forge/NeoForge: `logoFile` de `META-INF/mods.toml`.
  - Sin icono o archivo inexistente: devuelve vacío sin error (la UI usa el genérico por tipo, sin ensuciar el log).
- `Engine.GetContentIcon` + `ModsService.GetContentIcon`: la pestaña Instalado y la nueva pestaña Mods de cada instancia muestran el icono real de cada mod/textura/shader.
- Corrección: a `.ModsInstalled_Icon` le faltaba el ajuste de `img` y reventaba la vista con iconos a tamaño natural (reportado con captura: engranaje gigante tapando las filas). Ahora `object-fit: cover` recortado al icono.

### 2. Modpacks: la instancia aparece solo al 100% y bloqueada mientras tanto

- `internal/Core/Launcher/Instance/Provision.go` (nuevo) + marcas en `Manager.go`: `BeginProvisioning` reserva el nombre (falla si existe o ya se crea), `List()`/`Get()` ocultan la instancia en creación y `assertUsable` bloquea lanzarla, modificarla, verificarla, borrarla o descargarle hasta que esté lista. El flujo interno del motor usa variantes `System` (`AddVersionSystem`, `InstallModLoaderSystem`, `UpdateConfigSystem`, `UpdateMetadataSystem`, `GameDirForSystem`) que saltan solo ese bloqueo.
- Al terminar: se fija la versión del loader recién instalado como **versión activa** (`setDefaultLoaderVersion`: ya no hay que ir a Versiones a seleccionarla, se puede jugar directo) y recién ahí se libera la marca y la instancia aparece. Icono y versión se aplican antes de liberarla.
- Al fallar o cancelar: `DeleteProvisioned` borra el directorio a medias y libera el nombre (verificado en `Provision_test.go`).
- Frontend: `ListProvisioning` + tarjetas bloqueadas "Creando…" en la lista de instancias (icono del pack, fase en vivo desde las sesiones, barra y Cancelar). El widget y el diálogo ya seguían la sesión.

### 3. Pestaña Mods en cada instancia

- `Instances/InstanceMods.vue` + `Styles/InstanceMods.scss` (nuevos): gestor compacto dentro del detalle (pestaña Resumen/Versiones/**Mods**/Capturas) con buscador, filtros por tipo con conteos, iconos reales, habilitar/deshabilitar, borrar con confirmación y botón Añadir (abre Mods con la instancia preseleccionada vía `pendingInstance` en `Mods/Store.ts`, que el diálogo aplica al abrir).

### 4. Descarga mostrada por pasos + Instalado mejorado

- `Mods/DownloadDialog.vue`: la fase de trabajo del modpack muestra pasos (Modpack → Minecraft → Modloader solo si el pack lo exige → Archivos) deducidos de las fases vistas, además del mensaje, contador y barra.
- `Mods/Content.vue`: la pestaña **Instalado** vive junto a Todos/Mods/… (se eliminó el control Explorar|Instalado de la cabecera, que quedaba mal) con separador propio; sidebar plegable; tarjetas con 2 chips como mucho.
- `Mods/Installed.vue`: botón Volver, buscador por nombre, filtro de estado (Todos/Activados/Deshabilitados) e iconos reales.

### 5. Banner comprimible en el detalle de instancia

- `Instances/Detail.vue`: botón junto a Cerrar para comprimir/expandir el hero. Comprimido oculta la imagen y la descripción, compacta icono y título, y el overlay de descarga pasa a flujo normal para seguir visible: se ve más contenido abajo.

### 6. Política de bindings: no se tocan a mano

- Aclarado con el owner: `frontend/bindings/` es salida de `wails3 generate bindings` y se regenera solo (incluso el plugin de Vite lo reescribe al compilar). Se revirtieron todas las ediciones manuales de rondas anteriores y se regeneró todo con el generador (224 métodos, 111 modelos), dejándolo tal cual sale. Regla: ante un método nuevo, regenerar y adaptar el frontend al formato generado, nunca editar el binding.

## Por qué

Feedback del owner con capturas: modpacks que dejaban instancias no listas a la vista, loader sin seleccionar por defecto, Instalado sin iconos/filtros y con el toggle feo en la cabecera, y la vista rota por iconos sin escalar. Más el banner que tapaba el contenido del detalle y la norma de no tocar bindings.

## API afectada

- `ModsService`: `+ GetContentIcon`; `InstanceService`: `+ ListProvisioning`.
- `InstanceManager`: `Begin/End/IsProvisioning`, `ListProvisioning`, `DeleteProvisioned`, variantes `System` (no son bindings directos salvo vía Engine).
- Eventos sin cambios (se reutilizan `modpack_*`/`modcontent_*`).

## Comportamiento anterior/nuevo

- Antes: la instancia del modpack aparecía desde el minuto cero y se podía abrir rota; la versión activa quedaba en el MC base; Instalado sin iconos ni filtros; hero fijo de 20.5rem.
- Ahora: la instancia aparece lista para jugar (icono del pack + loader activo); Instalado y Mods por instancia gestionan con iconos; hero comprimible.

## Cómo verificar

- `go build ./...` en `launcher/` — pasa.
- `go test -count=1 ./internal/Core/Mods/... ./internal/Core/Launcher/Instance/...` — 15 tests en verde (iconos fabric/forge/pack.png/ausente, provisioning oculta/bloquea/limpia).
- `bun run type-check` y `bun run build` en `launcher/frontend` — limpios (contra bindings recién generados, sin edición manual).
- Manual: instalar FO como nueva instancia → tarjeta Creando… bloqueada → aparece lista con icono y loader activo; Mods en el detalle con iconos; comprimir el hero.
