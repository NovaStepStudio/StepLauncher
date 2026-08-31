import { ref, computed, watch } from 'vue';
import type { LocalTrack } from './LocalStore';
import { localTracks, ensureTrackMeta } from './LocalStore';
import { addToHistory } from './HistoryStore';

// Estado global del reproductor local (música REAL, no música de fondo).
// Usa un único HTMLAudioElement y Blob URLs creadas desde ReadAbsoluteFile.

export const currentTrack = ref<LocalTrack | null>(null);
export const queue = ref<LocalTrack[]>([]);
export const currentIndex = ref(-1);
export const playing = ref(false);
export const currentTime = ref(0);
export const duration = ref(0);
export const volume = ref(0.85);
export const playMode = ref<'queue' | 'shuffle' | 'repeat-one'>('queue');
// Nuevos estados separados para los controles pedidos
export const shuffleMode = ref<'linear' | 'shuffle'>('linear');
export const repeatMode = ref<'queue' | 'one'>('queue');
export const loadError = ref('');
export const isLoadingTrack = ref(false);

const shuffledOrder = ref<number[]>([]);
const audioUrlCache = new Map<string, string>();
const blobUrlCache = new Map<string, string>();

let audio: HTMLAudioElement | null = null;
let wantsPlay = false;
let retryTimer: number | null = null;

function mimeOf(path: string): string {
    const ext = path.split('.').pop()?.toLowerCase() ?? '';
    switch (ext) {
        case 'mp3': return 'audio/mpeg';
        case 'wav': return 'audio/wav';
        case 'ogg': return 'audio/ogg';
        case 'm4a': return 'audio/mp4';
        case 'flac': return 'audio/flac';
        default: return 'audio/mpeg';
    }
}

function restoreVolume(): void {
    try {
        const v = Number(localStorage.getItem('stl_local_music_volume'));
        if (Number.isFinite(v) && v >= 0 && v <= 1) volume.value = v;
    } catch (_e) {}
}

function persistVolume(): void {
    try {
        localStorage.setItem('stl_local_music_volume', String(volume.value));
    } catch (_e) {}
}

let volumeTimer: number | null = null;

async function persistQueue(): Promise<void> {
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.SetNowPlayingQueue === 'function') {
            const tracks = queue.value.map((t) => t.path);
            await mod.SetNowPlayingQueue(tracks, currentIndex.value, currentTrack.value?.path || '').catch(() => {});
        } else {
            // Fallback localStorage si el binding aún no está generado
            try { localStorage.setItem('stl_nowplaying_queue', JSON.stringify({ tracks: queue.value.map((t) => t.path), idx: currentIndex.value, cur: currentTrack.value?.path || '' })); } catch (_e) {}
        }
    } catch {
        try { localStorage.setItem('stl_nowplaying_queue', JSON.stringify({ tracks: queue.value.map((t) => t.path), idx: currentIndex.value, cur: currentTrack.value?.path || '' })); } catch (_e) {}
    }
}

let persistTimer: number | null = null;
function schedulePersist(): void {
    if (persistTimer !== null) window.clearTimeout(persistTimer);
    persistTimer = window.setTimeout(() => { void persistQueue(); persistTimer = null; }, 400);
}

function shuffleArray(arr: number[]): number[] {
    const a = [...arr];
    for (let i = a.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        const tmp = a[i]!; a[i] = a[j]!; a[j] = tmp;
    }
    return a;
}

export function buildShuffledQueue(): void {
    const len = queue.value.length;
    if (len <= 1) {
        shuffledOrder.value = len ? [0] : [];
        return;
    }
    const indices = Array.from({ length: len }, (_, i) => i);
    const shuffled = shuffleArray(indices);
    const cur = currentIndex.value;
    if (cur >= 0 && cur < len) {
        const at = shuffled.indexOf(cur);
        if (at > 0) {
            shuffled.splice(at, 1);
            shuffled.unshift(cur);
        }
    }
    shuffledOrder.value = shuffled;
}

