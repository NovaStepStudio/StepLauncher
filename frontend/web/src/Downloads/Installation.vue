<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue';
import {
    IconDownload,
    IconX,
    IconSearch,
    IconPlayerPause,
    IconRotateClockwise,
    IconPlayerStop,
    IconCheck,
    IconAlertTriangle,
    IconChevronDown,
} from '@tabler/icons-vue';
import { CLOSE_OVERLAYS_EVENT, idleOptions } from '@/Common/Stores/Idle';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';
import { registerDownload, clearDownload as clearCentralDownload } from './Store';
import { refreshAfterDownload, selectVersion, launchGame, installedVersions } from '@/Launcher/Store';
import { downloads } from '@/Instances/Store';
import { EventsOn } from '@wailsjs/runtime/runtime';
import {
    FetchVersionManifest,
    StartFullDownload,
    GetDownloadStatus,
    PauseDownload,
    ResumeDownload,
    CancelDownload,
    GetModLoaderVersions,
    ListDownloads,
    InstallModLoader,
} from '@wailsjs/go/main/App';
import type { downloader, modloader } from '@wailsjs/go/models';

import iconVanilla from '../../assets/icons/minecraft.png';
import iconFabric from '../../assets/icons/fabric.png';
import iconForge from '../../assets/icons/forge.png';
import iconNeoForge from '../../assets/icons/neoforge.png';
import iconQuilt from '../../assets/icons/quilt.png';
import iconLegacyFabric from '../../assets/icons/legacyfabric.png';

type DownloadProgress = downloader.DownloadProgress;
type DownloadInfo = {
    id: string;
    version: string;
    state: string;
    error?: string;
};
type VersionEntry = { id: string; type: string; releaseTime: string };
type LoaderVersion = modloader.LoaderVersion;

const props = defineProps<{
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
}>();

type Phase = 'setup' | 'installing' | 'done' | 'error';

const phase = ref<Phase>('setup');
const manifestLoaded = ref(false);
const manifestError = ref('');
const loadingManifest = ref(false);

const versions = ref<VersionEntry[]>([]);
const latestRelease = ref('');
const latestSnapshot = ref('');
const query = ref('');
const selectedVersion = ref('');

const LOADERS: Array<{ id: string; name: string; icon: string; tag: string }> = [
    { id: 'vanilla', name: 'Vanilla', icon: iconVanilla, tag: 'Original' },
    { id: 'fabric', name: 'Fabric', icon: iconFabric, tag: 'Ligero' },
    { id: 'forge', name: 'Forge', icon: iconForge, tag: 'Modding clásico' },
    { id: 'neoforge', name: 'NeoForged', icon: iconNeoForge, tag: 'Moderno' },
    { id: 'quilt', name: 'Quilt', icon: iconQuilt, tag: 'Fork de Fabric' },
    { id: 'legacyfabric', name: 'Legacy Fabric', icon: iconLegacyFabric, tag: 'Versiones antiguas' },
];

const selectedLoader = ref('vanilla');
const loaderVersions = ref<LoaderVersion[]>([]);
const loaderVersion = ref('');

const progress = ref<DownloadProgress | null>(null);
const downloadState = ref('');
const dlId = ref('');
const errorMessage = ref('');

const pendingLoaderInstall = ref(false);
const loaderPhase = ref<'idle' | 'running' | 'done' | 'error'>('idle');
const loaderSession = ref('');
const loaderStatus = ref('');
const loaderProgress = ref(0);
const loaderTotal = ref(0);
const loaderLogs = ref<string[]>([]);

const gameDone = ref(false);

type Summary = { files: number; mb: number };
const finalSummary = ref<Summary | null>(null);

const detailsOpen = ref(false);

const RING_R = 52;
const RING_CIRC = 2 * Math.PI * RING_R;
const ringOffset = computed(() =>
    RING_CIRC * (1 - clampPct(progress.value?.percent ?? 0) / 100)
);

const ringMode = computed<'percent' | 'busy' | 'done'>(() => {
    if (loaderPhase.value === 'running') return 'busy';
    const st = downloadState.value || progress.value?.state || '';
    if (st === 'completed') return 'done';
    if (st === 'verifying' || st === 'redownloading') return 'busy';
    return 'percent';
});

const ringClass = computed(() => ({
    done: ringMode.value === 'done',
    busy: ringMode.value === 'busy',
}));

const stepHint = computed(() => {
    const st = downloadState.value || progress.value?.state || '';
    if (st === 'verifying') {
        return progress.value?.currentSection === 'extracting_natives'
            ? 'Extrayendo las bibliotecas nativas. Puede tardar unos segundos…'
            : 'Comprobando la integridad de los archivos descargados…';
    }
    if (st === 'redownloading') {
        return 'Reparando los archivos que fallaron la verificación…';
    }
    return '';
});

let eventOffs: Array<() => void> = [];

function fmt(n: number, digits = 1): string {
    if (!Number.isFinite(n)) return '0';
    return n.toLocaleString('es-ES', { maximumFractionDigits: digits });
}

function fmtMb(mb: number): string {
    if (!Number.isFinite(mb) || mb <= 0) return '0 MB';
    if (mb >= 1024) return `${fmt(mb / 1024)} GB`;
    return `${fmt(mb)} MB`;
}

function fmtSpeed(mbps: number): string {
    if (!Number.isFinite(mbps) || mbps <= 0) return '0 MB/s';
    if (mbps >= 1024) return `${fmt(mbps / 1024)} GB/s`;
    return `${fmt(mbps)} MB/s`;
}

function clampPct(v: number): number {
    return Math.min(100, Math.max(0, v));
}

function riskText(e: unknown): string {
    const msg = (e as { message?: string })?.message;
    if (msg) return msg;
    return String(e ?? 'Error desconocido');
}

const SECTION_LABELS: Record<string, string> = {
    client: 'Cliente',
    libraries: 'Librerías',
    natives: 'Nativas',
    assets: 'Assets',
    asset_index: 'Índice assets',
    java: 'Java',
    extracting_natives: 'Nativas',
};

function sectionLabel(s: string): string {
    return SECTION_LABELS[s] ?? s;
}

