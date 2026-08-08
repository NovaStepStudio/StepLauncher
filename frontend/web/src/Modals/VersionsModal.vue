<script setup lang="ts">
import { watch, onMounted, onUnmounted } from 'vue';
import { IconBox, IconX } from '@tabler/icons-vue';
import VersionsView from './VersionsView.vue';
import { loadVersions, loadProfiles } from '../Stores/Launcher';
import { CLOSE_OVERLAYS_EVENT } from '../Stores/Idle';

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

.VersionsModal_Overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: var(--filter-blur);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 100;
}

.VersionsModal_Dialog {
    width: 100%;
    height: 100%;
    background: var(--background-modal-primary);
    display: flex;
    flex-direction: column;
    position: relative;
    overflow: hidden;
}

.VersionsModal_Head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.15rem 1.5rem 0.85rem;
    background: var(--background-modal-primary);
    border-bottom: var(--border-modal-style);
    flex-shrink: 0;
    position: relative;
    overflow: hidden;
}

.VersionsModal_Title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    position: relative;
    z-index: 1;
}

.VersionsModal_Icon {
    width: 2.4rem;
    height: 2.4rem;
    border-radius: 0.6rem;
    background: color-mix(in srgb, var(--background-modal-primary) 50%, gray 5%);
    border: var(--border-style);
    display: flex;
    justify-content: center;
    align-items: center;
    color: color-mix(in srgb, var(--background-modal-primary) 25%, white 85%);
    flex-shrink: 0;

    svg {
        width: 1.25rem;
        height: 1.25rem;
    }
}

.VersionsModal_Titles {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;

    h3 {
        margin: 0;
        font-family: var(--font-primary), Arial, sans-serif;
        font-size: calc(1rem * var(--font-size-primary, 1));
        font-weight: 600;
    }

    p {
        margin: 0;
        font-size: 0.68rem;
        opacity: 0.5;
    }
}

.VersionsModal_Close {
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    border-radius: 0.4rem;
    display: flex;
    justify-content: center;
    align-items: center;
    height: 2rem;
    width: 2rem;
    transition: background 150ms, color 150ms;

    svg {
        width: 1.1rem;
        height: 1.1rem;
    }

    &:hover {
        background: #1111;
        color: var(--text-primary);
    }
}

.VersionsModal_Body {
    flex: 1;
    min-height: 0;
    display: flex;
    overflow: hidden;

    :deep(.Vers) {
        flex: 1;
        min-width: 0;
        display: flex;
    }
}

.VersionsModal-enter-active,
.VersionsModal-leave-active {
    transition: opacity 200ms ease;

    .VersionsModal_Dialog {
        transition: transform 220ms ease, opacity 200ms ease;
    }
}

.VersionsModal-enter-from,
.VersionsModal-leave-to {
    opacity: 0;

    .VersionsModal_Dialog {
        transform: translateY(8px) scale(0.97);
        opacity: 0;
    }
}
</style>
