<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import {
    IconSettings,
    IconDownload,
    IconDeviceGamepad,
    IconHome,
    IconFolder,
    IconChevronDown,
    IconCpu,
    IconPalette,
    IconInfoCircle,
} from '@tabler/icons-vue';
import SettingsModal, { type SectionConfig } from './Modals/SettingsModal.vue';
import GeneralSettings from './Layouts/Sections/Settings/GeneralSettings.vue';
import MinecraftSettings from './Layouts/Sections/Settings/MinecraftSettings.vue';
import PersonalizationSettings from './Layouts/Sections/Settings/PersonalizationSettings.vue';
import AboutSettings from './Layouts/Sections/Settings/AboutSettings.vue';
import { setUIScale, applyPersonalization, personalization, loadLocal, loadLocalFresh } from './stores/ui';

const showSettings = ref(false);

const settingsSections: SectionConfig[] = [
    { name: 'General', icon: IconHome, component: GeneralSettings },
    { name: 'Minecraft', icon: IconCpu, component: MinecraftSettings },
    { name: 'Personalización', icon: IconPalette, component: PersonalizationSettings },
    { name: 'Acerca de', icon: IconInfoCircle, component: AboutSettings },
];

const bg = computed(() => personalization.value?.background ?? null);
const bgImageUrl = ref('');
const bgVideoUrl = ref('');
const dynamicUrls = ref<string[]>([]);
const dynamicIndex = ref(0);
const videoReady = ref(false);
const videoRef = ref<HTMLVideoElement | null>(null);
const videoRetries = ref(0);
const MAX_VIDEO_RETRIES = 2;
let dynamicTimer: number | null = null;
let videoCheckTimer: number | null = null;

function stopVideoCheck() {
    if (videoCheckTimer !== null) {
        window.clearInterval(videoCheckTimer);
        videoCheckTimer = null;
    }
}

function startVideoCheck() {
    stopVideoCheck();
    if (bg.value?.type !== 'video') return;
    videoCheckTimer = window.setInterval(() => {
        const v = videoRef.value;
        if (v && v.readyState >= 2) {
            videoReady.value = true;
            stopVideoCheck();
        }
    }, 250);
}

function onVideoReady() {
    videoReady.value = true;
    stopVideoCheck();
}

async function onVideoError() {
    const b = bg.value;
    if (!b || b.type !== 'video' || !b.videoPath || videoRetries.value >= MAX_VIDEO_RETRIES) {
        onVideoReady();
        return;
    }
    videoRetries.value++;
    const fresh = await loadLocalFresh(b.videoPath);
    if (fresh) bgVideoUrl.value = fresh;
    if (videoRef.value) videoRef.value.load();
}

async function refreshBackground() {
    const b = bg.value;
    dynamicIndex.value = 0;
    videoReady.value = false;
    videoRetries.value = 0;
    bgImageUrl.value = b?.type === 'image' ? await loadLocal(b.imagePath ?? '') : '';
    bgVideoUrl.value = b?.type === 'video' ? await loadLocal(b.videoPath ?? '') : '';
    if (b?.type === 'dynamic' && Array.isArray(b.dynamicImages)) {
        dynamicUrls.value = await Promise.all(b.dynamicImages.map((rel: string) => loadLocal(rel)));
    } else {
        dynamicUrls.value = [];
    }
    startVideoCheck();
}

watch(bgVideoUrl, () => {
    videoReady.value = false;
    startVideoCheck();
});

const dynamicImage = computed(() => {
    if (!dynamicUrls.value.length) return '';
    return dynamicUrls.value[dynamicIndex.value % dynamicUrls.value.length];
});

function startDynamicTimer() {
    stopDynamicTimer();
    const b = bg.value;
    if (!b || b.type !== 'dynamic' || dynamicUrls.value.length < 2) return;
    const ms = Math.max(3, b.dynamicInterval) * 1000;
    dynamicTimer = window.setInterval(() => {
        if (b.dynamicOrder === 'random') {
            dynamicIndex.value = Math.floor(Math.random() * dynamicUrls.value.length);
        } else {
            dynamicIndex.value = (dynamicIndex.value + 1) % dynamicUrls.value.length;
        }
    }, ms);
}