function stateBadge(s: string): { text: string; tone: 'accent' | 'warn' | 'ok' | 'muted' } {
    switch (s) {
        case 'downloading':
            return { text: 'Descargando', tone: 'accent' };
        case 'paused':
            return { text: 'En pausa', tone: 'warn' };
        case 'verifying':
            return progress.value?.currentSection === 'extracting_natives'
                ? { text: 'Extrayendo bibliotecas nativas', tone: 'accent' }
                : { text: 'Verificando archivos', tone: 'accent' };
        case 'redownloading':
            return { text: 'Re-descargando archivos', tone: 'warn' };
        case 'pending':
            return { text: 'Iniciando descarga…', tone: 'accent' };
        case 'completed':
            return { text: 'Descarga completada', tone: 'ok' };
        case 'cancelled':
            return { text: 'Cancelado', tone: 'muted' };
        case 'error':
            return { text: 'Error', tone: 'muted' };
        default:
            return { text: 'Preparando…', tone: 'accent' };
    }
}

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

const FILTER_TABS: Array<{ id: 'release' | 'snapshot' | 'old'; label: string }> = [
    { id: 'release', label: 'Releases' },
    { id: 'snapshot', label: 'Snapshots' },
    { id: 'old', label: 'Antiguas' },
];

const versionFilter = ref<'release' | 'snapshot' | 'old'>('release');

const GROUP_META: Array<{ type: string; label: string }> = [
    { type: 'release', label: 'Releases' },
    { type: 'snapshot', label: 'Snapshots' },
    { type: 'old_beta', label: 'Beta antiguas' },
    { type: 'old_alpha', label: 'Alpha antiguas' },
];

function groupVisible(type: string): boolean {
    switch (versionFilter.value) {
        case 'release':
            return type === 'release';
        case 'snapshot':
            return type === 'snapshot';
        case 'old':
            return type === 'old_beta' || type === 'old_alpha';
        default:
            return false;
    }
}

const versionGroups = computed(() => {
    const q = query.value.trim().toLowerCase();
    const groups: Array<{ type: string; label: string; items: VersionEntry[] }> = [];
    for (const meta of GROUP_META) {
        if (!groupVisible(meta.type)) continue;
        const items = versions.value
            .filter((v) => v.type === meta.type)
            .filter((v) => !q || v.id.toLowerCase().includes(q))
            .sort((a, b) => compareVersions(b.id, a.id));
        if (items.length) groups.push({ type: meta.type, label: meta.label, items });
    }
    return groups;
});

const visibleVersionCount = computed(() =>
    versionGroups.value.reduce((acc, g) => acc + g.items.length, 0)
);

function pickDefaultVersion() {
    const release = versions.value.find((v) => v.type === 'release');
    const snap = versions.value.find((v) => v.type === 'snapshot');
    selectedVersion.value = release?.id ?? snap?.id ?? versions.value[0]?.id ?? '';
}

function isLatest(id: string, type: string): boolean {
    if (type === 'release') return latestRelease.value === id;
    if (type === 'snapshot') return latestSnapshot.value === id;
    return false;
}

function isInstalled(id: string): boolean {
    return installedVersions.value.some((v) => v.id === id);
}

async function loadManifest() {
    if (manifestLoaded.value) return;
    loadingManifest.value = true;
    manifestError.value = '';
    try {
        const m = await FetchVersionManifest();
        versions.value = (m?.versions ?? []).map((v) => ({
            id: v.id,
            type: v.type ?? 'release',
            releaseTime: v.releaseTime ?? '',
        }));
        latestRelease.value = m?.latest?.release ?? '';
        latestSnapshot.value = m?.latest?.snapshot ?? '';
        manifestLoaded.value = true;
        pickDefaultVersion();
    } catch (e: any) {
        manifestError.value = riskText(e);
    } finally {
        loadingManifest.value = false;
    }
}

const manualLoaderVersion = ref(false);

type CompatStatus = 'loading' | 'ok' | 'empty' | 'error';
type CompatEntry = { status: CompatStatus; versions: LoaderVersion[] };

const compatData = ref<Record<string, CompatEntry>>({});
const autoLoaderResolved = ref('');

let compatToken = 0;

function compatOf(loader: string): CompatEntry | undefined {
    return compatData.value[loader];
}

function isLoaderSupported(loader: string): boolean {
    if (loader === 'vanilla') return true;
    const s = compatOf(loader)?.status;
    return s === 'ok' || s === 'loading';
}

function compatStatus(loader: string): CompatStatus {
    if (loader === 'vanilla') return 'ok';
    return compatOf(loader)?.status ?? 'loading';
}

function recommendFrom(list: LoaderVersion[]): string {
    const stable = list?.find((v) => v.stable);
    return stable?.loaderVersion ?? list?.[0]?.loaderVersion ?? '';
}

function pickSelectedLoaderVersion() {
    const e = compatOf(selectedLoader.value);
    if (!e || e.status !== 'ok') return;
    if (manualLoaderVersion.value) {
        loaderVersions.value = e.versions;
    } else {
        autoLoaderResolved.value = recommendFrom(e.versions);
    }
}

async function refreshCompat(mc: string) {
    const token = ++compatToken;
    const ids = LOADERS.map((l) => l.id).filter((id) => id !== 'vanilla');
    for (const id of ids) compatData.value[id] = { status: 'loading', versions: [] };
    loaderVersions.value = [];
    loaderVersion.value = '';
    autoLoaderResolved.value = '';
    await Promise.allSettled(
        ids.map(async (id) => {
            try {
                const list = await GetModLoaderVersions(id, mc);
                if (token !== compatToken) return;
                compatData.value[id] = {
                    status: list && list.length ? 'ok' : 'empty',
                    versions: list ?? [],
                };
            } catch {
                if (token !== compatToken) return;
                compatData.value[id] = { status: 'error', versions: [] };
            }
        })
    );
    if (token !== compatToken || selectedVersion.value !== mc) return;
    pickSelectedLoaderVersion();
}

