// Tarjeta de banner del panel: imagen panorámica del perfil (1920×1080
// exactos, PNG/GIF/JPG/WebP de hasta 8 MB). Las medidas se verifican en
// cliente antes de subir; la API tiene la última palabra.
<script setup lang="ts">
import { ref } from 'vue';
import { IconPhoto, IconAlertCircle, IconCircleCheck } from '@tabler/icons-vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { subirBanner } from '@/Auth/Api';
import { ApiError } from '@/Auth/Api';

const { profile, conAuth } = useAuth();

const previo = ref('');
const error = ref('');
const aviso = ref('');
const subiendo = ref(false);
const pendiente = ref<File | null>(null);

const TIPOS = ['image/png', 'image/gif', 'image/jpeg', 'image/webp'];

function dimensiones(archivo: File): Promise<{ w: number; h: number }> {
    return new Promise((resolve, reject) => {
        const url = URL.createObjectURL(archivo);
        const img = new Image();
        img.onload = () => {
            const dims = { w: img.naturalWidth, h: img.naturalHeight };
            URL.revokeObjectURL(url);
            resolve(dims);
        };
        img.onerror = () => {
            URL.revokeObjectURL(url);
            reject(new Error('no-imagen'));
        };
        img.src = url;
    });
}

async function elegir(evento: Event) {
    const input = evento.target as HTMLInputElement;
    const archivo = input.files && input.files[0] ? input.files[0] : null;
    error.value = '';
    aviso.value = '';
    if (previo.value) URL.revokeObjectURL(previo.value);
    previo.value = '';
    pendiente.value = null;
    if (!archivo) return;
    if (!TIPOS.includes(archivo.type)) {
        error.value = 'Solo se permiten imágenes PNG, GIF, JPEG o WebP.';
        return;
    }
    if (archivo.size > 8 * 1024 * 1024) {
        error.value = 'El banner no puede superar los 8 MB.';
        return;
    }
    try {
        const dims = await dimensiones(archivo);
        if (dims.w > 1920 || dims.h > 1080) {
            error.value = `El banner no puede superar 1920×1080 (el tuyo es de ${dims.w}×${dims.h}).`;
            return;
        }
    } catch {
        error.value = 'No se pudo leer esa imagen.';
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
        const res = await conAuth((token) => subirBanner(token, pendiente.value as File));
        if (profile.value) profile.value.bannerUrl = res.bannerUrl;
        pendiente.value = null;
        if (previo.value) URL.revokeObjectURL(previo.value);
        previo.value = '';
        aviso.value = 'Banner actualizado.';
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        subiendo.value = false;
    }
}
</script>

<template>
    <section class="Card" aria-label="Banner">
        <div class="Head">
            <div class="Tag">
                <IconPhoto stroke="2" />
                <span>Banner</span>
            </div>
        </div>
        <div class="Preview">
            <img v-if="previo || profile?.bannerUrl" :src="previo || profile?.bannerUrl || ''" alt="Banner de tu perfil">
            <span v-else class="Empty">Sin banner: tu cabecera usa el fondo clásico.</span>
        </div>
        <div class="Controls">
            <label class="GhostBtn" for="dash-banner">Elegir imagen</label>
            <input id="dash-banner" type="file" accept="image/png,image/gif,image/jpeg,image/webp" hidden @change="elegir">
            <button v-if="pendiente" type="button" class="BtnPrimary" :disabled="subiendo" @click="subir">
                {{ subiendo ? 'Subiendo...' : 'Subir banner' }}
            </button>
            <small class="Hint">Hasta 1920×1080, 8 MB.</small>
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
    .Preview{
        display: flex;
        justify-content: center;
        align-items: center;
        width: 100%;
        aspect-ratio: 16 / 9;
        border-radius: .6rem;
        overflow: hidden;
        border: 1px solid #ffffff18;
        background: #00000080;
        img{
            width: 100%;
            height: 100%;
            object-fit: cover;
        }
        .Empty{
            font-size: .75rem;
            opacity: .5;
            padding: 0 1rem;
            text-align: center;
        }
    }
    .Controls{
        display: flex;
        align-items: center;
        gap: .6rem;
        flex-wrap: wrap;
        .Hint{
            font-size: .7rem;
            opacity: .45;
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
