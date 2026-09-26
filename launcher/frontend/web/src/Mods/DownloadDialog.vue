<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import {
    IconX, IconDownload, IconPuzzle, IconLoader2, IconRotateClockwise, IconFile,
    IconCheck, IconAlertTriangle, IconPlayerStop,
} from '@tabler/icons-vue';
import { Events } from '@wailsio/runtime';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';
import {
    fetchProjectVersions, formatBytes, formatDate, formatCount, useModrinth,
} from '@/Common/Composables/useModrinth';
import type { ModrinthVersion } from '@/Common/Composables/useModrinth';
import { VERSION_TYPES } from './Store';
import { pendingInstance } from './Store';
import { ListInstances } from '@wailsjs/StepLauncher/internal/Services/Instance/instanceservice';
import {
    GetInstalledInstanceModLoader, GetInstalledModLoader,
} from '@wailsjs/StepLauncher/internal/Services/ModLoader/modloaderservice';
import {
    GlobalGameDir, InstanceGameDir, InstallModContent, InstallModpack, CancelModContent,
} from '@wailsjs/StepLauncher/internal/Services/Mods/modsservice';
import { loadInstances, ensureContentEvents, registerContentMeta, latestLoaderFeed } from '@/Instances/Store';

const props = defineProps<{
    slugOrId: string;
    title: string;
    iconUrl?: string | null;
    projectType?: string;
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
}>();

const modrinthTags = useModrinth();

const versions = ref<ModrinthVersion[]>([]);
const loading = ref(false);
const error = ref('');

const mcVersion = ref('');
const loader = ref('');
const showSnapshots = ref(false);
const channel = ref<'any' | 'release' | 'beta' | 'alpha'>('any');

const mcOptionsFull = computed(() => {
    const set = new Set<string>();
    for (const v of versions.value) for (const gv of v.game_versions) set.add(gv);
    return [...set].sort((a, b) => b.localeCompare(a));
});

const mcOptions = computed(() => {
    if (showSnapshots.value) return mcOptionsFull.value;
    // Sin snapshots: solo releases
    if (modrinthTags.tagsState.gameVersions.length) {
        const releaseSet = new Set(
            modrinthTags.tagsState.gameVersions
                .filter((g) => g.version_type === 'release')
                .map((g) => g.version)
        );
        const filtered = mcOptionsFull.value.filter((v) => releaseSet.has(v));
        // Fallback heurístico si ningún release coincide (ej. datos aún no clasificados)
        if (filtered.length) return filtered;
    }
    return mcOptionsFull.value.filter((v) => /^\d+(\.\d+)+$/.test(v));
});

const loaderOptions = computed(() => {
    const set = new Set<string>();
    for (const v of versions.value) for (const l of v.loaders) set.add(l);
    return [...set];
});

const isSingleLoader = computed(() => loaderOptions.value.length <= 1);

// La versión concreta del mod que coincide con MC + loader + canal elegidos:
const matchedVersion = computed(() => {
    if (!mcVersion.value || !loader.value) return null;
    const candidates = versions.value
        .filter((v) => v.game_versions.includes(mcVersion.value) && v.loaders.includes(loader.value))
        .filter((v) => channel.value === 'any' || v.version_type === channel.value)
        .sort((a, b) => {
            const rank = { release: 0, beta: 1, alpha: 2 } as const;
            const r = rank[a.version_type as keyof typeof rank] - rank[b.version_type as keyof typeof rank];
            if (r !== 0) return r;
            return b.date_published.localeCompare(a.date_published);
        });
    return candidates[0] ?? null;
});

const primaryFile = computed(() => {
    const v = matchedVersion.value;
    if (!v) return null;
    return v.files.find((f) => f.primary) ?? v.files[0] ?? null;
});

// ---------- Destino de la instalación ----------

const isModpack = computed(() => (props.projectType ?? '').toLowerCase() === 'modpack');

