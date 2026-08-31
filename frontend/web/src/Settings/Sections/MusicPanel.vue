<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import {
    IconMusic,
    IconFolder,
    IconFolderPlus,
    IconTrash,
    IconRefresh,
    IconSearch,
    IconLoader2,
    IconPhoto,
    IconDisc,
    IconSquare,
    IconPalette,
    IconEye,
    IconEyeOff,
    IconList,
    IconPlaylist,
    IconUpload,
    IconInfoCircle,
    IconClock,
    IconDatabase,
    IconFileMusic,
    IconCheck,
    IconAlertCircle,
    IconChevronDown,
    IconChevronUp,
    IconDroplet,
    IconX,
} from '@tabler/icons-vue';
import {
    GetMusicPanelConfig,
    UpdateMusicPanelConfig,
    PickMusicFolder,
    ScanMusicFolder,
    GetMusicFolders,
    AddMusicFolder,
    RemoveMusicFolder,
    ScanAllMusicFolders,
    GetMusicScanProgress,
    GetMusicLibraryStats,
    RebuildMusicIndex,
    ClearMusicCache,
    CancelMusicScan,
} from '@wailsjs/StepLauncher/internal/Services/Music/musicservice';
import { loadLocalLibrary } from '@/Music/LocalStore';

const musicFolder = ref('');
const musicFolders = ref<string[]>([]);
const foldersExpanded = ref(false);
const coverStyle = ref<'square' | 'disc'>('square');
const nowPlayingCover = ref<'square' | 'disc'>('square');
const nowPlayingHuge = ref(false);
const hugeCoverSize = ref(20); // 15..30 — se guarda como rem pero se muestra sin unidad
const colorMode = ref<'vibrant' | 'dominant' | 'muted' | 'least' | 'random'>('vibrant');
const coverOpacity = ref(100); // 0-100, persiste como 0..1 en coverOpacity (opacidad del ::after del panel Música)
const pageSize = ref(20);
const showCovers = ref(true);
const autoScan = ref<'off' | 'hourly' | 'daily' | 'onLaunch'>('off');
const smtcSource = ref<'auto' | 'background' | 'library'>('auto');
const msg = ref('');
const msgOk = ref(true);
const scanning = ref(false);
const foundCount = ref(0);
const scanProgress = ref<any>({ scanning: false, discovered: 0, processed: 0, skipped: 0, added: 0, updated: 0, removed: 0, currentFile: '', folder: '' });
const libraryStats = ref<any>({ totalTracks: 0, totalFolders: 0, cachedTracks: 0, cachedCovers: 0, databaseSizeBytes: 0, lastScanUnix: 0 });

let pollTimer: number | null = null;

function setMsg(t: string, ok = true) {
    msg.value = t;
    msgOk.value = ok;
    if (ok) setTimeout(() => { if (msg.value === t) msg.value = ''; }, 4000);
}

function formatBytes(b: number): string {
    if (!b || b <= 0) return '0 B';
    const u = ['B', 'KB', 'MB', 'GB'];
    let i = 0;
    let v = b;
    while (v >= 1024 && i < u.length - 1) { v /= 1024; i++; }
    return `${v.toFixed(i === 0 ? 0 : 1)} ${u[i]}`;
}
function formatDate(unix: number): string {
    if (!unix) return 'Nunca';
    try { return new Date(unix * 1000).toLocaleString('es-ES'); } catch { return '—'; }
}

const colorModeOptions = [
    { value: 'vibrant', label: 'Vibrante', desc: 'Colores más vivos de la carátula' },
    { value: 'dominant', label: 'Dominante', desc: 'Color principal de la carátula' },
    { value: 'muted', label: 'Suave', desc: 'Tonos apagados y elegantes' },
    { value: 'least', label: 'Sutil', desc: 'Color menos saturado' },
    { value: 'random', label: 'Aleatorio', desc: 'Color diferente cada vez' },
] as const;

