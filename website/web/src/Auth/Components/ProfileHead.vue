// Cabecera de perfil compartida (panel propio y preview de comunidad):
// portada, avatar solapado, nombre, handle y stats en línea. La edición
// (avatar/banner) es opt-in vía `editable`; las acciones de la derecha
// van por slot para que cada página ponga las suyas.
<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { RouterLink } from 'vue-router';
import { IconPencil } from '@tabler/icons-vue';

export interface StatCabecera {
    valor: string;
    etiqueta: string;
    to?: string;
}

const props = withDefaults(
    defineProps<{
        bannerUrl?: string | null;
        avatarUrl?: string | null;
        inicial?: string;
        eyebrow?: string;
        titulo?: string;
        handle?: string | null;
        stats?: StatCabecera[];
        editable?: boolean;
        textoBanner?: string;
        deshabilitado?: boolean;
        tamano?: 'grande' | 'compacto';
        enLinea?: boolean | null;
    }>(),
    {
        bannerUrl: null,
        avatarUrl: null,
        inicial: '?',
        eyebrow: '',
        titulo: '',
        handle: null,
        stats: () => [],
        editable: false,
        textoBanner: 'Banner',
        deshabilitado: false,
        tamano: 'grande',
        enLinea: null,
    },
);

const emit = defineEmits<{
    (e: 'editar-avatar'): void;
    (e: 'editar-banner'): void;
}>();

// Si el avatar no carga (URL rota), se muestra la inicial.
const roto = ref(false);

watch(
    () => props.avatarUrl,
    () => {
        roto.value = false;
    },
);

const conBanner = computed(() => !!props.bannerUrl);
</script>

<template>
    <div class="ProfileHead" :class="{ compacto: tamano === 'compacto' }">
        <div class="Cover">
            <img v-if="conBanner" :src="bannerUrl || ''" alt="" loading="eager" decoding="async">
            <div v-else class="CoverGlow"></div>
            <div class="Scrim"></div>
            <button
                v-if="editable"
                type="button"
                class="EditBanner"
                :disabled="deshabilitado"
                @click="emit('editar-banner')"
            >
                <IconPencil stroke="2" />
                {{ textoBanner }}
            </button>
        </div>
        <div class="IdentityRow">
            <div class="Avatar" :class="{ on: enLinea }">
                <img v-if="avatarUrl && !roto" :src="avatarUrl" alt="" @error="roto = true">
                <span v-else>{{ inicial }}</span>
                <button
                    v-if="editable"
                    type="button"
                    class="EditAvatar"
                    aria-label="Cambiar avatar"
                    title="Cambiar avatar"
                    :disabled="deshabilitado"
                    @click="emit('editar-avatar')"
                >
                    <IconPencil stroke="2" />
                </button>
            </div>
            <div class="Who">
                <small v-if="eyebrow" class="Eyebrow">{{ eyebrow }}</small>
                <h1>{{ titulo }}</h1>
                <span v-if="handle" class="Handle">{{ handle }}</span>
            </div>
            <div class="Acciones">
                <slot name="acciones"></slot>
            </div>
        </div>
        <nav v-if="stats.length > 0" class="StatsInline" aria-label="Resumen">
            <template v-for="(s, i) in stats" :key="`${s.etiqueta}-${i}`">
                <i v-if="i > 0"></i>
                <RouterLink v-if="s.to" :to="s.to"><b>{{ s.valor }}</b> {{ s.etiqueta }}</RouterLink>
                <span v-else><b>{{ s.valor }}</b> {{ s.etiqueta }}</span>
            </template>
        </nav>
    </div>
</template>

