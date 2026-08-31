<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import { Browser } from '@wailsio/runtime';
import {
    IconSearch, IconDownload, IconPuzzle, IconPackage, IconPalette, IconWallpaper,
    IconLayoutGrid, IconList, IconRotateClockwise, IconBox, IconLoader2,
    IconChevronLeft, IconChevronRight, IconClock, IconTrash, IconX,
} from '@tabler/icons-vue';
import {
    activeTab, selectTab, MOD_TYPE_TABS,
    LOADER_OPTIONS, SORT_OPTIONS, LIMIT_OPTIONS,
    viewMode, setViewMode, buildModsFacets, environmentLabel,
} from './Store';
import {
    useModrinth, formatCount,
    clearSearchHistory, addSearchHistory, getSearchHistory,
} from '@/Common/Composables/useModrinth';
import type { SearchHit } from '@/Common/Composables/useModrinth';
import { heavyPanel } from '@/Common/Overlays/Store';

import iconFabric from '../../assets/icons/fabric.png';
import iconForge from '../../assets/icons/forge.png';
import iconNeoForge from '../../assets/icons/neoforge.png';
import iconQuilt from '../../assets/icons/quilt.png';
import iconLegacyFabric from '../../assets/icons/legacyfabric.png';

const emit = defineEmits<{
    (e: 'open', slugOrId: string, hit: SearchHit): void;
    (e: 'download', hit: SearchHit): void;
}>();

const LOADER_ICONS: Record<string, string> = {
    fabric: iconFabric,
    forge: iconForge,
    neoforge: iconNeoForge,
    quilt: iconQuilt,
    'legacy-fabric': iconLegacyFabric,
};

const TAB_ICONS: Record<string, unknown> = {
    all: IconPuzzle,
    mod: IconPuzzle,
    modpack: IconPackage,
    shader: IconPalette,
    resourcepack: IconWallpaper,
};

const modrinth = useModrinth();

const query = ref('');
const loader = ref('');
const mcVersion = ref('');
const sortBy = ref<'any' | 'relevance' | 'downloads' | 'follows' | 'newest' | 'updated'>('any');
const limit = ref(20);
const showHistory = ref(false);

// Historial de búsqueda (usa directamente el ref del store para mutaciones reactivas)
const searchHistory = modrinth.searchHistory;

const releases = computed(() => modrinth.tagsState.gameVersions.filter((v) => v.version_type === 'release'));
const snapshots = computed(() => modrinth.tagsState.gameVersions.filter((v) => v.version_type === 'snapshot'));

// Buscador de versión de Minecraft del sidebar (estilo menú de instalación).
const mcQuery = ref('');
const mcTab = ref<'release' | 'snapshot'>('release');

const mcGroups = computed(() => {
    const q = mcQuery.value.trim().toLowerCase();
    const make = (label: string, items: typeof releases.value) => {
        const filtered = items.filter((v) => !q || v.version.toLowerCase().includes(q));
        return { label, items: filtered };
    };
    if (mcTab.value === 'release') return [make('Releases', releases.value)];
    return [make('Snapshots', snapshots.value)];
});

const mcEmpty = computed(() => !mcGroups.value.some((g) => g.items.length));

function runSearch(debounce = 0): void {
    void modrinth.search(
        {
            query: query.value.trim() || undefined,
            // Filtro server-side por project_type vía facets — nunca filtrar hit.project_type en cliente
            facets: buildModsFacets({
                type: activeTab.value,
                loader: loader.value || undefined,
                mcVersion: mcVersion.value || undefined,
            }),
            index: sortBy.value === 'any' ? undefined : sortBy.value,
            limit: limit.value,
        },
        debounce
    );
}

function retry(): void {
    runSearch(0);
}

function selectHistoryItem(q: string): void {
    query.value = q;
    showHistory.value = false;
    runSearch(0);
}

function clearHistory(): void {
    clearSearchHistory();
    // Forzar actualización reactiva
    searchHistory.value.splice(0);
}

function removeHistoryItem(h: string): void {
    const i = searchHistory.value.indexOf(h);
    if (i >= 0) {
        searchHistory.value.splice(i, 1);
        try { localStorage.setItem('modrinth_search_history', JSON.stringify(searchHistory.value)); } catch (_e) {}
    }
}

