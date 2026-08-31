<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import {
    IconPlayerPlay, IconPlayerPause, IconPlayerSkipBack, IconPlayerSkipForward,
    IconRepeat, IconRepeatOnce, IconArrowsShuffle, IconMusic, IconX,
    IconVolume, IconVolumeOff, IconPlaylist, IconDisc, IconLibrary,
} from '@tabler/icons-vue';
import { personalization } from '@/Common/Stores/Ui';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import {
    musicList, playing as bgPlaying, currentTime as bgCurrentTime,
    togglePlay as bgTogglePlay, next as bgNext, prev as bgPrev, seekTo as bgSeekTo, setPlayMode as bgSetPlayMode, playMode as bgPlayMode,
    currentMeta, formatDuration, loadMusicList, musicCoverStyle, musicDiscRotation,
    loadError as bgLoadError,
    volume as bgVolume, setVolume as bgSetVolume,
} from '@/Common/Stores/Music';
import {
    heavyPanel,
    accountsOpen, versionsOpen, newsOpen, welcomeOpen, previewOpen,
} from '@/Common/Overlays/Store';
import { localTracks } from '@/Music/LocalStore';
import {
    currentTrack as launcherTrack, playing as launcherPlaying, currentTime as launcherCurrentTime, duration as launcherDuration,
    queue as launcherQueue, queueTracks as launcherQueueTracks, volume as launcherVolume, playMode as launcherPlayMode,
    shuffleMode as launcherShuffleMode, repeatMode as launcherRepeatMode,
    togglePlay as launcherTogglePlay, next as launcherNext, prev as launcherPrev, seekTo as launcherSeekTo, setVolume as launcherSetVolume, setPlayMode as launcherSetPlayMode,
    toggleShuffleMode as launcherToggleShuffle, toggleRepeatMode as launcherToggleRepeat, playTrack as launcherPlayTrack,
    loadError as launcherLoadError,
} from '@/Music/PlayerStore';

const expanded = ref(false);
const volOpen = ref(false);
const queueOpen = ref(false);
const volBtn = ref<HTMLButtonElement | null>(null);
const volMenuRight = ref('0.5rem');
const volMenuBottom = ref('3.6rem');

// Fuente del widget: fondo vs biblioteca del launcher
const launcherSource = ref<'background' | 'launcher'>((() => {
    try { const v = localStorage.getItem('stl_bgm_source'); if (v === 'launcher' || v === 'background') return v as any; } catch (_e) {}
    return 'background';
})());
const isLauncher = computed(() => launcherSource.value === 'launcher');
function setSource(s: 'background' | 'launcher'): void {
    launcherSource.value = s;
    try { localStorage.setItem('stl_bgm_source', s); } catch (_e) {}
    queueOpen.value = false;
}

// Datos unificados según fuente
const meta = computed(() => currentMeta());
const bgDuration = computed(() => meta.value?.duration ?? 0);
const cover = computed(() => musicCoverStyle());
const isDisc = computed(() => cover.value === 'disc');
const isBg = computed(() => cover.value === 'background');
const spinEnabled = computed(() => isDisc.value && musicDiscRotation());
const multi = computed(() => isLauncher.value ? launcherQueue.value.length > 1 : musicList.value.length > 1);
const multiModes = computed(() => isLauncher.value ? launcherQueue.value.length > 2 : musicList.value.length > 2);

const playing = computed(() => isLauncher.value ? launcherPlaying.value : bgPlaying.value);
const currentTime = computed(() => isLauncher.value ? launcherCurrentTime.value : bgCurrentTime.value);
const duration = computed(() => isLauncher.value ? (launcherDuration.value || launcherTrack.value?.duration || 0) : bgDuration.value);
const loadError = computed(() => isLauncher.value ? launcherLoadError.value : bgLoadError.value);
const volume = computed(() => isLauncher.value ? launcherVolume.value : bgVolume.value);
const playMode = computed(() => isLauncher.value ? launcherPlayMode.value : bgPlayMode.value);
const shuffleMode = computed(() => launcherShuffleMode.value);
const repeatMode = computed(() => launcherRepeatMode.value);
const volLevel = computed(() => Math.round(volume.value * 100));
const isShuffle = computed(() => isLauncher.value ? shuffleMode.value === 'shuffle' : playMode.value === 'shuffle');
const isRepeatOne = computed(() => isLauncher.value ? repeatMode.value === 'one' : playMode.value === 'repeat-one');

