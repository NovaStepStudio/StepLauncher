<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import {
    IconSettings,
    IconDownload,
    IconDeviceGamepad,
    IconHome,
    IconChevronDown,
    IconCpu,
    IconPalette,
    IconInfoCircle,
    IconUsers,
    IconCheck,
    IconBell,
    IconNews,
    IconPhoto,
    IconPuzzle,
    IconBox,
} from '@tabler/icons-vue';
import SettingsModal, { type SectionConfig } from './Modals/SettingsModal.vue';
import InstallationModal from './Modals/InstallationModal.vue';
import GeneralSettings from './Layouts/Sections/Settings/GeneralSettings.vue';
import MinecraftSettings from './Layouts/Sections/Settings/MinecraftSettings.vue';
import PersonalizationSettings from './Layouts/Sections/Settings/PersonalizationSettings.vue';
import AboutSettings from './Layouts/Sections/Settings/AboutSettings.vue';
import AccountsSettings from './Layouts/Sections/Settings/AccountsSettings.vue';
import AccountsModal from './Modals/AccountsModal.vue';
import LoginProgressModal from './Modals/LoginProgressModal.vue';
import VersionsModal from './Modals/VersionsModal.vue';
import CrashModal from './Modals/CrashModal.vue';
import ScreenshotsModal from './Modals/ScreenshotsModal.vue';
import UpdateModal from './Modals/UpdateModal.vue';
import NewsModal from './Modals/NewsModal.vue';
import {
    selectedLabel,
    loadAccounts,
    autoRefresh,
    refreshAllAccounts,
    selectedAccountId,
    accountAvatars,
    accounts,
    setSelected,
    typeLabel,
    ACCOUNT_LOGIN_START_EVENT,
} from './Stores/Accounts';
import {
    hasVersions,
    canLaunch,
    selectedVersion,
    selectedProfile,
    profiles,
    launching,
    launchGame,
    launchMsg,
    launchError,
    launchPrepare,
    launchPrepareText,
    loadVersions,
    loadProfiles,
    refreshAfterDownload,
    crashInfo,
} from './Stores/Launcher';
import { EventsOn } from '@wailsjs/runtime/runtime';
import { ListScreenshots } from '@wailsjs/go/main/App';
import { setUIScale, applyPersonalization, personalization, uiScale, loadLocal, loadLocalFresh } from './Stores/Ui';
import { ensureCustomFonts, fontByType, isBuiltinFont, cleanFontName } from './Stores/Fonts';
import { startIdleTracking, stopIdleTracking, CLOSE_OVERLAYS_EVENT } from './Stores/Idle';
import { download, isDownloading } from './Stores/Downloads';
import { bindUpdateEvents, checkForUpdates } from './Stores/Update';
import { bindNewsEvents } from './Stores/News';

const showSettings = ref(false);
const showAccounts = ref(false);
const showLogin = ref(false);
const showInstall = ref(false);
const showVersions = ref(false);
const showCrash = ref(false);
const showShots = ref(false);
const showNews = ref(false);

const playLabel = computed(() => {
    if (!launching.value) return 'Jugar';
    return launchPrepare.value.active ? 'Descargando…' : 'Lanzando…';
});

const playHint = computed(() => {
    if (launchError.value) return launchError.value;
    if (launchPrepare.value.active) return launchPrepareText.value;
    return launchMsg.value;
});

watch(crashInfo, (val) => {
    showCrash.value = !!val;
});

const hasShots = ref(false);

const SHOTS_REFRESH_EVENT = 'sl:shots-refresh';

async function checkShots() {
    try {
        const list = await ListScreenshots();
        hasShots.value = (list?.length ?? 0) > 0;
    } catch {
        hasShots.value = false;
    }
}

function onGameClosed() {
    checkShots();
    window.dispatchEvent(new CustomEvent(SHOTS_REFRESH_EVENT));
}

function openShots() {
    showShots.value = true;
    checkShots();
}

const widgetPercent = computed(() =>
    Math.round(Math.min(100, Math.max(0, download.value?.percent ?? 0)))
);
const widgetVisible = computed(() => download.value !== null && isDownloading.value && !showInstall.value);

