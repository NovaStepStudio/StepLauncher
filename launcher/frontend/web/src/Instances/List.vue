<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import {
    IconPlus, IconStar, IconStarFilled, IconDeviceGamepad, IconPhoto, IconSettings,
    IconDots, IconTrash, IconPencil, IconX, IconClock, IconFolder,
    IconCopy, IconSearch, IconBox, IconPin, IconPinFilled, IconDownload,
    IconLoader2, IconPlayerStop, IconEye, IconEyeOff, IconAlertTriangle, IconCheck,
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
    isInstanceVerifying,
    provisioning,
    loadProvisioning,
    contentDls,
    cancelContentDownload,
    failedContent,
    dismissFailedContent,
} from './Store';
import { loadLocal } from '@/Common/Stores/Ui';
import { isOffline } from '@/Common/Stores/Connectivity';
import OfflineBadge from '@/Common/Components/OfflineBadge.vue';

import iconVanilla from '../../assets/icons/minecraft.webp';
import iconFabric from '../../assets/icons/fabric.webp';
import iconForge from '../../assets/icons/forge.webp';
import iconNeoForge from '../../assets/icons/neoforge.webp';
import iconQuilt from '../../assets/icons/quilt.webp';
import iconLegacyFabric from '../../assets/icons/legacyfabric.webp';

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
// Las "Creando…" se pueden ocultar de la rejilla sin cancelarlas.
const hideCreating = ref(false);

const failedList = computed(() => Object.values(failedContent.value).sort((a, b) => b.at - a.at));

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
            (inst.group || '').toLowerCase().includes(q) ||
            (inst.tags ?? []).some((t) => t.toLowerCase().includes(q))
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

function provProgress(p: { sessionId: string }): { text: string; pct: number } {
    const s = contentDls.value[p.sessionId];
    if (!s) return { text: 'Preparando…', pct: 0 };
    const total = s.total > 0 ? s.total : 0;
    const pct = total > 0 ? Math.min(100, Math.round((s.progress / total) * 100)) : 0;
    const text = s.message || 'Creando la instancia…';
    return { text, pct };
}

async function cancelProv(sessionId: string) {
    if (!sessionId) return;
    await cancelContentDownload(sessionId);
    await loadProvisioning();
}

async function loadAllDetails() {
    for (const inst of instances.value) void loadDetails(inst.name);
}

// Carga perezosa: List.vue ya no dispara loadInstances al montarse.
// La carga la orquesta Instances.vue cuando heavyPanel === 'instances'.
// Aquí solo se escuchan eventos de UI y se cargan detalles cuando la lista existe.
onMounted(() => {
    window.addEventListener('click', closeMenuOnClick);
    // Si el padre ya cargó la lista antes de montar este hijo (reapertura),
    // cargar los detalles de las tarjetas inmediatamente.
    if (instances.value.length) void loadAllDetails();
    void loadProvisioning();
});

