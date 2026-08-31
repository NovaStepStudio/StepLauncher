<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import {
    IconX, IconDownload, IconPuzzle, IconLoader2, IconRotateClockwise, IconFile,
} from '@tabler/icons-vue';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';
import {
    fetchProjectVersions, formatBytes, formatDate, formatCount, useModrinth,
} from '@/Common/Composables/useModrinth';
import type { ModrinthVersion } from '@/Common/Composables/useModrinth';
import { VERSION_TYPES } from './Store';

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

watch(
    () => props.slugOrId,
    () => {
        if (!props.slugOrId) return;
        versions.value = [];
        mcVersion.value = '';
        loader.value = '';
        channel.value = 'any';
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

function onCloseOverlays() {
    close();
}

useOverlayEscape(close, { priority: 2, isActive: () => props.visible });

onMounted(() => {
    void modrinthTags.loadTags();
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
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

                <div v-else-if="error" class="ModsDl_State error">
                    <b>No se pudieron cargar las versiones</b>
                    <span>{{ error }}</span>
                    <button class="SsBtn SsBtnPrimary" @click="retry">
                        <IconRotateClockwise stroke="2" /> Reintentar
                    </button>
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
                </div>

                <p v-else class="ModsDl_State">Este proyecto no tiene versiones disponibles.</p>
            </div>

            <footer class="ModsDl_Footer">
                <button class="SsBtn" @click="close">Cancelar</button>
                <button
                    class="SsBtn SsBtnPrimary ModsDl_Confirm"
                    :disabled="!matchedVersion || loading"
                    title="La instalación de mods se habilitará con el sistema de descargas"
                >
                    <IconDownload stroke="2" /> Descargar
                </button>
            </footer>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use './Styles/DownloadDialog.scss';
</style>
