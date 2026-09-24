// Tarjeta de amigos del panel: buscar jugadores y enviar solicitudes,
// gestionar recibidas/enviadas y ver la lista de amigos.
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import {
    IconSearch, IconUserPlus, IconCheck, IconX, IconTrash,
    IconAlertCircle, IconCircleCheck,
} from '@tabler/icons-vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import {
    buscarJugadores, enviarSolicitud, listarSolicitudes, aceptarSolicitud,
    rechazarSolicitud, cancelarSolicitud, listarAmigos, romperAmistad,
    listarBloqueos, bloquear, desbloquear,
    type Jugador, type Solicitud, type Amigo, type Bloqueado,
} from '@/Auth/Api';
import { ApiError } from '@/Auth/Api';

const POR_TANDA = 20;

const { conAuth } = useAuth();

const busqueda = ref('');
const resultados = ref<Jugador[]>([]);
const buscando = ref(false);
const errorBusqueda = ref('');
const aviso = ref('');
const enviandoA = ref('');

const recibidas = ref<Solicitud[]>([]);
const enviadas = ref<Solicitud[]>([]);
const amigos = ref<Amigo[]>([]);
const totalAmigos = ref(0);
const bloqueos = ref<Bloqueado[]>([]);
const inputBloqueo = ref('');
const errorBloqueo = ref('');
const visibles = ref(POR_TANDA);
const cargando = ref(true);
const trabajando = ref(false);
const error = ref('');
const confirmarQuite = ref('');

function nombreDe(u: { username: string; displayName: string }): string {
    return u.displayName || u.username;
}

// Avatares que fallaron al cargar (misma idea que en header/panel).
const avataresRotos = ref(new Set<string>());

function avatarOk(url: string | null | undefined): boolean {
    return !!url && !avataresRotos.value.has(url);
}

function romperAvatar(url: string | null | undefined) {
    if (url) avataresRotos.value.add(url);
}

const pendientesRecibidas = computed(() => recibidas.value.filter((s) => s.status === 'pending'));
const pendientesEnviadas = computed(() => enviadas.value.filter((s) => s.status === 'pending'));
const amigosVisibles = computed(() => amigos.value.slice(0, visibles.value));

async function cargarTodo(): Promise<void> {
    cargando.value = true;
    error.value = '';
    try {
        const [rec, env, ami, blo] = await Promise.all([
            conAuth((token) => listarSolicitudes(token, 'incoming')),
            conAuth((token) => listarSolicitudes(token, 'sent')),
            conAuth((token) => listarAmigos(token)),
            conAuth((token) => listarBloqueos(token)),
        ]);
        recibidas.value = rec.requests;
        enviadas.value = env.requests;
        amigos.value = ami.friends;
        totalAmigos.value = ami.count;
        bloqueos.value = blo.blocked;
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        cargando.value = false;
    }
}

async function buscar() {
    const q = busqueda.value.trim();
    errorBusqueda.value = '';
    aviso.value = '';
    resultados.value = [];
    if (q.length < 1) return;
    buscando.value = true;
    try {
        const res = await conAuth((token) => buscarJugadores(token, q, 10));
        resultados.value = res.users;
        if (res.users.length === 0) errorBusqueda.value = 'Nadie coincide con esa búsqueda.';
    } catch (err) {
        errorBusqueda.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        buscando.value = false;
    }
}

async function solicitar(username: string) {
    errorBusqueda.value = '';
    aviso.value = '';
    enviandoA.value = username;
    try {
        await conAuth((token) => enviarSolicitud(token, username));
        aviso.value = `Solicitud enviada a ${username}.`;
        resultados.value = resultados.value.filter((u) => u.username !== username);
        await cargarTodo();
    } catch (err) {
        errorBusqueda.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        enviandoA.value = '';
    }
}

