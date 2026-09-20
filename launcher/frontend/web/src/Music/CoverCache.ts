import { ref } from 'vue';

// Cache de carátulas en disco: cache/covers con TTL 24h gestionado en Go.
// Este store expone helpers para el frontend y un pequeño cache en memoria de Blob URLs.

export interface CoverCacheInfo {
    total: number;
    valid: number;
    expired: number;
    size: number;
}

const memoryUrls = new Map<string, string>();
const pending = new Set<string>();

function dataUriToBlobUrl(dataUri: string): string {
    if (!dataUri.startsWith('data:')) return dataUri;
    try {
        const comma = dataUri.indexOf(',');
        if (comma < 0) return dataUri;
        const header = dataUri.slice(5, comma);
        const b64 = dataUri.slice(comma + 1);
        const mime = header.split(';')[0] || 'image/jpeg';
        const bin = atob(b64);
        const bytes = new Uint8Array(bin.length);
        for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
        const blob = new Blob([bytes as BlobPart], { type: mime });
        const url = URL.createObjectURL(blob);
        return url;
    } catch {
        return '';
    }
}

function blobToBase64(blob: Blob): Promise<{ b64: string; mime: string }> {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => {
            const res = String(reader.result || '');
            const comma = res.indexOf(',');
            if (comma < 0) return reject(new Error('data uri inválido'));
            const header = res.slice(5, comma);
            const mime = header.split(';')[0] || blob.type || 'image/jpeg';
            const b64 = res.slice(comma + 1);
            resolve({ b64, mime });
        };
        reader.onerror = () => reject(reader.error);
        reader.readAsDataURL(blob);
    });
}

// Obtiene la carátula cacheada como data URI si existe y no expiró (24h). Retorna '' si no hay cache válido.
export async function getCachedCoverDataUri(trackPath: string): Promise<string> {
    const key = String(trackPath || '').trim();
    if (!key) return '';
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.GetCachedCover !== 'function') return '';
        const uri: string = await mod.GetCachedCover(key);
        if (typeof uri === 'string' && uri.startsWith('data:')) return uri;
        return '';
    } catch {
        return '';
    }
}

// Intenta obtener Blob URL desde cache en disco. Si está en memoria, lo reutiliza.
export async function getCachedCoverUrl(trackPath: string): Promise<string> {
    const key = String(trackPath || '').trim();
    if (!key) return '';
    const mem = memoryUrls.get(key);
    if (mem) return mem;
    const dataUri = await getCachedCoverDataUri(key);
    if (!dataUri) return '';
    const blobUrl = dataUriToBlobUrl(dataUri);
    if (blobUrl) {
        if (memoryUrls.size >= 32) {
            const first = memoryUrls.keys().next().value as string | undefined;
            if (first) {
                const old = memoryUrls.get(first);
                if (old) URL.revokeObjectURL(old);
                memoryUrls.delete(first);
            }
        }
        memoryUrls.set(key, blobUrl);
    }
    return blobUrl;
}

// Guarda una carátula en disco: coverUrl es Blob URL, object URL o data URI. Extrae bytes y persiste.
export async function saveCachedCover(trackPath: string, coverUrl: string): Promise<void> {
    const key = String(trackPath || '').trim();
    if (!key || !coverUrl) return;
    if (pending.has(key)) return;
    pending.add(key);
    try {
        let b64 = '';
        let mime = 'image/jpeg';
        if (coverUrl.startsWith('data:')) {
            const comma = coverUrl.indexOf(',');
            const header = coverUrl.slice(5, comma);
            mime = header.split(';')[0] || 'image/jpeg';
            b64 = coverUrl.slice(comma + 1);
        } else if (coverUrl.startsWith('blob:')) {
            try {
                const res = await fetch(coverUrl);
                const blob = await res.blob();
                const conv = await blobToBase64(blob);
                b64 = conv.b64;
                mime = conv.mime || mime;
            } catch {
                return;
            }
        } else {
            // URL remota o file path no soportado directamente
            return;
        }
        if (!b64) return;
        try {
            const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
            if (typeof mod.SaveCachedCover === 'function') {
                await mod.SaveCachedCover(key, b64, mime).catch(() => {});
            }
        } catch (_e) {}
        // También poblar memoria
        if (!memoryUrls.has(key) && coverUrl.startsWith('data:')) {
            const blobUrl = dataUriToBlobUrl(coverUrl);
            if (blobUrl) memoryUrls.set(key, blobUrl);
        }
    } finally {
        pending.delete(key);
    }
}

