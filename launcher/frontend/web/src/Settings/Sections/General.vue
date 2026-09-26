<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { setUIScale, applyPersonalization, uiScale, personalization } from '@/Common/Stores/Ui';
import { saveIdleOptions, CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { checkForUpdates as requestUpdateCheck, checking as updateChecking } from '@/Updates/Store';
import { GetConfig, ResetConfig, SetCheckForUpdatesOnStart, SetLaunchAfterInstall } from '@wailsjs/StepLauncher/internal/Services/Config/configservice';
import { SetHideLauncher, SetUIScale, UpdatePersonalization } from '@wailsjs/StepLauncher/internal/Services/Appearance/appearanceservice';
import { SetRichPresenceEnabled } from '@wailsjs/StepLauncher/internal/Services/Account/accountservice';

const hideLauncher = ref(true);
const richPresence = ref(true);
const launchAfterInstall = ref(false);

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

const checkOnStart = ref(false);

async function loadConfig() {
    try {
        const cfg = await GetConfig();
        if (cfg) {
            hideLauncher.value = cfg.launcher?.hideLauncherOnLaunch ?? true;
            richPresence.value = (cfg.launcher as any)?.richPresence ?? true;
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
            checkOnStart.value = cfg.launcher?.checkForUpdatesOnStart ?? false;
            launchAfterInstall.value = cfg.launcher?.launchAfterInstall ?? false;
        }
    } catch (_e) {}
}

async function saveHideLauncher() {
    try {
        await SetHideLauncher(hideLauncher.value);
    } catch (_e) {}
}

async function saveRichPresence() {
    try {
        await SetRichPresenceEnabled(richPresence.value);
    } catch (_e) {}
}

async function saveZoom() {
    try {
        await SetUIScale(uiScale.value);
    } catch (_e) {}
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
        await UpdatePersonalization(p as any);
    } catch (_e) {}
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
});

async function resetConfig() {
    showResetConfirm.value = false;
    try {
        await ResetConfig();
        window.location.reload();
    } catch (_e) {}
    await loadConfig();
}

async function checkUpdates() {
    try {
        await requestUpdateCheck(false);
    } catch (_e) {}
}

async function saveCheckOnStart() {
    try {
        await SetCheckForUpdatesOnStart(checkOnStart.value);
    } catch (_e) {}
}

async function saveLaunchAfterInstall() {
    try {
        await SetLaunchAfterInstall(launchAfterInstall.value);
    } catch (_e) {}
}
</script>

<template>
    <div class="Ss">

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><line x1="11" y1="8" x2="11" y2="14"/><line x1="8" y1="11" x2="14" y2="11"/></svg>
                <span>Tamaño</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Escala de la interfaz</span>
                    <span class="SsDesc">Agranda o achica todo el launcher, de 50% a 200%. Atajo rápido: Ctrl + y Ctrl −.</span>
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
                <span>Al jugar</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Ocultar el launcher al jugar</span>
                    <span class="SsDesc">Minimiza el launcher a la bandeja mientras el juego está abierto y lo restaura al cerrarlo. Ahorra recursos.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="hideLauncher" @change="saveHideLauncher"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Estado en Discord (Rich Presence)</span>
                    <span class="SsDesc">Muestra en tu perfil de Discord qué versión estás jugando. Requiere tener Discord abierto.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="richPresence" @change="saveRichPresence"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Jugar automáticamente al terminar de instalar</span>
                    <span class="SsDesc">Lanza el juego solo cuando una descarga termina. Ideal si instalás versiones pesadas y te alejás.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="launchAfterInstall" @change="saveLaunchAfterInstall"><span class="SsTgS"></span></label>
                </div>
            </div>
        </div>

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
                <span>Efectos visuales</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Animaciones</span>
                    <span class="SsDesc">Transiciones suaves al abrir y cerrar ventanas. Apagalas si tu PC va justo.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="animations" @change="saveRendimiento"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Desenfoque de fondo (blur)</span>
                    <span class="SsDesc">Aplica blur detrás de las ventanas. Queda lindo, pero usa GPU.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="blur" @change="saveRendimiento"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Sombras de ventanas</span>
                    <span class="SsDesc">Profundidad bajo ventanas y diálogos. Solo estético.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="shadows" @change="saveRendimiento"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Resplandor en el texto</span>
                    <span class="SsDesc">Un glow sutil sobre títulos y textos para darles relieve.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="textShadow" @change="saveRendimiento"><span class="SsTgS"></span></label>
                </div>
            </div>
            <template v-if="textShadow">
                <div class="SsRow">
                    <div class="SsInfo">
                    <span class="SsLabel">Intensidad del resplandor</span>
                    <span class="SsDesc">De ×0.50 (sutil) a ×2.00 (marcado).</span>
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
                <span>En tu ausencia</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Cerrar ventanas al ausentarte</span>
                    <span class="SsDesc">Si no tocás nada por un rato, cierra los paneles abiertos y vuelve al inicio. Tus descargas siguen en curso.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="autoCloseModals" @change="saveIdle"><span class="SsTgS"></span></label>
                </div>
            </div>
            <template v-if="autoCloseModals">
                <div class="SsRow">
                    <div class="SsInfo">
                    <span class="SsLabel">Tiempo de espera</span>
                    <span class="SsDesc">Minutos de inactividad antes de cerrar las ventanas (1–10).</span>
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
                    <span class="SsLabel">Verificar la personalización</span>
                    <span class="SsDesc">Revisa cada cierto tiempo que tus colores y tipografías sigan aplicados, por si algo externo los pisa.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="configCheckEnabled" @change="saveIdle"><span class="SsTgS"></span></label>
                </div>
            </div>
            <template v-if="configCheckEnabled">
                <div class="SsRow">
                    <div class="SsInfo">
                    <span class="SsLabel">Frecuencia de revisión</span>
                    <span class="SsDesc">Cada cuántos minutos se hace la comprobación (1–10).</span>
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
                <span>Zona de peligro</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Restablecer ajustes de fábrica</span>
                    <span class="SsDesc">Borra tu configuración y personalización, y deja todo como recién instalado. Tus mundos e instancias no se tocan.</span>
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
                    <span class="SsLabel">Buscar actualizaciones al abrir</span>
                    <span class="SsDesc">Consulta GitHub cada vez que iniciás el launcher y te avisa si hay versión nueva.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="checkOnStart" @change="saveCheckOnStart"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Buscar ahora</span>
                    <span class="SsDesc">Comprueba manualmente si hay una versión nueva disponible.</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn SsBtnPrimary" :disabled="updateChecking" @click="checkUpdates">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                        {{ updateChecking ? 'Buscando…' : 'Buscar' }}
                    </button>
                </div>
            </div>
        </div>

        <Teleport to="body">
            <div v-if="showResetConfirm" class="ConfirmOverlay" @click.self="showResetConfirm = false">
                <div class="ConfirmDialog">
                    <h3>Restablecer todo</h3>
                    <p>¿Seguro? Se borran tus ajustes y personalización (mundos e instancias a salvo) y no se puede deshacer.</p>
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
