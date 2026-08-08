<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import { IconX } from '@tabler/icons-vue';
import { installedVersions, createProfile, updateProfile, type LauncherProfile } from '../Stores/Launcher';
import { GetConfig } from '@wailsjs/go/main/App';

const props = defineProps<{
    visible: boolean;
    editing?: LauncherProfile | null;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
    (e: 'done', ok: boolean, message: string): void;
}>();

const busy = ref(false);
const msg = ref('');
const msgOk = ref(true);

const form = reactive({
    name: '',
    version: '',
    resWidth: '',
    resHeight: '',
    fullscreen: false,
    gameDir: '',
    javaExec: '',
    javaArgs: '',
    icon: '',
});

const iconInput = ref<HTMLInputElement | null>(null);

function clearIcon() {
    form.icon = '';
}

function fileToDataUrl(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => {
            const img = new Image();
            img.onload = () => {
                const size = 96;
                const canvas = document.createElement('canvas');
                canvas.width = size;
                canvas.height = size;
                const ctx = canvas.getContext('2d');
                if (!ctx) return reject(new Error('no canvas'));
                ctx.imageSmoothingEnabled = true;
                ctx.clearRect(0, 0, size, size);
                ctx.drawImage(img, 0, 0, size, size);
                resolve(canvas.toDataURL('image/png'));
            };
            img.onerror = () => reject(new Error('decode'));
            img.src = reader.result as string;
        };
        reader.onerror = () => reject(new Error('read'));
        reader.readAsDataURL(file);
    });
}

async function onIconFile(e: Event) {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (!file) return;
    if (!/^image\/(png|jpeg|webp|gif)$/i.test(file.type)) {
        msg.value = 'Formato de imagen no soportado (usa PNG, JPG, WEBP o GIF).';
        msgOk.value = false;
        return;
    }
    if (file.size > 3 * 1024 * 1024) {
        msg.value = 'La imagen no debe superar los 3 MB.';
        msgOk.value = false;
        return;
    }
    try {
        form.icon = await fileToDataUrl(file);
        msg.value = '';
    } catch {
        msg.value = 'No se pudo leer la imagen.';
        msgOk.value = false;
    }
}

watch(
    () => props.visible,
    (v) => {
        if (!v) return;
        msg.value = '';
        busy.value = false;
        const e = props.editing;
        form.name = e?.name ?? '';
        form.version = e?.version ?? '';
        form.resWidth = e?.resWidth ? String(e.resWidth) : '';
        form.resHeight = e?.resHeight ? String(e.resHeight) : '';
        form.fullscreen = !!e?.fullscreen;
        form.gameDir = e?.gameDir ?? '';
        form.javaExec = e?.javaExec ?? '';
        form.javaArgs = e?.javaArgs ?? '';
        form.icon = e?.icon ?? '';
        if (!e) {
            prefillWindowDefaults();
        }
    }
);

async function prefillWindowDefaults() {
    try {
        const cfg = await GetConfig();
        const mc = cfg?.minecraftConfig;
        if (!mc) return;
        if (mc.windowWidth > 0) form.resWidth = String(mc.windowWidth);
        if (mc.windowHeight > 0) form.resHeight = String(mc.windowHeight);
        form.fullscreen = !!mc.fullscreen;
    } catch {
    }
}

function close() {
    emit('update:visible', false);
}

async function submit() {
    if (busy.value) return;
    busy.value = true;
    msg.value = '';

    if ((form.resWidth && !form.resHeight) || (!form.resWidth && form.resHeight)) {
        msg.value = 'Indica el ancho y el alto de la resolución, o déjalo vacío.';
        msgOk.value = false;
        busy.value = false;
        return;
    }

    const payload: LauncherProfile = {
        name: form.name.trim(),
        version: form.version || '',
        gameDir: form.gameDir.trim() || '',
        javaExec: form.javaExec.trim() || '',
        javaArgs: form.javaArgs.trim() || '',
        fullscreen: form.fullscreen,
        icon: form.icon || '',
        createdAt: props.editing?.createdAt ?? new Date().toISOString(),
        lastUsed: props.editing?.lastUsed,
    };

    if (form.resWidth && form.resHeight) {
        payload.resWidth = parseInt(form.resWidth, 10);
        payload.resHeight = parseInt(form.resHeight, 10);
        if (!(payload.resWidth > 0) || !(payload.resHeight > 0)) {
            msg.value = 'La resolución debe ser mayor que cero.';
            msgOk.value = false;
            busy.value = false;
            return;
        }
    }

    const err = props.editing
        ? await updateProfile(props.editing.name, payload)
        : await createProfile(payload);
    busy.value = false;
    if (err) {
        msg.value = err;
        msgOk.value = false;
        return;
    }
    emit('done', true, props.editing ? 'Perfil actualizado correctamente.' : 'Perfil creado correctamente.');
    close();
}
</script>

