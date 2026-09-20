// Previsualización pública de un perfil: portada, datos, Minecraft,
// cosméticos que lleva puestos y acciones según la relación
// (amistad, solicitudes, bloqueos). Ruta protegida por el guard.
<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
    IconUserPlus, IconCheck, IconAlertCircle, IconShieldX,
    IconUserOff, IconCalendar,
} from '@tabler/icons-vue';
import SkinStage from '@/Auth/Dashboard/Components/SkinStage.vue';
import ProfileHead, { type StatCabecera } from '@/Auth/Components/ProfileHead.vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import {
    verPerfil, enviarSolicitud, aceptarSolicitud, rechazarSolicitud, cancelarSolicitud,
    romperAmistad, bloquear, desbloquear, buscarJugadores,
    type PreviewComunidad, type CosmeticoEquipado, type Jugador,
} from '@/Auth/Api';
import { ApiError } from '@/Auth/Api';
import { urlFirmada } from '@/Auth/firmadas';
import { applySeo, DEFAULT_IMAGE, SITE_URL } from '@/Common/Composables/useSeo';

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

const ruta = useRoute();
const router = useRouter();
const { conAuth } = useAuth();

const identificador = computed(() => {
    const p = ruta.params.id;
    const crudo = Array.isArray(p) ? (p[0] ?? '') : (p ?? '');
    return crudo.trim().replace(/^@/, '');
});

type Estado = 'cargando' | 'lista' | 'no-encontrado' | 'error' | 'buscar';

const estado = ref<Estado>('cargando');
const error = ref('');
const datos = ref<PreviewComunidad | null>(null);
const bloqueo = ref<{ reason: 'blocked_you' | 'blocked_by_you'; message: string } | null>(null);
const equipo = ref<Array<{ item: CosmeticoEquipado; url: string }>>([]);
const trabajando = ref(false);
const aviso = ref('');

// Sin identificador: buscador para encontrar jugadores.
const busqueda = ref('');
const resultados = ref<Jugador[]>([]);
const buscando = ref(false);
const errorBusqueda = ref('');

const statsPerfil = computed<StatCabecera[]>(() => {
    if (!datos.value) return [];
    const stats: StatCabecera[] = [
        { valor: `${equipo.value.length}`, etiqueta: equipo.value.length === 1 ? 'puesto' : 'puestos' },
        { valor: datos.value.minecraft.lastVersion ?? '—', etiqueta: 'último MC' },
    ];
    return stats;
});

const inicialPerfil = computed(() => (datos.value?.profile.displayName.trim().charAt(0) || '?').toUpperCase());

function fecha(iso: string | null | undefined): string {
    if (!iso) return '—';
    try {
        return new Date(iso).toLocaleDateString('es-AR', { day: 'numeric', month: 'short', year: 'numeric' });
    } catch {
        return '—';
    }
}

// El SEO de la ruta ya puso "@id"; acá se enriquece con el nombre real
// cuando el perfil carga, para que lo compartido coincida con lo visible.
function updateProfileSeo(): void {
    const id = identificador.value.trim().replace(/^@/, '');
    if (!id) return;
    const url = `${SITE_URL}/community/account/${id}`;
    if (estado.value === 'no-encontrado') {
        applySeo({
            title: 'Jugador no encontrado - StepLauncher',
            description: `No encontramos a @${id} en la comunidad StepLauncher. Revisá el usuario e intentá de nuevo.`,
            image: DEFAULT_IMAGE,
            url,
            robots: 'noindex, nofollow',
        });
        return;
    }
    if (!datos.value) return;
    const nombre = datos.value.profile.displayName.trim() || datos.value.profile.username;
    const usuario = datos.value.profile.username;
    const puestos = equipo.value.length;
    const mc = datos.value.minecraft.lastVersion ?? '—';
    applySeo({
        title: `${nombre} (@${usuario}) en StepLauncher`,
        description: `Mirá el perfil de ${nombre} (@${usuario}) en la comunidad StepLauncher: ${puestos} ${puestos === 1 ? 'puesto' : 'puestos'}, último MC ${mc} y más.`,
        image: DEFAULT_IMAGE,
        url,
        type: 'profile',
        robots: 'noindex, nofollow',
    });
}

