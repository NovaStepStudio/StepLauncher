# Cambios de StepLauncher 2.4.1 (Change-3) — Música de fondo global

- **Fecha**: 2026-08-16
- **Versión**: 2.4.1 (en desarrollo)
- **Estado**: implementado y verificado
- **Release**: en desarrollo — aún no mencionado en ninguna release.

## Qué cambió

Nueva opción de personalización global "Música de fondo": el usuario puede subir audios (MP3/WAV/OGG/M4A, máx. 10 minutos y 15 MB, hasta 5 pistas), se guardan en `cache/audio` y se registran en `launcher_assets.json` con la estructura `"audio": [{ "name": "...", "path": "..." }]`. Un widget flotante la reproduce en segundo plano con carátula y metadatos extraídos con `music-metadata`.

### 1. Backend: pistas de audio en launcher_assets.json (`internal/Core/Assets/Assets.go`)

- Nuevo `MusicSlot { name, path }` y campo `Assets.Music []MusicSlot json:"audio"` (estructura exacta pedida).
- Constantes `CacheDir = "cache"`, `AudioSubDir = "audio"`, `MaxMusicTracks = 5` y `MaxAudioBytes = 15 MB`.
- `AudioDirOf(rootDir)` → `<root>/cache/audio`; `IsAudioExt` valida `.mp3/.wav/.ogg/.m4a`.
- `normalizeMusic`: conserva solo pistas con nombre y ruta relativa `cache/audio/...`, sin duplicados (convive con la normalización de fuentes).
- `ListMusic()`, `AddMusic(name, src)` (valida extensión, tamaño ≤ 15 MB y máx. 5 pistas; copia con nombre sanitizado y deduplicado; registra la ruta relativa) y `RemoveMusic(name)` (borra del JSON y del archivo si no lo comparte otra pista).
- La copia es atómica con respecto al JSON: si `Save` falla se elimina el archivo copiado.

### 2. Backend: bindings (`internal/Handlers/Music.go` + wrappers en `app.go`)

- `PickMusicFile()`: diálogo nativo con filtro `*.mp3;*.wav;*.ogg;*.m4a`.
- `ListMusic() []assets.MusicSlot`.
- `AddMusic(name, src)` → ruta relativa.
- `RemoveMusic(name)`.
- `ReadMusicFile(src)`: lee el archivo elegido (solo extensiones de audio, ≤ 15 MB) y devuelve sus bytes para que el frontend extraiga metadatos y duración con `music-metadata` antes de importar.
- Bindings regenerados con `wails generate module`.

### 3. Backend: config global (`internal/Config/Config.go`)

- `Personalization.BackgroundMusic MusicConfig { enabled bool, position string }` con `position` ∈ `top-left | bottom-center`; default `{false, "top-left"}` y sanitizado en `sanitize()`.

### 4. Frontend: store y metadatos (`Common/Stores/Music.ts`, nuevo)

- Lista `musicList`, índice actual, `playing`, `volume` (persistido en localStorage, máx. 100%), tiempo, URL de la pista (`currentSrc`) y `metaCache` con los metadatos por pista.
- `validateAndReadMusic`: valida extensión, tamaño (≤ 15 MB) y duración (≤ 10 min) con `music-metadata` (`parseBlob`) sobre los bytes de `ReadMusicFile`.
- `refreshMetas`: extrae título/artista/duración y la carátula embebida (objectURL) de cada pista ya registrada.
- `loadMusicList/addMusic/removeMusic/playTrack/togglePlay/next/prev/setVolume/seekTo/stopMusic`; al desactivar la música desde Ajustes se detiene la reproducción.
- `Common/Stores/Ui.ts`: `Personalization.backgroundMusic` + normalización; `mimeOf` ahora mapea `mp3/wav/ogg/m4a` a sus MIME de audio (clave para que el `<audio>` reproduzca los blobs de `ReadLocalFile`).

### 5. Frontend: widget flotante (`Common/Widgets/BackgroundMusic.vue` + SCSS)

