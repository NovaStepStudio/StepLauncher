<script setup lang="ts">
import { IconMusic, IconLibrary, IconPlaylist, IconPlayerPlay } from '@tabler/icons-vue';
import { capturaAlAzar } from '@/Common/Composables/capturas';

const captura = capturaAlAzar('MusicPanel');
</script>

<template>
    <div class="Music" v-reveal>
        <div class="TextSide">
            <div class="Tag">
                <IconMusic stroke="2" />
                <span>Música</span>
            </div>
            <div class="Description">
                <h2>Tu música, mientras jugás.</h2>
                <p>Biblioteca musical integrada al launcher: tu colección local, tus listas y tu cola, sin abrir otra app.</p>
            </div>
            <div class="Points sl-stagger">
                <div class="Point">
                    <IconLibrary stroke="2" />
                    <div class="PointTxt">
                        <b>Biblioteca local</b>
                        <small>Toda tu música en un solo lugar, con resumen y estadísticas</small>
                    </div>
                </div>
                <div class="Point">
                    <IconPlaylist stroke="2" />
                    <div class="PointTxt">
                        <b>Listas de reproducción</b>
                        <small>Creá y gestioná tus playlists como quieras</small>
                    </div>
                </div>
                <div class="Point">
                    <IconPlayerPlay stroke="2" />
                    <div class="PointTxt">
                        <b>Ahora suena + cola</b>
                        <small>Reproductor con cola de reproducción a la vista</small>
                    </div>
                </div>
            </div>
        </div>
        <div class="Image">
            <img :src="captura" alt="Biblioteca de música de StepLauncher" loading="lazy" decoding="async">
            <div class="ImageTag">
                <b>Tu música, a tu ritmo</b>
                <small>Biblioteca y reproductor integrados</small>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Music{
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: calc(100% - 10rem);
    padding: 4rem 5rem;
    margin: auto;
    gap: 2rem;
    div{
        display:flex;
        justify-content:center;
        align-items:center;
    }
    .TextSide{
        width: 50%;
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
    }
    .Image{
        position: relative;
        margin: 5rem;
        img{
            transform: perspective(120px) rotateY(-2deg);
            border-radius: .5rem;
            object-fit: cover;
            width: 80dvh;
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
    .Music{
        flex-direction:column;
        width: calc(100% - 3rem);
        padding: 3rem 1.5rem;
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
