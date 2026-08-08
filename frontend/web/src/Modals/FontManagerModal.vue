<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onUnmounted } from 'vue';
import { cleanFontName, fontByPath, fontByType, type LauncherAssets, type FontSlotData } from '../Stores/Fonts';
import { CLOSE_OVERLAYS_EVENT } from '../Stores/Idle';

const props = defineProps<{
    visible: boolean;
    assets: LauncherAssets;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
    (e: 'changed'): void;
}>();

function onCloseOverlays() {
    emit('update:visible', false);
}

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});

const files = ref<string[]>([]);
const names = ref<Record<string, string>>({});
const msg = ref('');
const busy = ref(false);

const step = ref<'list' | 'import'>('list');
const pickedPath = ref('');
const pickedName = ref('');
const importSlot = ref<'primary' | 'secundary'>('primary');
const importName = ref('');

let nameSaveTimer: number | null = null;

function fontRef(file: string): string {
    return 'launcher/fonts/' + file;
}

function entryFor(file: string): FontSlotData | undefined {
    return fontByPath(props.assets, fontRef(file));
}

function slotLabel(file: string): string {
    const e = entryFor(file);
    if (e?.type === 'primary') return 'Tipografía principal';
    if (e?.type === 'secundary') return 'Tipografía secundaria';
    return '';
}

async function reload() {
    try {
        const list = await (window as any).go?.main?.App?.ListFontFiles?.();
        files.value = Array.isArray(list) ? list : [];
        const next: Record<string, string> = {};
        for (const f of files.value) {
            next[f] = names.value[f] || entryFor(f)?.name || cleanFontName(f);
        }
        names.value = next;
    } catch { }
}

function withTimeout<T>(promise: Promise<T> | undefined, ms: number, message: string): Promise<T> {
    return new Promise<T>((resolve, reject) => {
        const timer = window.setTimeout(() => reject(new Error(message)), ms);
        (promise ?? Promise.resolve(undefined as unknown as T)).then(
            (v: T) => { window.clearTimeout(timer); resolve(v); },
            (e: unknown) => { window.clearTimeout(timer); reject(e); }
        );
    });
}

const delay = (ms: number) => new Promise<void>((r) => window.setTimeout(r, ms));

function onNameInput() {
    if (nameSaveTimer !== null) {
        window.clearTimeout(nameSaveTimer);
    }
    nameSaveTimer = window.setTimeout(persistNames, 500);
}

async function persistNames() {
    if (nameSaveTimer !== null) {
        window.clearTimeout(nameSaveTimer);
        nameSaveTimer = null;
    }
    if (busy.value) return;
    const list = [...(props.assets.fonts ?? [])].map((e) => ({ ...e }));
    let changed = false;
    for (const f of files.value) {
        const name = (names.value[f] ?? '').trim();
        if (!name) continue;
        const ref = fontRef(f);
        const existing = list.find((e) => e.path === ref);
        if (existing && existing.name !== name) {
            existing.name = name;
            changed = true;
        }
    }
    if (!changed) return;
    try {
        await (window as any).go?.main?.App?.SaveLauncherAssets?.({ fonts: list });
        emit('changed');
    } catch { }
}

async function pickFont() {
    if (busy.value) return;
    busy.value = true;
    msg.value = '';
    try {
        const p = await (window as any).go?.main?.App?.PickFontFile?.();
        if (typeof p === 'string' && p) {
            const base = p.split(/[\\/]/).pop() ?? 'tipografia';
            pickedPath.value = p;
            pickedName.value = base;
            importName.value = cleanFontName(base);
            importSlot.value = fontByType(props.assets, 'primary')?.path ? 'secundary' : 'primary';
            step.value = 'import';
        }
    } catch (e: any) {
        msg.value = e?.message ?? 'No se pudo abrir el diálogo de selección.';
    } finally {
        busy.value = false;
    }
}

async function confirmImport() {
    if (busy.value) return;
    const name = importName.value.trim();
    if (!name) {
        msg.value = 'Ponle un nombre a la tipografía antes de importarla.';
        return;
    }
    busy.value = true;
    msg.value = '';
    try {
        const rel = await (window as any).go?.main?.App?.ImportFont?.(pickedPath.value);
        if (typeof rel !== 'string' || !rel) {
            msg.value = 'No se pudo importar la tipografía.';
            return;
        }
        const list = [...(props.assets.fonts ?? [])].map((e) => ({ ...e }));
        for (const e of list) {
            if (e.type === importSlot.value && e.path !== rel) e.type = '';
        }
        const idx = list.findIndex((e) => e.path === rel);
        const existing = list[idx];
        if (existing) {
            existing.type = importSlot.value;
            existing.name = name;
        } else {
            list.push({ type: importSlot.value, name, path: rel });
        }
        await (window as any).go?.main?.App?.SaveLauncherAssets?.({ fonts: list });
        msg.value = `Tipografía importada como ${importSlot.value === 'primary' ? 'principal' : 'secundaria'}: ${name}`;
        step.value = 'list';
        await reload();
        emit('changed');
    } catch (e: any) {
        msg.value = e?.message ?? 'Error al importar la tipografía.';
    } finally {
        busy.value = false;
    }
}

