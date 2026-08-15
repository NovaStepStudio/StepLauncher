<script setup lang="ts">
import { IconBrandGithub, IconBook, IconStar, IconHeart, IconExternalLink, IconInfoCircle, IconLink, IconHeartFilled } from '@tabler/icons-vue';
import appInfo from '../../../../../wails.json';

const appName = appInfo.name ?? 'StepLauncher';
const appVersion = appInfo.version ?? '0.0.0';

const links = [
    {
        label: 'Repositorio oficial',
        desc: 'Todo el código fuente en GitHub: issues, releases, contribuciones y el desarrollo del motor.',
        url: 'https://github.com/NovaStepStudio/StepLauncher',
        icon: IconBrandGithub,
    },
    {
        label: 'Documentación',
        desc: 'Guías de instalación, uso del launcher y documentación técnica del motor en NovaHub.',
        url: 'https://novastepstudios.pages.dev',
        icon: IconBook,
    },
    {
        label: 'NovaStepStudio',
        desc: 'La organización detrás de StepLauncher: explora todos los proyectos open-source del estudio.',
        url: 'https://github.com/NovaStepStudio',
        icon: IconStar,
    },
    {
        label: 'Wails',
        desc: 'El framework (Go + WebView2) que da vida a la interfaz nativa de escritorio del launcher.',
        url: 'https://wails.io',
        icon: IconHeart,
    },
];

function openUrl(url: string) {
    const rt = (window as any).runtime;
    if (rt?.BrowserOpenURL) {
        rt.BrowserOpenURL(url);
        return;
    }
    window.open(url, '_blank');
}
</script>

<template>
<div class="Ss">

    <div class="SsGroup">
        <div class="SsGroupHead">
            <IconInfoCircle :size="'15'" :stroke="'2'" />
            <span>Acerca de</span>
        </div>

        <div class="SsHero">
            <img class="SsHeroLogo" src="../../../assets/logo-step.png" alt="StepLauncher">
            <div class="SsHeroTitle">
                <h3>{{ appName }}</h3>
                <span class="SsHeroVersion">v{{ appVersion }}</span>
            </div>
            <p>
                Launcher de Minecraft creado por <strong>NovaStepStudio</strong>.
                Open-source e impulsado por <strong>Wails</strong>: fondos, colores y fuentes
                personalizables, con un rendimiento pensado para cualquier PC.
            </p>
            <button class="SsBtn SsBtnPrimary" @click="openUrl('https://github.com/NovaStepStudio/StepLauncher')">
                <IconBrandGithub :size="'15'" :stroke="'2'" />
                Ir al repositorio
            </button>
        </div>
    </div>

    <div class="SsGroup">
        <div class="SsGroupHead">
            <IconLink :size="'15'" :stroke="'2'" />
            <span>Recursos</span>
        </div>

        <div class="SsLinks">
            <button v-for="l in links" :key="l.url" class="SsLink" @click="openUrl(l.url)">
                <component :is="l.icon" class="SsLinkIcon" :size="'18'" :stroke="'1.75'" />
                <span class="SsLinkText">
                    <span class="SsLinkLabel">{{ l.label }}</span>
                    <span class="SsLinkDesc">{{ l.desc }}</span>
                </span>
                <IconExternalLink class="SsLinkArrow" :size="'16'" :stroke="'2'" />
            </button>
        </div>
    </div>

    <div class="SsGroup">
        <div class="SsGroupHead">
            <IconHeartFilled :size="'15'" :stroke="'2'" />
            <span>Créditos</span>
        </div>

        <div class="SsCredit">
            <img class="SsWailsLogo" src="../../../assets/logo-wails.png" alt="Wails">
            <div class="SsCreditText">
                <h4>Construido con Wails</h4>
                <p>
                    Wails v2 combina Go y el WebView2 para crear aplicaciones de
                    escritorio nativas y ligeras. Todo el motor del launcher está
                    escrito en Go y la interfaz en Vue 3.
                </p>
            </div>
            <button class="SsBtn" @click="openUrl('https://wails.io')">
                <IconExternalLink :size="'13'" :stroke="'2'" />
                wails.io
            </button>
        </div>
    </div>

</div>
</template>

<style scoped lang="scss">
@use '../Styles/About.scss';
</style>
