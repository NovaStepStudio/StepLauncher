<script setup lang="ts">
import { ref, computed, onMounted, onActivated, onUnmounted } from 'vue';
import { setUIScale, applyPersonalization, uiScale, personalization } from '@/Common/Stores/Ui';
import { saveIdleOptions, CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { checkForUpdates as requestUpdateCheck, checking as updateChecking } from '@/Updates/Store';

const concurrentDownloads = ref(4);
const maxMbps = ref(0);
const hideLauncher = ref(true);
const verifyIntegrity = ref(true);
const richPresence = ref(true);
const launchAfterInstall = ref(false);

interface IntegrityStatus {
    state: string;
    phase: string;
    scope: string;
    percent: number;
    tasksTotal: number;
    tasksDone: number;
    filesMissing: number;
    filesRestored: number;
    filesCorrupt: number;
    filesSkipped: number;
    versionsScanned: number;
}

const integrityScope = ref('todo');
const integrityStatus = ref<IntegrityStatus | null>(null);
const integrityBusy = ref(false);
let integrityTimer: number | null = null;

const integrityPercent = computed(() => integrityStatus.value?.percent ?? 0);

const integrityPhaseLabel = computed(() => {
    switch (integrityStatus.value?.phase) {
        case 'indexing': return 'Recorriendo JSON de versiones';
        case 'existence': return 'Comprobando archivos y descargando faltantes';
        case 'retry': return 'Reintentando descargas pendientes';
        case 'verify': return 'Verificando SHA1 y tamaño';
        case 'done': return 'Terminado';
        default: return '';
    }
});

const integrityDoneText = computed(() => {
    const st = integrityStatus.value;
    if (!st || st.state === 'running') return '';
    if (st.state === 'completed') {
        return `Completado: ${st.versionsScanned} versiones, ${st.filesRestored} archivos restaurados, ${st.filesSkipped} descartados`;
    }
    if (st.state === 'cancelled') return 'Verificación cancelada';
    if (st.state === 'error') return 'Verificación terminó con error (ver logs del backend)';
    return '';
});

function stopIntegrityPolling() {
    if (integrityTimer !== null) {
        window.clearInterval(integrityTimer);
        integrityTimer = null;
    }
}

async function pollIntegrity() {
    try {
        const st = await (window as any).go?.main?.App?.IntegrityStatus?.();
        if (!st) return;
        integrityStatus.value = st;
        if (st.state !== 'running') {
            stopIntegrityPolling();
            integrityBusy.value = false;
            if (st.state === 'completed' || st.state === 'cancelled' || st.state === 'error') {
                window.setTimeout(() => {
                    if (integrityStatus.value && integrityStatus.value.state === st.state) {
                        integrityStatus.value = null;
                    }
                }, 6000);
            }
        }
    } catch { }
}

async function syncIntegrityFromBackend() {
    try {
        const st = await (window as any).go?.main?.App?.IntegrityStatus?.();
        if (!st) return;
        integrityStatus.value = st;
        if (st.state === 'running') {
            integrityBusy.value = true;
            if (integrityTimer === null) {
                integrityTimer = window.setInterval(pollIntegrity, 500);
            }
        }
    } catch { }
}

async function setIntegrityScope(scope: string) {
    integrityScope.value = scope;
    try {
        await (window as any).go?.main?.App?.SetIntegritySector?.(scope);
    } catch { }
}

async function startIntegrityCheck() {
    if (integrityBusy.value) return;
    integrityBusy.value = true;
    integrityStatus.value = null;
    try {
        await (window as any).go?.main?.App?.SetIntegritySector?.(integrityScope.value);
        await (window as any).go?.main?.App?.StartIntegrityCheck?.(integrityScope.value);
    } catch { }
    stopIntegrityPolling();
    integrityTimer = window.setInterval(pollIntegrity, 500);
    await pollIntegrity();
}

const zoom = computed(() => uiScale.value);

const animations = ref(true);
const blur = ref(true);
const shadows = ref(true);
const textShadow = ref(false);
const textShadowIntensity = ref(1);

const autoCloseModals = ref(true);
const idleMinutes = ref(1);
const configCheckEnabled = ref(true);
const configCheckMinutes = ref(3);

const authVerify = ref(true);
const proxyEnabled = ref(false);
const proxyHost = ref('');
const proxyPort = ref(8080);
const proxyUser = ref('');
const proxyPass = ref('');

const checkOnStart = ref(false);

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
    { id: 'normal', label: 'Normal' },
    { id: 'minecraft', label: 'Minecraft' },
    { id: 'portable', label: 'Portable' },
    { id: 'custom', label: 'Personalizada' },
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
        const info = await (window as any).go?.main?.App?.GetDirectorySettings?.();
        if (info) {
            dirInfo.value = info;
            dirMode.value = info.mode ?? 'normal';
            customPath.value = info.customPath ?? '';
            if (info.mode === 'minecraft') separateGameDir.value = false;
        }
        const sep = await (window as any).go?.main?.App?.GetSeparateGameDir?.();
        if (typeof sep === 'boolean') separateGameDir.value = sep;
    } catch { }
}

