<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue';
import {
    IconX, IconBox,
} from '@tabler/icons-vue';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import {
    loadInstances, loadDetails,
    deleteInstance, cloneInstance,
} from './Store';
import InstancesView from './List.vue';
import InstanceDetailView from './Detail.vue';
import {
    heavyPanel, openHeavyPanel, closeHeavyPanel,
    openDialog, ask, shotsInstance, shotsReturn,
} from '@/Common/Overlays/Store';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

const view = ref<'list' | 'detail'>('list');
const selected = ref('');

const deleting = ref(false);
const cloning = ref(false);

function openDetail(name: string) {
    selected.value = name;
    view.value = 'detail';
    void loadDetails(name);
}

function toGrid() {
    view.value = 'list';
    selected.value = '';
    void loadInstances();
}

function newInstance() {
    openDialog('instances-form', { editing: null }, { done: onFormDone, created: onCreated });
}

function openEdit(name: string) {
    openDialog('instances-form', { editing: name }, { done: onFormDone, created: onCreated });
}

function openSettings(name: string) {
    openDialog('instances-settings', { name }, { done: onSettingsDone });
}

function openShots(name: string) {
    openHeavyPanel('shots');
    shotsInstance.value = name;
    shotsReturn.value = true;
}

function openDownload(name: string) {
    openDialog('instances-download', { name }, { done: onDlDone });
}

function onDlDone() {
    if (selected.value) void loadDetails(selected.value);
}

async function requestClone(name: string) {
    const r = await ask({
        title: `Clonar "${name}"`,
        message: 'Se copiarán las versiones descargadas de esta instancia (sin re-descargarlas).',
        confirmLabel: 'Clonar',
        inputLabel: 'Nombre de la copia',
        inputValue: `${name}-copia`,
    });
    if (!r.confirmed) return;
    const target = r.value.trim();
    if (!target || cloning.value) return;
    cloning.value = true;
    const err = await cloneInstance(name, target);
    cloning.value = false;
    if (err) {
        await ask({ title: 'No se pudo clonar', message: err, confirmLabel: 'Aceptar' });
        return;
    }
    void loadInstances();
    openDetail(target);
}

async function requestDelete(name: string) {
    const r = await ask({
        title: '¿Eliminar la instancia?',
        message: `Se borrarán sus versiones descargadas, modloaders, capturas y toda la carpeta de "${name}". Esta acción no se puede deshacer.`,
        confirmLabel: 'Eliminar',
        danger: true,
    });
    if (!r.confirmed || deleting.value) return;
    deleting.value = true;
    const err = await deleteInstance(name);
    deleting.value = false;
    if (err) {
        await ask({ title: 'No se pudo eliminar', message: err, confirmLabel: 'Aceptar' });
        return;
    }
    if (view.value === 'detail' && selected.value === name) toGrid();
    void loadInstances();
}

function onFormDone() {
    void loadInstances();
    if (view.value === 'detail') void loadDetails(selected.value);
}

function onCreated(name: string) {
    void loadInstances();
    openDetail(name);
    openDownload(name);
}

function onSettingsDone() {
    void loadDetails(selected.value);
}

function close() {
    closeHeavyPanel('instances');
}

function onCloseOverlays() {
    close();
}

useOverlayEscape(close, { isActive: () => heavyPanel.value === 'instances' });

watch(heavyPanel, (p) => {
    if (p !== 'instances') return;
    if (shotsReturn.value) {
        shotsReturn.value = false;
        if (selected.value) void loadDetails(selected.value);
        return;
    }
    view.value = 'list';
    selected.value = '';
    void loadInstances();
});

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});
</script>

<template>
    <div class="InstancesModal_Overlay">
        <header v-if="view === 'list'" class="InstancesModal_Head">
            <div class="InstancesModal_Title">
                <span class="InstancesModal_Icon"><IconBox stroke="2" /></span>
                <div class="InstancesModal_Titles">
                    <h3>Instancias</h3>
                    <p>Tu biblioteca de mundos y versiones</p>
                </div>
            </div>
            <button class="InstancesModal_Close" title="Cerrar" @click="close">
                <IconX stroke="2" />
            </button>
        </header>

        <div class="InstancesModal_Body">
            <InstancesView
                v-if="view === 'list'"
                @open="openDetail"
                @new="newInstance"
                @download="openDownload"
                @edit="openEdit"
                @settings="openSettings"
                @clone="requestClone"
                @shots="openShots"
                @delete="requestDelete"
            />
            <InstanceDetailView
                v-else
                :name="selected"
                @close="toGrid"
                @exit="close"
                @edit="openEdit"
                @settings="openSettings"
                @shots="openShots"
                @download="openDownload"
                @delete="requestDelete"
            />
        </div>
    </div>
</template>

<style scoped lang="scss">
@use './Styles/Instances.scss';
</style>