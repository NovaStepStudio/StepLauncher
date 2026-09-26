# Cambios de StepLauncher 2.5.0 (Change-2) — Música de fondo: reproducción fiable, volumen en Ajustes y carátula como fondo

- **Fecha**: 2026-08-17
- **Versión**: 2.5.0 (en desarrollo)
- **Estado**: implementado y verificado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

Tanda de mejoras de la música de fondo pedida por el usuario: la música se importaba pero **no se reproducía** en la app exportada, faltaba un controlador de volumen (debe vivir en Ajustes, no en el widget) y el estilo de carátula "Fondo del widget" se rediseñó por completo.

### 1. Fix de reproducción: `<audio>` nativo + data URI base64 (regresión del fix 2.4.1)

- **Causa raíz**: el changelog 2.4.1 (Change-3) documentó con evidencia real que en WebView2 los **blob URLs en el scheme `wails://` cargan pero pueden quedarse mudos** ("el estado marca 'reproduciendo' en falso") y el fix verificado fue servir el audio como **data URI base64** con su MIME real. Al migrar a Wails v3 (2.5.0 Change-1), `Common/Stores/Music.ts` se reescribió con **Howler.js (Web Audio API) + blob URLs**, regresando al mecanismo roto: `decodeAudioData`/XHR del blob falla en el scheme `wails://`, `onloaderror` solo hacía `console.warn` y el usuario pulsaba play sin que sonara nada.
- **Solución** (`Common/Stores/Music.ts`):
  - Se eliminó Howler.js: la reproducción vuelve al **`<audio>` nativo** (el motor verificado en 2.4.1), creado con `new Audio()` fuera del DOM para que Vue nunca lo destruya (sobrevive a `v-show`, Teleport y al ciclo de vida del widget).
  - La fuente ya no es un blob URL: `musicDataUri()` construye el **data URI base64** pegando directamente la cadena base64 que entrega `ReadLocalFile` (sin `atob`/reconversión pesada), con cache de 2 pistas en memoria.
  - Reintentos: `wantsPlay` + `canplay` + timeout de 800 ms si el navegador bloquea el primer `play()`.
  - **Feedback visible**: nuevo estado `loadError` en el store; el widget muestra un aviso con el motivo si una pista no se puede leer o cargar (antes el fallo era silencioso).
  - Se arregló el check muerto de `loadMusicList` (comparaba la lista nueva contra sí misma; ahora contra la anterior) y el bucle nativo (`audio.loop`) sigue funcionando con 1 pista o modo "repetir esta pista".
  - Media Session (SMTC), modos de reproducción, seek y metadatos se conservan intactos.
  - Dependencia eliminada: `howler` y `@types/howler` salen de `frontend/package.json` (verificado con grep: solo `Music.ts` los usaba) y del lockfile con `bun install`.

### 2. Validación de importación con el motor real

- `validateAndReadMusic` ya no depende de `music-metadata` para validar: los bytes del archivo se convierten a un data URI y se miden con un `<audio>` temporal (`measureAudioDuration`, el MISMO motor que reproduce). Así se valida de verdad que la pista **sonará en WebView2** y la duración real (≤ 10 min) sale del navegador, no de un parser de metadatos.
- Un archivo que el `<audio>` no pueda decodificar se rechaza con "El archivo no parece un audio válido para StepLauncher" (antes se importaba y luego no sonaba, el síntoma exacto del usuario).
- `music-metadata` se conserva solo para `refreshMetas` (título, artista y carátula en el widget).

### 3. Fix: doble extensión al importar con mayúsculas (`internal/Core/Assets/Assets.go`)

- `AddMusic` recortaba la extensión con `strings.TrimSuffix` (case-sensitive): un nombre enviado como `Song.MP3` se copiaba como `Song.MP3.mp3`. Ahora el recorte se hace por longitud sobre la extensión ya normalizada a minúsculas.

### 4. Controlador de volumen en Ajustes → Música de fondo (no en el widget)

