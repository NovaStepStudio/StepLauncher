<script setup lang="ts">
import { ref, reactive, watch, onUnmounted } from 'vue';
import { IconX } from '@tabler/icons-vue';
import {
    loadDetails,
    detailOf,
    updateConfig,
    startInstanceVerify,
    cancelInstanceVerify,
    refreshInstanceVerify,
    instanceVerifyOf,
    isInstanceVerifying,
    createInstanceBackup,
} from './Store';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

const props = defineProps<{
    visible: boolean;
    name: string;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
    (e: 'done', ok: boolean, message: string): void;
}>();

const busy = ref(false);
const msg = ref('');
const msgOk = ref(true);
const loading = ref(false);

const form = reactive({
    maxRam: '',
    useOfficialJava: true,
    javaExec: '',
    fullscreen: false,
    customResolution: false,
    resWidth: '',
    resHeight: '',
    gcEnabled: false,
    gcPreset: '',
    gpuEnabled: false,
    gpuPreference: '',
    detailedLogs: false,
    javaArgs: '',
    gameArgs: '',
    backupSchedule: 'off',
    backupInterval: '',
});

const GC_PRESETS = [
    { value: 'g1gc_basic', label: 'G1GC básico' },
    { value: 'g1gc_optimized', label: 'G1GC optimizado' },
    { value: 'zgc', label: 'ZGC (mucha RAM)' },
    { value: 'shenandoah', label: 'Shenandoah' },
    { value: 'none', label: 'Desactivado' },
];

const GPU_PREFERENCES = [
    { value: '', label: 'Automático' },
    { value: 'dgpu', label: 'GPU dedicada' },
    { value: 'igpu', label: 'GPU integrada' },
];

const BACKUP_SCHEDULES = [
    { value: 'off', label: 'Desactivados' },
    { value: 'session', label: 'Al cerrar una sesión de juego' },
    { value: 'hours', label: 'Cada X horas' },
    { value: 'days', label: 'Cada X días' },
];

function toInt(s: string): number | undefined {
    const n = Number(s);
    return Number.isFinite(n) && n > 0 ? Math.round(n) : undefined;
}

// El backend guarda la RAM en MB, pero el campo se edita en GB. Los valores
// guardados por versiones antiguas (< 1024) se escribieron con la etiqueta GB
// pero en unidades de MB (bug previo), así que se migran tratándolos como GB.
function ramToGb(mb: number): number {
    if (mb >= 1024) return Math.round((mb / 1024) * 10) / 10;
    return mb;
}

// Divide una línea de argumentos en tokens respetando comillas (misma lógica
// que splitArgs del backend): "a b" se convierte en ["a", "b"] y '"a b" c' en
// ['a b', 'c'].
function splitArgsLocal(s: string): string[] {
    const out: string[] = [];
    let cur = '';
    let inQuote = false;
    for (const ch of s) {
        if (ch === '"') {
            inQuote = !inQuote;
        } else if (/\s/.test(ch) && !inQuote) {
            if (cur) {
                out.push(cur);
                cur = '';
            }
        } else {
            cur += ch;
        }
    }
    if (cur) out.push(cur);
    return out;
}

async function load() {
    if (!props.name) return;
    loading.value = true;
    await loadDetails(props.name);
    const d = detailOf(props.name);
    const c = d?.config ?? {};
    form.maxRam = c.maxRam ? String(ramToGb(c.maxRam)) : '';
    form.useOfficialJava = c.useOfficialJava !== false;
    form.javaExec = c.javaExec ?? '';
    form.fullscreen = !!c.fullscreen;
    form.customResolution = !!c.customResolution;
    form.resWidth = c.resWidth ? String(c.resWidth) : '';
    form.resHeight = c.resHeight ? String(c.resHeight) : '';
    const gc = c.gcPreset && c.gcPreset !== 'auto' ? c.gcPreset : '';
    form.gcEnabled = !!gc;
    form.gcPreset = gc;
    form.gpuEnabled = !!c.gpuPreference;
    form.gpuPreference = c.gpuPreference ?? '';
    form.detailedLogs = !!c.detailedLogs;
    form.javaArgs = (c.javaArgs ?? []).join(' ');
    form.gameArgs = (c.gameArgs ?? []).join(' ');
    form.backupSchedule = c.backupSchedule ?? 'off';
    form.backupInterval = c.backupInterval ? String(c.backupInterval) : '';
    loading.value = false;
    void refreshVerify();
}