function onQueryBlur(): void {
    const q = query.value.trim();
    if (q) {
        addSearchHistory(q);
        searchHistory.value = getSearchHistory();
    }
}

// Solo buscar cuando el panel de Mods está visible; evita peticiones al abrir la app
watch([query, loader, mcVersion, activeTab, sortBy, limit], () => {
    if (heavyPanel.value !== 'mods') return;
    runSearch(350);
});

// Calcular skeleton count basado en columnas visibles
const skeletonCount = computed(() => {
    const width = window.innerWidth;
    const sidebarWidth = 232; // 14.5rem
    const padding = 48; // 1.5rem * 2
    const gap = 12.8; // 0.8rem
    const minCardWidth = 250;
    const available = width - sidebarWidth - padding;
    const cols = Math.max(1, Math.floor((available + gap) / (minCardWidth + gap)));
    const rows = 3;
    return cols * rows;
});

function loadersOf(hit: SearchHit): string[] {
    const known = new Set(LOADER_OPTIONS.map((l) => l.id));
    return hit.categories.filter((c) => known.has(c));
}

const page = computed(() => modrinth.searchState.page);
const totalPages = computed(() => modrinth.searchState.totalPages);

const pageItems = computed<(number | 'ellipsis')[]>(() => {
    const total = totalPages.value;
    const cur = page.value;
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

function goToPage(p: number): void {
    void modrinth.goToPage(p);
}

const rangeStart = computed(() => (page.value - 1) * limit.value + 1);
const rangeEnd = computed(() => Math.min(page.value * limit.value, modrinth.searchState.totalHits));

function openSite(url: string): void {
    void Browser.OpenURL(url);
}

function closeHistoryOnClickOutside(e: MouseEvent): void {
    const target = e.target as HTMLElement;
    if (!target.closest('.ModsView_SearchWrap')) {
        showHistory.value = false;
    }
}

let modsInitialized = false;
function ensureModsInit(): void {
    if (modsInitialized) return;
    modsInitialized = true;
    void modrinth.loadTags();
    runSearch(0);
}

onMounted(() => {
    // El panel existe desde el arranque (v-show) pero NO hace fetch hasta entrar.
    if (heavyPanel.value === 'mods') ensureModsInit();
    document.addEventListener('click', closeHistoryOnClickOutside);
});
watch(() => heavyPanel.value, (v) => {
    if (v === 'mods') ensureModsInit();
});

onUnmounted(() => {
    modrinth.abortSearch();
    document.removeEventListener('click', closeHistoryOnClickOutside);
});
</script>

<template>
    <div class="ModsView">
        <header class="ModsView_Bar">
            <div class="ModsView_BarInfo">
                <h2>Explorar Modrinth</h2>
                <p>Contenido para tu cliente de Minecraft</p>
            </div>
            <div class="ModsView_BarActions">
                <div class="ModsView_SearchWrap" @click.stop="showHistory = true">
                    <label class="ModsView_Search">
                        <IconSearch stroke="2" />
                        <input v-model="query" type="text" placeholder="Buscar mods, shaders, texturas…" autocomplete="off" spellcheck="false" @focus="showHistory = true" @blur="onQueryBlur" />
                    </label>
                    <Transition name="ModsHistory">
                        <div v-if="showHistory && searchHistory.length" class="ModsView_History">
                            <div class="ModsView_HistoryHead">
                                <span>Búsquedas recientes</span>
                                <button class="ModsView_HistoryClear" title="Limpiar historial" @click.stop="clearHistory">
                                    <IconTrash stroke="2" />
                                </button>
                            </div>
                            <button v-for="h in searchHistory" :key="h" class="ModsView_HistoryItem" @click.stop="selectHistoryItem(h)">
                                <IconClock stroke="2" />
                                <span>{{ h }}</span>
                                <IconX stroke="1.5" class="ModsView_HistoryRemove" @click.stop="removeHistoryItem(h)" title="Eliminar" />
                            </button>
                        </div>
                    </Transition>
                </div>
                <select class="SsSel ModsView_Sel" v-model="sortBy" title="Ordenar resultados">
                    <option v-for="s in SORT_OPTIONS" :key="s.id" :value="s.id">{{ s.label }}</option>
                    <option value="newest">Más recientes</option>
                    <option value="updated">Actualizados</option>
                </select>
                <select class="SsSel ModsView_Sel" v-model.number="limit" title="Elementos por página">
                    <option v-for="n in LIMIT_OPTIONS" :key="n" :value="n">{{ n }}</option>
                </select>
                <button class="ModsView_Reload" title="Recargar resultados" :disabled="modrinth.searchState.loading" @click="runSearch(0)">
                    <IconRotateClockwise stroke="2" :class="{ spin: modrinth.searchState.loading }" />
                </button>
                <div class="ModsView_ViewToggle" title="Vista">
                    <button :class="{ on: viewMode === 'grid' }" title="Vista en cards" @click="setViewMode('grid')">
                        <IconLayoutGrid stroke="2" />
                    </button>
                    <button :class="{ on: viewMode === 'list' }" title="Vista en filas" @click="setViewMode('list')">
                        <IconList stroke="2" />
                    </button>
                </div>
            </div>
        </header>

        <div class="ModsView_Tabs">
            <button
                v-for="tab in MOD_TYPE_TABS"
                :key="tab.type"
                class="ModsView_Tab"
                :class="{ on: activeTab === tab.type }"
                :title="tab.description"
                @click="selectTab(tab.type)"
            >
                <component :is="TAB_ICONS[tab.type]" stroke="2" />
                {{ tab.label }}
            </button>
        </div>

        <div class="ModsView_Body">
            <aside class="ModsView_Sidebar">
                <section class="ModsView_FilterGroup">
                    <h4 class="ModsView_FilterTitle">ModLoader</h4>
                    <div class="ModsView_LoaderList">
                        <button class="ModsView_LoaderChip" :class="{ on: loader === '' }" @click="loader = ''">
                            Todos
                        </button>
                        <button
                            v-for="l in LOADER_OPTIONS"
                            :key="l.id"
                            class="ModsView_LoaderChip"
                            :class="{ on: loader === l.id }"
                            @click="loader = l.id"
                        >
                            <img v-if="LOADER_ICONS[l.id]" :src="LOADER_ICONS[l.id]" alt="" />
                            {{ l.label }}
                        </button>
                    </div>
                </section>

                <section class="ModsView_FilterGroup">
                    <h4 class="ModsView_FilterTitle">Versión de Minecraft</h4>
                    <div class="ModsView_Vc">
                        <div class="ModsView_VcSearch">
                            <IconSearch stroke="2" />
                            <input v-model="mcQuery" type="text" placeholder="Buscar versión…" autocomplete="off" spellcheck="false" />
                        </div>
                        <div class="ModsView_VcTabs">
                            <button :class="{ on: mcTab === 'release' }" @click="mcTab = 'release'">Releases</button>
                            <button :class="{ on: mcTab === 'snapshot' }" @click="mcTab = 'snapshot'">Snapshots</button>
                        </div>
                        <button
                            class="ModsView_VcAll"
                            :class="{ on: mcVersion === '' }"
                            @click="mcVersion = ''"
                        >
                            Todas las versiones
                        </button>
                        <div class="ModsView_VcList">
                            <template v-for="g in mcGroups" :key="g.label">
                                <div class="ModsView_VcGroupHead">
                                    <span>{{ g.label }}</span>
                                    <em>{{ g.items.length }}</em>
                                </div>
                                <button
                                    v-for="v in g.items"
                                    :key="v.version"
                                    class="ModsView_VcItem"
                                    :class="{ on: mcVersion === v.version }"
                                    @click="mcVersion = v.version"
                                >
                                    <span>{{ v.version }}</span>
                                    <em v-if="v.major">MAYOR</em>
                                </button>
                            </template>
                            <p v-if="mcEmpty" class="ModsView_VcEmpty">
                                Sin resultados para «{{ mcQuery }}»
                            </p>
                        </div>
                    </div>
                    <p v-if="modrinth.tagsState.error" class="ModsView_SidebarError">
                        No se pudieron cargar los filtros ({{ modrinth.tagsState.error }})
                    </p>
                </section>
            </aside>

            <div class="ModsView_Main">
                <div v-if="!modrinth.searchState.error && !modrinth.searchState.hits.length && modrinth.searchState.loading" class="ModsView_Grid">
                    <article v-for="i in skeletonCount" :key="i" class="ModsCard ModsCard_Skeleton">
                        <div class="ModsCard_Body">
                            <div class="ModsCard_Line">
                                <span class="ModsCard_Icon ModsCard_SkIcon" />
                                <div class="ModsCard_Titles">
                                    <div class="ModsCard_SkLine" />
                                    <div class="ModsCard_SkLine short" />
                                </div>
                            </div>
                            <div class="ModsCard_SkLine" />
                            <div class="ModsCard_SkLine short" />
                            <div class="ModsCard_SkChips">
                                <span /><span /><span />
                            </div>
                            <div class="ModsCard_SkBtn" />
                        </div>
                    </article>
                </div>

                <div v-else-if="modrinth.searchState.error" class="ModsView_EmptyCard">
                    <span class="ModsView_EmptyIcon"><IconBox stroke="1.4" /></span>
                    <b class="ModsView_EmptyTitle">No se pudo conectar con Modrinth</b>
                    <p class="ModsView_EmptyText">{{ modrinth.searchState.error }}</p>
                    <button class="SsBtn SsBtnPrimary ModsView_EmptyBtn" @click="retry">
                        <IconRotateClockwise stroke="2" /> Reintentar
                    </button>
                </div>

                <div v-else-if="!modrinth.searchState.hits.length" class="ModsView_EmptyCard">
                    <span class="ModsView_EmptyIcon"><IconSearch stroke="1.4" /></span>
                    <b class="ModsView_EmptyTitle">No se encontraron proyectos</b>
                    <p class="ModsView_EmptyText">
                        Prueba con otra búsqueda o quita algún filtro lateral.
                    </p>
                </div>

                <template v-else>
                    <div v-if="viewMode === 'grid'" class="ModsView_Grid">
                        <article
                            v-for="hit in modrinth.searchState.hits"
                            :key="hit.project_id"
                            class="ModsCard"
                            @click="emit('open', hit.slug ?? hit.project_id, hit)"
                        >
                            <div class="ModsCard_Body">
                                <div class="ModsCard_Line">
                                    <span class="ModsCard_Icon">
                                        <img v-if="hit.icon_url" :src="hit.icon_url" alt="" loading="lazy" decoding="async" />
                                        <IconPuzzle v-else stroke="1.5" />
                                    </span>
                                    <div class="ModsCard_Titles">
                                        <span class="ModsCard_Title" :title="hit.title">{{ hit.title }}</span>
                                        <span class="ModsCard_Sub">por {{ hit.author || 'Desconocido' }}</span>
                                    </div>
                                    <span
                                        class="ModsCard_Type"
                                        :title="MOD_TYPE_TABS.find((t) => t.type === hit.project_type)?.label ?? hit.project_type"
                                    >
                                        <component :is="TAB_ICONS[hit.project_type] ?? IconPuzzle" stroke="2" />
                                    </span>
                                </div>

                                <p class="ModsCard_Desc">{{ hit.description }}</p>

                                <div class="ModsCard_Meta">
                                    <span v-for="l in loadersOf(hit).slice(0, 3)" :key="l" class="ModsCard_Cat" :title="l">
                                        <img v-if="LOADER_ICONS[l]" :src="LOADER_ICONS[l]" alt="" />
                                        {{ l }}
                                    </span>
                                    <span v-if="loadersOf(hit).length > 3" class="ModsCard_Cat">+{{ loadersOf(hit).length - 3 }}</span>
                                    <span v-if="environmentLabel(hit.environment)" class="ModsCard_Cat ModsCard_Env" :title="hit.environment.join(', ')">
                                        {{ environmentLabel(hit.environment) }}
                                    </span>
                                    <span class="ModsCard_Dl">
                                        <IconDownload stroke="2" /> {{ formatCount(hit.downloads ?? 0) }}
                                    </span>
                                </div>

                                <button
                                    class="ModsCard_Download"
                                    :title="`Descargar ${hit.title}`"
                                    @click.stop="emit('download', hit)"
                                >
                                    <IconDownload stroke="2" /> Descargar
                                </button>
                            </div>
                        </article>
                    </div>

                    <div v-else class="ModsView_List">
                        <article
                            v-for="hit in modrinth.searchState.hits"
                            :key="hit.project_id"
                            class="ModsRow"
                            @click="emit('open', hit.slug ?? hit.project_id, hit)"
                        >
                            <span class="ModsRow_Icon">
                                <img v-if="hit.icon_url" :src="hit.icon_url" alt="" loading="lazy" decoding="async" />
                                <IconPuzzle v-else stroke="1.5" />
                            </span>
                            <div class="ModsRow_Info">
                                <span class="ModsRow_Title">{{ hit.title }}</span>
                                <span class="ModsRow_Sub">
                                    {{ hit.author || 'Desconocido' }} · {{ hit.description }}
                                </span>
                            </div>
                            <div class="ModsRow_Chips">
                                <span v-for="l in loadersOf(hit).slice(0, 3)" :key="l" class="ModsRow_Chip" :title="l">
                                    <img v-if="LOADER_ICONS[l]" :src="LOADER_ICONS[l]" alt="" />
                                    {{ l }}
                                </span>
                                <span v-if="environmentLabel(hit.environment)" class="ModsRow_Chip ModsRow_Env" :title="hit.environment.join(', ')">
                                    {{ environmentLabel(hit.environment) }}
                                </span>
                                <span class="ModsRow_Chip ModsRow_Dl">
                                    <IconDownload stroke="2" /> {{ formatCount(hit.downloads ?? 0) }}
                                </span>
                            </div>
                            <button class="ModsRow_DlBtn" :title="`Descargar ${hit.title}`" @click.stop="emit('download', hit)">
                                <IconDownload stroke="2" />
                            </button>
                        </article>
                    </div>

                    <footer class="ModsView_Footer">
                        <div class="ModsView_FooterLeft">
                            <a class="ModsView_Brand" href="#" title="Abrir modrinth.com" @click.prevent="openSite('https://modrinth.com')">
                                <span class="ModsView_BrandLogo">M</span>
                                <span>Contenido de <b>Modrinth</b></span>
                            </a>
                            <span v-if="modrinth.searchState.loading" class="ModsView_FooterState loading">
                                <IconLoader2 class="ModsView_Spin" stroke="2" /> Cargando…
                            </span>
                            <span v-else class="ModsView_FooterState">
                                Mostrando <b>{{ rangeStart }}–{{ rangeEnd }}</b> de <b>{{ modrinth.searchState.totalHits }}</b>
                                <template v-if="totalPages > 1"> · Página <b>{{ page }}</b> de <b>{{ totalPages }}</b></template>
                            </span>
                        </div>

                        <nav v-if="totalPages > 1" class="ModsView_Pages">
                            <button
                                class="ModsView_PgBtn"
                                :disabled="page <= 1 || modrinth.searchState.loading"
                                title="Página anterior"
                                @click="goToPage(page - 1)"
                            >
                                <IconChevronLeft stroke="2" />
                            </button>
                            <template v-for="(item, i) in pageItems" :key="`${item}-${i}`">
                                <span v-if="item === 'ellipsis'" class="ModsView_PgGap">…</span>
                                <button
                                    v-else
                                    class="ModsView_PgBtn"
                                    :class="{ on: item === page }"
                                    :disabled="modrinth.searchState.loading"
                                    @click="goToPage(item)"
                                >
                                    {{ item }}
                                </button>
                            </template>
                            <button
                                class="ModsView_PgBtn"
                                :disabled="page >= totalPages || modrinth.searchState.loading"
                                title="Página siguiente"
                                @click="goToPage(page + 1)"
                            >
                                <IconChevronRight stroke="2" />
                            </button>
                        </nav>
                    </footer>
                </template>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
@use './Styles/Content.scss';
</style>