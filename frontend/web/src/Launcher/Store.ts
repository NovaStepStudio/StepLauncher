import { ref, computed } from 'vue';
import { EventsOn, WindowHide, WindowShow } from '@wailsjs/runtime/runtime';

export interface InstalledVersion {
    id: string;
    type: string;
}

export interface LauncherProfile {
    name: string;
    version?: string;
    gameDir?: string;
    javaExec?: string;
    javaArgs?: string;
    resWidth?: number;
    resHeight?: number;
    fullscreen: boolean;
    modLoader?: string;
    modLoaderVersion?: string;
    icon?: string;
    createdAt: string;
    lastUsed?: string;
    customProperties?: Record<string, string>;
}

const goApp = () => (window as any)?.go?.main?.App;

export const installedVersions = ref<InstalledVersion[]>([]);
export const selectedVersion = ref('');
export const profiles = ref<Record<string, LauncherProfile>>({});
export const selectedProfile = ref('');

export const launching = ref(false);
export const launchMsg = ref('');
export const launchError = ref('');

export const launchPrepare = ref<{ active: boolean; phase: string; current: number; total: number; label: string }>({
    active: false,
    phase: '',
    current: 0,
    total: 0,
    label: '',
});

type PhaseText = { base: string; batch: string; withLabel: string };

const DEFAULT_PHASE_TEXT: PhaseText = {
    base: 'Descargando archivos faltantes…',
    batch: 'Descargando archivos faltantes (%d/%d)…',
    withLabel: 'Descargando %s (%d/%d)…',
};

const PHASE_TEXT: Record<string, PhaseText> = {
    libraries: DEFAULT_PHASE_TEXT,
    natives: {
        base: 'Extrayendo archivos nativos…',
        batch: 'Extrayendo archivos nativos (%d/%d)…',
        withLabel: 'Extrayendo %s (%d/%d)…',
    },
};

export const launchPrepareText = computed(() => {
    const p = launchPrepare.value;
    if (!p.active) return '';
    const meta = (p.phase ? PHASE_TEXT[p.phase] : undefined) ?? DEFAULT_PHASE_TEXT;
    if (p.total > 0 && p.label) {
        return sprintf(meta.withLabel, p.label, p.current, p.total);
    }
    if (p.total > 0) {
        return sprintf(meta.batch, p.current, p.total);
    }
    return meta.base;
});

function sprintf(format: string, ...args: (string | number)[]): string {
    let i = 0;
    return format.replace(/%[sd]/g, () => String(args[i++] ?? ''));
}

export const launchingPhaseLabel = computed(() => {
    const p = launchPrepare.value;
    if (!p.active) return 'Lanzando…';
    if (p.phase === 'natives') return 'Extrayendo…';
    if (p.phase === 'libraries') return 'Descargando…';
    return 'Lanzando…';
});

function resetLaunchPrepare(): void {
    launchPrepare.value = { active: false, phase: '', current: 0, total: 0, label: '' };
}

function onPrepareProgress(raw: unknown): void {
    let obj: any = null;
    try {
        const s = typeof raw === 'string' ? raw : JSON.stringify(raw ?? '');
        obj = JSON.parse(s);
    } catch { }
    const d = obj?.data ?? obj;
    if (!d || typeof d !== 'object') return;
    if (d.finished) {
        resetLaunchPrepare();
        return;
    }
    launchPrepare.value = {
        active: true,
        phase: d.phase ?? '',
        current: d.current ?? 0,
        total: d.total ?? 0,
        label: d.label ?? '',
    };
}

let prepareSubs: (() => void)[] | null = null;

function subscribeLaunchPrepare() {
    if (prepareSubs) return;
    if (!(window as any).runtime) return;
    prepareSubs = [
        EventsOn('game_prepare', (raw) => {
            if (!launching.value) return;
            onPrepareProgress(raw);
        }),
    ];
}

export interface CrashInfo {
    id?: string;
    pid?: number;
    version?: string;
    instanceId?: string;
    playerName?: string;
    status?: string;
    exitCode?: number;
    crashLog?: string;
    crashLogText?: string;
    gameOutputText?: string;
    crashReason?: string;
    crashCategory?: string;
    launcherLogPath?: string;
    minecraftLogPath?: string;
    jvmLogPath?: string;
    launchInfo?: string;
    uptimeMs?: number;
    timestamp?: string;
    javaExec?: string;
    maxRam?: number;
    vanillaVersion?: string;
}

export const crashInfo = ref<CrashInfo | null>(null);

