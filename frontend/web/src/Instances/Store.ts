import { ref, computed, watch } from 'vue';
import { EventsOn } from '@wailsjs/runtime/runtime';
import { hideOnLaunchIfEnabled } from '@/Launcher/Store';
import {
    downloads as dlStore,
    registerDownload,
    clearDownload as clearCentralDownload,
    activeDownloads as dlActiveDownloads,
} from '@/Downloads/Store';
import {
    ListInstances,
    GetInstance,
    CreateInstance as CreateInstanceBinding,
    UpdateInstanceMetadata,
    DeleteInstance,
    UpdateInstanceConfig,
    AddInstanceVersion,
    CancelInstanceDownload,
    VerifyInstance,
    LaunchInstance,
    CloneInstance,
    InstallInstanceModLoader,
    GetInstalledInstanceModLoader,
    RemoveInstanceModLoaderState,
    GetInstanceStats as GetInstanceStatsBinding,
    OpenInstanceFolder as OpenInstanceFolderBinding,
} from '@wailsjs/go/main/App';
import type { instance } from '@wailsjs/go/models';

export interface InstanceInfo {
    name: string;
    title: string;
    versions: string[];
    favorite: boolean;
    pinned: boolean;
    group: string;
    lastPlayed: string;
    playTime: number;
}

export interface InstanceMetadata {
    id: string;
    name: string;
    title: string;
    description: string;
    icon: string;
    banner: string;
    group: string;
    tags: string[];
    favorite: boolean;
    pinned: boolean;
    createdAt: string;
    lastPlayed: string;
    playTime: number;
    versions: string[];
    configPath: string;
}

export interface InstanceLaunchConfig {
    version?: string;
    javaExec?: string;
    minRam?: number;
    maxRam?: number;
    useOfficialJava?: boolean;
    fullscreen?: boolean;
    hardwareAcceleration?: boolean;
    gcPreset?: string;
    gpuPreference?: string;
    customResolution?: boolean;
    resWidth?: number;
    resHeight?: number;
}

export interface CreateInstanceReq {
    name: string;
    version?: string;
    title?: string;
    description?: string;
    icon?: string;
    banner?: string;
    group?: string;
    tags?: string[];
    favorite?: boolean;
    pinned?: boolean;
}

export interface UpdateMetadataReq {
    title?: string;
    description?: string;
    icon?: string;
    banner?: string;
    group?: string;
    tags?: string[];
    favorite?: boolean;
    pinned?: boolean;
}

export interface VersionStat {
    version: string;
    playCount: number;
    totalPlayed: number;
    firstPlayed: number;
    lastPlayed: number;
}

export interface InstanceStats {
    totalPlayTime: number;
    totalSessions: number;
    firstPlayed: number;
    lastPlayed: number;
    weeklyPlayTime: number;
    weeklySessions: number;
    weeklyVersions: string[];
    versions: VersionStat[];
    running: boolean;
}

export interface InstanceDetails {
    meta: InstanceMetadata;
    config: InstanceLaunchConfig;
}

export interface InstanceDownloadState {
    dlId: string;
    version: string;
    state: string;
    percent: number;
    mbDownloaded: number;
    mbTotal: number;
    filesDownloaded: number;
    filesTotal: number;
    error?: string;
}

export interface VerifyResult {
    valid: boolean;
    version: string;
    issues: Array<{ type: string; file: string; message: string }>;
}

export interface InstalledLoaderInfo {
    loaderType: string;
    loaderVersion: string;
    minecraftVersion: string;
    versionJsonId?: string;
    installerJarPath?: string;
    installedAt?: number;
}

// Estado en vivo de la instalación de un modloader sobre una instancia
// (alimentado por los eventos modloader_* del backend). El modal gestiona su
// propia vista; este estado permite que la tarjeta y el banner del detalle
// sigan mostrando la instalación aunque el modal se cierre.
export interface InstanceLoaderDl {
    loader: string;
    loaderVersion: string;
    mcVersion: string;
    phase: 'resolving' | 'downloading' | 'installing' | 'done' | 'error';
    message: string;
    progress: number;
    total: number;
}

