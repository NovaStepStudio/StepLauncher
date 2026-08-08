<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import { IconX } from '@tabler/icons-vue';
import {
    createAccount,
    updateAccount,
    loginAuthlib,
    ACCOUNT_LOGIN_START_EVENT,
    typeLabel,
    typeDescription,
    type AccountInfo,
    type AuthlibLoginReq,
} from '../Stores/Accounts';

const props = defineProps<{
    visible: boolean;
    mode: 'offline' | 'authlib';
    editing?: AccountInfo | null;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
    (e: 'done', ok: boolean, message: string): void;
}>();

const busy = ref(false);
const msg = ref('');
const msgOk = ref(true);

const offlineForm = reactive<{ name: string; username: string }>({ name: '', username: '' });
const authForm = reactive<AuthlibLoginReq>({ authServerUrl: '', username: '', password: '' });

watch(
    () => props.visible,
    (v) => {
        if (!v) return;
        msg.value = '';
        busy.value = false;
        offlineForm.name = props.editing?.name ?? '';
        offlineForm.username = props.editing?.username ?? '';
        authForm.authServerUrl = '';
        authForm.username = '';
        authForm.password = '';
    }
);

function close() {
    emit('update:visible', false);
}

async function submitOffline() {
    if (busy.value) return;
    busy.value = true;
    msg.value = '';
    const err = props.editing
        ? await updateAccount(props.editing.id, { type: 'offline', name: offlineForm.name, username: offlineForm.username })
        : await createAccount({ type: 'offline', name: offlineForm.name, username: offlineForm.username });
    busy.value = false;
    if (err) {
        msg.value = err;
        msgOk.value = false;
        return;
    }
    emit('done', true, props.editing ? 'Cuenta actualizada correctamente.' : 'Cuenta añadida correctamente.');
    close();
}

function submitAuth() {
    if (busy.value) return;
    busy.value = true;
    msg.value = '';
    const res = loginAuthlib(authForm);
    busy.value = false;
    if (typeof res === 'string') {
        msg.value = res;
        msgOk.value = false;
        return;
    }
    window.dispatchEvent(new CustomEvent(ACCOUNT_LOGIN_START_EVENT));
    close();
}
</script>

<template>
    <Teleport to="body">
        <Transition name="AccountForm">
            <div v-if="visible" class="AccountForm_Overlay" @click.self="close">
                <div class="AccountForm_Dialog">
                    <div class="AccountForm_Head">
                        <div class="AccountForm_Titles">
                            <h3>{{ editing ? 'Editar cuenta' : (mode === 'offline' ? 'Añadir cuenta local' : 'Iniciar sesión') }}</h3>
                            <span class="AccountForm_Badge">{{ typeLabel(mode) }}</span>
                        </div>
                        <button class="AccountForm_Close" title="Cerrar" @click="close">
                            <IconX stroke="2" />
                        </button>
                    </div>

                    <div class="AccountForm_Body">
                        <p class="AccountForm_Desc">{{ typeDescription(mode) }}</p>

                        <template v-if="mode === 'offline'">
                            <label class="AccountForm_Field">
                                <span>Nombre de jugador</span>
                                <input class="SsIn" v-model="offlineForm.username" placeholder="Steve" autocomplete="off" />
                            </label>
                            <label class="AccountForm_Field">
                                <span>Etiqueta (opcional)</span>
                                <input class="SsIn" v-model="offlineForm.name" placeholder="Mi cuenta offline" autocomplete="off" />
                            </label>
                            <p class="AccountForm_Note">Solo juega con un nombre: no necesita conexión ni contraseña.</p>
                        </template>

                        <template v-else>
                            <label class="AccountForm_Field">
                                <span>Servidor de la cuenta</span>
                                <input class="SsIn" v-model="authForm.authServerUrl" placeholder="https://auth.mi-servidor.net" autocomplete="off" />
                            </label>
                            <label class="AccountForm_Field">
                                <span>Email o usuario</span>
                                <input class="SsIn" v-model="authForm.username" placeholder="tucorreo@ejemplo.com" autocomplete="off" />
                            </label>
                            <label class="AccountForm_Field">
                                <span>Contraseña</span>
                                <input class="SsIn" type="password" v-model="authForm.password" placeholder="••••••••" autocomplete="off" />
                            </label>
                            <p class="AccountForm_Note">Se comprueba tu cuenta contra el servidor y se guarda la sesión. Verás el progreso en una ventana propia.</p>
                        </template>

                        <p v-if="msg" :class="['AccountForm_Msg', { error: !msgOk }]">{{ msg }}</p>
                    </div>

                    <div class="AccountForm_Footer">
                        <button class="SsBtn" :disabled="busy" @click="close">Cancelar</button>
                        <button
                            v-if="mode === 'offline'"
                            class="SsBtn SsBtnPrimary"
                            :disabled="busy || !offlineForm.username.trim()"
                            @click="submitOffline"
                        >{{ editing ? 'Guardar cambios' : 'Crear cuenta' }}</button>
                        <button
                            v-else
                            class="SsBtn SsBtnPrimary"
                            :disabled="busy || !authForm.authServerUrl.trim() || !authForm.username.trim() || !authForm.password"
                            @click="submitAuth"
                        >Iniciar sesión</button>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use '../Styles/Settings.scss' as *;