- **Arriba a la izquierda**: panel fijo siempre visible con carátula, título/artista, barra de progreso clicable, controles (anterior/reproducir-pausar/siguiente) y volumen (0–100%).
- **Abajo en el centro**: pequeño botón circular (♪) que despliega y oculta el panel; si no hay pista seleccionada muestra un aviso con botón de cierre.
- `<audio>` global manejado por el store: al terminar la pista avanza automáticamente a la siguiente; si se pulsa Anterior con más de 3 s reproducidos, reinicia la pista.
- Se monta desde `App.vue` solo si la personalización lo activa y existe al menos una pista.

### 6. Frontend: sección en Ajustes (`Settings/Sections/Personalization.vue`)

- Nuevo grupo "Música de fondo": switch global activar/desactivar, selector de posición del widget, botón "Añadir audio" (diálogo nativo + validaciones), lista de pistas (carátula, nombre real del metadato, duración) con botón Quitar, y avisos de error.

## Por qué

El usuario pidió una opción de personalización global activable/desactivable para poner un audio de fondo en el launcher, con requisitos estrictos (formato, duración y peso), almacenamiento en `cache/audio` + `launcher_assets.json` (estructura `"audio"`), máx. 5 pistas, widget con dos posiciones (top-left fijo / bottom-center plegable), volumen máx. 100% y controles mínimos (siguiente/anterior), además de metadatos y carátula para darle estilo al widget.

## API afectada

- `Assets` gana `Music []MusicSlot json:"audio"`; nuevos métodos `ListMusic/AddMusic/RemoveMusic` y helpers `AudioDirOf/IsAudioExt`.
- `Personalization.BackgroundMusic` (config global).
- Bindings Wails nuevos: `PickMusicFile`, `ListMusic`, `AddMusic`, `RemoveMusic`, `ReadMusicFile`.
- Dependencia frontend nueva: `music-metadata` (extracción de metadatos y duración en el navegador).

## Comportamiento anterior/nuevo

- Antes no existía música de fondo → ahora: subir hasta 5 audios (MP3/WAV/OGG/M4A, ≤ 10 min, ≤ 15 MB) desde Personalización, que se guardan en `cache/audio/` y se registran en `launcher_assets.json` bajo `"audio"`.
- Widget flotante con carátula (si el audio la lleva), título/artista reales, progreso, anterior/siguiente/reproducir-pausar y volumen hasta 100%.
- La música se detiene si se desactiva la opción y al eliminar la pista en reproducción.

## Ajustes posteriores (rediseño)

