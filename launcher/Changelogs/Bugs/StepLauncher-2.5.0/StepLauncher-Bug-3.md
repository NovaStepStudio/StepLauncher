# Bugs de StepLauncher 2.5.0 (Bug-3) — Manifiesto roto sin red + interruptor de proxy imposible de activar

- **Fecha**: 2026-09-24
- **Versión**: 2.5.0
- **Estado**: corregido y verificado
- **Release**: en desarrollo — aún no mencionado en ninguna release.

## El bug en cuestión

Dos problemas encadenados reportados por el owner:

1. **Carga de versiones rota desde el manifiesto**: el binding fallaba con `fetch manifest: Get "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json": net/http: HTTP/1.x transport connection broken: malformed HTTP status code "\x00..."`. Sin red no había lista de versiones y el error era terminal aunque el launcher ya hubiera descargado el manifiesto antes. Causas combinadas: el fallback a caché solo valía si la copia estaba **expirada pero aún en disco**; la limpieza horaria (`internal/Core/Cache/Cache.go:258`) **borraba** los expirados al cabo de una hora; y `RefreshManifests` (`internal/Handlers/Engine/Versions.go:213`) **borraba toda la categoría antes de descargar**, así que un refresco con red rota destruía la única copia buena y dejaba el error permanente.

2. **El interruptor del proxy no se podía activar**: al encenderlo sin dirección, `frontend/web/src/Settings/Sections/Network.vue` apagaba el toggle a la fuerza, pero los campos de dirección estaban ocultos tras `v-if="proxyEnabled"`, así que la dirección **nunca se podía escribir** (punto muerto). Además `normalizeProxy` (`internal/Config/Config.go:101`) **borraba** host/puerto/credenciales al desactivar el proxy incompleto, destruyendo lo escrito.

## Qué afectaba y qué hacía

- **Manifiesto**: instalación de versiones e instancias bloqueada sin red, sin forma de recuperar lo ya descargado.
- **Ajustes > Red**: imposible configurar un proxy manual como vía de escape.

## Solución final

Simple y sin magia de red: **guardar y servir caché**.

- **`internal/Core/Cache/Cache.go` (`cleanup`)**: la categoría `manifest` queda exenta del auto-borrado por caducidad. Una copia expirada sigue sirviendo la lista de versiones; solo se elimina con limpieza explícita o al sobrescribirse.
- **`internal/Handlers/Engine/Versions.go` (`GetVersions`, `FetchVersionManifest`)**: ante **cualquier** fallo (red, HTTP distinto de 200, respuesta inválida) se sirve la copia guardada, fresca o expirada. Solo falla si no hay nada guardado.
- **`internal/Handlers/Engine/Versions.go` (`RefreshManifests`)**: ya no borra antes de descargar. Descarga primero y, con datos en mano (de red o de caché), invalida solo los derivados por tipo para que se regeneren desde `full`. Si la red falla, retorna la copia guardada sin borrar nada.
- **`frontend/web/src/Settings/Sections/Network.vue`**: la fila de dirección (host+puerto) **siempre visible**; al activar con datos incompletos se avisa pero **sin apagar el interruptor ni persistir**. Sin cambios de estilos.
- **`internal/Config/Config.go` (`normalizeProxy`)**: al desactivar un proxy incompleto **se conservan** los datos a medio escribir (antes se borraban).

## Comportamiento anterior/nuevo

- Antes: sin red = error terminal; refrescar sin red = caché destruida; proxy manual = imposible de activar.
- Ahora: sin red = se sirve la última copia guardada; refrescar sin red = conserva y retorna la copia; el proxy manual se puede activar escribiendo host/puerto (Clash: HTTP `7890`).

## Verificación

- `go build ./...` en `launcher/` — OK.
- `go vet` sobre `internal/Handlers/Engine`, `internal/Core/Cache`, `internal/Config` — OK.
- `bun run build` en `launcher/frontend` (type-check + Vite) — OK.
