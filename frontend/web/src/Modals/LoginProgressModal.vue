<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue';
import { pendingLogin, cancelLogin, type AuthLoginResult } from '../Stores/Accounts';
import { CLOSE_OVERLAYS_EVENT } from '../Stores/Idle';

const props = defineProps<{
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
}>();

const state = ref<'pending' | 'ok' | 'error'>('pending');
const error = ref('');

let currentPromise: Promise<AuthLoginResult> | null = null;
let closeTimer: number | null = null;

function reset() {
    state.value = 'pending';
    error.value = '';
    if (closeTimer !== null) {
        window.clearTimeout(closeTimer);
        closeTimer = null;
    }
}

function applyResult(res: AuthLoginResult) {
    if (res.ok) {
        console.log('[LoginProgressModal] login correcto');
        state.value = 'ok';
        closeTimer = window.setTimeout(() => {
            emit('update:visible', false);
        }, 1200);
    } else {
        console.warn('[LoginProgressModal] login fallido:', res.error);
        state.value = 'error';
        error.value = res.error;
    }
}

function onCancel() {
    console.log('[LoginProgressModal] cancelando login...');
    currentPromise = null;
    reset();
    emit('update:visible', false);
    cancelLogin();
}

function onCloseOverlays() {
    reset();
    emit('update:visible', false);
}

watch(
    () => props.visible,
    async (v) => {
        if (!v) {
            currentPromise = null;
            return;
        }
        reset();
        const p = pendingLogin();
        if (!p) {
            console.warn('[LoginProgressModal] abierto sin login en curso; cerrando');
            emit('update:visible', false);
            return;
        }
        currentPromise = p;
        const res = await p;
        if (currentPromise !== p) return;
        applyResult(res);
    }
);

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    if (closeTimer !== null) {
        window.clearTimeout(closeTimer);
        closeTimer = null;
    }
});
</script>

<template>
    <Teleport to="body">
        <Transition name="LoginProgressModal">
            <div v-if="visible" class="LoginProgressModal_Overlay" @click.self="emit('update:visible', false)">
                <div class="LoginProgressModal_Dialog">
                    <div class="LoginProgressModal_Head">
                        <span class="LoginProgressModal_Icon">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
                        </span>
                        <div class="LoginProgressModal_Titles">
                            <h3>Iniciando sesión...</h3>
                            <p>Autenticando contra el servidor Yggdrasil</p>
                        </div>
                    </div>

                    <div class="LoginProgressModal_Body">
                        <template v-if="state === 'pending'">
                            <img class="LoginProgressModal_Spinner" src="../../assets/gif/loading.gif" alt="">
                            <p class="LoginProgressModal_Text">Comprobando credenciales. Esto puede tardar unos segundos.</p>
                        </template>
                        <template v-else-if="state === 'ok'">
                            <svg class="LoginProgressModal_Ok" width="34" height="34" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
                            <p class="LoginProgressModal_Text">Sesión iniciada correctamente.</p>
                        </template>
                        <template v-else>
                            <svg class="LoginProgressModal_Err" width="34" height="34" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                            <p class="LoginProgressModal_Text">{{ error }}</p>
                        </template>
                    </div>

                    <div class="LoginProgressModal_Footer">
                        <button
                            v-if="state === 'pending'"
                            class="SsBtn"
                            @click="onCancel"
                        >Cancelar</button>
                        <button
                            v-else
                            class="SsBtn"
                            @click="emit('update:visible', false)"
                        >Cerrar</button>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use '../Styles/Settings.scss' as *;

.LoginProgressModal_Overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: var(--filter-blur);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 110;
}

.LoginProgressModal_Dialog {
    width: 22rem;
    max-width: 90vw;
    background: var(--background-modal-primary);
    border: var(--border-modal-style);
    border-radius: 0.85rem;
    box-shadow: var(--shadow-settings-normal) #000a;
    display: flex;
    flex-direction: column;
    padding: 1.25rem 1.35rem 1.1rem;
    gap: 1rem;
}

.LoginProgressModal_Head {
    display: flex;
    align-items: center;
    gap: 0.75rem;
}

.LoginProgressModal_Icon {
    width: 2.4rem;
    height: 2.4rem;
    flex-shrink: 0;
    border-radius: 0.6rem;
    background: linear-gradient(135deg, color-mix(in srgb, var(--background-button-primary) 35%, transparent), color-mix(in srgb, var(--background-button-primary) 10%, transparent));
    border: 1px solid color-mix(in srgb, var(--background-button-primary) 30%, transparent);
    display: flex;
    justify-content: center;
    align-items: center;
    color: color-mix(in srgb, var(--background-button-primary) 55%, white 45%);
}

.LoginProgressModal_Titles {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;

    h3 {
        margin: 0;
        font-family: var(--font-primary), Arial, sans-serif;
        font-size: calc(0.95rem * var(--font-size-primary, 1));
        font-weight: 600;
    }

    p {
        margin: 0;
        font-size: 0.68rem;
        opacity: 0.5;
    }
}

.LoginProgressModal_Body {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.7rem;
    padding: 0.5rem 0;
}

.LoginProgressModal_Spinner {
    width: 2.2rem;
    height: auto;
    image-rendering: pixelated;
}

.LoginProgressModal_Ok {
    color: var(--color-success);
}

.LoginProgressModal_Err {
    color: color-mix(in srgb, var(--color-error) 50%, white 50%);
}

.LoginProgressModal_Text {
    margin: 0;
    font-size: 0.76rem;
    line-height: 1.5;
    opacity: 0.6;
    text-align: center;
}

.LoginProgressModal_Footer {
    display: flex;
    justify-content: center;
}

.LoginProgressModal-enter-active,
.LoginProgressModal-leave-active {
    transition: opacity 180ms ease;

    .LoginProgressModal_Dialog {
        transition: transform 200ms ease, opacity 180ms ease;
    }
}

.LoginProgressModal-enter-from,
.LoginProgressModal-leave-to {
    opacity: 0;

    .LoginProgressModal_Dialog {
        transform: scale(0.95);
        opacity: 0;
    }
}
</style>