async function submit() {
    if (busy.value || !props.name) return;
    busy.value = true;
    msg.value = '';
    msgOk.value = true;
    const cfg: Record<string, unknown> = {};
    const maxRam = toInt(form.maxRam);
    if (maxRam) {
        cfg.minRam = 512;
        cfg.maxRam = maxRam * 1024;
    }
    cfg.useOfficialJava = form.useOfficialJava;
    if (!form.useOfficialJava && form.javaExec.trim()) cfg.javaExec = form.javaExec.trim();
    cfg.fullscreen = form.fullscreen;
    cfg.customResolution = form.customResolution;
    if (form.customResolution) {
        const w = toInt(form.resWidth);
        const h = toInt(form.resHeight);
        if (w) cfg.resWidth = w;
        if (h) cfg.resHeight = h;
    }
    cfg.gcPreset = form.gcEnabled && form.gcPreset ? form.gcPreset : '';
    if (form.gpuEnabled && form.gpuPreference) cfg.gpuPreference = form.gpuPreference;
    cfg.detailedLogs = form.detailedLogs;
    // Los argumentos propios se envían SIEMPRE (aunque vacíos): así se pueden
    // vaciar y volver a heredar los argumentos globales del launcher.
    cfg.javaArgs = splitArgsLocal(form.javaArgs);
    cfg.gameArgs = splitArgsLocal(form.gameArgs);
    cfg.backupSchedule = form.backupSchedule;
    const interval = toInt(form.backupInterval);
    if (interval) cfg.backupInterval = interval;
    const err = await updateConfig(props.name, cfg as any);
    busy.value = false;
    if (err) {
        msg.value = err;
        msgOk.value = false;
        return;
    }
    emit('done', true, 'Configuración de la instancia guardada.');
    emit('update:visible', false);
}

// ---- Verificación de integridad (solo esta instancia; la bloquea durante el
// proceso) ----

const verifyMsg = ref('');
const verifyPoll = ref<number | null>(null);

async function refreshVerify() {
    if (!props.name) return;
    await refreshInstanceVerify(props.name);
    const st = instanceVerifyOf(props.name);
    if (st?.state === 'verifying' && verifyPoll.value === null) {
        verifyPoll.value = window.setInterval(() => void refreshVerify(), 600);
    }
    if (st && st.state !== 'verifying' && verifyPoll.value !== null) {
        window.clearInterval(verifyPoll.value);
        verifyPoll.value = null;
        if (st.state === 'done') {
            verifyMsg.value = st.issues > 0
                ? `Verificación terminada: ${st.found} versiones, ${st.issues} problemas encontrados.`
                : `Verificación terminada: ${st.found} versiones sin problemas.`;
        } else if (st.state === 'cancelled') {
            verifyMsg.value = 'Verificación cancelada.';
        } else if (st.state === 'error') {
            verifyMsg.value = st.error || 'La verificación terminó con error.';
        }
    }
}

async function onVerify() {
    if (!props.name) return;
    verifyMsg.value = '';
    const err = await startInstanceVerify(props.name);
    if (err) {
        verifyMsg.value = err;
        return;
    }
    verifyMsg.value = 'Verificando… la instancia no se puede utilizar hasta que termine.';
}

async function onCancelVerify() {
    if (!props.name) return;
    await cancelInstanceVerify(props.name);
    verifyMsg.value = 'Cancelando…';
}

async function onBackupNow() {
    if (!props.name || busy.value) return;
    busy.value = true;
    const res = await createInstanceBackup(props.name);
    busy.value = false;
    if (!res.ok) {
        msg.value = res.error ?? 'No se pudo crear el backup.';
        msgOk.value = false;
        return;
    }
    msg.value = res.path ? `Backup creado en ${res.path}` : 'Backup creado.';
    msgOk.value = true;
    await loadDetails(props.name);
}

function close() {
    emit('update:visible', false);
}

watch(
    () => props.visible,
    (v) => {
        if (v) {
            msg.value = '';
            msgOk.value = true;
            verifyMsg.value = '';
            void load();
        }
    },
    { immediate: true }
);

onUnmounted(() => {
    if (verifyPoll.value !== null) {
        window.clearInterval(verifyPoll.value);
        verifyPoll.value = null;
    }
});

