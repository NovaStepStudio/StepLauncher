<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue';
import { nextColorFieldId, openColorFieldId, previewColorFieldId } from '../../../Stores/Colorfield';
import { CLOSE_OVERLAYS_EVENT } from '../../../Stores/Idle';

interface Rgb { r: number; g: number; b: number }

const props = withDefaults(defineProps<{ modelValue?: string; recents?: string[]; preview?: boolean; example?: boolean }>(), {
    modelValue: '#000000',
    recents: () => [],
    preview: false,
    example: false,
});
const emit = defineEmits(['update:modelValue', 'preview']);

const open = ref(false);
const hue = ref(0);
const sat = ref(0);
const val = ref(0);
const alpha = ref(100);
const format = ref<'hex' | 'rgb'>('hex');

const id = nextColorFieldId();
const pinned = ref(false);
let closeTimer: number | null = null;
let dragging = false;
let moveX = 0;
let moveY = 0;
let dragRect: DOMRect | null = null;
let anchorSat = 0;
let anchorVal = 0;
let anchorHue = 0;
let anchorAlpha = 0;
let dragKind: 'sv' | 'hue' | 'alpha' | null = null;
let lockActive = false;
let skipMove = 0;

function scheduleClose(ms = 5000) {
    cancelClose();
    closeTimer = window.setTimeout(() => {
        closeTimer = null;
        commitIfDirty();
        open.value = false;
        pinned.value = false;
        setPreviewing(false);
        if (openColorFieldId.value === id) openColorFieldId.value = null;
    }, ms);
}

function cancelClose() {
    if (closeTimer !== null) {
        window.clearTimeout(closeTimer);
        closeTimer = null;
    }
}

function closeNow() {
    cancelClose();
    unlockPointer();
    commitIfDirty();
    open.value = false;
    pinned.value = false;
    setPreviewing(false);
    if (openColorFieldId.value === id) openColorFieldId.value = null;
}

function onToggle() {
    if (open.value) {
        closeNow();
    } else {
        open.value = true;
        pinned.value = true;
        openColorFieldId.value = id;
        scheduleClose();
    }
}

function onLeave() {
    if (dragging) return;
    if (pinned.value) {
        scheduleClose();
    }
}

function onPanelDown() {
    dragging = true;
    cancelClose();
}

function onPanelUp() {
    dragging = false;
    if (open.value && pinned.value) {
        scheduleClose();
    }
}

function endInteraction() {
    dragging = false;
    if (open.value && pinned.value) {
        scheduleClose();
    }
}

function tryLock(el: HTMLElement) {
    if (document.pointerLockElement) return;
    try {
        const p = el.requestPointerLock() as unknown as Promise<void>;
        if (p && typeof p.catch === 'function') p.catch(() => {});
    } catch {
    }
}

function unlockPointer() {
    if (document.pointerLockElement) {
        document.exitPointerLock();
    }
}

function onPointerLockChange() {
    const el = document.pointerLockElement;
    lockActive = el === svCanvas.value || el === hueBar.value || el === alphaBar.value;
    if (lockActive) {
        skipMove = 1;
    }
}

function onCloseOverlays() {
    closeNow();
}

watch(openColorFieldId, (v) => {
    if (v !== null && v !== id) closeNow();
});

const previewing = ref(false);
const rootEl = ref<HTMLElement | null>(null);
let previewAnim: Animation | null = null;