const displayTitle = computed(() => isLauncher.value ? (launcherTrack.value?.title || 'Sin reproducción') : (meta.value?.title || 'Sin reproducción'));
const displayArtist = computed(() => isLauncher.value ? (launcherTrack.value?.artist || 'Biblioteca') : (meta.value?.artist || 'Fondo'));
const displayCoverUrl = computed(() => isLauncher.value ? (launcherTrack.value?.coverUrl || '') : (meta.value?.coverUrl || ''));

const launcherHasMusic = computed(() => localTracks.value.length > 0 || launcherQueue.value.length > 0);
// El widget se muestra si hay contenido en la fuente activa
const show = computed(() => {
    if (isLauncher.value) return launcherHasMusic.value;
    const p = personalization.value?.backgroundMusic;
    return !!p?.enabled && musicList.value.length > 0;
});

async function onTogglePlay() {
    if (isLauncher.value) await launcherTogglePlay();
    else await bgTogglePlay();
}

async function onNext() {
    if (isLauncher.value) await launcherNext();
    else await bgNext();
}

async function onPrev() {
    if (isLauncher.value) await launcherPrev();
    else await bgPrev();
}

function onMode(m: 'queue' | 'shuffle' | 'repeat-one') {
    if (isLauncher.value) {
        launcherSetPlayMode(m);
        if (m !== 'repeat-one' && launcherQueue.value.length > 1) void launcherNext();
    } else {
        bgSetPlayMode(m);
        if (m !== 'repeat-one' && musicList.value.length > 1) void bgNext();
    }
}

function onToggleShuffle(): void {
    if (isLauncher.value) launcherToggleShuffle();
    else {
        // Para fondo: alternar entre queue y shuffle
        if (bgPlayMode.value === 'shuffle') bgSetPlayMode('queue');
        else bgSetPlayMode('shuffle');
    }
}
function onToggleRepeat(): void {
    if (isLauncher.value) launcherToggleRepeat();
    else {
        if (bgPlayMode.value === 'repeat-one') bgSetPlayMode('queue');
        else bgSetPlayMode('repeat-one');
    }
}

async function onQueuePlay(t: any): Promise<void> {
    if (isLauncher.value) await launcherPlayTrack(t, launcherQueueTracks.value as any);
}

function onQueueRowClick(t: any): void { void onQueuePlay(t); }

function onSeek(e: Event) {
    const v = Number((e.target as HTMLInputElement).value);
    if (isLauncher.value) void launcherSeekTo(v);
    else bgSeekTo(v);
}

function toggleExpand() {
    expanded.value = !expanded.value;
    // Al plegar la card se cierra también el menú de volumen.
    if (!expanded.value) volOpen.value = false;
}

// ---- Menú de volumen --------------------------------------------------------
// El menú vive FUERA de la card (es hermano, no hijo), por lo que el
// overflow: hidden de la card (modo carátula de fondo) nunca lo recorta y no
// hace falta tocarlo. Se ancla con position: absolute al contenedor .Bgm
// (que es fijo a todo lo ancho del viewport), así los px de right/bottom son
// relativos a la pantalla. Nada de position: fixed en el propio menú.

function toggleVolumeMenu() {
    if (volOpen.value) {
        volOpen.value = false;
        return;
    }
    queueOpen.value = false;
    const el = volBtn.value;
    if (el) {
        const r = el.getBoundingClientRect();
        volMenuRight.value = `${Math.max(8, window.innerWidth - r.right)}px`;
        volMenuBottom.value = `${window.innerHeight - r.top + 8}px`;
    }
    volOpen.value = true;
}

function onVolume(e: Event) {
    const v = Number((e.target as HTMLInputElement).value) / 100;
    if (isLauncher.value) launcherSetVolume(v);
    else bgSetVolume(v);
}

