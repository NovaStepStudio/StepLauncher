<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { IconSearch, IconFolderPlus, IconLibrary, IconPlaylist, IconMusic, IconDisc, IconTrash, IconPencil, IconPlayerPlay, IconChevronLeft, IconChevronRight, IconStar, IconPin, IconArrowLeft, IconRefresh, IconArrowsShuffle, IconClock, IconCalendar, IconChevronDown, IconChevronUp, IconPhoto } from '@tabler/icons-vue';
import { localTracks, totalTracks, ensureCovers, ensureTrackMeta, refreshCovers } from '../LocalStore';
import { playlists, loadingPlaylists, deletePlaylist, updatePlaylist } from '../PlaylistStore';
import { playTrack, currentTrack } from '../PlayerStore';
import { useCoverPalette, getPaletteColor } from '@/Common/Composables/useCoverPalette';
import PlaylistForm from './PlaylistForm.vue';
import PlaylistCard from './PlaylistCard.vue';
import { GetMusicPanelConfig } from '@wailsjs/StepLauncher/internal/Services/Music/musicservice';

const emit = defineEmits<{ (e: 'create-playlist'): void }>();

// Estado: vista grid vs detalle subpanel
const selectedId = ref<string>('');
const query = ref('');
const bibPage = ref(1);
const bibPageSize = ref(20);
const showEdit = ref(false);
const editTarget = ref<any>(null);
const isRefreshing = ref(false);
const playlistSearch = ref('');
const isLoadingPlaylist = ref(false);
const heroCollapsed = ref(false);
const sortGrid = ref('reciente');
const sortDetail = ref('orden');

watch(playlists, (list) => {
    if (!list.length) selectedId.value = '';
    else if (selectedId.value && !list.some((p) => p.id === selectedId.value)) selectedId.value = '';
}, { immediate: true });

const viewMode = computed(() => selectedId.value ? 'detail' : 'grid');
const selectedPlaylist = computed(() => playlists.value.find((p) => p.id === selectedId.value) ?? null);

const playlistTracksCache = ref<Map<string, any>>(new Map());
const bibliotecaTracksRaw = ref<any[]>([]);
const pageCoverMap = ref<Map<string, string>>(new Map());

// Carga ligera: solo metadata sin covers (variant 'none') para consulta rápida sin decodificar imágenes
async function loadPlaylistTracks(pl: any) {
    if (!pl || !Array.isArray(pl.trackPaths) || !pl.trackPaths.length) {
        bibliotecaTracksRaw.value = [];
        pageCoverMap.value.clear();
        isLoadingPlaylist.value = false;
        return;
    }
    isLoadingPlaylist.value = true;
    pageCoverMap.value.clear();
    try {
        const paths: string[] = pl.trackPaths;
        // 1) Batch ligero sin covers para tener títulos/duración instantáneamente (usa índice, no decodifica imágenes)
        try {
            const { GetMusicTracksBatch } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
            const batch: any = await GetMusicTracksBatch(paths, 'none');
            if (Array.isArray(batch) && batch.length) {
                const map = new Map<string, any>();
                batch.forEach((t: any) => map.set(t.path, t));
                const out: any[] = paths.map((p: string) => {
                    const t = map.get(p);
                    if (t) return { id: t.path, path: t.path, fileName: t.fileName || p.split(/[\\/]/).pop() || p, title: t.title || t.fileName || p.split(/[\\/]/).pop() || p, artist: t.artist || 'Desconocido', duration: Number(t.duration) || 0, coverUrl: '', hasCover: !!t.hasCover, album: t.album || '' };
                    const existing = playlistTracksCache.value.get(p);
                    if (existing) return existing;
                    return { id: p, path: p, fileName: p.split(/[\\/]/).pop() || p, title: p.split(/[\\/]/).pop()?.replace(/\.[^.]+$/, '') || p, artist: 'Desconocido', duration: 0, coverUrl: '', hasCover: false };
                });
                bibliotecaTracksRaw.value = out;
                // No retornamos: dejamos que el watcher de paginación cargue thumbs solo de la página visible
                return;
            }
        } catch (_e) {}
        // Fallback: usar localTracks si batch falla
        const trackMap = new Map<string, any>();
        localTracks.value.forEach((t) => trackMap.set(t.path, t));
        const list = paths.map((p) => trackMap.get(p)).filter(Boolean) as any[];
        const missing = paths.filter((p) => !trackMap.has(p)).map((p) => ({
            id: p, path: p, fileName: p.split(/[\\/]/).pop() || p, title: p.split(/[\\/]/).pop()?.replace(/\.[^.]+$/, '') || p, artist: 'Desconocido', duration: 0, coverUrl: ''
        } as any));
        bibliotecaTracksRaw.value = [...list, ...missing];
    } finally {
        isLoadingPlaylist.value = false;
    }
}

