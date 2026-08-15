<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import { IconX } from '@tabler/icons-vue';
import { loadDetails, detailOf, updateConfig } from './Store';
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
});

const GC_PRESETS = [
    { value: 'g1gc_basic', label: 'G1GC básico' },
    { value: 'g1gc_optimized', label: 'G1GC optimizado' },
    { value: 'zgc', label: 'ZGC (mucha RAM)' },
    { value: 'shenandoah', label: 'Shenandoah' },
];

const GPU_PREFERENCES = [
    { value: '', label: 'Automático' },
    { value: 'dgpu', label: 'GPU dedicada' },
    { value: 'igpu', label: 'GPU integrada' },
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
    loading.value = false;
}

watch(
    () => props.visible,
    (v) => {
        if (v) {
            msg.value = '';
            msgOk.value = true;
            void load();
        }
    },
    { immediate: true }
);

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

function close() {
    emit('update:visible', false);
}

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