// Tarjeta de notificaciones del panel: bandeja de la cuenta
// (listar con paginado, filtrar no leídas, marcar y descartar).
<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { IconCheck, IconTrash, IconChevronLeft, IconChevronRight, IconAlertCircle } from '@tabler/icons-vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { useNotifications } from '@/Auth/Composables/useNotifications';
import {
    listarNotificaciones,
    marcarLeida,
    marcarTodasLeidas,
    borrarNotificacion,
    type NotificationItem,
} from '@/Auth/Api';
import { ApiError } from '@/Auth/Api';

const POR_PAGINA = 10;

const { conAuth } = useAuth();
const { refrescar } = useNotifications();

const items = ref<NotificationItem[]>([]);
const noLeidas = ref(0);
const soloNuevas = ref(false);
const pagina = ref(1);
const hayMas = ref(false);
const cargando = ref(false);
const trabajando = ref(false);
const error = ref('');

function fecha(iso: string): string {
    if (!iso) return '—';
    try {
        return new Date(iso).toLocaleDateString('es-AR', { day: 'numeric', month: 'short', year: 'numeric' });
    } catch {
        return '—';
    }
}

async function cargar(): Promise<void> {
    cargando.value = true;
    error.value = '';
    try {
        const res = await conAuth((token) =>
            listarNotificaciones(token, POR_PAGINA, (pagina.value - 1) * POR_PAGINA, soloNuevas.value),
        );
        items.value = res.notifications;
        noLeidas.value = res.unreadCount;
        hayMas.value = res.notifications.length === POR_PAGINA;
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        cargando.value = false;
    }
}

watch([soloNuevas], () => {
    pagina.value = 1;
    cargar();
});

watch(pagina, () => cargar());

async function leer(id: string) {
    error.value = '';
    trabajando.value = true;
    try {
        const actualizada = await conAuth((token) => marcarLeida(token, id));
        items.value = items.value.map((n) => (n.id === id ? actualizada : n));
        if (soloNuevas.value) items.value = items.value.filter((n) => n.id !== id);
        await refrescar();
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        trabajando.value = false;
    }
}

async function leerTodas() {
    error.value = '';
    trabajando.value = true;
    try {
        await conAuth((token) => marcarTodasLeidas(token));
        if (soloNuevas.value) {
            items.value = [];
        } else {
            items.value = items.value.map((n) => ({ ...n, read: true }));
        }
        await refrescar();
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        trabajando.value = false;
    }
}

async function descartar(id: string) {
    error.value = '';
    trabajando.value = true;
    try {
        await conAuth((token) => borrarNotificacion(token, id));
        items.value = items.value.filter((n) => n.id !== id);
        await refrescar();
        if (items.value.length === 0 && pagina.value > 1) pagina.value -= 1;
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        trabajando.value = false;
    }
}

onMounted(() => cargar());
</script>

<template>
    <section class="Card" aria-label="Notificaciones">
        <div class="Row">
            <button
                type="button"
                class="GhostBtn"
                :class="{ active: soloNuevas }"
                :aria-pressed="soloNuevas"
                @click="soloNuevas = !soloNuevas"
            >
                Solo no leídos
            </button>
            <button type="button" class="GhostBtn" :disabled="trabajando || noLeidas === 0" @click="leerTodas">
                <IconCheck stroke="2" />
                Marcar todas
            </button>
            <span v-if="!cargando" class="Count">{{ noLeidas }} sin leer</span>
        </div>
        <p v-if="cargando" class="State">Cargando tus notificaciones...</p>
        <p v-else-if="items.length === 0 && !error" class="State">No tenés notificaciones por acá. Todo al día.</p>
        <ul v-else class="List">
            <li v-for="n in items" :key="n.id" :class="{ unread: !n.read }">
                <div class="Txt">
                    <b>{{ n.title }}</b>
                    <p>{{ n.body }}</p>
                    <small>{{ fecha(n.createdAt) }}</small>
                </div>
                <div class="RowBtns">
                    <button
                        v-if="!n.read"
                        type="button"
                        class="MiniBtn"
                        :disabled="trabajando"
                        aria-label="Marcar como leído"
                        title="Marcar como leído"
                        @click="leer(n.id)"
                    >
                        <IconCheck stroke="2" />
                    </button>
                    <button
                        type="button"
                        class="MiniBtn danger"
                        :disabled="trabajando"
                        aria-label="Descartar aviso"
                        title="Descartar"
                        @click="descartar(n.id)"
                    >
                        <IconTrash stroke="2" />
                    </button>
                </div>
            </li>
        </ul>
        <nav v-if="!cargando && (pagina > 1 || hayMas)" class="Pagination" aria-label="Paginar notificaciones">
            <button
                type="button"
                class="PageBtn"
                :disabled="pagina <= 1"
                aria-label="Página anterior"
                @click="pagina > 1 && (pagina -= 1)"
            >
                <IconChevronLeft stroke="2" />
            </button>
            <span class="PageInfo">Página {{ pagina }}</span>
            <button
                type="button"
                class="PageBtn"
                :disabled="!hayMas"
                aria-label="Página siguiente"
                @click="hayMas && (pagina += 1)"
            >
                <IconChevronRight stroke="2" />
            </button>
        </nav>
        <p v-if="error" class="FormError" role="alert">
            <IconAlertCircle stroke="2" />
            {{ error }}
        </p>
    </section>
