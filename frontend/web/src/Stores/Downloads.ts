import { ref, computed } from 'vue';

export interface ActiveDownload {
    id: string;
    version: string;
    loader: string;
    state: string;
    percent: number;
    mb: number;
    speedMbps: number;
}

export const download = ref<ActiveDownload | null>(null);

const ACTIVE_STATES = ['pending', 'downloading', 'paused', 'verifying', 'redownloading'];

export const isDownloading = computed(() => {
    const d = download.value;
    return !!d && ACTIVE_STATES.includes(d.state);
});

export function setDownload(d: ActiveDownload) {
    download.value = d;
}

export function clearDownload() {
    download.value = null;
}