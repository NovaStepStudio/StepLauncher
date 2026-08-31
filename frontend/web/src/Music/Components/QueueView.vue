<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { IconPlaylist, IconChevronLeft, IconChevronRight, IconMusic, IconPlayerPlay, IconTrash, IconArrowsShuffle, IconGripVertical, IconX, IconPlayerSkipForward, IconDotsVertical, IconMenu2 } from '@tabler/icons-vue';
import { queueTracks, playTrack, setQueue, removeFromQueue, moveInQueue, addNext } from '../PlayerStore';
import { currentTrack } from '../PlayerStore';

const colaPage = ref(1);
const colaPageSize = 20;
const draggedOrdered = ref<number | null>(null);
const compactMenu = ref(true);
const openMenuPath = ref<string | null>(null);

const colaTotalPages = computed(() => Math.max(1, Math.ceil(queueTracks.value.length / colaPageSize)));
const colaRangeStart = computed(() => queueTracks.value.length === 0 ? 0 : (colaPage.value - 1) * colaPageSize + 1);
const colaRangeEnd = computed(() => Math.min(colaPage.value * colaPageSize, queueTracks.value.length));

// Si está sonando una pista de la página 4, al abrir la cola debe estar en la página 4
watch([() => currentTrack.value?.path, queueTracks], ([path]) => {
    if (!path) return;
    const idx = queueTracks.value.findIndex((t) => t.path === path);
    if (idx >= 0) {
        const p = Math.floor(idx / colaPageSize) + 1;
        if (p !== colaPage.value) colaPage.value = p;
    }
}, { immediate: true });
const queuePaginatedRaw = computed(() => {
    const s = (colaPage.value - 1) * colaPageSize;
    return queueTracks.value.slice(s, s + colaPageSize);
});
const queueCoverMap = ref<Map<string, string>>(new Map());
async function fetchQueueCovers(paths: string[]): Promise<void> {
    const need = paths.filter((p) => !queueCoverMap.value.has(p));
    if (!need.length) return;
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.GetMusicTracksBatch === 'function') {
            const batch: any = await mod.GetMusicTracksBatch(need, 'thumb');
            if (Array.isArray(batch)) batch.forEach((t: any) => { const c = t.coverThumb || t.coverRaw; if (c) queueCoverMap.value.set(t.path, c); });
        }
    } catch (_e) {}
}
watch(queuePaginatedRaw, (list) => { if (list.length) void fetchQueueCovers(list.map((t: any) => t.path)); }, { immediate: true });
const queuePaginated = computed(() => queuePaginatedRaw.value.map((t: any) => ({
    ...t,
    coverUrl: (t as any).coverUrl || queueCoverMap.value.get(t.path) || '',
})));
const colaPageItems = computed<(number|'ellipsis')[]>(() => {
    const total = colaTotalPages.value, cur = colaPage.value;
    if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);
    const items: (number|'ellipsis')[] = [1];
    const lo = Math.max(2, cur - 1), hi = Math.min(total - 1, cur + 1);
    if (lo > 2) items.push('ellipsis');
    for (let p = lo; p <= hi; p++) items.push(p);
    if (hi < total - 1) items.push('ellipsis');
    items.push(total);
    return items;
});
function goToColaPage(p: number): void { if (p >= 1 && p <= colaTotalPages.value) colaPage.value = p; }
function formatDuration(sec: number): string {
    if (!Number.isFinite(sec) || sec <= 0) return '—';
    const m = Math.floor(sec / 60), s = Math.floor(sec % 60);
    return `${m}:${String(s).padStart(2,'0')}`;
}
async function onPlay(t: any): Promise<void> { await playTrack(t, queueTracks.value as any); }
function clearQueue(): void { setQueue([]); }
function onRemove(t: any): void { removeFromQueue(t.path); }
function onAddNext(t: any): void { addNext(t as any); }
function onDragStart(_e: DragEvent, orderedIdx: number): void { draggedOrdered.value = orderedIdx; }
function onDragOver(e: DragEvent): void { e.preventDefault(); }
function onDrop(_e: DragEvent, dropOrderedIdx: number): void {
    if (draggedOrdered.value === null) return;
    const from = draggedOrdered.value;
    const to = dropOrderedIdx;
    draggedOrdered.value = null;
    if (from !== to) moveInQueue(from, to);
}
function toggleMenu(path: string): void { openMenuPath.value = openMenuPath.value === path ? null : path; }
function closeMenu(): void { openMenuPath.value = null; }
</script>

