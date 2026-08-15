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
} from './Store';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

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
    if (view.value !== 'reader') return;
    if (e.key === 'ArrowLeft') {
        e.preventDefault();
        goBack();
    } else if (e.key === 'ArrowRight') {
        e.preventDefault();
        readerForward();
    }
}

useOverlayEscape(close, { isActive: () => props.visible });

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
                                    <img class="News_Spinner" src="../../assets/gif/chicken_jockey_run.gif" alt="">
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
@use './Styles/News.scss';
</style>