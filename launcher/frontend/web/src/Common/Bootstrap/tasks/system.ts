import type { BootstrapTask } from '../types';
import { GetConfig } from '@wailsjs/StepLauncher/internal/Services/Config/configservice';
import { startIdleTracking } from '@/Common/Stores/Idle';
import { bindUpdateEvents, checkForUpdates } from '@/Updates/Store';
import { bindNewsEvents } from '@/News/Store';
import { isOffline } from '@/Common/Stores/Connectivity';
import { applyPersonalization } from '@/Common/Stores/Ui';

export const systemTask: BootstrapTask = {
    id: 'system',
    label: 'Inicializando subsistemas',
    description: 'Idle tracking verificacion de config updater y noticias',
    weight: 1,
    async run(ctx) {
        const cfg = ctx.shared.config ?? await GetConfig().catch(() => null);
        if (!ctx.shared.config && cfg) ctx.shared.config = cfg;

        try {
            const idle = cfg?.idle ?? {};
            ctx.log('debug', `Idle autoClose ${idle.autoCloseModals ?? true} cada ${idle.idleMinutes ?? 1} minutos`);
            startIdleTracking(
                {
                    autoCloseModals: idle.autoCloseModals ?? true,
                    idleMinutes: idle.idleMinutes ?? 1,
                    configCheckEnabled: idle.configCheckEnabled ?? true,
                    configCheckMinutes: idle.configCheckMinutes ?? 3,
                },
                () => {
                    window.dispatchEvent(new CustomEvent('sl:close-overlays'));
                    ctx.log('debug', 'Idle overlays cerrados');
                },
                async () => {
                    try {
                        const fresh = await GetConfig();
                        if (fresh?.personalization) {
                            applyPersonalization(fresh.personalization as any);
                        }
                    } catch (_e) {}
                }
            );
            ctx.log('success', 'Sistema de inactividad activo');
        } catch (e: any) {
            ctx.log('warn', `Idle no iniciado ${e?.message ?? e}`);
        }

        try {
            bindUpdateEvents();
            ctx.log('debug', 'Updater vinculado');
            bindNewsEvents();
            ctx.log('debug', 'Noticias vinculadas');
        } catch (e: any) {
            ctx.log('warn', `Eventos de sistema ${e?.message ?? e}`);
        }

        const shouldCheck = cfg?.launcher?.checkForUpdatesOnStart;
        if (shouldCheck) {
            if (isOffline.value || (typeof navigator !== 'undefined' && !navigator.onLine)) {
                ctx.log('debug', 'Check de updates omitido sin conexión');
            } else {
                ctx.log('info', 'Comprobando actualizaciones en segundo plano');
                checkForUpdates(true).catch(() => {});
            }
        } else {
            ctx.log('debug', 'Check de updates omitido desactivado');
        }
    },
};
