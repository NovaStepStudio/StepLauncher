<script setup lang="ts">
import { ref, computed, watch, reactive, onMounted, onUnmounted } from 'vue';
import { marked } from 'marked';
import { Browser } from '@wailsio/runtime';
import {
    IconArrowLeft, IconDownload, IconPuzzle, IconUsers, IconClock,
    IconPhoto, IconBox, IconTextCaption, IconChevronLeft, IconChevronRight,
    IconX, IconRotateClockwise, IconLoader2, IconExternalLink,
    IconZoomIn, IconZoomOut, IconWallpaper,
} from '@tabler/icons-vue';
import { DownloadGalleryImageAsBackground } from '@wailsjs/StepLauncher/internal/Services/Appearance/appearanceservice';
import { GetConfig } from '@wailsjs/StepLauncher/internal/Services/Config/configservice';
import { useModrinth, formatCount, formatBytes, formatDate } from '@/Common/Composables/useModrinth';
import { VERSION_TYPES, typeLabel, environmentLabel } from './Store';
import type { ModrinthVersion } from '@/Common/Composables/useModrinth';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';
import { heavyPanel } from '@/Common/Overlays/Store';
import { applyPersonalization } from '@/Common/Stores/Ui';
import { useBackground } from '@/Common/Composables/useBackground';

import iconFabric from '../../assets/icons/fabric.webp';
import iconForge from '../../assets/icons/forge.webp';
import iconNeoForge from '../../assets/icons/neoforge.webp';
import iconQuilt from '../../assets/icons/quilt.webp';
import iconLegacyFabric from '../../assets/icons/legacyfabric.webp';

const props = defineProps<{
    slugOrId: string;
    author?: string;
}>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'exit'): void;
    (e: 'download', project: { slug: string; title: string; icon_url: string | null; project_type: string }): void;
}>();

const modrinth = useModrinth();

// Panel propio para ESC: prioridad 1 (sub-vista dentro de Mods, por encima del panel 0)
// Así ESC primero vuelve a la lista y solo el segundo ESC cierra el overlay Mods.
useOverlayEscape(() => emit('close'), {
    priority: 1,
    isActive: () => heavyPanel.value === 'mods' && lightboxIndex.value === null,
});

const followersCount = computed(() => {
    const p: any = project.value;
    return p?.followers ?? p?.follows ?? 0;
});

const LOADER_ICONS: Record<string, string> = {
    fabric: iconFabric,
    forge: iconForge,
    neoforge: iconNeoForge,
    quilt: iconQuilt,
    'legacy-fabric': iconLegacyFabric,
};

marked.use({ gfm: true, breaks: true });

const mdCache = new Map<string, string>();

function renderMarkdown(md: string, key: string): string {
    const hit = mdCache.get(key);
    if (hit !== undefined) return hit;
    try {
        const out = marked.parse(md ?? '');
        let html = typeof out === 'string' ? out : '';
        // Limpieza observada en fetch real (Sodium, Create):
        // - Create trae muchos <p>&nbsp;</p> que generan bloques vacíos apilados "I I I"
        // - Sodium trae <center><img></center> y badges <a><img> que deben permanecer inline
        html = html
            .replace(/<p>\s*(&nbsp;|\u00A0|\s)*\s*<\/p>/gi, '')
            .replace(/<p>\s*<br\s*\/?>\s*<\/p>/gi, '');
        if (mdCache.size > 80) mdCache.clear();
        mdCache.set(key, html);
        return html;
    } catch {
        return '';
    }
}

const project = computed(() => modrinth.projectState.project);
const versions = computed(() => modrinth.projectState.versions);
const hasVersions = computed(() => versions.value.length > 0);

// Autor mostrado: prop (desde búsqueda) o fetch al team si viene vacío
const fetchedAuthor = ref('');
const displayAuthor = computed(() => props.author?.trim() || fetchedAuthor.value || 'Desconocido');

