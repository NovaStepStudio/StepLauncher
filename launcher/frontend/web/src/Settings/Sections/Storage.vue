<script setup lang="ts">
import { ref, computed, onMounted, onActivated } from 'vue';
import { GetDirectorySettings, PickDirectory, SetDirectoryMode, RestartApp } from '@wailsjs/StepLauncher/internal/Services/System/systemservice';
import { GetSeparateGameDir, SetSeparateGameDir, GetCacheInfo, ClearAllCache } from '@wailsjs/StepLauncher/internal/Services/Config/configservice';
import { RefreshManifests } from '@wailsjs/StepLauncher/internal/Services/Download/downloadservice';

interface DirectorySettings {
    mode: string;
    customPath: string;
    configured: boolean;
    workDir: string;
    normalDir: string;
    minecraftDir: string;
    minecraftExists: boolean;
    portableDir: string;
}
const dirInfo = ref<DirectorySettings | null>(null);
const dirMode = ref<string>('normal');
const customPath = ref('');
const dirBusy = ref(false);
const dirMsg = ref('');
const dirMsgOk = ref(true);
const separateGameDir = ref(true);

const dirModes = [
    { id: 'normal', label: 'Normal (recomendada)' },
    { id: 'minecraft', label: 'Usar mi .minecraft' },
    { id: 'portable', label: 'Portable (USB)' },
    { id: 'custom', label: 'Personalizada…' },
];

const dirChanged = computed(() => {
    const info = dirInfo.value;
    if (!info) return false;
    if (dirMode.value !== info.mode) return true;
    if (dirMode.value === 'custom' && customPath.value.trim() !== info.customPath) return true;
    return false;
});

async function loadDirectorySettings() {
    try {
        const info = await GetDirectorySettings();
        if (info) {
            dirInfo.value = info;
            dirMode.value = info.mode ?? 'normal';
            customPath.value = info.customPath ?? '';
            if (info.mode === 'minecraft') separateGameDir.value = false;
        }
        const sep = await GetSeparateGameDir();
        if (typeof sep === 'boolean') separateGameDir.value = sep;
    } catch (_e) {}
}

async function pickDirectory() {
    try {
        const p = await PickDirectory();
        if (p) customPath.value = p;
    } catch (_e) {}
}

async function saveDirectory() {
    if (dirBusy.value) return;
    dirBusy.value = true;
    dirMsg.value = '';
    dirMsgOk.value = true;
    try {
        await SetDirectoryMode(
            dirMode.value,
            dirMode.value === 'custom' ? customPath.value.trim() : ''
        );
        if (dirChanged.value) {
            dirMsg.value = 'Carpeta guardada. Reiniciando el launcher…';
            dirMsgOk.value = true;
            try {
                await RestartApp();
            } catch (_e) {}
            return;
        }
        dirMsg.value = 'Carpeta actualizada sin reinicio.';
        dirMsgOk.value = true;
        dirBusy.value = false;
        await loadDirectorySettings();
    } catch {
        dirBusy.value = false;
        dirMsg.value = 'No se pudo cambiar la carpeta. Revisá permisos y que la ruta exista.';
        dirMsgOk.value = false;
    }
}

function onDirModeChange() {
    if (dirMode.value === 'minecraft') separateGameDir.value = false;
}

async function saveSeparateGameDir() {
    try {
        await SetSeparateGameDir(separateGameDir.value);
    } catch (_e) {}
}

interface CacheInfo {
    totalEntries: number;
    totalBytes?: number;
    categories: Record<string, number>;
    sizes?: Record<string, number>;
}
const cacheInfo = ref<CacheInfo | null>(null);

const cacheTotal = computed(() => cacheInfo.value?.totalEntries ?? 0);

function fmtBytes(bytes: number | undefined): string {
    if (typeof bytes !== 'number' || bytes <= 0) return '0 MB';
    const mb = bytes / (1024 * 1024);
    if (mb >= 1024) return `${(mb / 1024).toFixed(2)} GB`;
    return `${mb.toFixed(2)} MB`;
}

const cacheSizeText = computed(() => fmtBytes(cacheInfo.value?.totalBytes));

const cacheDetail = computed(() => {
    const info = cacheInfo.value;
    if (!info) return 'Cargando...';
    const parts = Object.entries(info.categories ?? {})
        .filter(([, n]) => n > 0)
        .map(([k, n]) => `${k}: ${n}`);
    return parts.length ? parts.join(' · ') : 'Vacía';
});

