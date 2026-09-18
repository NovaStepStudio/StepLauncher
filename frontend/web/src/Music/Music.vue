<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { IconMusic, IconX, IconHome, IconLibrary, IconDisc, IconPlaylist, IconFolderPlus, IconLayoutSidebarLeftCollapse, IconLayoutSidebarLeftExpand } from '@tabler/icons-vue';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';
import { heavyPanel, closeHeavyPanel } from '@/Common/Overlays/Store';
import { localTracks, totalTracks, loadLocalLibrary, isScanning, scanProgress, ensureCovers, ensureMusicStoreInitialized } from './LocalStore';
import { playlists, loadPlaylists, createPlaylist, updatePlaylist } from './PlaylistStore';
import { currentTrack, playing, currentTime, duration, progressPct, volume, playMode, shuffleMode, repeatMode, queueTracks, togglePlay, next, prev, nextNext, prevPrev, seekForward10, seekBackward10, toggleShuffleMode, toggleRepeatMode, setVolume, setPlayMode, seekTo, ensurePlayerStoreInitialized } from './PlayerStore';
import MenuView from './Components/MenuView.vue';
import MusicView from './Components/MusicView.vue';
import LibraryView from './Components/LibraryView.vue';
import NowPlayingView from './Components/NowPlayingView.vue';
import QueueView from './Components/QueueView.vue';
import PlayerBar from './Components/PlayerBar.vue';
import PlaylistForm from './Components/PlaylistForm.vue';
import { useCoverPalette, paletteToCss, getPaletteColor } from '@/Common/Composables/useCoverPalette';

type Section = 'menu' | 'musica' | 'biblioteca' | 'ahora' | 'cola';
const section = ref<Section>('menu');
const showPlaylistForm = ref(false);
const menuCollapsed = ref(false);

const fetchedDisplayRaw = ref('');
const fetchedPlayerThumb = ref('');
const panelColorMode = ref('vibrant');
const panelCoverOpacity = ref(1); // 0..1 — opacidad del ::after con --panel-cover-color (controlado en Ajustes → Música)
watch(() => currentTrack.value?.path, async (p) => {
    fetchedDisplayRaw.value = '';
    fetchedPlayerThumb.value = '';
    if (!p) return;
    const cur: any = currentTrack.value as any;
    // Si no tiene carátula, no pedir
    const hasCover = cur?.hasCover !== false;
    if (!hasCover && !cur?.coverUrl && !cur?.coverRaw) { /* sin carátula, no pedir */ }
    else {
        try {
            const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
            if (typeof mod.GetMusicTracksBatch === 'function') {
                const batch: any = await mod.GetMusicTracksBatch([p], 'raw');
                if (Array.isArray(batch) && batch[0]?.coverRaw) { fetchedDisplayRaw.value = batch[0].coverRaw; }
                else if (Array.isArray(batch) && batch[0]?.coverThumb) { fetchedDisplayRaw.value = batch[0].coverThumb; }
            }
            if (!fetchedDisplayRaw.value && typeof mod.GetMusicCoverBase64 === 'function') {
                const raw: string = await mod.GetMusicCoverBase64(p, 'raw').catch(() => '');
                if (raw) fetchedDisplayRaw.value = raw;
            }
        } catch (_e) {}
    }
    // Thumb baja para player (10KB) – siempre thumb, no raw
    if (cur?.coverUrl) {
        fetchedPlayerThumb.value = cur.coverUrl;
    } else if (hasCover) {
        try {
            const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
            if (typeof mod.GetMusicTracksBatch === 'function') {
                const batch: any = await mod.GetMusicTracksBatch([p], 'thumb');
                if (Array.isArray(batch) && (batch[0]?.coverThumb || batch[0]?.coverRaw)) fetchedPlayerThumb.value = batch[0].coverThumb || batch[0].coverRaw;
            }
            if (!fetchedPlayerThumb.value && typeof mod.GetMusicCoverBase64 === 'function') {
                const thumb: string = await mod.GetMusicCoverBase64(p, 'thumb').catch(() => '');
                if (thumb) fetchedPlayerThumb.value = thumb;
            }
        } catch {}
    }
}, { immediate: true });

