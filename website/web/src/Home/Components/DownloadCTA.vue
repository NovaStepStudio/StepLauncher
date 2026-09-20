<script setup lang="ts">
import { onMounted } from 'vue';
import { IconDownload, IconBrandGithub, IconBrandWindows, IconBrandApple, IconTerminal, IconArrowRight } from '@tabler/icons-vue';
import { useVersion } from '@/Common/Composables/useVersion';

const { version, load } = useVersion();

onMounted(() => load());
</script>

<template>
    <div class="DownloadCTA" id="download" v-reveal>
        <div class="CTAInner sl-stagger">
            <div class="Text">
                <span class="Tag">
                    <IconDownload stroke="2" />
                    Descarga
                </span>
                <h2>¿Qué esperás? Descargá StepLauncher.</h2>
                <p>Gratis, open source y listo para jugar en Windows, Linux y macOS. Tu próxima instancia está a un clic.</p>
                <div class="Buttons">
                    <RouterLink class="BtnPrimary" to="/download">
                        <IconDownload stroke="2" />
                        {{ version ? `Descargar ${version}` : 'Descargar' }}
                    </RouterLink>
                    <a class="Btn" href="https://github.com/NovaStepStudio/StepLauncher" target="_blank" rel="noopener">
                        <IconBrandGithub stroke="2" />
                        GitHub
                    </a>
                </div>
                <div class="MiniInfo">
                    <span>Hecho por NovaStepStudio</span>
                    <i></i>
                    <span>GPL-3.0</span>
                </div>
            </div>
            <div class="Version">
                <small>Última versión</small>
                <b>{{ version ? version : '···' }}</b>
                <div class="Platforms">
                    <span>
                        <IconBrandWindows stroke="2" />
                        Windows
                    </span>
                    <span>
                        <IconBrandApple stroke="2" />
                        macOS
                    </span>
                    <span>
                        <IconTerminal stroke="2" />
                        Linux
                    </span>
                </div>
                <RouterLink class="HistoryLink" to="/changelog">
                    Ver historial
                    <IconArrowRight stroke="2" />
                </RouterLink>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.DownloadCTA{
    position: relative;
    width: 100%;
    padding: 8rem 0;
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1;
    &::after{
        content:'';
        position:absolute;
        inset:0;
        width: 100%;
        height:100%;
        mask: linear-gradient(transparent,#0008 85%);
        background: url('../../../assets/background/1.webp');
        background-position: center center;
        background-size: cover;
        z-index: -1;
    }
    .CTAInner{
        display: grid;
        grid-template-columns: 1.2fr .8fr;
        align-items: center;
        gap: 3rem;
        width: calc(100% - 10rem);
        max-width: 60rem;
        margin: auto;
        .Text{
            display: flex;
            flex-direction: column;
            align-items: flex-start;
            text-align: left;
            gap: 1rem;
            .Tag{
                display: flex;
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
            h2{
                margin: 0;
                font-family: 'Lexend';
                font-size: 1.7rem;
                font-weight: 600;
                line-height: 1.3;
            }
            p{
                margin: 0;
                font-size: .9rem;
                line-height: 1.6;
                opacity: .65;
                max-width: 26rem;
            }
            .Buttons{
                display: flex;
                gap: .6rem;
                a{
                    display: flex;
                    justify-content: center;
                    align-items: center;
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
                display: flex;
                align-items: center;
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
        .Version{
            display: flex;
            flex-direction: column;
            align-items: flex-start;
            gap: .6rem;
            padding-left: 3rem;
            border-left: 1px solid #ffffff1e;
            small{
                font-size: .7rem;
                text-transform: uppercase;
                letter-spacing: .12em;
                opacity: .45;
                font-family: 'Lexend';
            }
            b{
                font-family: 'Lexend';
                font-size: 2.4rem;
                font-weight: 700;
                line-height: 1;
            }
            .Platforms{
                display: flex;
                gap: 1rem;
                span{
                    display: flex;
                    align-items: center;
                    gap: .35rem;
                    font-size: .75rem;
                    opacity: .6;
                    svg{
                        width: 1rem;
                        height: 1rem;
                    }
                }
            }
            .HistoryLink{
                display: flex;
                align-items: center;
                gap: .35rem;
                margin-top: .3rem;
                font-size: .8rem;
                font-weight: 600;
                color: #ffffffa6;
                text-decoration: none;
                transition: color 150ms, gap 150ms;
                svg{
                    width: 1rem;
                    height: 1rem;
                }
                &:hover{
                    color: #fff;
                    gap: .55rem;
                }
            }
        }
    }
}
@media (max-width: 1000px){
    .DownloadCTA{
        padding: 5rem 1.5rem;
        .CTAInner{
            grid-template-columns: 1fr;
            gap: 2rem;
            width: 100%;
            .Version{
                padding-left: 0;
                padding-top: 1.6rem;
                border-left: 0;
                border-top: 1px solid #ffffff1e;
            }
            .Buttons{
                flex-direction: column;
                width: 100%;
                a{
                    width: 100%;
                }
            }
        }
    }
}
</style>