const userMenuOpen = ref(false);

function toggleUserMenu() {
    userMenuOpen.value = !userMenuOpen.value;
}

function closeUserMenu() {
    userMenuOpen.value = false;
}

async function useAccount(id: string) {
    closeUserMenu();
    try {
        await setSelected(id);
    } catch { }
}

function openAccountsManager() {
    closeUserMenu();
    showAccounts.value = true;
}

function openInstallFromVersions() {
    showVersions.value = false;
    showInstall.value = true;
}

async function onPlay() {
    await launchGame();
}

function onUserMenuDocClick() {
    closeUserMenu();
}

const splashVisible = ref(true);
let splashHideTimer: number | null = null;
let splashFailsafeTimer: number | null = null;

let accountEventOffs: (() => void)[] = [];
let accountWindowHandlers: { type: string; handler: () => void }[] = [];
let downloadEventOff: (() => void) | null = null;
let gameEventOffs: (() => void)[] = [];

const zoomIndicatorVisible = ref(false);
let zoomIndicatorTimer: number | null = null;

function flashZoomIndicator() {
    zoomIndicatorVisible.value = true;
    if (zoomIndicatorTimer !== null) {
        window.clearTimeout(zoomIndicatorTimer);
    }
    zoomIndicatorTimer = window.setTimeout(() => {
        zoomIndicatorVisible.value = false;
        zoomIndicatorTimer = null;
    }, 1300);
}

const settingsSections: SectionConfig[] = [
    { name: 'General', icon: IconHome, component: GeneralSettings },
    { name: 'Minecraft', icon: IconCpu, component: MinecraftSettings },
    { name: 'Cuentas', icon: IconUsers, component: AccountsSettings },
    { name: 'Personalización', icon: IconPalette, component: PersonalizationSettings },
    { name: 'Acerca de', icon: IconInfoCircle, component: AboutSettings },
];

const bg = computed(() => personalization.value?.background ?? null);
const bgImageUrl = ref('');
const bgVideoUrl = ref('');
const dynamicUrls = ref<string[]>([]);
const dynamicIndex = ref(0);
const videoReady = ref(false);
const videoRef = ref<HTMLVideoElement | null>(null);
const videoRetries = ref(0);
const MAX_VIDEO_RETRIES = 2;
let dynamicTimer: number | null = null;
let videoCheckTimer: number | null = null;

function stopVideoCheck() {
    if (videoCheckTimer !== null) {
        window.clearInterval(videoCheckTimer);
        videoCheckTimer = null;
    }
}

function startVideoCheck() {
    stopVideoCheck();
    if (bg.value?.type !== 'video') return;
    videoCheckTimer = window.setInterval(() => {
        const v = videoRef.value;
        if (v && v.readyState >= 2) {
            videoReady.value = true;
            stopVideoCheck();
        }
    }, 250);
}

function onVideoReady() {
    videoReady.value = true;
    stopVideoCheck();
}

async function onVideoError() {
    const b = bg.value;
    if (!b || b.type !== 'video' || !b.videoPath || videoRetries.value >= MAX_VIDEO_RETRIES) {
        onVideoReady();
        return;
    }
    videoRetries.value++;
    const fresh = await loadLocalFresh(b.videoPath);
    if (fresh) bgVideoUrl.value = fresh;
    if (videoRef.value) videoRef.value.load();
}

async function refreshBackground() {
    const b = bg.value;
    dynamicIndex.value = 0;
    videoReady.value = false;
    videoRetries.value = 0;
    bgImageUrl.value = b?.type === 'image' ? await loadLocal(b.imagePath ?? '') : '';
    bgVideoUrl.value = b?.type === 'video' ? await loadLocal(b.videoPath ?? '') : '';
    if (b?.type === 'dynamic' && Array.isArray(b.dynamicImages)) {
        dynamicUrls.value = await Promise.all(b.dynamicImages.map((rel: string) => loadLocal(rel)));
    } else {
        dynamicUrls.value = [];
    }
    startVideoCheck();
}

watch(bgVideoUrl, () => {
    videoReady.value = false;
    startVideoCheck();
});

