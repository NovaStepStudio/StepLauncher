import { ref } from 'vue';
import { Events } from '@wailsio/runtime';
import { CheckForUpdates as LegacyCheck, ApplyUpdate } from '@wailsjs/StepLauncher/internal/Services/System/systemservice';
// Bindings del updater Wails (generados vía wails3). Se importan como namespace para
// tolerar que aún no hayan sido regenerados (ts-ignore).
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
import * as WailsApp from '@wailsjs/StepLauncher/internal/Services/System/systemservice';

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

// Estado legacy (Engine.CheckForUpdates)
export const updateInfo = ref<UpdateInfo | null>(null);
export const checking = ref(false);
export const modalVisible = ref(false);

// Estado Wails updater headless (github)
export interface WailsProgress {
    written: number;
    total: number;
    rate: number;
    provider?: string;
}
export const wailsRelease = ref<any | null>(null);
export const wailsProgress = ref<WailsProgress | null>(null);
export const wailsState = ref<string>('idle');
export const wailsError = ref<string | null>(null);
export const wailsDownloading = ref(false);
export const wailsReady = ref(false);
export const wailsCurrentVersion = ref<string>('');

let bound = false;
let lastWasAuto = false;

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

function wailsFlagsVersion(): string {
    try {
        const flags = (window as unknown as { _wails?: { flags?: Record<string, unknown> } })._wails?.flags;
        const v = flags?.version;
        if (typeof v === 'string' && v.length > 0) return v;
    } catch (_e) {}
    return '';
}

function toUpdateInfoFromWails(rel: any): UpdateInfo {
    if (!rel) return emptyInfo();
    // Fuente única: wailsCurrentVersion (poblado vía GetLauncherVersionInfo / flags / meta)
    // Si aún vacío, dispara fetch async y deja que la UI se actualice sola.
    let cur = wailsCurrentVersion.value || wailsFlagsVersion();
    if (!cur) {
        const getter = (WailsApp as any).GetAppVersion as (() => Promise<string>) | undefined;
        if (typeof getter === 'function') {
            getter().then((v: string) => { if (v && !wailsCurrentVersion.value) wailsCurrentVersion.value = String(v); }).catch(() => {});
        }
        cur = '';
    }
    return {
        hasUpdate: true,
        latestVersion: String(rel.version ?? ''),
        currentVersion: cur,
        releaseUrl: rel.htmlUrl ?? rel.url ?? '',
        releaseName: String(rel.name ?? ''),
        releaseDate: rel.publishedAt ?? '',
        notes: String(rel.notes ?? ''),
        hasUpdater: true,
        updaterUrl: '',
        platform: rel.artifact?.platform ?? '',
        error: '',
    };
}

