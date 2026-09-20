<script setup lang="ts">
import { IconBrandGolang, IconCpu, IconBrandVue, IconPackage } from '@tabler/icons-vue';
</script>

<template>
    <div class="Stack" v-reveal>
        <div class="Head">
            <div class="Tag">
                <span>Tecnología</span>
            </div>
            <div class="Description">
                <h2>Nativo por dentro, moderno por fuera.</h2>
                <p>Backend en Go con Wails v3 y frontend en Vue 3 + TypeScript: potencia sin el peso de Electron.</p>
            </div>
        </div>
        <div class="Cards sl-stagger">
            <div class="Card">
                <span class="CardIcon"><IconBrandGolang stroke="2" /></span>
                <div class="CardTxt">
                    <b>Go 1.26</b>
                    <small>Descargas concurrentes, lanzamiento y verificación en el backend.</small>
                </div>
            </div>
            <div class="Card">
                <span class="CardIcon"><IconCpu stroke="2" /></span>
                <div class="CardTxt">
                    <b>Wails v3</b>
                    <small>App de escritorio nativa con WebView del sistema.</small>
                </div>
            </div>
            <div class="Card">
                <span class="CardIcon"><IconBrandVue stroke="2" /></span>
                <div class="CardTxt">
                    <b>Vue 3 + TS</b>
                    <small>Interfaz por dominios, rápida y personalizable.</small>
                </div>
            </div>
            <div class="Card">
                <span class="CardIcon"><IconPackage stroke="2" /></span>
                <div class="CardTxt">
                    <b>Vite + Bun</b>
                    <small>Builds veloces para Windows, Linux y macOS.</small>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Stack{
    width: calc(100% - 10rem);
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 4rem 5rem;
    margin: auto;
    gap: 1.5rem;
    background: #ffffff05;
    border-top: 1px solid #ffffff0d;
    border-bottom: 1px solid #ffffff0d;
    div{
        display:flex;
        justify-content:center;
        align-items:center;
    }
    .Head{
        flex-direction:column;
        gap: 1rem;
        text-align:center;
        .Tag{
            font-size: .7rem;
            text-transform: uppercase;
            letter-spacing: .12em;
            padding: .25rem .65rem;
            border-radius: 99rem;
            border: 1px solid #ffffff25;
            background: #ffffff0d;
            opacity: .85;
        }
        .Description{
            flex-direction:column;
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
                max-width: 32rem;
            }
        }
    }
    .Cards{
        gap: .8rem;
        align-items:stretch;
        width: 100%;
        max-width: 60rem;
        .Card{
            flex: 1;
            align-items:flex-start;
            gap: .7rem;
            padding: .9rem 1rem;
            border-radius: .7rem;
            border: 1px solid #ffffff14;
            background: #ffffff08;
            .CardIcon{
                display:flex;
                justify-content:center;
                align-items:center;
                width: 2.4rem;
                height: 2.4rem;
                border-radius: .55rem;
                background: #ffffff10;
                border: 1px solid #ffffff14;
                flex-shrink: 0;
                svg{
                    width: 1.2rem;
                    height: 1.2rem;
                }
            }
            .CardTxt{
                flex-direction:column;
                align-items:flex-start;
                text-align:left;
                gap: .3rem;
                b{
                    font-size: .82rem;
                    font-family: 'Lexend';
                }
                small{
                    font-size: .72rem;
                    opacity: .55;
                    line-height: 1.5;
                }
            }
        }
    }
}
@media (max-width: 1000px){
    .Stack{
        width: calc(100% - 3rem);
        padding: 3rem 1.5rem;
        .Cards{
            flex-direction:column;
        }
    }
}
</style>
