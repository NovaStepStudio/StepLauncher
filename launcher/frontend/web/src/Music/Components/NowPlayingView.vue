<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue';
import { IconMusic, IconPlaylist, IconSparkles, IconChevronLeft, IconChevronRight, IconPlayerPlay, IconDisc, IconGripVertical, IconX, IconPlayerSkipForward, IconDotsVertical, IconMenu2 } from '@tabler/icons-vue';
import { currentTrack, currentTime, duration, progressPct, queueTracks, seekTo, playTrack, removeFromQueue, moveInQueue, addNext } from '../PlayerStore';
import { useCoverPalette, getPaletteColor } from '@/Common/Composables/useCoverPalette';
import { GetMusicPanelConfig, UpdateMusicPanelConfig } from '@wailsjs/StepLauncher/internal/Services/Music/musicservice';

// Carátula Predominante: switch + tamaño configurable 15..30 (se guarda en rem, se muestra sin unidad)
const nowPlayingHuge = ref(false);
const hugeCoverSize = ref(20); // 15..30 — valor en config, display sin "rem"
const nowPlayingCoverStyle = ref<'square' | 'disc'>('square');
const colorMode = ref('vibrant');

async function refreshConfig(): Promise<void> {
    try {
        const cfg: any = await GetMusicPanelConfig();
        nowPlayingHuge.value = !!cfg?.nowPlayingHuge;
        const rawSize = (cfg as any)?.nowPlayingHugeSize;
        if (typeof rawSize === 'number' && Number.isFinite(rawSize)) {
            hugeCoverSize.value = Math.round(Math.max(15, Math.min(30, rawSize)));
        } else {
            hugeCoverSize.value = 20;
        }
        const npc = cfg?.nowPlayingCover as string;
        if (npc === 'disc') nowPlayingCoverStyle.value = 'disc';
        else nowPlayingCoverStyle.value = cfg?.coverStyle === 'disc' ? 'disc' : 'square';
        colorMode.value = (cfg as any)?.colorMode || 'vibrant';
    } catch (_e) {}
}
function onHugePreview(e: Event): void {
    const v = (e as CustomEvent).detail;
    if (typeof v === 'number' && Number.isFinite(v)) {
        hugeCoverSize.value = Math.round(Math.max(15, Math.min(30, v)));
    }
}
onMounted(async () => {
    await refreshConfig();
    window.addEventListener('stl:music-config-changed', refreshConfig as any);
    window.addEventListener('stl:music-folder-changed', refreshConfig as any);
    window.addEventListener('stl:huge-cover-size-preview', onHugePreview as any);
});
onUnmounted(() => {
    window.removeEventListener('stl:music-config-changed', refreshConfig as any);
    window.removeEventListener('stl:music-folder-changed', refreshConfig as any);
    window.removeEventListener('stl:huge-cover-size-preview', onHugePreview as any);
});
async function toggleHuge() {
    nowPlayingHuge.value = !nowPlayingHuge.value;
    try {
        const cfg: any = await GetMusicPanelConfig();
        await UpdateMusicPanelConfig({ ...cfg, nowPlayingHuge: nowPlayingHuge.value, nowPlayingCover: nowPlayingCoverStyle.value } as any);
    } catch (_e) {}
}

// Cover principal: alta resolución con carga perezosa en baja latencia
const fetchedRaw = ref('');
const isFetchingCover = ref(false);
watch(() => currentTrack.value?.path, async (p) => {
    fetchedRaw.value = '';
    if (!p) return;
    const cur = currentTrack.value as any;
    // Si ya tiene raw original alta (>30KB base64), no pedir de nuevo
    if (cur?.coverRaw && String(cur.coverRaw).startsWith('data:') && String(cur.coverRaw).length > 50000) return;
    isFetchingCover.value = true;
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.GetMusicTracksBatch === 'function') {
            const batch: any = await mod.GetMusicTracksBatch([p], 'raw');
            if (Array.isArray(batch) && batch[0]?.coverRaw && String(batch[0].coverRaw).length > 30000) { fetchedRaw.value = batch[0].coverRaw; return; }
            if (Array.isArray(batch) && batch[0]?.coverThumb && !fetchedRaw.value) { fetchedRaw.value = batch[0].coverThumb; return; }
        }
        if (typeof mod.GetMusicCoverBase64 === 'function') {
            const raw: string = await mod.GetMusicCoverBase64(p, 'raw').catch(() => '');
            if (raw && raw.length > 30000) { fetchedRaw.value = raw; return; }
            if (raw) fetchedRaw.value = raw;
        }
    } catch (_e) {} finally { isFetchingCover.value = false; }
}, { immediate: true });