</template>

<style scoped lang="scss">
.Card{
    display: flex;
    flex-direction: column;
    gap: 1.1rem;
    padding: 1.5rem;
    border-radius: .9rem;
    border: 1px solid #ffffff18;
    background: #ffffff08;
    .Row{
        display: flex;
        align-items: center;
        gap: .5rem;
        flex-wrap: wrap;
        .Count{
            margin-left: auto;
            font-size: .7rem;
            opacity: .45;
            white-space: nowrap;
        }
    }
    .GhostBtn{
        display: flex;
        justify-content: center;
        align-items: center;
        gap: .4rem;
        padding: .5rem 1rem;
        border-radius: .55rem;
        border: 1px solid #ffffff25;
        background: #ffffff08;
        color: #fff;
        font-size: .78rem;
        font-weight: 600;
        font-family: inherit;
        cursor: pointer;
        transition: background 150ms, color 150ms, opacity 150ms;
        svg{
            width: 1rem;
            height: 1rem;
        }
        &:hover:not(:disabled){
            background: #ffffff14;
        }
        &:disabled{
            opacity: .5;
            cursor: default;
        }
        &.active{
            background: #fff;
            color: #000;
            border-color: #fff;
        }
    }
    .State{
        margin: 0;
        text-align: center;
        font-size: .8rem;
        opacity: .55;
    }
    .List{
        margin: 0;
        padding: 0;
        list-style: none;
        display: flex;
        flex-direction: column;
        gap: .6rem;
        li{
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            gap: .8rem;
            padding: .85rem 1rem;
            border-radius: .7rem;
            border: 1px solid #ffffff14;
            background: #00000060;
            transition: border-color 150ms, background 150ms;
            &:hover{
                border-color: #ffffff25;
                background: #ffffff08;
            }
            .Txt{
                display: flex;
                flex-direction: column;
                gap: .25rem;
                min-width: 0;
                b{
                    display: flex;
                    align-items: center;
                    gap: .45rem;
                    font-size: .82rem;
                    font-family: 'Lexend';
                }
                p{
                    margin: 0;
                    font-size: .78rem;
                    line-height: 1.55;
                    opacity: .7;
                    overflow-wrap: anywhere;
                }
                small{
                    font-size: .68rem;
                    opacity: .45;
                }
            }
            &.unread{
                border-color: #ffffff2e;
                background: #ffffff0d;
                .Txt b::before{
                    content: '';
                    flex-shrink: 0;
                    width: .45rem;
                    height: .45rem;
                    border-radius: 99rem;
                    background: #fff;
                }
            }
            .RowBtns{
                display: flex;
                gap: .4rem;
                flex-shrink: 0;
                .MiniBtn{
                    display: flex;
                    justify-content: center;
                    align-items: center;
                    min-width: 2.2rem;
                    height: 2.2rem;
                    padding: 0 .5rem;
                    border-radius: .55rem;
                    border: 1px solid #ffffff25;
                    background: #ffffff08;
                    color: #fff;
                    cursor: pointer;
                    transition: background 150ms, opacity 150ms;
                    svg{
                        width: 1rem;
                        height: 1rem;
                    }
                    &:hover:not(:disabled){
                        background: #ffffff14;
                    }
                    &:disabled{
                        opacity: .5;
                        cursor: wait;
                    }
                    &.danger:hover:not(:disabled){
                        background: #ff6b6b22;
                    }
                }
            }
        }
    }
    .Pagination{
        display: flex;
        justify-content: center;
        align-items: center;
        gap: .5rem;
        .PageBtn{
            display: flex;
            justify-content: center;
            align-items: center;
            min-width: 2.2rem;
            height: 2.2rem;
            padding: 0 .5rem;
            border-radius: .55rem;
            border: 1px solid #ffffff25;
            background: #ffffff08;
            color: #ffffffa6;
            cursor: pointer;
            transition: background 150ms, color 150ms, opacity 150ms;
            svg{
                width: 1rem;
                height: 1rem;
            }
            &:hover:not(:disabled){
                background: #ffffff14;
                color: #fff;
            }
            &:disabled{
                opacity: .3;
                cursor: default;
            }
        }
        .PageInfo{
            font-size: .75rem;
            font-family: 'Lexend';
            opacity: .6;
        }
    }
    .FormError{
        display: flex;
        align-items: center;
        gap: .5rem;
        margin: 0;
        padding: .65rem .85rem;
        border-radius: .55rem;
        border: 1px solid #ff9d9d45;
        background: #ff6b6b14;
        font-size: .8rem;
        line-height: 1.5;
        svg{
            flex-shrink: 0;
            width: 1.1rem;
            height: 1.1rem;
            color: #ff9d9d;
        }
    }
}
</style>
