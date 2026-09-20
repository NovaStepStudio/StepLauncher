<script setup lang="ts">
import { onMounted } from 'vue';
import { IconBrandGithub, IconStar, IconGitFork } from '@tabler/icons-vue';
import { useReleases } from '@/Common/Composables/useReleases';

const { repo, load } = useReleases();

onMounted(() => load());
</script>

<template>
    <div class="Studio" v-reveal>
        <div class="Brand sl-stagger">
            <div class="BrandName">
                <img src="../../../assets/logo-step-white.png" alt="StepLauncher" loading="lazy" decoding="async">
                <h2>Hecho por NovaStepStudio.</h2>
            </div>
            <p>StepLauncher es gratis y open source bajo licencia GPL-3.0. El código, los releases y el roadmap están a la vista en GitHub.</p>
            <div class="Buttons">
                <a class="BtnPrimary" href="https://github.com/NovaStepStudio/StepLauncher" target="_blank" rel="noopener">
                    <IconBrandGithub stroke="2" />
                    Ver en GitHub
                </a>
                <RouterLink class="Btn" to="/download">
                    Descargar
                </RouterLink>
            </div>
        </div>
        <div class="Stats sl-stagger">
            <div class="Stat">
                <IconStar stroke="2" />
                <div class="StatTxt">
                    <b>{{ repo.stars }}</b>
                    <small>estrellas en GitHub</small>
                </div>
            </div>
            <div class="Stat">
                <IconGitFork stroke="2" />
                <div class="StatTxt">
                    <b>{{ repo.forks }}</b>
                    <small>forks del repo</small>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Studio{
    width: calc(100% - 10rem);
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 3rem;
    margin: auto;
    padding: 4rem 5rem;
    .Brand{
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        gap: 1rem;
        max-width: 30rem;
        .BrandName{
            display: flex;
            align-items: center;
            gap: 1rem;
            img{
                width: 3.5rem;
                height: 3.5rem;
                object-fit: contain;
                flex-shrink: 0;
            }
            h2{
                margin: 0;
                font-family: 'Lexend';
                font-size: 1.7rem;
                font-weight: 600;
                line-height: 1.25;
            }
        }
        p{
            margin: 0;
            font-size: .9rem;
            line-height: 1.6;
            opacity: .65;
        }
        .Buttons{
            display: flex;
            flex-wrap: wrap;
            gap: .6rem;
            a{
                display: flex;
                justify-content: center;
                align-items: center;
                gap: .45rem;
                padding: .6rem 1.3rem;
                border-radius: .5rem;
                font-size: .85rem;
                font-weight: 600;
                text-decoration: none;
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
    }
    .Stats{
        display: flex;
        flex-direction: column;
        gap: 1.6rem;
        padding-left: 3rem;
        border-left: 1px solid #ffffff14;
        .Stat{
            display: flex;
            align-items: center;
            gap: .9rem;
            svg{
                width: 1.4rem;
                height: 1.4rem;
                opacity: .7;
                flex-shrink: 0;
            }
            .StatTxt{
                display: flex;
                flex-direction: column;
                align-items: flex-start;
                gap: .1rem;
                b{
                    font-family: 'Lexend';
                    font-size: 1.9rem;
                    font-weight: 700;
                    line-height: 1;
                }
                small{
                    font-size: .72rem;
                    opacity: .5;
                }
            }
        }
    }
}
@media (max-width: 1000px){
    .Studio{
        flex-direction: column;
        align-items: stretch;
        gap: 2rem;
        width: calc(100% - 3rem);
        padding: 3rem 1.5rem;
        .Brand{
            max-width: 100%;
        }
        .Stats{
            flex-direction: row;
            padding-left: 0;
            padding-top: 1.6rem;
            border-left: 0;
            border-top: 1px solid #ffffff14;
            .Stat{
                flex: 1;
            }
        }
    }
}
</style>
