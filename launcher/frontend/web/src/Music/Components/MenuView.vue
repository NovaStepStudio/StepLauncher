<script setup lang="ts">
import { computed, onMounted, watch, ref } from 'vue';
import { IconFolderPlus, IconPlayerPlay, IconLibrary, IconDisc, IconClock, IconMusic, IconHeart, IconFolderOpen, IconSparkles, IconHistory, IconChevronDown, IconChevronUp, IconArrowsMaximize, IconArrowsMinimize } from '@tabler/icons-vue';
import { localTracks, totalTracks, isScanning } from '../LocalStore';
import { playlists } from '../PlaylistStore';
import { currentTrack, playTrack } from '../PlayerStore';
import { musicHistory, loadHistory } from '../HistoryStore';
import { useCoverPalette, paletteToCss } from '@/Common/Composables/useCoverPalette';

const heroCollapsed = ref(false);

const props = defineProps<{
    coverUrl: string;
    coverBgStyle: Record<string, string>;
}>();

const emit = defineEmits<{
    (e: 'open-biblioteca'): void;
    (e: 'open-ahora'): void;
}>();

const { palette } = useCoverPalette(() => props.coverUrl);
const bgStyle = computed(() => paletteToCss(palette.value, 0.22));

const totalHits = computed(() => totalTracks.value || localTracks.value.length);
const totalPlaylists = computed(() => playlists.value.length);
const historyCount = computed(() => musicHistory.value.length);

// Recientemente añadidas: últimas 6 del índice (no solo página 20)
const recentRaw = computed(() => [...localTracks.value].slice(-6).reverse());
const volverRaw = computed(() => {
    if (musicHistory.value.length) return musicHistory.value.slice(0, 6).map((h) => ({
        path: h.path, title: h.title, artist: h.artist, coverUrl: h.coverUrl || '', fileName: h.path.split(/[\\/]/).pop() || h.title, duration: 0,
    })) as any[];
    return localTracks.value.slice(0, 6);
});

const recentCoverMap = ref<Map<string, string>>(new Map());
const volverCoverMap = ref<Map<string, string>>(new Map());

async function fetchCoversFor(paths: string[], map: Map<string,string>): Promise<void> {
    const need = paths.filter((p) => p && !map.has(p));
    if (!need.length) return;
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.GetMusicTracksBatch === 'function') {
            const batch: any = await mod.GetMusicTracksBatch(need, 'thumb');
            if (Array.isArray(batch)) batch.forEach((t: any) => { const c = t.coverThumb || t.coverRaw; if (c) map.set(t.path, c); });
        }
    } catch {}
}

watch(recentRaw, (list) => {
    if (list.length) void fetchCoversFor(list.map((t:any)=>t.path), recentCoverMap.value);
}, { immediate: true });
watch(volverRaw, (list) => {
    if (list.length) void fetchCoversFor(list.map((t:any)=>t.path), volverCoverMap.value);
}, { immediate: true });

const recent = computed(() => recentRaw.value.map((t:any) => ({ ...t, coverUrl: recentCoverMap.value.get(t.path) || t.coverUrl || '' })));
const volver = computed(() => volverRaw.value.map((t:any) => ({ ...t, coverUrl: volverCoverMap.value.get(t.path) || t.coverUrl || '' })));

onMounted(() => { void loadHistory(); });

async function playRecent(t: any): Promise<void> {
    const list: any[] = recent.value as any;
    const found = list.find((x) => x.path === t.path) || t;
    // Cola = las 6 recientes, no toda la biblioteca paginada (respeta lo que ves)
    if (found?.path) await playTrack({ id: found.path, path: found.path, fileName: found.fileName || found.title, title: found.title, artist: found.artist || 'Desconocido', duration: found.duration || 0, coverUrl: found.coverUrl || '' } as any, list as any);
}

async function playHistory(t: any): Promise<void> {
    const list: any[] = volver.value as any;
    const found = list.find((x) => x.path === t.path) || t;
    if (found?.path) await playTrack({ id: found.path, path: found.path, fileName: found.fileName || found.title, title: found.title, artist: found.artist || 'Desconocido', duration: found.duration || 0, coverUrl: found.coverUrl || '' } as any, list as any);
}
</script>

