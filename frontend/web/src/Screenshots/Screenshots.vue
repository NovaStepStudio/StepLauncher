<script setup lang="ts">
import { ref, computed, reactive, watch, onMounted, onUnmounted } from 'vue';
import { IconPhoto, IconX, IconZoomIn, IconZoomOut, IconChevronLeft, IconChevronRight } from '@tabler/icons-vue';
import { ListScreenshots, ReadLocalFile } from '@wailsjs/go/main/App';
import type { Handlers } from '@wailsjs/go/models';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import {
    heavyPanel, openHeavyPanel, closeHeavyPanel,
    shotsInstance, shotsReturn,
} from '@/Common/Overlays/Store';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

const SHOTS_REFRESH_EVENT = 'sl:shots-refresh';

const shots = ref<Handlers.ScreenshotInfo[]>([]);
const loading = ref(false);
const error = ref('');

const thumbCache = reactive(new Map<string, string>());
const fullCache = reactive(new Map<string, string>());

const THUMB_MAX = 480;

const previewIndex = ref<number | null>(null);
const zoom = ref(1);
const pan = reactive({ x: 0, y: 0 });
const stageEl = ref<HTMLElement | null>(null);
const baseSize = ref<{ w: number; h: number } | null>(null);
const dragState = ref<{ x: number; y: number; sx: number; sy: number } | null>(null);

const MIN_ZOOM = 0.5;
const MAX_ZOOM = 4;

const preview = computed(() =>
    previewIndex.value === null ? null : shots.value[previewIndex.value]
);

const previewUrl = computed(() =>
    preview.value ? loadUrl(preview.value.path, fullCache, 0) : ''
);

const imgStyle = computed(() => ({
    transform: `translate(calc(-50% + ${pan.x}px), calc(-50% + ${pan.y}px)) scale(${zoom.value})`,
}));

function mime(path: string): string {
    const lower = path.toLowerCase();
    if (lower.endsWith('.jpg') || lower.endsWith('.jpeg')) return 'image/jpeg';
    if (lower.endsWith('.gif')) return 'image/gif';
    return 'image/png';
}

function bytesToBlob(data: unknown, rel: string): Blob | null {
    let bytes: Uint8Array;
    if (typeof data === 'string') {
        const bin = atob(data);
        bytes = new Uint8Array(bin.length);
        for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
    } else if (Array.isArray(data)) {
        bytes = new Uint8Array(data);
    } else if (data instanceof Uint8Array) {
        bytes = data;
    } else {
        return null;
    }
    return new Blob([bytes.slice().buffer as ArrayBuffer], { type: mime(rel) });
}

function decodeScaled(blob: Blob, maxDim: number): Promise<string> {
    return new Promise((resolve, reject) => {
        const url = URL.createObjectURL(blob);
        const img = new Image();
        img.onload = () => {
            const scale =
                maxDim > 0
                    ? Math.min(1, maxDim / Math.max(img.naturalWidth, img.naturalHeight))
                    : 1;
            if (scale >= 1) {
                resolve(url);
                return;
            }
            const canvas = document.createElement('canvas');
            canvas.width = Math.max(1, Math.round(img.naturalWidth * scale));
            canvas.height = Math.max(1, Math.round(img.naturalHeight * scale));
            const ctx = canvas.getContext('2d');
            if (!ctx) {
                resolve(url);
                return;
            }
            ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
            URL.revokeObjectURL(url);
            canvas.toBlob(
                (b) => resolve(b ? URL.createObjectURL(b) : url),
                'image/jpeg',
                0.82
            );
        };
        img.onerror = () => {
            URL.revokeObjectURL(url);
            reject(new Error('imagen no decodificable'));
        };
        img.src = url;
    });
}

function loadUrl(rel: string, cache: Map<string, string>, maxDim: number): string {
    const cached = cache.get(rel);
    if (cached) return cached;
    ReadLocalFile(rel)
        .then(async (data) => {
            if (cache.has(rel)) return;
            const blob = bytesToBlob(data, rel);
            if (!blob) return;
            const url = await decodeScaled(blob, maxDim);
            if (!cache.has(rel)) cache.set(rel, url);
        })
        .catch(() => { });
    return '';
}

function thumbOf(rel: string): string {
    return loadUrl(rel, thumbCache, THUMB_MAX);
}