const cover = computed(() => {
    // CARÁTULA Ahora Suena SIEMPRE raw alta, nunca thumb 128 baja
    if (fetchedRaw.value && fetchedRaw.value.length > 30000) return fetchedRaw.value;
    const cur: any = currentTrack.value as any;
    if (cur?.coverRaw && String(cur.coverRaw).startsWith('data:') && String(cur.coverRaw).length > 30000) return cur.coverRaw;
    if (fetchedRaw.value) return fetchedRaw.value;
    return '';
});
const { palette } = useCoverPalette(() => cover.value);
// Panel completo con linear-gradient arriba→abajo usando color predominante/configurado
const panelBg = computed(() => {
    if (!currentTrack.value || !palette.value) return {};
    const col = getPaletteColor(palette.value, colorMode.value);
    return {
        background: `linear-gradient(180deg, ${col}22 0%, ${col}12 18%, var(--background-modal-primary) 55%)`,
    } as Record<string, string>;
});

const sidePage = ref(1);
const sidePageSize = 10;
const sideTotalPages = computed(() => Math.max(1, Math.ceil(queueTracks.value.length / sidePageSize)));
const sidePaginatedRaw = computed(() => {
    const s = (sidePage.value - 1) * sidePageSize;
    return queueTracks.value.slice(s, s + sidePageSize);
});
const sideCoverMap = ref<Map<string, string>>(new Map());
const draggedSide = ref<number | null>(null);
const compactSideMenu = ref(true);
const openSideMenuPath = ref<string | null>(null);
async function fetchSideCovers(paths: string[]): Promise<void> {
    const need = paths.filter((p) => !sideCoverMap.value.has(p));
    if (!need.length) return;
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.GetMusicTracksBatch === 'function') {
            const batch: any = await mod.GetMusicTracksBatch(need, 'thumb');
            if (Array.isArray(batch)) batch.forEach((t: any) => { const c = t.coverThumb || t.coverRaw; if (c) sideCoverMap.value.set(t.path, c); });
        }
    } catch (_e) {}
}
watch(sidePaginatedRaw, (list) => {
    if (list.length) void fetchSideCovers(list.map((t: any) => t.path));
}, { immediate: true });
const sidePaginated = computed(() => sidePaginatedRaw.value.map((t: any) => ({
    ...t,
    coverUrl: (t as any).coverUrl || sideCoverMap.value.get(t.path) || '',
})));
watch([() => currentTrack.value?.path, queueTracks], ([path]) => {
    if (!path) return;
    const idx = queueTracks.value.findIndex((t) => t.path === path);
    if (idx >= 0) {
        const p = Math.floor(idx / sidePageSize) + 1;
        if (p !== sidePage.value) sidePage.value = p;
    }
}, { immediate: true });
function onSideRemove(t: any): void { removeFromQueue(t.path); }
function onSideNext(t: any): void { addNext(t as any); }
function onSideDragStart(_e: DragEvent, orderedIdx: number): void { draggedSide.value = orderedIdx; }
function onSideDragOver(e: DragEvent): void { e.preventDefault(); }
function onSideDrop(_e: DragEvent, dropOrderedIdx: number): void {
    if (draggedSide.value === null) return;
    const from = draggedSide.value;
    const to = dropOrderedIdx;
    draggedSide.value = null;
    if (from !== to) moveInQueue(from, to);
}
function toggleSideMenu(path: string): void { openSideMenuPath.value = openSideMenuPath.value === path ? null : path; }
function closeSideMenu(): void { openSideMenuPath.value = null; }