function retrySelectedLoader() {
    const loader = selectedLoader.value;
    const mc = selectedVersion.value;
    if (!mc || loader === 'vanilla') return;
    const token = compatToken;
    compatData.value[loader] = { status: 'loading', versions: [] };
    GetModLoaderVersions(loader, mc)
        .then((list) => {
            if (token !== compatToken) return;
            compatData.value[loader] = {
                status: list && list.length ? 'ok' : 'empty',
                versions: list ?? [],
            };
            if (selectedVersion.value === mc) pickSelectedLoaderVersion();
        })
        .catch(() => {
            if (token !== compatToken) return;
            compatData.value[loader] = { status: 'error', versions: [] };
        });
}

let lastMc = '';
watch(
    [selectedVersion, selectedLoader, manualLoaderVersion],
    () => {
        const mc = selectedVersion.value;
        if (!mc) return;
        if (mc !== lastMc) {
            lastMc = mc;
            refreshCompat(mc);
        } else {
            loaderVersions.value = [];
            loaderVersion.value = '';
            autoLoaderResolved.value = '';
            pickSelectedLoaderVersion();
        }
    },
    { immediate: false }
);

const effectiveLoaderVersion = computed(() => {
    if (selectedLoader.value === 'vanilla') return '';
    if (manualLoaderVersion.value) return loaderVersion.value;
    return autoLoaderResolved.value;
});

const installReady = computed(() => {
    if (!selectedVersion.value || !manifestLoaded.value || manifestError.value) return false;
    if (selectedLoader.value === 'vanilla') return true;
    return effectiveLoaderVersion.value !== '';
});

const selectedLoaderStatus = computed(() => compatStatus(selectedLoader.value));

const selectedLoaderHint = computed(() => {
    const mc = selectedVersion.value;
    const s = selectedLoaderStatus.value;
    if (s === 'empty') return mc ? `No disponible para ${mc}` : 'No disponible para esta versión';
    if (s === 'error') return mc ? `No se pudo verificar la compatibilidad con ${mc}` : 'No se pudo verificar';
    if (s === 'ok' && !manualLoaderVersion.value && autoLoaderResolved.value)
        return `Se usará la versión recomendada: ${autoLoaderResolved.value}`;
    return '';
});

const selectedLoaderIcon = computed(() => {
    const l = LOADERS.find((x) => x.id === selectedLoader.value);
    return l?.icon ?? LOADERS[0]?.icon ?? '';
});

function resetRun() {
    clearCentralDownload(dlId.value);
    detailsOpen.value = false;
    progress.value = null;
    downloadState.value = '';
    dlId.value = '';
    errorMessage.value = '';
    gameDone.value = false;
    finalSummary.value = null;
    pendingLoaderInstall.value = false;
    loaderPhase.value = 'idle';
    loaderSession.value = '';
    loaderStatus.value = '';
    loaderProgress.value = 0;
    loaderTotal.value = 0;
    loaderLogs.value = [];
    lastLogLine = '';
    autoLaunchHandled.value = false;
}

function syncDownloadWidget() {
    const p = progress.value;
    if (!p || !dlId.value) {
        clearCentralDownload(dlId.value);
        return;
    }
    registerDownload({
        id: dlId.value,
        label: `Minecraft ${selectedVersion.value}`,
        version: selectedVersion.value,
        kind: 'version',
        state: downloadState.value || p.state || '',
        percent: clampPct(p.percent ?? 0),
        mbDownloaded: p.mbDownloaded ?? 0,
        mbTotal: p.mbTotal ?? 0,
        filesDownloaded: p.filesDownloaded ?? 0,
        filesTotal: p.filesTotal ?? 0,
        speedMbps: p.speedMbps ?? 0,
    });
}

function applyProgress(p: DownloadProgress) {
    if (!p) return;
    progress.value = p;
    downloadState.value = p.state ?? '';
    syncDownloadWidget();
    if (p.state === 'completed') handleGameCompleted();
}

function handleGameCompleted() {
    if (gameDone.value) return;
    gameDone.value = true;
    const pr = progress.value;
    finalSummary.value = {
        files: pr?.filesDownloaded ?? 0,
        mb: pr?.mbDownloaded ?? 0,
    };
    if (pendingLoaderInstall.value) {
        startLoaderInstall();
    } else {
        phase.value = 'done';
    }
}

async function onInstall() {
    if (!selectedVersion.value) return;
    resetRun();
    phase.value = 'installing';
    errorMessage.value = '';
    try {
        const info = await StartFullDownload(selectedVersion.value);
        dlId.value = info?.id ?? '';
        if (!dlId.value) {
            phase.value = 'error';
            errorMessage.value = 'No se pudo iniciar la descarga (motor no disponible).';
            return;
        }
        try {
            const st = await GetDownloadStatus(dlId.value);
            applyProgress(st as DownloadProgress);
        } catch { }
        if (selectedLoader.value !== 'vanilla') {
            pendingLoaderInstall.value = true;
        }
    } catch (e: any) {
        phase.value = 'error';
        errorMessage.value = riskText(e);
    }
}

async function startLoaderInstall() {
    loaderPhase.value = 'running';
    loaderStatus.value = 'Preparando la instalación del loader…';
    loaderLogLine('Preparando la instalación del ModLoader…');
    try {
        const res = await InstallModLoader(selectedLoader.value, effectiveLoaderVersion.value, selectedVersion.value, '');
        loaderSession.value = res?.sessionId ?? '';
    } catch (e: any) {
        loaderPhase.value = 'error';
        phase.value = 'error';
        errorMessage.value = `Fallo al instalar el loader: ${riskText(e)}`;
    }
}

