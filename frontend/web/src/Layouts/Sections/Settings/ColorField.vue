<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';

interface Rgb { r: number; g: number; b: number }

const props = withDefaults(defineProps<{ modelValue?: string; recents?: string[] }>(), {
    modelValue: '#000000',
    recents: () => [],
});
const emit = defineEmits(['update:modelValue']);

const open = ref(false);
const hue = ref(0);
const sat = ref(0);
const val = ref(0);
const alpha = ref(100);
const format = ref<'hex' | 'rgb'>('hex');

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
    const hex = t.match(/^#([0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/);
    if (hex) {
        let h = hex[1] ?? '';
        let a = 100;
        if (h.length === 8) {
            a = Math.round((parseInt(h.slice(6, 8), 16) / 255) * 100);
            h = h.slice(0, 6);
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
}

watch(() => props.modelValue, (v) => {
    if (String(v ?? '').trim() === outputText()) return;
    syncFromValue(v);
}, { immediate: true });

onMounted(drawSV);

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

// El drag solo previsualiza en el popover; el valor real se emite al soltar.

function onSVDown(e: PointerEvent) {
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    const c = svCanvas.value;
    if (!c) return;
    const rect = c.getBoundingClientRect();
    sat.value = clamp((e.clientX - rect.left) / rect.width, 0, 1);
    val.value = 1 - clamp((e.clientY - rect.top) / rect.height, 0, 1);
    drawSV();
}

function onSVDrag(e: PointerEvent) {
    if (!(e.buttons & 1)) return;
    onSVDown(e);
}

function onSVUp() {
    commit();
}

const hueThumbPos = computed(() => `${(hue.value / 360) * 100}%`);
const hueBg = 'linear-gradient(to top, #f00 0%, #ff0 17%, #0f0 33%, #0ff 50%, #00f 67%, #f0f 83%, #f00 100%)';

function onHueDown(e: PointerEvent) {
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    const el = hueBar.value;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    hue.value = clamp((e.clientY - rect.top) / rect.height, 0, 1) * 360;
    drawSV();
}

function onHueDrag(e: PointerEvent) {
    if (!(e.buttons & 1)) return;
    onHueDown(e);
}

function onHueUp() {
    commit();
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
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    const el = alphaBar.value;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    alpha.value = Math.round(clamp((e.clientX - rect.left) / rect.width, 0, 1) * 100);
}

function onAlphaDrag(e: PointerEvent) {
    if (!(e.buttons & 1)) return;
    onAlphaDown(e);
}

function onAlphaUp() {
    commit();
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

// commit guarda el color: solo se llama al soltar el click (o al elegir
// un color del historial o escribir el codigo a mano).
function commit() {
    emit('update:modelValue', outputText());
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

function applyRecent(v: string) {
    syncFromValue(v);
    commit();
}
</script>

<template>
<div class="Cf" :class="{ open }" @mouseenter="open = true" @mouseleave="open = false">
    <button type="button" class="CfTrigger" @click="open = !open" title="Personalizar color">
        <span class="CfSwatch"><span class="CfSwatchColor" :style="{ background: previewStyle }"></span></span>
        <span class="CfValue">{{ shortValue }}</span>
        <svg class="CfCaret" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
    </button>

    <div class="CfBody">
        <div class="CfBodyInner">
            <div class="CfMain">
                <canvas
                    ref="svCanvas"
                    class="CfSV"
                    width="220"
                    height="220"
                    @pointerdown="onSVDown"
                    @pointermove="onSVDrag"
                    @pointerup="onSVUp"
                ></canvas>
                <div
                    ref="hueBar"
                    class="CfHue"
                    :style="{ background: hueBg }"
                    @pointerdown="onHueDown"
                    @pointermove="onHueDrag"
                    @pointerup="onHueUp"
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
                <input class="CfSnippet" :value="outputText()" spellcheck="false" @change="onSnippet">
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
        </div>
    </div>
</div>
</template>

<style scoped lang="scss">
.Cf {
    width: 100%;
    max-width: 250px;
}

.CfTrigger {
    display: inline-flex;
    align-items: center;
    gap: .5rem;
    width: 100%;
    padding: .35rem .55rem .35rem .4rem;
    border-radius: .5rem;
    border: 1px solid var(--color-border);
    background: var(--color-button);
    color: var(--text-secondary);
    font-family: var(--font-secundary);
    font-size: .72rem;
    cursor: pointer;
    transition: border-color 120ms, background 120ms;

    &:hover,
    .Cf.open & {
        border-color: var(--color-border-modal);
        background: color-mix(in srgb, var(--color-button) 85%, #fff 6%);
    }
}

.CfSwatch {
    position: relative;
    width: 1.6rem;
    height: 1.6rem;
    flex-shrink: 0;
    border-radius: .4rem;
    border: 1px solid var(--color-border);
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

.CfMain {
    display: flex;
    gap: .6rem;
    align-items: stretch;
    padding-top: .55rem;
}

.CfSV {
    width: 5rem;
    aspect-ratio: 1 / 1;
    border-radius: .5rem;
    border: 1px solid var(--color-border);
    cursor: crosshair;
    touch-action: none;
    display: block;
    background: #000;
}

.CfHue {
    position: relative;
    width: .5rem;
    flex-shrink: 0;
    border-radius: .5rem;
    border: 1px solid var(--color-border);
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
    border: 1px solid var(--color-border);
    background-color: var(--color-modal);
    cursor: ew-resize;
    touch-action: none;
}

.CfAlphaVal {
    flex-shrink: 0;
    width: 2.5rem;
    text-align: right;
    font-size: .7rem;
    font-family: var(--font-secundary);
    color: var(--text-secondary);
}

.CfOut {
    display: flex;
    align-items: center;
    gap: .45rem;
    padding-top: .6rem;
}

.CfFmt {
    display: flex;
    border: 1px solid var(--color-border);
    border-radius: .45rem;
    overflow: hidden;
    flex-shrink: 0;

    button {
        padding: .35rem .55rem;
        background: transparent;
        border: none;
        color: var(--text-secondary);
        font-family: var(--font-secundary);
        font-size: .68rem;
        cursor: pointer;
        transition: background 100ms, color 100ms;

        &.active {
            background: var(--background-sidebar-items);
            color: var(--text-primary);
        }
    }
}

.CfSnippet {
    flex: 1;
    min-width: 0;
    padding: .4rem .55rem;
    border-radius: .45rem;
    border: 1px solid var(--color-border);
    background: var(--color-modal);
    color: var(--text-primary);
    font-family: 'Consolas', 'Courier New', monospace;
    font-size: .72rem;
    letter-spacing: .02em;

    &:focus {
        border-color: var(--color-border-modal);
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
    border: 1px solid var(--color-border);
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
</style>
