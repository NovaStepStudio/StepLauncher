<script setup lang="ts">
import { ref, computed } from 'vue';
import {
    IconSearch,
    IconCheck,
    IconDownload,
    IconUserPlus,
} from '@tabler/icons-vue';
import {
    installedVersions,
    selectedVersion,
    selectVersion,
    groupedVersions,
    profiles,
    selectedProfile,
    setSelectedProfile,
    deleteProfile,
    loadVersions,
    loadProfiles,
    type LauncherProfile,
} from '../Stores/Launcher';
import ProfileFormModal from './ProfileFormModal.vue';

const emit = defineEmits<{
    (e: 'open-download'): void;
}>();

type Tab = 'versions' | 'profiles';
const tab = ref<Tab>('versions');

const query = ref('');
const busy = ref(false);
const msg = ref('');
const msgOk = ref(true);

const formVisible = ref(false);
const editing = ref<LauncherProfile | null>(null);
const pendingDelete = ref<LauncherProfile | null>(null);

const visibleGroups = computed(() => {
    const q = query.value.trim().toLowerCase();
    if (!q) return groupedVersions.value;
    return groupedVersions.value
        .map((g) => ({ ...g, items: g.items.filter((v) => v.id.toLowerCase().includes(q)) }))
        .filter((g) => g.items.length > 0);
});

const visibleCount = computed(() =>
    visibleGroups.value.reduce((acc, g) => acc + g.items.length, 0)
);

const profileList = computed(() => Object.values(profiles.value));

function typeBadge(type: string): string {
    switch (type) {
        case 'release': return 'Release';
        case 'snapshot': return 'Snapshot';
        case 'old_beta': return 'Beta';
        case 'old_alpha': return 'Alpha';
        default: return type || 'Release';
    }
}

function typeHue(type: string): string {
    switch (type) {
        case 'release': return 'IsRelease';
        case 'snapshot': return 'IsSnapshot';
        case 'old_beta': return 'IsBeta';
        case 'old_alpha': return 'IsAlpha';
        default: return 'IsOther';
    }
}

function setMsg(text: string, ok = true) {
    msg.value = text;
    msgOk.value = ok;
}

function openCreate() {
    editing.value = null;
    formVisible.value = true;
}

function openEdit(p: LauncherProfile) {
    editing.value = p;
    formVisible.value = true;
}

function onFormDone(ok: boolean, message: string) {
    setMsg(message, ok);
    loadProfiles();
}

async function select(id: string) {
    if (busy.value) return;
    busy.value = true;
    const err = await setSelectedProfile(id);
    busy.value = false;
    if (err) setMsg(err, false);
}

function askDelete(p: LauncherProfile) {
    pendingDelete.value = p;
}

async function confirmDelete() {
    if (!pendingDelete.value) return;
    const target = pendingDelete.value;
    pendingDelete.value = null;
    if (busy.value) return;
    busy.value = true;
    setMsg('');
    const err = await deleteProfile(target.name);
    busy.value = false;
    if (err) {
        setMsg(err, false);
        return;
    }
    setMsg('Perfil eliminado.');
}

function profileSub(p: LauncherProfile): string {
    const parts: string[] = [];
    parts.push(p.version && p.version.trim() ? p.version.trim() : 'Cualquier versión');
    if (p.resWidth && p.resHeight) parts.push(`${p.resWidth}×${p.resHeight}`);
    if (p.fullscreen) parts.push('Pantalla completa');
    if (p.gameDir && p.gameDir.trim()) parts.push('Carpeta propia');
    return parts.join(' · ');
}
</script>