function formatDuration(sec: number): string {
    if (!Number.isFinite(sec) || sec <= 0) return '0:00';
    const m = Math.floor(sec / 60), s = Math.floor(sec % 60);
    return `${m}:${String(s).padStart(2,'0')}`;
}

async function onPlay(t: any): Promise<void> {
    await playTrack(t, queueTracks.value as any);
}

// Limpia título/artista tal cual debe verse — elimina BOM/FFFD y controles que se ven como "☐" al final del nombre.
function cleanDisplay(s: any): string {
    let str = String(s ?? '').trim();
    str = str.replace(/[\uFEFF\uFFFD]/g, '');
    str = str.replace(/[\x00-\x1F\x7F-\x9F]/g, '');
    str = str.replace(/[\u200B\u200C\u200D\u2060\u00AD\u034F]/g, '');
    return str.trim();
}
const safeTitle = computed(() => cleanDisplay(currentTrack.value?.title) || 'Nada en reproducción');
const safeArtist = computed(() => cleanDisplay(currentTrack.value?.artist) || 'Elige una pista');

// Tamaño carátula predominante 15..30 — el tamaño afecta al contenedor (width y height), la carátula es auto con max 100% para que el borde se adapte
const hugeCoverStyle = computed(() => {
    const s = Math.max(15, Math.min(30, Number(hugeCoverSize.value) || 20));
    const isDisc = nowPlayingCoverStyle.value === 'disc';
    if (isDisc) {
        // Disco: contenedor cuadrado fijo, imagen cover
        return {
            width: `${s}rem`,
            height: `${s}rem`,
            minWidth: '3rem',
            minHeight: '3rem',
            maxWidth: `min(${s}rem, 85vw)`,
            maxHeight: `min(${s}rem, 85vw)`,
            aspectRatio: '1',
        } as Record<string, string>;
    } else {
        // Cuadrado: contenedor auto, limitado a s rem en ambos ejes, imagen max 100% dentro (no sobresale)
        return {
            width: 'auto',
            height: 'auto',
            minWidth: '3rem',
            minHeight: '3rem',
            maxWidth: `min(${s}rem, 85vw)`,
            maxHeight: `min(${s}rem, 55vh)`,
        } as Record<string, string>;
    }
});
const hugeMetaStyle = computed(() => {
    const s = Math.max(15, Math.min(30, Number(hugeCoverSize.value) || 20));
    return { maxWidth: `${s}rem` } as Record<string, string>;
});

const d = computed(() => duration.value || currentTrack.value?.duration || 0);
</script>