function stopDynamicTimer() {
    if (dynamicTimer !== null) {
        window.clearInterval(dynamicTimer);
        dynamicTimer = null;
    }
}

watch(() => personalization.value?.background, () => {
    refreshBackground();
});

watch(dynamicUrls, (urls) => {
    if (urls.length) startDynamicTimer();
    else stopDynamicTimer();
});

onMounted(async () => {
    try {
        const cfg = await (window as any).go?.main?.App?.GetConfig?.();
        const raw = cfg?.personalization?.uiScale;
        if (typeof raw === 'number' && raw >= 50 && raw <= 200) {
            setUIScale(raw);
        }
        if (cfg?.personalization) {
            applyPersonalization(cfg.personalization);
        }
    } catch { /* */ }
    await refreshBackground();
    startDynamicTimer();
});

onUnmounted(() => {
    stopDynamicTimer();
    stopVideoCheck();
});
</script>

<template>
    <div class="BackgroundLayer" v-if="bg && bg.type !== 'none' && (bgImageUrl || bgVideoUrl || dynamicImage)">
        <Transition name="BgFade">
            <img v-if="bg.type === 'image' && bgImageUrl" :src="bgImageUrl" alt="">
            <video v-else-if="bg.type === 'video' && bgVideoUrl" ref="videoRef" :src="bgVideoUrl" autoplay muted loop playsinline preload="auto" @loadeddata="onVideoReady" @canplay="onVideoReady" @playing="onVideoReady" @error="onVideoError"></video>
            <img v-else-if="bg.type === 'dynamic' && dynamicImage" :key="dynamicIndex" :src="dynamicImage" alt="">
        </Transition>
        <div v-if="bg.type === 'video' && bgVideoUrl && !videoReady" class="BgLoading">
            <img src="../assets/gif/loading.gif" alt="">
        </div>
    </div>
    <main class="MainContent">
        <div class="Sidebar">
            <div class="Item" @click="showSettings = true">
                <IconSettings class="Item_Icon" stroke="2"/>
                <label class="Item_Label">Configuracion</label>
            </div>
            <div class="Item">
                <IconDownload class="Item_Icon" stroke="2"/>
                <label class="Item_Label">Descargar una version</label>
            </div>
        </div>

        <SettingsModal v-model:visible="showSettings" :sections="settingsSections" />
        <div class="Content">
            <div class="BottomControlVersion">
                <div class="VersionSelected">
                    <div class="ImageVersion">
                        <img src="../assets/not_found/not_found_version.png" loading="lazy" decoding="async" fetchpriority="high">
                    </div>
                    <div class="InfoVersion">
                        <p>Version Seleccionada :</p>
                        <h5>1.12.2</h5>
                    </div>
                </div>
                <div class="PlayButton">
                    <IconDeviceGamepad class="Icon" stroke="2"/>
                    <h1>Jugar</h1>
                    <div class="Decoration">
                        <img src="../assets/decorations/chicken.png" class="Chicken" loading="lazy" decoding="async" fetchpriority="high">
                        <img src="../assets/decorations/steve_and_alex.png" loading="lazy" decoding="async" fetchpriority="high">
                    </div>
                </div>
                <div class="Options">
                    <button>
                        <IconFolder stroke="2"/>
                        <label>Abrir Carpeta</label>
                    </button>
                </div>
            </div>
            <div class="TopOptions">
                <div class="UserCard">
                    <div class="Avatar">
                        <img src="../assets/not_found/avatar_not_found.png" loading="lazy" decoding="async" fetchpriority="low">
                    </div>
                    <div class="Username">
                        <h1>Usuario</h1>
                        <p>Sin Cuenta</p>
                    </div>
                    <button class="ExpandButtonProfiles">
                        <IconChevronDown stroke="2"/>
                    </button>
                </div>
            </div>
        </div>
    </main>
</template>

<style scoped lang="scss">
@use './Styles/App.scss';
</style>
