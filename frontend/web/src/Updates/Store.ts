import { ref } from 'vue';
import { EventsOn } from '@wailsjs/runtime/runtime';

export interface UpdateInfo {
    hasUpdate: boolean;
    latestVersion: string;
    currentVersion: string;
    releaseUrl: string;
    releaseName: string;
    releaseDate: string;
    notes: string;
    hasUpdater: boolean;
    updaterUrl: string;
    platform: string;
    error: string;
}

export const updateInfo = ref<UpdateInfo | null>(null);
export const checking = ref(false);
export const modalVisible = ref(false);

let bound = false;
let lastWasAuto = false;

export function bindUpdateEvents() {
    if (bound) return;
    bound = true;
    try {
        EventsOn('update_check', (raw: unknown) => {
            try {
                const str = typeof raw === 'string' ? raw : JSON.stringify(raw ?? '');
                const info = JSON.parse(str) as UpdateInfo;
                updateInfo.value = info;
                if (!lastWasAuto || info.hasUpdate) {
                    modalVisible.value = true;
                }
            } catch {
                updateInfo.value = { ...emptyInfo(), error: 'Respuesta inválida del servidor.' };
                if (!lastWasAuto) modalVisible.value = true;
            } finally {
                checking.value = false;
            }
        });
    } catch {
        checking.value = false;
    }
}

function emptyInfo(): UpdateInfo {
    return {
        hasUpdate: false,
        latestVersion: '',
        currentVersion: '',
        releaseUrl: '',
        releaseName: '',
        releaseDate: '',
        notes: '',
        hasUpdater: false,
        updaterUrl: '',
        platform: '',
        error: '',
    };
}

export async function checkForUpdates(auto = false) {
    lastWasAuto = auto;
    updateInfo.value = null;
    checking.value = true;
    if (!auto) modalVisible.value = true;
    try {
        await (window as any).go?.main?.App?.CheckForUpdates?.();
    } catch {
        checking.value = false;
        updateInfo.value = { ...emptyInfo(), error: 'No se pudo comprobar las actualizaciones.' };
        if (!auto) modalVisible.value = true;
    }
}

export async function installUpdate() {
    try {
        await (window as any).go?.main?.App?.ApplyUpdate?.();
    } catch { }
}

export function closeUpdateModal() {
    modalVisible.value = false;
}