async function fetchPageCovers(pagePaths: string[]): Promise<void> {
    if (!pagePaths.length) return;
    // Pedir thumb para los visibles — no filtrar por hasCover del batch 'none' (puede venir false)
    // Se pide para todo lo no cacheado; el backend responde vacío si no hay carátula
    const toFetch = pagePaths.filter((p) => !pageCoverMap.value.has(p));
    if (!toFetch.length) return;
    try {
        const { GetMusicTracksBatch } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        const batch: any = await GetMusicTracksBatch(toFetch, 'thumb');
        if (Array.isArray(batch)) {
            let mutated = false;
            batch.forEach((t: any) => {
                const thumb = t.coverThumb || t.coverRaw || '';
                if (thumb) {
                    pageCoverMap.value.set(t.path, thumb);
                    playlistTracksCache.value.set(t.path, { ...t, coverUrl: thumb });
                    mutated = true;
                } else if (t.hasCover === false) {
                    // Marcar como sin carátula para no re-pedir
                    pageCoverMap.value.set(t.path, '');
                    mutated = true;
                }
            });
            // ref<Map> no es reactivo a .set(), forzar actualización
            if (mutated) {
                pageCoverMap.value = new Map(pageCoverMap.value);
                playlistTracksCache.value = new Map(playlistTracksCache.value);
            }
        }
    } catch (_e) {}
}
watch(() => selectedPlaylist.value?.id, (id) => {
    if (id) void loadPlaylistTracks(selectedPlaylist.value);
    else bibliotecaTracksRaw.value = [];
}, { immediate: true });
watch(() => (selectedPlaylist.value as any)?.trackPaths, () => {
    if (selectedPlaylist.value) void loadPlaylistTracks(selectedPlaylist.value);
}, { deep: true });

const bibliotecaTracks = computed(() => {
    const q = query.value.trim().toLowerCase();
    let all = bibliotecaTracksRaw.value;
    if (q) all = all.filter((t) => t.title.toLowerCase().includes(q) || t.artist.toLowerCase().includes(q) || t.fileName.toLowerCase().includes(q));
    const order = sortDetail.value;
    if (order === 'orden') return all;
    const arr = [...all];
    switch (order) {
        case 'titulo-asc': return arr.sort((a,b)=> String(a.title||'').localeCompare(String(b.title||''),'es',{sensitivity:'base'}));
        case 'titulo-desc': return arr.sort((a,b)=> String(b.title||'').localeCompare(String(a.title||''),'es',{sensitivity:'base'}));
        case 'artista-asc': return arr.sort((a,b)=> String(a.artist||'').localeCompare(String(b.artist||''),'es',{sensitivity:'base'}));
        case 'artista-desc': return arr.sort((a,b)=> String(b.artist||'').localeCompare(String(a.artist||''),'es',{sensitivity:'base'}));
        case 'duracion-asc': return arr.sort((a,b)=> (Number(a.duration)||0) - (Number(b.duration)||0));
        case 'duracion-desc': return arr.sort((a,b)=> (Number(b.duration)||0) - (Number(a.duration)||0));
        default: return all;
    }
});

// UI ligera: usa solo cache de Go (TotalDuration/TrackCount). No sumar duraciones en frontend (O(n) + I/O)
const totalDuration = computed(() => {
    const cached = (selectedPlaylist.value as any)?.totalDuration;
    if (typeof cached === 'number' && cached >= 0) return cached;
    return 0;
});
const validCount = computed(() => {
    const cached = (selectedPlaylist.value as any)?.trackCount;
    if (typeof cached === 'number' && cached >= 0) return cached;
    const total = (selectedPlaylist.value as any)?.trackPaths?.length ?? bibliotecaTracks.value.length;
    return total;
});
const missingCount = computed(() => {
    const total = (selectedPlaylist.value as any)?.trackPaths?.length ?? bibliotecaTracks.value.length;
    return Math.max(0, total - validCount.value);
});

