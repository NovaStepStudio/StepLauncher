<script setup lang="ts">
import { IconHeartHandshake, IconTypography, IconPuzzle, IconCpu, IconExternalLink } from '@tabler/icons-vue';

interface Credito {
    nombre: string;
    uso: string;
    licencia: string;
    url: string;
}

// Tipografías que usa StepLauncher en su interfaz.
const tipografias: Credito[] = [
    {
        nombre: 'Inter',
        uso: 'Texto general de la interfaz del launcher y la web: legible en tamaños chicos.',
        licencia: 'OFL-1.1',
        url: 'https://fonts.google.com/specimen/Inter',
    },
    {
        nombre: 'Fredoka',
        uso: 'Títulos y destacados con tono redondeado, propios de la identidad del launcher.',
        licencia: 'OFL-1.1',
        url: 'https://fonts.google.com/specimen/Fredoka',
    },
    {
        nombre: 'Lexend',
        uso: 'Marca, números y encabezados: la que ves en el logo, versiones y cifras.',
        licencia: 'OFL-1.1',
        url: 'https://fonts.google.com/specimen/Lexend',
    },
];

// Librerías del frontend (Vue + utilidades de UI/UX del launcher).
const librerias: Credito[] = [
    {
        nombre: 'Vue',
        uso: 'Framework de la interfaz del launcher y de esta web.',
        licencia: 'MIT',
        url: 'https://vuejs.org',
    },
    {
        nombre: '@tabler/icons-vue',
        uso: 'Set de iconos de toda la interfaz.',
        licencia: 'MIT',
        url: 'https://tabler.io/icons',
    },
    {
        nombre: 'sass-embedded',
        uso: 'Compilador de los estilos SCSS del proyecto.',
        licencia: 'MIT',
        url: 'https://sass-lang.com',
    },
    {
        nombre: 'marked',
        uso: 'Renderizado de Markdown (notas de versión y textos con formato).',
        licencia: 'MIT',
        url: 'https://github.com/markedjs/marked',
    },
    {
        nombre: 'music-metadata',
        uso: 'Lectura de metadatos de audio para el reproductor de música del launcher.',
        licencia: 'MIT',
        url: 'https://github.com/Borewit/music-metadata',
    },
];

// Base nativa en Go: puente de escritorio y utilidades del backend.
const nativo: Credito[] = [
    {
        nombre: 'Wails3',
        uso: 'Puente entre el backend en Go y el frontend en Vue para la app de escritorio (incluye @wailsio/runtime).',
        licencia: 'MIT',
        url: 'https://wails.io',
    },
    {
        nombre: 'go-winio',
        uso: 'Pipes con nombre y utilidades de Windows para la comunicación del backend.',
        licencia: 'MIT',
        url: 'https://github.com/Microsoft/go-winio',
    },
    {
        nombre: 'imaging',
        uso: 'Procesado de imágenes en Go (redimensionar, recortar y convertir).',
        licencia: 'MIT',
        url: 'https://github.com/disintegration/imaging',
    },
    {
        nombre: 'audiometa',
        uso: 'Metadatos de audio en el backend Go, complemento del reproductor.',
        licencia: 'MIT',
        url: 'https://github.com/simonhull/audiometa',
    },
];
</script>

