<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { setUIScale, applyPersonalization } from '../../../stores/ui';

const concurrentDownloads = ref(4);
const maxMbps = ref(0);
const zoom = ref(100);

const animations = ref(true);
const blur = ref(true);
const shadows = ref(true);
let fullPersonalization: any = null;

const authVerify = ref(true);
const proxyEnabled = ref(false);
const proxyHost = ref('');
const proxyPort = ref(8080);
const proxyUser = ref('');
const proxyPass = ref('');

interface CacheInfo {
    totalEntries: number;
    categories: Record<string, number>;
}
const cacheInfo = ref<CacheInfo | null>(null);

onMounted(async () => {
    try {
        const cfg = await (window as any).go?.main?.App?.GetConfig?.();
        if (cfg) {
            concurrentDownloads.value = cfg.launcher?.concurrentDownloads ?? 4;
            maxMbps.value = cfg.launcher?.maxMbps ?? 0;
            const raw = cfg.personalization?.uiScale;
            zoom.value = (typeof raw === 'number' && raw >= 50 && raw <= 200) ? Math.round(raw) : 100;
            setUIScale(zoom.value);
            fullPersonalization = cfg.personalization ?? null;
            animations.value = cfg.personalization?.animations ?? true;
            blur.value = cfg.personalization?.blur ?? true;
            shadows.value = cfg.personalization?.shadows ?? true;
            const mc = cfg.minecraftConfig ?? {};
            authVerify.value = mc.authVerify ?? true;
            proxyEnabled.value = mc.proxyEnabled ?? false;
            proxyHost.value = mc.proxyHost ?? '';
            proxyPort.value = mc.proxyPort ?? 8080;
            proxyUser.value = mc.proxyUser ?? '';
            proxyPass.value = mc.proxyPass ?? '';
        }
    } catch { /* */ }

    await refreshCache();
});

async function refreshCache() {
    try {
        const info = await (window as any).go?.main?.App?.GetCacheInfo?.();
        if (info && typeof info.totalEntries === 'number') {
            cacheInfo.value = info;
        }
    } catch { /* */ }
}

async function clearCache() {
    try {
        await (window as any).go?.main?.App?.ClearAllCache?.();
    } catch { /* */ }
    await refreshCache();
}

async function saveDownloads() {
    try {
        await (window as any).go?.main?.App?.SetConcurrentDownloads?.(concurrentDownloads.value);
    } catch { /* */ }
}

async function saveAuthVerify() {
    try {
        await (window as any).go?.main?.App?.SetAuthVerify?.(authVerify.value);
    } catch { /* */ }
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
    } catch { /* */ }
}

async function saveMbps() {
    try {
        await (window as any).go?.main?.App?.SetMaxMbps?.(maxMbps.value);
    } catch { /* */ }
}

async function saveZoom() {
    setUIScale(zoom.value);
    try {
        await (window as any).go?.main?.App?.SetUIScale?.(zoom.value);
    } catch { /* */ }
}

async function saveRendimiento() {
    const p = { ...(fullPersonalization ?? {}), animations: animations.value, blur: blur.value, shadows: shadows.value };
    fullPersonalization = p;
    applyPersonalization(p as any);
    try {
        await (window as any).go?.main?.App?.UpdatePersonalization?.(p);
    } catch { /* */ }
}

const showResetConfirm = ref(false);

async function resetConfig() {
    showResetConfirm.value = false;
    try {
        await (window as any).go?.main?.App?.ResetConfig?.();
    } catch { /* */ }
}

async function checkUpdates() {
    try {
        await (window as any).go?.main?.App?.CheckForUpdates?.();
    } catch { /* */ }
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
                        <button class="SsStepBtn" :disabled="zoom <= 50" @click="zoom = Math.max(50, zoom - 10); saveZoom()">−</button>
                        <span class="SsStepVal">{{ zoom }}%</span>
                        <button class="SsStepBtn" :disabled="zoom >= 200" @click="zoom = Math.min(200, zoom + 10); saveZoom()">+</button>
                    </div>
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
                    <span class="SsDesc">Manifiestos de NovaCore y fondos importados almacenados localmente.</span>
                </div>
                <span class="SsValue">{{ cacheInfo ? cacheInfo.totalEntries : 0 }} archivos</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Limpiar caché</span>
                    <span class="SsDesc">Elimina los manifiestos del motor y los archivos temporales descargados.</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn SsBtnDanger" @click="clearCache">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                        Limpiar
                    </button>
                </div>
            </div>
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
                    <span class="SsLabel">Versión del launcher</span>
                    <span class="SsDesc">Comprueba si hay una versión nueva disponible.</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn SsBtnPrimary" @click="checkUpdates">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                        Buscar Actualizaciones
                    </button>
                </div>
            </div>
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
@use '../../../Styles/Settings.scss';

.ConfirmOverlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 100;
}

.ConfirmDialog {
    width: 22rem;
    max-width: 90vw;
    background: var(--background-modal-primray);
    border: var(--border-modal-style);
    border-radius: 0.75rem;
    box-shadow: var(--shadow-settings-normal) #0008;
    padding: 1.25rem 1.35rem;

    h3 {
        margin: 0;
        font-family: var(--font-primary), Arial, sans-serif;
        font-size: 0.95rem;
        font-weight: 600;
    }

    p {
        margin: 0.6rem 0 0;
        font-size: 0.78rem;
        line-height: 1.5;
        opacity: 0.55;
    }
}

.ConfirmActions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    margin-top: 1.1rem;
}
</style>
