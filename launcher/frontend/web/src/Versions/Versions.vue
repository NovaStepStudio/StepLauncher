<script setup lang="ts">
import { watch, onMounted, onUnmounted } from 'vue';
import { IconBox, IconX } from '@tabler/icons-vue';
import VersionsView from './Content.vue';
import { loadVersions, loadProfiles } from '@/Launcher/Store';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

const props = defineProps<{
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
    (e: 'open-download'): void;
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
        if (v) {
            loadVersions();
            loadProfiles();
        }
    }
);
</script>

<template>
    <Teleport to="body">
        <Transition name="VersionsModal">
            <div v-if="visible" class="VersionsModal_Overlay" @click.self="emit('update:visible', false)">
                <div class="VersionsModal_Dialog">
                    <div class="VersionsModal_Head">
                        <div class="VersionsModal_Title">
                            <span class="VersionsModal_Icon"><IconBox stroke="2" /></span>
                            <div class="VersionsModal_Titles">
                                <h3>Versiones y perfiles</h3>
                                <p>Elige con qué version jugar y crea tus propios perfiles</p>
                            </div>
                        </div>
                        <button class="VersionsModal_Close" title="Cerrar" @click="emit('update:visible', false)">
                            <IconX stroke="2" />
                        </button>
                    </div>
                    <div class="VersionsModal_Body">
                        <VersionsView @open-download="emit('open-download')" />
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use './Styles/Versions.scss';
</style>
