<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue';
import { dialogs, closeDialog, closeAllDialogs, confirmState, resolveConfirm } from './Store';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import Confirm from './Confirm.vue';
import AccountForm from '@/Accounts/Form.vue';
import ProfileForm from '@/Versions/ProfileForm.vue';
import FontManager from '@/Settings/FontManager.vue';
import InstanceForm from '@/Instances/Form.vue';
import InstanceSettings from '@/Instances/Settings.vue';
import InstanceDownload from '@/Instances/Download.vue';

const REGISTRY: Record<string, unknown> = {
    'account-form': AccountForm,
    'profile-form': ProfileForm,
    'font-manager': FontManager,
    'instances-form': InstanceForm,
    'instances-settings': InstanceSettings,
    'instances-download': InstanceDownload,
};

function onCloseOverlays() {
    closeAllDialogs();
    resolveConfirm({ confirmed: false, value: '' });
}

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});
</script>

<template>
    <Confirm v-if="confirmState" v-bind="confirmState.opts" />
    <component
        v-for="d in dialogs"
        :key="d.id"
        :is="REGISTRY[d.name]"
        :visible="true"
        v-bind="d.props"
        v-on="d.listeners"
        @update:visible="closeDialog(d.id)"
    />
</template>
