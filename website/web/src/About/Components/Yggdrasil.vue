<script setup lang="ts">
import { IconKey, IconServer, IconShirt } from '@tabler/icons-vue';
</script>

<template>
    <div class="Yggdrasil" v-reveal>
        <div class="Head">
            <div class="Tag">
                <span>Yggdrasil público</span>
            </div>
            <div class="Description">
                <h2>Tu cuenta, tu identidad en el juego.</h2>
                <p>StepLauncher incluye un servicio de autenticación compatible con Yggdrasil y abierto a todos: sirve para entrar a servidores que lo acepten y para que cualquier servidor lo adopte.</p>
            </div>
        </div>
        <div class="Cards sl-stagger">
            <div class="Card">
                <span class="CardIcon"><IconKey stroke="2" /></span>
                <div class="CardTxt">
                    <b>Cuenta propia</b>
                    <small>Modo offline para jugar sin verificar tu identidad, o modo online para entrar con tu cuenta a servidores compatibles: tu nombre, tu UUID y tus texturas viajan con vos.</small>
                </div>
            </div>
            <div class="Card">
                <span class="CardIcon"><IconServer stroke="2" /></span>
                <div class="CardTxt">
                    <b>Servidores externos</b>
                    <small>El launcher permite agregar direcciones Yggdrasil de terceros. Son servicios ajenos: revisá sus reglas antes de usarlos.</small>
                </div>
            </div>
            <div class="Card">
                <span class="CardIcon"><IconShirt stroke="2" /></span>
                <div class="CardTxt">
                    <b>Skin y capa en el juego</b>
                    <small>Lo que subas a tu perfil viaja con vos al servidor: tu skin y tu capa se ven donde juegues.</small>
                </div>
            </div>
        </div>
        <div class="Note">
            <b>Descarga oficial, sin atajos</b>
            <p>El juego y los modloaders (Forge, Fabric, Quilt, NeoForge) se descargan desde fuentes oficiales y se ejecutan en tu PC, siguiendo las políticas de Mojang y Microsoft. StepLauncher no envía tu cuenta a Mojang o Microsoft: en modo online tu identidad solo llega al servicio Yggdrasil que elegiste y al servidor donde jugás. Yggdrasil no otorga licencias ni reemplaza la cuenta oficial: los servidores premium siguen pidiendo su autenticación. Proyecto independiente, sin afiliación con Mojang Studios ni Microsoft.</p>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Yggdrasil{
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
    .Note{
        flex-direction:column;
        align-items:flex-start;
        gap: .4rem;
        width: 100%;
        max-width: 60rem;
        padding: .9rem 1rem;
        border-radius: .7rem;
        border: 1px solid #ffffff14;
        background: #00000080;
        b{
            font-size: .78rem;
            font-family: 'Lexend';
        }
        p{
            margin: 0;
            font-size: .75rem;
            line-height: 1.6;
            opacity: .6;
        }
    }
}
@media (max-width: 1000px){
    .Yggdrasil{
        width: calc(100% - 3rem);
        padding: 3rem 1.5rem;
        .Cards{
            flex-direction:column;
        }
    }
}
</style>
