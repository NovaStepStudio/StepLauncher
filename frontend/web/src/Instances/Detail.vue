<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';
import { heavyPanel } from '@/Common/Overlays/Store';
import {
    IconArrowLeft, IconStar, IconStarFilled, IconPencil, IconSettings, IconTrash,
    IconBox, IconX, IconCheck, IconDownload, IconClock, IconDeviceGamepad, IconPhoto,
    IconFolderOpen, IconHistory, IconInfoCircle, IconPlayerStop,
} from '@tabler/icons-vue';
import {
    ListInstanceScreenshots,
    ReadLocalFile,
    RemoveInstanceVersion,
} from '@wailsjs/go/main/App';
import {
    detailOf,
    loadDetails,
    updateConfig,
    toggleFavorite,
    launchInstance,
    launching,
    getInstanceStats,
    formatPlayTime,
    openInstanceFolder,
    loaderOf,
    loaderLabel,
    downloads,
    cancelDownload,
    dlStateText,
    loaderDlOf,
    loaderDlStateText,
    isInstanceBusy,
    type InstanceStats,
    type InstalledLoaderInfo,
} from './Store';
import { loadLocal } from '@/Common/Stores/Ui';

import iconVanilla from '../../assets/icons/minecraft.png';
import iconFabric from '../../assets/icons/fabric.png';
import iconForge from '../../assets/icons/forge.png';
import iconNeoForge from '../../assets/icons/neoforge.png';
import iconQuilt from '../../assets/icons/quilt.png';
import iconLegacyFabric from '../../assets/icons/legacyfabric.png';

const LOADER_ICONS: Record<string, string> = {
    vanilla: iconVanilla,
    fabric: iconFabric,
    forge: iconForge,
    neoforge: iconNeoForge,
    quilt: iconQuilt,
    legacyfabric: iconLegacyFabric,
};

function loaderIcon(): string {
    const key = (loaderOf(props.name)?.loaderType ?? 'vanilla').toLowerCase();
    return LOADER_ICONS[key] ?? iconVanilla;
}

function loaderTitle(): string {
    const l = loaderOf(props.name);
    if (!l) return 'Vanilla (sin modloader)';
    return `${loaderLabel(l)} ${l.loaderVersion ?? ''} para ${l.minecraftVersion ?? ''}`;
}

function loaderVersionId(): string {
    return loaderOf(props.name)?.versionJsonId ?? '';
}

function isLoaderVersion(v: string): boolean {
    const l = loaderOf(props.name) as InstalledLoaderInfo | null;
    if (!l) return false;
    if (loaderVersionId() && v === loaderVersionId()) return true;
    if (v === l.minecraftVersion) return false;
    const key = (l.loaderType ?? '').toLowerCase();
    return key !== '' && v.toLowerCase().includes(key);
}

const props = defineProps<{
    name: string;
}>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'exit'): void;
    (e: 'edit', name: string): void;
    (e: 'settings', name: string): void;
    (e: 'delete', name: string): void;
    (e: 'shots', name: string): void;
    (e: 'download', name: string): void;
}>();

const d = computed(() => detailOf(props.name));
const loadingDetail = ref(true);
const heroIcon = ref('');
const heroBanner = ref('');

const flash = ref('');
const flashOk = ref(true);
function setFlash(text: string, ok = true) {
    flash.value = text;
    flashOk.value = ok;
    if (ok) window.setTimeout(() => { if (flash.value === text) flash.value = ''; }, 4000);
}

watch(
    () => [d.value?.meta?.icon, d.value?.meta?.banner] as const,
    async () => {
        const m = d.value?.meta;
        heroIcon.value = '';
        heroBanner.value = '';
        if (m?.icon) heroIcon.value = await loadLocal(m.icon);
        if (m?.banner) heroBanner.value = await loadLocal(m.banner);
    },
    { immediate: true }
);

// ---------- Pestañas ----------

type Tab = 'resumen' | 'versiones' | 'capturas';
const tab = ref<Tab>('resumen');

// ---------- Acciones ----------

const playMsg = ref('');
const playing = computed(() => launching.value[props.name] ?? false);

