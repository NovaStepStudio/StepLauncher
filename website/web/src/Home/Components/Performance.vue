<script setup lang="ts">
import { IconBolt, IconFeather, IconDevices, IconRocket } from '@tabler/icons-vue';
import { capturaAlAzar } from '@/Common/Composables/capturas';

const captura = capturaAlAzar('Playing');
</script>

<template>
    <div class="Performance" v-reveal>
        <div class="Head">
            <div class="Tag">
                <IconRocket stroke="2" />
                <span>Rendimiento</span>
            </div>
            <div class="Description">
                <h2>Rápido, liviano y multiplataforma.</h2>
                <p>App nativa hecha con Wails + Go: arranca rápido, descarga en serio y corre en tu sistema, sea cual sea.</p>
            </div>
        </div>
        <div class="Cards sl-stagger">
            <div class="Card">
                <span class="CardIcon"><IconBolt stroke="2" /></span>
                <div class="CardTxt">
                    <b>Rápido de verdad</b>
                    <small>Descargas concurrentes de 1 a 16 hilos, con pausa, reanudación y verificación SHA1.</small>
                </div>
            </div>
            <div class="Card">
                <span class="CardIcon"><IconFeather stroke="2" /></span>
                <div class="CardTxt">
                    <b>Liviano</b>
                    <small>Binario nativo sin Electron pesado. Interfaz Vue fluida que no come RAM.</small>
                </div>
            </div>
            <div class="Card">
                <span class="CardIcon"><IconDevices stroke="2" /></span>
                <div class="CardTxt">
                    <b>Multiplataforma</b>
                    <small>Windows, Linux y macOS con la misma experiencia y tus datos donde elijas.</small>
                </div>
            </div>
        </div>
        <div class="Image">
            <img :src="captura" alt="Menú de juego de StepLauncher" loading="lazy" decoding="async">
            <div class="ImageTag">
                <b>Listo para jugar</b>
                <small>Detecta tu Java y a jugar</small>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Performance{
    width: calc(100% - 10rem);
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 4rem 0;
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
    .Image{
        position: relative;
        margin-top: 1rem;
        img{
            border-radius: .5rem;
            object-fit: cover;
            width: 120dvh;
            max-width: 100%;
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
    .Performance{
        width: calc(100% - 3rem);
        padding: 3rem 1.5rem;
        .Cards{
            flex-direction:column;
        }
        .Image{
            img{
                width: 100%;
                max-width: 34rem;
            }
        }
    }
}
</style>
