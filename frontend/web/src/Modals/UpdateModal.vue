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
import { updateInfo, checking, modalVisible, installUpdate, closeUpdateModal } from '../Stores/Update';
import { CLOSE_OVERLAYS_EVENT } from '../Stores/Idle';

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
                            <img class="UpdateModal_Spinner" src="../../assets/gif/loading.gif" alt="">
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
@use '../Styles/Settings.scss' as *;

.UpdateModal_Overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: var(--filter-blur);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 140;
}

.UpdateModal_Dialog {
    width: 32rem;
    max-width: 92vw;
    max-height: 88vh;
    background: var(--background-modal-primary);
    border: var(--border-modal-style);
    border-radius: 0.85rem;
    box-shadow: var(--shadow-settings-normal) #000a;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.UpdateModal_Head {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1.1rem 1.35rem 0.85rem;
    border-bottom: var(--border-modal-style);
    background: #0005;
}

.UpdateModal_Icon {
    width: 2.4rem;
    height: 2.4rem;
    flex-shrink: 0;
    border-radius: 0.6rem;
    background: linear-gradient(135deg, color-mix(in srgb, var(--color-success) 35%, transparent), color-mix(in srgb, var(--color-success) 10%, transparent));
    border: 1px solid color-mix(in srgb, var(--color-success) 35%, transparent);
    display: flex;
    justify-content: center;
    align-items: center;
    color: var(--color-success);

    &.none {
        background: linear-gradient(135deg, color-mix(in srgb, var(--background-button-primary) 40%, transparent), color-mix(in srgb, var(--background-button-primary) 12%, transparent));
        border-color: color-mix(in srgb, var(--background-button-primary) 35%, transparent);
        color: var(--text-secondary);
    }

    &.err {
        background: linear-gradient(135deg, color-mix(in srgb, var(--color-error) 35%, transparent), color-mix(in srgb, var(--color-error) 10%, transparent));
        border-color: color-mix(in srgb, var(--color-error) 35%, transparent);
        color: var(--color-error);
    }
}

.UpdateModal_Icon svg {
    width: 1.35rem;
    height: 1.35rem;
}

.UpdateModal_Titles {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    min-width: 0;

    h3 {
        margin: 0;
        font-family: var(--font-primary), Arial, sans-serif;
        font-size: calc(0.95rem * var(--font-size-primary, 1));
        font-weight: 600;
    }

    p {
        margin: 0;
        font-size: 0.68rem;
        opacity: 0.5;
    }
}

.UpdateModal_Close {
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

.UpdateModal_Body {
    display: flex;
    flex-direction: column;
    gap: 0.9rem;
    padding: 1.1rem 1.35rem;
    overflow-y: auto;
}

.UpdateModal_State {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1.25rem 0;
    justify-content: center;
    font-size: 0.8rem;
    opacity: 0.75;
}

.UpdateModal_Spinner {
    width: 1.1rem;
    height: auto;
    image-rendering: pixelated;
}

.UpdateModal_Error {
    display: flex;
    align-items: flex-start;
    gap: 0.7rem;
    padding: 0.85rem 0.95rem;
    border-radius: 0.5rem;
    background: color-mix(in srgb, var(--color-error) 12%, transparent);
    border: 1px solid color-mix(in srgb, var(--color-error) 30%, transparent);
    color: var(--color-error);

    & > svg {
        width: 1.25rem;
        height: 1.25rem;
        flex-shrink: 0;
    }

    & > div {
        display: flex;
        flex-direction: column;
        gap: 0.15rem;
    }
}

.UpdateModal_ErrorTitle {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
}

.UpdateModal_ErrorText {
    font-size: 0.72rem;
    line-height: 1.5;
    opacity: 0.75;
}

.UpdateModal_Update {
    display: flex;
    flex-direction: column;
    gap: 0.9rem;
}

.UpdateModal_Versions {
    display: flex;
    align-items: center;
    gap: 0.6rem;
}

.UpdateModal_VersionBox {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    padding: 0.6rem 0.75rem;
    border-radius: 0.5rem;
    background: var(--control-bg);
    border: 1px solid var(--control-border);

    &.UpdateModal_VersionBoxNew {
        border-color: color-mix(in srgb, var(--color-success) 45%, transparent);
        background: color-mix(in srgb, var(--color-success) 8%, transparent);
    }
}

.UpdateModal_Key {
    font-size: 0.62rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    opacity: 0.45;
}

.UpdateModal_Val {
    font-size: 0.95rem;
    font-weight: 700;
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
}

.UpdateModal_Arrow {
    font-size: 1.1rem;
    opacity: 0.5;
}

.UpdateModal_Release {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
}

.UpdateModal_ReleaseName {
    font-size: 0.82rem;
    font-weight: 600;
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
}

.UpdateModal_ReleaseDate {
    font-size: 0.68rem;
    opacity: 0.5;
}

.UpdateModal_Notes {
    max-height: 10rem;
    overflow-y: auto;
    padding: 0.75rem 0.9rem;
    border-radius: 0.5rem;
    background: rgba(0, 0, 0, 0.35);
    border: var(--border-style);
    font-size: 0.72rem;
    line-height: 1.55;
    white-space: pre-wrap;
    word-break: break-word;
    opacity: 0.85;

    pre {
        margin: 0;
        font-family: inherit;
        white-space: pre-wrap;
        word-break: break-word;
    }
}

.UpdateModal_Explain {
    margin: 0;
    font-size: 0.72rem;
    line-height: 1.5;
    opacity: 0.6;
}

.UpdateModal_StateDone {
    display: flex;
    align-items: center;
    gap: 0.7rem;
    padding: 1.25rem 0;
    justify-content: center;
    font-size: 0.82rem;
    color: var(--color-success);

    svg {
        width: 1.4rem;
        height: 1.4rem;
    }
}

.UpdateModal_Footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 0.85rem 1.35rem 1.1rem;
    border-top: var(--border-modal-style);
}

.UpdateModal_Link {
    color: var(--text-secondary);
}

.UpdateModal-enter-active,
.UpdateModal-leave-active {
    transition: opacity 180ms ease;

    .UpdateModal_Dialog {
        transition: transform 200ms ease, opacity 180ms ease;
    }
}

.UpdateModal-enter-from,
.UpdateModal-leave-to {
    opacity: 0;

    .UpdateModal_Dialog {
        transform: translateY(8px) scale(0.97);
        opacity: 0;
    }
}
</style>