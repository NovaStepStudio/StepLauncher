import { ref, watch } from 'vue';
import { parseBlob } from 'music-metadata';
import { loadLocal, mimeOf, personalization } from './Ui';
import { ReadLocalFile } from '@wailsjs/StepLauncher/internal/Services/System/systemservice';
import { ListMusic, ReadMusicFile, AddMusic, RemoveMusic, UpdatePersonalization } from '@wailsjs/StepLauncher/internal/Services/Appearance/appearanceservice';
import { Events } from '@wailsio/runtime';

// Tipos de la música de fondo. La lista vive en launcher_assets.json (campo
// "audio") y los archivos en cache/audio/.
export interface MusicSlot {
    name: string;
    path: string;
}

// Metadatos extraídos con music-metadata: título/artista/duración y la
// carátula embebida (objectURL), para darle estilo al widget.
export interface TrackMeta {
    title: string;
    artist: string;
    duration: number;
    coverUrl: string;
}

export const MUSIC_EXTENSIONS = ['mp3', 'wav', 'ogg', 'm4a'];
export const MAX_MUSIC_TRACKS = 5;
export const MAX_MUSIC_SECONDS = 600; // 10 minutos
export const MAX_MUSIC_BYTES = 15 * 1024 * 1024; // 15 MB

export const musicList = ref<MusicSlot[]>([]);
export const currentIndex = ref(-1);
export const playing = ref(false);
export const volume = ref(0.8);
export const currentTime = ref(0);
export const loadError = ref('');
export const metaCache = ref<Record<string, TrackMeta>>({});

// Modos de reproducción: cola (repetición al terminar), aleatorio o repetir
// la misma pista en bucle.
export type PlayMode = 'queue' | 'shuffle' | 'repeat-one';
export const playMode = ref<PlayMode>('queue');

// Reproducción con el elemento <audio> nativo. El audio se sirve como data URI
// base64 con su MIME real (los bytes vienen del backend): en WebView2 los blob
// URLs en el scheme wails:// cargan pero pueden quedarse mudos (ver
// Changelogs/Changes/StepLauncher-2.4.1/StepLauncher-Change-3.md), mientras
// que los data URIs son fiables en cualquier scheme.
let audio: HTMLAudioElement | null = null;

// Intención de reproducción: si el usuario quiere que suene, se reintenta
// automáticamente cuando el recurso esté listo (canplay) sin esperar a nada más.
let wantsPlay = false;

// ---- Fuente del audio ----
// ReadLocalFile devuelve los bytes en base64; el data URI se construye pegando
// esa cadena al prefijo con el MIME real, sin reconversiones pesadas en JS.
const audioDataCache = new Map<string, string>();

function bytesToBase64(data: Uint8Array): string {
    let bin = '';
    const chunk = 0x8000;
    for (let i = 0; i < data.length; i += chunk) {
        bin += String.fromCharCode(...data.subarray(i, i + chunk));
    }
    return btoa(bin);
}

export function bytesToDataUri(data: Uint8Array, mime: string): string {
    return `data:${mime};base64,${bytesToBase64(data)}`;
}

async function musicDataUri(rel: string): Promise<string> {
    const key = String(rel ?? '').replace(/\\/g, '/');
    if (!key) return '';
    const cached = audioDataCache.get(key);
    if (cached) return cached;
    try {
        const res: any = await ReadLocalFile(key);
        if (!res) return '';
        const mime = mimeOf(key) || 'audio/mpeg';
        let uri: string;
        if (typeof res === 'string') {
            // El binding ya entrega los bytes como base64: se pega directo.
            uri = `data:${mime};base64,${res}`;
        } else if (res instanceof Uint8Array) {
            uri = bytesToDataUri(res, mime);
        } else if (Array.isArray(res)) {
            uri = bytesToDataUri(new Uint8Array(res), mime);
        } else {
            return '';
        }
        if (audioDataCache.size >= 2) {
            const first = audioDataCache.keys().next().value;
            if (first !== undefined) {
                audioDataCache.delete(first);
            }
        }
        audioDataCache.set(key, uri);
        return uri;
    } catch {
        return '';
    }
}

