<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import {
    accounts,
    loadAccounts,
    deleteAccount,
    setSelected,
    refreshAccount,
    refreshAllAccounts,
    selectedAccountId,
    typeLabel,
    accountAvatars,
    fetchAccountAvatar,
    type AccountInfo,
} from './Store';
import { openDialog, ask } from '@/Common/Overlays/Store';
import { EventsOn } from '@wailsjs/runtime/runtime';

const busy = ref(false);
const msg = ref('');
const msgOk = ref(false);

const formMode = ref<'offline' | 'authlib'>('offline');
const editing = ref<AccountInfo | null>(null);

function setMsg(text: string, ok = true) {
    msg.value = text;
    msgOk.value = ok;
}

function openCreate(mode: 'offline' | 'authlib') {
    editing.value = null;
    formMode.value = mode;
    openDialog('account-form', { mode: formMode.value, editing: editing.value }, { done: onFormDone });
}

function openEdit(a: AccountInfo) {
    editing.value = a;
    formMode.value = 'offline';
    openDialog('account-form', { mode: formMode.value, editing: editing.value }, { done: onFormDone });
}

function onFormDone(ok: boolean, message: string) {
    if (ok) {
        setMsg(message);
        loadAccounts();
    } else {
        setMsg(message, false);
        loadAccounts();
    }
}

function onRefreshResult(data: any) {
    const payload = typeof data === 'string' ? JSON.parse(data) : data;
    if (!payload?.id) return;
    if (payload.ok) {
        setMsg(payload.renewed ? 'Sesión renovada correctamente.' : 'Sesión verificada correctamente.');
    } else {
        setMsg(payload?.error ?? 'No se pudo renovar la sesión.', false);
    }
    loadAccounts();
}

async function remove(a: AccountInfo) {
    if (busy.value) return;
    const r = await ask({
        title: 'Eliminar cuenta',
        message: `¿Eliminar la cuenta «${a.username}»? Esta acción no se puede deshacer.`,
        confirmLabel: 'Eliminar',
        danger: true,
    });
    if (!r.confirmed) return;
    busy.value = true;
    setMsg('');
    const err = await deleteAccount(a.id);
    busy.value = false;
    if (err) {
        setMsg(err, false);
        return;
    }
    setMsg('Cuenta eliminada.');
}

async function select(id: string) {
    if (busy.value) return;
    busy.value = true;
    const err = await setSelected(id);
    busy.value = false;
    if (err) setMsg(err, false);
}

async function renew(a: AccountInfo) {
    if (busy.value) return;
    busy.value = true;
    setMsg('Comprobando la sesión...');
    const err = await refreshAccount(a.id);
    if (err) {
        setMsg(err, false);
        busy.value = false;
    }
}

async function renewAll() {
    if (busy.value) return;
    busy.value = true;
    setMsg('Refrescando datos...');
    const res = await refreshAllAccounts();
    busy.value = false;
    loadAccounts();
    if (res && res !== '0') {
        setMsg('Refrescando sesiones y skins...');
    } else {
        setMsg('Datos actualizados.');
    }
}

function sessionLabel(a: AccountInfo): string {
    if (a.type === 'offline') return '';
    if (!a.hasToken) return 'Sin sesión';
    return a.sessionValid ? 'Sesión OK' : 'Sesión expirada';
}

function sessionClass(a: AccountInfo): string {
    if (a.type === 'offline' || !a.hasToken) return '';
    return a.sessionValid ? 'ok' : 'bad';
}

let offRefresh: (() => void) | null = null;

onMounted(() => {
    offRefresh = EventsOn('account_refresh', onRefreshResult);
    loadAccounts();
    for (const a of accounts.value) {
        if (a.type === 'authlib') fetchAccountAvatar(a.id);
    }
});

onUnmounted(() => {
    offRefresh?.();
});
</script>

