<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { DetectJavaInstallations, TotalRAMGB, SetMaxRAM, MaxRAMGB, GetMinecraftConfig, UpdateMinecraftConfig, GetConfig, SetVerifyBeforeLaunch } from '@wailsjs/StepLauncher/internal/Services/Config/configservice';
import { isOffline } from '@/Common/Stores/Connectivity';

const hwEnabled = ref(true);
const hwAccel = ref(true);
const gpuType = ref('auto');
const gpuPreset = ref('');

const javaMode = ref('auto');
const javaCustomPath = ref('');

const totalRAM = ref(0);
const maxRAM = ref(2);

interface JavaDetected {
    path: string;
    name: string;
    version: string;
    display: string;
}
const detectedJava = ref<JavaDetected[]>([]);
const selectedJava = ref('');

const winW = ref(854);
const winH = ref(480);
const fullscreen = ref(false);

const javaArgs = ref('');
const gameArgs = ref('');

const offlineMode = ref(false);
const compatMode = ref(false);
const detailedLogs = ref(false);

const verifyBeforeLaunch = ref(true);

function parseDetected(list: string[]): JavaDetected[] {
    const out: JavaDetected[] = [];
    for (const entry of list) {
        let path = entry;
        let version = '';
        const idx = entry.lastIndexOf(' (');
        if (idx > 0 && entry.endsWith(')')) {
            path = entry.slice(0, idx);
            version = entry.slice(idx + 2, -1);
        }
        if (!path || out.some((j) => j.path === path)) continue;
        out.push({
            path,
            name: version,
            version,
            display: version ? `Java ${version}` : 'Java',
        });
    }
    return out;
}

const javaSelectOptions = computed(() => {
    return [...detectedJava.value];
});

async function scanJava() {
    try {
        const list = await DetectJavaInstallations?.();
        if (Array.isArray(list)) detectedJava.value = parseDetected(list);
    } catch (_e) {}
}

function pickJava() {
    if (!selectedJava.value) return;
    javaCustomPath.value = selectedJava.value;
    save();
}

async function detectRAM() {
    try {
        const total = await TotalRAMGB?.();
        if (total && typeof total === 'number') {
            totalRAM.value = total;
            return;
        }
    } catch (_e) {}

    if ((navigator as any).deviceMemory) {
        totalRAM.value = Math.round((navigator as any).deviceMemory as number);
        return;
    }

    totalRAM.value = 8;
}

async function saveRAM() {
    try {
        await SetMaxRAM?.(maxRAM.value);
    } catch (_e) {}
}

onMounted(async () => {
    await detectRAM();

    try {
        const cur = await MaxRAMGB?.();
        if (cur && typeof cur === 'number') maxRAM.value = cur;
    } catch (_e) {}

    try {
        const c = await GetMinecraftConfig?.();
        if (c) {
            hwEnabled.value = c.hardwareEnabled ?? true;
            hwAccel.value = c.hardwareAcceleration ?? true;
            gpuType.value = c.gpuType ?? 'auto';
            gpuPreset.value = c.gpuPreset ?? '';
            javaMode.value = c.javaMode ?? 'auto';
            javaCustomPath.value = c.javaCustomPath ?? '';
            winW.value = c.windowWidth ?? 854;
            winH.value = c.windowHeight ?? 480;
            fullscreen.value = c.fullscreen ?? false;
            javaArgs.value = c.javaArgs ?? '';
            gameArgs.value = c.gameArgs ?? '';
            offlineMode.value = c.offlineMode ?? false;
            compatMode.value = c.compatMode ?? false;
            detailedLogs.value = c.detailedLogs ?? false;
        }
    } catch (_e) {}
    try {
        const cfg = await GetConfig?.();
        if (cfg?.launcher) {
            verifyBeforeLaunch.value = cfg.launcher.verifyBeforeLaunch ?? true;
        }
    } catch (_e) {}

    await scanJava();
    if (javaCustomPath.value && detectedJava.value.some((j) => j.path === javaCustomPath.value)) {
        selectedJava.value = javaCustomPath.value;
    }
});

