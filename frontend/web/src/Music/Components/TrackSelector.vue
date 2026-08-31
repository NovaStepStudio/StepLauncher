<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { IconSearch, IconCheck, IconMusic, IconSelectAll, IconX, IconLayoutGrid, IconList, IconLoader2, IconPhoto } from '@tabler/icons-vue';
import type { LocalTrack } from '../LocalStore';
import { totalTracks as globalTotal } from '../LocalStore';

// Nota: Cover thumb baja resolución (128x128) se usa siempre, no raw

const props = defineProps<{
    modelValue: string[];
    tracks: LocalTrack[];
    title?: string;
}>();

const emit = defineEmits<{
    (e: 'update:modelValue', v: string[]): void;
}>();

const query = ref('');
const page = ref(1);
const pageSize = 20;
const gridMode = ref(false);
const isLoading = ref(false);
const internalTracks = ref<LocalTrack[]>([]);
const totalServer = ref(0);
const hasFetched = ref(false);

let debounceTimer: any = null;

const selectedSet = computed(() => new Set(props.modelValue));

// Intento server-side: si hay backend disponible, paginamos desde Go con thumb baja resolución
async function fetchPage(): Promise<void> {
    isLoading.value = true;
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.GetMusicTracksPaged === 'function') {
            const offset = (page.value - 1) * pageSize;
            const res: any = await mod.GetMusicTracksPaged(offset, pageSize, 'thumb', query.value.trim());
            if (res && Array.isArray(res.tracks)) {
                const tracks: LocalTrack[] = res.tracks.map((t: any) => ({
                    id: t.path,
                    path: t.path,
                    fileName: t.fileName || (t.path.split(/[\\/]/).pop() || ''),
                    title: t.title || t.fileName || 'Desconocido',
                    artist: t.artist || 'Desconocido',
                    duration: Number(t.duration) || 0,
                    coverUrl: t.coverThumb || t.coverRaw || '',
                    coverRaw: t.coverRaw || t.coverThumb || '',
                    hasCover: !!t.hasCover,
                }));
                internalTracks.value = tracks;
                totalServer.value = Number(res.total) || tracks.length;
                hasFetched.value = true;
                return;
            }
        }
        throw new Error('no backend');
    } catch (_e) {
        // Fallback client-side sobre props.tracks (solo si backend falla)
        const q = query.value.trim().toLowerCase();
        let filtered: LocalTrack[] = props.tracks;
        if (q) filtered = props.tracks.filter((t) =>
            t.title.toLowerCase().includes(q) ||
            t.artist.toLowerCase().includes(q) ||
            t.fileName.toLowerCase().includes(q) ||
            t.path.toLowerCase().includes(q)
        );
        totalServer.value = filtered.length;
        const s = (page.value - 1) * pageSize;
        internalTracks.value = filtered.slice(s, s + pageSize);
        hasFetched.value = true;
    } finally {
        isLoading.value = false;
    }
}

// Debounce query
watch(query, () => {
    page.value = 1;
    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => { void fetchPage(); }, 280);
});
watch(page, () => { void fetchPage(); });

onMounted(() => { void fetchPage(); });

// Si props.tracks cambia y aún no hay fetch server, reintentar
watch(() => props.tracks.length, () => {
    if (!hasFetched.value) void fetchPage();
});

// Fallback si globalTotal cambia (biblioteca cargó después)
watch(globalTotal, () => {
    if (hasFetched.value && totalServer.value < globalTotal.value) void fetchPage();
});

const displayTracks = computed(() => internalTracks.value);
const totalPages = computed(() => Math.max(1, Math.ceil(totalServer.value / pageSize)));
const filteredCount = computed(() => totalServer.value);
const hasTracks = computed(() => totalServer.value > 0 || props.tracks.length > 0 || isLoading.value);

const pageItems = computed<(number|'ellipsis')[]>(() => {
    const total = totalPages.value, cur = page.value;
    if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);
    const items: (number|'ellipsis')[] = [1];
    const lo = Math.max(2, cur - 1), hi = Math.min(total - 1, cur + 1);
    if (lo > 2) items.push('ellipsis');
    for (let p = lo; p <= hi; p++) items.push(p);
    if (hi < total - 1) items.push('ellipsis');
    items.push(total);
    return items;
});