- El widget ya no vive en `App.vue`: se auto-monta en el body desde `Main.ts` (host `#bgm-host`) y decide por sí mismo cuándo mostrarse (personalización activada + al menos una pista). `App.vue` queda sin líneas de música.
- Fix de reproducción: el `<audio>` estaba solo en la variante arriba-izquierda, por eso no sonaba en abajo-centro; ahora vive en la raíz del widget y siempre está montado cuando el widget se muestra.
- La lista de pistas se carga al montar el widget (arranque de la app), así el widget aparece con una sola pista y sin abrir Ajustes; al arrancar se selecciona la primera pista sin reproducirla.
- **Posición única**: el widget solo vive abajo en el centro (botón que despliega/oculta la card). Se eliminó la posición arriba-izquierda (backend + Ajustes + widget).
- **Rediseño visual**: paleta con degradado derivado del color personalizado del launcher (`--color-tag`) hacia rosa vibrante; componente más grande (carátula 4.1 rem, card 25 rem, controles mayores); animación de apertura con rebote (`bgm-in`); z-index moderado (25) para quedar sobre la UI pero bajo modales (100+).
- **Carátula configurable**: estilo "Disco" (gira siempre, incluso en pausa, si la rotación está activada) o "Cuadrado"; ambos configurables desde Ajustes → Música de fondo (selector de estilo + switch de rotación).
- **Ocultación global**: el widget entero (incluido el botón de abrir) se oculta al dispararse `CLOSE_OVERLAYS_EVENT` (idle que cierra overlays) y reaparece con la próxima interacción.
- **Controles según cantidad de pistas**: con 1 pista solo play/pausa; con 2+ se añaden anterior/siguiente; con 3+ aparecen los modos de reproducción (repetir cola, aleatorio, repetir esta pista en bucle sin avanzar).
- **Sin volumen**: se eliminó el controlador de volumen del widget (se mantiene el estado interno en el store, no visible en la UI).
- El botón de abrir reacciona a la música: ecualizador mini animado cuando suena, halo pulsante y escalado al pasar el ratón.
- **Estilo 100 % variables**: el widget usa exclusivamente las variables de personalización del launcher (colores, fondos, bordes, transiciones, tipografías y `--shadow-settings-normal` con color explícito como el resto del proyecto); sin colores fijos propios.
- **Fix barra de progreso**: el contenedor pasó de `v-if` a `v-show`, así el `<audio>` (y sus handlers de tiempo) ya no se destruye al ocultarse el widget por idle; la barra y el contador avanzan en vivo.
- **Fix botón de abrir**: sombras con color explícito (nada de glow blanco), fondo `--background-button-primary` y transición `out-in` en el desplegable para que el botón y la card no se crucen al alternar.
- **Animación de apertura**: deslizamiento vertical (slide-up) con fade, sin scale ni popup; `mode="out-in"`.
- **Nuevo estilo de carátula "Fondo"**: la carátula se convierte en el fondo del widget, en proporción 1:1 anclada a la derecha con altura 100 %, recortada con `overflow: hidden` y una máscara que la desvanece de derecha (visible) a izquierda (transparente).
- **Bucle infinito con una sola pista**: con 1 pista (o modo "repetir esta pista") el `<audio>` usa `loop` nativo; nunca avanza.
- **Reproducción robusta**: el cambio de pista ahora es limpio (`pause → src → load → play`), se reintenta automáticamente cuando el recurso está listo (`oncanplay`, flag `wantsPlay`) en lugar de esperar el poll de 2 s, y la URL del blob anterior se revoca con 1.5 s de retraso para no romper la decodificación de la pista que se estaba reproduciendo.
- **Salto automático al cambiar de modo**: al activar "Repetir cola" o "Aleatorio" el reproductor pasa automáticamente a la siguiente pista y la reproduce (aleatorio → pista al azar, distinta de la actual).
- **Apertura más rápida**: la transición de apertura se acortó (salida 0.12 s + entrada 0.22 s).
- **Fix audio en la app exportada (WebView2)**: el `<audio>` ya no usa blob URLs (en el scheme `wails://` cargan pero pueden quedarse mudos y el estado marca "reproduciendo" en falso); ahora el sonido se sirve como **data URI base64** con su MIME real (`musicDataUri`, con cache de 2 pistas), fiable en cualquier scheme. Además `setTrack` hace un cambio limpio (`pause → src → load → play`) y `wantsPlay` + `oncanplay` reintentan la reproducción en cuanto el recurso está listo.
- **Controles multimedia del sistema (Media Session / SMTC)**: el reproductor se integra con `navigator.mediaSession` — Windows reconoce el reproductor con pausar/reproducir y anterior/siguiente, y recibe la metadata del launcher (título, artista, álbum y carátula). La metadata se actualiza al cambiar de pista, al extraerla con music-metadata y al pausar/reanudar.

## Cómo verificar

- `go build ./...` (raíz): compila sin errores.
- `wails generate module` + `bun run build` (frontend): type-check + build correctos.
- En Ajustes → Personalización → Música de fondo: activar el switch, añadir un MP3 con carátula (debe pasar las validaciones de formato/duración/peso; un archivo > 15 MB o > 10 min debe rechazarse) y comprobar que aparece en la lista con carátula, título y duración.
- Verificar `cache/audio/<archivo>` y la entrada `"audio": [...]` en `launcher_assets.json`.
- Comprobar el widget en ambas posiciones (arriba-izquierda y abajo-centro con botón desplegable), controles anterior/siguiente, volumen al 100% y avance automático al terminar una pista.