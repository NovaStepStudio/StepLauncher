// Tarjeta de avatar del panel: muestra el actual (o la inicial del usuario)
// y permite subir uno nuevo (PNG/JPEG/WebP de hasta 2 MB).
<script setup lang="ts">
import { computed, ref } from 'vue';
import { IconPhoto, IconUpload, IconAlertCircle, IconCircleCheck } from '@tabler/icons-vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { subirAvatar } from '@/Auth/Api';
import { ApiError } from '@/Auth/Api';

const { profile, nombre, conAuth } = useAuth();

const previo = ref('');
const error = ref('');
const aviso = ref('');
const subiendo = ref(false);
const pendiente = ref<File | null>(null);

const inicial = computed(() => (nombre.value.trim().charAt(0) || '?').toUpperCase());

function elegir(evento: Event) {
    const input = evento.target as HTMLInputElement;
    const archivo = input.files && input.files[0] ? input.files[0] : null;
    error.value = '';
    aviso.value = '';
    if (previo.value) URL.revokeObjectURL(previo.value);
    previo.value = '';
    pendiente.value = null;
    if (!archivo) return;
    const esImagen = archivo.type === 'image/png' || archivo.type === 'image/jpeg' || archivo.type === 'image/webp';
    if (!esImagen) {
        error.value = 'Solo se permiten imágenes PNG, JPEG o WebP.';
        return;
    }
    if (archivo.size > 2 * 1024 * 1024) {
        error.value = 'La imagen no puede superar los 2 MB.';
        return;
    }
    pendiente.value = archivo;
    previo.value = URL.createObjectURL(archivo);
}

async function subir() {
    if (!pendiente.value) return;
    error.value = '';
    aviso.value = '';
    subiendo.value = true;
    try {
        const res = await conAuth((token) => subirAvatar(token, pendiente.value as File));
        if (profile.value) profile.value.avatarUrl = res.avatarUrl;
        pendiente.value = null;
        if (previo.value) URL.revokeObjectURL(previo.value);
        previo.value = '';
        aviso.value = 'Avatar actualizado.';
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        subiendo.value = false;
    }
}
</script>

<template>
    <section class="Card" aria-label="Avatar">
        <div class="Head">
            <div class="Tag">
                <IconPhoto stroke="2" />
                <span>Avatar</span>
            </div>
        </div>
        <div class="Body">
            <div class="Preview">
                <img v-if="previo || profile?.avatarUrl" :src="previo || profile?.avatarUrl || ''" alt="Avatar de tu cuenta">
                <span v-else class="Initial">{{ inicial }}</span>
            </div>
            <div class="Controls">
                <label class="GhostBtn" for="dash-avatar">Elegir imagen</label>
                <input id="dash-avatar" type="file" accept="image/png,image/jpeg,image/webp" hidden @change="elegir">
                <button v-if="pendiente" type="button" class="BtnPrimary" :disabled="subiendo" @click="subir">
                    <IconUpload stroke="2" />
                    {{ subiendo ? 'Subiendo...' : 'Subir avatar' }}
                </button>
                <small class="Hint">PNG, JPEG o WebP de hasta 2 MB.</small>
            </div>
        </div>
        <p v-if="error" class="FormError" role="alert">
            <IconAlertCircle stroke="2" />
            {{ error }}
        </p>
        <p v-if="aviso" class="Ok" role="status">
            <IconCircleCheck stroke="2" />
            {{ aviso }}
        </p>
    </section>
</template>

<style scoped lang="scss">
.Card{
    display: flex;
    flex-direction: column;
    gap: 1.1rem;
    padding: 1.5rem;
    border-radius: .9rem;
    border: 1px solid #ffffff18;
    background: #ffffff08;
    .Head{
        display: flex;
        justify-content: space-between;
        align-items: center;
        .Tag{
            display: flex;
            justify-content: center;
            align-items: center;
            gap: .4rem;
            font-size: .7rem;
            text-transform: uppercase;
            letter-spacing: .12em;
            padding: .25rem .65rem;
            border-radius: 99rem;
            border: 1px solid #ffffff25;
            background: #ffffff0d;
            font-family: 'Lexend';
            svg{
                width: .95rem;
                height: .95rem;
            }
        }
    }
    .Body{
        display: flex;
        align-items: center;
        gap: 1.2rem;
        .Preview{
            display: flex;
            justify-content: center;
            align-items: center;
            flex-shrink: 0;
            width: 5.5rem;
            height: 5.5rem;
            border-radius: 99rem;
            overflow: hidden;
            border: 1px solid #ffffff25;
            background: #ffffff10;
            img{
                width: 100%;
                height: 100%;
                object-fit: cover;
            }
            .Initial{
                font-family: 'Lexend';
                font-size: 2rem;
                font-weight: 700;
                opacity: .8;
            }
        }
        .Controls{
            display: flex;
            flex-direction: column;
            align-items: flex-start;
            gap: .6rem;
            .Hint{
                font-size: .7rem;
                opacity: .45;
            }
        }
    }
    .GhostBtn{
        display: flex;
        justify-content: center;
        align-items: center;
        gap: .4rem;
        padding: .5rem 1rem;
        border-radius: .55rem;
        border: 1px solid #ffffff25;
        background: #ffffff08;
        color: #fff;
        font-size: .78rem;
        font-weight: 600;
        font-family: inherit;
        cursor: pointer;
        transition: background 150ms;
        &:hover{
            background: #ffffff14;
        }
    }
    .BtnPrimary{
        display: flex;
        justify-content: center;
        align-items: center;
        gap: .4rem;
        padding: .5rem 1.1rem;
        border-radius: .55rem;
        border: 1px solid #fff;
        background: #fff;
        color: #000;
        font-size: .78rem;
        font-weight: 700;
        font-family: inherit;
        cursor: pointer;
        transition: filter 150ms, opacity 150ms;
        svg{
            width: 1rem;
            height: 1rem;
        }
        &:hover:not(:disabled){
            filter: brightness(.85);
        }
        &:disabled{
            opacity: .6;
            cursor: wait;
        }
    }
    .FormError{
        display: flex;
        align-items: center;
        gap: .5rem;
        margin: 0;
        padding: .65rem .85rem;
        border-radius: .55rem;
        border: 1px solid #ff9d9d45;
        background: #ff6b6b14;
        font-size: .8rem;
        line-height: 1.5;
        svg{
            flex-shrink: 0;
            width: 1.1rem;
            height: 1.1rem;
            color: #ff9d9d;
        }
    }
    .Ok{
        display: flex;
        align-items: center;
        gap: .5rem;
        margin: 0;
        font-size: .8rem;
        color: #9dffb0;
        svg{
            width: 1.1rem;
            height: 1.1rem;
        }
    }
}
</style>
