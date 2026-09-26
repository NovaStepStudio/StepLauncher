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
                <span class="SsLabel">Ajustes de GPU personalizados</span>
                <span class="SsDesc">Si lo apagás, Minecraft usa su configuración gráfica por defecto e ignora lo de abajo.</span>
            </div>
            <div class="SsCtrl">
                <label class="SsTg"><input type="checkbox" v-model="hwEnabled" @change="save"><span class="SsTgS"></span></label>
            </div>
        </div>
        <template v-if="hwEnabled">
            <div class="SsRow">
                <div class="SsInfo">
                <span class="SsLabel">Aceleración por hardware</span>
                <span class="SsDesc">Usa tu GPU para renderizar: más FPS y menos carga de CPU.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="hwAccel" @change="save"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                <span class="SsLabel">GPU preferida</span>
                <span class="SsDesc">En equipos con dos gráficas, forzá cuál usa el juego. Automático suele acertar.</span>
                </div>
                <div class="SsCtrl">
                    <select class="SsSel" v-model="gpuType" @change="save">
                        <option value="">Desactivado</option>
                        <option value="auto">Automático</option>
                        <option value="dedicated">Dedicada (NVIDIA/AMD)</option>
                        <option value="integrated">Integrada (Intel/AMD APU)</option>
                    </select>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                <span class="SsLabel">Perfil de la GPU</span>
                <span class="SsDesc">Rendimiento prioriza FPS, Calidad prioriza imagen, Balanceado reparte.</span>
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
                <span>Si el juego crashea al abrir o queda en pantalla negra, apagá estos ajustes y probá de nuevo.</span>
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
                <span class="SsLabel">RAM para Minecraft (−Xmx)</span>
                <span class="SsDesc">Cuánta memoria le asignás al juego. Tu PC tiene {{ totalRAM }} GB en total.</span>
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
                <span>Regla práctica: asigná la mitad de tu RAM total. Darle de más puede empeorar el rendimiento.</span>
        </div>
    </div>

    <div class="SsGroup">
        <div class="SsGroupHead">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 3l-1 8h10l-1-8H8z"/><path d="M4 16l2 4h12l2-4H4z"/><path d="M12 11v5"/><path d="M9 21h6"/></svg>
            <span>Java</span>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Versión de Java</span>
                <span class="SsDesc">Automático elige el Java correcto según la versión de Minecraft. Cambialo solo si sabés lo que hacés.</span>
            </div>
            <div class="SsCtrl">
                <select class="SsSel" v-model="javaMode" @change="save">
                    <option value="auto">Automático</option>
                    <option value="system">Instalado en mi PC</option>
                    <option value="official">El oficial de Minecraft</option>
                    <option value="custom">Ruta manual</option>
                </select>
            </div>
        </div>
        <template v-if="javaMode === 'system'">
            <div class="SsRow">
                <div class="SsInfo">
                <span class="SsLabel">Java detectados</span>
                <span class="SsDesc">Elegí uno de los Java encontrados en tu PC, o tocá Buscar para re-escanear.</span>
                </div>
                <div class="SsCtrl">
                    <select class="SsSel SsSelJava" v-model="selectedJava" @change="pickJava">
                        <option value="" disabled>Ninguno seleccionado</option>
                        <option v-for="j in javaSelectOptions" :key="j.path" :value="j.path">{{ j.display }}</option>
                    </select>
                    <button class="SsBtn" @click="scanJava" title="Volver a detectar los Java instalados">
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                        Buscar
                    </button>
                </div>
            </div>
        </template>
        <template v-else-if="javaMode === 'custom'">
            <div class="SsRow">
                <div class="SsInfo">
                <span class="SsLabel">Ruta manual a Java</span>
                <span class="SsDesc">Pegá la ruta completa al javaw.exe que querés usar (solo Windows).</span>
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
                <span class="SsLabel">Tamaño inicial de la ventana</span>
                <span class="SsDesc">Ancho × alto en píxeles con los que abre el juego (después lo podés cambiar dentro del juego).</span>
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
                <span class="SsLabel">Abrir en pantalla completa</span>
                <span class="SsDesc">El juego arranca ocupando todo el monitor. Después se alterna con F11.</span>
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
                <span class="SsLabel">Argumentos de la JVM</span>
                <span class="SsDesc">Flags extra para Java (ej. recolector de basura). Solo si sabés lo que hacés: un flag malo impide abrir el juego.</span>
            </div>
        </div>
        <div class="SsArg">
            <input class="SsIn SsInW" v-model="javaArgs" placeholder="Ej: -XX:+UseG1GC" @change="save">
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Argumentos del juego</span>
                <span class="SsDesc">Flags que recibe Minecraft al arrancar (ej. resolución). También solo para avanzados.</span>
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
                <span class="SsLabel">Verificar archivos antes de jugar</span>
                <span class="SsDesc">Comprueba por hash que el juego esté completo antes de abrirlo. Apagalo para lanzar más rápido si ya verificaste todo.</span>
            </div>
            <div class="SsCtrl">
                <label class="SsTg"><input type="checkbox" v-model="verifyBeforeLaunch" :disabled="isOffline" @change="saveVerifyBeforeLaunch"><span class="SsTgS"></span></label>
            </div>
        </div>
        <div v-if="isOffline" class="SsTip">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
            <span>Sin conexión, la verificación queda forzada para evitar arranques rotos. Se libera sola al volver internet.</span>
        </div>
        <div v-else class="SsTip">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
            <span>Si se cae internet al lanzar, esta verificación se activa sola para no abrir el juego con archivos incompletos.</span>
        </div>
    </div>

    <div class="SsGroup">
        <div class="SsGroupHead">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
            <span>Comportamiento</span>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Modo sin conexión</span>
                <span class="SsDesc">Permite jugar offline con tu última sesión. Los servidores online igual exigen internet.</span>
            </div>
            <div class="SsCtrl">
                <label class="SsTg"><input type="checkbox" v-model="offlineMode" @change="save"><span class="SsTgS"></span></label>
            </div>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Modo compatibilidad (PCs viejos)</span>
                <span class="SsDesc">Arranque conservador para equipos antiguos o drivers problemáticos. Probalo si el juego no abre.</span>
            </div>
            <div class="SsCtrl">
                <label class="SsTg"><input type="checkbox" v-model="compatMode" @change="save"><span class="SsTgS"></span></label>
            </div>
        </div>
        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Logs detallados</span>
                <span class="SsDesc">Guarda información extra del arranque. Activálo solo si vas a reportar un error, pesa más.</span>
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
