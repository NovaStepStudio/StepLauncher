<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import {
    IconNews,
    IconX,
    IconRefresh,
    IconAlertCircle,
    IconSearch,
    IconArrowLeft,
    IconArrowRight,
    IconChevronRight,
} from '@tabler/icons-vue';
import { marked } from 'marked';
import { BrowserOpenURL } from '@wailsjs/runtime/runtime';
import {
    indexState,
    indexLoading,
    details,
    detailLoading,
    releaseErrors,
    changelogs,
    changelogLoading,
    changelogErrors,
    docs,
    docLoading,
    docErrors,
    readerStack,
    readerIndex,
    isReading,
    currentEntry,
    currentDocUrl,
    currentMarkdown,
    refreshNews,
    preloadDetails,
    loadChangelog,
    loadMarkdown,
    openReaderVersion,
    openReaderUrl,
    readerBack,
    readerForward,
    readerTo,
    closeReader,
    ensureReaderDoc,
    reloadDetail,
} from '../Stores/News';
import { CLOSE_OVERLAYS_EVENT } from '../Stores/Idle';

const NEWS_BASE_URL = 'https://wpnvconaefhdmvgvqbsv.supabase.co/storage/v1/object/public/news';

marked.use({ gfm: true, breaks: true });

const props = defineProps<{
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
}>();

const TYPE_META: Record<string, { label: string; short: string; cls: string }> = {
    major: { label: 'Actualización importante', short: 'Importante', cls: 'Major' },
    changes: { label: 'Cambios', short: 'Cambios', cls: 'Changes' },
    improvements: { label: 'Mejoras', short: 'Mejoras', cls: 'Improvements' },
    bugfix: { label: 'Correcciones', short: 'Correcciones', cls: 'Bugfix' },
    internal: { label: 'Interno', short: 'Interno', cls: 'Internal' },
};

function typeMeta(type: string) {
    return TYPE_META[type] ?? { label: 'Noticia', short: 'Noticia', cls: 'Changes' };
}

function fmtShortDate(date: string): string {
    if (!date) return '';
    const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(date);
    if (!m) return date;
    try {
        return new Intl.DateTimeFormat('es-ES', { day: 'numeric', month: 'short', year: 'numeric' }).format(
            new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]))
        );
    } catch {
        return date;
    }
}

const view = computed(() => (isReading.value ? 'reader' : 'grid'));

function docFileName(url: string): string {
    try {
        return url.split('/').pop() ?? url;
    } catch {
        return url;
    }
}

function crumbLabel(entry: { version: string; url: string; label: string }): string {
    if (entry.url && entry.label !== entry.url.split('/').pop()) return docFileName(entry.url);
    return entry.label;
}

// ---------- Grilla ----------

const searchQuery = ref('');
const activeType = ref('');

const filteredEntries = computed(() => {
    const q = searchQuery.value.trim().toLowerCase();
    return indexState.value.content.filter((e) => {
        const d = detailOf(e.version);
        if (activeType.value && d && d.type !== activeType.value) return false;
        if (!q) return true;
        if (e.version.toLowerCase().includes(q)) return true;
        if (d?.title && d.title.toLowerCase().includes(q)) return true;
        return false;
    });
});

function detailOf(version: string) {
    return details.get(version);
}

const hasActiveFilters = computed(() => !!searchQuery.value.trim() || !!activeType.value);

function clearFilters() {
    searchQuery.value = '';
    activeType.value = '';
}

function cardState(version: string) {
    const loading = !!detailLoading.get(version);
    const error = releaseErrors.get(version) ?? '';
    return { loading, error };
}

function onCardOpen(version: string) {
    openReaderVersion(version);
}

function retryCard(version: string) {
    reloadDetail(version);
}

const skeletonCards = Array.from({ length: 6 });

// ---------- Lector ----------

const htmlCache = new Map<string, string>();

const markdownHtml = computed(() => {
    const md = currentMarkdown.value;
    if (!md) return '';
    const key = currentDocUrl.value || 'root';
    const hit = htmlCache.get(key);
    if (hit !== undefined) return hit;
    try {
        const out = marked.parse(md);
        const html = typeof out === 'string' ? out : '';
        if (htmlCache.size > 60) htmlCache.clear();
        htmlCache.set(key, html);
        return html;
    } catch {
        return '';
    }
});

const docLabel = computed(() => {
    const i = readerIndex.value;
    const stack = readerStack.value;
    if (i < 0 || i >= stack.length) return '';
    return stack[i]!.label;
});

function docState() {
    const url = currentDocUrl.value;
    const e = currentEntry.value;
    if (!e) return { loading: false, error: '', loaded: false };
    if (url) {
        return {
            loading: !!docLoading.get(url),
            error: docErrors.get(url) ?? '',
            loaded: docs.has(url),
        };
    }
    if (e.version) {
        return {
            loading: !!changelogLoading.get(e.version),
            error: changelogErrors.get(e.version) ?? '',
            loaded: changelogs.has(e.version),
        };
    }
    return { loading: false, error: '', loaded: false };
}

watch(
    () => currentDocUrl.value,
    () => ensureReaderDoc(),
    { immediate: true }
);

function canBack() {
    return readerIndex.value > 0;
}

function canForward() {
    return readerIndex.value >= 0 && readerIndex.value < readerStack.value.length - 1;
}

