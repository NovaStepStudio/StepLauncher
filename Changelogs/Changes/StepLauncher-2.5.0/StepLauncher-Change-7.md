# Cambios de StepLauncher 2.5.0 (Step-7) — Panel de Música reconstruido: música REAL, selector visual y estilos completos

- **Fecha**: 2026-08-29
- **Versión**: 2.5.0
- **Estado**: implementado y verificado
- **Release**: en desarrollo — aún no mencionado en ninguna release.

## Qué cambió

Se reconstruyó por completo el dominio `Music/` del frontend y su integración con el backend, porque el panel estaba a medias: vistas que no recibían props, SCSS vacíos, playlists basadas en escribir rutas a mano, y reproducción ligada a la música de fondo (máx. 5 pistas, 15 MB) en lugar de la biblioteca real del usuario.

### 1. Nuevo motor de reproducción local (`PlayerStore.ts`) — música REAL

**Archivo nuevo**: `frontend/web/src/Music/PlayerStore.ts` (inglés, mayúscula inicial).
- Estado global separado de `Common/Stores/Music.ts` (música de fondo). Usa un único `HTMLAudioElement`, `Blob URLs` creadas desde `ReadAbsoluteFile` (base64 → `Uint8Array` → `Blob` → `URL.createObjectURL`), caché LRU de 6 URLs con `URL.revokeObjectURL`.
- Maneja `currentTrack`, `queue`, `currentIndex`, `playing`, `currentTime`, `duration`, `volume`, `playMode` (`queue` | `shuffle` | `repeat-one`), `shuffledOrder`, `loadError`, `isLoadingTrack`.
- Cola barajada persistente (igual que la música de fondo pero para biblioteca real), `syncLoop`, `MediaSession` (`play`/`pause`/`next`/`prev`), persistencia de volumen en `localStorage` (`stl_local_music_volume`) y `watch` para sincronizar `audio.volume`.
- API: `setQueue`, `playTrack(track, queueOverride?)`, `playPath`, `togglePlay`, `next`, `prev`, `seekTo`, `setVolume`, `setPlayMode`, `stop`, `progressPct`, `displayCover`, `displayMeta`, `queueTracks`.

### 2. Selector visual de pistas (`TrackSelector.vue` + `TrackSelector.scss`)

**Archivos nuevos**: `frontend/web/src/Music/Components/TrackSelector.vue` y `frontend/web/src/Music/Styles/TrackSelector.scss`.
- Reemplaza el `textarea` de rutas. Muestra la música REAL que ya tiene el usuario (`localTracks` de `LocalStore.ts`), con búsqueda en vivo (título/artista/archivo/ruta), checkboxes, carátulas, duración, paginación (50 por página), selección por filtro y limpieza.
- Acciones: “Todos filtrados”, “Quitar filtrados”, “Limpiar”. Badge de `seleccionadas · coinciden`. Empty states diferenciados (sin biblioteca vs sin resultados).
- Estilos completos con `var(--control-bg)`, `var(--progress-color)`, hover, estado `selected`, grid de paginación.

### 3. `PlaylistForm.vue` reescrito — sin escribir rutas

- Antes: `textarea` “una ruta por línea”.
- Ahora: formulario con título, tarjetas toggle `Favorita`/`Anclada`, color y carátula custom con preview, y `TrackSelector` embebido. Validación: título requerido y al menos una pista seleccionada. Modo `create`/`edit` (props `initial*` + `mode`). Mensajes `error`/`warn`/`ok` según estado de `localTracks`.
- Mantiene `useOverlayEscape` (prioridad 2) y transición `PlaylistForm`.

### 4. `Music.vue` orquestador corregido

