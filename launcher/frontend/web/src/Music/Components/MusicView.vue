<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { IconSearch, IconChevronLeft, IconChevronRight, IconMusic, IconPlayerPlay, IconPlayerPause, IconRefresh, IconFolderOpen, IconSettings } from '@tabler/icons-vue';
import { localTracks, pagedTracks, totalTracks, isScanning, scanError, loadLocalLibrary, fetchMusicPage, scanProgress } from '../LocalStore';
import { currentTrack, playing, playTrack } from '../PlayerStore';
import { settingsOpen } from '@/Common/Overlays/Store';

const query = ref('');
const pageSize = ref(20);
const page = ref(1);
const isFetching = ref(false);
const sortOrder = ref('path');
const sortedTotal = ref(0);

let queryTimer: any = null;

function sortTracks(list: any[], order: string): any[] {
    const arr = [...list];
    switch (order) {
        case 'titulo-asc': return arr.sort((a,b) => String(a.title||'').localeCompare(String(b.title||''), 'es', { sensitivity:'base' }));
        case 'titulo-desc': return arr.sort((a,b) => String(b.title||'').localeCompare(String(a.title||''), 'es', { sensitivity:'base' }));
        case 'artista-asc': return arr.sort((a,b) => String(a.artist||'').localeCompare(String(b.artist||''), 'es', { sensitivity:'base' }));
        case 'artista-desc': return arr.sort((a,b) => String(b.artist||'').localeCompare(String(a.artist||''), 'es', { sensitivity:'base' }));
        case 'duracion-asc': return arr.sort((a,b) => (Number(a.duration)||0) - (Number(b.duration)||0));
        case 'duracion-desc': return arr.sort((a,b) => (Number(b.duration)||0) - (Number(a.duration)||0));
        default: return arr;
    }
}

// Carga paginada desde Go con carátulas thumb baja resolución (128x128, ~10KB) — no raw 500KB
async function loadPage(): Promise<void> {
    isFetching.value = true;
    try {
        const q = query.value.trim();
        // Si hay orden no default, ordenar globalmente: pedir todo sin covers y paginar cliente
        if (sortOrder.value !== 'path') {
            const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
            if (typeof mod.GetMusicTracksPaged === 'function') {
                const res: any = await mod.GetMusicTracksPaged(0, 5000, 'none', q);
                if (res && Array.isArray(res.tracks)) {
                    let all: any[] = res.tracks.map((t: any) => ({
                        id: t.path, path: t.path, fileName: t.fileName || t.path.split(/[\\/]/).pop()||'', title: t.title||t.fileName||'Desconocido', artist: t.artist||'Desconocido', duration: Number(t.duration)||0, coverUrl: '', hasCover: !!t.hasCover, fileName2: t.fileName
                    }));
                    all = sortTracks(all, sortOrder.value);
                    sortedTotal.value = all.length;
                    const s = (page.value - 1) * pageSize.value;
                    const slice = all.slice(s, s + pageSize.value);
                    // Pedir thumbs solo para el slice visible (baja resolución)
                    if (slice.length) {
                        try {
                            const batch: any = await mod.GetMusicTracksBatch(slice.map((x:any)=>x.path), 'thumb');
                            if (Array.isArray(batch)) {
                                const m = new Map<string, any>(); batch.forEach((t:any)=> m.set(t.path, t));
                                slice.forEach((t:any)=> { const b=m.get(t.path); if(b) t.coverUrl = b.coverThumb||b.coverRaw||''; });
                            }
                        } catch {}
                    }
                    pagedTracks.value = slice as any;
                    return;
                }
            }
        }
        sortedTotal.value = 0;
        const offset = (page.value - 1) * pageSize.value;
        await fetchMusicPage(offset, pageSize.value, q, 'thumb');
        // Ordenar solo la página actual si se pidió orden (fallback)
        if (sortOrder.value !== 'path' && pagedTracks.value.length) {
            pagedTracks.value = sortTracks([...pagedTracks.value], sortOrder.value) as any;
        }
    } finally {
        isFetching.value = false;
    }
}

watch(query, () => {
    page.value = 1;
    if (queryTimer) clearTimeout(queryTimer);
    queryTimer = setTimeout(() => { void loadPage(); }, 280);
});
watch([page, pageSize, sortOrder], () => { void loadPage(); });