async function pickDirectory() {
    try {
        const p = await (window as any).go?.main?.App?.PickDirectory?.();
        if (p) customPath.value = p;
    } catch { }
}

async function saveDirectory() {
    if (dirBusy.value) return;
    dirBusy.value = true;
    dirMsg.value = '';
    dirMsgOk.value = true;
    try {
        const err = await (window as any).go?.main?.App?.SetDirectoryMode?.(
            dirMode.value,
            dirMode.value === 'custom' ? customPath.value.trim() : ''
        );
        if (err) {
            dirMsg.value = typeof err === 'string' ? err : 'No se pudo cambiar la carpeta';
            dirMsgOk.value = false;
            dirBusy.value = false;
            return;
        }
        if (dirChanged.value) {
            dirMsg.value = 'Directorio guardado. Reiniciando el launcher...';
            dirMsgOk.value = true;
            try {
                await (window as any).go?.main?.App?.RestartApp?.();
            } catch { }
            return;
        }
        dirMsg.value = 'Directorio actualizado.';
        dirMsgOk.value = true;
        dirBusy.value = false;
        await loadDirectorySettings();
    } catch {
        dirBusy.value = false;
        dirMsg.value = 'No se pudo cambiar la carpeta';
        dirMsgOk.value = false;
    }
}

function onDirModeChange() {
    if (dirMode.value === 'minecraft') separateGameDir.value = false;
}

async function saveSeparateGameDir() {
    try {
        await (window as any).go?.main?.App?.SetSeparateGameDir?.(separateGameDir.value);
    } catch { }
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
    return parts.length ? parts.join(' · ') : 'Caché vacía';
});

async function loadConfig() {
    try {
        const cfg = await (window as any).go?.main?.App?.GetConfig?.();
        if (cfg) {
            concurrentDownloads.value = cfg.launcher?.concurrentDownloads ?? 4;
            maxMbps.value = cfg.launcher?.maxMbps ?? 0;
            hideLauncher.value = cfg.launcher?.hideLauncherOnLaunch ?? true;
            verifyIntegrity.value = cfg.launcher?.verifyIntegrity ?? true;
            richPresence.value = cfg.richPresence?.enabled ?? true;
            animations.value = cfg.personalization?.animations ?? true;
            blur.value = cfg.personalization?.blur ?? true;
            shadows.value = cfg.personalization?.shadows ?? true;
            textShadow.value = cfg.personalization?.textShadow ?? false;
            textShadowIntensity.value = cfg.personalization?.textShadowIntensity ?? 1;
            const idle = cfg.idle ?? {};
            autoCloseModals.value = idle.autoCloseModals ?? true;
            idleMinutes.value = idle.idleMinutes ?? 1;
            configCheckEnabled.value = idle.configCheckEnabled ?? true;
            configCheckMinutes.value = idle.configCheckMinutes ?? 3;
            const mc = cfg.minecraftConfig ?? {};
            authVerify.value = mc.authVerify ?? true;
            proxyEnabled.value = mc.proxyEnabled ?? false;
            proxyHost.value = mc.proxyHost ?? '';
            proxyPort.value = mc.proxyPort ?? 8080;
            proxyUser.value = mc.proxyUser ?? '';
            proxyPass.value = mc.proxyPass ?? '';
            checkOnStart.value = cfg.launcher?.checkForUpdatesOnStart ?? false;
            launchAfterInstall.value = cfg.launcher?.launchAfterInstall ?? false;
            integrityScope.value = cfg.launcher?.integritySector ?? 'todo';
        }
    } catch { }
}

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    loadConfig();
});

