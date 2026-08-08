<script setup lang="ts">
import { watch, onMounted, onUnmounted } from 'vue';
import { IconUsers, IconX } from '@tabler/icons-vue';
import AccountsView from './AccountsView.vue';
import { loadAccounts } from '../Stores/Accounts';
import { CLOSE_OVERLAYS_EVENT } from '../Stores/Idle';

const props = defineProps<{
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
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

.AccountsModal_Overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: var(--filter-blur);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 100;
}

.AccountsModal_Dialog {
    width: 100%;
    height:100%;
    background: var(--background-modal-primary);
    display: flex;
    flex-direction: column;
    position: relative;
    overflow: hidden;
}

.AccountsModal_Head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.15rem 1.5rem 0.85rem;
    background: var(--background-modal-primary);
    border-bottom: var(--border-modal-style);
    flex-shrink: 0;
}

.AccountsModal_Title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
}

.AccountsModal_Icon {
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

.AccountsModal_Titles {
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

.AccountsModal_Close {
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

.AccountsModal_Body {
    flex: 1;
    min-height: 0;
    display: flex;
    overflow: hidden;

    :deep(.Acct) {
        flex: 1;
        min-width: 0;
        display: flex;
    }
}

.AccountsModal-enter-active,
.AccountsModal-leave-active {
    transition: opacity 200ms ease;

    .AccountsModal_Dialog {
        transition: transform 220ms ease, opacity 200ms ease;
    }
}

.AccountsModal-enter-from,
.AccountsModal-leave-to {
    opacity: 0;

    .AccountsModal_Dialog {
        transform: translateY(8px) scale(0.97);
        opacity: 0;
    }
}
</style>