const coverStyleOptions = [
    { value: 'square', label: 'Cuadrado', icon: IconSquare },
    { value: 'disc', label: 'Disco', icon: IconDisc },
] as const;

async function refreshStats() {
    try {
        const s: any = await GetMusicLibraryStats();
        libraryStats.value = s || { totalTracks: 0, totalFolders: 0, cachedTracks: 0, cachedCovers: 0, databaseSizeBytes: 0, lastScanUnix: 0 };
    } catch (_e) {}
    try {
        const p: any = await GetMusicScanProgress();
        scanProgress.value = p || scanProgress.value;
        scanning.value = !!p?.scanning;
    } catch (_e) {}
}

function startPolling() {
    if (pollTimer) return;
    pollTimer = window.setInterval(async () => {
        try {
            const p: any = await GetMusicScanProgress();
            if (p) {
                scanProgress.value = p;
                scanning.value = !!p.scanning;
                if (p.scanning) {
                    foundCount.value = Number(p.discovered) || foundCount.value;
                }
            }
            const s: any = await GetMusicLibraryStats().catch(() => null);
            if (s) libraryStats.value = s;
        } catch (_e) {}
    }, 400);
}
function stopPolling() {
    if (pollTimer) { clearInterval(pollTimer); pollTimer = null; }
}

let folderChangedHandler: (() => void) | null = null;
onMounted(async () => {
    folderChangedHandler = () => { void refreshStats(); void loadLocalLibrary().catch(() => {}); };
    window.addEventListener('stl:music-folder-changed', folderChangedHandler);
    try {
        const cfg = await GetMusicPanelConfig();
        let folders: string[] = [];
        try {
            const f = await GetMusicFolders();
            if (Array.isArray(f) && f.length) folders = f as string[];
        } catch (_e) {}
        if (!folders.length && cfg) {
            if (Array.isArray((cfg as any).musicFolders) && (cfg as any).musicFolders.length) folders = (cfg as any).musicFolders;
            else if (cfg.musicFolder) folders = [cfg.musicFolder];
        }
        musicFolders.value = folders;
        musicFolder.value = folders[0] || cfg?.musicFolder || '';
        if (cfg) {
            const cs = (cfg.coverStyle as any);
            coverStyle.value = cs === 'disc' ? 'disc' : 'square';
            const npc = (cfg as any).nowPlayingCover as any;
            nowPlayingCover.value = npc === 'disc' ? 'disc' : 'square';
            nowPlayingHuge.value = !!(cfg as any).nowPlayingHuge;
            // Tamaño carátula predominante 15..30 — clamp y default 20
            const rawSize = (cfg as any).nowPlayingHugeSize;
            if (typeof rawSize === 'number' && Number.isFinite(rawSize)) {
                hugeCoverSize.value = Math.round(Math.max(15, Math.min(30, rawSize)));
            } else {
                hugeCoverSize.value = 20;
            }
            colorMode.value = (cfg.colorMode as any) ?? 'vibrant';
            // Opacidad 0..1 -> 0..100 para el slider
            const rawOp = (cfg as any).coverOpacity;
            if (typeof rawOp === 'number' && Number.isFinite(rawOp)) {
                coverOpacity.value = Math.round(Math.max(0, Math.min(1, rawOp)) * 100);
            } else {
                coverOpacity.value = 100;
            }
            pageSize.value = cfg.pageSize ?? 20;
            showCovers.value = cfg.showCovers ?? true;
            autoScan.value = (cfg as any).autoScan || 'off';
            smtcSource.value = (cfg as any).smtcSource || 'auto';
        }
        await refreshStats();
        startPolling();
        setupAutoScan();
        if (musicFolders.value.length) {
            // no auto scan, solo mostrar stats
        }
    } catch (_e) {}
});
onUnmounted(() => {
    stopPolling();
    teardownAutoScan();
    if (folderChangedHandler) {
        window.removeEventListener('stl:music-folder-changed', folderChangedHandler);
        folderChangedHandler = null;
    }
});