onActivated(async () => {
    await refreshCache();
    await syncIntegrityFromBackend();
    await loadDirectorySettings();
});

async function refreshCache() {
    try {
        const info = await (window as any).go?.main?.App?.GetCacheInfo?.();
        if (info && typeof info.totalEntries === 'number') {
            cacheInfo.value = info;
        }
    } catch { }
}

async function clearCache() {
    try {
        await (window as any).go?.main?.App?.ClearAllCache?.();
    } catch { }
    await refreshCache();
}

async function refreshManifests() {
    try {
        await (window as any).go?.main?.App?.RefreshManifests?.();
    } catch { }
    await refreshCache();
}

async function saveDownloads() {
    try {
        await (window as any).go?.main?.App?.SetConcurrentDownloads?.(concurrentDownloads.value);
    } catch { }
}

async function saveAuthVerify() {
    try {
        await (window as any).go?.main?.App?.SetAuthVerify?.(authVerify.value);
    } catch { }
}

async function saveProxy() {
    try {
        await (window as any).go?.main?.App?.SetProxy?.(
            proxyEnabled.value,
            proxyHost.value,
            proxyPort.value,
            proxyUser.value,
            proxyPass.value
        );
    } catch { }
}

async function saveMbps() {
    try {
        await (window as any).go?.main?.App?.SetMaxMbps?.(maxMbps.value);
    } catch { }
}

async function saveHideLauncher() {
    try {
        await (window as any).go?.main?.App?.SetHideLauncher?.(hideLauncher.value);
    } catch { }
}

async function saveVerifyIntegrity() {
    try {
        await (window as any).go?.main?.App?.SetVerifyIntegrity?.(verifyIntegrity.value);
    } catch { }
}

async function saveRichPresence() {
    try {
        await (window as any).go?.main?.App?.SetRichPresenceEnabled?.(richPresence.value);
    } catch { }
}

async function saveZoom() {
    try {
        await (window as any).go?.main?.App?.SetUIScale?.(uiScale.value);
    } catch { }
}

function stepZoom(delta: number) {
    const next = Math.min(200, Math.max(50, uiScale.value + delta));
    setUIScale(next);
    saveZoom();
}

async function saveRendimiento() {
    const p = {
        ...(personalization.value ?? {}),
        uiScale: uiScale.value,
        animations: animations.value,
        blur: blur.value,
        shadows: shadows.value,
        textShadow: textShadow.value,
        textShadowIntensity: textShadowIntensity.value,
    };
    applyPersonalization(p as any);
    try {
        await (window as any).go?.main?.App?.UpdatePersonalization?.(p);
    } catch { }
}

function stepTextShadowIntensity(delta: number) {
    textShadowIntensity.value = Math.round(Math.min(2, Math.max(0.5, textShadowIntensity.value + delta)) * 100) / 100;
    saveRendimiento();
}

async function saveIdle() {
    await saveIdleOptions({
        autoCloseModals: autoCloseModals.value,
        idleMinutes: idleMinutes.value,
        configCheckEnabled: configCheckEnabled.value,
        configCheckMinutes: configCheckMinutes.value,
    });
}

const showResetConfirm = ref(false);

function onCloseOverlays() {
    showResetConfirm.value = false;
}

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    loadConfig();
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    stopIntegrityPolling();
});

