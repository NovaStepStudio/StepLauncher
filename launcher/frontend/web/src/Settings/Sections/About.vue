<script setup lang="ts">
import { ref } from 'vue';
import { IconBrandGithub, IconBook, IconStar, IconHeart, IconExternalLink, IconInfoCircle, IconLink, IconHeartFilled, IconPackage, IconChevronDown, IconX, IconCode, IconMusic, IconMarkdown, IconPalette, IconTools } from '@tabler/icons-vue';
import { Browser } from '@wailsio/runtime';
import { useAppVersion } from '@/Common/Composables/useAppVersion';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

const appName = 'StepLauncher';
const { appVersion } = useAppVersion();

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
    Browser.OpenURL(url).catch(() => window.open(url, '_blank'));
}

const showCredits = ref(false);
const showLibModal = ref(false);
const selectedLib = ref<any>(null);

const thirdPartyLibs = [
    { name: 'music-metadata', desc: 'Lee metadatos de audio (título, artista, carátula) para la música de fondo.', url: 'https://github.com/Borewit/music-metadata', icon: IconMusic },
    { name: 'marked', desc: 'Convierte Markdown a HTML para las noticias y changelogs.', url: 'https://github.com/markedjs/marked', icon: IconMarkdown },
    { name: 'typescript', desc: 'Tipado estático para todo el frontend. Hace el código más seguro y mantenible.', url: 'https://www.typescriptlang.org/', icon: IconCode },
    { name: 'sass-embedded', desc: 'Compila los estilos SCSS del launcher a CSS.', url: 'https://github.com/sass/dart-sass', icon: IconPalette },
    { name: 'vue', desc: 'Framework progresivo con el que está hecha toda la interfaz.', url: 'https://vuejs.org/', icon: IconStar },
];

const internalLibs = [
    { name: 'Launcher', desc: 'Motor de lanzamiento de Minecraft (perfiles, argumentos, natives, classpath).', path: 'internal/Core/Launcher' },
    { name: 'Downloader', desc: 'Gestor de descargas con verificación, reintentos y límite de velocidad.', path: 'internal/Core/Downloader' },
    { name: 'ModLoader', desc: 'Soporte para Fabric, Forge, NeoForge, Quilt y LegacyFabric.', path: 'internal/Core/ModLoader' },
    { name: 'Accounts', desc: 'Gestión de cuentas offline y Authlib con refresco automático.', path: 'internal/Core/Accounts' },
    { name: 'Assets', desc: 'Gestión de fondos, música y tipografías (launcher_assets.json).', path: 'internal/Core/Assets' },
    { name: 'Tray', desc: 'Bandeja del sistema y menú contextual.', path: 'internal/Tray' },
    { name: 'RichPresence', desc: 'Presencia en Discord.', path: 'internal/RichPresence' },
    { name: 'Updater', desc: 'Actualizador vía GitHub releases.', path: 'internal/Updater' },
    { name: 'Config', desc: 'Persistencia y migración de launcher_config.json.', path: 'internal/Config' },
    { name: 'Platform', desc: 'Detección de RAM, plataforma y utilidades del sistema.', path: 'internal/Core/Platform' },
];

function openLib(lib: any) {
    selectedLib.value = lib;
    showLibModal.value = true;
}

function closeLibModal() {
    showLibModal.value = false;
    selectedLib.value = null;
}

useOverlayEscape(closeLibModal, { isActive: () => showLibModal.value });

