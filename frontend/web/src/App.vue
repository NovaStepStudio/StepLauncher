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
import SettingsModal, { type SectionConfig } from '@/Settings/Settings.vue';
import InstallationModal from '@/Downloads/Installation.vue';
import GeneralSettings from '@/Settings/Sections/General.vue';
import MinecraftSettings from '@/Settings/Sections/Minecraft.vue';
import PersonalizationSettings from '@/Settings/Sections/Personalization.vue';
import AboutSettings from '@/Settings/Sections/About.vue';
import AccountsSettings from '@/Settings/Sections/Accounts.vue';
import AccountsModal from '@/Accounts/Manager.vue';
import InstancesModal from '@/Instances/Instances.vue';
import LoginProgressModal from '@/Login/Progress.vue';
import VersionsModal from '@/Versions/Versions.vue';
import CrashModal from '@/Crash/Crash.vue';
import ScreenshotsModal from '@/Screenshots/Screenshots.vue';
import UpdateModal from '@/Updates/Update.vue';
import NewsModal from '@/News/News.vue';
import WelcomeModal from '@/Welcome/Welcome.vue';
import PersonalizationPreviewModal from '@/Settings/PersonalizationPreview.vue';
import DownloadWidget from '@/Downloads/Widget.vue';
import DialogHost from '@/Common/Overlays/Host.vue';
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
} from '@/Accounts/Store';
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
    launchingPhaseLabel,
    loadVersions,
    loadProfiles,
    refreshAfterDownload,
    crashInfo,
    onGameCrash,
    maybeShowWindow,
    hideLaunchMessage,
} from '@/Launcher/Store';
import { EventsOn } from '@wailsjs/runtime/runtime';
import { ListScreenshots } from '@wailsjs/go/main/App';
import { setUIScale, applyPersonalization, personalization, uiScale, loadLocal, loadLocalFresh } from '@/Common/Stores/Ui';
import { ensureCustomFonts, fontByType, isBuiltinFont, cleanFontName } from '@/Common/Stores/Fonts';
import { startIdleTracking, stopIdleTracking, CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { anyAllActive, allActiveDownloads } from '@/Instances/Store';
import {
    heavyPanel, openHeavyPanel, closeHeavyPanel,
    shotsInstance, shotsReturn,
    settingsOpen, accountsOpen, loginOpen, installOpen, versionsOpen,
    crashOpen, newsOpen, welcomeOpen, previewOpen,
    PERSONALIZATION_PREVIEW_EVENT,
} from '@/Common/Overlays/Store';
import { bindUpdateEvents, checkForUpdates } from '@/Updates/Store';
import { bindNewsEvents } from '@/News/Store';

const playLabel = computed(() => {
    if (!launching.value) return 'Jugar';
    return launchingPhaseLabel.value;
});

const playHint = computed(() => {
    if (launchError.value) return launchError.value;
    if (launchPrepare.value.active) return launchPrepareText.value;
    if (launching.value) return launchMsg.value;
    return '';
});

watch(crashInfo, (val) => {
    crashOpen.value = !!val;
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
    openHeavyPanel('shots');
    shotsInstance.value = null;
    shotsReturn.value = false;
    checkShots();
}

function openInstances() {
    openHeavyPanel('instances');
}

const widgetVisible = computed(() => anyAllActive.value && !installOpen.value);

function openWidget() {
    const d = allActiveDownloads.value[0];
    if (!d) return;
    if (d.kind === 'version') installOpen.value = true;
    else openInstances();
}

const userMenuOpen = ref(false);

const mainMenuHidden = computed(
    () =>
        !!heavyPanel.value ||
        accountsOpen.value ||
        versionsOpen.value ||
        newsOpen.value ||
        welcomeOpen.value ||
        previewOpen.value
);

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
    accountsOpen.value = true;
}

function openInstallFromVersions() {
    versionsOpen.value = false;
    installOpen.value = true;
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

function openPersonalizationPreview() {
    settingsOpen.value = false;
    previewOpen.value = true;
}

function handleIdle() {
    settingsOpen.value = false;
    accountsOpen.value = false;
    loginOpen.value = false;
    installOpen.value = false;
    versionsOpen.value = false;
    newsOpen.value = false;
    welcomeOpen.value = false;
    previewOpen.value = false;
    closeHeavyPanel('shots');
    closeHeavyPanel('instances');
    closeUserMenu();
    window.dispatchEvent(new CustomEvent(CLOSE_OVERLAYS_EVENT));
}

async function verifyConfigTick() {
    try {
        if (settingsOpen.value) return;
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
        accountsOpen.value = false;
        loginOpen.value = true;
    };
    window.addEventListener(ACCOUNT_LOGIN_START_EVENT, onLoginStart);
    accountWindowHandlers = [{ type: ACCOUNT_LOGIN_START_EVENT, handler: onLoginStart }];
    window.addEventListener(PERSONALIZATION_PREVIEW_EVENT, openPersonalizationPreview);

    const splashStart = Date.now();
    const SPLASH_MIN_MS = 900;
    const SPLASH_MAX_MS = 6000;
    let splashHidden = false;
    let showWelcomeAfterSplash = false;
    function hideSplash() {
        if (splashHidden) return;
        splashHidden = true;
        const wait = Math.max(0, SPLASH_MIN_MS - (Date.now() - splashStart));
        splashHideTimer = window.setTimeout(() => {
            splashVisible.value = false;
            if (showWelcomeAfterSplash) welcomeOpen.value = true;
            splashHideTimer = null;
        }, wait);
    }
    splashFailsafeTimer = window.setTimeout(hideSplash, SPLASH_MAX_MS);

    try {
        showWelcomeAfterSplash = (await (window as any).go?.main?.App?.GetFirstLaunch?.()) === true;
    } catch {
        showWelcomeAfterSplash = false;
    }

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
        EventsOn('game_started', hideLaunchMessage),
        EventsOn('game_exited', onGameClosed),
        EventsOn('game_crashed', (data: unknown) => {
            onGameCrash(data);
            void maybeShowWindow();
            onGameClosed();
        }),
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
    window.removeEventListener(PERSONALIZATION_PREVIEW_EVENT, openPersonalizationPreview);
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
                <img class="SplashScreen_Loader" src="../assets/gif/chicken_jockey_run.gif" alt="">
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
            <img src="../assets/gif/chicken_jockey_run.gif" alt="">
        </div>
    </div>
    <Transition name="ZoomFade">
        <div v-if="zoomIndicatorVisible" class="ZoomIndicator">{{ uiScale }}%</div>
    </Transition>
    <main class="MainContent" :class="{ menuHidden: mainMenuHidden }">
        <div class="Sidebar">
            <div v-if="hasShots" class="Item" @click="openShots">
                <IconPhoto class="Item_Icon" stroke="2"/>
                <label class="Item_Label">Fotos</label>
            </div>
            <div class="Item" @click="openInstances">
                <IconBox class="Item_Icon" stroke="2"/>
                <label class="Item_Label">Instancias</label>
            </div>
            <div class="Item">
                <IconPuzzle class="Item_Icon" stroke="2"/>
                <label class="Item_Label">Mods</label>
            </div>
            <div class="Item" @click="installOpen = true">
                <IconDownload class="Item_Icon" stroke="2"/>
                <label class="Item_Label">Descargas</label>
            </div>
        </div>

        <Transition name="DownloadWidget">
            <DownloadWidget v-if="widgetVisible" @open="openWidget" />
        </Transition>
        <div class="Content">
            <div v-if="hasVersions" class="BottomControlVersion">
                <div class="VersionSelected" @click="versionsOpen = true" title="Elegir versión o perfil">
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
                    <div class="OptionOther" @click="newsOpen = true">
                        <IconNews stroke="2"/>
                        <label class="OptionLabel">Noticias</label>
                    </div>
                    <div class="OptionOther" @click="settingsOpen = true">
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

    <SettingsModal v-model:visible="settingsOpen" :sections="settingsSections" />
    <AccountsModal v-model:visible="accountsOpen" />
    <Transition name="InstancesModal">
        <InstancesModal v-show="heavyPanel === 'instances'" />
    </Transition>
    <LoginProgressModal v-model:visible="loginOpen" />
    <InstallationModal v-model:visible="installOpen" />
    <VersionsModal v-model:visible="versionsOpen" @open-download="openInstallFromVersions" />
    <CrashModal v-model:visible="crashOpen" />
    <Transition name="ScreenshotsModal">
        <ScreenshotsModal v-show="heavyPanel === 'shots'" />
    </Transition>
    <UpdateModal />
    <NewsModal v-model:visible="newsOpen" />
    <WelcomeModal v-model:visible="welcomeOpen" />
    <PersonalizationPreviewModal v-model:visible="previewOpen" />
    <DialogHost />
</template>

<style scoped lang="scss">
@use './Common/Styles/App/App.scss';
</style>
