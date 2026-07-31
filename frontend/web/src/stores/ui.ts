import { ref, watch } from 'vue';

export const uiScale = ref(100);

export interface BackgroundConfig {
    type: 'none' | 'image' | 'video' | 'dynamic';
    imagePath: string;
    videoPath: string;
    dynamicImages: string[];
    dynamicOrder: 'sequential' | 'random';
    dynamicInterval: number;
}

export interface ThemeColors {
    sidebar: string;
    modal: string;
    buttons: string;
    borderModal: string;
    border: string;
}

export interface Personalization {
    uiScale: number;
    background: BackgroundConfig;
    fontPrimary: string;
    fontSecondary: string;
    colors: ThemeColors;
    recentColors: string[];
    animations: boolean;
    blur: boolean;
    shadows: boolean;
}

export const personalization = ref<Personalization | null>(null);

export function setUIScale(percent: number) {
    if (typeof percent !== 'number' || percent < 50 || percent > 200) {
        uiScale.value = 100;
        return;
    }
    uiScale.value = Math.round(percent);
}

function mimeOf(rel: string): string {
    const ext = rel.split('.').pop()?.toLowerCase() ?? '';
    switch (ext) {
        case 'mp4': return 'video/mp4';
        case 'webm': return 'video/webm';
        case 'gif': return 'image/gif';
        case 'png': return 'image/png';
        case 'jpg':
        case 'jpeg': return 'image/jpeg';
        case 'webp': return 'image/webp';
        case 'bmp': return 'image/bmp';
        default: return 'application/octet-stream';
    }
}

const localCache = new Map<string, string>();

// loadLocal lee un archivo del workdir via binding Go y devuelve una blob URL
// lista para usar en <img>/<video>. Funciona en el webview con wails dev o build.
// OJO: Wails v2 serializa []byte como base64 string en el binding, por eso hay
// que decodificarla con atob antes de crear el blob.
// loadLocalFresh re-lee el archivo ignorando la caché, creando una blob URL nueva.
// Se usa para reintentar la carga de videos de fondo cuando la anterior falla.
export async function loadLocalFresh(rel: string): Promise<string> {
    const key = String(rel ?? '').replace(/\\/g, '/');
    if (!key) return '';
    const old = localCache.get(key);
    localCache.delete(key);
    const url = await loadLocal(key);
    if (old && old !== url) {
        try { URL.revokeObjectURL(old); } catch { /* */ }
    }
    return url;
}

export async function loadLocal(rel: string): Promise<string> {
    const key = String(rel ?? '').replace(/\\/g, '/');
    if (!key) return '';
    const cached = localCache.get(key);
    if (cached) return cached;
    try {
        const res = await (window as any).go?.main?.App?.ReadLocalFile?.(key);
        if (!res) return '';
        let data: Uint8Array;
        if (typeof res === 'string') {
            const bin = atob(res);
            data = new Uint8Array(bin.length);
            for (let i = 0; i < bin.length; i++) data[i] = bin.charCodeAt(i);
        } else if (res instanceof Uint8Array) {
            data = res;
        } else if (Array.isArray(res)) {
            data = new Uint8Array(res);
        } else {
            return '';
        }
        const url = URL.createObjectURL(new Blob([new Uint8Array(data)], { type: mimeOf(key) }));
        localCache.set(key, url);
        return url;
    } catch {
        return '';
    }
}

watch(
    uiScale,
    (v) => {
        if (typeof document !== 'undefined') {
            document.documentElement.style.zoom = String(v / 100);
        }
    },
    { immediate: true }
);

function applyRootVar(name: string, value: string) {
    if (typeof document === 'undefined') return;
    if (value && value.trim()) {
        document.documentElement.style.setProperty(name, value.trim());
    }
}

export function applyPersonalization(p: Personalization | null) {
    if (!p || typeof document === 'undefined') {
        personalization.value = p;
        return;
    }
    personalization.value = p;

    const root = document.documentElement;
    const c = p.colors ?? {};

    applyRootVar('--color-sidebar', c.sidebar);
    applyRootVar('--color-modal', c.modal);
    applyRootVar('--color-button', c.buttons);
    applyRootVar('--color-border-modal', c.borderModal);
    applyRootVar('--color-border', c.border);

    applyRootVar('--background-sidebar', c.sidebar);
    applyRootVar('--background-sidebar-items', c.sidebar);
    applyRootVar('--background-bottom-control-version', c.sidebar);
    applyRootVar('--background-play-button', c.buttons);
    applyRootVar('--background-button-primary', c.buttons);
    applyRootVar('--background-modal-primray', c.modal);
    applyRootVar('--border-modal-style', `1px solid color-mix(in srgb, ${c.borderModal} 50%, gray 25%)`);
    applyRootVar('--border-style', `1px solid color-mix(in srgb, ${c.border} 50%, gray 20%)`);

    const fontPrimary = p.fontPrimary === 'system' ? 'system-ui' : `'${p.fontPrimary}'`;
    const fontSecondary = p.fontSecondary === 'system' ? 'system-ui' : `'${p.fontSecondary}'`;
    applyRootVar('--font-primary', fontPrimary);
    applyRootVar('--font-secundary', fontSecondary);

    root.dataset.anim = p.animations ? 'on' : 'off';
    root.dataset.blur = p.blur ? 'on' : 'off';
    root.dataset.shadows = p.shadows ? 'on' : 'off';
}