export function onGameCrash(raw: unknown): void {
    let obj: any = null;
    try {
        const s = typeof raw === 'string' ? raw : JSON.stringify(raw ?? '');
        obj = JSON.parse(s);
    } catch { }
    const d = obj?.data ?? obj;
    if (!d || typeof d !== 'object') return;
    crashInfo.value = {
        id: d.id,
        pid: d.pid,
        version: d.version,
        instanceId: d.instanceId,
        playerName: d.playerName,
        status: d.status,
        exitCode: d.exitCode,
        crashLog: d.crashLog,
        crashLogText: d.crashLogText,
        gameOutputText: d.gameOutputText,
        crashReason: d.crashReason,
        crashCategory: d.crashCategory,
        launcherLogPath: d.launcherLogPath,
        minecraftLogPath: d.minecraftLogPath,
        jvmLogPath: d.jvmLogPath,
        launchInfo: d.launchInfo,
        uptimeMs: d.uptimeMs,
        timestamp: d.timestamp,
        javaExec: d.javaExec,
        maxRam: d.maxRam,
        vanillaVersion: d.vanillaVersion,
    };
}

export function clearCrash(): void {
    crashInfo.value = null;
}

let launchMsgTimer: ReturnType<typeof setTimeout> | null = null;

function setLaunchMessage(msg: string, isError = false, persist = false): void {
    if (launchMsgTimer !== null) {
        clearTimeout(launchMsgTimer);
        launchMsgTimer = null;
    }
    if (isError) {
        launchError.value = msg;
        launchMsg.value = '';
    } else {
        launchMsg.value = msg;
        launchError.value = '';
    }
    if (persist) {
        launchMsgTimer = null;
        return;
    }
    launchMsgTimer = setTimeout(() => {
        launchMsg.value = '';
        launchError.value = '';
        launchMsgTimer = null;
    }, 6000);
}

export function hideLaunchMessage(): void {
    if (launchMsgTimer !== null) {
        clearTimeout(launchMsgTimer);
        launchMsgTimer = null;
    }
    launchMsg.value = '';
    launchError.value = '';
}

export const hasVersions = computed(() => installedVersions.value.length > 0);

export const canLaunch = computed(() => {
    if (selectedProfile.value) {
        const p = profiles.value[selectedProfile.value];
        if (p?.version?.trim()) return true;
    }
    return !!selectedVersion.value;
});

const TYPE_ORDER: Record<string, number> = {
    release: 0,
    snapshot: 1,
    old_beta: 2,
    old_alpha: 3,
};

export const VERSION_GROUP_META: Array<{ type: string; label: string }> = [
    { type: 'release', label: 'Releases' },
    { type: 'snapshot', label: 'Snapshots' },
    { type: 'old_beta', label: 'Beta antiguas' },
    { type: 'old_alpha', label: 'Alpha antiguas' },
    { type: 'other', label: 'Otras' },
];

function compareVersions(a: string, b: string): number {
    const pa = a.split('.').map((n) => parseInt(n, 10) || 0);
    const pb = b.split('.').map((n) => parseInt(n, 10) || 0);
    for (let i = 0; i < Math.max(pa.length, pb.length); i++) {
        const va = pa[i] ?? 0;
        const vb = pb[i] ?? 0;
        if (va !== vb) return va - vb;
    }
    return 0;
}

export const groupedVersions = computed(() => {
    const groups: Array<{ type: string; label: string; items: InstalledVersion[] }> = [];
    for (const meta of VERSION_GROUP_META) {
        const items = installedVersions.value
            .filter((v) => (meta.type === 'other' ? !(TYPE_ORDER[v.type] !== undefined) : v.type === meta.type))
            .sort((a, b) => compareVersions(b.id, a.id));
        if (items.length) groups.push({ type: meta.type, label: meta.label, items });
    }
    return groups;
});

export async function loadVersions(): Promise<void> {
    try {
        const list = await goApp()?.ListDownloadedVersions?.();
        if (Array.isArray(list)) {
            installedVersions.value = (list as InstalledVersion[]).filter((v) => v && typeof v.id === 'string');
        }
    } catch { }
    try {
        const last = await goApp()?.GetSelectedVersion?.();
        if (typeof last === 'string' && last && installedVersions.value.some((v) => v.id === last)) {
            selectedVersion.value = last;
        }
    } catch { }
    ensureVersionSelected();
}

export async function loadProfiles(): Promise<void> {
    try {
        const p = await goApp()?.ListProfiles?.();
        if (p && typeof p === 'object') profiles.value = p as Record<string, LauncherProfile>;
    } catch { }
    try {
        const sel = await goApp()?.GetSelectedProfile?.();
        if (typeof sel === 'string') selectedProfile.value = sel;
    } catch { }
    ensureVersionSelected();
    syncProfileVersion();
}

function syncProfileVersion(): void {
    const p = selectedProfile.value ? profiles.value[selectedProfile.value] : undefined;
    const v = p?.version?.trim();
    if (v) selectedVersion.value = v;
}