<template>
    <div class="MusicSection">
        <div class="MusicHero" :class="{ 'is-collapsed': heroCollapsed }" :style="{ ...bgStyle, ...coverBgStyle }">
            <img v-if="coverUrl && !heroCollapsed" class="MusicHero_Bg" :src="coverUrl" alt="" />
            <div v-if="!heroCollapsed" class="MusicHero_Glow" :style="palette ? { background: `radial-gradient(600px 200px at 20% 0%, ${palette.vibrant}18, transparent 60%)` } : {}" />
            <button class="MusicHero_Toggle" :title="heroCollapsed ? 'Expandir hero' : 'Contraer hero – más espacio para música'" @click="heroCollapsed = !heroCollapsed">
                <IconChevronUp v-if="!heroCollapsed" :size="14" stroke="2" />
                <IconChevronDown v-else :size="14" stroke="2" />
            </button>
            <template v-if="!heroCollapsed">
            <div class="MusicHero_Left">
                <span class="MusicHero_Kicker"><IconSparkles :size="12" stroke="2" /> Tu biblioteca</span>
                <h2>Tu música, a tu ritmo</h2>
                <p>Toda tu música en un solo lugar. Gestiona tu colección y crea listas de reproducción seleccionando lo que ya tienes.</p>
                <div class="MusicHero_Actions">
                    <button class="SsBtn SsBtnPrimary" @click="emit('open-biblioteca')"><IconFolderPlus :size="14" stroke="2" /> Abrir biblioteca</button>
                    <button class="SsBtn MusicGhostBtn" @click="emit('open-ahora')"><IconPlayerPlay :size="14" stroke="2" /> Ahora suena</button>
                </div>
                <div v-if="isScanning" class="MusicHero_ScanHint">Cargando… {{ totalHits }} pistas encontradas</div>
            </div>
            <div class="MusicHero_Card">
                <div class="MusicHero_Cover"><img v-if="coverUrl" :src="coverUrl" alt="" /><IconMusic v-else :size="20" stroke="1.5" /></div>
                <div class="MusicHero_Info">
                    <span class="MusicHero_Label">{{ currentTrack ? 'Reproduciendo ahora' : 'Sin reproducción' }}</span>
                    <b class="MusicHero_Title">{{ currentTrack?.title ?? 'Nada en reproducción' }}</b>
                    <span class="MusicHero_Artist">{{ currentTrack?.artist ?? 'Selecciona una pista' }}</span>
                </div>
            </div>
            </template>
            <template v-else>
            <div class="MusicHero_CollapsedInfo">
                <span class="MusicHero_Cover mini"><img v-if="coverUrl" :src="coverUrl" alt="" /><IconMusic v-else :size="14" stroke="1.5" /></span>
                <b>Tu música</b>
                <span>{{ totalHits }} pistas · {{ totalPlaylists }} listas de reproducción</span>
                <button class="SsBtn SsBtnPrimary small" @click="emit('open-biblioteca')"><IconLibrary :size="12" stroke="2" /> Biblioteca</button>
            </div>
            </template>
        </div>

        <div class="MusicStatsGrid">
            <div class="MusicStat"><span class="MusicStat_Icon"><IconLibrary :size="16" stroke="2" /></span><div><b>{{ totalHits }}</b><span>Pistas</span></div></div>
            <div class="MusicStat"><span class="MusicStat_Icon"><IconDisc :size="16" stroke="2" /></span><div><b>{{ totalPlaylists }}</b><span>Listas de reproducción</span></div></div>
            <div class="MusicStat"><span class="MusicStat_Icon"><IconClock :size="16" stroke="2" /></span><div><b>{{ historyCount }}</b><span>Escuchadas</span></div></div>
            <div class="MusicStat"><span class="MusicStat_Icon"><IconHistory :size="16" stroke="2" /></span><div><b>{{ totalHits }}</b><span>En biblioteca</span></div></div>
        </div>

        <div v-if="!totalHits" class="MusicEmptyHero">
            <span class="MusicEmptyHero_Icon"><IconFolderOpen :size="20" stroke="1.5" /></span>
            <b>Empieza por tu carpeta de música</b>
            <p>Ve a <b>Ajustes → Música</b> y elige tu carpeta. La música se cargará automáticamente.</p>
        </div>

        <div class="MusicRecentGrid">
            <section class="MusicRecent">
                <h4><IconClock :size="14" stroke="2" /> Recientemente añadidas</h4>
                <div class="MusicRows small">
                    <div v-for="t in recent" :key="t.path" class="MusicRow small" @click="playRecent(t)">
                        <span class="MusicRow_Cover small" :class="{ 'is-loading': (t as any).hasCover && !(t as any).coverUrl }"><img v-if="(t as any).coverUrl" :src="(t as any).coverUrl" alt="" loading="lazy" /><span v-if="(t as any).hasCover && !(t as any).coverUrl" class="MusicRow_Shimmer" style="position:absolute; inset:0;"></span><IconMusic v-if="!(t as any).coverUrl" :size="14" stroke="1.5" style="position:relative; z-index:1;" /></span>
                        <span class="MusicRow_Info"><span class="MusicRow_Title">{{ t.title }}</span><span class="MusicRow_Sub">{{ t.artist }}</span></span>
                        <span class="MusicRow_Dur">{{ t.duration ? `${Math.floor(t.duration/60)}:${String(Math.floor(t.duration%60)).padStart(2,'0')}` : '—' }}</span>
                        <button class="MusicRow_Play small" title="Reproducir" @click.stop="playRecent(t)"><IconPlayerPlay :size="12" stroke="2" /></button>
                    </div>
                    <p v-if="!recent.length" class="MusicMiniEmpty">Nada por aquí aún — escanea tu carpeta.</p>
                </div>
            </section>
            <section class="MusicRecent">
                <h4><IconHistory :size="14" stroke="2" /> Volver a escuchar</h4>
                <div class="MusicRows small">
                    <div v-for="t in volver" :key="t.path" class="MusicRow small" @click="playHistory(t)">
                        <span class="MusicRow_Cover small" :class="{ 'is-loading': (t as any).hasCover && !(t as any).coverUrl }"><img v-if="(t as any).coverUrl" :src="(t as any).coverUrl" alt="" loading="lazy" /><span v-if="(t as any).hasCover && !(t as any).coverUrl" class="MusicRow_Shimmer" style="position:absolute; inset:0;"></span><IconDisc v-if="!(t as any).coverUrl" :size="14" stroke="1.5" style="position:relative; z-index:1;" /></span>
                        <span class="MusicRow_Info"><span class="MusicRow_Title">{{ (t as any).title }}</span><span class="MusicRow_Sub">{{ (t as any).artist }}</span></span>
                        <button class="MusicRow_Play small" title="Reproducir" @click.stop="playHistory(t)"><IconPlayerPlay :size="12" stroke="2" /></button>
                    </div>
                    <p v-if="!volver.length" class="MusicMiniEmpty">Escucha algo y aparecerá aquí. Se guarda en launcher_music_history.json</p>
                </div>
            </section>
        </div>
    </div>
</template>

<style scoped lang="scss">
@use '../Styles/MenuView.scss';
</style>
