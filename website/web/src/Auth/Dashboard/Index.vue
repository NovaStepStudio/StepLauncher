// Panel de la cuenta con estética propia de app (sin hero tipo landing):
// cabecera compacta con avatar, saludo y sesión, glow superior sutil y
// subtabs deslizantes. Ruta protegida por el guard del router.
<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { IconUser, IconShieldCheck, IconShirt, IconUsers, IconBell, IconLogout, IconPencil, IconAlertCircle, IconCircleCheck } from '@tabler/icons-vue';
import SliderTabs from '@/Auth/Components/SliderTabs.vue';
import ProfileHead, { type StatCabecera } from '@/Auth/Components/ProfileHead.vue';
import ProfileCard from './Components/ProfileCard.vue';
import SecurityCard from './Components/SecurityCard.vue';
import CosmeticosCard from './Components/FilesCard.vue';
import AmigosCard from './Components/FriendsCard.vue';
import AvisosCard from './Components/NotificationsCard.vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { useNotifications } from '@/Auth/Composables/useNotifications';
import { subirAvatar, subirBanner, listarAmigos } from '@/Auth/Api';
import { ApiError } from '@/Auth/Api';

type SubTab = 'perfil' | 'seguridad' | 'cosmeticos' | 'amigos' | 'notificaciones';

const pestanas = [
    { key: 'perfil', label: 'Perfil', icon: IconUser },
    { key: 'seguridad', label: 'Seguridad', icon: IconShieldCheck },
    { key: 'cosmeticos', label: 'Cosméticos', icon: IconShirt },
    { key: 'amigos', label: 'Amigos', icon: IconUsers },
    { key: 'notificaciones', label: 'Notificaciones', icon: IconBell },
] as const;

const VALIDAS: SubTab[] = ['perfil', 'seguridad', 'cosmeticos', 'amigos', 'notificaciones'];

const SECCIONES: Record<SubTab, { titulo: string; desc: string }> = {
    perfil: { titulo: 'Tu perfil', desc: 'Cómo te ve el resto y tus datos de jugador.' },
    seguridad: { titulo: 'Seguridad', desc: 'Correo, contraseña y quién puede encontrarte.' },
    cosmeticos: { titulo: 'Cosméticos', desc: 'Tus skins y capas, en 3D.' },
    amigos: { titulo: 'Amigos', desc: 'Buscá jugadores y gestioná amistades.' },
    notificaciones: { titulo: 'Notificaciones', desc: 'Lo último que pasó en tu cuenta.' },
};

const seccion = computed(() => SECCIONES[tab.value]);
const insigniaAvisos = computed(() => (noLeidas.value > 99 ? '99+' : `${noLeidas.value}`));

function insigniaDe(key: SubTab): string {
    if (key === 'amigos') return totalAmigos.value > 0 ? `${totalAmigos.value}` : '';
    if (key === 'notificaciones') return noLeidas.value > 0 ? insigniaAvisos.value : '';
    return '';
}

function tabDe(query: unknown): SubTab {
    if (query === 'avisos') return 'notificaciones';
    return typeof query === 'string' && (VALIDAS as string[]).includes(query) ? (query as SubTab) : 'perfil';
}

const ruta = useRoute();
const router = useRouter();
const { nombre, profile, refrescarPerfil, logout, conAuth } = useAuth();
const { noLeidas, refrescar: refrescarAvisos } = useNotifications();

const totalAmigos = ref(0);
const mcJugado = computed(() => profile.value?.lastMcVersion || '—');

const tab = ref<SubTab>(tabDe(ruta.query.tab));

watch(
    () => ruta.query.tab,
    (nueva) => {
        tab.value = tabDe(nueva);
    },
);

watch(tab, (nueva) => {
    router.replace({ path: '/dashboard', query: { ...ruta.query, tab: nueva } });
});

const inicial = computed(() => (nombre.value.trim().charAt(0) || '?').toUpperCase());
const bannerFondo = computed(() => profile.value?.bannerUrl || '');

