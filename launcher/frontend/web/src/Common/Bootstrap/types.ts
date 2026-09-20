/**
 * Tipos centrales del sistema de Bootstrap.
 * Cada fase del arranque es una tarea trazable con peso para el progreso.
 */

export type BootstrapLogLevel = 'info' | 'success' | 'warn' | 'error' | 'debug';

export type BootstrapStepStatus = 'pending' | 'running' | 'done' | 'error' | 'skipped';

export interface BootstrapStep {
    id: string;
    label: string;
    description?: string;
    /** Peso relativo para el cálculo de progreso (default 1). */
    weight: number;
    status: BootstrapStepStatus;
    startedAt?: number;
    finishedAt?: number;
    durationMs?: number;
    error?: string;
}

export interface BootstrapLog {
    id: number;
    time: string;
    level: BootstrapLogLevel;
    stepId: string;
    message: string;
}

export type BootstrapStatus = 'idle' | 'running' | 'done' | 'error';

export interface BootstrapState {
    status: BootstrapStatus;
    currentId: string | null;
    currentLabel: string;
    progress: number; // 0-100
    steps: BootstrapStep[];
    logs: BootstrapLog[];
    startedAt: number | null;
    finishedAt: number | null;
    showWelcome: boolean;
    error: string | null;
    /** Ruta de config cargada, versión, etc para mostrar en splash. */
    meta: {
        version?: string;
        workDir?: string;
        java?: string;
    };
}

export interface BootstrapContext {
    /** Registra un log asociado al paso actual. */
    log: (level: BootstrapLogLevel, message: string) => void;
    /** Actualiza el label del paso actual (para sub-fases). */
    setStepLabel: (label: string) => void;
    /** Estado mutable compartido entre tareas. */
    state: BootstrapState;
    /** Datos ya cargados que pueden compartir tareas. */
    shared: {
        config?: any;
        launcherAssets?: any;
        firstLaunch?: boolean;
    };
}

export interface BootstrapTask {
    id: string;
    label: string;
    description?: string;
    weight?: number;
    /** Si falla y critical=true, el bootstrap entero entra en error. */
    critical?: boolean;
    /** Tiempo máximo en ms (0 = sin límite). */
    timeoutMs?: number;
    run: (ctx: BootstrapContext) => Promise<void>;
}

export interface BootstrapOptions {
    minSplashMs?: number;
    maxSplashMs?: number;
}
