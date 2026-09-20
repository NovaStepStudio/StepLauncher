import { ref, computed } from 'vue';
import type { BootstrapState, BootstrapStep, BootstrapLog } from './types';
import { createLogger, trimLogs } from './logger';

const DEFAULT_STEPS: Omit<BootstrapStep, 'status'>[] = [
    { id: 'welcome', label: 'Comprobando primer arranque', weight: 0.5 },
    { id: 'config', label: 'Cargando configuración', weight: 2 },
    { id: 'appearance', label: 'Aplicando personalización', weight: 1.5 },
    { id: 'system', label: 'Inicializando subsistemas', weight: 1 },
    { id: 'accounts', label: 'Sincronizando cuentas', weight: 1.5 },
    { id: 'versions', label: 'Cargando versiones y perfiles', weight: 1.5 },
    { id: 'events', label: 'Conectando eventos', weight: 1 },
    { id: 'background', label: 'Preparando fondos', weight: 1 },
];

function initialSteps(): BootstrapStep[] {
    return DEFAULT_STEPS.map((s) => ({ ...s, status: 'pending' as const }));
}

export const bootstrapState = ref<BootstrapState>({
    status: 'idle',
    currentId: null,
    currentLabel: 'Preparando StepLauncher…',
    progress: 0,
    steps: initialSteps(),
    logs: [],
    startedAt: null,
    finishedAt: null,
    showWelcome: false,
    error: null,
    meta: {},
});

export const bootstrapVisible = computed(() => bootstrapState.value.status !== 'done');

export const bootstrapProgress = computed(() => bootstrapState.value.progress);

const logger = createLogger((entry) => {
    const s = bootstrapState.value;
    s.logs = trimLogs([...s.logs, entry]);
});

export function log(level: BootstrapLog['level'], message: string, stepId?: string) {
    const current = stepId ?? bootstrapState.value.currentId ?? 'bootstrap';
    logger.log(level, current, message);
}

export function resetBootstrap() {
    bootstrapState.value = {
        status: 'idle',
        currentId: null,
        currentLabel: 'Preparando StepLauncher…',
        progress: 0,
        steps: initialSteps(),
        logs: [],
        startedAt: null,
        finishedAt: null,
        showWelcome: false,
        error: null,
        meta: {},
    };
}

export function setStepStatus(id: string, status: BootstrapStep['status'], error?: string) {
    const step = bootstrapState.value.steps.find((s) => s.id === id);
    if (!step) return;
    step.status = status;
    if (error) step.error = error;
    if (status === 'running') step.startedAt = Date.now();
    if (status === 'done' || status === 'error' || status === 'skipped') {
        step.finishedAt = Date.now();
        if (step.startedAt) step.durationMs = step.finishedAt - step.startedAt;
    }
}

export function setCurrentStep(id: string, label?: string) {
    bootstrapState.value.currentId = id;
    if (label) bootstrapState.value.currentLabel = label;
    else {
        const step = bootstrapState.value.steps.find((s) => s.id === id);
        if (step) bootstrapState.value.currentLabel = step.label;
    }
}

export function updateProgress() {
    const steps = bootstrapState.value.steps;
    const total = steps.reduce((a, s) => a + s.weight, 0);
    const done = steps.reduce((a, s) => {
        if (s.status === 'done' || s.status === 'skipped') return a + s.weight;
        if (s.status === 'running') return a + s.weight * 0.35;
        return a;
    }, 0);
    bootstrapState.value.progress = Math.min(99, Math.round((done / total) * 100));
}

export function finishBootstrap() {
    bootstrapState.value.progress = 100;
    bootstrapState.value.status = 'done';
    bootstrapState.value.finishedAt = Date.now();
    bootstrapState.value.currentLabel = '¡Listo!';
}

export function failBootstrap(error: string) {
    bootstrapState.value.status = 'error';
    bootstrapState.value.error = error;
    bootstrapState.value.finishedAt = Date.now();
}

export function useBootstrap() {
    return {
        state: bootstrapState,
        visible: bootstrapVisible,
        progress: bootstrapProgress,
        log,
        reset: resetBootstrap,
    };
}