let autoScanTimer: number | null = null;
function setupAutoScan() {
    teardownAutoScan();
    if (autoScan.value === 'off') return;
    if (autoScan.value === 'onLaunch' && musicFolders.value.length) {
        // Escanear al iniciar si hay carpetas y el índice está vacío o es viejo
        void scan();
        return;
    }
    const interval = autoScan.value === 'hourly' ? 60 * 60 * 1000 : 24 * 60 * 60 * 1000;
    autoScanTimer = window.setInterval(() => { void scan(); }, interval);
}
function teardownAutoScan() {
    if (autoScanTimer) { clearInterval(autoScanTimer); autoScanTimer = null; }
}

async function save(shouldRescan = false) {
    try {
        // Sincronizar musicFolder con primera carpeta — nunca reintroducir carpeta por defecto
        musicFolder.value = musicFolders.value[0] || '';
        try {
            const mod = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
            // Siempre persistir lista exacta (vacía limpia todo), nunca dejar fantasma
            await mod.SetMusicFolders(musicFolders.value as any);
            if (!musicFolders.value.length) {
                try { await mod.ClearMusicCache(); } catch (_e) {}
            }
        } catch (_e) {}
        // Sincronizar nowPlayingCover con el estilo elegido
        nowPlayingCover.value = coverStyle.value as any;
        const clampedSize = Math.max(15, Math.min(30, Number(hugeCoverSize.value) || 20));
        hugeCoverSize.value = clampedSize;
        await UpdateMusicPanelConfig({
            musicFolder: musicFolders.value[0] || '',
            musicFolders: musicFolders.value as any,
            coverStyle: coverStyle.value,
            nowPlayingCover: nowPlayingCover.value as any,
            nowPlayingHuge: nowPlayingHuge.value as any,
            nowPlayingHugeSize: clampedSize as any,
            colorMode: colorMode.value,
            coverOpacity: Math.max(0, Math.min(1, coverOpacity.value / 100)) as any,
            pageSize: pageSize.value,
            showCovers: showCovers.value,
            allowAbsolute: true,
            autoScan: autoScan.value as any,
            smtcSource: smtcSource.value as any,
        } as any);
        setupAutoScan();
        // SMTC: notificar cambio para que PlayerStore y MusicStore se coordinen
        window.dispatchEvent(new CustomEvent('stl:smtc-source-changed', { detail: smtcSource.value }));
        localStorage.setItem('stl_smtc_source', smtcSource.value);
        setMsg(musicFolders.value.length ? `Guardado ${musicFolders.value.length} carpeta(s)` : 'Configuración guardada');
        if (shouldRescan && (musicFolders.value.length || musicFolder.value.trim())) {
            void loadLocalLibrary();
            window.dispatchEvent(new CustomEvent('stl:music-folder-changed'));
        } else {
            // Solo estilo: notificar sin re-escanear
            window.dispatchEvent(new CustomEvent('stl:music-config-changed'));
        }
        await refreshStats();
    } catch (e: any) {
        setMsg(e?.message ?? 'No se pudo guardar', false);
    }
}

async function pickFolder() {
    try {
        const p = await PickMusicFolder();
        if (!p) return;
        if (!musicFolders.value.includes(p)) {
            musicFolders.value = [...musicFolders.value, p];
            try { await AddMusicFolder(p); } catch (_e) {}
        }
        musicFolder.value = musicFolders.value[0] || p;
        await save(true);
        void scan();
    } catch (e: any) {
        setMsg(e?.message ?? 'No se pudo elegir carpeta', false);
    }
}

async function removeFolder(folder: string) {
    try {
        musicFolders.value = musicFolders.value.filter((f) => f !== folder);
        musicFolder.value = musicFolders.value[0] || '';
        try { await RemoveMusicFolder(folder); } catch (_e) {}
        if (!musicFolders.value.length) {
            try {
                const mod = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
                await mod.ClearMusicCache();
            } catch (_e) {}
        }
        await save(true);
        // No auto-escanear si no quedan carpetas — evita re-crear carpeta fantasma
        if (musicFolders.value.length) void scan();
        else {
            foundCount.value = 0;
            await refreshStats();
            void loadLocalLibrary();
        }
        setMsg(`Carpeta quitada`);
    } catch (e: any) {
        setMsg(e?.message ?? 'No se pudo quitar', false);
    }
}