export function bindUpdateEvents() {
    if (bound) return;
    bound = true;
    // Intenta capturar versión actual desde flags y API lo antes posible (reconocimiento concreto)
    if (!wailsCurrentVersion.value) {
        const v = wailsFlagsVersion();
        if (v) wailsCurrentVersion.value = v;
        else {
            const getter = (WailsApp as any).GetLauncherVersionInfo as (() => Promise<any>) | undefined;
            const simpleGetter = (WailsApp as any).GetAppVersion as (() => Promise<string>) | undefined;
            if (typeof getter === 'function') {
                getter().then((info: any) => {
                    if (info?.currentVersion) wailsCurrentVersion.value = String(info.currentVersion);
                }).catch(() => {});
            } else if (typeof simpleGetter === 'function') {
                simpleGetter().then((v: string) => { if (v) wailsCurrentVersion.value = String(v); }).catch(() => {});
            } else if (typeof (WailsApp as any).WailsGetCurrentVersion === 'function') {
                (WailsApp as any).WailsGetCurrentVersion().then((v: string) => { if (v) wailsCurrentVersion.value = String(v); }).catch(() => {});
            }
        }
    }
    // --- Legacy engine event ---
    try {
        Events.On('update_check', ({ data: raw }: any) => {
            try {
                const str = typeof raw === 'string' ? raw : JSON.stringify(raw ?? '');
                const info = JSON.parse(str) as UpdateInfo;
                updateInfo.value = info;
                const isOfflineNow = typeof navigator !== 'undefined' && !navigator.onLine;
                const isConnError = !!info.error && (String(info.error).toLowerCase().includes('no se pudo conectar') || String(info.error).toLowerCase().includes('sin conexión'));
                if (isOfflineNow && isConnError && lastWasAuto) {
                    console.debug('[updater] update_check silenciado sin conexión');
                    return;
                }
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

    // --- Wails updater headless events (github) ---
    const wailsOn = (name: string, cb: (e: any) => void) => {
        try { Events.On(name, cb as any); } catch { /* ignore */ }
    };

    wailsOn('wails:updater:check-started', () => {
        if (!lastWasAuto) checking.value = true;
        else checking.value = true;
        wailsState.value = 'checking';
        wailsError.value = null;
        wailsProgress.value = null;
        wailsReady.value = false;
        wailsDownloading.value = false;
    });

    wailsOn('wails:updater:meta', ({ data: meta }: any) => {
        const m = meta?.data ?? meta;
        if (m?.currentVersion) {
            wailsCurrentVersion.value = String(m.currentVersion);
        }
        // fallback: si no vino meta, intenta leer flags o binding
        if (!wailsCurrentVersion.value) {
            const v = wailsFlagsVersion();
            if (v) wailsCurrentVersion.value = v;
            else {
                const getter = (WailsApp as any).WailsGetCurrentVersion as (() => Promise<string>) | undefined;
                if (typeof getter === 'function') {
                    getter().then((v) => { if (v) wailsCurrentVersion.value = String(v); }).catch(() => {});
                }
            }
        }
    });

    wailsOn('wails:updater:update-available', ({ data: rel }: any) => {
        const release = rel?.data ?? rel;
        wailsRelease.value = release;
        // Asegura que tenemos versión actual antes de mapear
        if (!wailsCurrentVersion.value) {
            const v = wailsFlagsVersion();
            if (v) wailsCurrentVersion.value = v;
        }
        updateInfo.value = toUpdateInfoFromWails(release);
        wailsState.value = 'available';
        checking.value = false;
        // Prompt requerido: "¿Quieres actualizar? Actualizar Ahora / Más tarde"
        // Se muestra incluso en auto-check si hay update.
        modalVisible.value = true;
    });

    wailsOn('wails:updater:no-update', () => {
        wailsState.value = 'up-to-date';
        checking.value = false;
        wailsRelease.value = null;
        // Asegura versión concreta para el mensaje "¡Estás al día! vX"
        if (!wailsCurrentVersion.value) {
            const v = wailsFlagsVersion();
            if (v) wailsCurrentVersion.value = v;
        }
        let cur = wailsCurrentVersion.value || wailsFlagsVersion() || '';
        // Intenta refrescar desde backend si sigue vacío (API centralizada)
        if (!cur) {
            const getter = (WailsApp as any).GetLauncherVersionInfo as (() => Promise<any>) | undefined;
            const simple = (WailsApp as any).GetAppVersion as (() => Promise<string>) | undefined;
            const p = typeof getter === 'function' ? getter().then((info: any) => info?.currentVersion) : (typeof simple === 'function' ? simple() : Promise.resolve(''));
            p.then((v: string) => {
                if (v) {
                    wailsCurrentVersion.value = String(v);
                    if (updateInfo.value && !updateInfo.value.currentVersion) updateInfo.value.currentVersion = wailsCurrentVersion.value;
                }
            }).catch(() => {});
            // Mientras resuelve, usa lo que haya (vacío) y la UI lo pintará al llegar
        }
        if (!lastWasAuto) {
            updateInfo.value = { ...emptyInfo(), currentVersion: cur };
            // también refleja en wailsCurrentVersion para Vue done
            if (!wailsCurrentVersion.value) wailsCurrentVersion.value = cur;
            modalVisible.value = true;
        }
    });

    wailsOn('wails:updater:download-started', ({ data: rel }: any) => {
        wailsState.value = 'downloading';
        wailsDownloading.value = true;
        wailsProgress.value = { written: 0, total: rel?.artifact?.size ?? 0, rate: 0 };
        modalVisible.value = true;
    });

    wailsOn('wails:updater:download-progress', ({ data: p }: any) => {
        const prog = p?.data ?? p;
        wailsProgress.value = prog;
        wailsState.value = 'downloading';
        wailsDownloading.value = true;
    });

    wailsOn('wails:updater:download-complete', () => {
        wailsState.value = 'verifying';
        wailsDownloading.value = false;
    });

    wailsOn('wails:updater:verifying', () => {
        wailsState.value = 'verifying';
    });

    wailsOn('wails:updater:installing', () => {
        wailsState.value = 'installing';
    });

    wailsOn('wails:updater:update-ready', ({ data: rel }: any) => {
        const release = rel?.data ?? rel;
        if (release) wailsRelease.value = release;
        wailsState.value = 'ready';
        wailsReady.value = true;
        wailsDownloading.value = false;
        wailsProgress.value = null;
        modalVisible.value = true;
    });

    wailsOn('wails:updater:error', ({ data: info }: any) => {
        // Si es por falta de internet, silenciar en auto-check
        const isOfflineNow = typeof navigator !== 'undefined' && !navigator.onLine;
        const err = info?.data ?? info;
        let msg = err?.message ?? err?.Message ?? 'Error desconocido del updater';
        const stage = err?.stage ?? err?.Stage ?? '';
        const raw = String(msg).toLowerCase();
        const isNoHost = raw.includes('no such host') || raw.includes('dial tcp') || raw.includes('lookup') || raw.includes('no internet') || raw.includes('network is unreachable');
        if (isOfflineNow || isNoHost) {
            console.debug('[updater] error silenciado sin conexión:', msg);
            wailsState.value = 'idle';
            checking.value = false;
            wailsDownloading.value = false;
            if (!lastWasAuto) {
                wailsError.value = 'Sin conexión — No se pudo comprobar las actualizaciones sin internet.';
                updateInfo.value = { ...emptyInfo(), error: wailsError.value as string };
                modalVisible.value = true;
            }
            return;
        }
        // Mensaje amigable para rate limit de GitHub (403 anónimo)
        if (raw.includes('rate limit') || raw.includes('403') || raw.includes('api rate limit exceeded')) {
            msg = 'GitHub está limitando las solicitudes (rate limit). Intenta de nuevo en unos minutos o configura un token GH_TOKEN para elevar el límite.';
        }
        wailsError.value = stage ? `${stage}: ${msg}` : String(msg);
        wailsState.value = 'error';
        checking.value = false;
        wailsDownloading.value = false;
        // También refleja en updateInfo para compatibilidad con modal existente
        updateInfo.value = { ...emptyInfo(), error: wailsError.value as string };
        // En auto-check, el rate limit es ruidoso: solo muestra si es manual o si no es rate limit.
        // Para rate limit en auto, no abrir modal automáticamente (solo log).
        const isRateLimit = String(wailsError.value).toLowerCase().includes('rate limit');
        if (lastWasAuto && isRateLimit) {
            console.warn('[updater] rate limit en auto-check, silenciado:', wailsError.value);
            // no abrir modal en auto si es rate limit
        } else {
            modalVisible.value = true;
        }
    });
}

export async function checkForUpdates(auto = false) {
    // Sin internet no buscar actualizaciones
    if (typeof navigator !== 'undefined' && !navigator.onLine) {
        console.debug('[updater] omitido sin conexión');
        checking.value = false;
        wailsState.value = 'idle';
        if (!auto) {
            updateInfo.value = { ...emptyInfo(), error: 'Sin conexión — No se pudo comprobar las actualizaciones sin internet.' };
            modalVisible.value = true;
        }
        return;
    }
    lastWasAuto = auto;
    // reset estados wails pero preserva modalVisible si es manual
    wailsError.value = null;
    wailsReady.value = false;
    wailsProgress.value = null;
    wailsState.value = 'checking';
    // Si es manual, muestra modal con spinner inmediatamente
    updateInfo.value = null;
    wailsRelease.value = null;
    checking.value = true;
    if (!auto) modalVisible.value = true;

    // Intenta Wails updater (GitHub directo, headless) primero
    const wailsCheck = (WailsApp as any).WailsCheckForUpdate as (() => Promise<void>) | undefined;
    if (typeof wailsCheck === 'function') {
        try {
            await wailsCheck();
            // el resultado llegará vía eventos wails:updater:*
        } catch (e: any) {
            // fallback a legacy si Wails falla (p.ej. no configurado)
            console.warn('[updater] WailsCheckForUpdate fallo, fallback legacy', e);
        }
    }

    // Siempre también el check legacy: es el que resuelve el instalador
    // en Windows y el enlace de instalación manual en Linux/macOS.
    try {
        await LegacyCheck();
    } catch {
        checking.value = false;
        updateInfo.value = { ...emptyInfo(), error: 'No se pudo comprobar las actualizaciones.' };
        if (!auto) modalVisible.value = true;
    }
}

export async function installUpdate() {
    // El flujo legacy es el que aplica la actualización de verdad:
    // en Windows descarga el *-installer.exe, lo ejecuta y cierra el
    // launcher; en Linux/macOS abre la release para instalación manual.
    // (El flujo Wails nunca debe instalar: en Windows pondría el
    // instalador como binario con el swap en caliente.)
    const info = updateInfo.value;
    if (info?.hasUpdate) {
        try {
            await ApplyUpdate();
        } catch (_e) {}
        return;
    }
    // Fallback Wails headless solo si el legacy no resolvió actualización
    // (p.ej. falló su check pero Wails sí detectó release).
    if (wailsReady.value) {
        restartUpdate();
        return;
    }
    const wailsInstall = (WailsApp as any).WailsInstallUpdate as (() => Promise<void>) | undefined;
    const hasWailsRelease = !!wailsRelease.value || wailsState.value === 'available';
    if (hasWailsRelease && typeof wailsInstall === 'function') {
        try {
            wailsState.value = 'downloading';
            wailsDownloading.value = true;
            await wailsInstall();
            return;
        } catch (e) {
            console.error('[updater] WailsInstallUpdate error', e);
            wailsError.value = String((e as any)?.message ?? e);
            wailsState.value = 'error';
            return;
        }
    }
    // Fallback legacy
    try {
        await ApplyUpdate();
    } catch (_e) {}
}

export async function restartUpdate() {
    const wailsRestart = (WailsApp as any).WailsRestart as (() => Promise<void>) | undefined;
    if (typeof wailsRestart === 'function') {
        try {
            await wailsRestart();
        } catch (e) {
            console.error('[updater] WailsRestart error', e);
        }
    }
}

export function skipUpdateVersion() {
    const skip = (WailsApp as any).WailsSkipVersion as ((v: string) => Promise<void>) | undefined;
    if (typeof skip === 'function' && wailsRelease.value?.version) {
        try { skip(wailsRelease.value.version); } catch (_e) {}
    }
    closeUpdateModal();
}

export function closeUpdateModal() {
    modalVisible.value = false;
    // No resetea wailsRelease para permitir reabrir, pero limpia progress si estaba ready/error
    if (wailsState.value === 'error' || wailsState.value === 'up-to-date') {
        wailsState.value = 'idle';
        wailsError.value = null;
    }
}
