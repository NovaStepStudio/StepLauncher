<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { IconX, IconPalette, IconMusic, IconSparkles, IconPin } from '@tabler/icons-vue';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';
import { localTracks, totalTracks } from '../LocalStore';
import TrackSelector from './TrackSelector.vue';

const props = defineProps<{
    visible: boolean;
    initialTitle?: string;
    initialTracks?: string[];
    initialFavorite?: boolean;
    initialPinned?: boolean;
    initialColor?: string;
    initialCover?: string;
    mode?: 'create' | 'edit';
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
    (e: 'submit', data: { title: string; favorite: boolean; pinned: boolean; color: string; cover: string; tracks: string[] }): void;
}>();

const title = ref('');
const favorite = ref(false);
const pinned = ref(false);
const color = ref('');
const cover = ref('');
const selectedTracks = ref<string[]>([]);

const hasLibrary = computed(() => localTracks.value.length > 0 || totalTracks.value > 0);
const selectedCount = computed(() => selectedTracks.value.length);
const totalAvailable = computed(() => totalTracks.value || localTracks.value.length);

watch(() => props.visible, (v) => {
    if (v) {
        title.value = props.initialTitle ?? '';
        favorite.value = props.initialFavorite ?? false;
        pinned.value = props.initialPinned ?? false;
        color.value = props.initialColor ?? '';
        cover.value = props.initialCover ?? '';
        selectedTracks.value = [...(props.initialTracks ?? [])];
    }
});

function close(): void {
    emit('update:visible', false);
}

function submit(): void {
    if (!title.value.trim()) return;
    if (!selectedTracks.value.length) return;
    emit('submit', {
        title: title.value.trim(),
        favorite: favorite.value,
        pinned: pinned.value,
        color: color.value.trim(),
        cover: cover.value.trim(),
        tracks: [...selectedTracks.value],
    });
    close();
}

useOverlayEscape(close, { isActive: () => props.visible, priority: 2 });

const isEdit = computed(() => props.mode === 'edit');
</script>

<template>
    <Teleport to="body">
        <Transition name="PlaylistForm">
            <div v-if="visible" class="PlaylistForm_Overlay" @click.self="close">
                <div class="PlaylistForm_Dialog">
                    <div class="PlaylistForm_Head">
                        <div class="PlaylistForm_HeadText">
                            <span class="PlaylistForm_Icon"><IconMusic :size="16" stroke="2" /></span>
                            <div>
                                <h3>{{ isEdit ? 'Editar lista de reproducción' : 'Nueva lista de reproducción' }}</h3>
                                <p>{{ hasLibrary ? `${totalAvailable} pistas disponibles · selecciona y listo` : 'Añade música primero en Ajustes → Música' }}</p>
                            </div>
                        </div>
                        <button class="PlaylistForm_Close" @click="close" title="Cerrar"><IconX :size="16" stroke="2" /></button>
                    </div>

                    <div class="PlaylistForm_Body is-split">
                        <div class="PlaylistForm_Column is-content">
                            <div class="PlaylistForm_ColumnHead">
                                <span>Contenido</span>
                                <em>Información de la lista de reproducción</em>
                            </div>
                            <label class="PlaylistForm_Field">
                                <span>Título *</span>
                                <input class="SsIn" v-model="title" placeholder="Ej: Favoritas, Para jugar, Lofi night" maxlength="60" />
                                <em class="PlaylistForm_Hint">{{ title.trim().length }}/60</em>
                            </label>

                            <div class="PlaylistForm_RowCards">
                                <label class="PlaylistForm_ToggleCard" :class="{ on: favorite }">
                                    <input type="checkbox" v-model="favorite" />
                                    <span class="PlaylistForm_ToggleCardIcon"><IconSparkles :size="16" stroke="2" /></span>
                                    <span><b>Favorita</b><em>Aparece primero</em></span>
                                </label>
                                <label class="PlaylistForm_ToggleCard" :class="{ on: pinned }">
                                    <input type="checkbox" v-model="pinned" />
                                    <span class="PlaylistForm_ToggleCardIcon"><IconPin :size="16" stroke="2" /></span>
                                    <span><b>Anclada</b><em>Fija arriba</em></span>
                                </label>
                            </div>

                            <label class="PlaylistForm_Field">
                                <span>Color personalizado</span>
                                <div class="PlaylistForm_ColorRow">
                                    <input class="SsIn" v-model="color" placeholder="#a78bfa o vacío = auto" />
                                    <span class="PlaylistForm_ColorPreview" :style="{ background: color || 'color-mix(in srgb, var(--progress-color) 18%, transparent)', borderColor: color || 'transparent' }" title="Vista previa">
                                        <IconPalette :size="12" stroke="2" />
                                    </span>
                                </div>
                                <em class="PlaylistForm_Hint">Vacío usa la carátula más colorida.</em>
                            </label>
                            <label class="PlaylistForm_Field">
                                <span>Carátula custom</span>
                                <input class="SsIn" v-model="cover" placeholder="C:\covers\mi.png o cache/covers/..." />
                            </label>
                        </div>

                        <div class="PlaylistForm_Column is-musics">
                            <div class="PlaylistForm_ColumnHead">
                                <span>Músicas</span>
                                <span class="PlaylistForm_SelectorCount">{{ selectedCount }} seleccionadas</span>
                            </div>
                            <TrackSelector v-model="selectedTracks" :tracks="localTracks" title="Selecciona de tu biblioteca" />
                            <p v-if="!hasLibrary" class="PlaylistForm_Hint error">Sin música aún. Configura la carpeta y se cargará automáticamente.</p>
                            <p v-else-if="selectedCount===0" class="PlaylistForm_Hint warn">Selecciona al menos una pista para crear la lista de reproducción.</p>
                            <p v-else class="PlaylistForm_Hint ok">{{ selectedCount }} pistas listas.</p>
                        </div>
                    </div>

                    <div class="PlaylistForm_Footer">
                        <button class="SsBtn" @click="close">Cancelar</button>
                        <button class="SsBtn SsBtnPrimary" :disabled="!title.trim() || selectedCount===0" @click="submit">{{ isEdit ? 'Guardar cambios' : `Crear (${selectedCount})` }}</button>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use '../../Common/Styles/Components' as *;

