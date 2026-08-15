<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import {
    IconBox,
    IconDeviceGamepad,
    IconDownload,
    IconPhoto,
    IconPuzzle,
    IconBell,
    IconNews,
    IconSettings,
    IconChevronDown,
    IconPalette,
    IconCheck,
    IconX,
    IconMaximize,
    IconMinimize,
    IconLayoutSidebarLeftCollapse,
    IconLayoutSidebarLeftExpand,
} from '@tabler/icons-vue';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';
import { loadLocal, personalization } from '@/Common/Stores/Ui';
import iconNotFoundVersion from '../../assets/not_found/not_found_version.png';
import avatarNotFound from '../../assets/not_found/avatar_not_found.png';
import chickenImg from '../../assets/decorations/chicken.png';
import steveAlexImg from '../../assets/decorations/steve_and_alex.png';

const props = defineProps<{
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
}>();

function close() {
    emit('update:visible', false);
}

useOverlayEscape(close, { isActive: () => props.visible });

function onCloseOverlays() {
    close();
}

const bg = computed(() => personalization.value?.background ?? null);

const bgImageUrl = ref('');

type BackgroundKind = 'none' | 'image' | 'dynamic' | 'bundled';
const selKind = ref<BackgroundKind>('none');
const selBundled = ref('');

// ---------- Fondos incluidos en el launcher ----------

const bgAssets = import.meta.glob('../../assets/background/*', {
    eager: true,
}) as Record<string, { default: string }>;

function prettyBgName(file: string): string {
    const base = file.replace(/\.[^.]+$/, '').toLowerCase();
    if (base === 'bg-welcome') return 'Bienvenida';
    if (base === 'bg-welcome-2') return 'Bienvenida 2';
    if (base === 'bg-welcome-3') return 'Bienvenida 3';
    if (base === 'bg') return 'Clásico';
    const pretty = base
        .replace(/^bg[-_]?/, '')
        .replace(/[-_]+/g, ' ')
        .trim()
        .replace(/\b\w/g, (c) => c.toUpperCase());
    return pretty || file;
}

const bundledBackgrounds = Object.entries(bgAssets)
    .map(([path, mod]) => {
        const file = path.split('/').pop() ?? '';
        return { file, name: prettyBgName(file), url: mod.default };
    })
    .sort((a, b) => a.name.localeCompare(b.name));

const mainOptions = computed(() => [
    { kind: 'none' as BackgroundKind, label: 'Sin fondo', badge: '' },
    {
        kind: 'image' as BackgroundKind,
        label: 'Imagen',
        badge: bgImageUrl.value ? 'tu fondo' : 'ejemplo',
    },
    {
        kind: 'dynamic' as BackgroundKind,
        label: 'Dinámico',
        badge: `${bundledBackgrounds.length} fondos · 10s`,
    },
]);

// ---------- Rotación del fondo dinámico ----------

const dynIndex = ref(0);
let dynTimer: number | null = null;

function startDynamic() {
    stopDynamic();
    if (!bundledBackgrounds.length) return;
    dynIndex.value = 0;
    dynTimer = window.setInterval(() => {
        dynIndex.value = (dynIndex.value + 1) % bundledBackgrounds.length;
    }, 10_000);
}

function stopDynamic() {
    if (dynTimer !== null) {
        window.clearInterval(dynTimer);
        dynTimer = null;
    }
}

watch(selKind, (k) => {
    if (k === 'dynamic') startDynamic();
    else stopDynamic();
});

const selUrl = computed(() => {
    switch (selKind.value) {
        case 'image':
            return bgImageUrl.value || bundledBackgrounds[0]?.url || '';
        case 'dynamic':
            return bundledBackgrounds[dynIndex.value]?.url || '';
        case 'bundled':
            return bundledBackgrounds.find((b) => b.file === selBundled.value)?.url ?? '';
        default:
            return '';
    }
});

