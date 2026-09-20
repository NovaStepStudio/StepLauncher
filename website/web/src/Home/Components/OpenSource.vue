<script setup lang="ts">
import { IconBrandGithub, IconScale, IconStar, IconUsers } from '@tabler/icons-vue';
</script>

<template>
    <div class="OpenSource" id="about" v-reveal>
        <div class="CodeCard">
            <div class="CodeHead">
                <div class="Dots">
                    <i></i>
                    <i></i>
                    <i></i>
                </div>
                <small>terminal</small>
            </div>
            <div class="CodeBody sl-stagger">
                <span><em>$</em> git clone https://github.com/NovaStepStudio/StepLauncher.git</span>
                <span><em>$</em> cd StepLauncher/launcher</span>
                <span><em>$</em> go build ./...</span>
            </div>
            <div class="CodeFoot">
                <img src="../../../assets/logo-step-white.png" alt="StepLauncher" loading="lazy" decoding="async">
                <div class="CodeFootTxt">
                    <b>NovaStepStudio / StepLauncher</b>
                    <small>GPL-3.0 • Go + Vue + Wails v3</small>
                </div>
                <span class="BadgeMain">Open Source</span>
            </div>
        </div>
        <div class="TextSide">
            <div class="Tag">
                <IconBrandGithub stroke="2" />
                <span>Open Source</span>
            </div>
            <div class="Description">
                <h2>Código abierto, sin letra chica.</h2>
                <p>StepLauncher es gratis y de código abierto bajo licencia GPL-3.0. El repo es un monorepo: la app vive en <code>launcher/</code>, esta web en <code>website/</code> y la API en <code>api/</code>. Podés ver el código, reportar issues y seguir cada release en GitHub.</p>
            </div>
            <div class="Points sl-stagger">
                <div class="Point">
                    <IconScale stroke="2" />
                    <div class="PointTxt">
                        <b>Licencia GPL-3.0</b>
                        <small>Libre para usar, estudiar y mejorar</small>
                    </div>
                </div>
                <div class="Point">
                    <IconStar stroke="2" />
                    <div class="PointTxt">
                        <b>Desarrollo público</b>
                        <small>Releases, changelog y roadmap a la vista</small>
                    </div>
                </div>
                <div class="Point">
                    <IconUsers stroke="2" />
                    <div class="PointTxt">
                        <b>Hecho por la comunidad</b>
                        <small>NovaStepStudio + contribuidores</small>
                    </div>
                </div>
            </div>
            <div class="Buttons">
                <a class="Btn" href="https://github.com/NovaStepStudio/StepLauncher" target="_blank" rel="noopener">
                    <IconBrandGithub stroke="2" />
                    Ver código
                </a>
                <RouterLink class="Btn" to="/about">
                    Cómo funciona
                </RouterLink>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.OpenSource{
    width: calc(100% - 10rem);
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 4rem 5rem;
    margin: auto;
    gap: 2rem;
    background: #ffffff05;
    border-top: 1px solid #ffffff0d;
    border-bottom: 1px solid #ffffff0d;
    div{
        display:flex;
        justify-content:center;
        align-items:center;
    }
    .CodeCard{
        width: 45%;
        flex-direction:column;
        align-items:stretch;
        border-radius: .7rem;
        border: 1px solid #ffffff14;
        background: #0b0b0d;
        box-shadow: 0 0 15px #000;
        overflow: hidden;
        .CodeHead{
            justify-content:space-between;
            padding: .6rem .9rem;
            border-bottom: 1px solid #ffffff10;
            .Dots{
                gap: .35rem;
                i{
                    width: .6rem;
                    height: .6rem;
                    border-radius: 99rem;
                    background: #ffffff25;
                    &:first-child{
                        background: #ff5f57;
                    }
                    &:nth-child(2){
                        background: #febc2e;
                    }
                    &:nth-child(3){
                        background: #28c840;
                    }
                }
            }
            small{
                font-size: .68rem;
                opacity: .4;
            }
        }
        .CodeBody{
            flex-direction:column;
            align-items:flex-start;
            gap: .45rem;
            padding: 1.1rem 1rem;
            font-family: monospace;
            font-size: .78rem;
            span{
                opacity: .75;
                em{
                    font-style: normal;
                    opacity: .5;
                    margin-right: .5rem;
                }
            }
        }
        .CodeFoot{
            justify-content:flex-start;
            gap: .6rem;
            padding: .8rem .9rem;
            border-top: 1px solid #ffffff10;
            background: #ffffff05;
            img{
                width: 2rem;
                height: 2rem;
                object-fit: contain;
            }
            .CodeFootTxt{
                flex-direction:column;
                align-items:flex-start;
                gap: .1rem;
                b{
                    font-size: .75rem;
                    font-family: 'Lexend';
                }
                small{
                    font-size: .65rem;
                    opacity: .5;
                }
            }
            .BadgeMain{
                margin-left: auto;
                font-size: .65rem;
                font-weight: 700;
                padding: .25rem .6rem;
                border-radius: 99rem;
                background: #fff;
                color: #000;
                border: 1px solid #fff;
            }
        }
    }
    .TextSide{
        width: 45%;
        flex-direction:column;
        align-items:flex-start;
        text-align:left;
        gap: 1rem;
        .Tag{
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
            flex-direction:column;
            align-items:flex-start;
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
                max-width: 27rem;
                code{
                    font-family: monospace;
                    font-size: .8rem;
                    padding: .1rem .35rem;
                    border-radius: .3rem;
                    border: 1px solid #ffffff20;
                    background: #ffffff10;
                    opacity: 1;
                }
            }
        }
        .Points{
            flex-direction:column;
            align-items:stretch;
            gap: .6rem;
            width: 100%;
            max-width: 27rem;
            .Point{
                justify-content:flex-start;
                gap: .7rem;
                padding: .75rem .85rem;
                border-radius: .7rem;
                border: 1px solid #ffffff14;
                background: #ffffff08;
                svg{
                    width: 1.2rem;
                    height: 1.2rem;
                    flex-shrink: 0;
                    opacity: .8;
                }
                .PointTxt{
                    flex-direction:column;
                    align-items:flex-start;
                    gap: .1rem;
                    b{
                        font-size: .82rem;
                        font-family: 'Lexend';
                    }
                    small{
                        font-size: .7rem;
                        opacity: .55;
                    }
                }
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
    }
}
@media (max-width: 1000px){
    .OpenSource{
        flex-direction:column;
        width: calc(100% - 3rem);
        padding: 3rem 1.5rem;
        .CodeCard{
            width: 100%;
            max-width: 34rem;
        }
        .TextSide{
            width: 100%;
            align-items:center;
            text-align:center;
            .Description{
                align-items:center;
            }
            .Point{
                text-align:left;
            }
        }
    }
}
</style>
