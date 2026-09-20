<script setup lang="ts">
import { IconDownload, IconBox, IconPlayerPlay } from '@tabler/icons-vue';
</script>

<template>
    <div class="HowItWorks" v-reveal>
        <div class="Head">
            <div class="Tag">
                <span>Cómo funciona</span>
            </div>
            <div class="Description">
                <h2>De la descarga a jugar en minutos.</h2>
                <p>Sin cuentas raras ni configuraciones eternas: instalás, armás tu instancia y jugás.</p>
            </div>
        </div>
        <div class="Steps sl-stagger">
            <div class="Step">
                <span class="Num">1</span>
                <div class="StepTxt">
                    <b><IconDownload stroke="2" /> Descargá la app</b>
                    <small>Windows, Linux o macOS desde la sección de descarga. Es instalary abrir: la primera vez una bienvenida guiada te deja todo listo.</small>
                </div>
            </div>
            <div class="Step">
                <span class="Num">2</span>
                <div class="StepTxt">
                    <b><IconBox stroke="2" /> Creá tu instancia</b>
                    <small>Elegí la versión de Minecraft, el Java que corresponda y, si querés, un modloader (Fabric, Quilt, Forge o NeoForge). Cada instancia guarda su propia configuración.</small>
                </div>
            </div>
            <div class="Step">
                <span class="Num">3</span>
                <div class="StepTxt">
                    <b><IconPlayerPlay stroke="2" /> Jugá</b>
                    <small>Apretá jugar y listo. Si algo falla, el historial de errores te dice qué pasó en español, sin códigos crípticos.</small>
                </div>
            </div>
        </div>
        <div class="Note">
            <b>Qué no es StepLauncher</b>
            <p>No es una cuenta premium ni vende Minecraft: necesitás tu copia del juego. No está afiliado a Mojang Studios ni a Microsoft. Es un proyecto independiente y gratuito, con el código a la vista bajo licencia GPL-3.0.</p>
        </div>
    </div>
</template>

<style scoped lang="scss">
.HowItWorks{
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
    .Steps{
        align-items:stretch;
        gap: .8rem;
        width: 100%;
        max-width: 60rem;
        .Step{
            flex: 1;
            align-items:flex-start;
            gap: .8rem;
            padding: 1rem;
            border-radius: .7rem;
            border: 1px solid #ffffff14;
            background: #ffffff08;
            .Num{
                display:flex;
                justify-content:center;
                align-items:center;
                width: 2rem;
                height: 2rem;
                border-radius: 99rem;
                background: #fff;
                color: #000;
                font-family: 'Lexend';
                font-size: .85rem;
                font-weight: 700;
                flex-shrink: 0;
            }
            .StepTxt{
                flex-direction:column;
                align-items:flex-start;
                text-align:left;
                gap: .3rem;
                b{
                    display:flex;
                    align-items:center;
                    gap: .4rem;
                    font-size: .82rem;
                    font-family: 'Lexend';
                    svg{
                        width: 1rem;
                        height: 1rem;
                    }
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
    .HowItWorks{
        width: calc(100% - 3rem);
        padding: 3rem 1.5rem;
        .Steps{
            flex-direction:column;
        }
    }
}
</style>