async function scan() {
    // PROHIBIDO: nunca usar carpeta por defecto — solo escanea carpetas explícitas del usuario
    const hasMulti = musicFolders.value.length > 0;
    if (!hasMulti && !musicFolder.value) {
        setMsg('Elige una carpeta primero', false);
        return;
    }
    scanning.value = true;
    try {
        let list: any = null;
        if (hasMulti) {
            try { list = await ScanAllMusicFolders(); } catch { list = await ScanMusicFolder(musicFolders.value[0]!); }
        } else {
            list = await ScanMusicFolder(musicFolder.value);
        }
        foundCount.value = Array.isArray(list) ? list.length : (libraryStats.value.totalTracks || 0);
        setMsg(`Escaneo completado: ${foundCount.value} pistas`);
        await refreshStats();
        void loadLocalLibrary();
        window.dispatchEvent(new CustomEvent('stl:music-folder-changed'));
        // Importar playlists encontradas en las carpetas escaneadas (no es sistema aparte)
        try {
            const { loadPlaylists } = await import('@/Music/PlaylistStore');
            await loadPlaylists();
            window.dispatchEvent(new CustomEvent('stl:playlists-changed'));
        } catch {}
        // También refrescar previewCovers ya cacheados
        window.dispatchEvent(new CustomEvent('stl:music-config-changed'));
    } catch (e: any) {
        const msg = String(e?.message ?? '');
        if (msg.includes('cancelado')) {
            setMsg('Escaneo cancelado', false);
            await refreshStats();
        } else {
            setMsg(msg || 'No se pudo escanear', false);
        }
    } finally {
        scanning.value = false;
    }
}

async function cancelScan() {
    try {
        await CancelMusicScan();
        setMsg('Cancelando escaneo…', false);
        // El backend pondrá scanning=false en el siguiente poll; optimismo local
        scanning.value = false;
    } catch (e: any) {
        setMsg(e?.message ?? 'No se pudo cancelar', false);
    }
}

async function rebuild() {
    scanning.value = true;
    setMsg('Reconstruyendo índice...');
    try {
        await RebuildMusicIndex();
        setMsg('Índice reconstruido');
        await refreshStats();
        void loadLocalLibrary();
    } catch (e: any) {
        setMsg(e?.message ?? 'Error al reconstruir', false);
    } finally {
        scanning.value = false;
    }
}
async function clearCache() {
    try {
        const n = await ClearMusicCache();
        setMsg(`Caché limpiada (${n} entradas)`);
        await refreshStats();
    } catch (e: any) {
        setMsg(e?.message ?? 'No se pudo limpiar', false);
    }
}

function onCoverOpacityInput(e: Event) {
    const v = Number((e.target as HTMLInputElement).value);
    coverOpacity.value = v;
    try { window.dispatchEvent(new CustomEvent('stl:panel-cover-opacity', { detail: v / 100 })); } catch (_e) {}
}
function onHugeSizeInput(e: Event) {
    const v = Math.round(Math.max(15, Math.min(30, Number((e.target as HTMLInputElement).value) || 20)));
    hugeCoverSize.value = v;
    // Preview instantáneo sin esperar a save(): NowPlayingView escucha stl:huge-cover-size-preview
    try { window.dispatchEvent(new CustomEvent('stl:huge-cover-size-preview', { detail: v })); } catch (_e) {}
}

async function importPl() {
    try {
        const { PickPlaylistFile, ImportPlaylistFile } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        const p = await PickPlaylistFile();
        if (!p) return;
        await ImportPlaylistFile(p);
        setMsg('Lista de reproducción importada');
    } catch (e: any) {
        setMsg(e?.message ?? 'No se pudo importar', false);
    }
}
</script>

