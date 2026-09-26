<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import {
    IconBox, IconPuzzle, IconPalette, IconWallpaper, IconReload, IconTrash,
    IconPlayerPause, IconPlayerPlay, IconDownload, IconLoader2, IconSearch, IconPlus,
    IconFolderOpen, IconArrowsLeftRight,
} from '@tabler/icons-vue';
import { Events } from '@wailsio/runtime';
import { formatBytes } from '@/Common/Composables/useModrinth';
import { loadLocal } from '@/Common/Stores/Ui';
import { openHeavyPanel } from '@/Common/Overlays/Store';
import { ask } from '@/Common/Overlays/Store';
import {
    ListInstalledContent, SetInstalledContentEnabled, DeleteInstalledContent, GetContentIcon,
    MoveInstalledContent, RevealInstalledContent,
} from '@wailsjs/StepLauncher/internal/Services/Mods/modsservice';
import { ListInstances } from '@wailsjs/StepLauncher/internal/Services/Instance/instanceservice';
import { preferInstance } from '@/Mods/Store';
import MoveTo from '@/Mods/MoveTo.vue';
import { ensureContentEvents } from './Store';

const props = defineProps<{
    name: string;
}>();

export interface InstanceModFile {
    kind: string;
    name: string;
    title: string;
    size: number;
    enabled: boolean;
}

const KIND_LABEL: Record<string, string> = {
    mod: 'Mod',
    shader: 'Shader',
    resourcepack: 'Textura',
};

const KIND_ICON: Record<string, unknown> = {
    mod: IconPuzzle,
    shader: IconPalette,
    resourcepack: IconWallpaper,
};

const kindFilter = ref<'all' | 'mod' | 'shader' | 'resourcepack'>('all');
const query = ref('');
const files = ref<InstanceModFile[]>([]);
const loading = ref(false);
const error = ref('');
const acting = ref('');
const iconUrls = ref<Record<string, string>>({});

const filtered = computed(() => {
    const q = query.value.trim().toLowerCase();
    return files.value.filter((f) => {
        if (kindFilter.value !== 'all' && f.kind !== kindFilter.value) return false;
        if (q && !f.name.toLowerCase().includes(q)) return false;
        return true;
    });
});

const counts = computed(() => {
    const c: Record<string, number> = { all: files.value.length, mod: 0, shader: 0, resourcepack: 0 };
    for (const f of files.value) c[f.kind] = (c[f.kind] ?? 0) + 1;
    return c;
});

async function refresh(): Promise<void> {
    if (loading.value) return;
    loading.value = true;
    error.value = '';
    try {
        const list = await ListInstalledContent('instance', props.name);
        files.value = (Array.isArray(list) ? list : []).filter((f) => f !== null) as InstanceModFile[];
        void refreshIcons();
    } catch (e: any) {
        error.value = e?.message ?? 'No se pudo leer el contenido.';
        files.value = [];
    } finally {
        loading.value = false;
    }
}

async function refreshIcons(): Promise<void> {
    const next: Record<string, string> = {};
    await Promise.allSettled(
        files.value.map(async (f) => {
            const key = `${f.kind}:${f.name}`;
            try {
                const rel = await GetContentIcon('instance', props.name, f.kind, f.name);
                if (rel) next[key] = await loadLocal(rel);
            } catch (_e) {}
        })
    );
    iconUrls.value = next;
}

function iconOf(f: InstanceModFile): string {
    return iconUrls.value[`${f.kind}:${f.name}`] ?? '';
}

async function toggle(f: InstanceModFile): Promise<void> {
    if (acting.value) return;
    acting.value = `${f.kind}:${f.name}`;
    try {
        await SetInstalledContentEnabled('instance', props.name, f.kind, f.name, !f.enabled);
        await refresh();
    } catch (e: any) {
        error.value = e?.message ?? 'No se pudo cambiar el estado.';
    } finally {
        acting.value = '';
    }
}

async function remove(f: InstanceModFile): Promise<void> {
    const r = await ask({
        title: `Borrar ${f.name}`,
        message: `Se eliminará de ${props.name}. Esta acción no se puede deshacer.`,
        confirmLabel: 'Borrar',
        cancelLabel: 'Cancelar',
        danger: true,
    });
    if (!r.confirmed) return;
    acting.value = `${f.kind}:${f.name}`;
    try {
        await DeleteInstalledContent('instance', props.name, f.kind, f.name);
        await refresh();
    } catch (e: any) {
        error.value = e?.message ?? 'No se pudo borrar el archivo.';
    } finally {
        acting.value = '';
    }
}

function addMore(): void {
    preferInstance(props.name);
    openHeavyPanel('mods');
}

// Mover a otro destino (global u otra instancia).
const movingKey = ref('');
const moveInstances = ref<Array<{ name: string; title: string }>>([]);

async function loadMoveInstances(): Promise<void> {
    try {
        const list = await ListInstances();
        moveInstances.value = (Array.isArray(list) ? list : [])
            .filter((i) => i !== null)
            .map((i) => ({ name: String((i as any).name ?? ''), title: String((i as any).title ?? (i as any).name ?? '') }))
            .filter((i) => i.name);
    } catch {
        moveInstances.value = [];
    }
}

function startMove(f: InstanceModFile): void {
    const key = `${f.kind}:${f.name}`;
    movingKey.value = movingKey.value === key ? '' : key;
    if (movingKey.value) void loadMoveInstances();
}

async function confirmMove(f: InstanceModFile, dest: string, inst: string): Promise<void> {
    movingKey.value = '';
    acting.value = `${f.kind}:${f.name}`;
    try {
        await MoveInstalledContent('instance', props.name, f.kind, f.name, dest, inst);
        await refresh();
    } catch (e: any) {
        error.value = e?.message ?? 'No se pudo mover el archivo.';
    } finally {
        acting.value = '';
    }
}