/** Atrás: sube en el historial del lector; en la raíz vuelve al panel principal de noticias. */
function goBack() {
    if (canBack()) {
        readerBack();
        return;
    }
    closeReader();
}

function isInternalUrl(url: string): boolean {
    return url.startsWith(NEWS_BASE_URL) || !/^https?:\/\//i.test(url);
}

function onDocClick(e: MouseEvent) {
    const target = e.target as HTMLElement | null;
    const anchor = target?.closest?.('a');
    if (!anchor) return;
    e.preventDefault();
    const href = anchor.getAttribute('href') ?? '';
    if (!href || href.startsWith('#')) return;
    const base = currentDocUrl.value || NEWS_BASE_URL;
    let url: string;
    try {
        url = new URL(href, base).toString();
    } catch {
        url = href;
    }
    if (isInternalUrl(url)) {
        if (url.startsWith('http') && !url.startsWith(NEWS_BASE_URL)) {
            try {
                BrowserOpenURL(url);
            } catch { }
            return;
        }
        openReaderUrl(url, docFileName(url));
        return;
    }
    try {
        BrowserOpenURL(url);
    } catch { }
}

function retryCurrent() {
    const state = docState();
    if (state.error) {
        const url = currentDocUrl.value;
        const e = currentEntry.value;
        if (url) {
            docErrors.delete(url);
            loadMarkdown(url);
        } else if (e?.version) {
            changelogErrors.delete(e.version);
            loadChangelog(e.version);
        }
    }
}

// ---------- Ciclo de vida ----------

function close() {
    emit('update:visible', false);
}

function onCloseOverlays() {
    close();
}

function onKeydown(e: KeyboardEvent) {
    if (!props.visible) return;
    if (e.key === 'Escape') {
        close();
        return;
    }
    if (view.value !== 'reader') return;
    if (e.key === 'ArrowLeft') {
        e.preventDefault();
        goBack();
    } else if (e.key === 'ArrowRight') {
        e.preventDefault();
        readerForward();
    }
}

watch(
    () => props.visible,
    (v) => {
        if (v) {
            closeReader();
            refreshNews();
        }
    }
);

watch(indexState, (idx) => {
    if (idx.ok && idx.content.length) {
        preloadDetails();
    }
});

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    window.addEventListener('keydown', onKeydown);
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    window.removeEventListener('keydown', onKeydown);
});
</script>