useOverlayEscape(close, { isActive: () => props.visible, priority: 2 });
</script>

<template>
    <Teleport to="body">
        <Transition name="InstModal">
            <div v-if="visible" class="InstSet_Overlay" @click.self="close">
                <div class="InstSet_Dialog">
                    <div class="InstSet_Head">
                        <div class="InstSet_Titles">
                            <h3>Configuración de la instancia</h3>
                            <span class="InstSet_Name">{{ name }}</span>
                        </div>
                        <button class="InstSet_Close" title="Cerrar" @click="close">
                            <IconX stroke="2" />
                        </button>
                    </div>

                    <div class="InstSet_Body">
                        <p v-if="loading" class="InstSet_Loading">Cargando configuración…</p>
                        <template v-else>
                            <div class="InstSet_Row">
                                <label class="InstSet_Field">
                                    <span>RAM máxima (GB)</span>
                                    <input class="SsIn" type="number" min="0" max="32" v-model="form.maxRam" placeholder="Auto" />
                                </label>
                            </div>

                            <div class="InstSet_SwRow">
                                <span class="InstSet_SwText">Utilizar el Java que tengas configurado</span>
                                <label class="SsTg">
                                    <input type="checkbox" v-model="form.useOfficialJava" />
                                    <span class="SsTgS"></span>
                                </label>
                            </div>
                            <label v-if="!form.useOfficialJava" class="InstSet_Field">
                                <span>Ruta al ejecutable de Java</span>
                                <input class="SsIn" v-model="form.javaExec" placeholder="C:\Program Files\Java\bin\java.exe" autocomplete="off" />
                            </label>

                            <div class="InstSet_SwRow">
                                <span class="InstSet_SwText">Iniciar en pantalla completa</span>
                                <label class="SsTg">
                                    <input type="checkbox" v-model="form.fullscreen" />
                                    <span class="SsTgS"></span>
                                </label>
                            </div>

                            <div class="InstSet_SwRow">
                                <span class="InstSet_SwText">Usar resolución personalizada</span>
                                <label class="SsTg">
                                    <input type="checkbox" v-model="form.customResolution" />
                                    <span class="SsTgS"></span>
                                </label>
                            </div>
                            <div v-if="form.customResolution" class="InstSet_Row">
                                <label class="InstSet_Field">
                                    <span>Ancho</span>
                                    <input class="SsIn" type="number" min="640" v-model="form.resWidth" placeholder="1280" />
                                </label>
                                <label class="InstSet_Field">
                                    <span>Alto</span>
                                    <input class="SsIn" type="number" min="480" v-model="form.resHeight" placeholder="720" />
                                </label>
                            </div>

                            <div class="InstSet_SwRow">
                                <span class="InstSet_SwText">
                                    Usar recolector de basura personalizado
                                    <small class="InstSet_SwSub">Aplica un GC distinto al automático al lanzar</small>
                                </span>
                                <label class="SsTg">
                                    <input type="checkbox" v-model="form.gcEnabled" />
                                    <span class="SsTgS"></span>
                                </label>
                            </div>
                            <label v-if="form.gcEnabled" class="InstSet_Field">
                                <span>Recolector de basura</span>
                                <select class="SsSel" v-model="form.gcPreset">
                                    <option v-for="g in GC_PRESETS" :key="g.value" :value="g.value">{{ g.label }}</option>
                                </select>
                            </label>

                            <div class="InstSet_SwRow">
                                <span class="InstSet_SwText">
                                    Logs detallados
                                    <small class="InstSet_SwSub">Fuerza el nivel debug del juego para un registro más verboso</small>
                                </span>
                                <label class="SsTg">
                                    <input type="checkbox" v-model="form.detailedLogs" />
                                    <span class="SsTgS"></span>
                                </label>
                            </div>

                            <div class="InstSet_SwRow">
                                <span class="InstSet_SwText">
                                    Argumentos de lanzamiento propios
                                    <small class="InstSet_SwSub">Sustituyen a los globales del launcher; vacíos para heredarlos</small>
                                </span>
                            </div>
                            <label class="InstSet_Field">
                                <span>Argumentos JVM</span>
                                <textarea class="SsIn InstSet_Args" v-model="form.javaArgs" rows="2" placeholder="-XX:+UseConcMarkSweepGC -Dminecraft.client.remix=false" autocomplete="off" spellcheck="false" />
                            </label>
                            <label class="InstSet_Field">
                                <span>Argumentos del juego</span>
                                <textarea class="SsIn InstSet_Args" v-model="form.gameArgs" rows="2" placeholder="--server mc.ejemplo.com --port 25565" autocomplete="off" spellcheck="false" />
                            </label>

                            <div class="InstSet_SwRow">
                                <span class="InstSet_SwText">
                                    Backups automáticos
                                    <small class="InstSet_SwSub">Copia comprimida de la instancia en <code>instances/backups/&lt;nombre&gt;.zip</code></small>
                                </span>
                            </div>
                            <label class="InstSet_Field">
                                <span>Frecuencia</span>
                                <select class="SsSel" v-model="form.backupSchedule">
                                    <option v-for="b in BACKUP_SCHEDULES" :key="b.value" :value="b.value">{{ b.label }}</option>
                                </select>
                            </label>
                            <label v-if="form.backupSchedule === 'hours' || form.backupSchedule === 'days'" class="InstSet_Field">
                                <span>Cada {{ form.backupSchedule === 'hours' ? 'X horas' : 'X días' }}</span>
                                <input class="SsIn" type="number" min="1" max="8760" v-model="form.backupInterval" :placeholder="form.backupSchedule === 'hours' ? '6' : '1'" />
                            </label>

                            <div class="InstSet_Verify">
                                <span class="InstSet_VerifyText">
                                    Verificar integridad
                                    <small class="InstSet_SwSub">Comprueba versiones, librerías y nativos de ESTA instancia. Mientras dure, la instancia no se puede utilizar.</small>
                                </span>
                                <button
                                    class="SsBtn InstSet_VerifyBtn"
                                    :disabled="busy || isInstanceVerifying(props.name)"
                                    @click="onVerify"
                                >
                                    {{ isInstanceVerifying(props.name) ? 'Verificando…' : 'Verificar' }}
                                </button>
                            </div>
                            <div v-if="isInstanceVerifying(props.name)" class="InstSet_VerifyProgress">
                                <div class="InstSet_VerifyBar">
                                    <span class="InstSet_VerifyBarFill" :style="{ width: (instanceVerifyOf(props.name)?.percent ?? 0) + '%' }"></span>
                                </div>
                                <span class="InstSet_VerifyState">
                                    {{ instanceVerifyOf(props.name)?.version || 'Preparando…' }} · {{ instanceVerifyOf(props.name)?.percent ?? 0 }}%
                                </span>
                                <button class="SsBtn InstSet_VerifyCancel" @click="onCancelVerify">Cancelar</button>
                            </div>
                            <p v-if="verifyMsg" class="InstSet_Msg">{{ verifyMsg }}</p>

                            <div class="InstSet_Verify">
                                <span class="InstSet_VerifyText">
                                    Backup manual
                                    <small class="InstSet_SwSub">Genera ahora el zip de la instancia en <code>instances/backups/</code></small>
                                </span>
                                <button class="SsBtn" :disabled="busy || isInstanceVerifying(props.name)" @click="onBackupNow">
                                    Crear backup ahora
                                </button>
                            </div>

                            <div class="InstSet_SwRow">
                                <span class="InstSet_SwText">
                                    Poner una preferencia de GPU
                                    <small class="InstSet_SwSub">Forzará la gráfica dedicada o integrada al lanzar</small>
                                </span>
                                <label class="SsTg">
                                    <input type="checkbox" v-model="form.gpuEnabled" />
                                    <span class="SsTgS"></span>
                                </label>
                            </div>
                            <label v-if="form.gpuEnabled" class="InstSet_Field">
                                <span>Preferencia de GPU</span>
                                <select class="SsSel" v-model="form.gpuPreference">
                                    <option v-for="g in GPU_PREFERENCES" :key="g.value" :value="g.value">{{ g.label }}</option>
                                </select>
                            </label>

                            <p v-if="msg" :class="['InstSet_Msg', { error: !msgOk }]">{{ msg }}</p>
                        </template>
                    </div>

                    <div class="InstSet_Footer">
                        <button class="SsBtn" :disabled="busy" @click="close">Cancelar</button>
                        <button class="SsBtn SsBtnPrimary InstSet_Submit" :disabled="busy || loading" @click="submit">
                            {{ busy ? 'Guardando…' : 'Guardar' }}
                        </button>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use './Styles/Settings.scss';
</style>