- Antes pasaba props incorrectas (p. ej. `MenuView` esperaba `displayMeta`, `totalHits`… pero recibía `cover-url`/`cover-bg-style`), y `PlayerBar` se renderizaba sin props.
- Ahora orquesta `LocalStore` + `PlaylistStore` + `PlayerStore`, expone `displayCover` + `paletteToCss`, gestiona `section` (`menu`|`musica`|`biblioteca`|`ahora`|`cola`), `showPlaylistForm`, `headerSub` (escaneando / sin pistas / reproduciendo), acciones de cabecera (“Nueva playlist” + chips de estado) y `PlayerBar` con props reales (`coverUrl`, `title`, `artist`, `currentTime`, `duration`, `progressPct`, `playing`, `volume`, `playMode`) y eventos (`prev`/`toggle`/`next`/`seek`/`setVolume`/`setMode`).
- `handleCreatePlaylist` crea y, si hay `favorite`/`pinned`/`color`/`cover`, hace `updatePlaylist` inmediato y cambia a `biblioteca`.

### 5. Vistas reescritas y autocontenidas (cada una lee sus stores)

- **`MusicView.vue`**: búsqueda + `pageSize` (20/50/100) + paginación + filas con cover/duración/botón play-pause, estado `active` si es `currentTrack`, `MusicAlert` para `scanError`, `MusicLoading` con dot pulsante, empty states con CTA “Escanear ahora”, `loadLocalLibrary` al montar. `onPlay` usa `playTrack(t, filtered)`.
- **`LibraryView.vue`**: hero con `bibCover`/`bibPalette` (usa `useCoverPalette`), grid de playlists (`MusicBibPl` con badges ★/📌, cover o grid 2×2 de 4 carátulas), `MusicToolbar` de búsqueda dentro de playlist, filas con `onPlay` usando `trackMap`, paginación propia, edición/eliminación, `PlaylistForm` en modo `edit` reutilizado.
- **`MenuView.vue`**: hero con `coverBgStyle` + `palette`, stats (`localTracks.length`, `playlists.length`, `queue.length`, `Offline`), CTA `Escaneando…`, `MusicEmptyHero` si no hay pistas, `recent` (últimas 6) y `volver` (cola o primeras 6) con `playTrack`.
- **`NowPlayingView.vue`**: layout `MusicNow_Huge` con `cover` + `palette`, barra de progreso clickeable + `range` oculto, controles `prev`/`toggle`/`next`, cola lateral `queueTracks.slice(0,14)` con `is-active`.
- **`ColaView.vue`**: `queueTracks` de `PlayerStore`, paginación 20, numeración, botón “Limpiar” (`setQueue([])`), `onPlay` con `playTrack(t, queueTracks)`, empty con icono y descripción.
- **`PlayerBar.vue`**: props completas, `MusicBar_Track` clickeable para seek, `MusicBar_RangeHidden` para arrastre preciso, modos `queue`/`shuffle`/`repeat-one` con `on`, volumen `range`, divisor, responsive (`@media max-width: 760px`).

### 6. SCSS completados (antes vacíos)

Se rellenaron los 6 ficheros que estaban vacíos (0 líneas) y se pulió el sistema de estilos del dominio, respetando `var(--color-*)` y `@use`:

- `Styles/MusicaView.scss` — toolbar sticky, `MusicSearch` con focus, `MusicRows` con scrollbar thin, `MusicRow` hover/active, `MusicEmpty`, `MusicFooter`/`MusicPages`.
- `Styles/BibliotecaView.scss` — `MusicBibHero` con `MusicBibHero_Bg` blur + `MusicBibHero_Glow`, `MusicBibCoverGrid` 2×2, `MusicBibPlaylists` grid auto-fill, `MusicBibPl` active/dashed, filas y footer.
- `Styles/MenuPrincipal.scss` — `MusicHero` con `::before` radial, `MusicHero_Card`, `MusicStatsGrid` 4→2 cols responsive, `MusicEmptyHero`, `MusicRecentGrid`, `MusicRow.small`.
- `Styles/AhoraSuenaView.scss` — `MusicNow_Layout` is-pro (flex + responsive), `MusicNow_Huge` con `MusicNow_HugeBgImg` blur, `MusicNow_Cover is-huge`, `MusicNow_Progress` + `MusicNow_Bar` + `MusicNow_Range`, controles y `MusicNow_Side`.
- `Styles/ColaView.scss` — toolbar con título, `MusicRow_Num`, `MusicRow_Cover`, paginación.
- `Styles/PlayerBar.scss` — `MusicBar` con `backdrop-filter`, `MusicBar_Left/Center/Right`, `MusicBar_Track` con `i` animado, `MusicBar_Modes` con `.on`, responsive wrap.