export function getQueueOrder(): number[] {
    const isShuf = shuffleMode.value === 'shuffle' || playMode.value === 'shuffle';
    if (isShuf && shuffledOrder.value.length === queue.value.length) return [...shuffledOrder.value];
    return Array.from({ length: queue.value.length }, (_, i) => i);
}

function nextIndex(mode: string, cur: number, len: number): number {
    if (len <= 1) return 0;
    const isShuf = mode === 'shuffle' || shuffleMode.value === 'shuffle';
    if (isShuf) {
        if (!shuffledOrder.value.length || shuffledOrder.value.length !== len) buildShuffledQueue();
        const pos = shuffledOrder.value.indexOf(cur);
        const nextPos = pos >= 0 ? (pos + 1) % shuffledOrder.value.length : 0;
        return shuffledOrder.value[nextPos] ?? 0;
    }
    return (cur + 1) % len;
}

function prevShuffleIndex(cur: number, len: number): number {
    if (!shuffledOrder.value.length || shuffledOrder.value.length !== len) buildShuffledQueue();
    const pos = shuffledOrder.value.indexOf(cur);
    const prevPos = pos >= 0 ? (pos - 1 + shuffledOrder.value.length) % shuffledOrder.value.length : 0;
    return shuffledOrder.value[prevPos] ?? 0;
}

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

function onAudioPlay(): void {
    wantsPlay = true;
    playing.value = true;
    loadError.value = '';
    updateMediaSession();
}

function onAudioPause(): void {
    playing.value = false;
    updateMediaSession();
}

function onAudioEnded(): void {
    playing.value = false;
    void next();
}

function onAudioTime(): void {
    if (audio) currentTime.value = Number(audio.currentTime) || 0;
}

function onAudioMeta(): void {
    if (audio) duration.value = Number(audio.duration) || 0;
}

function onAudioCanPlay(): void {
    loadError.value = '';
    if (wantsPlay && !playing.value) void audio?.play().catch(() => {});
}

function onAudioError(): void {
    const err = audio?.error;
    loadError.value = err ? `No se pudo reproducir (código ${err.code}).` : 'No se pudo reproducir la pista.';
    playing.value = false;
}

