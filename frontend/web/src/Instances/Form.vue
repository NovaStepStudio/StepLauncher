<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { IconX, IconPhotoPlus } from '@tabler/icons-vue';
import { createInstance, updateMetadata, detailOf } from './Store';
import { loadLocal } from '@/Common/Stores/Ui';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

const goApp = () => (window as any)?.go?.main?.App;

const props = defineProps<{
    visible: boolean;
    editing: string | null;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
    (e: 'done', ok: boolean, message: string): void;
    (e: 'created', name: string): void;
}>();

const isEdit = computed(() => !!props.editing);

const busy = ref(false);
const msg = ref('');
const msgOk = ref(true);

const name = ref('');
const title = ref('');
const description = ref('');
const favorite = ref(false);

const iconUrl = ref('');
const bannerUrl = ref('');

watch(
    () => props.visible,
    (v) => {
        if (!v) return;
        name.value = '';
        title.value = '';
        description.value = '';
        favorite.value = false;
        iconUrl.value = '';
        bannerUrl.value = '';
        msg.value = '';
        msgOk.value = true;
        if (isEdit.value) {
            const d = detailOf(props.editing as string);
            title.value = d?.meta.title ?? '';
            description.value = d?.meta.description ?? '';
            favorite.value = !!d?.meta.favorite;
            if (d?.meta.icon) void loadLocal(d.meta.icon).then((u) => (iconUrl.value = u));
            if (d?.meta.banner) void loadLocal(d.meta.banner).then((u) => (bannerUrl.value = u));
        }
    },
    { immediate: true }
);

function cleanName(value: string): boolean {
    if (!value.trim()) {
        msg.value = 'Escribe un nombre para la instancia.';
        msgOk.value = false;
        return false;
    }
    if (value.includes('..') || /[\\/]/.test(value)) {
        msg.value = 'El nombre no puede contener separadores de ruta ni "..".';
        msgOk.value = false;
        return false;
    }
    return true;
}

async function submit() {
    if (busy.value) return;
    busy.value = true;
    msg.value = '';
    msgOk.value = true;
    if (isEdit.value && props.editing) {
        const err = await updateMetadata(props.editing, {
            title: title.value.trim(),
            description: description.value.trim(),
            favorite: favorite.value,
        });
        busy.value = false;
        if (err) {
            msg.value = err;
            msgOk.value = false;
            return;
        }
        emit('done', true, 'Instancia actualizada.');
        emit('update:visible', false);
        return;
    }
    if (!cleanName(name.value)) {
        busy.value = false;
        return;
    }
    const finalName = name.value.trim();
    const err = await createInstance({
        name: finalName,
        title: title.value.trim() || undefined,
        description: description.value.trim() || undefined,
        favorite: favorite.value || undefined,
    });
    busy.value = false;
    if (err) {
        msg.value = err;
        msgOk.value = false;
        return;
    }
    emit('done', true, `Instancia "${finalName}" creada.`);
    emit('created', finalName);
    emit('update:visible', false);
}

async function pickAsset(kind: 'icon' | 'banner') {
    if (busy.value || !props.editing) return;
    busy.value = true;
    msg.value = '';
    msgOk.value = true;
    try {
        const picked = await goApp()?.PickInstanceAssetFile?.();
        if (!picked) {
            busy.value = false;
            return;
        }
        const rel = await goApp()?.ImportInstanceAsset?.(props.editing, kind, picked);
        if (!rel) {
            msg.value = 'No se pudo importar la imagen.';
            msgOk.value = false;
            busy.value = false;
            return;
        }
        const err = await updateMetadata(props.editing, { [kind]: rel });
        busy.value = false;
        if (err) {
            msg.value = err;
            msgOk.value = false;
            return;
        }
        const url = await loadLocal(rel);
        if (kind === 'icon') iconUrl.value = url;
        else bannerUrl.value = url;
    } catch (e: any) {
        busy.value = false;
        msg.value = e?.message ?? 'No se pudo importar la imagen.';
        msgOk.value = false;
    }
}

