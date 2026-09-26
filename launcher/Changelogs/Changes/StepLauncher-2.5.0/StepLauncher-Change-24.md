# Cambios de StepLauncher 2.5.0 (Step-24) — Diagnóstico del 404 de Forge: versión corta del manifiesto + loader no fatal

- **Fecha**:   2026-09-25
- **Versión**: 2.5.0
- **Estado**:  implementado y verificado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Diagnóstico (con logs del owner y evidencia)

- Log: `The Pixelmon Modpack-2` (MC 1.16.5) descarga bien su Minecraft y falla en `download forge installer: HTTP 404`.
- Causa raíz: el manifiesto `.mrpack` declara Forge en corto (`36.2.34`, formato Modrinth) pero el Maven de Forge exige la versión completa en la ruta (`1.16.5-36.2.34/forge-1.16.5-36.2.34-installer.jar`). El provider construía la URL con la corta tal cual → 404 siempre.
- Verificado contra el `maven-metadata.xml` real de Forge: todas las versiones son completas (`1.16.5-36.2.x`, `1.20.1-47.x`). De paso: `47.4.20` (Better MC) no existe en Maven (47.4 llega a 47.4.5): si un manifiesto pide una versión inexistente, el 404 es inevitable y entra el camino no fatal.
- Los flujos manuales no sufrían esto porque `GetVersions` ya devuelve versiones completas.

## Qué cambió

### 1. Normalización de la versión de Forge del manifiesto

- `MrpackIndex.RequiredLoader` normaliza con `NormalizeLoaderVersion`: Forge corta (`36.2.34` + MC `1.16.5`) → completa (`1.16.5-36.2.34`); si ya viene completa no se toca; fabric/quilt/neoforge no se tocan (usan corta en sus repos).

### 2. Fallo del loader no fatal (a pedido del owner)

- Si el loader no se puede instalar (404, versión inexistente, Java…), la instalación **sigue**: instancia (nueva o existente) o global se crean igual con mods, overrides y texturas, y se avisa con `Sin modloader X Y (motivo). El contenido está instalado; pon el loader desde Instancias > Descargar.`
- El aviso viaja en `Message` del evento `modpack_installed` (campo nuevo en el done) y lo muestran el diálogo, el widget y las sesiones; la versión activa queda en el MC base para que el loader se ponga luego desde el menú de descarga de la instancia.
- El error de descarga del orquestador ahora incluye la URL (`download forge installer (<url>): …`) para diagnosticar sin adivinar.

## Cómo verificar

- `go build ./...` — pasa.
- `go test -count=1 ./internal/Core/Mods/...` — 17 tests en verde (+ normalización forge corta/completa/otros).
- `bun run build` — limpio.
- Manual: reinstalar el Pixelmon (1.16.5 + forge corto del manifiesto) → debe instalar `1.16.5-36.2.x` sin 404; con versión inexistente → instancia creada con aviso.
