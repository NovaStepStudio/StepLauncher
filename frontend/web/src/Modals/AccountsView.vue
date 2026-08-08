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
} from '../Stores/Accounts';
import AccountFormModal from './AccountFormModal.vue';
import { EventsOn } from '@wailsjs/runtime/runtime';

const busy = ref(false);
const msg = ref('');
const msgOk = ref(false);

const formVisible = ref(false);
const formMode = ref<'offline' | 'authlib'>('offline');
const editing = ref<AccountInfo | null>(null);
const pendingDelete = ref<AccountInfo | null>(null);

function setMsg(text: string, ok = true) {
    msg.value = text;
    msgOk.value = ok;
}

function openCreate(mode: 'offline' | 'authlib') {
    editing.value = null;
    formMode.value = mode;
    formVisible.value = true;
}

function openEdit(a: AccountInfo) {
    editing.value = a;
    formMode.value = 'offline';
    formVisible.value = true;
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
    pendingDelete.value = a;
}

async function confirmDelete() {
    if (!pendingDelete.value) return;
    const target = pendingDelete.value;
    pendingDelete.value = null;
    if (busy.value) return;
    busy.value = true;
    setMsg('');
    const err = await deleteAccount(target.id);
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

        <AccountFormModal v-model:visible="formVisible" :mode="formMode" :editing="editing" @done="onFormDone" />

        <Teleport to="body">
            <Transition name="AccountForm">
                <div v-if="pendingDelete" class="AccountForm_Overlay" @click.self="pendingDelete = null">
                    <div class="AccountForm_Dialog">
                        <div class="AccountForm_Head">
                            <div class="AccountForm_Titles">
                                <h3>Eliminar cuenta</h3>
                            </div>
                            <button class="AccountForm_Close" title="Cerrar" @click="pendingDelete = null">
                                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                            </button>
                        </div>
                        <div class="AccountForm_Body">
                            <p class="AccountForm_Note">¿Eliminar la cuenta «{{ pendingDelete?.username }}»? Esta acción no se puede deshacer.</p>
                        </div>
                        <div class="AccountForm_Footer">
                            <button class="SsBtn" @click="pendingDelete = null">Cancelar</button>
                            <button class="SsBtn SsBtnDanger" @click="confirmDelete">Eliminar</button>
                        </div>
                    </div>
                </div>
            </Transition>
        </Teleport>
    </div>
</template>

<style scoped lang="scss">
@use '../Styles/Settings.scss' as *;
.Acct {
    flex: 1;
    min-width: 0;
    display: flex;
    height: 100%;
}

.Acct_Body {
    flex: 1;
    min-width: 0;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    padding-bottom: 1rem;
}

.Acct_Info {
    margin: 1.1rem 2rem 0;
    font-size: 0.78rem;
    line-height: 1.55;
    opacity: 0.55;

    strong {
        opacity: 0.85;
    }
}

.Acct_List {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    margin: 1.1rem 2rem 1rem;
}

.Acct_Item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    background: var(--control-bg);
    border: 1px solid var(--control-border);
    border-radius: 0.65rem;
    padding: 0.6rem 0.85rem;
    cursor: pointer;
    transition: background 150ms, border-color 150ms, opacity 150ms, transform 150ms;

    &:hover {
        background: rgba(255, 255, 255, 0.07);
        border-color: rgba(255, 255, 255, 0.16);
        transform: translateY(-1px);
    }

    &.active {
        border-color: color-mix(in srgb, var(--background-button-primary) 35%, gray 50%);
        background:color-mix(in srgb, var(--background-button-primary) 85%, gray 10%);
    }

    &.dim {
        opacity: 0.55;

        &:hover {
            opacity: 0.8;
        }
    }
}

.Acct_Avatar {
    width: 2.3rem;
    height: 2.3rem;
    flex-shrink: 0;
    border-radius: 50%;
    display: flex;
    justify-content: center;
    align-items: center;
    overflow: hidden;

    img {
        width: 100%;
        height: 100%;
        object-fit: cover;
        border-radius: 50%;
        image-rendering: pixelated;
    }

    span {
        font-family: var(--font-primary), Arial, sans-serif;
        font-size: 1rem;
        font-weight: 600;
        color: #fff;
        text-shadow: 0 1px 2px rgba(0, 0, 0, 0.35);
    }
}

.Acct_Main {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
}

