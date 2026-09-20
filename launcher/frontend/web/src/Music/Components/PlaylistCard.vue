<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { IconMusic, IconStar, IconPin, IconPlaylist, IconPlayerPlay, IconPencil } from '@tabler/icons-vue';
import { localTracks } from '../LocalStore';
import type { Playlist } from '../PlaylistStore';
import { GetMusicTracksBatch } from '@wailsjs/StepLauncher/internal/Services/Music/musicservice';
import { GetMusicPanelConfig } from '@wailsjs/StepLauncher/internal/Services/Music/musicservice';
import { useCoverPalette, getPaletteColor } from '@/Common/Composables/useCoverPalette';

const props = defineProps<{
    playlist: Playlist;
}>();

const emit = defineEmits<{
    (e: 'open', id: string): void;
    (e: 'play', id: string): void;
    (e: 'edit', id: string): void;
}>();

const resolvedCustomCover = ref('');

function isDataOrBlob(u: string): boolean {
    return u.startsWith('data:') || u.startsWith('blob:') || u.startsWith('http://') || u.startsWith('https://');
}

async function resolveCustomCover(): Promise<void> {
    const raw = String((props.playlist as any).customCover || '').trim();
    if (!raw) { resolvedCustomCover.value = ''; return; }
    if (isDataOrBlob(raw)) { resolvedCustomCover.value = raw; return; }
    // Es ruta de archivo (absoluta o cache/...): intentar leer vía backend como imagen
    try {
        const { ReadAbsoluteFile } = await import('@wailsjs/StepLauncher/internal/Services/System/systemservice');
        const data: any = await ReadAbsoluteFile(raw).catch(() => null);
        if (!data) { resolvedCustomCover.value = raw; return; }
        let bytes: Uint8Array | null = null;
        if (typeof data === 'string') {
            const bin = atob(data as string);
            bytes = new Uint8Array(bin.length);
            for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
        } else if (data instanceof Uint8Array) bytes = data;
        else if (Array.isArray(data)) bytes = new Uint8Array(data as any);
        if (!bytes || bytes.length === 0) { resolvedCustomCover.value = raw; return; }
        const ext = raw.split('.').pop()?.toLowerCase() ?? 'jpeg';
        const mime = ext === 'png' ? 'image/png' : ext === 'webp' ? 'image/webp' : ext === 'gif' ? 'image/gif' : 'image/jpeg';
        const blob = new Blob([bytes as BlobPart], { type: mime });
        const url = URL.createObjectURL(blob);
        resolvedCustomCover.value = url;
    } catch {
        resolvedCustomCover.value = raw;
    }
}

watch(() => (props.playlist as any).customCover, () => { void resolveCustomCover(); });
onMounted(() => { void resolveCustomCover(); });

const coverTracks = ref<(any | null)[]>([null, null, null, null]);
const hasAnyCover = computed(() => coverTracks.value.some((t) => !!t?.coverUrl));
const hasCustom = computed(() => !!resolvedCustomCover.value || !!(props.playlist as any).customCover);
const isLoadingCover = ref(false);

// Config para colores y estilos
const colorMode = ref('vibrant');
const coverStyle = ref('square');
onMounted(async () => {
    try {
        const cfg: any = await GetMusicPanelConfig();
        colorMode.value = cfg?.colorMode || 'vibrant';
        coverStyle.value = cfg?.coverStyle || 'square';
    } catch (_e) {}
});
if (typeof window !== 'undefined') {
    window.addEventListener('stl:music-config-changed', async () => {
        try {
            const cfg: any = await GetMusicPanelConfig();
            colorMode.value = cfg?.colorMode || 'vibrant';
            coverStyle.value = cfg?.coverStyle || 'square';
        } catch (_e) {}
    });
}

// Carga carátulas para las 4 pistas: usa previewCovers cacheado en Go (instantáneo) si existe, si no pide batch
async function loadCoverTracks() {
    const preview: string[] = (props.playlist as any).previewCovers || [];
    if (preview.length) {
        // previewCovers ya son thumbs 128 data URI cacheados en playlist JSON – lectura instantánea sin I/O
        const out: (any | null)[] = [];
        for (let i = 0; i < 4; i++) {
            const url = preview[i];
            if (url) out.push({ coverUrl: url, path: `preview-${i}` });
            else out.push(null);
        }
        coverTracks.value = out;
        isLoadingCover.value = false;
        return;
    }
    const paths: string[] = (props.playlist as any).trackPaths || [];
    const out: (any | null)[] = [];
    isLoadingCover.value = true;
    try {
        const batchPaths = paths.slice(0, 4).filter(Boolean) as string[];
        if (batchPaths.length) {
            try {
                const batch: any = await GetMusicTracksBatch(batchPaths, 'thumb');
                if (Array.isArray(batch)) {
                    const map = new Map<string, any>();
                    batch.forEach((t: any) => map.set(t.path, t));
                    for (let i = 0; i < 4; i++) {
                        const p = paths[i];
                        if (!p) { out.push(null); continue; }
                        const t = map.get(p);
                        if (t) {
                            out.push({ path: t.path, title: t.title, artist: t.artist, coverUrl: t.coverThumb || t.coverRaw || '', fileName: t.fileName });
                        } else {
                            const local = localTracks.value.find((x) => x.path === p);
                            out.push(local ? { ...local } : { path: p, title: p.split(/[\\/]/).pop() || p, artist: 'No encontrada', coverUrl: '' });
                        }
                    }
                } else {
                    throw new Error('no batch');
                }
            } catch {
                for (let i = 0; i < 4; i++) {
                    const p = paths[i];
                    if (!p) { out.push(null); continue; }
                    const local = localTracks.value.find((x) => x.path === p);
                    out.push(local ? { ...local } : null);
                }
            }
        } else {
            for (let i = 0; i < 4; i++) out.push(null);
        }
    } finally {
        isLoadingCover.value = false;
    }
    coverTracks.value = out;
}
watch(() => (props.playlist as any).previewCovers, () => { void loadCoverTracks(); }, { deep: true });
watch(() => (props.playlist as any).trackPaths, () => { void loadCoverTracks(); }, { immediate: true, deep: true });
watch(() => (props.playlist as any).id, () => { void loadCoverTracks(); });
onMounted(() => { void loadCoverTracks(); });