<style scoped lang="scss">
.ProfileHead{
    width: 100%;
    .Cover{
        position: relative;
        height: clamp(15rem, 36dvh, 24rem);
        border-radius: 1rem;
        border: 1px solid #ffffff2e;
        overflow: hidden;
        background: #0b0b0f;
        box-shadow: 0 12px 48px #00000080;
        img{
            width: 100%;
            height: 100%;
            object-fit: cover;
        }
        .CoverGlow{
            position: absolute;
            inset: 0;
            background:
                radial-gradient(30rem 10rem at 15% 0%, #ffffff16, transparent),
                radial-gradient(26rem 12rem at 90% 100%, #ffffff0d, transparent);
        }
        .Scrim{
            position: absolute;
            inset: 0;
            background: linear-gradient(transparent 55%, #000000a6);
            pointer-events: none;
        }
    }
    &.compacto{
        .Cover{
            height: clamp(11rem, 28dvh, 17rem);
        }
        .IdentityRow{
            gap: 1.4rem;
            .Avatar{
                width: 6.5rem;
                height: 6.5rem;
                span{
                    font-size: 2.2rem;
                }
            }
            .Who{
                h1{
                    font-size: 2.1rem;
                }
            }
        }
    }
    .EditBanner{
        position: absolute;
        right: .9rem;
        bottom: .9rem;
        display: flex;
        justify-content: center;
        align-items: center;
        gap: .4rem;
        padding: .45rem .85rem;
        border-radius: 99rem;
        border: 1px solid #ffffff25;
        background: #000000b3;
        backdrop-filter: blur(8px);
        color: #fff;
        font-size: .72rem;
        font-weight: 600;
        font-family: inherit;
        cursor: pointer;
        transition: background 150ms, opacity 150ms;
        svg{
            width: .95rem;
            height: .95rem;
        }
        &:hover:not(:disabled){
            background: #ffffff1c;
        }
        &:disabled{
            opacity: .5;
            cursor: wait;
        }
    }
    .IdentityRow{
        display: flex;
        align-items: flex-end;
        gap: 1.2rem;
        padding: 0 .5rem;
        .Avatar{
            position: relative;
            display: flex;
            justify-content: center;
            align-items: center;
            flex-shrink: 0;
            width: 5.6rem;
            height: 5.6rem;
            margin-top: .5rem;
            border-radius: 99rem;
            overflow: hidden;
            border: 3px solid #000;
            outline: 1px solid #ffffff30;
            background: #ffffff10;
            box-shadow: 0 8px 28px #000000cc;
            transition: outline-color 150ms, box-shadow 150ms;
            &.on{
                outline-color: #4caf50;
                box-shadow: 0 0 22px #4caf5066, 0 8px 28px #000000cc;
            }
            img{
                width: 100%;
                height: 100%;
                object-fit: cover;
            }
            span{
                font-family: 'Lexend';
                font-size: 2rem;
                font-weight: 700;
                opacity: .85;
            }
            .EditAvatar{
                position: absolute;
                inset: auto 0 0 0;
                display: flex;
                justify-content: center;
                align-items: center;
                padding: .3rem 0 .5rem 0;
                border: 0;
                background: #000000b3;
                backdrop-filter: blur(4px);
                color: #fff;
                cursor: pointer;
                opacity: 0;
                transition: opacity 150ms, background 150ms;
                svg{
                    width: .95rem;
                    height: .95rem;
                }
                &:hover:not(:disabled){
                    background: #ffffff30;
                }
                &:disabled{
                    cursor: wait;
                }
            }
            &:hover .EditAvatar,
            &:focus-within .EditAvatar{
                opacity: 1;
            }
        }
        .Who{
            display: flex;
            flex-direction: column;
            align-items: flex-start;
            gap: .25rem;
            flex: 1;
            min-width: 0;
            .Eyebrow{
                font-size: .65rem;
                font-weight: 700;
                text-transform: uppercase;
                letter-spacing: .14em;
                padding: .2rem .6rem;
                border-radius: 99rem;
                border: 1px solid #ffffff25;
                background: #ffffff0d;
                opacity: .8;
            }
            h1{
                margin: 0;
                font-family: 'Lexend';
                font-size: 1.9rem;
                font-weight: 700;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
                max-width: 100%;
            }
            .Handle{
                font-size: .8rem;
                font-weight: 600;
                opacity: .55;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
                max-width: 100%;
            }
        }
        .Acciones{
            display: flex;
            align-items: center;
            flex-wrap: wrap;
            gap: .5rem;
            flex-shrink: 0;
            padding-bottom: .2rem;
        }
    }
    .StatsInline{
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        gap: .4rem .7rem;
        padding: .9rem .5rem 0 .5rem;
        font-size: .82rem;
        a{
            color: #fff;
            text-decoration: none;
            border-bottom: 1px solid transparent;
            transition: border-color 150ms;
            &:hover{
                border-color: #ffffff60;
            }
        }
        span{
            opacity: .55;
        }
        b{
            font-family: 'Lexend';
            font-weight: 700;
        }
        i{
            width: 3px;
            height: 3px;
            flex-shrink: 0;
            border-radius: 99rem;
            background: #ffffff40;
        }
    }
}
@media (hover: none){
    .ProfileHead{
        .IdentityRow{
            .Avatar{
                .EditAvatar{
                    opacity: 1;
                }
            }
        }
    }
}
@media (max-width: 600px){
    .ProfileHead{
        .Cover{
            height: 11rem;
        }
        &.compacto{
            .Cover{
                height: 9rem;
            }
        }
        .IdentityRow{
            flex-wrap: wrap;
            .Who{
                flex: 1 1 12rem;
                h1{
                    font-size: 1.5rem;
                    white-space: normal;
                }
            }
        }
        .Acciones{
            width: 100%;
        }
    }
}
</style>
