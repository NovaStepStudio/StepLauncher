<script setup lang="ts">
import { onMounted, onUnmounted, computed } from 'vue';
import {
    IconArrowUpCircle,
    IconCheck,
    IconX,
    IconAlertTriangle,
    IconExternalLink,
    IconDownload,
    IconRefresh,
    IconRocket,
} from '@tabler/icons-vue';
import {
    updateInfo, checking, modalVisible, installUpdate, closeUpdateModal,
    wailsRelease, wailsState, wailsProgress, wailsReady, wailsDownloading, wailsError,
    wailsCurrentVersion,
    restartUpdate,
} from './Store';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

const isWindows = computed(() => (updateInfo.value?.platform ?? '') === 'windows');

const wailsAvailable = computed(() => wailsState.value === 'available' && !!wailsRelease.value);
const wailsProgressPct = computed(() => {
    const p = wailsProgress.value;
    if (!p || !p.total) return 0;
    return Math.min(100, Math.round((p.written / p.total) * 100));
});
const wailsProgressText = computed(() => {
    const p = wailsProgress.value;
    if (!p) return '';
    const fmt = (n: number) => (n / (1024 * 1024)).toFixed(2) + ' MB';
    if (p.total) return `${fmt(p.written)} / ${fmt(p.total)} · ${wailsProgressPct.value}%`;
    return `${fmt(p.written)} descargados`;
});

const currentDisplay = computed(() => wailsCurrentVersion.value || updateInfo.value?.currentVersion || '');

const subtitle = computed(() => {
    if (checking.value) return 'Consultando la última release de GitHub…';
    if (wailsState.value === 'downloading' && wailsDownloading.value) return `Descargando actualización… ${wailsProgressPct.value}%`;
    if (wailsState.value === 'verifying') return 'Verificando la descarga…';
    if (wailsState.value === 'installing') return 'Instalando actualización…';
    if (wailsState.value === 'ready' || wailsReady.value) return 'Actualización lista para aplicar';
    if (wailsError.value) return 'No se pudo completar la actualización';
    if (wailsAvailable.value) return '¿Quieres actualizar? Hay una versión nueva disponible';
    if (wailsState.value === 'up-to-date') return `Tienes la última versión instalada (v${currentDisplay.value})`;
    const info = updateInfo.value;
    if (!info) return 'Sistema de actualizaciones';
    if (info.error) return 'No se pudo completar la comprobación';
    if (info.hasUpdate) return `Hay una versión nueva disponible (v${info.latestVersion})`;
    return `Tienes la última versión instalada (v${currentDisplay.value})`;
});

const installLabel = computed(() => {
    if (wailsReady.value) return 'Reiniciar ahora';
    if (wailsAvailable.value) return 'Actualizar ahora';
    const info = updateInfo.value;
    if (!info) return 'Actualizar';
    if (isWindows.value && info.hasUpdater) return 'Actualizar ahora';
    if (info.hasUpdate) return 'Abrir GitHub';
    return 'Descargar desde GitHub';
});

const installText = computed(() => {
    if (wailsReady.value) return 'La actualización ya está descargada. El launcher se reiniciará para completar la instalación.';
    if (wailsAvailable.value) return 'Se descargará la nueva versión desde GitHub y se instalará automáticamente. ¿Quieres actualizar ahora o más tarde?';
    const info = updateInfo.value;
    if (!info) return '';
    if (isWindows.value && info.hasUpdater) {
        return 'Se descargará el instalador (steplauncher-…-installer.exe), el launcher se cerrará y el instalador completará la actualización automáticamente.';
    }
    if (info.hasUpdate) {
        return 'Hay una nueva actualización disponible y en tu sistema debes instalarla manualmente: se abrirá la release en GitHub para que descargues el paquete de tu plataforma (.deb, .rpm, .AppImage, .dmg o .app).';
    }
    return 'Se abrirá la última release en el navegador para que descargues la nueva versión manualmente.';
});

function close() {
    closeUpdateModal();
}

function onActualizarAhora() {
    if (wailsReady.value) {
        restartUpdate();
        return;
    }
    installUpdate();
}

function onMasTarde() {
    close();
}

useOverlayEscape(close, { isActive: () => modalVisible.value });

function onCloseOverlays() {
    close();
}

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});
</script>