// Paleta para color predominante de la card
const firstCoverUrl = computed(() => coverTracks.value.find((t) => t?.coverUrl)?.coverUrl || '');
const { palette } = useCoverPalette(() => firstCoverUrl.value);
const cardColor = computed(() => {
    const custom = (props.playlist as any).customColor as string;
    if (custom && custom.trim()) return custom;
    return getPaletteColor(palette.value, colorMode.value);
});

const displayTitle = computed(() => props.playlist.title || 'Sin título');
const count = computed(() => ((props.playlist as any).trackPaths?.length ?? 0));
const fav = computed(() => !!(props.playlist as any).favorite);
const pinned = computed(() => !!(props.playlist as any).pinned);

function onOpen(): void { emit('open', props.playlist.id); }
function onPlay(e: MouseEvent): void { e.stopPropagation(); emit('play', props.playlist.id); }
function onEdit(e: MouseEvent): void { e.stopPropagation(); emit('edit', props.playlist.id); }
</script>

<template>
    <div class="PlCard" :class="{ fav: fav, pinned: pinned }" @click="onOpen" role="button" tabindex="0" @keydown.enter="onOpen" @keydown.space.prevent="onOpen" :style="cardColor ? { '--card-accent': cardColor } as any : {}">
        <div class="PlCard_CoverWrap">
            <!-- Custom cover prioritaria: ocupa todo el cuadro -->
            <div v-if="hasCustom" class="PlCard_Cover is-custom" :class="{ 'is-disc': coverStyle==='disc' }">
                <img v-if="resolvedCustomCover" :src="resolvedCustomCover" alt="" loading="lazy" :class="{ 'is-disc-img': coverStyle==='disc' }" />
                <div v-else-if="(playlist as any).customCover" class="PlCard_CoverPlaceholder">
                    <img :src="(playlist as any).customCover" alt="" loading="lazy" @error="resolvedCustomCover = ''" />
                </div>
                <IconPlaylist v-else :size="20" stroke="1.4" style="min-width:20px; min-height:20px;" />
                <button class="PlCard_Play" title="Reproducir lista de reproducción" @click="onPlay"><IconPlayerPlay :size="14" stroke="2" /></button>
            </div>
            <!-- Collage 2x2 con las primeras 4 carátulas -->
            <div v-else class="PlCard_Cover is-grid" :class="{ empty: !hasAnyCover, 'is-disc-grid': coverStyle==='disc' }">
                <template v-if="isLoadingCover">
                    <span v-for="i in 4" :key="i" class="PlCard_Cell is-loading"><span class="PlCard_Shimmer"></span></span>
                </template>
                <template v-else-if="hasAnyCover">
                    <span v-for="(t,i) in coverTracks" :key="i" class="PlCard_Cell" :class="{ 'is-disc-cell': coverStyle==='disc' }">
                        <img v-if="t?.coverUrl" :src="t.coverUrl" alt="" loading="lazy" :class="{ 'is-disc-img': coverStyle==='disc' }" />
                        <IconMusic v-else :size="14" stroke="1.2" style="min-width:14px; min-height:14px;" />
                    </span>
                </template>
                <template v-else>
                    <span class="PlCard_Cell is-empty"><IconMusic :size="18" stroke="1.3" style="min-width:18px; min-height:18px;" /></span>
                    <span class="PlCard_Cell is-empty"><IconMusic :size="18" stroke="1.3" style="min-width:18px; min-height:18px;" /></span>
                    <span class="PlCard_Cell is-empty"><IconMusic :size="18" stroke="1.3" style="min-width:18px; min-height:18px;" /></span>
                    <span class="PlCard_Cell is-empty"><IconMusic :size="18" stroke="1.3" style="min-width:18px; min-height:18px;" /></span>
                </template>
                <button class="PlCard_Play" title="Reproducir lista de reproducción" @click="onPlay"><IconPlayerPlay :size="14" stroke="2" /></button>
            </div>
            <span v-if="fav" class="PlCard_Badge is-fav" title="Favorita"><IconStar :size="10" stroke="2" /></span>
            <span v-if="pinned" class="PlCard_Badge is-pin" title="Anclada"><IconPin :size="10" stroke="2" /></span>
        </div>
        <div class="PlCard_Info">
            <b class="PlCard_Title" :title="displayTitle">{{ displayTitle }}</b>
            <span class="PlCard_Sub">{{ count }} pista{{ count===1 ? '' : 's' }}<template v-if="fav"> · Favorita</template><template v-if="pinned"> · Anclada</template></span>
            <div class="PlCard_Actions">
                <button class="PlCard_Btn is-play" @click.stop="onPlay"><IconPlayerPlay :size="12" stroke="2" /> Reproducir</button>
                <button class="PlCard_Btn" @click.stop="onEdit"><IconPencil :size="12" stroke="2" /> Editar</button>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
@use '../Styles/PlaylistCard.scss';
</style>