// Elemento unificado para widget y panel: combina las descargas centrales
// (download_*) con las instalaciones de modloaders (modloader_*), que no pasan
// por el store de descargas. El widget así permanece visible durante todo el
// proceso de instalación de un modloader, no solo durante la descarga de la
// versión.
export interface MergedActiveDownload {
    id: string;
    label: string;
    version: string;
    kind: 'version' | 'instance' | 'loader';
    state: string;
    percent: number;
    mbDownloaded: number;
    mbTotal: number;
    filesDownloaded: number;
    filesTotal: number;
    speedMbps: number;
    error?: string;
    phase?: string;
    message?: string;
    loader?: string;
    cancellable: boolean;
}

const ACTIVE_LOADER_PHASES = ['resolving', 'downloading', 'installing'];

export const allActiveDownloads = computed<MergedActiveDownload[]>(() => {
    const list: MergedActiveDownload[] = dlActiveDownloads.value.map((d) => ({
        ...d,
        cancellable: true,
    }));
    for (const [inst, ld] of Object.entries(loaderDls.value)) {
        if (!ACTIVE_LOADER_PHASES.includes(ld.phase)) continue;
        const total = ld.total > 0 ? ld.total : 0;
        const percent = total > 0 ? Math.min(100, (ld.progress / total) * 100) : 0;
        list.push({
            id: `loader-${inst}`,
            label: inst,
            version: ld.loaderVersion ? `${ld.loaderVersion} · MC ${ld.mcVersion}` : `MC ${ld.mcVersion}`,
            kind: 'loader',
            state: ld.phase,
            percent,
            mbDownloaded: 0,
            mbTotal: 0,
            filesDownloaded: ld.progress,
            filesTotal: total,
            speedMbps: 0,
            phase: ld.phase,
            message: ld.message,
            loader: ld.loader,
            cancellable: false,
        });
    }
    return list.sort((a, b) => a.id.localeCompare(b.id));
});

export const anyAllActive = computed(() => allActiveDownloads.value.length > 0);

export const instances = ref<InstanceInfo[]>([]);
export const details = ref<Record<string, InstanceDetails | null>>({});
export const loaders = ref<Record<string, InstalledLoaderInfo | null>>({});
export const downloads = ref<Record<string, InstanceDownloadState>>({});
export const loaderDls = ref<Record<string, InstanceLoaderDl>>({});
export const launching = ref<Record<string, boolean>>({});
export const loadingList = ref(false);

const dlToInstance = new Map<string, string>();
const loaderSessionToInstance = new Map<string, string>();

const ACTIVE_DL = ['pending', 'downloading', 'paused', 'verifying', 'redownloading'];
const TERMINAL_DL = ['completed', 'cancelled', 'error'];

// El map por instancia (para tarjetas, banners y modales) se deriva de la lista
// central de descargas: el registro de eventos download_* vive solo en
// Downloads/Store.ts y aquí se proyecta por instancia. Al pasar a un estado
// terminal se refresca la lista de instancias.
watch(
    dlStore,
    (all) => {
        const next: Record<string, InstanceDownloadState> = {};
        let transitioned = false;
        for (const [id, d] of Object.entries(all)) {
            const inst = dlToInstance.get(id);
            if (!inst) continue;
            const prev = downloads.value[inst];
            next[inst] = {
                dlId: id,
                version: d.version,
                state: d.state,
                percent: d.percent,
                mbDownloaded: d.mbDownloaded,
                mbTotal: d.mbTotal,
                filesDownloaded: d.filesDownloaded,
                filesTotal: d.filesTotal,
                error: d.error,
            };
            if (prev && ACTIVE_DL.includes(prev.state) && TERMINAL_DL.includes(d.state)) {
                transitioned = true;
            }
        }
        downloads.value = next;
        if (transitioned) void loadInstances();
    },
    { deep: true }
);

