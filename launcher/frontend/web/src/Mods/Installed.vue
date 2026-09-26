<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import {
    IconArrowLeft, IconBox, IconPuzzle, IconPalette, IconWallpaper, IconReload, IconTrash,
    IconPlayerPause, IconPlayerPlay, IconX, IconDownload, IconLoader2, IconSearch,
    IconFolderOpen, IconArrowsLeftRight, IconChevronDown, IconAlertTriangle, IconCheck,
} from '@tabler/icons-vue';
import { Events } from '@wailsio/runtime';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { ask } from '@/Common/Overlays/Store';
import { heavyPanel } from '@/Common/Overlays/Store';
import { formatBytes } from '@/Common/Composables/useModrinth';
import { loadLocal } from '@/Common/Stores/Ui';
import { ListInstances } from '@wailsjs/StepLauncher/internal/Services/Instance/instanceservice';
import {
    ListInstalledContent, SetInstalledContentEnabled, DeleteInstalledContent, GetContentIcon,
    MoveInstalledContent, RevealInstalledContent,
} from '@wailsjs/StepLauncher/internal/Services/Mods/modsservice';
import {
    contentDls, contentMeta, failedContent, dismissFailedContent, cancelContentDownload, ensureContentEvents,
} from '@/Instances/Store';
import MoveTo from './MoveTo.vue';

const emit = defineEmits<{
    (e: 'back'): void;
}>();

export interface InstalledFile {
    kind: string;
    name: string;
    title: string;
    size: number;
    enabled: boolean;
}

const KIND_TABS = [
    { id: 'all', label: 'Todo' },
    { id: 'mod', label: 'Mods' },
    { id: 'shader', label: 'Shaders' },
    { id: 'resourcepack', label: 'Texturas' },
] as const;

type KindFilter = (typeof KIND_TABS)[number]['id'];

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

const destination = ref<'global' | string>('global');
const instances = ref<Array<{ name: string; title: string }>>([]);
const kindFilter = ref<KindFilter>('all');
const statusFilter = ref<'all' | 'on' | 'off'>('all');
const query = ref('');
const files = ref<InstalledFile[]>([]);
const loading = ref(false);
const error = ref('');
const acting = ref('');

const destLabel = computed(() => {
    if (destination.value === 'global') return 'Juego global';
    const found = instances.value.find((i) => i.name === destination.value);
    return found?.title || destination.value;
});

// Sesiones en curso (las muestra el widget): con botón para cancelarlas.
const activeSessions = computed(() =>
    Object.values(contentDls.value).filter((s) => !s.done && !s.errorMsg)
);

// Fallos persistentes: se quedan hasta descartarlos para que ningún error
// pase en silencio aunque el diálogo estuviera cerrado.
const failedSessions = computed(() =>
    Object.values(failedContent.value).sort((a, b) => b.at - a.at)
);

const filtered = computed(() => {
    const q = query.value.trim().toLowerCase();
    return files.value.filter((f) => {
        if (kindFilter.value !== 'all' && f.kind !== kindFilter.value) return false;
        if (statusFilter.value === 'on' && !f.enabled) return false;
        if (statusFilter.value === 'off' && f.enabled) return false;
        if (q && !f.name.toLowerCase().includes(q)) return false;
        return true;
    });
});

// Iconos reales extraídos del contenido (pack.png o icon del mod). Se
// resuelven por backend y se sirven con loadLocal; lo que no tenga icono
// conserva el genérico por tipo.
const iconUrls = ref<Record<string, string>>({});

async function refreshIcons(): Promise<void> {
    const next: Record<string, string> = {};
    const dest = destination.value === 'global' ? 'global' : 'instance';
    const inst = destination.value === 'global' ? '' : destination.value;
    await Promise.allSettled(
        files.value.map(async (f) => {
            const key = `${f.kind}:${f.name}`;
            try {
                const rel = await GetContentIcon(dest, inst, f.kind, f.name);
                if (rel) next[key] = await loadLocal(rel);
            } catch (_e) {}
        })
    );
    iconUrls.value = next;
}

function iconOf(f: InstalledFile): string {
    return iconUrls.value[`${f.kind}:${f.name}`] ?? '';
}

const counts = computed(() => {
    const c: Record<string, number> = { all: files.value.length, mod: 0, shader: 0, resourcepack: 0 };
    for (const f of files.value) c[f.kind] = (c[f.kind] ?? 0) + 1;
    return c;
});

