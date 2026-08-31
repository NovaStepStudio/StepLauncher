import type { BootstrapLog, BootstrapLogLevel } from './Types';

let seq = 0;
const MAX_LOGS = 180;

export function createLogger(
    onLog: (entry: BootstrapLog) => void
) {
    return {
        log(level: BootstrapLogLevel, stepId: string, message: string) {
            const entry: BootstrapLog = {
                id: ++seq,
                time: new Date().toLocaleTimeString('es-ES', { hour12: false }),
                level,
                stepId,
                message,
            };
            onLog(entry);
            const prefix = `[Bootstrap:${stepId}]`;
            switch (level) {
                case 'error': console.error(prefix, message); break;
                case 'warn': console.warn(prefix, message); break;
                case 'debug': console.debug(prefix, message); break;
                default: console.log(prefix, message); break;
            }
            // Persiste en los logs del launcher (Go) sin bloquear la UI
            try {
                import('@wailsjs/StepLauncher/internal/Services/System/systemservice').then((m) => {
                    if (typeof m.BootstrapLog === 'function') {
                        m.BootstrapLog(level, `[${stepId}] ${message}`).catch(() => {});
                    }
                }).catch(() => {});
            } catch (_e) {}
        },
    };
}

export function trimLogs(logs: BootstrapLog[]): BootstrapLog[] {
    if (logs.length <= MAX_LOGS) return logs;
    return logs.slice(logs.length - MAX_LOGS);
}