const SUBDIR_BY_TYPE: Record<string, string> = {
    mod: 'mods',
    resourcepack: 'resourcepacks',
    shader: 'shaderpacks',
};

const destSubdir = computed(() => SUBDIR_BY_TYPE[(props.projectType ?? '').toLowerCase()] ?? '');

type DestKind = 'global' | 'instance';

const destination = ref<DestKind>('global');
// Destino del modpack: nueva instancia (con su icono), una existente o el
// juego global. Nada se impone: el usuario elige dónde va cada pack.
type ModpackDest = 'new' | 'instance' | 'global';

const mpDestination = ref<ModpackDest>('new');
const instances = ref<Array<{ name: string; title: string }>>([]);
const loadingInstances = ref(false);
const selectedInstance = ref('');
const globalGameDir = ref('');
const instanceGameDir = ref('');
const modpackName = ref('');
const loaderHint = ref('');
const checkingLoader = ref(false);

const destPath = computed(() => {
    const base = destination.value === 'instance' ? instanceGameDir.value : globalGameDir.value;
    if (!base) return '';
    return destSubdir.value ? `${base}/${destSubdir.value}` : base;
});

function sanitizeInstanceName(raw: string): string {
    const clean = raw.trim().replace(/[/\\]/g, '-').replace(/\s+/g, ' ').trim();
    return clean.slice(0, 48) || 'modpack';
}

async function refreshInstances(): Promise<void> {
    loadingInstances.value = true;
    try {
        const list = await ListInstances();
        instances.value = (Array.isArray(list) ? list : [])
            .filter((i) => i !== null)
            .map((i) => ({ name: String((i as any).name ?? ''), title: String((i as any).title ?? (i as any).name ?? '') }))
            .filter((i) => i.name);
        if (!selectedInstance.value && instances.value.length) {
            selectedInstance.value = instances.value[0]!.name;
        }
    } catch {
        instances.value = [];
    } finally {
        loadingInstances.value = false;
    }
}

async function refreshGlobalDir(): Promise<void> {
    try {
        globalGameDir.value = await GlobalGameDir();
    } catch {
        globalGameDir.value = '';
    }
}

async function refreshInstanceDir(): Promise<void> {
    instanceGameDir.value = '';
    if (destination.value !== 'instance' || !selectedInstance.value) return;
    try {
        instanceGameDir.value = await InstanceGameDir(selectedInstance.value);
    } catch {
        instanceGameDir.value = '';
    }
}

// Aviso informativo (no bloqueante) si el destino no tiene modloader: los
// .jar de mods solo cargan con Fabric/Forge/Quilt/NeoForge instalado.
async function refreshLoaderHint(): Promise<void> {
    loaderHint.value = '';
    if (isModpack.value || (props.projectType ?? '').toLowerCase() !== 'mod') return;
    if (destination.value === 'instance' && !selectedInstance.value) return;
    checkingLoader.value = true;
    try {
        const info = destination.value === 'instance'
            ? await GetInstalledInstanceModLoader(selectedInstance.value)
            : await GetInstalledModLoader(globalGameDir.value || '.');
        if (!info) {
            loaderHint.value = destination.value === 'instance'
                ? 'Esta instancia no tiene modloader: el mod se guardará pero no cargará hasta instalar Fabric, Forge, Quilt o NeoForge.'
                : 'El juego global no tiene modloader detectado: el mod se guardará pero no cargará hasta instalar un loader.';
        }
    } catch {
        // Si no se puede comprobar, no se molesta: la descarga sigue igual.
    } finally {
        checkingLoader.value = false;
    }
}

watch(destination, () => {
    void refreshInstanceDir().then(() => void refreshLoaderHint());
});

watch(selectedInstance, () => {
    void refreshInstanceDir().then(() => void refreshLoaderHint());
});

// ---------- Instalación (progreso por eventos del backend) ----------

type Phase = 'pick' | 'working' | 'done' | 'error';