// CARÁTULA hero SIEMPRE alta resolución (raw original), NUNCA thumb 128 baja
// Fondo puede ser thumb, pero carátula nunca
const displayCover = computed(() => {
    if (fetchedDisplayRaw.value) return fetchedDisplayRaw.value;
    const cur: any = currentTrack.value as any;
    if (cur?.coverRaw && String(cur.coverRaw).startsWith('data:')) return cur.coverRaw;
    return '';
});
const displayCoverForHero = computed(() => displayCover.value);
const displayCoverForBg = computed(() => {
    // Fondo borroso puede usar thumb baja para no gastar RAM, pero si no hay thumb usa raw
    const cur: any = currentTrack.value as any;
    return cur?.coverUrl || displayCover.value || '';
});
const { palette } = useCoverPalette(() => displayCover.value);
const coverBgStyle = computed(() => paletteToCss(palette.value, 0.22));
const totalDisplay = computed(() => totalTracks.value || localTracks.value.length);
// Panel completo: BG #0005 + BG:after con color predominante un poquito negro (como PlayerBar)
// La opacidad del ::after se controla con --panel-cover-opacity (0 transparente .. 1 opaco) desde Ajustes → Música
const musicPanelBg = computed(() => {
    if (!currentTrack.value || !palette.value) {
        // Sin pista: mantener opacidad configurada por si se cambia antes de reproducir
        return { '--panel-cover-opacity': String(Math.max(0, Math.min(1, panelCoverOpacity.value))) } as Record<string, string>;
    }
    const raw = getPaletteColor(palette.value, panelColorMode.value);
    const col = `color-mix(in srgb, ${raw} 88%, black 12%)`;
    return {
        '--panel-cover-color': col,
        '--panel-cover-opacity': String(Math.max(0, Math.min(1, panelCoverOpacity.value))),
    } as Record<string, string>;
});
const playerCoverUrl = computed(() => {
    const cur: any = currentTrack.value as any;
    if (cur?.coverUrl) return cur.coverUrl;
    if (fetchedPlayerThumb.value) return fetchedPlayerThumb.value;
    return '';
});
const playerCoverColor = computed(() => {
    if (!currentTrack.value || !palette.value) return '';
    return getPaletteColor(palette.value, panelColorMode.value);
});

function close(): void { closeHeavyPanel('music'); }
useOverlayEscape(close, { isActive: () => heavyPanel.value === 'music' });

async function eagerLoad(): Promise<void> {
    // Desactivado para listas: 0 MB. Solo NowPlaying carga 1 cover raw bajo demanda
}

// --- Carga perezosa: NO se ejecuta nada hasta que el usuario entra al panel ---
let _musicInitialized = false;
const isPanelLoading = ref(false);
const panelLoadError = ref('');
async function initializeMusicPanel(): Promise<void> {
    if (_musicInitialized) {
        // Reapertura: refrescar por si se cambió la carpeta en Ajustes
        void loadLocalLibrary().then(() => void eagerLoad());
        void loadPlaylists();
        return;
    }
    _musicInitialized = true;
    isPanelLoading.value = true;
    panelLoadError.value = '';
    try { ensureMusicStoreInitialized(); } catch (_e) {}
    try { ensurePlayerStoreInitialized(); } catch (_e) {}
    void loadPlaylists();
    try {
        await loadLocalLibrary().then(() => void eagerLoad());
    } catch (e: any) {
        panelLoadError.value = e?.message ?? 'No se pudo cargar la biblioteca';
    } finally {
        isPanelLoading.value = false;
    }
}

onMounted(async () => {
    if (heavyPanel.value === 'music') void initializeMusicPanel();
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        const cfg: any = await mod.GetMusicPanelConfig().catch(() => null);
        if (cfg?.colorMode) panelColorMode.value = cfg.colorMode;
        const rawOp = (cfg as any)?.coverOpacity;
        if (typeof rawOp === 'number' && Number.isFinite(rawOp)) {
            panelCoverOpacity.value = Math.max(0, Math.min(1, rawOp));
        }
    } catch {}
});
watch(() => heavyPanel.value, (v) => {
    if (v === 'music') void initializeMusicPanel();
});
watch(localTracks, () => { if (_musicInitialized) void eagerLoad(); });
watch(queueTracks, () => { if (_musicInitialized) void eagerLoad(); });
watch(currentTrack, (t) => { if (_musicInitialized && t && !t.coverUrl) void ensureCovers([t as any]); });
if (typeof window !== 'undefined') {
    const reloadPanelConfig = async () => {
        try {
            const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
            const cfg: any = await mod.GetMusicPanelConfig().catch(() => null);
            if (cfg?.colorMode) panelColorMode.value = cfg.colorMode;
            const rawOp = (cfg as any)?.coverOpacity;
            if (typeof rawOp === 'number' && Number.isFinite(rawOp)) {
                panelCoverOpacity.value = Math.max(0, Math.min(1, rawOp));
            }
        } catch {}
    };
    window.addEventListener('stl:music-config-changed', reloadPanelConfig);
    window.addEventListener('stl:music-folder-changed', reloadPanelConfig);
    // Preview instantáneo mientras se arrastra el slider de opacidad en Ajustes → Música
    window.addEventListener('stl:panel-cover-opacity', ((e: CustomEvent<number>) => {
        const v = (e as CustomEvent).detail;
        if (typeof v === 'number' && Number.isFinite(v)) {
            panelCoverOpacity.value = Math.max(0, Math.min(1, v));
        }
    }) as EventListener);
}