async function loadInstanceOptions(): Promise<void> {
    try {
        const list = await ListInstances();
        instances.value = (Array.isArray(list) ? list : [])
            .filter((i) => i !== null)
            .map((i) => ({ name: String((i as any).name ?? ''), title: String((i as any).title ?? (i as any).name ?? '') }))
            .filter((i) => i.name);
    } catch {
        instances.value = [];
    }
}

async function refresh(): Promise<void> {
    if (loading.value) return;
    loading.value = true;
    error.value = '';
    try {
        const dest = destination.value === 'global' ? 'global' : 'instance';
        const inst = destination.value === 'global' ? '' : destination.value;
        const list = await ListInstalledContent(dest, inst);
        files.value = (Array.isArray(list) ? list : []).filter((f) => f !== null) as InstalledFile[];
        void refreshIcons();
    } catch (e: any) {
        error.value = e?.message ?? 'No se pudo leer el contenido instalado.';
        files.value = [];
    } finally {
        loading.value = false;
    }
}

async function toggle(f: InstalledFile): Promise<void> {
    const key = `${f.kind}:${f.name}`;
    if (acting.value) return;
    acting.value = key;
    try {
        const dest = destination.value === 'global' ? 'global' : 'instance';
        const inst = destination.value === 'global' ? '' : destination.value;
        await SetInstalledContentEnabled(dest, inst, f.kind, f.name, !f.enabled);
        await refresh();
    } catch (e: any) {
        error.value = e?.message ?? 'No se pudo cambiar el estado.';
    } finally {
        acting.value = '';
    }
}

async function remove(f: InstalledFile): Promise<void> {
    const r = await ask({
        title: `Borrar ${f.name}`,
        message: `Se eliminará de ${destLabel.value}. Esta acción no se puede deshacer.`,
        confirmLabel: 'Borrar',
        cancelLabel: 'Cancelar',
        danger: true,
    });
    if (!r.confirmed) return;
    const key = `${f.kind}:${f.name}`;
    acting.value = key;
    try {
        const dest = destination.value === 'global' ? 'global' : 'instance';
        const inst = destination.value === 'global' ? '' : destination.value;
        await DeleteInstalledContent(dest, inst, f.kind, f.name);
        await refresh();
    } catch (e: any) {
        error.value = e?.message ?? 'No se pudo borrar el archivo.';
    } finally {
        acting.value = '';
    }
}

async function cancelSession(sessionId: string): Promise<void> {
    await cancelContentDownload(sessionId);
}

function sessionIcon(s: { sessionId: string }): string {
    return contentMeta.value[s.sessionId]?.iconUrl ?? '';
}

function sessionDest(s: { sessionId: string }): string {
    const dest = contentMeta.value[s.sessionId]?.dest ?? '';
    return dest || 'destino';
}

// Mover entre destinos: abre el selector inline bajo la fila.
const movingKey = ref('');
const expandedSession = ref('');

function startMove(f: InstalledFile): void {
    movingKey.value = movingKey.value === `${f.kind}:${f.name}` ? '' : `${f.kind}:${f.name}`;
}

async function confirmMove(f: InstalledFile, dest: string, inst: string): Promise<void> {
    const fromDest = destination.value === 'global' ? 'global' : 'instance';
    const fromInst = destination.value === 'global' ? '' : destination.value;
    movingKey.value = '';
    acting.value = `${f.kind}:${f.name}`;
    try {
        await MoveInstalledContent(fromDest, fromInst, f.kind, f.name, dest, inst);
        await refresh();
    } catch (e: any) {
        error.value = e?.message ?? 'No se pudo mover el archivo.';
    } finally {
        acting.value = '';
    }
}

async function reveal(f: InstalledFile): Promise<void> {
    try {
        const dest = destination.value === 'global' ? 'global' : 'instance';
        const inst = destination.value === 'global' ? '' : destination.value;
        await RevealInstalledContent(dest, inst, f.kind, f.name);
    } catch (e: any) {
        error.value = e?.message ?? 'No se pudo abrir la ubicación.';
    }
}

function displayName(f: InstalledFile): string {
    return f.title && f.title !== f.name ? f.title : f.name;
}

function onInstalledBackend(): void {
    // Tras instalar algo desde el diálogo, la lista se actualiza sola.
    void refresh();
}