async function setPreviewing(v: boolean) {
    if (!props.preview) return;
    const el = rootEl.value;
    if (v) {
        if (previewing.value) return;
        const from = el ? el.getBoundingClientRect() : null;
        previewing.value = true;
        previewColorFieldId.value = id;
        await nextTick();
        const to = el ? el.getBoundingClientRect() : null;
        if (el && from && to) {
            previewAnim?.cancel();
            const dx = from.left - to.left;
            const dy = from.top - to.top;
            previewAnim = el.animate(
                [
                    { transform: `translate(calc(-50% + ${dx}px), calc(-50% + ${dy}px)) scale(.92)`, opacity: 0 },
                    { transform: 'translate(-50%, -50%) scale(1)', opacity: 1 },
                ],
                { duration: 200, easing: 'cubic-bezier(.165, .84, .44, 1)', fill: 'both' }
            );
        }
    } else {
        if (!previewing.value) return;
        const from = el ? el.getBoundingClientRect() : null;
        previewing.value = false;
        previewColorFieldId.value = null;
        await nextTick();
        const to = el ? el.getBoundingClientRect() : null;
        if (el && from && to) {
            previewAnim?.cancel();
            const dx = from.left - to.left;
            const dy = from.top - to.top;
            previewAnim = el.animate(
                [
                    { transform: `translate(${dx}px, ${dy}px) scale(1)`, opacity: 1 },
                    { transform: 'translate(0px, 0px) scale(.92)', opacity: 1 },
                ],
                { duration: 200, easing: 'cubic-bezier(.165, .84, .44, 1)', fill: 'none' }
            );
        }
    }
}

const svCanvas = ref<HTMLCanvasElement | null>(null);
const hueBar = ref<HTMLDivElement | null>(null);
const alphaBar = ref<HTMLDivElement | null>(null);

const rgb = computed<Rgb>(() => hsvToRgb(hue.value, sat.value, val.value));

function clamp(n: number, min: number, max: number) {
    return Math.min(max, Math.max(min, n));
}

function rgbToHsv(r: number, g: number, b: number): { h: number; s: number; v: number } {
    const rn = r / 255, gn = g / 255, bn = b / 255;
    const max = Math.max(rn, gn, bn), min = Math.min(rn, gn, bn);
    const d = max - min;
    let h = 0;
    if (d !== 0) {
        if (max === rn) h = ((gn - bn) / d) % 6;
        else if (max === gn) h = (bn - rn) / d + 2;
        else h = (rn - gn) / d + 4;
        h *= 60;
        if (h < 0) h += 360;
    }
    return { h, s: max === 0 ? 0 : d / max, v: max };
}

function hsvToRgb(h: number, s: number, v: number): Rgb {
    const c = v * s;
    const x = c * (1 - Math.abs(((h / 60) % 2) - 1));
    const m = v - c;
    let r = 0, g = 0, b = 0;
    if (h < 60) { r = c; g = x; }
    else if (h < 120) { r = x; g = c; }
    else if (h < 180) { g = c; b = x; }
    else if (h < 240) { g = x; b = c; }
    else if (h < 300) { r = x; b = c; }
    else { r = c; b = x; }
    return { r: Math.round((r + m) * 255), g: Math.round((g + m) * 255), b: Math.round((b + m) * 255) };
}

function parseColor(v: string): { r: number; g: number; b: number; a: number } | null {
    const t = (v ?? '').trim();
    const hex = t.match(/^#([0-9a-fA-F]{3}|[0-9a-fA-F]{4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/);
    if (hex) {
        let h = hex[1] ?? '';
        let a = 100;
        if (h.length === 8) {
            a = Math.round((parseInt(h.slice(6, 8), 16) / 255) * 100);
            h = h.slice(0, 6);
        } else if (h.length === 4) {
            a = Math.round((parseInt(h.slice(3, 4), 16) * 17) / 2.55);
            h = h.slice(0, 3);
        }
        if (h.length === 3) {
            h = h.slice(0, 1) + h.slice(0, 1) + h.slice(1, 2) + h.slice(1, 2) + h.slice(2, 3) + h.slice(2, 3);
        }
        return { r: parseInt(h.slice(0, 2), 16), g: parseInt(h.slice(2, 4), 16), b: parseInt(h.slice(4, 6), 16), a };
    }
    const m = t.match(/^rgba?\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*(?:,\s*([\d.]+)\s*)?\)$/);
    if (m) {
        return {
            r: clamp(parseInt(m[1] ?? '0', 10), 0, 255),
            g: clamp(parseInt(m[2] ?? '0', 10), 0, 255),
            b: clamp(parseInt(m[3] ?? '0', 10), 0, 255),
            a: m[4] === undefined ? 100 : Math.round(clamp(parseFloat(m[4]), 0, 1) * 100),
        };
    }
    return null;
}

