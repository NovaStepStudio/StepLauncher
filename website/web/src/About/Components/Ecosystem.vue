<script setup lang="ts">
import { IconAppWindow, IconWorld, IconServer } from '@tabler/icons-vue';
</script>

<template>
    <div class="Ecosystem" v-reveal>
        <div class="Head">
            <div class="Tag">
                <span>El ecosistema</span>
            </div>
            <div class="Description">
                <h2>Un monorepo, tres piezas.</h2>
                <p>Launcher, web y API viven en el mismo repositorio pero se desarrollan y despliegan por separado: lo que cambia en uno no rompe al otro.</p>
            </div>
        </div>
        <div class="Cards sl-stagger">
            <div class="Card">
                <span class="CardIcon"><IconAppWindow stroke="2" /></span>
                <div class="CardTxt">
                    <b>launcher/ — la app</b>
                    <small>El programa de escritorio: backend en Go con Wails v3 y frontend en Vue 3. Descargas, lanzamiento, mods, instancias, cuentas y música.</small>
                </div>
            </div>
            <div class="Card">
                <span class="CardIcon"><IconWorld stroke="2" /></span>
                <div class="CardTxt">
                    <b>website/ — esta web</b>
                    <small>Landing, descargas, historial y tu panel de cuenta. Vue 3 + Vite desplegada como sitio estático en Cloudflare Pages.</small>
                </div>
            </div>
            <div class="Card">
                <span class="CardIcon"><IconServer stroke="2" /></span>
                <div class="CardTxt">
                    <b>api/ — el backend</b>
                    <small>La API de cuentas y comunidad: Cloudflare Workers con validación estricta y datos en Supabase. Versionada para no romper nada.</small>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Ecosystem{
    width: calc(100% - 10rem);
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 4rem 5rem;
    margin: auto;
    gap: 1.5rem;
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
        align-items:stretch;
        gap: .8rem;
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
    .Ecosystem{
        width: calc(100% - 3rem);
        padding: 3rem 1.5rem;
        .Cards{
            flex-direction:column;
        }
    }
}
</style>