.Acct_ItemTop {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.Acct_Username {
    font-size: 0.84rem;
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.Acct_Badge {
    font-size: 0.6rem;
    padding: 0.12rem 0.4rem;
    border-radius: 0.3rem;
    background: color-mix(in srgb, var(--background-button-primary) 15%, transparent);
    color: color-mix(in srgb, var(--background-button-primary) 55%, white 45%);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    white-space: nowrap;
}

.Acct_Session {
    font-size: 0.6rem;
    padding: 0.12rem 0.4rem;
    border-radius: 0.3rem;
    white-space: nowrap;

    &.ok {
        background: color-mix(in srgb, var(--color-success) 15%, transparent);
        color: var(--color-success);
    }

    &.bad {
        background: color-mix(in srgb, var(--color-error) 15%, transparent);
        color: color-mix(in srgb, var(--color-error) 85%, white 15%);
    }
}

.Acct_Sub {
    font-size: 0.68rem;
    opacity: 0.45;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.Acct_Tools {
    display: flex;
    gap: 0.35rem;
    flex-shrink: 0;
}

.Acct_Tool {
    background: none;
    border: 1px solid var(--control-border);
    color: var(--text-secondary);
    cursor: pointer;
    border-radius: 0.4rem;
    display: flex;
    justify-content: center;
    align-items: center;
    height: 1.9rem;
    width: 1.9rem;
    transition: background 150ms, color 150ms, border-color 150ms;

    &:hover:not(:disabled) {
        background: color-mix(in srgb, var(--background-button-primary) 18%, transparent);
        border-color: color-mix(in srgb, var(--background-button-primary) 40%, transparent);
        color: color-mix(in srgb, var(--background-button-primary) 55%, white 45%);
    }

    &:disabled {
        opacity: 0.4;
        cursor: default;
    }
}

.Acct_ToolDanger {
    &:hover:not(:disabled) {
        background: color-mix(in srgb, var(--color-error) 16%, transparent);
        border-color: color-mix(in srgb, var(--color-error) 40%, transparent);
        color: color-mix(in srgb, var(--color-error) 50%, white 50%);
    }
}

.Acct_Empty {
    margin: 1rem 2rem;
    text-align: center;
    opacity: 0.55;
    font-size: 0.78rem;
    padding: 1.25rem 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;

    svg {
        opacity: 0.5;
    }
}

.Acct_Selected {
    margin: 0.25rem 2rem 1rem;
    display: flex;
    align-items: center;
    gap: 0.45rem;
    font-size: 0.68rem;
    opacity: 0.55;

    svg {
        flex-shrink: 0;
        opacity: 0.6;
    }

    strong {
        opacity: 0.85;
    }
}

.Acct_Sidebar {
    width: 20rem;
    flex-shrink: 0;
    background: var(--background-sidebar);
    border-left: var(--border-modal-style);
    padding: 1.1rem 0.85rem 1.25rem;
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 0.5rem;
    overflow-y: auto;

    .SsBtn.Acct_SideBtn {
        width: 100%;
        justify-content: flex-start;
        align-items: center;
        text-align: left;
        white-space: normal;
        padding: 0.6rem 0.75rem;
        background: var(--background-button-primary);
        border: var(--border-style);
        color: var(--text-primary);
        text-shadow: var(--text-shadow-primary, none);

        svg {
            width: 0.95rem;
            height: 0.95rem;
            flex-shrink: 0;
            opacity: 0.8;
        }

        &:hover:not(:disabled) {
            background: color-mix(in srgb, var(--background-button-primary) 55%, gray 25%);
            border-color: color-mix(in srgb, var(--background-modal-primary, #111) 50%, gray 40%);
        }

        &:disabled {
            opacity: 0.4;
        }
    }
}

.Acct_SideHead {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    padding: 0 0.15rem 0.35rem;
}

.Acct_SideLabel {
    font-size: 0.62rem;
    text-transform: uppercase;
    letter-spacing: 0.09em;
    opacity: 0.4;
}

.Acct_SideDesc {
    margin: 0;
    font-size: 0.66rem;
    line-height: 1.45;
    opacity: 0.5;
}

.Acct_SideBtnTxt {
    display: flex;
    flex-direction: column;
    gap: 0.12rem;
    min-width: 0;
}

.Acct_SideBtnTitle {
    font-size: calc(0.72rem * var(--font-size-secundary, 1));
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.Acct_SideBtnSub {
    font-size: 0.6rem;
    opacity: 0.55;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.Acct_SideDivider {
    width: 100%;
    height: 1px;
    background: rgba(255, 255, 255, 0.08);
    margin: 0.15rem 0;
}

.Acct_Msg {
    margin: 0.5rem 2rem 1.1rem;
    font-size: 0.75rem;
    color: color-mix(in srgb, var(--background-button-primary) 55%, white 45%);

    &.error {
        color: var(--color-error);
    }
}

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

    h3 {
        margin: 0;
        font-family: var(--font-primary), Arial, sans-serif;
        font-size: calc(0.95rem * var(--font-size-primary, 1));
        font-weight: 600;
    }
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

.AccountForm_Note {
    margin: 0;
    font-size: 0.66rem;
    line-height: 1.45;
    opacity: 0.42;
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