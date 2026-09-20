<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue';
import { pendingLogin, cancelLogin, type AuthLoginResult } from '@/Accounts/Store';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

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

useOverlayEscape(() => emit('update:visible', false), { isActive: () => props.visible });

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
                            <img class="LoginProgressModal_Spinner" src="../../assets/gif/chicken_jockey_run.gif" alt="">
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
@use './Styles/Progress.scss';
</style>