// Descarga activa de la instancia (estado en vivo alimentado por el store
// con los eventos download_*: sigue visible aunque el modal se haya cerrado).
const instDl = computed(() => downloads.value[props.name] ?? null);

// Instalación de modloader en curso (eventos modloader_*): también visible en
// el banner aunque el modal de instalación se haya cerrado.
const instLdr = computed(() => loaderDlOf(props.name));

function fmtMb(mb: number): string {
    if (!Number.isFinite(mb) || mb <= 0) return '0 MB';
    if (mb >= 1024) return `${(mb / 1024).toLocaleString('es-ES', { maximumFractionDigits: 1 })} GB`;
    return `${mb.toLocaleString('es-ES', { maximumFractionDigits: 1 })} MB`;
}

// Ocupada: descarga activa, modloader instalándose o partida lanzándose.
const opBusy = computed(() => isInstanceBusy(props.name));

async function play() {
    if (playing.value) return;
    playMsg.value = 'Lanzando…';
    const err = await launchInstance(props.name);
    playMsg.value = err ? `Error: ${err}` : 'Juego lanzado. Al cerrar Minecraft el estado se actualizará.';
}

const activeVersion = computed(() => d.value?.config?.version ?? '');

async function toggleFav() {
    const meta = d.value?.meta;
    if (!meta) return;
    const err = await toggleFavorite(props.name, !meta.favorite);
    if (err) setFlash(err, false);
}

async function openFolder() {
    const err = await openInstanceFolder(props.name);
    if (err) setFlash(err, false);
}

const busy = ref(false);

async function setActive(version: string) {
    const err = await updateConfig(props.name, { version });
    setFlash(err ? `No se pudo cambiar: ${err}` : `Versión activa: ${version}`, !err);
}

async function removeVersionFromInstance(version: string, e: Event) {
    e.stopPropagation();
    if (busy.value) return;
    busy.value = true;
    try {
        await RemoveInstanceVersion(props.name, version);
        await loadDetails(props.name);
        setFlash(`Versión ${version} eliminada.`);
    } catch (err: any) {
        setFlash(`No se pudo eliminar: ${err?.message ?? 'error desconocido'}`, false);
    } finally {
        busy.value = false;
    }
}

function versionPlays(version: string): string {
    const s = stats.value?.versions?.find((x) => x.version === version);
    if (!s || (s.playCount ?? 0) <= 0) return 'Nunca jugada';
    const n = s.playCount === 1 ? 'partida' : 'partidas';
    return `${s.playCount} ${n} · ${formatPlayTime(s.totalPlayed)} jugados`;
}

// ---------- Estadísticas ----------

const stats = ref<InstanceStats | null>(null);

async function refreshStats() {
    stats.value = await getInstanceStats(props.name);
}

// ---------- Capturas ----------

type ShotInfo = { name: string; path: string; size: number; time: string };
const shotThumbs = ref<ShotInfo[]>([]);
const shotsLoading = ref(false);
const shotsTotal = ref(0);
const thumbCache = new Map<string, string>();

async function refreshShots() {
    shotsLoading.value = true;
    try {
        const list = ((await ListInstanceScreenshots(props.name)) ?? []) as unknown as ShotInfo[];
        shotsTotal.value = list.length;
        shotThumbs.value = list.slice(0, 6);
    } catch {
        shotsTotal.value = 0;
        shotThumbs.value = [];
    } finally {
        shotsLoading.value = false;
    }
}

function thumbOf(rel: string): string {
    const cached = thumbCache.get(rel);
    if (cached) return cached;
    ReadLocalFile(rel)
        .then(async (data) => {
            if (thumbCache.has(rel)) return;
            const url = await blobThumb(data, rel);
            if (url && !thumbCache.has(rel)) thumbCache.set(rel, url);
        })
        .catch(() => { });
    return '';
}