function formatDuration(sec: number): string {
    if (!Number.isFinite(sec) || sec <= 0) return '—';
    const m = Math.floor(sec / 60), s = Math.floor(sec % 60);
    return `${m}:${String(s).padStart(2,'0')}`;
}
function formatLongDuration(sec: number): string {
    if (!sec || !Number.isFinite(sec) || sec <= 0) return '—';
    const d = Math.floor(sec / 86400);
    const h = Math.floor((sec % 86400) / 3600);
    const m = Math.floor((sec % 3600) / 60);
    const s = Math.floor(sec % 60);
    if (d > 0) return `${d}d ${h}h ${m}m`;
    if (h > 0) return `${h}h ${m}m ${s}s`;
    if (m > 0) return `${m}min ${s}s`;
    return `${s}s`;
}
function formatDate(iso: string): string {
    if (!iso) return '—';
    try { const d = new Date(iso); return d.toLocaleDateString('es-ES', { day: '2-digit', month: 'short', year: 'numeric' }); } catch { return iso; }
}

const bibTotalPages = computed(() => Math.max(1, Math.ceil(bibliotecaTracks.value.length / bibPageSize.value)));
const bibRangeStart = computed(() => bibliotecaTracks.value.length === 0 ? 0 : (bibPage.value - 1) * bibPageSize.value + 1);
const bibRangeEnd = computed(() => Math.min(bibPage.value * bibPageSize.value, bibliotecaTracks.value.length));
// Paginación ligera: slice + merge con covers cacheados de la página (baja resolución solo para visibles)
const bibliotecaPaginatedRaw = computed(() => {
    const s = (bibPage.value - 1) * bibPageSize.value;
    return bibliotecaTracks.value.slice(s, s + bibPageSize.value);
});
const bibliotecaPaginated = computed(() => {
    return bibliotecaPaginatedRaw.value.map((t: any) => ({
        ...t,
        coverUrl: pageCoverMap.value.get(t.path) || t.coverUrl || '',
    }));
});
const bibPageItems = computed<(number|'ellipsis')[]>(() => {
    const total = bibTotalPages.value, cur = bibPage.value;
    if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);
    const items: (number|'ellipsis')[] = [1];
    const lo = Math.max(2, cur - 1), hi = Math.min(total - 1, cur + 1);
    if (lo > 2) items.push('ellipsis');
    for (let p = lo; p <= hi; p++) items.push(p);
    if (hi < total - 1) items.push('ellipsis');
    items.push(total);
    return items;
});

watch([query, selectedId, sortDetail], () => bibPage.value = 1);
watch(bibliotecaPaginatedRaw, (list) => {
    if (list.length) void fetchPageCovers(list.map((t: any) => t.path));
}, { immediate: true });

function goToBibPage(p: number): void { if (p >= 1 && p <= bibTotalPages.value) bibPage.value = p; }

async function onPlay(t: any): Promise<void> {
    const track = bibliotecaTracksRaw.value.find((x) => x.path === t.path) || t;
    const queue = bibliotecaTracksRaw.value.filter((x: any) => x.artist !== 'No encontrada') as any;
    if (track) await playTrack(track, queue.length ? queue : [track]);
    else if (t.path) {
        await playTrack({ id: t.path, path: t.path, fileName: t.fileName, title: t.title, artist: t.artist, duration: t.duration, coverUrl: '' } as any, queue);
    }
}
async function onPlayAll(shuffle = false): Promise<void> {
    const queue = bibliotecaTracksRaw.value.filter((x: any) => x.artist !== 'No encontrada') as any;
    if (!queue.length) return;
    let toPlay = queue;
    if (shuffle) {
        toPlay = [...queue].sort(() => Math.random() - 0.5);
    }
    await playTrack(toPlay[0], toPlay);
}
async function onDelete(id: string): Promise<void> {
    await deletePlaylist(id);
    selectedId.value = '';
}
function onEdit(pl: any): void { editTarget.value = pl; showEdit.value = true; }
async function handleEditSubmit(data: { title: string; favorite: boolean; pinned: boolean; color: string; cover: string; tracks: string[] }): Promise<void> {
    if (!editTarget.value) return;
    await updatePlaylist(editTarget.value.id, {
        title: data.title,
        favorite: data.favorite,
        pinned: data.pinned,
        customColor: data.color,
        customCover: data.cover,
        trackPaths: data.tracks,
    } as any);
    showEdit.value = false;
}
function openPlaylist(id: string): void { selectedId.value = id; window.scrollTo({ top: 0 }); }
function closeDetail(): void { selectedId.value = ''; query.value = ''; bibPage.value = 1; }
async function playPlaylistById(id: string): Promise<void> {
    const pl: any = playlists.value.find((p) => p.id === id);
    if (!pl || !Array.isArray(pl.trackPaths) || !pl.trackPaths.length) return;
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        const batch: any = await mod.GetMusicTracksBatch(pl.trackPaths, 'none');
        let queue: any[] = [];
        if (Array.isArray(batch) && batch.length) {
            const map = new Map<string, any>(); batch.forEach((t: any) => map.set(t.path, t));
            queue = pl.trackPaths.map((p: string) => {
                const t = map.get(p);
                if (t) return { id: t.path, path: t.path, fileName: t.fileName || p.split(/[\\/]/).pop() || p, title: t.title || t.fileName || p.split(/[\\/]/).pop() || p, artist: t.artist || 'Desconocido', duration: Number(t.duration) || 0, coverUrl: '' };
                return { id: p, path: p, fileName: p.split(/[\\/]/).pop() || p, title: p.split(/[\\/]/).pop()?.replace(/\.[^.]+$/, '') || p, artist: 'Desconocido', duration: 0, coverUrl: '' };
            });
        } else {
            queue = pl.trackPaths.map((p: string) => ({ id: p, path: p, fileName: p.split(/[\\/]/).pop() || p, title: p.split(/[\\/]/).pop()?.replace(/\.[^.]+$/, '') || p, artist: 'Desconocido', duration: 0, coverUrl: '' }));
        }
        if (queue.length) await playTrack(queue[0], queue as any);
    } catch {}
}
function editPlaylistById(id: string): void {
    const pl: any = playlists.value.find((p) => p.id === id);
    if (pl) { editTarget.value = pl; showEdit.value = true; }
}