<template>
    <Teleport to="body">
        <Transition name="ProfileForm">
            <div v-if="visible" class="ProfileForm_Overlay" @click.self="close">
                <div class="ProfileForm_Dialog">
                    <div class="ProfileForm_Head">
                        <div class="ProfileForm_Titles">
                            <h3>{{ editing ? 'Editar perfil' : 'Nuevo perfil' }}</h3>
                            <span class="ProfileForm_Badge">Perfil de juego</span>
                        </div>
                        <button class="ProfileForm_Close" title="Cerrar" @click="close">
                            <IconX stroke="2" />
                        </button>
                    </div>

                    <div class="ProfileForm_Body">
                        <p class="ProfileForm_Desc">
                            Guarda una configuración propia de juego. Lo que dejes vacío se toma de la
                            configuración del launcher al pulsar Jugar.
                        </p>

                        <label class="ProfileForm_Field">
                            <span>Nombre del perfil</span>
                            <input class="SsIn" v-model="form.name" placeholder="Mi perfil" autocomplete="off" :disabled="!!editing" />
                        </label>
                        <label class="ProfileForm_Field">
                            <span>Versión</span>
                            <select class="SsIn" v-model="form.version">
                                <option value="">Cualquier versión (usa la del menú)</option>
                                <option v-for="v in installedVersions" :key="v.id" :value="v.id">{{ v.id }}</option>
                            </select>
                        </label>

                        <div class="ProfileForm_Icon">
                            <span class="ProfileForm_FieldName">Icono del perfil</span>
                            <div class="ProfileForm_IconPreview">
                                <span class="ProfileForm_Avatar">
                                    <img v-if="form.icon" :src="form.icon" alt="" loading="lazy">
                                    <span v-else>{{ form.name.trim().slice(0, 1).toUpperCase() || '?' }}</span>
                                </span>
                                <div class="ProfileForm_IconActions">
                                    <button class="SsBtn" @click="iconInput?.click()">
                                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                                        Importar imagen
                                    </button>
                                    <button v-if="form.icon" class="SsBtn SsBtnDanger" @click="clearIcon">Quitar</button>
                                </div>
                                <input ref="iconInput" type="file" accept="image/png,image/jpeg,image/webp,image/gif" class="ProfileForm_IconFile" @change="onIconFile">
                            </div>
                            <span class="ProfileForm_Hint">Sube tu propia imagen. Sin icono se usa la inicial del nombre.</span>
                        </div>

                        <div class="ProfileForm_Row">
                            <label class="ProfileForm_Field">
                                <span>Resolución — ancho</span>
                                <input class="SsIn" type="number" min="1" v-model="form.resWidth" placeholder="—" />
                            </label>
                            <label class="ProfileForm_Field">
                                <span>Resolución — alto</span>
                                <input class="SsIn" type="number" min="1" v-model="form.resHeight" placeholder="—" />
                            </label>
                        </div>

                        <label class="ProfileForm_Field ProfileForm_Cb">
                            <input type="checkbox" v-model="form.fullscreen" />
                            <span>Pantalla completa</span>
                        </label>

                        <label class="ProfileForm_Field">
                            <span>Carpeta del juego (opcional)</span>
                            <input class="SsIn" v-model="form.gameDir" placeholder="C:\Minecraft\Perfil1" autocomplete="off" />
                        </label>
                        <label class="ProfileForm_Field">
                            <span>Java (opcional)</span>
                            <input class="SsIn" v-model="form.javaExec" placeholder="Ruta de javaw.exe / java" autocomplete="off" />
                        </label>
                        <label class="ProfileForm_Field">
                            <span>Argumentos de Java (opcional)</span>
                            <input class="SsIn" v-model="form.javaArgs" placeholder="-XX:+UseG1GC -Xmx4G" autocomplete="off" />
                        </label>

                        <p class="ProfileForm_Note">
                            Los campos que no se rellenan heredan la configuración del launcher.
                        </p>

                        <p v-if="msg" :class="['ProfileForm_Msg', { error: !msgOk }]">{{ msg }}</p>
                    </div>

                    <div class="ProfileForm_Footer">
                        <button class="SsBtn" :disabled="busy" @click="close">Cancelar</button>
                        <button
                            class="SsBtn SsBtnPrimary"
                            :disabled="busy || !form.name.trim()"
                            @click="submit"
                        >{{ editing ? 'Guardar cambios' : 'Crear perfil' }}</button>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use '../Styles/Settings.scss' as *;

