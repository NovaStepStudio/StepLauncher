// Kit oficial de marca: banner e iconos con descarga directa.
// Los archivos viven en `web/public/branding/` (se sirven tal cual en /branding/*)
// y son copia de la carpeta `branding/` de la raíz del monorepo.
<script setup lang="ts">
import { IconPalette, IconDownload } from '@tabler/icons-vue';

interface Pieza {
    archivo: string;
    nombre: string;
    detalle: string;
    peso: string;
}

const base = '/branding/';

// Avatar del diseñador del kit (URL pública de Discord).
const avatarDisenador = 'https://cdn.discordapp.com/avatars/940098668114087969/30abbeaa8b096e4bc3654d26bcbfec14.png?size=2048';

const banners: Pieza[] = [
    {
        archivo: 'banner-002.png',
        nombre: 'Banner oficial',
        detalle: 'El banner actual de StepLauncher: portada de la web, imagen OpenGraph y cabecera de los READMEs.',
        peso: '≈ 1,3 MB',
    },
    {
        archivo: 'banner-001.png',
        nombre: 'Banner alternativo',
        detalle: 'Primera variante del banner oficial, con la misma identidad.',
        peso: '≈ 1,2 MB',
    },
    {
        archivo: 'banner-003-rectangle.png',
        nombre: 'Banner rectangular',
        detalle: 'Formato rectangular pensado para el repo: ideal como cabecera de READMEs.',
        peso: '≈ 334 KB',
    },
];

const iconos: Pieza[] = [
    {
        archivo: 'appicon-neon.png',
        nombre: 'Icono neón',
        detalle: 'El icono principal: lo usan la app, la web y los builds.',
        peso: '≈ 868 KB',
    },
    {
        archivo: 'appicon-normal.png',
        nombre: 'Icono normal',
        detalle: 'Variante estándar del icono para fondos claros.',
        peso: '≈ 930 KB',
    },
    {
        archivo: 'appicon-black.png',
        nombre: 'Icono negro',
        detalle: 'Variante negra: la usa el launcher en la bandeja del sistema.',
        peso: '≈ 32 KB',
    },
    {
        archivo: 'appicon-white.png',
        nombre: 'Icono blanco',
        detalle: 'Variante blanca para fondos oscuros.',
        peso: '≈ 444 KB',
    },
    {
        archivo: 'appicon-nobg-white.png',
        nombre: 'Icono blanco sin fondo',
        detalle: 'La usan la cabecera y el pie de esta web.',
        peso: '≈ 9 KB',
    },
    {
        archivo: 'appicon-tray.png',
        nombre: 'Icono de bandeja',
        detalle: 'Variante para la bandeja del sistema.',
        peso: '≈ 910 KB',
    },
];

// Baja suave al kit sin tocar la URL (nada de #hash).
function irAlKit() {
    document.getElementById('kit')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
}
</script>