function goToPage(p: number): void { if (p >= 1 && p <= totalPages.value) page.value = p; }

function isSelected(path: string): boolean {
    return selectedSet.value.has(path);
}

function toggle(path: string): void {
    const s = new Set(props.modelValue);
    if (s.has(path)) s.delete(path);
    else s.add(path);
    emit('update:modelValue', Array.from(s));
}

function selectAllFiltered(): void {
    const s = new Set(props.modelValue);
    displayTracks.value.forEach((t) => s.add(t.path));
    // También seleccionar todos los filtrados server side: necesita fetch all IDs? Por ahora solo página visible
    // Para seleccionar todos filtrados reales, necesitamos listar IDs; hacemos fallback: añadir todos los visibles paginados
    // Si usuario quiere todos, puede usar selectAllFiltered repetido navegando páginas o botón adicional
    emit('update:modelValue', Array.from(s));
}

async function selectAllMatching(): Promise<void> {
    // Selecciona todas las pistas que coinciden con query (server side)
    try {
        const mod: any = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        if (typeof mod.GetMusicTracksPaged === 'function') {
            const res: any = await mod.GetMusicTracksPaged(0, 5000, 'none', query.value.trim());
            if (res && Array.isArray(res.tracks)) {
                const s = new Set(props.modelValue);
                res.tracks.forEach((t: any) => s.add(t.path));
                emit('update:modelValue', Array.from(s));
                return;
            }
        }
    } catch (_e) {}
    selectAllFiltered();
}

function clearFiltered(): void {
    const s = new Set(props.modelValue);
    const pageSet = new Set(displayTracks.value.map((t) => t.path));
    for (const p of Array.from(s)) if (pageSet.has(p)) s.delete(p);
    emit('update:modelValue', Array.from(s));
}

function clearAll(): void {
    emit('update:modelValue', []);
}

function formatDuration(sec: number): string {
    if (!Number.isFinite(sec) || sec <= 0) return '—';
    const m = Math.floor(sec / 60), s = Math.floor(sec % 60);
    return `${m}:${String(s).padStart(2, '0')}`;
}

const selectedCount = computed(() => props.modelValue.length);
</script>

