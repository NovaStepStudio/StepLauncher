<script setup lang="ts">
import { ref, reactive, watch } from 'vue';
import { IconX } from '@tabler/icons-vue';
import { installedVersions, createProfile, updateProfile, type LauncherProfile } from '@/Launcher/Store';
import { GetConfig } from '@wailsjs/go/main/App';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

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
    },
    { immediate: true }
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

useOverlayEscape(close, { isActive: () => props.visible, priority: 2 });

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
@use './Styles/ProfileForm.scss';
</style>