<template>
    <div class="Vers">
        <div class="Vers_Body">
            <div class="Vers_Tabs">
                <button class="Vers_Tab" :class="{ active: tab === 'versions' }" @click="tab = 'versions'">
                    Versiones
                    <em v-if="installedVersions.length">{{ installedVersions.length }}</em>
                </button>
                <button class="Vers_Tab" :class="{ active: tab === 'profiles' }" @click="tab = 'profiles'">
                    Perfiles
                    <em v-if="profileList.length">{{ profileList.length }}</em>
                </button>
            </div>

            <template v-if="tab === 'versions'">
                <div class="Vers_Info">
                    Estas son las versiones <strong>descargadas</strong>. Toca la que quieras usar y se
                    marcará en el menú: el botón <strong>Jugar</strong> la lanzará con la configuración del
                    launcher (o la del perfil activo si lo tienes).
                </div>

                <div v-if="installedVersions.length" class="Vers_Search">
                    <IconSearch />
                    <input v-model="query" type="text" placeholder="Buscar versión…" spellcheck="false" />
                    <span v-if="query" class="Vers_SearchCount">{{ visibleCount }} resultados</span>
                </div>

                <div v-if="installedVersions.length" class="Vers_List">
                    <template v-for="group in visibleGroups" :key="group.type">
                        <div class="Vers_GroupHead">
                            <span>{{ group.label }}</span>
                            <em>{{ group.items.length }}</em>
                        </div>
                        <button
                            v-for="v in group.items"
                            :key="v.id"
                            class="Vers_Item"
                            :class="{ active: v.id === selectedVersion }"
                            @click="selectVersion(v.id)"
                        >
                            <span class="Vers_ItemBadge" :class="'Vers_TypeHue_' + typeHue(v.type)">{{ typeBadge(v.type) }}</span>
                            <span class="Vers_ItemId">{{ v.id }}</span>
                            <IconCheck v-if="v.id === selectedVersion" class="Vers_ItemCheck" stroke="2.4" />
                        </button>
                    </template>
                    <div v-if="!visibleGroups.length" class="Vers_Empty">
                        <IconSearch stroke="2" />
                        <span>Sin resultados para «{{ query }}»</span>
                    </div>
                </div>
                <div v-else class="Vers_Empty">
                    <svg width="34" height="34" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/></svg>
                    <p>Aún no hay versiones descargadas. Descarga una desde el botón «Descargar versión».</p>
                </div>

                <div class="Vers_Selected">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
                    <span v-if="selectedVersion">Versión activa: jugarás con <strong>{{ selectedVersion }}</strong>.</span>
                    <span v-else>Sin versión activa: toca una versión para seleccionarla.</span>
                </div>
            </template>

            <template v-else>
                <div class="Vers_Info">
                    Un perfil es una forma de juego con <strong>su propia configuración</strong>: fija una
                    versión, resolución, Java o carpeta. Al pulsar <strong>Jugar</strong> se usa la
                    configuración del perfil y lo que el perfil no define se toma de la configuración del
                    launcher.
                </div>

                <div v-if="profileList.length" class="Vers_List">
                    <div
                        v-for="p in profileList"
                        :key="p.name"
                        :class="['Vers_Item', 'Vers_ProfileItem', { active: p.name === selectedProfile, dim: p.name !== selectedProfile }]"
                        @click="select(p.name)"
                    >
                        <span class="Vers_ProfileAvatar">
                            <img v-if="p.icon" :src="p.icon" alt="" loading="lazy">
                            <span v-else>{{ p.name.slice(0, 1).toUpperCase() }}</span>
                        </span>
                        <div class="Vers_ProfileMain">
                            <span class="Vers_ProfileName">{{ p.name }}</span>
                            <span class="Vers_ProfileSub">{{ profileSub(p) }}</span>
                        </div>
                        <div class="Vers_Tools">
                            <button class="Vers_Tool" title="Editar" :disabled="busy" @click.stop="openEdit(p)">
                                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.12 2.12 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                            </button>
                            <button class="Vers_Tool Vers_ToolDanger" title="Eliminar" :disabled="busy" @click.stop="askDelete(p)">
                                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                            </button>
                        </div>
                    </div>
                </div>
                <div v-else class="Vers_Empty">
                    <svg width="34" height="34" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
                    <p>Aún no hay perfiles. Crea el tuyo para darle un uso mayor a tus versiones.</p>
                </div>

                <div class="Vers_Selected">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
                    <span v-if="selectedProfile">Perfil activo: al pulsar Jugar se usa <strong>{{ selectedProfile }}</strong>.</span>
                    <span v-else>Sin perfil activo: Jugar usa la configuración del launcher.</span>
                </div>
            </template>

            <p v-if="msg" :class="['Vers_Msg', { error: !msgOk }]">{{ msg }}</p>
        </div>

        <aside class="Vers_Sidebar">
            <div class="Vers_SideHead">
                <span class="Vers_SideLabel">Versiones y Perfiles</span>
                <p class="Vers_SideDesc">Toca una versión para jugar con ella. Los perfiles guardan tu configuración por forma de juego.</p>
            </div>

            <button class="SsBtn Vers_SideBtn" @click="emit('open-download')">
                <IconDownload class="Vers_SideIcon" stroke="2" />
                <span class="Vers_SideBtnTxt">
                    <span class="Vers_SideBtnTitle">Descargar versión</span>
                    <span class="Vers_SideBtnSub">Instala una versión desde Mojang</span>
                </span>
            </button>
            <button class="SsBtn Vers_SideBtn" :disabled="!installedVersions.length" @click="openCreate">
                <IconUserPlus class="Vers_SideIcon" stroke="2" />
                <span class="Vers_SideBtnTxt">
                    <span class="Vers_SideBtnTitle">Crear perfil</span>
                    <span class="Vers_SideBtnSub">Configuración propia de juego</span>
                </span>
            </button>
        </aside>

        <ProfileFormModal v-model:visible="formVisible" :editing="editing" @done="onFormDone" />

        <Teleport to="body">
            <Transition name="ProfileForm">
                <div v-if="pendingDelete" class="ProfileForm_Overlay" @click.self="pendingDelete = null">
                    <div class="ProfileForm_Dialog">
                        <div class="ProfileForm_Head">
                            <div class="ProfileForm_Titles">
                                <h3>Eliminar perfil</h3>
                            </div>
                            <button class="ProfileForm_Close" title="Cerrar" @click="pendingDelete = null">
                                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                            </button>
                        </div>
                        <div class="ProfileForm_Body">
                            <p class="ProfileForm_Note">¿Eliminar el perfil «{{ pendingDelete?.name }}»? Esta acción no se puede deshacer.</p>
                        </div>
                        <div class="ProfileForm_Footer">
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

