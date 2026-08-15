<script setup lang="ts">
import { ref, onMounted, onActivated } from 'vue';
import { applyPersonalization, loadLocal, personalization } from '@/Common/Stores/Ui';
import { ensureCustomFonts, isBuiltinFont, cleanFontName, fontByType, type LauncherAssets } from '@/Common/Stores/Fonts';
import ColorField from '@/Common/Components/ColorField.vue';
import { openDialog, PERSONALIZATION_PREVIEW_EVENT } from '@/Common/Overlays/Store';

const bgType = ref('none');
const bgImage = ref('');
const bgVideo = ref('');
const dynamicImages = ref<string[]>([]);
const dynamicOrder = ref('sequential');
const dynamicInterval = ref(10);

const dynamicUrls = ref<string[]>([]);

const fontPrimary = ref('Lexend');
const fontSecondary = ref('Inter');
const fontPrimaryColor = ref('#ffffff');
const fontSecondaryColor = ref('#cfcfd6');
const fontPrimarySize = ref(1);
const fontSecondarySize = ref(1);

const colorSidebar = ref('#0005');
const colorModal = ref('#111');
const colorBorderModal = ref('#494949');
const colorBorder = ref('rgba(37, 37, 37, 0.3)');
const colorProgress = ref('#5ed89a');
const colorError = ref('#ff6b6b');
const colorSuccess = ref('#5ed89a');
const colorTag = ref('#a974ff');
const colorWarning = ref('#ffb347');
const colorPlayButton = ref('#111');
const colorButtonPrimary = ref('#111');
const recentColors = ref<string[]>([]);

const errorMsg = ref('');

const fontOptions = ref(['Lexend', 'Inter', 'Fredoka', 'system']);
const fontFiles = ref<string[]>([]);
const assetsRef = ref<LauncherAssets>({ fonts: [] });
const maxDynamicImages = 10;

function openFontManager() {
    openDialog('font-manager', { assets: assetsRef.value }, { changed: onFontsChanged });
}

async function loadFontData() {
    let assets: any = null;
    let files: string[] = [];
    try {
        assets = await (window as any).go?.main?.App?.GetLauncherAssets?.();
    } catch { }
    try {
        const f = await (window as any).go?.main?.App?.ListFontFiles?.();
        files = Array.isArray(f) ? f : [];
    } catch { }
    assetsRef.value = { fonts: Array.isArray(assets?.fonts) ? assets.fonts : [] };
    fontFiles.value = files;
    const custom: string[] = [];
    const seen = new Set<string>();
    const push = (n: string) => {
        const t = (n ?? '').trim();
        if (t && !isBuiltinFont(t) && !seen.has(t)) {
            seen.add(t);
            custom.push(t);
        }
    };
    for (const e of assetsRef.value.fonts) {
        if (!e) continue;
        push((e.name ?? '').trim() || (e.path ? cleanFontName(e.path) : ''));
    }
    for (const f of fontFiles.value) {
        const ref = 'launcher/fonts/' + f;
        if (assetsRef.value.fonts.some((e) => e?.path === ref)) continue;
        push(cleanFontName(f));
    }
    fontOptions.value = ['Lexend', 'Inter', 'Fredoka', ...custom, 'system'];
}

function slotFontName(a: LauncherAssets, slot: string): string {
    const e = fontByType(a, slot);
    if (!e) return '';
    return (e.name ?? '').trim() || (e.path ? cleanFontName(e.path) : '');
}

async function syncFontSelects() {
    const a = assetsRef.value;
    const pName = slotFontName(a, 'primary');
    const sName = slotFontName(a, 'secundary');
    const opts = fontOptions.value;
    let changed = false;
    const realign = (cur: string, registered: string, fallback: string) => {
        if (isBuiltinFont(cur)) return null;
        if (registered && registered !== cur) return registered;
        if (!opts.includes(cur)) return registered || fallback;
        return null;
    };
    const pNext = realign(fontPrimary.value, pName, 'Lexend');
    if (pNext !== null && pNext !== fontPrimary.value) {
        fontPrimary.value = pNext;
        changed = true;
    }
    const sNext = realign(fontSecondary.value, sName, 'Inter');
    if (sNext !== null && sNext !== fontSecondary.value) {
        fontSecondary.value = sNext;
        changed = true;
    }
    await ensureCustomFonts(assetsRef.value);
    if (changed) {
        save();
    } else {
        localApply();
    }
}