function blobThumb(data: unknown, rel: string): Promise<string> {
    return new Promise((resolve) => {
        let bytes: Uint8Array;
        if (typeof data === 'string') {
            const bin = atob(data);
            bytes = new Uint8Array(bin.length);
            for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
        } else if (Array.isArray(data)) {
            bytes = new Uint8Array(data);
        } else if (data instanceof Uint8Array) {
            bytes = data;
        } else {
            resolve('');
            return;
        }
        let mimeType = 'image/png';
        const lower = rel.toLowerCase();
        if (lower.endsWith('.jpg') || lower.endsWith('.jpeg')) mimeType = 'image/jpeg';
        else if (lower.endsWith('.gif')) mimeType = 'image/gif';
        const url = URL.createObjectURL(new Blob([bytes.slice().buffer as ArrayBuffer], { type: mimeType }));
        const img = new Image();
        img.onload = () => {
            const scale = Math.min(1, 480 / Math.max(img.naturalWidth, img.naturalHeight));
            if (scale >= 1) {
                resolve(url);
                return;
            }
            const canvas = document.createElement('canvas');
            canvas.width = Math.max(1, Math.round(img.naturalWidth * scale));
            canvas.height = Math.max(1, Math.round(img.naturalHeight * scale));
            const ctx = canvas.getContext('2d');
            if (!ctx) {
                resolve(url);
                return;
            }
            ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
            URL.revokeObjectURL(url);
            canvas.toBlob((b) => resolve(b ? URL.createObjectURL(b) : url), 'image/jpeg', 0.82);
        };
        img.onerror = () => {
            URL.revokeObjectURL(url);
            resolve('');
        };
        img.src = url;
    });
}

// ---------- Arranque ----------

useOverlayEscape(() => emit('close'), { isActive: () => heavyPanel.value === 'instances', priority: 1 });

onMounted(() => {
    void loadDetails(props.name).then(() => {
        loadingDetail.value = false;
    });
    void refreshStats();
    void refreshShots();
});
</script>

