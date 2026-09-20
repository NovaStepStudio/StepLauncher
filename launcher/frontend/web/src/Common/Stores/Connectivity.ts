import { ref, computed } from 'vue';

// Estado reactivo de conectividad basado en navigator.onLine
export const isOnline = ref(typeof navigator !== 'undefined' ? navigator.onLine : true);
export const isOffline = computed(() => !isOnline.value);

let initialized = false;

export const CONNECTIVITY_ONLINE_EVENT = 'sl:connectivity-online';
export const CONNECTIVITY_OFFLINE_EVENT = 'sl:connectivity-offline';

function updateOnlineStatus(): void {
    const wasOffline = !isOnline.value;
    isOnline.value = navigator.onLine;
    if (typeof window !== 'undefined') {
        if (isOnline.value && wasOffline) {
            window.dispatchEvent(new CustomEvent(CONNECTIVITY_ONLINE_EVENT));
        } else if (!isOnline.value && !wasOffline) {
            window.dispatchEvent(new CustomEvent(CONNECTIVITY_OFFLINE_EVENT));
        }
    }
}

export function initConnectivity(): void {
    if (initialized || typeof window === 'undefined') return;
    initialized = true;
    isOnline.value = navigator.onLine;
    window.addEventListener('online', updateOnlineStatus);
    window.addEventListener('offline', updateOnlineStatus);
}

export function stopConnectivity(): void {
    if (!initialized || typeof window === 'undefined') return;
    window.removeEventListener('online', updateOnlineStatus);
    window.removeEventListener('offline', updateOnlineStatus);
    initialized = false;
}
