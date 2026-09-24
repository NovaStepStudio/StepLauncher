# Bugs de StepLauncher 2.5.0 (Bug-4) — Proxy manual solo para Minecraft + transporte del launcher siempre directo sin HTTP/2

- **Fecha**: 2026-09-24
- **Versión**: 2.5.0
- **Estado**: corregido y verificado
- **Release**: en desarrollo — aún no mencionado en ninguna release.

## El bug en cuestión

Regla de arquitectura del owner: el proxy de Ajustes > Red es **SOLO para el juego de Minecraft** y nunca debe afectar a peticiones del launcher. Pero el cliente HTTP manual se aplicaba a descargadores y metadatos de modloaders, y una docena de clientes sueltos (`News`, `Auth`, `Updater`, `Gallery`, actualizador, `hasInternet`, descargas de librerías con `http.DefaultClient`) heredaban el proxy del sistema por usar el transporte por defecto de Go. Además, el error `malformed HTTP status code` con bytes binarios apunta a una capa de inspección transparente y/o a la negociación HTTP/2, así que el transporte del launcher pasa a HTTP/1.1 explícito.

## Qué afectaba y qué hacía

- Cualquier proxy (manual o del sistema) podía contaminar manifiesto, metadatos, noticias, auth, updater y galería con respuestas no-HTTP.
- `http.DefaultClient` en descargas de librerías (`internal/Core/Launcher/Launcher.go`) heredaba `HTTP_PROXY` del entorno sin que el usuario lo supiera.

## Solución final

- **`internal/Core/Downloader/Download.go`**: `DefaultTransport` con `Proxy: nil` explícito, `ForceAttemptHTTP2: false` y `TLSNextProto` vacío (sin h2 por ALPN). `NewConfiguredHTTPClient` ya no aplica proxy manual (firma intacta, parámetros ignorados): solo aplica el límite de velocidad. Se eliminó el bloque de parseo SOCKS/HTTP (~130 líneas) y sus imports.
- **`internal/Core/Downloader/Rate.go`**: fallback a `DefaultTransport` en vez de `http.DefaultTransport`.
- **`internal/Handlers/Engine/Config.go` (`applyNetworkConfig`)** y **`internal/Handlers/Engine/Launch.go`**: comentarios corregidos (proxy solo vía flags JVM). El juego sigue recibiendo `-Dhttp.proxyHost`/`-DsocksProxyHost` en `internal/Core/Launcher/Launcher.go` sin cambios.
- **Clientes migrados al transporte directo** (mismos timeouts, solo cambia el `Transport`): `News.go`, `Auth/Injector.go`, `Auth/Authlib.go` (4 sitios), `Updater/worker_provider.go`, `Handlers/Gallery.go`, `Handlers/Engine/Update.go`, `Launcher.go` (`hasInternet`, pre-verificación de auth y 3 descargas de librerías que usaban `http.DefaultClient`).
- **`frontend/web/src/Settings/Sections/Network.vue`**: etiqueta y descripción ("Usar proxy solo para el juego — Solo lo usa Minecraft al lanzarse. El launcher siempre va en directo"). Sin cambios de estilos.
- Verificado: ya no queda ningún `http.DefaultClient`, `http.DefaultTransport` ni `&http.Client{Timeout...}` sin transporte explícito en todo `launcher/`.

## Comportamiento anterior/nuevo

- Antes: el proxy manual y el del sistema podían colarse en cualquier petición del launcher; HTTP/2 negociado por defecto.
- Ahora: todo el launcher va en directo por HTTP/1.1; el proxy manual solo llega a la JVM del juego.

## Verificación

- `go build ./...` en `launcher/` — OK.
- `go vet ./...` en `launcher/` — OK.
- `bun run build` en `launcher/frontend` — OK.