async function onRefreshCovers(): Promise<void> {
    if (!selectedPlaylist.value || isRefreshing.value) return;
    isRefreshing.value = true;
    try {
        const tracks = bibliotecaTracks.value.filter((t: any) => t.artist !== 'No encontrada') as any;
        if (tracks.length) await refreshCovers(tracks);
    } finally { isRefreshing.value = false; }
}

// Grid filtros + ordenar por
const filteredPlaylists = computed(() => {
    const q = playlistSearch.value.trim().toLowerCase();
    if (!q) return playlists.value;
    return playlists.value.filter((pl: any) => pl.title.toLowerCase().includes(q));
});

const sortedPlaylists = computed(() => {
    const list = [...filteredPlaylists.value];
    const order = sortGrid.value;
    if (order === 'reciente') {
        // Ancladas primero, luego favoritas, luego por updatedAt descendente
        return list.sort((a: any, b: any) => {
            if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
            if (a.favorite !== b.favorite) return a.favorite ? -1 : 1;
            const da = a.updatedAt || a.createdAt || '';
            const db = b.updatedAt || b.createdAt || '';
            return String(db).localeCompare(String(da));
        });
    }
    if (order === 'nombre-asc') return list.sort((a:any,b:any)=> String(a.title||'').localeCompare(String(b.title||''),'es',{sensitivity:'base'}));
    if (order === 'nombre-desc') return list.sort((a:any,b:any)=> String(b.title||'').localeCompare(String(a.title||''),'es',{sensitivity:'base'}));
    if (order === 'pistas-asc') return list.sort((a:any,b:any)=> (Number(a.trackPaths?.length||a.trackCount||0) - Number(b.trackPaths?.length||b.trackCount||0)));
    if (order === 'pistas-desc') return list.sort((a:any,b:any)=> (Number(b.trackPaths?.length||b.trackCount||0) - Number(a.trackPaths?.length||a.trackCount||0)));
    if (order === 'duracion-asc') return list.sort((a:any,b:any)=> (Number(a.totalDuration||0) - Number(b.totalDuration||0)));
    if (order === 'duracion-desc') return list.sort((a:any,b:any)=> (Number(b.totalDuration||0) - Number(a.totalDuration||0)));
    return list;
});

const detailCustomBlob = ref('');
const colorMode = ref('vibrant');
const coverStyle = ref('square');
onMounted(async () => {
    try {
        const cfg: any = await GetMusicPanelConfig();
        colorMode.value = cfg?.colorMode || 'vibrant';
        coverStyle.value = cfg?.coverStyle || 'square';
    } catch (_e) {}
});
if (typeof window !== 'undefined') {
    window.addEventListener('stl:music-folder-changed', async () => {
        try {
            const cfg: any = await GetMusicPanelConfig();
            colorMode.value = cfg?.colorMode || 'vibrant';
            coverStyle.value = cfg?.coverStyle || 'square';
        } catch (_e) {}
    });
}
const detailColor = computed(() => getPaletteColor(detailPalette.value, colorMode.value));