let eventOffs: Array<() => void> = [];

watch(destination, () => void refresh());

watch(
    () => heavyPanel.value,
    (v) => {
        if (v === 'mods') void refresh();
    }
);

function onCloseOverlays(): void {}

onMounted(() => {
    ensureContentEvents();
    void loadInstanceOptions().then(() => void refresh());
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    eventOffs = [
        Events.On('modcontent_installed', onInstalledBackend),
        Events.On('modpack_installed', onInstalledBackend),
    ];
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    eventOffs.forEach((off) => off());
    eventOffs = [];
});
</script>

<template>
    <div class="ModsInstalled">
        <header class="ModsInstalled_Bar">
            <div class="ModsInstalled_BarLeft">
                <button class="ModsInstalled_Back" title="Volver a explorar" @click="emit('back')">
                    <IconArrowLeft stroke="2" />
                </button>
                <div class="ModsInstalled_BarInfo">
                    <h2>Instalado</h2>
                    <p>Activa, desactiva o borra tu contenido en {{ destLabel }}</p>
                </div>
            </div>
            <div class="ModsInstalled_BarActions">
                <select v-model="destination" class="SsSel" title="Destino">
                    <option value="global">Juego global</option>
                    <option v-for="i in instances" :key="i.name" :value="i.name">
                        {{ i.title || i.name }}
                    </option>
                </select>
                <button class="ModsInstalled_Reload" title="Recargar" :disabled="loading" @click="refresh">
                    <IconReload stroke="2" :class="{ spin: loading }" />
                </button>
            </div>
        </header>

        <div class="ModsInstalled_Tools">
            <label class="ModsInstalled_Search">
                <IconSearch stroke="2" />
                <input v-model="query" type="text" placeholder="Buscar por nombre…" autocomplete="off" spellcheck="false" />
            </label>
            <select v-model="statusFilter" class="SsSel" title="Estado">
                <option value="all">Todos</option>
                <option value="on">Activados</option>
                <option value="off">Deshabilitados</option>
            </select>
        </div>

        <div v-if="failedSessions.length" class="ModsInstalled_Failed">
            <div v-for="s in failedSessions" :key="s.sessionId" class="ModsInstalled_Session ModsInstalled_SessionFail">
                <span class="ModsInstalled_SessionIcon">
                    <img v-if="s.iconUrl" :src="s.iconUrl" alt="" />
                    <IconAlertTriangle v-else stroke="2" />
                </span>
                <div class="ModsInstalled_SessionInfo">
                    <b>No se pudo instalar {{ s.label }}</b>
                    <span :title="s.error">{{ s.error }}</span>
                </div>
                <button class="SsBtn" title="Descartar este aviso" @click="dismissFailedContent(s.sessionId)">
                    <IconCheck stroke="2" /> Entendido
                </button>
            </div>
        </div>

        <div v-if="activeSessions.length" class="ModsInstalled_Active">
            <div v-for="s in activeSessions" :key="s.sessionId" class="ModsInstalled_Session">
                <span class="ModsInstalled_SessionIcon">
                    <img v-if="sessionIcon(s)" :src="sessionIcon(s)" alt="" />
                    <IconDownload v-else stroke="2" />
                </span>
                <div class="ModsInstalled_SessionInfo">
                    <b>{{ s.label }}</b>
                    <span>{{ s.message || 'Instalando…' }}</span>
                </div>
                <button
                    class="ModsInstalled_SessionToggle"
                    :title="expandedSession === s.sessionId ? 'Ocultar detalle' : 'Ver detalle'"
                    @click="expandedSession = expandedSession === s.sessionId ? '' : s.sessionId"
                >
                    <IconChevronDown stroke="2" :class="{ flip: expandedSession === s.sessionId }" />
                </button>
                <button class="SsBtn" title="Cancelar" @click="cancelSession(s.sessionId)">
                    <IconX stroke="2" /> Cancelar
                </button>
                <div v-if="expandedSession === s.sessionId" class="ModsInstalled_SessionDetail">
                    <span>Destino: {{ sessionDest(s) }}</span>
                    <div v-if="s.total > 0" class="ModsInstalled_SessionBar">
                        <div
                            class="ModsInstalled_SessionFill"
                            :style="{ width: Math.min(100, Math.max(0, (s.progress / Math.max(1, s.total)) * 100)) + '%' }"
                        ></div>
                    </div>
                    <span v-if="s.total > 0">{{ Math.min(s.progress, s.total) }}/{{ s.total }} archivos</span>
                </div>
            </div>
        </div>

        <div class="ModsInstalled_Tabs">
            <button
                v-for="t in KIND_TABS"
                :key="t.id"
                class="ModsInstalled_Tab"
                :class="{ on: kindFilter === t.id }"
                @click="kindFilter = t.id"
            >
                {{ t.label }} <em>{{ counts[t.id] ?? 0 }}</em>
            </button>
        </div>

        <div v-if="loading && !files.length" class="ModsInstalled_Empty">
            <IconLoader2 class="spin" stroke="2" />
            <p>Leyendo el contenido instalado…</p>
        </div>

        <div v-else-if="error && !files.length" class="ModsInstalled_Empty">
            <b>No se pudo leer el contenido</b>
            <p>{{ error }}</p>
            <button class="SsBtn SsBtnPrimary" @click="refresh">
                <IconReload stroke="2" /> Reintentar
            </button>
        </div>

        <div v-else-if="!filtered.length" class="ModsInstalled_Empty">
            <span class="ModsInstalled_EmptyIcon"><IconBox stroke="1.4" /></span>
            <b>Nada instalado aquí</b>
            <p>Descarga mods, shaders o texturas desde Explorar y aparecerán en esta lista.</p>
        </div>

        <div v-else class="ModsInstalled_List">
            <template v-for="f in filtered" :key="`${f.kind}:${f.name}`">
                <article
                    class="ModsInstalled_Row"
                    :class="{ off: !f.enabled }"
                >
                    <span class="ModsInstalled_Icon">
                        <img v-if="iconOf(f)" :src="iconOf(f)" :alt="f.name" loading="lazy" decoding="async" />
                        <component v-else :is="KIND_ICON[f.kind] ?? IconBox" stroke="1.6" />
                    </span>
                    <div class="ModsInstalled_Info">
                        <span class="ModsInstalled_Name" :title="f.name">{{ displayName(f) }}</span>
                        <span class="ModsInstalled_Sub" :title="f.title && f.title !== f.name ? f.name : undefined">
                            {{ KIND_LABEL[f.kind] ?? f.kind }} · {{ formatBytes(f.size) }}
                            <template v-if="f.title && f.title !== f.name"> · {{ f.name }}</template>
                            <template v-if="!f.enabled"> · Deshabilitado</template>
                        </span>
                    </div>
                    <div class="ModsInstalled_Actions">
                        <button
                            class="SsBtn ModsInstalled_IconBtn"
                            :disabled="!!acting"
                            :title="f.enabled ? 'Deshabilitar (no cargará en el juego)' : 'Habilitar'"
                            @click="toggle(f)"
                        >
                            <component :is="f.enabled ? IconPlayerPause : IconPlayerPlay" stroke="2" />
                        </button>
                        <button
                            class="SsBtn ModsInstalled_IconBtn"
                            :disabled="!!acting"
                            title="Mover a otro destino"
                            @click="startMove(f)"
                        >
                            <IconArrowsLeftRight stroke="2" />
                        </button>
                        <button
                            class="SsBtn ModsInstalled_IconBtn"
                            :disabled="!!acting"
                            title="Abrir ubicación del archivo"
                            @click="reveal(f)"
                        >
                            <IconFolderOpen stroke="2" />
                        </button>
                        <button
                            class="SsBtn ModsInstalled_IconBtn ModsInstalled_Danger"
                            :disabled="!!acting"
                            title="Borrar archivo"
                            @click="remove(f)"
                        >
                            <IconTrash stroke="2" />
                        </button>
                    </div>
                </article>
                <MoveTo
                    v-if="movingKey === `${f.kind}:${f.name}`"
                    :file-name="f.name"
                    :from-label="destLabel"
                    :from-destination="destination === 'global' ? 'global' : 'instance'"
                    :from-instance="destination === 'global' ? '' : destination"
                    :instances="instances"
                    @confirm="(dest, inst) => confirmMove(f, dest, inst)"
                    @cancel="movingKey = ''"
                />
            </template>
        </div>
    </div>
</template>

<style scoped lang="scss">
@use './Styles/Installed.scss';
</style>