.Vers {
    flex: 1;
    min-width: 0;
    display: flex;
    height: 100%;
}

.Vers_Body {
    flex: 1;
    min-width: 0;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    padding-bottom: 1rem;
}

.Vers_Tabs {
    display: flex;
    gap: 0.5rem;
    padding: 1.1rem 2rem 0;
    flex-shrink: 0;
}

.Vers_Tab {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    background: var(--control-bg);
    border: 1px solid var(--control-border);
    border-radius: 0.5rem;
    color: var(--text-secondary);
    font-size: 0.72rem;
    padding: 0.45rem 0.9rem;
    cursor: pointer;
    transition: background 150ms, border-color 150ms, color 150ms;

    em {
        font-style: normal;
        font-size: 0.6rem;
        padding: 0.1rem 0.3rem;
        border-radius: 0.3rem;
        background: rgba(255, 255, 255, 0.08);
    }

    &:hover {
        background: color-mix(in srgb, var(--background-button-primary) 50%, gray 10%);
        color: var(--text-primary);
    }

    &.active {
        border-color: color-mix(in srgb, var(--background-button-primary) 55%, gray 50%);
        background-color: color-mix(in srgb, var(--background-button-primary) 80%, gray 15%);

        em {
            background: color-mix(in srgb, var(--background-button-primary) 50%, gray 25%);
            color: color-mix(in srgb, var(--background-button-primary) 55%, white 45%);
        }
    }
}

.Vers_Info {
    margin: 1rem 2rem 0;
    font-size: 0.78rem;
    line-height: 1.55;
    opacity: 0.55;

    strong {
        opacity: 0.85;
    }
}

