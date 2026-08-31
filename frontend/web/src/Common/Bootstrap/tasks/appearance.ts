import type { BootstrapTask } from '../Types';
import { loadLocal, personalization } from '@/Common/Stores/Ui';

export const appearanceTask: BootstrapTask = {
    id: 'appearance',
    label: 'Aplicando personalizacion',
    weight: 1.5,
    async run(ctx) {
        ctx.log('info', 'Resolviendo tema y fondos');
        const bg = personalization.value?.background;
        if (!bg || bg.type === 'none') {
            ctx.log('debug', 'Sin fondo configurado');
            return;
        }
        ctx.log('info', `Fondo tipo ${bg.type}`);
        try {
            if (bg.type === 'image' && bg.imagePath) {
                ctx.setStepLabel('Cargando imagen de fondo');
                const url = await loadLocal(bg.imagePath);
                ctx.log(url ? 'success' : 'warn', url ? `Imagen cargada ${bg.imagePath}` : `Imagen no encontrada ${bg.imagePath}`);
            } else if (bg.type === 'video' && bg.videoPath) {
                ctx.setStepLabel('Cargando video de fondo');
                const url = await loadLocal(bg.videoPath);
                ctx.log(url ? 'success' : 'warn', url ? `Video cargado ${bg.videoPath}` : `Video no encontrado ${bg.videoPath}`);
            } else if (bg.type === 'dynamic' && bg.dynamicImages?.length) {
                ctx.setStepLabel(`Cargando ${bg.dynamicImages.length} fondos dinamicos`);
                const results = await Promise.all(bg.dynamicImages.map((p: string) => loadLocal(p)));
                const ok = results.filter(Boolean).length;
                ctx.log('info', `Fondos dinamicos ${ok} de ${bg.dynamicImages.length} cargados`);
            }
        } catch (e: any) {
            ctx.log('warn', `Fondos ${e?.message ?? e}`);
        }

        const enabled = personalization.value?.backgroundMusic?.enabled;
        ctx.log('debug', enabled ? 'Musica de fondo habilitada' : 'Musica de fondo desactivada');
    },
};

export const backgroundTask: BootstrapTask = {
    id: 'background',
    label: 'Preparando fondos',
    weight: 1,
    async run(ctx) {
        ctx.log('debug', 'Fondos en segundo plano listos');
        await new Promise((r) => setTimeout(r, 80));
    },
};
