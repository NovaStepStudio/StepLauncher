/**
 * useBackground — extraído de App.vue para reducir el componente raíz.
 * Gestiona imagen / vídeo / dinámico + timers + estado de vídeo.
 */
import { ref, computed, watch } from 'vue';
import { personalization, loadLocal, loadLocalFresh } from '@/Common/Stores/Ui';

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

export function useBackground() {
    return {
        bg,
        bgImageUrl,
        bgVideoUrl,
        dynamicImage,
        dynamicIndex,
        dynamicUrls,
        videoReady,
        videoRef,
        refreshBackground,
        startDynamicTimer,
        stopDynamicTimer,
        stopVideoCheck,
        onVideoReady,
        onVideoError,
    };
}