const phase = ref<Phase>('pick');
const workMessage = ref('');
const workProgress = ref(0);
const workTotal = ref(0);
// Fases vistas en esta sesión (para la lista de pasos del modpack).
const seenWork = ref<string[]>([]);
const doneMessage = ref('');
const sessionId = ref('');

function markSeen(key: string): void {
    if (!seenWork.value.includes(key)) seenWork.value = [...seenWork.value, key];
}

// Pasos del modpack deducidos de las fases vistas: pack → Minecraft →
// loader (solo si el pack lo exige, es decir, si se vio instalando) →
// archivos → listo.
const modpackSteps = computed(() => {
    const seen = seenWork.value;
    const finished = phase.value === 'done';
    const working = phase.value === 'working';
    const hasLoader = seen.includes('installing');
    const hasFiles = seen.includes('files');
    const packDone = finished || seen.includes('downloading') || hasLoader || hasFiles;
    const mcDone = finished || hasLoader || hasFiles;
    const loaderDone = finished || hasFiles;
    const steps = [
        { id: 'pack', label: 'Modpack', done: packDone, active: working && !packDone },
        { id: 'mc', label: 'Minecraft', done: mcDone, active: working && packDone && !mcDone },
        { id: 'files', label: 'Archivos', done: finished, active: working && hasFiles && !finished },
    ];
    if (hasLoader || working) {
        steps.splice(2, 0, { id: 'loader', label: 'Modloader', done: loaderDone, active: working && hasLoader && !hasFiles });
    }
    return steps;
});

// Progreso del loader incrustado: mientras el paso es Modloader se muestra su
// mensaje y porcentaje reales (vienen de los eventos modloader_* globales).
const loaderStep = computed(() => modpackSteps.value.find((s) => s.id === 'loader' && s.active) ?? null);

const loaderFeedLine = computed(() => {
    if (!loaderStep.value) return null;
    const feed = latestLoaderFeed();
    if (!feed || !feed.message) return null;
    const pct = feed.total > 0 ? Math.min(100, Math.round((feed.progress / feed.total) * 100)) : null;
    return { message: feed.message, pct };
});

function close(): void {
    emit('update:visible', false);
}

function retry(): void {
    void loadVersions();
}

async function loadVersions(): Promise<void> {
    if (loading.value) return;
    loading.value = true;
    error.value = '';
    try {
        versions.value = await fetchProjectVersions(props.slugOrId, { includeChangelog: false });
        const uniqueFull = [...new Set(versions.value.flatMap((v) => v.game_versions))].sort((a, b) => b.localeCompare(a));
        // Elegir MC inicial respetando snapshot flag (por defecto solo releases)
        let initialMc = '';
        if (showSnapshots.value) {
            initialMc = uniqueFull[0] ?? '';
        } else {
            const releaseOpts = mcOptions.value;
            initialMc = releaseOpts[0] ?? uniqueFull[0] ?? '';
        }
        mcVersion.value = initialMc;
        const loaders = [...new Set(versions.value.flatMap((v) => v.loaders))];
        loader.value = loaders[0] ?? '';
        // Canal por defecto: si hay releases, priorizar release, si no cualquiera
        const hasRelease = versions.value.some((v) => v.version_type === 'release');
        channel.value = hasRelease ? 'any' : 'any';
    } catch (err) {
        error.value = (err as Error)?.message ?? 'No se pudieron cargar las versiones';
    } finally {
        loading.value = false;
    }
}

function parseEvent(raw: unknown): any {
    try {
        const s = typeof raw === 'string' ? raw : JSON.stringify(raw ?? '');
        return JSON.parse(s);
    } catch {
        return null;
    }
}

function isMine(e: any): boolean {
    return !!e && (!e.sessionId || !sessionId.value || e.sessionId === sessionId.value);
}