<template>
    <div class="Acct">
        <div class="Acct_Body">
            <p class="Acct_Info">
                Estas son tus cuentas. La cuenta <strong>marcada</strong> es la que se usa al pulsar <strong>Jugar</strong>:
                una cuenta <strong>local</strong> solo necesita un nombre, y una <strong>en línea</strong> usa tu usuario
                y contraseña para jugar con tu skin y tu nombre.
            </p>

            <div v-if="accounts.length" class="Acct_List">
                <div
                    v-for="a in accounts"
                    :key="a.id"
                    :class="['Acct_Item', { active: a.id === selectedAccountId, dim: a.id !== selectedAccountId }]"
                    @click="select(a.id)"
                >
                    <div class="Acct_Avatar">
                        <img v-if="accountAvatars[a.id]" :src="accountAvatars[a.id]" alt="" />
                        <span v-else>{{ a.username.slice(0, 1).toUpperCase() }}</span>
                    </div>
                    <div class="Acct_Main">
                        <div class="Acct_ItemTop">
                            <span class="Acct_Username">{{ a.username }}</span>
                            <span class="Acct_Badge">{{ typeLabel(a.type) }}</span>
                            <span v-if="a.type === 'authlib'" :class="['Acct_Session', sessionClass(a)]">{{ sessionLabel(a) }}</span>
                        </div>
                        <span class="Acct_Sub">
                            <template v-if="a.type === 'authlib' && a.serverName">{{ a.serverName }}</template>
                            <template v-else-if="a.type === 'authlib' && a.authServerUrl">{{ a.authServerUrl }}</template>
                            <template v-else>{{ a.uuid }}</template>
                        </span>
                    </div>
                    <div class="Acct_Tools">
                        <button v-if="a.type === 'authlib'" class="Acct_Tool" title="Refrescar sesión" :disabled="busy" @click.stop="renew(a)">
                            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M1 4v6h6"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
                        </button>
                        <button v-if="a.type === 'offline'" class="Acct_Tool" title="Editar" :disabled="busy" @click.stop="openEdit(a)">
                            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.12 2.12 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                        </button>
                        <button class="Acct_Tool Acct_ToolDanger" title="Eliminar" :disabled="busy" @click.stop="remove(a)">
                            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                        </button>
                    </div>
                </div>
            </div>
            <div v-else class="Acct_Empty">
                <svg width="34" height="34" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
                <p>Aún no hay cuentas. Añade una para poder jugar.</p>
            </div>

            <div class="Acct_Selected">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
                <span v-if="selectedAccountId">Cuenta activa: usa <strong>la marcada</strong> al pulsar Jugar.</span>
                <span v-else>Sin cuenta activa: toca una cuenta para seleccionarla.</span>
            </div>

            <p v-if="msg" :class="['Acct_Msg', { error: !msgOk }]">{{ msg }}</p>
        </div>

        <aside class="Acct_Sidebar">
            <div class="Acct_SideHead">
                <span class="Acct_SideLabel">Cuentas</span>
                <p class="Acct_SideDesc">Toca una cuenta para jugar con ella. Añade la tuya y empieza a jugar.</p>
            </div>

            <button class="SsBtn Acct_SideBtn" :disabled="busy" @click="openCreate('offline')">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                <span class="Acct_SideBtnTxt">
                    <span class="Acct_SideBtnTitle">Añadir cuenta local</span>
                    <span class="Acct_SideBtnSub">Solo con un nombre, sin sesión</span>
                </span>
            </button>
            <button class="SsBtn Acct_SideBtn" :disabled="busy" @click="openCreate('authlib')">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0z"/><path d="M6 21v-2a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v2"/></svg>
                <span class="Acct_SideBtnTxt">
                    <span class="Acct_SideBtnTitle">Iniciar sesión con tu cuenta</span>
                    <span class="Acct_SideBtnSub">Usuario y contraseña de tu servidor</span>
                </span>
            </button>

            <div class="Acct_SideDivider"></div>

            <button class="SsBtn Acct_SideBtn" :disabled="busy" @click="renewAll">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M1 4v6h6"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
                <span class="Acct_SideBtnTxt">
                    <span class="Acct_SideBtnTitle">Refrescar datos</span>
                    <span class="Acct_SideBtnSub">Actualiza la skin y la sesión</span>
                </span>
            </button>
        </aside>
    </div>
</template>

<style scoped lang="scss">
@use './Styles/Content.scss';
</style>