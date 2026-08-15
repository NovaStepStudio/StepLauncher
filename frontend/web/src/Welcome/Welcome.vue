<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
import {
    IconArrowLeft,
    IconArrowRight,
    IconBox,
    IconCheck,
    IconDeviceGamepad,
    IconDownload,
    IconPalette,
    IconPhoto,
    IconUserPlus,
    IconWorld,
} from '@tabler/icons-vue';
import { applyPersonalization, personalization } from '@/Common/Stores/Ui';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';
import bgWelcome1 from '../../assets/background/bg-welcome.webp';
import bgWelcome2 from '../../assets/background/bg-welcome-2.webp';
import bgWelcome3 from '../../assets/background/bg-welcome-3.webp';
import havingFun from '../../assets/decorations/having-fun.webp';
import steveAlex from '../../assets/decorations/steve_and_alex.png';
import minecraftIcon from '../../assets/icons/minecraft.png';
import {
    createAccount,
    loginAuthlib,
    loadAccounts,
    type AuthlibLoginReq,
} from '@/Accounts/Store';
import chickenRun from '../../assets/gif/chicken_jockey_run.gif';
import logostep from '../../assets/logo-step.png';

const props = defineProps<{
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
}>();

type Step = 'welcome' | 'customize' | 'directory' | 'account';

const step = ref<Step>('welcome');
const backStep = ref<Step>('welcome');
const mode = ref<'offline' | 'authlib'>('offline');
const busy = ref(false);
const msg = ref('');
const msgOk = ref(true);
const switching = ref(false);

interface DirectorySettings {
    mode: string;
    customPath: string;
    configured: boolean;
    workDir: string;
    normalDir: string;
    minecraftDir: string;
    minecraftExists: boolean;
    portableDir: string;
}

const dirInfo = ref<DirectorySettings>({
    mode: 'normal',
    customPath: '',
    configured: true,
    workDir: '',
    normalDir: '',
    minecraftDir: '',
    minecraftExists: false,
    portableDir: '',
});
const dirConfigured = ref(false);
const dirMode = ref<string>('normal');
const customPath = ref('');
const dirBusy = ref(false);
const dirMsg = ref('');
const dirMsgOk = ref(true);
const restartPending = ref(false);

const dirModes = [
    { id: 'normal', name: 'Normal', desc: 'Carpeta predeterminada (.StepLauncher)' },
    { id: 'minecraft', name: 'Minecraft', desc: 'Usa la carpeta .minecraft oficial' },
    { id: 'portable', name: 'Portable', desc: 'Junto al ejecutable del launcher' },
    { id: 'custom', name: 'Personalizada', desc: 'Elige tú la carpeta' },
];

const dirPreview = computed(() => {
    switch (dirMode.value) {
        case 'minecraft':
            return dirInfo.value.minecraftDir;
        case 'portable':
            return dirInfo.value.portableDir;
        case 'custom':
            return customPath.value.trim();
        default:
            return dirInfo.value.normalDir;
    }
});

const dirChanged = computed(() => {
    if (dirMode.value !== dirInfo.value.mode) return true;
    if (dirMode.value === 'custom' && customPath.value.trim() !== dirInfo.value.customPath) return true;
    return false;
});

const totalSteps = () => (dirConfigured.value ? 3 : 4);

const offlineForm = reactive<{ name: string; username: string }>({ name: '', username: '' });
const authForm = reactive<AuthlibLoginReq>({ authServerUrl: '', username: '', password: '' });

const toggles = reactive<{ animations: boolean; blur: boolean; shadows: boolean }>({
    animations: true,
    blur: true,
    shadows: true,
});

const palettes = [
    { name: 'Default', accent: '#9a9a9a', buttonPrimary: '#111', tag: '#a974ff', progress: '#5ed89a' },
    { name: 'Esmeralda', accent: '#5ed89a', buttonPrimary: '#1f9e6e', tag: '#5ed89a', progress: '#5ed89a' },
    { name: 'Púrpura', accent: '#a974ff', buttonPrimary: '#7a4dff', tag: '#a974ff', progress: '#a974ff' },
    { name: 'Azul', accent: '#4da3ff', buttonPrimary: '#2f7fe0', tag: '#5aa7ff', progress: '#5aa7ff' },
    { name: 'Naranja', accent: '#ff9f45', buttonPrimary: '#d97a1f', tag: '#ffb347', progress: '#ffb347' },
    { name: 'Rosa', accent: '#ff7ac6', buttonPrimary: '#d63f8f', tag: '#ff7ac6', progress: '#ff7ac6' },
    { name: 'Rojo', accent: '#ff6b6b', buttonPrimary: '#c43c3c', tag: '#ff6b6b', progress: '#ff6b6b' },
];

