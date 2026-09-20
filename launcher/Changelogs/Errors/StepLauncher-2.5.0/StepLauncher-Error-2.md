# Errores de StepLauncher 2.5.0 (Error-2) — El tray nunca aparecía: nunca se cableó y `SystemTray.Run()` era un no-op durante el arranque

- **Fecha**:     2026-08-17
- **Versión**:   2.5.0 (en desarrollo)
- **Estado**:    corregido
- **Release**:   en desarrollo — aún no mencionado en ninguna release.

## Síntoma

Tras la migración a Wails v3, el icono del tray del área de notificaciones **no aparecía nunca**: ni durante la sesión de desarrollo, ni al ejecutar el binario. No había ningún icono en la bandeja del sistema con el que abrir el menú del launcher.

## Causa raíz

Dos causas independientes, ambas en la migración v2 → v3:

1. **`Tray.Setup()` nunca se invocaba**: `App.go` creaba el gestor (`tray.New(...)`) y conectaba `OnEngineEvent`, pero **ninguna llamada a `Setup(icon)`** registraba el tray. El paquete `internal/Tray` fue creado en la migración sin cablear su punto de entrada (grep de `Setup` solo lo encuentra en su propia definición).

2. **`SystemTray.Run()` es un no-op durante `ServiceStartup`**: en Wails v3 beta.9, `systemtray.go`:
   ```go
   func (s *SystemTray) Run() {
       if globalApplication == nil || globalApplication.running == false {
           return
       }
       ...
   }
   ```
   `ServiceStartup` se ejecuta dentro de `startup()` de `application.go`, en la línea 677, **antes** de `a.running = true` (línea 736). Aunque `Setup` se hubiera llamado desde el servicio, `Run()` retornaba en silencio, `impl` quedaba a nil y el icono jamás se creaba (`systemtray_windows.go` solo crea el icono en `run()` del impl).

## Diagnóstico y evidencia

- `grep -r "tray.Setup\|Setup("`: el único `Setup` relacionado es su definición en `internal/Tray/Tray.go`; en `App.go` solo hay `tray.New(...)` y `OnEngineEvent`.
- Lectura del fuente de Wails v3 beta.9 (`C:\Users\...\go\pkg\mod\github.com\wailsapp\wails\v3@v3.0.0-beta.9\pkg\application\`):
  - `systemtray.go:133-141` — guard de `Run()` contra `running == false`.
  - `application.go:671-737` — `startup()` ejecuta `a.startupService` antes de `a.running = true`.
  - `application_windows.go:148-157` — `m.run()` emite `events.Windows.ApplicationStarted` desde el bucle de mensajes, **después** de que la app está marcada como en marcha.
- Confirmación en `events_common_windows.go:9`: `events.Windows.ApplicationStarted → events.Common.ApplicationStarted`.

## Solución aplicada

1. **Cablear el tray** (`App.go`): icono embebido con `//go:embed build/appicon.png` (`var appIcon []byte`) y `a.tray.Setup(appIcon)` al final de `ServiceStartup`, después de `handler.Startup()`.
2. **Diferir `Run()` al evento correcto** (`internal/Tray/Tray.go`): `Setup` registra `t.app.Event.OnApplicationEvent(events.Common.ApplicationStarted, ...)` y es ahí donde se llama `t.sysTray.Run()` — la app ya está `running` y el bucle de mensajes puede crear el icono (los handlers de clic, icono y menú se configuran antes, como marca el patrón de la doc oficial).
3. **Limpieza al cerrar**: nuevo `Tray.Shutdown()` que llama a `sysTray.Destroy()` (la doc recomienda destruir el tray siempre al apagar); se invoca desde `ServiceShutdown`.

## Regla aprendida

- **En Wails v3, `SystemTray.Run()` debe llamarse cuando la app ya está en marcha** (`running == true`). Desde un `ServiceStartup` es un no-op silencioso: si el tray no aparece, verificar que `Run()` se ejecuta tras el arranque (p. ej. en el evento `Common.ApplicationStarted`).
- **Verificar el cableado completo del ciclo de vida**: un paquete creado en una migración puede estar definido pero nunca conectado (el tray tenía `New`/`OnEngineEvent` y faltaba el `Setup`). El grep de consumidores al terminar una migración evita estos "objetos huérfanos".

## Verificación

- `go build ./...` — EXIT 0 (el embed `build/appicon.png` resuelve).
- Ejecución en dev: el icono de StepLauncher aparece en la bandeja del sistema con su menú, y se destruye al salir.