- El store ya tenía `setVolume` (0-100 %, persistido en localStorage) pero **ninguna UI lo exponía**. El usuario lo pidió explícitamente en la configuración, no en el widget (el widget no lleva volumen y sigue sin llevarlo).
- `internal/Config/Config.go`: `MusicConfig` gana `Volume float64 json:"volume"` (default `0.8`, sanitizado a 0-1 en `sanitize()`).
- `Common/Stores/Ui.ts`: `MusicConfig.volume` + normalización (fallback 0.8).
- `Common/Stores/Music.ts`: la fuente de verdad del volumen es la personalización global (el localStorage se conserva como respaldo de versiones anteriores) y un `watch` aplica en vivo el volumen de la config al `<audio>`.
- `Settings/Sections/Personalization.vue`: nueva fila "Volumen" con slider (0-100 %, paso 5) que aplica el volumen en vivo al arrastrar y lo persiste en la personalización al soltar (`@input` → `setVolume`, `@change` → `save()`). Estilos nuevos `SsVol`/`SsVolRange`/`SsVolVal` en `Settings/Styles/Personalization.scss` (range con `accent-color: var(--progress-color)`, mismo patrón que el seek del widget).

### 5. Rediseño del estilo de carátula "Fondo del widget"

El modo `coverStyle: 'background'` ahora usa la carátula **dos veces** (`Common/Widgets/BackgroundMusic.vue` + `Styles/BackgroundMusic.scss`):

- **Fondo de todo el widget**: la carátula ocupa el 100 % del ancho y alto de la card (`position: absolute; inset: 0`), con `object-fit: cover`, **`filter: blur(10px) saturate(1.35)`** para que "pegue" con el resto y una **animación de paneo vertical de arriba hacia abajo** muy lenta y suave (`bgm-bg-pan`, 30 s, `ease-in-out`, `infinite alternate`), con escala extra `scale(1.14)` para que el desplazamiento nunca deje ver los bordes. Se respeta `prefers-reduced-motion`.
- **Veladura de legibilidad**: un degradado sutil (22 % → 48 %) del color modal del launcher (`--background-modal-primary`) mantiene el texto y los controles legibles sin tapar la carátula.
- **Carátula completa en recuadro**: nuevo `.Bgm_Cover--bg` (6 rem, esquinas redondeadas, sombra del launcher + borde derivado de `--control-border`) que muestra la imagen completa sin blur, para poder verla bien; si no hay carátula, icono de música como en los otros estilos.

## Por qué

El usuario reportó que la música se importaba pero no se reproducía (regresión documentada del fix de 2.4.1 al migrar a v3), pidió un controlador de volumen dentro de la configuración de la música —no en el widget— y una estética más vistosa para la carátula como fondo: dos imágenes (una de vista completa y otra como fondo con paneo lento, blur y saturate).

## API afectada

- `MusicConfig.Volume float64` (config global `backgroundMusic.volume`).
- Store `Common/Stores/Music.ts`: motor interno reescrito (Howler → `<audio>` nativo + data URIs); API pública intacta + nuevo estado `loadError`.
- Dependencias frontend: eliminadas `howler` y `@types/howler`.
- Sin cambios en bindings de Wails.

## Comportamiento anterior/nuevo

- Reproducción: pulsar play no sonaba nada (fallo silencioso en WebView2 con blob URLs) → el `<audio>` nativo con data URI base64 suena en cualquier scheme y muestra el motivo si algo falla.
- Importación: un audio que no se pudiera decodificar se importaba igual → se valida con el motor real antes de importar (decodificabilidad + duración ≤ 10 min).
- Importar `Song.MP3` → `Song.MP3.mp3` → `Song.mp3`.
- Volumen: existía en el store sin UI → slider en Ajustes → Música de fondo, persistido en la personalización global.
- Carátula "Fondo del widget": imagen 1:1 anclada a la derecha con máscara → fondo completo del widget con blur + saturate + paneo vertical lento y recuadro con la carátula completa.

## Cómo verificar

- `go build ./...` (raíz): compila sin errores.
- `bun run type-check` en `frontend/`: pasa sin errores (el usuario trabaja en dev; no se ejecuta `bun run build` mientras `wails dev` esté activo para no cortar el proxy).
- En Ajustes → Personalización → Música de fondo: añadir un MP3, activar la música y reproducir desde el widget — debe sonar (incluida la app exportada). Quitar el archivo de `cache/audio` a mano y pulsar play → el widget muestra el aviso de error.
- Arrastrar el slider de volumen con música sonando → el nivel cambia en vivo; reiniciar el launcher → el volumen persiste.
- Estilo "Fondo del widget": la carátula cubre toda la card con blur + saturate y paneo vertical lento, y el recuadro muestra la carátula completa.