<template>
    <Teleport to="body">
        <Transition name="NewsModal">
            <div v-if="visible" class="News_Overlay" @click.self="close">
                <div class="News_Dialog">
                    <header class="News_Head">
                        <span class="News_HeadIcon"><IconNews stroke="2" /></span>
                        <div class="News_Titles">
                            <h3>Noticias</h3>
                            <p>{{ view === 'reader' ? (docLabel || 'Leyendo documento') + ' • ' : '' }}Actualizaciones del launcher</p>
                        </div>
                        <button
                            class="News_Refresh"
                            :class="{ loading: indexLoading }"
                            :disabled="indexLoading"
                            title="Buscar novedades"
                            @click="refreshNews"
                        >
                            <IconRefresh stroke="2" />
                        </button>
                        <button class="News_Close" title="Cerrar (Esc)" @click="close">
                            <IconX stroke="2" />
                        </button>
                    </header>

                    <Transition name="NewsView" mode="out-in">
                        <section v-if="view === 'grid'" key="grid" class="News_View">
                            <div class="News_GridToolbar">
                                <div class="News_Search">
                                    <IconSearch class="News_SearchIcon" stroke="2" />
                                    <input
                                        v-model="searchQuery"
                                        type="text"
                                        placeholder="Buscar versión o título…"
                                        spellcheck="false"
                                    />
                                    <button v-if="searchQuery" class="News_SearchClear" title="Limpiar búsqueda" @click="searchQuery = ''">
                                        <IconX stroke="2" />
                                    </button>
                                </div>
                                <div class="News_FilterRow">
                                    <div class="News_FilterChips">
                                        <button
                                            class="News_FilterChip"
                                            :class="{ active: activeType === '' }"
                                            @click="activeType = ''"
                                        >
                                            Todas
                                        </button>
                                        <button
                                            v-for="[type, meta] in Object.entries(TYPE_META)"
                                            :key="type"
                                            class="News_FilterChip"
                                            :class="['News_FilterChip--' + meta.cls, { active: activeType === type }]"
                                            @click="activeType = type"
                                        >
                                            {{ meta.short }}
                                        </button>
                                    </div>
                                    <div class="News_FilterActions">
                                        <span class="News_CountLabel">
                                            {{ filteredEntries.length }}
                                            {{ filteredEntries.length === 1 ? 'novedad' : 'novedades' }}
                                        </span>
                                        <button v-if="hasActiveFilters" class="News_ClearFilters" @click="clearFilters">
                                            Limpiar filtros
                                        </button>
                                    </div>
                                </div>
                            </div>

                            <div class="News_Body">
                                <template v-if="indexLoading && !indexState.content.length">
                                    <div class="News_Grid">
                                        <div v-for="(_, i) in skeletonCards" :key="i" class="News_Card News_Card--skeleton">
                                            <div class="News_CardTop">
                                                <span class="News_Sk News_Sk--chip"></span>
                                                <span class="News_Sk News_Sk--chip News_Sk--chipWide"></span>
                                            </div>
                                            <div class="News_CardBody">
                                                <span class="News_Sk News_Sk--title"></span>
                                                <span class="News_Sk News_Sk--line"></span>
                                                <span class="News_Sk News_Sk--line News_Sk--short"></span>
                                            </div>
                                            <div class="News_CardFoot">
                                                <span class="News_Sk News_Sk--mini"></span>
                                                <span class="News_Sk News_Sk--mini News_Sk--miniWide"></span>
                                            </div>
                                        </div>
                                    </div>
                                </template>

                                <div v-else-if="!indexState.ok" class="News_Empty">
                                    <IconAlertCircle class="News_EmptyIcon" stroke="1.5" />
                                    <p>{{ indexState.error || 'No se pudieron cargar las noticias.' }}</p>
                                    <button class="News_Retry" @click="refreshNews">Reintentar</button>
                                </div>

                                <div v-else-if="!indexState.content.length" class="News_Empty">
                                    <p>Todavía no hay novedades publicadas.</p>
                                </div>

                                <div v-else-if="!filteredEntries.length" class="News_Empty">
                                    <IconSearch class="News_EmptyIcon" stroke="1.5" />
                                    <p>No hay novedades que coincidan con tu búsqueda.</p>
                                    <button class="News_Retry" @click="clearFilters">Limpiar filtros</button>
                                </div>

                                <div v-else class="News_Grid">
                                    <article
                                        v-for="entry in filteredEntries"
                                        :key="entry.version"
                                        class="News_Card"
                                        role="button"
                                        tabindex="0"
                                        :title="details.get(entry.version)?.title"
                                        @click="onCardOpen(entry.version)"
                                        @keydown.enter="onCardOpen(entry.version)"
                                    >
                                        <template v-if="detailOf(entry.version)">
                                            <div class="News_CardTop">
                                                <span class="News_CardVersion">v{{ entry.version }}</span>
                                                <span
                                                    v-if="entry.version === indexState.latest"
                                                    class="News_LatestTag"
                                                >
                                                    Última
                                                </span>
                                                <span
                                                    class="News_TypeTag"
                                                    :class="'News_TypeTag--' + typeMeta(detailOf(entry.version)!.type).cls"
                                                >
                                                    {{ typeMeta(detailOf(entry.version)!.type).short }}
                                                </span>
                                            </div>
                                            <div class="News_CardBody">
                                                <span class="News_CardTitle">{{ detailOf(entry.version)!.title }}</span>
                                                <span v-if="detailOf(entry.version)!.body" class="News_CardText">
                                                    {{ detailOf(entry.version)!.body }}
                                                </span>
                                            </div>
                                            <footer class="News_CardFoot">
                                                <span class="News_CardDate">{{ fmtShortDate(detailOf(entry.version)!.date) }}</span>
                                                <span class="News_CardOpen">
                                                    Abrir <IconChevronRight class="News_CardArrow" stroke="2" />
                                                </span>
                                            </footer>
                                        </template>

                                        <template v-else-if="cardState(entry.version).error">
                                            <div class="News_CardTop">
                                                <span class="News_CardVersion">v{{ entry.version }}</span>
                                            </div>
                                            <div class="News_CardBody">
                                                <span class="News_CardError">{{ cardState(entry.version).error }}</span>
                                            </div>
                                            <footer class="News_CardFoot">
                                                <span class="News_CardDate">—</span>
                                                <button class="News_CardRetry" @click.stop="retryCard(entry.version)">
                                                    Reintentar
                                                </button>
                                            </footer>
                                        </template>

                                        <template v-else>
                                            <div class="News_CardTop">
                                                <span class="News_Sk News_Sk--chip"></span>
                                                <span class="News_Sk News_Sk--chip News_Sk--chipWide"></span>
                                            </div>
                                            <div class="News_CardBody">
                                                <span class="News_Sk News_Sk--title"></span>
                                                <span class="News_Sk News_Sk--line"></span>
                                                <span class="News_Sk News_Sk--line News_Sk--short"></span>
                                            </div>
                                            <div class="News_CardFoot">
                                                <span class="News_Sk News_Sk--mini"></span>
                                                <span class="News_Sk News_Sk--mini News_Sk--miniWide"></span>
                                            </div>
                                        </template>
                                    </article>
                                </div>
                            </div>
                        </section>

                        <section v-else key="reader" class="News_View">
                            <div class="News_ReaderBar">
                                <div class="News_ReaderNav">
                                    <button
                                        class="News_BackPill"
                                        :title="canBack() ? 'Retroceder (←)' : 'Volver al panel de noticias (←)'"
                                        @click="goBack"
                                    >
                                        <IconArrowLeft stroke="2" />
                                        <span>{{ canBack() ? 'Volver' : 'Noticias' }}</span>
                                    </button>
                                    <nav class="News_Breadcrumb">
                                        <span class="News_CrumbSep">/</span>
                                        <template v-if="readerStack.length > 1">
                                            <template v-for="(entry, i) in readerStack.slice(0, -1)" :key="i">
                                                <button class="News_Crumb" :title="entry.label" @click="readerTo(i)">
                                                    {{ crumbLabel(entry) }}
                                                </button>
                                                <span class="News_CrumbSep">/</span>
                                            </template>
                                        </template>
                                        <span v-if="readerStack.length" class="News_Crumb News_Crumb--current" :title="docLabel">
                                            {{ docLabel }}
                                        </span>
                                    </nav>
                                </div>
                                <div class="News_ReaderEnd">
                                    <button
                                        class="News_NavBtn"
                                        :disabled="!canForward()"
                                        title="Adelante (→)"
                                        @click="readerForward"
                                    >
                                        <IconArrowRight stroke="2" />
                                    </button>
                                </div>
                            </div>

                            <div class="News_ReaderBody">
                                <div v-if="docState().loading" class="News_Empty">
                                    <img class="News_Spinner" src="../../assets/gif/loading.gif" alt="">
                                    <p>Cargando documento…</p>
                                </div>
                                <div v-else-if="docState().error" class="News_Empty">
                                    <IconAlertCircle class="News_EmptyIcon" stroke="1.5" />
                                    <p>{{ docState().error }}</p>
                                    <button class="News_Retry" @click="retryCurrent">Reintentar</button>
                                </div>
                                <div v-else-if="markdownHtml" class="News_Md" v-html="markdownHtml" @click="onDocClick"></div>
                                <div v-else class="News_Empty">
                                    <p>Este documento no trae contenido.</p>
                                </div>
                            </div>
                        </section>
                    </Transition>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
