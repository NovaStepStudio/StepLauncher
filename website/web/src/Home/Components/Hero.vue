<script setup lang="ts">
import { onMounted } from 'vue';
import { IconBrandGithub, IconDownload } from '@tabler/icons-vue';
import { useVersion } from '@/Common/Composables/useVersion';

const { version, load } = useVersion();

onMounted(() => load());
</script>

<template>
    <div class="FirstPrew" id="home">
        <div class="TextAndButtons">
            <div class="AppName sl-enter" style="--sl-delay: 0s">
                <img class="sl-float" src="../../../assets/logo-step-white.png" alt="StepLauncher" loading="eager" decoding="async" fetchpriority="high">
                <h1>StepLauncher</h1>
            </div>
            <div class="Badges sl-enter" style="--sl-delay: .08s">
                <span v-if="version" class="BadgeMain">{{ version }}</span>
                <span v-else class="Badge">Buscando versión...</span>
                <span class="Badge">Windows • Linux • macOS</span>
            </div>
            <div class="Description sl-enter" style="--sl-delay: .16s">
                <h2>Tu Minecraft, sin vueltas.</h2>
                <p>Launcher moderno, rápido y multiplataforma para Minecraft: Java Edition. Gestioná versiones, modloaders, instancias y cuentas desde una interfaz limpia.</p>
            </div>
            <div class="Buttons sl-enter" style="--sl-delay: .24s">
                <RouterLink class="BtnPrimary" to="/download">
                    <IconDownload stroke="2" />
                    Descargar
                </RouterLink>
                <a class="Btn" href="https://github.com/NovaStepStudio/StepLauncher" target="_blank" rel="noopener">
                    <IconBrandGithub stroke="2" />
                    GitHub
                </a>
            </div>
            <div class="MiniInfo sl-enter" style="--sl-delay: .32s">
                <span>Gratis y open source</span>
                <i></i>
                <span>Hecho con Wails + Vue</span>
            </div>
        </div>
        <div class="Image sl-fade" style="--sl-delay: .2s">
            <img src="../../../assets/capturesPreview/MainMenu.png" alt="Menú principal de StepLauncher" loading="eager" decoding="async" fetchpriority="high">
            <div class="ImageTag">
                <b>Menú principal</b>
                <small>Interfaz real del launcher</small>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.FirstPrew{
    position:relative;
    width: 100%;
    min-height: 100dvh;
    display: flex;
    justify-content: space-between;
    align-items: center;
    z-index: 1;
    &::after{
        content:'';
        position:absolute;
        inset:0;
        width: 100%;
        height:100%;
        mask: linear-gradient(#0008 50%,transparent);
        background: url('../../../assets/background/2.webp');
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
        width: 50%;
        flex-direction:column;
        align-items:flex-start;
        text-align:left;
        padding-left: 6rem;
        gap: 1rem;
        .AppName{
            gap: .5rem;
            font-family: 'Lexend';
            img{
                width: 5rem;
            }
            h1{
                margin: 0;
                font-size: 2.6rem;
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
            align-items:flex-start;
            gap: .5rem;
            max-width: 28rem;
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
            a{
                display:flex;
                justify-content:center;
                align-items:center;
                gap: .45rem;
                padding: .6rem 1.3rem;
                border-radius: .5rem;
                font-size: .85rem;
                font-weight: 600;
                text-decoration: none;
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
    .Image {
        position: relative;
        margin: 5rem;
        img{
            transform: perspective(120px) rotateY(-2deg);
            border-radius: .5rem;
            object-fit: cover;
            width: 85dvh;
            box-shadow: 0 0 15px #000;
            border: 1px solid #ffffff18;
        }
        .ImageTag{
            position: absolute;
            left: 1rem;
            bottom: 1rem;
            flex-direction: column;
            align-items: flex-start;
            gap: .1rem;
            padding: .55rem .8rem;
            border-radius: .5rem;
            border: 1px solid #ffffff20;
            background: #000000b3;
            transform: perspective(120px) rotateY(-2deg);
            b{
                font-size: .75rem;
                font-family: 'Lexend';
            }
            small{
                font-size: .65rem;
                opacity: .55;
            }
        }
    }
}
@media (max-width: 1000px){
    .FirstPrew{
        flex-direction:column;
        justify-content:center;
        gap: 1rem;
        padding: 5rem 1.5rem 2rem 1.5rem;
        .TextAndButtons{
            width: 100%;
            padding-left: 0;
            align-items:center;
            text-align:center;
            .Description{
                align-items:center;
            }
        }
        .Image{
            margin: 0;
            img{
                width: 100%;
                max-width: 34rem;
            }
        }
    }
}
</style>