<template>
    <div class="MusicSection" @click="closeMenu">
        <div class="MusicToolbar">
            <span class="MusicToolbar_Title"><IconPlaylist :size="16" stroke="2" /> Cola de reproducción</span>
            <div class="MusicToolbar_Right">
                <button class="SsBtn SsBtnSmall" :title="compactMenu ? 'Mostrar botones' : 'Modo menú compacto'" @click="compactMenu = !compactMenu"><IconMenu2 :size="12" stroke="2" /> {{ compactMenu ? 'Expandir botones' : 'Modo menú' }}</button>
                <span class="MusicCount">{{ queueTracks.length }} pistas en cola</span>
                <button class="SsBtn SsBtnSmall SsBtnDanger" :disabled="!queueTracks.length" @click="clearQueue"><IconTrash :size="12" stroke="2" /> Limpiar</button>
            </div>
        </div>

        <div v-if="!queueTracks.length" class="MusicEmpty">
            <span class="MusicEmpty_Icon"><IconArrowsShuffle :size="22" stroke="1.5" /></span>
            <b>Cola vacía</b>
            <p class="MusicEmpty_Desc">Reproduce una pista desde <b>Música</b> o una <b>lista de reproducción</b> y se llenará automáticamente. Respeta el modo aleatorio.</p>
        </div>

        <div v-else class="MusicRows">
            <div v-for="(t, idx) in queuePaginated" :key="t.path" class="MusicRow" :class="{ active: currentTrack?.path === t.path, dragging: draggedOrdered === (colaPage-1)*colaPageSize+idx }" draggable="true" @dragstart="onDragStart($event, (colaPage-1)*colaPageSize+idx)" @dragover="onDragOver($event)" @drop="onDrop($event, (colaPage-1)*colaPageSize+idx)" @click="onPlay(t)">
                <span class="MusicRow_Drag" title="Arrastrar para reordenar"><IconGripVertical :size="14" stroke="2" /></span>
                <span class="MusicRow_Num">{{ (colaPage - 1) * colaPageSize + idx + 1 }}</span>
                <span class="MusicRow_Cover" :class="{ 'is-loading': (t as any).hasCover && !(t as any).coverUrl }"><img v-if="(t as any).coverUrl" :src="(t as any).coverUrl" alt="" loading="lazy" /><span v-if="(t as any).hasCover && !(t as any).coverUrl" class="MusicRow_Shimmer"></span><IconMusic v-if="!(t as any).coverUrl" :size="14" stroke="1.5" style="position:relative; z-index:1;" /></span>
                <div class="MusicRow_Info"><span class="MusicRow_Title">{{ (t as any).title }}</span><span class="MusicRow_Sub">{{ (t as any).artist }} · {{ (t as any).fileName }}</span></div>
                <span class="MusicRow_Dur">{{ formatDuration((t as any).duration) }}</span>
                <template v-if="!compactMenu">
                    <button class="MusicRow_IconBtn" title="Añadir a continuación" @click.stop="onAddNext(t)"><IconPlayerSkipForward :size="14" stroke="2" /></button>
                    <button class="MusicRow_IconBtn danger" title="Quitar de la cola" @click.stop="onRemove(t)"><IconX :size="14" stroke="2" /></button>
                </template>
                <div v-else class="MusicRow_MenuWrap">
                    <button class="MusicRow_IconBtn" title="Más opciones" @click.stop="toggleMenu((t as any).path)"><IconDotsVertical :size="14" stroke="2" /></button>
                    <div v-if="openMenuPath === (t as any).path" class="MusicRow_Menu">
                        <button @click.stop="onAddNext(t); closeMenu()"><IconPlayerSkipForward :size="12" stroke="2" /> Añadir a continuación</button>
                        <button class="danger" @click.stop="onRemove(t); closeMenu()"><IconX :size="12" stroke="2" /> Quitar de la cola</button>
                    </div>
                </div>
                <button class="MusicRow_Play" title="Reproducir" @click.stop="onPlay(t)"><IconPlayerPlay :size="14" stroke="2" /></button>
            </div>
        </div>

        <footer v-if="queueTracks.length" class="MusicFooter">
            <span class="MusicFooter_Info">Mostrando <b>{{ colaRangeStart }}–{{ colaRangeEnd }}</b> de <b>{{ queueTracks.length }}</b></span>
            <nav v-if="colaTotalPages>1" class="MusicPages">
                <button class="MusicPgBtn" :disabled="colaPage<=1" @click="colaPage--"><IconChevronLeft :size="14" stroke="2" /></button>
                <template v-for="(it,i) in colaPageItems" :key="`${it}-${i}`">
                    <span v-if="it==='ellipsis'" class="MusicPgGap">…</span>
                    <button v-else class="MusicPgBtn" :class="{ on: it===colaPage }" @click="goToColaPage(it as number)">{{ it }}</button>
                </template>
                <button class="MusicPgBtn" :disabled="colaPage>=colaTotalPages" @click="colaPage++"><IconChevronRight :size="14" stroke="2" /></button>
            </nav>
        </footer>
    </div>
</template>

<style scoped lang="scss">
@use '../Styles/QueueView.scss';
</style>