const dynamicImage = computed(() => {
    if (!dynamicUrls.value.length) return '';
    return dynamicUrls.value[dynamicIndex.value % dynamicUrls.value.length];
});

function startDynamicTimer() {
    stopDynamicTimer();
    const b = bg.value;
    if (!b || b.type !== 'dynamic' || dynamicUrls.value.length < 2) return;
    const ms = Math.max(3, b.dynamicInterval) * 1000;
    dynamicTimer = window.setInterval(() => {
        if (b.dynamicOrder === 'random') {
            dynamicIndex.value = Math.floor(Math.random() * dynamicUrls.value.length);
        } else {
            dynamicIndex.value = (dynamicIndex.value + 1) % dynamicUrls.value.length;
        }
    }, ms);
}

function stopDynamicTimer() {
    if (dynamicTimer !== null) {
        window.clearInterval(dynamicTimer);
        dynamicTimer = null;
    }
}

watch(() => personalization.value?.background, () => {
    refreshBackground();
});

watch(dynamicUrls, (urls) => {
    if (urls.length) startDynamicTimer();
    else stopDynamicTimer();
});

function onKeydown(e: KeyboardEvent) {
    const target = e.target as HTMLElement | null;
    if (target && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable)) return;
    if (!(e.ctrlKey || e.metaKey)) return;
    const k = e.key.toLowerCase();
    let next: number | null = null;
    if (k === '=' || k === '+') next = uiScale.value + 10;
    else if (k === '-' || k === '_') next = uiScale.value - 10;
    else if (k === '0') next = 100;
    if (next === null) return;
    e.preventDefault();
    setUIScale(next);
    flashZoomIndicator();
    try {
        (window as any).go?.main?.App?.SetUIScale?.(uiScale.value);
    } catch { }
}

function handleIdle() {
    showSettings.value = false;
    showAccounts.value = false;
    showLogin.value = false;
    showInstall.value = false;
    showVersions.value = false;
    showShots.value = false;
    showNews.value = false;
    closeUserMenu();
    window.dispatchEvent(new CustomEvent(CLOSE_OVERLAYS_EVENT));
}

async function verifyConfigTick() {
    try {
        if (showSettings.value) return;
        const cfg = await (window as any).go?.main?.App?.GetConfig?.();
        if (cfg?.personalization) {
            applyPersonalization(cfg.personalization);
        }
    } catch { }
}