function dotFor(kind: BackgroundKind): string {
    switch (kind) {
        case 'image':
            return bgImageUrl.value || bundledBackgrounds[0]?.url || '';
        case 'dynamic':
            return bundledBackgrounds[dynIndex.value]?.url || '';
        default:
            return '';
    }
}

async function refreshBackground() {
    const b = bg.value;
    bgImageUrl.value = b?.type === 'image' ? await loadLocal(b.imagePath ?? '') : '';
    selKind.value = b?.type === 'dynamic' ? 'dynamic' : b?.type === 'image' ? 'image' : 'none';
}

// ---------- Controles de la vista previa ----------

const sideOpen = ref(true);
const fullscreen = ref(false);
const paletaOpen = ref(true);
    
    const colorItems = computed(() => {
        const c = (personalization.value?.colors ?? {}) as Record<string, string>;
        const defs: Array<{ key: string; label: string }> = [
        { key: 'buttonPrimary', label: 'Botón principal' },
        { key: 'tag', label: 'Etiquetas' },
        { key: 'progress', label: 'Progreso' },
        { key: 'playButton', label: 'Botón jugar' },
        { key: 'buttons', label: 'Controles' },
        { key: 'modal', label: 'Fondo modal' },
        { key: 'sidebar', label: 'Barra lateral' },
        { key: 'borderModal', label: 'Bordes de modal' },
        { key: 'border', label: 'Bordes' },
        { key: 'success', label: 'Éxito' },
        { key: 'warning', label: 'Avisos' },
        { key: 'error', label: 'Errores' },
        ];
        return defs.map((d) => ({ ...d, value: c[d.key] ?? '' }));
    });
    
    const textColorItems = computed(() => [
    { label: 'Texto primario', value: personalization.value?.fontPrimaryColor ?? '#ffffff' },
    { label: 'Texto secundario', value: personalization.value?.fontSecondaryColor ?? '#a9a9b2' },
    ]);
    
    const fontPrimaryName = computed(() => {
        const f = personalization.value?.fontPrimary;
        return !f || f === 'system' ? 'Sistema' : f;
    });
    
    const fontSecondaryName = computed(() => {
        const f = personalization.value?.fontSecondary;
        return !f || f === 'system' ? 'Sistema' : f;
    });
    
    const fontPrimarySizePct = computed(() =>
    `${Math.round((personalization.value?.fontPrimarySize ?? 1) * 100)}%`
    );
    
    const fontSecondarySizePct = computed(() =>
    `${Math.round((personalization.value?.fontSecondarySize ?? 1) * 100)}%`
    );

watch(
    () => props.visible,
    (v) => {
        if (v) {
            void refreshBackground();
        } else {
            stopDynamic();
        }
    }
);

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
    stopDynamic();
});
</script>

<template>
    <Teleport to="body">
        <div v-if="visible" class="PvRoot">
            <header class="PvBar">
                <div class="PvBar_Title">
                    <IconPalette stroke="2" />
                    <h2>Vista previa del launcher</h2>
                    <span>Así se verá con tus colores</span>
                </div>
                <div class="PvBar_Tools">
                    <button
                    class="PvBar_Tool"
                    :class="{ on: paletaOpen }"
                    @click="paletaOpen = !paletaOpen"
                    title="Mostrar u ocultar la paleta de colores y textos"
                    >
                    <IconPalette stroke="2" />
                </button>
                <button
                class="PvBar_Tool"
                @click="fullscreen = !fullscreen"
                :title="fullscreen ? 'Salir de pantalla completa' : 'Pantalla completa'"
                >
                <IconMinimize v-if="fullscreen" stroke="2" />
                <IconMaximize v-else stroke="2" />
            </button>
            <button
            class="PvBar_Tool"
            :class="{ on: sideOpen }"
            @click="sideOpen = !sideOpen"
            title="Mostrar u ocultar la lista de fondos"
            >
            <IconLayoutSidebarLeftCollapse v-if="sideOpen" stroke="2" />
            <IconLayoutSidebarLeftExpand v-else stroke="2" />
        </button>
        <button class="PvBar_Close" @click="close" title="Cerrar vista previa">
            <IconX stroke="2" />
        </button>
    </div>
