# StepLauncher-Error-17: Rich Presence no aparecía en Discord (net.Dial("unix") no puede abrir las named pipes de Discord en Windows)

- **Fecha**: 2026-08-07
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (en desarrollo, funcionalidad Rich Presence nueva)

---

## 1. Síntoma

Con Discord abierto y la opción de presencia activada, el perfil de Discord no mostraba ninguna actividad: ni "Navegando por el menú" ni "Jugando Minecraft {version}". El launcher logueaba reintentos de conexión en bucle.

## 2. Causa raíz

La primera implementación de `internal/RichPresence` intentaba conectar a las named pipes de Discord (`\\?\pipe\discord-ipc-0` / `\\.\pipe\discord-ipc-0`) con `net.DialTimeout("unix", addr, ...)` (sockets AF_UNIX de Go, soportados desde Go 1.23 en Windows). La pipe `discord-ipc-0` SÍ existía (comprobado con `Get-ChildItem \\.\pipe\` y Discord en ejecución), pero el dial fallaba con `connect: No connection could be made because the target machine actively refused it` para las 20 candidatas (0..9 en ambos formatos).

Los sockets AF_UNIX de Go en Windows solo pueden hablar con pipes creadas también como AF_UNIX; no sirven para abrir named pipes clásicas creadas con `CreateNamedPipe` (que es como Discord crea las suyas). Por eso el handshake nunca llegaba a enviarse.

## 3. Diagnóstico y evidencia

- `Get-Process` mostraba Discord en ejecución (8 procesos) y `Get-ChildItem \\.\pipe\ | Where-Object Name -like '*discord*'` listaba `discord-ipc-0` y las pipes de `DiscordSystemHelper`: la pipe existía.
- Programa de prueba independiente (`Temp\opencode\discordtest`) probando `net.Dial("unix", ...)` contra las 20 candidatas: TODAS fallaban con "actively refused".
- El mismo programa usando `winio.DialPipe("\\.\pipe\discord-ipc-0", &timeout)` conectaba al instante: handshake respondía `{"cmd":"DISPATCH","evt":"READY",...}` y `SET_ACTIVITY` respondía `{"cmd":"SET_ACTIVITY","data":{"details":"StepLauncher","state":"Probando presencia","type":0,"name":"StepLauncher","application_id":"1438239391666405396",...}}` — la presencia se veía en Discord.

## 4. Solución aplicada

- `internal/RichPresence/RichPresence.go` (`dialDiscord`): en Windows se conecta con `winio.DialPipe` (librería oficial de Microsoft para named pipes, el mismo mecanismo que usa la librería `rich-go` de la comunidad) a `\\.\pipe\discord-ipc-0..9`; en Unix se mantiene `net.DialTimeout("unix", ...)`.
- Dependencia añadida: `github.com/Microsoft/go-winio v0.6.2` en `go.mod`.

## 5. Regla aprendida

En Windows, para conectar a named pipes de terceros (Discord IPC, etc.) usar SIEMPRE `winio.DialPipe` (go-winio); `net.Dial("unix", "\\.\pipe\...")` solo funciona contra pipes creadas como AF_UNIX y falla con "actively refused" contra pipes clásicas de `CreateNamedPipe`.

## 6. Verificación

- `go build ./...` OK (con go-winio v0.6.2).
- Programa de prueba con `winio.DialPipe`: handshake READY + `SET_ACTIVITY` aceptado y visible en el perfil de Discord.
