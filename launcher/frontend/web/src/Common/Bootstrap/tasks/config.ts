import type { BootstrapTask } from '../types';
import { GetConfig } from '@wailsjs/StepLauncher/internal/Services/Config/configservice';
import { GetLauncherAssets, UpdatePersonalization } from '@wailsjs/StepLauncher/internal/Services/Appearance/appearanceservice';
import { setUIScale, applyPersonalization } from '@/Common/Stores/Ui';
import { ensureCustomFonts, fontByType, isBuiltinFont, cleanFontName } from '@/Common/Stores/Fonts';

export const configTask: BootstrapTask = {
    id: 'config',
    label: 'Cargando configuracion',
    description: 'Lee la config del launcher, aplica escalado y personalizacion, y cura fuentes.',
    weight: 2,
    critical: true,
    async run(ctx) {
        ctx.log('info', 'Leyendo configuracion persistida');
        let cfg: any = null;
        try {
            cfg = await GetConfig();
            ctx.shared.config = cfg;
            if (cfg?.launcher?.version || cfg?.launcherVersion) {
                ctx.state.meta.version = cfg?.launcherVersion ?? cfg?.launcher?.version ?? '';
            }
            ctx.log('success', `Config cargada trabajo ${cfg?.workDir ?? 'desconocido'}`);
        } catch (e: any) {
            ctx.log('warn', `No se pudo leer la config ${e?.message ?? e}`);
            cfg = ctx.shared.config ?? {};
        }

        const rawScale = cfg?.personalization?.uiScale;
        if (typeof rawScale === 'number' && rawScale >= 50 && rawScale <= 200) {
            setUIScale(rawScale);
            ctx.log('debug', `Escalado UI ${rawScale} por ciento`);
        }

        if (cfg?.personalization) {
            applyPersonalization(cfg.personalization as any);
            ctx.log('info', 'Tema y colores aplicados');
        }

        try {
            ctx.log('debug', 'Verificando fuentes personalizadas');
            const assets: any = await GetLauncherAssets();
            ctx.shared.launcherAssets = assets;
            await ensureCustomFonts(assets);
            const p = cfg?.personalization;
            if (p && Array.isArray(assets?.fonts) && assets.fonts.length) {
                const eP = fontByType(assets, 'primary');
                const eS = fontByType(assets, 'secundary');
                const pName = eP ? ((eP.name ?? '').trim() || (eP.path ? cleanFontName(eP.path) : '')) : '';
                const sName = eS ? ((eS.name ?? '').trim() || (eS.path ? cleanFontName(eS.path) : '')) : '';
                const next = { ...p };
                let changed = false;
                if (pName && !isBuiltinFont(p.fontPrimary) && p.fontPrimary !== pName) {
                    next.fontPrimary = pName; changed = true;
                    ctx.log('info', `Fuente primaria corregida a ${pName}`);
                }
                if (sName && !isBuiltinFont(p.fontSecondary) && p.fontSecondary !== sName) {
                    next.fontSecondary = sName; changed = true;
                    ctx.log('info', `Fuente secundaria corregida a ${sName}`);
                }
                if (changed) {
                    applyPersonalization(next);
                    await UpdatePersonalization(next);
                    ctx.log('success', 'Personalizacion de fuentes sincronizada');
                } else {
                    ctx.log('debug', 'Fuentes verificadas sin cambios');
                }
            }
        } catch (e: any) {
            ctx.log('warn', `Fuentes ${e?.message ?? 'error silencioso'}`);
        }

        ctx.log('success', 'Configuracion lista');
    },
};