function onModContentEvent(raw: unknown): void {
    const e = parseEvent(raw);
    if (!isMine(e)) return;
    switch (e?.type) {
        case 'modcontent_resolving':
            workMessage.value = `Preparando la instalación en ${e.dest ?? 'el destino'}…`;
            break;
        case 'modcontent_downloading':
            workMessage.value = `Descargando ${e.file ?? 'el archivo'}…`;
            break;
        case 'modcontent_installed':
            phase.value = 'done';
            doneMessage.value = e.dest ? `Instalado en ${e.dest}` : 'Instalación completada';
            break;
        case 'modcontent_error':
            phase.value = 'error';
            error.value = e.error ?? 'No se pudo instalar el archivo.';
            break;
    }
}

function onModpackEvent(raw: unknown): void {
    const e = parseEvent(raw);
    if (!isMine(e)) return;
    switch (e?.type) {
        case 'modpack_resolving':
            markSeen('resolving');
            workMessage.value = e.message ?? 'Instalando el modpack…';
            break;
        case 'modpack_downloading':
            markSeen('downloading');
            if (Number(e.total ?? 0) > 1) markSeen('files');
            workMessage.value = e.message ?? 'Instalando el modpack…';
            workProgress.value = Number(e.progress ?? 0);
            workTotal.value = Number(e.total ?? 0);
            break;
        case 'modpack_installing':
            markSeen('installing');
            workMessage.value = e.message ?? 'Instalando el modpack…';
            workProgress.value = Number(e.progress ?? 0);
            workTotal.value = Number(e.total ?? 0);
            break;
        case 'modpack_installed':
            phase.value = 'done';
            doneMessage.value = (typeof e.message === 'string' && e.message)
                || (e.instance ? `Modpack instalado en la instancia ${e.instance}` : 'Modpack instalado');
            void loadInstances();
            break;
        case 'modpack_error':
            phase.value = 'error';
            error.value = e.error ?? 'No se pudo instalar el modpack.';
            break;
    }
}

async function onDownload(): Promise<void> {
    const file = primaryFile.value;
    if (!file) return;
    const hashes = (file as any).hashes ?? {};
    phase.value = 'working';
    error.value = '';
    doneMessage.value = '';
    workMessage.value = 'Iniciando la descarga…';
    workProgress.value = 0;
    workTotal.value = 0;
    seenWork.value = [];
    try {
        if (isModpack.value) {
            if (mpDestination.value === 'instance' && !selectedInstance.value) {
                phase.value = 'error';
                error.value = 'Elige la instancia de destino.';
                return;
            }
            const res = await InstallModpack({
                title: props.title,
                fileUrl: file.url,
                fileName: file.filename,
                fileSize: file.size,
                hashSha1: hashes.sha1 ?? '',
                hashSha512: hashes.sha512 ?? '',
                destination: mpDestination.value,
                instance: selectedInstance.value,
                instanceName: sanitizeInstanceName(modpackName.value || props.title),
                iconUrl: props.iconUrl ?? '',
            } as any);
            sessionId.value = res?.sessionId ?? '';
            if (sessionId.value) {
                registerContentMeta(sessionId.value, {
                    iconUrl: props.iconUrl ?? '',
                    title: props.title,
                    dest: mpDestination.value === 'new'
                        ? `nueva instancia ${sanitizeInstanceName(modpackName.value || props.title)}`
                        : mpDestination.value === 'instance'
                            ? `instancia ${selectedInstance.value}`
                            : 'juego global',
                });
            }
        } else {
            if (destination.value === 'instance' && !selectedInstance.value) {
                phase.value = 'error';
                error.value = 'Elige la instancia de destino.';
                return;
            }
            const res = await InstallModContent({
                projectType: props.projectType ?? 'mod',
                title: props.title,
                fileUrl: file.url,
                fileName: file.filename,
                fileSize: file.size,
                hashSha1: hashes.sha1 ?? '',
                hashSha512: hashes.sha512 ?? '',
                iconUrl: props.iconUrl ?? '',
                destination: destination.value,
                instance: selectedInstance.value,
            } as any);
            sessionId.value = res?.sessionId ?? '';
            if (sessionId.value) {
                registerContentMeta(sessionId.value, {
                    iconUrl: props.iconUrl ?? '',
                    title: props.title,
                    dest: destination.value === 'instance' ? `instancia ${selectedInstance.value}` : 'juego global',
                });
            }
        }
    } catch (err) {
        phase.value = 'error';
        error.value = (err as Error)?.message ?? 'No se pudo iniciar la descarga.';
    }
}