<template>
    <div class="TrackSelector">
        <div class="TrackSelector_Head">
            <span class="TrackSelector_Title">{{ title ?? 'Selecciona pistas' }}</span>
            <span class="TrackSelector_Badge">{{ selectedCount }} seleccionadas · {{ filteredCount }} coinciden</span>
        </div>

        <div class="TrackSelector_Toolbar">
            <label class="TrackSelector_Search">
                <IconSearch stroke="2" />
                <input v-model="query" type="text" placeholder="Buscar por título, artista o archivo…" spellcheck="false" />
                <button v-if="query" class="TrackSelector_Clear" @click="query=''"><IconX stroke="2" /></button>
                <IconLoader2 v-if="isLoading" class="TrackSelector_Spinner spinning" :size="14" stroke="2" />
            </label>
            <div class="TrackSelector_Actions">
                <button class="SsBtn SsBtnSmall" :disabled="!displayTracks.length" @click="selectAllMatching"><IconSelectAll stroke="2" /> Todos filtrados</button>
                <button class="SsBtn SsBtnSmall" :disabled="selectedCount===0" @click="clearFiltered">Quitar visibles</button>
                <button class="SsBtn SsBtnSmall SsBtnGhost" :disabled="selectedCount===0" @click="clearAll">Limpiar</button>
                <span class="TrackSelector_Divider"></span>
                <button class="TrackSelector_ViewBtn" :class="{ on: !gridMode }" title="Vista lista" @click="gridMode=false"><IconList :size="14" stroke="2" /></button>
                <button class="TrackSelector_ViewBtn" :class="{ on: gridMode }" title="Vista grilla 2x2" @click="gridMode=true"><IconLayoutGrid :size="14" stroke="2" /></button>
            </div>
        </div>

        <!-- Paginación superior -->
        <div v-if="totalPages > 1" class="TrackSelector_Pager is-top">
            <span class="TrackSelector_FootInfo">Página {{ page }} de {{ totalPages }} · {{ filteredCount }} pistas</span>
            <nav class="TrackSelector_Pages">
                <button class="TrackSelector_Pg" :disabled="page<=1" @click="goToPage(page-1)">‹</button>
                <template v-for="(it,i) in pageItems" :key="`${it}-${i}`">
                    <span v-if="it==='ellipsis'" class="TrackSelector_Gap">…</span>
                    <button v-else class="TrackSelector_Pg" :class="{ on: it===page }" @click="goToPage(it as number)">{{ it }}</button>
                </template>
                <button class="TrackSelector_Pg" :disabled="page>=totalPages" @click="goToPage(page+1)">›</button>
            </nav>
        </div>

        <div v-if="isLoading && !displayTracks.length" class="TrackSelector_List is-loading">
            <div v-for="i in 6" :key="i" class="TrackSelector_Row is-skeleton">
                <span class="TrackSelector_Cover is-skeleton"><span class="TrackSelector_Shimmer"></span></span>
                <span class="TrackSelector_Info"><b class="TrackSelector_TrackTitle skeleton"></b><em class="TrackSelector_TrackSub skeleton"></em></span>
                <span class="TrackSelector_Dur skeleton"></span>
            </div>
        </div>

        <div v-else-if="!hasTracks" class="TrackSelector_Empty">
            <span class="TrackSelector_EmptyIcon"><IconMusic stroke="1.5" /></span>
            <b>No hay música para elegir</b>
            <p>Configura tu carpeta en Ajustes → Música y escanea.</p>
        </div>

        <div v-else-if="!displayTracks.length" class="TrackSelector_Empty small">
            <p>Sin resultados para “{{ query }}”.</p>
        </div>

        <div v-else class="TrackSelector_List" :class="{ 'is-grid': gridMode, 'is-loading': isLoading }">
            <div v-if="isLoading" class="TrackSelector_Overlay"><IconLoader2 :size="16" stroke="2" class="spinning" /> Cargando carátulas en baja resolución…</div>
            <label v-for="t in displayTracks" :key="t.path" class="TrackSelector_Row" :class="{ selected: isSelected(t.path) }">
                <input type="checkbox" :checked="isSelected(t.path)" @change="toggle(t.path)" />
                <span class="TrackSelector_Cover" :class="{ 'is-thumb': !!t.coverUrl, 'is-loading': (t as any).hasCover && !(t as any).coverUrl }">
                    <img v-if="t.coverUrl" :src="t.coverUrl" alt="" loading="lazy" />
                    <span v-else-if="(t as any).hasCover && isLoading" class="TrackSelector_CoverPlaceholder"><IconPhoto :size="14" stroke="1.5" class="pulse" /></span>
                    <span v-else class="TrackSelector_CoverPlaceholder"><IconMusic stroke="1.5" :size="16" /></span>
                </span>
                <span class="TrackSelector_Info">
                    <b class="TrackSelector_TrackTitle" :title="t.title">{{ t.title }}</b>
                    <em class="TrackSelector_TrackSub">{{ t.artist }} · {{ t.fileName }}</em>
                </span>
                <span class="TrackSelector_Dur">{{ formatDuration(t.duration) }}</span>
                <span class="TrackSelector_Check" :class="{ on: isSelected(t.path) }"><IconCheck stroke="2.5" /></span>
            </label>
        </div>

        <div v-if="totalPages > 1" class="TrackSelector_Footer">
            <span class="TrackSelector_FootInfo">Página {{ page }} de {{ totalPages }} · {{ filteredCount }} pistas</span>
            <nav class="TrackSelector_Pages">
                <button class="TrackSelector_Pg" :disabled="page<=1" @click="goToPage(page-1)">‹</button>
                <template v-for="(it,i) in pageItems" :key="`${it}-${i}`">
                    <span v-if="it==='ellipsis'" class="TrackSelector_Gap">…</span>
                    <button v-else class="TrackSelector_Pg" :class="{ on: it===page }" @click="goToPage(it as number)">{{ it }}</button>
                </template>
                <button class="TrackSelector_Pg" :disabled="page>=totalPages" @click="goToPage(page+1)">›</button>
            </nav>
        </div>
    </div>
</template>

<style scoped lang="scss">
@use '../Styles/TrackSelector.scss';
</style>