</header>

<div class="PvBody">
    <Transition name="PvSide">
        <aside v-if="sideOpen" class="PvSide">
                    <span class="PvSide_Title">Fondos</span>
                    <button
                        v-for="o in mainOptions"
                        :key="o.kind"
                        type="button"
                        class="PvSide_Item"
                        :class="{ on: selKind === o.kind }"
                        @click="selKind = o.kind"
                    >
                        <span
                            class="PvSide_Dot"
                            :style="dotFor(o.kind) ? { backgroundImage: `url(${dotFor(o.kind)})` } : {}"
                        ></span>
                        <span class="PvSide_Txt">
                            <b>{{ o.label }}</b>
                            <small>{{ o.badge || '—' }}</small>
                        </span>
                        <IconCheck v-if="selKind === o.kind" class="PvSide_Check" stroke="2" />
                    </button>

                    <span class="PvSide_Title PvSide_TitleSplit">Fondos incluidos</span>
                    <button
                        v-for="b in bundledBackgrounds"
                        :key="b.file"
                        type="button"
                        class="PvSide_Item"
                        :class="{ on: selKind === 'bundled' && selBundled === b.file }"
                        @click="selKind = 'bundled'; selBundled = b.file"
                    >
                        <span class="PvSide_Dot" :style="{ backgroundImage: `url(${b.url})` }"></span>
                        <span class="PvSide_Txt">
                            <b>{{ b.name }}</b>
                            <small>{{ b.file }}</small>
                        </span>
                        <IconCheck v-if="selKind === 'bundled' && selBundled === b.file" class="PvSide_Check" stroke="2" />
                    </button>
                    <span class="PvSide_Hint">El fondo dinámico alterna entre todos los fondos incluidos cada 10 segundos.</span>
    </aside>
</Transition>

