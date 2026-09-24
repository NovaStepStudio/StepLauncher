// Sección de cosméticos del panel: SOLO skins y capas de Minecraft con
// render 3D por card (SkinStage), modal ampliado, descarga y borrado.
// El listado de la API también trae avatar y banner (historial de subidas):
// acá se excluyen siempre. Paginado de 10 por página + filtro por tipo.
// La API no filtra en servidor, así que se juntan páginas en un caché local
// (con tope) sin traer todo de golpe: nunca hay más de 10 visores a la vez.
<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { IconUpload, IconChevronLeft, IconChevronRight, IconAlertCircle, IconCircleCheck } from '@tabler/icons-vue';
import SliderTabs from '@/Auth/Components/SliderTabs.vue';
import CosmeticCard from './CosmeticCard.vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { listarArchivos, subirArchivo, type FileUpload, type KindArchivo } from '@/Auth/Api';
import { ApiError } from '@/Auth/Api';
import { olvidarTodas } from '@/Auth/firmadas';

const POR_PAGINA = 10;
const RELLENO = 50;

type Filtro = 'all' | KindArchivo;

const filtros = [
    { key: 'all', label: 'Todos' },
    { key: 'skin', label: 'Skins' },
    { key: 'cape', label: 'Capas' },
] as const;

const { conAuth } = useAuth();

const filtro = ref<Filtro>('all');
const pagina = ref(1);
const enPagina = ref<FileUpload[]>([]);
const hayMas = ref(false);
const quota = ref('');
const cargando = ref(false);
const subiendo = ref<KindArchivo | null>(null);
const error = ref('');
const aviso = ref('');

// Caché en orden de servidor para el filtro por tipo (Map conserva inserción).
const cache = new Map<string, FileUpload>();
let siguienteOffset = 0;
let agotado = false;

// Solo skins y capas: avatar y banner viven en el historial pero se
// gestionan desde la cabecera, jamás como cosméticos.
function esCosmetico(a: FileUpload): boolean {
    return a.kind === 'skin' || a.kind === 'cape';
}

function visiblesSegunFiltro(): FileUpload[] {
    const todos = [...cache.values()].filter(esCosmetico);
    if (filtro.value === 'all') return todos;
    return todos.filter((a) => a.kind === filtro.value);
}

async function cargar(): Promise<void> {
    cargando.value = true;
    error.value = '';
    try {
        let juntados = visiblesSegunFiltro();
        let vueltas = 0;
        while (juntados.length < pagina.value * POR_PAGINA && !agotado && vueltas < 20) {
            vueltas += 1;
            const res = await conAuth((token) => listarArchivos(token, RELLENO, siguienteOffset));
            for (const u of res.uploads) cache.set(u.id, u);
            siguienteOffset += res.uploads.length;
            if (res.uploads.length < RELLENO) agotado = true;
            quota.value = `Límite diario: ${res.dailyQuota.limit} subidas`;
            juntados = visiblesSegunFiltro();
        }
        enPagina.value = juntados.slice((pagina.value - 1) * POR_PAGINA, pagina.value * POR_PAGINA);
        hayMas.value = juntados.length > pagina.value * POR_PAGINA || !agotado;
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        cargando.value = false;
    }
}

watch(filtro, () => {
    pagina.value = 1;
    siguienteOffset = 0;
    agotado = false;
    cargar();
});

watch(pagina, () => cargar());

function paginaSiguiente() {
    if (hayMas.value) pagina.value += 1;
}

function paginaAnterior() {
    if (pagina.value > 1) pagina.value -= 1;
}

async function elegir(kind: KindArchivo, evento: Event) {
    const input = evento.target as HTMLInputElement;
    const archivo = input.files && input.files[0] ? input.files[0] : null;
    input.value = '';
    error.value = '';
    aviso.value = '';
    if (!archivo) return;
    if (archivo.type !== 'image/png') {
        error.value = 'Solo se permiten archivos PNG.';
        return;
    }
    if (archivo.size > 1024 * 1024) {
        error.value = 'Cada archivo puede pesar como máximo 1 MB.';
        return;
    }
    subiendo.value = kind;
    try {
        await conAuth((token) => subirArchivo(token, kind, archivo));
        aviso.value = kind === 'skin' ? 'Skin subida.' : 'Capa subida.';
        cache.clear();
        olvidarTodas();
        siguienteOffset = 0;
        agotado = false;
        pagina.value = 1;
        await cargar();
    } catch (err) {
        error.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        subiendo.value = null;
    }
}

async function trasBorrado(id: string) {
    cache.delete(id);
    aviso.value = 'Cosmético borrado.';
    if (enPagina.value.length <= 1 && pagina.value > 1) {
        pagina.value -= 1;
    } else {
        await cargar();
    }
}

const vacio = computed(() => !cargando.value && enPagina.value.length === 0 && !error.value);

onMounted(() => cargar());
</script>