function toggleQueue() {
    if (queueOpen.value) { queueOpen.value = false; return; }
    volOpen.value = false;
    queueOpen.value = true;
}

// Clic fuera del botón (o del propio menú) o Escape cierran el menú de volumen.
function onWindowDown(e: MouseEvent) {
    if (!volOpen.value && !queueOpen.value) return;
    const t = e.target as HTMLElement | null;
    if (t?.closest('.Bgm_Volume') || t?.closest('.Bgm_VolMenu') || t?.closest('.Bgm_QueueBtn') || t?.closest('.Bgm_Queue')) return;
    volOpen.value = false;
    queueOpen.value = false;
}

function onEscapeKey(e: KeyboardEvent) {
    if (e.key === 'Escape') { volOpen.value = false; queueOpen.value = false; }
}

// El menú principal se oculta solo con ciertos paneles (igual que App.vue:158 mainMenuHidden).
// Respeta el orden de ocultación: solo se oculta al abrir x panel, no con cualquier overlay.
const menuHidden = computed(() =>
    !!heavyPanel.value
    || accountsOpen.value
    || versionsOpen.value
    || newsOpen.value
    || welcomeOpen.value
    || previewOpen.value
);

// Al cerrar todos los overlays (idle, etc.) el widget entero se oculta,
// incluido el botón de abrir; reaparece con la próxima interacción.
const hiddenByOverlay = ref(false);
let lastActivity = 0;

watch(menuHidden, (hidden) => {
    if (hidden) {
        expanded.value = false;
        volOpen.value = false;
        queueOpen.value = false;
    }
});

function onCloseOverlays() {
    hiddenByOverlay.value = true;
    expanded.value = false;
    volOpen.value = false;
    queueOpen.value = false;
}

function onActivity() {
    const now = Date.now();
    if (now - lastActivity < 1500) return;
    lastActivity = now;
    if (hiddenByOverlay.value) hiddenByOverlay.value = false;
}

onMounted(() => {
    void loadMusicList();
    // loadLocalLibrary ya NO se llama aquí: es carga perezosa del panel de Música.
    // El widget reacciona automáticamente cuando el panel puebla localTracks.
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    window.addEventListener('mousemove', onActivity, { passive: true });
    window.addEventListener('mousedown', onActivity, { passive: true });
    window.addEventListener('wheel', onActivity, { passive: true });
    window.addEventListener('mousedown', onWindowDown);
    window.addEventListener('keydown', onEscapeKey);
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    window.removeEventListener('mousemove', onActivity);
    window.removeEventListener('mousedown', onActivity);
    window.removeEventListener('wheel', onActivity);
    window.removeEventListener('mousedown', onWindowDown);
    window.removeEventListener('keydown', onEscapeKey);
});
</script>