const statsCabecera = computed<StatCabecera[]>(() => [
    { valor: `${totalAmigos.value}`, etiqueta: totalAmigos.value === 1 ? 'amigo' : 'amigos', to: '/dashboard?tab=amigos' },
    { valor: `${noLeidas.value}`, etiqueta: noLeidas.value === 1 ? 'notificación' : 'notificaciones', to: '/dashboard?tab=notificaciones' },
    { valor: mcJugado.value, etiqueta: 'último MC' },
]);

const cargando = ref(true);
const error = ref('');
const saliendo = ref(false);

onMounted(async () => {
    try {
        await refrescarPerfil();
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        cargando.value = false;
    }
    // Stats de la cabecera: mejor esfuerzo, no bloquean el panel.
    try {
        const [ami] = await Promise.all([
            conAuth((token) => listarAmigos(token)),
            refrescarAvisos(),
        ]);
        totalAmigos.value = ami.count;
    } catch {
        // Se conservan los ceros.
    }
});

async function salir() {
    saliendo.value = true;
    try {
        await logout();
        await router.push('/');
    } finally {
        saliendo.value = false;
    }
}

// --- Avatar y banner directo desde la cabecera (sin tabs) ---------------------

const inputAvatar = ref<HTMLInputElement | null>(null);
const inputBanner = ref<HTMLInputElement | null>(null);
const mensajeImg = ref('');
const errorImg = ref('');
const subiendoImg = ref(false);

const TIPOS_AVATAR = ['image/png', 'image/jpeg', 'image/webp'];
const TIPOS_BANNER = ['image/png', 'image/gif', 'image/jpeg', 'image/webp'];

function dimensiones(archivo: File): Promise<{ w: number; h: number }> {
    return new Promise((resolve, reject) => {
        const url = URL.createObjectURL(archivo);
        const img = new Image();
        img.onload = () => {
            const dims = { w: img.naturalWidth, h: img.naturalHeight };
            URL.revokeObjectURL(url);
            resolve(dims);
        };
        img.onerror = () => {
            URL.revokeObjectURL(url);
            reject(new Error('no-imagen'));
        };
        img.src = url;
    });
}

async function elegirAvatar(evento: Event) {
    const input = evento.target as HTMLInputElement;
    const archivo = input.files && input.files[0] ? input.files[0] : null;
    input.value = '';
    mensajeImg.value = '';
    errorImg.value = '';
    if (!archivo) return;
    if (!TIPOS_AVATAR.includes(archivo.type)) {
        errorImg.value = 'Solo se permiten imágenes PNG, JPEG o WebP.';
        return;
    }
    if (archivo.size > 2 * 1024 * 1024) {
        errorImg.value = 'El avatar no puede superar los 2 MB.';
        return;
    }
    subiendoImg.value = true;
    try {
        const res = await conAuth((token) => subirAvatar(token, archivo));
        if (profile.value) profile.value.avatarUrl = res.avatarUrl;
        mensajeImg.value = 'Avatar actualizado.';
    } catch (err) {
        errorImg.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        subiendoImg.value = false;
    }
}

async function elegirBanner(evento: Event) {
    const input = evento.target as HTMLInputElement;
    const archivo = input.files && input.files[0] ? input.files[0] : null;
    input.value = '';
    mensajeImg.value = '';
    errorImg.value = '';
    if (!archivo) return;
    if (!TIPOS_BANNER.includes(archivo.type)) {
        errorImg.value = 'Solo se permiten imágenes PNG, GIF, JPEG o WebP.';
        return;
    }
    if (archivo.size > 8 * 1024 * 1024) {
        errorImg.value = 'El banner no puede superar los 8 MB.';
        return;
    }
    try {
        const dims = await dimensiones(archivo);
        if (dims.w > 1920 || dims.h > 1080) {
            errorImg.value = `El banner no puede superar 1920×1080 (el tuyo es de ${dims.w}×${dims.h}).`;
            return;
        }
    } catch {
        errorImg.value = 'No se pudo leer esa imagen.';
        return;
    }
    subiendoImg.value = true;
    try {
        const res = await conAuth((token) => subirBanner(token, archivo));
        if (profile.value) profile.value.bannerUrl = res.bannerUrl;
        mensajeImg.value = 'Banner actualizado.';
    } catch (err) {
        errorImg.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        subiendoImg.value = false;
    }
}
</script>

