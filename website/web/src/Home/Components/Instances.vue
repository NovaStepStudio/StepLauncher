<script setup lang="ts">
import { IconBox, IconCopy, IconShieldCheck, IconSettings } from '@tabler/icons-vue';
import { capturaAlAzar } from '@/Common/Composables/capturas';

const captura = capturaAlAzar('Instances');
</script>

<template>
    <div class="Instances" v-reveal>
        <div class="TextSide">
            <div class="Tag">
                <IconBox stroke="2" />
                <span>Instancias</span>
            </div>
            <div class="Description">
                <h2>Un perfil para cada forma de jugar.</h2>
                <p>Cada instancia tiene su versión, su configuración y su carpeta propia. Jugá vanilla, con mods o con snapshots sin que se pisen entre sí.</p>
            </div>
            <div class="Points sl-stagger">
                <div class="Point">
                    <IconSettings stroke="2" />
                    <div class="PointTxt">
                        <b>Configuración propia</b>
                        <small>RAM, Java, resolución y argumentos por instancia</small>
                    </div>
                </div>
                <div class="Point">
                    <IconCopy stroke="2" />
                    <div class="PointTxt">
                        <b>Cloná y organizá</b>
                        <small>Duplicá, agrupá por tags y marcá favoritas</small>
                    </div>
                </div>
                <div class="Point">
                    <IconShieldCheck stroke="2" />
                    <div class="PointTxt">
                        <b>Verificación y respaldo</b>
                        <small>Verificá archivos y creá backups cuando quieras</small>
                    </div>
                </div>
            </div>
        </div>
        <div class="Image">
            <img :src="captura" alt="Lista de instancias de StepLauncher" loading="lazy" decoding="async">
            <div class="ImageTag">
                <b>Mis instancias</b>
                <small>Versiones y perfiles independientes</small>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Instances{
    width: calc(100% - 10rem);
    padding: 4rem 5rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin: auto;
    gap: 2rem;
    div{
        display:flex;
        justify-content:center;
        align-items:center;
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
            transform: perspective(120px) rotateY(-1deg);
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
    .Instances{
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
