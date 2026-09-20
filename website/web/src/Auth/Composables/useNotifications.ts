// Contador global de no leídas (campana + tarjeta lo comparten).

import { ref } from 'vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { listarNotificaciones } from '@/Auth/Api';

const noLeidas = ref(0);

export function useNotifications() {
    async function refrescar(): Promise<void> {
        const auth = useAuth();
        if (!auth.autenticado.value) {
            noLeidas.value = 0;
            return;
        }
        try {
            const pagina = await auth.conAuth((token) => listarNotificaciones(token, 1, 0, false));
            noLeidas.value = pagina.unreadCount;
        } catch {
            // Se conserva el valor previo: el contador es orientativo.
        }
    }

    return { noLeidas, refrescar };
}