async function saveAssets() {
    try {
        await (window as any).go?.main?.App?.SaveLauncherAssets?.(assetsRef.value);
        await ensureCustomFonts(assetsRef.value);
    } catch { }
}

async function onFontChange(slot: 'primary' | 'secundary') {
    const val = slot === 'primary' ? fontPrimary.value : fontSecondary.value;
    const list = (assetsRef.value.fonts ?? []).map((e) => ({ ...e }));
    let assetsChanged = false;
    if (isBuiltinFont(val)) {
        const cur = fontByType(assetsRef.value, slot);
        if (cur && (cur.name || cur.path)) {
            for (const e of list) {
                if (e.type === slot && e.path) e.type = '';
            }
            assetsChanged = true;
        }
} else if (val) {
        let target = list.find((e) => e.name === val && e.path)
            ?? list.find((e) => e.path && cleanFontName(e.path) === val);
        let created = false;
        if (!target) {
            const file = fontFiles.value.find((f) => cleanFontName(f) === val);
            if (file) {
                target = { type: slot, name: val, path: 'launcher/fonts/' + file };
                list.push(target);
                created = true;
            }
        }
        const current = list.find((e) => e.type === slot && e.path);
        if (target?.path && (created || current?.path !== target.path)) {
            for (const e of list) {
                if (e.type === slot && e.path) e.type = '';
            }
            target.type = slot;
            assetsChanged = true;
        }
    }
    if (assetsChanged) {
        assetsRef.value = { fonts: list };
        await saveAssets();
    }
    save();
}

async function onFontsChanged() {
    await loadFontData();
    await syncFontSelects();
}

function buildPersonalization() {
    const base = (personalization.value ?? {}) as any;
    return {
        ...base,
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
        fontPrimaryColor: fontPrimaryColor.value,
        fontSecondaryColor: fontSecondaryColor.value,
        fontPrimarySize: fontPrimarySize.value,
        fontSecondarySize: fontSecondarySize.value,
        colors: {
            sidebar: colorSidebar.value,
            modal: colorModal.value,
            borderModal: colorBorderModal.value,
            border: colorBorder.value,
            progress: colorProgress.value,
            playButton: colorPlayButton.value,
            buttonPrimary: colorButtonPrimary.value,
            error: colorError.value,
            success: colorSuccess.value,
            tag: colorTag.value,
            warning: colorWarning.value,
        },
        recentColors: recentColors.value,
    };
}

function trackRecents() {
    const used = [colorSidebar.value, colorModal.value, colorBorderModal.value, colorBorder.value, colorProgress.value, colorPlayButton.value, colorButtonPrimary.value, colorError.value, colorSuccess.value, colorTag.value, colorWarning.value];
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
    } catch { }
    await refreshPreviews();
}

function localApply() {
    applyPersonalization(buildPersonalization() as any);
}