watch(
    () => [project.value?.id, props.author] as const,
    async ([id, propAuthor]) => {
        if (propAuthor?.trim()) {
            fetchedAuthor.value = '';
            return;
        }
        if (!id) {
            fetchedAuthor.value = '';
            return;
        }
        try {
            const res = await fetch(`https://api.modrinth.com/v2/project/${encodeURIComponent(id)}/members`, {
                headers: { Accept: 'application/json' },
            });
            if (!res.ok) return;
            const members: any[] = await res.json();
            const owner = members.find((m) => String(m.role ?? '').toLowerCase() === 'owner') ?? members[0];
            const name = owner?.user?.username ?? owner?.user?.name ?? '';
            if (name) fetchedAuthor.value = String(name);
        } catch (_e) {}
    },
    { immediate: true }
);

const descriptionHtml = computed(() => {
    const p = project.value;
    if (!p) return '';
    return renderMarkdown(p.body ?? '', `body:${p.id}`);
});

// Los enlaces del markdown y del panel de detalles SIEMPRE se abren en el
// navegador del sistema, nunca dentro de la ventana de Wails.
function openLink(url: string): void {
    void Browser.OpenURL(url);
}

function onMdClick(e: MouseEvent): void {
    const anchor = (e.target as HTMLElement | null)?.closest?.('a') as HTMLAnchorElement | null;
    if (!anchor) return;
    e.preventDefault();
    void Browser.OpenURL(anchor.href);
}

// ---------- Pestañas ----------

const tab = ref<'description' | 'gallery' | 'versions'>('description');
const galleryCount = computed(() => project.value?.gallery?.length ?? 0);
const hasGallery = computed(() => galleryCount.value > 0);

// ---------- Versiones del proyecto ----------

const mcFilter = ref('');
const typeFilter = ref('');

const projectMcVersions = computed(() => {
    const set = new Set<string>();
    for (const v of versions.value) for (const gv of v.game_versions) set.add(gv);
    return [...set].sort((a, b) => b.localeCompare(a));
});

const filteredVersions = computed(() =>
    versions.value.filter((v) => {
        if (mcFilter.value && !v.game_versions.includes(mcFilter.value)) return false;
        if (typeFilter.value && v.version_type !== typeFilter.value) return false;
        return true;
    })
);

// ---------- Paginación de versiones (evita lag con 1k+ versiones como Fabric API) ----------
const versionsPage = ref(1);
const versionsPerPage = ref(20);
const VERSIONS_LIMIT_OPTIONS = [10, 20, 50, 100] as const;

const totalVersionPages = computed(() => Math.max(1, Math.ceil(filteredVersions.value.length / versionsPerPage.value)));
const pagedVersions = computed(() => {
    const start = (versionsPage.value - 1) * versionsPerPage.value;
    return filteredVersions.value.slice(start, start + versionsPerPage.value);
});
const versionRangeStart = computed(() => (filteredVersions.value.length ? (versionsPage.value - 1) * versionsPerPage.value + 1 : 0));
const versionRangeEnd = computed(() => Math.min(versionsPage.value * versionsPerPage.value, filteredVersions.value.length));
const versionPageItems = computed<(number | 'ellipsis')[]>(() => {
    const total = totalVersionPages.value;
    const cur = versionsPage.value;
    if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);
    const items: (number | 'ellipsis')[] = [1];
    const lo = Math.max(2, cur - 1);
    const hi = Math.min(total - 1, cur + 1);
    if (lo > 2) items.push('ellipsis');
    for (let p = lo; p <= hi; p++) items.push(p);
    if (hi < total - 1) items.push('ellipsis');
    items.push(total);
    return items;
});

function goToVersionPage(p: number): void {
    const t = Math.min(Math.max(1, p), totalVersionPages.value);
    if (t === versionsPage.value) return;
    versionsPage.value = t;
}

// Reset a página 1 al cambiar filtros o al cargar nuevo proyecto
watch([mcFilter, typeFilter, versionsPerPage], () => {
    versionsPage.value = 1;
});
watch(filteredVersions, () => {
    if (versionsPage.value > totalVersionPages.value) versionsPage.value = totalVersionPages.value;
});

const selectedVersionId = ref('');
const selectedVersion = computed(() => versions.value.find((v) => v.id === selectedVersionId.value) ?? null);

function selectVersion(v: ModrinthVersion): void {
    selectedVersionId.value = v.id;
    void modrinth.loadChangelog(v.id);
}