// ---- Integración con el sistema (Media Session / controles multimedia) -----
// Windows y otros sistemas reconocen el reproductor vía navigator.mediaSession:
// pausar/reproducir, anterior/siguiente y la metadata con carátula.

let sessionBound = false;

function bindMediaSession() {
    if (sessionBound) return;
    const ms = (navigator as any).mediaSession;
    if (!ms) return;
    sessionBound = true;
    const noop = () => { };
    ms.setActionHandler?.('play', () => { void togglePlay(); });
    ms.setActionHandler?.('pause', () => { void togglePlay(); });
    ms.setActionHandler?.('previoustrack', () => { void prev(); });
    ms.setActionHandler?.('nexttrack', () => { void next(); });
    ms.setActionHandler?.('seekbackward', noop);
    ms.setActionHandler?.('seekforward', noop);
}

function getSMTCSource(): string {
    try {
        return localStorage.getItem('stl_smtc_source') || 'auto';
    } catch { return 'auto'; }
}
function updateMediaSession() {
    const ms = (navigator as any).mediaSession;
    if (!ms) return;
    const src = getSMTCSource();
    if (src === 'library') return; // biblioteca tiene prioridad
    if (src === 'auto') {
        try { if ((window as any).__stlLauncherPlaying) return; } catch (_e) {}
    }
    bindMediaSession();
    try {
        const m = currentMeta();
        ms.metadata = new MediaMetadata({
            // El sistema operativo muestra la pista con el nombre del launcher
            // delante (SMTC de Windows, Centro de control de macOS, etc.).
            title: `StepLauncher - ${m?.title || 'Música de fondo'}`,
            artist: m?.artist || 'StepLauncher',
            album: 'StepLauncher',
            artwork: m?.coverUrl ? [{ src: m.coverUrl, sizes: '512x512' }] : [],
        });
    } catch (_e) {}
    ms.playbackState = playing.value ? 'playing' : 'paused';
}

export function setPlayMode(m: PlayMode) {
    // Si activa aleatorio, crea una cola barajada persistente
    if (m === 'shuffle' && playMode.value !== 'shuffle') {
        buildShuffledQueue();
    }
    playMode.value = m;
    syncLoop();
}

// Cola barajada persistente: al activar aleatorio se genera una lista
// desordenada de índices que se guarda en memoria y se respeta en
// siguiente/anterior, en vez de elegir random en cada salto.
const shuffledQueue = ref<number[]>([]);
const shufflePos = ref(0);

function shuffleArray(arr: number[]): number[] {
    const a = [...arr];
    for (let i = a.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        const tmp = a[i]!; a[i] = a[j]!; a[j] = tmp;
    }
    return a;
}

export function buildShuffledQueue(): void {
    const len = musicList.value.length;
    if (len <= 1) {
        shuffledQueue.value = len ? [0] : [];
        shufflePos.value = 0;
        return;
    }
    const indices = Array.from({ length: len }, (_, i) => i);
    const shuffled = shuffleArray(indices);
    // Asegura que la pista actual quede al inicio de la cola barajada
    // para no saltar bruscamente al activar aleatorio
    const cur = currentIndex.value;
    if (cur >= 0 && cur < len) {
        const at = shuffled.indexOf(cur);
        if (at > 0) {
            shuffled.splice(at, 1);
            shuffled.unshift(cur);
        }
    }
    shuffledQueue.value = shuffled;
    shufflePos.value = 0;
}

export function getShuffledQueue(): number[] {
    return [...shuffledQueue.value];
}

export function getQueueOrder(): number[] {
    if (playMode.value === 'shuffle' && shuffledQueue.value.length) return [...shuffledQueue.value];
    return Array.from({ length: musicList.value.length }, (_, i) => i);
}

// Con una sola pista (o en modo "repetir esta pista") el audio repite en
// bucle infinito de forma nativa.
function syncLoop() {
    if (audio) {
        audio.loop = musicList.value.length <= 1 || playMode.value === 'repeat-one';
    }
}