<template>
    <section class="Card" aria-label="Cosméticos">
        <small v-if="quota" class="Quota">{{ quota }}</small>
        <div class="Uploads">
            <label class="Drop" for="dash-skin">
                <IconUpload stroke="2" />
                <b>{{ subiendo === 'skin' ? 'Subiendo...' : 'Subir skin' }}</b>
                <small>PNG de hasta 1 MB</small>
            </label>
            <input id="dash-skin" type="file" accept="image/png" hidden @change="elegir('skin', $event)">
            <label class="Drop" for="dash-cape">
                <IconUpload stroke="2" />
                <b>{{ subiendo === 'cape' ? 'Subiendo...' : 'Subir capa' }}</b>
                <small>PNG de hasta 1 MB</small>
            </label>
            <input id="dash-cape" type="file" accept="image/png" hidden @change="elegir('cape', $event)">
        </div>
        <SliderTabs v-model="filtro" :tabs="filtros" />
        <p v-if="cargando" class="State">Cargando tus cosméticos...</p>
        <div v-else-if="vacio" class="Empty">
            <IconUpload stroke="2" />
            <b>Nada por acá todavía</b>
            <p>Subí tu primera skin o capa y aparece en 3D.</p>
        </div>
        <div v-else class="Grid sl-stagger">
            <CosmeticCard v-for="a in enPagina" :key="a.id" :item="a" @borrado="trasBorrado" />
        </div>
        <nav v-if="!cargando && (pagina > 1 || hayMas)" class="Pagination" aria-label="Paginar cosméticos">
            <button
                type="button"
                class="PageBtn"
                :disabled="pagina <= 1"
                aria-label="Página anterior"
                @click="paginaAnterior"
            >
                <IconChevronLeft stroke="2" />
            </button>
            <span class="PageInfo">Página {{ pagina }}</span>
            <button
                type="button"
                class="PageBtn"
                :disabled="!hayMas"
                aria-label="Página siguiente"
                @click="paginaSiguiente"
            >
                <IconChevronRight stroke="2" />
            </button>
        </nav>
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
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 1.1rem;
    padding: 1.5rem;
    border-radius: .9rem;
    border: 1px solid #ffffff18;
    background: #ffffff08;
    .Quota{
        position: absolute;
        top: 1.4rem;
        right: 1.5rem;
        font-size: .68rem;
        opacity: .4;
        white-space: nowrap;
    }
    .Uploads{
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: .6rem;
    }
    .Drop{
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        gap: .3rem;
        padding: 1.1rem .8rem;
        border-radius: .7rem;
        border: 1px dashed #ffffff2e;
        background: #00000060;
        color: #fff;
        cursor: pointer;
        text-align: center;
        transition: background 150ms, border-color 150ms, transform 150ms;
        svg{
            width: 1.3rem;
            height: 1.3rem;
            opacity: .7;
        }
        b{
            font-size: .8rem;
            font-family: 'Lexend';
        }
        small{
            font-size: .68rem;
            opacity: .45;
        }
        &:hover{
            background: #ffffff0d;
            border-color: #ffffff45;
            transform: translateY(-1px);
        }
    }
    .State{
        margin: 0;
        text-align: center;
        font-size: .8rem;
        opacity: .55;
    }
    .Empty{
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: .4rem;
        padding: 2.2rem 1.5rem;
        border-radius: .8rem;
        border: 1px dashed #ffffff25;
        background: #00000060;
        text-align: center;
        svg{
            width: 1.8rem;
            height: 1.8rem;
            opacity: .4;
        }
        b{
            font-family: 'Lexend';
            font-size: .9rem;
        }
        p{
            margin: 0;
            font-size: .78rem;
            opacity: .55;
        }
    }
    .Grid{
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
        gap: .7rem;
    }
    .Pagination{
        display: flex;
        justify-content: center;
        align-items: center;
        gap: .5rem;
        .PageBtn{
            display: flex;
            justify-content: center;
            align-items: center;
            min-width: 2.2rem;
            height: 2.2rem;
            padding: 0 .5rem;
            border-radius: .55rem;
            border: 1px solid #ffffff25;
            background: #ffffff08;
            color: #ffffffa6;
            cursor: pointer;
            transition: background 150ms, color 150ms, opacity 150ms;
            svg{
                width: 1rem;
                height: 1rem;
            }
            &:hover:not(:disabled){
                background: #ffffff14;
                color: #fff;
            }
            &:disabled{
                opacity: .3;
                cursor: default;
            }
        }
        .PageInfo{
            font-size: .75rem;
            font-family: 'Lexend';
            opacity: .6;
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
@media (max-width: 600px){
    .Card{
        min-width: 0;
        padding: 1rem;
        padding-top: 2.6rem;
        .Quota{
            top: 1rem;
            right: 1rem;
            left: 1rem;
            white-space: normal;
            text-align: right;
        }
        .Uploads{
            grid-template-columns: minmax(0, 1fr);
        }
        .Drop{
            min-height: 2.75rem;
        }
        .Grid{
            grid-template-columns: minmax(0, 1fr);
        }
    }
}
</style>