let switchTimer: number | null = null;

const stepOrder = computed<Step[]>(() =>
    dirConfigured.value
        ? ['welcome', 'customize', 'account']
        : ['welcome', 'directory', 'customize', 'account']
);

const stepIndex = (s: Step) => stepOrder.value.indexOf(s) + 1;
const currentStepIndex = () => stepIndex(step.value);

const bgKey = computed(() => step.value);
const bgImage = computed(() =>
    step.value === 'welcome' ? bgWelcome1 : step.value === 'customize' ? bgWelcome2 : bgWelcome3
);

const activePaletteName = computed(() => palettes.find((p) => isPaletteActive(p))?.name ?? 'Default');

const canSubmit = () => {
    if (mode.value === 'offline') return offlineForm.username.trim().length > 0;
    return authForm.authServerUrl.trim().length > 0 && authForm.username.trim().length > 0 && authForm.password.length > 0;
};

function syncToggles() {
    const p = personalization.value;
    toggles.animations = p?.animations ?? true;
    toggles.blur = p?.blur ?? true;
    toggles.shadows = p?.shadows ?? true;
}

function isPaletteActive(pal: (typeof palettes)[number]) {
    const c = personalization.value?.colors;
    return !!c && c.buttonPrimary.toLowerCase() === pal.buttonPrimary.toLowerCase();
}

watch(
    () => props.visible,
    (v) => {
        if (switchTimer !== null) {
            window.clearTimeout(switchTimer);
            switchTimer = null;
        }
        step.value = 'welcome';
        backStep.value = 'welcome';
        mode.value = 'offline';
        busy.value = false;
        msg.value = '';
        msgOk.value = true;
        switching.value = false;
        offlineForm.name = '';
        offlineForm.username = '';
        authForm.authServerUrl = '';
        authForm.username = '';
        authForm.password = '';
        syncToggles();
        loadDirInfo();
    }
);

async function loadDirInfo() {
    try {
        const info = await (window as any)?.go?.main?.App?.GetDirectorySettings?.();
        if (!info) return;
        dirInfo.value = info;
        dirConfigured.value = info.configured;
        dirMode.value = info.mode ?? 'normal';
        customPath.value = info.customPath ?? '';
        dirMsg.value = '';
        dirMsgOk.value = true;
        restartPending.value = false;
    } catch { }
}

function nextFromCustomize() {
    goTo('account');
}

function useMinecraftDir() {
    dirMode.value = 'minecraft';
    customPath.value = '';
}

async function pickCustomDir() {
    try {
        const p = await (window as any)?.go?.main?.App?.PickDirectory?.();
        if (p) customPath.value = p;
    } catch { }
}

async function saveDirectory() {
    if (dirBusy.value) return;
    dirBusy.value = true;
    dirMsg.value = '';
    dirMsgOk.value = true;
    try {
        const err = await (window as any)?.go?.main?.App?.SetDirectoryMode?.(
            dirMode.value,
            dirMode.value === 'custom' ? customPath.value.trim() : ''
        );
        if (err) {
            dirMsg.value = typeof err === 'string' ? err : 'No se pudo configurar la carpeta';
            dirMsgOk.value = false;
            dirBusy.value = false;
            return;
        }
        dirConfigured.value = true;
        dirBusy.value = false;
        if (dirChanged.value) {
            restartPending.value = true;
            try {
                await (window as any)?.go?.main?.App?.RestartApp?.();
            } catch { }
        } else {
            goTo('customize');
        }
    } catch {
        dirBusy.value = false;
        dirMsg.value = 'No se pudo configurar la carpeta';
        dirMsgOk.value = false;
    }
}

function close() {
    emit('update:visible', false);
}

useOverlayEscape(close, { isActive: () => props.visible });