function refresh() {
    if (heavyPanel.value !== 'shots') return;
    loading.value = true;
    error.value = '';
    const load = shotsInstance.value
        ? (window as any)?.go?.main?.App?.ListInstanceScreenshots?.(shotsInstance.value)
        : ListScreenshots();
    Promise.resolve(load)
        .then((list: Handlers.ScreenshotInfo[] | null | undefined) => {
            shots.value = list ?? [];
        })
        .catch(() => {
            error.value = 'No se pudieron cargar las capturas.';
        })
        .finally(() => {
            loading.value = false;
        });
}

function previewIndexByPath(path: string): number {
    return Math.max(0, shots.value.findIndex((s) => s.path === path));
}

function openPreview(path: string) {
    previewIndex.value = previewIndexByPath(path);
    resetPreviewZoom();
}

function close() {
    previewIndex.value = null;
    resetPreviewZoom();
    if (shotsReturn.value) {
        openHeavyPanel('instances');
    } else {
        closeHeavyPanel('shots');
    }
}

function onCloseOverlays() {
    close();
}

function onShotsRefresh() {
    refresh();
}

function resetPreviewZoom() {
    zoom.value = 1;
    pan.x = 0;
    pan.y = 0;
    baseSize.value = null;
    dragState.value = null;
}

function closePreview() {
    previewIndex.value = null;
    resetPreviewZoom();
}

function move(dir: number) {
    if (previewIndex.value === null || !shots.value.length) return;
    previewIndex.value = (previewIndex.value + dir + shots.value.length) % shots.value.length;
    resetPreviewZoom();
}

function setZoom(v: number) {
    zoom.value = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, Math.round(v * 100) / 100));
    clampPan();
}

function biteZoom(delta: number) {
    setZoom(zoom.value + delta);
}

function clampPan() {
    const st = stageEl.value;
    const bs = baseSize.value;
    if (!st || !bs) return;
    const vw = bs.w * zoom.value;
    const vh = bs.h * zoom.value;
    const mx = Math.max(0, (vw - st.clientWidth) / 2);
    const my = Math.max(0, (vh - st.clientHeight) / 2);
    pan.x = Math.max(-mx, Math.min(mx, pan.x));
    pan.y = Math.max(-my, Math.min(my, pan.y));
}

function onWheel(e: WheelEvent) {
    const st = stageEl.value;
    const bs = baseSize.value;
    if (previewIndex.value === null || !st || !bs) return;
    const next = e.deltaY < 0 ? zoom.value * 1.12 : zoom.value / 1.12;
    if (next < MIN_ZOOM - 0.001 || next > MAX_ZOOM + 0.001) return;
    const rect = st.getBoundingClientRect();
    const cx = e.clientX - rect.left - st.clientWidth / 2;
    const cy = e.clientY - rect.top - st.clientHeight / 2;
    const ratio = next / zoom.value;
    pan.x = cx - (cx - pan.x) * ratio;
    pan.y = cy - (cy - pan.y) * ratio;
    zoom.value = Math.round(next * 100) / 100;
    clampPan();
}

function onPanStart(e: MouseEvent) {
    const st = stageEl.value;
    if (!st || zoom.value <= 1) return;
    dragState.value = { x: e.clientX, y: e.clientY, sx: pan.x, sy: pan.y };
}

function onPanMove(e: MouseEvent) {
    const d = dragState.value;
    if (!d) return;
    pan.x = d.sx + (e.clientX - d.x);
    pan.y = d.sy + (e.clientY - d.y);
    clampPan();
}

function onPanEnd() {
    dragState.value = null;
}

function onImgLoad(e: Event) {
    const img = e.target as HTMLImageElement;
    const st = stageEl.value;
    if (!st || img.naturalWidth <= 0) return;
    const fit = Math.min(
        1,
        (st.clientWidth * 0.9) / img.naturalWidth,
        (st.clientHeight * 0.8) / img.naturalHeight
    );
    baseSize.value = {
        w: Math.round(img.naturalWidth * fit),
        h: Math.round(img.naturalHeight * fit),
    };
    clampPan();
}

function onKeydown(e: KeyboardEvent) {
    if (previewIndex.value === null) return;
    if (e.key === 'ArrowLeft') move(-1);
    if (e.key === 'ArrowRight') move(1);
    if (e.key === '+' || e.key === '=') biteZoom(0.25);
    if (e.key === '-') biteZoom(-0.25);
}

useOverlayEscape(
    () => (previewIndex.value !== null ? closePreview() : close()),
    { isActive: () => heavyPanel.value === 'shots' }
);

function fmtSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