// Índice de la siguiente pista según el modo: con cola avanza en orden;
// con aleatorio respeta la cola barajada persistente.
function nextIndex(mode: PlayMode, cur: number, len: number): number {
    if (len <= 1) return 0;
    if (mode === 'shuffle') {
        if (!shuffledQueue.value.length || shuffledQueue.value.length !== len) buildShuffledQueue();
        const pos = shuffledQueue.value.indexOf(cur);
        const nextPos = pos >= 0 ? (pos + 1) % shuffledQueue.value.length : 0;
        shufflePos.value = nextPos;
        return shuffledQueue.value[nextPos] ?? 0;
    }
    return (cur + 1) % len;
}

function prevShuffledIndex(cur: number, len: number): number {
    if (!shuffledQueue.value.length || shuffledQueue.value.length !== len) buildShuffledQueue();
    const pos = shuffledQueue.value.indexOf(cur);
    const prevPos = pos >= 0 ? (pos - 1 + shuffledQueue.value.length) % shuffledQueue.value.length : 0;
    shufflePos.value = prevPos;
    return shuffledQueue.value[prevPos] ?? 0;
}

function restoreVolume() {
    // La fuente de verdad del volumen es la personalización global; el
    // localStorage se conserva como respaldo de versiones anteriores.
    const fromConfig = personalization.value?.backgroundMusic?.volume;
    if (typeof fromConfig === 'number' && fromConfig >= 0 && fromConfig <= 1) {
        volume.value = fromConfig;
        return;
    }
    try {
        const v = Number(localStorage.getItem('stl_music_volume'));
        if (Number.isFinite(v) && v >= 0 && v <= 1) volume.value = v;
    } catch (_e) {}
}

function persistVolume() {
    try {
        localStorage.setItem('stl_music_volume', String(volume.value));
    } catch (_e) {}
    // El volumen también se refleja en la personalización global (Ajustes →
    // Música de fondo) y se persiste en la config, para que el slider de
    // Ajustes muestre el mismo valor y sobreviva al reinicio.
    const p = personalization.value;
    if (p?.backgroundMusic && p.backgroundMusic.volume !== volume.value) {
        p.backgroundMusic.volume = volume.value;
        void UpdatePersonalization?.(p as any).catch(() => { });
    }
}

// ---- Elemento <audio> global ----
// Vive fuera del DOM (creado con new Audio()), así Vue nunca lo destruye:
// sobrevive a v-show, Teleport y al ciclo de vida del widget.

function ensureAudio(): HTMLAudioElement | null {
    if (typeof document === 'undefined') return null;
    if (!audio) {
        const a = new Audio();
        a.preload = 'auto';
        a.volume = volume.value;
        a.addEventListener('play', onAudioPlay);
        a.addEventListener('pause', onAudioPause);
        a.addEventListener('ended', onAudioEnded);
        a.addEventListener('timeupdate', onAudioTime);
        a.addEventListener('loadedmetadata', onAudioMeta);
        a.addEventListener('canplay', onAudioCanPlay);
        a.addEventListener('error', onAudioError);
        audio = a;
    }
    return audio;
}

function onAudioPlay() {
    wantsPlay = true;
    playing.value = true;
    loadError.value = '';
    updateMediaSession();
}

function onAudioPause() {
    playing.value = false;
    updateMediaSession();
}

function onAudioEnded() {
    playing.value = false;
    // En bucle nativo no llega ended (loop = true).
    void next();
}

function onAudioTime() {
    if (audio) {
        currentTime.value = Number(audio.currentTime) || 0;
    }
}

function onAudioMeta() {
    const a = audio;
    if (!a) return;
    const cur = musicList.value[currentIndex.value];
    if (cur && Number.isFinite(a.duration) && a.duration > 0) {
        const t = metaCache.value[cur.path];
        if (t && !t.duration) {
            metaCache.value = {
                ...metaCache.value,
                [cur.path]: { ...t, duration: a.duration },
            };
        }
    }
}

function onAudioCanPlay() {
    loadError.value = '';
    // Reintento rápido si el usuario pulsó play antes de que cargara.
    if (wantsPlay && !playing.value) {
        void audio?.play().catch(() => { });
    }
}

function onAudioError() {
    const err = audio?.error;
    loadError.value = err
        ? `No se pudo cargar la pista (código ${err.code}).`
        : 'No se pudo cargar la pista.';
    playing.value = false;
    console.warn('[Music] error al cargar el audio:', err);
}

