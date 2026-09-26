<script setup lang="ts">
import { ref, computed, onMounted, onActivated, onUnmounted } from 'vue';
import { IntegrityStatus as IntegrityStatusBinding, SetIntegritySector, StartIntegrityCheck, GetConfig, SetVerifyIntegrity } from '@wailsjs/StepLauncher/internal/Services/Config/configservice';

const verifyIntegrity = ref(true);

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
        case 'indexing': return 'Inventariando archivos esperados';
        case 'existence': return 'Descargando lo que falta';
        case 'retry': return 'Reintentando descargas fallidas';
        case 'verify': return 'Verificando hashes';
        case 'done': return 'Completado';
        default: return 'Preparando la revisión…';
    }
});

const integrityDoneText = computed(() => {
    const st = integrityStatus.value;
    if (!st || st.state === 'running') return '';
    if (st.state === 'completed') {
        return `Listo: revisé ${st.versionsScanned} versiones y reparé ${st.filesRestored} archivos.`;
    }
    if (st.state === 'cancelled') return 'Revisión cancelada por vos.';
    if (st.state === 'error') return 'Falló la revisión. Revisá los logs para el detalle técnico.';
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
        const st = await IntegrityStatusBinding();
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
    } catch (_e) {}
}

async function syncIntegrityFromBackend() {
    try {
        const st = await IntegrityStatusBinding();
        if (!st) return;
        integrityStatus.value = st;
        if (st.state === 'running') {
            integrityBusy.value = true;
            if (integrityTimer === null) {
                integrityTimer = window.setInterval(pollIntegrity, 500);
            }
        }
    } catch (_e) {}
}

async function setIntegrityScope(scope: string) {
    integrityScope.value = scope;
    try {
        await SetIntegritySector(scope);
    } catch (_e) {}
}

async function startIntegrityCheck() {
    if (integrityBusy.value) return;
    integrityBusy.value = true;
    integrityStatus.value = null;
    try {
        await SetIntegritySector(integrityScope.value);
        await StartIntegrityCheck(integrityScope.value);
    } catch (_e) {}
    stopIntegrityPolling();
    integrityTimer = window.setInterval(pollIntegrity, 500);
    await pollIntegrity();
}

async function loadConfig() {
    try {
        const cfg = await GetConfig();
        if (cfg) {
            verifyIntegrity.value = cfg.launcher?.verifyIntegrity ?? true;
            integrityScope.value = cfg.launcher?.integritySector ?? 'todo';
        }
    } catch (_e) {}
}

async function saveVerifyIntegrity() {
    try {
        await SetVerifyIntegrity(verifyIntegrity.value);
    } catch (_e) {}
}

onMounted(() => {
    loadConfig();
});

onActivated(async () => {
    await syncIntegrityFromBackend();
});

onUnmounted(() => {
    stopIntegrityPolling();
});
</script>

<template>
    <div class="Ss">

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><path d="M9 12l2 2 4-4"/></svg>
                <span>Verificación automática</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Verificar cada descarga (hash)</span>
                    <span class="SsDesc">Comprueba la firma de cada archivo al bajarlo. Es un poco más lento, pero evita arranques rotos. Recomendado: dejar activo.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="verifyIntegrity" @change="saveVerifyIntegrity"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsTip">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
                <span>Esto no hace nada ahora: actúa solo mientras descargás.</span>
            </div>
        </div>

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><polyline points="9 11 11 13 15 9"/></svg>
                <span>Verificación manual</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Alcance de la revisión</span>
                    <span class="SsDesc">Elegí qué revisar: todo el contenido, solo los archivos base del juego o solo tus instancias (mods, mundos y configs).</span>
                </div>
                <div class="SsCtrl">
                    <div class="SsSeg">
                        <button :class="{ active: integrityScope === 'todo' }" :disabled="integrityBusy" @click="setIntegrityScope('todo')">Todo</button>
                        <button :class="{ active: integrityScope === 'global' }" :disabled="integrityBusy" @click="setIntegrityScope('global')">Juego base</button>
                        <button :class="{ active: integrityScope === 'instances' }" :disabled="integrityBusy" @click="setIntegrityScope('instances')">Instancias</button>
                    </div>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Reparar archivos ahora</span>
                    <span class="SsDesc">Detecta faltantes y corruptos por hash y los re-descarga. Puede tardar varios minutos según tu disco e internet.</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn SsBtnPrimary" :disabled="integrityBusy" @click="startIntegrityCheck">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
                        {{ integrityBusy ? `Revisando… ${integrityPercent}%` : (integrityStatus && integrityStatus.state === 'completed' ? 'Listo' : 'Revisar ahora') }}
                    </button>
                </div>
            </div>
            <template v-if="integrityBusy">
                <div class="SsRow">
                    <div class="SsInfo">
                        <span class="SsDesc">{{ integrityPhaseLabel }}</span>
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
            <div class="SsTip">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
                <span>Esto sí trabaja ahora mismo. Para prevenir (no curar), usá la verificación automática de arriba.</span>
            </div>
        </div>

    </div>
</template>

<style scoped lang="scss">
@use '../Styles/General.scss';
</style>
