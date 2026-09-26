<script setup lang="ts">
import { IconBrandGithub, IconGlobe, IconStar, IconHeart, IconExternalLink, IconInfoCircle, IconLink, IconShieldCheck, IconUsers } from '@tabler/icons-vue';
import { Browser } from '@wailsio/runtime';
import { useAppVersion } from '@/Common/Composables/useAppVersion';
import { accountsOpen } from '@/Common/Overlays/Store';
import bannerImg from '../../../assets/banner.webp';
import logoStep from '../../../assets/logo-step.png';
import logoWails from '../../../assets/logo-wails.png';

const appName = 'StepLauncher';
const { appVersion } = useAppVersion();

// Base de la web pública: créditos y legales completos viven ahí y no se duplican en el launcher.
const webBase = 'https://steplauncher.pages.dev';
const creditsUrl = `${webBase}/credits`;

const links = [
    {
        label: 'Repositorio oficial',
        desc: 'Código fuente, issues y releases en GitHub.',
        url: 'https://github.com/NovaStepStudio/StepLauncher',
        icon: IconBrandGithub,
    },
    {
        label: 'Web oficial',
        desc: 'Descargas, historial de cambios, preguntas y branding.',
        url: webBase,
        icon: IconGlobe,
    },
    {
        label: 'NovaStepStudio',
        desc: 'La organización detrás de StepLauncher y sus proyectos open source.',
        url: 'https://github.com/NovaStepStudio',
        icon: IconStar,
    },
    {
        label: 'Créditos a Terceros',
        desc: 'Tipografías, librerías y herramientas open source que hacen posible el launcher.',
        url: creditsUrl,
        icon: IconHeart,
    },
];

// Textos legales completos en la web: el launcher solo enlaza, no duplica el contenido.
const legalLinks = [
    {
        label: 'Política de privacidad',
        desc: 'Qué datos guarda tu cuenta, para qué se usan y cómo pedir que se borren.',
        url: `${webBase}/privacy`,
        icon: IconShieldCheck,
    },
    {
        label: 'Términos y condiciones',
        desc: 'Las reglas para usar StepLauncher y tu cuenta, cortas y en español.',
        url: `${webBase}/terms`,
        icon: IconShieldCheck,
    },
];

function openUrl(url: string) {
    Browser.OpenURL(url).catch(() => window.open(url, '_blank'));
}

// Abre el gestor de cuentas por encima de los ajustes.
function openAccounts() {
    accountsOpen.value = true;
}
</script>

