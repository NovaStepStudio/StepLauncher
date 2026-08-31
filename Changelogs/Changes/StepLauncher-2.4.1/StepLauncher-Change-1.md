# Cambios de StepLauncher 2.4.1 (Step-1)

- **Fecha**: 2026-08-15
- **Versión**: 2.4.1 (en desarrollo)
- **Estado**: implementado y verificado
- **Release**: en desarrollo — aún no mencionado en ninguna release.
- **Estado**: implementado y verificado.

## Qué cambió

Soporte para versiones de terceros (p. ej. BatMod) cuyos `version.json` declaran las librerías y los nativos solo con `url`/`sha1`/`size`, omitiendo el campo `path`. Antes esas versiones no descargaban ni extraían ningún nativo ("Extracted 0 native file(s)") y algunas librerías no se podían descargar ("cannot resolve download").

### 1. Normalización de paths (`internal/Core/Downloader/Normalize.go`, nuevo)

- Nueva función `NormalizeVersion(ver)` que completa `Artifact.Path` cuando falta: para `downloads.artifact` se deriva de la coordenada maven (`MavenPath(name)`); para cada `downloads.classifiers` se deriva con el clasificador (`MavenPath(name + ":" + classifier)`), produciendo `<artifact>-<version>-<classifier>.jar` en la ruta maven correcta.
- Nueva función `NativeClassifierKey(lib, os, arch)` que resuelve la key del clasificador nativo (incluido `${arch}`), usada por la resolución de nativos.
- Se llama a `NormalizeVersion` en todos los puntos que cargan un `VersionJSON`:
  - `internal/Core/Launcher/Launcher.go` (`loadVersion`, incluido el merge con `inheritsFrom`).
  - `internal/Core/Downloader/Manager.go` (`runDownload`, tras obtener el JSON del manifest).
  - `internal/Handlers/Engine/Integrity.go` (`loadVersionJSON`).
  - `internal/Core/Launcher/Instance/Verify.go` (verificación de instancias).

### 2. Defensa en profundidad en los resolutores

Aunque la normalización ya rellena los paths, los resolutores ahora derivan la ruta maven si el `path` sigue vacío, para no depender del orden de llamadas:

- `internal/Core/Launcher/Helpers/Classpath.go` (`ResolveLibraryDownload`): antes devolvía `librariesDir` (directorio raíz) como destino cuando `path` estaba vacío, por lo que `downloadMissingLibraries` no encontraba la librería en su mapa y fallaba con "cannot resolve download" (p. ej. `com.mojang:netty:1.6` y `com.mojang:realms:1.7.39`). Ahora usa `MavenPath(name)`.
- `internal/Core/Launcher/Helpers/Natives.go` (`resolveNativeJar` y `ResolveNativeJarDownload`): derivan el path del clasificador cuando falta; antes devolvían vacío y la extracción terminaba con 0 archivos.
- `internal/Core/Downloader/Tasks.go` (`addLibraryTasks` y `addNativeTasks`): mismo fallback maven para el destino de descarga (antes `addNativeTasks` generaba el jar sin el sufijo del clasificador, apuntando a un archivo que no existe).

### 3. Decodificador de logs XML de log4j2 (`internal/Core/Launcher/Utils/Logdecoder.go`, nuevo)

- Nuevo `Log4j2XMLWriter`: convierte a texto plano el output del juego que llega en formato XML de log4j2 (layout `log4j2-xml` usado por BatMod y otras versiones de terceros). Cada evento `<log4j:Event timestamp=... level=... thread=...><log4j:Message><![CDATA[...]]></log4j:Message></log4j:Event>` se emite como `[HH:mm:ss] [LEVEL] [thread] mensaje`; las cabeceras del stream (`<?xml...`, `<log4j:eventSet...`, `</log4j:eventSet>`) se descartan y cualquier otra línea se reenvía tal cual (las versiones vanilla con log4j2 de texto plano no se ven afectadas).
- `internal/Core/Launcher/Utils/Process.go` (`LaunchProcess`): stdout y stderr del juego pasan ahora por el writer decodificador; se devuelve también el writer para poder vaciarlo al final.
- `internal/Core/Launcher/Launcher.go` (`waitForExit`): hace `Flush()` del decodificador antes de cerrar el archivo del log, para no perder la última línea o un evento XML incompleto.

## Por qué

El `version.json` de BatMod (y otros de terceros) declara `downloads` sin el campo `path`, y el launcher asumía que siempre venía. Eso rompía la descarga de librerías (netty-1.6, realms-1.7.39 quedaban en `[MISSING]`) y la descarga/extracción de los jars nativos (lwjgl-platform, jinput-platform, twitch-platform, twitch-external-platform), dejando la carpeta `natives` vacía y el juego muriendo con `UnsatisfiedLinkError`. Además, el output XML de log4j2 de BatMod se volcaba crudo al log del juego, haciéndolo ilegible.

## API afectada

- Nuevas funciones exportadas en `internal/Core/Downloader`: `NormalizeVersion`, `NativeClassifierKey`.
- `LaunchProcess` (paquete `internal/Core/Launcher/Utils`) devuelve un tercer valor `*Log4j2XMLWriter` (solo lo usa `Launcher.go`).
- Sin cambios en bindings de Wails ni en el frontend.

## Comportamiento anterior/nuevo

- BatMod: "Extracted 0 native file(s)" y `[MISSING]` en `com.mojang:netty:1.6` / `com.mojang:realms:1.7.39` → descarga las librerías y los 4 jars nativos de Windows y los extrae a `versions/BatMod/natives` (lwjgl 613 KB, jinput 155 KB, twitch 580 KB, twitch-external 7,4 MB).
- Log del juego de BatMod: XML crudo de log4j2 → líneas legibles `[HH:mm:ss] [LEVEL] [thread] mensaje`; las versiones con log4j2 en texto plano se escriben igual que antes.

## Cómo verificar

- `go build ./...` y `go test ./internal/Core/Downloader/... ./internal/Core/Launcher/Utils/...` (tests nuevos: `Normalize_test.go` y `Logdecoder_test.go`).
- Lanzar BatMod: el launcher debe descargar `netty-1.6.jar`, `realms-1.7.39.jar` y los 4 jars nativos, y loguear "Extracted N native file(s)"; el log del juego debe verse como texto plano con `[INFO] [Client thread] ...`.
- Lanzar una versión vanilla (p. ej. 1.8.9) y comprobar que descarga y extrae nativos igual que antes (14 archivos).