async function resetConfig() {
    showResetConfirm.value = false;
    try {
        await (window as any).go?.main?.App?.ResetConfig?.();
        window.location.reload();
    } catch { }
    await loadConfig();
}

async function checkUpdates() {
    try {
        await requestUpdateCheck(false);
    } catch { }
}

async function saveCheckOnStart() {
    try {
        await (window as any).go?.main?.App?.SetCheckForUpdatesOnStart?.(checkOnStart.value);
    } catch { }
}

async function saveLaunchAfterInstall() {
    try {
        await (window as any).go?.main?.App?.SetLaunchAfterInstall?.(launchAfterInstall.value);
    } catch { }
}
</script>

<template>
    <div class="Ss">

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><line x1="11" y1="8" x2="11" y2="14"/><line x1="8" y1="11" x2="14" y2="11"/></svg>
                <span>Interfaz</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Zoom de la interfaz</span>
                    <span class="SsDesc">Escala el tamaño de todo el launcher, del 50% al 200%.</span>
                </div>
                <div class="SsCtrl">
                    <div class="SsStep">
                        <button class="SsStepBtn" :disabled="zoom <= 50" @click="stepZoom(-10)">−</button>
                        <span class="SsStepVal">{{ zoom }}%</span>
                        <button class="SsStepBtn" :disabled="zoom >= 200" @click="stepZoom(10)">+</button>
                    </div>
                </div>
            </div>
        </div>

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="21" x2="9" y2="9"/></svg>
                <span>Comportamiento</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Ocultar launcher al abrir Minecraft</span>
                    <span class="SsDesc">Al pulsar Jugar, la ventana del launcher se oculta mientras el juego está abierto y reaparece al cerrarlo o si el juego se cierra.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="hideLauncher" @change="saveHideLauncher"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Verificación de integridad de archivos</span>
                    <span class="SsDesc">Comprueba el SHA1 de cada archivo descargado por el motor (re-descarga los corruptos). Al desactivarlo las descargas son más rápidas pero pueden quedar archivos dañados.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="verifyIntegrity" @change="saveVerifyIntegrity"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Presencia en Discord</span>
                    <span class="SsDesc">Muestra tu actividad en tu perfil de Discord (Rich Presence): navegando por el launcher o jugando a Minecraft. Requiere tener Discord abierto.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="richPresence" @change="saveRichPresence"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Lanzar Minecraft al terminar una instalación</span>
                    <span class="SsDesc">Cuando una descarga de versión (con o sin ModLoader) termine, el launcher la abre automáticamente y lanza el juego sin tocar nada más.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="launchAfterInstall" @change="saveLaunchAfterInstall"><span class="SsTgS"></span></label>
                </div>
            </div>
        </div>

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
                <span>Internet</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Descargas Concurrentes</span>
                    <span class="SsDesc">Número de archivos que se descargan al mismo tiempo.</span>
                </div>
                <div class="SsCtrl">
                    <div class="SsStep">
                        <button class="SsStepBtn" :disabled="concurrentDownloads <= 1" @click="concurrentDownloads--; saveDownloads()">−</button>
                        <span class="SsStepVal">{{ concurrentDownloads }}</span>
                        <button class="SsStepBtn" :disabled="concurrentDownloads >= 16" @click="concurrentDownloads++; saveDownloads()">+</button>
                    </div>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Máximo de Mbps</span>
                    <span class="SsDesc">Límite de velocidad de descarga. 0 = Ilimitado.</span>
                </div>
                <div class="SsCtrl">
                    <div class="SsStep">
                        <button class="SsStepBtn" :disabled="maxMbps <= 0" @click="maxMbps = Math.max(0, maxMbps - 5); saveMbps()">−</button>
                        <span class="SsStepVal">{{ maxMbps === 0 ? 'Ilimitado' : maxMbps + ' Mbps' }}</span>
                        <button class="SsStepBtn" :disabled="maxMbps >= 500" @click="maxMbps += 5; saveMbps()">+</button>
                    </div>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Verificar servidor</span>
                    <span class="SsDesc">Valida la conexión con el servidor de autenticación al iniciar sesión.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="authVerify" @change="saveAuthVerify"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Proxy de Minecraft</span>
                    <span class="SsDesc">Enruta el tráfico de Minecraft (descargas del juego, skins y conexión a servidores) a través de un servidor proxy. Solo afecta al juego, no al launcher.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="proxyEnabled" @change="saveProxy"><span class="SsTgS"></span></label>
                </div>
            </div>
            <template v-if="proxyEnabled">
                <div class="SsRow">
                    <div class="SsInfo">
                        <span class="SsLabel">Host y puerto</span>
                        <span class="SsDesc">Dirección y puerto del proxy que usará Minecraft.</span>
                    </div>
                    <div class="SsGrid">
                        <input class="SsIn" v-model="proxyHost" placeholder="Host" @change="saveProxy">
                        <input class="SsIn" type="number" v-model.number="proxyPort" placeholder="Puerto" @change="saveProxy">
                    </div>
                </div>
                <div class="SsRow">
                    <div class="SsInfo">
                        <span class="SsLabel">Credenciales</span>
                        <span class="SsDesc">Usuario y contraseña del proxy (opcional).</span>
                    </div>
                    <div class="SsGrid">
                        <input class="SsIn" v-model="proxyUser" placeholder="Usuario" @change="saveProxy">
                        <input class="SsIn" type="password" v-model="proxyPass" placeholder="Contraseña" @change="saveProxy">
                    </div>
                </div>
            </template>
        </div>

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
                <span>Cache</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Elementos en caché</span>
                    <span class="SsDesc">Manifiestos del motor y fondos importados almacenados localmente.</span>
                    <span class="SsCacheDetail">{{ cacheDetail }}</span>
                </div>
                <span class="SsValue">{{ cacheTotal }} archivos · {{ cacheSizeText }}</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Limpiar caché</span>
                    <span class="SsDesc">Elimina los manifiestos del motor y los archivos temporales descargados.</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn" @click="refreshManifests">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                        Refrescar manifiestos
                    </button>
                    <button class="SsBtn SsBtnDanger" @click="clearCache">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                        Limpiar
                    </button>
                </div>
            </div>
        </div>

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
                <span>Directorio del launcher</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Ubicación de la carpeta de trabajo</span>
                    <span class="SsDesc">Dónde se guardan versiones, instancias, cuentas y configuración. Al cambiar no se copian los datos: solo se copia la configuración si la nueva carpeta no tiene una propia.</span>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Modo de directorio</span>
                    <span class="SsDesc">Normal usa la carpeta predeterminada .StepLauncher. Minecraft usa la carpeta oficial .minecraft sin sobrescribir sus archivos. Portable crea .StepLauncher junto al ejecutable. Personalizada usa la ruta que elijas.</span>
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
                        <span class="SsLabel">Ruta personalizada</span>
                        <span class="SsDesc">Carpeta donde se guardará todo el launcher.</span>
                    </div>
                    <div class="SsCtrl SsDirPick">
                        <input class="SsIn" v-model="customPath" placeholder="C:\Users\TuUsuario\AppData\Roaming\.StepLauncher" />
                        <button class="SsBtn" :disabled="dirBusy" @click="pickDirectory">Examinar</button>
                    </div>
                </div>
            </template>
            <template v-if="dirInfo?.minecraftExists && dirMode !== 'minecraft'">
                <div class="SsRow">
                    <div class="SsInfo">
                        <span class="SsLabel">Instalación de Minecraft detectada</span>
                        <span class="SsDesc">Se encontró la carpeta {{ dirInfo.minecraftDir }}. ¿Quieres usarla como carpeta de trabajo?</span>
                    </div>
                    <div class="SsCtrl">
                        <button class="SsBtn SsBtnPrimary" @click="dirMode = 'minecraft'; separateGameDir = false">Usar .minecraft</button>
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
                    <span class="SsLabel">Carpeta de juego separada</span>
                    <span class="SsDesc">Guarda lo que genera Minecraft (saves, mods, opciones) en una subcarpeta "game". Al desactivarlo se usa la carpeta de trabajo directamente. En modo Minecraft siempre está desactivado.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="separateGameDir" :disabled="dirMode === 'minecraft'" title="No disponible en modo Minecraft" @change="saveSeparateGameDir"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Aplicar cambio</span>
                    <span class="SsDesc">El launcher se reinicia automáticamente para usar el nuevo directorio.</span>
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
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
                <span>Rendimiento</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Animaciones</span>
                    <span class="SsDesc">Transiciones y efectos animados de la interfaz.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="animations" @change="saveRendimiento"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Filtro Blur</span>
                    <span class="SsDesc">Desenfoque de fondo en paneles y modales.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="blur" @change="saveRendimiento"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Sombras</span>
                    <span class="SsDesc">Sombras proyectadas por ventanas y elementos.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="shadows" @change="saveRendimiento"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Sombra de texto</span>
                    <span class="SsDesc">Brillo tipo neón en las letras, usando el color de la letra.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="textShadow" @change="saveRendimiento"><span class="SsTgS"></span></label>
                </div>
            </div>
            <template v-if="textShadow">
                <div class="SsRow">
                    <div class="SsInfo">
                        <span class="SsLabel">Intensidad del brillo</span>
                        <span class="SsDesc">Qué tan marcado es el resplandor neón de las letras.</span>
                    </div>
                    <div class="SsCtrl">
                        <div class="SsStep">
                            <button class="SsStepBtn" :disabled="textShadowIntensity <= 0.5" @click="stepTextShadowIntensity(-0.25)">−</button>
                            <span class="SsStepVal">×{{ textShadowIntensity.toFixed(2) }}</span>
                            <button class="SsStepBtn" :disabled="textShadowIntensity >= 2" @click="stepTextShadowIntensity(0.25)">+</button>
                        </div>
                    </div>
                </div>
            </template>
        </div>

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
                <span>Reposo y mantenimiento</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Cerrar modales en reposo</span>
                    <span class="SsDesc">Si no interactúas con la app durante un tiempo, cierra automáticamente los modales abiertos (tipografías, configuración, colores) y vuelve al menú principal.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="autoCloseModals" @change="saveIdle"><span class="SsTgS"></span></label>
                </div>
            </div>
            <template v-if="autoCloseModals">
                <div class="SsRow">
                    <div class="SsInfo">
                        <span class="SsLabel">Minutos de inactividad</span>
                        <span class="SsDesc">Tiempo sin interacción antes de cerrar los modales.</span>
                    </div>
                    <div class="SsCtrl">
                        <div class="SsStep">
                            <button class="SsStepBtn" :disabled="idleMinutes <= 1" @click="idleMinutes = Math.max(1, idleMinutes - 1); saveIdle()">−</button>
                            <span class="SsStepVal">{{ idleMinutes }} min</span>
                            <button class="SsStepBtn" :disabled="idleMinutes >= 10" @click="idleMinutes = Math.min(10, idleMinutes + 1); saveIdle()">+</button>
                        </div>
                    </div>
                </div>
            </template>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Verificar configuración</span>
                    <span class="SsDesc">Revisa periódicamente que la configuración guardada (colores, tipografías, sombra de texto...) siga aplicada y la reaplica si hace falta.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="configCheckEnabled" @change="saveIdle"><span class="SsTgS"></span></label>
                </div>
            </div>
            <template v-if="configCheckEnabled">
                <div class="SsRow">
                    <div class="SsInfo">
                        <span class="SsLabel">Intervalo de verificación</span>
                        <span class="SsDesc">Cada cuánto tiempo se comprueba la configuración.</span>
                    </div>
                    <div class="SsCtrl">
                        <div class="SsStep">
                            <button class="SsStepBtn" :disabled="configCheckMinutes <= 1" @click="configCheckMinutes = Math.max(1, configCheckMinutes - 1); saveIdle()">−</button>
                            <span class="SsStepVal">{{ configCheckMinutes }} min</span>
                            <button class="SsStepBtn" :disabled="configCheckMinutes >= 10" @click="configCheckMinutes = Math.min(10, configCheckMinutes + 1); saveIdle()">+</button>
                        </div>
                    </div>
                </div>
            </template>
        </div>

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
                <span>Configuración</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Restablecer configuración</span>
                    <span class="SsDesc">Vuelve a los valores por defecto de todos los ajustes del launcher.</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn SsBtnDanger" @click="showResetConfirm = true">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                        Restablecer
                    </button>
                </div>
            </div>
        </div>

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                <span>Actualizaciones</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Buscar actualizaciones al iniciar</span>
                    <span class="SsDesc">Al abrir el launcher comprueba en segundo plano si hay una versión nueva de StepLauncher. La actualización nunca es obligatoria.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="checkOnStart" @change="saveCheckOnStart"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Versión del launcher</span>
                    <span class="SsDesc">Comprueba si hay una versión nueva disponible.</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn SsBtnPrimary" :disabled="updateChecking" @click="checkUpdates">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                        {{ updateChecking ? 'Buscando…' : 'Buscar Actualizaciones' }}
                    </button>
                </div>
            </div>
        </div>

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><polyline points="9 11 11 13 15 9"/></svg>
                <span>Integridad</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Sector a verificar</span>
                    <span class="SsDesc">Alcance del proceso: todo el launcher, solo las versiones globales o solo las instancias.</span>
                </div>
                <div class="SsCtrl">
                    <div class="SsSeg">
                        <button :class="{ active: integrityScope === 'todo' }" :disabled="integrityBusy" @click="setIntegrityScope('todo')">Todo</button>
                        <button :class="{ active: integrityScope === 'global' }" :disabled="integrityBusy" @click="setIntegrityScope('global')">Global</button>
                        <button :class="{ active: integrityScope === 'instances' }" :disabled="integrityBusy" @click="setIntegrityScope('instances')">Instancias</button>
                    </div>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Verificar integridad de archivos</span>
                    <span class="SsDesc">Recorre todos los JSON de las versiones descargadas, descarga los archivos que faltan, reintenta los que fallan y verifica el SHA1 y el tamaño de cada uno. Proceso lento; los archivos no recuperables quedan registrados en el backend.</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn SsBtnPrimary" :disabled="integrityBusy" @click="startIntegrityCheck">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
                        {{ integrityBusy ? `Verificando… ${integrityPercent}%` : (integrityStatus && integrityStatus.state === 'completed' ? 'Completado' : 'Verificar integridad') }}
                    </button>
                </div>
            </div>
            <template v-if="integrityBusy">
                <div class="SsRow">
                    <div class="SsInfo">
                        <span class="SsDesc">Fase: {{ integrityPhaseLabel }}</span>
                    </div>
                    <div class="SsCtrl">
                        <div class="SsIntegrityBar">
                            <div class="SsIntegrityBarFill" :style="{ width: integrityPercent + '%' }"></div>
                        </div>
                    </div>
                </div>
            </template>
            <template v-else-if="integrityDoneText">
                <div class="SsRow">
                    <div class="SsInfo">
                        <span class="SsDesc">{{ integrityDoneText }}</span>
                    </div>
                </div>
            </template>
        </div>

        <Teleport to="body">
            <div v-if="showResetConfirm" class="ConfirmOverlay" @click.self="showResetConfirm = false">
                <div class="ConfirmDialog">
                    <h3>Restablecer configuración</h3>
                    <p>¿Seguro que quieres volver todos los ajustes del launcher a los valores por defecto? Esta acción no se puede deshacer.</p>
                    <div class="ConfirmActions">
                        <button class="SsBtn" @click="showResetConfirm = false">Cancelar</button>
                        <button class="SsBtn SsBtnDanger" @click="resetConfig">
                            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                            Restablecer
                        </button>
                    </div>
                </div>
            </div>
        </Teleport>

    </div>
</template>

<style scoped lang="scss">
@use '../Styles/General.scss';
</style>
