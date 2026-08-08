import { loadLocal } from './Ui';

export interface FontSlotData {
    type: string;
    name: string;
    path: string;
}

export interface LauncherAssets {
    fonts: FontSlotData[];
}

export const BUILTIN_FONTS = ['Lexend', 'Inter', 'Fredoka', 'system'];

export const FONT_EXTENSIONS = ['.ttf', '.otf', '.woff', '.woff2'];

export function fontByType(assets: LauncherAssets | null, type: string): FontSlotData | undefined {
    return assets?.fonts?.find((f) => f?.type === type);
}

export function fontByPath(assets: LauncherAssets | null, path: string): FontSlotData | undefined {
    return assets?.fonts?.find((f) => f?.path === path);
}

export function fontBaseName(name: string): string {
    const base = name.split(/[\\/]/).pop() ?? name;
    const i = base.lastIndexOf('.');
    return i > 0 ? base.slice(0, i) : base;
}

export function cleanFontName(name: string): string {
    const trimmed = String(name ?? '').trim();
    if (!trimmed) return '';
    let out = fontBaseName(trimmed);
    const suffix = out.match(/\s+\((\d+)\)$/);
    if (suffix) out = out.slice(0, suffix.index);
    const dash = out.match(/-(\d+)$/);
    if (dash) out = out.slice(0, dash.index);
    return out.trim();
}

export function isBuiltinFont(name: string): boolean {
    return BUILTIN_FONTS.includes(name);
}

const loadedFaces = new Map<string, FontFace>();

export async function ensureCustomFonts(assets: LauncherAssets | null): Promise<void> {
    if (!assets || typeof document === 'undefined') return;
    const wanted = new Set<string>();
    const slots = Array.isArray(assets.fonts) ? assets.fonts : [];
    for (const slot of slots) {
        if (!slot || !slot.path) continue;
        const displayName = (slot.name ?? '').trim() || cleanFontName(slot.path);
        if (!displayName) continue;
        const key = `${displayName}\u0000${slot.path}`;
        wanted.add(key);
        if (loadedFaces.has(key)) continue;
        const url = await loadLocal(slot.path);
        if (!url) continue;
        try {
            const face = new FontFace(displayName, `url(${url})`);
            await face.load();
            document.fonts.add(face);
            loadedFaces.set(key, face);
        } catch { }
    }
    for (const [key, face] of [...loadedFaces]) {
        if (wanted.has(key)) continue;
        try {
            document.fonts.delete(face);
        } catch { }
        loadedFaces.delete(key);
    }
}
