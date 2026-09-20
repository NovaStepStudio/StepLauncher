<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted } from 'vue';
import { bootstrapState } from './state';

const showLogs = ref(false);
const logsRef = ref<HTMLDivElement | null>(null);

const progress = computed(() => bootstrapState.value.progress);
const currentLabel = computed(() => bootstrapState.value.currentLabel);
const steps = computed(() => bootstrapState.value.steps);
const logs = computed(() => bootstrapState.value.logs);
const status = computed(() => bootstrapState.value.status);
const error = computed(() => bootstrapState.value.error);

const elapsed = ref(0);
let elapsedTimer: number | null = null;

onMounted(() => {
    elapsedTimer = window.setInterval(() => {
        if (bootstrapState.value.startedAt) {
            elapsed.value = Math.round((Date.now() - bootstrapState.value.startedAt) / 100) / 10;
        }
    }, 100);
});

watch(logs, async () => {
    if (!showLogs.value) return;
    await nextTick();
    if (logsRef.value) logsRef.value.scrollTop = logsRef.value.scrollHeight;
});

const visibleLogs = computed(() => logs.value.slice(-80));

function toggleLogs() {
    showLogs.value = !showLogs.value;
}

function levelClass(l: string) {
    switch (l) {
        case 'success': return 'log-success';
        case 'error': return 'log-error';
        case 'warn': return 'log-warn';
        case 'debug': return 'log-debug';
        default: return 'log-info';
    }
}

const visible = computed(() => bootstrapState.value.status !== 'done');
</script>

<template>
    <Transition name="SplashFade">
        <div v-if="visible" class="BootstrapSplash">
            <div class="BootstrapSplash_Glow" />

            <div class="BootstrapSplash_Center">
                <img class="BootstrapSplash_Logo" src="../../../assets/logo-step.png" alt="StepLauncher" draggable="false" />
                <div class="BootstrapSplash_Title">StepLauncher</div>
                <div class="BootstrapSplash_Subtitle">Minecraft Launcher — NovaStepStudio</div>

                <div class="BootstrapSplash_StepLabel">{{ currentLabel }}</div>

                <div class="BootstrapSplash_ProgressWrap">
                    <div class="BootstrapSplash_Bar">
                        <div class="BootstrapSplash_Fill" :style="{ width: progress + '%' }" />
                        <div class="BootstrapSplash_Shine" :style="{ left: progress + '%' }" />
                    </div>
                    <div class="BootstrapSplash_ProgressText">{{ progress }}%</div>
                </div>

                <div class="BootstrapSplash_Steps">
                    <div
                        v-for="s in steps"
                        :key="s.id"
                        class="StepDot"
                        :class="{
                            done: s.status === 'done',
                            running: s.status === 'running',
                            error: s.status === 'error',
                            pending: s.status === 'pending',
                        }"
                        :title="s.label"
                    >
                        <span class="dot" />
                        <span class="label">{{ s.label }}</span>
                        <span v-if="s.durationMs" class="dur">{{ s.durationMs }}ms</span>
                    </div>
                </div>

                <div v-if="error" class="BootstrapSplash_Error">
                    <span>⚠ {{ error }}</span>
                </div>
            </div>

            <div class="BootstrapSplash_Bottom">
                <div class="BootstrapSplash_BottomLeft">
                    <span class="elapsed">{{ elapsed.toFixed(1) }}s</span>
                    <span class="sep">·</span>
                    <span class="status">{{ status === 'error' ? 'Error' : status === 'running' ? 'Iniciando…' : 'Listo' }}</span>
                    <span v-if="logs.length" class="sep">·</span>
                    <button v-if="logs.length" class="LogsToggle" @click="toggleLogs">
                        {{ showLogs ? 'Ocultar logs' : `Ver logs (${logs.length})` }}
                    </button>
                </div>
                <div class="BootstrapSplash_BottomRight">
                    <img class="BootstrapSplash_Loader" src="../../../assets/gif/chicken_jockey_run.gif" alt="" />
                </div>
            </div>

            <Transition name="LogsFade">
                <div v-if="showLogs" ref="logsRef" class="BootstrapSplash_Logs">
                    <div v-for="l in visibleLogs" :key="l.id" class="LogLine" :class="levelClass(l.level)">
                        <span class="time">{{ l.time }}</span>
                        <span class="lvl">[{{ l.level }}]</span>
                        <span class="step">{{ l.stepId }}</span>
                        <span class="msg">{{ l.message }}</span>
                    </div>
                    <div v-if="!logs.length" class="LogEmpty">Sin logs aún…</div>
                </div>
            </Transition>
        </div>
    </Transition>
</template>

<style scoped lang="scss">
@use './Styles/SplashScreen.scss';
</style>
