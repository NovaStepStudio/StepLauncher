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
    IconNews,
    IconPhoto,
    IconPuzzle,
    IconBox,
    IconWorld,
    IconShieldCheck,
    IconDatabase,
    IconMusic,
} from '@tabler/icons-vue';
import SettingsModal, { type SectionConfig } from '@/Settings/Settings.vue';
import InstallationModal from '@/Downloads/Installation.vue';
import GeneralSettings from '@/Settings/Sections/General.vue';
import MinecraftSettings from '@/Settings/Sections/Minecraft.vue';
import PersonalizationSettings from '@/Settings/Sections/Personalization.vue';
import AboutSettings from '@/Settings/Sections/About.vue';
import AccountsSettings from '@/Settings/Sections/Accounts.vue';
import NetworkSettings from '@/Settings/Sections/Network.vue';
import DownloadsSettings from '@/Settings/Sections/Downloads.vue';
import IntegritySettings from '@/Settings/Sections/Integrity.vue';
import StorageSettings from '@/Settings/Sections/Storage.vue';
import MusicPanelSettings from '@/Settings/Sections/MusicPanel.vue';
import AccountsModal from '@/Accounts/Manager.vue';
import InstancesModal from '@/Instances/Instances.vue';
import ModsModal from '@/Mods/Mods.vue';
import MusicModal from '@/Music/Music.vue';
import ScreenshotsModal from '@/Screenshots/Screenshots.vue';
import LoginProgressModal from '@/Login/Progress.vue';
import VersionsModal from '@/Versions/Versions.vue';
import CrashModal from '@/Crash/Crash.vue';
import UpdateModal from '@/Updates/Update.vue';
import NewsModal from '@/News/News.vue';
import WelcomeModal from '@/Welcome/Welcome.vue';
import PersonalizationPreviewModal from '@/Settings/PersonalizationPreview.vue';
import DownloadWidget from '@/Downloads/Widget.vue';
import DialogHost from '@/Common/Overlays/Host.vue';
import SplashScreen from '@/Common/Bootstrap/SplashScreen.vue';
import { bootstrapState } from '@/Common/Bootstrap/state';
import { useBackground } from '@/Common/Composables/useBackground';
import {
    selectedLabel,
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
    crashInfo,
} from '@/Launcher/Store';
import { Events } from '@wailsio/runtime';
import { ListScreenshots, SetUIScale } from '@wailsjs/StepLauncher/internal/Services/Appearance/appearanceservice';
import { setUIScale, personalization, uiScale } from '@/Common/Stores/Ui';
import { stopIdleTracking, CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { anyAllActive, allActiveDownloads } from '@/Instances/Store';
import {
    heavyPanel, openHeavyPanel, closeHeavyPanel,
    shotsInstance, shotsReturn,
    settingsOpen, accountsOpen, loginOpen, installOpen, versionsOpen,
    crashOpen, newsOpen, welcomeOpen, previewOpen,
    PERSONALIZATION_PREVIEW_EVENT,
} from '@/Common/Overlays/Store';
import { isOffline, initConnectivity, CONNECTIVITY_ONLINE_EVENT } from '@/Common/Stores/Connectivity';
import OfflineBadge from '@/Common/Components/OfflineBadge.vue';

// ——— Fondos ———
const {
    bg,
    bgImageUrl,
    bgVideoUrl,
    dynamicImage,
    dynamicIndex,
    videoReady,
    videoRef,
    refreshBackground,
    startDynamicTimer,
    onVideoReady,
    onVideoError,
} = useBackground();

// ——— Bootstrap → Welcome ———
watch(() => bootstrapState.value.showWelcome, (v) => {
    if (v && bootstrapState.value.status === 'done') welcomeOpen.value = true;
});
watch(() => bootstrapState.value.status, (s) => {
    if (s === 'done' && bootstrapState.value.showWelcome) welcomeOpen.value = true;
});

// ——— Jugar ———
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
watch(crashInfo, (val) => { crashOpen.value = !!val; });

// ——— Capturas ———
const hasShots = ref(false);
const SHOTS_REFRESH_EVENT = 'sl:shots-refresh';
async function checkShots() {
    try {
        const list = await ListScreenshots();
        hasShots.value = (list?.length ?? 0) > 0;
    } catch { hasShots.value = false; }
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
function openInstances() { openHeavyPanel('instances'); }
function openMods() { openHeavyPanel('mods'); }
function openMusic() { openHeavyPanel('music'); }

// ——— Descargas widget ———
const widgetVisible = computed(() => anyAllActive.value && !installOpen.value);
function openWidget() {
    const d = allActiveDownloads.value[0];
    if (!d) return;
    if (d.kind === 'version') installOpen.value = true;
    else openInstances();
}

// ——— Usuario ———
const userMenuOpen = ref(false);
const mainMenuHidden = computed(
    () => !!heavyPanel.value || accountsOpen.value || versionsOpen.value || newsOpen.value || welcomeOpen.value || previewOpen.value
);
function toggleUserMenu() { userMenuOpen.value = !userMenuOpen.value; }
function closeUserMenu() { userMenuOpen.value = false; }
async function useAccount(id: string) {
    closeUserMenu();
    try { await setSelected(id); } catch (_e) {}
}
function openAccountsManager() { closeUserMenu(); accountsOpen.value = true; }
function openInstallFromVersions() { versionsOpen.value = false; installOpen.value = true; }
async function onPlay() { await launchGame(); }
function onUserMenuDocClick() { closeUserMenu(); }

// ——— Zoom ———
const zoomIndicatorVisible = ref(false);
let zoomIndicatorTimer: number | null = null;
function flashZoomIndicator() {
    zoomIndicatorVisible.value = true;
    if (zoomIndicatorTimer !== null) window.clearTimeout(zoomIndicatorTimer);
    zoomIndicatorTimer = window.setTimeout(() => { zoomIndicatorVisible.value = false; zoomIndicatorTimer = null; }, 1300);
}

// ——— Ajustes secciones ———
const settingsSections: SectionConfig[] = [
    { name: 'General', icon: IconHome, component: GeneralSettings },
    { name: 'Personalización', icon: IconPalette, component: PersonalizationSettings },
    { name: 'Minecraft', icon: IconCpu, component: MinecraftSettings },
    { name: 'Red', icon: IconWorld, component: NetworkSettings },
    { name: 'Descargas', icon: IconDownload, component: DownloadsSettings },
    { name: 'Integridad', icon: IconShieldCheck, component: IntegritySettings },
    { name: 'Almacenamiento', icon: IconDatabase, component: StorageSettings },
    { name: 'Cuentas', icon: IconUsers, component: AccountsSettings },
    { name: 'Música', icon: IconMusic, component: MusicPanelSettings },
    { name: 'Acerca de', icon: IconInfoCircle, component: AboutSettings },
];

// ——— Atajos y preview ———
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
    try { SetUIScale(uiScale.value); } catch (_e) {}
}
function openPersonalizationPreview() {
    settingsOpen.value = false;
    previewOpen.value = true;
}
function closeAllPanels() {
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
    closeHeavyPanel('mods');
    closeHeavyPanel('music');
    closeUserMenu();
    window.dispatchEvent(new CustomEvent(CLOSE_OVERLAYS_EVENT));
}

// ——— Ciclo de vida (solo UI, el bootstrap lo orquesta Main.ts) ———
let gameShotOffs: (() => void)[] = [];
let accountWindowHandlers: { type: string; handler: () => void }[] = [];

onMounted(async () => {
    initConnectivity();
    window.addEventListener('keydown', onKeydown);
    document.addEventListener('click', onUserMenuDocClick);
    checkShots();

    // Al recuperar internet, reintentar subsistemas que lo requieren
    const onReconnect = async () => {
        try {
            const { GetConfig } = await import('@wailsjs/StepLauncher/internal/Services/Config/configservice');
            const cfg = await GetConfig().catch(() => null);
            if (cfg?.launcher?.checkForUpdatesOnStart) {
                const { checkForUpdates } = await import('@/Updates/Store');
                checkForUpdates(true).catch(() => {});
            }
        } catch (_e) {}
        try {
            const { NewsRefreshIndex } = await import('@wailsjs/StepLauncher/internal/Services/System/systemservice');
            NewsRefreshIndex().catch(() => {});
        } catch (_e) {}
    };
    window.addEventListener(CONNECTIVITY_ONLINE_EVENT, onReconnect);
    accountWindowHandlers.push({ type: CONNECTIVITY_ONLINE_EVENT, handler: onReconnect as any });

    const onLoginStart = () => { accountsOpen.value = false; loginOpen.value = true; };
    window.addEventListener(ACCOUNT_LOGIN_START_EVENT, onLoginStart);
    accountWindowHandlers.push({ type: ACCOUNT_LOGIN_START_EVENT, handler: onLoginStart });
    window.addEventListener(PERSONALIZATION_PREVIEW_EVENT, openPersonalizationPreview);

    // Fondos: el bootstrap ya aplicó personalización y precargó blobs;
    // aquí solo se asegura que la UI refleje el estado actual.
    await refreshBackground();
    startDynamicTimer();

    // Sincroniza escalado por si el bootstrap lo cambió
    setUIScale(uiScale.value);

    // Refresco de capturas al cerrar juego (el bootstrap ya loguea el evento)
    gameShotOffs = [
        Events.On('game_exited', onGameClosed),
        Events.On('game_stopped', onGameClosed),
        Events.On('game_crashed', onGameClosed),
    ];
});

onUnmounted(() => {
    window.removeEventListener('keydown', onKeydown);
    document.removeEventListener('click', onUserMenuDocClick);
    stopIdleTracking();
    accountWindowHandlers.forEach((h) => window.removeEventListener(h.type, h.handler));
    accountWindowHandlers = [];
    window.removeEventListener(PERSONALIZATION_PREVIEW_EVENT, openPersonalizationPreview);
    gameShotOffs.forEach((off) => off());
    gameShotOffs = [];
    if (zoomIndicatorTimer !== null) { window.clearTimeout(zoomIndicatorTimer); zoomIndicatorTimer = null; }
});
</script>

<template>
    <SplashScreen />
    <div class="BackgroundLayer" v-if="bg && bg.type !== 'none' && (bgImageUrl || bgVideoUrl || dynamicImage)">
        <Transition name="BgFade">
            <img v-if="bg.type === 'image' && bgImageUrl" :src="bgImageUrl" alt="">
            <video v-else-if="bg.type === 'video' && bgVideoUrl" ref="videoRef" :src="bgVideoUrl" autoplay muted loop playsinline preload="auto" @loadeddata="onVideoReady" @canplay="onVideoReady" @playing="onVideoReady" @error="onVideoError"></video>
            <img v-else-if="bg.type === 'dynamic' && dynamicImage" :key="dynamicIndex" :src="dynamicImage" alt="">
        </Transition>
        <div v-if="bg.type === 'video' && bgVideoUrl && !videoReady" class="BgLoading">
            <img src="../assets/gif/chicken_jockey_run.gif" alt="">
        </div>
        <Transition name="BgAttributionFade">
            <div v-if="bg.type === 'image' && bg.imageAuthor && bg.imageModName" class="BgAttribution">
                <span class="BgAttribution_Author">{{ bg.imageAuthor }}</span>
                <span class="BgAttribution_Dot">•</span>
                <span class="BgAttribution_Mod">{{ bg.imageModName }}</span>
            </div>
        </Transition>
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
            <div class="Item" @click="openMusic">
                <IconMusic class="Item_Icon" stroke="2"/>
                <label class="Item_Label">Música</label>
            </div>
            <div
                class="Item"
                :class="{ offline: isOffline }"
                :title="isOffline ? 'Sin conexión — Mods requiere internet y no está disponible sin conexión.' : undefined"
                @click="isOffline ? undefined : openMods()"
            >
                <IconPuzzle class="Item_Icon" stroke="2"/>
                <label class="Item_Label">Mods</label>
                <OfflineBadge v-if="isOffline" tooltip="right" />
            </div>
            <div
                class="Item"
                :class="{ offline: isOffline }"
                :title="isOffline ? 'Sin conexión — Descargas requiere internet y no está disponible sin conexión.' : undefined"
                @click="isOffline ? undefined : (installOpen = true)"
            >
                <IconDownload class="Item_Icon" stroke="2"/>
                <label class="Item_Label">Descargas</label>
                <OfflineBadge v-if="isOffline" tooltip="right" />
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
                        <img v-else src="../assets/icons/minecraft.png" loading="lazy" decoding="async" fetchpriority="high">
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
    <Transition name="ModsModal">
        <ModsModal v-show="heavyPanel === 'mods'" />
    </Transition>
    <Transition name="MusicModal">
        <MusicModal v-show="heavyPanel === 'music'" />
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