const changelogHtml = computed(() => {
    if (!modrinth.changelogState.changelog) return '';
    return renderMarkdown(modrinth.changelogState.changelog, `cl:${modrinth.changelogState.versionId}`);
});

// ---------- Galería / lightbox avanzado (estilo Screenshots) ----------

const lightboxIndex = ref<number | null>(null);
const lightboxList = computed(() => {
    const g = project.value?.gallery ?? [];
    return g.filter((img) => img.url);
});

const zoom = ref(1);
const pan = reactive({ x: 0, y: 0 });
const stageEl = ref<HTMLElement | null>(null);
const baseSize = ref<{ w: number; h: number } | null>(null);
const dragState = ref<{ x: number; y: number; sx: number; sy: number } | null>(null);

const MIN_ZOOM = 0.5;
const MAX_ZOOM = 4;

const previewTitle = computed(() => {
    if (lightboxIndex.value === null) return '';
    const item = lightboxList.value[lightboxIndex.value];
    return item?.title ?? item?.description ?? `Imagen ${(lightboxIndex.value ?? 0) + 1}`;
});

const imgStyle = computed(() => ({
    transform: `translate(calc(-50% + ${pan.x}px), calc(-50% + ${pan.y}px)) scale(${zoom.value})`,
}));

const previewImgError = ref(false);
const previewHighResUrl = computed(() => {
    if (lightboxIndex.value === null) return '';
    const item: any = lightboxList.value[lightboxIndex.value];
    if (!item) return '';
    // Modrinth devuelve url (miniatura _350.webp) y raw_url (original alta calidad)
    return item.raw_url || item.url || '';
});
const previewSrc = computed(() => {
    if (lightboxIndex.value === null) return '';
    const item: any = lightboxList.value[lightboxIndex.value];
    if (!item) return '';
    if (previewImgError.value) return item.url;
    return (item.raw_url || item.url) as string;
});

function onPreviewImgError(): void {
    previewImgError.value = true;
}

// Estado para "Colocar como fondo"
const isSettingBg = ref(false);
const bgSetMsg = ref('');
const bgSetError = ref('');

async function setAsBackground(): Promise<void> {
    if (lightboxIndex.value === null || isSettingBg.value) return;
    const item = lightboxList.value[lightboxIndex.value];
    if (!item) return;
    isSettingBg.value = true;
    bgSetMsg.value = '';
    bgSetError.value = '';
    let lastErr: any = null;
    const author = displayAuthor.value !== 'Desconocido' ? displayAuthor.value : (props.author || 'Desconocido');
    const modName = project.value?.title ?? 'Mod';
    const title = item.title ?? item.description ?? '';
    const candidates = [previewHighResUrl.value || item.url, item.url].filter((v, i, a) => v && a.indexOf(v) === i);
    for (const url of candidates) {
        try {
            const rel = await DownloadGalleryImageAsBackground(url, author, modName, title);
            if (rel) {
                // Refrescar personalización en el frontend para que el fondo y el label se actualicen al instante
                try {
                    const cfg = await GetConfig();
                    if (cfg?.personalization) {
                        applyPersonalization(cfg.personalization as any);
                        await useBackground().refreshBackground();
                    }
                } catch (_e) {}
                bgSetMsg.value = 'Fondo aplicado';
                isSettingBg.value = false;
                setTimeout(() => { bgSetMsg.value = ''; }, 2500);
                return;
            }
        } catch (e: any) {
            lastErr = e;
            // Si es 404 en alta resolución, intentar con original
            const msg = String(e?.message ?? '');
            if (!msg.includes('404') && !msg.includes('respondió')) {
                break;
            }
        }
    }
    bgSetError.value = lastErr?.message ?? 'No se pudo colocar como fondo';
    isSettingBg.value = false;
    setTimeout(() => { bgSetError.value = ''; }, 4000);
}

function resetPreviewZoom(): void {
    zoom.value = 1;
    pan.x = 0;
    pan.y = 0;
    baseSize.value = null;
    dragState.value = null;
}

function closeLightbox(): void {
    lightboxIndex.value = null;
    resetPreviewZoom();
}