.Vers_Search {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 1rem 2rem 0.75rem;
    padding: 0.5rem 0.75rem;
    background: var(--control-bg);
    border: 1px solid var(--control-border);
    border-radius: 0.55rem;

    svg {
        width: 0.9rem;
        height: 0.9rem;
        opacity: 0.5;
        flex-shrink: 0;
    }

    input {
        flex: 1;
        min-width: 0;
        background: none;
        border: none;
        outline: none;
        color: var(--text-primary);
        font-family: var(--font-secundary), Arial, sans-serif;
        font-size: 0.74rem;

        &::placeholder {
            color: var(--text-secondary);
            opacity: 0.5;
        }
    }
}

.Vers_SearchCount {
    font-size: 0.62rem;
    opacity: 0.5;
    white-space: nowrap;
}

.Vers_List {
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
    margin: 0.25rem 2rem 1rem;
}

.Vers_GroupHead {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-top: 0.5rem;

    span {
        font-size: 0.62rem;
        text-transform: uppercase;
        letter-spacing: 0.09em;
        opacity: 0.4;
    }

    em {
        font-style: normal;
        font-size: 0.6rem;
        opacity: 0.35;
    }
}

.Vers_Item {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    width: 100%;
    text-align: left;
    background: var(--control-bg);
    border: 1px solid var(--control-border);
    border-radius: 0.55rem;
    padding: 0.55rem 0.8rem;
    cursor: pointer;
    color: var(--text-primary);
    font-family: var(--font-secundary), Arial, sans-serif;
    transition: background 150ms, border-color 150ms, transform 150ms;

    &:hover {
        border-color: color-mix(in srgb, var(--background-button-primary) 50%, gray 40%);
        background: color-mix(in srgb, var(--background-button-primary) 50%, gray 20%);
        transform: translateY(-1px);
    }

    &.active {
        background: color-mix(in srgb, var(--background-button-primary) 50%, gray 15%);
        border-color: rgba(255, 255, 255, 0.16);
    }

    &.dim {
        opacity: 0.55;

        &:hover {
            opacity: 0.8;
        }
    }
}

.Vers_ItemBadge {
    font-size: 0.58rem;
    padding: 0.1rem 0.4rem;
    border-radius: 0.3rem;
    background: color-mix(in srgb, var(--background-button-primary) 15%, transparent);
    color: color-mix(in srgb, var(--background-button-primary) 55%, white 45%);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    white-space: nowrap;
    flex-shrink: 0;
}

.Vers_TypeHue_IsRelease {
    background: color-mix(in srgb, var(--color-success) 16%, transparent);
    color: color-mix(in srgb, var(--color-success) 50%, white 50%);
}

.Vers_TypeHue_IsSnapshot {
    background: color-mix(in srgb, var(--color-warning) 16%, transparent);
    color: color-mix(in srgb, var(--color-warning) 50%, white 50%);
}

.Vers_TypeHue_IsBeta {
    background: color-mix(in srgb, var(--color-warning) 16%, transparent);
    color: var(--color-warning);
}

.Vers_TypeHue_IsAlpha {
    background: color-mix(in srgb, var(--color-error) 16%, transparent);
    color: color-mix(in srgb, var(--color-error) 50%, white 50%);
}

.Vers_TypeHue_IsOther {
    background: rgba(255, 255, 255, 0.1);
    color: #c9c9d6;
}

.Vers_ItemId {
    flex: 1;
    min-width: 0;
    font-size: 0.8rem;
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.Vers_ItemCheck {
    width: 0.95rem;
    height: 0.95rem;
    color: var(--color-success);
    flex-shrink: 0;
}

.Vers_ProfileItem {
    cursor: pointer;
}

.Vers_ProfileAvatar {
    width: 2.2rem;
    height: 2.2rem;
    flex-shrink: 0;
    border-radius: 0.5rem;
    border: var(--border-style);
    background: color-mix(in srgb, var(--background-button-primary) 50%, transparent);
    overflow: hidden;
    display: flex;
    justify-content: center;
    align-items: center;
    font-family: var(--font-primary), Arial, sans-serif;
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);

    img {
        width: 100%;
        height: 100%;
        object-fit: cover;
        image-rendering: pixelated;
    }
}