let retryTimer: number | null = null;

function stopAudio() {
    if (retryTimer !== null) {
        window.clearTimeout(retryTimer);
        retryTimer = null;
    }
    wantsPlay = false;
    if (audio) {
        audio.pause();
        audio.removeAttribute('src');
        audio.load();
    }
    playing.value = false;
}

// Carga (y opcionalmente reproduce) la pista `i` con el <audio> nativo. El
// audio se sirve como data URI y el navegador lo decodifica al reproducir.
async function loadTrack(i: number, autoPlay: boolean): Promise<void> {
    const list = musicList.value;
    if (!list.length) return;
    const idx = ((i % list.length) + list.length) % list.length;
    const slot = list[idx];
    if (!slot) return;
    const uri = await musicDataUri(slot.path);
    if (!uri) {
        loadError.value = `No se pudo leer el archivo de "${slot.name}".`;
        console.warn('[Music] no se pudo leer la pista:', slot.path);
        return;
    }
    const a = ensureAudio();
    if (!a) return;
    stopAudio();
    loadError.value = '';
    currentIndex.value = idx;
    currentTime.value = 0;
    updateMediaSession();

    a.src = uri;
    a.volume = volume.value;
    a.loop = musicList.value.length <= 1 || playMode.value === 'repeat-one';
    a.load();
    if (autoPlay) {
        // Pausar el reproductor de biblioteca para exclusión mutua
        try { (window as any).__pauseLauncher?.(); } catch (_e) {}
        wantsPlay = true;
        void a.play().catch(() => {
            // El navegador pudo bloquear el play inicial; se reintenta en
            // canplay (o tras 800 ms) sin molestar al usuario.
            if (retryTimer !== null) window.clearTimeout(retryTimer);
            retryTimer = window.setTimeout(() => {
                retryTimer = null;
                if (wantsPlay && !playing.value) void a.play().catch(() => { });
            }, 800);
        });
    }
}

export function isMusicEnabled(): boolean {
    return !!personalization.value?.backgroundMusic?.enabled;
}

export function musicPosition(): 'top-left' | 'bottom-center' {
    return personalization.value?.backgroundMusic?.position ?? 'bottom-center';
}

export function musicCoverStyle(): 'disc' | 'square' | 'background' {
    return personalization.value?.backgroundMusic?.coverStyle ?? 'disc';
}

export function musicDiscRotation(): boolean {
    return personalization.value?.backgroundMusic?.discRotation ?? true;
}

export async function loadMusicList(): Promise<void> {
    try {
        const list = await ListMusic();
        const nextList = (Array.isArray(list) ? list : []) as MusicSlot[];
        const prev = musicList.value;
        musicList.value = nextList;
        syncLoop();
        if (nextList.length === 0) {
            shuffledQueue.value = [];
            shufflePos.value = 0;
            stopMusic();
            return;
        }
        // Si estaba en aleatorio, regenerar cola para reflejar nueva lista
        if (playMode.value === 'shuffle') buildShuffledQueue();
        if (currentIndex.value >= nextList.length) {
            currentIndex.value = nextList.length - 1;
        }
        if (currentIndex.value >= 0 && !nextList.some((s) => s.path === prev[currentIndex.value]?.path)) {
            stopMusic();
        }
        if (currentIndex.value < 0) {
            await loadTrack(0, false);
        }
        void refreshMetas(nextList);
    } catch {
        musicList.value = [];
        shuffledQueue.value = [];
    }
}