`Styles/TrackSelector.scss` (nuevo) y `Styles/Library.scss`/`Music.scss` existentes se conservaron como base del layout `MusicLayout`/`MusicMenu`.

### 7. Correcciones de `LocalStore.ts` y tipado

- `ScanMusicFolder` ahora maneja `null` (`const list: any` + `Array.isArray(list) ? list as string[] : []`).
- `Blob` creation tipada como `BlobPart` (`arr as BlobPart`, `data as unknown as BlobPart`) para `vue-tsc`.
- `PlaylistStore` y `PlayerStore` dejan de mezclar música de fondo con biblioteca real.

## Por qué

- Las playlists no podían ser “escribir la ruta”; el usuario debe ver su música real y seleccionar, sin teclear rutas absolutas/relativas.
- El panel cargaba la música de fondo (límite 5 pistas, 15 MB, 10 min) en lugar de la biblioteca local configurada en `Config.MusicPanelConfig` (`musicFolder`, `coverStyle`, `colorMode`, `pageSize`, `showCovers`, `allowAbsolute`). El config ya exponía todo lo necesario (`GetMusicPanelConfig`, `ScanMusicFolder`, `ReadAbsoluteFile`, `ListPlaylists`…), pero el frontend no lo usaba.
- Muchas vistas estaban incompletas: `Music.vue` pasaba props que no coincidían con las vistas hijas, y los `Styles/*.scss` específicos estaban vacíos, por lo que filas, heroes y barras se veían sin estilo o desalineadas.
- Se pidió mejorar todo lo posible y que cualquier archivo nuevo empiece en mayúscula y en inglés: se crearon `PlayerStore.ts`, `TrackSelector.vue` y `TrackSelector.scss` siguiendo esa norma.

## API afectada

- **Bindings Go (sin cambios)**: `GetMusicPanelConfig`, `UpdateMusicPanelConfig`, `PickMusicFolder`, `ScanMusicFolder`, `ReadAbsoluteFile`, `ListPlaylists`, `CreatePlaylist`, `UpdatePlaylist`, `DeletePlaylist`, `ImportPlaylistFile` siguen igual; ahora se usan para música REAL.
- **Stores frontend**:
  - Nuevo `Music/PlayerStore.ts` (exporta `currentTrack`, `queue`, `playing`, `playTrack`, etc.).
  - Nuevo `Music/Components/TrackSelector.vue` (props `modelValue: string[]`, `tracks: LocalTrack[]`, emite `update:modelValue`).
  - `PlaylistStore.ts` y `LocalStore.ts` mantienen su API, con tipado corregido.
- **Componentes**: `Music.vue`, `MusicView.vue`, `LibraryView.vue`, `MenuView.vue`, `NowPlayingView.vue`, `ColaView.vue`, `PlayerBar.vue`, `PlaylistForm.vue` cambian props/eventos para usar `PlayerStore` en vez de `Common/Stores/Music`.

## Comportamiento anterior/nuevo

- **Anterior**: playlists con `textarea` de rutas (“`C:\Musica\cancion.mp3` por línea”); carátulas a veces dummy; vistas con props desalineadas y SCSS vacíos; reproducción recortada a música de fondo.
- **Nuevo**: al crear/editar playlist se abre un `TrackSelector` con tu música REAL (paginada, con búsqueda y carátulas); la biblioteca se escanea desde `musicFolder` (mp3/wav/ogg/m4a/flac, máx. 2000) y se extrae metadata con `music-metadata`; la reproducción es local vía `Blob URL` sin copiar a `cache/audio`; estilos completos con `backdrop-filter`, paletas de carátula (`useCoverPalette` → `paletteToCss`), heroes con glow, grids responsive y `PlayerBar` funcional.