// totalHits viene del servidor (índice), no del slice local de 20; si hay orden global usa sortedTotal
const totalHits = computed(() => sortOrder.value !== 'path' && sortedTotal.value ? sortedTotal.value : totalTracks.value);
const totalPages = computed(() => Math.max(1, Math.ceil(totalHits.value / pageSize.value)));
const rangeStart = computed(() => totalHits.value === 0 ? 0 : (page.value - 1) * pageSize.value + 1);
const rangeEnd = computed(() => Math.min(page.value * pageSize.value, totalHits.value));
const paginated = computed(() => pagedTracks.value);

const pageItems = computed<(number|'ellipsis')[]>(() => {
    const total = totalPages.value, cur = page.value;
    if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);
    const items: (number|'ellipsis')[] = [1];
    const lo = Math.max(2, cur - 1), hi = Math.min(total - 1, cur + 1);
    if (lo > 2) items.push('ellipsis');
    for (let p = lo; p <= hi; p++) items.push(p);
    if (hi < total - 1) items.push('ellipsis');
    items.push(total);
    return items;
});

function goToPage(p: number): void { if (p >= 1 && p <= totalPages.value) page.value = p; }

function formatDuration(sec: number): string {
    if (!Number.isFinite(sec) || sec <= 0) return '—';
    const m = Math.floor(sec / 60), s = Math.floor(sec % 60);
    return `${m}:${String(s).padStart(2,'0')}`;
}

async function onPlay(t: typeof pagedTracks.value[number]): Promise<void> {
    // Cola completa: toda la carpeta filtrada/ordenada, no solo los 20 visibles
    const q = query.value.trim();
    let fullQueue: any[] = [];
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        const res: any = await mod.GetMusicTracksPaged(0, 5000, 'none', q);
        if (res && Array.isArray(res.tracks) && res.tracks.length) {
            fullQueue = res.tracks.map((x: any) => ({ id: x.path, path: x.path, fileName: x.fileName || x.path.split(/[\\/]/).pop() || '', title: x.title || x.fileName || 'Desconocido', artist: x.artist || 'Desconocido', duration: Number(x.duration) || 0, coverUrl: '' }));
            if (sortOrder.value !== 'path') fullQueue = sortTracks(fullQueue, sortOrder.value);
        }
    } catch {}
    if (!fullQueue.length) fullQueue = pagedTracks.value as any;
    if (!fullQueue.some((x: any) => x.path === t.path)) fullQueue = [t, ...fullQueue];
    await playTrack(t, fullQueue as any);
}

function isCurrent(t: typeof pagedTracks.value[number]): boolean {
    return currentTrack.value?.path === t.path;
}

async function reload(): Promise<void> { page.value = 1; await loadLocalLibrary(true); await loadPage(); }

onMounted(() => { void loadPage(); });
</script>

