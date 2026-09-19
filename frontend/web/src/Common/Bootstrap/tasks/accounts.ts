import type { BootstrapTask } from '../types';
import { loadAccounts, autoRefresh, refreshAllAccounts } from '@/Accounts/Store';

export const accountsTask: BootstrapTask = {
    id: 'accounts',
    label: 'Sincronizando cuentas',
    weight: 1.5,
    async run(ctx) {
        const t0 = performance.now();
        try {
            ctx.log('info', 'Cargando cuentas guardadas');
            await loadAccounts();
            const { accounts } = await import('@/Accounts/Store');
            const count = accounts.value.length;
            ctx.log('success', count ? `Cuentas ${count} encontradas` : 'Sin cuentas guardadas');
            if (autoRefresh.value) {
                ctx.setStepLabel('Refrescando sesiones');
                ctx.log('info', 'Auto refresh de sesiones activo validando');
                await refreshAllAccounts();
                ctx.log('success', 'Sesiones validadas');
            } else {
                ctx.log('debug', 'Auto refresh desactivado');
            }
            const d = Math.round(performance.now() - t0);
            ctx.log('debug', `Cuentas listas en ${d}ms`);
        } catch (e: any) {
            ctx.log('warn', `Cuentas ${e?.message ?? e}`);
        }
    },
};