## Revisión 2 — 2026-08-29 — Ajustes tras feedback del usuario

Tras la primera entrega el usuario reportó: la música no cargaba sin tocar “Escanear” manualmente, el diálogo de playlist seguía sin estilo, aparecían textos “Offline 100% local” y “Música Real” no deseados, el scroll no tenía estilo, “Ahora Suena” duplicaba controles, faltaba paginación en dos secciones, el default debía ser 20 (no 50), “Volver a escuchar” debía ser historial persistente y el widget de fondo debía poder reproducir también la biblioteca con mini cola.

### 1. Carga automática (sin “Escanear” manual)

- **Backend** `internal/Handlers/App.go:ScanMusicFolder` ahora resuelve rutas relativas contra `engine.ConfigManager().RootDir()` (`if !filepath.IsAbs(folder) { folder = filepath.Join(RootDir, folder) }`) y hace `filepath.Clean`, por lo que `musicas` o `C:/...` funcionan igual que `ReadAbsoluteFile`.
- **Frontend** `Settings/Sections/MusicPanel.vue:save/pickFolder/scan`: tras guardar la carpeta se despacha `window.dispatchEvent('stl:music-folder-changed')` y se hace `void loadLocalLibrary()` automáticamente. `Music/LocalStore.ts` escucha ese evento y hace `setTimeout(() => void loadLocalLibrary(), 900)` al iniciar, por lo que la biblioteca se carga al abrir el launcher y cada vez que se abre el panel (`Music.vue:watch(heavyPanel)`).
- `MusicView.vue` y `PlayerStore.ts:blobUrlFor` ya no normalizan con `replace(/\\/g,'/')` (usa la ruta tal cual para `ReadAbsoluteFile`).

### 2. Limpieza de textos “Offline” y “Música Real”

- Eliminados de `Music.vue:headerSub` (`100% offline` → `${len} pistas en tu biblioteca`), `MenuView.vue:kicker` y stats (`Pistas reales` → `Pistas`, `Offline` → `En biblioteca`, `100% local` eliminado), `LibraryView.vue:hero` (`100% offline` y `tu música REAL`), `MusicView.vue:empty` (`Se cargará tu música REAL` → `La biblioteca se carga automáticamente`), `PlayerBar.vue:subtitle` (`Tu música real · offline` → `Selecciona una pista`), `NowPlayingView.vue:kicker` (`Ahora suena · Música real` → `Ahora suena`) y `PlaylistForm.vue` (`carátulas reales` → `primeras carátulas`). Se mantiene solo en comentarios internos.

### 3. Scroll con estilo y `PlaylistForm` pulido

- Todos los contenedores con scroll (`MusicMain`, `MusicRows`, `MusicBibPlaylists`, `MusicSection`, `TrackSelector_List`, `MusicNow_Side`, `PlaylistForm_Body`, `ColaView`, `Bgm_QueueList`) ahora usan `scrollbar-width: thin; scrollbar-color: color-mix(in srgb, var(--text-primary) 14%, transparent) transparent;` + `::-webkit-scrollbar` 6px con thumb redondeado, respetando `var(--text-primary)`.
- `PlaylistForm.vue` importa `@use '../../Common/Styles/Components' as *;` y su `Body` ya tiene scrollbar estilizado; se eliminó el comentario `<!-- Selector REAL -->` y se mejoró el hint de cover.
- Todos los `Styles/*.scss` nuevos ahora incluyen `@use '../../Common/Styles/Components' as *;` al inicio para heredar `SsBtn`, `SsIn`, `SsSel` correctamente (antes los botones se veían sin estilo porque el `scoped` no incluía los componentes).

### 4. “Ahora Suena” sin controles duplicados y con paginación