export async function refreshCachedCover(trackPath: string): Promise<void> {
    const key = String(trackPath || '').trim();
    if (!key) return;
    // Limpiar memoria
    const mem = memoryUrls.get(key);
    if (mem) {
        URL.revokeObjectURL(mem);
        memoryUrls.delete(key);
    }
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.RefreshCachedCover === 'function') {
            await mod.RefreshCachedCover(key).catch(() => {});
        }
    } catch (_e) {}
}

export async function getCoverCacheInfo(): Promise<CoverCacheInfo> {
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.GetCoverCacheInfo === 'function') {
            const info = await mod.GetCoverCacheInfo();
            return info as CoverCacheInfo;
        }
        if (typeof mod.GetCoverCacheStatus === 'function') {
            const info = await mod.GetCoverCacheStatus();
            return info as CoverCacheInfo;
        }
    } catch (_e) {}
    return { total: 0, valid: 0, expired: 0, size: 0 };
}

export async function clearCoverCache(): Promise<number> {
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.ClearCoverCache === 'function') {
            const n = await mod.ClearCoverCache();
            return Number(n) || 0;
        }
    } catch (_e) {}
    // Limpiar memoria
    for (const url of memoryUrls.values()) URL.revokeObjectURL(url);
    memoryUrls.clear();
    return 0;
}

// Guarda una imagen arbitraria (por ejemplo extraída con music-metadata) directamente desde bytes base64
export async function saveCoverBytes(trackPath: string, bytes: Uint8Array, mime: string): Promise<void> {
    if (!bytes || bytes.length === 0) return;
    let b64 = '';
    try {
        let bin = '';
        const chunk = 0x8000;
        for (let i = 0; i < bytes.length; i += chunk) bin += String.fromCharCode(...bytes.subarray(i, i + chunk));
        b64 = btoa(bin);
    } catch {
        return;
    }
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.SaveCachedCover === 'function') {
            await mod.SaveCachedCover(trackPath, b64, mime || 'image/jpeg').catch(() => {});
        }
    } catch (_e) {}
}

export async function saveMusicMetadata(trackPath: string, title: string, artist: string, duration: number, coverData: string, mime: string): Promise<void> {
    const key = String(trackPath || '').trim();
    if (!key) return;
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.SaveMusicMetadata === 'function') {
            await mod.SaveMusicMetadata(key, title, artist, duration, coverData || '', mime || 'image/jpeg').catch(() => {});
        } else if (coverData) {
            // Fallback a cover cache
            await mod.SaveCachedCover?.(key, coverData, mime || 'image/jpeg').catch(() => {});
        }
    } catch (_e) {}
}

export async function getAllMusicMetadata(): Promise<Record<string, any>> {
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.GetAllMusicMetadata === 'function') {
            const data = await mod.GetAllMusicMetadata();
            if (data && typeof data === 'object') return data as Record<string, any>;
        }
    } catch (_e) {}
    return {};
}

export async function getMusicMetadata(trackPath: string): Promise<any | null> {
    const key = String(trackPath || '').trim();
    if (!key) return null;
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.GetMusicMetadata === 'function') {
            const entry = await mod.GetMusicMetadata(key).catch(() => null);
            if (entry) return entry;
        }
    } catch (_e) {}
    return null;
}

export const coverCacheReady = ref(false);