let lastLogLine = '';
const loaderLogEl = ref<HTMLElement | null>(null);
function loaderLogLine(line: string): void {
    // Los instaladores oficiales pueden emitir códigos ANSI (colores/progreso):
    // se limpian para que el log se vea como texto plano.
    const clean = line.replace(/\u001b\[[0-9;?]*[a-zA-Z]/g, '');
    if (clean === '' || clean === lastLogLine) return;
    lastLogLine = clean;
    loaderLogs.value = [...loaderLogs.value, clean].slice(-40);
    void nextTick(() => {
        const el = loaderLogEl.value;
        if (el) el.scrollTop = el.scrollHeight;
    });
}

async function onPause() {
    if (!dlId.value) return;
    try {
        await PauseDownload(dlId.value);
        downloadState.value = 'paused';
        syncDownloadWidget();
    } catch (e: any) {
        errorMessage.value = riskText(e);
    }
}

async function onResume() {
    if (!dlId.value) return;
    try {
        await ResumeDownload(dlId.value);
        downloadState.value = 'downloading';
        syncDownloadWidget();
    } catch (e: any) {
        errorMessage.value = riskText(e);
    }
}

async function onCancel() {
    pendingLoaderInstall.value = false;
    if (dlId.value) {
        try {
            await CancelDownload(dlId.value);
        } catch { }
    }
    resetRun();
    phase.value = 'setup';
}

function onCloseOverlays() {
    emit('update:visible', false);
}

function closeModal() {
    emit('update:visible', false);
}

useOverlayEscape(closeModal, { isActive: () => props.visible });

async function syncFromBackend() {
    try {
        // Recuperación de estado al abrir el modal desde cualquier parte:
        // si hay una descarga activa y no pertenece a una instancia (las de
        // instancia las gestiona su propio modal), se retoma su visualización
        // con la versión que estaba descargando; si el estado anterior era un
        // resumen cerrado (done/error) de una instalación ya terminada, se
        // restablece el menú de instalación para no tener que pulsar
        // "Nueva instalación" a mano.
        const ACTIVE = ['pending', 'downloading', 'paused', 'verifying', 'redownloading'];
        const dl = (await ListDownloads()) ?? [];
        const instIds = new Set(Object.values(downloads.value).map((x) => x.dlId));
        const active = dl.find(
            (d: DownloadInfo) => ACTIVE.includes(d.state) && !instIds.has(d.id)
        );
        if (active) {
            selectedVersion.value = active.version;
            dlId.value = active.id;
            pendingLoaderInstall.value = false;
            phase.value = 'installing';
            const st = await GetDownloadStatus(active.id);
            if (st) applyProgress(st as DownloadProgress);
            return;
        }
        if (phase.value === 'done' || phase.value === 'error') {
            resetRun();
            phase.value = 'setup';
        }
    } catch { }
}

watch(
    () => props.visible,
    (v) => {
        if (!v) return;
        loadManifest();
        syncFromBackend();
    }
);

let doneResetTimer: number | null = null;

const autoLaunchHandled = ref(false);

async function maybeAutoLaunch() {
    if (autoLaunchHandled.value) return;
    let cfg: any = null;
    try {
        cfg = await (window as any).go?.main?.App?.GetConfig?.();
    } catch { }
    if (!cfg?.launcher?.launchAfterInstall) return;
    autoLaunchHandled.value = true;
    const installed = selectedVersion.value;
    try {
        await refreshAfterDownload();
        if (installed) selectVersion(installed);
    } catch { }
    emit('update:visible', false);
    window.setTimeout(() => {
        void launchGame();
    }, 300);
}

// El resumen de una instalación terminada se mantiene durante el tiempo que
// el usuario tenga configurado como "Minutos de Inactividad" (antes 5 min).
function scheduleDoneReset() {
    if (doneResetTimer !== null) {
        window.clearTimeout(doneResetTimer);
    }
    const minutes = Math.max(1, idleOptions.value.idleMinutes);
    doneResetTimer = window.setTimeout(() => {
        doneResetTimer = null;
        resetRun();
        phase.value = 'setup';
    }, minutes * 60 * 1000);
}

watch(phase, (p) => {
    if (doneResetTimer !== null) {
        window.clearTimeout(doneResetTimer);
        doneResetTimer = null;
    }
    if (p === 'done') {
        scheduleDoneReset();
        void maybeAutoLaunch();
    }
});

function parseDownloadEvent(raw: unknown): { id: string; data: any } | null {
    try {
        const s = typeof raw === 'string' ? raw : JSON.stringify(raw ?? '');
        const obj = JSON.parse(s) as { id?: string; data?: any };
        if (!obj || typeof obj.id !== 'string' || typeof obj.data !== 'object') return null;
        return { id: obj.id, data: obj.data };
    } catch {
        return null;
    }
}

function parseModLoaderEvent(raw: unknown): any {
    try {
        const s = typeof raw === 'string' ? raw : JSON.stringify(raw ?? '');
        return JSON.parse(s);
    } catch {
        return null;
    }
}

function onDownloadProgress(raw: unknown) {
    const evt = parseDownloadEvent(raw);
    if (!evt || !dlId.value || evt.id !== dlId.value) return;
    applyProgress(evt.data as DownloadProgress);
}

function onDownloadState(raw: unknown) {
    const evt = parseDownloadEvent(raw);
    if (!evt || !dlId.value || evt.id !== dlId.value) return;
    const st = evt.data?.state as string;
    downloadState.value = st;
    syncDownloadWidget();
    if (st === 'completed') handleGameCompleted();
    if (st === 'cancelled') {
        pendingLoaderInstall.value = false;
        clearCentralDownload(dlId.value);
    }
}

function onDownloadError(raw: unknown) {
    const evt = parseDownloadEvent(raw);
    if (!evt || !dlId.value || evt.id !== dlId.value) return;
    pendingLoaderInstall.value = false;
    clearCentralDownload(dlId.value);
    phase.value = 'error';
    errorMessage.value = evt.data?.error ?? 'Error desconocido durante la descarga.';
}

function onModLoaderEvent(raw: unknown) {
    const e = parseModLoaderEvent(raw);
    if (!e || !e.type) return;
    if (loaderSession.value && e.sessionId && e.sessionId !== loaderSession.value) return;
    loaderPhase.value = 'running';
    switch (e.type) {
        case 'modloader_resolving':
            loaderStatus.value = `Resolviendo versión de ${e.loader}…`;
            loaderLogLine(`Resolviendo versión de ${e.loader}…`);
            break;
        case 'modloader_downloading':
            loaderStatus.value = `${e.message ?? 'Descargando…'}`;
            loaderProgress.value = e.progress ?? 0;
            loaderTotal.value = e.total ?? 0;
            loaderLogLine(e.message ?? `Descargando librerías del ModLoader (${e.progress ?? 0}/${e.total ?? 0})…`);
            break;
        case 'modloader_installing':
            loaderStatus.value = e.message ?? 'Instalando…';
            loaderLogLine(e.message ?? 'Instalando el ModLoader…');
            break;
        case 'modloader_installed':
            loaderPhase.value = 'done';
            loaderStatus.value = `${e.loader} ${e.version ?? ''} instalado.`;
            loaderLogLine(`${e.loader} ${e.version ?? ''} instalado correctamente.`);
            phase.value = 'done';
            break;
        case 'modloader_error':
            loaderPhase.value = 'error';
            phase.value = 'error';
            loaderLogLine(`Error: ${e.error ?? 'desconocido'}`);
            errorMessage.value = `Fallo al instalar el loader: ${e.error ?? 'desconocido'}`;
            break;
    }
}

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    eventOffs = [
        EventsOn('download_progress', onDownloadProgress),
        EventsOn('download_state', onDownloadState),
        EventsOn('download_error', onDownloadError),
        EventsOn('modloader_resolving', onModLoaderEvent),
        EventsOn('modloader_downloading', onModLoaderEvent),
        EventsOn('modloader_installing', onModLoaderEvent),
        EventsOn('modloader_installed', onModLoaderEvent),
        EventsOn('modloader_error', onModLoaderEvent),
    ];
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    eventOffs.forEach((off) => off());
    eventOffs = [];
    if (doneResetTimer !== null) {
        window.clearTimeout(doneResetTimer);
        doneResetTimer = null;
    }
});
</script>