function toggleCredits() {
    showCredits.value = !showCredits.value;
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
                    Wails v3 combina Go y el WebView2 para crear aplicaciones de
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

    <div class="SsGroup">
        <div class="SsGroupHead">
            <IconPackage :size="'15'" :stroke="'2'" />
            <span>Bibliotecas</span>
        </div>

        <div class="SsRow">
            <div class="SsInfo">
                <span class="SsLabel">Creditos de terceros</span>
                <span class="SsDesc">Librerias y componentes que hacen posible StepLauncher.</span>
            </div>
            <div class="SsCtrl">
                <button class="SsBtn" @click="toggleCredits">
                    <IconPackage :size="'13'" :stroke="'2'" />
                    {{ showCredits ? 'Ocultar' : 'Ver' }}
                    <IconChevronDown :size="'13'" :stroke="'2'" :class="{ rotated: showCredits }" style="transition: transform 0.2s;" :style="{ transform: showCredits ? 'rotate(180deg)' : 'rotate(0deg)' }" />
                </button>
            </div>
        </div>

        <Transition name="SsCollapse">
            <div v-if="showCredits" class="SsCreditsCollapse">
                <div class="SsWailsHighlight">
                    <div class="SsWailsHighlightHead">
                        <IconHeartFilled :size="'14'" :stroke="'2'" />
                        <span>Mayor aporte al proyecto</span>
                    </div>
                    <h4>Wails 3</h4>
                    <p>Gracias Wails por crear un excelente framework. Esperamos que Wails 3 siga creciendo asi, muchas gracias Wails 3 sin ti no pudimos hacer StepLauncher.</p>
                    <button class="SsBtn SsBtnPrimary" @click="openUrl('https://wails.io')">
                        <IconExternalLink :size="'13'" :stroke="'2'" />
                        wails.io
                    </button>
                </div>

                <div class="SsCreditsSection">
                    <h5><IconTools :size="'12'" :stroke="'2'" /> Librerias de terceros</h5>
                    <div class="SsCreditsList">
                        <button v-for="lib in thirdPartyLibs" :key="lib.name" class="SsCreditItem" @click="openLib(lib)">
                            <component :is="lib.icon" class="SsCreditItemIcon" :size="'16'" :stroke="'1.75'" />
                            <span class="SsCreditItemText">
                                <span class="SsCreditItemName">{{ lib.name }}</span>
                                <span class="SsCreditItemDesc">{{ lib.desc }}</span>
                            </span>
                            <IconExternalLink class="SsCreditItemArrow" :size="'13'" :stroke="'2'" />
                        </button>
                    </div>
                </div>

                <div class="SsCreditsSection">
                    <h5><IconCode :size="'12'" :stroke="'2'" /> Componentes internos del launcher</h5>
                    <p class="SsCreditsHint">Modulos del backend que merecen credito por su aporte al proyecto.</p>
                    <div class="SsCreditsList">
                        <button v-for="lib in internalLibs" :key="lib.name" class="SsCreditItem" @click="openLib(lib)">
                            <IconPackage class="SsCreditItemIcon" :size="'16'" :stroke="'1.75'" />
                            <span class="SsCreditItemText">
                                <span class="SsCreditItemName">{{ lib.name }}</span>
                                <span class="SsCreditItemDesc">{{ lib.desc }}</span>
                            </span>
                            <span class="SsCreditItemPath">{{ lib.path }}</span>
                        </button>
                    </div>
                </div>
            </div>
        </Transition>
    </div>

    <Teleport to="body">
        <Transition name="SsModalFade">
            <div v-if="showLibModal" class="SsLibModalOverlay" @click.self="closeLibModal">
                <div class="SsLibModal">
                    <div class="SsLibModalHead">
                        <h3>{{ selectedLib?.name }}</h3>
                        <button class="SsLibModalClose" @click="closeLibModal">
                            <IconX :size="'14'" :stroke="'2'" />
                        </button>
                    </div>
                    <p class="SsLibModalDesc">{{ selectedLib?.desc }}</p>
                    <p v-if="selectedLib?.path" class="SsLibModalPath">{{ selectedLib?.path }}</p>
                    <div class="SsLibModalActions">
                        <button class="SsBtn" @click="closeLibModal">Cerrar</button>
                        <button v-if="selectedLib?.url" class="SsBtn SsBtnPrimary" @click="openUrl(selectedLib.url)">
                            <IconExternalLink :size="'13'" :stroke="'2'" />
                            Abrir sitio
                        </button>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>

</div>
</template>

<style scoped lang="scss">
@use '../Styles/About.scss';

.SsCreditsCollapse {
    padding: 0 2rem 1.25rem;
    display: flex;
    flex-direction: column;
    gap: 1.1rem;
}

.SsWailsHighlight {
    padding: 1rem 1.1rem;
    border-radius: 0.6rem;
    border: 1px solid color-mix(in srgb, var(--color-success) 30%, transparent);
    background: color-mix(in srgb, var(--color-success) 10%, var(--control-bg));
    display: flex;
    flex-direction: column;
    gap: 0.55rem;

    h4 {
        margin: 0;
        font-size: 0.95rem;
        font-weight: 700;
        font-family: var(--font-primary), Arial, sans-serif;
    }

    p {
        margin: 0;
        font-size: 0.75rem;
        line-height: 1.55;
        opacity: 0.75;
    }
}

.SsWailsHighlightHead {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    opacity: 0.6;
    color: var(--color-success);
}

.SsCreditsSection {
    h5 {
        margin: 0 0 0.55rem;
        display: flex;
        align-items: center;
        gap: 0.35rem;
        font-size: 0.78rem;
        font-weight: 600;
        opacity: 0.85;
    }
}

.SsCreditsHint {
    margin: 0 0 0.6rem;
    font-size: 0.7rem;
    opacity: 0.5;
    line-height: 1.4;
}

.SsCreditsList {
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
}

.SsCreditItem {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    width: 100%;
    padding: 0.7rem 0.85rem;
    border-radius: 0.55rem;
    border: 1px solid var(--control-border);
    background: var(--control-bg);
    text-align: left;
    cursor: pointer;
    font-family: inherit;
    color: var(--text-primary);
    transition: var(--transition);

    &:hover {
        background: rgba(255, 255, 255, 0.07);
        border-color: rgba(255, 255, 255, 0.14);
        transform: translateY(-1px);
    }
}

.SsCreditItemIcon {
    flex-shrink: 0;
    color: var(--text-secondary);
}

.SsCreditItemText {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
}

.SsCreditItemName {
    font-size: 0.78rem;
    font-weight: 600;
}

.SsCreditItemDesc {
    font-size: 0.68rem;
    opacity: 0.5;
    line-height: 1.4;
}

.SsCreditItemPath {
    font-size: 0.62rem;
    opacity: 0.4;
    font-family: monospace;
    flex-shrink: 0;
}

.SsCreditItemArrow {
    flex-shrink: 0;
    opacity: 0.35;
}

.SsLibModalOverlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    padding: 1rem;
}