function syncFromValue(v: string) {
    const c = parseColor(v);
    if (!c) return;
    const hsv = rgbToHsv(c.r, c.g, c.b);
    hue.value = hsv.h;
    sat.value = hsv.s;
    val.value = hsv.v;
    alpha.value = c.a;
    drawSV();
    lastEmitted = outputText();
}

let lastEmitted = '';

watch(() => props.modelValue, (v) => {
    const s = String(v ?? '').trim();
    if (s === lastEmitted) return;
    if (s === outputText()) { lastEmitted = s; return; }
    syncFromValue(v);
}, { immediate: true });

function onWindowBlur() {
    if (!dragKind) return;
    dragKind = null;
    dragRect = null;
    unlockPointer();
    commit();
    setPreviewing(false);
    endInteraction();
}

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    document.addEventListener('pointerlockchange', onPointerLockChange);
    window.addEventListener('blur', onWindowBlur);
    window.addEventListener('keydown', onKeydown);
    drawSV();
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    document.removeEventListener('pointerlockchange', onPointerLockChange);
    window.removeEventListener('blur', onWindowBlur);
    window.removeEventListener('keydown', onKeydown);
    unlockPointer();
    if (closeTimer !== null) {
        window.clearTimeout(closeTimer);
        closeTimer = null;
    }
    previewAnim?.cancel();
    if (previewColorFieldId.value === id) {
        previewColorFieldId.value = null;
    }
});

function drawSV() {
    const c = svCanvas.value;
    if (!c) return;
    const ctx = c.getContext('2d');
    if (!ctx) return;
    const w = c.width, h = c.height;
    ctx.fillStyle = `hsl(${hue.value}, 100%, 50%)`;
    ctx.fillRect(0, 0, w, h);
    const gw = ctx.createLinearGradient(0, 0, w, 0);
    gw.addColorStop(0, '#ffffff');
    gw.addColorStop(1, 'rgba(255,255,255,0)');
    ctx.fillStyle = gw;
    ctx.fillRect(0, 0, w, h);
    const gb = ctx.createLinearGradient(0, 0, 0, h);
    gb.addColorStop(0, 'rgba(0,0,0,0)');
    gb.addColorStop(1, '#000000');
    ctx.fillStyle = gb;
    ctx.fillRect(0, 0, w, h);
    const px = sat.value * w;
    const py = (1 - val.value) * h;
    ctx.beginPath();
    ctx.arc(px, py, 8, 0, Math.PI * 2);
    ctx.strokeStyle = '#ffffff';
    ctx.lineWidth = 2.5;
    ctx.stroke();
    ctx.beginPath();
    ctx.arc(px, py, 6.5, 0, Math.PI * 2);
    ctx.strokeStyle = 'rgba(0,0,0,0.55)';
    ctx.lineWidth = 1;
    ctx.stroke();
}

function onSVDown(e: PointerEvent) {
    const c = svCanvas.value;
    if (!c) return;
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    const rect = c.getBoundingClientRect();
    sat.value = clamp((e.clientX - rect.left) / rect.width, 0, 1);
    val.value = 1 - clamp((e.clientY - rect.top) / rect.height, 0, 1);
    drawSV();
    anchorSat = sat.value;
    anchorVal = val.value;
    moveX = 0;
    moveY = 0;
    dragRect = rect;
    dragKind = 'sv';
    skipMove = 0;
    tryLock(c);
    setPreviewing(true);
    emit('preview', outputText());
}

function onSVDrag(e: PointerEvent) {
    if (dragKind !== 'sv') return;
    if (!dragRect) return;
    if (!lockActive && e.buttons === 0) {
        onSVUp();
        return;
    }
    if (skipMove > 0) {
        skipMove--;
        return;
    }
    moveX += e.movementX;
    moveY += e.movementY;
    sat.value = clamp(anchorSat + moveX / dragRect.width, 0, 1);
    val.value = clamp(anchorVal - moveY / dragRect.height, 0, 1);
    drawSV();
    emit('preview', outputText());
}