function finishOnboarding() {
    try {
        (window as any)?.go?.main?.App?.SetFirstLaunchDone?.();
    } catch { }
}

function goTo(next: Step) {
    backStep.value = step.value;
    switching.value = true;
    switchTimer = window.setTimeout(() => {
        step.value = next;
        switchTimer = window.setTimeout(() => {
            switching.value = false;
            switchTimer = null;
        }, 420);
    }, 350);
}

async function applyPalette(pal: (typeof palettes)[number]) {
    const p = personalization.value;
    if (!p) return;
    const next = {
        ...p,
        colors: {
            ...p.colors,
            buttonPrimary: pal.buttonPrimary,
            tag: pal.tag,
            progress: pal.progress,
        },
    };
    applyPersonalization(next);
    try {
        await (window as any)?.go?.main?.App?.UpdatePersonalization?.(next);
    } catch { }
}

async function toggle(kind: 'animations' | 'blur' | 'shadows') {
    const p = personalization.value;
    if (!p) return;
    const next = { ...p };
    if (kind === 'animations') next.animations = !next.animations;
    else if (kind === 'blur') next.blur = !next.blur;
    else next.shadows = !next.shadows;
    toggles[kind] = !toggles[kind];
    applyPersonalization(next);
    try {
        await (window as any)?.go?.main?.App?.UpdatePersonalization?.(next);
    } catch { }
}

async function submit() {
    if (busy.value || !canSubmit()) return;
    busy.value = true;
    msg.value = '';
    msgOk.value = true;
    let err = '';
    if (mode.value === 'offline') {
        err = await createAccount({ type: 'offline', name: offlineForm.name, username: offlineForm.username });
    } else {
        const res = await loginAuthlib(authForm);
        err = typeof res === 'string' ? res : res.ok ? '' : res.error;
    }
    busy.value = false;
    if (err) {
        msg.value = err;
        msgOk.value = false;
        return;
    }
    await loadAccounts();
    finishOnboarding();
    close();
}
</script>

