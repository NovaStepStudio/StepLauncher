import type { BootstrapTask } from '../types';
import { Events } from '@wailsio/runtime';
import { refreshAfterDownload, onGameCrash, hideLaunchMessage, maybeShowWindow } from '@/Launcher/Store';

let bound = false;

export const eventsTask: BootstrapTask = {
    id: 'events',
    label: 'Conectando eventos',
    weight: 1,
    async run(ctx) {
        if (bound) {
            ctx.log('debug', 'Eventos ya vinculados, omitiendo');
            return;
        }
        bound = true;
        ctx.log('info', 'Suscribiendo eventos de descarga');
        Events.On('download_state', ({ data: raw }: any) => {
            try {
                const s = typeof raw === 'string' ? raw : JSON.stringify(raw ?? '');
                const obj = JSON.parse(s) as { data?: { state?: string } };
                if (obj?.data?.state === 'completed') void refreshAfterDownload();
            } catch (_e) {}
        });
        ctx.log('success', 'Eventos de descarga listos');

        ctx.log('info', 'Suscribiendo eventos de juego');
        Events.On('game_started', async () => {
            hideLaunchMessage();
        });
        Events.On('game_crashed', async ({ data }: any) => {
            onGameCrash(data);
            void maybeShowWindow();
            checkShotsLog(ctx);
        });
        Events.On('game_exited', () => checkShotsLog(ctx));
        Events.On('game_stopped', () => checkShotsLog(ctx));
        ctx.log('success', 'Eventos de juego listos');

        ctx.log('info', 'Suscribiendo eventos de bandeja y cuentas');
        Events.On('tray_open_settings', () => {
            import('@/Common/Overlays/Store').then(({ settingsOpen }) => { settingsOpen.value = true; });
        });
        Events.On('tray_open_instances', () => {
            import('@/Common/Overlays/Store').then(({ openHeavyPanel }) => openHeavyPanel('instances'));
        });
        Events.On('tray_open_downloads', () => {
            import('@/Common/Overlays/Store').then(({ closeHeavyPanel, installOpen }) => {
                closeHeavyPanel('shots'); closeHeavyPanel('instances'); closeHeavyPanel('mods');
                installOpen.value = true;
            });
        });
        Events.On('tray_open_music', () => {
            import('@/Common/Overlays/Store').then(({ openHeavyPanel }) => openHeavyPanel('music'));
        });
        // Compat: recargar interfaz pedida desde el tray también puede llegar como evento
        Events.On('tray_reload', () => {
            try { window.location.reload(); } catch (_e) {}
        });
        // Cuentas
        Events.On('account_login', async () => {
            const { loadAccounts } = await import('@/Accounts/Store');
            void loadAccounts();
        });
        Events.On('account_refresh', async () => {
            const { loadAccounts } = await import('@/Accounts/Store');
            void loadAccounts();
        });
        Events.On('account_refresh_all', async () => {
            const { loadAccounts } = await import('@/Accounts/Store');
            void loadAccounts();
        });
        ctx.log('success', 'Eventos de bandeja y cuentas listos');
    },
};

function checkShotsLog(ctx: any) {
    ctx.log('debug', 'Juego cerrado revisando capturas');
}
