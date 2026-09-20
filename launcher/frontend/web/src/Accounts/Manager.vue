<script setup lang="ts">
import { watch, onMounted, onUnmounted } from 'vue';
import { IconUsers, IconX } from '@tabler/icons-vue';
import AccountsView from './Content.vue';
import { loadAccounts } from './Store';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

const props = defineProps<{
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
}>();

function onCloseOverlays() {
    emit('update:visible', false);
}

useOverlayEscape(() => emit('update:visible', false), { isActive: () => props.visible });

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});

watch(
    () => props.visible,
    (v) => {
        if (v) loadAccounts();
    }
);
</script>

<template>
    <Teleport to="body">
        <Transition name="AccountsModal">
            <div v-if="visible" class="AccountsModal_Overlay" @click.self="emit('update:visible', false)">
                <div class="AccountsModal_Dialog">
                    <div class="AccountsModal_Head">
                        <div class="AccountsModal_Title">
                            <span class="AccountsModal_Icon"><IconUsers stroke="2" /></span>
                            <div class="AccountsModal_Titles">
                                <h3>Cuentas</h3>
                                <p>Gestiona tus sesiones de juego</p>
                            </div>
                        </div>
                        <button class="AccountsModal_Close" title="Cerrar" @click="emit('update:visible', false)">
                            <IconX stroke="2" />
                        </button>
                    </div>
                    <div class="AccountsModal_Body">
                        <AccountsView />
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use './Styles/Manager.scss';
</style>