function isDetailDataOrBlob(u: string): boolean {
    return u.startsWith('data:') || u.startsWith('blob:') || u.startsWith('http://') || u.startsWith('https://');
}
watch(() => (selectedPlaylist.value as any)?.customCover, async (raw: string) => {
    const prev = detailCustomBlob.value;
    if (prev && prev.startsWith('blob:')) { try { URL.revokeObjectURL(prev); } catch (_e) {} }
    detailCustomBlob.value = '';
    const r = String(raw || '').trim();
    if (!r) return;
    if (isDetailDataOrBlob(r)) { detailCustomBlob.value = r; return; }
    try {
        const { ReadAbsoluteFile } = await import('@wailsjs/StepLauncher/internal/Services/System/systemservice');
        const data: any = await ReadAbsoluteFile(r).catch(() => null);
        if (!data) { detailCustomBlob.value = r; return; }
        let bytes: Uint8Array | null = null;
        if (typeof data === 'string') {
            const bin = atob(data as string);
            bytes = new Uint8Array(bin.length);
            for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
        } else if (data instanceof Uint8Array) bytes = data;
        else if (Array.isArray(data)) bytes = new Uint8Array(data as any);
        if (!bytes || bytes.length === 0) { detailCustomBlob.value = r; return; }
        const ext = r.split('.').pop()?.toLowerCase() ?? 'jpeg';
        const mime = ext === 'png' ? 'image/png' : ext === 'webp' ? 'image/webp' : ext === 'gif' ? 'image/gif' : 'image/jpeg';
        const blob = new Blob([bytes as BlobPart], { type: mime });
        const url = URL.createObjectURL(blob);
        detailCustomBlob.value = url;
    } catch {
        detailCustomBlob.value = r;
    }
}, { immediate: true });

// Paleta para detalle: si hay customCover usa su blob, si no usa previewCovers cacheado (instantáneo) o primera carátula cargada
const detailCoverUrl = computed(() => {
    if (detailCustomBlob.value) return detailCustomBlob.value;
    const preview: string[] | undefined = (selectedPlaylist.value as any)?.previewCovers;
    if (preview && preview.length && preview[0]) return preview[0] as string;
    for (const t of bibliotecaTracks.value) {
        const c = pageCoverMap.value.get(t.path) || (t as any).coverUrl;
        if (c) return c;
    }
    return '';
});
const { palette: detailPalette } = useCoverPalette(() => detailCoverUrl.value);

// Cover para mosaic en detalle hero - usa previewCovers cacheado si existe (sin I/O), si no pide las 4 visibles
const detailFirstFour = computed(() => bibliotecaTracks.value.slice(0, 4));
const detailFirstFourDisplay = computed(() => {
    const preview: string[] | undefined = (selectedPlaylist.value as any)?.previewCovers;
    if (preview && preview.length) {
        return detailFirstFour.value.map((t: any, i: number) => ({
            ...t,
            coverUrl: preview[i] || pageCoverMap.value.get(t.path) || t.coverUrl || '',
        }));
    }
    return detailFirstFour.value.map((t: any) => ({
        ...t,
        coverUrl: pageCoverMap.value.get(t.path) || t.coverUrl || '',
    }));
});
watch(detailFirstFour, (list) => {
    const preview: string[] | undefined = (selectedPlaylist.value as any)?.previewCovers;
    if (preview && preview.length) return;
    if (list.length) void fetchPageCovers(list.map((t: any) => t.path));
}, { immediate: true });

function isCurrent(t: any): boolean { return currentTrack.value?.path === t.path; }

// Placeholder para covers que tardan
const loadingCovers = ref<Set<string>>(new Set());
function isCoverLoading(path: string): boolean { return loadingCovers.value.has(path); }
</script>