async function cargar(): Promise<void> {
    estado.value = 'cargando';
    error.value = '';
    aviso.value = '';
    datos.value = null;
    bloqueo.value = null;
    equipo.value = [];
    const id = identificador.value.trim();
    if (!id) {
        estado.value = 'buscar';
        return;
    }
    try {
        const res = await conAuth((token) => verPerfil(token, id));
        if (res.blocked) {
            bloqueo.value = { reason: res.reason, message: res.message };
            estado.value = 'lista';
            return;
        }
        datos.value = res;
        estado.value = 'lista';
        // URLs firmadas de lo equipado (solo skins/capas se renderizan).
        const piezas = res.equippedCosmetics.filter((c) => c.kind === 'skin' || c.kind === 'cape');
        const urls = await Promise.all(
            piezas.map(async (item) => {
                try {
                    return { item, url: await urlFirmada(conAuth, item.id) };
                } catch {
                    return null;
                }
            }),
        );
        equipo.value = urls.filter((u): u is { item: CosmeticoEquipado; url: string } => u !== null);
        updateProfileSeo();
    } catch (err) {
        if (err instanceof ApiError && err.code === 'profile_not_found') {
            estado.value = 'no-encontrado';
            updateProfileSeo();
        } else {
            estado.value = 'error';
            error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
        }
    }
}

async function actuar(accion: () => Promise<unknown>, okMsg: string): Promise<void> {
    error.value = '';
    aviso.value = '';
    trabajando.value = true;
    try {
        await accion();
        if (okMsg) aviso.value = okMsg;
        await cargar();
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        trabajando.value = false;
    }
}

function username(): string {
    return datos.value?.profile.username ?? '';
}

function solicitar() {
    const u = username();
    if (u) actuar((() => conAuth((token) => enviarSolicitud(token, u))), `Solicitud enviada a ${u}.`);
}

function aceptar() {
    const rid = datos.value?.relationship.requestId;
    if (rid) actuar(() => conAuth((token) => aceptarSolicitud(token, rid)), 'Amistad aceptada.');
}

function rechazar() {
    const rid = datos.value?.relationship.requestId;
    if (rid) actuar(() => conAuth((token) => rechazarSolicitud(token, rid)), '');
}

function retirar() {
    const rid = datos.value?.relationship.requestId;
    if (rid) actuar(() => conAuth((token) => cancelarSolicitud(token, rid)), 'Solicitud retirada.');
}

function romper() {
    const fid = datos.value?.profile.userId;
    if (fid) actuar(() => conAuth((token) => romperAmistad(token, fid)), 'Amistad eliminada.');
}

function bloquearUsuario() {
    const u = username();
    if (u) actuar(() => conAuth((token) => bloquear(token, u)), 'Jugador bloqueado.');
}

function desbloquearUsuario() {
    // El aviso de bloqueo no trae userId: solo se puede directo si la URL es UUID.
    const id = identificador.value.trim();
    if (UUID_RE.test(id)) actuar(() => conAuth((token) => desbloquear(token, id)), 'Jugador desbloqueado.');
}

function irAmigos() {
    router.push('/dashboard?tab=amigos');
}

async function buscar() {
    const q = busqueda.value.trim();
    errorBusqueda.value = '';
    resultados.value = [];
    if (!q) return;
    buscando.value = true;
    try {
        const res = await conAuth((token) => buscarJugadores(token, q.replace(/^@/, ''), 10));
        resultados.value = res.users;
        if (res.users.length === 0) errorBusqueda.value = 'Nadie coincide con esa búsqueda.';
    } catch (err) {
        errorBusqueda.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        buscando.value = false;
    }
}

watch(identificador, () => cargar());
onMounted(() => cargar());
</script>