async function aceptar(id: string) {
    error.value = '';
    aviso.value = '';
    trabajando.value = true;
    try {
        await conAuth((token) => aceptarSolicitud(token, id));
        aviso.value = 'Amistad aceptada.';
        await cargarTodo();
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        trabajando.value = false;
    }
}

async function rechazar(id: string) {
    error.value = '';
    trabajando.value = true;
    try {
        await conAuth((token) => rechazarSolicitud(token, id));
        await cargarTodo();
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        trabajando.value = false;
    }
}

async function cancelar(id: string) {
    error.value = '';
    trabajando.value = true;
    try {
        await conAuth((token) => cancelarSolicitud(token, id));
        aviso.value = 'Solicitud cancelada.';
        await cargarTodo();
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        trabajando.value = false;
    }
}

async function quitar(friendId: string) {    if (confirmarQuite.value !== friendId) {
        confirmarQuite.value = friendId;
        return;
    }
    confirmarQuite.value = '';
    error.value = '';
    trabajando.value = true;
    try {
        await conAuth((token) => romperAmistad(token, friendId));
        amigos.value = amigos.value.filter((a) => a.userId !== friendId);
        totalAmigos.value = Math.max(0, totalAmigos.value - 1);
        aviso.value = 'Amistad eliminada.';
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        trabajando.value = false;
    }
}

function nombreBloqueado(b: Bloqueado): string {
    const u = b.user;
    if ('username' in u && u.username) return `@${u.username}`;
    return `UUID ${u.userId.slice(0, 8)}…`;
}

function idBloqueado(b: Bloqueado): string {
    return b.user.userId;
}

async function bloquearDirecto() {
    const identifier = inputBloqueo.value.trim();
    errorBloqueo.value = '';
    aviso.value = '';
    if (!identifier) return;
    trabajando.value = true;
    try {
        await conAuth((token) => bloquear(token, identifier));
        aviso.value = 'Jugador bloqueado.';
        inputBloqueo.value = '';
        const blo = await conAuth((token) => listarBloqueos(token));
        bloqueos.value = blo.blocked;
    } catch (err) {
        errorBloqueo.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        trabajando.value = false;
    }
}

async function desbloquearUsuario(userId: string) {
    errorBloqueo.value = '';
    aviso.value = '';
    trabajando.value = true;
    try {
        await conAuth((token) => desbloquear(token, userId));
        bloqueos.value = bloqueos.value.filter((b) => idBloqueado(b) !== userId);
        aviso.value = 'Jugador desbloqueado.';
    } catch (err) {
        errorBloqueo.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        trabajando.value = false;
    }
}

onMounted(() => cargarTodo());
</script>