export const sortedInstances = computed(() => {
    return [...instances.value].sort((a, b) => {
        if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
        if (a.favorite !== b.favorite) return a.favorite ? -1 : 1;
        return (b.lastPlayed || '').localeCompare(a.lastPlayed || '');
    });
});

function parsePayload(raw: unknown): any {
    if (raw && typeof raw === 'object') return raw;
    try {
        return JSON.parse(String(raw ?? ''));
    } catch {
        return {};
    }
}

let eventsOff: (() => void)[] | null = null;

function ensureEvents() {
    if (eventsOff) return;
    eventsOff = [
        // Instalación de modloader sobre una instancia: el estado vive en el
        // store (loaderDls) para que banner y tarjeta lo muestren aunque el
        // modal de instalación esté cerrado. El modal tiene su propia vista.
        // Las descargas download_* se gestionan en Downloads/Store.ts (lista
        // central) y se proyectan aquí por instancia con el watch de dlStore.
        EventsOn('modloader_resolving', (raw: any) => updateModLoaderEvent(raw)),
        EventsOn('modloader_downloading', (raw: any) => updateModLoaderEvent(raw)),
        EventsOn('modloader_installing', (raw: any) => updateModLoaderEvent(raw)),
        EventsOn('modloader_installed', (raw: any) => updateModLoaderEvent(raw)),
        EventsOn('modloader_error', (raw: any) => updateModLoaderEvent(raw)),
    ];
}

function updateModLoaderEvent(raw: unknown): void {
    const e = parsePayload(raw) as {
        type?: string;
        sessionId?: string;
        loader?: string;
        version?: string;
        mcVersion?: string;
        message?: string;
        error?: string;
        progress?: number;
        total?: number;
    };
    const inst = e?.sessionId ? loaderSessionToInstance.get(e.sessionId) : undefined;
    if (!inst || !e?.type) return;
    const prev = loaderDls.value[inst];
    const next: InstanceLoaderDl = {
        loader: e.loader ?? prev?.loader ?? 'modloader',
        loaderVersion: e.version ?? prev?.loaderVersion ?? '',
        mcVersion: e.mcVersion ?? prev?.mcVersion ?? '',
        phase: prev?.phase ?? 'resolving',
        message: e.message ?? prev?.message ?? '',
        progress: Number(e.progress ?? 0),
        total: Number(e.total ?? 0),
    };
    switch (e.type) {
        case 'modloader_resolving':
            next.phase = 'resolving';
            break;
        case 'modloader_downloading':
            next.phase = 'downloading';
            next.progress = Number(e.progress ?? prev?.progress ?? 0);
            next.total = Number(e.total ?? prev?.total ?? 0);
            break;
        case 'modloader_installing':
            next.phase = 'installing';
            break;
        case 'modloader_installed':
            next.phase = 'done';
            void loadInstalledLoader(inst);
            window.setTimeout(() => clearLoaderDl(inst), 8000);
            break;
        case 'modloader_error':
            next.phase = 'error';
            next.message = e.error ?? e.message ?? 'Error desconocido';
            window.setTimeout(() => clearLoaderDl(inst), 12000);
            break;
    }
    loaderDls.value = { ...loaderDls.value, [inst]: next };
}

// Registra la sesión de instalación de un modloader para poder mapear los
// eventos modloader_* a la instancia. Lo llaman los modales tras recibir el
// sessionId del binding.
export function registerLoaderSession(
    name: string,
    sessionId: string,
    loader: string,
    loaderVersion: string,
    mcVersion: string
): void {
    ensureEvents();
    loaderSessionToInstance.set(sessionId, name);
    loaderDls.value = {
        ...loaderDls.value,
        [name]: {
            loader,
            loaderVersion,
            mcVersion,
            phase: 'resolving',
            message: 'Preparando…',
            progress: 0,
            total: 0,
        },
    };
}

export function clearLoaderDl(instName: string): void {
    if (!loaderDls.value[instName]) return;
    const next = { ...loaderDls.value };
    delete next[instName];
    loaderDls.value = next;
}