function lightboxPrev(): void {
    if (lightboxIndex.value === null || !lightboxList.value.length) return;
    lightboxIndex.value = (lightboxIndex.value - 1 + lightboxList.value.length) % lightboxList.value.length;
    resetPreviewZoom();
}

function lightboxNext(): void {
    if (lightboxIndex.value === null || !lightboxList.value.length) return;
    lightboxIndex.value = (lightboxIndex.value + 1) % lightboxList.value.length;
    resetPreviewZoom();
}

function openLightbox(idx: number): void {
    lightboxIndex.value = idx;
    resetPreviewZoom();
}

function setZoom(v: number): void {
    zoom.value = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, Math.round(v * 100) / 100));
    clampPan();
}

function biteZoom(delta: number): void {
    setZoom(zoom.value + delta);
}

function clampPan(): void {
    const st = stageEl.value;
    const bs = baseSize.value;
    if (!st || !bs) return;
    const vw = bs.w * zoom.value;
    const vh = bs.h * zoom.value;
    const mx = Math.max(0, (vw - st.clientWidth) / 2);
    const my = Math.max(0, (vh - st.clientHeight) / 2);
    pan.x = Math.max(-mx, Math.min(mx, pan.x));
    pan.y = Math.max(-my, Math.min(my, pan.y));
}

function onWheel(e: WheelEvent): void {
    const st = stageEl.value;
    const bs = baseSize.value;
    if (lightboxIndex.value === null || !st || !bs) return;
    const next = e.deltaY < 0 ? zoom.value * 1.12 : zoom.value / 1.12;
    if (next < MIN_ZOOM - 0.001 || next > MAX_ZOOM + 0.001) return;
    const rect = st.getBoundingClientRect();
    const cx = e.clientX - rect.left - st.clientWidth / 2;
    const cy = e.clientY - rect.top - st.clientHeight / 2;
    const ratio = next / zoom.value;
    pan.x = cx - (cx - pan.x) * ratio;
    pan.y = cy - (cy - pan.y) * ratio;
    zoom.value = Math.round(next * 100) / 100;
    clampPan();
}

function onPanStart(e: MouseEvent): void {
    const st = stageEl.value;
    if (!st || zoom.value <= 1) return;
    dragState.value = { x: e.clientX, y: e.clientY, sx: pan.x, sy: pan.y };
}

function onPanMove(e: MouseEvent): void {
    const d = dragState.value;
    if (!d) return;
    pan.x = d.sx + (e.clientX - d.x);
    pan.y = d.sy + (e.clientY - d.y);
    clampPan();
}

function onPanEnd(): void {
    dragState.value = null;
}

function onImgLoad(e: Event): void {
    const img = e.target as HTMLImageElement;
    const st = stageEl.value;
    if (!st || img.naturalWidth <= 0) return;
    const fit = Math.min(
        1,
        (st.clientWidth * 0.9) / img.naturalWidth,
        (st.clientHeight * 0.8) / img.naturalHeight
    );
    baseSize.value = {
        w: Math.round(img.naturalWidth * fit),
        h: Math.round(img.naturalHeight * fit),
    };
    clampPan();
}

function onGalleryKeydown(e: KeyboardEvent): void {
    if (lightboxIndex.value === null) return;
    if (e.key === 'ArrowLeft') { e.preventDefault(); lightboxPrev(); }
    else if (e.key === 'ArrowRight') { e.preventDefault(); lightboxNext(); }
    else if (e.key === '+' || e.key === '=') { e.preventDefault(); biteZoom(0.25); }
    else if (e.key === '-' || e.key === '_') { e.preventDefault(); biteZoom(-0.25); }
}

// ESC para el lightbox: prioridad 2 (por encima del propio Detail y del panel Mods)
useOverlayEscape(closeLightbox, {
    priority: 2,
    isActive: () => lightboxIndex.value !== null && heavyPanel.value === 'mods',
});

watch(lightboxIndex, (v) => {
    if (v !== null) resetPreviewZoom();
    previewImgError.value = false;
    bgSetMsg.value = '';
    bgSetError.value = '';
});

onMounted(() => {
    window.addEventListener('keydown', onGalleryKeydown);
});
onUnmounted(() => {
    window.removeEventListener('keydown', onGalleryKeydown);
});