.ProfileForm_Overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: var(--filter-blur);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 115;
}

.ProfileForm_Dialog {
    width: 26rem;
    max-width: 92vw;
    max-height: 88vh;
    background: var(--background-modal-primary);
    border: var(--border-modal-style);
    border-radius: 0.85rem;
    box-shadow: var(--shadow-settings-normal) #000a;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.ProfileForm_Head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.75rem;
    padding: 1.1rem 1.35rem 0.85rem;
    border-bottom: var(--border-modal-style);
    background: #0005;
}

.ProfileForm_Titles {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    min-width: 0;

    h3 {
        margin: 0;
        font-family: var(--font-primary), Arial, sans-serif;
        font-size: calc(0.95rem * var(--font-size-primary, 1));
        font-weight: 600;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
}

.ProfileForm_Badge {
    font-size: 0.6rem;
    padding: 0.12rem 0.4rem;
    border-radius: 0.3rem;
    background: color-mix(in srgb, var(--background-button-primary) 15%, transparent);
    color: color-mix(in srgb, var(--background-button-primary) 55%, white 45%);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    white-space: nowrap;
    flex-shrink: 0;
}

.ProfileForm_Close {
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    border-radius: 0.4rem;
    display: flex;
    justify-content: center;
    align-items: center;
    height: 2rem;
    width: 2rem;
    flex-shrink: 0;
    transition: background 150ms, color 150ms;

    svg {
        width: 1.1rem;
        height: 1.1rem;
    }

    &:hover {
        background: #1111;
        color: var(--text-primary);
    }
}

.ProfileForm_Body {
    display: flex;
    flex-direction: column;
    gap: 0.8rem;
    padding: 1.1rem 1.35rem 0.25rem;
    overflow-y: auto;
}

.ProfileForm_Desc {
    margin: 0;
    font-size: 0.7rem;
    line-height: 1.45;
    opacity: 0.5;
}

.ProfileForm_Row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.8rem;
}

.ProfileForm_Field {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;

    > span {
        font-size: 0.7rem;
        opacity: 0.55;
    }

    .SsIn {
        width: 100%;
        box-sizing: border-box;
    }
}

.ProfileForm_Cb {
    flex-direction: row;
    align-items: center;
    gap: 0.5rem;

    input {
        accent-color: var(--background-button-primary);
        width: 0.95rem;
        height: 0.95rem;
        cursor: pointer;
    }

    > span {
        opacity: 0.6;
    }
}

.ProfileForm_Icon {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.ProfileForm_FieldName {
    font-size: 0.7rem;
    opacity: 0.55;
}

.ProfileForm_IconPreview {
    display: flex;
    align-items: center;
    gap: 0.6rem;
}

.ProfileForm_Avatar {
    width: 3rem;
    height: 3rem;
    flex-shrink: 0;
    border-radius: 0.6rem;
    border: var(--border-style);
    background: color-mix(in srgb, var(--background-button-primary) 60%, transparent);
    overflow: hidden;
    display: flex;
    justify-content: center;
    align-items: center;
    font-family: var(--font-primary), Arial, sans-serif;
    font-size: 1.15rem;
    font-weight: 600;
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);

    img {
        width: 100%;
        height: 100%;
        object-fit: cover;
        image-rendering: pixelated;
    }
}

.ProfileForm_IconActions {
    display: flex;
    gap: 0.4rem;
    flex-wrap: wrap;

    .SsBtn {
        padding: 0.4rem 0.7rem;
    }
}

.ProfileForm_IconFile {
    display: none;
}

.ProfileForm_Hint {
    font-size: 0.66rem;
    line-height: 1.45;
    opacity: 0.42;
}

.ProfileForm_Note {
    margin: 0;
    font-size: 0.66rem;
    line-height: 1.45;
    opacity: 0.42;
}

.ProfileForm_Msg {
    margin: 0;
    font-size: 0.72rem;
    color: color-mix(in srgb, var(--background-button-primary) 55%, white 45%);

    &.error {
        color: var(--color-error);
    }
}

.ProfileForm_Footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 0.85rem 1.35rem 1.1rem;
}

.ProfileForm-enter-active,
.ProfileForm-leave-active {
    transition: opacity 180ms ease;

    .ProfileForm_Dialog {
        transition: transform 200ms ease, opacity 180ms ease;
    }
}

.ProfileForm-enter-from,
.ProfileForm-leave-to {
    opacity: 0;

    .ProfileForm_Dialog {
        transform: translateY(8px) scale(0.97);
        opacity: 0;
    }
}
</style>