const sections = [
    { id: 'menu' as Section, label: 'Inicio', icon: IconHome, desc: 'Resumen' },
    { id: 'musica' as Section, label: 'Música', icon: IconMusic, desc: `${localTracks.value.length} pistas` },
    { id: 'biblioteca' as Section, label: 'Biblioteca', icon: IconLibrary, desc: `${playlists.value.length} listas de reproducción` },
    { id: 'ahora' as Section, label: 'Ahora suena', icon: IconDisc, desc: playing.value ? 'Reproduciendo' : 'En pausa' },
    { id: 'cola' as Section, label: 'Cola de Reproducción', icon: IconPlaylist, desc: 'Siguiente' },
] as const;

const headerSub = computed(() => {
    if (isScanning.value) return 'Escaneando tu biblioteca…';
    const total = totalTracks.value || localTracks.value.length;
    if (!total) return 'Configura tu carpeta en Ajustes → Música';
    if (playing.value && currentTrack.value) return `${currentTrack.value.title} · ${currentTrack.value.artist}`;
    return `${total} pistas en tu biblioteca`;
});

async function handleCreatePlaylist(data: { title: string; favorite: boolean; pinned: boolean; color: string; cover: string; tracks: string[] }): Promise<void> {
    const err = await createPlaylist(data.title, data.tracks);
    if (!err && (data.favorite || data.pinned || data.color || data.cover)) {
        const created = playlists.value[playlists.value.length - 1];
        if (created) {
            await updatePlaylist(created.id, { favorite: data.favorite, pinned: data.pinned, customColor: data.color, customCover: data.cover } as any);
        }
    }
    if (!err) section.value = 'biblioteca';
}
</script>