.NewsModal-enter-active,
.NewsModal-leave-active {
    transition: opacity 160ms ease;
}
.NewsModal-enter-from,
.NewsModal-leave-to {
    opacity: 0;
}

/* ---------- Contenedor ---------- */

.News_Overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: var(--filter-blur);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 130;
}

.News_Dialog {
    width: 100%;
    height: 100%;
    background: var(--background-modal-primary);
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.News_View {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
}

.NewsView-enter-active,
.NewsView-leave-active {
    transition: opacity 130ms ease, transform 130ms ease;
}
.NewsView-enter-from {
    opacity: 0;
    transform: translateX(10px);
}
.NewsView-leave-to {
    opacity: 0;
    transform: translateX(-10px);
}

/* ---------- Cabecera ---------- */

.News_Head {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1.1rem 1.35rem 0.85rem;
    border-bottom: var(--border-modal-style);
    background: #0005;
    flex-shrink: 0;
}

.News_HeadIcon {
    width: 2.4rem;
    height: 2.4rem;
    flex-shrink: 0;
    border-radius: 0.6rem;
    background: linear-gradient(135deg, color-mix(in srgb, var(--background-button-primary) 35%, transparent), color-mix(in srgb, var(--background-button-primary) 10%, transparent));
    border: 1px solid color-mix(in srgb, var(--background-button-primary) 30%, transparent);
    display: flex;
    justify-content: center;
    align-items: center;
    color: color-mix(in srgb, var(--background-button-primary) 55%, white 45%);

    svg {
        width: 1.2rem;
        height: 1.2rem;
    }
}

.News_Titles {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    min-width: 0;

    h3 {
        font-family: var(--font-primary), Arial, sans-serif;
        font-size: 1.05rem;
        margin: 0;
        color: var(--text-primary);
    }

    p {
        font-size: 0.72rem;
        margin: 0;
        opacity: 0.55;
        color: var(--text-secondary);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
}

.News_Refresh,
.News_Close {
    width: 2rem;
    height: 2rem;
    flex-shrink: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    border-radius: 0.5rem;
    border: 1px solid var(--control-border);
    background: var(--control-bg);
    color: var(--text-primary);
    cursor: pointer;
    transition: background 120ms, border-color 120ms, color 120ms;

    svg {
        width: 1rem;
        height: 1rem;
    }

    &:disabled {
        opacity: 0.6;
        cursor: default;
    }
}

.News_Refresh:hover:not(:disabled) {
    background: color-mix(in srgb, var(--background-button-primary) 18%, transparent);
    border-color: color-mix(in srgb, var(--background-button-primary) 45%, transparent);
}

.News_Refresh.loading svg {
    animation: NewsSpin 0.8s linear infinite;
}

@keyframes NewsSpin {
    to {
        transform: rotate(360deg);
    }
}

.News_Close:hover {
    background: color-mix(in srgb, var(--color-error) 20%, transparent);
    border-color: color-mix(in srgb, var(--color-error) 40%, transparent);
    color: var(--color-error);
}

/* ---------- Toolbar (grilla) ---------- */

.News_GridToolbar {
    display: flex;
    flex-direction: column;
    gap: 0.55rem;
    padding: 0.75rem 1.35rem 0.6rem;
    border-bottom: var(--border-modal-style);
    background: #0003;
    flex-shrink: 0;
}

.News_Search {
    position: relative;
    display: flex;
    align-items: center;

    input {
        width: 100%;
        padding: 0.5rem 2.2rem 0.5rem 2.1rem;
        border-radius: 0.55rem;
        border: 1px solid var(--control-border);
        background: var(--control-bg);
        color: var(--text-primary);
        text-shadow: var(--text-shadow-primary, none);
        font-family: var(--font-secundary), Arial, sans-serif;
        font-size: 0.75rem;
        outline: none;
        transition: border-color 140ms, background 140ms;

        &::placeholder {
            color: var(--text-secondary);
            opacity: 0.6;
        }

        &:focus {
            border-color: color-mix(in srgb, var(--background-button-primary) 55%, transparent);
            background: color-mix(in srgb, var(--control-bg) 78%, var(--background-button-primary) 6%);
        }
    }
}

.News_SearchIcon {
    position: absolute;
    left: 0.7rem;
    width: 0.95rem;
    height: 0.95rem;
    color: var(--text-secondary);
    opacity: 0.7;
    pointer-events: none;
}

.News_SearchClear {
    position: absolute;
    right: 0.45rem;
    width: 1.5rem;
    height: 1.5rem;
    display: flex;
    justify-content: center;
    align-items: center;
    border: none;
    border-radius: 0.4rem;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;

    svg {
        width: 0.85rem;
        height: 0.85rem;
    }

    &:hover {
        background: color-mix(in srgb, var(--color-error) 16%, transparent);
        color: var(--color-error);
    }
}

.News_FilterRow {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.6rem;
    flex-wrap: wrap;
}

.News_FilterChips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
    min-width: 0;
}

.News_FilterActions {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    margin-left: auto;
    flex-shrink: 0;
}

.News_FilterChip {
    font-family: var(--font-secundary), Arial, sans-serif;
    font-size: 0.62rem;
    font-weight: 600;
    letter-spacing: 0.02em;
    padding: 0.32rem 0.8rem;
    border-radius: 99rem;
    border: 1px solid var(--control-border);
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    transition: background 130ms, border-color 130ms, color 130ms;

    &:hover {
        border-color: color-mix(in srgb, var(--background-button-primary) 40%, transparent);
        color: var(--text-primary);
    }

    &.active {
        border-color: color-mix(in srgb, var(--background-button-primary) 55%, transparent);
        background: color-mix(in srgb, var(--background-button-primary) 20%, transparent);
        color: var(--text-primary);
    }

    &--Major.active {
        border-color: color-mix(in srgb, var(--color-error) 50%, transparent);
        background: color-mix(in srgb, var(--color-error) 16%, transparent);
        color: var(--color-error);
    }

    &--Changes.active {
        border-color: color-mix(in srgb, var(--color-tag) 50%, transparent);
        background: color-mix(in srgb, var(--color-tag) 16%, transparent);
        color: var(--color-tag);
    }

    &--Improvements.active {
        border-color: color-mix(in srgb, var(--color-success) 50%, transparent);
        background: color-mix(in srgb, var(--color-success) 16%, transparent);
        color: var(--color-success);
    }

    &--Bugfix.active {
        border-color: color-mix(in srgb, var(--color-warning) 50%, transparent);
        background: color-mix(in srgb, var(--color-warning) 16%, transparent);
        color: var(--color-warning);
    }

    &--Internal.active {
        border-color: color-mix(in srgb, var(--text-secondary) 45%, transparent);
        background: color-mix(in srgb, var(--text-secondary) 12%, transparent);
        color: var(--text-secondary);
    }
}

.News_CountLabel {
    font-size: 0.64rem;
    font-weight: 600;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    opacity: 0.55;
    color: var(--text-secondary);
    white-space: nowrap;
}

.News_ClearFilters {
    font-family: var(--font-secundary), Arial, sans-serif;
    font-size: 0.64rem;
    font-weight: 600;
    padding: 0.2rem 0.55rem;
    border-radius: 0.4rem;
    border: 1px solid var(--control-border);
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    transition: background 130ms, border-color 130ms, color 130ms;

    &:hover {
        border-color: color-mix(in srgb, var(--color-error) 40%, transparent);
        background: color-mix(in srgb, var(--color-error) 12%, transparent);
        color: var(--color-error);
    }
}

/* ---------- Grilla de cards ---------- */

.News_Body {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 1.1rem 1.4rem 1.6rem;
}

.News_Grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(min(13.5rem, 100%), 1fr));
    gap: 0.9rem;
    max-width: 96rem;
    margin: 0 auto;
}

