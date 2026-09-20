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
} from './Store';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

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
    },
    { immediate: true }
);

function close() {
    emit('update:visible', false);
}

useOverlayEscape(close, { isActive: () => props.visible, priority: 2 });

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
@use './Styles/Form.scss';
</style>