<template>
    <div class="Community">
        <div class="Wrap">
            <p v-if="estado === 'cargando'" class="State">Buscando ese perfil...</p>
            <div v-else-if="estado === 'buscar'" class="Search sl-fade">
                <h1>Comunidad</h1>
                <p>Buscá un jugador por su usuario para ver su perfil: <b>/community/account/@usuario</b>.</p>
                <form class="SearchRow" @submit.prevent="buscar">
                    <input v-model="busqueda" type="search" placeholder="@usuario..." aria-label="Buscar jugador">
                    <button type="submit" class="GhostBtn" :disabled="buscando">
                        {{ buscando ? 'Buscando...' : 'Buscar' }}
                    </button>
                </form>
                <small v-if="errorBusqueda" class="FieldError">{{ errorBusqueda }}</small>
                <ul v-if="resultados.length > 0" class="Results">
                    <li v-for="u in resultados" :key="u.username">
                        <RouterLink class="Who link" :to="`/community/account/@${u.username}`">
                            <img v-if="u.avatarUrl" :src="u.avatarUrl" :alt="u.username" loading="lazy">
                            <i v-else>{{ ((u.displayName || u.username).charAt(0) || '?').toUpperCase() }}</i>
                            <span class="Names">
                                <b>{{ u.displayName || u.username }}</b>
                                <small>@{{ u.username }}</small>
                            </span>
                        </RouterLink>
                    </li>
                </ul>
            </div>
            <div v-else-if="estado === 'no-encontrado'" class="Notice sl-fade">
                <IconUserOff stroke="2" />
                <b>Ese perfil no existe o es privado.</b>
                <p>Puede que el jugador haya desactivado "que puedan buscarte".</p>
                <button type="button" class="GhostBtn" @click="irAmigos">Buscar jugadores</button>
            </div>
            <p v-else-if="estado === 'error'" class="FormError" role="alert">
                <IconAlertCircle stroke="2" />
                {{ error }}
            </p>
            <template v-else-if="bloqueo">
                <div class="Notice sl-fade">
                    <IconShieldX stroke="2" />
                    <b>{{ bloqueo.reason === 'blocked_you' ? 'No podés ver este perfil.' : 'Lo tenés bloqueado.' }}</b>
                    <p>{{ bloqueo.message }}</p>
                    <button
                        v-if="bloqueo.reason === 'blocked_by_you' && UUID_RE.test(identificador.trim())"
                        type="button"
                        class="GhostBtn"
                        :disabled="trabajando"
                        @click="desbloquearUsuario"
                    >
                        Desbloquear
                    </button>
                    <button v-else-if="bloqueo.reason === 'blocked_by_you'" type="button" class="GhostBtn" @click="irAmigos">
                        Ver mi lista negra
                    </button>
                </div>
                <p v-if="error" class="FormError" role="alert">
                    <IconAlertCircle stroke="2" />
                    {{ error }}
                </p>
            </template>
            <template v-else-if="datos">
                <ProfileHead
                    class="sl-enter"
                    style="--sl-delay: 0s"
                    :banner-url="datos.profile.bannerUrl"
                    :avatar-url="datos.profile.avatarUrl"
                    :inicial="inicialPerfil"
                    eyebrow="Comunidad"
                    :titulo="datos.profile.displayName"
                    :handle="`@${datos.profile.username}`"
                    :stats="statsPerfil"
                    :en-linea="datos.profile.isOnline"
                >
                    <template #acciones>
                        <template v-if="datos.relationship.status === 'self'">
                            <RouterLink class="BtnPrimary" to="/dashboard">Mi panel</RouterLink>
                        </template>
                        <template v-else-if="datos.relationship.status === 'friends'">
                            <button type="button" class="GhostBtn" :disabled="trabajando" @click="romper">Dejar de ser amigos</button>
                            <button type="button" class="GhostBtn danger" :disabled="trabajando" @click="bloquearUsuario">Bloquear</button>
                        </template>
                        <template v-else-if="datos.relationship.status === 'pending_sent'">
                            <button type="button" class="GhostBtn" :disabled="trabajando" @click="retirar">Retirar solicitud</button>
                            <button type="button" class="GhostBtn danger" :disabled="trabajando" @click="bloquearUsuario">Bloquear</button>
                        </template>
                        <template v-else-if="datos.relationship.status === 'pending_received'">
                            <button type="button" class="BtnPrimary" :disabled="trabajando" @click="aceptar">Aceptar</button>
                            <button type="button" class="GhostBtn" :disabled="trabajando" @click="rechazar">Rechazar</button>
                            <button type="button" class="GhostBtn danger" :disabled="trabajando" @click="bloquearUsuario">Bloquear</button>
                        </template>
                        <template v-else>
                            <button type="button" class="BtnPrimary" :disabled="trabajando" @click="solicitar">
                                <IconUserPlus stroke="2" />
                                Añadir amigo
                            </button>
                            <button type="button" class="GhostBtn danger" :disabled="trabajando" @click="bloquearUsuario">Bloquear</button>
                        </template>
                    </template>
                </ProfileHead>
                <div v-if="datos.profile.bio" class="Bio sl-enter" style="--sl-delay: .08s">
                    <p>{{ datos.profile.bio }}</p>
                </div>
                <div class="Meta sl-enter" style="--sl-delay: .12s">
                    <span v-if="datos.profile.isOnline" class="Chip on">
                        En línea
                    </span>
                    <span class="Chip">
                        <IconCalendar stroke="2" />
                        Miembro desde {{ fecha(datos.profile.memberSince) }}
                    </span>
                    <span v-if="datos.relationship.status === 'friends' && datos.relationship.friendsSince" class="Chip">
                        <IconCheck stroke="2" />
                        Amigos desde {{ fecha(datos.relationship.friendsSince) }}
                    </span>
                    <span v-else-if="datos.relationship.status === 'pending_sent'" class="Chip">Solicitud enviada</span>
                    <span v-else-if="datos.relationship.status === 'pending_received'" class="Chip">Quiere ser tu amigo</span>
                </div>
                <p v-if="error" class="FormError" role="alert">
                    <IconAlertCircle stroke="2" />
                    {{ error }}
                </p>
                <p v-if="aviso" class="Ok" role="status">
                    <IconCheck stroke="2" />
                    {{ aviso }}
                </p>
                <div v-if="equipo.length > 0" class="Equipo sl-enter" style="--sl-delay: .2s">
                    <h2>Lleva puesto</h2>
                    <div class="Grid">
                        <article v-for="e in equipo" :key="e.item.id" class="Pieza">
                            <SkinStage
                                :skin="e.item.kind === 'skin' ? e.url : null"
                                :cape="e.item.kind === 'cape' ? e.url : null"
                                :ancho="220"
                                :alto="250"
                            />
                            <b>{{ e.item.name }}</b>
                            <small>{{ e.item.kind === 'skin' ? 'Skin' : 'Capa' }}</small>
                        </article>
                    </div>
                </div>
            </template>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Community{
    display: flex;
    flex-direction: column;
    align-items: center;
    background: #000;
    padding: 4.5rem 1.5rem 0 1.5rem;
}
.Wrap{
    display: flex;
    flex-direction: column;
    gap: 1.2rem;
    width: 100%;
    max-width: 70rem;
    margin-inline: auto;
    padding-bottom: 4rem;
}
.State{
    text-align: center;
    font-size: .85rem;
    opacity: .55;
}
.Search{
    display: flex;
    flex-direction: column;
    gap: .8rem;
    width: 100%;
    max-width: 38rem;
    align-self: center;
    padding: 2rem 0;
    h1{
        margin: 0;
        font-family: 'Lexend';
        font-size: 1.8rem;
        font-weight: 700;
    }
    p{
        margin: 0;
        font-size: .85rem;
        opacity: .6;
        b{
            font-family: monospace;
            font-weight: 400;
            opacity: 1;
        }
    }
    .SearchRow{
        display: flex;
        gap: .5rem;
        input{
            flex: 1;
            min-width: 0;
            padding: .65rem .85rem;
            border-radius: .55rem;
            border: 1px solid #ffffff25;
            background: #00000080;
            color: #fff;
            font-size: .85rem;
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
        font-size: .75rem;
        color: #ff9d9d;
    }
    .Results{
        margin: 0;
        padding: 0;
        list-style: none;
        display: flex;
        flex-direction: column;
        li{
            padding: .5rem 0;
            border-bottom: 1px solid #ffffff10;
            &:last-child{
                border-bottom: 0;
            }
        }
    }
    .Who{
        display: flex;
        align-items: center;
        gap: .6rem;
        color: #fff;
        text-decoration: none;
        border-radius: .5rem;
        transition: opacity 150ms;
        &:hover{
            opacity: .75;
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
                font-size: .82rem;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }
            small{
                font-size: .7rem;
                opacity: .5;
            }
        }
    }
}
.Notice{
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: .6rem;
    width: 100%;
    max-width: 44rem;
    align-self: center;
    padding: 3rem 1.5rem;
    border-radius: .9rem;
    border: 1px solid #ffffff18;
    background: #ffffff08;
    text-align: center;
    svg{
        width: 2.2rem;
        height: 2.2rem;
        opacity: .5;
    }
    b{
        font-family: 'Lexend';
        font-size: 1.05rem;
    }
    p{
        margin: 0;
        font-size: .82rem;
        opacity: .6;
        max-width: 24rem;
    }
}
.Bio{
    p{
        margin: 0;
        padding-left: 1rem;
        border-left: 2px solid #ffffff30;
        font-size: .88rem;
        line-height: 1.65;
        opacity: .85;
        overflow-wrap: anywhere;
        white-space: pre-line;
    }
}
.Meta{
    display: flex;
    flex-wrap: wrap;
    gap: .5rem;
    .Chip{
        display: flex;
        align-items: center;
        gap: .4rem;
        padding: .4rem .75rem;
        border-radius: 99rem;
        border: 1px solid #ffffff18;
        background: #ffffff08;
        font-size: .75rem;
        svg{
            width: .95rem;
            height: .95rem;
            opacity: .65;
        }
        &.on{
            border-color: #4caf5060;
            color: #a5e8a8;
        }
    }
}
.BtnPrimary{
    display: flex;
    justify-content: center;
    align-items: center;
    gap: .45rem;
    padding: .6rem 1.2rem;
    border-radius: .55rem;
    border: 1px solid #fff;
    background: #fff;
    color: #000;
    font-size: .82rem;
    font-weight: 700;
    font-family: inherit;
    text-decoration: none;
    cursor: pointer;
    transition: filter 150ms, opacity 150ms;
    svg{
        width: 1.05rem;
        height: 1.05rem;
    }
    &:hover:not(:disabled){
        filter: brightness(.85);
    }
    &:disabled{
        opacity: .6;
        cursor: wait;
    }
}
.GhostBtn{
    display: flex;
    justify-content: center;
    align-items: center;
    gap: .4rem;
    padding: .6rem 1.1rem;
    border-radius: .55rem;
    border: 1px solid #ffffff25;
    background: #ffffff08;
    color: #fff;
    font-size: .8rem;
    font-weight: 600;
    font-family: inherit;
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
    svg{
        flex-shrink: 0;
        width: 1.1rem;
        height: 1.1rem;
        color: #ff9d9d;
    }
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
.Equipo{
    display: flex;
    flex-direction: column;
    gap: .8rem;
    h2{
        margin: 0;
        font-family: 'Lexend';
        font-size: 1.1rem;
        font-weight: 700;
    }
    .Grid{
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
        gap: .7rem;
        .Pieza{
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: .3rem;
            padding: 1rem;
            border-radius: .8rem;
            border: 1px solid #ffffff14;
            background: #00000060;
            b{
                font-size: .8rem;
                font-family: 'Lexend';
                text-align: center;
            }
            small{
                font-size: .65rem;
                text-transform: uppercase;
                letter-spacing: .08em;
                opacity: .5;
            }
        }
    }
}
@media (max-width: 600px){
    .Community{
        padding: 4rem 1rem 0 1rem;
    }
    .Wrap{
        padding-bottom: 3rem;
    }
}
</style>