<template>
    <div class="MusicSection MusicNow" :style="panelBg">
        <div class="MusicNow_Layout is-pro">
            <!-- Modo ENORME: carátula gigante original arriba, título+artista debajo, sin progreso -->
            <template v-if="nowPlayingHuge">
                <div class="MusicNow_Huge is-pro is-huge-mode" :style="palette ? { borderColor: `color-mix(in srgb, ${palette.vibrant} 22%, transparent)` } : {}">
                    <img v-if="cover" class="MusicNow_HugeBgImg" :src="cover" alt="" />
                    <div class="MusicNow_HugeBg" :style="palette ? { background: `radial-gradient(800px 500px at 30% 18%, ${palette.vibrant}18, transparent 62%), radial-gradient(600px 400px at 85% 85%, ${palette.dominant}16, transparent 60%), linear-gradient(180deg, transparent 45%, rgba(0,0,0,0.45) 100%)` } : {}" />
                    <div class="MusicNow_HugeCoverOnly" :class="{ 'is-disc': nowPlayingCoverStyle === 'disc', 'is-loading': isFetchingCover && !cover }" :style="hugeCoverStyle">
                        <img v-if="cover" :src="cover" alt="" :class="{ 'is-disc-img': nowPlayingCoverStyle === 'disc' }" />
                        <span v-else-if="isFetchingCover" class="MusicNow_Shimmer"></span>
                        <IconMusic v-else :size="32" stroke="1.2" />
                        <IconDisc v-if="!cover && nowPlayingCoverStyle==='disc' && !isFetchingCover" :size="32" stroke="1.2" style="position:absolute; opacity:0.3;" />
                    </div>
                    <div class="MusicNow_HugeMetaOnly" :style="hugeMetaStyle">
                        <h3 class="MusicNow_Title is-huge" :title="safeTitle">{{ safeTitle }}</h3>
                        <p class="MusicNow_Artist is-huge">{{ safeArtist }}</p>
                    </div>
                </div>
            </template>
            <template v-else>
            <div class="MusicNow_Huge is-pro" :style="palette ? { borderColor: `color-mix(in srgb, ${palette.vibrant} 22%, transparent)` } : {}">
                <img v-if="cover" class="MusicNow_HugeBgImg" :src="cover" alt="" />
                <div class="MusicNow_HugeBg" :style="palette ? { background: `radial-gradient(800px 500px at 30% 18%, ${palette.vibrant}24, transparent 62%), radial-gradient(600px 400px at 85% 85%, ${palette.dominant}20, transparent 60%), linear-gradient(180deg, transparent 45%, rgba(0,0,0,0.55) 100%)` } : {}" />
                <div class="MusicNow_CoverWrap is-row is-pro">
                    <div class="MusicNow_Cover is-huge is-pro" :class="{ 'is-disc': nowPlayingCoverStyle==='disc', 'is-loading': isFetchingCover && !cover }" :style="palette ? { boxShadow: `0 16px 48px ${palette.dominant}35` } : {}">
                        <img v-if="cover" :src="cover" alt="" :class="{ 'is-disc-img': nowPlayingCoverStyle==='disc' }" />
                        <span v-else-if="isFetchingCover" class="MusicNow_Shimmer"></span>
                        <IconMusic v-else :size="28" stroke="1.4" />
                        <IconDisc v-if="nowPlayingCoverStyle==='disc' && cover" :size="16" stroke="1.5" style="position:absolute; inset:0; margin:auto; width:28%; height:28%; opacity:0.9; color:white; filter:drop-shadow(0 1px 2px rgba(0,0,0,0.6));" />
                    </div>
                    <div class="MusicNow_HugeInfo is-side is-pro">
                        <span class="MusicNow_Kicker"><IconSparkles :size="12" stroke="2" /> Ahora suena</span>
                        <h3 class="MusicNow_Title is-huge" :title="safeTitle">{{ safeTitle }}</h3>
                        <p class="MusicNow_Artist is-huge">{{ safeArtist }}</p>
                        <span class="MusicNow_Album">{{ currentTrack ? `${formatDuration(d)} · ${currentTrack.fileName}` : 'Sin cola · Abre Música o Biblioteca' }}</span>
                        <div class="MusicNow_Progress is-pro">
                            <span>{{ formatDuration(currentTime) }}</span>
                            <div class="MusicNow_Bar" @click="(e: MouseEvent) => { const rect = (e.currentTarget as HTMLElement).getBoundingClientRect(); const pct = (e.clientX - rect.left)/rect.width; seekTo(pct * d); }">
                                <i :style="{ width: progressPct + '%' }" />
                                <input class="MusicNow_Range" type="range" :min="0" :max="d || 100" :value="currentTime" step="1" @input="(e: any) => seekTo(Number(e.target.value))" />
                            </div>
                            <span>{{ formatDuration(d) }}</span>
                        </div>
                    </div>
                </div>
            </div>
            </template>
            <aside class="MusicNow_Side is-pro" @click="closeSideMenu">
                <h4 style="display:flex; align-items:center; justify-content:space-between; gap:0.5rem;"><span style="display:flex; align-items:center; gap:0.4rem;"><IconPlaylist :size="14" stroke="2" /> A continuación <em style="opacity:0.5; font-weight:400;">— {{ queueTracks.length }}</em></span><button class="SsBtn SsBtnSmall" :title="compactSideMenu ? 'Mostrar todos los botones' : 'Modo menú compacto para cola estrecha'" @click.stop="compactSideMenu = !compactSideMenu"><IconMenu2 :size="12" stroke="2" /> {{ compactSideMenu ? 'Botones' : 'Menú' }}</button></h4>
                <p class="MusicNow_SideHint">{{ queueTracks.length ? 'Paginado · respeta aleatorio' : 'Añade pistas desde Música o pon una lista de reproducción' }}</p>
                <div class="MusicRows small is-side">
                    <div v-for="(t, idx) in sidePaginated" :key="t.path" class="MusicRow small" :class="{ 'is-active': currentTrack?.path === t.path, dragging: draggedSide === (sidePage-1)*sidePageSize+idx }" draggable="true" @dragstart="onSideDragStart($event, (sidePage-1)*sidePageSize+idx)" @dragover="onSideDragOver($event)" @drop="onSideDrop($event, (sidePage-1)*sidePageSize+idx)" @click="onPlay(t)">
                        <span class="MusicRow_Drag small" title="Arrastrar"><IconGripVertical :size="12" stroke="2" /></span>
                        <span class="MusicRow_Cover small" :class="{ 'is-loading': (t as any).hasCover && !(t as any).coverUrl }"><img v-if="(t as any).coverUrl" :src="(t as any).coverUrl" alt="" loading="lazy" /><span v-if="(t as any).hasCover && !(t as any).coverUrl" class="MusicRow_Shimmer"></span><IconMusic v-if="!(t as any).coverUrl" :size="14" stroke="1.5" style="position:relative; z-index:1;" /></span>
                        <span class="MusicRow_Info"><span class="MusicRow_Title">{{ cleanDisplay(t.title) }}</span><span class="MusicRow_Sub">{{ cleanDisplay(t.artist) }}</span></span>
                        <span class="MusicRow_Dur">{{ formatDuration(t.duration) }}</span>
                        <template v-if="!compactSideMenu">
                            <button class="MusicRow_IconBtn small" title="Añadir a continuación" @click.stop="onSideNext(t)"><IconPlayerSkipForward :size="12" stroke="2" /></button>
                            <button class="MusicRow_IconBtn small danger" title="Quitar de la cola" @click.stop="onSideRemove(t)"><IconX :size="12" stroke="2" /></button>
                        </template>
                        <div v-else class="MusicRow_MenuWrap">
                            <button class="MusicRow_IconBtn small" title="Más opciones" @click.stop="toggleSideMenu((t as any).path)"><IconDotsVertical :size="12" stroke="2" /></button>
                            <div v-if="openSideMenuPath === (t as any).path" class="MusicRow_Menu">
                                <button @click.stop="onSideNext(t); closeSideMenu()"><IconPlayerSkipForward :size="12" stroke="2" /> Añadir a continuación</button>
                                <button class="danger" @click.stop="onSideRemove(t); closeSideMenu()"><IconX :size="12" stroke="2" /> Quitar</button>
                            </div>
                        </div>
                        <button class="MusicRow_Play small" title="Reproducir" @click.stop="onPlay(t)"><IconPlayerPlay :size="12" stroke="2" /></button>
                    </div>
                    <p v-if="!queueTracks.length" class="MusicMiniEmpty">Cola vacía. Reproduce algo y aparecerá aquí.</p>
                </div>
                <div v-if="sideTotalPages > 1" class="MusicNow_SideFooter">
                    <span class="MusicNow_SideInfo">{{ sidePage }} / {{ sideTotalPages }}</span>
                    <nav class="MusicNow_SidePages">
                        <button class="MusicPgBtn small" :disabled="sidePage<=1" @click="sidePage--"><IconChevronLeft :size="12" stroke="2" /></button>
                        <button class="MusicPgBtn small" :disabled="sidePage>=sideTotalPages" @click="sidePage++"><IconChevronRight :size="12" stroke="2" /></button>
                    </nav>
                </div>
            </aside>
        </div>
    </div>
</template>

<style scoped lang="scss">
@use '../Styles/NowPlayingView.scss';
</style>
