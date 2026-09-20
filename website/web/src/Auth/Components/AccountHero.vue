// Cabecera compacta de las páginas de cuenta (login, registro, panel).
// Sigue el patrón FirstPrew de Download/Changelog con el fondo 1.webp.
<script setup lang="ts">
defineProps<{
    titulo: string;
    subtitulo: string;
    descripcion: string;
    insignia?: string;
}>();
</script>

<template>
    <div class="FirstPrew">
        <div class="TextAndButtons">
            <div class="AppName sl-enter" style="--sl-delay: 0s">
                <img class="sl-float" src="../../../assets/logo-step-white.png" alt="StepLauncher" loading="eager" decoding="async" fetchpriority="high">
                <h1>{{ titulo }}</h1>
            </div>
            <div class="Badges sl-enter" style="--sl-delay: .08s">
                <span v-if="insignia" class="BadgeMain">{{ insignia }}</span>
                <span class="Badge">Cuenta StepLauncher</span>
            </div>
            <div class="Description sl-enter" style="--sl-delay: .16s">
                <h2>{{ subtitulo }}</h2>
                <p>{{ descripcion }}</p>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.FirstPrew{
    position: relative;
    width: 100%;
    min-height: 46dvh;
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
        background: linear-gradient(#000000b3, #000000b3), url('../../../assets/background/1.webp');
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
    }
}
</style>