async function onCancelWork(): Promise<void> {
    if (sessionId.value) {
        try {
            await CancelModContent(sessionId.value);
        } catch (_e) {}
    }
    sessionId.value = '';
    phase.value = 'pick';
    workMessage.value = '';
}

function resetDialog(): void {
    phase.value = 'pick';
    error.value = '';
    doneMessage.value = '';
    workMessage.value = '';
    workProgress.value = 0;
    workTotal.value = 0;
    seenWork.value = [];
    sessionId.value = '';
}

watch(
    () => props.slugOrId,
    () => {
        if (!props.slugOrId) return;
        versions.value = [];
        mcVersion.value = '';
        loader.value = '';
        channel.value = 'any';
        modpackName.value = props.title ?? '';
        resetDialog();
        void loadVersions();
    },
    { immediate: true }
);

watch(showSnapshots, () => {
    // Si el MC actual ya no está en la lista filtrada, saltar al primero disponible
    if (!mcOptions.value.includes(mcVersion.value)) {
        mcVersion.value = mcOptions.value[0] ?? mcOptionsFull.value[0] ?? '';
    }
});

watch(mcOptions, (opts) => {
    if (!opts.includes(mcVersion.value) && opts.length) {
        mcVersion.value = opts[0]!;
    }
});

watch(
    () => props.visible,
    (v) => {
        if (!v) return;
        initDialog();
    }
);

// El Host monta el diálogo ya visible: el watcher no se dispara al abrir,
// así que la carga inicial va en onMounted (el watcher queda por seguridad).
function initDialog(): void {
    resetDialog();
    modpackName.value = props.title ?? '';
    void refreshGlobalDir();
    void refreshInstances().then(() => {
        // Si se abrió desde una instancia ("Añadir"), preseleccionarla.
        if (pendingInstance.value && instances.value.some((i) => i.name === pendingInstance.value)) {
            selectedInstance.value = pendingInstance.value!;
            if (isModpack.value) mpDestination.value = 'instance';
            else destination.value = 'instance';
        }
        return void refreshInstanceDir().then(() => void refreshLoaderHint());
    });
}

function onCloseOverlays() {
    close();
}

useOverlayEscape(close, { priority: 2, isActive: () => props.visible });

let eventOffs: Array<() => void> = [];

onMounted(() => {
    void modrinthTags.loadTags();
    // El store sigue la sesión en el widget aunque este diálogo se cierre.
    ensureContentEvents();
    // Carga inicial (el Host monta ya visible: el watcher no se dispara).
    initDialog();
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    eventOffs = [
        Events.On('modcontent_resolving', ({ data: raw }: any) => onModContentEvent(raw)),
        Events.On('modcontent_downloading', ({ data: raw }: any) => onModContentEvent(raw)),
        Events.On('modcontent_installed', ({ data: raw }: any) => onModContentEvent(raw)),
        Events.On('modcontent_error', ({ data: raw }: any) => onModContentEvent(raw)),
        Events.On('modpack_resolving', ({ data: raw }: any) => onModpackEvent(raw)),
        Events.On('modpack_downloading', ({ data: raw }: any) => onModpackEvent(raw)),
        Events.On('modpack_installing', ({ data: raw }: any) => onModpackEvent(raw)),
        Events.On('modpack_installed', ({ data: raw }: any) => onModpackEvent(raw)),
        Events.On('modpack_error', ({ data: raw }: any) => onModpackEvent(raw)),
    ];
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    eventOffs.forEach((off) => off());
    eventOffs = [];
});
</script>