function onPreviewColor(field: 'sidebar' | 'border', v: string) {
    const p = buildPersonalization();
    const colors = { ...p.colors };
    colors[field] = v;
    applyPersonalization({ ...p, colors } as any);
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

function openPreview() {
    window.dispatchEvent(new CustomEvent(PERSONALIZATION_PREVIEW_EVENT));
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
        fontPrimaryColor.value = p.fontPrimaryColor ?? '#ffffff';
        fontSecondaryColor.value = p.fontSecondaryColor ?? '#cfcfd6';
        fontPrimarySize.value = p.fontPrimarySize ?? 1;
        fontSecondarySize.value = p.fontSecondarySize ?? 1;
        const c = p.colors ?? {};
        colorSidebar.value = c.sidebar ?? '#0005';
        colorModal.value = c.modal ?? '#111';
        colorBorderModal.value = c.borderModal ?? '#494949';
        colorBorder.value = c.border ?? 'rgba(37, 37, 37, 0.3)';
        colorProgress.value = c.progress ?? '#5ed89a';
        colorError.value = c.error ?? '#ff6b6b';
        colorSuccess.value = c.success ?? '#5ed89a';
        colorTag.value = c.tag ?? '#a974ff';
        colorWarning.value = c.warning ?? '#ffb347';
        colorPlayButton.value = c.playButton ?? '#111';
        colorButtonPrimary.value = c.buttonPrimary ?? '#111';
        recentColors.value = Array.isArray(p.recentColors) ? p.recentColors : [];
        localApply();
    } catch { }
    try { await refreshPreviews(); } catch { }
    try {
        await loadFontData();
        await syncFontSelects();
    } catch { }
});

let firstActivate = true;
onActivated(async () => {
    if (firstActivate) {
        firstActivate = false;
        return;
    }
    try {
        await loadFontData();
        await syncFontSelects();
    } catch { }
});

