<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { IconHistory, IconBrandGithub, IconListDetails, IconSearch, IconChevronLeft, IconChevronRight } from '@tabler/icons-vue';
import { channelOf, useReleases, type ReleaseChannel } from '@/Common/Composables/useReleases';
import ReleaseNotes from './Components/ReleaseNotes.vue';

const { releases, loading, error, load } = useReleases();

type Filtro = 'all' | ReleaseChannel;
const filtro = ref<Filtro>('all');
const busqueda = ref('');

const filtros: Array<{ key: Filtro; label: string }> = [
    { key: 'all', label: 'Todas' },
    { key: 'stable', label: 'Estables' },
    { key: 'beta', label: 'Betas' },
    { key: 'alpha', label: 'Alphas' },
];

const visibles = computed(() => {
    const q = busqueda.value.trim().toLowerCase();
    return releases.value.filter((r) => {
        if (r.draft) return false;
        if (filtro.value !== 'all' && channelOf(r.tag, r.prerelease) !== filtro.value) return false;
        if (q && !`${r.tag} ${r.name}`.toLowerCase().includes(q)) return false;
        return true;
    });
});

const sinReleases = computed(() => releases.value.filter((r) => !r.draft).length === 0);

// Paginación: 5 por página para que 500 versiones no maten el scroll.
const PAGE_SIZE = 5;
const pagina = ref(1);
const totalPaginas = computed(() => Math.max(1, Math.ceil(visibles.value.length / PAGE_SIZE)));
const enPagina = computed(() => {
    const actual = Math.min(pagina.value, totalPaginas.value);
    return visibles.value.slice((actual - 1) * PAGE_SIZE, actual * PAGE_SIZE);
});

// Ventana de hasta 5 números alrededor de la página actual.
const paginasVisibles = computed(() => {
    const total = totalPaginas.value;
    const actual = Math.min(pagina.value, total);
    const fin = Math.min(total, Math.max(5, actual + 2));
    const inicio = Math.max(1, fin - 4);
    const arr: number[] = [];
    for (let i = inicio; i <= fin; i++) arr.push(i);
    return arr;
});

// Al filtrar o buscar se vuelve a la primera página.
watch([filtro, busqueda], () => {
    pagina.value = 1;
});

function irAPagina(n: number) {
    pagina.value = Math.min(Math.max(1, n), totalPaginas.value);
    scrollToList();
}

