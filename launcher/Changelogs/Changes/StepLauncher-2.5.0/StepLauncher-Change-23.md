# Cambios de StepLauncher 2.5.0 (Step-23) — Iconos NeoForge y por proyecto, mover entre destinos, provisioning sin ocultar y banner animado

- **Fecha**:   2026-09-25
- **Versión**: 2.5.0
- **Estado**:  implementado y verificado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

### 1. Iconos que faltaban (diagnóstico con jars reales del owner)

- Causa 1: los mods NeoForge nuevos usan `META-INF/neoforge.mods.toml` y solo se leía `mods.toml`. Ahora se prueban ambos (`Icon.go`: `forgeIcon` recorre los dos nombres).
- Causa 2: `quilt.mod.json` guarda icono y nombre en `quilt_loader.metadata`, no en `metadata` raíz: se leen ambas rutas.
- Causa 3 (sin arreglo posible desde el archivo): jars sin metadatos (p. ej. Essential) o con `logoFile` comentado (JEI): ahí el genérico es lo correcto. Para cubrirlos, al descargar un mod se guarda el icono del proyecto (`CacheProjectIcon` en `cache/content-icons/by-name/`, con `IconURL` nuevo en `ModContentRequest`) y `ExtractContentIcon` lo usa como respaldo.
- Además: `ContentDisplayName` (nombre bonito de fabric/quilt/toml) y campo `title` nuevo en `InstalledFile`: las filas muestran "Just Enough Items" con el filename como subtítulo.

### 2. Mover contenido y abrir su ubicación

- Backend: `MoveInstalledContent` (entre global e instancias, conservando habilitado, sin sobrescribir) y `RevealInstalledContent` (explorador con el archivo seleccionado; `InstalledPath` + `RevealInExplorer` por plataforma) + `ModsService.Move/Reveal`.
- Frontend: componente `Mods/MoveTo.vue` (selector global/instancia) integrado en Instalado y en Mods por instancia; botones de carpeta en cada fila.

### 3. Provisioning sin esconder nada + sesiones con detalle

- Bug crítico: el bloque de "Creando…" usaba `v-if`/`v-else` con la rejilla y ocultaba todas las instancias durante la creación. Ahora son independientes: la tarjeta bloqueada va arriba y la lista sigue visible.
- Sesiones de Instalado con icono del proyecto (`contentMeta` registrado al iniciar cada descarga) y detalle expandible (destino, barra x/y archivos).
- Encabezado de Instalado reagrupado (volver + título a la izquierda, acciones a la derecha).

### 4. Progreso real 0-100% y feed del loader

- La descarga del Minecraft del modpack emite porcentaje y MB (`ensureInstanceVersion`/`ensureGlobalVersion` con callback de progreso, `total=100`): widget, diálogo y tarjetas lo muestran de 0 a 100.
- `modloaderFeed` en el store (último evento por sesión de loader): el diálogo muestra su mensaje y % mientras el paso es Modloader.

### 5. Banner comprimido con animación y con imagen

- Se revierte el `display:none`: comprimido es una tira animada (`transition` en `min-height`, opacidad e icono) que conserva la imagen tenue, con fila compacta y overlay de descarga en flujo.

### 6. Bug visual de iconos gigantes

- `.ModsInstalled_Icon` no limitaba el `img`: los iconos salían a tamaño natural rompiendo las filas (captura del owner). Ahora `overflow:hidden` + `object-fit:cover`.

### 7. Corrección: la tarjeta "Creando…" ya no tapa la lista (y vive en la rejilla)

- Bug (UI, reportado con captura): el bloque de provisioning usaba `v-if`/`v-else` encadenado con la rejilla y, al haber una creación en curso, ocultaba todas las instancias.
- Primero se separó el bloque; a pedido del owner se integró **dentro de la rejilla** como tarjeta bloqueada (sin abrir, fijar, editar ni jugar; solo progreso y Cancelar), en la misma cuadrícula que las demás. Solo frontend (`Instances/List.vue` + estilos); sin tocar el backend.

### 8. Corrección: creando compacta sin banner, con filtro y errores visibles + diálogo que cargaba vacío

- La tarjeta en creación ahora es compacta **sin banner** (icono, título, fase, barra, Cancelar) y la rejilla usa `align-items: start` para que las demás cards no se estiren.
- Nuevo filtro `Creando N` (ojo mostrar/ocultar) en la fila de filtros: oculta las no listas sin cancelarlas.
- Los fallos (p. ej. instalador Forge con HTTP 404) ya no pasan en silencio: quedan como tarjeta de error en instancias y fila en Instalado, con el motivo y botón Entendido para descartarlas (`failedContent` en el store; el widget sigue limpiándose solo).
- Bug real del diálogo de descarga: el Host lo monta ya visible, así que el `watch(visible)` nunca se disparaba y **nunca cargaba la lista de instancias** (de ahí el "No hay instancias" teniéndolas). Ahora la carga va en `onMounted` vía `initDialog()` (+ preselección desde instancia).
- Solo frontend; backend sin compilar.

## API afectada

- `ModsService`: `+ MoveInstalledContent`, `+ RevealInstalledContent`; `ModContentRequest.iconUrl`; `InstalledFile.title`.
- Bindings 100% generados por wails3 (226 métodos), sin edición manual, tal como salen.

## Cómo verificar

- `go build ./...` — pasa; `go vet` — limpio.
- `go test -count=1 ./internal/Core/Mods/... ./internal/Core/Launcher/Instance/...` — 18 tests en verde (toml neoforge, caché por nombre, títulos, mover).
- `bun run type-check` y `bun run build` — limpios contra bindings recién generados.
- Manual: iconos en Pixelmon (JEI seguirá genérico: no trae logo); mover fabric-api global→instancia; revelar; crear modpack viendo 0-100% y feed del loader; comprimir banner animado.