export function loaderDlOf(name: string): InstanceLoaderDl | null {
    return loaderDls.value[name] ?? null;
}

// La instancia está ocupada (no se puede jugar/descargar/editar) mientras
// tiene una descarga activa, una instalación de modloader en curso o un
// lanzamiento en marcha.
export function isInstanceBusy(name: string): boolean {
    if (launching.value[name]) return true;
    const ld = loaderDls.value[name];
    if (ld && (ld.phase === 'resolving' || ld.phase === 'downloading' || ld.phase === 'installing')) {
        return true;
    }
    const dl = downloads.value[name];
    if (!dl) return false;
    const ACTIVE = ['pending', 'downloading', 'paused', 'verifying', 'redownloading'];
    return ACTIVE.includes(dl.state);
}

export function loaderDlStateText(ld: InstanceLoaderDl): string {
    switch (ld.phase) {
        case 'resolving':
            return `Resolviendo ${ld.loader}…`;
        case 'downloading':
            return ld.message || `Descargando ${ld.loader}…`;
        case 'installing':
            return ld.message || `Instalando ${ld.loader}…`;
        case 'done':
            return `${ld.loader} instalado`;
        case 'error':
            return `Error: ${ld.message || 'no se pudo instalar'}`;
        default:
            return 'Instalando…';
    }
}

export function clearDownload(instName: string): void {
    const dl = downloads.value[instName];
    if (dl?.dlId) clearCentralDownload(dl.dlId);
    const next = { ...downloads.value };
    delete next[instName];
    downloads.value = next;
}

export function dlStateText(state: string): string {
    switch (state) {
        case 'downloading':
            return 'Descargando';
        case 'paused':
            return 'En pausa';
        case 'verifying':
            return 'Verificando archivos';
        case 'redownloading':
            return 'Re-descargando archivos';
        case 'pending':
            return 'Iniciando descarga…';
        case 'completed':
            return 'Descarga completada';
        case 'cancelled':
            return 'Cancelado';
        case 'error':
            return 'Error';
        default:
            return 'Preparando…';
    }
}

export async function loadInstances(): Promise<void> {
    loadingList.value = true;
    try {
        const list = await ListInstances();
        if (Array.isArray(list)) instances.value = list;
    } catch { }
    loadingList.value = false;
}

export async function loadDetails(name: string): Promise<InstanceDetails | null> {
    try {
        const res = await GetInstance(name);
        if (!res?.metadata) return null;
        const meta = res.metadata as InstanceMetadata;
        const cfg = (res.config ?? {}) as InstanceLaunchConfig;
        if (!Array.isArray(meta.versions)) meta.versions = [];
        details.value = { ...details.value, [name]: { meta, config: cfg } };
        return details.value[name] ?? null;
    } catch {
        return null;
    } finally {
        void loadInstalledLoader(name);
    }
}

export function detailOf(name: string): InstanceDetails | null {
    return details.value[name] ?? null;
}

// ---------- Modloader instalado ----------

export const LOADER_LABELS: Record<string, string> = {
    vanilla: 'Vanilla',
    fabric: 'Fabric',
    forge: 'Forge',
    neoforge: 'NeoForge',
    quilt: 'Quilt',
    legacyfabric: 'Legacy Fabric',
};

export function loaderOf(name: string): InstalledLoaderInfo | null {
    return loaders.value[name] ?? null;
}

export function loaderLabel(l: InstalledLoaderInfo | null): string {
    if (!l) return 'Vanilla';
    const key = (l.loaderType ?? '').toLowerCase();
    return LOADER_LABELS[key] ?? l.loaderType ?? 'ModLoader';
}

// Carga el estado del modloader instalado en la instancia (loader-state.json).
// Separado de loadDetails para que un fallo aquí no rompa la carga del detalle:
// si no hay modloader instalado queda registrado como null (Vanilla).
export async function loadInstalledLoader(name: string): Promise<InstalledLoaderInfo | null> {
    try {
        const info = await getInstalledInstanceModLoader(name);
        loaders.value = { ...loaders.value, [name]: info };
        return info;
    } catch {
        return null;
    }
}