// Scroll suave al listado sin tocar la URL (nada de #hash).
function scrollToList() {
    document.getElementById('history')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

onMounted(() => load());
</script>

<template>
    <div class="Changelog">
        <div class="FirstPrew">
            <div class="TextAndButtons">
                <div class="AppName sl-enter" style="--sl-delay: 0s">
                    <img class="sl-float" src="../../assets/logo-step-white.png" alt="StepLauncher" loading="eager" decoding="async" fetchpriority="high">
                    <h1>Historial</h1>
                </div>
                <div class="Badges sl-enter" style="--sl-delay: .08s">
                    <span v-if="loading" class="Badge">Buscando releases...</span>
                    <span v-else class="BadgeMain">{{ visibles.length }} {{ visibles.length === 1 ? 'versión' : 'versiones' }}</span>
                    <span class="Badge">Notas de GitHub</span>
                </div>
                <div class="Description sl-enter" style="--sl-delay: .16s">
                    <h2>Cada versión, con sus notas.</h2>
                    <p>El historial completo de actualizaciones con el markdown original de cada release, directo de GitHub.</p>
                </div>
                <div class="Buttons sl-enter" style="--sl-delay: .24s">
                    <button class="BtnPrimary" type="button" @click="scrollToList">
                        <IconListDetails stroke="2" />
                        Ver historial
                    </button>
                    <a class="Btn" href="https://github.com/NovaStepStudio/StepLauncher/releases" target="_blank" rel="noopener">
                        <IconBrandGithub stroke="2" />
                        Todos los releases
                    </a>
                </div>
                <div class="MiniInfo sl-enter" style="--sl-delay: .32s">
                    <span>Gratis y open source</span>
                    <i></i>
                    <span>Publicado en GitHub</span>
                </div>
            </div>
        </div>
        <div class="History" id="history" v-reveal>
            <div class="Head">
                <div class="Tag">
                    <IconHistory stroke="2" />
                    <span>Actualizaciones</span>
                </div>
                <div class="Description">
                    <h2>Qué cambió en cada versión.</h2>
                    <p>Tocá una versión para leer sus notas completas.</p>
                </div>
            </div>
            <p v-if="loading" class="State">Buscando el historial en GitHub...</p>
            <div v-else-if="error" class="StateCard">
                <b>{{ error }}</b>
                <a class="Btn" href="https://github.com/NovaStepStudio/StepLauncher/releases" target="_blank" rel="noopener">Ir a los releases</a>
            </div>
            <template v-else>
                <div class="Filters" role="tablist" aria-label="Filtrar por canal">
                    <button
                        v-for="f in filtros"
                        :key="f.key"
                        role="tab"
                        type="button"
                        class="Filter"
                        :class="{ active: filtro === f.key }"
                        :aria-selected="filtro === f.key"
                        @click="filtro = f.key"
                    >
                        {{ f.label }}
                    </button>
                    <label class="Search">
                        <IconSearch stroke="2" />
                        <input v-model="busqueda" type="search" placeholder="Buscar versión..." aria-label="Buscar versión">
                    </label>
                </div>
                <div v-if="visibles.length > 0" class="Notes sl-stagger">
                    <ReleaseNotes v-for="(r, i) in enPagina" :key="r.tag" :release="r" :open="i === 0" />
                </div>
                <nav v-if="totalPaginas > 1" class="Pagination" aria-label="Paginar historial">
                    <button
                        type="button"
                        class="PageBtn"
                        :disabled="pagina <= 1"
                        aria-label="Página anterior"
                        @click="irAPagina(pagina - 1)"
                    >
                        <IconChevronLeft stroke="2" />
                    </button>
                    <button
                        v-for="n in paginasVisibles"
                        :key="n"
                        type="button"
                        class="PageBtn Num"
                        :class="{ active: n === pagina }"
                        :aria-current="n === pagina ? 'page' : undefined"
                        @click="irAPagina(n)"
                    >
                        {{ n }}
                    </button>
                    <button
                        type="button"
                        class="PageBtn"
                        :disabled="pagina >= totalPaginas"
                        aria-label="Página siguiente"
                        @click="irAPagina(pagina + 1)"
                    >
                        <IconChevronRight stroke="2" />
                    </button>
                </nav>
                <div v-else class="StateCard">
                    <b>{{ sinReleases ? 'Todavía no hay releases publicados.' : 'Ninguna versión coincide con ese filtro.' }}</b>
                    <a v-if="sinReleases" class="Btn" href="https://github.com/NovaStepStudio/StepLauncher" target="_blank" rel="noopener">Ver el repo</a>
                    <p v-else>Probá con otro canal o limpiá la búsqueda.</p>
                </div>
            </template>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Changelog{
    display: flex;
    flex-direction: column;
    background: #000;
}
.FirstPrew{
    position: relative;
    width: 100%;
    min-height: 58dvh;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-top: 4rem;
    z-index: 1;
    &::after{
        content: '';
        position: absolute;
        inset: 0;
        width: 100%;
        height: 100%;
        mask: linear-gradient(#0008 50%, transparent);
        background: linear-gradient(#000000b3, #000000b3), url('../../assets/background/4.webp');
        background-position: center center;
        background-size: cover;
        z-index: -1;
    }
    div{
        display: flex;
        justify-content: center;
        align-items: center;
    }
    .TextAndButtons{
        width: 100%;
        flex-direction: column;
        align-items: center;
        text-align: center;
        gap: 1rem;
        padding: 0 1.5rem;
        .AppName{
            gap: .5rem;
            font-family: 'Lexend';
            img{
                width: 4rem;
            }
            h1{
                margin: 0;
                font-size: 2.4rem;
            }
        }
        .Badges{
            gap: .5rem;
            .BadgeMain,
            .Badge{
                font-size: .7rem;
                padding: .25rem .65rem;
                border-radius: 99rem;
                border: 1px solid #ffffff25;
                background: #ffffff0d;
            }
            .BadgeMain{
                background: #fff;
                color: #000;
                font-weight: 700;
                border-color: #fff;
            }
        }
        .Description{
            flex-direction: column;
            align-items: center;
            gap: .5rem;
            max-width: 30rem;
            h2{
                margin: 0;
                font-family: 'Lexend';
                font-size: 1.6rem;
                font-weight: 600;
            }
            p{
                margin: 0;
                font-size: .9rem;
                line-height: 1.6;
                opacity: .65;
            }
        }
        .Buttons{
            gap: .6rem;
            a,
            button{
                display: flex;
                justify-content: center;
                align-items: center;
                gap: .45rem;
                padding: .6rem 1.3rem;
                border-radius: .5rem;
                font-size: .85rem;
                font-weight: 600;
                font-family: inherit;
                text-decoration: none;
                cursor: pointer;
                transition: filter 150ms, background 150ms, transform 150ms;
                svg{
                    width: 1.1rem;
                    height: 1.1rem;
                }
            }
            .BtnPrimary{
                background: #fff;
                color: #000;
                border: 1px solid #fff;
                &:hover{
                    filter: brightness(.85);
                    transform: translateY(-1px);
                }
            }
            .Btn{
                background: #ffffff10;
                color: #fff;
                border: 1px solid #ffffff25;
                &:hover{
                    background: #ffffff1c;
                    transform: translateY(-1px);
                }
            }
        }
        .MiniInfo{
            gap: .6rem;
            font-size: .72rem;
            opacity: .5;
            i{
                width: 3px;
                height: 3px;
                border-radius: 99rem;
                background: #fff;
            }
        }
    }
}
.History{
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 2.5rem 1.5rem 4rem 1.5rem;
    gap: 1.5rem;
    .Head{
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;
        text-align: center;
        .Tag{
            display: flex;
            justify-content: center;
            align-items: center;
            gap: .4rem;
            font-size: .7rem;
            text-transform: uppercase;
            letter-spacing: .12em;
            padding: .25rem .65rem;
            border-radius: 99rem;
            border: 1px solid #ffffff25;
            background: #ffffff0d;
            opacity: .85;
            svg{
                width: .95rem;
                height: .95rem;
            }
        }
        .Description{
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: .5rem;
            h2{
                margin: 0;
                font-family: 'Lexend';
                font-size: 1.6rem;
                font-weight: 600;
            }
            p{
                margin: 0;
                font-size: .9rem;
                line-height: 1.6;
                opacity: .65;
            }
        }
    }
    .State{
        font-size: .85rem;
        opacity: .55;
    }
    .Filters{
        display: flex;
        justify-content: center;
        align-items: center;
        flex-wrap: wrap;
        gap: .5rem;
        width: 100%;
        max-width: 52rem;
        .Filter{
            padding: .45rem 1rem;
            border-radius: 99rem;
            border: 1px solid #ffffff25;
            background: #ffffff08;
            color: #ffffffa6;
            font-size: .78rem;
            font-weight: 600;
            font-family: 'Lexend';
            cursor: pointer;
            transition: background 150ms, color 150ms, border-color 150ms;
            &:hover{
                background: #ffffff14;
                color: #fff;
            }
            &.active{
                background: #fff;
                color: #000;
                border-color: #fff;
            }
        }
        .Search{
            display: flex;
            align-items: center;
            gap: .45rem;
            padding: .45rem .9rem;
            border-radius: 99rem;
            border: 1px solid #ffffff25;
            background: #ffffff08;
            transition: border-color 150ms;
            &:focus-within{
                border-color: #ffffff50;
            }
            svg{
                width: 1rem;
                height: 1rem;
                flex-shrink: 0;
                opacity: .55;
            }
            input{
                width: 9rem;
                border: 0;
                outline: 0;
                background: transparent;
                color: #fff;
                font-size: .78rem;
                font-family: inherit;
                &::placeholder{
                    color: #ffffff55;
                }
                &::-webkit-search-cancel-button{
                    cursor: pointer;
                }
            }
        }
    }
    .StateCard{
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;
        padding: 1.5rem 2rem;
        border-radius: .9rem;
        border: 1px solid #ffffff18;
        background: transparent;
        b{
            font-size: .85rem;
            font-family: 'Lexend';
        }
        p{
            margin: 0;
            font-size: .8rem;
            opacity: .55;
        }
        .Btn{
            padding: .6rem 1.3rem;
            border-radius: .5rem;
            font-size: .85rem;
            font-weight: 600;
            text-decoration: none;
            background: #ffffff10;
            color: #fff;
            border: 1px solid #ffffff25;
            &:hover{
                background: #ffffff1c;
            }
        }
    }
    .Notes{
        display: flex;
        flex-direction: column;
        align-items: stretch;
        gap: .7rem;
        width: 100%;
        max-width: 52rem;
    }
    .Pagination{
        display: flex;
        justify-content: center;
        align-items: center;
        gap: .4rem;
        width: 100%;
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
            font-size: .8rem;
            font-weight: 600;
            font-family: 'Lexend';
            cursor: pointer;
            transition: background 150ms, color 150ms, border-color 150ms, opacity 150ms;
            svg{
                width: 1rem;
                height: 1rem;
            }
            &:hover:not(:disabled){
                background: #ffffff14;
                color: #fff;
            }
            &.active{
                background: #fff;
                color: #000;
                border-color: #fff;
            }
            &:disabled{
                opacity: .3;
                cursor: default;
            }
        }
    }
}
</style>