<template>
    <div class="Ss">
        <!-- BIBLIOTECA -->
        <div class="SsGroup">
            <div class="SsGroupHead">
                <IconFolder :size="16" stroke="2" />
                <span>Biblioteca local</span>
                <span class="SsBadge" v-if="libraryStats.totalTracks">{{ libraryStats.totalTracks }} pistas</span>
            </div>
            <div class="SsRow" style="flex-direction:column; align-items:stretch; gap:0.6rem;">
                <div style="display:flex; gap:0.4rem; flex-wrap:wrap;">
                    <button class="SsBtn SsBtnPrimary" @click="pickFolder"><IconFolderPlus :size="14" stroke="2" /> Añadir carpeta</button>
                    <button class="SsBtn" :disabled="scanning" @click="scan"><IconSearch :size="14" stroke="2" /> {{ scanning ? 'Analizando…' : 'Re-escanear' }}</button>
                    <button class="SsBtn SsBtnGhost" :disabled="scanning" @click="rebuild" title="Reconstruir índice desde cero"><IconRefresh :size="14" stroke="2" /> Reconstruir</button>
                </div>

                <!-- Progreso escaneo vivo -->
                <div v-if="scanning || scanProgress.scanning" class="SsProgressBox">
                    <div class="SsProgressHead">
                        <IconLoader2 :size="14" stroke="2" class="spinning" />
                        <span>Escaneando{{ scanProgress.folder ? ' — ' + String(scanProgress.folder).split(/[\\/]/).pop() : '' }}</span>
                        <span class="SsProgressPct">{{ scanProgress.discovered ? `${scanProgress.processed || 0}/${scanProgress.discovered}` : `${scanProgress.current}/${scanProgress.total}` }}</span>
                        <button class="SsBtn SsBtnSmall SsBtnDanger" style="margin-left:auto; padding:0.2rem 0.5rem;" @click="cancelScan"><IconX :size="12" stroke="2" /> Cancelar</button>
                    </div>
                    <div class="SsProgressBar"><i :style="{ width: scanProgress.total ? `${Math.min(100, (scanProgress.current/scanProgress.total)*100)}%` : (scanProgress.discovered ? `${Math.min(100, (scanProgress.processed/scanProgress.discovered)*100)}%` : '45%') }"></i></div>
                    <div class="SsProgressGrid">
                        <span><IconFileMusic :size="12" stroke="2" /> {{ scanProgress.discovered || 0 }} descubiertas</span>
                        <span><IconCheck :size="12" stroke="2" /> {{ scanProgress.processed || 0 }} procesadas</span>
                        <span><IconEyeOff :size="12" stroke="2" /> {{ scanProgress.skipped || 0 }} omitidas</span>
                        <span><IconUpload :size="12" stroke="2" /> {{ scanProgress.added || 0 }} nuevas</span>
                        <span><IconRefresh :size="12" stroke="2" /> {{ scanProgress.updated || 0 }} actualizadas</span>
                        <span><IconTrash :size="12" stroke="2" /> {{ scanProgress.removed || 0 }} eliminadas</span>
                    </div>
                    <span v-if="scanProgress.currentFile" class="SsProgressFile" :title="scanProgress.currentFile">{{ String(scanProgress.currentFile).split(/[\\/]/).pop() }}</span>
                </div>

                <!-- Stats -->
                <div class="SsStatsGrid">
                    <div class="SsStat"><IconFileMusic :size="14" stroke="2" /><b>{{ libraryStats.totalTracks }}</b><em>pistas indexadas</em></div>
                    <div class="SsStat"><IconFolder :size="14" stroke="2" /><b>{{ libraryStats.totalFolders || musicFolders.length }}</b><em>carpetas</em></div>
                    <div class="SsStat"><IconDatabase :size="14" stroke="2" /><b>{{ formatBytes(libraryStats.databaseSizeBytes) }}</b><em>índice</em></div>
                    <div class="SsStat"><IconClock :size="14" stroke="2" /><b>{{ formatDate(libraryStats.lastScanUnix) }}</b><em>último escaneo</em></div>
                </div>
                <div class="SsStatsGrid" style="margin-top:0.4rem;">
                    <div class="SsStat small"><IconPhoto :size="12" stroke="2" /><b>{{ libraryStats.cachedCovers }}</b><em>carátulas</em></div>
                    <div class="SsStat small"><IconInfoCircle :size="12" stroke="2" /><b>{{ libraryStats.cachedTracks }}</b><em>en caché</em></div>
                    <div class="SsStat small"><IconAlertCircle :size="12" stroke="2" /><b>{{ libraryStats.missingCovers }}</b><em>sin carátula</em></div>
                    <button class="SsBtn SsBtnSmall" @click="clearCache"><IconTrash :size="12" stroke="2" /> Limpiar caché</button>
                </div>

                <div style="display:flex; align-items:center; justify-content:space-between; gap:0.5rem; margin-top:0.4rem;">
                    <span style="font-size:0.72rem; opacity:0.7; font-weight:600;"><IconFolder :size="12" stroke="2" /> {{ musicFolders.length }} carpeta(s)</span>
                    <button class="SsBtn SsBtnSmall" @click="foldersExpanded = !foldersExpanded"><template v-if="foldersExpanded"><IconChevronUp :size="12" stroke="2" /> Ocultar</template><template v-else><IconChevronDown :size="12" stroke="2" /> Ver</template></button>
                </div>
                <div v-if="foldersExpanded" style="display:flex; flex-direction:column; gap:0.35rem; margin-top:0.4rem;">
                    <div v-if="musicFolders.length" style="display:flex; flex-direction:column; gap:0.35rem;">
                        <div v-for="f in musicFolders" :key="f" style="display:flex; align-items:center; gap:0.5rem; padding:0.45rem 0.6rem; border-radius:0.5rem; background:color-mix(in srgb, var(--control-bg) 60%, transparent); border:1px solid color-mix(in srgb, var(--text-primary) 6%, transparent);">
                            <IconFolder :size="12" stroke="2" style="opacity:0.6; flex-shrink:0;" />
                            <span style="flex:1; min-width:0; font-size:0.72rem; white-space:nowrap; overflow:hidden; text-overflow:ellipsis;" :title="f">{{ f }}</span>
                            <button class="SsBtn SsBtnSmall" style="padding:0.2rem 0.5rem;" @click="removeFolder(f)"><IconTrash :size="12" stroke="2" /> Quitar</button>
                        </div>
                    </div>
                    <div v-else style="padding:0.6rem; border-radius:0.5rem; border:1px dashed color-mix(in srgb, var(--text-primary) 10%, transparent); text-align:center; opacity:0.6; font-size:0.72rem;">
                        Sin carpetas — añade al menos una
                    </div>
                </div>
            </div>
            <div class="SsTip">
                <IconInfoCircle :size="12" stroke="2" />
                <span v-if="foundCount">Último escaneo: {{ foundCount }} archivos. Índice incremental por tamaño/fecha.</span>
                <span v-else>El escaneo es incremental: solo procesa archivos nuevos o modificados.</span>
            </div>
        </div>

        <!-- PERSONALIZACIÓN -->
        <div class="SsGroup">
            <div class="SsGroupHead">
                <IconPalette :size="16" stroke="2" />
                <span>Personalización del panel</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel"><IconDisc :size="12" stroke="2" style="margin-right:4px;" /> Estilo de carátula</span>
                    <span class="SsDesc">Elige entre disco o carátula nítida.</span>
                </div>
                <div class="SsCtrl">
                    <select class="SsSel" v-model="coverStyle" @change="() => save()">
                        <option v-for="o in coverStyleOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
                    </select>
                </div>
            </div>

            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel"><IconPhoto :size="12" stroke="2" style="margin-right:4px;" /> Carátula Predominante</span>
                    <span class="SsDesc">En Ahora suena muestra solo la carátula enorme con título (1rem) y artista (0.65rem). Actívalo para ajustar el tamaño.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="nowPlayingHuge" @change="() => save()" /><span class="SsTgS"></span></label>
                </div>
            </div>

            <div v-if="nowPlayingHuge" class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel"><IconPhoto :size="12" stroke="2" style="margin-right:4px;" /> Tamaño de la carátula</span>
                    <span class="SsDesc">Desliza entre 15 (compacto) y 30 (inmersivo). El mínimo siempre es 3 — se muestra sin unidad “rem”.</span>
                </div>
                <div class="SsCtrl" style="gap:0.6rem; flex-wrap:nowrap;">
                    <span style="font-size:0.66rem; opacity:0.5; font-weight:600;">15</span>
                    <input
                        type="range"
                        min="15"
                        max="30"
                        step="1"
                        :value="hugeCoverSize"
                        @input="onHugeSizeInput"
                        @change="() => save()"
                        style="width: 128px; accent-color: var(--accent, #5ed89a);"
                    />
                    <span style="font-size:0.66rem; opacity:0.5; font-weight:600;">30</span>
                    <span style="min-width:2.4rem; text-align:center; font-size:0.72rem; font-weight:800; opacity:0.95; padding:0.18rem 0.4rem; border-radius:0.35rem; background:color-mix(in srgb, var(--accent, #5ed89a) 14%, transparent); border:1px solid color-mix(in srgb, var(--accent, #5ed89a) 18%, transparent);">
                        {{ hugeCoverSize }}
                    </span>
                </div>
            </div>

            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel"><IconPalette :size="12" stroke="2" style="margin-right:4px;" /> Color de las tarjetas</span>
                    <span class="SsDesc">Se aplica al fondo de listas de reproducción y filas. Usa la carátula si tiene color.</span>
                </div>
                <div class="SsCtrl">
                    <select class="SsSel" v-model="colorMode" @change="() => save()">
                        <option v-for="o in colorModeOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
                    </select>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel"><IconDroplet :size="12" stroke="2" style="margin-right:4px;" /> Opacidad del fondo</span>
                    <span class="SsDesc">Qué tan visible es el color de la carátula en el fondo del panel Música (el degradado superior). 0 transparente, 100 opaco.</span>
                </div>
                <div class="SsCtrl" style="gap:0.6rem;">
                    <input
                        type="range"
                        min="0"
                        max="100"
                        step="5"
                        :value="coverOpacity"
                        @input="onCoverOpacityInput"
                        @change="() => save()"
                        style="width: 128px; accent-color: var(--accent, #5ed89a);"
                    />
                    <span style="min-width: 2.8rem; text-align: right; font-size: 0.72rem; font-weight: 700; opacity: 0.9;">{{ coverOpacity }}%</span>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel"><IconEye :size="12" stroke="2" style="margin-right:4px;" /> Mostrar carátulas en filas</span>
                    <span class="SsDesc">Ocultar mejora rendimiento con 1k+ pistas.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="showCovers" @change="() => save()" /><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel"><IconList :size="12" stroke="2" style="margin-right:4px;" /> Pistas por página</span>
                    <span class="SsDesc">Para no cargar 1k filas de golpe.</span>
                </div>
                <div class="SsCtrl">
                    <select class="SsSel" v-model.number="pageSize" @change="() => save()">
                        <option :value="20">20</option>
                        <option :value="50">50</option>
                        <option :value="100">100</option>
                    </select>
                </div>
            </div>
        </div>

        <div class="SsGroup">
            <div class="SsGroupHead">
                <IconClock :size="16" stroke="2" />
                <span>Automatización</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel"><IconRefresh :size="12" stroke="2" style="margin-right:4px;" /> Re-escanear automáticamente</span>
                    <span class="SsDesc">Mantén tu biblioteca actualizada sin intervención.</span>
                </div>
                <div class="SsCtrl">
                    <select class="SsSel" v-model="autoScan" @change="() => save()">
                        <option value="off">Desactivado</option>
                        <option value="hourly">Cada hora</option>
                        <option value="daily">Cada día</option>
                        <option value="onLaunch">Al iniciar el launcher</option>
                    </select>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel"><IconDisc :size="12" stroke="2" style="margin-right:4px;" /> Control multimedia del sistema (SMTC)</span>
                    <span class="SsDesc">Elige qué música controla Windows: fondo o biblioteca. “Auto” decide según qué esté sonando.</span>
                </div>
                <div class="SsCtrl">
                    <select class="SsSel" v-model="smtcSource" @change="() => save()">
                        <option value="auto">Automático</option>
                        <option value="background">Solo música de fondo</option>
                        <option value="library">Solo biblioteca</option>
                    </select>
                </div>
            </div>
            <div class="SsTip">
                <IconInfoCircle :size="12" stroke="2" />
                <span>“Cada día” escanea a la misma hora del último escaneo. “Al iniciar” solo si hay carpetas configuradas.</span>
            </div>
        </div>

        <div class="SsGroup">
            <div class="SsGroupHead">
                <IconPlaylist :size="16" stroke="2" />
                <span>Listas de reproducción</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel"><IconUpload :size="12" stroke="2" style="margin-right:4px;" /> Importar lista de reproducción</span>
                    <span class="SsDesc">Soporta .m3u/.m3u8 y .json. Rutas relativas funcionan.</span>
                </div>
                <div class="SsCtrl">
                    <button class="SsBtn SsBtnPrimary" @click="importPl"><IconUpload :size="14" stroke="2" /> Importar</button>
                </div>
            </div>
        </div>

        <p v-if="msg" :class="['SsTip', { error: !msgOk }]" style="margin-top:0.6rem; opacity:1;"><IconInfoCircle :size="12" stroke="2" style="margin-right:4px;" /> {{ msg }}</p>
    </div>