<div class="PvStage">
    <div class="PvMain" :class="{ full: fullscreen }">
        <div class="BackgroundLayer">
            <img v-if="selUrl" :src="selUrl" alt="" draggable="false" />
        </div>
        
        <main class="MainContent">
            <div class="Sidebar">
                <div class="Item" title="Fotos">
                    <IconPhoto class="Item_Icon" stroke="2" />
                    <label class="Item_Label">Fotos</label>
                </div>
                <div class="Item" title="Instancias">
                    <IconBox class="Item_Icon" stroke="2" />
                    <label class="Item_Label">Instancias</label>
                </div>
                <div class="Item" title="Mods">
                    <IconPuzzle class="Item_Icon" stroke="2" />
                    <label class="Item_Label">Mods</label>
                </div>
                <div class="Item" title="Descargas">
                    <IconDownload class="Item_Icon" stroke="2" />
                    <label class="Item_Label">Descargas</label>
                </div>
            </div>
            
            <div class="Content">
                <div class="BottomControlVersion">
                    <div class="VersionSelected" title="Elegir versión o perfil">
                        <div class="ImageVersion">
                            <img :src="iconNotFoundVersion" alt="" draggable="false" />
                        </div>
                        <div class="InfoVersion">
                            <p>Version Perfil • Fabric :</p>
                            <h5>1.21.4</h5>
                        </div>
                    </div>
                    <div class="PlayBlock">
                        <div class="PlayButton">
                            <IconDeviceGamepad class="Icon" stroke="2" />
                            <h1>JUGAR</h1>
                            <div class="Decoration">
                                <img :src="chickenImg" class="Chicken" alt="" draggable="false" />
                                <img :src="steveAlexImg" alt="" draggable="false" />
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="TopOptions">
                    <div class="Others">
                        <div class="OptionOther" title="Notificaciones">
                            <IconBell stroke="2" />
                            <label class="OptionLabel">Notificaciones</label>
                        </div>
                        <div class="OptionOther" title="Noticias">
                            <IconNews stroke="2" />
                            <label class="OptionLabel">Noticias</label>
                        </div>
                        <div class="OptionOther" title="Configuración">
                            <IconSettings stroke="2" />
                            <label class="OptionLabel">Configuracion</label>
                        </div>
                    </div>
                    <div class="UserCardWrap">
                        <div class="UserCard">
                            <div class="Avatar">
                                <img :src="avatarNotFound" alt="" draggable="false" />
                            </div>
                            <div class="Username">
                                <h1>Steve</h1>
                                <p>Cuenta local</p>
                            </div>
                            <button class="ExpandButtonProfiles" tabindex="-1">
                                <IconChevronDown stroke="2" />
                            </button>
                        </div>
                    </div>
                </div>
            </div>
            
            <div v-if="paletaOpen" class="PvPalette">
                <div class="PvPalette_Card">
                    <div class="PvPalette_Head">
                        <span class="PvPalette_Title">
                            <IconPalette stroke="2" />
                            Paleta y tipografía
                        </span>
                        <button class="PvPalette_Close" title="Ocultar paleta" @click="paletaOpen = false">
                            <IconX stroke="2" />
                        </button>
                    </div>
                    <div class="PvPalette_Body">
                        <div class="PvPalette_Col">
                            <span class="PvPalette_ColTitle">Colores</span>
                            <div v-for="item in colorItems" :key="item.key" class="PvPalette_Swatch">
                                <span class="PvPalette_Dot" :style="{ background: item.value }"></span>
                                <span class="PvPalette_Name">{{ item.label }}</span>
                                <code>{{ item.value }}</code>
                            </div>
                        </div>
                        <div class="PvPalette_Col">
                            <span class="PvPalette_ColTitle">Textos</span>
                            <div class="PvPalette_Font">
                                <span class="PvPalette_FontLabel">Tipografía primaria</span>
                                <b class="PvPalette_Sample primary">StepLauncher JUGAR 123</b>
                                <span class="PvPalette_FontMeta">
                                    {{ fontPrimaryName }} · {{ fontPrimarySizePct }}
                                </span>
                            </div>
                            <div class="PvPalette_Font">
                                <span class="PvPalette_FontLabel">Tipografía secundaria</span>
                                <b class="PvPalette_Sample secondary">Instancias · Configuración · 72%</b>
                                <span class="PvPalette_FontMeta">
                                    {{ fontSecondaryName }} · {{ fontSecondarySizePct }}
                                </span>
                            </div>
                            <div class="PvPalette_Font">
                                <span class="PvPalette_FontLabel">Colores de texto</span>
                                <div class="PvPalette_TextRow">
                                    <div v-for="t in textColorItems" :key="t.label" class="PvPalette_Swatch">
                                        <span class="PvPalette_Dot" :style="{ background: t.value }"></span>
                                        <span class="PvPalette_Name">{{ t.label }}</span>
                                        <code>{{ t.value }}</code>
                                    </div>
                                </div>
                                <span class="PvPalette_FontMeta">Muestra con sombra de texto activa</span>
                                <span class="PvPalette_ShadowSample">Texto con sombra</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </main>
    </div>
    
    <div v-show="!fullscreen" class="PvExamples">
        <span class="PvEx_Badge">Lista</span>
        <div class="PvEx_Progress">
            <div class="PvEx_ProgressTrack">
                <span class="PvEx_ProgressFill"></span>
            </div>
            <span class="PvEx_Pct">72%</span>
        </div>
        <span class="PvEx_Msg good">Descarga completada ✓</span>
        <span class="PvEx_Msg warn">⚠ Espacio insuficiente</span>
        <span class="PvEx_Msg bad">✕ Error al conectar</span>
        <span class="PvEx_Btn ghost">Cancelar</span>
        <span class="PvEx_Btn primary">Aceptar</span>
        <span class="PvEx_Btn danger">Eliminar</span>
    </div>
</div>
</div>
</div>
</Teleport>
</template>

<style scoped lang="scss">
@use '../Common/Styles/App/App.scss';
@use './Styles/PersonalizationPreview.scss';
</style>