- `NowPlayingView.vue` eliminó `MusicNow_Controls is-huge` (prev/play/next duplicados junto a la carátula) — ahora solo queda el `PlayerBar` inferior. También simplificó `MusicNow_Album` (`Local ·` removido).
- La barra lateral “A continuación” ahora está paginada: `sidePage`/`sidePageSize=10`, `sideTotalPages`, `sidePaginated`, footer con `MusicNow_SideFooter` + `MusicPgBtn` (estilos añadidos en `AhoraSuenaView.scss`). Cumple el requisito “Ahora suena (Barra lateral) debe tener paginación”.

### 5. Paginación default 20 y en “buscar para añadir a playlist”

- `MusicView.vue:pageSize` 50 → 20, `TrackSelector.vue:pageSize` 50 → 20, `LibraryView.vue:bibPageSize` 30 → 20. `Settings/Sections/MusicPanel.vue` ya tenía default 20, ahora coherente.
- `TrackSelector.vue` ya tenía paginación; se verificó que el footer (`TrackSelector_Footer`/`TrackSelector_Pages`) se muestra con `totalPages > 1` y respeta el estilo `SsBtn`.

### 6. Historial `launcher_music_history.json` para “Volver a escuchar”

- **Backend nuevo**: `internal/Core/MusicHistory/History.go` (Manager con `path = RootDir/launcher_music_history.json`, `load`/`save` con `os.Rename` tmp, `Add` con `PlayCount` y `PlayedAt` RFC3339, límite 100, `List`/`Clear`/`SortedByPlays`), `internal/Config/Config.go` añadido `FileMusicHistory`/`ExtraKeyMusicHistory`/`ExtraData.MusicHistory` + `Default` + `sanitize` + `RegisterExtraFile`, `internal/Handlers/App.go` con `musicHistory *musichistory.Manager` + `init` + `GetMusicHistory`/`AddMusicHistory`/`ClearMusicHistory`, `app.go` expone los tres bindings (import `musichistory`).
- **Frontend nuevo**: `frontend/web/src/Music/HistoryStore.ts` (`musicHistory`, `loadHistory`, `addToHistory`, `clearHistory` con fallback `localStorage` `stl_music_history`), `PlayerStore.ts` importa `addToHistory` y lo llama en `loadTrack` cuando `autoPlay`.
- `MenuView.vue` ahora importa `musicHistory`/`loadHistory`, `historyCount` para stats, `volver` mapea `musicHistory.slice(0,6)` a `LocalTrack` placeholder si no está en `localTracks`, y `playHistory` busca en `localTracks` o crea un track sintético. Texto actualizado: “Se guarda en launcher_music_history.json”.

### 7. Widget `BackgroundMusic` con dos fuentes + mini cola

- `Common/Widgets/BackgroundMusic.vue` ahora maneja `launcherSource` (`localStorage stl_bgm_source`, default `background`), `isLauncher`, `setSource`, datos unificados (`displayTitle`/`displayArtist`/`displayCoverUrl`, `playing`/`currentTime`/`duration`/`volume`/`playMode` computados según fuente), `show` (`launcherHasMusic` vs `backgroundMusic.enabled`), `onTogglePlay`/`onNext`/`onPrev`/`onMode`/`onVolume` bifurcados, `queueOpen` + `toggleQueue` + `onWindowDown`/`onEscape` actualizados, `onMounted` carga también `localTracks` (`import('@/Music/LocalStore').then(m=>m.loadLocalLibrary())`).
- Template: `Bgm_Source` con dos `Bgm_SourceBtn` (`Fondo` con `IconDisc` y `Biblioteca` con `IconLibrary`), cover condicional (`isLauncher ? displayCoverUrl : meta.coverUrl`), `Bgm_QueueBtn` junto al volumen y dos paneles `Bgm_VolMenu` + `Bgm_Queue` (lista de 10 pistas, `Bgm_QueueHead`/`Bgm_QueueList`/`Bgm_QueueRow`/`Bgm_QueueCover`/`Bgm_QueueInfo`, empty states diferenciados).
- `BackgroundMusic.scss` añadido bloque para `.Bgm_Source`, `.Bgm_SourceBtn`, `.Bgm_QueueBtn`, `.Bgm_Queue`, etc., con `scrollbar` estilizado y `backdrop-filter`.

