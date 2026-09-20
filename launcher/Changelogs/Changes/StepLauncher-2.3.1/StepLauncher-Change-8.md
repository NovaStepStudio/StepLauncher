# Changes/StepLauncher-2.3.1/StepLauncher-Change-8.md

- **Fecha**: 2026-08-09
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.

## Qué cambió

Nueva opción en Ajustes → General (al final de todo): botón **"Verificar integridad de archivos"** que recorre TODOS los JSON de las versiones descargadas (globales e instancias, o solo el sector elegido) y repara/verifica los archivos de forma lenta pero segura. Proceso manual independiente del toggle `verifyIntegrity` existente (ese solo controla la verificación SHA1 en caliente durante descargas).

### 1. Flujo del verificador (`internal/Handlers/Engine/Integrity.go`, nuevo)

Al pulsar el botón, el motor ejecuta en goroutine propia (sin bloquear bindings, `context` de cancelación, pool con `ConcurrentDownloads*3` máx. 64 usando el HTTP client compartido):

- **Indexación**: recorrido de TODOS los JSON (`versions/<id>/<id>.json` globales y `instances/<nombre>/versions/<id>/<id>.json` de instancias) respetando la estructura de carpetas de cada tipo: el global usa `WorkDir` y la instancia usa `WorkDir/shared` + `InstanceVersionDir` en la carpeta de la instancia. Las tareas esperadas se construyen con `downloader.BuildTasks` (client, libraries, natives, assets, java) y se deduplican por ruta (librerías/assets compartidas se verifican una vez). Los JSON ilegibles se registran en el backend y se saltan.
- **Fase 1 (existencia)**: si falta un archivo se descarga con hasta **3 intentos**; si falla 3 veces, se cancela **ese archivo** (no la versión ni el proceso), se registra en el backend (`e.log.Warn`) y se guarda en memoria (lista pendiente).
- **Fase 2 (reintento)**: vuelta a "verificar todos los archivos": se re-comprueba la existencia de los pendientes y se re-descargan con hasta **5 intentos**; si un archivo sigue sin responder se saltea y queda registrado en el backend (`e.log.Error`).
- **Fase 3 (hash y tamaño)**: verificación final de TODOS los archivos por **SHA1 y tamaño**; los corruptos se borran, se re-descargan (5 intentos) y, si no se recuperan, se saltean y registran en el backend.
- **Sector**: `StartIntegrityCheck(scope)` acepta `todo` / `global` / `instances`; `IntegrityStatus()` devuelve el progreso (fase, %, versiones escaneadas, faltantes, restaurados, corruptos, descartados, lista `skipped` máx. 200) y `CancelIntegrityCheck()` cancela la ejecución.

### 2. Config (`internal/Config/Config.go`)

Nuevo campo `launcher.integritySector` (string `todo`/`global`/`instances`, default `todo`), sanitizado en `sanitize()` y persistido en `launcher_config.json` vía `SetIntegritySector`.

### 3. Bindings

`internal/Handlers/App.go` y `app.go`: `StartIntegrityCheck(scope string) error`, `CancelIntegrityCheck()`, `IntegrityStatus()`, `SetIntegritySector(s)`, `GetIntegritySector()` (se registran en el runtime en el próximo `wails build`/`wails dev`).

### 4. UI (`GeneralSettings.vue` + `GeneralSettings.scss`)

Nuevo grupo "Integridad" al final de General (tras "Actualizaciones"):

- Selector de sector segmentado (Todo / Global / Instancias), elección guardada en config.
- Botón "Verificar integridad de archivos": mientras corre muestra "Verificando… X%" con barra de progreso discreta y se deshabilita; al terminar muestra resumen breve ("Completado: N versiones, M restaurados, K descartados") que se limpia solo. Sin modales ni errores en la UI: las incidencias solo quedan en el backend.
- Polling de `IntegrityStatus` cada 500 ms mientras corre; al volver al panel se sincroniza si hay una verificación en curso.

## Por qué

No había forma manual de reparar/verificar instalaciones completas (globales o de instancias) sin reinstalar cada versión; el toggle existente solo afectaba a descargas nuevas. El proceso por fases (3 intentos → memoria → 5 intentos → saltear y registrar) evita descargas infinitas y deja constancia de cada archivo no recuperable solo en el backend, como pidió el usuario.

## API afectada

- `launcher_config.json`: nuevo campo `integritySector` (ausente → `todo` por sanitize).
- Bindings Wails nuevos: `StartIntegrityCheck`, `CancelIntegrityCheck`, `IntegrityStatus`, `SetIntegritySector`, `GetIntegritySector`.
- Reutiliza `downloader.BuildTasks`, `downloader.DownloadFile` (reintentos + stall timeout), `downloader.VerifySHA1` y `downloader.FileExists`; la fase 3 añade comprobación de tamaño además del SHA1.

## Comportamiento anterior/nuevo

- Anterior: no existía verificación manual; los archivos dañados solo se detectaban (o no) al descargar.
- Nuevo: botón con sector configurable que recorre los JSON, descarga faltantes (3 intentos), reintenta pendientes (5 intentos), verifica SHA1+tamaño de todo y registra en el backend cada archivo descartado.

## Cómo verificar

- `go build ./...` en la raíz: pasa.
- `bun run build` en `frontend/` (vue-tsc + sass + vite): pasa.
- En el launcher: Ajustes → General → grupo Integridad → elegir sector → "Verificar integridad"; mientras corre el botón muestra % y barra; al terminar, resumen. Los logs `[Integrity]` aparecen en el archivo del motor.

## Pendientes (verificables al ejecutar)

- `wails build` (o `wails dev`) para registrar los 5 bindings nuevos en el runtime; sin eso la UI no encuentra las funciones (los `go` bindings de `frontend/wailsjs` se regeneran ahí).
- Prueba real del ciclo de reintentos (3 → 5 → saltear) con red cortada a mitad: los archivos deben quedar en `[Integrity]` WARN/ERROR del log y el proceso debe terminar en completado con los descartados contados.