function close() {
    emit('update:visible', false);
}

useOverlayEscape(close, { isActive: () => props.visible, priority: 2 });
</script>

<template>
    <Teleport to="body">
        <Transition name="InstForm">
            <div v-if="visible" class="InstForm_Overlay" @click.self="close">
                <div class="InstForm_Dialog">
                    <div class="InstForm_Head">
                        <div class="InstForm_Titles">
                            <h3>{{ isEdit ? 'Personalizar instancia' : 'Nueva instancia' }}</h3>
                            <span v-if="isEdit && editing" class="InstForm_Name">{{ editing }}</span>
                            <span v-else class="InstForm_Sub">Después de crearla elegirás la versión y el modloader</span>
                        </div>
                        <button class="InstForm_Close" title="Cerrar" @click="close">
                            <IconX stroke="2" />
                        </button>
                    </div>

                    <div class="InstForm_Body">
                        <template v-if="!isEdit">
                            <label class="InstForm_Field">
                                <span>Nombre <em class="req">*</em></span>
                                <input class="SsIn" v-model="name" placeholder="Mi mundo 1.20" autocomplete="off" />
                            </label>
                        </template>

                        <label class="InstForm_Field">
                            <span>Título</span>
                            <input class="SsIn" v-model="title" placeholder="Nombre visible en la tarjeta" autocomplete="off" />
                        </label>

                        <label class="InstForm_Field">
                            <span>Descripción</span>
                            <textarea class="SsIn InstForm_Textarea" v-model="description" rows="3" placeholder="¿De qué va esta instancia?" />
                        </label>

                        <template v-if="isEdit">
                            <div class="InstForm_Assets">
                                <span class="InstForm_AssetsTitle">Imágenes de la instancia</span>
                                <div class="InstForm_AssetsRow">
                                    <div class="InstForm_Asset">
                                        <span class="InstForm_AssetLabel">Icono</span>
                                        <div class="InstForm_AssetBox" :class="{ has: !!iconUrl }">
                                            <img v-if="iconUrl" :src="iconUrl" alt="" />
                                            <IconPhotoPlus v-else stroke="1.5" />
                                        </div>
                                        <button class="SsBtn" :disabled="busy" @click="pickAsset('icon')">Cambiar</button>
                                    </div>
                                    <div class="InstForm_Asset">
                                        <span class="InstForm_AssetLabel">Banner</span>
                                        <div class="InstForm_AssetBox InstForm_AssetBoxWide" :class="{ has: !!bannerUrl }">
                                            <img v-if="bannerUrl" :src="bannerUrl" alt="" />
                                            <IconPhotoPlus v-else stroke="1.5" />
                                        </div>
                                        <button class="SsBtn" :disabled="busy" @click="pickAsset('banner')">Cambiar</button>
                                    </div>
                                </div>
                            </div>
                            <p class="InstForm_Note">El icono y el banner se guardan dentro de la carpeta de la instancia y se borran con ella.</p>
                            <p v-if="!editing" class="InstForm_Note">Necesitas una instancia creada para importar imágenes.</p>
                        </template>

                        <div class="InstForm_SwRow">
                            <span class="InstForm_SwText">Marcar como favorita</span>
                            <label class="SsTg">
                                <input type="checkbox" v-model="favorite" />
                                <span class="SsTgS"></span>
                            </label>
                        </div>

                        <p v-if="msg" :class="['InstForm_Msg', { error: !msgOk }]">{{ msg }}</p>
                    </div>

                    <div class="InstForm_Footer">
                        <button class="SsBtn" :disabled="busy" @click="close">Cancelar</button>
                        <button class="SsBtn SsBtnPrimary InstForm_Submit" :disabled="busy || (!isEdit && !name.trim())" @click="submit">
                            {{ busy ? 'Guardando…' : (isEdit ? 'Guardar cambios' : 'Crear instancia') }}
                        </button>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use './Styles/Form.scss';
</style>