async function refreshCache() {
    try {
        const info = await GetCacheInfo();
        if (info && typeof info.totalEntries === 'number') {
            cacheInfo.value = info as unknown as CacheInfo;
        }
    } catch (_e) {}
}

async function clearCache() {
    try {
        await ClearAllCache();
    } catch (_e) {}
    await refreshCache();
}

async function refreshManifests() {
    try {
        await RefreshManifests();
    } catch (_e) {}
    await refreshCache();
}

onMounted(() => {
    loadDirectorySettings();
});

onActivated(async () => {
    await refreshCache();
    await loadDirectorySettings();
});
</script>

<template>
    <div class="Ss">

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
                <span>Dónde se guarda todo</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Carpeta del launcher</span>
                    <span class="SsDesc">Acá viven tus instancias, versiones, cuentas y mundos. Si la cambiás, solo se migra tu configuración: respaldá tus instancias antes.</span>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Modo de carpeta</span>
                    <span class="SsDesc">Normal (recomendado): todo dentro del launcher. Minecraft: usa tu .minecraft oficial. Portable: todo junto al ejecutable, ideal para USB. Personalizada: la que vos elijas.</span>
                </div>
                <div class="SsCtrl">
                    <select class="SsSel" v-model="dirMode" @change="onDirModeChange">
                        <option v-for="m in dirModes" :key="m.id" :value="m.id">{{ m.label }}</option>
                    </select>
                </div>
            </div>
            <template v-if="dirMode === 'custom'">
                <div class="SsRow">
                    <div class="SsInfo">
                    <span class="SsLabel">Carpeta personalizada</span>
                    <span class="SsDesc">Elegí la carpeta raíz donde vive todo el launcher.</span>
                    </div>
                    <div class="SsCtrl SsDirPick">
                        <input class="SsIn" v-model="customPath" placeholder="C:\MiLauncher" />
                        <button class="SsBtn" :disabled="dirBusy" @click="pickDirectory">Examinar</button>
                    </div>
                </div>
            </template>
            <template v-if="dirInfo?.minecraftExists && dirMode !== 'minecraft'">
                <div class="SsRow">
                    <div class="SsInfo">
                    <span class="SsLabel">¿Usar tu .minecraft existente?</span>
                    <span class="SsDesc">Detectamos {{ dirInfo.minecraftDir }} con contenido. Podés adoptarla como carpeta del launcher.</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn SsBtnPrimary" @click="dirMode = 'minecraft'; separateGameDir = false">Usar esa carpeta</button>
                    </div>
                </div>
            </template>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Carpeta actual</span>
                    <span class="SsDesc">{{ dirInfo?.workDir || 'Cargando...' }}</span>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Carpeta «game» separada</span>
                    <span class="SsDesc">Guarda instancias y mundos en una subcarpeta «game» aparte del resto. Recomendado: orden y backups más simples.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="separateGameDir" :disabled="dirMode === 'minecraft'" title="No disponible con Minecraft" @change="saveSeparateGameDir"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Guardar cambio de carpeta</span>
                    <span class="SsDesc">Aplica el cambio y reinicia solo el launcher (el juego no se toca).</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn SsBtnPrimary" :disabled="dirBusy || !dirChanged" @click="saveDirectory">
                        {{ dirBusy ? 'Guardando...' : 'Guardar y reiniciar' }}
                    </button>
                </div>
            </div>
            <p v-if="dirMsg" :class="['SsDirMsg', { error: !dirMsgOk }]">{{ dirMsg }}</p>
        </div>

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
                <span>Archivos temporales y caché</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Contenido de la caché</span>
                    <span class="SsDesc">{{ cacheDetail }}</span>
                </div>
                <span class="SsValue">{{ cacheTotal }} archivos · {{ cacheSizeText }}</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Liberar espacio</span>
                    <span class="SsDesc">Borra temporales (miniaturas, índices viejos). Tus instancias y mundos están a salvo.</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn" @click="refreshManifests">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                        Recalcular
                    </button>
                    <button class="SsBtn SsBtnDanger" @click="clearCache">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                        Borrar caché
                    </button>
                </div>
            </div>
        </div>

    </div>
</template>

<style scoped lang="scss">
@use '../Styles/General.scss';
</style>
