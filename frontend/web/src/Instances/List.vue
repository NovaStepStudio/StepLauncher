<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import {
    IconPlus, IconStar, IconStarFilled, IconDeviceGamepad, IconPhoto, IconSettings,
    IconDots, IconTrash, IconPencil, IconX, IconClock, IconFolder,
    IconCopy, IconSearch, IconBox, IconPin, IconPinFilled, IconDownload,
} from '@tabler/icons-vue';
import {
    instances,
    sortedInstances,
    details,
    downloads,
    launching,
    loadingList,
    loadInstances,
    loadDetails,
    toggleFavorite,
    togglePin,
    cancelDownload,
    launchInstance,
    formatPlayTime,
    loaderOf,
    loaderLabel,
    loaderDlOf,
    loaderDlStateText,
    isInstanceBusy,
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

function loaderIcon(name: string): string {
    const key = (loaderOf(name)?.loaderType ?? 'vanilla').toLowerCase();
    return LOADER_ICONS[key] ?? iconVanilla;
}

function loaderChipTitle(name: string): string {
    const l = loaderOf(name);
    if (!l) return '';
    return `${loaderLabel(l)} ${l.loaderVersion ?? ''} para ${l.minecraftVersion ?? ''}`;
}

const emit = defineEmits<{
    (e: 'open', name: string): void;
    (e: 'new'): void;
    (e: 'download', name: string): void;
    (e: 'edit', name: string): void;
    (e: 'settings', name: string): void;
    (e: 'clone', name: string): void;
    (e: 'shots', name: string): void;
    (e: 'delete', name: string): void;
}>();

const menuFor = ref<string | null>(null);
const assetUrls = ref<Record<string, { icon?: string; banner?: string }>>({});

const search = ref('');
const filter = ref<'all' | 'fav' | 'pin'>('all');
const groupFilter = ref('');

const groups = computed(() =>
    [...new Set(instances.value.map((i) => i.group).filter(Boolean))]
        .sort((a, b) => a.localeCompare(b))
);

const filtered = computed(() => {
    const q = search.value.trim().toLowerCase();
    return sortedInstances.value.filter((inst) => {
        if (filter.value === 'fav' && !inst.favorite) return false;
        if (filter.value === 'pin' && !inst.pinned) return false;
        if (groupFilter.value && (inst.group || '') !== groupFilter.value) return false;
        if (!q) return true;
        const title = (inst.title || inst.name).toLowerCase();
        return (
            title.includes(q) ||
            inst.name.toLowerCase().includes(q) ||
            (inst.group || '').toLowerCase().includes(q)
        );
    });
});

const totalCount = computed(() => instances.value.length);
const favoritesCount = computed(() => instances.value.filter((i) => i.favorite).length);
const pinsCount = computed(() => instances.value.filter((i) => i.pinned).length);

async function refreshAssets() {
    const next: Record<string, { icon?: string; banner?: string }> = {};
    for (const inst of sortedInstances.value) {
        const d = details.value[inst.name];
        if (!d) continue;
        const entry: { icon?: string; banner?: string } = {};
        if (d.meta.icon) entry.icon = await loadLocal(d.meta.icon);
        if (d.meta.banner) entry.banner = await loadLocal(d.meta.banner);
        next[inst.name] = entry;
    }
    assetUrls.value = next;
}

watch(
    () => sortedInstances.value.map((i) => `${i.name}:${details.value[i.name]?.meta?.icon ?? ''}:${details.value[i.name]?.meta?.banner ?? ''}`),
    () => void refreshAssets(),
    { immediate: true }
);

function openMenu(name: string, e: Event) {
    e.stopPropagation();
    menuFor.value = menuFor.value === name ? null : name;
}

function closeMenuOnClick() {
    menuFor.value = null;
}

async function play(name: string) {
    if (launching.value[name]) return;
    await launchInstance(name);
}

async function toggleStar(name: string) {
    const inst = instances.value.find((i) => i.name === name);
    if (!inst) return;
    const err = await toggleFavorite(name, !inst.favorite);
    if (!err) await loadInstances();
}

async function togglePinned(name: string) {
    const inst = instances.value.find((i) => i.name === name);
    if (!inst) return;
    const err = await togglePin(name, !inst.pinned);
    if (!err) await loadInstances();
}

async function cancelDl(name: string) {
    await cancelDownload(name);
}

async function loadAllDetails() {
    for (const inst of instances.value) void loadDetails(inst.name);
}

onMounted(() => {
    void loadInstances().then(loadAllDetails);
    window.addEventListener('click', closeMenuOnClick);
});

onUnmounted(() => {
    window.removeEventListener('click', closeMenuOnClick);
});
</script>

<template>
    <div class="InstView">
        <header class="InstView_Bar">
            <div class="InstView_BarInfo">
                <h2>Mis instancias</h2>
                <p>
                    {{ totalCount }} {{ totalCount === 1 ? 'instancia' : 'instancias' }}
                    <template v-if="favoritesCount"> · {{ favoritesCount }} favorit{{ favoritesCount === 1 ? 'a' : 'as' }}</template>
                </p>
            </div>
            <div class="InstView_BarActions">
                <label class="InstView_Search">
                    <IconSearch stroke="2" />
                    <input v-model="search" type="text" placeholder="Buscar instancia…" autocomplete="off" spellcheck="false" />
                </label>
                <button class="SsBtn SsBtnPrimary InstView_NewBtn" @click="emit('new')">
                    <IconPlus stroke="2" /> Nueva instancia
                </button>
            </div>
        </header>

        <div class="InstView_Filters" v-if="totalCount">
            <div class="InstView_FilterChips">
                <button class="InstView_FilterChip" :class="{ on: filter === 'all' }" @click="filter = 'all'">
                    Todas
                </button>
                <button class="InstView_FilterChip" :class="{ on: filter === 'fav' }" @click="filter = 'fav'">
                    <IconStar stroke="2" /> Favoritas <em>{{ favoritesCount }}</em>
                </button>
                <button class="InstView_FilterChip" :class="{ on: filter === 'pin' }" @click="filter = 'pin'">
                    <IconPin stroke="2" /> Fijadas <em>{{ pinsCount }}</em>
                </button>
            </div>
            <select v-if="groups.length" class="SsSel InstView_GroupSel" v-model="groupFilter">
                <option value="">Todos los grupos</option>
                <option v-for="g in groups" :key="g" :value="g">{{ g }}</option>
            </select>
        </div>

        <p v-if="loadingList && !sortedInstances.length" class="InstView_Empty">Cargando instancias…</p>

        <div v-else-if="!sortedInstances.length" class="InstView_EmptyCard">
            <span class="InstView_EmptyIcon"><IconBox stroke="1.4" /></span>
            <b class="InstView_EmptyTitle">Aún no tienes instancias</b>
            <p class="InstView_EmptyText">
                Las instancias son mundos de juego independientes con sus propias versiones, modloaders,
                configuraciones y capturas. Los recursos compartidos del launcher se reutilizan automáticamente.
            </p>
            <button class="SsBtn SsBtnPrimary InstView_EmptyBtn" @click="emit('new')">
                <IconPlus stroke="2" /> Crear mi primera instancia
            </button>
        </div>

        <p v-else-if="!filtered.length" class="InstView_Empty">
            No hay instancias que coincidan con «{{ search }}».
        </p>

        <div v-else class="InstView_Grid">
            <article
                v-for="inst in filtered"
                :key="inst.name"
                class="InstCard"
                :class="{ fav: inst.favorite, pin: inst.pinned }"
                @click="emit('open', inst.name)"
            >
                <div class="InstCard_Banner" :class="{ hasImg: assetUrls[inst.name]?.banner }">
                    <img v-if="assetUrls[inst.name]?.banner" :src="assetUrls[inst.name]?.banner" alt="" loading="lazy" />
                    <div class="InstCard_BannerGrad" />
                    <div v-if="downloads[inst.name]" class="InstCard_BannerDl" title="Descarga en curso">
                        <span
                            class="InstCard_BannerDlFill"
                            :style="{ width: (downloads[inst.name]?.percent ?? 0) + '%' }"
                        />
                    </div>
                    <div
                        v-else-if="loaderDlOf(inst.name) && ['resolving', 'downloading', 'installing'].includes(loaderDlOf(inst.name)!.phase)"
                        class="InstCard_Ldr"
                        :title="loaderDlStateText(loaderDlOf(inst.name)!)"
                    >
                        <span class="InstCard_LdrPulse"></span>
                        <span class="InstCard_LdrText">{{ loaderDlStateText(loaderDlOf(inst.name)!) }}</span>
                        <span
                            v-if="loaderDlOf(inst.name)!.total > 0"
                            class="InstCard_LdrBar"
                        ><span :style="{ width: Math.min(100, (loaderDlOf(inst.name)!.progress / loaderDlOf(inst.name)!.total) * 100) + '%' }" /></span>
                    </div>
                    <div class="InstCard_BannerTop">
                        <span v-if="inst.group" class="InstCard_Group" :title="inst.group">{{ inst.group }}</span>
                        <span class="InstCard_BannerBtns">
                            <button
                                class="InstCard_Pin"
                                :class="{ on: inst.pinned }"
                                :title="inst.pinned ? 'Quitar de fijadas' : 'Fijar instancia'"
                                @click.stop="togglePinned(inst.name)"
                            >
                                <IconPinFilled v-if="inst.pinned" stroke="2" />
                                <IconPin v-else stroke="2" />
                            </button>
                            <button
                                class="InstCard_Star"
                                :class="{ on: inst.favorite }"
                                :title="inst.favorite ? 'Quitar de favoritas' : 'Marcar como favorita'"
                                @click.stop="toggleStar(inst.name)"
                            >
                                <IconStarFilled v-if="inst.favorite" stroke="2" />
                                <IconStar v-else stroke="2" />
                            </button>
                        </span>
                    </div>
                </div>

                <div class="InstCard_Body">
                    <div class="InstCard_Line">
                        <span class="InstCard_Icon">
                            <img v-if="assetUrls[inst.name]?.icon" :src="assetUrls[inst.name]?.icon" alt="" />
                            <IconPhoto v-else stroke="1.5" />
                        </span>
                        <div class="InstCard_Titles">
                            <span class="InstCard_Title" :title="inst.title || inst.name">{{ inst.title || inst.name }}</span>
                            <span class="InstCard_Sub">
                                <IconFolder stroke="2" /> {{ inst.name }}
                            </span>
                        </div>
                    </div>

                    <div class="InstCard_Meta">
                        <span v-if="inst.versions.length" class="InstCard_Chip">
                            <IconBox stroke="2" /> {{ inst.versions.length }} versión{{ inst.versions.length !== 1 ? 'es' : '' }}
                        </span>
                        <span v-if="loaderOf(inst.name)" class="InstCard_Chip InstCard_LoaderChip" :title="loaderChipTitle(inst.name)">
                            <img :src="loaderIcon(inst.name)" alt="" /> {{ loaderLabel(loaderOf(inst.name)) }}
                        </span>
                        <span class="InstCard_Chip"><IconClock stroke="2" /> {{ formatPlayTime(inst.playTime) }}</span>
                    </div>

                    <div class="InstCard_Actions">
                        <button
                            class="InstCard_Play"
                            :class="{ disabled: isInstanceBusy(inst.name) }"
                            :disabled="isInstanceBusy(inst.name)"
                            @click.stop="play(inst.name)"
                        >
                            <IconDeviceGamepad stroke="2" />
                            {{ launching[inst.name] ? 'Lanzando…' : isInstanceBusy(inst.name) ? 'Ocupada…' : 'Jugar' }}
                        </button>
                        <button
                            class="InstCard_DlBtn"
                            title="Descargar versión"
                            :disabled="isInstanceBusy(inst.name)"
                            @click.stop="emit('download', inst.name)"
                        >
                            <IconDownload stroke="2" />
                        </button>
                        <span class="InstCard_MenuWrap">
                            <button class="InstCard_MenuDots" title="Opciones" @click="openMenu(inst.name, $event)">
                                <IconDots stroke="2" />
                            </button>
                            <div v-if="menuFor === inst.name" class="InstCard_MenuDrop" @click.stop>
                                <button @click="menuFor = null; emit('edit', inst.name)">
                                    <IconPencil stroke="2" /> Editar
                                </button>
                                <button @click="menuFor = null; emit('clone', inst.name)">
                                    <IconCopy stroke="2" /> Clonar
                                </button>
                                <button @click="menuFor = null; emit('settings', inst.name)">
                                    <IconSettings stroke="2" /> Configurar
                                </button>
                                <button @click="menuFor = null; emit('shots', inst.name)">
                                    <IconPhoto stroke="2" /> Capturas
                                </button>
                                <button class="danger" @click="menuFor = null; emit('delete', inst.name)">
                                    <IconTrash stroke="2" /> Eliminar
                                </button>
                            </div>
                        </span>
                    </div>

                    <div v-if="downloads[inst.name]" class="InstCard_Dl">
                        <div class="InstCard_DlBar">
                            <span
                                class="InstCard_DlFill"
                                :style="{ width: (downloads[inst.name]?.percent ?? 0) + '%' }"
                            />
                        </div>
                        <div class="InstCard_DlMeta">
                            <span v-if="downloads[inst.name]?.state === 'verifying' || downloads[inst.name]?.state === 'redownloading'">
                                Verificando · {{ downloads[inst.name]?.filesDownloaded ?? 0 }}/{{ downloads[inst.name]?.filesTotal ?? 0 }}
                            </span>
                            <span v-else>{{ Math.round(downloads[inst.name]?.percent ?? 0) }}% · {{ downloads[inst.name]?.version }}</span>
                            <button title="Cancelar" @click.stop="cancelDl(inst.name)">
                                <IconX stroke="2" />
                            </button>
                        </div>
                    </div>
                </div>
            </article>
        </div>
    </div>
</template>

<style scoped lang="scss">
@use './Styles/List.scss';
</style>