<template>
    <Teleport to="body">
        <Transition name="ModsDl">
            <div v-if="visible" class="ModsDl_Overlay" @click.self="close">
                <div class="ModsDl_Card">
            <header class="ModsDl_Head">
                <div class="ModsDl_Title">
                    <span class="ModsDl_Icon">
                        <img v-if="iconUrl" :src="iconUrl" alt="" />
                        <IconPuzzle v-else stroke="1.5" />
                    </span>
                    <div class="ModsDl_Titles">
                        <h3>Descargar {{ title }}</h3>
                        <p>Elige la versión de Minecraft y el modloader</p>
                    </div>
                </div>
                <button class="ModsDl_Close" title="Cerrar" @click="close">
                    <IconX stroke="2" />
                </button>
            </header>

            <div class="ModsDl_Body">
                <p v-if="loading" class="ModsDl_State">
                    <IconLoader2 class="spin" stroke="2" /> Cargando versiones…
                </p>

                <div v-else-if="error && phase !== 'working'" class="ModsDl_State error">
                    <b>{{ phase === 'error' ? 'No se pudo instalar' : 'No se pudieron cargar las versiones' }}</b>
                    <span>{{ error }}</span>
                    <button class="SsBtn SsBtnPrimary" @click="retry">
                        <IconRotateClockwise stroke="2" /> Reintentar
                    </button>
                </div>

                <div v-else-if="phase === 'working'" class="ModsDl_Work">
                    <p class="ModsDl_State">
                        <IconLoader2 class="spin" stroke="2" /> {{ workMessage || 'Instalando…' }}
                    </p>
                    <ol v-if="isModpack" class="ModsDl_Steps">
                        <li v-for="s in modpackSteps" :key="s.id" :class="{ done: s.done, active: s.active }">
                            <span class="ModsDl_StepDot">
                                <IconCheck v-if="s.done" stroke="2.6" />
                                <IconLoader2 v-else-if="s.active" class="spin" stroke="2" />
                            </span>
                            {{ s.label }}
                        </li>
                    </ol>
                    <p v-if="workTotal > 0" class="ModsDl_WorkCount">
                        {{ Math.min(workProgress, workTotal) }}/{{ workTotal }} archivos
                    </p>
                    <div v-if="workTotal > 0" class="ModsDl_Bar">
                        <div
                            class="ModsDl_BarFill"
                            :style="{ width: Math.min(100, Math.max(0, (workProgress / Math.max(1, workTotal)) * 100)) + '%' }"
                        ></div>
                    </div>
                    <p v-if="loaderFeedLine" class="ModsDl_LoaderFeed" :title="loaderFeedLine.message">
                        {{ loaderFeedLine.message }}<template v-if="loaderFeedLine.pct !== null"> · {{ loaderFeedLine.pct }}%</template>
                    </p>
                    <p class="ModsDl_WorkHint">Puedes cerrar este diálogo: la instalación sigue en segundo plano.</p>
                </div>

                <div v-else-if="phase === 'done'" class="ModsDl_State ok">
                    <span class="ModsDl_OkIcon"><IconCheck stroke="2.4" /></span>
                    <b>Instalación completada</b>
                    <span>{{ doneMessage }}</span>
                    <button class="SsBtn SsBtnPrimary" @click="close">Cerrar</button>
                </div>

                <div v-else-if="versions.length" class="ModsDl_Fields">
                    <label class="ModsDl_Field">
                        <span class="ModsDl_FieldLabel">Versión de Minecraft</span>
                        <select v-model="mcVersion" class="SsSel" :disabled="!mcOptions.length">
                            <option v-for="v in mcOptions" :key="v" :value="v">{{ v }}</option>
                        </select>
                    </label>

                    <label class="ModsDl_Check">
                        <input v-model="showSnapshots" type="checkbox" />
                        <span class="ModsDl_CheckBox"></span>
                        <span class="ModsDl_CheckLabel">Incluir snapshots</span>
                        <span class="ModsDl_CheckHint">Muestra versiones semanales y snapshots</span>
                    </label>

                    <label class="ModsDl_Field">
                        <span class="ModsDl_FieldLabel">Tipo de ModLoader</span>
                        <select v-model="loader" class="SsSel" :disabled="isSingleLoader">
                            <option v-for="l in loaderOptions" :key="l" :value="l">{{ l }}</option>
                        </select>
                        <span v-if="isSingleLoader" class="ModsDl_FieldHint">Este proyecto solo ofrece {{ loaderOptions[0] ?? 'un' }} loader</span>
                    </label>

                    <label class="ModsDl_Field">
                        <span class="ModsDl_FieldLabel">Canal de versión</span>
                        <select v-model="channel" class="SsSel">
                            <option value="any">Cualquiera</option>
                            <option v-for="t in VERSION_TYPES" :key="t.id" :value="t.id">{{ t.label }}</option>
                        </select>
                    </label>

                    <div v-if="matchedVersion" class="ModsDl_Summary">
                        <div class="ModsDl_SummaryLine">
                            <span class="ModsDl_SummaryName">{{ matchedVersion.version_number }}</span>
                            <span
                                class="ModsDl_SummaryType"
                                :class="{
                                    Release: matchedVersion.version_type === 'release',
                                    Beta: matchedVersion.version_type === 'beta',
                                    Alpha: matchedVersion.version_type === 'alpha',
                                }"
                            >
                                {{ VERSION_TYPES.find((t) => t.id === matchedVersion?.version_type)?.label ?? matchedVersion?.version_type }}
                            </span>
                        </div>
                        <p v-if="primaryFile" class="ModsDl_SummaryFile">
                            <IconFile stroke="2" /> {{ primaryFile.filename }} · {{ formatBytes(primaryFile.size) }}
                        </p>
                        <p class="ModsDl_SummaryMeta">
                            {{ formatDate(matchedVersion.date_published) }} · {{ formatCount(matchedVersion.downloads) }} descargas
                        </p>
                    </div>
                    <p v-else class="ModsDl_NoMatch">No hay versión compatible con esa combinación. Prueba otro loader, canal o versión de MC.</p>

                    <div v-if="isModpack" class="ModsDl_Dest">
                        <span class="ModsDl_FieldLabel">Destino del modpack</span>
                        <div class="ModsDl_Radios">
                            <label class="ModsDl_Radio" :class="{ on: mpDestination === 'new' }">
                                <input v-model="mpDestination" type="radio" value="new" />
                                <span class="ModsDl_RadioDot"></span>
                                <span class="ModsDl_RadioText">Nueva instancia</span>
                            </label>
                            <label class="ModsDl_Radio" :class="{ on: mpDestination === 'instance' }">
                                <input v-model="mpDestination" type="radio" value="instance" />
                                <span class="ModsDl_RadioDot"></span>
                                <span class="ModsDl_RadioText">Instancia</span>
                            </label>
                            <label class="ModsDl_Radio" :class="{ on: mpDestination === 'global' }">
                                <input v-model="mpDestination" type="radio" value="global" />
                                <span class="ModsDl_RadioDot"></span>
                                <span class="ModsDl_RadioText">Juego global</span>
                            </label>
                        </div>
                        <label v-if="mpDestination === 'new'" class="ModsDl_Field">
                            <input
                                v-model="modpackName"
                                class="SsIn ModsDl_NameInput"
                                type="text"
                                maxlength="48"
                                placeholder="Nombre de la instancia"
                                spellcheck="false"
                                autocomplete="off"
                            />
                            <span class="ModsDl_FieldHint">Nacerá con el icono del modpack y su Minecraft + modloader.</span>
                        </label>
                        <label v-if="mpDestination === 'instance'" class="ModsDl_Field">
                            <select v-model="selectedInstance" class="SsSel" :disabled="loadingInstances || !instances.length">
                                <option v-for="i in instances" :key="i.name" :value="i.name">
                                    {{ i.title || i.name }}
                                </option>
                            </select>
                            <span v-if="!loadingInstances && !instances.length" class="ModsDl_FieldHint">
                                No hay instancias: crea una desde el panel de Instancias o elige otra opción.
                            </span>
                        </label>
                        <p class="ModsDl_DestHint">
                            <template v-if="mpDestination === 'new'">
                                Se creará la instancia, se descargará su Minecraft, se instalará
                                su modloader y se copiarán la configuración (overrides), los mods
                                y las texturas del pack.
                            </template>
                            <template v-else-if="mpDestination === 'instance'">
                                Se comprobará el Minecraft del pack en la instancia, se instalará
                                su modloader si le falta y se copiarán overrides, mods y texturas.
                            </template>
                            <template v-else>
                                Se comprobará el Minecraft del pack en el juego global, se
                                instalará su modloader si falta y se copiarán overrides, mods y
                                texturas.
                            </template>
                        </p>
                    </div>

                    <div v-else class="ModsDl_Dest">
                        <span class="ModsDl_FieldLabel">Destino de la instalación</span>
                        <div class="ModsDl_Radios">
                            <label class="ModsDl_Radio" :class="{ on: destination === 'global' }">
                                <input v-model="destination" type="radio" value="global" />
                                <span class="ModsDl_RadioDot"></span>
                                <span class="ModsDl_RadioText">Juego global</span>
                            </label>
                            <label class="ModsDl_Radio" :class="{ on: destination === 'instance' }">
                                <input v-model="destination" type="radio" value="instance" />
                                <span class="ModsDl_RadioDot"></span>
                                <span class="ModsDl_RadioText">Instancia</span>
                            </label>
                        </div>
                        <label v-if="destination === 'instance'" class="ModsDl_Field">
                            <select v-model="selectedInstance" class="SsSel" :disabled="loadingInstances || !instances.length">
                                <option v-for="i in instances" :key="i.name" :value="i.name">
                                    {{ i.title || i.name }}
                                </option>
                            </select>
                            <span v-if="!loadingInstances && !instances.length" class="ModsDl_FieldHint">
                                No hay instancias: crea una desde el panel de Instancias.
                            </span>
                        </label>
                        <p v-if="destPath" class="ModsDl_DestPath" :title="destPath">
                            <IconFile stroke="2" /> {{ destPath }}/{{ primaryFile?.filename ?? '' }}
                        </p>
                        <p v-if="loaderHint" class="ModsDl_LoaderWarn">
                            <IconAlertTriangle stroke="2" /> {{ loaderHint }}
                        </p>
                    </div>
                </div>

                <p v-else class="ModsDl_State">Este proyecto no tiene versiones disponibles.</p>
            </div>

            <footer class="ModsDl_Footer">
                <button v-if="phase === 'working'" class="SsBtn" @click="onCancelWork">
                    <IconPlayerStop stroke="2" /> Cancelar instalación
                </button>
                <template v-else>
                    <button class="SsBtn" @click="close">Cancelar</button>
                    <button
                        class="SsBtn SsBtnPrimary ModsDl_Confirm"
                        :disabled="!matchedVersion || loading || (isModpack ? (mpDestination === 'new' ? !sanitizeInstanceName(modpackName || title) : mpDestination === 'instance' && !selectedInstance) : destination === 'instance' && !selectedInstance)"
                        @click="onDownload"
                    >
                        <IconDownload stroke="2" /> Descargar
                    </button>
                </template>
            </footer>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use './Styles/DownloadDialog.scss';
</style>