function onSVUp() {
    dragKind = null;
    unlockPointer();
    dragRect = null;
    commit();
    setPreviewing(false);
    endInteraction();
}

function onWheelHue(e: WheelEvent) {
    hue.value = clamp(hue.value + e.deltaY * 0.2, 0, 360);
    drawSV();
    emit('preview', outputText());
}

function onKeydown(e: KeyboardEvent) {
    if (!open.value) return;
    if (e.key !== 'ArrowUp' && e.key !== 'ArrowDown') return;
    const t = e.target as HTMLElement | null;
    if (t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.isContentEditable)) return;
    e.preventDefault();
    const step = e.shiftKey ? 15 : 5;
    hue.value = clamp(hue.value + (e.key === 'ArrowUp' ? -step : step), 0, 360);
    drawSV();
    emit('preview', outputText());
}

const hueThumbPos = computed(() => `${(hue.value / 360) * 100}%`);
const hueBg = 'linear-gradient(to bottom, #f00 0%, #ff0 17%, #0f0 33%, #0ff 50%, #00f 67%, #f0f 83%, #f00 100%)';

function onHueDown(e: PointerEvent) {
    const el = hueBar.value;
    if (!el) return;
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    const rect = el.getBoundingClientRect();
    hue.value = clamp((e.clientY - rect.top) / rect.height, 0, 1) * 360;
    drawSV();
    anchorHue = hue.value;
    moveY = 0;
    dragRect = rect;
    dragKind = 'hue';
    skipMove = 0;
    tryLock(e.currentTarget as HTMLElement);
    setPreviewing(true);
    emit('preview', outputText());
}

function onHueDrag(e: PointerEvent) {
    if (dragKind !== 'hue') return;
    if (!dragRect) return;
    if (e.buttons === 0) {
        onHueUp();
        return;
    }
    if (skipMove > 0) {
        skipMove--;
        return;
    }
    moveY += e.movementY;
    hue.value = clamp(anchorHue + (moveY / dragRect.height) * 360, 0, 360);
    drawSV();
    emit('preview', outputText());
}

function onHueUp() {
    dragKind = null;
    unlockPointer();
    dragRect = null;
    commit();
    setPreviewing(false);
    endInteraction();
}

const alphaThumbPos = computed(() => `${alpha.value}%`);
const alphaBg = computed(() =>
    `linear-gradient(to right, rgba(${rgb.value.r},${rgb.value.g},${rgb.value.b},0) 0%, rgba(${rgb.value.r},${rgb.value.g},${rgb.value.b},1) 100%)`
);
const alphaStyle = computed(() => ({
    backgroundImage: `${alphaBg.value}, linear-gradient(45deg, rgba(255,255,255,0.14) 25%, transparent 25%, transparent 75%, rgba(255,255,255,0.14) 75%), linear-gradient(45deg, rgba(255,255,255,0.14) 25%, transparent 25%, transparent 75%, rgba(255,255,255,0.14) 75%)`,
    backgroundSize: '100% 100%, 10px 10px, 10px 10px',
    backgroundPosition: '0 0, 0 0, 5px 5px',
}));

function onAlphaDown(e: PointerEvent) {
    const el = alphaBar.value;
    if (!el) return;
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    const rect = el.getBoundingClientRect();
    alpha.value = Math.round(clamp((e.clientX - rect.left) / rect.width, 0, 1) * 100);
    anchorAlpha = alpha.value;
    moveX = 0;
    dragRect = rect;
    dragKind = 'alpha';
    skipMove = 0;
    tryLock(e.currentTarget as HTMLElement);
    setPreviewing(true);
    emit('preview', outputText());
}

function onAlphaDrag(e: PointerEvent) {
    if (dragKind !== 'alpha') return;
    if (!dragRect) return;
    if (e.buttons === 0) {
        onAlphaUp();
        return;
    }
    if (skipMove > 0) {
        skipMove--;
        return;
    }
    moveX += e.movementX;
    alpha.value = Math.round(clamp(anchorAlpha + (moveX / dragRect.width) * 100, 0, 100));
    emit('preview', outputText());
}

