# Release 2.5.0 — StepLauncher (Beta)

- **Fecha**: 2026-08-31
- **Versión**: 2.5.0
- **Estado**: beta

## Resumen

Beta de la actualización más grande desde la 2.3.1: migración completa a **Wails v3**, nuevo **panel de Mods** (Modrinth cliente), **sistema Bootstrap** con splash trazable, reorganización de **Ajustes** para usuario promedio, **panel de Música reconstruido** con biblioteca real y selector visual, **capa de Servicios** desacoplada, **índice persistente de biblioteca musical**, **infra de build multiplataforma** y **configuración extendida**. Esta beta consolida 11 cambios, 4 errores y 2 bugs de la rama 2.5.0 en desarrollo, lista para pruebas internas antes de la final.

> **Nota beta**: esta versión es de pruebas. Puede contener regresiones menores. La actualización final 2.5.0 llegará tras validar el índice `index_v2.gob.gz` con bibliotecas de 3000+ pistas.

---

## Funcionalidades nuevas

### 1. Migración completa a Wails v3 (`Change-1`)
- Backend `go 1.26.4` con `wails/v3 v3.0.0-beta.9`, `App` como servicio, `RuntimeBridge`, frontend con `@wailsio/runtime`.

### 2. Música de fondo fiable (`Change-2`)
- `<audio>` nativo + data URI base64, volumen en Ajustes, `coverStyle: background`.

### 3. Tray renovado (`Change-3`)
- Un solo icono, clics y submenús clicables, cuenta con radio.

### 4. Panel de Mods (`Change-4`)
- Dominio `Mods/` maqueta Modrinth cliente.

### 5. Sistema Bootstrap (`Change-5`)
- `Common/Bootstrap/` con 8 tareas, `SplashScreen` con logs, `Main.ts` orquestador.

### 6. Ajustes reorganizados (`Change-6`)
- 4 secciones nuevas: Network, Downloads, Integrity, Storage.

### 7. Panel de Música reconstruido (`Change-7`)
- `PlayerStore` + `TrackSelector` + 5 vistas, historial y widget dual.

### 8. Capa de Servicios + motor musical (`Change-8`)
- 9 servicios desacoplados, `Index` `index_v2.gob.gz`, scanner incremental.

### 9. Infra de build (`Change-9`)
- `build/config.yml` v3, `Taskfile.yml`, `frontend/bindings`.

### 10. Configuración extendida (`Change-10`)
- `VerifyBeforeLaunch`, `ImageAuthor/ModName/Url`, `PruneOrphanGallery`.

### 11. Stores transversales (`Change-11`)
- `Connectivity`, `OfflineBadge`, `useBackground`, `useCoverPalette`.

---

## Errores corregidos

- [StepLauncher-Error-1: `wails3 dev/build` rotos por `build/ios`](../../Errors/StepLauncher-2.5.0/StepLauncher-Error-1.md) — corregido.
- [StepLauncher-Error-2: Tray nunca aparecía](../../Errors/StepLauncher-2.5.0/StepLauncher-Error-2.md) — corregido.
- [StepLauncher-Error-3: Dos iconos de tray](../../Errors/StepLauncher-2.5.0/StepLauncher-Error-3.md) — corregido.
- [StepLauncher-Error-4: Verificación antes de lanzar ignorada y fondos duplicados](../../Errors/StepLauncher-2.5.0/StepLauncher-Error-4.md) — corregido.

## Bugs resueltos

- [StepLauncher-Bug-1: Acumulación infinita de fondos de galería](../../Bugs/StepLauncher-2.5.0/StepLauncher-Bug-1.md) — corregido.
- [StepLauncher-Bug-2: Widget con textos largos y carátulas sin revoke](../../Bugs/StepLauncher-2.5.0/StepLauncher-Bug-2.md) — corregido.

---

## Notas beta

- Instalar la beta reemplazando el ejecutable 2.3.1; la configuración se migra sola.
- Música: escaneo inicial crea `cache/music/index_v2.gob.gz`, migración automática de shards antiguos.
- Reportar cualquier `PruneOrphanGallery` anómalo.
