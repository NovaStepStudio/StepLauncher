import { ref, watch } from 'vue';

export const uiScale = ref(100);

function applyUIScaleZoom(v: number) {
    if (typeof document !== 'undefined') {
        document.documentElement.style.zoom = String(v / 100);
    }
}

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
    progress: string;
    playButton: string;
    buttonPrimary: string;
    error: string;
    success: string;
    tag: string;
    warning: string;
}

export interface Personalization {
    uiScale: number;
    background: BackgroundConfig;
    fontPrimary: string;
    fontSecondary: string;
    fontPrimaryColor: string;
    fontSecondaryColor: string;
    fontPrimarySize: number;
    fontSecondarySize: number;
    colors: ThemeColors;
    recentColors: string[];
    animations: boolean;
    blur: boolean;
    shadows: boolean;
    textShadow: boolean;
    textShadowIntensity: number;
}

export const personalization = ref<Personalization | null>(null);

export function setUIScale(percent: number) {
    const v = Math.round(Number(percent));
    if (Number.isNaN(v)) {
        uiScale.value = 100;
        applyUIScaleZoom(100);
        return;
    }
    uiScale.value = Math.min(200, Math.max(50, v));
    applyUIScaleZoom(uiScale.value);
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

export async function loadLocalFresh(rel: string): Promise<string> {
    const key = String(rel ?? '').replace(/\\/g, '/');
    if (!key) return '';
    const old = localCache.get(key);
    localCache.delete(key);
    const url = await loadLocal(key);
    if (old && old !== url) {
        try { URL.revokeObjectURL(old); } catch { }
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
    (v) => applyUIScaleZoom(v),
    { immediate: true }
);

function applyRootVar(name: string, value: string) {
    if (typeof document === 'undefined') return;
    if (value && value.trim()) {
        document.documentElement.style.setProperty(name, value.trim());
    }
}

function normalizePersonalization(p: any, cur: any): Personalization {
    const num = (v: any, fb: number) => (typeof v === 'number' && Number.isFinite(v) ? v : fb);
    const bool = (v: any, fb: boolean) => (typeof v === 'boolean' ? v : fb);
    const str = (v: any, fb: string) => (typeof v === 'string' && v.trim() ? v : fb);
    const arr = (v: any, fb: string[]) => (Array.isArray(v) ? v : fb);
    const colorsIn = p?.colors ?? cur?.colors ?? {};
    const bgIn = p?.background ?? cur?.background ?? {};
    return {
        uiScale: num(p?.uiScale, num(cur?.uiScale, 100)),
        background: {
            type: ['none', 'image', 'video', 'dynamic'].includes(bgIn.type) ? bgIn.type : 'none',
            imagePath: str(bgIn.imagePath, ''),
            videoPath: str(bgIn.videoPath, ''),
            dynamicImages: arr(bgIn.dynamicImages, []),
            dynamicOrder: bgIn.dynamicOrder === 'random' ? 'random' : 'sequential',
            dynamicInterval: num(bgIn.dynamicInterval, 10),
        },
        fontPrimary: str(p?.fontPrimary, str(cur?.fontPrimary, 'Lexend')),
        fontSecondary: str(p?.fontSecondary, str(cur?.fontSecondary, 'Inter')),
        fontPrimaryColor: str(p?.fontPrimaryColor, str(cur?.fontPrimaryColor, '#ffffff')),
        fontSecondaryColor: str(p?.fontSecondaryColor, str(cur?.fontSecondaryColor, '#cfcfd6')),
        fontPrimarySize: num(p?.fontPrimarySize, num(cur?.fontPrimarySize, 1)),
        fontSecondarySize: num(p?.fontSecondarySize, num(cur?.fontSecondarySize, 1)),
        colors: {
            sidebar: str(colorsIn.sidebar, '#0005'),
            modal: str(colorsIn.modal, '#111'),
            buttons: str(colorsIn.buttons, '#111'),
            borderModal: str(colorsIn.borderModal, '#494949'),
            border: str(colorsIn.border, 'rgba(37, 37, 37, 0.3)'),
            progress: str(colorsIn.progress, '#5ed89a'),
            playButton: str(colorsIn.playButton, '#111'),
            buttonPrimary: str(colorsIn.buttonPrimary, '#111'),
            error: str(colorsIn.error, '#ff6b6b'),
            success: str(colorsIn.success, '#5ed89a'),
            tag: str(colorsIn.tag, '#a974ff'),
            warning: str(colorsIn.warning, '#ffb347'),
        },
        recentColors: arr(p?.recentColors, arr(cur?.recentColors, [])),
        animations: bool(p?.animations, bool(cur?.animations, true)),
        blur: bool(p?.blur, bool(cur?.blur, true)),
        shadows: bool(p?.shadows, bool(cur?.shadows, true)),
        textShadow: bool(p?.textShadow, bool(cur?.textShadow, false)),
        textShadowIntensity: num(p?.textShadowIntensity, num(cur?.textShadowIntensity, 1)),
    };
}

export function applyPersonalization(p: Personalization | null) {
    if (!p || typeof document === 'undefined') {
        personalization.value = p;
        return;
    }
    const normalized = normalizePersonalization(p, personalization.value ?? null);
    personalization.value = normalized;

    setUIScale(normalized.uiScale);

    const root = document.documentElement;
    const c = normalized.colors;

    applyRootVar('--background-sidebar', c.sidebar);
    applyRootVar('--background-sidebar-items', c.sidebar);
    applyRootVar('--background-bottom-control-version', c.sidebar);
    applyRootVar('--background-modal-primary', c.modal);
    applyRootVar('--border-modal-style', `1px solid ${c.borderModal}`);
    applyRootVar('--border-style', `1px solid ${c.border}`);
    applyRootVar('--progress-color', c.progress);
    applyRootVar('--background-play-button', c.playButton);
    applyRootVar('--background-button-primary', c.buttonPrimary);
    applyRootVar('--color-error', c.error);
    applyRootVar('--color-success', c.success);
    applyRootVar('--color-tag', c.tag);
    applyRootVar('--color-warning', c.warning);

    const fontPrimary = normalized.fontPrimary === 'system' ? 'system-ui' : `'${normalized.fontPrimary}'`;
    const fontSecondary = normalized.fontSecondary === 'system' ? 'system-ui' : `'${normalized.fontSecondary}'`;
    applyRootVar('--font-primary', fontPrimary);
    applyRootVar('--font-secundary', fontSecondary);

    applyRootVar('--text-primary', normalized.fontPrimaryColor);
    applyRootVar('--text-secondary', normalized.fontSecondaryColor);

    applyRootVar('--font-size-primary', String(normalized.fontPrimarySize));
    applyRootVar('--font-size-secundary', String(normalized.fontSecondarySize));

    root.dataset.anim = normalized.animations ? 'on' : 'off';
    root.dataset.blur = normalized.blur ? 'on' : 'off';
    root.dataset.shadows = normalized.shadows ? 'on' : 'off';

    root.dataset.textshadow = normalized.textShadow ? 'on' : 'off';
    applyRootVar('--text-shadow-intensity', String(normalized.textShadowIntensity));
}
