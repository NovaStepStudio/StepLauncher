<script setup lang="ts">
import { onMounted, onUnmounted, computed } from 'vue';
import {
    IconArrowUpCircle,
    IconCheck,
    IconX,
    IconAlertTriangle,
    IconExternalLink,
    IconDownload,
} from '@tabler/icons-vue';
import { updateInfo, checking, modalVisible, installUpdate, closeUpdateModal } from './Store';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

const isWindows = computed(() => (updateInfo.value?.platform ?? '') === 'windows');

const subtitle = computed(() => {
    const info = updateInfo.value;
    if (checking.value) return 'Consultando la última release de GitHub…';
    if (!info) return 'Sistema de actualizaciones';
    if (info.error) return 'No se pudo completar la comprobación';
    if (info.hasUpdate) return `Hay una versión nueva disponible (v${info.latestVersion})`;
    return `Tienes la última versión instalada (v${info.currentVersion})`;
});

const installLabel = computed(() => {
    const info = updateInfo.value;
    if (!info) return 'Actualizar';
    if (isWindows.value && info.hasUpdater) return 'Actualizar ahora';
    return 'Descargar desde GitHub';
});

const installText = computed(() => {
    const info = updateInfo.value;
    if (!info) return '';
    if (isWindows.value && info.hasUpdater) {
        return 'Se descargará el StepLauncher-Updater.exe, el launcher se cerrará y el actualizador completará la instalación automáticamente.';
    }
    return 'Se abrirá la última release en el navegador para que descargues la nueva versión manualmente.';
});

function close() {
    closeUpdateModal();
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
                                <span>¡Estás al día! Tienes la versión v{{ updateInfo.currentVersion }} instalada.</span>
                            </div>
                        </template>
                    </div>

                    <div class="UpdateModal_Footer">
                        <button v-if="updateInfo?.hasUpdate" class="SsBtn UpdateModal_Link" @click="close">
                            <IconExternalLink :size="'14'" :stroke="'2'" />
                            Más tarde
                        </button>
                        <button
                            v-if="updateInfo?.hasUpdate"
                            class="SsBtn SsBtnPrimary"
                            @click="installUpdate"
                        >
                            <IconDownload :size="'14'" :stroke="'2'" />
                            {{ installLabel }}
                        </button>
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