import type { BootstrapTask, BootstrapOptions, BootstrapContext } from './types';
import {
    bootstrapState,
    setStepStatus,
    setCurrentStep,
    updateProgress,
    finishBootstrap,
    failBootstrap,
    log,
} from './state';
import { welcomeTask } from './tasks/welcome';
import { configTask } from './tasks/config';
import { appearanceTask, backgroundTask } from './tasks/appearance';
import { systemTask } from './tasks/system';
import { accountsTask } from './tasks/accounts';
import { versionsTask } from './tasks/versions';
import { eventsTask } from './tasks/events';

const ALL_TASKS: BootstrapTask[] = [
    welcomeTask,
    configTask,
    appearanceTask,
    systemTask,
    accountsTask,
    versionsTask,
    eventsTask,
    backgroundTask,
];

export function getBootstrapTasks(): BootstrapTask[] {
    return ALL_TASKS;
}

function withTimeout<T>(p: Promise<T>, ms: number, label: string): Promise<T> {
    if (!ms) return p;
    return new Promise((resolve, reject) => {
        const t = window.setTimeout(() => reject(new Error(`Timeout ${ms}ms en ${label}`)), ms);
        p.then((v) => { clearTimeout(t); resolve(v); }, (e) => { clearTimeout(t); reject(e); });
    });
}

export async function runBootstrap(opts: BootstrapOptions = {}): Promise<void> {
    const minMs = opts.minSplashMs ?? 900;
    const maxMs = opts.maxSplashMs ?? 6500;
    const started = Date.now();
    bootstrapState.value.status = 'running';
    bootstrapState.value.startedAt = started;
    bootstrapState.value.progress = 2;
    bootstrapState.value.error = null;
    log('info', `Iniciando StepLauncher bootstrap minimo ${minMs}ms maximo ${maxMs}ms`);

    let failsafe: number | null = window.setTimeout(() => {
        log('warn', 'Failsafe de splash alcanzado forzando cierre');
        // No aborta tareas, solo asegura que el splash no quede colgado.
        if (bootstrapState.value.status === 'running') {
            finishBootstrap();
        }
    }, maxMs);

    const ctx: BootstrapContext = {
        log(level, message) {
            const cur = bootstrapState.value.currentId ?? 'bootstrap';
            log(level, message, cur);
        },
        setStepLabel(label) {
            if (bootstrapState.value.currentId) {
                const step = bootstrapState.value.steps.find((s) => s.id === bootstrapState.value.currentId);
                if (step) bootstrapState.value.currentLabel = label;
            }
        },
        state: bootstrapState.value,
        shared: {},
    };

    for (const task of ALL_TASKS) {
        const stepId = task.id;
        setCurrentStep(stepId, task.label);
        setStepStatus(stepId, 'running');
        updateProgress();
        ctx.log('info', `Iniciando ${task.label}`);
        const t0 = performance.now();
        try {
            const p = task.run({ ...ctx, log: (lvl, msg) => log(lvl, msg, stepId), setStepLabel: ctx.setStepLabel, state: bootstrapState.value, shared: ctx.shared });
            await withTimeout(p, task.timeoutMs ?? 0, task.label);
            const dur = Math.round(performance.now() - t0);
            setStepStatus(stepId, 'done');
            log('success', `${task.label} completado en ${dur}ms`, stepId);
        } catch (e: any) {
            const msg = e?.message ?? String(e);
            setStepStatus(stepId, 'error', msg);
            log('error', `${task.label} error ${msg}`, stepId);
            if (task.critical) {
                failBootstrap(msg);
                if (failsafe !== null) { clearTimeout(failsafe); failsafe = null; }
                throw e;
            }
            // No crítico: se continúa con el resto
        }
        updateProgress();
        // Pausa breve para que el splash sea legible en tareas muy rápidas
        await new Promise((r) => setTimeout(r, 40));
    }

    // Asegura tiempo minimo de splash para evitar flash
    const elapsed = Date.now() - started;
    const wait = Math.max(0, minMs - elapsed);
    if (wait > 0) {
        log('debug', `Esperando ${wait}ms para cumplir splash minimo`, 'bootstrap');
        await new Promise((r) => setTimeout(r, wait));
    }

    if (failsafe !== null) { clearTimeout(failsafe); failsafe = null; }
    const totalMs = Date.now() - started;
    const failed = bootstrapState.value.steps.filter((s) => s.status === 'error').length;
    finishBootstrap();
    log('success', `Bootstrap completado en ${totalMs}ms`, 'bootstrap');
    // Registra en los logs del launcher el tiempo total y los pasos fallidos
    try {
        const { BootstrapComplete } = await import('@wailsjs/StepLauncher/internal/Services/System/systemservice');
        await BootstrapComplete(totalMs, failed);
    } catch (_e) {}

    // Deja visible el splash 120ms extra para animacion
    await new Promise((r) => setTimeout(r, 120));
}