<template>
    <Teleport to="body">
        <Transition name="UpdateModal">
            <div v-if="modalVisible" class="UpdateModal_Overlay" @click.self="close">
                <div class="UpdateModal_Dialog">
                    <div class="UpdateModal_Head">
                        <span
                            class="UpdateModal_Icon"
                            :class="{ none: updateInfo && !updateInfo.hasUpdate && !updateInfo.error, err: updateInfo && !!updateInfo.error }"
                        >
                            <IconArrowUpCircle stroke="2" />
                        </span>
                        <div class="UpdateModal_Titles">
                            <h3>{{ checking ? 'Buscando actualizaciones…' : 'Actualizaciones' }}</h3>
                            <p>{{ subtitle }}</p>
                        </div>
                        <button class="UpdateModal_Close" title="Cerrar" @click="close">
                            <IconX stroke="2" />
                        </button>
                    </div>

                    <div class="UpdateModal_Body">
                        <div v-if="checking" class="UpdateModal_State">
                            <img class="UpdateModal_Spinner" src="../../assets/gif/chicken_jockey_run.gif" alt="">
                            <span>Consultando la última release de GitHub…</span>
                        </div>

                        <!-- Wails headless: progreso de descarga -->
                        <div v-else-if="wailsState === 'downloading' && wailsDownloading" class="UpdateModal_State" style="flex-direction: column; gap: 0.8rem;">
                            <img class="UpdateModal_Spinner" src="../../assets/gif/chicken_jockey_run.gif" alt="">
                            <span>Descargando actualización… {{ wailsProgressPct }}%</span>
                            <span style="font-size: 0.72rem; opacity: 0.6;">{{ wailsProgressText }}</span>
                            <div style="width: 100%; height: 6px; background: var(--control-bg); border-radius: 99px; overflow: hidden; border: 1px solid var(--control-border);">
                                <div :style="{ width: wailsProgressPct + '%', height: '100%', background: 'var(--color-success)', transition: 'width 0.3s' }"></div>
                            </div>
                        </div>
                        <div v-else-if="wailsState === 'verifying' || wailsState === 'installing'" class="UpdateModal_State" style="flex-direction: column;">
                            <img class="UpdateModal_Spinner" src="../../assets/gif/chicken_jockey_run.gif" alt="">
                            <span>{{ wailsState === 'verifying' ? 'Verificando la descarga…' : 'Instalando actualización…' }}</span>
                        </div>
                        <div v-else-if="wailsReady" class="UpdateModal_State UpdateModal_StateDone" style="flex-direction: column; gap: 0.6rem;">
                            <IconRocket stroke="2" />
                            <span>¡Actualización descargada! Reinicia para aplicar la nueva versión.</span>
                            <span style="font-size: 0.72rem; opacity: 0.6;">{{ installText }}</span>
                        </div>
                        <div v-else-if="wailsError" class="UpdateModal_Error">
                            <IconAlertTriangle stroke="2" />
                            <div>
                                <span class="UpdateModal_ErrorTitle">Error de actualización</span>
                                <span class="UpdateModal_ErrorText">{{ wailsError }}</span>
                            </div>
                        </div>
                        <!-- Wails available: prompt ¿Quieres actualizar? -->
                        <div v-else-if="wailsAvailable" class="UpdateModal_Update">
                            <div class="UpdateModal_Versions">
                                <div class="UpdateModal_VersionBox">
                                    <span class="UpdateModal_Key">Versión actual</span>
                                    <span class="UpdateModal_Val">v{{ currentDisplay || '...' }}</span>
                                </div>
                                <div class="UpdateModal_Arrow">→</div>
                                <div class="UpdateModal_VersionBox UpdateModal_VersionBoxNew">
                                    <span class="UpdateModal_Key">Nueva versión</span>
                                    <span class="UpdateModal_Val">v{{ wailsRelease?.version || updateInfo?.latestVersion }}</span>
                                </div>
                            </div>
                            <div class="UpdateModal_Release">
                                <span class="UpdateModal_ReleaseName">
                                    {{ wailsRelease?.name || updateInfo?.releaseName || `StepLauncher v${wailsRelease?.version || updateInfo?.latestVersion}` }}
                                </span>
                                <span v-if="wailsRelease?.publishedAt || updateInfo?.releaseDate" class="UpdateModal_ReleaseDate">
                                    {{ wailsRelease?.publishedAt || updateInfo?.releaseDate }}
                                </span>
                            </div>
                            <div v-if="wailsRelease?.notes || updateInfo?.notes" class="UpdateModal_Notes">
                                <pre>{{ wailsRelease?.notes || updateInfo?.notes }}</pre>
                            </div>
                            <p class="UpdateModal_Explain">{{ installText }}</p>
                        </div>

                        <template v-else-if="updateInfo">
                            <div v-if="updateInfo.error" class="UpdateModal_Error">
                                <IconAlertTriangle stroke="2" />
                                <div>
                                    <span class="UpdateModal_ErrorTitle">No se pudo comprobar</span>
                                    <span class="UpdateModal_ErrorText">{{ updateInfo.error }}</span>
                                </div>
                            </div>

                            <div v-else-if="updateInfo.hasUpdate" class="UpdateModal_Update">
                                <div class="UpdateModal_Versions">
                                    <div class="UpdateModal_VersionBox">
                                        <span class="UpdateModal_Key">Versión actual</span>
                                        <span class="UpdateModal_Val">v{{ updateInfo.currentVersion }}</span>
                                    </div>
                                    <div class="UpdateModal_Arrow">→</div>
                                    <div class="UpdateModal_VersionBox UpdateModal_VersionBoxNew">
                                        <span class="UpdateModal_Key">Nueva versión</span>
                                        <span class="UpdateModal_Val">v{{ updateInfo.latestVersion }}</span>
                                    </div>
                                </div>
                                <div class="UpdateModal_Release">
                                    <span class="UpdateModal_ReleaseName">
                                        {{ updateInfo.releaseName || `StepLauncher v${updateInfo.latestVersion}` }}
                                    </span>
                                    <span v-if="updateInfo.releaseDate" class="UpdateModal_ReleaseDate">
                                        {{ updateInfo.releaseDate }}
                                    </span>
                                </div>
                                <div v-if="updateInfo.notes" class="UpdateModal_Notes">
                                    <pre>{{ updateInfo.notes }}</pre>
                                </div>
                                <p class="UpdateModal_Explain">{{ installText }}</p>
                            </div>

                            <div v-else class="UpdateModal_State UpdateModal_StateDone">
                                <IconCheck stroke="2" />
                                <span>¡Estás al día! Tienes la versión v{{ currentDisplay }} instalada.</span>
                            </div>
                        </template>
                        <!-- Wails up-to-date sin updateInfo (fallback) -->
                        <div v-else-if="wailsState === 'up-to-date'" class="UpdateModal_State UpdateModal_StateDone">
                            <IconCheck stroke="2" />
                            <span>¡Estás al día! Tienes la versión v{{ currentDisplay }} instalada.</span>
                        </div>
                    </div>

                    <div class="UpdateModal_Footer">
                        <!-- Wails ready: Actualizar ahora = Reiniciar -->
                        <template v-if="wailsReady">
                            <button class="SsBtn UpdateModal_Link" @click="onMasTarde">
                                <IconX :size="'14'" :stroke="'2'" />
                                Más tarde
                            </button>
                            <button class="SsBtn SsBtnPrimary" @click="onActualizarAhora">
                                <IconRocket :size="'14'" :stroke="'2'" />
                                Reiniciar ahora
                            </button>
                        </template>
                        <!-- Wails downloading/verifying/installing: solo spinner, botón deshabilitado -->
                        <template v-else-if="wailsDownloading || wailsState === 'verifying' || wailsState === 'installing'">
                            <button class="SsBtn" disabled>Descargando… {{ wailsProgressPct }}%</button>
                        </template>
                        <!-- Wails available o legacy hasUpdate: Actualizar Ahora / Más tarde -->
                        <template v-else-if="wailsAvailable || updateInfo?.hasUpdate">
                            <button class="SsBtn UpdateModal_Link" @click="onMasTarde">
                                <IconExternalLink :size="'14'" :stroke="'2'" />
                                Más tarde
                            </button>
                            <button class="SsBtn SsBtnPrimary" @click="onActualizarAhora">
                                <IconDownload :size="'14'" :stroke="'2'" />
                                {{ installLabel }}
                            </button>
                        </template>
                        <template v-else-if="wailsError">
                            <button class="SsBtn UpdateModal_Link" @click="close">
                                <IconX :size="'14'" :stroke="'2'" />
                                Cerrar
                            </button>
                            <button class="SsBtn SsBtnPrimary" @click="onMasTarde">
                                <IconRefresh :size="'14'" :stroke="'2'" />
                                Entendido
                            </button>
                        </template>
                        <button v-else class="SsBtn SsBtnPrimary" @click="close">
                            <IconCheck :size="'14'" :stroke="'2'" />
                            Entendido
                        </button>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use './Styles/Update.scss';
</style>