// Re-extrae (o reutiliza) los metadatos de cada pista para el widget.
async function refreshMetas(list: MusicSlot[]) {
    const missing = list.filter((s) => !metaCache.value[s.path]);
    for (const s of missing) {
        try {
            const url = await loadLocal(s.path);
            if (!url) continue;
            const res = await fetch(url);
            const blob = await res.blob();
            const meta = await parseBlob(blob, { duration: true, skipCovers: false });
            const pic = meta.common.picture?.[0];
            let coverUrl = '';
            if (pic?.data) {
                const coverBytes = new Uint8Array(pic.data);
                const coverBuf = coverBytes.buffer.slice(coverBytes.byteOffset, coverBytes.byteOffset + coverBytes.byteLength) as ArrayBuffer;
                coverUrl = URL.createObjectURL(new Blob([coverBuf], { type: pic.format || 'image/jpeg' }));
            }
            const trackMeta: TrackMeta = {
                title: meta.common.title?.trim() || s.name,
                artist: meta.common.artist?.trim() || 'Desconocido',
                duration: meta.format.duration ?? 0,
                coverUrl,
            };
            metaCache.value = { ...metaCache.value, [s.path]: trackMeta };
            // Si es la pista en reproducción, se actualiza la metadata del
            // sistema (SMTC) ahora que se conoce título/artista/carátula.
            if (musicList.value[currentIndex.value]?.path === s.path) {
                updateMediaSession();
            }
        } catch {
            metaCache.value = {
                ...metaCache.value,
                [s.path]: { title: s.name, artist: 'Desconocido', duration: 0, coverUrl: '' },
            };
        }
    }
}

export async function playTrack(i: number): Promise<void> {
    await loadTrack(i, true);
}

export async function togglePlay(): Promise<void> {
    if (currentIndex.value < 0 || !musicList.value.length) {
        if (musicList.value.length) await loadTrack(0, true);
        return;
    }
    const a = audio;
    if (!a || !a.src) {
        await loadTrack(currentIndex.value, true);
        return;
    }
    if (playing.value) {
        wantsPlay = false;
        a.pause();
    } else {
        try { (window as any).__pauseLauncher?.(); } catch (_e) {}
        wantsPlay = true;
        void a.play().catch(() => {
            if (wantsPlay && !playing.value) void a.play().catch(() => { });
        });
    }
}

export async function next(): Promise<void> {
    if (!musicList.value.length) return;
    const i = currentIndex.value < 0 ? 0 : nextIndex(playMode.value, currentIndex.value, musicList.value.length);
    await loadTrack(i, true);
}

export async function prev(): Promise<void> {
    if (!musicList.value.length) return;
    if (audio && currentTime.value > 3) {
        await seekTo(0);
        return;
    }
    const len = musicList.value.length;
    let i: number;
    if (playMode.value === 'shuffle') {
        i = prevShuffledIndex(currentIndex.value, len);
    } else {
        i = currentIndex.value <= 0 ? len - 1 : currentIndex.value - 1;
    }
    await loadTrack(i, true);
}

let volumeTimer: number | null = null;

export function setVolume(v: number) {
    const clamped = Math.min(1, Math.max(0, Number(v) || 0));
    volume.value = clamped;
    if (audio) audio.volume = clamped;
    if (volumeTimer !== null) {
        window.clearTimeout(volumeTimer);
    }
    volumeTimer = window.setTimeout(persistVolume, 300);
}

export async function seekTo(t: number) {
    if (audio && Number.isFinite(t)) {
        try {
            audio.currentTime = Math.max(0, t);
            currentTime.value = Math.max(0, t);
        } catch (_e) {}
    }
}

export function stopMusic() {
    stopAudio();
    currentIndex.value = -1;
    currentTime.value = 0;
    updateMediaSession();
}

// Lee el archivo seleccionado con el diálogo nativo y lo valida con el MISMO
// motor que reproduce la música (el <audio> nativo): extensión soportada,
// <= 15 MB y duración <= 10 minutos medida sobre un data URI real. Así lo que
// se importa siempre se puede reproducir en WebView2.
export async function validateAndReadMusic(srcPath: string): Promise<{ path: string; name: string; ext: string }> {
    const ext = (srcPath.split('.').pop() ?? '').toLowerCase();
    if (!MUSIC_EXTENSIONS.includes(ext)) {
        throw new Error('Formato no soportado. Usa MP3, WAV, OGG o M4A.');
    }
    const bytes = await ReadMusicFile(srcPath) as Uint8Array | string | null;
    if (!bytes) {
        throw new Error('No se pudo leer el archivo.');
    }
    let data: Uint8Array;
    if (typeof bytes === 'string') {
        const bin = atob(bytes);
        data = new Uint8Array(bin.length);
        for (let i = 0; i < bin.length; i++) data[i] = bin.charCodeAt(i);
    } else if (bytes instanceof Uint8Array) {
        data = bytes;
    } else if (Array.isArray(bytes)) {
        data = new Uint8Array(bytes);
    } else {
        throw new Error('No se pudo leer el archivo.');
    }
    if (data.byteLength > MAX_MUSIC_BYTES) {
        throw new Error('El audio no debe pesar más de 15 MB.');
    }
    const mime = ext === 'mp3' ? 'audio/mpeg' : ext === 'wav' ? 'audio/wav' : ext === 'ogg' ? 'audio/ogg' : 'audio/mp4';
    const uri = bytesToDataUri(data, mime);
    const duration = await measureAudioDuration(uri);
    if (duration <= 0) {
        throw new Error('El archivo no parece un audio válido para StepLauncher.');
    }
    if (duration > MAX_MUSIC_SECONDS) {
        throw new Error('El audio no debe durar más de 10 minutos.');
    }
    const base = (srcPath.split(/[\\/]/).pop() ?? 'audio');
    const name = base.replace(/\.[^.]+$/, '');
    return { path: srcPath, name, ext };
}