<template>
    <div class="MusicSection">
        <!-- MODO GRID: tarjetas de playlists -->
        <template v-if="viewMode==='grid'">
            <div class="MusicBibHeader">
                <div class="MusicBibHeader_Left">
                    <h3><IconLibrary :size="18" stroke="2" /> Tus listas de reproducción</h3>
                    <p v-if="playlists.length">{{ playlists.length }} lista{{ playlists.length===1 ? '' : 's' }} de reproducción · {{ totalTracks || localTracks.length }} pistas en biblioteca</p>
                    <p v-else>Crea tu primera colección con lo que ya tienes</p>
                </div>
                <div class="MusicBibHeader_Actions">
                    <label v-if="playlists.length" class="MusicSearch is-small"><IconSearch :size="14" stroke="2" /><input v-model="playlistSearch" type="text" placeholder="Buscar lista de reproducción…" spellcheck="false" /></label>
                    <label v-if="playlists.length" class="MusicToolbar_Sort"><select v-model="sortGrid" class="SsSel"><option value="reciente">Orden: Reciente</option><option value="nombre-asc">Nombre A → Z</option><option value="nombre-desc">Nombre Z → A</option><option value="pistas-desc">Más pistas</option><option value="pistas-asc">Menos pistas</option><option value="duracion-desc">Duración ↓</option><option value="duracion-asc">Duración ↑</option></select></label>
                    <button class="SsBtn SsBtnPrimary" @click="emit('create-playlist')"><IconFolderPlus :size="14" stroke="2" /> Nueva lista de reproducción</button>
                </div>
            </div>

            <div v-if="loadingPlaylists" class="MusicBibCards" style="display:flex; justify-content:center; align-items:center; min-height:220px;">
                <span style="display:inline-flex; align-items:center; gap:0.5rem; opacity:0.6; font-size:0.75rem;"><IconRefresh :size="14" stroke="2" style="animation: spin 1s linear infinite;" /> Cargando listas de reproducción...</span>
            </div>
            <div v-else-if="sortedPlaylists.length" class="MusicBibCards">
                <PlaylistCard
                    v-for="pl in sortedPlaylists"
                    :key="pl.id"
                    :playlist="pl as any"
                    @open="openPlaylist"
                    @play="playPlaylistById"
                    @edit="editPlaylistById"
                />
                <button class="PlCard is-add" @click="emit('create-playlist')">
                    <div class="PlCard_CoverWrap">
                        <div class="PlCard_Cover is-add"><IconFolderPlus :size="22" stroke="1.8" /></div>
                    </div>
                    <div class="PlCard_Info"><b class="PlCard_Title">Nueva lista de reproducción</b><span class="PlCard_Sub"> Con lo que tienes</span></div>
                </button>
            </div>
            <div v-else class="MusicBibEmpty">
                <span class="MusicBibEmptyIcon"><IconPlaylist :size="22" stroke="1.5" /></span>
                <b>{{ playlistSearch ? 'Sin resultados' : 'Aún no tienes listas de reproducción' }}</b>
                <p v-if="!playlistSearch">Crea una y selecciona tus pistas visibles — sin escribir rutas.</p>
                <p v-else>No hay listas de reproducción para “{{ playlistSearch }}”.</p>
                <button v-if="!playlistSearch" class="SsBtn SsBtnPrimary" @click="emit('create-playlist')"><IconFolderPlus :size="14" stroke="2" /> Crear primera</button>
                <button v-else class="SsBtn" @click="playlistSearch=''">Limpiar búsqueda</button>
            </div>
        </template>

        <!-- MODO DETALLE: subpanel completo -->
        <template v-else-if="selectedPlaylist">
            <div class="MusicBibDetail">
                <button class="MusicBibBack" @click="closeDetail"><IconArrowLeft :size="16" stroke="2" /> Volver a listas de reproducción</button>

                <div class="MusicBibDetail_Hero" :class="{ 'is-collapsed': heroCollapsed }" :style="detailPalette ? { ['--cover-vibrant' as any]: detailColor, ['--cover-muted' as any]: detailPalette.muted, ['--cover-dominant' as any]: detailPalette.dominant } : {}">
                    <img v-if="detailCoverUrl && !heroCollapsed" class="MusicBibDetail_Bg" :src="detailCoverUrl" alt="" />
                    <div v-if="!heroCollapsed" class="MusicBibDetail_Glow" :style="detailPalette ? { background: `radial-gradient(700px 260px at 18% 0%, ${detailColor}18, transparent 65%)` } : {}" />
                    <button class="MusicBibHero_Toggle" :title="heroCollapsed ? 'Expandir hero' : 'Contraer hero'" @click="heroCollapsed = !heroCollapsed">
                        <IconChevronUp v-if="!heroCollapsed" :size="14" stroke="2" />
                        <IconChevronDown v-else :size="14" stroke="2" />
                    </button>
                    <template v-if="!heroCollapsed">
                    <div class="MusicBibDetail_Cover" :class="{ hasGrid: !detailCustomBlob && !(selectedPlaylist as any).customCover && detailFirstFourDisplay.some((t:any)=>t.coverUrl), 'is-disc': coverStyle==='disc' }">
                        <img v-if="detailCustomBlob" :src="detailCustomBlob" alt="" />
                        <img v-else-if="(selectedPlaylist as any).customCover" :src="(selectedPlaylist as any).customCover" alt="" @error="(e:any)=> e.target.style.display='none'" />
                        <div v-else-if="detailFirstFourDisplay.some((t:any)=>t.coverUrl)" class="MusicBibCoverGrid is-large">
                            <span v-for="(t,i) in detailFirstFourDisplay" :key="i" class="MusicBibCoverCell"><img v-if="(t as any).coverUrl" :src="(t as any).coverUrl" alt="" loading="lazy" /><IconMusic v-else :size="16" stroke="1.2" /></span>
                        </div>
                        <IconLibrary v-else :size="28" stroke="1.4" />
                    </div>
                    <div class="MusicBibDetail_Info">
                        <span class="MusicBibDetail_Kicker">Lista de reproducción · {{ validCount }} pista{{ validCount===1?'':'s' }}<template v-if="missingCount"> · {{ missingCount }} no encontrada{{ missingCount===1?'':'s' }}</template></span>
                        <h2 :title="selectedPlaylist.title">{{ selectedPlaylist.title }}</h2>
                        <div class="MusicBibDetail_Meta">
                            <span><IconMusic :size="12" stroke="2" /> {{ bibliotecaTracks.length }} pistas totales</span>
                            <span><IconClock :size="12" stroke="2" /> {{ formatLongDuration(totalDuration) }}</span>
                            <span><IconCalendar :size="12" stroke="2" /> {{ formatDate((selectedPlaylist as any).updatedAt || (selectedPlaylist as any).createdAt) }}</span>
                            <span v-if="(selectedPlaylist as any).favorite" class="is-fav"><IconStar :size="12" stroke="2" /> Favorita</span>
                            <span v-if="(selectedPlaylist as any).pinned" class="is-pin"><IconPin :size="12" stroke="2" /> Anclada</span>
                        </div>
                        <p v-if="(selectedPlaylist as any).customColor" class="MusicBibDetail_ColorHint"><span class="dot" :style="{ background: (selectedPlaylist as any).customColor }"></span> Color personalizado {{ (selectedPlaylist as any).customColor }}</p>
                        <div class="MusicBibDetail_Actions">
                            <button class="SsBtn SsBtnPrimary" :disabled="!validCount" @click="onPlayAll(false)"><IconPlayerPlay :size="14" stroke="2" /> Reproducir</button>
                            <button class="SsBtn" :disabled="!validCount" @click="onPlayAll(true)"><IconArrowsShuffle :size="14" stroke="2" /> Aleatorio</button>
                            <button class="SsBtn" @click="onEdit(selectedPlaylist)"><IconPencil :size="14" stroke="2" /> Editar</button>
                            <button class="SsBtn" :disabled="isRefreshing" @click="onRefreshCovers"><IconRefresh :size="14" stroke="2" :class="{ spinning: isRefreshing }" /> {{ isRefreshing ? 'Refrescando…' : 'Refrescar carátulas' }}</button>
                            <button class="SsBtn SsBtnDanger" @click="onDelete(selectedPlaylist.id)"><IconTrash :size="14" stroke="2" /> Eliminar</button>
                        </div>
                    </div>
                    </template>
                    <template v-else>
                    <div class="MusicBibDetail_CollapsedInfo">
                        <span class="MusicBibDetail_Cover mini" :class="{ 'is-disc': coverStyle==='disc' }">
                            <img v-if="detailCustomBlob" :src="detailCustomBlob" alt="" />
                            <img v-else-if="(selectedPlaylist as any).customCover" :src="(selectedPlaylist as any).customCover" alt="" />
                            <IconLibrary v-else :size="14" stroke="1.4" />
                        </span>
                        <b :title="selectedPlaylist.title">{{ selectedPlaylist.title }}</b>
                        <span>{{ validCount }} pistas · {{ formatLongDuration(totalDuration) }}</span>
                        <button class="SsBtn SsBtnPrimary small" :disabled="!validCount" @click="onPlayAll(false)"><IconPlayerPlay :size="12" stroke="2" /> Reproducir</button>
                    </div>
                    </template>
                </div>

                <div class="MusicToolbar is-detail">
                    <label class="MusicSearch"><IconSearch :size="14" stroke="2" /><input v-model="query" type="text" placeholder="Buscar en esta lista de reproducción…" spellcheck="false" /></label>
                    <label class="MusicToolbar_Sort"><select v-model="sortDetail" class="SsSel"><option value="orden">Orden: Original</option><option value="titulo-asc">Título A → Z</option><option value="titulo-desc">Título Z → A</option><option value="artista-asc">Artista A → Z</option><option value="artista-desc">Artista Z → A</option><option value="duracion-asc">Duración ↑</option><option value="duracion-desc">Duración ↓</option></select></label>
                    <select v-model.number="bibPageSize" class="SsSel" title="Por página" style="margin-left:0.4rem;">
                        <option :value="20">20</option>
                        <option :value="50">50</option>
                        <option :value="100">100</option>
                    </select>
                    <span class="MusicCount">{{ bibliotecaTracks.length }} pistas · {{ formatLongDuration(totalDuration) }}</span>
                </div>
                <div v-if="isLoadingPlaylist" class="MusicRows is-detail is-loading">
                    <div v-for="i in 8" :key="i" class="MusicRow is-skeleton">
                        <span class="MusicRow_Cover is-skeleton"><span class="MusicRow_Shimmer"></span></span>
                        <div class="MusicRow_Info"><span class="MusicRow_Title skeleton"></span><span class="MusicRow_Sub skeleton"></span></div>
                        <span class="MusicRow_Dur skeleton"></span>
                    </div>
                    <p class="MusicLoadingHint"><IconRefresh :size="12" stroke="2" class="spinning" /> Cargando listas de reproduccion…</p>
                </div>
                <div v-else class="MusicRows is-detail">
                    <div v-for="t in bibliotecaPaginated" :key="t.path" class="MusicRow" :class="{ missing: t.artist==='No encontrada', active: isCurrent(t) }" @click="onPlay(t)">
                        <span class="MusicRow_Cover" :class="{ playing: isCurrent(t), 'is-loading': (t as any).hasCover && !(t as any).coverUrl }">
                            <img v-if="(t as any).coverUrl" :src="(t as any).coverUrl" alt="" loading="lazy" />
                            <span v-else-if="(t as any).hasCover" class="MusicRow_Shimmer"></span>
                            <IconDisc v-else :size="14" stroke="1.6" />
                        </span>
                        <div class="MusicRow_Info"><span class="MusicRow_Title" :title="t.title">{{ t.title }}</span><span class="MusicRow_Sub">{{ t.artist }} · {{ t.fileName }}</span></div>
                        <span class="MusicRow_Dur">{{ formatDuration(t.duration) }}</span>
                        <button class="MusicRow_Play" :class="{ on: isCurrent(t) }" title="Reproducir" @click.stop="onPlay(t)"><IconPlayerPlay :size="14" stroke="2" /></button>
                    </div>
                    <div v-if="!bibliotecaPaginated.length" class="MusicEmpty" style="padding:1.4rem;"><p class="MusicEmpty_Desc">{{ query ? `Sin resultados para “${query}”.` : 'Esta lista de reproducción está vacía.' }}</p></div>
                </div>
                <footer class="MusicFooter"><span class="MusicFooter_Info">Mostrando <b>{{ bibRangeStart }}–{{ bibRangeEnd }}</b> de <b>{{ bibliotecaTracks.length }}</b></span><nav v-if="bibTotalPages>1" class="MusicPages"><button class="MusicPgBtn" :disabled="bibPage<=1" @click="goToBibPage(bibPage-1)"><IconChevronLeft :size="14" stroke="2" /></button><template v-for="(it,i) in bibPageItems" :key="`${it}-${i}`"><span v-if="it==='ellipsis'" class="MusicPgGap">…</span><button v-else class="MusicPgBtn" :class="{ on: it===bibPage }" @click="goToBibPage(it as number)">{{ it }}</button></template><button class="MusicPgBtn" :disabled="bibPage>=bibTotalPages" @click="goToBibPage(bibPage+1)"><IconChevronRight :size="14" stroke="2" /></button></nav></footer>
            </div>
        </template>

        <PlaylistForm
            :visible="showEdit"
            :initial-title="editTarget?.title"
            :initial-tracks="editTarget?.trackPaths"
            :initial-favorite="editTarget?.favorite"
            :initial-pinned="editTarget?.pinned"
            :initial-color="editTarget?.customColor"
            :initial-cover="editTarget?.customCover"
            mode="edit"
            @update:visible="showEdit = $event"
            @submit="handleEditSubmit"
        />
    </div>
</template>

<style scoped lang="scss">
@use '../Styles/LibraryView.scss';
</style>
