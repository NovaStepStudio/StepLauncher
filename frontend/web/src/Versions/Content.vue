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
} from '@/Launcher/Store';
import { openDialog, ask } from '@/Common/Overlays/Store';

const emit = defineEmits<{
    (e: 'open-download'): void;
}>();

type Tab = 'versions' | 'profiles';
const tab = ref<Tab>('versions');

const query = ref('');
const busy = ref(false);
const msg = ref('');
const msgOk = ref(true);

const editing = ref<LauncherProfile | null>(null);

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
    openDialog('profile-form', { editing: null }, { done: onFormDone });
}

function openEdit(p: LauncherProfile) {
    editing.value = p;
    openDialog('profile-form', { editing: p }, { done: onFormDone });
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

async function askDelete(p: LauncherProfile) {
    const r = await ask({
        title: 'Eliminar perfil',
        message: `¿Eliminar el perfil «${p.name}»? Esta acción no se puede deshacer.`,
        confirmLabel: 'Eliminar',
        danger: true,
    });
    if (!r.confirmed || busy.value) return;
    busy.value = true;
    setMsg('');
    const err = await deleteProfile(p.name);
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
    </div>
</template>

<style scoped lang="scss">
@use './Styles/Content.scss';
</style>