export async function loadAllDetails(): Promise<void> {
    const names = [...instances.value.map((i) => i.name), ...Object.keys(details.value)];
    await Promise.allSettled([...new Set(names)].map((n) => loadDetails(n)));
}

export async function createInstance(req: CreateInstanceReq): Promise<string> {
    try {
        const res = await CreateInstanceBinding(req as instance.CreateInstanceReq);
        await loadInstances();
        if (res?.metadata?.name) await loadDetails(res.metadata.name);
        if (res?.downloadId && res?.metadata?.name) {
            ensureEvents();
            dlToInstance.set(res.downloadId, res.metadata.name);
            registerDownload({
                id: res.downloadId,
                label: res.metadata.name,
                version: req.version ?? '',
                kind: 'instance',
                state: 'pending',
                percent: 0,
                mbDownloaded: 0,
                mbTotal: 0,
                filesDownloaded: 0,
                filesTotal: 0,
                speedMbps: 0,
            });
            downloads.value = {
                ...downloads.value,
                [res.metadata.name]: {
                    dlId: res.downloadId,
                    version: req.version ?? '',
                    state: 'pending',
                    percent: 0,
                    mbDownloaded: 0,
                    mbTotal: 0,
                    filesDownloaded: 0,
                    filesTotal: 0,
                },
            };
        }
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo crear la instancia.';
    }
}

export async function updateMetadata(name: string, req: UpdateMetadataReq): Promise<string> {
    try {
        await UpdateInstanceMetadata(name, req);
        await loadInstances();
        await loadDetails(name);
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo actualizar la instancia.';
    }
}

export async function toggleFavorite(name: string, favorite: boolean): Promise<string> {
    return updateMetadata(name, { favorite });
}

export async function togglePin(name: string, pinned: boolean): Promise<string> {
    return updateMetadata(name, { pinned });
}

export async function getInstanceStats(name: string): Promise<InstanceStats | null> {
    try {
        const res = await GetInstanceStatsBinding(name);
        if (!res) return null;
        const stats = res as unknown as InstanceStats;
        if (!Array.isArray(stats.versions)) stats.versions = [];
        if (!Array.isArray(stats.weeklyVersions)) stats.weeklyVersions = [];
        return stats;
    } catch {
        return null;
    }
}

export async function openInstanceFolder(name: string): Promise<string> {
    try {
        await OpenInstanceFolderBinding(name);
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo abrir la carpeta de la instancia.';
    }
}

export async function deleteInstance(name: string): Promise<string> {
    try {
        await DeleteInstance(name);
        // Refresco optimista: la card desaparece de inmediato aunque la
        // recarga de la lista falle o tarde; luego se recarga desde el backend.
        instances.value = instances.value.filter((i) => i.name !== name);
        const d = { ...details.value };
        delete d[name];
        details.value = d;
        const l = { ...loaders.value };
        delete l[name];
        loaders.value = l;
        clearDownload(name);
        clearLoaderDl(name);
        // Limpia los mapeos internos para que los eventos tardíos de descarga
        // o modloader no vuelvan a crear estado para la instancia borrada.
        for (const [id, inst] of dlToInstance) {
            if (inst === name) dlToInstance.delete(id);
        }
        for (const [sid, inst] of loaderSessionToInstance) {
            if (inst === name) loaderSessionToInstance.delete(sid);
        }
        await loadInstances();
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo eliminar la instancia.';
    }
}

export async function cloneInstance(name: string, newName: string): Promise<string> {
    try {
        await CloneInstance(name, newName, false);
        await loadInstances();
        await loadDetails(newName);
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo clonar la instancia.';
    }
}

export async function updateConfig(name: string, cfg: InstanceLaunchConfig): Promise<string> {
    try {
        await UpdateInstanceConfig(name, cfg as instance.InstanceLaunchConfig);
        await loadDetails(name);
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo guardar la configuración.';
    }
}

