<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { IconBrandWindows, IconBrandApple, IconTerminal, IconFile, IconCalendar, IconTag, IconDownload } from '@tabler/icons-vue';
import { assetsByPlatform, formatSize, formatDate, archOf, kindOf, type GithubRelease, type PlatformKey } from '@/Common/Composables/useReleases';

const props = defineProps<{
    release: GithubRelease;
    badge: string;
}>();

const groups = computed(() => assetsByPlatform(props.release));

const platforms = [
    { key: 'windows', label: 'Windows', icon: IconBrandWindows },
    { key: 'macos', label: 'macOS', icon: IconBrandApple },
    { key: 'linux', label: 'Linux', icon: IconTerminal },
    { key: 'other', label: 'Otros', icon: IconFile },
] as const;

const active = ref<PlatformKey>('windows');

function detectarSistema(): PlatformKey {
    const ua = navigator.userAgent.toLowerCase();
    if (ua.includes('win')) return 'windows';
    if (ua.includes('mac')) return 'macos';
    if (ua.includes('linux')) return 'linux';
    return 'windows';
}

function elegirTabInicial() {
    const g = groups.value;
    const sistema = detectarSistema();
    if (g[sistema].length > 0) {
        active.value = sistema;
        return;
    }
    const primero = (Object.keys(g) as PlatformKey[]).find((k) => g[k].length > 0);
    if (primero) active.value = primero;
}

watch(groups, () => elegirTabInicial(), { immediate: true });

const totalArchivos = computed(() =>
    (Object.keys(groups.value) as PlatformKey[]).reduce((acc, k) => acc + groups.value[k].length, 0),
);

// Índice de la tab activa: mueve el span deslizante (cada columna mide lo
// mismo, así que translateX(índice * 100%) cae justo en su columna).
const activeIndex = computed(() => Math.max(0, platforms.findIndex((p) => p.key === active.value)));

// Delay escalonado por archivo (tope 560ms para que no se haga eterno).
function delayDe(i: number): string {
    return `${Math.min(i * 70, 560)}ms`;
}
</script>

<template>
    <div class="ReleaseCard">
        <div class="ReleaseHead">
            <div class="ReleaseTitle">
                <span class="BadgeMain">{{ badge }}</span>
                <b>{{ release.tag }}</b>
            </div>
            <div class="ReleaseMeta">
                <span><IconTag stroke="2" /> {{ release.name }}</span>
                <span><IconCalendar stroke="2" /> {{ formatDate(release.publishedAt) }}</span>
                <span class="Count">{{ totalArchivos }} archivos</span>
            </div>
        </div>
        <div class="Tabs" role="tablist" :style="'--tab-index: ' + activeIndex">
            <span class="TabSlider" aria-hidden="true"></span>
            <button
                v-for="p in platforms"
                :key="p.key"
                role="tab"
                class="Tab"
                :class="{ active: active === p.key, empty: groups[p.key].length === 0 }"
                :aria-selected="active === p.key"
                @click="active = p.key"
            >
                <component :is="p.icon" stroke="2" />
                <span>{{ p.label }}</span>
                <i>{{ groups[p.key].length }}</i>
            </button>
        </div>
        <div class="TabBody sl-fade" :key="active" role="tabpanel">
            <div v-if="groups[active].length > 0" class="Assets">
                <a
                    v-for="(a, i) in groups[active]"
                    :key="a.url"
                    class="Asset sl-item"
                    :style="'--sl-delay: ' + delayDe(i)"
                    :href="a.url"
                    target="_blank"
                    rel="noopener"
                >
                    <span class="AssetTxt">
                        <b>{{ a.name }}</b>
                        <small>
                            <i v-if="archOf(a.name)" class="MiniBadge">{{ archOf(a.name) }}</i>
                            <i v-if="kindOf(a.name)" class="MiniBadge">{{ kindOf(a.name) }}</i>
                            {{ formatSize(a.size) }} • {{ a.downloads }} descargas
                        </small>
                    </span>
                    <IconDownload stroke="2" />
                </a>
            </div>
            <small v-else class="Empty">Sin archivos para este sistema en este release.</small>
        </div>
        <a class="AllReleases" :href="release.url" target="_blank" rel="noopener">Ver este release en GitHub</a>
    </div>
</template>