function onAlphaUp() {
    dragKind = null;
    unlockPointer();
    dragRect = null;
    commit();
    setPreviewing(false);
    endInteraction();
}

function alphaText(a: number): string {
    const v = (a / 100).toFixed(2);
    const clean = v.replace(/0+$/, '').replace(/\.$/, '');
    return clean === '' ? '0' : clean;
}

function outputText(): string {
    const p = rgb.value;
    if (format.value === 'hex') {
        const to2 = (n: number) => n.toString(16).padStart(2, '0');
        if (alpha.value >= 100) return `#${to2(p.r)}${to2(p.g)}${to2(p.b)}`;
        return `#${to2(p.r)}${to2(p.g)}${to2(p.b)}${to2(Math.round((alpha.value / 100) * 255))}`;
    }
    if (alpha.value >= 100) return `rgb(${p.r}, ${p.g}, ${p.b})`;
    return `rgba(${p.r}, ${p.g}, ${p.b}, ${alphaText(alpha.value)})`;
}

function commit() {
    const v = outputText();
    lastEmitted = v;
    emit('update:modelValue', v);
}

function commitIfDirty() {
    if (outputText() !== lastEmitted) commit();
}

const previewStyle = computed(() => `rgba(${rgb.value.r}, ${rgb.value.g}, ${rgb.value.b}, ${alpha.value / 100})`);
const shortValue = computed(() => {
    const v = String(props.modelValue ?? '').trim();
    return v.length > 16 ? v.slice(0, 15) + '…' : v;
});

function onSnippet(e: Event) {
    const v = (e.target as HTMLInputElement).value;
    if (!parseColor(v)) return;
    syncFromValue(v);
    commit();
}

function onSnippetFocus() {
    cancelClose();
}

function onSnippetBlur() {
    if (pinned.value) {
        scheduleClose();
    }
}

function applyRecent(v: string) {
    syncFromValue(v);
    commit();
}
</script>