// ---------- Acciones ----------

function openDownload(): void {
    const p = project.value;
    if (!p) return;
    emit('download', {
        slug: p.slug ?? p.id,
        title: p.title,
        icon_url: p.icon_url,
        project_type: p.project_type,
    });
}

function retry(): void {
    void modrinth.loadProject(props.slugOrId);
}

const VERSION_TYPE_CLASS: Record<string, string> = {
    release: 'Release',
    beta: 'Beta',
    alpha: 'Alpha',
};

function versionTypeLabel(t: string): string {
    return VERSION_TYPES.find((vt) => vt.id === t)?.label ?? t;
}

// Resumen legible de las versiones de Minecraft de una build del mod:
// "1.20.1" (una), "1.20.1+" (misma rama mayor) o "1.8.9 · 1.20.4" (extremos).
function mcShort(gv: readonly string[]): string {
    const first = gv[0];
    if (!first) return '—';
    if (gv.length === 1) return first;
    const last = gv[gv.length - 1] ?? first;
    if (first.split('.')[0] === last.split('.')[0]) return `${first}+`;
    return `${first} · ${last}`;
}

// La columna del changelog solo existe si la versión seleccionada tiene
// changelog (o está cargándolo o falló); si no, se oculta por completo.
const showChangelog = computed(() => {
    if (!selectedVersionId.value) return false;
    const s = modrinth.changelogState;
    return s.loading || !!s.error || !!s.changelog;
});

watch(
    () => props.slugOrId,
    () => {
        tab.value = 'description';
        mcFilter.value = '';
        typeFilter.value = '';
        selectedVersionId.value = '';
        versionsPage.value = 1;
        modrinth.clearChangelog();
        void modrinth.loadProject(props.slugOrId);
    }
);

// Al cargar las versiones, seleccionar la release más reciente por defecto
// para que el changelog esté siempre visible.
watch(
    versions,
    (vs) => {
        if (selectedVersionId.value || !vs.length) return;
        const first = vs.find((v) => v.version_type === 'release') ?? vs[0];
        if (first) selectVersion(first);
    },
    { immediate: true }
);
</script>