function stopAudio(): void {
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

function syncLoop(): void {
    if (audio) audio.loop = queue.value.length <= 1 || repeatMode.value === 'one' || playMode.value === 'repeat-one';
}

async function blobUrlFor(path: string): Promise<string> {
    const key = String(path ?? '').trim();
    if (!key) return '';
    const cached = blobUrlCache.get(key);
    if (cached) return cached;
    try {
        const { ReadAbsoluteFile } = await import('@wailsjs/StepLauncher/internal/Services/System/systemservice');
        const res: any = await ReadAbsoluteFile(key);
        if (!res) return '';
        let bytes: Uint8Array | null = null;
        if (typeof res === 'string') {
            const bin = atob(res as string);
            bytes = new Uint8Array(bin.length);
            for (let k = 0; k < bin.length; k++) bytes[k] = bin.charCodeAt(k);
        } else if (res instanceof Uint8Array) bytes = res;
        else if (Array.isArray(res)) bytes = new Uint8Array(res as any);
        else return '';
        if (!bytes || bytes.length === 0) return '';
        const mime = mimeOf(key);
        const blob = new Blob([bytes as any], { type: mime });
        const url = URL.createObjectURL(blob);
        if (blobUrlCache.size >= 6) {
            const first = blobUrlCache.keys().next().value as string | undefined;
            if (first) {
                const old = blobUrlCache.get(first);
                if (old) URL.revokeObjectURL(old);
                blobUrlCache.delete(first);
            }
        }
        blobUrlCache.set(key, url);
        return url;
    } catch {
        return '';
    }
}

function syncCurrentTrackMeta(path: string): void {
    if (!path) return;
    const updated = localTracks.value.find((x) => x.path === path);
    if (!updated) return;
    if (currentTrack.value?.path !== path) return;
    const cur = currentTrack.value;
    const merged: LocalTrack = {
        ...cur,
        title: updated.title || cur.title,
        artist: updated.artist || cur.artist,
        duration: updated.duration || cur.duration,
        coverUrl: updated.coverUrl || cur.coverUrl,
        coverRaw: (updated as any).coverRaw || (cur as any).coverRaw || '',
        fileName: updated.fileName || cur.fileName,
    } as LocalTrack;
    // Solo actualizar si hay cambios para evitar loops
    if (merged.artist !== cur.artist || merged.coverUrl !== cur.coverUrl || (merged as any).coverRaw !== (cur as any).coverRaw || merged.duration !== cur.duration || merged.title !== cur.title) {
        currentTrack.value = merged;
        // También sincronizar en la cola
        const qi = queue.value.findIndex((x) => x.path === path);
        if (qi >= 0) {
            const cq = [...queue.value];
            cq[qi] = { ...cq[qi]!, artist: merged.artist, duration: merged.duration, coverUrl: merged.coverUrl, coverRaw: (merged as any).coverRaw, title: merged.title } as LocalTrack;
            queue.value = cq;
        }
        if (merged.duration && !duration.value) duration.value = merged.duration;
        updateMediaSession();
    }
}

async function loadTrack(index: number, autoPlay: boolean): Promise<void> {
    if (!queue.value.length) return;
    const idx = ((index % queue.value.length) + queue.value.length) % queue.value.length;
    const track0 = queue.value[idx];
    if (!track0) return;
    let track: LocalTrack = track0;
    // Si la pista en cola no tiene metadatos, intentar hidratar desde la biblioteca antes de reproducir
    const lib = localTracks.value.find((x) => x.path === track.path);
    if (lib && (!track.coverUrl || !track.artist || track.artist === 'Desconocido')) {
        track = { ...track, title: lib.title || track.title, artist: lib.artist || track.artist, duration: lib.duration || track.duration, coverUrl: lib.coverUrl || track.coverUrl, coverRaw: (lib as any).coverRaw || (track as any).coverRaw || '', fileName: lib.fileName || track.fileName } as LocalTrack;
        // Actualizar cola en memoria
        const cq = [...queue.value];
        cq[idx] = track;
        queue.value = cq;
    }
    const url = await blobUrlFor(track.path);
    if (!url) {
        loadError.value = `No se pudo leer "${track.title}". Verifica la carpeta en Ajustes → Música.`;
        return;
    }
    const a = ensureAudio();
    if (!a) return;
    stopAudio();
    loadError.value = '';
    isLoadingTrack.value = true;
    currentIndex.value = idx;
    currentTrack.value = track;
    currentTime.value = 0;
    duration.value = track.duration || 0;
    updateMediaSession();
    a.src = url;
    a.volume = volume.value;
    a.loop = playMode.value === 'repeat-one' || queue.value.length <= 1;
    a.load();
    if (autoPlay) {
        try { (window as any).__pauseBg?.(); } catch (_e) {}
        wantsPlay = true;
        void a.play().catch(() => {
            if (retryTimer !== null) window.clearTimeout(retryTimer);
            retryTimer = window.setTimeout(() => {
                retryTimer = null;
                if (wantsPlay && !playing.value) void a.play().catch(() => {});
            }, 700);
        });
        void addToHistory(track);
    }
    isLoadingTrack.value = false;
    schedulePersist();
    // Cargar metadatos reales y sincronizar el reproductor cuando terminen
    const loadPath = track.path;
    const loadDuration = track.duration;
    void ensureTrackMeta([track]).then(() => {
        syncCurrentTrackMeta(loadPath);
        // Si la pista ganó duración, actualizar el duration del audio si aún es 0
        const upd = localTracks.value.find((x) => x.path === loadPath);
        if (upd?.duration && (!duration.value || duration.value === loadDuration)) {
            duration.value = upd.duration;
        }
    });
}

// Integración Media Session (controles del sistema)
let sessionBound = false;
function bindMediaSession(): void {
    if (sessionBound) return;
    const ms = (navigator as any).mediaSession;
    if (!ms) return;
    sessionBound = true;
    ms.setActionHandler?.('play', () => { void togglePlay(); });
    ms.setActionHandler?.('pause', () => { void togglePlay(); });
    ms.setActionHandler?.('previoustrack', () => { void prev(); });
    ms.setActionHandler?.('nexttrack', () => { void next(); });
    ms.setActionHandler?.('seekbackward', () => {});
    ms.setActionHandler?.('seekforward', () => {});
}

function getSMTCSource(): string {
    try { return localStorage.getItem('stl_smtc_source') || 'auto'; } catch { return 'auto'; }
}
function updateMediaSession(): void {
    const ms = (navigator as any).mediaSession;
    if (!ms) return;
    const src = getSMTCSource();
    if (src === 'background') return;
    if (src === 'auto') {
        try {
            const bgPlaying = (window as any).__stlBgPlaying;
            if (bgPlaying) return;
        } catch (_e) {}
    }
    bindMediaSession();
    try {
        const t = currentTrack.value;
        ms.metadata = new MediaMetadata({
            title: t?.title || 'Sin reproducción',
            artist: t?.artist || 'Biblioteca local',
            album: 'StepLauncher',
            artwork: t?.coverUrl ? [{ src: t.coverUrl, sizes: '512x512' }] : [],
        });
    } catch (_e) {}
    ms.playbackState = playing.value ? 'playing' : 'paused';
}

// API pública

export function setQueue(tracks: LocalTrack[], startPath?: string): void {
    queue.value = [...tracks];
    if (shuffleMode.value === 'shuffle' || playMode.value === 'shuffle') buildShuffledQueue();
    syncLoop();
    if (startPath) {
        const idx = queue.value.findIndex((t) => t.path === startPath);
        if (idx >= 0) {
            currentIndex.value = idx;
            currentTrack.value = queue.value[idx] ?? null;
            schedulePersist();
            return;
        }
    }
    if (queue.value.length && currentIndex.value < 0) {
        currentIndex.value = 0;
        currentTrack.value = queue.value[0] ?? null;
    }
    if (!queue.value.length) {
        currentIndex.value = -1;
        currentTrack.value = null;
        stopAudio();
    }
    schedulePersist();
}

export function removeFromQueue(path: string): void {
    const rawIdx = queue.value.findIndex((t) => t.path === path);
    if (rawIdx < 0) return;
    const wasCurrent = currentIndex.value === rawIdx;
    queue.value.splice(rawIdx, 1);
    if (shuffledOrder.value.length) {
        const pos = shuffledOrder.value.indexOf(rawIdx);
        if (pos >= 0) shuffledOrder.value.splice(pos, 1);
        shuffledOrder.value = shuffledOrder.value.map((i) => (i > rawIdx ? i - 1 : i));
        if (shuffledOrder.value.length !== queue.value.length) buildShuffledQueue();
    }
    if (wasCurrent) {
        if (!queue.value.length) {
            currentIndex.value = -1;
            currentTrack.value = null;
            stopAudio();
        } else {
            const newIdx = Math.min(rawIdx, queue.value.length - 1);
            currentIndex.value = newIdx;
            currentTrack.value = queue.value[newIdx] ?? null;
        }
    } else if (rawIdx < currentIndex.value) {
        currentIndex.value -= 1;
    }
    syncLoop();
    schedulePersist();
}

export function moveInQueue(fromOrdered: number, toOrdered: number): void {
    if (fromOrdered === toOrdered) return;
    const order = getQueueOrder();
    const len = order.length;
    if (fromOrdered < 0 || fromOrdered >= len || toOrdered < 0 || toOrdered >= len) return;
    const isShuf = shuffleMode.value === 'shuffle' || playMode.value === 'shuffle';
    if (isShuf) {
        const newOrder = [...shuffledOrder.value];
        if (newOrder.length !== len) buildShuffledQueue();
        const arr = [...shuffledOrder.value];
        const [moved] = arr.splice(fromOrdered, 1);
        if (moved === undefined) return;
        arr.splice(toOrdered, 0, moved);
        shuffledOrder.value = arr;
    } else {
        const q = [...queue.value];
        const [moved] = q.splice(fromOrdered, 1);
        if (moved === undefined) return;
        q.splice(toOrdered, 0, moved);
        queue.value = q;
        if (currentIndex.value === fromOrdered) currentIndex.value = toOrdered;
        else if (fromOrdered < currentIndex.value && toOrdered >= currentIndex.value) currentIndex.value -= 1;
        else if (fromOrdered > currentIndex.value && toOrdered <= currentIndex.value) currentIndex.value += 1;
    }
    schedulePersist();
}

export function addNext(track: LocalTrack): void {
    if (!track?.path) return;
    const existing = queue.value.findIndex((t) => t.path === track.path);
    if (existing >= 0) {
        if (existing === currentIndex.value) return;
        queue.value.splice(existing, 1);
        if (existing < currentIndex.value) currentIndex.value -= 1;
        if (shuffledOrder.value.length) {
            const pos = shuffledOrder.value.indexOf(existing);
            if (pos >= 0) shuffledOrder.value.splice(pos, 1);
            shuffledOrder.value = shuffledOrder.value.map((i) => (i > existing ? i - 1 : i));
        }
    }
    const insertAt = currentIndex.value >= 0 ? currentIndex.value + 1 : queue.value.length;
    // Hidratar si es necesario (buscar en biblioteca)
    const lib = localTracks.value.find((x) => x.path === track.path);
    const toInsert: LocalTrack = lib ? { ...track, title: lib.title || track.title, artist: lib.artist || track.artist, coverUrl: lib.coverUrl || track.coverUrl } as LocalTrack : track;
    queue.value.splice(insertAt, 0, toInsert);
    if (shuffleMode.value === 'shuffle' || playMode.value === 'shuffle') {
        if (shuffledOrder.value.length !== queue.value.length - 1) buildShuffledQueue();
        else {
            shuffledOrder.value = shuffledOrder.value.map((i) => (i >= insertAt ? i + 1 : i));
            const order = getQueueOrder();
            const curOrderedPos = order.indexOf(currentIndex.value);
            const insertOrdered = curOrderedPos >= 0 ? curOrderedPos + 1 : shuffledOrder.value.length;
            shuffledOrder.value.splice(insertOrdered, 0, insertAt);
        }
    }
    syncLoop();
    schedulePersist();
}

export function getOrderedIndexForPath(path: string): number {
    const order = getQueueOrder();
    const tracks = order.map((i) => queue.value[i]);
    return tracks.findIndex((t) => t?.path === path);
}

export async function playTrack(track: LocalTrack, queueOverride?: LocalTrack[]): Promise<void> {
    // Intentar hidratar metadatos desde la biblioteca antes de reproducir
    let hydrated = track;
    const lib = localTracks.value.find((x) => x.path === track.path);
    if (lib) hydrated = { ...track, title: lib.title || track.title, artist: lib.artist || track.artist, duration: lib.duration || track.duration, coverUrl: lib.coverUrl || track.coverUrl, coverRaw: (lib as any).coverRaw || (track as any).coverRaw || '', fileName: lib.fileName || track.fileName } as LocalTrack;
    else {
        // Si no está en biblioteca (pista huérfana de playlist), intentar cargar metadatos igualmente
        try { await ensureTrackMeta([track]); } catch (_e) {}
        const after = localTracks.value.find((x) => x.path === track.path);
        if (after) hydrated = { ...track, title: after.title || track.title, artist: after.artist || track.artist, duration: after.duration || track.duration, coverUrl: after.coverUrl || track.coverUrl, coverRaw: (after as any).coverRaw || (track as any).coverRaw || '', fileName: after.fileName || track.fileName } as LocalTrack;
    }
    // Si la cola viene de la playlist, asegurarse de que toda la cola tenga al menos intento de carátulas (no bloqueante)
    if (queueOverride && queueOverride.length) {
        const mapped = queueOverride.map((t) => {
            const l = localTracks.value.find((x) => x.path === t.path);
            return l ? ({ ...t, title: l.title || t.title, artist: l.artist || t.artist, duration: l.duration || t.duration, coverUrl: l.coverUrl || t.coverUrl, coverRaw: (l as any).coverRaw || (t as any).coverRaw || '' } as LocalTrack) : t;
        });
        queue.value = mapped;
        // Disparar carga de carátulas para la cola visible sin bloquear
        const need = mapped.filter((x) => !x.coverUrl).slice(0, 8);
        if (need.length) void ensureTrackMeta(need as any);
        hydrated = mapped.find((x) => x.path === hydrated.path) ?? hydrated;
    } else if (!queue.value.length) {
        queue.value = [hydrated];
    } else if (!queue.value.some((t) => t.path === hydrated.path)) {
        queue.value = [hydrated, ...queue.value];
    }
    // Sincronizar el track a reproducir con la versión hidratada en cola
    const idx = queue.value.findIndex((t) => t.path === hydrated.path);
    await loadTrack(idx >= 0 ? idx : 0, true);
}

export async function playPath(path: string, queueOverride?: LocalTrack[]): Promise<void> {
    const list = queueOverride ?? queue.value;
    const t = list.find((x) => x.path === path);
    if (t) await playTrack(t, queueOverride);
}

export async function togglePlay(): Promise<void> {
    if (!queue.value.length && !currentTrack.value) return;
    if (currentIndex.value < 0 && queue.value.length) {
        await loadTrack(0, true);
        return;
    }
    const a = audio;
    if (!a || !a.src) {
        const idx = currentIndex.value >= 0 ? currentIndex.value : 0;
        await loadTrack(idx, true);
        return;
    }
    if (playing.value) {
        wantsPlay = false;
        a.pause();
    } else {
        try { (window as any).__pauseBg?.(); } catch (_e) {}
        wantsPlay = true;
        void a.play().catch(() => {});
    }
}

export async function next(): Promise<void> {
    if (!queue.value.length) return;
    const len = queue.value.length;
    const cur = currentIndex.value < 0 ? 0 : currentIndex.value;
    const nxt = nextIndex(playMode.value, cur, len);
    await loadTrack(nxt, true);
}

export async function prev(): Promise<void> {
    if (!queue.value.length) return;
    if (audio && currentTime.value > 3) {
        await seekTo(0);
        return;
    }
    const len = queue.value.length;
    let idx: number;
    if (playMode.value === 'shuffle') idx = prevShuffleIndex(currentIndex.value, len);
    else idx = currentIndex.value <= 0 ? len - 1 : currentIndex.value - 1;
    await loadTrack(idx, true);
}

export function setVolume(v: number): void {
    const clamped = Math.min(1, Math.max(0, Number(v) || 0));
    volume.value = clamped;
    if (audio) audio.volume = clamped;
    if (volumeTimer !== null) window.clearTimeout(volumeTimer);
    volumeTimer = window.setTimeout(persistVolume, 250);
}

export async function seekTo(t: number): Promise<void> {
    if (audio && Number.isFinite(t)) {
        try {
            audio.currentTime = Math.max(0, t);
            currentTime.value = Math.max(0, t);
        } catch (_e) {}
    }
}

export function setPlayMode(m: 'queue' | 'shuffle' | 'repeat-one'): void {
    if (m === 'shuffle' && playMode.value !== 'shuffle') buildShuffledQueue();
    playMode.value = m;
    // Sincronizar nuevos estados
    if (m === 'shuffle') shuffleMode.value = 'shuffle';
    else if (m === 'queue') shuffleMode.value = 'linear';
    if (m === 'repeat-one') repeatMode.value = 'one';
    else if (m === 'queue') repeatMode.value = 'queue';
    syncLoop();
}

export function toggleShuffleMode(): void {
    if (shuffleMode.value === 'linear') {
        shuffleMode.value = 'shuffle';
        playMode.value = 'shuffle';
        buildShuffledQueue();
    } else {
        shuffleMode.value = 'linear';
        if (playMode.value === 'shuffle') playMode.value = 'queue';
    }
    syncLoop();
}

export function toggleRepeatMode(): void {
    if (repeatMode.value === 'queue') {
        repeatMode.value = 'one';
        playMode.value = 'repeat-one';
    } else {
        repeatMode.value = 'queue';
        if (playMode.value === 'repeat-one') playMode.value = 'queue';
    }
    syncLoop();
}

export async function nextNext(): Promise<void> {
    if (!queue.value.length) return;
    // Salta una y va a la siguiente (avanza 2)
    const len = queue.value.length;
    if (len <= 2) return next();
    const cur = currentIndex.value < 0 ? 0 : currentIndex.value;
    const first = nextIndex(shuffleMode.value === 'shuffle' ? 'shuffle' : 'queue', cur, len);
    const second = nextIndex(shuffleMode.value === 'shuffle' ? 'shuffle' : 'queue', first, len);
    await loadTrack(second, true);
}

export async function prevPrev(): Promise<void> {
    if (!queue.value.length) return;
    const len = queue.value.length;
    if (len <= 2) return prev();
    const cur = currentIndex.value < 0 ? 0 : currentIndex.value;
    let first: number;
    if (shuffleMode.value === 'shuffle') first = prevShuffleIndex(cur, len);
    else first = cur <= 0 ? len - 1 : cur - 1;
    let second: number;
    if (shuffleMode.value === 'shuffle') second = prevShuffleIndex(first, len);
    else second = first <= 0 ? len - 1 : first - 1;
    await loadTrack(second, true);
}

export async function seekForward10(): Promise<void> {
    const d = duration.value || currentTrack.value?.duration || 0;
    const nt = Math.min(d || 1e9, (audio?.currentTime || currentTime.value) + 10);
    await seekTo(nt);
}

export async function seekBackward10(): Promise<void> {
    const nt = Math.max(0, (audio?.currentTime || currentTime.value) - 10);
    await seekTo(nt);
}

export function stop(): void {
    stopAudio();
    currentIndex.value = -1;
    currentTrack.value = null;
    currentTime.value = 0;
    duration.value = 0;
    updateMediaSession();
}

export const progressPct = computed(() => {
    const d = duration.value || currentTrack.value?.duration || 0;
    if (!d || !Number.isFinite(d) || d <= 0) return 0;
    return Math.min(100, Math.max(0, (currentTime.value / d) * 100));
});

export const displayCover = computed(() => (currentTrack.value as any)?.coverRaw || currentTrack.value?.coverUrl || '');
export const displayMeta = computed(() => {
    if (!currentTrack.value) return null;
    return {
        title: currentTrack.value.title,
        artist: currentTrack.value.artist,
        duration: duration.value || currentTrack.value.duration,
        coverUrl: (currentTrack.value as any).coverRaw || currentTrack.value.coverUrl,
        fileName: currentTrack.value.fileName,
        path: currentTrack.value.path,
    };
});

export const queueTracks = computed(() => {
    const order = getQueueOrder();
    return order.map((i) => queue.value[i]).filter(Boolean) as LocalTrack[];
});

restoreVolume();
if (typeof window !== 'undefined') {
    try { (window as any).__pauseLauncher = () => { try { if (audio) audio.pause(); playing.value = false; updateMediaSession(); } catch (_e) {} }; } catch (_e) {}
}
// Inicialización perezosa del PlayerStore: la restauración de cola solo
// debe ocurrir cuando el usuario entra al panel de Música, no al arrancar.
let _playerInitDone = false;
let _playerFolderHandler: (() => void) | null = null;
export function ensurePlayerStoreInitialized(): void {
    if (_playerInitDone || typeof window === 'undefined') return;
    _playerInitDone = true;
    // Reintento de restauración de cola cuando cambia la carpeta musical
    _playerFolderHandler = () => setTimeout(() => void tryRestoreQueue(), 600);
    window.addEventListener('stl:music-folder-changed', _playerFolderHandler);
    // Restauración inicial diferida (ya estamos dentro del panel)
    setTimeout(() => void tryRestoreQueue(), 300);
}
watch(volume, (v) => { if (audio) audio.volume = v; });

// Señal global para que el Media Session de fondo ceda ante el reproductor
function syncGlobalMediaFlag(): void {
    try {
        (window as any).__stlLauncherPlaying = playing.value;
        (window as any).__stlLauncherHasTrack = !!currentTrack.value;
    } catch (_e) {}
}
watch(playing, syncGlobalMediaFlag);
watch(currentTrack, syncGlobalMediaFlag);
syncGlobalMediaFlag();

// Si la biblioteca actualiza metadatos (cover/artist) del track en reproducción, reflejarlo al instante
watch(localTracks, () => {
    if (currentTrack.value) syncCurrentTrackMeta(currentTrack.value.path);
}, { deep: false });

watch(playMode, syncLoop);
watch(shuffleMode, syncLoop);
watch(repeatMode, syncLoop);
watch(queue, schedulePersist, { deep: false });
watch(currentIndex, schedulePersist);
watch(currentTrack, () => { updateMediaSession(); syncGlobalMediaFlag(); });
watch(playing, () => { updateMediaSession(); });
// Eager: cuando la cola cambia (restaurada o desde playlist), precargar carátulas sin esperar a reproducir
watch(queue, (q) => {
    const need = q.filter((t) => !t.coverUrl).slice(0, 8);
    if (need.length) void ensureTrackMeta(need as any);
}, { deep: false });

// Restaurar cola temporal desde launcher_music_nowplaying.json si existe y la cola está vacía
async function tryRestoreQueue(): Promise<void> {
    let q: any = null;
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.GetNowPlayingQueue === 'function') {
            q = await mod.GetNowPlayingQueue();
        } else {
            throw new Error('no binding');
        }
    } catch {
        try {
            const raw = localStorage.getItem('stl_nowplaying_queue');
            if (raw) q = JSON.parse(raw);
        } catch (_e) {}
    }
    if (q && Array.isArray(q.tracks) && q.tracks.length && !queue.value.length) {
        let tries = 0;
        const attempt = () => {
            if (localTracks.value.length) {
                const map = new Map(localTracks.value.map((t) => [t.path, t]));
                const restored = (q.tracks as string[]).map((p: string) => map.get(p)).filter(Boolean) as LocalTrack[];
                if (restored.length) {
                    queue.value = restored;
                    const idx = typeof q.currentIndex === 'number' && q.currentIndex >= 0 && q.currentIndex < restored.length ? q.currentIndex : (typeof q.idx === 'number' ? q.idx : 0);
                    currentIndex.value = idx;
                    currentTrack.value = restored[idx] ?? null;
                    syncLoop();
                }
            } else if (tries++ < 12) {
                setTimeout(attempt, 700);
            }
        };
        attempt();
    }
}
// tryRestoreQueue ya no se auto-ejecuta al importar; lo dispara ensurePlayerStoreInitialized()
// cuando el usuario entra por primera vez al panel de Música.
