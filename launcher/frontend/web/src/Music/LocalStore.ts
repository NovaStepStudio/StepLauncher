import { ref, shallowRef } from 'vue';

// Estado ligero: solo página visible + total, Go es autoridad.
export interface LocalTrack {
    id: string;
    path: string;
    fileName: string;
    title: string;
    artist: string;
    duration: number;
    coverUrl: string;
    coverRaw?: string;
    hasCover?: boolean;
    album?: string;
}

// Biblioteca indexada en Go; frontend solo consume páginas.
export const localTracks = shallowRef<LocalTrack[]>([]); // compat: alias a pagedTracks
export const pagedTracks = shallowRef<LocalTrack[]>([]);
export const totalTracks = ref(0);
export const isScanning = ref(false);
export const scanError = ref('');
export const scanProgress = ref<{ current: number; total: number; folder: string; currentFile: string; scanning: boolean; discovered?: number; processed?: number; skipped?: number; added?: number; updated?: number; removed?: number; errors?: number }>({ current: 0, total: 0, folder: '', currentFile: '', scanning: false });

let pollTimer: any = null;
function startPolling(): void {
    if (pollTimer) return;
    pollTimer = setInterval(async () => {
        try {
            const { GetMusicScanProgress } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
            const p: any = await GetMusicScanProgress();
            scanProgress.value = {
                current: Number(p.current) || 0,
                total: Number(p.total) || 0,
                folder: String(p.folder || ''),
                currentFile: String(p.currentFile || ''),
                scanning: !!p.scanning,
                discovered: Number(p.discovered) || 0,
                processed: Number(p.processed) || 0,
                skipped: Number(p.skipped) || 0,
                added: Number(p.added) || 0,
                updated: Number(p.updated) || 0,
                removed: Number(p.removed) || 0,
                errors: Number(p.errors) || 0,
            };
        } catch (_e) {}
    }, 200);
}
function stopPolling(): void {
    if (pollTimer) { clearInterval(pollTimer); pollTimer = null; }
}

// Carga inicial: solo lee del índice, no escanea a menos que se fuerce o índice vacío.
// PROHIBIDO: nunca usar carpeta por defecto — solo escanea si hay carpetas explícitas del usuario.
export async function loadLocalLibrary(forceScan = false): Promise<void> {
    // Si no se fuerza y ya hay datos en índice, solo cargar página
    if (!forceScan) {
        try {
            const { GetMusicLibraryStats } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
            const stats: any = await GetMusicLibraryStats().catch(() => null);
            if (stats && Number(stats.totalTracks) > 0) {
                await fetchMusicPage(0, 20, '', 'thumb');
                return;
            }
            // Si hay carpetas pero índice vacío, sí escanear
            const { GetMusicFolders } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
            const folders: any = await GetMusicFolders().catch(() => []);
            if (!folders || !folders.length) {
                await fetchMusicPage(0, 20, '', 'thumb');
                return;
            }
            // Índice vacío pero hay carpetas -> forzar scan
            forceScan = true;
        } catch {
            // Fallback a scan si no se pudo verificar
        }
    }
    isScanning.value = true;
    scanError.value = '';
    scanProgress.value = { current: 0, total: 0, folder: '', currentFile: '', scanning: true };
    startPolling();
    try {
        if (forceScan) {
            const { ScanAllMusicFolders } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
            await ScanAllMusicFolders().catch(() => null);
        }
        await fetchMusicPage(0, 20, '', 'thumb');
    } catch (e: any) {
        scanError.value = e?.message ?? 'No se pudo cargar';
    } finally {
        isScanning.value = false;
        setTimeout(() => stopPolling(), 1200);
    }
}

