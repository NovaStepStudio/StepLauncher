import type { BootstrapTask } from '../Types';
import { loadVersions, loadProfiles, installedVersions, profiles } from '@/Launcher/Store';

export const versionsTask: BootstrapTask = {
    id: 'versions',
    label: 'Cargando versiones y perfiles',
    weight: 1.5,
    async run(ctx) {
        const t0 = performance.now();
        try {
            ctx.log('info', 'Leyendo versiones instaladas');
            await loadVersions();
            ctx.log('success', `Versiones ${installedVersions.value.length} instaladas`);
            ctx.setStepLabel('Cargando perfiles');
            ctx.log('info', 'Cargando perfiles del launcher');
            await loadProfiles();
            const pCount = Object.keys(profiles.value ?? {}).length;
            ctx.log('success', pCount ? `Perfiles ${pCount} cargados` : 'Sin perfiles');
            ctx.log('debug', `Versiones y perfiles en ${Math.round(performance.now() - t0)}ms`);
        } catch (e: any) {
            ctx.log('warn', `Versiones y perfiles ${e?.message ?? e}`);
        }
    },
};