onMounted(async () => {
    window.addEventListener('keydown', onKeydown);
    document.addEventListener('click', onUserMenuDocClick);
    checkShots();

    const onLoginStart = () => {
        showAccounts.value = false;
        showLogin.value = true;
    };
    window.addEventListener(ACCOUNT_LOGIN_START_EVENT, onLoginStart);
    accountWindowHandlers = [{ type: ACCOUNT_LOGIN_START_EVENT, handler: onLoginStart }];

    const splashStart = Date.now();
    const SPLASH_MIN_MS = 900;
    const SPLASH_MAX_MS = 6000;
    let splashHidden = false;
    function hideSplash() {
        if (splashHidden) return;
        splashHidden = true;
        const wait = Math.max(0, SPLASH_MIN_MS - (Date.now() - splashStart));
        splashHideTimer = window.setTimeout(() => {
            splashVisible.value = false;
            splashHideTimer = null;
        }, wait);
    }
    splashFailsafeTimer = window.setTimeout(hideSplash, SPLASH_MAX_MS);

    async function healFontNames(cfg: any) {
        try {
            const assets = await (window as any).go?.main?.App?.GetLauncherAssets?.();
            await ensureCustomFonts(assets);
            const p = cfg?.personalization;
            if (!p || !assets?.fonts?.length) return;
            const eP = fontByType(assets, 'primary');
            const eS = fontByType(assets, 'secundary');
            const pName = eP ? ((eP.name ?? '').trim() || (eP.path ? cleanFontName(eP.path) : '')) : '';
            const sName = eS ? ((eS.name ?? '').trim() || (eS.path ? cleanFontName(eS.path) : '')) : '';
            const next = { ...p };
            let changed = false;
            if (pName && !isBuiltinFont(p.fontPrimary) && p.fontPrimary !== pName) {
                next.fontPrimary = pName;
                changed = true;
            }
            if (sName && !isBuiltinFont(p.fontSecondary) && p.fontSecondary !== sName) {
                next.fontSecondary = sName;
                changed = true;
            }
            if (changed) {
                applyPersonalization(next);
                await (window as any).go?.main?.App?.UpdatePersonalization?.(next);
            }
        } catch { }
    }

    let loadedCfg: any = null;
    try {
        const cfg = await (window as any).go?.main?.App?.GetConfig?.();
        loadedCfg = cfg;
        const raw = cfg?.personalization?.uiScale;
        if (typeof raw === 'number' && raw >= 50 && raw <= 200) {
            setUIScale(raw);
        }
        if (cfg?.personalization) {
            applyPersonalization(cfg.personalization);
        }
        await healFontNames(cfg);
    } catch { }

    const idleCfg = loadedCfg?.idle ?? {};
    startIdleTracking(
        {
            autoCloseModals: idleCfg.autoCloseModals ?? true,
            idleMinutes: idleCfg.idleMinutes ?? 1,
            configCheckEnabled: idleCfg.configCheckEnabled ?? true,
            configCheckMinutes: idleCfg.configCheckMinutes ?? 3,
        },
        handleIdle,
        verifyConfigTick
    );
    verifyConfigTick();

    bindUpdateEvents();
    bindNewsEvents();
    if (loadedCfg?.launcher?.checkForUpdatesOnStart) {
        checkForUpdates(true);
    }

    await refreshBackground();
    startDynamicTimer();

    try {
        await loadAccounts();
        if (autoRefresh.value) {
            await refreshAllAccounts();
        }
    } catch { }
    const offLoginEvt = EventsOn('account_login', () => loadAccounts());
    const offRefreshEvt = EventsOn('account_refresh', () => loadAccounts());
    const offRefreshAllEvt = EventsOn('account_refresh_all', () => loadAccounts());
    accountEventOffs = [offLoginEvt, offRefreshEvt, offRefreshAllEvt];

try {
        await loadVersions();
        await loadProfiles();
    } catch { }
    downloadEventOff = EventsOn('download_state', (raw: unknown) => {
        try {
            const s = typeof raw === 'string' ? raw : JSON.stringify(raw ?? '');
            const obj = JSON.parse(s) as { data?: { state?: string } };
            if (obj?.data?.state === 'completed') refreshAfterDownload();
        } catch { }
    });

    gameEventOffs = [
        EventsOn('game_exited', onGameClosed),
        EventsOn('game_crashed', onGameClosed),
        EventsOn('game_stopped', onGameClosed),
    ];

    setUIScale(uiScale.value);
    if (splashFailsafeTimer !== null) {
        window.clearTimeout(splashFailsafeTimer);
        splashFailsafeTimer = null;
    }
    hideSplash();
});

onUnmounted(() => {
    window.removeEventListener('keydown', onKeydown);
    document.removeEventListener('click', onUserMenuDocClick);
    stopIdleTracking();
    accountEventOffs.forEach((off) => off());
    accountEventOffs = [];
    accountWindowHandlers.forEach((h) => window.removeEventListener(h.type, h.handler));
    accountWindowHandlers = [];
    if (downloadEventOff) {
        downloadEventOff();
        downloadEventOff = null;
    }
    gameEventOffs.forEach((off) => off());
    gameEventOffs = [];
    if (zoomIndicatorTimer !== null) {
        window.clearTimeout(zoomIndicatorTimer);
        zoomIndicatorTimer = null;
    }
    if (splashHideTimer !== null) {
        window.clearTimeout(splashHideTimer);
        splashHideTimer = null;
    }
    if (splashFailsafeTimer !== null) {
        window.clearTimeout(splashFailsafeTimer);
        splashFailsafeTimer = null;
    }
    stopDynamicTimer();
    stopVideoCheck();
});
</script>

