<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { applyPersonalization, loadLocal } from '../../../stores/ui';
import ColorField from './ColorField.vue';

const bgType = ref('none');
const bgImage = ref('');
const bgVideo = ref('');
const dynamicImages = ref<string[]>([]);
const dynamicOrder = ref('sequential');
const dynamicInterval = ref(10);

const dynamicUrls = ref<string[]>([]);

const fontPrimary = ref('Lexend');
const fontSecondary = ref('Inter');

const colorSidebar = ref('#0005');
const colorModal = ref('#111');
const colorButtons = ref('#111');
const colorBorderModal = ref('#494949');
const colorBorder = ref('rgba(37, 37, 37, 0.3)');
const recentColors = ref<string[]>([]);

const animations = ref(true);
const blur = ref(true);
const shadows = ref(true);

const errorMsg = ref('');

const fontOptions = ['Lexend', 'Inter', 'Fredoka', 'system'];
const maxDynamicImages = 10;

function buildPersonalization() {
    return {
        background: {
            type: bgType.value,
            imagePath: bgImage.value,
            videoPath: bgVideo.value,
            dynamicImages: dynamicImages.value,
            dynamicOrder: dynamicOrder.value,
            dynamicInterval: dynamicInterval.value,
        },
        fontPrimary: fontPrimary.value,
        fontSecondary: fontSecondary.value,
        colors: {
            sidebar: colorSidebar.value,
            modal: colorModal.value,
            buttons: colorButtons.value,
            borderModal: colorBorderModal.value,
            border: colorBorder.value,
        },
        recentColors: recentColors.value,
        animations: animations.value,
        blur: blur.value,
        shadows: shadows.value,
    };
}

function trackRecents() {
    const used = [colorSidebar.value, colorModal.value, colorButtons.value, colorBorderModal.value, colorBorder.value];
    const next = [...recentColors.value];
    for (const c of used) {
        if (!c) continue;
        const dup = next.find((h) => h.toLowerCase() === c.toLowerCase());
        if (!dup) next.push(c);
    }
    recentColors.value = next.slice(-12);
}

async function refreshPreviews() {
    if (bgType.value === 'dynamic') {
        dynamicUrls.value = await Promise.all(dynamicImages.value.map((rel) => loadLocal(rel)));
    } else {
        dynamicUrls.value = [];
    }
}

async function save() {
    trackRecents();
    const p = buildPersonalization();
    applyPersonalization(p as any);
    try {
        await (window as any).go?.main?.App?.UpdatePersonalization?.(p);
    } catch { /* */ }
    await refreshPreviews();
}

function localApply() {
    applyPersonalization(buildPersonalization() as any);
}

async function pickBackground(kind: 'image' | 'video'): Promise<string | null> {
    errorMsg.value = '';
    try {
        const rel = await (window as any).go?.main?.App?.PickBackgroundFile?.(kind);
        if (typeof rel === 'string') {
            if (rel) return rel;
            return null;
        }
    } catch (e: any) {
        errorMsg.value = e?.message ?? 'No se pudo importar el archivo.';
    }
    return null;
}

async function onImagePick() {
    const rel = await pickBackground('image');
    if (rel) {
        bgImage.value = rel;
        bgType.value = 'image';
        save();
    }
}

async function onVideoPick() {
    const rel = await pickBackground('video');
    if (rel) {
        bgVideo.value = rel;
        bgType.value = 'video';
        save();
    }
}

async function onDynamicPick() {
    if (dynamicImages.value.length >= maxDynamicImages) {
        errorMsg.value = `El fondo dinámico admite hasta ${maxDynamicImages} imágenes.`;
        return;
    }
    const rel = await pickBackground('image');
    if (rel) {
        dynamicImages.value.push(rel);
        save();
    }
}

function removeDynamic(i: number) {
    dynamicImages.value.splice(i, 1);
    if (dynamicImages.value.length === 0) bgType.value = 'none';
    save();
}

function clearImage() {
    bgImage.value = '';
    bgType.value = 'none';
    save();
}