<template>
    <Teleport to="body">
        <div v-show="show && !hiddenByOverlay" class="Bgm" :class="{ 'Bgm--menuHidden': menuHidden }">

            <Transition name="bgm-up" mode="out-in">
                <div v-if="!expanded" class="Bgm_Fab" :class="{ live: playing }" title="Música de fondo" @click="toggleExpand">
                    <IconMusic stroke="2" />
                    <span class="Bgm_FabEq"><i v-for="n in 3" :key="n"></i></span>
                </div>

                <div v-else class="Bgm_Card" :class="{ 'Bgm_Card--bg': isBg && !isLauncher }">
                    <div v-if="isBg && !isLauncher" class="Bgm_Bg">
                        <img v-if="meta?.coverUrl" class="Bgm_BgImage" :src="meta.coverUrl" alt="" loading="lazy" />
                    </div>
                    <div v-if="isLauncher && launcherTrack?.coverUrl" class="Bgm_Bg">
                        <img class="Bgm_BgImage" :src="launcherTrack.coverUrl" alt="" loading="lazy" />
                    </div>

                    <button class="Bgm_Close" title="Ocultar widget" @click="toggleExpand">
                        <IconX stroke="2" />
                    </button>

                    <div class="Bgm_Source">
                        <button class="Bgm_SourceBtn" :class="{ active: !isLauncher }" @click="setSource('background')" title="Música de fondo">
                            <IconDisc stroke="2" /> Fondo
                        </button>
                        <button class="Bgm_SourceBtn" :class="{ active: isLauncher }" @click="setSource('launcher')" title="Tu biblioteca">
                            <IconLibrary stroke="2" /> Biblioteca
                        </button>
                    </div>

                    <div
                        v-if="!isBg || isLauncher"
                        class="Bgm_Cover"
                        :class="{
                            'Bgm_Disc': isDisc && !isLauncher,
                            'Bgm_Square': !isDisc || isLauncher,
                            spinning: spinEnabled && !isLauncher,
                            empty: isLauncher ? !displayCoverUrl : !meta?.coverUrl,
                        }"
                    >
                        <img v-if="isLauncher ? displayCoverUrl : meta?.coverUrl" :src="isLauncher ? displayCoverUrl : meta?.coverUrl || ''" alt="" />
                        <IconMusic v-else stroke="1.5" />
                        <span v-if="isDisc && !isLauncher" class="Bgm_Hole"></span>
                    </div>

                    <div v-else class="Bgm_Cover Bgm_Cover--bg" :class="{ empty: !meta?.coverUrl }">
                        <img v-if="meta?.coverUrl" :src="meta.coverUrl" alt="" />
                        <IconMusic v-else stroke="1.5" />
                    </div>

                    <div class="Bgm_Body">
                        <div class="Bgm_Head">
                            <div class="Bgm_Info">
                                <span class="Bgm_Title" :title="displayTitle">{{ displayTitle }}</span>
                                <span class="Bgm_Artist" :title="displayArtist">{{ displayArtist }}</span>
                            </div>
                            <div class="Bgm_Eq" :class="{ live: playing }">
                                <i v-for="n in 4" :key="n"></i>
                            </div>
                        </div>

                        <div v-if="loadError" class="Bgm_Error" :title="loadError">
                            <span>{{ loadError }}</span>
                        </div>

                        <div class="Bgm_SeekRow">
                            <span class="Bgm_Time">{{ formatDuration(currentTime) }}</span>
                            <input
                                class="Bgm_Seek"
                                type="range"
                                min="0"
                                :max="duration || 0"
                                step="1"
                                :value="currentTime"
                                :disabled="!duration"
                                @input="onSeek"
                            />
                            <span class="Bgm_Time">{{ formatDuration(duration) }}</span>
                        </div>

                        <div class="Bgm_Controls">
                            <div class="Bgm_Transport">
                                <button v-if="multi" class="Bgm_Btn" title="Anterior" @click="onPrev">
                                    <IconPlayerSkipBack stroke="2" />
                                </button>
                                <button class="Bgm_Btn Bgm_BtnPlay" :title="playing ? 'Pausar' : 'Reproducir'" @click="onTogglePlay">
                                    <IconPlayerPause v-if="playing" stroke="2" />
                                    <IconPlayerPlay v-else stroke="2" />
                                </button>
                                <button v-if="multi" class="Bgm_Btn" title="Siguiente" @click="onNext">
                                    <IconPlayerSkipForward stroke="2" />
                                </button>
                            </div>
                            <div class="Bgm_Right">
                                <div class="Bgm_Modes">
                                    <button class="Bgm_Mode" :class="{ active: isShuffle }" title="Aleatorio" @click="onToggleShuffle">
                                        <IconArrowsShuffle stroke="2" />
                                    </button>
                                    <button class="Bgm_Mode" :class="{ active: isRepeatOne }" title="Bucle" @click="onToggleRepeat">
                                        <IconRepeatOnce v-if="isRepeatOne" stroke="2" />
                                        <IconRepeat v-else stroke="2" />
                                    </button>
                                </div>
                                <span class="Bgm_Divider"></span>
                                <button v-if="isLauncher" class="Bgm_QueueBtn" :class="{ active: queueOpen }" title="Ver cola" @click="toggleQueue">
                                    <IconPlaylist stroke="2" />
                                </button>
                                <div class="Bgm_Volume">
                                    <button
                                        ref="volBtn"
                                        class="Bgm_VolBtn"
                                        :class="{ active: volOpen }"
                                        :title="`Volumen: ${volLevel}%`"
                                        @click="toggleVolumeMenu"
                                    >
                                        <IconVolumeOff v-if="volLevel === 0" stroke="2" />
                                        <IconVolume v-else stroke="2" />
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </Transition>

            <Transition name="bgm-vol">
                <div
                    v-if="volOpen"
                    class="Bgm_VolMenu"
                    :style="{
                        '--vol-right': volMenuRight,
                        '--vol-bottom': volMenuBottom,
                    }"
                >
                    <IconVolumeOff v-if="volLevel === 0" class="Bgm_VolIcon" stroke="2" />
                    <IconVolume v-else class="Bgm_VolIcon" stroke="2" />
                    <input
                        class="Bgm_VolRange"
                        type="range"
                        min="0"
                        max="100"
                        step="1"
                        :value="volLevel"
                        title="Volumen"
                        @input="onVolume"
                    />
                    <span class="Bgm_VolVal">{{ volLevel }}%</span>
                </div>
            </Transition>

            <Transition name="bgm-vol">
                <div v-if="queueOpen" class="Bgm_Queue" :style="{ '--vol-right': volMenuRight, '--vol-bottom': volMenuBottom }">
                    <div class="Bgm_QueueHead">
                        <span><IconPlaylist stroke="2" /> Cola</span>
                        <span class="Bgm_QueueCount">{{ isLauncher ? launcherQueueTracks.length : musicList.length }} pistas</span>
                    </div>
                    <div class="Bgm_QueueList">
                        <template v-if="isLauncher">
                            <div v-for="(t, i) in launcherQueueTracks.slice(0, 10)" :key="t.path" class="Bgm_QueueRow" :class="{ active: launcherTrack?.path === t.path }" style="cursor:pointer;" @click="onQueueRowClick(t)">
                                <span class="Bgm_QueueNum">{{ i + 1 }}</span>
                                <span class="Bgm_QueueCover"><img v-if="t.coverUrl" :src="t.coverUrl" alt="" loading="lazy" /><IconMusic v-else stroke="1.5" /></span>
                                <span class="Bgm_QueueInfo"><b>{{ t.title }}</b><em>{{ t.artist }}</em></span>
                                <button class="Bgm_QueuePlay" title="Reproducir" @click.stop="onQueueRowClick(t)"><IconPlayerPlay :size="12" stroke="2" /></button>
                            </div>
                            <p v-if="!launcherQueueTracks.length" class="Bgm_QueueEmpty">Cola vacía. Reproduce desde tu biblioteca.</p>
                        </template>
                        <template v-else>
                            <div v-for="(t, i) in musicList.slice(0, 10)" :key="t.path" class="Bgm_QueueRow">
                                <span class="Bgm_QueueNum">{{ i + 1 }}</span>
                                <span class="Bgm_QueueCover"><img v-if="currentMeta()?.coverUrl && i===0" :src="currentMeta()?.coverUrl || ''" alt="" /><IconMusic v-else stroke="1.5" /></span>
                                <span class="Bgm_QueueInfo"><b>{{ t.name }}</b><em>{{ t.path }}</em></span>
                            </div>
                            <p v-if="!musicList.length" class="Bgm_QueueEmpty">Sin pistas de fondo. Añade en Ajustes → Personalización.</p>
                        </template>
                    </div>
                </div>
            </Transition>
        </div>
    </Teleport>
</template>

<style scoped lang="scss">
@use './Styles/BackgroundMusic.scss';

.Bgm {
    transition: opacity 400ms ease, transform 400ms ease, filter 400ms ease, visibility 0s linear 400ms;
}

.Bgm.Bgm--menuHidden {
    opacity: 0;
    visibility: hidden;
    transform: scale(.92) translateY(-12px);
    filter: blur(8px);
    pointer-events: none;
    transition: opacity 400ms ease, transform 400ms ease, filter 400ms ease, visibility 0s linear 400ms;
}

@media (max-width: 380px) {
    .Bgm_Card { width: calc(100vw - 1rem); padding: 0.7rem 0.8rem; }
}
</style>