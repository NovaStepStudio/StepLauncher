// Escenario 3D de un cosmético con skinview3d (librería existente del
// ecosistema, sin renderer propio). El visor se crea solo cuando entra en
// viewport y se pausa al salir: así 10 cards no funden la GPU a la vez.
// Las capas se muestran sobre un maniquí neutro generado en local
// (data-URL, sin CORS ni dependencias externas).
<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { SkinViewer } from 'skinview3d';

const props = withDefaults(
    defineProps<{
        skin?: string | null;
        cape?: string | null;
        ancho?: number;
        alto?: number;
        girar?: boolean;
        zoom?: boolean;
    }>(),
    { skin: null, cape: null, ancho: 220, alto: 250, girar: false, zoom: false },
);

// Maniquí gris de 64x64 para lucir capas sin skin del usuario.
let maniqui: string | null = null;

function urlManiqui(): string {
    if (maniqui) return maniqui;
    const lienzo = document.createElement('canvas');
    lienzo.width = 64;
    lienzo.height = 64;
    const ctx = lienzo.getContext('2d');
    if (ctx) {
        ctx.fillStyle = '#b9b9b9';
        ctx.fillRect(0, 0, 64, 64);
        maniqui = lienzo.toDataURL('image/png');
    }
    return maniqui ?? '';
}

const lienzo = ref<HTMLCanvasElement | null>(null);
const listo = ref(false);
const fallo = ref(false);

let visor: SkinViewer | null = null;
let observador: IntersectionObserver | null = null;

async function crear() {
    if (visor || !lienzo.value) return;
    try {
        visor = new SkinViewer({
            canvas: lienzo.value,
            width: props.ancho,
            height: props.alto,
            model: 'auto-detect',
        });
        visor.autoRotate = props.girar;
        visor.autoRotateSpeed = 0.8;
        visor.controls.enableZoom = props.zoom;
        visor.controls.enablePan = false;
        const piel = props.skin || (props.cape ? urlManiqui() : '');
        if (piel) await visor.loadSkin(piel);
        if (props.cape) await visor.loadCape(props.cape);
        if (!visor.disposed) {
            visor.renderPaused = false;
            listo.value = true;
        }
    } catch {
        destruir();
        fallo.value = true;
    }
}

function destruir() {
    observador?.disconnect();
    observador = null;
    if (visor && !visor.disposed) visor.dispose();
    visor = null;
}

onMounted(() => {
    if (typeof IntersectionObserver === 'undefined') {
        crear();
        return;
    }
    observador = new IntersectionObserver(
        (entradas) => {
            for (const entrada of entradas) {
                if (entrada.isIntersecting) {
                    if (!visor && !fallo.value) crear();
                    else if (visor && !visor.disposed) visor.renderPaused = false;
                } else if (visor && !visor.disposed) {
                    visor.renderPaused = true;
                }
            }
        },
        { rootMargin: '100px' },
    );
    if (lienzo.value) observador.observe(lienzo.value);
});

onBeforeUnmount(() => destruir());
</script>

<template>
    <div class="Stage" :style="{ width: `${ancho}px`, height: `${alto}px` }">
        <canvas ref="lienzo" :width="ancho" :height="alto"></canvas>
        <span v-if="!listo && !fallo" class="Cargando">Cargando 3D...</span>
        <span v-if="fallo" class="Fallo">No se pudo cargar la vista 3D.</span>
    </div>
</template>

<style scoped lang="scss">
.Stage{
    position: relative;
    display: flex;
    justify-content: center;
    align-items: center;
    max-width: 100%;
    border-radius: .6rem;
    background: radial-gradient(circle at 50% 35%, #ffffff14, transparent 70%);
    overflow: hidden;
    canvas{
        max-width: 100%;
        height: auto;
    }
    .Cargando,
    .Fallo{
        position: absolute;
        inset: auto 0 0.6rem 0;
        text-align: center;
        font-size: .7rem;
        opacity: .5;
        pointer-events: none;
    }
}
</style>
