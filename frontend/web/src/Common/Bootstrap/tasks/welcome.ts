import type { BootstrapTask } from '../Types';
import { GetFirstLaunch } from '@wailsjs/StepLauncher/internal/Services/System/systemservice';

export const welcomeTask: BootstrapTask = {
    id: 'welcome',
    label: 'Comprobando bienvenida',
    weight: 0.5,
    async run(ctx) {
        try {
            const first = await GetFirstLaunch();
            ctx.shared.firstLaunch = first === true;
            ctx.state.showWelcome = first === true;
            if (first) ctx.log('info', 'Primer arranque detectado se mostrara bienvenida');
            else ctx.log('debug', 'Arranque normal');
        } catch (e: any) {
            ctx.log('warn', `No se pudo verificar primer arranque ${e?.message ?? e}`);
            ctx.state.showWelcome = false;
        }
    },
};