watch(
    [heavyPanel, shotsInstance],
    () => {
        if (heavyPanel.value === 'shots') refresh();
    }
);

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    window.addEventListener(SHOTS_REFRESH_EVENT, onShotsRefresh);
    window.addEventListener('keydown', onKeydown);
    refresh();
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    window.removeEventListener(SHOTS_REFRESH_EVENT, onShotsRefresh);
    window.removeEventListener('keydown', onKeydown);
});
</script>

<template>
    <div class="Shots_Overlay">
        <div v-if="preview" class="Shots_Viewer" @click.self="closePreview">
                    <div class="Shots_PreviewHead">
                        <button class="Shots_PreviewBack" title="Volver a la galería (Esc)" @click="closePreview">
                            <IconChevronLeft stroke="2" /> Galería
                        </button>
                        <span class="Shots_PreviewName" :title="preview.name">{{ preview.name }}</span>
                        <span class="Shots_PreviewCount">{{ (previewIndex ?? 0) + 1 }} / {{ shots.length }}</span>
                        <div class="Shots_PreviewTools">
                            <button title="Cerrar (Esc)" @click="close">
                                <IconX stroke="2" />
                            </button>
                        </div>
                    </div>
                    <div
                        ref="stageEl"
                        class="Shots_PreviewStage"
                        :class="{ zoomed: zoom > 1 }"
                        @wheel.prevent="onWheel"
                        @mousedown="onPanStart"
                        @mousemove="onPanMove"
                        @mouseup="onPanEnd"
                        @mouseleave="onPanEnd"
                    >
                        <img
                            v-if="previewUrl"
                            :src="previewUrl"
                            :alt="preview.name"
                            class="Shots_PreviewImg"
                            :style="imgStyle"
                            draggable="false"
                            @load="onImgLoad"
                        />
                        <IconPhoto v-else class="Shots_PreviewFallback" stroke="1.5" />
                    </div>
                    <div class="Shots_PreviewZoomBar">
                        <button title="Alejar (−)" @click.stop="biteZoom(-0.25)">
                            <IconZoomOut stroke="2" />
                        </button>
                        <span>{{ Math.round(zoom * 100) }}%</span>
                        <button title="Acercar (+)" @click.stop="biteZoom(0.25)">
                            <IconZoomIn stroke="2" />
                        </button>
                    </div>
                    <button
                        v-if="shots.length > 1"
                        class="Shots_Nav Shots_NavPrev"
                        title="Anterior (←)"
                        @click.stop="move(-1)"
                    >
                        <IconChevronLeft stroke="2" />
                    </button>
                    <button
                        v-if="shots.length > 1"
                        class="Shots_Nav Shots_NavNext"
                        title="Siguiente (→)"
                        @click.stop="move(1)"
                    >
                        <IconChevronRight stroke="2" />
                    </button>
                </div>

                <div v-else class="Shots_Dialog">
                    <div class="Shots_Head">
                        <span class="Shots_Icon"><IconPhoto stroke="2" /></span>
                        <div class="Shots_Titles">
                            <h3>{{ shotsInstance ? `Capturas de ${shotsInstance}` : 'Fotos' }}</h3>
                            <p>{{ shotsInstance ? 'Capturas de Minecraft tomadas dentro de esta instancia' : 'Capturas de Minecraft guardadas en el juego' }}</p>
                        </div>
                        <button class="Shots_Close" title="Cerrar" @click="close">
                            <IconX stroke="2" />
                        </button>
                    </div>

                    <div class="Shots_Body">
                        <p v-if="error" class="Shots_Empty">{{ error }}</p>
                        <p v-else-if="loading" class="Shots_Empty">Buscando capturas…</p>
                        <p v-else-if="!shots.length" class="Shots_Empty">
                            Aún no hay capturas. Toma una en Minecraft (F2) y, al cerrar el juego, aparecerá aquí.
                        </p>
                        <div v-else class="Shots_Grid">
                            <button
                                v-for="shot in shots"
                                :key="shot.path"
                                class="Shots_Card"
                                @click="openPreview(shot.path)"
                            >
                                <span class="Shots_Thumb">
                                    <img
                                        v-if="thumbOf(shot.path)"
                                        :src="thumbOf(shot.path)"
                                        :alt="shot.name"
                                        loading="lazy"
                                    />
                                    <IconPhoto v-else class="Shots_ThumbFallback" stroke="1.5" />
                                </span>
                                <span class="Shots_Meta">
                                    <span class="Shots_Name" :title="shot.name">{{ shot.name }}</span>
                                    <span class="Shots_Sub">{{ fmtSize(shot.size) }} · {{ shot.time }}</span>
                                </span>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
</template>

<style scoped lang="scss">
@use './Styles/Screenshots.scss';
</style>