<template>
    <section class="Card" aria-label="Amigos">
        <form class="Search" @submit.prevent="buscar">
            <div class="SearchRow">
                <input v-model="busqueda" type="search" placeholder="Usuario, email o UUID..." aria-label="Buscar jugador">
                <button type="submit" class="GhostBtn" :disabled="buscando">
                    <IconSearch stroke="2" />
                    {{ buscando ? 'Buscando...' : 'Buscar' }}
                </button>
            </div>
            <small v-if="errorBusqueda" class="FieldError">{{ errorBusqueda }}</small>
            <p v-if="aviso" class="Ok" role="status">
                <IconCircleCheck stroke="2" />
                {{ aviso }}
            </p>
            <ul v-if="resultados.length > 0" class="Results">
                <li v-for="u in resultados" :key="u.username">
                    <RouterLink class="Who link" :to="`/community/account/@${u.username}`">
                        <img v-if="avatarOk(u.avatarUrl)" :src="u.avatarUrl || ''" :alt="u.username" loading="lazy" @error="romperAvatar(u.avatarUrl)">
                        <i v-else>{{ (nombreDe(u).charAt(0) || '?').toUpperCase() }}</i>
                        <span class="Names">
                            <b>{{ nombreDe(u) }}</b>
                            <small>@{{ u.username }}</small>
                        </span>
                    </RouterLink>
                    <button type="button" class="MiniBtn" :disabled="!!enviandoA" :aria-label="`Añadir a ${u.username}`" @click="solicitar(u.username)">
                        <IconUserPlus stroke="2" />
                        {{ enviandoA === u.username ? 'Enviando...' : 'Añadir' }}
                    </button>
                </li>
            </ul>
        </form>
        <p v-if="cargando" class="State">Cargando tus amigos...</p>
        <template v-else>
            <div v-if="pendientesRecibidas.length > 0" class="Group">
                <b>Quieren ser tus amigos ({{ pendientesRecibidas.length }})</b>
                <ul>
                    <li v-for="s in pendientesRecibidas" :key="s.id">
                        <RouterLink v-if="s.user" class="Who link" :to="`/community/account/@${s.user.username}`">
                            <img v-if="avatarOk(s.user?.avatarUrl)" :src="s.user?.avatarUrl || ''" :alt="s.user?.username" loading="lazy" @error="romperAvatar(s.user?.avatarUrl)">
                            <i v-else>{{ ((s.user?.displayName || s.user?.username || '?').charAt(0) || '?').toUpperCase() }}</i>
                            <span class="Names">
                                <b>{{ s.user ? nombreDe(s.user) : 'Jugador' }}</b>
                                <small v-if="s.user">@{{ s.user.username }}</small>
                            </span>
                        </RouterLink>
                        <span v-else class="Who">
                            <i>J</i>
                            <span class="Names">
                                <b>Jugador</b>
                            </span>
                        </span>
                        <span class="RowBtns">
                            <button type="button" class="MiniBtn ok" :disabled="trabajando" aria-label="Aceptar" @click="aceptar(s.id)">
                                <IconCheck stroke="2" />
                            </button>
                            <button type="button" class="MiniBtn danger" :disabled="trabajando" aria-label="Rechazar" @click="rechazar(s.id)">
                                <IconX stroke="2" />
                            </button>
                        </span>
                    </li>
                </ul>
            </div>
            <div v-if="pendientesEnviadas.length > 0" class="Group">
                <b>Esperando respuesta ({{ pendientesEnviadas.length }})</b>
                <ul>
                    <li v-for="s in pendientesEnviadas" :key="s.id">
                        <RouterLink v-if="s.user" class="Who link" :to="`/community/account/@${s.user.username}`">
                            <img v-if="avatarOk(s.user?.avatarUrl)" :src="s.user?.avatarUrl || ''" :alt="s.user?.username" loading="lazy" @error="romperAvatar(s.user?.avatarUrl)">
                            <i v-else>{{ ((s.user?.displayName || s.user?.username || '?').charAt(0) || '?').toUpperCase() }}</i>
                            <span class="Names">
                                <b>{{ s.user ? nombreDe(s.user) : 'Jugador' }}</b>
                                <small v-if="s.user">@{{ s.user.username }}</small>
                            </span>
                        </RouterLink>
                        <span v-else class="Who">
                            <i>J</i>
                            <span class="Names">
                                <b>Jugador</b>
                            </span>
                        </span>
                        <button type="button" class="MiniBtn" :disabled="trabajando" @click="cancelar(s.id)">
                            Cancelar
                        </button>
                    </li>
                </ul>
            </div>
            <div class="Group">
                <b>Mis amigos ({{ totalAmigos }})</b>
                <p v-if="amigos.length === 0" class="State">Todavía no tenés amigos. Buscá a alguien arriba.</p>
                <ul v-else>
                    <li v-for="a in amigosVisibles" :key="a.userId">
                        <RouterLink class="Who link" :to="`/community/account/@${a.username}`">
                            <img v-if="avatarOk(a.avatarUrl)" :src="a.avatarUrl || ''" :alt="a.username" loading="lazy" @error="romperAvatar(a.avatarUrl)">
                            <i v-else>{{ (nombreDe(a).charAt(0) || '?').toUpperCase() }}</i>
                            <span class="Names">
                                <b>{{ nombreDe(a) }}</b>
                                <small>@{{ a.username }}</small>
                            </span>
                        </RouterLink>
                        <button
                            type="button"
                            class="MiniBtn danger"
                            :class="{ confirm: confirmarQuite === a.userId }"
                            :disabled="trabajando"
                            :aria-label="confirmarQuite === a.userId ? 'Confirmar' : `Dejar de ser amigo de ${a.username}`"
                            @click="quitar(a.userId)"
                        >
                            <IconTrash stroke="2" />
                            <i v-if="confirmarQuite === a.userId">¿Seguro?</i>
                        </button>
                    </li>
                </ul>
                <button
                    v-if="amigos.length > visibles"
                    type="button"
                    class="GhostBtn more"
                    @click="visibles += POR_TANDA"
                >
                    Mostrar más ({{ amigos.length - visibles }} restantes)
                </button>
            </div>
            <div class="Group">
                <b>Lista negra ({{ bloqueos.length }})</b>
                <form class="BlockRow" @submit.prevent="bloquearDirecto">
                    <input v-model="inputBloqueo" type="text" placeholder="Usuario, email o UUID..." aria-label="Bloquear jugador">
                    <button type="submit" class="MiniBtn danger" :disabled="trabajando">Bloquear</button>
                </form>
                <small v-if="errorBloqueo" class="FieldError">{{ errorBloqueo }}</small>
                <p v-if="bloqueos.length === 0" class="State">Todavía no bloqueaste a nadie.</p>
                <ul v-else>
                    <li v-for="b in bloqueos" :key="idBloqueado(b)">
                        <span class="Who">
                            <i>B</i>
                            <span class="Names">
                                <b>{{ nombreBloqueado(b) }}</b>
                            </span>
                        </span>
                        <button type="button" class="MiniBtn" :disabled="trabajando" @click="desbloquearUsuario(idBloqueado(b))">
                            Desbloquear
                        </button>
                    </li>
                </ul>
            </div>
        </template>
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
    .Search{
        display: flex;
        flex-direction: column;
        gap: .6rem;
        .SearchRow{
            display: flex;
            gap: .5rem;
            input{
                flex: 1;
                min-width: 0;
                padding: .6rem .85rem;
                border-radius: .55rem;
                border: 1px solid #ffffff25;
                background: #00000080;
                color: #fff;
                font-size: .82rem;
                font-family: inherit;
                outline: none;
                transition: border-color 150ms;
                &::placeholder{
                    color: #ffffff45;
                }
                &:focus{
                    border-color: #ffffff60;
                }
            }
        }
        .FieldError{
            font-size: .72rem;
            color: #ff9d9d;
        }
    }
    .Results,
    .Group ul{
        margin: 0;
        padding: 0;
        list-style: none;
        display: flex;
        flex-direction: column;
        li{
            display: flex;
            justify-content: space-between;
            align-items: center;
            gap: .8rem;
            padding: .6rem 0;
            border-bottom: 1px solid #ffffff10;
            &:last-child{
                border-bottom: 0;
            }
        }
    }
    .Group{
        display: flex;
        flex-direction: column;
        gap: .4rem;
        b{
            font-size: .78rem;
            font-family: 'Lexend';
            opacity: .8;
        }
        .BlockRow{
            display: flex;
            gap: .5rem;
            input{
                flex: 1;
                min-width: 0;
                padding: .55rem .8rem;
                border-radius: .55rem;
                border: 1px solid #ffffff25;
                background: #00000080;
                color: #fff;
                font-size: .78rem;
                font-family: inherit;
                outline: none;
                transition: border-color 150ms;
                &::placeholder{
                    color: #ffffff45;
                }
                &:focus{
                    border-color: #ffffff60;
                }
            }
        }
    }
    .Who{
        display: flex;
        align-items: center;
        gap: .6rem;
        min-width: 0;
        &.link{
            color: #fff;
            text-decoration: none;
            border-radius: .5rem;
            transition: opacity 150ms;
            &:hover{
                opacity: .75;
            }
        }
        img,
        i{
            width: 2.2rem;
            height: 2.2rem;
            flex-shrink: 0;
            border-radius: 99rem;
            object-fit: cover;
            transition: outline-color 150ms, box-shadow 150ms;
            &.on{
                outline: 2px solid #4caf50;
                outline-offset: 1px;
                box-shadow: 0 0 10px #4caf5066;
            }
        }
        i{
            display: flex;
            justify-content: center;
            align-items: center;
            font-style: normal;
            font-family: 'Lexend';
            font-size: .85rem;
            font-weight: 700;
            background: #ffffff18;
        }
        .Names{
            display: flex;
            flex-direction: column;
            min-width: 0;
            b{
                font-size: .8rem;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }
            small{
                font-size: .68rem;
                opacity: .5;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }
        }
    }
    .GhostBtn{
        display: flex;
        justify-content: center;
        align-items: center;
        gap: .4rem;
        padding: .55rem 1rem;
        border-radius: .55rem;
        border: 1px solid #ffffff25;
        background: #ffffff08;
        color: #fff;
        font-size: .78rem;
        font-weight: 600;
        font-family: inherit;
        cursor: pointer;
        white-space: nowrap;
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
        &.more{
            align-self: center;
            margin-top: .4rem;
        }
    }
    .MiniBtn{
        display: flex;
        justify-content: center;
        align-items: center;
        gap: .35rem;
        padding: .45rem .8rem;
        border-radius: .55rem;
        border: 1px solid #ffffff25;
        background: #ffffff08;
        color: #fff;
        font-size: .74rem;
        font-weight: 600;
        font-family: inherit;
        cursor: pointer;
        white-space: nowrap;
        transition: background 150ms, opacity 150ms;
        svg{
            width: 1rem;
            height: 1rem;
        }
        i{
            font-style: normal;
            font-size: .7rem;
            font-weight: 700;
        }
        &:hover:not(:disabled){
            background: #ffffff14;
        }
        &:disabled{
            opacity: .5;
            cursor: wait;
        }
        &.ok:hover:not(:disabled){
            background: #4caf5030;
        }
        &.danger{
            &:hover:not(:disabled){
                background: #ff6b6b22;
            }
            &.confirm{
                border-color: #ff9d9d60;
                background: #ff6b6b22;
            }
        }
    }
    .RowBtns{
        display: flex;
        gap: .4rem;
        flex-shrink: 0;
    }
    .State{
        margin: 0;
        font-size: .8rem;
        opacity: .55;
    }
    .Ok{
        display: flex;
        align-items: center;
        gap: .5rem;
        margin: 0;
        font-size: .8rem;
        color: #9dffb0;
        svg{
            width: 1.1rem;
            height: 1.1rem;
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
@media (max-width: 600px){
    .Card{
        min-width: 0;
        padding: 1rem;
        .Search{
            .SearchRow{
                flex-direction: column;
                align-items: stretch;
                .GhostBtn{
                    width: 100%;
                    min-height: 2.75rem;
                }
            }
        }
        .Group{
            min-width: 0;
            .BlockRow{
                flex-direction: column;
                align-items: stretch;
                .MiniBtn{
                    width: 100%;
                    min-height: 2.75rem;
                }
            }
        }
        .Results,
        .Group ul{
            li{
                flex-wrap: wrap;
                .Who{
                    flex: 1 1 8rem;
                }
            }
        }
        .GhostBtn{
            &.more{
                width: 100%;
                min-height: 2.75rem;
                white-space: normal;
                text-align: center;
            }
        }
    }
}
</style>
