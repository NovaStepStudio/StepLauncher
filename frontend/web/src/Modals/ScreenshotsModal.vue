<script setup lang="ts">
import { ref, computed, reactive, watch, onMounted, onUnmounted } from 'vue';
import { IconPhoto, IconX, IconZoomIn, IconZoomOut, IconChevronLeft, IconChevronRight } from '@tabler/icons-vue';
import { ListScreenshots, ReadLocalFile } from '@wailsjs/go/main/App';
import type { Handlers } from '@wailsjs/go/models';
import { CLOSE_OVERLAYS_EVENT } from '../Stores/Idle';

const SHOTS_REFRESH_EVENT = 'sl:shots-refresh';

const props = defineProps<{
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
}>();

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
    if (!props.visible) return;
    loading.value = true;
    error.value = '';
    ListScreenshots()
        .then((list) => {
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
    emit('update:visible', false);
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
    if (!props.visible) return;
    if (e.key === 'Escape') {
        if (previewIndex.value !== null) {
            closePreview();
        } else {
            close();
        }
        return;
    }
    if (previewIndex.value === null) return;
    if (e.key === 'ArrowLeft') move(-1);
    if (e.key === 'ArrowRight') move(1);
    if (e.key === '+' || e.key === '=') biteZoom(0.25);
    if (e.key === '-') biteZoom(-0.25);
}

function fmtSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

watch(
    () => props.visible,
    (v) => {
        if (v) {
            refresh();
        } else {
            previewIndex.value = null;
            resetPreviewZoom();
        }
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
    <Teleport to="body">
        <Transition name="ScreenshotsModal">
            <div v-if="visible" class="Shots_Overlay" @click.self="close">
                <div class="Shots_Dialog">
                    <div class="Shots_Head">
                        <span class="Shots_Icon"><IconPhoto stroke="2" /></span>
                        <div class="Shots_Titles">
                            <h3>Fotos</h3>
                            <p>Capturas de Minecraft guardadas en el juego</p>
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

                <Transition name="ScreenshotsPreview">
                    <div v-if="preview" class="Shots_Preview" @click.self="closePreview">
                        <div class="Shots_PreviewHead">
                            <span class="Shots_PreviewName" :title="preview.name">{{ preview.name }}</span>
                            <div class="Shots_PreviewTools">
                                <button title="Cerrar (Esc)" @click="closePreview">
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
                </Transition>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use '../Styles/Settings.scss' as *;

.ScreenshotsModal-enter-active,
.ScreenshotsModal-leave-active {
    transition: opacity 160ms ease;
}
.ScreenshotsModal-enter-from,
.ScreenshotsModal-leave-to {
    opacity: 0;
}

.ScreenshotsPreview-enter-active,
.ScreenshotsPreview-leave-active {
    transition: opacity 150ms ease;
}
.ScreenshotsPreview-enter-from,
.ScreenshotsPreview-leave-to {
    opacity: 0;
}

.Shots_Overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: var(--filter-blur);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 130;
}

.Shots_Dialog {
    width: 100%;
    height: 100%;
    background: var(--background-modal-primary);
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.Shots_Head {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1.1rem 1.35rem 0.85rem;
    border-bottom: var(--border-modal-style);
    background: #0005;
}

.Shots_Icon {
    width: 2.4rem;
    height: 2.4rem;
    flex-shrink: 0;
    border-radius: 0.6rem;
    background: linear-gradient(135deg, color-mix(in srgb, var(--background-button-primary) 35%, transparent), color-mix(in srgb, var(--background-button-primary) 10%, transparent));
    border: 1px solid color-mix(in srgb, var(--background-button-primary) 30%, transparent);
    display: flex;
    justify-content: center;
    align-items: center;
    color: color-mix(in srgb, var(--background-button-primary) 55%, white 45%);
}

.Shots_Titles {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    min-width: 0;

    h3 {
        font-family: var(--font-primary), Arial, sans-serif;
        font-size: 1.05rem;
        margin: 0;
        color: var(--text-primary);
    }

    p {
        font-size: 0.72rem;
        margin: 0;
        opacity: 0.55;
        color: var(--text-secondary);
    }
}

.Shots_Close {
    width: 2rem;
    height: 2rem;
    flex-shrink: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    border-radius: 0.5rem;
    border: 1px solid var(--control-border);
    background: var(--control-bg);
    color: var(--text-primary);
    cursor: pointer;
    transition: background 120ms, border-color 120ms;

    &:hover {
        background: color-mix(in srgb, var(--color-error) 20%, transparent);
        border-color: color-mix(in srgb, var(--color-error) 40%, transparent);
        color: var(--color-error);
    }
}

.Shots_Body {
    flex: 1;
    overflow-y: auto;
    padding: 1rem 1.35rem 1.25rem;
}

.Shots_Empty {
    margin: 0;
    padding: 2.5rem 1rem;
    text-align: center;
    opacity: 0.6;
    font-size: 0.82rem;
}

.Shots_Grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(13.5rem, 1fr));
    gap: 0.9rem;
}

.Shots_Card {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.55rem;
    border-radius: 0.7rem;
    border: 1px solid var(--control-border);
    background: var(--control-bg);
    cursor: pointer;
    text-align: left;
    transition: border-color 150ms, background 150ms, transform 150ms;

    &:hover {
        border-color: color-mix(in srgb, var(--background-button-primary) 45%, transparent);
        background: color-mix(in srgb, var(--background-button-primary) 8%, transparent);
        transform: translateY(-2px);
    }
}

.Shots_Thumb {
    aspect-ratio: 16 / 9;
    border-radius: 0.45rem;
    overflow: hidden;
    display: flex;
    justify-content: center;
    align-items: center;
    background: #000a;

    img {
        width: 100%;
        height: 100%;
        object-fit: cover;
    }
}

.Shots_ThumbFallback {
    width: 1.6rem;
    height: 1.6rem;
    opacity: 0.4;
}

.Shots_Meta {
    display: flex;
    flex-direction: column;
    gap: 0.12rem;
    min-width: 0;
}

.Shots_Name {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.Shots_Sub {
    font-size: 0.7rem;
    letter-spacing: 0.01em;
    color: var(--text-secondary);
    opacity: 0.9;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.Shots_Preview {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.82);
    backdrop-filter: var(--filter-blur);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 140;
}

.Shots_PreviewHead {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    z-index: 5;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    padding: 0.8rem 1rem;
    background: linear-gradient(to bottom, #0008, transparent);
}

.Shots_PreviewName {
    font-size: 0.8rem;
    color: #fff;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.Shots_PreviewTools {
    display: flex;
    gap: 0.4rem;
    flex-shrink: 0;

    button {
        width: 2rem;
        height: 2rem;
        display: flex;
        justify-content: center;
        align-items: center;
        border-radius: 0.5rem;
        border: 1px solid var(--control-border-strong);
        background: var(--control-bg-soft);
        color: #fff;
        cursor: pointer;
        transition: background 150ms, border-color 150ms;

        &:hover {
            background: color-mix(in srgb, var(--background-button-primary) 25%, transparent);
            border-color: color-mix(in srgb, var(--background-button-primary) 50%, transparent);
        }
    }
}

.Shots_PreviewStage {
    position: absolute;
    top: 3.4rem;
    left: 0;
    right: 0;
    bottom: 0;
    overflow: hidden;
    cursor: default;

    &.zoomed {
        cursor: grab;

        &:active {
            cursor: grabbing;
        }
    }
}

.Shots_PreviewImg {
    position: absolute;
    left: 50%;
    top: 50%;
    width: auto;
    height: auto;
    max-width: 90%;
    max-height: 80%;
    aspect-ratio: 16 / 9;
    object-fit: contain;
    transform-origin: center;
    will-change: transform;
    border-radius: 0.25rem;
    user-select: none;
    -webkit-user-drag: none;
}

.Shots_PreviewFallback {
    position: absolute;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    width: 3rem;
    height: 3rem;
    opacity: 0.4;
}

.Shots_PreviewZoomBar {
    position: absolute;
    left: 50%;
    bottom: 1.1rem;
    transform: translateX(-50%);
    z-index: 7;
    display: flex;
    align-items: center;
    gap: 0.15rem;
    padding: 0.25rem;
    border-radius: 0.65rem;
    background: rgba(10, 10, 14, 0.75);
    border: 1px solid var(--control-border-strong);
    backdrop-filter: blur(6px);

    button {
        width: 1.9rem;
        height: 1.9rem;
        display: flex;
        justify-content: center;
        align-items: center;
        border-radius: 0.45rem;
        border: none;
        background: transparent;
        color: #fff;
        cursor: pointer;
        transition: background 120ms;

        &:hover {
            background: rgba(255, 255, 255, 0.12);
        }
    }

    span {
        min-width: 3rem;
        text-align: center;
        font-size: 0.72rem;
        font-weight: 600;
        font-variant-numeric: tabular-nums;
        color: #fff;
        user-select: none;
    }
}

.Shots_Nav {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    z-index: 6;
    width: 2.6rem;
    height: 4.5rem;
    display: flex;
    justify-content: center;
    align-items: center;
    border-radius: 0.5rem;
    border: 1px solid var(--control-border-strong);
    background: var(--control-bg-soft);
    color: #fff;
    cursor: pointer;
    transition: background 150ms, border-color 150ms;

    &:hover {
        background: color-mix(in srgb, var(--background-button-primary) 25%, transparent);
        border-color: color-mix(in srgb, var(--background-button-primary) 50%, transparent);
    }
}

.Shots_NavPrev {
    left: 0.8rem;
}

.Shots_NavNext {
    right: 0.8rem;
}
</style>