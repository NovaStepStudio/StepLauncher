<script setup lang="ts">
import { IconBox, IconPuzzle, IconMusic, IconPalette } from '@tabler/icons-vue';
</script>

<template>
    <div class="Project" v-reveal>
        <div class="Head">
            <div class="Tag">
                <span>El proyecto</span>
            </div>
            <div class="Description">
                <h2>Todo Minecraft en una sola app.</h2>
                <p>Sin launchers pesados ni mil programas: versiones, instancias, mods, cuentas y música conviven en StepLauncher.</p>
            </div>
        </div>
        <div class="Cards sl-stagger">
            <div class="Card">
                <span class="CardIcon"><IconBox stroke="2" /></span>
                <div class="CardTxt">
                    <b>Instancias y versiones</b>
                    <small>Cada perfil con su versión, su Java y su configuración. Cloná, verificá y organizá.</small>
                </div>
            </div>
            <div class="Card">
                <span class="CardIcon"><IconPuzzle stroke="2" /></span>
                <div class="CardTxt">
                    <b>Mods integrados</b>
                    <small>Modrinth dentro del launcher: mods, modpacks, shaders y texturas con todos los loaders.</small>
                </div>
            </div>
            <div class="Card">
                <span class="CardIcon"><IconMusic stroke="2" /></span>
                <div class="CardTxt">
                    <b>Música incluida</b>
                    <small>Biblioteca local, playlists y cola de reproducción sin salir del launcher.</small>
                </div>
            </div>
            <div class="Card">
                <span class="CardIcon"><IconPalette stroke="2" /></span>
                <div class="CardTxt">
                    <b>Personalización total</b>
                    <small>Fondos, colores, fuentes y efectos con vista previa en vivo.</small>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Project{
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
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: .8rem;
        width: 100%;
        max-width: 60rem;
        .Card{
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
    .Project{
        width: calc(100% - 3rem);
        padding: 3rem 1.5rem;
        .Cards{
            grid-template-columns: 1fr;
        }
    }
}
</style>
