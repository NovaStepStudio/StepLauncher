# Errores de StepLauncher 2.5.0 (Error-3) — Dos iconos de tray en la bandeja: `Run()` se ejecutaba dos veces

- **Fecha**:     2026-08-17
- **Versión**:   2.5.0 (en desarrollo)
- **Estado**:    corregido
- **Release**:   en desarrollo — aún no mencionado en ninguna release.

## Síntoma

Al ejecutar el launcher aparecían **dos iconos idénticos** de StepLauncher en el área de notificaciones, cada uno con su propio menú funcional.

## Causa raíz

`SystemTrayManager.New()` (system_tray_manager.go) ya orquesta el arranque del tray: si la app no está en marcha, lo añade a `pendingRun` (`runOrDeferToAppRun`) y `startup()` lo ejecuta al empezar el bucle de mensajes. El fix del Error-2 añadió además un listener en `events.Common.ApplicationStarted` que llamaba `sysTray.Run()` explícitamente: `Run()` se invocaba **dos veces** (una por el deferred de `New()`, otra por el evento), y en Windows cada `run()` crea su propia ventana de mensajes + `Shell_NotifyIcon` → dos iconos.

## Diagnóstico y evidencia

- Lectura del fuente de Wails v3 beta.9:
  - `system_tray_manager.go:16-27` — `New()` → `stm.app.runOrDeferToAppRun(newSystemTray)`.
  - `application.go:1000-1013` — `runOrDeferToAppRun`: si `!running` añade el tray a `pendingRun`.
  - `application.go:745-750` — `startup()` ejecuta cada `pendingRun` con `go pending.Run()`.
  - `systemtray.go:133-141` — `Run()` crea el impl (`newSystemTrayImpl`) y llama `impl.run()`.
  - `systemtray_windows.go:206-321` — `run()` crea el `hwnd` de mensajes y registra el icono en la bandeja.
- Conclusión: tray `New()` + defer automático + `Run()` manual en `ApplicationStarted` = dos `impl.run()` = dos iconos.

## Solución aplicada

- Eliminar de `Setup()` el listener de `ApplicationStarted` y la llamada explícita a `Run()`: `SystemTray.New()` ya difiere el arranque al momento correcto. `Setup` solo configura icono, tooltip, menú y handlers de clic.
- `go build ./...` — EXIT 0.

## Regla aprendida

- **En Wails v3 no hay que llamar a `SystemTray.Run()` a mano**: `SystemTrayManager.New()` ya lo difiere al arranque de la app (`runOrDeferToAppRun`). Llamar `Run()` manualmente (aunque sea desde `ApplicationStarted`) crea un segundo tray. Configurar icono/menú/handlers ANTES del arranque es suficiente porque `run()` los consume de los campos del `SystemTray`.

## Verificación

- `go build ./...` — EXIT 0.
- Ejecución: aparece **un solo** icono de StepLauncher en la bandeja con su menú completo; al salir se destruye.