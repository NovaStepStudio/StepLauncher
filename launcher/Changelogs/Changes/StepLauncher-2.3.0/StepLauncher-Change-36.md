# StepLauncher-Change-36.md

## Fecha
2026-08-07

## Release
StepLauncher-2.3.0 — se mencionó por primera vez en esta release.

## Cambio
Nueva funcionalidad: **Presencia de Discord (Rich Presence)**. Un apartado nuevo en `internal/` (`internal/RichPresence`) que maneja la presencia del launcher en Discord: en el menú muestra "Navegando por el menú", al lanzar "Lanzando Minecraft" y al jugar "Jugando Minecraft" con el texto "StepLauncher {versión}" y contador de tiempo. Solo texto, sin imágenes (Discord muestra el icono de la app automáticamente). Configurable desde Ajustes → Comportamiento ("Presencia en Discord"). Multiplataforma: Windows (named pipes con go-winio), Linux/macOS (socket AF_UNIX, incluyendo Snap/Flatpak).

## Por qué
Poder mostrar en el perfil de Discord la actividad del launcher y del juego, como hacen la mayoría de launchers.

## Qué se hizo
1. **Paquete nuevo `internal/RichPresence/RichPresence.go`**:
   - Cliente del protocolo IPC local de Discord implementado directamente (frames opcode+payload JSON): handshake con `client_id`, `SET_ACTIVITY`, ping/pong y detección de cierre de pipe.
   - En Windows conecta a `\\.\pipe\discord-ipc-0..9` con `winio.DialPipe` (los sockets AF_UNIX de Go no pueden abrir las named pipes de Discord); en Unix usa socket AF_UNIX (ver Error-17).
   - `Manager` con bucle de reconexión en segundo plano (Discord puede no estar abierto): guarda la actividad deseada y la reenvía al reconectar; nunca bloquea al llamante (mutex + goroutine, sin locks reentrantes).
   - `SetEnabled` (desactivar limpia la presencia al momento), `SetActivity(details, state, start)`, `Close`.
   - App ID de Discord Developer: `1438239391666405396` (`DiscordClientID`).
2. **`internal/Config/Config.go`**: nueva sección `richPresence.enabled` como `*bool` (nil → activado, patrón de `VerifyIntegrity`), `EnabledValue()`, `SetRichPresenceEnabled`, default en `Default()` y `sanitize`.
2. **`internal/Config/Config.go`**: nueva sección `richPresence.enabled` como `*bool` (nil → activado, patrón de `VerifyIntegrity`), `EnabledValue()`, `SetRichPresenceEnabled`, default en `Default()` y `sanitize`.
3. **`internal/Handlers/App.go`**: manager creado en `NewApp` con log al logger del engine; `Startup` activa según config y muestra el estado de menú; `SetEventCallback` ahora envuelve el callback para traducir los eventos de juego del motor a la presencia (`game_starting` → "StepLauncher {version}" / "Lanzando Minecraft", `game_started` → "StepLauncher {version}" / "Jugando Minecraft" con timestamp, `game_exited`/`game_crashed`/`game_stopped` → menú, comprobando si queda otra partida abierta); bindings `GetRichPresenceConfig` y `SetRichPresenceEnabled`; `Shutdown` cierra el manager.
4. **`app.go` (raíz)**: bindings delegados `GetRichPresenceConfig`/`SetRichPresenceEnabled` y cierre del handler en `shutdown`.
5. **Frontend `GeneralSettings.vue`**: toggle "Presencia en Discord" en Ajustes → Comportamiento (carga desde `cfg.richPresence?.enabled ?? true` y guarda con `SetRichPresenceEnabled`).

## API afectada
- Bindings Wails nuevos: `App.GetRichPresenceConfig()`, `App.SetRichPresenceEnabled(v bool)` (se regeneran con `wails build`/`wails dev`).
- `launcher_config.json`: nueva sección `richPresence` (ausente → activada por defecto).
- Dependencia nueva: `github.com/Microsoft/go-winio v0.6.2` (conexión a named pipes en Windows).

## Cómo verificar
- `go build ./...` OK y `bun run build` (type-check) OK.
- Con Discord abierto y la opción activada: al abrir el launcher aparece "StepLauncher" / "Navegando por el menú"; al pulsar Jugar, "StepLauncher {version}" / "Lanzando Minecraft" y luego "StepLauncher {version}" / "Jugando Minecraft" con contador de tiempo; al cerrar el juego vuelve al menú.
- Ajustes → Comportamiento: desactivar "Presencia en Discord" limpia la presencia al instante; reactivarla la vuelve a mostrar.
- Si Discord no está abierto, el launcher sigue funcionando y la presencia aparece al abrir Discord (reintentos en segundo plano).
- En Linux/macOS: compilar para esas plataformas; el socket se busca en `XDG_RUNTIME_DIR`, `TMPDIR`, `/tmp` y rutas Snap/Flatpak de Discord.

## Notas
- Presencia solo con texto: Discord muestra automáticamente el icono de la app registrado en Discord Developer Portal (app `1438239391666405396`); no se usan assets de imagen.
- Durante el desarrollo se evaluó la librería `rich-go` (ver Error-17): internamente implementa exactamente el mismo protocolo IPC (pipe `discord-ipc-0` + frames `SET_ACTIVITY`), pero sin reconexión, sin leer respuestas y con socket global; la implementación propia añade reconexión automática, ping/pong y validación, manteniendo las mismas rutas multiplataforma.
- El `*bool` de `richPresence.enabled` evita que una config guardada sin el campo se interprete como "desactivado".