export function ensureVersionSelected(): void {
    const ids = new Set(installedVersions.value.map((v) => v.id));
    if (ids.has(selectedVersion.value)) return;
    const sorted = [...installedVersions.value].sort((a, b) => {
        const oa = TYPE_ORDER[a.type] ?? 9;
        const ob = TYPE_ORDER[b.type] ?? 9;
        if (oa !== ob) return oa - ob;
        return compareVersions(b.id, a.id);
    });
    selectedVersion.value = sorted[0]?.id ?? '';
}

export function selectVersion(id: string): void {
    if (!installedVersions.value.some((v) => v.id === id)) return;
    if (selectedProfile.value) {
        selectedProfile.value = '';
        void dismissPersistedProfile();
    }
    selectedVersion.value = id;
    hideLaunchMessage();
    void persistSelectedVersion(id);
}

export async function persistSelectedVersion(id: string): Promise<void> {
    try {
        await goApp()?.SetSelectedVersion?.(id);
    } catch { }
}

async function dismissPersistedProfile(): Promise<void> {
    try {
        await goApp()?.SetSelectedProfile?.('');
    } catch { }
}

export async function createProfile(p: LauncherProfile): Promise<string> {
    try {
        await goApp()?.CreateProfile?.(p);
        await loadProfiles();
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo crear el perfil.';
    }
}

export async function updateProfile(name: string, p: LauncherProfile): Promise<string> {
    try {
        await goApp()?.UpdateProfile?.(name, p);
        await loadProfiles();
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo actualizar el perfil.';
    }
}

export async function deleteProfile(name: string): Promise<string> {
    try {
        await goApp()?.DeleteProfile?.(name);
        if (selectedProfile.value === name) selectedProfile.value = '';
        await loadProfiles();
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo eliminar el perfil.';
    }
}

export async function setSelectedProfile(name: string): Promise<string> {
    try {
        await goApp()?.SetSelectedProfile?.(name || '');
        selectedProfile.value = name || '';
        await loadProfiles();
        const v = selectedVersion.value;
        if (v) void persistSelectedVersion(v);
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo seleccionar el perfil.';
    }
}

export async function launchGame(): Promise<string> {
    hideLaunchMessage();
    const prof = selectedProfile.value ? profiles.value[selectedProfile.value] : undefined;
    const effVersion = prof?.version?.trim() ? prof.version : selectedVersion.value;
    if (!effVersion) {
        setLaunchMessage('Elige una versión descargada para poder jugar.', true);
        return launchError.value;
    }
    const label = selectedProfile.value
        ? `Lanzando ${effVersion} con el perfil ${selectedProfile.value}…`
        : `Lanzando ${effVersion}…`;
    launching.value = true;
    setLaunchMessage(label, false, true);
    resetLaunchPrepare();
    subscribeLaunchPrepare();
    try {
        const resp = await goApp()?.LaunchMinecraft?.({
            Version: effVersion,
            Profile: selectedProfile.value || '',
        });
        if (resp?.id) {
            setLaunchMessage(label, false, true);
            hideOnLaunchIfEnabled();
        }
    } catch (e: any) {
        setLaunchMessage(e?.message ?? 'No se pudo lanzar Minecraft.', true);
    } finally {
        launching.value = false;
        resetLaunchPrepare();
        if (!launchError.value) setLaunchMessage(label, false);
    }
    return launchError.value;
}

let windowHideSubs: (() => void)[] | null = null;

export function subscribeWindowHideRestore() {
    if (windowHideSubs) return;
    if (!(window as any).runtime) return;
    windowHideSubs = [
        EventsOn('game_exited', () => maybeShowWindow()),
        EventsOn('game_crashed', (data) => {
            onGameCrash(data);
            maybeShowWindow();
        }),
        EventsOn('game_stopped', () => maybeShowWindow()),
    ];
}

export async function maybeShowWindow() {
    try {
        const games = await goApp()?.ListGames?.();
        const running = Array.isArray(games) && games.some((g) => g.status === 'running' || g.status === 'starting');
        if (!running) WindowShow();
    } catch {
        WindowShow();
    }
}

export async function hideOnLaunchIfEnabled() {
    try {
        const cfg = await goApp()?.GetConfig?.();
        if (cfg?.launcher?.hideLauncherOnLaunch === false) return;
        const games = await goApp()?.ListGames?.();
        const running = Array.isArray(games) && games.some((g) => g.status === 'running' || g.status === 'starting');
        if (!running) return;
        subscribeWindowHideRestore();
        WindowHide();
    } catch { }
}

export async function refreshAfterDownload(): Promise<void> {
    await loadVersions();
    await loadProfiles();
}