async function assign(slot: 'primary' | 'secundary', file: string) {
    if (busy.value) return;
    busy.value = true;
    msg.value = '';
    const name = (names.value[file] ?? '').trim() || cleanFontName(file);
    names.value[file] = name;
    const ref = fontRef(file);
    const list = [...(props.assets.fonts ?? [])].map((e) => ({ ...e }));
    for (const e of list) {
        if (e.type === slot && e.path !== ref) e.type = '';
    }
    const idx = list.findIndex((e) => e.path === ref);
    const existing = list[idx];
    if (existing) {
        existing.type = slot;
        existing.name = name;
    } else {
        list.push({ type: slot, name, path: ref });
    }
    try {
        await (window as any).go?.main?.App?.SaveLauncherAssets?.({ fonts: list });
        msg.value = `Tipografía asignada como ${slot === 'primary' ? 'principal' : 'secundaria'}: ${name}`;
        await reload();
        emit('changed');
    } catch (e: any) {
        msg.value = e?.message ?? 'Error al guardar la tipografía.';
    } finally {
        busy.value = false;
    }
}

async function remove(file: string) {
    if (busy.value) return;
    busy.value = true;
    msg.value = '';
    const ref = fontRef(file);
    try {
        const next: LauncherAssets = {
            fonts: (props.assets.fonts ?? []).filter((e) => e.path !== ref),
        };
        await (window as any).go?.main?.App?.SaveLauncherAssets?.(next);
        await reload();
        await nextTick();
        emit('changed');
        await delay(150);
        await withTimeout(
            (window as any).go?.main?.App?.DeleteFontFile?.(file),
            8000,
            'La eliminación tardó demasiado (el archivo puede estar en uso).'
        );
        msg.value = `Fuente eliminada: ${cleanFontName(file)}`;
        await reload();
    } catch (e: any) {
        msg.value = e?.message ?? 'Error al eliminar la fuente.';
        await reload();
    } finally {
        busy.value = false;
    }
}

watch(
    () => props.visible,
    (v) => {
        if (v) {
            msg.value = '';
            reload();
        }
    }
);

watch(
    () => props.assets,
    () => {
        const next: Record<string, string> = {};
        for (const f of files.value) {
            next[f] = entryFor(f)?.name || names.value[f] || cleanFontName(f);
        }
        names.value = next;
    }
);
</script>

<template>
    <Teleport to="body">
        <div v-if="visible" class="FontManager_Overlay" @click.self="emit('update:visible', false)">
            <div class="FontManager_Dialog">
                <div class="FontManager_Head">
                    <h3>Gestionar tipografías</h3>
                    <button class="FontManager_Close" @click="emit('update:visible', false)">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                    </button>
                </div>

                <template v-if="step === 'list'">
                    <p class="FontManager_Info">Pulsa <strong>Importar tipografía</strong> y selecciona el archivo con el diálogo del sistema. Podrás ponerle el nombre que quieras y usarla como principal o secundaria. Edita el <strong>nombre</strong> de cada fuente: se guarda automáticamente.</p>

                    <div class="FontManager_Actions">
                        <button class="SsBtn SsBtnPrimary" :disabled="busy" @click="pickFont">
                            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                            Importar tipografía
                        </button>
                        <button class="SsBtn" :disabled="busy" @click="reload">
                            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                            Actualizar
                        </button>
                    </div>

                    <div v-if="files.length" class="FontManager_List">
                        <div v-for="f in files" :key="f" class="FontManager_Item">
                            <div class="FontManager_ItemInfo">
                                <span class="FontManager_FileName">{{ names[f] || cleanFontName(f) }}</span>
                                <span v-if="slotLabel(f)" class="FontManager_Badge">{{ slotLabel(f) }}</span>
                                <span v-else class="FontManager_Unassigned">Sin asignar</span>
                            </div>
                            <div class="FontManager_ItemRow">
                                <input class="SsIn" v-model="names[f]" :placeholder="cleanFontName(f)" @input="onNameInput" />
                                <div class="FontManager_ItemBtns">
                                    <button class="SsBtn" :disabled="busy" @click="assign('primary', f)">Primaria</button>
                                    <button class="SsBtn" :disabled="busy" @click="assign('secundary', f)">Secundaria</button>
                                    <button class="SsBtn SsBtnDanger" :disabled="busy" @click="remove(f)">Quitar</button>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div v-else class="FontManager_Empty">
                        <p>No hay tipografías importadas todavía.</p>
                    </div>
                </template>

                <template v-else>
                    <div class="FontManager_Import">
                        <div class="FontManager_ImportFile">
                            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 7V4h16v3"/><path d="M9 20h6"/><path d="M12 4v16"/></svg>
                            <span>{{ cleanFontName(pickedName) }}</span>
                        </div>
                        <div class="FontManager_ImportRow">
                            <span class="FontManager_ImportLabel">Usar como</span>
                            <div class="FontManager_ImportSlots">
                                <button class="SsBtn" :class="{ 'SsBtnPrimary': importSlot === 'primary' }" :disabled="busy" @click="importSlot = 'primary'">Primaria</button>
                                <button class="SsBtn" :class="{ 'SsBtnPrimary': importSlot === 'secundary' }" :disabled="busy" @click="importSlot = 'secundary'">Secundaria</button>
                            </div>
                        </div>
                        <div class="FontManager_ImportRow">
                            <span class="FontManager_ImportLabel">Nombre</span>
                            <input class="SsIn" v-model="importName" placeholder="Nombre de la tipografía" />
                        </div>
                        <p class="FontManager_ImportHint">Ponle un nombre y elige el tipo: el bloque se crea en launcher_assets.json solo con estos datos.</p>
                        <div class="FontManager_ImportBtns">
                            <button class="SsBtn" :disabled="busy" @click="step = 'list'">Cancelar</button>
                            <button class="SsBtn SsBtnPrimary" :disabled="busy || !importName.trim()" @click="confirmImport">
                                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                                Importar
                            </button>
                        </div>
                    </div>
                </template>

                <p v-if="msg" class="FontManager_Msg">{{ msg }}</p>
            </div>
        </div>
    </Teleport>