.SsLibModal {
    width: 22rem;
    max-width: 90vw;
    background: var(--background-modal-primary);
    border: var(--border-modal-style);
    border-radius: 0.75rem;
    box-shadow: var(--shadow-settings-normal) #0008;
    padding: 1.2rem 1.3rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.SsLibModalHead {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;

    h3 {
        margin: 0;
        font-size: 0.95rem;
        font-weight: 700;
    }
}

.SsLibModalClose {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 0.25rem;
    border-radius: 0.35rem;
    display: flex;

    &:hover {
        background: rgba(255, 255, 255, 0.08);
        color: var(--text-primary);
    }
}

.SsLibModalDesc {
    margin: 0;
    font-size: 0.78rem;
    line-height: 1.55;
    opacity: 0.7;
}

.SsLibModalPath {
    margin: 0;
    font-size: 0.68rem;
    font-family: monospace;
    opacity: 0.45;
    background: var(--control-bg);
    border: 1px solid var(--control-border);
    border-radius: 0.4rem;
    padding: 0.35rem 0.5rem;
}

.SsLibModalActions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    margin-top: 0.25rem;
}

.SsCollapse-enter-active, .SsCollapse-leave-active {
    transition: all 0.22s ease;
    overflow: hidden;
}
.SsCollapse-enter-from, .SsCollapse-leave-to {
    opacity: 0;
    transform: translateY(-6px);
}

.SsModalFade-enter-active, .SsModalFade-leave-active {
    transition: opacity 0.18s ease;
}
.SsModalFade-enter-from, .SsModalFade-leave-to {
    opacity: 0;
}
</style>
