<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue';
import { IconX, IconPuzzle } from '@tabler/icons-vue';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { heavyPanel, closeHeavyPanel, openDialog, dialogs } from '@/Common/Overlays/Store';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';
import { useModrinth } from '@/Common/Composables/useModrinth';
import type { SearchHit } from '@/Common/Composables/useModrinth';
import ModsContent from './Content.vue';
import ModsDetail from './Detail.vue';
import ModsInstalled from './Installed.vue';

const modrinth = useModrinth();

const view = ref<'list' | 'detail' | 'installed'>('list');
const selected = ref('');
const selectedHit = ref<SearchHit | null>(null);

function openDetail(slugOrId: string, hit?: SearchHit) {
    selected.value = slugOrId;
    selectedHit.value = hit ?? null;
    view.value = 'detail';
    void modrinth.loadProject(slugOrId);
}

function toGrid() {
    view.value = 'list';
    selected.value = '';
    selectedHit.value = null;
    modrinth.clearProject();
}

function showInstalled() {
    selected.value = '';
    selectedHit.value = null;
    modrinth.clearProject();
    modrinth.abortSearch();
    view.value = 'installed';
}

interface DownloadTarget {
    slug: string | null;
    title: string;
    icon_url: string | null;
    project_type: string;
    project_id?: string | null;
}

function openDownload(target: DownloadTarget) {
    openDialog('mods-download', {
        slugOrId: target.slug ?? target.project_id ?? target.title,
        title: target.title,
        iconUrl: target.icon_url,
        projectType: target.project_type,
    });
}

function closeModsDownloadDialogs(): void {
    for (let i = dialogs.value.length - 1; i >= 0; i--) {
        const d = dialogs.value[i];
        if (d?.name === 'mods-download') dialogs.value.splice(i, 1);
    }
}

function close() {
    closeModsDownloadDialogs();
    closeHeavyPanel('mods');
}

function onCloseOverlays() {
    close();
}

useOverlayEscape(close, { isActive: () => heavyPanel.value === 'mods' });

watch(heavyPanel, (p) => {
    if (p !== 'mods') {
        closeModsDownloadDialogs();
        return;
    }
    if (view.value !== 'installed') {
        toGrid();
        modrinth.abortSearch();
    }
});

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    modrinth.abortSearch();
});
</script>

<template>
    <div class="ModsModal_Overlay">
        <header v-if="view !== 'detail'" class="ModsModal_Head">
            <div class="ModsModal_Title">
                <span class="ModsModal_Icon"><IconPuzzle stroke="2" /></span>
                <div class="ModsModal_Titles">
                    <h3>Mods</h3>
                    <p>Descarga mods, modpacks, shaders y texturas desde Modrinth</p>
                </div>
            </div>
            <button class="ModsModal_Close" title="Cerrar" @click="close">
                <IconX stroke="2" />
            </button>
        </header>

        <div class="ModsModal_Body">
            <ModsContent v-if="view === 'list'" @open="openDetail" @download="openDownload" @installed="showInstalled" />
            <ModsInstalled v-else-if="view === 'installed'" @back="toGrid" />
            <ModsDetail
                v-else
                :slug-or-id="selected"
                :author="selectedHit?.author ?? ''"
                @close="toGrid"
                @exit="close"
                @download="openDownload"
            />
        </div>
    </div>
</template>

<style scoped lang="scss">
@use './Styles/Mods.scss';
</style>