.News_Card {
    display: flex;
    flex-direction: column;
    gap: 0.7rem;
    padding: 1rem;
    border-radius: 0.85rem;
    border: 1px solid var(--control-border-strong, var(--control-border));
    background: linear-gradient(180deg, var(--control-bg-soft) 0%, var(--control-bg) 100%);
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
    font-family: var(--font-secundary), Arial, sans-serif;
    text-align: left;
    cursor: pointer;
    transition: border-color 150ms, background 150ms, transform 150ms, box-shadow 150ms;

    &:hover,
    &:focus-visible {
        border-color: color-mix(in srgb, var(--background-button-primary) 55%, transparent);
        background: color-mix(in srgb, var(--background-button-primary) 8%, transparent);
        transform: translateY(-3px);
        box-shadow: var(--shadow-settings-normal) #0006;

        .News_CardArrow {
            transform: translateX(2px);
        }
    }
}

.News_CardTop {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    flex-wrap: wrap;
}

.News_CardVersion {
    font-family: var(--font-primary), Arial, sans-serif;
    font-size: 0.72rem;
    font-weight: 600;
    padding: 0.26rem 0.5rem;
    border-radius: 0.45rem;
    background: color-mix(in srgb, var(--background-button-primary) 25%, transparent);
    border: 1px solid color-mix(in srgb, var(--background-button-primary) 40%, transparent);
    color: color-mix(in srgb, var(--background-button-primary) 85%, white 45%);
}

.News_LatestTag {
    font-size: 0.52rem;
    font-weight: 800;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    padding: 0.18rem 0.45rem;
    margin-left: auto;
    border-radius: 99rem;
    background: color-mix(in srgb, var(--color-success) 16%, transparent);
    border: 1px solid color-mix(in srgb, var(--color-success) 38%, transparent);
    color: var(--color-success);
}