</template>

<style scoped lang="scss">
@use '../Styles/Settings.scss';

.FontManager_Overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 100;
}

.FontManager_Dialog {
    width: 34rem;
    max-width: 92vw;
    max-height: 85vh;
    overflow-y: auto;
    background: var(--background-modal-primary);
    border: var(--border-modal-style);
    border-radius: 0.75rem;
    box-shadow: var(--shadow-settings-normal) #0008;
    padding: 1.25rem 1.35rem;
}

.FontManager_Head {
    display: flex;
    justify-content: space-between;
    align-items: center;

    h3 {
        margin: 0;
        font-family: var(--font-primary), Arial, sans-serif;
        font-size: calc(0.95rem * var(--font-size-primary, 1));
        font-weight: 600;
    }
}

.FontManager_Close {
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    border-radius: 0.35rem;
    display: flex;
    justify-content: center;
    align-items: center;
    height: 1.85rem;
    width: 1.85rem;
    transition: background 150ms, color 150ms;

    &:hover {
        background: #1111;
        color: var(--text-primary);
    }
}

.FontManager_Info {
    margin: 0.6rem 0 0;
    font-size: 0.78rem;
    line-height: 1.5;
    opacity: 0.55;

    strong {
        opacity: 0.85;
    }
}

.FontManager_Actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 0.9rem;
    flex-wrap: wrap;
}

.FontManager_List {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    margin-top: 1rem;
}

.FontManager_Item {
    background: var(--control-bg);
    border: 1px solid var(--control-border);
    border-radius: 0.5rem;
    padding: 0.7rem 0.85rem;
    display: flex;
    flex-direction: column;
    gap: 0.55rem;
}

.FontManager_ItemInfo {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
}

.FontManager_FileName {
    font-size: 0.8rem;
    font-weight: 500;
}

.FontManager_Badge {
    font-size: 0.62rem;
    padding: 0.15rem 0.45rem;
    border-radius: 0.3rem;
    background: color-mix(in srgb, var(--background-button-primary) 60%, gray 30%);
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
    text-transform: uppercase;
    letter-spacing: 0.05em;
}

.FontManager_Unassigned {
    font-size: 0.62rem;
    padding: 0.15rem 0.45rem;
    border-radius: 0.3rem;
    background: var(--control-bg);
    opacity: 0.45;
    text-transform: uppercase;
    letter-spacing: 0.05em;
}

.FontManager_ItemRow {
    display: flex;
    gap: 0.5rem;
    align-items: center;
    flex-wrap: wrap;

    .SsIn {
        flex: 1;
        min-width: 10rem;
    }
}

.FontManager_ItemBtns {
    display: flex;
    gap: 0.4rem;
    flex-shrink: 0;
}

.FontManager_Empty {
    margin-top: 1rem;
    text-align: center;
    opacity: 0.5;
    font-size: 0.78rem;
    padding: 0.5rem 0;
}

.FontManager_Import {
    margin-top: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    background: var(--control-bg);
    border: 1px solid var(--control-border);
    border-radius: 0.5rem;
    padding: 0.9rem 1rem;
}

.FontManager_ImportFile {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.8rem;
    font-weight: 500;
    word-break: break-all;

    svg {
        flex-shrink: 0;
        opacity: 0.6;
    }
}

.FontManager_ImportRow {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
}

.FontManager_ImportLabel {
    font-size: 0.72rem;
    opacity: 0.55;
    width: 5rem;
    flex-shrink: 0;
}

.FontManager_ImportSlots {
    display: flex;
    gap: 0.4rem;
}

.FontManager_ImportHint {
    margin: 0;
    font-size: 0.7rem;
    opacity: 0.5;
}

.FontManager_ImportBtns {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    margin-top: 0.25rem;
}

.FontManager_Msg {
    margin: 0.8rem 0 0;
    font-size: 0.75rem;
    opacity: 0.7;
}
</style>