<template>
    <Transition name="SplashFade">
        <div v-if="splashVisible" class="SplashScreen">
            <img class="SplashScreen_Logo" src="../assets/logo-step.png" alt="StepLauncher" draggable="false">
            <div class="SplashScreen_Bottom">
                <span class="SplashScreen_Text">Cargando configuración...</span>
                <img class="SplashScreen_Loader" src="../assets/gif/loading.gif" alt="">
            </div>
        </div>
    </Transition>
    <div class="BackgroundLayer" v-if="bg && bg.type !== 'none' && (bgImageUrl || bgVideoUrl || dynamicImage)">
        <Transition name="BgFade">
            <img v-if="bg.type === 'image' && bgImageUrl" :src="bgImageUrl" alt="">
            <video v-else-if="bg.type === 'video' && bgVideoUrl" ref="videoRef" :src="bgVideoUrl" autoplay muted loop playsinline preload="auto" @loadeddata="onVideoReady" @canplay="onVideoReady" @playing="onVideoReady" @error="onVideoError"></video>
            <img v-else-if="bg.type === 'dynamic' && dynamicImage" :key="dynamicIndex" :src="dynamicImage" alt="">
        </Transition>
        <div v-if="bg.type === 'video' && bgVideoUrl && !videoReady" class="BgLoading">
            <img src="../assets/gif/loading.gif" alt="">
        </div>
    </div>
    <Transition name="ZoomFade">
        <div v-if="zoomIndicatorVisible" class="ZoomIndicator">{{ uiScale }}%</div>
    </Transition>
    <main class="MainContent">
        <div class="Sidebar">
            <div class="Item">
                <IconHome class="Item_Icon" stroke="2"/>
                <label class="Item_Label">Inicio</label>
            </div>
            <div v-if="hasShots" class="Item" @click="openShots">
                <IconPhoto class="Item_Icon" stroke="2"/>
                <label class="Item_Label">Fotos</label>
            </div>
            <div class="Item">
                <IconBox class="Item_Icon" stroke="2"/>
                <label class="Item_Label">Instancias</label>
            </div>
            <div class="Item">
                <IconPuzzle class="Item_Icon" stroke="2"/>
                <label class="Item_Label">Mods</label>
            </div>
            <div class="Item" @click="showInstall = true">
                <IconDownload class="Item_Icon" stroke="2"/>
                <label class="Item_Label">Descargas</label>
            </div>
        </div>

        <SettingsModal v-model:visible="showSettings" :sections="settingsSections" />
        <AccountsModal v-model:visible="showAccounts" />
        <LoginProgressModal v-model:visible="showLogin" />
        <InstallationModal v-model:visible="showInstall" />
        <VersionsModal v-model:visible="showVersions" @open-download="openInstallFromVersions" />
        <CrashModal v-model:visible="showCrash" />
        <ScreenshotsModal v-model:visible="showShots" />
        <UpdateModal />
        <NewsModal v-model:visible="showNews" />

        <Transition name="DownloadWidget">
            <button v-if="widgetVisible" class="DownloadWidget" @click="showInstall = true" title="Ver descarga">
                <span class="DownloadWidget_Head">
                    <span class="DownloadWidget_Icon"><IconDownload stroke="2" /></span>
                    <span class="DownloadWidget_Txt">
                        <span class="DownloadWidget_Title">Descargando {{ download?.version }}</span>
                        <span class="DownloadWidget_Sub">Progreso {{ widgetPercent }}%</span>
                    </span>
                </span>
                <span class="DownloadWidget_Bar">
                    <span class="DownloadWidget_BarFill" :style="{ width: widgetPercent + '%' }"></span>
                </span>
            </button>
        </Transition>
        <div class="Content">
            <div v-if="hasVersions" class="BottomControlVersion">
                <div class="VersionSelected" @click="showVersions = true" title="Elegir versión o perfil">
                    <div class="ImageVersion">
                        <img v-if="selectedProfile && profiles[selectedProfile]?.icon" :src="profiles[selectedProfile]?.icon" alt="" loading="lazy" decoding="async" fetchpriority="high">
                        <img v-else src="../assets/not_found/not_found_version.png" loading="lazy" decoding="async" fetchpriority="high">
                    </div>
                    <div class="InfoVersion">
                        <p>Version {{ selectedProfile ? `Perfil • ${selectedProfile}` : 'Seleccionada' }} :</p>
                        <h5>{{ selectedProfile && profiles[selectedProfile]?.version?.trim() ? profiles[selectedProfile]?.version : selectedVersion }}</h5>
                    </div>
                </div>
                <div class="PlayBlock">
                    <div
                        class="PlayButton"
                        :class="{ disabled: !canLaunch || launching }"
                        :disabled="!canLaunch || launching"
                        @click="onPlay"
                    >
                        <IconDeviceGamepad class="Icon" stroke="2"/>
                        <h1>{{ playLabel }}</h1>
                        <div class="Decoration">
                            <img src="../assets/decorations/chicken.png" class="Chicken" loading="lazy" decoding="async" fetchpriority="high">
                            <img src="../assets/decorations/steve_and_alex.png" loading="lazy" decoding="async" fetchpriority="high">
                        </div>
                    </div>
                    <Transition name="LaunchMsgFade">
                        <div v-if="playHint" :class="['LaunchMsg', { error: !!launchError }]">
                            <span>{{ playHint }}</span>
                        </div>
                    </Transition>
                </div>
            </div>
            <div class="TopOptions">
                <div class="Others">
                    <div class="OptionOther">
                        <IconBell stroke="2"/>
                        <label class="OptionLabel">Notificaciones</label>
                    </div>
                    <div class="OptionOther" @click="showNews = true">
                        <IconNews stroke="2"/>
                        <label class="OptionLabel">Noticias</label>
                    </div>
                    <div class="OptionOther" @click="showSettings = true">
                        <IconSettings stroke="2"/>
                        <label class="OptionLabel">Configuracion</label>
                    </div>
                </div>
                <div class="UserCardWrap" @click.stop>
                    <div class="UserCard" @click="toggleUserMenu">
                        <div class="Avatar">
                            <img v-if="accountAvatars[selectedAccountId]" :src="accountAvatars[selectedAccountId]" alt="" loading="lazy" decoding="async" fetchpriority="low">
                            <img v-else src="../assets/not_found/avatar_not_found.png" alt="" loading="lazy" decoding="async" fetchpriority="low">
                        </div>
                        <div class="Username">
                            <h1>{{ selectedLabel.name }}</h1>
                            <p>{{ selectedLabel.sub }}</p>
                        </div>
                        <button class="ExpandButtonProfiles" :class="{ open: userMenuOpen }" @click.stop="toggleUserMenu">
                            <IconChevronDown stroke="2"/>
                        </button>
                    </div>

                    <Transition name="UserMenuFade">
                        <div v-if="userMenuOpen" class="UserMenu">
                            <div class="UserMenu_Head">Cambiar de cuenta</div>
                            <button v-for="a in accounts" :key="a.id" class="UserMenu_Item" :class="{ active: a.id === selectedAccountId }" @click.stop="useAccount(a.id)">
                                <span class="UserMenu_Avatar">
                                    <img v-if="accountAvatars[a.id]" :src="accountAvatars[a.id]" alt="">
                                    <span v-else>{{ a.username.slice(0, 1).toUpperCase() }}</span>
                                </span>
                                <span class="UserMenu_Txt">
                                    <span class="UserMenu_Name">{{ a.username }}</span>
                                    <span class="UserMenu_Sub">{{ typeLabel(a.type) }}</span>
                                </span>
                                <IconCheck v-if="a.id === selectedAccountId" class="UserMenu_Check" stroke="2" />
                            </button>
                            <div v-if="!accounts.length" class="UserMenu_Empty">Aún no hay cuentas. Añade una desde “Gestionar cuentas”.</div>
                            <div class="UserMenu_Divider"></div>
                            <button class="UserMenu_Item UserMenu_Manage" @click.stop="openAccountsManager">
                                <IconUsers class="UserMenu_Icon" stroke="2" />
                                <span class="UserMenu_Txt">
                                    <span class="UserMenu_Name">Gestionar cuentas</span>
                                </span>
                            </button>
                        </div>
                    </Transition>
                </div>
            </div>
        </div>
    </main>