<template>
    <Teleport to="body">
        <Transition name="WelcomeModal">
            <div v-if="visible" class="WelcomeModal_Overlay">
                <div class="WelcomeModal_Stage">
                    <Transition name="WelcomeBg" mode="out-in">
                        <img v-if="!switching" :key="bgKey" class="WelcomeModal_StageImg" :src="bgImage" alt="" draggable="false" />
                    </Transition>
                    <div class="WelcomeModal_StageShade"></div>
                    <Transition name="WelcomeSwitch">
                        <div v-if="switching" class="WelcomeModal_StageChicken">
                            <img :src="chickenRun" alt="" draggable="false" />
                        </div>
                    </Transition>

                    <Transition name="WelcomePreview">
                        <div v-if="step === 'customize' && !switching" class="WelcomeModal_Preview">
                            <div class="WelcomeModal_PreviewHead">
                                <IconPalette stroke="2" />
                                <span>Vista previa</span>
                                <small>{{ activePaletteName }}</small>
                            </div>
                            <div class="WelcomeModal_PreviewMock">
                                <div class="WelcomeModal_PreviewSide">
                                    <span class="WelcomeModal_PreviewSideItem on"><IconBox stroke="2" /></span>
                                    <span class="WelcomeModal_PreviewSideItem"><IconPhoto stroke="2" /></span>
                                    <span class="WelcomeModal_PreviewSideItem"><IconDownload stroke="2" /></span>
                                </div>
                                <div class="WelcomeModal_PreviewMain">
                                    <div class="WelcomeModal_PreviewVersion">
                                        <span class="WelcomeModal_PreviewThumb">
                                            <img :src="minecraftIcon" loading="lazy" decoding="async">
                                        </span>
                                        <span class="WelcomeModal_PreviewVerTxt">
                                            <b>1.21.4</b>
                                            <small>Fabric • Perfil</small>
                                        </span>
                                        <span class="WelcomeModal_PreviewBadge">Lista</span>
                                    </div>
                                    <div class="WelcomeModal_PreviewPlay">
                                        <IconDeviceGamepad stroke="2" />
                                        <b>JUGAR</b>
                                    </div>
                                    <div class="WelcomeModal_PreviewRow">
                                        <span class="WelcomeModal_PreviewBar" :key="activePaletteName">
                                            <span class="fill"></span>
                                        </span>
                                        <span class="WelcomeModal_PreviewPct">72%</span>
                                    </div>
                                    <span class="WelcomeModal_PreviewMsg"><IconCheck stroke="2" /> Descarga completada</span>
                                </div>
                            </div>
                        </div>
                    </Transition>

                    <Transition name="WelcomeDecor">
                        <div v-if="step === 'account' && !switching" class="WelcomeModal_Decor">
                            <img class="WelcomeModal_DecorMain" :src="havingFun" alt="" draggable="false" />
                            <img class="WelcomeModal_DecorChars" :src="steveAlex" alt="" draggable="false" />
                        </div>
                    </Transition>
                </div>

                <div class="WelcomeModal_Content">
                    <Transition name="WelcomeSlide" mode="out-in">
                        <section v-if="step === 'welcome'" :key="'welcome'" class="WelcomeModal_Step WelcomeModal_Welcome">
                        <div class="WelcomeModal_Steps">
                            <i v-for="n in totalSteps" :key="n" :class="{ on: n <= currentStepIndex() }"></i>
                        </div>
                        <div class="Title">
                            <img :src="logostep" alt="" draggable="false" />
                            <div>
                                <h1>StepLauncher</h1>
                                <span class="WelcomeModal_Badge">Primera vez</span>
                            </div>
                        </div>
                        <h2>¡Bienvenido a StepLauncher!</h2>
                        <p v-if="!dirConfigured">
                            Tu launcher de Minecraft está listo. Antes de continuar, elige la carpeta donde se
                            guardarán tus versiones, instancias y configuraciones.
                        </p>
                        <p v-else>
                            Tu launcher de Minecraft está listo. Antes de jugar necesitas una cuenta,
                            pero primero puedes dejar tu launcher a tu gusto.
                        </p>
                        <div class="WelcomeModal_Cards">
                            <button v-if="!dirConfigured" class="WelcomeModal_Card is-main" @click="goTo('directory')">
                                <span class="WelcomeModal_CardIcon"><IconBox stroke="2" /></span>
                                <span class="WelcomeModal_CardTxt">
                                    <b>Configurar la carpeta</b>
                                    <small>Dónde se guardarán versiones e instancias</small>
                                </span>
                                <span class="WelcomeModal_CardBadge">Obligatoria</span>
                            </button>
                            <template v-else>
                                <button class="WelcomeModal_Card is-main" @click="goTo('customize')">
                                    <span class="WelcomeModal_CardIcon"><IconPalette stroke="2" /></span>
                                    <span class="WelcomeModal_CardTxt">
                                        <b>Personalizar el launcher</b>
                                        <small>Colores de acento y efectos visuales</small>
                                    </span>
                                    <span class="WelcomeModal_CardBadge">Recomendado</span>
                                </button>
                                <button class="WelcomeModal_Card" @click="goTo('account')">
                                    <span class="WelcomeModal_CardIcon"><IconUserPlus stroke="2" /></span>
                                    <span class="WelcomeModal_CardTxt">
                                        <b>Configurar mi cuenta</b>
                                        <small>Necesaria para poder jugar</small>
                                    </span>
                                    <span class="WelcomeModal_CardBadge">Obligatoria</span>
                                </button>
                            </template>
                        </div>
                        <div class="WelcomeModal_Features">
                            <span><IconDeviceGamepad stroke="2" /> Jugar</span>
                            <span><IconBox stroke="2" /> Instancias</span>
                            <span><IconWorld stroke="2" /> Skins</span>
                            <span><IconPalette stroke="2" /> Personalizable</span>
                        </div>
                    </section>

                    <section v-else-if="step === 'directory'" :key="'directory'" class="WelcomeModal_Step WelcomeModal_Directory">
                        <button class="WelcomeModal_Back" @click="goTo(backStep)">
                            <IconArrowLeft stroke="2" /> Volver
                        </button>
                        <div class="WelcomeModal_Steps">
                            <i v-for="n in totalSteps" :key="n" :class="{ on: n <= currentStepIndex() }"></i>
                        </div>
                        <div class="WelcomeModal_FormHead">
                            <h3>Carpeta del launcher</h3>
                            <p>Elige dónde se guardan las versiones, instancias y configuraciones. Podrás cambiarlo más tarde desde los ajustes.</p>
                        </div>

                        <div v-if="dirInfo.minecraftExists && dirMode !== 'minecraft'" class="WelcomeModal_DirNotice">
                            <IconBox stroke="2" />
                            <span>
                                <b>Detectamos una instalación de Minecraft</b>
                                <small>Se encontró la carpeta {{ dirInfo.minecraftDir }}. ¿Quieres usarla?</small>
                            </span>
                            <button class="WelcomeModal_Btn WelcomeModal_BtnGhost" @click="useMinecraftDir">Usar</button>
                        </div>

                        <select class="WelcomeModal_Select" v-model="dirMode">
                                <option v-for="m in dirModes" :key="m.id" :value="m.id">{{ m.name }} — {{ m.desc }}</option>
                            </select>
                        <p class="WelcomeModal_DirHint">Se guardará en: {{ dirPreview || '—' }}</p>

                        <template v-if="dirMode === 'custom'">
                            <label class="WelcomeModal_Field">
                                <span>Ruta personalizada</span>
                                <div class="WelcomeModal_DirPickRow">
                                    <input class="WelcomeModal_Input" v-model="customPath" placeholder="C:\Users\TuUsuario\AppData\Roaming\.StepLauncher" autocomplete="off" />
                                    <button class="WelcomeModal_Btn WelcomeModal_BtnGhost" :disabled="dirBusy" @click="pickCustomDir">Examinar</button>
                                </div>
                            </label>
                        </template>

                        <p v-if="dirMsg" :class="['WelcomeModal_Msg', { error: !dirMsgOk }]">{{ dirMsg }}</p>

                        <div class="WelcomeModal_Footer">
                            <button
                                class="WelcomeModal_Btn WelcomeModal_BtnPrimary"
                                :disabled="dirBusy || (dirMode === 'custom' && !customPath.trim())"
                                @click="saveDirectory"
                            >
                                <img v-if="restartPending" class="WelcomeModal_Spinner" :src="chickenRun" alt="" draggable="false" />
                                {{ restartPending ? 'Reiniciando...' : dirChanged ? 'Guardar y reiniciar' : 'Continuar' }}
                                <IconArrowRight v-if="!restartPending" stroke="2" />
                            </button>
                        </div>
                    </section>

                    <section v-else-if="step === 'customize'" :key="'customize'" class="WelcomeModal_Step WelcomeModal_Customize">
                        <button class="WelcomeModal_Back" @click="goTo('welcome')">
                            <IconArrowLeft stroke="2" /> Volver
                        </button>
                        <div class="WelcomeModal_Steps">
                            <i v-for="n in totalSteps" :key="n" :class="{ on: n <= currentStepIndex() }"></i>
                        </div>
                        <div class="WelcomeModal_FormHead">
                            <h3>Configurar el launcher</h3>
                            <p>Elige un acento de color y tus preferencias. Podrás cambiarlo cuando quieras.</p>
                        </div>

                        <div class="WelcomeModal_Palettes">
                            <button
                                v-for="pal in palettes"
                                :key="pal.name"
                                class="WelcomeModal_Palette"
                                :class="{ active: isPaletteActive(pal) }"
                                @click="applyPalette(pal)"
                            >
                                <span class="WelcomeModal_PaletteDot" :style="{ background: pal.accent }"></span>
                                <small>{{ pal.name }}</small>
                                <IconCheck v-if="isPaletteActive(pal)" class="WelcomeModal_PaletteCheck" stroke="2" />
                            </button>
                        </div>

                        <div class="WelcomeModal_Toggles">
                            <label class="WelcomeModal_ToggleRow">
                                <span class="WelcomeModal_ToggleTxt">
                                    <b>Animaciones</b>
                                    <small>Transiciones y efectos visuales</small>
                                </span>
                                <button type="button" class="WelcomeModal_Toggle" :class="{ on: toggles.animations }" @click="toggle('animations')">
                                    <span></span>
                                </button>
                            </label>
                            <label class="WelcomeModal_ToggleRow">
                                <span class="WelcomeModal_ToggleTxt">
                                    <b>Desenfoque</b>
                                    <small>Fondo difuminado en paneles y modales</small>
                                </span>
                                <button type="button" class="WelcomeModal_Toggle" :class="{ on: toggles.blur }" @click="toggle('blur')">
                                    <span></span>
                                </button>
                            </label>
                            <label class="WelcomeModal_ToggleRow">
                                <span class="WelcomeModal_ToggleTxt">
                                    <b>Sombras</b>
                                    <small>Sombras en tarjetas y elementos</small>
                                </span>
                                <button type="button" class="WelcomeModal_Toggle" :class="{ on: toggles.shadows }" @click="toggle('shadows')">
                                    <span></span>
                                </button>
                            </label>
                        </div>

                        <div class="WelcomeModal_Footer">
                            <button class="WelcomeModal_Btn WelcomeModal_BtnPrimary" @click="nextFromCustomize">
                                Continuar
                                <IconArrowRight stroke="2" />
                            </button>
                        </div>
                    </section>

                    <section v-else :key="'account'" class="WelcomeModal_Step WelcomeModal_Account">
                        <button class="WelcomeModal_Back" @click="goTo(backStep)">
                            <IconArrowLeft stroke="2" /> Volver
                        </button>
                        <div class="WelcomeModal_Steps">
                            <i v-for="n in totalSteps" :key="n" :class="{ on: n <= currentStepIndex() }"></i>
                        </div>
                        <div class="WelcomeModal_FormHead">
                            <h3>Tu primera cuenta</h3>
                            <p>Este paso es obligatorio: sin cuenta no podrás pulsar Jugar.</p>
                        </div>

                        <div class="WelcomeModal_Modes">
                            <button
                                class="WelcomeModal_Mode"
                                :class="{ active: mode === 'offline' }"
                                @click="mode = 'offline'"
                            >
                                <IconUserPlus stroke="2" />
                                <span>
                                    <b>Cuenta local</b>
                                    <small>Solo un nombre</small>
                                </span>
                            </button>
                            <button
                                class="WelcomeModal_Mode"
                                :class="{ active: mode === 'authlib' }"
                                @click="mode = 'authlib'"
                            >
                                <IconWorld stroke="2" />
                                <span>
                                    <b>En línea</b>
                                    <small>Servidor AuthLib</small>
                                </span>
                            </button>
                        </div>

                        <template v-if="mode === 'offline'">
                            <label class="WelcomeModal_Field">
                                <span>Nombre de jugador</span>
                                <input class="WelcomeModal_Input" v-model="offlineForm.username" placeholder="Steve" autocomplete="off" />
                            </label>
                            <label class="WelcomeModal_Field">
                                <span>Etiqueta (opcional)</span>
                                <input class="WelcomeModal_Input" v-model="offlineForm.name" placeholder="Mi cuenta offline" autocomplete="off" />
                            </label>
                        </template>

                        <template v-else>
                            <label class="WelcomeModal_Field">
                                <span>Servidor de la cuenta</span>
                                <input class="WelcomeModal_Input" v-model="authForm.authServerUrl" placeholder="https://auth.mi-servidor.net" autocomplete="off" />
                            </label>
                            <label class="WelcomeModal_Field">
                                <span>Email o usuario</span>
                                <input class="WelcomeModal_Input" v-model="authForm.username" placeholder="tucorreo@ejemplo.com" autocomplete="off" />
                            </label>
                            <label class="WelcomeModal_Field">
                                <span>Contraseña</span>
                                <input class="WelcomeModal_Input" type="password" v-model="authForm.password" placeholder="••••••••" autocomplete="off" />
                            </label>
                        </template>

                        <p v-if="msg" :class="['WelcomeModal_Msg', { error: !msgOk }]">{{ msg }}</p>

                        <div class="WelcomeModal_Footer">
                            <button class="WelcomeModal_Btn WelcomeModal_BtnGhost" :disabled="busy" @click="goTo(backStep)">Volver</button>
                            <button
                                class="WelcomeModal_Btn WelcomeModal_BtnPrimary"
                                :disabled="busy || !canSubmit()"
                                @click="submit"
                            >
                                <img v-if="busy" class="WelcomeModal_Spinner" :src="chickenRun" alt="" draggable="false" />
                                {{ mode === 'offline' ? 'Crear cuenta' : 'Iniciar sesión' }}
                            </button>
                        </div>
                    </section>
                    </Transition>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use './Styles/Welcome.scss';
</style>