// Carga un data URI en un <audio> temporal y devuelve su duración en segundos
// (0 si no se pudo reproducir). Es la misma ruta que usa el reproductor, así
// que valida de verdad que la pista sonará.
function measureAudioDuration(src: string): Promise<number> {
    return new Promise((resolve) => {
        const a = new Audio();
        let done = false;
        const finish = (d: number) => {
            if (done) return;
            done = true;
            a.removeAttribute('src');
            a.load();
            resolve(d);
        };
        a.preload = 'metadata';
        a.addEventListener('loadedmetadata', () => finish(Number.isFinite(a.duration) ? a.duration : 0), { once: true });
        a.addEventListener('error', () => finish(0), { once: true });
        a.src = src;
        a.load();
        // Salvaguarda: si el navegador tarda demasiado, se descarta.
        window.setTimeout(() => finish(0), 15000);
    });
}

// Añade un audio de fondo (después de validateAndReadMusic) y refresca la
// lista. Devuelve el mensaje de error o '' si fue bien.
export async function addMusic(srcPath: string, name: string, ext: string): Promise<string> {
    try {
        const rel = await AddMusic(name + '.' + ext, srcPath);
        if (!rel) return 'No se pudo añadir el audio.';
        await loadMusicList();
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo añadir el audio.';
    }
}

export async function removeMusic(name: string): Promise<string> {
    try {
        await RemoveMusic(name);
        if (currentIndex.value >= 0 && musicList.value[currentIndex.value]?.name === name) {
            stopMusic();
        }
        await loadMusicList();
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo eliminar el audio.';
    }
}

export function currentMeta(): TrackMeta | null {
    const slot = musicList.value[currentIndex.value];
    if (!slot) return null;
    return metaCache.value[slot.path] ?? { title: slot.name, artist: 'Desconocido', duration: 0, coverUrl: '' };
}

export function formatDuration(sec: number): string {
    if (!Number.isFinite(sec) || sec <= 0) return '0:00';
    const m = Math.floor(sec / 60);
    const s = Math.floor(sec % 60);
    return `${m}:${String(s).padStart(2, '0')}`;
}

restoreVolume();
if (typeof window !== 'undefined') {
    try { (window as any).__pauseBg = () => { try { if (audio) audio.pause(); playing.value = false; updateMediaSession(); } catch (_e) {} }; } catch (_e) {}
}

// El volumen de la personalización global (Ajustes → Música de fondo) se
// aplica en vivo al reproductor.
watch(
    () => personalization.value?.backgroundMusic?.volume,
    (v) => {
        if (typeof v === 'number' && v >= 0 && v <= 1) {
            setVolume(v);
        }
    }
);

// Al desactivar la música de fondo en Ajustes se detiene la reproducción.
watch(
    () => personalization.value?.backgroundMusic?.enabled,
    (enabled) => {
        if (!enabled) {
            wantsPlay = false;
            if (audio) audio.pause();
            playing.value = false;
        }
    }
);

// El system tray (ítem "Pausar / Reproducir") alterna la música de fondo: el
// backend emite este evento al hacer clic en el menú del área de notificaciones.
Events.On('music_tray_toggle', () => {
    void togglePlay();
});