export async function addVersion(name: string, version: string): Promise<string> {
    try {
        const res = await AddInstanceVersion(name, { version });
        if (res?.downloadId) {
            ensureEvents();
            dlToInstance.set(res.downloadId, name);
            registerDownload({
                id: res.downloadId,
                label: name,
                version,
                kind: 'instance',
                state: 'pending',
                percent: 0,
                mbDownloaded: 0,
                mbTotal: 0,
                filesDownloaded: 0,
                filesTotal: 0,
                speedMbps: 0,
            });
            downloads.value = {
                ...downloads.value,
                [name]: {
                    dlId: res.downloadId,
                    version,
                    state: 'pending',
                    percent: 0,
                    mbDownloaded: 0,
                    mbTotal: 0,
                    filesDownloaded: 0,
                    filesTotal: 0,
                },
            };
        }
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo añadir la versión.';
    }
}

export async function cancelDownload(instName: string): Promise<string> {
    const dl = downloads.value[instName];
    if (!dl?.dlId) return '';
    try {
        await CancelInstanceDownload(dl.dlId);
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo cancelar la descarga.';
    }
}

export async function verifyInstance(name: string): Promise<VerifyResult[]> {
    try {
        const res = await VerifyInstance(name);
        return Array.isArray(res) ? (res as VerifyResult[]) : [];
    } catch (e: any) {
        return [
            {
                valid: false,
                version: '',
                issues: [{ type: 'error', file: '', message: e?.message ?? 'No se pudo verificar.' }],
            },
        ];
    }
}

export async function launchInstance(name: string): Promise<string> {
    try {
        const res = await LaunchInstance(name, '', '', '', '', '');
        if (!res?.id && !res?.pid) {
            return 'No se pudo lanzar la instancia.';
        }
        void hideOnLaunchIfEnabled();
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo lanzar la instancia.';
    }
}

export function setLaunching(name: string, on: boolean): void {
    launching.value = { ...launching.value, [name]: on };
}

export async function installInstanceModLoader(
    name: string,
    loader: string,
    loaderVersion: string,
    mcVersion: string
): Promise<{ ok: boolean; sessionId?: string; error?: string }> {
    try {
        const res = await InstallInstanceModLoader(name, loader, loaderVersion, mcVersion);
        if (!res?.sessionId) return { ok: false, error: 'No se pudo iniciar la instalación del modloader.' };
        return { ok: true, sessionId: res.sessionId };
    } catch (e: any) {
        return { ok: false, error: e?.message ?? 'No se pudo instalar el modloader.' };
    }
}

export async function getInstalledInstanceModLoader(name: string): Promise<InstalledLoaderInfo | null> {
    try {
        const res = await GetInstalledInstanceModLoader(name);
        if (!res) return null;
        return res as unknown as InstalledLoaderInfo;
    } catch {
        return null;
    }
}

export async function removeInstanceModLoaderState(name: string): Promise<string> {
    try {
        await RemoveInstanceModLoaderState(name);
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo quitar el modloader.';
    }
}

async function refreshAfterGame(): Promise<void> {
    await loadInstances();
    for (const inst of instances.value) void loadDetails(inst.name);
}

let gameOffs: (() => void)[] | null = null;

export function bindGameEvents() {
    if (gameOffs) return;
    gameOffs = [
        EventsOn('game_exited', () => void refreshAfterGame()),
        EventsOn('game_crashed', () => void refreshAfterGame()),
        EventsOn('game_stopped', () => void refreshAfterGame()),
    ];
}

export function unbindGameEvents() {
    gameOffs?.forEach((off) => off());
    gameOffs = null;
}

export function formatPlayTime(totalSeconds: number): string {
    if (!totalSeconds || totalSeconds <= 0) return 'Sin jugar';
    const h = Math.floor(totalSeconds / 3600);
    const m = Math.floor((totalSeconds % 3600) / 60);
    if (h > 0) return `${h}h ${m}m`;
    return `${m}m`;
}