<template>
<div ref="rootEl" class="Cf" :class="{ open, previewing }" @mouseleave="onLeave">
    <button type="button" class="CfTrigger" @click="onToggle" title="Clic para fijar el selector (se cierra solo a los 5s)">
        <span class="CfSwatch"><span class="CfSwatchColor" :style="{ background: previewStyle }"></span></span>
        <span class="CfValue">{{ shortValue }}</span>
        <svg class="CfCaret" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
    </button>

    <div class="CfBody" @pointerdown="onPanelDown" @pointerup="onPanelUp">
        <div class="CfBodyInner">
            <div class="CfPanel">
                <div class="CfMain">
                    <canvas
                        ref="svCanvas"
                        class="CfSV"
                        width="220"
                        height="220"
                        @pointerdown="onSVDown"
                        @pointermove="onSVDrag"
                        @pointerup="onSVUp"
                        @pointercancel="onSVUp"
                        @wheel.prevent="onWheelHue"
                    ></canvas>
                    <div
                        ref="hueBar"
                        class="CfHue"
                        :style="{ background: hueBg }"
                        @pointerdown="onHueDown"
                        @pointermove="onHueDrag"
                        @pointerup="onHueUp"
                        @pointercancel="onHueUp"
                        @wheel.prevent="onWheelHue"
                    >
                        <span class="CfThumb CfHueThumb" :style="{ top: hueThumbPos }"></span>
                    </div>
                </div>

                <div class="CfAlpha">
                    <div
                        ref="alphaBar"
                        class="CfAlphaBar"
                        :style="alphaStyle"
                        @pointerdown="onAlphaDown"
                        @pointermove="onAlphaDrag"
                        @pointerup="onAlphaUp"
                        @pointercancel="onAlphaUp"
                    >
                        <span class="CfThumb CfAlphaThumb" :style="{ left: alphaThumbPos }"></span>
                    </div>
                    <span class="CfAlphaVal">{{ alpha }}%</span>
                </div>

                <div class="CfOut">
                    <div class="CfFmt">
                        <button :class="{ active: format === 'hex' }" @click="format = 'hex'">HEX</button>
                        <button :class="{ active: format === 'rgb' }" @click="format = 'rgb'">RGB</button>
                    </div>
                    <input class="CfSnippet" :value="outputText()" spellcheck="false" @change="onSnippet" @focus="onSnippetFocus" @input="onSnippetFocus" @blur="onSnippetBlur">
                </div>

                <div v-if="recents.length" class="CfRecents">
                    <button
                        v-for="c in recents"
                        :key="c"
                        class="CfRecent"
                        :style="{ background: c }"
                        :class="{ active: c.toLowerCase() === String(modelValue).toLowerCase() }"
                        :title="c"
                        @click="applyRecent(c)"
                    ></button>
                </div>

                <div class="CfHint">Rueda o flechas ↑/↓ (Shift: 15°): tono</div>
            </div>
        </div>
    </div>

    <div v-if="props.example && previewing" class="CfReplica" aria-hidden="true">
        <div class="CfReplicaDialog">
            <div class="CfReplicaHead">
                <span class="CfReplicaIcon">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
                </span>
                <div class="CfReplicaTitles">
                    <span class="CfReplicaTitle">Descargando el juego</span>
                    <span class="CfReplicaSub">1.20.1 · Fabric 0.15.11</span>
                </div>
            </div>
            <div class="CfReplicaBody">
                <div class="CfReplicaRing">
                    <svg viewBox="0 0 120 120">
                        <circle class="CfReplicaRingTrack" cx="60" cy="60" r="52" />
                        <circle class="CfReplicaRingFill" cx="60" cy="60" r="52" :stroke="previewStyle" stroke-dasharray="326.7" stroke-dashoffset="117.6" transform="rotate(-90 60 60)" />
                    </svg>
                    <span class="CfReplicaRingPct">64%</span>
                </div>
                <div class="CfReplicaCol">
                    <div class="CfReplicaLine">
                        Descargado <b>1.9 MB</b> de 2.8 MB
                    </div>
                    <div class="CfReplicaLine sub">
                        <b>4.2 MB/s</b> · assets
                    </div>
                    <div class="CfReplicaChip">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
                        Detalles técnicos
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
</template>

<style scoped lang="scss">
.Cf {
    width: 100%;
    max-width: 250px;
}

.Cf.previewing {
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    z-index: 200;
    width: max-content;
    max-width: min(250px, 90vw);
    padding: .4rem;
    border-radius: .6rem;
    background: var(--background-modal-primary);
    border: var(--border-modal-style);
    box-shadow: var(--shadow-settings-normal) #0008;
}

.CfTrigger {
    display: inline-flex;
    align-items: center;
    gap: .5rem;
    width: 100%;
    padding: .35rem .55rem .35rem .4rem;
    border-radius: .45rem;
    border: var(--border-style);
    background: var(--background-modal-primary);
    color: var(--text-secondary);
    text-shadow: var(--text-shadow-secundary, none);
    font-family: var(--font-secundary);
    font-size: calc(.72rem * var(--font-size-secundary, 1));
    cursor: pointer;
    transition: border-color 120ms, background 120ms;

    &:hover,
    .Cf.open & {
        border-color: rgba(255, 255, 255, 0.14);
        background: rgba(255, 255, 255, 0.08);
    }
}

.CfSwatch {
    position: relative;
    width: 1.6rem;
    height: 1.6rem;
    flex-shrink: 0;
    border-radius: .4rem;
    border: var(--border-style);
    overflow: hidden;
    background-image:
        linear-gradient(45deg, rgba(255, 255, 255, 0.14) 25%, transparent 25%, transparent 75%, rgba(255, 255, 255, 0.14) 75%),
        linear-gradient(45deg, rgba(255, 255, 255, 0.14) 25%, transparent 25%, transparent 75%, rgba(255, 255, 255, 0.14) 75%);
    background-size: 10px 10px, 10px 10px;
    background-position: 0 0, 5px 5px;
}