export async function fetchMusicPage(offset: number, limit: number, query = '', coverVariant: 'thumb' | 'raw' | 'both' | 'none' = 'thumb'): Promise<{ tracks: LocalTrack[]; total: number }> {
    try {
        const { GetMusicTracksPaged } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        const page: any = await GetMusicTracksPaged(offset, limit, coverVariant, query);
        if (page && Array.isArray(page.tracks)) {
            // Limpia título/artista tal cual debe verse — elimina BOM/FFFD y controles que se ven como "☐"
            const clean = (s: any): string => {
                let str = String(s ?? '').trim();
                // elimina BOM, reemplazo y controles/zero-width
                str = str.replace(/[\uFEFF\uFFFD]/g, '');
                str = str.replace(/[\x00-\x1F\x7F-\x9F]/g, '');
                str = str.replace(/[\u200B\u200C\u200D\u2060\u00AD\u034F]/g, '');
                return str.trim();
            };
            const tracks: LocalTrack[] = page.tracks.map((t: any) => ({
                id: t.path,
                path: t.path,
                fileName: t.fileName || (t.path.split(/[\\/]/).pop() || ''),
                title: clean(t.title) || clean(t.fileName) || 'Desconocido',
                artist: clean(t.artist) || 'Desconocido',
                duration: Number(t.duration) || 0,
                coverUrl: t.coverThumb || t.coverRaw || '',
                coverRaw: t.coverRaw || t.coverThumb || '',
                hasCover: !!t.hasCover,
                album: clean(t.album) || '',
            }));
            pagedTracks.value = tracks;
            localTracks.value = tracks; // compat: localTracks = página actual
            totalTracks.value = Number(page.total) || tracks.length;
            return { tracks, total: totalTracks.value };
        }
    } catch (_e) {}
    pagedTracks.value = [];
    localTracks.value = [];
    totalTracks.value = 0;
    return { tracks: [], total: 0 };
}

// No-ops: Go es fuente única, frontend no parsea audio.
export async function ensureTrackMeta(_tracks: LocalTrack[]): Promise<void> {
    // Intencionalmente vacío: metadata viene de Go index via fetchMusicPage
    return;
}
export async function ensureCovers(_tracks: LocalTrack[]): Promise<void> {
    // Covers se piden bajo demanda via CoverCache.getCachedCoverUrl (lazy)
    return;
}
export async function refreshCovers(tracks: LocalTrack[]): Promise<void> {
    // Invalida covers de esas pistas y recarga página
    try {
        const { RefreshCachedCover } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        for (const t of tracks) {
            try { await RefreshCachedCover(t.path); } catch (_e) {}
        }
    } catch (_e) {}
}
export async function refreshCoverFor(path: string): Promise<void> {
    try {
        const { RefreshCachedCover } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        await RefreshCachedCover(path);
    } catch (_e) {}
}

// Inicialización perezosa: solo cuando entra al panel Música
let _musicInitDone = false;
let _musicFolderHandler: (() => void) | null = null;
let _musicFocusHandler: (() => void) | null = null;
export function ensureMusicStoreInitialized(): void {
    if (_musicInitDone || typeof window === 'undefined') return;
    _musicInitDone = true;
    _musicFolderHandler = () => { void loadLocalLibrary(); };
    _musicFocusHandler = () => {
        if (!isScanning.value && !pagedTracks.value.length) void loadLocalLibrary();
    };
    window.addEventListener('stl:music-folder-changed', _musicFolderHandler);
    window.addEventListener('focus', _musicFocusHandler);
}
export function getDemoTracks(count = 1000): LocalTrack[] {
    const colors = ['#5ee9a8', '#a78bfa', '#f472b6', '#60a5fa', '#fbbf24', '#34d399'];
    return Array.from({ length: count }, (_, i) => {
        const bg = colors[i % colors.length]!;
        const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200"><rect width="100%" height="100%" rx="24" fill="${bg}"/><text x="50%" y="54%" dominant-baseline="middle" text-anchor="middle" font-family="Inter, Arial, sans-serif" font-size="42" font-weight="800" fill="#111">${(i % 100) + 1}</text></svg>`;
        return {
            id: `demo-${i + 1}`,
            path: `demo/${i + 1}.mp3`,
            fileName: `pista_${i + 1}.mp3`,
            title: `Canción de prueba ${i + 1}`,
            artist: ['NovaStep', 'StepBeats', 'Lofi Step', 'Pixel Wave'][i % 4] as string,
            duration: 90 + (i % 210),
            coverUrl: `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`,
        };
    });
}