.News_CardBody {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    flex: 1;
}

.News_CardTitle {
    font-family: var(--font-primary), Arial, sans-serif;
    font-size: 0.95rem;
    font-weight: 600;
    line-height: 1.35;
    color: var(--text-primary);
}

.News_CardText {
    font-size: 0.72rem;
    line-height: 1.55;
    color: var(--text-secondary);
    opacity: 0.9;
    overflow-wrap: break-word;
}

.News_CardError {
    font-size: 0.72rem;
    line-height: 1.5;
    color: var(--color-error);
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
    overflow: hidden;
}

.News_CardFoot {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    margin-top: auto;
    padding-top: 0.55rem;
    border-top: 1px solid var(--control-border);
}

.News_CardDate {
    font-size: 0.66rem;
    opacity: 0.6;
    color: var(--text-secondary);
}

.News_CardOpen {
    display: flex;
    align-items: center;
    gap: 0.2rem;
    font-size: 0.66rem;
    font-weight: 700;
    letter-spacing: 0.03em;
    text-transform: uppercase;
    color: color-mix(in srgb, var(--background-button-primary) 80%, white 20%);
}

.News_CardRetry {
    font-family: var(--font-secundary), Arial, sans-serif;
    font-size: 0.62rem;
    font-weight: 700;
    padding: 0.25rem 0.6rem;
    border-radius: 0.45rem;
    border: 1px solid color-mix(in srgb, var(--color-error) 40%, transparent);
    background: color-mix(in srgb, var(--color-error) 14%, transparent);
    color: var(--color-error);
    cursor: pointer;
    transition: background 130ms, border-color 130ms;

    &:hover {
        background: color-mix(in srgb, var(--color-error) 24%, transparent);
        border-color: color-mix(in srgb, var(--color-error) 60%, transparent);
    }
}

.News_CardArrow {
    width: 0.85rem;
    height: 0.85rem;
    transition: transform 150ms;
}

/* ---------- Skeletons ---------- */

.News_Card--skeleton {
    cursor: default;
    pointer-events: none;
}

.News_Sk {
    position: relative;
    display: block;
    overflow: hidden;
    border-radius: 0.35rem;
    background: linear-gradient(
        90deg,
        var(--control-bg) 25%,
        color-mix(in srgb, var(--control-bg-soft) 70%, transparent) 50%,
        var(--control-bg) 75%
    );
    background-size: 200% 100%;
    animation: NewsShimmer 1.3s linear infinite;

    &--chip {
        width: 3.4rem;
        height: 1.15rem;
    }

    &--chipWide {
        width: 5rem;
    }

    &--title {
        width: 82%;
        height: 0.85rem;
    }

    &--line {
        width: 100%;
        height: 0.6rem;
    }

    &--short {
        width: 58%;
    }

    &--mini {
        width: 4.5rem;
        height: 0.7rem;
    }

    &--miniWide {
        width: 6rem;
    }
}

@keyframes NewsShimmer {
    0% {
        background-position: 200% 0;
    }
    100% {
        background-position: -200% 0;
    }
}

/* ---------- Lector ---------- */

.News_ReaderBar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.8rem;
    padding: 0.6rem 1.35rem;
    border-bottom: var(--border-modal-style);
    background: #0003;
    flex-shrink: 0;
}

.News_ReaderNav {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    min-width: 0;
}

.News_BackPill {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    height: 2rem;
    padding: 0 0.75rem;
    flex-shrink: 0;
    border-radius: 0.5rem;
    border: 1px solid var(--control-border);
    background: var(--control-bg);
    color: var(--text-primary);
    font-family: var(--font-secundary), Arial, sans-serif;
    font-size: 0.72rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 120ms, border-color 120ms, color 120ms;

    svg {
        width: 0.95rem;
        height: 0.95rem;
    }

    &:hover {
        border-color: color-mix(in srgb, var(--background-button-primary) 45%, transparent);
        background: color-mix(in srgb, var(--background-button-primary) 16%, transparent);
    }
}

.News_Breadcrumb {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    min-width: 0;
    flex-wrap: nowrap;
}

.News_CrumbSep {
    opacity: 0.35;
    font-size: 0.72rem;
    color: var(--text-secondary);
    flex-shrink: 0;
}

.News_Crumb {
    font-family: var(--font-secundary), Arial, sans-serif;
    font-size: 0.74rem;
    font-weight: 600;
    color: var(--text-secondary);
    background: transparent;
    border: none;
    cursor: pointer;
    padding: 0.1rem 0.2rem;
    white-space: nowrap;
    max-width: 15rem;
    overflow: hidden;
    text-overflow: ellipsis;

    &:hover:not(.News_Crumb--current) {
        color: var(--text-primary);
        text-decoration: underline;
    }

    &--current {
        color: var(--text-primary);
        cursor: default;
        pointer-events: none;
    }
}

.News_ReaderEnd {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    flex-shrink: 0;
}

.News_NavBtn {
    width: 2rem;
    height: 2rem;
    flex-shrink: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    border-radius: 0.5rem;
    border: 1px solid var(--control-border);
    background: var(--control-bg);
    color: var(--text-primary);
    cursor: pointer;
    transition: background 120ms, border-color 120ms, color 120ms;

    svg {
        width: 1rem;
        height: 1rem;
    }

    &:hover:not(:disabled) {
        border-color: color-mix(in srgb, var(--background-button-primary) 45%, transparent);
        background: color-mix(in srgb, var(--background-button-primary) 16%, transparent);
    }

    &:disabled {
        opacity: 0.35;
        cursor: default;
        pointer-events: none;
    }
}