<template>
    <div class="Dashboard">
        <div class="DashHead sl-enter" style="--sl-delay: 0s;">
        <ProfileHead
            tamano="compacto"
            :banner-url="bannerFondo || null"
            :avatar-url="profile?.avatarUrl || null"
            :inicial="inicial"
            eyebrow="Mi cuenta"
            :titulo="cargando ? 'Cargando...' : nombre ? `Hola, ${nombre}` : 'Tu panel'"
            :handle="profile?.username ? `@${profile.username}` : null"
            :stats="statsCabecera"
            :editable="true"
            :en-linea="profile?.isOnline ?? null"
            :texto-banner="bannerFondo ? 'Banner' : 'Añadir banner'"
            :deshabilitado="subiendoImg || cargando"
            @editar-avatar="inputAvatar?.click()"
            @editar-banner="inputBanner?.click()"
        >
            <template #acciones>
                <button type="button" class="LogoutBtn" :disabled="saliendo || cargando" @click="salir">
                    <IconLogout stroke="2" />
                    {{ saliendo ? 'Cerrando...' : 'Salir' }}
                </button>
            </template>
        </ProfileHead>
        </div>
        <input ref="inputAvatar" type="file" accept="image/png,image/jpeg,image/webp" hidden @change="elegirAvatar">
        <input ref="inputBanner" type="file" accept="image/png,image/gif,image/jpeg,image/webp" hidden @change="elegirBanner">
        <div class="Body" v-reveal>
            <p v-if="cargando" class="State">Cargando tu cuenta...</p>
            <p v-else-if="error" class="FormError" role="alert">
                <IconAlertCircle stroke="2" />
                {{ error }}
            </p>
            <template v-else>
                <div class="Layout">
                    <nav class="Rail" aria-label="Secciones de la cuenta">
                        <button
                            v-for="p in pestanas"
                            :key="p.key"
                            type="button"
                            :class="{ active: tab === p.key }"
                            :aria-current="tab === p.key ? 'page' : undefined"
                            @click="tab = p.key"
                        >
                            <component :is="p.icon" stroke="2" />
                            {{ p.label }}
                            <i v-if="insigniaDe(p.key)" class="Badge">{{ insigniaDe(p.key) }}</i>
                        </button>
                    </nav>
                    <SliderTabs class="Bar" v-model="tab" :tabs="pestanas" />
                    <div class="Content">
                        <div class="SectionHead">
                            <h2>{{ seccion.titulo }}</h2>
                            <p>{{ seccion.desc }}</p>
                        </div>
                        <div class="TabBody sl-fade" :key="tab" role="tabpanel">
                            <ProfileCard v-if="tab === 'perfil'" />
                            <SecurityCard v-else-if="tab === 'seguridad'" />
                            <CosmeticosCard v-else-if="tab === 'cosmeticos'" />
                            <AmigosCard v-else-if="tab === 'amigos'" />
                            <AvisosCard v-else />
                        </div>
                        <p v-if="errorImg" class="FormError" role="alert">
                            <IconAlertCircle stroke="2" />
                            {{ errorImg }}
                        </p>
                        <p v-if="mensajeImg" class="Ok" role="status">
                            <IconCircleCheck stroke="2" />
                            {{ mensajeImg }}
                        </p>
                    </div>
                </div>
            </template>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Dashboard{
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    background: #000;
    overflow: hidden;
    padding: 4.5rem 1.5rem 0 1.5rem;
    &::before{
        content: '';
        position: absolute;
        top: 0;
        left: 50%;
        width: min(70rem, 120dvw);
        height: 22rem;
        transform: translateX(-50%);
        background: radial-gradient(closest-side, #ffffff17, transparent);
        pointer-events: none;
    }
}
.DashHead{
    width: 100%;
    max-width: 70rem;
    .LogoutBtn{
        display: flex;
        justify-content: center;
        align-items: center;
        gap: .45rem;
        flex-shrink: 0;
        padding: .6rem 1.1rem;
        border-radius: .55rem;
        border: 1px solid #ff9d9d45;
        background: transparent;
        color: #ffb3b3;
        font-size: .8rem;
        font-weight: 600;
        font-family: inherit;
        cursor: pointer;
        transition: background 150ms, opacity 150ms;
        svg{
            width: 1.05rem;
            height: 1.05rem;
        }
        &:hover:not(:disabled){
            background: #ff6b6b1c;
        }
        &:disabled{
            opacity: .5;
            cursor: wait;
        }
    }
}
@media (hover: none){
    .DashHead{
        .IdentityRow{
            .Avatar{
                .EditAvatar{
                    opacity: 1;
                }
            }
        }
    }
}
.Body{
    position: relative;
    width: 100%;
    max-width: 70rem;
    display: flex;
    flex-direction: column;
    align-items: stretch;
    padding: 0 0 4rem 0;
    .State{
        text-align: center;
        font-size: .85rem;
        opacity: .55;
    }
    .Layout{
        display: grid;
        grid-template-columns: 12.5rem minmax(0, 1fr);
        gap: 2.5rem;
        align-items: start;
        width: 100%;
        padding-top: 1.8rem;
        border-top: 1px solid #ffffff10;
        margin-top: 1.8rem;
    }
    .Rail{
        position: sticky;
        top: 4.5rem;
        display: flex;
        flex-direction: column;
        gap: .15rem;
        padding: .4rem;
        border-radius: .8rem;
        border: 1px solid #ffffff14;
        background: #ffffff05;
        button{
            display: flex;
            align-items: center;
            gap: .6rem;
            width: 100%;
            padding: .6rem .8rem;
            border-radius: .55rem;
            border: 0;
            background: transparent;
            color: #ffffffa6;
            font-size: .82rem;
            font-weight: 600;
            font-family: 'Lexend';
            cursor: pointer;
            text-align: left;
            transition: background 150ms, color 150ms, box-shadow 150ms;
            svg{
                width: 1.05rem;
                height: 1.05rem;
                flex-shrink: 0;
                opacity: .75;
            }
            .Badge{
                margin-left: auto;
                font-style: normal;
                font-size: .62rem;
                font-weight: 800;
                min-width: 1.25rem;
                height: 1.25rem;
                display: flex;
                justify-content: center;
                align-items: center;
                padding: 0 .3rem;
                border-radius: 99rem;
                background: #fff;
                color: #000;
            }
            &:hover{
                background: #ffffff0d;
                color: #fff;
            }
            &.active{
                background: #ffffff10;
                color: #fff;
                box-shadow: inset 2px 0 0 #fff;
            }
        }
    }
    .Bar{
        display: none;
    }
    .Content{
        display: flex;
        flex-direction: column;
        gap: 1.2rem;
        min-width: 0;
    }
    .SectionHead{
        h2{
            margin: 0;
            font-family: 'Lexend';
            font-size: 1.5rem;
            font-weight: 700;
        }
        p{
            margin: .3rem 0 0 0;
            font-size: .85rem;
            opacity: .55;
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
        justify-content: center;
        gap: .5rem;
        margin: 0;
        font-size: .8rem;
        color: #9dffb0;
        svg{
            width: 1.1rem;
            height: 1.1rem;
        }
    }
    .TabBody{
        width: 100%;
    }
}
@media (max-width: 900px){
    .Body{
        .Layout{
            grid-template-columns: minmax(0, 1fr);
            gap: 1.2rem;
        }
        .Rail{
            display: none;
        }
        .Bar{
            display: grid;
        }
    }
}
@media (max-width: 600px){
    .Dashboard{
        padding: 4rem 1rem 0 1rem;
    }
    .DashHead{
        .LogoutBtn{
            width: 100%;
        }
    }
}
</style>
