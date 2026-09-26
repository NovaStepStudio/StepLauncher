<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { GetConfig, SetConcurrentDownloads, SetMaxMbps } from '@wailsjs/StepLauncher/internal/Services/Config/configservice';

const concurrentDownloads = ref(4);
const maxMbps = ref(0);

async function loadConfig() {
    try {
        const cfg = await GetConfig();
        if (cfg) {
            concurrentDownloads.value = cfg.launcher?.concurrentDownloads ?? 4;
            maxMbps.value = cfg.launcher?.maxMbps ?? 0;
        }
    } catch (_e) {}
}

onMounted(() => {
    loadConfig();
});

async function saveDownloads() {
    try {
        await SetConcurrentDownloads(concurrentDownloads.value);
    } catch (_e) {}
}

async function saveMbps() {
    try {
        await SetMaxMbps(maxMbps.value);
    } catch (_e) {}
}
</script>

<template>
    <div class="Ss">

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                <span>Velocidad</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Descargas en paralelo</span>
                    <span class="SsDesc">Cuántos archivos bajan a la vez (1–8). Más es más rápido, pero satura tu conexión y puede fallar en redes flojas.</span>
                </div>
                <div class="SsCtrl">
                    <div class="SsStep">
                        <button class="SsStepBtn" :disabled="concurrentDownloads <= 1" @click="concurrentDownloads--; saveDownloads()">−</button>
                        <span class="SsStepVal">{{ concurrentDownloads }}</span>
                        <button class="SsStepBtn" :disabled="concurrentDownloads >= 8" @click="concurrentDownloads++; saveDownloads()">+</button>
                    </div>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Tope de velocidad</span>
                    <span class="SsDesc">Máximo en Mbps. En 0 (sin límite) usa todo tu ancho de banda: útil si nadie más usa internet.</span>
                </div>
                <div class="SsCtrl">
                    <div class="SsStep">
                        <button class="SsStepBtn" :disabled="maxMbps <= 0" @click="maxMbps = Math.max(0, maxMbps - 5); saveMbps()">−</button>
                        <span class="SsStepVal">{{ maxMbps === 0 ? 'Sin límite' : maxMbps + ' Mbps' }}</span>
                        <button class="SsStepBtn" :disabled="maxMbps >= 500" @click="maxMbps += 5; saveMbps()">+</button>
                    </div>
                </div>
            </div>
        </div>

        <div class="SsTip">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
            <span>¿Descargas corruptas o juego que no abre? Activá la verificación automática en <strong>Integridad</strong>.</span>
        </div>

    </div>
</template>

<style scoped lang="scss">
@use '../Styles/General.scss';
</style>