.News_ReaderBody {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 1.4rem 1.4rem 2.5rem;
}

.News_Md {
    max-width: 60rem;
    margin: 0 auto;
    padding: 0.2rem 0;
    font-family: var(--font-secundary), Arial, sans-serif;
    font-size: 0.86rem;
    line-height: 1.75;
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
    overflow-wrap: break-word;

    > :first-child {
        margin-top: 0 !important;
    }

    :deep(h1),
    :deep(h2),
    :deep(h3),
    :deep(h4),
    :deep(h5),
    :deep(h6) {
        font-family: var(--font-primary), Arial, sans-serif;
        color: var(--text-primary);
        line-height: 1.3;
        text-shadow: var(--text-shadow-primary, none);
        margin: 1.7em 0 0.55em;
    }

    :deep(h1) {
        font-size: 1.6rem;
        padding-bottom: 0.4rem;
        border-bottom: 1px solid color-mix(in srgb, var(--text-secondary) 28%, transparent);
        margin-bottom: 1em;
    }

    :deep(h2) {
        font-size: 1.22rem;
        margin-top: 1.9em;
    }

    :deep(h3) {
        font-size: 1.02rem;
    }

    :deep(h4) {
        font-size: 0.9rem;
        opacity: 0.9;
    }

    :deep(h5),
    :deep(h6) {
        font-size: 0.82rem;
        opacity: 0.85;
        text-transform: uppercase;
        letter-spacing: 0.04em;
    }

    :deep(p) {
        margin: 0 0 0.6em;
    }

    :deep(ul),
    :deep(ol) {
        margin: 0 0 0.7em;
        padding-left: 1.5em;

        li {
            margin: 0.28em 0;

            &::marker {
                color: var(--text-secondary);
                font-weight: 600;
            }

            & > p {
                margin-bottom: 0.3em;
            }
        }
    }

    :deep(li > ul),
    :deep(li > ol) {
        margin: 0.2em 0;
    }

    :deep(a) {
        color: color-mix(in srgb, var(--color-tag) 60%, var(--text-primary) 40%);
        text-decoration: underline;
        text-underline-offset: 3px;
        text-decoration-thickness: 1px;
        cursor: pointer;
        border-radius: 0.2rem;
        transition: color 120ms, background 120ms;

        &:hover {
            color: var(--text-primary);
            background: color-mix(in srgb, var(--color-tag) 16%, transparent);
        }

        &[href$='.md']::after,
        &[href*='.md?']::after {
            content: ' ⇢';
            opacity: 0.75;
            font-size: 0.85em;
        }
    }

    :deep(strong) {
        font-weight: 700;
        color: var(--text-primary);
    }

    :deep(em) {
        opacity: 0.92;
    }

    :deep(code) {
        font-family: 'Consolas', 'Courier New', monospace;
        font-size: 0.82em;
        padding: 0.18em 0.45em;
        border-radius: 0.35rem;
        background: color-mix(in srgb, var(--control-bg) 80%, black 20%);
        border: 1px solid var(--control-border);
        color: color-mix(in srgb, var(--color-tag) 78%, white 22%);
    }

    :deep(pre) {
        margin: 0.9em 0;
        padding: 1rem 1.15rem;
        border-radius: 0.7rem;
        background: #000a;
        border: 1px solid var(--control-border);
        box-shadow: var(--shadow-settings-normal) #0005;
        overflow-x: auto;

        code {
            padding: 0;
            background: transparent;
            border: none;
            color: var(--text-primary);
            font-size: 0.8em;
            line-height: 1.6;
            white-space: pre;
        }
    }

    :deep(blockquote) {
        margin: 0.9em 0;
        padding: 0.7rem 1.1rem;
        border-left: 3px solid color-mix(in srgb, var(--color-tag) 60%, transparent);
        background: var(--control-bg);
        border-radius: 0.3rem 0.5rem 0.5rem 0.3rem;
        color: var(--text-secondary);

        p {
            margin-bottom: 0.3em;
        }
    }

    :deep(hr) {
        margin: 1.6em 0;
        border: none;
        height: 1px;
        background: linear-gradient(90deg, transparent, var(--control-border) 20%, var(--control-border) 80%, transparent);
    }

    :deep(img) {
        max-width: 100%;
        border-radius: 0.6rem;
        border: 1px solid var(--control-border);
    }

    :deep(table) {
        margin: 1em 0;
        border-collapse: collapse;
        font-size: 0.82em;
        width: 100%;
        border-radius: 0.5rem;
        overflow: hidden;

        th,
        td {
            border: 1px solid var(--control-border);
            padding: 0.45rem 0.75rem;
            text-align: left;
        }

        th {
            background: var(--control-bg);
            font-weight: 700;
            border-top: none;
            border-left: none;
        }

        tr:nth-child(even) td {
            background: var(--control-bg-soft);
        }
    }

    :deep(input[type='checkbox']) {
        margin-right: 0.45em;
        accent-color: var(--color-tag);
    }
}

/* ---------- Estados ---------- */