<style scoped lang="scss">
.ReleaseCard{
    width: 100%;
    max-width: 60rem;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 1.2rem;
    .ReleaseHead{
        display:flex;
        flex-direction:column;
        align-items:center;
        text-align:center;
        gap: .6rem;
        .ReleaseTitle{
            display:flex;
            align-items:center;
            gap: .6rem;
            .BadgeMain{
                font-size: .65rem;
                font-weight: 700;
                text-transform: uppercase;
                letter-spacing: .08em;
                padding: .25rem .6rem;
                border-radius: 99rem;
                background: #fff;
                color: #000;
                border: 1px solid #fff;
            }
            b{
                font-family: 'Lexend';
                font-size: 1.3rem;
            }
        }
        .ReleaseMeta{
            display:flex;
            align-items:center;
            justify-content:center;
            flex-wrap: wrap;
            gap: .5rem 1rem;
            span{
                display:flex;
                align-items:center;
                gap: .35rem;
                font-size: .75rem;
                opacity: .55;
                svg{
                    width: .95rem;
                    height: .95rem;
                }
            }
            .Count{
                padding: .15rem .55rem;
                border-radius: 99rem;
                border: 1px solid #ffffff25;
                background: #ffffff0d;
                opacity: .85;
            }
        }
    }
    .Tabs{
        position: relative;
        display: grid;
        grid-template-columns: repeat(4, minmax(0, 1fr));
        padding: .35rem;
        border-radius: .7rem;
        border: 1px solid #ffffff14;
        background: #ffffff05;
        overflow-x: auto;
        .TabSlider{
            position: absolute;
            top: .35rem;
            bottom: .35rem;
            left: .35rem;
            width: calc((100% - .7rem) / 4);
            border-radius: .5rem;
            background: #fff;
            box-shadow: 0 2px 12px #00000080;
            transform: translateX(calc(var(--tab-index, 0) * 100%));
            transition: transform .38s cubic-bezier(.22, .61, .36, 1);
            pointer-events: none;
        }
        .Tab{
            position: relative;
            z-index: 1;
            display:flex;
            justify-content:center;
            align-items:center;
            gap: .45rem;
            padding: .55rem .8rem;
            border-radius: .5rem;
            border: 1px solid transparent;
            background: transparent;
            color: #ffffffa6;
            font-size: .8rem;
            font-weight: 600;
            font-family: 'Lexend';
            cursor: pointer;
            white-space: nowrap;
            transition: color 150ms;
            svg{
                width: 1.05rem;
                height: 1.05rem;
                flex-shrink: 0;
            }
            i{
                font-style: normal;
                font-size: .65rem;
                font-weight: 700;
                min-width: 1.3rem;
                height: 1.3rem;
                display:flex;
                justify-content:center;
                align-items:center;
                padding: 0 .3rem;
                border-radius: 99rem;
                border: 1px solid #ffffff25;
                background: #ffffff10;
                transition: background 150ms, border-color 150ms;
            }
            &:hover:not(.active){
                color: #fff;
            }
            &.active{
                background: transparent;
                color: #000;
                border-color: transparent;
                i{
                    background: #00000015;
                    border-color: #00000025;
                }
            }
            &.empty:not(.active){
                opacity: .45;
            }
        }
    }
    .TabBody{
        display:flex;
        flex-direction:column;
        .Assets{
            display:grid;
            grid-template-columns: repeat(2, minmax(0, 1fr));
            gap: .45rem;
            .Asset:first-child:nth-last-child(1){
                grid-column: 1 / -1;
            }
            .Asset{
                display:flex;
                justify-content:space-between;
                align-items:center;
                gap: .8rem;
                padding: .6rem .8rem;
                border-radius: .5rem;
                border: 1px solid #ffffff14;
                background: #ffffff08;
                color: #fff;
                text-decoration: none;
                transition: background 150ms, transform 150ms;
                svg{
                    width: 1.1rem;
                    height: 1.1rem;
                    flex-shrink: 0;
                    opacity: .7;
                }
                .AssetTxt{
                    display:flex;
                    flex-direction:column;
                    align-items:flex-start;
                    gap: .1rem;
                    min-width: 0;
                    b{
                        font-size: .78rem;
                        white-space: nowrap;
                        overflow: hidden;
                        text-overflow: ellipsis;
                        max-width: 100%;
                    }
                    small{
                        font-size: .68rem;
                        opacity: .5;
                        display:flex;
                        align-items:center;
                        gap: .3rem;
                        .MiniBadge{
                            font-style: normal;
                            font-size: .6rem;
                            font-weight: 700;
                            text-transform: uppercase;
                            letter-spacing: .06em;
                            padding: .1rem .4rem;
                            border-radius: 99rem;
                            border: 1px solid #ffffff25;
                            background: #ffffff10;
                            opacity: 1;
                        }
                    }
                }
                &:hover{
                    background: #ffffff14;
                    transform: translateY(-1px);
                }
            }
        }
        .Empty{
            font-size: .72rem;
            opacity: .4;
            text-align: center;
            padding: 1rem 0;
        }
    }
    .AllReleases{
        align-self: center;
        font-size: .78rem;
        color: #ffffffa6;
        text-decoration: none;
        &:hover{
            color: #fff;
        }
    }
}
@media (max-width: 700px){
    .ReleaseCard{
        .TabBody{
            .Assets{
                grid-template-columns: 1fr;
            }
        }
    }
}
@media (max-width: 600px){
    .ReleaseCard{
        .Tabs{
            display: flex;
            .TabSlider{
                display: none;
            }
            .Tab{
                flex: 1 0 auto;
                &.active{
                    background: #fff;
                }
            }
        }
    }
}
@media (prefers-reduced-motion: reduce){
    .ReleaseCard{
        .Tabs{
            .TabSlider{
                transition: none;
            }
        }
    }
}
</style>