<template>
    <div v-if="loadingDetail" class="InstDet InstDetLoading">
        <p>Abriendo la instancia…</p>
    </div>

    <div v-else class="InstDet">
        <!-- Hero -->
        <div class="InstDet_Hero" :class="{ hasBanner: heroBanner }">
            <img v-if="heroBanner" :src="heroBanner" alt="" class="InstDet_BannerImg" />
            <div class="InstDet_HeroScrim" />

            <!-- Visualizador temporal: muestra la descarga/verificación en curso
                 o la instalación del modloader aunque el modal se haya cerrado -->
            <div v-if="instDl || instLdr" class="InstDet_DlOverlay">
                <div v-if="instDl" class="InstDet_DlOverlayMain">
                    <div class="InstDet_DlOverlayHead">
                        <span class="InstDet_DlOverlayState">
                            <span class="InstDet_DlOverlayDot"></span>
                            {{ dlStateText(instDl.state) }} · {{ instDl.version }}
                        </span>
                        <span class="InstDet_DlOverlayPct">{{ Math.round(instDl.percent ?? 0) }}%</span>
                    </div>
                    <div class="InstDet_DlOverlayBar">
                        <span :style="{ width: (instDl.percent ?? 0) + '%' }" />
                    </div>
                    <div class="InstDet_DlOverlayMeta">
                        {{ fmtMb(instDl.mbDownloaded) }} / {{ fmtMb(instDl.mbTotal) }}
                        <template v-if="instDl.filesTotal && instDl.filesTotal > 0">
                            · {{ Math.min(instDl.filesDownloaded ?? 0, instDl.filesTotal) }}/{{ instDl.filesTotal }} archivos
                        </template>
                    </div>
                    <div v-if="instDl.state === 'verifying' || instDl.state === 'redownloading'" class="InstDet_DlOverlayFiles">
                        <span v-if="instDl.filesTotal && instDl.filesTotal > 0">
                            {{ Math.min(instDl.filesDownloaded ?? 0, instDl.filesTotal) }}/{{ instDl.filesTotal }}
                            archivos verificados
                        </span>
                        <span v-else>Preparando la verificación…</span>
                    </div>
                    <div class="InstDet_DlOverlayActions">
                        <button class="InstDet_DlOverlayCancel" title="Cancelar descarga" @click="cancelDownload(props.name)">
                            <IconPlayerStop stroke="2" /> Cancelar
                        </button>
                        <button class="InstDet_DlOverlayPanel" title="Abrir el panel de descargas" @click="emit('download', props.name)">
                            <IconDownload stroke="2" /> Panel
                        </button>
                    </div>
                </div>
                <div v-else-if="instLdr" class="InstDet_DlOverlayMain">
                    <div class="InstDet_DlOverlayHead">
                        <span class="InstDet_DlOverlayState">
                            <span class="InstDet_DlOverlayDot" :class="{ pulse: instLdr.phase !== 'done' && instLdr.phase !== 'error' }"></span>
                            {{ loaderDlStateText(instLdr) }}
                        </span>
                        <span
                            v-if="instLdr.total > 0 && instLdr.phase === 'downloading'"
                            class="InstDet_DlOverlayPct"
                        >{{ Math.round((instLdr.progress / instLdr.total) * 100) }}%</span>
                        <span v-else-if="instLdr.phase === 'done'" class="InstDet_DlOverlayOk">OK</span>
                    </div>
                    <div v-if="instLdr.total > 0 && instLdr.phase === 'downloading'" class="InstDet_DlOverlayBar">
                        <span :style="{ width: (instLdr.progress / instLdr.total) * 100 + '%' }" />
                    </div>
                    <div v-else-if="instLdr.phase === 'resolving' || instLdr.phase === 'installing'" class="InstDet_DlOverlayPulse"><span></span></div>
                    <div v-if="instLdr.phase === 'error'" class="InstDet_DlOverlayError">{{ instLdr.message }}</div>
                    <div v-else class="InstDet_DlOverlayMeta">
                        {{ instLdr.loader }} {{ instLdr.loaderVersion }}
                        <template v-if="instLdr.mcVersion"> · MC {{ instLdr.mcVersion }}</template>
                    </div>
                    <div class="InstDet_DlOverlayActions">
                        <button class="InstDet_DlOverlayPanel" title="Abrir el panel de descargas" @click="emit('download', props.name)">
                            <IconDownload stroke="2" /> Panel
                        </button>
                    </div>
                </div>
            </div>

            <button class="InstDet_Back" title="Volver a la lista" @click="emit('close')">
                <IconArrowLeft stroke="2" /> Instancias
            </button>
            <button class="InstDet_Close" title="Cerrar" @click="emit('exit')">
                <IconX stroke="2" />
            </button>

            <div class="InstDet_HeroMain">
                <span class="InstDet_Icon">
                    <img v-if="heroIcon" :src="heroIcon" :alt="d?.meta?.title ?? ''" />
                    <IconBox v-else stroke="1.5" />
                </span>
                <div class="InstDet_Identity">
                    <h3>{{ d?.meta?.title || d?.meta?.name }}</h3>
                    <p v-if="d?.meta?.description">{{ d?.meta?.description }}</p>
                    <div class="InstDet_Chips">
                        <span class="InstDet_Chip"><IconClock stroke="2" /> {{ formatPlayTime(d?.meta?.playTime ?? 0) }}</span>
                        <span class="InstDet_Chip InstDet_LoaderChip" :title="loaderTitle()">
                            <img :src="loaderIcon()" alt="" /> {{ loaderLabel(loaderOf(props.name)) }}
                        </span>
                        <span v-if="d?.meta?.lastPlayed" class="InstDet_Chip">
                            <IconHistory stroke="2" /> Última partida {{ d?.meta?.lastPlayed?.slice(0, 10) }}
                        </span>
                    </div>
                </div>
                <div class="InstDet_HeroActions">
                    <button
                        class="InstDet_Play"
                        :class="{ disabled: playing || !activeVersion || opBusy }"
                        :disabled="playing || !activeVersion || opBusy"
                        @click="play()"
                    >
                        <IconDeviceGamepad stroke="2" />
                        {{ playing ? 'Lanzando…' : instDl ? 'Descargando…' : instLdr ? 'Instalando…' : 'Jugar' }}
                    </button>
                    <button class="InstDet_DlHero" title="Descargar versión" :disabled="opBusy" @click="emit('download', props.name)">
                        <IconDownload stroke="2" /> {{ instDl ? 'Descargando…' : instLdr ? 'Instalando…' : 'Descargar' }}
                    </button>
                    <button class="InstDet_Tool" :class="{ on: d?.meta?.favorite }" title="Favorita" @click="toggleFav()">
                        <IconStarFilled v-if="d?.meta?.favorite" stroke="2" />
                        <IconStar v-else stroke="2" />
                    </button>
                    <button class="InstDet_Tool" title="Abrir carpeta de la instancia" @click="openFolder()">
                        <IconFolderOpen stroke="2" />
                    </button>
                    <button class="InstDet_Tool" title="Editar" @click="emit('edit', props.name)">
                        <IconPencil stroke="2" />
                    </button>
                    <button class="InstDet_Tool" title="Configurar" @click="emit('settings', props.name)">
                        <IconSettings stroke="2" />
                    </button>
                    <button class="InstDet_Tool danger" title="Eliminar instancia" @click="emit('delete', props.name)">
                        <IconTrash stroke="2" />
                    </button>
                </div>
            </div>
        </div>

        <!-- Pestañas -->
        <div class="InstDet_Tabs">
            <button :class="{ active: tab === 'resumen' }" @click="tab = 'resumen'">
                <IconInfoCircle stroke="2" /> Resumen
            </button>
            <button :class="{ active: tab === 'versiones' }" @click="tab = 'versiones'">
                <IconBox stroke="2" /> Versiones instaladas <em>{{ d?.meta?.versions?.length ?? 0 }}</em>
            </button>
            <button :class="{ active: tab === 'capturas' }" @click="tab = 'capturas'">
                <IconPhoto stroke="2" /> Capturas
            </button>
        </div>

        <div class="InstDet_Body">
            <p v-if="playMsg" class="InstDet_Flash">{{ playMsg }}</p>
            <p v-if="flash" :class="['InstDet_Flash', { error: !flashOk }]">{{ flash }}</p>

            <!-- Resumen -->
            <template v-if="tab === 'resumen'">
                <div class="InstDet_Tiles">
                    <div class="InstDet_Tile">
                        <span class="InstDet_TileLabel"><IconDeviceGamepad stroke="2" /> Versión activa</span>
                        <b>{{ activeVersion || '—' }}</b>
                    </div>
                    <div class="InstDet_Tile">
                        <span class="InstDet_TileLabel"><img :src="loaderIcon()" class="InstDet_TileLoaderImg" alt="" /> Modloader</span>
                        <b :title="loaderTitle()">
                            {{ loaderLabel(loaderOf(props.name)) }}<template v-if="loaderOf(props.name)"> · {{ loaderOf(props.name)?.loaderVersion }}</template>
                        </b>
                    </div>
                    <div class="InstDet_Tile">
                        <span class="InstDet_TileLabel"><IconClock stroke="2" /> Tiempo jugado</span>
                        <b>{{ formatPlayTime(d?.meta?.playTime ?? 0) }}</b>
                    </div>
                    <div class="InstDet_Tile">
                        <span class="InstDet_TileLabel"><IconHistory stroke="2" /> Última partida</span>
                        <b>{{ d?.meta?.lastPlayed?.slice(0, 10) || 'Nunca jugada' }}</b>
                    </div>
                    <div class="InstDet_Tile">
                        <span class="InstDet_TileLabel"><IconDownload stroke="2" /> Sesiones</span>
                        <b>{{ stats?.totalSessions ?? '—' }}</b>
                    </div>
                </div>

                <div v-if="stats" class="InstDet_Week">
                    <div class="InstDet_SectionHead">
                        <span>Últimos 7 días</span>
                        <span v-if="stats.running" class="InstDet_WeekRunning">En juego ahora</span>
                    </div>
                    <div class="InstDet_WeekGrid">
                        <div class="InstDet_WeekItem">
                            <label><IconClock stroke="2" /> Tiempo jugado</label>
                            <b>{{ formatPlayTime(stats.weeklyPlayTime) }}</b>
                        </div>
                        <div class="InstDet_WeekItem">
                            <label><IconDeviceGamepad stroke="2" /> Sesiones</label>
                            <b>{{ stats.weeklySessions }}</b>
                        </div>
                        <div class="InstDet_WeekItem">
                            <label><IconBox stroke="2" /> Versiones usadas</label>
                            <b>{{ stats.weeklyVersions.length || '—' }}</b>
                        </div>
                        <div class="InstDet_WeekItem">
                            <label><IconHistory stroke="2" /> Primera partida</label>
                            <b>{{ stats.firstPlayed ? new Date(stats.firstPlayed * 1000).toLocaleDateString() : '—' }}</b>
                        </div>
                    </div>
                    <div v-if="stats.versions.length" class="InstDet_WeekVersions">
                        <span v-for="v in stats.versions" :key="v.version" class="InstDet_WeekVersion">
                            {{ v.version }}
                            <em>{{ v.playCount }}× · {{ formatPlayTime(v.totalPlayed) }}</em>
                        </span>
                    </div>
                </div>
            </template>

            <!-- Versiones instaladas -->
            <template v-else-if="tab === 'versiones'">
                <div class="InstDet_SectionHead">
                    <span>Versiones instaladas</span>
                    <button class="SsBtn SsBtnPrimary InstDet_DlBtn" :disabled="opBusy" @click="emit('download', props.name)">
                        <IconDownload stroke="2" /> {{ instDl ? 'Descargando…' : instLdr ? 'Instalando…' : 'Añadir versión' }}
                    </button>
                </div>

                <div v-if="d?.meta?.versions?.length" class="InstDet_VersionRows">
                    <div
                        v-for="v in d?.meta?.versions ?? []"
                        :key="v"
                        class="InstDet_VersionRow"
                        :class="{ active: v === activeVersion, clickable: v !== activeVersion }"
                        :title="v === activeVersion ? 'Versión activa' : `Usar ${v}`"
                        @click="v !== activeVersion && setActive(v)"
                    >
                        <span class="InstDet_VersionActive">
                            <IconCheck v-if="v === activeVersion" stroke="2.4" />
                            <img v-else-if="isLoaderVersion(v)" :src="loaderIcon()" class="InstDet_VersionLoaderIcon" alt="" :title="`Versión de ${loaderLabel(loaderOf(props.name))}`" />
                            <IconBox v-else stroke="1.6" />
                        </span>
                        <span class="InstDet_VersionMeta">
                            <span class="InstDet_VersionId">{{ v }}</span>
                            <span class="InstDet_VersionSub">{{ isLoaderVersion(v) ? `Versión de ${loaderLabel(loaderOf(props.name))}` : versionPlays(v) }}</span>
                        </span>
                        <span v-if="v === activeVersion" class="InstDet_VersionTag">Activa</span>
                        <span v-else-if="isLoaderVersion(v)" class="InstDet_VersionTag loader">{{ loaderLabel(loaderOf(props.name)) }}</span>

                        <div class="InstDet_VersionTools">
                            <button
                                v-if="v !== activeVersion"
                                class="InstDet_TinyBtn"
                                @click.stop="setActive(v)"
                            ><IconCheck stroke="2" /> Usar</button>
                            <button class="InstDet_TinyIcon" title="Eliminar esta versión" @click.stop="removeVersionFromInstance(v, $event)">
                                <IconX stroke="2" />
                            </button>
                        </div>
                    </div>
                </div>
                <p v-else class="InstDet_Muted">
                    Esta instancia no tiene versiones descargadas aún.
                    Pulsa «Añadir versión» para descargar una.
                </p>
            </template>

            <!-- Capturas -->
            <template v-else>
                <div class="InstDet_SectionHead">
                    <span>Capturas</span>
                    <button class="SsBtn SsBtnPrimary InstDet_DlBtn" :disabled="!shotsTotal" @click="emit('shots', props.name)">
                        <IconFolderOpen stroke="2" /> Abrir visor
                    </button>
                </div>

                <div v-if="shotsLoading" class="InstDet_Muted">Buscando capturas…</div>
                <div v-else-if="shotThumbs.length" class="InstDet_Shots">
                    <button v-for="s in shotThumbs" :key="s.path" class="InstDet_Shot" @click="emit('shots', props.name)">
                        <img v-if="thumbOf(s.path)" :src="thumbOf(s.path)" :alt="s.name" loading="lazy" />
                        <IconBox v-else stroke="1.5" />
                        <span class="InstDet_ShotMeta">{{ s.name }}</span>
                    </button>
                </div>
                <p v-else class="InstDet_Muted">
                    Aún no hay capturas en esta instancia. Toma una con F2 y aparecerá aquí.
                </p>
            </template>
        </div>
    </div>
</template>

<style scoped lang="scss">
@use './Styles/Detail.scss';
</style>