<template>
    <div class="ModsDetail">
        <header class="ModsDetail_Head">
            <button class="ModsDetail_Back" title="Volver a la lista" @click="emit('close')">
                <IconArrowLeft stroke="2" />
            </button>

            <template v-if="project">
                <span class="ModsDetail_Icon">
                    <img v-if="project.icon_url" :src="project.icon_url" alt="" />
                    <IconPuzzle v-else stroke="1.5" />
                </span>
                <div class="ModsDetail_Titles">
                    <h2>{{ project.title }}</h2>
                    <p>por {{ displayAuthor === 'Desconocido' ? 'desconocido' : displayAuthor }} <template v-if="project.slug">· {{ project.slug }}</template></p>
                </div>
                <div class="ModsDetail_Badges">
                    <span class="ModsDetail_Badge ModsDetail_Type">{{ typeLabel(project.project_type) }}</span>
                    <span v-if="environmentLabel(project.environment)" class="ModsDetail_Badge ModsDetail_Env">{{ environmentLabel(project.environment) }}</span>
                    <span class="ModsDetail_Badge" title="Seguidores"><IconUsers stroke="2" /> {{ formatCount(followersCount) }}</span>
                    <span class="ModsDetail_Badge"><IconDownload stroke="2" /> {{ formatCount(project.downloads ?? 0) }}</span>
                    <span class="ModsDetail_Badge"><IconClock stroke="2" /> {{ formatDate(project.updated) }}</span>
                </div>
            </template>
            <div v-else class="ModsDetail_Titles">
                <h2>Cargando…</h2>
            </div>

            <div class="ModsDetail_HeadActions">
                <button
                    class="ModsDetail_Dl"
                    :disabled="!project || modrinth.projectState.loading"
                    :title="project ? `Descargar ${project.title}` : ''"
                    @click="openDownload"
                >
                    <IconDownload stroke="2" /> Descargar
                </button>
                <button class="ModsDetail_Exit" title="Cerrar panel" @click="emit('exit')">
                    <IconX stroke="2" />
                </button>
            </div>
        </header>

        <div v-if="modrinth.projectState.error && !project" class="ModsDetail_Error">
            <b>No se pudo cargar el proyecto</b>
            <p>{{ modrinth.projectState.error }}</p>
            <button class="SsBtn SsBtnPrimary" @click="retry">
                <IconRotateClockwise stroke="2" /> Reintentar
            </button>
        </div>

        <template v-else-if="project">
            <nav class="ModsDetail_Tabs">
                <button :class="{ on: tab === 'description' }" @click="tab = 'description'">
                    <IconTextCaption stroke="2" /> Descripción
                </button>
                <button v-if="hasGallery" :class="{ on: tab === 'gallery' }" @click="tab = 'gallery'">
                    <IconPhoto stroke="2" /> Galería <em>{{ galleryCount }}</em>
                </button>
                <button :class="{ on: tab === 'versions' }" @click="tab = 'versions'">
                    <IconBox stroke="2" /> Versiones <em>{{ versions.length }}</em>
                </button>
            </nav>

            <div class="ModsDetail_Body">
                <div v-if="tab === 'description'" class="ModsDetail_DescWrap">
                    <div class="ModsDetail_DescMd" @click="onMdClick">
                        <div v-if="descriptionHtml" class="ModsDetail_Md" v-html="descriptionHtml"></div>
                        <p v-else-if="project.description" class="ModsDetail_Plain">{{ project.description }}</p>
                        <p v-else class="ModsDetail_Plain">Este proyecto no tiene descripción.</p>
                    </div>

                    <aside class="ModsDetail_Info">
                        <h4 class="ModsDetail_InfoTitle">Detalles del mod</h4>
                        <dl class="ModsDetail_InfoList">
                            <div>
                                <dt>Tipo</dt>
                                <dd>{{ typeLabel(project.project_type) }}</dd>
                            </div>
                            <div>
                                <dt>Autor</dt>
                                <dd>{{ displayAuthor }}</dd>
                            </div>
                            <div>
                                <dt>Descargas</dt>
                                <dd>{{ formatCount(project.downloads ?? 0) }}</dd>
                            </div>
                            <div>
                                <dt>Seguidores</dt>
                                <dd>{{ formatCount(followersCount) }}</dd>
                            </div>
                            <div>
                                <dt>Publicado</dt>
                                <dd>{{ formatDate(project.published) }}</dd>
                            </div>
                            <div>
                                <dt>Actualizado</dt>
                                <dd>{{ formatDate(project.updated) }}</dd>
                            </div>
                            <div v-if="environmentLabel(project.environment)">
                                <dt>Entorno</dt>
                                <dd>{{ environmentLabel(project.environment) }}</dd>
                            </div>
                            <div v-if="project.license?.name">
                                <dt>Licencia</dt>
                                <dd>{{ project.license.name }}</dd>
                            </div>
                        </dl>
                        <a
                            v-if="project.source_url"
                            class="ModsDetail_InfoLink"
                            href="#"
                            @click.prevent="openLink(project.source_url)"
                        >
                            <IconExternalLink stroke="2" /> Ver código fuente
                        </a>
                    </aside>
                </div>

                <div v-else-if="tab === 'gallery'" class="ModsDetail_Scroll">
                    <div v-if="lightboxList.length" class="ModsDetail_Gallery">
                        <button
                            v-for="(img, i) in lightboxList"
                            :key="img.url"
                            class="ModsDetail_GalItem"
                            @click="openLightbox(i)"
                        >
                            <img :src="img.url" alt="" loading="lazy" decoding="async" />
                            <span v-if="img.title" class="ModsDetail_GalTitle">{{ img.title }}</span>
                        </button>
                    </div>
                    <p v-else class="ModsDetail_Plain">Este proyecto no tiene galería.</p>

                    <Transition name="ModsLightbox">
                        <div v-if="lightboxIndex !== null" class="ModsDetail_Lightbox" @click.self="closeLightbox">
                            <div class="ModsDetail_LbViewer" @click.self="closeLightbox">
                                <div class="ModsDetail_LbHead">
                                    <button class="ModsDetail_LbBack" title="Volver a la galería (Esc)" @click="closeLightbox">
                                        <IconChevronLeft stroke="2" /> Galería
                                    </button>
                                    <span class="ModsDetail_LbTitle" :title="previewTitle">{{ previewTitle }}</span>
                                    <span class="ModsDetail_LbCount">{{ (lightboxIndex ?? 0) + 1 }} / {{ lightboxList.length }}</span>
                                    <div class="ModsDetail_LbTools">
                                        <button
                                            class="ModsDetail_LbBgBtn"
                                            :disabled="isSettingBg"
                                            title="Colocar como fondo"
                                            @click.stop="setAsBackground"
                                        >
                                            <IconWallpaper stroke="2" />
                                            <span>{{ isSettingBg ? 'Aplicando…' : 'Colocar como fondo' }}</span>
                                        </button>
                                        <button title="Cerrar (Esc)" @click="closeLightbox">
                                            <IconX stroke="2" />
                                        </button>
                                    </div>
                                </div>
                                <div
                                    ref="stageEl"
                                    class="ModsDetail_LbStage"
                                    :class="{ zoomed: zoom > 1 }"
                                    @wheel.prevent="onWheel"
                                    @mousedown="onPanStart"
                                    @mousemove="onPanMove"
                                    @mouseup="onPanEnd"
                                    @mouseleave="onPanEnd"
                                >
                                    <img
                                        :src="previewSrc"
                                        :alt="previewTitle"
                                        class="ModsDetail_LbImg"
                                        :style="imgStyle"
                                        draggable="false"
                                        @load="onImgLoad"
                                        @error="onPreviewImgError"
                                    />
                                </div>
                                <div v-if="bgSetMsg || bgSetError" class="ModsDetail_LbBgMsg" :class="{ error: !!bgSetError }">
                                    {{ bgSetMsg || bgSetError }}
                                </div>
                                <div class="ModsDetail_LbZoomBar">
                                    <button title="Alejar (−)" @click.stop="biteZoom(-0.25)">
                                        <IconZoomOut stroke="2" />
                                    </button>
                                    <span>{{ Math.round(zoom * 100) }}%</span>
                                    <button title="Acercar (+)" @click.stop="biteZoom(0.25)">
                                        <IconZoomIn stroke="2" />
                                    </button>
                                </div>
                                <button
                                    v-if="lightboxList.length > 1"
                                    class="ModsDetail_LbNav left"
                                    title="Anterior (←)"
                                    @click.stop="lightboxPrev"
                                >
                                    <IconChevronLeft stroke="2" />
                                </button>
                                <button
                                    v-if="lightboxList.length > 1"
                                    class="ModsDetail_LbNav right"
                                    title="Siguiente (→)"
                                    @click.stop="lightboxNext"
                                >
                                    <IconChevronRight stroke="2" />
                                </button>
                            </div>
                        </div>
                    </Transition>
                </div>

                <div v-else-if="hasVersions" class="ModsDetail_VWrap" :class="{ wide: !showChangelog }">
                    <div class="ModsDetail_VListCol">
                        <div class="ModsDetail_VFilters">
                            <select v-model="mcFilter" class="SsSel">
                                <option value="">Todas las versiones de MC</option>
                                <option v-for="gv in projectMcVersions" :key="gv" :value="gv">{{ gv }}</option>
                            </select>
                            <select v-model="typeFilter" class="SsSel">
                                <option value="">Todos los tipos</option>
                                <option v-for="t in VERSION_TYPES" :key="t.id" :value="t.id">{{ t.label }}</option>
                            </select>
                            <select v-model.number="versionsPerPage" class="SsSel ModsDetail_VPerPage" title="Versiones por página">
                                <option v-for="n in VERSIONS_LIMIT_OPTIONS" :key="n" :value="n">{{ n }}/pág</option>
                            </select>
                        </div>

                        <div class="ModsDetail_VList">
                            <article
                                v-for="v in pagedVersions"
                                :key="v.id"
                                class="ModsDetail_VItem"
                                :class="{ on: v.id === selectedVersionId }"
                                @click="selectVersion(v)"
                            >
                                <div class="ModsDetail_VMain">
                                    <span class="ModsDetail_VType" :class="VERSION_TYPE_CLASS[v.version_type]">
                                        {{ versionTypeLabel(v.version_type) }}
                                    </span>
                                    <div class="ModsDetail_VTitles">
                                        <span class="ModsDetail_VName">{{ v.version_number }}</span>
                                        <span class="ModsDetail_VSub">
                                            <IconBox stroke="2" /> {{ mcShort(v.game_versions) }}
                                        </span>
                                    </div>
                                </div>
                                <div class="ModsDetail_VMeta">
                                    <span v-if="v.loaders?.length" class="ModsDetail_VChips">
                                        <template v-for="l in v.loaders.slice(0, 3)" :key="l">
                                            <img
                                                v-if="LOADER_ICONS[l]"
                                                :src="LOADER_ICONS[l]"
                                                :alt="l"
                                                :title="l"
                                            />
                                            <span v-else class="ModsDetail_VChipFallback" :title="l">{{ l }}</span>
                                        </template>
                                    </span>
                                    <span class="ModsDetail_VInfo">
                                        {{ formatDate(v.date_published) }} · {{ formatCount(v.downloads) }} descargas
                                        <template v-if="v.files[0]"> · {{ formatBytes(v.files[0].size) }}</template>
                                    </span>
                                </div>
                            </article>

                            <p v-if="!filteredVersions.length" class="ModsDetail_Plain">
                                No hay versiones con esos filtros.
                            </p>
                        </div>

                        <footer v-if="filteredVersions.length" class="ModsDetail_VFooter">
                            <span class="ModsDetail_VFooterState">
                                Mostrando <b>{{ versionRangeStart }}–{{ versionRangeEnd }}</b> de <b>{{ filteredVersions.length }}</b>
                                <template v-if="totalVersionPages > 1"> · Página <b>{{ versionsPage }}</b> de <b>{{ totalVersionPages }}</b></template>
                            </span>
                            <nav v-if="totalVersionPages > 1" class="ModsDetail_VPages">
                                <button class="ModsDetail_VPgBtn" :disabled="versionsPage <= 1" title="Página anterior" @click="goToVersionPage(versionsPage - 1)">
                                    <IconChevronLeft stroke="2" />
                                </button>
                                <template v-for="(item, i) in versionPageItems" :key="`${item}-${i}`">
                                    <span v-if="item === 'ellipsis'" class="ModsDetail_VPgGap">…</span>
                                    <button v-else class="ModsDetail_VPgBtn" :class="{ on: item === versionsPage }" @click="goToVersionPage(item as number)">{{ item }}</button>
                                </template>
                                <button class="ModsDetail_VPgBtn" :disabled="versionsPage >= totalVersionPages" title="Página siguiente" @click="goToVersionPage(versionsPage + 1)">
                                    <IconChevronRight stroke="2" />
                                </button>
                            </nav>
                        </footer>
                    </div>

                    <div v-if="showChangelog" class="ModsDetail_VChangelog" @click="onMdClick">
                        <h4 class="ModsDetail_ChangelogTitle">
                            Changelog de {{ modrinth.changelogState.versionNumber || selectedVersion?.version_number }}
                        </h4>
                        <p v-if="modrinth.changelogState.loading" class="ModsDetail_ChangelogLoading">
                            <IconLoader2 class="spin" stroke="2" /> Cargando changelog…
                        </p>
                        <p v-else-if="modrinth.changelogState.error" class="ModsDetail_Plain">
                            No se pudo cargar el changelog ({{ modrinth.changelogState.error }})
                        </p>
                        <div v-else-if="changelogHtml" class="ModsDetail_Md" v-html="changelogHtml"></div>
                    </div>
                </div>

                <div v-else class="ModsDetail_Scroll">
                    <p class="ModsDetail_Plain">Este proyecto no tiene versiones publicadas.</p>
                </div>
            </div>
        </template>

        <div v-else class="ModsDetail_Loading">
            <IconLoader2 class="spin" stroke="2" /> Cargando proyecto…
        </div>
    </div>
</template>

<style scoped lang="scss">
@use './Styles/Detail.scss';
</style>