<template>
    <div class="MusicModal_Overlay" :class="{ 'has-gradient': !!currentTrack }" :style="musicPanelBg">
        <header class="MusicModal_Head">
            <div class="MusicModal_Title">
                <span class="MusicModal_Icon"><IconMusic :size="18" stroke="2" /></span>
                <div class="MusicModal_Titles">
                    <h3>Música</h3>
                    <p :title="headerSub">{{ headerSub }}</p>
                    <span v-if="isScanning" style="display:inline-flex; align-items:center; gap:0.3rem; margin-top:0.2rem; font-size:0.68rem; color:var(--accent);"><span class="spin" style="display:inline-block; width:12px; height:12px; border:2px solid currentColor; border-top-color:transparent; border-radius:50%; animation: spin 0.8s linear infinite;"></span> Escaneando {{ scanProgress.discovered ? `${scanProgress.processed||0}/${scanProgress.discovered}` : `${scanProgress.current||0}` }}…</span>
                </div>
            </div>
            <div class="MusicHead_Actions">
                <button class="SsBtn SsBtnPrimary" title="Nueva lista de reproducción con selector visual" @click="showPlaylistForm = true">
                    <IconFolderPlus :size="14" stroke="2" /> Nueva lista de reproducción
                </button>
                <span class="MusicHead_Chip" :class="{ subtle: !playing }">{{ playing ? '● En reproducción' : '○ En pausa' }}</span>
                <span class="MusicHead_Chip subtle">{{ totalDisplay }} pistas</span>
            </div>
            <button class="MusicModal_Close" title="Cerrar" @click="close"><IconX :size="16" stroke="2" /></button>
        </header>

        <div class="MusicLayout" :class="{ 'is-menu-collapsed': menuCollapsed }">
            <aside class="MusicMenu" :class="{ collapsed: menuCollapsed }">
                <div class="MusicMenu_Toggle">
                    <button class="MusicMenu_CollapseBtn" :class="{ collapsed: menuCollapsed }" :title="menuCollapsed ? 'Mostrar barra lateral' : 'Ocultar barra lateral'" @click="menuCollapsed = !menuCollapsed">
                        <component :is="menuCollapsed ? IconLayoutSidebarLeftExpand : IconLayoutSidebarLeftCollapse" :size="18" stroke="2" />
                        <span v-if="!menuCollapsed" class="MusicMenu_CollapseLabel">Ocultar menú</span>
                    </button>
                </div>
                <nav class="MusicMenu_Nav">
                    <button v-for="s in sections" :key="s.id" class="MusicMenu_Item" :class="{ active: section===s.id }" @click="section=s.id" :title="menuCollapsed ? s.label : undefined">
                        <component :is="s.icon" :size="16" stroke="2" />
                        <span v-if="!menuCollapsed" class="MusicMenu_Txt"><b>{{ s.label }}</b><em>{{ s.id==='musica' ? `${localTracks.length} pistas` : s.id==='biblioteca' ? `${playlists.length} listas` : s.desc }}</em></span>
                    </button>
                </nav>
                <div v-if="!menuCollapsed" class="MusicMenu_Foot">
                    <div class="MusicMenu_Stats">
                        <span><IconMusic :size="12" stroke="2" /> {{ localTracks.length }} pistas</span>
                        <span><IconLibrary :size="12" stroke="2" /> {{ playlists.length }} listas de reproducción</span>
                    </div>
                    <p class="MusicMenu_Hint">Gestiona tu biblioteca y crea listas de reproducción.</p>
                </div>
            </aside>

            <main class="MusicMain">
                <div v-if="isPanelLoading" class="MusicPanel_LoadingOverlay">
                    <span class="MusicPanel_Spinner"></span>
                    <h4>Cargando tu biblioteca musical</h4>
                    <p>Preparando tus pistas y carátulas… Esto puede tardar unos segundos la primera vez.</p>
                    <span v-if="isScanning" class="MusicPanel_ScanDetail">Escaneando {{ scanProgress.discovered ? `${scanProgress.processed || 0}/${scanProgress.discovered}` : `${scanProgress.current || 0}` }} archivos… {{ scanProgress.currentFile ? scanProgress.currentFile.split(/[\\/]/).pop() : '' }}</span>
                    <span v-else class="MusicPanel_ScanDetail">Leyendo índice local…</span>
                    <p v-if="panelLoadError" class="MusicPanel_Error">{{ panelLoadError }}</p>
                </div>
                <template v-else>
                <MenuView v-if="section==='menu'" :cover-url="displayCover" :cover-bg-style="coverBgStyle" @open-biblioteca="section='biblioteca'" @open-ahora="section='ahora'" />
                <MusicView v-else-if="section==='musica'" />
                <LibraryView v-else-if="section==='biblioteca'" @create-playlist="showPlaylistForm = true" />
                <NowPlayingView v-else-if="section==='ahora'" />
                <QueueView v-else />
                </template>
            </main>
        </div>

        <PlayerBar
            :cover-url="playerCoverUrl"
            :cover-color="playerCoverColor"
            :title="currentTrack?.title || ''"
            :artist="currentTrack?.artist || ''"
            :current-time="currentTime"
            :duration="duration || currentTrack?.duration || 0"
            :progress-pct="progressPct"
            :playing="playing"
            :volume="volume"
            :play-mode="playMode"
            :shuffle-mode="shuffleMode"
            :repeat-mode="repeatMode"
            @prev="prev"
            @prev-prev="prevPrev"
            @toggle="togglePlay"
            @next="next"
            @next-next="nextNext"
            @seek="seekTo"
            @seek-backward="seekBackward10"
            @seek-forward="seekForward10"
            @toggle-shuffle="toggleShuffleMode"
            @toggle-repeat="toggleRepeatMode"
            @set-volume="setVolume"
            @set-mode="setPlayMode"
        />

        <PlaylistForm
            :visible="showPlaylistForm"
            @update:visible="showPlaylistForm = $event"
            @submit="handleCreatePlaylist"
        />
    </div>
</template>

<style scoped lang="scss">
@use './Styles/Music.scss';
@use './Styles/Library.scss';

.MusicPanel_LoadingOverlay {
    flex: 1;
    min-height: 320px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.85rem;
    padding: 2.5rem 1.5rem;
    text-align: center;
    background: var(--background-modal-primary);
}

.MusicPanel_Spinner {
    width: 38px;
    height: 38px;
    border: 3px solid color-mix(in srgb, var(--background-button-primary) 20%, transparent);
    border-top-color: var(--background-button-primary);
    border-radius: 50%;
    animation: MusicPanelSpin 0.7s linear infinite;
}

.MusicPanel_LoadingOverlay h4 {
    margin: 0;
    font-size: 0.96rem;
    font-weight: 700;
    color: var(--text-primary);
}

.MusicPanel_LoadingOverlay > p {
    margin: 0;
    max-width: 28rem;
    font-size: 0.78rem;
    line-height: 1.5;
    opacity: 0.62;
    color: var(--text-secondary);
}

.MusicPanel_ScanDetail {
    font-size: 0.72rem;
    opacity: 0.75;
    color: var(--background-button-primary);
    font-weight: 600;
}

.MusicPanel_Error {
    color: var(--color-error) !important;
    opacity: 1 !important;
    font-weight: 600;
}

@keyframes MusicPanelSpin {
    to { transform: rotate(360deg); }
}
</style>