## Cómo verificar (revisión 2)

- `go build ./...` — ok (incluye `MusicHistory`).
- `bun run type-check` — 0 errores.
- `bun run build-only` — `✓ built in 13s` (6571 módulos).
- Smoke revisado:
  1. Elegir carpeta en Ajustes → Música → la biblioteca se carga sola sin pulsar Escanear (evento + auto-load).
  2. Abrir Música → Inicio sin “Offline” ni “Música Real”, stats con `Escuchadas`/`En biblioteca`, “Volver a escuchar” muestra historial persistente.
  3. Crear playlist → diálogo estilizado con `TrackSelector` paginado (20) y scroll con thumb.
  4. “Ahora Suena” sin controles duplicados, barra lateral paginada (10).
  5. Widget abajo → tabs Fondo/Biblioteca, botón cola expande mini panel con scroll estilizado.

## Revisión 3 — 2026-08-29 — Dialogo grande | Contenido | Músicas |, sin emojis, single checkbox, auto 100% y optimización de memoria/CPU

El usuario reportó que el diálogo de playlist seguía pequeño, con 2 checkboxes visibles y emojis, y que aún había que ir a `Configuración > Música > Escanear` para que cargara. Además, al abrir el panel el launcher subía a >20% CPU y >500 MB RAM por mantener carátulas y cola en memoria, los controles eran pocos y el volumen no era igual al del widget.

### 1. PlaylistForm grande con layout | Contenido | Músicas |

- `Music/Components/PlaylistForm.vue` (y `HistoryStore.ts`): diálogo ahora `width: 860px` (antes 640px), `max-height: 88vh`, body con `display: grid; grid-template-columns: 340px 1fr` (`.is-split`), responsive a 1 col en <760px. Columna izquierda `is-content` (sticky) con `Contenido` (título, `Favorita`/`Anclada`, color, cover) y columna derecha `is-musics` con `Músicas` (`TrackSelector` con `max-height: 380px`). ColumnHead con `Contenido`/`Músicas` y badge de seleccionadas.
- Sin emojis: `📌` reemplazado por `<IconPin>` y `★` removido de `LibraryView.vue:138,117` (ahora solo `IconStar`/`IconPin` badges y texto “Favorita”/“Anclada” sin símbolos).
- Single checkbox: `TrackSelector.scss:168` ahora oculta el `input[type=checkbox]` nativo (`position:absolute; opacity:0`) y deja solo el círculo custom `.TrackSelector_Check` (antes se veían dos).
- `TrackSelector.vue` y `Styles/TrackSelector.scss` ya importan `Components` y la paginación es por 20 (default), con `max-height: 380px` en el contexto del diálogo.

### 2. Carga 100% automática (nunca más Escanear manual)

- `Settings/Sections/MusicPanel.vue:26,39,54` ahora en `onMounted` si hay `musicFolder` hace `void scan()` inmediato y tras `pickFolder`/`save` hace `void scan()` + `void loadLocalLibrary()` + `dispatchEvent('stl:music-folder-changed')`. Tip cambiado a “La biblioteca se carga automáticamente… Re-escanea solo si añades música nueva” y botón a “Re-escanear” (opcional).
- `Music/LocalStore.ts:144` auto-load con reintentos exponenciales (6 intentos, `700ms → 1200*attempt`), escucha `stl:music-folder-changed` y `focus` para recargar si sigue vacío. `MusicView.vue` empty ahora ofrece `Abrir ajustes` (`settingsOpen=true`) además de `Recargar`.
- `Music/Music.vue:watch(heavyPanel)` y `BackgroundMusic.vue:onMounted` también disparan `loadLocalLibrary`, por lo que abrir el panel o el widget ya carga sin pasar por Ajustes.

### 3. Optimización CPU/RAM — carátulas y cola no en memoria