</template>

<style scoped lang="scss">
@use '../Styles/General.scss';
.spinning { animation: spin 0.9s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
.SsProgressBox { padding:0.6rem; border-radius:0.5rem; background:color-mix(in srgb, var(--control-bg) 60%, transparent); border:1px solid color-mix(in srgb, var(--text-primary) 8%, transparent); display:flex; flex-direction:column; gap:0.4rem; }
.SsProgressHead { display:flex; align-items:center; gap:0.4rem; font-size:0.72rem; font-weight:600; }
.SsProgressPct { margin-left:auto; opacity:0.6; font-size:0.68rem; }
.SsProgressBar { height:4px; border-radius:999px; background:color-mix(in srgb, var(--text-primary) 10%, transparent); overflow:hidden; }
.SsProgressBar i { display:block; height:100%; background:var(--accent, #5ed89a); transition: width 0.2s; }
.SsProgressGrid { display:grid; grid-template-columns:repeat(3,1fr); gap:0.3rem; font-size:0.66rem; opacity:0.8; }
.SsProgressGrid span { display:flex; align-items:center; gap:0.25rem; }
.SsProgressFile { font-size:0.66rem; opacity:0.6; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
.SsStatsGrid { display:grid; grid-template-columns:repeat(4,1fr); gap:0.4rem; }
.SsStat { display:flex; flex-direction:column; align-items:center; gap:0.15rem; padding:0.5rem 0.3rem; border-radius:0.5rem; background:color-mix(in srgb, var(--control-bg) 50%, transparent); border:1px solid color-mix(in srgb, var(--text-primary) 6%, transparent); font-size:0.68rem; }
.SsStat b { font-size:0.82rem; }
.SsStat em { font-style:normal; opacity:0.6; font-size:0.62rem; }
.SsStat.small { padding:0.35rem 0.2rem; }
.SsBadge { margin-left:auto; font-size:0.62rem; padding:0.15rem 0.4rem; border-radius:999px; background:var(--accent, #5ed89a); color:#111; font-weight:700; }
</style>
