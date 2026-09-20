<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import {
    IconPlayerPlay, IconPlayerPause, IconPlayerSkipBack, IconPlayerSkipForward,
    IconPlayerTrackPrev, IconPlayerTrackNext,
    IconRewindBackward10, IconRewindForward10,
    IconVolume, IconVolumeOff,
    IconRepeat, IconRepeatOnce, IconArrowsShuffle, IconArrowRightBar, IconMusic
} from '@tabler/icons-vue';

const props = defineProps<{
    coverUrl: string;
    coverColor?: string;
    title: string;
    artist: string;
    currentTime: number;
    duration: number;
    progressPct: number;
    playing: boolean;
    volume: number;
    playMode: string;
    shuffleMode?: 'linear' | 'shuffle';
    repeatMode?: 'queue' | 'one';
}>();

const emit = defineEmits<{
    (e: 'prev'): void;
    (e: 'prev-prev'): void;
    (e: 'toggle'): void;
    (e: 'next'): void;
    (e: 'next-next'): void;
    (e: 'seek', v: number): void;
    (e: 'seek-backward'): void;
    (e: 'seek-forward'): void;
    (e: 'toggle-shuffle'): void;
    (e: 'toggle-repeat'): void;
    (e: 'setVolume', v: number): void;
    (e: 'setMode', m: 'queue' | 'shuffle' | 'repeat-one'): void;
}>();

const volOpen = ref(false);
const volBtn = ref<HTMLButtonElement | null>(null);
const volWrap = ref<HTMLDivElement | null>(null);

function onWindowDown(e: MouseEvent) {
    if (!volOpen.value) return;
    const t = e.target as HTMLElement | null;
    if (t?.closest('.MusicBar_VolWrap') || t?.closest('.MusicBar_VolMenu')) return;
    volOpen.value = false;
}
function onEsc(e: KeyboardEvent) { if (e.key === 'Escape') volOpen.value = false; }
onMounted(() => { window.addEventListener('mousedown', onWindowDown); window.addEventListener('keydown', onEsc); });
onUnmounted(() => { window.removeEventListener('mousedown', onWindowDown); window.removeEventListener('keydown', onEsc); });

function formatDuration(sec: number): string {
    if (!Number.isFinite(sec) || sec <= 0) return '0:00';
    const m = Math.floor(sec / 60), s = Math.floor(sec % 60);
    return `${m}:${String(s).padStart(2,'0')}`;
}

function onSeek(e: Event): void {
    const val = Number((e.target as HTMLInputElement).value);
    emit('seek', val);
}

function toggleVol(): void { volOpen.value = !volOpen.value; }
function onVolInput(e: Event): void { emit('setVolume', Number((e.target as HTMLInputElement).value) / 100); }

function clean(s: string): string {
    let str = String(s ?? '').trim();
    str = str.replace(/[\uFEFF\uFFFD]/g, '');
    str = str.replace(/[\x00-\x1F\x7F-\x9F]/g, '');
    str = str.replace(/[\u200B\u200C\u200D\u2060\u00AD\u034F]/g, '');
    return str.trim();
}

// Compat: si no pasan shuffleMode/repeatMode, derivar de playMode
const isShuffle = () => (props.shuffleMode ?? (props.playMode === 'shuffle' ? 'shuffle' : 'linear')) === 'shuffle';
const isRepeatOne = () => (props.repeatMode ?? (props.playMode === 'repeat-one' ? 'one' : 'queue')) === 'one';
</script>

<template>
    <div class="MusicBar" :style="coverColor ? { '--player-cover-color': coverColor } as any : {}">
        <div class="MusicBar_Left">
            <span class="MusicBar_Cover"><img v-if="coverUrl" :src="coverUrl" alt="" /><IconMusic v-else stroke="1.5" /></span>
            <div class="MusicBar_Info">
                <b :title="clean(title)">{{ clean(title) || 'Sin reproducción' }}</b>
                <span :title="clean(artist)">{{ clean(artist) || 'Selecciona una pista' }}</span>
            </div>
        </div>

        <div class="MusicBar_Center">
            <div class="MusicBar_Progress">
                <span>{{ formatDuration(currentTime) }}</span>
                <div class="MusicBar_Track" @click="(ev: MouseEvent) => { const r = (ev.currentTarget as HTMLElement).getBoundingClientRect(); const pct = (ev.clientX - r.left)/r.width; emit('seek', pct * (duration||0)); }">
                    <i :style="{ width: progressPct + '%' }" />
                </div>
                <span>{{ formatDuration(duration) }}</span>
            </div>
            <div class="MusicBar_Controls is-below">
                <button class="MusicBar_Btn" title="Anterior anterior" @click="emit('prev-prev')"><IconPlayerTrackPrev stroke="2" /></button>
                <button class="MusicBar_Btn" title="Anterior" @click="emit('prev')"><IconPlayerSkipBack stroke="2" /></button>
                <button class="MusicBar_Btn" title="-10s" @click="emit('seek-backward')"><IconRewindBackward10 stroke="2" /></button>
                <button class="MusicBar_Btn is-play" :title="playing ? 'Pausar' : 'Reproducir'" @click="emit('toggle')">
                    <IconPlayerPause v-if="playing" stroke="2" />
                    <IconPlayerPlay v-else stroke="2" />
                </button>
                <button class="MusicBar_Btn" title="+10s" @click="emit('seek-forward')"><IconRewindForward10 stroke="2" /></button>
                <button class="MusicBar_Btn" title="Siguiente" @click="emit('next')"><IconPlayerSkipForward stroke="2" /></button>
                <button class="MusicBar_Btn" title="Siguiente siguiente" @click="emit('next-next')"><IconPlayerTrackNext stroke="2" /></button>
            </div>
            <input class="MusicBar_RangeHidden" type="range" :min="0" :max="duration || 100" :value="currentTime" step="1" @input="onSeek" />
        </div>

        <div class="MusicBar_Right">
            <div class="MusicBar_Modes">
                <button class="MusicBar_Mode" :class="{ on: isShuffle() }" title="Aleatorio" @click="emit('toggle-shuffle')">
                    <IconArrowsShuffle v-if="isShuffle()" stroke="2" />
                    <IconArrowRightBar v-else stroke="2" />
                </button>
                <button class="MusicBar_Mode" :class="{ on: isRepeatOne() }" title="Bucle" @click="emit('toggle-repeat')">
                    <IconRepeatOnce v-if="isRepeatOne()" stroke="2" />
                    <IconRepeat v-else stroke="2" />
                </button>
            </div>
            <span class="MusicBar_Divider" />
            <div ref="volWrap" class="MusicBar_VolWrap">
                <button ref="volBtn" class="MusicBar_Btn" :class="{ active: volOpen }" :title="`Volumen ${Math.round(volume*100)}%`" @click="toggleVol">
                    <IconVolumeOff v-if="volume===0" stroke="2" />
                    <IconVolume v-else stroke="2" />
                </button>
                <Transition name="MusicVol">
                    <div v-if="volOpen" class="MusicBar_VolMenu">
                        <IconVolumeOff v-if="volume===0" class="MusicBar_VolIcon" stroke="2" />
                        <IconVolume v-else class="MusicBar_VolIcon" stroke="2" />
                        <input class="MusicBar_VolRange" type="range" min="0" max="100" :value="Math.round(volume*100)" @input="onVolInput" />
                        <span class="MusicBar_VolVal">{{ Math.round(volume*100) }}%</span>
                    </div>
                </Transition>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
@use '../Styles/PlayerBar.scss';
</style>