// Cuando la lista de instancias cambie (primera carga perezosa), hidratar detalles
watch(() => instances.value.length, (len) => {
    if (len > 0) void loadAllDetails();
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
                <span class="offline-wrap" style="position: relative; display: inline-flex;">
                    <button
                        class="SsBtn SsBtnPrimary InstView_NewBtn"
                        :class="{ offline: isOffline }"
                        :disabled="isOffline"
                        :title="isOffline ? 'Sin conexión — Crear instancia requiere internet y no está disponible sin conexión.' : undefined"
                        @click="isOffline ? undefined : emit('new')"
                    >
                        <IconPlus stroke="2" /> Nueva instancia
                    </button>
                    <OfflineBadge v-if="isOffline" placement="inside" tooltip="left" message="Sin conexión — Crear instancia requiere internet y no está disponible sin conexión." />
                </span>
            </div>
        </header>

        <div class="InstView_Filters" v-if="totalCount || provisioning.length">
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
                <button
                    v-if="provisioning.length"
                    class="InstView_FilterChip"
                    :class="{ on: !hideCreating }"
                    :title="hideCreating ? 'Mostrar las que se están creando' : 'Ocultar las que se están creando (siguen en curso)'"
                    @click="hideCreating = !hideCreating"
                >
                    <component :is="hideCreating ? IconEyeOff : IconEye" stroke="2" /> Creando <em>{{ provisioning.length }}</em>
                </button>
            </div>
            <select v-if="groups.length" class="SsSel InstView_GroupSel" v-model="groupFilter">
                <option value="">Todos los grupos</option>
                <option v-for="g in groups" :key="g" :value="g">{{ g }}</option>
            </select>
        </div>

        <p v-if="loadingList && !sortedInstances.length && !provisioning.length && !failedList.length" class="InstView_Empty">Cargando instancias…</p>

        <div v-else-if="!sortedInstances.length && !provisioning.length && !failedList.length" class="InstView_EmptyCard">
            <span class="InstView_EmptyIcon"><IconBox stroke="1.4" /></span>
            <b class="InstView_EmptyTitle">Aún no tienes instancias</b>
            <p class="InstView_EmptyText">
                Las instancias son mundos de juego independientes con sus propias versiones, modloaders,
                configuraciones y capturas. Los recursos compartidos del launcher se reutilizan automáticamente.
            </p>
            <span class="offline-wrap" style="position: relative; display: inline-flex;">
                <button
                    class="SsBtn SsBtnPrimary InstView_EmptyBtn"
                    :class="{ offline: isOffline }"
                    :disabled="isOffline"
                    :title="isOffline ? 'Sin conexión — Crear instancia requiere internet y no está disponible sin conexión.' : undefined"
                    @click="isOffline ? undefined : emit('new')"
                >
                    <IconPlus stroke="2" /> Crear mi primera instancia
                </button>
                <OfflineBadge v-if="isOffline" placement="inside" tooltip="bottom" message="Sin conexión — Crear instancia requiere internet y no está disponible sin conexión." />
            </span>
        </div>

        <p v-else-if="!filtered.length && !provisioning.length && !failedList.length" class="InstView_Empty">
            No hay instancias que coincidan con «{{ search }}».
        </p>

        <div v-else class="InstView_Grid">
            <!-- En creación por modpack: parte de la rejilla, bloqueada (sin
                 abrir, fijar, editar ni jugar hasta estar lista). Compacta y
                 sin banner. Se puede ocultar con el filtro sin cancelarla. -->
            <template v-if="!hideCreating">
                <article
                    v-for="p in provisioning"
                    :key="'prov-' + p.name"
                    class="InstCard InstCardProv InstCardMini"
                    :title="`${p.title} se está creando y aún no está lista: no se puede abrir ni modificar`"
                >
                    <div class="InstCard_Body">
                        <div class="InstCard_Line">
                            <span class="InstCard_Icon">
                                <img v-if="p.iconUrl" :src="p.iconUrl" alt="" />
                                <IconLoader2 v-else class="spin" stroke="2" />
                            </span>
                            <div class="InstCard_Titles">
                                <span class="InstCard_Title">{{ p.title }}</span>
                                <span class="InstCard_Sub">
                                    <span class="InstCardProv_Lock">
                                        <IconLoader2 class="spin" stroke="2" /> Creando…
                                    </span>
                                </span>
                            </div>
                        </div>
                        <p class="InstCardProv_Msg" :title="provProgress(p).text">{{ provProgress(p).text }}</p>
                        <div class="InstCardProv_Bar">
                            <span :style="{ width: provProgress(p).pct + '%' }" />
                        </div>
                        <div class="InstCardProv_Actions">
                            <button class="SsBtn" title="Cancelar la creación" @click="cancelProv(p.sessionId)">
                                <IconPlayerStop stroke="2" /> Cancelar
                            </button>
                        </div>
                    </div>
                </article>
            </template>
            <!-- Fallos de instalación: siempre visibles hasta descartarlos,
                 para que ningún error pase en silencio. -->
            <article
                v-for="f in failedList"
                :key="'fail-' + f.sessionId"
                class="InstCard InstCardFail"
                :title="`Falló la instalación de ${f.label}`"
            >
                <div class="InstCard_Body">
                    <div class="InstCard_Line">
                        <span class="InstCard_Icon">
                            <img v-if="f.iconUrl" :src="f.iconUrl" alt="" />
                            <IconAlertTriangle v-else stroke="2" />
                        </span>
                        <div class="InstCard_Titles">
                            <span class="InstCard_Title">{{ f.label }}</span>
                            <span class="InstCard_Sub">
                                <span class="InstCardFail_Lock">No se pudo instalar</span>
                            </span>
                        </div>
                    </div>
                    <p class="InstCardProv_Msg InstCardFail_Msg" :title="f.error">{{ f.error }}</p>
                    <div class="InstCardProv_Actions">
                        <button class="SsBtn" title="Descartar este aviso" @click="dismissFailedContent(f.sessionId)">
                            <IconCheck stroke="2" /> Entendido
                        </button>
                    </div>
                </div>
            </article>
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
                        <span v-if="isInstanceVerifying(inst.name)" class="InstCard_Chip InstCard_VerifyChip" title="Verificando integridad… la instancia no se puede utilizar">
                            Verificando…
                        </span>
                    </div>

                    <div v-if="inst.tags?.length" class="InstCard_Tags">
                        <span v-for="t in inst.tags.slice(0, 4)" :key="t" class="InstCard_Tag">{{ t }}</span>
                        <span v-if="inst.tags.length > 4" class="InstCard_TagMore">+{{ inst.tags.length - 4 }}</span>
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
                        <span class="offline-wrap" style="position: relative; display: inline-flex;">
                            <button
                                class="InstCard_DlBtn"
                                :class="{ offline: isOffline }"
                                :title="isOffline ? 'Sin conexión — Descargar versión requiere internet y no está disponible sin conexión.' : 'Descargar versión'"
                                :disabled="isInstanceBusy(inst.name) || isOffline"
                                @click.stop="isOffline ? undefined : emit('download', inst.name)"
                            >
                                <IconDownload stroke="2" />
                            </button>
                            <OfflineBadge v-if="isOffline" placement="inside" tooltip="top" message="Sin conexión — Descargar versión requiere internet y no está disponible sin conexión." />
                        </span>
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