.AccountForm_Overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: var(--filter-blur);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 115;
}

.AccountForm_Dialog {
    width: 24rem;
    max-width: 92vw;
    background: var(--background-modal-primary);
    border: var(--border-modal-style);
    border-radius: 0.85rem;
    box-shadow: var(--shadow-settings-normal) #000a;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.AccountForm_Head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.75rem;
    padding: 1.1rem 1.35rem 0.85rem;
    border-bottom: var(--border-modal-style);
    background: #0005;
}

.AccountForm_Titles {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    min-width: 0;

    h3 {
        margin: 0;
        font-family: var(--font-primary), Arial, sans-serif;
        font-size: calc(0.95rem * var(--font-size-primary, 1));
        font-weight: 600;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
}

.AccountForm_Badge {
    font-size: 0.6rem;
    padding: 0.12rem 0.4rem;
    border-radius: 0.3rem;
    background: color-mix(in srgb, var(--background-button-primary) 15%, transparent);
    color: color-mix(in srgb, var(--background-button-primary) 55%, white 45%);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    white-space: nowrap;
    flex-shrink: 0;
}

.AccountForm_Close {
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
    flex-shrink: 0;
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

.AccountForm_Body {
    display: flex;
    flex-direction: column;
    gap: 0.85rem;
    padding: 1.1rem 1.35rem 0.25rem;
}

.AccountForm_Desc {
    margin: 0;
    font-size: 0.7rem;
    line-height: 1.45;
    opacity: 0.5;
}

.AccountForm_Field {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;

    > span {
        font-size: 0.7rem;
        opacity: 0.55;
    }

    .SsIn {
        width: 100%;
        box-sizing: border-box;
    }
}

.AccountForm_Note {
    margin: 0;
    font-size: 0.66rem;
    line-height: 1.45;
    opacity: 0.42;
}

.AccountForm_Msg {
    margin: 0;
    font-size: 0.72rem;
    color: color-mix(in srgb, var(--background-button-primary) 55%, white 45%);

    &.error {
        color: var(--color-error);
    }
}

.AccountForm_Footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 0.85rem 1.35rem 1.1rem;
}

.AccountForm-enter-active,
.AccountForm-leave-active {
    transition: opacity 180ms ease;

    .AccountForm_Dialog {
        transition: transform 200ms ease, opacity 180ms ease;
    }
}

.AccountForm-enter-from,
.AccountForm-leave-to {
    opacity: 0;

    .AccountForm_Dialog {
        transform: translateY(8px) scale(0.97);
        opacity: 0;
    }
}
</style>