function clearVideo() {
    bgVideo.value = '';
    bgType.value = 'none';
    save();
}

onMounted(async () => {
    try {
        const cfg = await (window as any).go?.main?.App?.GetConfig?.();
        const p = cfg?.personalization ?? {};
        const b = p.background ?? {};
        bgType.value = b.type ?? 'none';
        bgImage.value = b.imagePath ?? '';
        bgVideo.value = b.videoPath ?? '';
        dynamicImages.value = Array.isArray(b.dynamicImages) ? b.dynamicImages : [];
        dynamicOrder.value = b.dynamicOrder ?? 'sequential';
        dynamicInterval.value = b.dynamicInterval ?? 10;
        fontPrimary.value = p.fontPrimary ?? 'Lexend';
        fontSecondary.value = p.fontSecondary ?? 'Inter';
        const c = p.colors ?? {};
        colorSidebar.value = c.sidebar ?? '#0005';
        colorModal.value = c.modal ?? '#111';
        colorButtons.value = c.buttons ?? '#111';
        colorBorderModal.value = c.borderModal ?? '#494949';
        colorBorder.value = c.border ?? 'rgba(37, 37, 37, 0.3)';
        recentColors.value = Array.isArray(p.recentColors) ? p.recentColors : [];
        animations.value = p.animations ?? true;
        blur.value = p.blur ?? true;
        shadows.value = p.shadows ?? true;
        localApply();
        await refreshPreviews();
    } catch { /* */ }
});
</script>