<template>
<div class="Ss">

    <div class="SsGroup">
        <div class="SsGroupHead">
            <IconInfoCircle :size="'15'" :stroke="'2'" />
            <span>Acerca de</span>
        </div>

        <div class="SsAboutTop">
            <div class="SsBanner">
                <img class="SsBannerImg" :src="bannerImg" alt="StepLauncher — Tu Minecraft, sin rodeos">
            </div>

            <div class="SsIdentity">
                <img class="SsIdentityLogo" :src="logoStep" alt="StepLauncher">
                <div class="SsIdentityMeta">
                    <div class="SsIdentityName">
                        <h3>{{ appName }}</h3>
                        <span class="SsIdentityVersion">v{{ appVersion }}</span>
                    </div>
                    <span class="SsIdentityTag">Tu Minecraft, sin rodeos.</span>
                </div>
            </div>

            <div class="SsChips">
                <span>Open source</span>
                <i></i>
                <span>GPL-3.0</span>
                <i></i>
                <span>Hecho con Wails + Vue</span>
            </div>

            <p class="SsLead">
                Launcher moderno, rápido y open source para <strong>Minecraft: Java Edition</strong>,
                creado por <strong>NovaStepStudio</strong>. Gestiona versiones, modloaders,
                instancias y cuentas, con personalización de fondos, colores y fuentes.
            </p>

            <div class="SsHeroActions">
                <button class="SsBtn SsBtnPrimary" @click="openUrl('https://github.com/NovaStepStudio/StepLauncher')">
                    <IconBrandGithub :size="'15'" :stroke="'2'" />
                    Ir al repositorio
                </button>
                <button class="SsBtn" @click="openUrl(webBase)">
                    <IconGlobe :size="'15'" :stroke="'2'" />
                    Web oficial
                </button>
            </div>
        </div>
    </div>

    <div class="SsGroup">
        <div class="SsGroupHead">
            <IconUsers :size="'15'" :stroke="'2'" />
            <span>Cuentas</span>
        </div>

        <div class="SsCreditCats">
            <div class="SsCreditCat">
                <span class="SsCreditCatName">Offline y online</span>
                <span class="SsCreditCatDesc">Juega sin cuenta (offline) o con tu cuenta StepLauncher u otro servidor Yggdrasil (online). Tus credenciales externas quedan solo en tu dispositivo.</span>
            </div>
            <div class="SsCreditCat">
                <span class="SsCreditCatName">Texturas y sesiones</span>
                <span class="SsCreditCatDesc">Skins, capas y sesiones de juego se gestionan desde el gestor de cuentas.</span>
            </div>
        </div>

        <div class="SsCreditActions">
            <button class="SsBtn SsBtnPrimary" @click="openAccounts">
                <IconUsers :size="'13'" :stroke="'2'" />
                Gestionar cuentas
            </button>
            <button class="SsBtn" @click="openUrl(`${webBase}/faq`)">
                <IconExternalLink :size="'13'" :stroke="'2'" />
                Preguntas frecuentes
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
            <IconHeart :size="'15'" :stroke="'2'" />
            <span>Créditos</span>
        </div>

        <div class="SsCredit">
            <img class="SsWailsLogo" :src="logoWails" alt="Wails">
            <div class="SsCreditText">
                <h4>Construido con Wails</h4>
                <p>
                    Wails v3 une el backend en Go con la interfaz en Vue 3 para una app
                    de escritorio nativa y ligera.
                </p>
            </div>
        </div>

        <div class="SsCreditCats">
            <div class="SsCreditCat">
                <span class="SsCreditCatName">Tipografías</span>
                <span class="SsCreditCatDesc">Inter · Fredoka · Lexend</span>
            </div>
            <div class="SsCreditCat">
                <span class="SsCreditCatName">Librerías UI/UX</span>
                <span class="SsCreditCatDesc">Vue · @tabler/icons-vue · sass-embedded · marked · music-metadata</span>
            </div>
            <div class="SsCreditCat">
                <span class="SsCreditCatName">Nativo</span>
                <span class="SsCreditCatDesc">Wails3 · go-winio · imaging · audiometa</span>
            </div>
        </div>

        <div class="SsCreditActions">
            <button class="SsBtn SsBtnPrimary" @click="openUrl(creditsUrl)">
                <IconExternalLink :size="'13'" :stroke="'2'" />
                Ver créditos a terceros
            </button>
            <span class="SsCreditNote">Autores, licencias y enlaces completos en la web. StepLauncher es GPL-3.0.</span>
        </div>
    </div>

    <div class="SsGroup">
        <div class="SsGroupHead">
            <IconShieldCheck :size="'15'" :stroke="'2'" />
            <span>Legal</span>
        </div>

        <div class="SsLinks">
            <button v-for="l in legalLinks" :key="l.url" class="SsLink" @click="openUrl(l.url)">
                <component :is="l.icon" class="SsLinkIcon" :size="'18'" :stroke="'1.75'" />
                <span class="SsLinkText">
                    <span class="SsLinkLabel">{{ l.label }}</span>
                    <span class="SsLinkDesc">{{ l.desc }}</span>
                </span>
                <IconExternalLink class="SsLinkArrow" :size="'16'" :stroke="'2'" />
            </button>
        </div>
    </div>

</div>
</template>

<style scoped lang="scss">
@use '../Styles/About.scss';
</style>
