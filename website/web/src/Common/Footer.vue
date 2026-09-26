<script setup lang="ts">
import { onMounted } from 'vue';
import { IconBrandGithub, IconDownload } from '@tabler/icons-vue';
import { useVersion } from '@/Common/Composables/useVersion';
import { useAuth } from '@/Auth/Composables/useAuth';

const { version, load } = useVersion();
const { autenticado } = useAuth();

onMounted(() => load());
</script>

<template>
    <div class="Footer">
        <div class="Content">
            <div class="Brand">
                <div class="BrandName">
                    <img src="../../assets/logo-step-white.png" alt="StepLauncher" loading="lazy" decoding="async">
                    <b>StepLauncher</b>
                </div>
                <p>Launcher moderno, rápido y open source para Minecraft: Java Edition. Hecho por NovaStepStudio.</p>
                <span v-if="version" class="BadgeMain">{{ version }}</span>
            </div>
            <div class="Columns">
                <div class="Column">
                    <b>Navegación</b>
                    <RouterLink to="/">Inicio</RouterLink>
                    <RouterLink to="/download">Descarga</RouterLink>
                    <RouterLink to="/changelog">Historial</RouterLink>
                    <RouterLink to="/branding">Branding</RouterLink>
                    <RouterLink to="/faq">FAQ</RouterLink>
                    <RouterLink to="/about">Acerca De</RouterLink>
                </div>
                <div class="Column">
                    <b>Proyecto</b>
                    <a href="https://github.com/NovaStepStudio/StepLauncher" target="_blank" rel="noopener">
                        <IconBrandGithub stroke="2" />
                        Github
                    </a>
                    <a href="https://github.com/NovaStepStudio/StepLauncher/releases" target="_blank" rel="noopener">
                        <IconDownload stroke="2" />
                        Releases
                    </a>
                </div>
                <div class="Column">
                    <b>Cuenta</b>
                    <RouterLink v-if="autenticado" to="/dashboard">Mi panel</RouterLink>
                    <RouterLink v-else to="/auth">Entrar</RouterLink>
                </div>
                <div class="Column">
                    <b>Legal</b>
                    <RouterLink to="/privacy">Privacidad</RouterLink>
                    <RouterLink to="/terms">Términos</RouterLink>
                    <RouterLink to="/credits">Créditos</RouterLink>
                </div>
            </div>
        </div>
        <div class="Bottom">
            <span>© 2026 NovaStepStudio • GPL-3.0</span>
            <i></i>
            <span>Hecho con Wails + Vue</span>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Footer{
    position: relative;
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 3rem 5rem;
    gap: 2rem;
    margin: auto;
    border-top: 1px solid #ffffff14;
    z-index: 1;
    &::after{
        content:'';
        position:absolute;
        inset:0;
        width: 100%;
        height:100%;
        background: linear-gradient(#000000e6,#000000f2), url('../../assets/background/5.webp');
        background-position: center center;
        background-size: cover;
        z-index: -1;
    }
    div{
        display:flex;
        justify-content:center;
        align-items:center;
    }
    .Content{
        width: 100%;
        justify-content: space-between;
        align-items: flex-start;
        gap: 2rem;
        .Brand{
            flex-direction:column;
            align-items:flex-start;
            gap: .7rem;
            max-width: 22rem;
            .BrandName{
                gap: .5rem;
                img{
                    width: 2.2rem;
                    height: 2.2rem;
                    object-fit: contain;
                }
                b{
                    font-family: 'Lexend';
                    font-size: 1.1rem;
                }
            }
            p{
                margin: 0;
                font-size: .8rem;
                line-height: 1.6;
                opacity: .55;
            }
            .BadgeMain{
                font-size: .65rem;
                font-weight: 700;
                padding: .25rem .6rem;
                border-radius: 99rem;
                background: #fff;
                color: #000;
                border: 1px solid #fff;
            }
        }
        .Columns{
            gap: 3rem;
            align-items: flex-start;
            .Column{
                flex-direction:column;
                align-items:flex-start;
                gap: .5rem;
                b{
                    font-size: .7rem;
                    text-transform: uppercase;
                    letter-spacing: .12em;
                    opacity: .4;
                    font-family: 'Lexend';
                    margin-bottom: .2rem;
                }
                a{
                    display:flex;
                    justify-content:center;
                    align-items:center;
                    gap: .4rem;
                    font-size: .82rem;
                    color: #ffffffa6;
                    text-decoration: none;
                    transition: color 150ms;
                    svg{
                        width: 1rem;
                        height: 1rem;
                    }
                    &:hover{
                        color: #fff;
                    }
                }
            }
        }
    }
    .Bottom{
        gap: .6rem;
        width: 100%;
        padding-top: 1.2rem;
        border-top: 1px solid #ffffff10;
        font-size: .72rem;
        opacity: .45;
        i{
            width: 3px;
            height: 3px;
            border-radius: 99rem;
            background: #fff;
        }
    }
}
@media (max-width: 800px){
    .Footer{
        width: calc(100% - 3rem);
        padding: 2.5rem 1.5rem calc(1.5rem + env(safe-area-inset-bottom, 0px)) 1.5rem;
        .Content{
            flex-direction:column;
            align-items: stretch;
            .Brand{
                align-items:center;
                text-align:center;
                max-width: 100%;
                p{
                    max-width: 26rem;
                }
            }
            .Columns{
                width: 100%;
                display: grid;
                grid-template-columns: repeat(2, minmax(0, 1fr));
                gap: 1.5rem 1rem;
                .Column{
                    a{
                        // Área táctil amplia sin cambiar el visual.
                        padding: .35rem 0;
                    }
                }
            }
        }
        .Bottom{
            flex-wrap: wrap;
            justify-content: center;
            text-align: center;
            row-gap: .3rem;
        }
    }
}
</style>