const round1 = (n: number) => Math.round(n * 10) / 10;
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
                <select class="SsSel" v-model="fontPrimary" @change="onFontChange('primary')">
                    <option v-for="f in fontOptions" :key="f" :value="f">{{ f === 'system' ? 'Del sistema' : cleanFontName(f) }}</option>
                </select>
            </div>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Tipografía secundaria</span>
                <span class="SsDesc">Textos y descripciones.</span>
            </div>
            <div class="SsCtrl">
                <select class="SsSel" v-model="fontSecondary" @change="onFontChange('secundary')">
                    <option v-for="f in fontOptions" :key="f" :value="f">{{ f === 'system' ? 'Del sistema' : cleanFontName(f) }}</option>
                </select>
            </div>
        </div>
        <div class="SsColorRow">
            <div class="SsInfo">
                <span class="SsLabel">Color de letra principal</span>
                <span class="SsDesc">Color de los títulos y elementos destacados.</span>
            </div>
            <ColorField v-model="fontPrimaryColor" :recents="recentColors" @update:model-value="save" />
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Tamaño de letra principal</span>
                <span class="SsDesc">Porcentaje respecto al tamaño original.</span>
            </div>
            <div class="SsCtrl">
                <div class="SsStep">
                    <button class="SsStepBtn" :disabled="fontPrimarySize <= 0.5" @click="fontPrimarySize = round1(fontPrimarySize - 0.1); save()">−</button>
                    <span class="SsStepVal">{{ Math.round(fontPrimarySize * 100) }}%</span>
                    <button class="SsStepBtn" :disabled="fontPrimarySize >= 2" @click="fontPrimarySize = round1(fontPrimarySize + 0.1); save()">+</button>
                </div>
            </div>
        </div>
        <div class="SsColorRow">
            <div class="SsInfo">
                <span class="SsLabel">Color de letra secundaria</span>
                <span class="SsDesc">Color de los textos y descripciones.</span>
            </div>
            <ColorField v-model="fontSecondaryColor" :recents="recentColors" @update:model-value="save" />
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Tamaño de letra secundaria</span>
                <span class="SsDesc">Porcentaje respecto al tamaño original.</span>
            </div>
            <div class="SsCtrl">
                <div class="SsStep">
                    <button class="SsStepBtn" :disabled="fontSecondarySize <= 0.5" @click="fontSecondarySize = round1(fontSecondarySize - 0.1); save()">−</button>
                    <span class="SsStepVal">{{ Math.round(fontSecondarySize * 100) }}%</span>
                    <button class="SsStepBtn" :disabled="fontSecondarySize >= 2" @click="fontSecondarySize = round1(fontSecondarySize + 0.1); save()">+</button>
                </div>
            </div>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Fuentes personalizadas</span>
                <span class="SsDesc">Importa tus propias tipografías para personalizar el launcher.</span>
            </div>
            <div class="SsCtrl">
                <button class="SsBtn SsBtnPrimary" @click="openFontManager">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 7V4h16v3"/><path d="M9 20h6"/><path d="M12 4v16"/></svg>
                    Gestionar tipografías
                </button>
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
                <span class="SsLabel">Barra lateral</span>
                <span class="SsDesc">Color de la barra lateral izquierda.</span>
            </div>
            <ColorField v-model="colorSidebar" :recents="recentColors" preview @update:model-value="save" @preview="(v: string) => onPreviewColor('sidebar', v)" />
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
            <ColorField v-model="colorBorder" :recents="recentColors" preview @update:model-value="save" @preview="(v: string) => onPreviewColor('border', v)" />
        </div>
        <div class="SsColorRow">
            <div class="SsInfo">
                <span class="SsLabel">Progreso de descarga</span>
                <span class="SsDesc">Color de la barra y del círculo de progreso al descargar.</span>
            </div>
            <ColorField v-model="colorProgress" :recents="recentColors" @update:model-value="save" />
        </div>
        <div class="SsColorRow">
            <div class="SsInfo">
                <span class="SsLabel">Error</span>
                <span class="SsDesc">Errores, acciones destructivas y botones de peligro.</span>
            </div>
            <ColorField v-model="colorError" :recents="recentColors" @update:model-value="save" />
        </div>
        <div class="SsColorRow">
            <div class="SsInfo">
                <span class="SsLabel">Éxito</span>
                <span class="SsDesc">Confirmaciones, listo y estados positivos.</span>
            </div>
            <ColorField v-model="colorSuccess" :recents="recentColors" @update:model-value="save" />
        </div>
        <div class="SsColorRow">
            <div class="SsInfo">
                <span class="SsLabel">Etiquetas</span>
                <span class="SsDesc">Etiquetas y badges informativos.</span>
            </div>
            <ColorField v-model="colorTag" :recents="recentColors" @update:model-value="save" />
        </div>
        <div class="SsColorRow">
            <div class="SsInfo">
                <span class="SsLabel">Aviso</span>
                <span class="SsDesc">Advertencias y avisos.</span>
            </div>
            <ColorField v-model="colorWarning" :recents="recentColors" @update:model-value="save" />
        </div>
        <div class="SsColorRow">
            <div class="SsInfo">
                <span class="SsLabel">Botón de jugar</span>
                <span class="SsDesc">Color del botón principal de jugar en la pantalla de inicio.</span>
            </div>
            <ColorField v-model="colorPlayButton" :recents="recentColors" @update:model-value="save" />
        </div>
        <div class="SsColorRow">
            <div class="SsInfo">
                <span class="SsLabel">Botones principales</span>
                <span class="SsDesc">Color de los botones primarios de la interfaz.</span>
            </div>
            <ColorField v-model="colorButtonPrimary" :recents="recentColors" @update:model-value="save" />
        </div>
        <div class="SsRow PsColorHead">
            <div class="SsInfo">
                <span class="SsLabel">Vista previa</span>
                <span class="SsDesc">Así se verán los colores semánticos en la interfaz.</span>
            </div>
        </div>
        <div class="PsExamples">
            <div class="PsExItem"><span class="PsExCheck">✓</span> Versión seleccionada</div>
            <span class="PsExBadge">Fabric</span>
            <span class="PsExText is-good">Descarga completada ✓</span>
            <span class="PsExText is-warn">⚠ Espacio insuficiente</span>
            <button type="button" class="PsExBtn" tabindex="-1">Eliminar</button>
            <span class="PsExText is-bad">✕ Error al conectar</span>
        </div>
        <div class="PsPreviewRow">
            <div class="SsInfo">
                <span class="SsLabel">Vista previa completa</span>
                <span class="SsDesc">Abre el launcher tal y como se verá con tus colores actuales.</span>
            </div>
            <div class="SsCtrl">
                <button class="SsBtn SsBtnPrimary" @click="openPreview">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                    Abrir vista previa
                </button>
            </div>
        </div>
    </div>

</div>
</template>

<style scoped lang="scss">
@use '../Styles/Personalization.scss';
</style>