async function reveal(f: InstanceModFile): Promise<void> {
    try {
        await RevealInstalledContent('instance', props.name, f.kind, f.name);
    } catch (e: any) {
        error.value = e?.message ?? 'No se pudo abrir la ubicación.';
    }
}

function displayName(f: InstanceModFile): string {
    return f.title && f.title !== f.name ? f.title : f.name;
}

function onInstalledBackend(): void {
    void refresh();
}

let eventOffs: Array<() => void> = [];

watch(
    () => props.name,
    () => void refresh()
);

onMounted(() => {
    ensureContentEvents();
    void refresh();
    eventOffs = [
        Events.On('modcontent_installed', onInstalledBackend),
        Events.On('modpack_installed', onInstalledBackend),
    ];
});

onUnmounted(() => {
    eventOffs.forEach((off) => off());
    eventOffs = [];
});
</script>

<template>
    <div class="InstMods">
        <div class="InstMods_Head">
            <span>Contenido de {{ name }}</span>
            <div class="InstMods_HeadActions">
                <label class="InstMods_Search">
                    <IconSearch stroke="2" />
                    <input v-model="query" type="text" placeholder="Buscar…" autocomplete="off" spellcheck="false" />
                </label>
                <button class="SsBtn" title="Recargar" :disabled="loading" @click="refresh">
                    <IconReload stroke="2" :class="{ spin: loading }" />
                </button>
                <button class="SsBtn SsBtnPrimary" title="Añadir desde Modrinth" @click="addMore">
                    <IconPlus stroke="2" /> Añadir
                </button>
            </div>
        </div>

        <div class="InstMods_Tabs">
            <button :class="{ on: kindFilter === 'all' }" @click="kindFilter = 'all'">
                Todo <em>{{ counts.all ?? 0 }}</em>
            </button>
            <button :class="{ on: kindFilter === 'mod' }" @click="kindFilter = 'mod'">
                Mods <em>{{ counts.mod ?? 0 }}</em>
            </button>
            <button :class="{ on: kindFilter === 'shader' }" @click="kindFilter = 'shader'">
                Shaders <em>{{ counts.shader ?? 0 }}</em>
            </button>
            <button :class="{ on: kindFilter === 'resourcepack' }" @click="kindFilter = 'resourcepack'">
                Texturas <em>{{ counts.resourcepack ?? 0 }}</em>
            </button>
        </div>

        <div v-if="loading && !files.length" class="InstMods_State">
            <IconLoader2 class="spin" stroke="2" /> Leyendo el contenido…
        </div>
        <div v-else-if="error && !files.length" class="InstMods_State">
            <b>No se pudo leer el contenido</b>
            <span>{{ error }}</span>
            <button class="SsBtn SsBtnPrimary" @click="refresh">Reintentar</button>
        </div>
        <div v-else-if="!filtered.length" class="InstMods_State">
            <IconBox stroke="1.4" />
            <p>Sin contenido todavía. Pulsa Añadir para descargar desde Modrinth.</p>
        </div>

        <div v-else class="InstMods_List">
            <template v-for="f in filtered" :key="`${f.kind}:${f.name}`">
                <article
                    class="InstMods_Row"
                    :class="{ off: !f.enabled }"
                >
                    <span class="InstMods_Icon">
                        <img v-if="iconOf(f)" :src="iconOf(f)" :alt="f.name" loading="lazy" decoding="async" />
                        <component v-else :is="KIND_ICON[f.kind] ?? IconBox" stroke="1.6" />
                    </span>
                    <div class="InstMods_Info">
                        <span class="InstMods_Name" :title="f.name">{{ displayName(f) }}</span>
                        <span class="InstMods_Sub">
                            {{ KIND_LABEL[f.kind] ?? f.kind }} · {{ formatBytes(f.size) }}
                            <template v-if="f.title && f.title !== f.name"> · {{ f.name }}</template>
                            <template v-if="!f.enabled"> · Deshabilitado</template>
                        </span>
                    </div>
                    <div class="InstMods_Actions">
                        <button
                            class="SsBtn"
                            :disabled="!!acting"
                            :title="f.enabled ? 'Deshabilitar' : 'Habilitar'"
                            @click="toggle(f)"
                        >
                            <component :is="f.enabled ? IconPlayerPause : IconPlayerPlay" stroke="2" />
                        </button>
                        <button
                            class="SsBtn"
                            :disabled="!!acting"
                            title="Mover a otro destino"
                            @click="startMove(f)"
                        >
                            <IconArrowsLeftRight stroke="2" />
                        </button>
                        <button
                            class="SsBtn"
                            :disabled="!!acting"
                            title="Abrir ubicación del archivo"
                            @click="reveal(f)"
                        >
                            <IconFolderOpen stroke="2" />
                        </button>
                        <button
                            class="SsBtn InstMods_Danger"
                            :disabled="!!acting"
                            title="Borrar"
                            @click="remove(f)"
                        >
                            <IconTrash stroke="2" />
                        </button>
                    </div>
                </article>
                <MoveTo
                    v-if="movingKey === `${f.kind}:${f.name}`"
                    :file-name="f.name"
                    :from-label="`instancia ${props.name}`"
                    from-destination="instance"
                    :from-instance="props.name"
                    :instances="moveInstances"
                    @confirm="(dest, inst) => confirmMove(f, dest, inst)"
                    @cancel="movingKey = ''"
                />
            </template>
        </div>
    </div>
</template>

<style scoped lang="scss">
@use './Styles/InstanceMods.scss';
</style>
