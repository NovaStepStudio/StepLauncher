import { ref } from 'vue';

export interface IdleOptions {
    autoCloseModals: boolean;
    idleMinutes: number;
    configCheckEnabled: boolean;
    configCheckMinutes: number;
}

export const idleOptions = ref<IdleOptions>({
    autoCloseModals: true,
    idleMinutes: 1,
    configCheckEnabled: true,
    configCheckMinutes: 3,
});

export const CLOSE_OVERLAYS_EVENT = 'sl:close-overlays';

const IDLE_EVENTS = ['mousemove', 'mousedown', 'keydown', 'wheel', 'touchstart', 'scroll'] as const;

let idleTimer: number | null = null;
let checkTimer: number | null = null;
let onIdle: (() => void) | null = null;
let onCheck: (() => void) | null = null;
let lastMark = 0;

function resetIdleTimer() {
    if (idleTimer !== null) {
        window.clearTimeout(idleTimer);
        idleTimer = null;
    }
    if (!idleOptions.value.autoCloseModals) return;
    const ms = Math.max(1, idleOptions.value.idleMinutes) * 60 * 1000;
    idleTimer = window.setTimeout(() => {
        idleTimer = null;
        onIdle?.();
    }, ms);
}

function onActivity() {
    const now = Date.now();
    if (now - lastMark < 3000) return;
    lastMark = now;
    resetIdleTimer();
}

function restartCheck() {
    if (checkTimer !== null) {
        window.clearInterval(checkTimer);
        checkTimer = null;
    }
    if (!idleOptions.value.configCheckEnabled) return;
    const ms = Math.max(1, idleOptions.value.configCheckMinutes) * 60 * 1000;
    checkTimer = window.setInterval(() => {
        onCheck?.();
    }, ms);
}

export function startIdleTracking(
    opts: IdleOptions,
    idleCb: () => void,
    checkCb: () => void
) {
    idleOptions.value = { ...idleOptions.value, ...opts };
    onIdle = idleCb;
    onCheck = checkCb;
    for (const ev of IDLE_EVENTS) {
        window.addEventListener(ev, onActivity, { passive: true });
    }
    lastMark = Date.now();
    resetIdleTimer();
    restartCheck();
}

export function stopIdleTracking() {
    for (const ev of IDLE_EVENTS) {
        window.removeEventListener(ev, onActivity);
    }
    if (idleTimer !== null) {
        window.clearTimeout(idleTimer);
        idleTimer = null;
    }
    if (checkTimer !== null) {
        window.clearInterval(checkTimer);
        checkTimer = null;
    }
    onIdle = null;
    onCheck = null;
}

export async function saveIdleOptions(opts: Partial<IdleOptions>) {
    idleOptions.value = { ...idleOptions.value, ...opts };
    resetIdleTimer();
    restartCheck();
    try {
        await (window as any).go?.main?.App?.SetIdle?.({
            autoCloseModals: idleOptions.value.autoCloseModals,
            idleMinutes: idleOptions.value.idleMinutes,
            configCheckEnabled: idleOptions.value.configCheckEnabled,
            configCheckMinutes: idleOptions.value.configCheckMinutes,
        });
    } catch { }
}
