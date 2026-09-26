# Cambios de StepLauncher 2.5.0 (Change-4) — Panel de Mods: maqueta del navegador de Modrinth (mods, modpacks, shaders y texturas de cliente)

- **Fecha**: 2026-08-18
- **Versión**: 2.5.0 (en desarrollo)
- **Estado**: implementado y verificado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

Nuevo dominio `frontend/web/src/Mods/` con el panel de descargas de contenido de Modrinth para el **cliente** (nada de servidores). Es una maqueta: el panel ya navega y muestra el layout definitivo con una card de ejemplo, pero **no carga datos reales** — la integración con la API llegará en una tarea posterior.

### 1. Investigación de la API de Modrinth (Labrinth v2)

Documentada en `Mods/Store.ts` (comentarios): base `https://api.modrinth.com/v2`, `GET /search` (facets con AND/OR en arrays, index relevance/downloads/follows/newest/updated, offset/limit), `GET /project/{id|slug}`, `GET /project/{id|slug}/version` (files con URL directa, filename, size y hashes sha1/sha512), `GET /project/{id|slug}/dependencies` y `GET /tags/...`. Tipos de proyecto: `mod`, `modpack`, `resourcepack` y `shader` — exactamente los cuatro que descargará el panel. Para garantizar contenido de cliente se usa la facet `environment` (acepta `client_only`, `client_and_server`, `client_only_server_optional`, `singleplayer_only`, `client_or_server`, `client_or_server_prefers_both`; excluye los `server_only*`).

### 2. Panel de Mods (maqueta)

- `Common/Overlays/Store.ts`: `HeavyPanel` ahora incluye `'mods'`.
- `Mods/Mods.vue` + `Mods/Styles/Mods.scss`: overlay de pantalla completa (mismo patrón que `Instances/Instances.vue`): cabecera con icono de puzle, título, subtítulo, botón cerrar, cierre con ESC (`useOverlayEscape`) y cierre por inactividad (`CLOSE_OVERLAYS_EVENT`).
- `Mods/Content.vue` + `Mods/Styles/Content.scss`: contenido del panel:
  - Barra superior con título "Explorar Modrinth" y buscador (deshabilitado, aún sin API).
  - Aviso persistente de que solo se descarga contenido de **cliente**.
  - Pestañas de tipo de proyecto: Todos, Mods, Modpacks, Shaders y Texturas (cambian el estado activo, sin carga todavía).
  - Aviso de "vista previa del diseño".
  - Grilla de cards con **una card de ejemplo** ("Sodium", datos estáticos con la misma forma que un `SearchHit` de `/search`): icono (fallback), título, autor, descripción, chips de categorías/loaders, versiones de Minecraft, descargas y seguidores formateados (`formatCount`: 75,4M / 124K), badge de tipo, badge de entorno ("Solo cliente") y botones Descargar (deshabilitado) y favorito (funcional localmente).
- `App.vue`: el ítem "Mods" del sidebar (que ya existía sin acción) ahora abre el panel con `openHeavyPanel('mods')`; se renderiza con `v-show` + `Transition name="ModsModal"`; `closeAllPanels()` también lo cierra. El menú principal se oculta automáticamente porque `mainMenuHidden` ya reacciona a `heavyPanel`.

## Por qué

El launcher no tenía forma de descargar mods, shaders o texturas. Este panel es la base visual y estructural (dominio feature-first con su `Store.ts`, componentes y `Styles/`) sobre la que se construirá la integración real con la API de Modrinth: tipos TS `ModrinthProject` mapeados 1:1 a `SearchHit`, facetas de cliente y formato de números ya listos.

## API afectada

- Frontend únicamente: `Common/Overlays/Store.ts` (tipo `HeavyPanel`), `App.vue`, y el dominio nuevo `Mods/` (4 archivos).
- Ningún binding de Wails cambió. No se toca backend.

## Comportamiento anterior/nuevo

- Ítem "Mods" del sidebar: sin acción → abre el panel de Mods a pantalla completa ocultando el menú principal (con la transición y blur habituales).
- No había forma de ver cómo se verán los mods → el panel muestra el layout completo con tabs, avisos de cliente y una card de ejemplo marcada como tal.

## Cómo verificar

- `bun run build` en `frontend/` (type-check + vite build): pasa.
- `go build ./...` en la raíz: compila sin errores.
- Ejecutar el launcher, pulsar "Mods" en el sidebar: se abre el panel a pantalla completa, el menú principal se oculta, ESC y la X lo cierran.
- Las pestañas Todos/Mods/Modpacks/Shaders/Texturas cambian el estado activo; la card de ejemplo se ve con sus chips, badge "Ejemplo", badge de entorno y botones.