.CfSwatchColor {
    position: absolute;
    inset: 0;
}

.CfValue {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: left;
    font-family: 'Consolas', 'Courier New', monospace;
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
}

.CfCaret {
    flex-shrink: 0;
    opacity: 0.5;
    transition: transform 150ms ease;
}

.Cf.open .CfCaret {
    transform: rotate(180deg);
}

.CfBody {
    display: grid;
    grid-template-rows: 0fr;
    transition: grid-template-rows 200ms ease;

    .Cf.open & {
        grid-template-rows: 1fr;
    }
}

.CfBodyInner {
    overflow: hidden;
}

.CfPanel {
    margin-top: .55rem;
    padding: .7rem;
    border-radius: .6rem;
    border: var(--border-style);
    background: color-mix(in srgb, var(--background-modal-primary) 90%, #fff 5%);
    box-shadow: 0 .6rem 1.2rem rgba(0, 0, 0, 0.35);
}

.CfMain {
    display: flex;
    gap: .7rem;
    align-items: stretch;
}

.CfReplica {
    position: fixed;
    top: 50%;
    left: calc(50% + 9.5rem);
    transform: translateY(-50%);
    z-index: 199;
    pointer-events: none;
    animation: CfReplicaIn 220ms cubic-bezier(.165, .84, .44, 1);
}

@keyframes CfReplicaIn {
    from {
        opacity: 0;
        transform: translateY(-50%) translateX(14px);
    }
    to {
        opacity: 1;
        transform: translateY(-50%) translateX(0);
    }
}

.CfReplicaDialog {
    width: 15.5rem;
    padding: .8rem;
    border-radius: .7rem;
    background: var(--background-modal-primary);
    border: var(--border-modal-style);
    box-shadow: var(--shadow-settings-normal) #0009;
    display: flex;
    flex-direction: column;
    gap: .6rem;
}

.CfReplicaHead {
    display: flex;
    align-items: center;
    gap: .55rem;
}

.CfReplicaIcon {
    width: 1.9rem;
    height: 1.9rem;
    flex-shrink: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    border-radius: .45rem;
    background: color-mix(in srgb, v-bind(previewStyle) 22%, transparent);
    color: var(--text-primary);

    svg {
        width: .95rem;
        height: .95rem;
    }
}

.CfReplicaTitles {
    display: flex;
    flex-direction: column;
    gap: .05rem;
    min-width: 0;
}

.CfReplicaTitle {
    font-family: var(--font-primary);
    font-size: calc(.72rem * var(--font-size-primary, 1));
    font-weight: 600;
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.CfReplicaSub {
    font-family: var(--font-secundary);
    font-size: .58rem;
    opacity: .5;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.CfReplicaBody {
    display: flex;
    align-items: center;
    gap: .7rem;
}

.CfReplicaRing {
    position: relative;
    width: 3.6rem;
    height: 3.6rem;
    flex-shrink: 0;

    svg {
        width: 100%;
        height: 100%;
        display: block;
    }
}

.CfReplicaRingTrack {
    fill: none;
    stroke: rgba(255, 255, 255, 0.07);
    stroke-width: 9;
}

.CfReplicaRingFill {
    fill: none;
    stroke-width: 9;
    stroke-linecap: round;
    transition: stroke 120ms ease;
}

.CfReplicaRingPct {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    font-family: var(--font-primary);
    font-size: calc(.72rem * var(--font-size-primary, 1));
    font-weight: 700;
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
}

.CfReplicaCol {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: .32rem;
}

.CfReplicaLine {
    font-family: var(--font-secundary);
    font-size: .6rem;
    color: var(--text-secondary);
    text-shadow: var(--text-shadow-secundary, none);
    opacity: .75;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;

    b {
        font-weight: 700;
        color: v-bind(previewStyle);
    }

    &.sub {
        opacity: .5;
    }
}

.CfReplicaChip {
    display: inline-flex;
    align-items: center;
    gap: .3rem;
    align-self: flex-start;
    padding: .18rem .5rem;
    border-radius: 99px;
    border: var(--border-style);
    background: rgba(0, 0, 0, 0.2);
    font-family: var(--font-secundary);
    font-size: .56rem;
    color: var(--text-secondary);
    text-shadow: var(--text-shadow-secundary, none);

    svg {
        width: .6rem;
        height: .6rem;
    }
}

.CfSV {
    width: 7rem;
    aspect-ratio: 1 / 1;
    border-radius: .5rem;
    border: var(--border-style);
    cursor: crosshair;
    touch-action: none;
    display: block;
    background: #000;
}

.CfHue {
    position: relative;
    width: .6rem;
    flex-shrink: 0;
    border-radius: .5rem;
    border: var(--border-style);
    cursor: ns-resize;
    touch-action: none;
}

.CfThumb {
    position: absolute;
    width: 1rem;
    height: 1rem;
    border-radius: 50%;
    border: 2px solid #fff;
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.6), 0 1px 4px rgba(0, 0, 0, 0.5);
    pointer-events: none;
}

.CfHueThumb {
    left: 50%;
    transform: translate(-50%, -50%);
}

.CfAlphaThumb {
    top: 50%;
    transform: translate(-50%, -50%);
}

.CfAlpha {
    display: flex;
    align-items: center;
    gap: .55rem;
    padding-top: .6rem;
}

.CfAlphaBar {
    position: relative;
    flex: 1;
    height: 1rem;
    border-radius: .4rem;
    border: var(--border-style);
    background-color: var(--background-modal-primary);
    cursor: ew-resize;
    touch-action: none;
}

.CfAlphaVal {
    flex-shrink: 0;
    width: 2.5rem;
    text-align: right;
    font-size: calc(.7rem * var(--font-size-secundary, 1));
    font-family: var(--font-secundary);
    color: var(--text-secondary);
    text-shadow: var(--text-shadow-secundary, none);
}

.CfOut {
    display: flex;
    align-items: center;
    gap: .45rem;
    padding-top: .6rem;
}

.CfFmt {
    display: flex;
    border: var(--border-style);
    border-radius: .45rem;
    overflow: hidden;
    flex-shrink: 0;

    button {
        padding: .35rem .55rem;
        background: transparent;
        border: none;
        color: var(--text-secondary);
        text-shadow: var(--text-shadow-secundary, none);
        font-family: var(--font-secundary);
        font-size: calc(.68rem * var(--font-size-secundary, 1));
        cursor: pointer;
        transition: background 100ms, color 100ms;

        &.active {
            background: var(--background-sidebar-items);
            color: var(--text-primary);
            text-shadow: var(--text-shadow-primary, none);
        }
    }
}

.CfSnippet {
    flex: 1;
    min-width: 0;
    padding: .4rem .55rem;
    border-radius: .45rem;
    border: var(--border-style);
    background: var(--background-modal-primary);
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
    font-family: 'Consolas', 'Courier New', monospace;
    font-size: .72rem;
    letter-spacing: .02em;

    &:focus {
        border-color: color-mix(in srgb, var(--background-modal-primary) 50%, gray 25%);
        outline: none;
    }
}

.CfRecents {
    display: flex;
    flex-wrap: wrap;
    gap: .4rem;
    padding-top: .6rem;
}

.CfRecent {
    width: 1.55rem;
    height: 1.55rem;
    border-radius: .4rem;
    border: var(--border-style);
    padding: 0;
    cursor: pointer;
    transition: transform 90ms ease, border-color 100ms ease;

    &:hover {
        transform: scale(1.12);
    }

    &.active {
        border-color: var(--text-primary);
        box-shadow: 0 0 0 1px var(--text-primary);
    }
}

.CfHint {
    padding-top: .55rem;
    text-align: center;
    font-family: var(--font-secundary);
    font-size: calc(.62rem * var(--font-size-secundary, 1));
    color: var(--text-secondary);
    text-shadow: var(--text-shadow-secundary, none);
    opacity: .72;
}
</style>