<template>
    <div class="Credits">
        <div class="Hero sl-enter">
            <div class="Tag">
                <IconHeartHandshake stroke="2" />
                <span>Comunidad</span>
            </div>
            <h1>Créditos a Terceros</h1>
            <p>StepLauncher se apoya en tipografías, librerías y herramientas open source. Acá están sus autores y licencias.</p>
            <small>Si usás StepLauncher, también estás usando su trabajo: gracias.</small>
        </div>
        <div class="Body" v-reveal>
            <section>
                <div class="SectionHead">
                    <IconTypography stroke="2" />
                    <h2>Tipografías</h2>
                </div>
                <p>Las tres familias que visten la interfaz, todas con licencia abierta SIL Open Font License.</p>
                <div class="Grid">
                    <article v-for="item in tipografias" :key="item.nombre" class="Card">
                        <b>{{ item.nombre }}</b>
                        <p>{{ item.uso }}</p>
                        <div class="Meta">
                            <span>{{ item.licencia }}</span>
                            <a :href="item.url" target="_blank" rel="noopener">
                                <IconExternalLink stroke="2" />
                                Ver proyecto
                            </a>
                        </div>
                    </article>
                </div>
            </section>
            <section>
                <div class="SectionHead">
                    <IconPuzzle stroke="2" />
                    <h2>Librerías UI/UX</h2>
                </div>
                <p>El frontend del launcher: framework, iconos, estilos y utilidades de contenido y audio.</p>
                <div class="Grid">
                    <article v-for="item in librerias" :key="item.nombre" class="Card">
                        <b>{{ item.nombre }}</b>
                        <p>{{ item.uso }}</p>
                        <div class="Meta">
                            <span>{{ item.licencia }}</span>
                            <a :href="item.url" target="_blank" rel="noopener">
                                <IconExternalLink stroke="2" />
                                Ver proyecto
                            </a>
                        </div>
                    </article>
                </div>
            </section>
            <section>
                <div class="SectionHead">
                    <IconCpu stroke="2" />
                    <h2>Nativo</h2>
                </div>
                <p>La base en Go que convierte la web en app de escritorio y resuelve audio, imágenes y sistema.</p>
                <div class="Grid">
                    <article v-for="item in nativo" :key="item.nombre" class="Card">
                        <b>{{ item.nombre }}</b>
                        <p>{{ item.uso }}</p>
                        <div class="Meta">
                            <span>{{ item.licencia }}</span>
                            <a :href="item.url" target="_blank" rel="noopener">
                                <IconExternalLink stroke="2" />
                                Ver proyecto
                            </a>
                        </div>
                    </article>
                </div>
            </section>
            <section class="Note">
                <h2>Nota de licencias</h2>
                <p>Cada proyecto mantiene su propia licencia y sus autores. Los nombres y enlaces de esta página pertenecen a sus respectivos dueños y se listan solo con fines de atribución. StepLauncher sigue siendo open source bajo GPL-3.0.</p>
            </section>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Credits{
    display:flex;
    flex-direction:column;
    align-items:center;
    background:#000;
    padding: 7rem 1.5rem 4rem 1.5rem;
    .Hero{
        display:flex;
        flex-direction:column;
        align-items:center;
        text-align:center;
        gap: .8rem;
        max-width: 42rem;
        margin-bottom: 2rem;
        .Tag{
            display:flex;
            align-items:center;
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
        h1{
            margin: 0;
            font-family: 'Lexend';
            font-size: 2.2rem;
            font-weight: 600;
        }
        p{
            margin: 0;
            font-size: .95rem;
            line-height: 1.6;
            opacity: .65;
        }
        small{
            font-size: .72rem;
            opacity: .4;
        }
    }
    .Body{
        width: 100%;
        max-width: 52rem;
        display:flex;
        flex-direction:column;
        gap: 2.2rem;
        section{
            .SectionHead{
                display:flex;
                align-items:center;
                gap: .5rem;
                margin-bottom: .4rem;
                svg{
                    width: 1.15rem;
                    height: 1.15rem;
                    opacity: .8;
                }
                h2{
                    margin: 0;
                    font-family: 'Lexend';
                    font-size: 1.05rem;
                    font-weight: 600;
                }
            }
            > p{
                margin: 0 0 1rem 0;
                font-size: .85rem;
                line-height: 1.7;
                opacity: .6;
            }
            .Grid{
                display: grid;
                grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
                gap: .7rem;
                .Card{
                    display:flex;
                    flex-direction:column;
                    gap: .45rem;
                    padding: 1rem 1.1rem;
                    border-radius: .8rem;
                    border: 1px solid #ffffff14;
                    background: #ffffff08;
                    transition: border-color 150ms, background 150ms;
                    &:hover{
                        border-color: #ffffff2e;
                        background: #ffffff0d;
                    }
                    b{
                        font-family: 'Lexend';
                        font-size: .9rem;
                        font-weight: 600;
                    }
                    p{
                        margin: 0;
                        font-size: .78rem;
                        line-height: 1.6;
                        opacity: .65;
                    }
                    .Meta{
                        display:flex;
                        justify-content:space-between;
                        align-items:center;
                        gap: .6rem;
                        margin-top: auto;
                        padding-top: .6rem;
                        span{
                            font-size: .65rem;
                            font-weight: 700;
                            padding: .2rem .55rem;
                            border-radius: 99rem;
                            border: 1px solid #ffffff25;
                            background: #ffffff0d;
                            opacity: .8;
                        }
                        a{
                            display:flex;
                            align-items:center;
                            gap: .3rem;
                            font-size: .75rem;
                            font-weight: 600;
                            color: #fff;
                            text-decoration: none;
                            opacity: .8;
                            transition: opacity 150ms;
                            svg{
                                width: .9rem;
                                height: .9rem;
                            }
                            &:hover{
                                opacity: 1;
                            }
                        }
                    }
                }
            }
        }
        .Note{
            padding: 1.1rem 1.2rem;
            border-radius: .8rem;
            border: 1px solid #ffffff14;
            background: #00000060;
            h2{
                margin: 0 0 .4rem 0;
                font-family: 'Lexend';
                font-size: .95rem;
                font-weight: 600;
            }
            p{
                margin: 0;
                font-size: .8rem;
                line-height: 1.7;
                opacity: .6;
            }
        }
    }
}
@media (max-width: 600px){
    .Credits{
        .Hero{
            h1{
                font-size: 1.8rem;
            }
        }
    }
}
</style>
