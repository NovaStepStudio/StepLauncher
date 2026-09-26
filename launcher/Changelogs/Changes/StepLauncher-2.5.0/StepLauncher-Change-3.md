# Cambios de StepLauncher 2.5.0 (Change-3) — Tray renovado: un solo icono, sesiones e instancias clicables, cuenta seleccionable y acciones rápidas

- **Fecha**: 2026-08-17
- **Versión**: 2.5.0 (en desarrollo)
- **Estado**: implementado y verificado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

Mejora completa del tray (`internal/Tray/Tray.go`) usando únicamente la API pública de system tray de Wails v3 (`app.SystemTray`, menús e items), con eventos al frontend para lanzar partidas, abrir vistas y sincronizar la cuenta. No se tocó código de Wails.

### 1. El tray aparece una sola vez (fix de arranque, ver [Error-2](../../Errors/StepLauncher-2.5.0/StepLauncher-Error-2.md) y [Error-3](../../Errors/StepLauncher-2.5.0/StepLauncher-Error-3.md))

- `App.go`: el tray **nunca se había cableado** — `tray.New()` existía pero `Setup(icon)` no se llamaba. Ahora el icono se embebe (`//go:embed build/appicon.png`) y `a.tray.Setup(appIcon)` se invoca al final de `ServiceStartup`.
- `internal/Tray/Tray.go`: **no se llama a `Run()` manualmente** — `SystemTrayManager.New()` ya difiere el arranque del tray al momento correcto (`runOrDeferToAppRun`). El intento previo de diferirlo a `ApplicationStarted` con `Run()` explícito creaba **dos iconos** en la bandeja (ver Error-3).
- Nuevo `Tray.Shutdown()` → `sysTray.Destroy()`, llamado desde `ServiceShutdown` para liberar el icono al salir.

### 2. Clics del icono

- **Clic izquierdo** (`OnClick`): acción principal — muestra/oculta la ventana principal.
- **Clic derecho** (`OnRightClick`): abre el menú explícitamente con `OpenMenu()`.
- **Doble clic** (`OnDoubleClick`): muestra y enfoca la ventana principal.

### 3. "Últimas Sesiones" e "Instancias" ahora son clicables

- `internal/Tray/Tray.go` emite dos eventos nuevos al frontend: `tray_launch_version` (con la versión) y `tray_launch_instance` (con el nombre de la instancia); el frontend reutiliza sus flujos de lanzamiento existentes con todo el feedback de UI (mensajes, descargas, ocultar ventana al lanzar):
  - `Launcher/Store.ts`: nuevo `launchVersion(version)` (mismo cuerpo que `launchGame` pero con versión explícita) + listener `Events.On('tray_launch_version', ...)`.
  - `Instances/Store.ts`: listener `Events.On('tray_launch_instance', ...)` → `launchInstance(name)` existente.

### 4. Submenú "Cuenta" con radio items

- `internal/Tray/Tray.go`: submenú "Cuenta" con hasta 5 cuentas (`AddRadio`), marcada la activa; al hacer clic, `engine.SetSelectedAccount(id)` persiste la selección y se emite `tray_account_selected` para sincronizar el frontend (`Accounts/Store.ts` actualiza `selectedAccountId` y recarga la lista). El menú se reconstruye para reflejar el nuevo radio marcado.

### 5. Acciones rápidas nuevas

- **Configuración** → evento `tray_open_settings` → `App.vue` abre el modal de Ajustes (`settingsOpen`).
- **Abrir Instancias** → evento `tray_open_instances` → `App.vue` abre el panel pesado de instancias (`openHeavyPanel('instances')`).
- **Abrir Descargas** → evento `tray_open_downloads` → `App.vue` abre el modal de instalación/descargas (`installOpen`).
- Los listeners se registran en `onMounted` de `App.vue` (mismo patrón que el resto de `Events.On`) y se limpian en `onUnmounted`.

### 6. Música como botón y menú limpio

- La música ya **no es un checkbox**: es un item normal ("Pausar Música" / "Reproducir Música") que alterna etiqueta y estado al hacer clic, con `menu.Update()` tras cada cambio y el estado conservado en el struct al reconstruir el menú.
- Se eliminó la cabecera de estado ("StepLauncher — Inactivo") y el atajo "Jugar de nuevo": el menú es compacto — submenús (Sesiones, Instancias, Cuenta) → música → ventana → Configuración/Instancias/Descargas → GitHub/Salir.

## Por qué

El tray anterior era estático: solo tenía menú (sin clics), no informaba de partidas en curso, la música era un item plano sin estado y el menú entero se reconstruía sin `Update()`. Además, tras la migración a v3 el tray no aparecía (nunca se cableó), y al cablearlo mal aparecían dos iconos.

## API afectada

- `internal/Tray/Tray.go` completo y `App.go` (embed del icono + `Setup` + `Shutdown`).
- Eventos emitidos por el tray: `tray_launch_version`, `tray_launch_instance`, `tray_account_selected`, `tray_open_settings`, `tray_open_instances`, `tray_open_downloads` (además del `music_tray_toggle` existente).
- Frontend: `Launcher/Store.ts` (nuevo `launchVersion` + listener), `Instances/Store.ts` (listener), `Accounts/Store.ts` (listener de sincronización) y `App.vue` (listeners de navegación).
- Ningún binding de Wails cambió.

## Comportamiento anterior/nuevo

- El tray no aparecía nunca → aparece **un solo** icono en la bandeja al arrancar y se destruye al salir.
- Clic izquierdo en el icono: nada → muestra/oculta el launcher. Clic derecho: menú vía `OpenMenu()`. Doble clic: muestra y enfoca.
- "Últimas Sesiones"/"Instancias": bloques planos deshabilitados → submenús con items "Jugar X" clicables.
- Cuenta: sin acceso desde el tray → submenú con hasta 5 cuentas y radio activo, selección persistente y sincronizada con la UI.
- Música: checkbox → botón que alterna la etiqueta.
- Nuevas acciones: Configuración, Abrir Instancias y Abrir Descargas abren sus vistas.

## Cómo verificar

- `go build ./...` en la raíz: compila sin errores. `bun run type-check` en `frontend/`: pasa.
- Ejecutar el launcher: aparece **un solo** icono en la bandeja.
- Clic izquierdo alterna la ventana, doble clic la muestra y enfoca, clic derecho abre el menú.
- "Jugar X" de los submenús lanza la partida con el feedback habitual (con el launcher visible u oculto).
- Submenú "Cuenta": cambiar la cuenta marca el radio nuevo, la UI del launcher refleja el cambio al instante y la selección persiste al reiniciar.
- "Configuración"/"Abrir Instancias"/"Abrir Descargas" abren sus respectivas vistas.
- "Pausar Música"/"Reproducir Música" alterna el item y la música del frontend.