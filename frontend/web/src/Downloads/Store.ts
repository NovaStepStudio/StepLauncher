import { ref, computed } from 'vue';
import { EventsOn } from '@wailsjs/runtime/runtime';

export interface ActiveDownload {
    id: string;
    label: string;
    version: string;
    kind: 'version' | 'instance';
    state: string;
    percent: number;
    mbDownloaded: number;
    mbTotal: number;
    filesDownloaded: number;
    filesTotal: number;
    speedMbps: number;
    error?: string;
}

// Lista central de descargas activas del launcher (versiones e instancias).
// Alimentada por los eventos download_* del backend; los modales registran su
// descarga con registerDownload y el widget la muestra sin importar el
// origen. Los IDs del backend son únicos por manager (ver-/inst-), por lo que
// las descargas simultáneas no se pisan.
export const downloads = ref<Record<string, ActiveDownload>>({});

const ACTIVE_STATES = ['pending', 'downloading', 'paused', 'verifying', 'redownloading'];
const TERMINAL_STATES = ['completed', 'cancelled', 'error'];

export const activeDownloads = computed(() =>
    Object.values(downloads.value)
        .filter((d) => ACTIVE_STATES.includes(d.state))
        .sort((a, b) => a.id.localeCompare(b.id))
);

export const anyActive = computed(() => activeDownloads.value.length > 0);

export const latest = computed(() => activeDownloads.value[0] ?? null);

export function downloadOf(id: string): ActiveDownload | null {
    return downloads.value[id] ?? null;
}

export function registerDownload(d: ActiveDownload): void {
    ensureDownloadEvents();
    const prev = downloads.value[d.id];
    downloads.value = { ...downloads.value, [d.id]: { ...prev, ...d } };
}

export function updateDownload(id: string, patch: Partial<ActiveDownload>): void {
    const prev = downloads.value[id];
    if (!prev) return;
    downloads.value = { ...downloads.value, [id]: { ...prev, ...patch } };
}

export function clearDownload(id: string): void {
    if (!downloads.value[id]) return;
    const next = { ...downloads.value };
    delete next[id];
    downloads.value = next;
}

function parsePayload(raw: unknown): any {
    if (raw && typeof raw === 'object') return raw;
    try {
        return JSON.parse(String(raw ?? ''));
    } catch {
        return {};
    }
}

let eventsOff: (() => void)[] | null = null;

// Registro único de los eventos de descarga: aquí se alimenta la lista central.
// Los modales mantienen sus propios listeners para el progreso detallado, por
// lo que no hay doble estado; este store solo cubre widget y panel.
function ensureDownloadEvents(): void {
    if (eventsOff) return;
    eventsOff = [
        EventsOn('download_progress', (raw: any) => {
            const p = parsePayload(raw);
            const id = p?.id;
            const data = p?.data;
            if (!id || !data || typeof data !== 'object') return;
            const prev = downloads.value[id];
            if (!prev) return;
            updateDownload(id, {
                state: data.state ?? prev.state,
                percent: Number(data.percent ?? 0),
                mbDownloaded: Number(data.mbDownloaded ?? 0),
                mbTotal: Number(data.mbTotal ?? 0),
                filesDownloaded: Number(data.filesDownloaded ?? 0),
                filesTotal: Number(data.filesTotal ?? 0),
                speedMbps: Number(data.speedMbps ?? 0),
            });
        }),
        EventsOn('download_state', (raw: any) => {
            const p = parsePayload(raw);
            const id = p?.id;
            if (!id) return;
            const state = p?.data?.state ?? '';
            const prev = downloads.value[id];
            if (!prev) return;
            updateDownload(id, { state });
            if (TERMINAL_STATES.includes(state)) {
                const delay = state === 'error' ? 8000 : 800;
                window.setTimeout(() => {
                    if (downloads.value[id]?.state === state) clearDownload(id);
                }, delay);
            }
        }),
        EventsOn('download_error', (raw: any) => {
            const p = parsePayload(raw);
            const id = p?.id;
            if (!id) return;
            const prev = downloads.value[id];
            if (!prev) return;
            updateDownload(id, { state: 'error', error: p?.data?.error ?? 'Error de descarga' });
            window.setTimeout(() => clearDownload(id), 12000);
        }),
    ];
}