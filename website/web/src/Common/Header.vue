<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { IconBrandGithub, IconMenu2, IconLogout, IconBell } from '@tabler/icons-vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { useNotifications } from '@/Auth/Composables/useNotifications';

const open = ref(false);
const router = useRouter();
const { autenticado, nombre, profile, logout } = useAuth();
const { noLeidas, refrescar } = useNotifications();

const inicial = computed(() => (nombre.value.trim().charAt(0) || '?').toUpperCase());
const insignia = computed(() => (noLeidas.value > 99 ? '99+' : `${noLeidas.value}`));

// Si el avatar no carga (URL rota o bucket sin lectura pública), se muestra
// la inicial en vez de una imagen rota.
const avatarRoto = ref(false);

watch(
    () => profile.value?.avatarUrl,
    () => {
        avatarRoto.value = false;
    },
);

async function salir() {
    open.value = false;
    await logout();
    await router.push('/');
}

onMounted(() => refrescar());

watch(autenticado, () => refrescar());
</script>

<template>
    <div class="Header">
        <RouterLink class="Brand" to="/" @click="open = false">
            <img src="../../assets/logo-step-white.png" alt="StepLauncher" loading="eager" decoding="async">
            <b>StepLauncher</b>
        </RouterLink>
        <div class="Links">
            <RouterLink to="/">Inicio</RouterLink>
            <RouterLink to="/download">Descarga</RouterLink>
            <RouterLink to="/changelog">Historial</RouterLink>
            <RouterLink to="/about">Acerca De</RouterLink>
            <a class="Github" href="https://github.com/NovaStepStudio/StepLauncher" target="_blank" rel="noopener">
                <IconBrandGithub stroke="2" />
                Github
            </a>
            <template v-if="autenticado">
                <RouterLink class="Bell" to="/dashboard?tab=notificaciones" aria-label="Notificaciones" title="Notificaciones">
                    <IconBell stroke="2" />
                    <i v-if="noLeidas > 0">{{ insignia }}</i>
                </RouterLink>
                <RouterLink class="Account" to="/dashboard" :title="nombre">
                    <span class="AvWrap" :class="{ on: profile?.isOnline }">
                        <img v-if="profile?.avatarUrl && !avatarRoto" :src="profile.avatarUrl" alt="" @error="avatarRoto = true">
                        <i v-else>{{ inicial }}</i>
                    </span>
                    <b>{{ nombre }}</b>
                </RouterLink>
                <button class="Exit" type="button" aria-label="Cerrar sesión" title="Cerrar sesión" @click="salir">
                    <IconLogout stroke="2" />
                </button>
            </template>
            <template v-else>
                <RouterLink class="Accent" to="/auth">Entrar</RouterLink>
            </template>
        </div>
        <button class="MenuBtn" @click="open = !open" aria-label="Menú">
            <IconMenu2 stroke="2" />
        </button>
        <div v-if="open" class="Menu">
            <RouterLink to="/" @click="open = false">Inicio</RouterLink>
            <RouterLink to="/download" @click="open = false">Descarga</RouterLink>
            <RouterLink to="/changelog" @click="open = false">Historial</RouterLink>
            <RouterLink to="/about" @click="open = false">Acerca De</RouterLink>
            <a href="https://github.com/NovaStepStudio/StepLauncher" target="_blank" rel="noopener">
                <IconBrandGithub stroke="2" />
                Github
            </a>
            <template v-if="autenticado">
                <RouterLink to="/dashboard" @click="open = false">Panel ({{ nombre }})</RouterLink>
                <RouterLink to="/dashboard?tab=notificaciones" @click="open = false">Notificaciones{{ noLeidas > 0 ? ` (${insignia})` : '' }}</RouterLink>
                <button class="ExitMenu" type="button" @click="salir">
                    <IconLogout stroke="2" />
                    Cerrar sesión
                </button>
            </template>
            <template v-else>
                <RouterLink to="/auth" @click="open = false">Entrar</RouterLink>
            </template>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Header{
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: .7rem 5rem;
    background: #000000a6;
    backdrop-filter: blur(12px);
    border-bottom: 1px solid #ffffff14;
    z-index: 100;
    box-sizing: border-box;
    div{
        display:flex;
        justify-content:center;
        align-items:center;
    }
    .Brand{
        display:flex;
        justify-content:center;
        align-items:center;
        gap: .5rem;
        text-decoration: none;
        color: #fff;
        img{
            width: 2rem;
            height: 2rem;
            object-fit: contain;
        }
        b{
            font-family: 'Lexend';
            font-size: 1rem;
            font-weight: 600;
        }
    }
    .Links{
        gap: .3rem;
        a{
            display:flex;
            justify-content:center;
            align-items:center;
            gap: .4rem;
            padding: .45rem .8rem;
            border-radius: .5rem;
            font-size: .82rem;
            color: #ffffffa6;
            text-decoration: none;
            transition: background 150ms, color 150ms;
            svg{
                width: 1rem;
                height: 1rem;
            }
            &:hover{
                background: #ffffff10;
                color: #fff;
            }
            &.router-link-exact-active{
                background: #ffffff10;
                color: #fff;
            }
        }
        .Accent{
            background: #fff;
            color: #000;
            font-weight: 600;
            &:hover{
                background: #ffffffd9;
                color: #000;
            }
            &.router-link-exact-active{
                background: #fff;
                color: #000;
            }
        }
        .Account{
            display: flex;
            justify-content: center;
            align-items: center;
            gap: .45rem;
            padding: .3rem .7rem .3rem .3rem;
            border-radius: 99rem;
            border: 1px solid #ffffff25;
            background: #ffffff0d;
            color: #fff;
            text-decoration: none;
            transition: background 150ms, border-color 150ms;
            img,
            i{
                width: 1.5rem;
                height: 1.5rem;
                border-radius: 99rem;
                object-fit: cover;
            }
            i{
                display: flex;
                justify-content: center;
                align-items: center;
                font-style: normal;
                font-family: 'Lexend';
                font-size: .7rem;
                font-weight: 700;
                background: #ffffff20;
                color: #fff;
            }
            .AvWrap{
                position: relative;
                display: flex;
                flex-shrink: 0;
                &.on{
                    img,
                    i{
                        outline: 2px solid #4caf50;
                        outline-offset: 1px;
                        box-shadow: 0 0 10px #4caf5066;
                    }
                }
            }
            b{
                max-width: 7rem;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
                font-size: .78rem;
                font-weight: 600;
            }
            &:hover{
                background: #ffffff14;
                border-color: #ffffff40;
            }
        }
        .Exit{
            display: flex;
            justify-content: center;
            align-items: center;
            width: 2rem;
            height: 2rem;
            border-radius: .5rem;
            border: 1px solid transparent;
            background: transparent;
            color: #ffffffa6;
            cursor: pointer;
            transition: background 150ms, color 150ms;
            svg{
                width: 1rem;
                height: 1rem;
            }
            &:hover{
                background: #ffffff10;
                color: #fff;
            }
        }
        .Bell{
            position: relative;
            i{
                position: absolute;
                top: .1rem;
                right: .1rem;
                display: flex;
                justify-content: center;
                align-items: center;
                min-width: 1rem;
                height: 1rem;
                padding: 0 .25rem;
                border-radius: 99rem;
                background: #fff;
                color: #000;
                font-style: normal;
                font-size: .58rem;
                font-weight: 800;
            }
        }
    }
    .MenuBtn{
        display: none;
        justify-content:center;
        align-items:center;
        width: 2.2rem;
        height: 2.2rem;
        border-radius: .5rem;
        border: 1px solid #ffffff25;
        background: #ffffff10;
        color: #fff;
        cursor: pointer;
        svg{
            width: 1.2rem;
            height: 1.2rem;
        }
    }
    .Menu{
        display: none;
    }
}
@media (max-width: 800px){
    .Header{
        padding: .7rem 1.5rem;
        .Links{
            display: none;
        }
        .MenuBtn{
            display: flex;
        }
        .Menu{
            position: absolute;
            top: 100%;
            left: 0;
            width: 100%;
            display: flex;
            flex-direction: column;
            align-items: stretch;
            gap: .2rem;
            padding: .6rem 1.5rem .8rem 1.5rem;
            background: #000000f2;
            border-bottom: 1px solid #ffffff14;
            box-sizing: border-box;
            a{
                display:flex;
                align-items:center;
                gap: .5rem;
                padding: .6rem .8rem;
                border-radius: .5rem;
                font-size: .85rem;
                color: #ffffffa6;
                text-decoration: none;
                svg{
                    width: 1rem;
                    height: 1rem;
                }
                &:hover{
                    background: #ffffff10;
                    color: #fff;
                }
                &.router-link-exact-active{
                    background: #ffffff10;
                    color: #fff;
                }
            }
            .ExitMenu{
                display: flex;
                align-items: center;
                gap: .5rem;
                padding: .6rem .8rem;
                border-radius: .5rem;
                border: 0;
                background: transparent;
                font-size: .85rem;
                font-family: inherit;
                color: #ffffffa6;
                cursor: pointer;
                text-align: left;
                svg{
                    width: 1rem;
                    height: 1rem;
                }
                &:hover{
                    background: #ffffff10;
                    color: #fff;
                }
            }
        }
    }
}
</style>