async function save() {
    try {
        await UpdateMinecraftConfig?.({
            hardwareEnabled: hwEnabled.value,
            hardwareAcceleration: hwAccel.value,
            gpuType: gpuType.value,
            gpuPreset: gpuPreset.value,
            javaMode: javaMode.value,
            javaCustomPath: javaCustomPath.value,
            windowWidth: winW.value,
            windowHeight: winH.value,
            fullscreen: fullscreen.value,
            javaArgs: javaArgs.value,
            gameArgs: gameArgs.value,
            offlineMode: offlineMode.value,
            compatMode: compatMode.value,
            detailedLogs: detailedLogs.value,
        } as any);
    } catch (_e) {}
}

async function saveVerifyBeforeLaunch() {
    try {
        await SetVerifyBeforeLaunch?.(verifyBeforeLaunch.value);
    } catch (_e) {}
}
</script>

<template>
<div class="Ss">

    <div class="SsGroup">
        <div class="SsGroupHead">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="6" height="6"/><path d="M15 2v2"/><path d="M15 20v2"/><path d="M2 15h2"/><path d="M20 15h2"/></svg>
            <span>Gráficos</span>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Usar opciones de gráfica</span>
                <span class="SsDesc">Si lo apagas, el juego usa lo que trae por defecto.</span>
            </div>
            <div class="SsCtrl">
                <label class="SsTg"><input type="checkbox" v-model="hwEnabled" @change="save"><span class="SsTgS"></span></label>
            </div>
        </div>
        <template v-if="hwEnabled">
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Aceleración</span>
                    <span class="SsDesc">Usa tu gráfica para que vaya más fluido.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="hwAccel" @change="save"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Qué gráfica usar</span>
                    <span class="SsDesc">Automático elige la mejor para tu PC.</span>
                </div>
                <div class="SsCtrl">
                    <select class="SsSel" v-model="gpuType" @change="save">
                        <option value="">Desactivado</option>
                        <option value="auto">Automático</option>
                        <option value="dedicated">Dedicada</option>
                        <option value="integrated">Integrada</option>
                    </select>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Modo de la gráfica</span>
                    <span class="SsDesc">Rendimiento va más rápido, Calidad se ve mejor.</span>
                </div>
                <div class="SsCtrl">
                    <select class="SsSel" v-model="gpuPreset" @change="save">
                        <option value="">Desactivado</option>
                        <option value="performance">Rendimiento</option>
                        <option value="balanced">Balanceado</option>
                        <option value="quality">Calidad</option>
                    </select>
                </div>
            </div>
            <div class="SsTip">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
                <span>Si el juego no abre, prueba apagando esto.</span>
            </div>
        </template>
    </div>

    <div class="SsGroup">
        <div class="SsGroupHead">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="6" height="6"/><path d="M15 2v2"/><path d="M15 20v2"/><path d="M2 15h2"/><path d="M20 15h2"/></svg>
            <span>Memoria</span>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Memoria para Minecraft</span>
                <span class="SsDesc">Cuánta memoria le das al juego. Tienes {{ totalRAM }} GB en total.</span>
            </div>
            <div class="SsCtrl">
                <div class="SsStep">
                    <button class="SsStepBtn" :disabled="maxRAM <= 1" @click="maxRAM--; saveRAM()">−</button>
                    <span class="SsStepVal">{{ maxRAM }} GB</span>
                    <button class="SsStepBtn" :disabled="totalRAM > 0 && maxRAM >= totalRAM - 1" @click="maxRAM++; saveRAM()">+</button>
                </div>
            </div>
        </div>
        <div class="SsTip">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
            <span>Tip: usa la mitad de tu memoria total.</span>
        </div>
    </div>

    <div class="SsGroup">
        <div class="SsGroupHead">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 3l-1 8h10l-1-8H8z"/><path d="M4 16l2 4h12l2-4H4z"/><path d="M12 11v5"/><path d="M9 21h6"/></svg>
            <span>Java</span>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Qué Java usar</span>
                <span class="SsDesc">Déjalo en Automático si no sabes qué es esto.</span>
            </div>
            <div class="SsCtrl">
                <select class="SsSel" v-model="javaMode" @change="save">
                    <option value="auto">Automático</option>
                    <option value="system">Elegir de mi PC</option>
                    <option value="official">El que trae Minecraft</option>
                    <option value="custom">Ruta manual</option>
                </select>
            </div>
        </div>
        <template v-if="javaMode === 'system'">
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Elige uno</span>
                    <span class="SsDesc">Busca los Java que tienes y elige.</span>
                </div>
                <div class="SsCtrl">
                    <select class="SsSel SsSelJava" v-model="selectedJava" @change="pickJava">
                        <option value="" disabled>Ninguno</option>
                        <option v-for="j in javaSelectOptions" :key="j.path" :value="j.path">{{ j.display }}</option>
                    </select>
                    <button class="SsBtn" @click="scanJava" title="Detectar Java instalados">
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                        Buscar
                    </button>
                </div>
            </div>
        </template>
        <template v-else-if="javaMode === 'custom'">
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Dónde está Java</span>
                    <span class="SsDesc">Pega la ruta a javaw.exe</span>
                </div>
                <div class="SsCtrl">
                    <input class="SsIn" v-model="javaCustomPath" placeholder="C:\ruta\javaw.exe" @change="save">
                </div>
            </div>
        </template>
    </div>

    <div class="SsGroup">
        <div class="SsGroupHead">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
            <span>Ventana</span>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Tamaño al abrir</span>
                <span class="SsDesc">Ancho y alto de la ventana del juego.</span>
            </div>
            <div class="SsGrid">
                <div class="SsField">
                    <span class="SsLabel">Ancho</span>
                    <input class="SsIn" type="number" v-model.number="winW" @change="save">
                </div>
                <div class="SsField">
                    <span class="SsLabel">Alto</span>
                    <input class="SsIn" type="number" v-model.number="winH" @change="save">
                </div>
            </div>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Pantalla completa</span>
                <span class="SsDesc">Que el juego ocupe toda la pantalla al abrir.</span>
            </div>
            <div class="SsCtrl">
                <label class="SsTg"><input type="checkbox" v-model="fullscreen" @change="save"><span class="SsTgS"></span></label>
            </div>
        </div>
    </div>

    <div class="SsGroup">
        <div class="SsGroupHead">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>
            <span>Avanzado</span>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Opciones para Java</span>
                <span class="SsDesc">Solo toca esto si sabes lo que haces.</span>
            </div>
        </div>
        <div class="SsArg">
            <input class="SsIn SsInW" v-model="javaArgs" placeholder="Ej: -XX:+UseG1GC" @change="save">
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Opciones para el juego</span>
                <span class="SsDesc">También solo para avanzados.</span>
            </div>
        </div>
        <div class="SsArg">
            <input class="SsIn SsInW" v-model="gameArgs" placeholder="Ej: --width 854" @change="save">
        </div>
    </div>

    <div class="SsGroup">
        <div class="SsGroupHead">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 12l2 2 4-4"/><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
            <span>Lanzamiento</span>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Verificar antes de lanzar Minecraft</span>
                <span class="SsDesc">Si está activo, comprueba que todos los archivos del juego estén completos antes de iniciar. Desactívalo para lanzar más rápido si sabes que todo está descargado.</span>
            </div>
            <div class="SsCtrl">
                <label class="SsTg"><input type="checkbox" v-model="verifyBeforeLaunch" :disabled="isOffline" @change="saveVerifyBeforeLaunch"><span class="SsTgS"></span></label>
            </div>
        </div>
        <div v-if="isOffline" class="SsTip">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
            <span>Sin internet esta opción está forzada y no se puede desactivar. Volverá a funcionar al recuperar la conexión.</span>
        </div>
        <div v-else class="SsTip">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
            <span>Sin internet esta verificación se activará automáticamente para evitar que el juego falle por archivos incompletos.</span>
        </div>
    </div>

    <div class="SsGroup">
        <div class="SsGroupHead">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
            <span>Más opciones</span>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Jugar sin internet</span>
                <span class="SsDesc">Te deja entrar aunque no tengas conexión.</span>
            </div>
            <div class="SsCtrl">
                <label class="SsTg"><input type="checkbox" v-model="offlineMode" @change="save"><span class="SsTgS"></span></label>
            </div>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Modo para PCs viejos</span>
                <span class="SsDesc">Si tu PC tiene muchos años y el juego falla, prueba esto.</span>
            </div>
            <div class="SsCtrl">
                <label class="SsTg"><input type="checkbox" v-model="compatMode" @change="save"><span class="SsTgS"></span></label>
            </div>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Guardar más info si falla</span>
                <span class="SsDesc">Útil si necesitas reportar un error.</span>
            </div>
            <div class="SsCtrl">
                <label class="SsTg"><input type="checkbox" v-model="detailedLogs" @change="save"><span class="SsTgS"></span></label>
            </div>
        </div>
    </div>

</div>
</template>

<style scoped lang="scss">
@use '../Styles/Minecraft.scss';
</style>