<template>
    <Teleport to="body">
        <Transition name="InstallationModal">
            <div v-if="visible" class="InstallationModal_Overlay" @click.self="closeModal">
                <div class="InstallationModal_Dialog">
                    <div class="InstallationModal_Head">
                        <span class="InstallationModal_Icon">
                            <img v-if="phase !== 'setup'" :src="selectedLoaderIcon" alt="" />
                            <IconDownload v-else stroke="2" />
                        </span>
                        <div class="InstallationModal_Titles">
                            <h3>Instalar Minecraft</h3>
                            <p>{{ phase === 'setup' ? 'Elige versión y loader para descargarla' : selectedVersion ? `Descargando ${selectedVersion}${selectedLoader !== 'vanilla' ? ` · ${selectedLoader}` : ''}` : 'Descargando…' }}</p>
                        </div>
                        <button class="InstallationModal_Close" title="Cerrar" @click="closeModal">
                            <IconX stroke="2" />
                        </button>
                    </div>

                    <div class="InstallationModal_Body">
                        <template v-if="phase === 'setup'">
                            <div class="InstallationModal_Setup">
                                <section class="InstallationModal_Panel main">
                                    <div class="InstallationModal_PanelHead">
                                        <span class="InstallationModal_Label">Versión</span>
                                        <span v-if="manifestLoaded" class="InstallationModal_Count">{{ visibleVersionCount }} versiones</span>
                                    </div>

                                    <div class="InstallationModal_Search">
                                        <IconSearch />
                                        <input
                                            v-model="query"
                                            type="text"
                                            placeholder="Buscar versión…"
                                            spellcheck="false"
                                        />
                                    </div>

                                    <div class="InstallationModal_FilterTabs">
                                        <button
                                            v-for="f in FILTER_TABS"
                                            :key="f.id"
                                            class="InstallationModal_FilterTab"
                                            :class="{ active: versionFilter === f.id }"
                                            @click="versionFilter = f.id"
                                        >{{ f.label }}</button>
                                    </div>

                                    <div v-if="loadingManifest || !manifestLoaded" class="InstallationModal_Loading">
                                        <span class="InstallationModal_Spinner" aria-hidden="true"></span>
                                        <span>Cargando versiones desde el manifest de Mojang…</span>
                                    </div>

                                    <div v-else-if="manifestError" class="InstallationModal_Notice error">
                                        <IconAlertTriangle />
                                        <span>No se pudo cargar el manifest: {{ manifestError }}</span>
                                    </div>

                                    <div v-else class="InstallationModal_List">
                                        <template v-for="group in versionGroups" :key="group.type">
                                            <div class="InstallationModal_GroupHead">
                                                <span>{{ group.label }}</span>
                                                <em>{{ group.items.length }}</em>
                                            </div>
                                            <button
                                                v-for="v in group.items"
                                                :key="v.id"
                                                class="InstallationModal_Version"
                                                :class="{ active: v.id === selectedVersion, reinstalling: isInstalled(v.id) }"
                                                :title="isInstalled(v.id) ? `Ya está instalada: se volverá a descargar (reinstalación)` : ''"
                                                @click="selectedVersion = v.id"
                                            >
                                                <span class="InstallationModal_VersionId">{{ v.id }}</span>
                                                <em v-if="isInstalled(v.id)" class="InstallationModal_VersionBadge installed">REINSTALAR</em>
                                                <em v-else-if="isLatest(v.id, v.type)" class="InstallationModal_VersionBadge latest">ÚLTIMA</em>
                                                <em v-else-if="v.type === 'snapshot'" class="InstallationModal_VersionBadge snap">SNAPSHOT</em>
                                                <em v-else-if="v.type !== 'release'" class="InstallationModal_VersionBadge legacy">ANTIGUA</em>
                                                <IconCheck v-if="v.id === selectedVersion" class="InstallationModal_VersionCheck" stroke="2.4" />
                                            </button>
                                        </template>
                                        <div v-if="!versionGroups.length" class="InstallationModal_Empty">
                                            <IconSearch stroke="2" />
                                            <span>Sin resultados para «{{ query }}»</span>
                                        </div>
                                    </div>

                                    <div v-if="isInstalled(selectedVersion)" class="InstallationModal_Notice info">
                                        <IconAlertTriangle />
                                        <span>Esta versión ya está instalada: <strong>se volverá a descargar</strong> (reinstalación).</span>
                                    </div>
                                </section>

                                <section class="InstallationModal_Panel side">
                                    <div class="InstallationModal_PanelHead">
                                        <span class="InstallationModal_Label">Modloader</span>
                                        <span v-if="selectedVersion" class="InstallationModal_Count">para {{ selectedVersion }}</span>
                                    </div>

                                    <div class="InstallationModal_Loaders">
                                        <button
                                            v-for="l in LOADERS"
                                            :key="l.id"
                                            class="InstallationModal_Loader"
                                            :class="{ active: selectedLoader === l.id, unavailable: !isLoaderSupported(l.id) }"
                                            :disabled="!isLoaderSupported(l.id)"
                                            :title="compatOf(l.id)?.status === 'empty' || compatOf(l.id)?.status === 'error'
                                                ? selectedVersion ? `No disponible para ${selectedVersion}` : 'No disponible'
                                                : `${l.name} · ${l.tag}`"
                                            @click="selectedLoader = l.id"
                                        >
                                            <img :src="l.icon" :alt="l.name" draggable="false" />
                                            <span class="InstallationModal_LoaderName">
                                                {{ l.name }}
                                                <em v-if="compatOf(l.id)?.status === 'empty'" class="InstallationModal_LoaderLack">No disponible para {{ selectedVersion }}</em>
                                                <em v-else-if="compatOf(l.id)?.status === 'error'" class="InstallationModal_LoaderLack">No se pudo verificar {{ selectedVersion }}</em>
                                                <em v-else>{{ l.tag }}</em>
                                            </span>
                                            <span v-if="compatOf(l.id)?.status === 'loading'" class="InstallationModal_LoaderSpinner">
                                                <span class="InstallationModal_Spinner" aria-hidden="true"></span>
                                            </span>
                                            <IconCheck v-else-if="isLoaderSupported(l.id) && selectedLoader === l.id" class="InstallationModal_LoaderCheck" stroke="2.4" />
                                        </button>
                                    </div>

                                    <div v-if="selectedVersion && (selectedLoaderStatus === 'empty' || selectedLoaderStatus === 'error')" class="InstallationModal_Unavailable">
                                        <IconAlertTriangle />
                                        <span>{{ selectedLoaderHint }}</span>
                                    </div>

                                    <label class="InstallationModal_Toggle">
                                        <span class="InstallationModal_ToggleBox">
                                            <input v-model="manualLoaderVersion" type="checkbox" />
                                            <span class="InstallationModal_ToggleTrack"><span></span></span>
                                        </span>
                                        <span class="InstallationModal_ToggleText">
                                            <strong>Elegir versión del modloader</strong>
                                            <em>Desactívalo para usar la recomendada automáticamente</em>
                                        </span>
                                    </label>

                                    <template v-if="selectedLoader !== 'vanilla' && selectedVersion">
                                        <template v-if="manualLoaderVersion">
                                            <div v-if="selectedLoaderStatus === 'loading'" class="InstallationModal_LoaderVersion loading">
                                                <span class="InstallationModal_Spinner" aria-hidden="true"></span>
                                                <span>Buscando versiones de {{ selectedLoader }}…</span>
                                            </div>
                                            <div v-else-if="selectedLoaderStatus === 'empty'" class="InstallationModal_LoaderVersion empty">
                                                <IconAlertTriangle />
                                                <span>No hay versiones de {{ selectedLoader }} para {{ selectedVersion }}.</span>
                                            </div>
                                            <div v-else-if="selectedLoaderStatus === 'error'" class="InstallationModal_LoaderVersion empty">
                                                <IconAlertTriangle />
                                                <span>No se pudo consultar {{ selectedLoader }} para {{ selectedVersion }}.</span>
                                                <button class="SsBtn" @click="retrySelectedLoader">Reintentar</button>
                                            </div>
                                            <div v-else class="InstallationModal_LoaderVersion">
                                                <span class="InstallationModal_Label inner">Todas las versiones</span>
                                                <select
                                                    v-model="loaderVersion"
                                                    class="SsSel InstallationModal_Select"
                                                >
                                                    <option v-for="lv in loaderVersions" :key="lv.loaderVersion" :value="lv.loaderVersion">
                                                        {{ lv.loaderVersion }}{{ lv.stable ? '  (recomendada)' : '' }}
                                                    </option>
                                                </select>
                                            </div>
                                        </template>
                                        <template v-else>
                                            <div v-if="selectedLoaderStatus === 'loading'" class="InstallationModal_LoaderVersion loading">
                                                <span class="InstallationModal_Spinner" aria-hidden="true"></span>
                                                <span>Comprobando {{ selectedLoader }} para {{ selectedVersion }}…</span>
                                            </div>
                                            <div v-else-if="selectedLoaderStatus === 'ok'" class="InstallationModal_AutoVersion">
                                                <IconCheck stroke="2.2" />
                                                <span>Se usará la versión recomendada <strong>{{ autoLoaderResolved }}</strong> automáticamente.</span>
                                            </div>
                                        </template>
                                    </template>
                                </section>
                            </div>
                        </template>

                        <template v-else-if="phase === 'installing'">
                            <div class="InstallationModal_Progress">
                                <div class="InstallationModal_StateRow">
                                    <span class="InstallationModal_StateVersion">{{ selectedVersion }}<template v-if="selectedLoader !== 'vanilla'"> · {{ selectedLoader }} {{ effectiveLoaderVersion }}</template></span>
                                </div>

                                <template v-if="loaderPhase !== 'running'">
                                <div class="InstallationModal_Ring" :class="ringClass">
                                    <svg viewBox="0 0 120 120" class="InstallationModal_RingSvg" aria-hidden="true">
                                        <circle class="InstallationModal_RingTrack" cx="60" cy="60" r="52" />
                                        <circle
                                            class="InstallationModal_RingFill"
                                            cx="60"
                                            cy="60"
                                            r="52"
                                            :stroke-dasharray="RING_CIRC"
                                            :stroke-dashoffset="ringOffset"
                                            transform="rotate(-90 60 60)"
                                        />
                                    </svg>
                                    <span v-if="ringMode === 'percent'" class="InstallationModal_RingPct">
                                        {{ fmt(clampPct(progress?.percent ?? 0), 0) }}<em>%</em>
                                    </span>
                                    <span v-else-if="ringMode === 'busy'" class="InstallationModal_RingCenter">
                                        <span class="InstallationModal_Spinner" aria-hidden="true"></span>
                                    </span>
                                    <span v-else class="InstallationModal_RingCenter done">
                                        <IconCheck stroke="2.6" />
                                    </span>
                                </div>

                                <div class="InstallationModal_Step">
                                    <span class="InstallationModal_StepBadge" :class="stateBadge(downloadState).tone">
                                        <span class="dot"></span>
                                        {{ stateBadge(downloadState).text }}
                                    </span>
                                    <span v-if="stepHint" class="InstallationModal_StepHint">{{ stepHint }}</span>
                                </div>

                                <div v-if="downloadState === 'verifying' || downloadState === 'redownloading'" class="InstallationModal_VerifyCount">
                                    <span v-if="progress?.filesTotal">
                                        {{ fmt(Math.min(progress.filesDownloaded, progress.filesTotal), 0) }}/{{ fmt(progress.filesTotal, 0) }}
                                        archivos verificados
                                    </span>
                                    <span v-else>Preparando la verificación…</span>
                                </div>

                                <div v-if="ringMode === 'percent'" class="InstallationModal_FriendlyLine">
                                    <span>Descargado <strong>{{ fmtMb(progress?.mbDownloaded ?? 0) }}</strong></span>
                                    <template v-if="progress?.mbTotal && progress.mbTotal > 0">
                                        <span>de {{ fmtMb(progress.mbTotal) }}</span>
                                    </template>
                                    <span>·</span>
                                    <span>{{ fmtSpeed(progress?.speedMbps ?? 0) }}</span>
                                </div>

                                <button class="InstallationModal_DetailsToggle" :class="{ open: detailsOpen }" @click="detailsOpen = !detailsOpen">
                                    <IconChevronDown stroke="2" />
                                    {{ detailsOpen ? 'Ocultar detalles' : 'Detalles técnicos' }}
                                </button>

                                <div v-if="detailsOpen" class="InstallationModal_Details">
                                    <div class="InstallationModal_DetailStats">
                                        <span class="InstallationModal_DetailStat">
                                            <b>{{ progress!.filesDownloaded }}/{{ progress!.filesTotal }}</b>
                                            <i>archivos</i>
                                        </span>
                                        <span class="InstallationModal_DetailStat">
                                            <b>{{ fmtMb(progress!.mbDownloaded) }}/{{ fmtMb(progress!.mbTotal || 0) }}</b>
                                            <i>megabytes</i>
                                        </span>
                                        <span class="InstallationModal_DetailStat">
                                            <b>{{ progress!.queuedCount }}</b>
                                            <i>en cola</i>
                                        </span>
                                        <span class="InstallationModal_DetailStat">
                                            <b>{{ fmtSpeed(progress!.speedMbps) }}</b>
                                            <i>velocidad</i>
                                        </span>
                                    </div>

                                    <div v-if="progress?.currentFile" class="InstallationModal_DetailBlock">
                                        <div class="InstallationModal_DetailTitle">Descargando</div>
                                        <div class="InstallationModal_DetailCurrent">
                                            <span class="InstallationModal_DetailCurrentName">{{ progress!.currentFile }}</span>
                                            <div class="InstallationModal_DetailBar">
                                                <div class="InstallationModal_DetailBarFill" :style="{ width: clampPct(progress!.currentProgress) + '%' }"></div>
                                            </div>
                                            <span class="InstallationModal_DetailCurrentMeta">
                                                <span>{{ sectionLabel(progress!.currentSection) || '—' }}</span>
                                                <em>{{ fmt(clampPct(progress!.currentProgress), 0) }}%</em>
                                            </span>
                                        </div>
                                    </div>

                                    <div v-if="(progress?.sections ?? []).length" class="InstallationModal_DetailBlock">
                                        <div class="InstallationModal_DetailTitle">
                                            Progreso por elemento<em>{{ progress!.sectionsCompleted.length }}/{{ progress!.sections.length }}</em>
                                        </div>
                                        <div class="InstallationModal_DetailList">
                                            <div
                                                v-for="s in progress!.sections"
                                                :key="s.name"
                                                class="InstallationModal_DetailRow"
                                                :class="{ done: s.totalFiles > 0 && s.doneFiles >= s.totalFiles }"
                                            >
                                                <div class="InstallationModal_DetailRowHead">
                                                    <span class="InstallationModal_DetailRowLabel">{{ sectionLabel(s.name) }}</span>
                                                    <em>{{ s.totalFiles ? fmt(s.doneFiles / s.totalFiles * 100, 0) : 0 }}%</em>
                                                </div>
                                                <div class="InstallationModal_DetailBar">
                                                    <div
                                                        class="InstallationModal_DetailBarFill"
                                                        :style="{ width: s.totalFiles ? clampPct(s.doneFiles / s.totalFiles * 100) + '%' : '0%' }"
                                                    ></div>
                                                </div>
                                                <span class="InstallationModal_DetailRowMeta">{{ s.doneFiles }}/{{ s.totalFiles }} archivos · {{ fmtMb(s.mbDownloaded) }}/{{ fmtMb(s.mbTotal) }}</span>
                                            </div>
                                        </div>
                                    </div>

                                    <div v-if="(progress?.activeFiles ?? []).length" class="InstallationModal_DetailBlock">
                                        <div class="InstallationModal_DetailTitle">Descargando ahora</div>
                                        <div class="InstallationModal_DetailList">
                                            <div v-for="f in progress!.activeFiles.slice(0, 4)" :key="f.name" class="InstallationModal_DetailFile">
                                                <div class="InstallationModal_DetailFileHead">
                                                    <span class="name">{{ f.name }}</span>
                                                    <em>{{ fmt(clampPct(f.percent), 0) }}%</em>
                                                </div>
                                                <div class="InstallationModal_DetailFileBar">
                                                    <div class="InstallationModal_DetailFileBarFill" :style="{ width: clampPct(f.percent) + '%' }"></div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>

                                    <div v-if="(progress?.queuedCount ?? 0) > 0" class="InstallationModal_DetailBlock">
                                        <div class="InstallationModal_DetailTitle">En cola · {{ progress!.queuedCount }}</div>
                                        <div class="InstallationModal_DetailChips">
                                            <span v-for="name in (progress?.queuedPreview ?? []).slice(0, 5)" :key="name" class="InstallationModal_DetailChip">{{ name }}</span>
                                            <span v-if="(progress?.queuedPreview ?? []).length < progress!.queuedCount" class="InstallationModal_DetailChip more">
                                                +{{ progress!.queuedCount - (progress?.queuedPreview ?? []).length }}
                                            </span>
                                        </div>
                                    </div>

                                    <div v-if="(progress?.log ?? []).length" class="InstallationModal_DetailBlock">
                                        <div class="InstallationModal_DetailTitle">Registro</div>
                                        <div class="InstallationModal_DetailLog">
                                            <span v-for="(line, li) in (progress?.log ?? []).slice(-10)" :key="li">{{ line }}</span>
                                        </div>
                                    </div>
                                </div>

                                </template>

                                <template v-else>
                                    <div class="InstallationModal_LoaderWarn">
                                        <IconAlertTriangle stroke="2" />
                                        <div>
                                            <strong>Instalador independiente y de consumo alto</strong>
                                            <span>El instalador de {{ selectedLoader }} se ejecuta como proceso aparte y puede consumir mucha memoria y CPU (Forge/NeoForge pueden pasar de los 600 MB). Ten paciencia: deja que instale todo y no cierres la aplicación.</span>
                                        </div>
                                    </div>

                                    <div class="InstallationModal_LoaderStatus running">
                                        <span class="InstallationModal_Spinner" aria-hidden="true"></span>
                                        <div class="InstallationModal_LoaderStatusText">
                                            <span class="title">Instalando {{ selectedLoader }} {{ effectiveLoaderVersion }}</span>
                                            <span class="detail">{{ loaderStatus }}</span>
                                        </div>
                                    </div>

                                    <div class="InstallationModal_LoaderBody">
                                        <div v-if="loaderTotal > 0" class="InstallationModal_LoaderBar">
                                            <div
                                                class="InstallationModal_LoaderBarFill"
                                                :style="{ width: clampPct(loaderTotal ? (loaderProgress / loaderTotal) * 100 : 0) + '%' }"
                                            ></div>
                                            <em>{{ fmt(clampPct(loaderTotal ? (loaderProgress / loaderTotal) * 100 : 0), 0) }}%</em>
                                        </div>
                                        <div v-if="loaderTotal > 0" class="InstallationModal_LoaderParts">
                                            <span>Descargando parte {{ fmt(loaderProgress, 0) }} de {{ fmt(loaderTotal, 0) }}</span>
                                        </div>
                                        <div ref="loaderLogEl" class="InstallationModal_LoaderLog">
                                            <span v-for="(line, li) in loaderLogs" :key="li">{{ line }}</span>
                                        </div>
                                    </div>
                                </template>

                                <div v-if="errorMessage" class="InstallationModal_Notice error">
                                    <IconAlertTriangle />
                                    <span>{{ errorMessage }}</span>
                                </div>
                            </div>
                        </template>

                        <template v-else-if="phase === 'done'">
                            <div class="InstallationModal_Done">
                                <span class="InstallationModal_DoneIcon"><IconCheck stroke="2.6" /></span>
                                <h4>
                                    {{ loaderPhase === 'done' ? '¡Listo!' : 'Instalación completada' }}
                                </h4>
                                <p>
                                    <template v-if="loaderPhase === 'done'">
                                        {{ selectedVersion }} con {{ selectedLoader }} {{ effectiveLoaderVersion }} está instalado y listo para jugar.
                                    </template>
                                    <template v-else>
                                        La versión {{ selectedVersion }} se descargó correctamente.
                                    </template>
                                </p>
                                <div v-if="finalSummary" class="InstallationModal_Summary">
                                    <div>
                                        <span class="k">Archivos descargados</span>
                                        <span class="v">{{ finalSummary.files }}</span>
                                    </div>
                                    <div>
                                        <span class="k">Datos transferidos</span>
                                        <span class="v">{{ fmtMb(finalSummary.mb) }}</span>
                                    </div>
                                </div>
                            </div>
                        </template>

                        <template v-else>
                            <div class="InstallationModal_Error">
                                <span class="InstallationModal_ErrorIcon"><IconAlertTriangle stroke="2.4" /></span>
                                <h4>Algo salió mal</h4>
                                <p>{{ errorMessage }}</p>
                            </div>
                        </template>
                    </div>

                    <div class="InstallationModal_Footer">
                        <template v-if="phase === 'setup'">
                            <button class="SsBtn" @click="closeModal">Cancelar</button>
                            <button
                                class="SsBtn SsBtnPrimary"
                                :disabled="!installReady"
                                @click="onInstall"
                            >
                                <IconDownload stroke="2" />
                                {{ selectedLoader === 'vanilla' ? 'Instalar' : `Instalar con ${selectedLoader}` }}
                            </button>
                        </template>

                        <template v-else-if="phase === 'installing'">
                            <template v-if="!gameDone">
                                <button
                                    v-if="downloadState === 'downloading'"
                                    class="SsBtn"
                                    @click="onPause"
                                >
                                    <IconPlayerPause stroke="2" />
                                    Pausar
                                </button>
                                <button
                                    v-else-if="downloadState === 'paused'"
                                    class="SsBtn"
                                    @click="onResume"
                                >
                                    <IconRotateClockwise stroke="2" />
                                    Reanudar
                                </button>
                                <button class="SsBtn SsBtnDanger" @click="onCancel">
                                    <IconPlayerStop stroke="2" />
                                    Cancelar
                                </button>
                            </template>
                            <template v-else-if="loaderPhase === 'running'">
                                <span class="InstallationModal_FooterNote">
                                    Instalando {{ selectedLoader }} {{ effectiveLoaderVersion }} · {{ loaderStatus }}
                                </span>
                                <button class="SsBtn" @click="closeModal">Cerrar</button>
                            </template>
                            <template v-else>
                                <button class="SsBtn SsBtnPrimary" @click="closeModal">
                                    <IconCheck stroke="2" />
                                    Listo
                                </button>
                            </template>
                        </template>

                        <template v-else>
                            <button class="SsBtn" @click="onCancel">Nueva instalación</button>
                            <button class="SsBtn SsBtnPrimary" @click="closeModal">Cerrar</button>
                        </template>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use './Styles/Installation.scss';
</style>