- `Music/LocalStore.ts:30` carga ligera: `loadLocalLibrary` ya no lee cada archivo con `ReadAbsoluteFile` + `parseBlob` para extraer carátula/duración; solo crea `LocalTrack` con `fileName`/`title` y `artist=Desconocido`/`duration=0`/`coverUrl=''` (o cache). Solo cuando una pista es visible se llama `ensureTrackMeta`/`ensureCovers` (lazy, máx. 12 a la vez, `pendingMeta` Set, `coverCache` LRU 40 con `URL.revokeObjectURL`, `metaCache`).
- `MusicView.vue`, `LibraryView.vue`, `MenuView.vue`, `NowPlayingView.vue`, `ColaView.vue`, `TrackSelector.vue` ahora hacen `watch(paginated, (list)=>void ensureCovers(list))` para cargar solo lo visible (20 por página).
- `Music/PlayerStore.ts:14` `blobUrlCache` limitado a 6, `coverCache` a 40; `queue` ahora se persiste en `launcher_music_nowplaying.json` temporal (`internal/Core/NowPlaying/NowPlaying.go` con `Queue{Tracks, CurrentIndex, CurrentPath}`, `Get/Set/Clear`, `app.go` bindings y `frontend` fallback `localStorage stl_nowplaying_queue` con `persistQueue` debounce 400ms y `tryRestoreQueue` con reintentos cuando `localTracks` esté disponible). No se mantiene la cola completa con carátulas en RAM permanente.

### 4. Controles completos y volumen como el widget

- `Music/PlayerStore.ts:15` nuevos estados `shuffleMode: 'linear'|'shuffle'` (`IconArrowRightBar` vs `IconArrowsShuffle`) y `repeatMode: 'queue'|'one'` (`IconRepeat` vs `IconRepeatOnce`) con `toggleShuffleMode`/`toggleRepeatMode`, `nextNext`/`prevPrev` (salto de 2, respeta shuffle), `seekForward10`/`seekBackward10` (±10s).
- `Music/Components/PlayerBar.vue` reescrito: props `shuffleMode`/`repeatMode`, 7 botones centrales (`PrevPrev` `IconPlayerTrackPrev`, `Prev` `IconPlayerSkipBack`, `-10` `IconRewindBackward10`, `PlayPause`, `+10` `IconRewindForward10`, `Next` `IconPlayerSkipForward`, `NextNext` `IconPlayerTrackNext`), modos con `isShuffle()`/`isRepeatOne()`, y volumen con popup idéntico al widget (`volOpen`/`volBtn`/`volWrap` + `onWindowDown`/`onEsc` + `Transition MusicVol` + `Bgm_VolMenu` clonado en `PlayerBar.scss` con `backdrop-filter` y `VolRange`).
- `Music/Music.vue` ahora pasa `shuffleMode`/`repeatMode` y escucha `prev-prev`/`next-next`/`seek-backward`/`seek-forward`/`toggle-shuffle`/`toggle-repeat`.
- `Music/PlayerBar.scss` añadido `MusicBar_VolWrap`/`MusicBar_VolMenu`/`MusicBar_VolIcon`/`MusicBar_VolRange`/`MusicBar_VolVal` + transiciones y responsive.

## Cómo verificar (revisión 3)

- `go build ./...` — ok (incluye `MusicHistory` + `NowPlaying`).
- `bun run type-check` — 0 errores (persistencia con `as any` fallback si bindings aún no regenerados).
- `bun run build-only` — `✓ 6571 modules` en ~14s.
- Smoke:
  1. Borrar `musicFolder`, abrir Música → vacío con `Abrir ajustes` → elegir carpeta → sin tocar Escanear, la lista aparece sola (lazy, sin pico de RAM).
  2. Crear playlist → diálogo 860px con `| Contenido | Músicas |`, sin emojis, 1 checkbox por fila, paginación 20.
  3. Abrir panel con 500+ pistas → CPU <5% y RAM estable (covers solo de las 20 visibles, cola en `launcher_music_nowplaying.json`).
  4. PlayerBar abajo con 7 controles + aleatorio/bucle de 2 estados + volumen popup igual al widget; `±10s` y `PrevPrev/NextNext` funcionan.