.News_Empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.6rem;
    padding: 2.5rem 1rem;
    text-align: center;
    opacity: 0.65;
    font-size: 0.82rem;

    p {
        margin: 0;
    }
}

.News_EmptyIcon {
    width: 2.2rem;
    height: 2.2rem;
    opacity: 0.5;
}

.News_Spinner {
    width: 1.1rem;
    height: auto;
    image-rendering: pixelated;
    margin: 0.2rem 0;
}

.News_Retry {
    font-family: var(--font-secundary), Arial, sans-serif;
    font-size: 0.72rem;
    font-weight: 600;
    padding: 0.45rem 1rem;
    border-radius: 0.5rem;
    border: 1px solid color-mix(in srgb, var(--background-button-primary) 45%, transparent);
    background: color-mix(in srgb, var(--background-button-primary) 18%, transparent);
    color: var(--text-primary);
    cursor: pointer;
    transition: background 140ms, border-color 140ms;

    &:hover {
        background: color-mix(in srgb, var(--background-button-primary) 30%, transparent);
        border-color: color-mix(in srgb, var(--background-button-primary) 65%, transparent);
    }
}

.News_TypeTag {
    font-size: 0.5rem;
    font-weight: 700;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    padding: 0.18rem 0.55rem;
    border-radius: 99rem;
    border: 1px solid transparent;
    white-space: nowrap;

    &--Major {
        background: color-mix(in srgb, var(--color-error) 14%, transparent);
        border-color: color-mix(in srgb, var(--color-error) 35%, transparent);
        color: var(--color-error);
    }

    &--Changes {
        background: color-mix(in srgb, var(--color-tag) 14%, transparent);
        border-color: color-mix(in srgb, var(--color-tag) 35%, transparent);
        color: var(--color-tag);
    }

    &--Improvements {
        background: color-mix(in srgb, var(--color-success) 14%, transparent);
        border-color: color-mix(in srgb, var(--color-success) 35%, transparent);
        color: var(--color-success);
    }

    &--Bugfix {
        background: color-mix(in srgb, var(--color-warning) 14%, transparent);
        border-color: color-mix(in srgb, var(--color-warning) 35%, transparent);
        color: var(--color-warning);
    }

    &--Internal {
        background: color-mix(in srgb, var(--text-secondary) 12%, transparent);
        border-color: color-mix(in srgb, var(--text-secondary) 30%, transparent);
        color: var(--text-secondary);
    }
}

/* ---------- Ajustes para ventanas pequeñas (escritorio) ---------- */

@media (max-width: 1280px) {
    .News_Grid {
        grid-template-columns: repeat(auto-fill, minmax(min(13rem, 100%), 1fr));
    }
}

@media (max-width: 1120px) {
    .News_Head,
    .News_GridToolbar,
    .News_ReaderBar {
        padding-left: 1.1rem;
        padding-right: 1.1rem;
    }

    .News_Body,
    .News_ReaderBody {
        padding-left: 1.1rem;
        padding-right: 1.1rem;
    }

    .News_Grid {
        grid-template-columns: repeat(auto-fill, minmax(min(12.5rem, 100%), 1fr));
    }
}

@media (max-width: 980px) {
    .News_Grid {
        grid-template-columns: repeat(auto-fill, minmax(min(13.5rem, 100%), 1fr));
        gap: 0.75rem;
    }

    .News_Md {
        font-size: 0.82rem;
        max-width: 100%;
    }
}

@media (max-width: 860px) {
    .News_Head {
        padding: 0.9rem 0.95rem 0.75rem;
        gap: 0.6rem;
    }

    .News_HeadIcon {
        width: 2.1rem;
        height: 2.1rem;
    }

    .News_Titles p {
        font-size: 0.66rem;
    }

    .News_GridToolbar {
        padding: 0.65rem 0.95rem 0.55rem;
    }

    .News_FilterActions {
        margin-left: 0;
    }

    .News_FilterRow {
        gap: 0.35rem;
    }

    .News_ReaderBar {
        padding: 0.5rem 0.95rem;
    }

    .News_Body,
    .News_ReaderBody {
        padding: 0.9rem 0.95rem 1.4rem;
    }

    .News_Grid {
        grid-template-columns: repeat(auto-fill, minmax(min(11.5rem, 100%), 1fr));
        gap: 0.7rem;
    }

    .News_BackPill span {
        display: none;
    }

    .News_BackPill {
        padding: 0 0.55rem;
    }

    .News_Crumb {
        max-width: 10rem;
    }
}

@media (max-width: 680px) {
    .News_Head,
    .News_GridToolbar,
    .News_ReaderBar,
    .News_Body,
    .News_ReaderBody {
        padding-left: 0.8rem;
        padding-right: 0.8rem;
    }

    .News_Grid {
        grid-template-columns: 1fr;
        gap: 0.65rem;
    }

    .News_Crumb {
        max-width: 7rem;
    }

    .News_Card {
        padding: 0.9rem;
    }
}

@media (max-height: 720px) {
    .News_Head {
        padding-top: 0.75rem;
        padding-bottom: 0.6rem;
    }

    .News_GridToolbar {
        padding-top: 0.55rem;
        padding-bottom: 0.45rem;
    }

    .News_ReaderBar {
        padding-top: 0.45rem;
        padding-bottom: 0.45rem;
    }

    .News_Body,
    .News_ReaderBody {
        padding-top: 0.8rem;
    }
}
</style>