<template>
    <div class="Branding">
        <div class="FirstPrew">
            <div class="TextAndButtons">
                <div class="AppName sl-enter" style="--sl-delay: 0s">
                    <img class="sl-float" src="/branding/appicon-neon.png" alt="StepLauncher" loading="eager" decoding="async" fetchpriority="high">
                    <h1>Branding</h1>
                </div>
                <div class="Badges sl-enter" style="--sl-delay: .08s">
                    <span class="BadgeMain">9 archivos PNG</span>
                    <span class="Badge">Descarga directa</span>
                </div>
                <div class="Description sl-enter" style="--sl-delay: .16s">
                    <h2>El kit oficial de marca.</h2>
                    <p>El banner y todos los iconos de StepLauncher en alta resolución, listos para descargar y usar donde los necesites.</p>
                </div>
                <div class="Buttons sl-enter" style="--sl-delay: .24s">
                    <button class="BtnPrimary" type="button" @click="irAlKit">
                        <IconPalette stroke="2" />
                        Ver el kit
                    </button>
                    <a class="Btn" href="https://www.behance.net/gallery/242802323/Experience-My-Portfolio-D" target="_blank" rel="noopener">
                        Portafolio del diseño
                    </a>
                </div>
                <div class="MiniInfo sl-enter" style="--sl-delay: .32s">
                    <span>Diseño por Reeduroz</span>
                    <i></i>
                    <span>PNG en alta resolución</span>
                </div>
            </div>
        </div>
        <div class="Kit" id="kit" v-reveal>
            <div class="Head">
                <div class="Tag">
                    <IconPalette stroke="2" />
                    <span>Kit oficial</span>
                </div>
                <div class="Description">
                    <h2>Todo el set, listo para descargar.</h2>
                    <p>Tocá descargar en la pieza que necesites: el archivo baja con su nombre original.</p>
                </div>
            </div>
            <article class="Designer" v-reveal>
                <img :src="avatarDisenador" alt="Reeduroz" loading="lazy" decoding="async">
                <div class="Who">
                    <small>Diseñador del branding</small>
                    <b>♦ シ 𝑅𝑒𝑒𝑑𝑢𝑟𝑜𝑠</b>
                    <p>Autor del banner y todos los iconos de este kit.</p>
                    <div class="Links">
                        <a href="https://www.behance.net/gallery/242802323/Experience-My-Portfolio-D" target="_blank" rel="noopener">Portafolio</a>
                        <a href="https://www.behance.net/reedurosxd" target="_blank" rel="noopener">Perfil de Behance</a>
                    </div>
                </div>
            </article>
            <article v-for="item in banners" :key="item.archivo" class="BannerCard">
                <img :src="`${base}${item.archivo}`" :alt="item.nombre" loading="lazy" decoding="async">
                <div class="Info">
                    <div class="Text">
                        <b>{{ item.nombre }}</b>
                        <small>{{ item.archivo }} · {{ item.peso }}</small>
                        <p>{{ item.detalle }}</p>
                    </div>
                    <a class="BtnPrimary" :href="`${base}${item.archivo}`" :download="item.archivo">
                        <IconDownload stroke="2" />
                        Descargar
                    </a>
                </div>
            </article>
            <div class="Grid sl-stagger">
                <article v-for="item in iconos" :key="item.archivo" class="Card">
                    <div class="Preview">
                        <img :src="`${base}${item.archivo}`" :alt="item.nombre" loading="lazy" decoding="async">
                    </div>
                    <b>{{ item.nombre }}</b>
                    <small>{{ item.archivo }} · {{ item.peso }}</small>
                    <p>{{ item.detalle }}</p>
                    <a class="Btn" :href="`${base}${item.archivo}`" :download="item.archivo">
                        <IconDownload stroke="2" />
                        Descargar
                    </a>
                </article>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Branding{
    display: flex;
    flex-direction: column;
    background: #000;
}
.FirstPrew{
    position: relative;
    width: 100%;
    min-height: 58dvh;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-top: 4rem;
    z-index: 1;
    &::after{
        content: '';
        position: absolute;
        inset: 0;
        width: 100%;
        height: 100%;
        mask: linear-gradient(#0008 50%, transparent);
        background: linear-gradient(#000000b3, #000000b3), url('../../assets/background/4.webp');
        background-position: center center;
        background-size: cover;
        z-index: -1;
    }
    div{
        display: flex;
        justify-content: center;
        align-items: center;
    }
    .TextAndButtons{
        width: 100%;
        flex-direction: column;
        align-items: center;
        text-align: center;
        gap: 1rem;
        padding: 0 1.5rem;
        .AppName{
            gap: .5rem;
            font-family: 'Lexend';
            img{
                width: 4rem;
            }
            h1{
                margin: 0;
                font-size: 2.4rem;
            }
        }
        .Badges{
            gap: .5rem;
            .BadgeMain,
            .Badge{
                font-size: .7rem;
                padding: .25rem .65rem;
                border-radius: 99rem;
                border: 1px solid #ffffff25;
                background: #ffffff0d;
            }
            .BadgeMain{
                background: #fff;
                color: #000;
                font-weight: 700;
                border-color: #fff;
            }
        }
        .Description{
            flex-direction: column;
            align-items: center;
            gap: .5rem;
            max-width: 30rem;
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
            }
        }
        .Buttons{
            gap: .6rem;
            a,
            button{
                display: flex;
                justify-content: center;
                align-items: center;
                gap: .45rem;
                padding: .6rem 1.3rem;
                border-radius: .5rem;
                font-size: .85rem;
                font-weight: 600;
                font-family: inherit;
                text-decoration: none;
                cursor: pointer;
                transition: filter 150ms, background 150ms, transform 150ms;
                svg{
                    width: 1.1rem;
                    height: 1.1rem;
                }
            }
            .BtnPrimary{
                background: #fff;
                color: #000;
                border: 1px solid #fff;
                &:hover{
                    filter: brightness(.85);
                    transform: translateY(-1px);
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
        .MiniInfo{
            gap: .6rem;
            font-size: .72rem;
            opacity: .5;
            i{
                width: 3px;
                height: 3px;
                border-radius: 99rem;
                background: #fff;
            }
        }
    }
}
.Kit{
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 2.5rem 1.5rem 4rem 1.5rem;
    gap: 1.5rem;
    .Head{
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;
        text-align: center;
        .Tag{
            display: flex;
            justify-content: center;
            align-items: center;
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
            display: flex;
            flex-direction: column;
            align-items: center;
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
            }
        }
    }
    .Designer{
        display: flex;
        align-items: center;
        gap: 1.2rem;
        width: 100%;
        max-width: 52rem;
        padding: 1.2rem 1.4rem;
        border-radius: .9rem;
        border: 1px solid #ffffff2e;
        background: linear-gradient(135deg, #ffffff14, #ffffff06);
        box-shadow: 0 0 28px #ffffff12, inset 0 1px 0 #ffffff14;
        img{
            width: 5rem;
            height: 5rem;
            flex-shrink: 0;
            border-radius: 99rem;
            object-fit: cover;
            border: 2px solid #ffffff40;
            box-shadow: 0 0 16px #ffffff25;
        }
        .Who{
            display: flex;
            flex-direction: column;
            align-items: flex-start;
            gap: .2rem;
            small{
                font-size: .65rem;
                text-transform: uppercase;
                letter-spacing: .12em;
                opacity: .45;
                font-family: 'Lexend';
            }
            b{
                font-family: 'Lexend';
                font-size: 1.05rem;
            }
            p{
                margin: 0;
                font-size: .8rem;
                opacity: .6;
            }
            .Links{
                display: flex;
                gap: .5rem;
                margin-top: .5rem;
                a{
                    padding: .45rem .9rem;
                    border-radius: .5rem;
                    border: 1px solid #ffffff25;
                    background: #ffffff10;
                    color: #fff;
                    font-size: .75rem;
                    font-weight: 600;
                    text-decoration: none;
                    transition: background 150ms;
                    &:hover{
                        background: #ffffff1c;
                    }
                }
            }
        }
    }
    .BannerCard{
        display: flex;
        flex-direction: column;
        width: 100%;
        max-width: 52rem;
        border-radius: .9rem;
        border: 1px solid #ffffff18;
        background: #ffffff08;
        overflow: hidden;
        img{
            width: 100%;
            display: block;
            object-fit: cover;
        }
        .Info{
            display: flex;
            justify-content: space-between;
            align-items: center;
            gap: 1rem;
            padding: 1rem 1.2rem;
            .Text{
                display: flex;
                flex-direction: column;
                gap: .25rem;
                b{
                    font-family: 'Lexend';
                    font-size: .95rem;
                }
                small{
                    font-size: .7rem;
                    opacity: .5;
                }
                p{
                    margin: .25rem 0 0 0;
                    font-size: .8rem;
                    line-height: 1.6;
                    opacity: .65;
                }
            }
        }
    }
    .Grid{
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
        gap: .7rem;
        width: 100%;
        max-width: 52rem;
        .Card{
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: .35rem;
            padding: 1rem;
            border-radius: .8rem;
            border: 1px solid #ffffff14;
            background: #00000060;
            text-align: center;
            .Preview{
                display: flex;
                justify-content: center;
                align-items: center;
                width: 100%;
                padding: 1.2rem;
                border-radius: .6rem;
                background: #b9bdc8;
                img{
                    width: 5rem;
                    height: 5rem;
                    object-fit: contain;
                }
            }
            b{
                margin-top: .4rem;
                font-size: .82rem;
                font-family: 'Lexend';
            }
            small{
                font-size: .65rem;
                opacity: .5;
            }
            p{
                margin: 0;
                font-size: .75rem;
                line-height: 1.6;
                opacity: .6;
            }
            .Btn{
                margin-top: .4rem;
            }
        }
    }
    .BtnPrimary,
    .Btn{
        display: flex;
        justify-content: center;
        align-items: center;
        gap: .45rem;
        padding: .6rem 1.2rem;
        border-radius: .55rem;
        font-size: .82rem;
        font-weight: 600;
        font-family: inherit;
        text-decoration: none;
        cursor: pointer;
        transition: filter 150ms, background 150ms;
        svg{
            width: 1.05rem;
            height: 1.05rem;
        }
    }
    .BtnPrimary{
        flex-shrink: 0;
        background: #fff;
        color: #000;
        border: 1px solid #fff;
        &:hover{
            filter: brightness(.85);
        }
    }
    .Btn{
        background: #ffffff10;
        color: #fff;
        border: 1px solid #ffffff25;
        &:hover{
            background: #ffffff1c;
        }
    }
}
@media (max-width: 600px){
    .Kit{
        padding: 2rem 1rem 3rem 1rem;
        .Designer{
            flex-direction: column;
            text-align: center;
            .Who{
                align-items: center;
                .Links{
                    justify-content: center;
                }
            }
        }
    .BannerCard{
            .Info{
                flex-direction: column;
                align-items: stretch;
                text-align: center;
                .Text{
                    align-items: center;
                }
            }
        }
    }
}
</style>
