<script setup lang="ts">
import { IconDownload, IconBrandGithub } from '@tabler/icons-vue';
import { useReleases } from '@/Common/Composables/useReleases';
import { useVersion } from '@/Common/Composables/useVersion';
import { onMounted } from 'vue';

const { main, loading, load } = useReleases();
const { version, load: loadVersion } = useVersion();

// Scroll suave a los archivos sin tocar la URL (nada de #hash).
function scrollToFiles() {
    document.getElementById('stable')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

onMounted(() => {
    load();
    loadVersion();
});
</script>

<template>
    <div class="FirstPrew">
        <div class="TextAndButtons">
            <div class="AppName sl-enter" style="--sl-delay: 0s">
                <img class="sl-float" src="../../../assets/logo-step-white.png" alt="StepLauncher" loading="eager" decoding="async" fetchpriority="high">
                <h1>Descargas</h1>
            </div>
            <div class="Badges sl-enter" style="--sl-delay: .08s">
                <span v-if="loading" class="Badge">Buscando última versión...</span>
                <span v-else-if="main" class="BadgeMain">{{ main.tag }}</span>
                <span v-else-if="version" class="BadgeMain">{{ version }}</span>
                <span v-else class="Badge">Versión en camino</span>
                <span class="Badge">Windows • Linux • macOS</span>
            </div>
            <div class="Description sl-enter" style="--sl-delay: .16s">
                <h2>Elegí tu sistema y a jugar.</h2>
                <p>Descargá el último release estable directo de GitHub. Si te gusta probar lo nuevo, abajo también tenés las betas.</p>
            </div>
            <div class="Buttons sl-enter" style="--sl-delay: .24s">
                <button class="BtnPrimary" type="button" @click="scrollToFiles">
                    <IconDownload stroke="2" />
                    Ver archivos
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
</template>

<style scoped lang="scss">
.FirstPrew{
    position:relative;
    width: 100%;
    min-height: 58dvh;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-top: 4rem;
    z-index: 1;
    &::after{
        content:'';
        position:absolute;
        inset:0;
        width: 100%;
        height:100%;
        mask: linear-gradient(#0008 50%,transparent);
        background: linear-gradient(#000000b3,#000000b3), url('../../../assets/background/4.webp');
        background-position: center center;
        background-size: cover;
        z-index: -1;
    }
    div{
        display:flex;
        justify-content:center;
        align-items:center;
    }
    .TextAndButtons{
        width: 100%;
        flex-direction:column;
        align-items:center;
        text-align:center;
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
            flex-direction:column;
            align-items:center;
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
                display:flex;
                justify-content:center;
                align-items:center;
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
</style>