</template>

<style scoped lang="scss">
@use './Styles/App.scss';

.UserCardWrap {
    position: relative;
}

.UserMenu {
    position: absolute;
    top: calc(100% + 0.5rem);
    right: 0;
    width: 12rem;
    max-height: min(60vh, 24rem);
    overflow-y: auto;
    background: var(--background-modal-primary);
    border: var(--border-modal-style);
    border-radius: 0.6rem;
    box-shadow: var(--shadow-settings-normal) #000a;
    padding: 0.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    z-index: 60;
}

.UserMenu_Head {
    font-size: 0.62rem;
    text-transform: uppercase;
    letter-spacing: 0.09em;
    opacity: 0.4;
    padding: 0.25rem 0.5rem 0.35rem;
}

.UserMenu_Item {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 0.55rem;
    padding: 0.45rem 0.5rem;
    border-radius: 0.45rem;
    background: transparent;
    border: 1px solid transparent;
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
    cursor: pointer;
    font-family: var(--font-secundary), Arial, sans-serif;
    text-align: left;

    &:hover {
        background: color-mix(in srgb, var(--background-button-primary) 45%, transparent);
    }

    &.active {
        border: var(--border-style);
        background: color-mix(in srgb, var(--background-button-primary) 35%, transparent);
    }
}