<template>
<div class="Ss">

    <div class="SsGroup">
        <div class="SsGroupHead">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="M21 15l-5-5L5 21"/></svg>
            <span>Fondo</span>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Tipo de fondo</span>
                <span class="SsDesc">Imagen estática, video animado o rotación dinámica.</span>
            </div>
            <div class="SsCtrl">
                <select class="SsSel" v-model="bgType" @change="save">
                    <option value="none">Ninguno</option>
                    <option value="image">Imagen</option>
                    <option value="video">Video animado</option>
                    <option value="dynamic">Dinámico</option>
                </select>
            </div>
        </div>

        <template v-if="bgType === 'image'">
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Fondo del launcher</span>
                    <span class="SsDesc">Imagen de fondo de la interfaz principal.</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn" @click="onImagePick">
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                        {{ bgImage ? 'Cambiar imagen' : 'Elegir imagen' }}
                    </button>
                    <button v-if="bgImage" class="SsBtn SsBtnDanger" @click="clearImage">Quitar</button>
                </div>
            </div>
        </template>

        <template v-else-if="bgType === 'video'">
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Fondo animado</span>
                    <span class="SsDesc">Un video de fondo. Debe ser MP4, GIF o WEBM, pesar menos de 20MB y tener resolución menor a 1080p.</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn" @click="onVideoPick">
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                        {{ bgVideo ? 'Cambiar video' : 'Elegir video' }}
                    </button>
                    <button v-if="bgVideo" class="SsBtn SsBtnDanger" @click="clearVideo">Quitar</button>
                </div>
            </div>
        </template>

        <template v-else-if="bgType === 'dynamic'">
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Imágenes del fondo</span>
                    <span class="SsDesc">Agrega hasta {{ maxDynamicImages }} imágenes y se irán alternando.</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn" :disabled="dynamicImages.length >= maxDynamicImages" @click="onDynamicPick">
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                        Agregar imagen ({{ dynamicImages.length }}/{{ maxDynamicImages }})
                    </button>
                </div>
            </div>
            <div v-for="(img, i) in dynamicImages" :key="img" class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Imagen {{ i + 1 }}</span>
                    <span class="SsDesc">{{ img.split(/[\\/]/).pop() }}</span>
                </div>
                <div class="SsCtrl">
                    <div class="SsBgThumb">
                        <img :src="dynamicUrls[i] || ''" alt="">
                    </div>
                    <button class="SsBtn SsBtnDanger" @click="removeDynamic(i)">Quitar</button>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Orden</span>
                    <span class="SsDesc">Secuencial o aleatorio.</span>
                </div>
                <div class="SsCtrl">
                    <select class="SsSel" v-model="dynamicOrder" @change="save">
                        <option value="sequential">Secuencial</option>
                        <option value="random">Aleatorio</option>
                    </select>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Tiempo por imagen</span>
                    <span class="SsDesc">Segundos que se muestra cada fondo.</span>
                </div>
                <div class="SsCtrl">
                    <div class="SsStep">
                        <button class="SsStepBtn" :disabled="dynamicInterval <= 3" @click="dynamicInterval = Math.max(3, dynamicInterval - 5); save()">−</button>
                        <span class="SsStepVal">{{ dynamicInterval }}s</span>
                        <button class="SsStepBtn" :disabled="dynamicInterval >= 300" @click="dynamicInterval += 5; save()">+</button>
                    </div>
                </div>
            </div>
        </template>

        <div v-if="errorMsg" class="SsTip SsTipDanger">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
            <span>{{ errorMsg }}</span>
        </div>
    </div>

    <div class="SsGroup">
        <div class="SsGroupHead">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 7V4h16v3"/><path d="M9 20h6"/><path d="M12 4v16"/></svg>
            <span>Tipografías</span>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Tipografía principal</span>
                <span class="SsDesc">Títulos y elementos destacados.</span>
            </div>
            <div class="SsCtrl">
                <select class="SsSel" v-model="fontPrimary" @change="save">
                    <option v-for="f in fontOptions" :key="f" :value="f">{{ f === 'system' ? 'Del sistema' : f }}</option>
                </select>
            </div>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Tipografía secundaria</span>
                <span class="SsDesc">Textos y descripciones.</span>
            </div>
            <div class="SsCtrl">
                <select class="SsSel" v-model="fontSecondary" @change="save">
                    <option v-for="f in fontOptions" :key="f" :value="f">{{ f === 'system' ? 'Del sistema' : f }}</option>
                </select>
            </div>
        </div>
    </div>

    <div class="SsGroup">
        <div class="SsGroupHead">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="13.5" cy="6.5" r=".5"/><circle cx="17.5" cy="10.5" r=".5"/><circle cx="8.5" cy="7.5" r=".5"/><circle cx="6.5" cy="12.5" r=".5"/><path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.555-2.503 5.555-5.554C21.965 6.012 17.461 2 12 2z"/></svg>
            <span>Colores</span>
        </div>
        <div class="SsColorRow">
            <div class="SsInfo">
                <span class="SsLabel">Sidebar</span>
                <span class="SsDesc">Color de la barra lateral izquierda.</span>
            </div>
            <ColorField v-model="colorSidebar" :recents="recentColors" @update:model-value="save" />
        </div>
        <div class="SsColorRow">
            <div class="SsInfo">
                <span class="SsLabel">Modal</span>
                <span class="SsDesc">Fondo de las ventanas modales.</span>
            </div>
            <ColorField v-model="colorModal" :recents="recentColors" @update:model-value="save" />
        </div>
        <div class="SsColorRow">
            <div class="SsInfo">
                <span class="SsLabel">Botones</span>
                <span class="SsDesc">Color de los botones y el botón de jugar.</span>
            </div>
            <ColorField v-model="colorButtons" :recents="recentColors" @update:model-value="save" />
        </div>
        <div class="SsColorRow">
            <div class="SsInfo">
                <span class="SsLabel">Bordes de modales</span>
                <span class="SsDesc">Bordes de las ventanas modales.</span>
            </div>
            <ColorField v-model="colorBorderModal" :recents="recentColors" @update:model-value="save" />
        </div>
        <div class="SsColorRow">
            <div class="SsInfo">
                <span class="SsLabel">Bordes generales</span>
                <span class="SsDesc">Bordes de tarjetas y elementos en general.</span>
            </div>
            <ColorField v-model="colorBorder" :recents="recentColors" @update:model-value="save" />
        </div>
    </div>

</div>
</template>

<style scoped lang="scss">
@use '../../../Styles/Settings.scss';
</style>