.PlaylistForm_Overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.52);
    backdrop-filter: blur(10px);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 120;
    padding: 1rem;
}
.PlaylistForm_Dialog {
    width: 860px;
    max-width: 96vw;
    max-height: 88vh;
    background: var(--background-modal-primary);
    border: var(--border-modal-style);
    border-radius: 0.95rem;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.32);
}
.PlaylistForm_Head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.8rem;
    padding: 1rem 1.2rem;
    border-bottom: var(--border-style);
    background: linear-gradient(180deg, color-mix(in srgb, var(--background-modal-primary) 96%, white 4%), var(--background-modal-primary));
    flex-shrink: 0;
}
.PlaylistForm_HeadText {
    display: flex;
    align-items: center;
    gap: 0.7rem;
    min-width: 0;
    div { display: flex; flex-direction: column; gap: 0.12rem; min-width: 0; }
    h3 { margin: 0; font-size: 0.98rem; font-weight: 800; color: var(--text-primary); letter-spacing: -0.01em; }
    p { margin: 0; font-size: 0.66rem; opacity: 0.55; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
}
.PlaylistForm_Icon {
    width: 2rem;
    height: 2rem;
    border-radius: 0.55rem;
    background: linear-gradient(135deg, color-mix(in srgb, var(--color-tag) 18%, transparent), color-mix(in srgb, var(--progress-color) 14%, transparent));
    border: 1px solid color-mix(in srgb, var(--color-tag) 14%, transparent);
    display: flex;
    justify-content: center;
    align-items: center;
    color: var(--text-primary);
    flex-shrink: 0;
}
.PlaylistForm_Close {
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    width: 1.9rem;
    height: 1.9rem;
    display: flex;
    justify-content: center;
    align-items: center;
    border-radius: 0.45rem;
    flex-shrink: 0;
    transition: var(--transition);
    &:hover { background: var(--control-bg); color: var(--text-primary); }
}
.PlaylistForm_Body {
    padding: 1rem 1.2rem;
    display: flex;
    flex-direction: column;
    gap: 0.95rem;
    overflow-y: auto;
    scrollbar-width: thin;
    scrollbar-color: color-mix(in srgb, var(--text-primary) 14%, transparent) transparent;
    &::-webkit-scrollbar { width: 6px; height: 6px; }
    &::-webkit-scrollbar-thumb { background: color-mix(in srgb, var(--text-primary) 14%, transparent); border-radius: 999px; }
    &::-webkit-scrollbar-track { background: transparent; }
    &.is-split {
        display: grid;
        grid-template-columns: 340px 1fr;
        gap: 1rem;
        align-items: start;
        @media (max-width: 760px) { grid-template-columns: 1fr; }
        .TrackSelector_List { max-height: 380px; }
    }
}
.PlaylistForm_Column {
    display: flex;
    flex-direction: column;
    gap: 0.85rem;
    min-width: 0;
    &.is-content { position: sticky; top: 0; }
    &.is-musics { min-height: 0; }
}
.PlaylistForm_ColumnHead {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-bottom: 0.4rem;
    border-bottom: 1px solid color-mix(in srgb, var(--text-primary) 8%, transparent);
    span:first-child { font-size: 0.78rem; font-weight: 800; color: var(--text-primary); letter-spacing: -0.01em; }
    em { font-size: 0.64rem; opacity: 0.5; font-style: normal; }
}
.PlaylistForm_Field {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    span { font-size: 0.74rem; font-weight: 600; color: var(--text-primary); }
}
.PlaylistForm_Hint {
    font-size: 0.62rem;
    opacity: 0.5;
    font-style: normal;
    line-height: 1.35;
    &.error { opacity: 1; color: var(--color-error); background: color-mix(in srgb, var(--color-error) 8%, transparent); padding: 0.3rem 0.5rem; border-radius: 0.4rem; border: 1px solid color-mix(in srgb, var(--color-error) 16%, transparent); }
    &.warn { opacity: 1; color: var(--color-warning); }
    &.ok { opacity: 1; color: var(--color-success); }
}
.PlaylistForm_RowCards {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.5rem;
}
.PlaylistForm_ToggleCard {
    display: flex;
    align-items: center;
    gap: 0.55rem;
    padding: 0.6rem 0.7rem;
    border-radius: 0.6rem;
    background: color-mix(in srgb, var(--control-bg) 70%, transparent);
    border: 1px solid color-mix(in srgb, var(--text-primary) 7%, transparent);
    cursor: pointer;
    transition: var(--transition);
    input { display: none; }
    span { display: flex; flex-direction: column; gap: 0.08rem; b { font-size: 0.76rem; font-weight: 700; color: var(--text-primary); line-height: 1; } em { font-size: 0.6rem; opacity: 0.5; font-style: normal; } }
    &:hover { background: var(--control-bg); transform: translateY(-1px); }
    &.on { background: color-mix(in srgb, var(--progress-color) 10%, var(--control-bg)); border-color: color-mix(in srgb, var(--progress-color) 18%, transparent); box-shadow: 0 2px 10px color-mix(in srgb, var(--progress-color) 12%, transparent); }
}
.PlaylistForm_ToggleCardIcon {
    width: 1.7rem;
    height: 1.7rem;
    border-radius: 0.45rem;
    background: var(--control-bg);
    border: var(--border-style);
    display: flex;
    justify-content: center;
    align-items: center;
    flex-shrink: 0;
    font-size: 0.9rem;
}
.PlaylistForm_Grid2 {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.7rem;
    @media (max-width: 560px) { grid-template-columns: 1fr; }
}
.PlaylistForm_ColorRow {
    display: flex;
    gap: 0.45rem;
    align-items: center;
    .SsIn { flex: 1; min-width: 0; }
}
.PlaylistForm_ColorPreview {
    width: 2rem;
    height: 2rem;
    border-radius: 0.5rem;
    border: 1px solid transparent;
    display: flex;
    justify-content: center;
    align-items: center;
    flex-shrink: 0;
    color: var(--text-primary);
    box-shadow: 0 1px 6px rgba(0,0,0,0.12);
}
.PlaylistForm_SelectorWrap {
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
}
.PlaylistForm_SelectorHead {
    display: flex;
    justify-content: space-between;
    align-items: center;
    span:first-child { font-size: 0.76rem; font-weight: 700; color: var(--text-primary); }
}
.PlaylistForm_SelectorCount {
    font-size: 0.64rem;
    padding: 0.2rem 0.5rem;
    border-radius: 999px;
    background: var(--control-bg);
    border: var(--border-style);
    opacity: 0.7;
}
.PlaylistForm_Footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 0.9rem 1.2rem;
    border-top: var(--border-style);
    background: color-mix(in srgb, var(--background-modal-primary) 97%, white 3%);
    flex-shrink: 0;
}
.PlaylistForm-enter-active, .PlaylistForm-leave-active { transition: all 0.18s ease; }
.PlaylistForm-enter-from, .PlaylistForm-leave-to { opacity: 0; transform: scale(0.98); }
</style>