.UserMenu_Avatar {
    width: 1.8rem;
    height: 1.8rem;
    border-radius: 50%;
    flex-shrink: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    overflow: hidden;
    background: color-mix(in srgb, var(--background-button-primary) 60%, transparent);

    img {
        width: 100%;
        height: 100%;
        object-fit: cover;
        border-radius: 50%;
        image-rendering: pixelated;
    }

    span {
        font-size: 0.7rem;
        font-weight: 600;
        color: var(--text-primary);
        text-shadow: var(--text-shadow-primary, none);
    }
}

.UserMenu_Txt {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.05rem;
}

.UserMenu_Name {
    font-size: 0.75rem;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.UserMenu_Sub {
    font-size: 0.62rem;
    opacity: 0.55;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.UserMenu_Check {
    width: 0.95rem;
    height: 0.95rem;
    color: var(--color-success);
    flex-shrink: 0;
}

.UserMenu_Icon {
    width: 1rem;
    height: 1rem;
    opacity: 0.75;
    flex-shrink: 0;
}

.UserMenu_Empty {
    font-size: 0.7rem;
    opacity: 0.5;
    padding: 0.5rem;
    text-align: center;
}

.UserMenu_Divider {
    height: 1px;
    background: rgba(255, 255, 255, 0.08);
    margin: 0.2rem 0;
}

.UserMenu_Manage {
    color: var(--text-secondary);
}

.UserMenuFade-enter-active,
.UserMenuFade-leave-active {
    transition: opacity 160ms ease, transform 160ms ease;
}

.UserMenuFade-enter-from,
.UserMenuFade-leave-to {
    opacity: 0;
    transform: translateY(-4px);
}

.ExpandButtonProfiles {
    svg {
        transition: transform 180ms ease;
    }

    &.open svg {
        transform: rotate(180deg);
    }
}

.PlayBlock {
    position: relative;
    display: flex;
    justify-content: center;
    align-items: center;
}

.PlayButton.disabled {
    opacity: 0.45;
    cursor: not-allowed;
}

.LaunchMsg {
    position: absolute;
    bottom: calc(100% + 0.55rem);
    right: 0;
    max-width: 16rem;
    font-family: var(--font-secundary), Arial, sans-serif;
    font-size: 0.66rem;
    padding: 0.4rem 0.6rem;
    border-radius: 0.45rem;
    background: var(--background-modal-primary);
    border: var(--border-modal-style);
    box-shadow: var(--shadow-settings-normal) #000a;
    color: var(--background-button-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;

    &.error {
        color: var(--color-error);
    }
}

.LaunchMsgFade-enter-active,
.LaunchMsgFade-leave-active {
    transition: opacity 160ms ease, transform 160ms ease;
}

.LaunchMsgFade-enter-from,
.LaunchMsgFade-leave-to {
    opacity: 0;
    transform: translateY(-4px);
}

.DownloadWidget {
    position: fixed;
    right: 0.75rem;
    bottom: 5rem;
    z-index: 5;
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
    width: 16rem;
    max-width: calc(100vw - 1.5rem);
    padding: 0.65rem 0.75rem;
    border-radius: 0.6rem;
    background: var(--background-modal-primary);
    border: var(--border-style);
    box-shadow: var(--shadow-settings-normal) #000a;
    backdrop-filter: var(--filter-blur);
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
    font-family: var(--font-secundary), Arial, sans-serif;
    text-align: left;
    cursor: pointer;
    transition: border-color 150ms, transform 150ms;

    &:hover {
        border-color: color-mix(in srgb, var(--background-button-primary) 75%, white 15%);
        transform: translateY(-2px);
    }
}

.DownloadWidget_Head {
    display: flex;
    align-items: center;
    gap: 0.55rem;
    min-width: 0;
}

.DownloadWidget_Icon {
    width: 2.1rem;
    height: 2.1rem;
    flex-shrink: 0;
    border-radius: 0.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background: color-mix(in srgb, var(--background-button-primary) 60%, transparent);
    border: var(--border-style);
    color: var(--text-primary);

    svg {
        width: 1.05rem;
        height: 1.05rem;
    }
}

.DownloadWidget_Txt {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.06rem;
}

.DownloadWidget_Title {
    font-size: 0.74rem;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.DownloadWidget_Sub {
    font-size: 0.64rem;
    opacity: 0.6;
}

.DownloadWidget_Bar {
    height: 0.42rem;
    border-radius: 0.3rem;
    background: rgba(0, 0, 0, 0.3);
    border: var(--border-style);
    overflow: hidden;
}

.DownloadWidget_BarFill {
    display: block;
    height: 100%;
    border-radius: inherit;
    background: var(--progress-color, var(--background-button-primary));
    transition: width 200ms ease;
}

.DownloadWidget-enter-active,
.DownloadWidget-leave-active {
    transition: opacity 160ms ease, transform 160ms ease;
}

.DownloadWidget-enter-from,
.DownloadWidget-leave-to {
    opacity: 0;
    transform: translateY(8px);
}

.ZoomIndicator {
    position: fixed;
    top: 0.9rem;
    left: 50%;
    transform: translateX(-50%);
    width: 3rem;
    height: 2rem;
    padding: 0 .5rem;
    border-radius: 0.5rem;
    background: var(--background-modal-primary);
    border: var(--border-modal-style);
    box-shadow: var(--shadow-settings-normal) #0008;
    display: flex;
    align-items: center;
    justify-content: center;
    font-family: var(--font-primary), Arial, sans-serif;
    font-size: calc(.85rem * var(--font-size-primary, 1));
    font-weight: 600;
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
    pointer-events: none;
    z-index: 200;
}

.ZoomFade-enter-active,
.ZoomFade-leave-active {
    transition: opacity 120ms ease;
}

.ZoomFade-enter-from,
.ZoomFade-leave-to {
    opacity: 0;
}

.SplashScreen {
    position: fixed;
    inset: 0;
    z-index: 300;
    display: flex;
    justify-content: center;
    align-items: center;
    background: color-mix(in srgb, #111 85%, black 15%);
    overflow: hidden;

    .SplashScreen_Logo {
        width: 7rem;
        max-width: 60vw;
        height: auto;
    }

    .SplashScreen_Bottom {
        position: absolute;
        inset: auto 0 0 0;
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 1rem;
        padding: 0 1.25rem 1.25rem 1.25rem;

        .SplashScreen_Text {
            font-family: var(--font-primary, 'Fredoka'), Arial, Helvetica, sans-serif;
            font-size: calc(0.85rem * var(--font-size-primary, 1));
            color: var(--text-primary, #fff);
            text-shadow: var(--text-shadow-primary, none);
            opacity: 0.75;
        }

        .SplashScreen_Loader {
            width: 1rem;
            height: auto;
            image-rendering: pixelated;
        }
    }
}

.SplashFade-enter-active,
.SplashFade-leave-active {
    transition: opacity 400ms ease;
}

.SplashFade-enter-from,
.SplashFade-leave-to {
    opacity: 0;
}
</style>