<template>
    <div class="MusicSection">
        <div class="MusicToolbar">
            <label class="MusicSearch">
                <IconSearch :size="16" stroke="2" />
                <input v-model="query" type="text" placeholder="Buscar por título, artista o archivo…" spellcheck="false" />
            </label>
            <div class="MusicToolbar_Right">
                <label class="MusicToolbar_Sort" title="Ordenar por"><select v-model="sortOrder" class="SsSel"><option value="path">Orden: Por defecto</option><option value="titulo-asc">Título A → Z</option><option value="titulo-desc">Título Z → A</option><option value="artista-asc">Artista A → Z</option><option value="artista-desc">Artista Z → A</option><option value="duracion-asc">Duración ↑</option><option value="duracion-desc">Duración ↓</option></select></label>
                <button class="SsBtn SsBtnSmall" title="Recargar biblioteca" :disabled="isScanning" @click="reload">
                    <IconRefresh :size="14" stroke="2" :class="{ spinning: isScanning }" /> {{ isScanning ? 'Cargando…' : 'Recargar' }}
                </button>
                <select v-model.number="pageSize" class="SsSel" title="Por página">
                    <option :value="20">20</option>
                    <option :value="50">50</option>
                    <option :value="100">100</option>
                </select>
                <span class="MusicCount">{{ isFetching ? 'Cargando…' : `${totalHits} pistas` }} · {{ totalHits }} totales</span>
            </div>
        </div>

        <div v-if="scanError" class="MusicAlert error">
            <IconFolderOpen :size="16" stroke="2" />
            <span>{{ scanError }}</span>
            <button class="SsBtn SsBtnSmall" @click="reload">Reintentar</button>
        </div>

        <div v-if="isScanning" class="MusicLoading">
            <span class="MusicLoading_Dot"></span>
            <span v-if="scanProgress.scanning && scanProgress.total > 0">Importando carpeta {{ scanProgress.current }} / {{ scanProgress.total }} · {{ scanProgress.folder ? scanProgress.folder.split(/[\\/]/).pop() : '' }}</span>
            <span v-else-if="scanProgress.scanning">Importando carpeta {{ scanProgress.current }} / {{ scanProgress.total || localTracks.length }}…</span>
            <span v-else>Escaneando… {{ localTracks.length }} pistas encontradas</span>
        </div>

        <div v-if="isScanning && !paginated.length" class="MusicRows">
            <div v-for="i in 8" :key="i" class="MusicRow is-skeleton">
                <span class="MusicRow_Cover is-skeleton"><span class="MusicRow_Shimmer"></span></span>
                <div class="MusicRow_Info"><span class="MusicRow_Title skeleton"></span><span class="MusicRow_Sub skeleton"></span></div>
                <span class="MusicRow_Dur skeleton"></span>
            </div>
        </div>
        <div v-else class="MusicRows">
            <div v-for="t in paginated" :key="t.path" class="MusicRow" :class="{ active: isCurrent(t) }" @click="onPlay(t)">
                <span class="MusicRow_Cover" :class="{ playing: isCurrent(t) && playing, 'is-loading': (t as any).hasCover && !(t as any).coverUrl }">
                    <img v-if="(t as any).coverUrl" :src="(t as any).coverUrl" alt="" loading="lazy" />
                    <span v-else-if="(t as any).hasCover" class="MusicRow_Shimmer" style="position:absolute; inset:0;"></span>
                    <IconMusic v-if="!(t as any).coverUrl" :size="16" stroke="1.5" style="position:relative; z-index:1;" />
                </span>
                <div class="MusicRow_Info">
                    <span class="MusicRow_Title" :title="t.title">{{ t.title }}</span>
                    <span class="MusicRow_Sub" :title="`${t.artist} · ${t.fileName}`">{{ t.artist }} · {{ t.fileName }}</span>
                </div>
                <span class="MusicRow_Dur">{{ formatDuration(t.duration) }}</span>
                <button class="MusicRow_Play" :class="{ on: isCurrent(t) && playing }" :title="isCurrent(t) && playing ? 'Pausar' : 'Reproducir'" @click.stop="onPlay(t)">
                    <IconPlayerPause v-if="isCurrent(t) && playing" :size="14" stroke="2" />
                    <IconPlayerPlay v-else :size="14" stroke="2" />
                </button>
            </div>

            <div v-if="!paginated.length && !isScanning" class="MusicEmpty">
                <span class="MusicEmpty_Icon"><IconMusic :size="22" stroke="1.4" /></span>
                <b v-if="!localTracks.length">Aún no hay música</b>
                <b v-else>Sin resultados</b>
                <p class="MusicEmpty_Desc">
                    <template v-if="!localTracks.length">Configura tu carpeta en <b>Ajustes → Música</b>. La biblioteca se carga automáticamente.</template>
                    <template v-else>No hay pistas para “{{ query }}”. Prueba otro término.</template>
                </p>
                <div v-if="!localTracks.length" style="display:flex; gap:0.5rem; justify-content:center; flex-wrap:wrap;">
                    <button class="SsBtn SsBtnPrimary" @click="settingsOpen = true"><IconSettings :size="14" stroke="2" /> Abrir ajustes</button>
                    <button class="SsBtn" @click="reload"><IconRefresh :size="14" stroke="2" /> Recargar</button>
                </div>
            </div>
        </div>

        <footer class="MusicFooter">
            <span class="MusicFooter_Info">Mostrando <b>{{ rangeStart }}–{{ rangeEnd }}</b> de <b>{{ totalHits }}</b> · Página <b>{{ page }}</b> de <b>{{ totalPages }}</b></span>
            <nav v-if="totalPages>1" class="MusicPages">
                <button class="MusicPgBtn" :disabled="page<=1" @click="goToPage(page-1)"><IconChevronLeft :size="14" stroke="2" /></button>
                <template v-for="(it,i) in pageItems" :key="`${it}-${i}`">
                    <span v-if="it==='ellipsis'" class="MusicPgGap">…</span>
                    <button v-else class="MusicPgBtn" :class="{ on: it===page }" @click="goToPage(it as number)">{{ it }}</button>
                </template>
                <button class="MusicPgBtn" :disabled="page>=totalPages" @click="goToPage(page+1)"><IconChevronRight :size="14" stroke="2" /></button>
            </nav>
        </footer>
    </div>
</template>

<style scoped lang="scss">
@use '../Styles/MusicView.scss';
</style>