.Vers_ProfileMain {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.12rem;
}

.Vers_ProfileName {
    font-size: 0.82rem;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.Vers_ProfileSub {
    font-size: 0.66rem;
    opacity: 0.5;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.Vers_Tools {
    display: flex;
    gap: 0.35rem;
    flex-shrink: 0;
}

.Vers_Tool {
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
        background: color-mix(in srgb, var(--background-button-primary) 20%, transparent);
        border-color: color-mix(in srgb, var(--background-button-primary) 45%, transparent);
        color: color-mix(in srgb, var(--background-button-primary) 55%, white 45%);
    }

    &:disabled {
        opacity: 0.4;
        cursor: default;
    }
}

.Vers_ToolDanger {
    &:hover:not(:disabled) {
        background: color-mix(in srgb, var(--color-error) 16%, transparent);
        border-color: color-mix(in srgb, var(--color-error) 40%, transparent);
        color: color-mix(in srgb, var(--color-error) 50%, white 50%);
    }
}

.Vers_Empty {
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

    p {
        margin: 0;
    }
}

.Vers_Selected {
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

.Vers_Msg {
    margin: 0.5rem 2rem 1.1rem;
    font-size: 0.75rem;
    color: color-mix(in srgb, var(--background-button-primary) 55%, white 45%);

    &.error {
        color: var(--color-error);
    }
}

.Vers_Sidebar {
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

    .SsBtn.Vers_SideBtn {
        width: 100%;
        justify-content: flex-start;
        align-items: center;
        text-align: left;
        white-space: normal;
        padding: 0.6rem 0.75rem;
        background-color: var(--background-button-primary);
        border: 1px solid color-mix(in srgb, var(--background-button-primary) 28%, transparent);
        color: var(--text-primary);
        text-shadow: var(--text-shadow-primary, none);

        &:hover:not(:disabled) {
            background-color: color-mix(in srgb, var(--background-button-primary) 50%, gray 15%);
            border-color: color-mix(in srgb, var(--background-button-primary) 50%, transparent);
        }

        &:disabled {
            opacity: 0.4;
        }
    }
}

.Vers_SideIcon {
    width: 0.95rem;
    height: 0.95rem;
    flex-shrink: 0;
    opacity: 0.8;
}

.Vers_SideHead {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    padding: 0 0.15rem 0.35rem;
}

.Vers_SideLabel {
    font-size: 0.62rem;
    text-transform: uppercase;
    letter-spacing: 0.09em;
    opacity: 0.4;
}

.Vers_SideDesc {
    margin: 0;
    font-size: 0.66rem;
    line-height: 1.45;
    opacity: 0.5;
}

.Vers_SideBtnTxt {
    display: flex;
    flex-direction: column;
    gap: 0.12rem;
    min-width: 0;
}

.Vers_SideBtnTitle {
    font-size: calc(0.72rem * var(--font-size-secundary, 1));
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.Vers_SideBtnSub {
    font-size: 0.6rem;
    opacity: 0.55;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.ProfileForm_Overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: var(--filter-blur);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 115;
}

.ProfileForm_Dialog {
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

.ProfileForm_Head {
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

.ProfileForm_Close {
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

.ProfileForm_Body {
    display: flex;
    flex-direction: column;
    gap: 0.85rem;
    padding: 1.1rem 1.35rem 0.25rem;
}

.ProfileForm_Note {
    margin: 0;
    font-size: 0.66rem;
    line-height: 1.45;
    opacity: 0.42;
}

.ProfileForm_Footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 0.85rem 1.35rem 1.1rem;
}

.ProfileForm-enter-active,
.ProfileForm-leave-active {
    transition: opacity 180ms ease;

    .ProfileForm_Dialog {
        transition: transform 200ms ease, opacity 180ms ease;
    }
}

.ProfileForm-enter-from,
.ProfileForm-leave-to {
    opacity: 0;

    .ProfileForm_Dialog {
        transform: translateY(8px) scale(0.97);
        opacity: 0;
    }
}
</style>
