<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue';

const emite = defineEmits<{ cerrar: [] }>();

// El vídeo manda: 128x72 a 60 fps con el audio adentro. El canvas pinta
// el fotograma actual del vídeo, así animación y audio comparten el
// mismo reloj y es imposible que se desincronicen (sin matemática de
// tiempo ni archivos de frames por separado).
const COLS = 128;
const ROWS = 72;
// Umbral de gris: igual que website/python/convert.py (>153 fondo).
const UMBRAL = 153;
const VIDEO_URL = '/bad_apple.mp4';

const contenedor = ref<HTMLElement | null>(null);
const lienzo = ref<HTMLCanvasElement | null>(null);
const reproductor = ref<HTMLVideoElement | null>(null);
const velo = ref<HTMLElement | null>(null);

let muestra: HTMLCanvasElement | null = null;
let ctxMuestra: CanvasRenderingContext2D | null = null;
let ctx: CanvasRenderingContext2D | null = null;
let puntoX: Float32Array | null = null;
let puntoY: Float32Array | null = null;
let radioPunto = 3;
let vistaAncho = 0;
let vistaAlto = 0;
// Rango visible de la grilla (modo cover: lo que se recorta no se dibuja).
let filaIni = 0;
let filaFin = ROWS;
let colIni = 0;
let colFin = COLS;

let raf = 0;
let ultimoTiempo = -1;
let repintar = false;
let arrancado = false;
let fallo = false;
const conFallo = ref(false);

let observador: ResizeObserver | null = null;

// Prepara el canvas a la medida real de la pantalla (DPR capado en 2:
// más resolución no se nota en puntos y funde la GPU del celu) y
// precalcula el centro de cada celda para no hacerlo por fotograma.
function prepararLienzo(): void {
    const zona = contenedor.value;
    const canvas = lienzo.value;
    if (!zona || !canvas) return;

    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    const ancho = zona.offsetWidth;
    const alto = zona.offsetHeight;

    vistaAncho = ancho;
    vistaAlto = alto;
    canvas.width = Math.floor(ancho * dpr);
    canvas.height = Math.floor(alto * dpr);
    canvas.style.width = `${ancho}px`;
    canvas.style.height = `${alto}px`;

    ctx = canvas.getContext('2d');
    if (!ctx) return;
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

    // Modo cover (como object-fit: cover): la grilla 16:9 llena toda la
    // pantalla sin estirarse; lo que sobra se recorta (en vertical se
    // recorta arriba/abajo). Los puntos siempre quedan circulares.
    const escala = Math.max(ancho / COLS, alto / ROWS);
    const despX = (ancho - COLS * escala) / 2;
    const despY = (alto - ROWS * escala) / 2;
    radioPunto = Math.max(0.6, escala * 0.42);

    puntoX = new Float32Array(COLS * ROWS);
    puntoY = new Float32Array(COLS * ROWS);
    for (let i = 0; i < COLS * ROWS; i += 1) {
        puntoX[i] = despX + ((i % COLS) + 0.5) * escala;
        puntoY[i] = despY + (((i / COLS) | 0) + 0.5) * escala;
    }
    // Solo las celdas que caen dentro de la pantalla se dibujan: en un
    // celu vertical se ahorra ~70% de los arcos por fotograma.
    colIni = Math.max(0, Math.floor(-despX / escala));
    colFin = Math.min(COLS, Math.ceil((ancho - despX) / escala));
    filaIni = Math.max(0, Math.floor(-despY / escala));
    filaFin = Math.min(ROWS, Math.ceil((alto - despY) / escala));
    // El próximo fotograma se dibuja con las posiciones nuevas.
    repintar = true;
}

// Dibuja el fotograma actual del vídeo como puntos blancos.
function pintar(): void {
    const video = reproductor.value;
    if (!video || !ctx || !ctxMuestra || !puntoX || !puntoY) return;
    if (video.readyState < 2) return;

    ctxMuestra.drawImage(video, 0, 0, COLS, ROWS);
    const datos = ctxMuestra.getImageData(0, 0, COLS, ROWS).data;

    ctx.clearRect(0, 0, vistaAncho, vistaAlto);
    const xs = puntoX;
    const ys = puntoY;
    const r = radioPunto;

    ctx.fillStyle = '#fff';
    ctx.beginPath();
    for (let fila = filaIni; fila < filaFin; fila += 1) {
        const base = fila * COLS;
        for (let col = colIni; col < colFin; col += 1) {
            const i = base + col;
            const j = i * 4;
            const lum = (datos[j]! * 299 + datos[j + 1]! * 587 + datos[j + 2]! * 114) / 1000;
            if (lum <= UMBRAL) {
                ctx.moveTo(xs[i]! + r, ys[i]!);
                ctx.arc(xs[i]!, ys[i]!, r, 0, Math.PI * 2);
            }
        }
    }
    ctx.fill();
}

function detenerBucle(): void {
    if (raf !== 0) {
        cancelAnimationFrame(raf);
        raf = 0;
    }
}

function bucle(): void {
    raf = requestAnimationFrame(bucle);
    const video = reproductor.value;
    if (!video || video.paused) return;
    // Solo se rasteriza cuando el vídeo avanzó de fotograma.
    if (video.currentTime === ultimoTiempo && !repintar) return;
    ultimoTiempo = video.currentTime;
    repintar = false;
    pintar();
}

// El primer toque es el gesto que los navegadores exigen para
// reproducir con sonido: arranca vídeo + audio juntos, del mismo reloj.
async function alPrimerClic(): Promise<void> {
    const video = reproductor.value;
    if (!video) return;
    try {
        await video.play();
        arrancado = true;
        if (velo.value) {
            velo.value.style.display = 'none';
        }
    } catch {
        // Bloqueado: el próximo toque lo reintenta.
    }
}

function alReproducir(): void {
    ultimoTiempo = -1;
    if (raf === 0) {
        raf = requestAnimationFrame(bucle);
    }
}

function alPausar(): void {
    detenerBucle();
}

function alError(): void {
    fallo = true;
    conFallo.value = true;
    if (import.meta.env.DEV) {
        console.warn('[BadApple] no se pudo cargar /bad_apple.mp4.');
    }
}

function alCambiarVisibilidad(): void {
    const video = reproductor.value;
    if (!video || !arrancado) return;
    if (document.hidden) {
        video.pause();
    } else {
        void video.play().catch(() => undefined);
    }
}

function alEscape(evento: KeyboardEvent): void {
    if (evento.key === 'Escape') {
        emite('cerrar');
    }
}

onMounted(() => {
    muestra = document.createElement('canvas');
    muestra.width = COLS;
    muestra.height = ROWS;
    ctxMuestra = muestra.getContext('2d', { willReadFrequently: true });

    prepararLienzo();
    const zona = contenedor.value;
    if (zona && typeof ResizeObserver !== 'undefined') {
        observador = new ResizeObserver(prepararLienzo);
        observador.observe(zona);
    }
    document.addEventListener('visibilitychange', alCambiarVisibilidad);
    window.addEventListener('keydown', alEscape);
});

onBeforeUnmount(() => {
    detenerBucle();
    document.removeEventListener('visibilitychange', alCambiarVisibilidad);
    window.removeEventListener('keydown', alEscape);
    observador?.disconnect();
    reproductor.value?.pause();
});
</script>

<template>
    <div ref="contenedor" class="BadApple">
        <canvas ref="lienzo"></canvas>
        <video
            ref="reproductor"
            :src="VIDEO_URL"
            preload="auto"
            loop
            playsinline
            @play="alReproducir"
            @pause="alPausar"
            @error="alError"
        ></video>
        <div v-if="!conFallo" ref="velo" class="Velo" @click="alPrimerClic">
            <span class="Icono">▶</span>
            <span class="Texto">Hacé clic para reproducir</span>
        </div>
        <p v-else class="Fallo">No se pudo cargar el easter egg.</p>
        <button type="button" class="Cerrar" aria-label="Cerrar" title="Cerrar" @click="emite('cerrar')">✕</button>
    </div>
</template>

<style scoped lang="scss">
.BadApple{
    position: fixed;
    inset: 0;
    z-index: 99999;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    background: #000;
    cursor: pointer;
    // Táctil: sin zoom por doble toque, sin resaltado gris ni rebote.
    touch-action: manipulation;
    overscroll-behavior: none;
    -webkit-tap-highlight-color: transparent;
    user-select: none;
    -webkit-user-select: none;
    canvas{
        display: block;
    }
    video{
        position: absolute;
        width: 1px;
        height: 1px;
        opacity: 0;
        pointer-events: none;
    }
    .Velo{
        position: absolute;
        inset: 0;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: .5rem;
        z-index: 1;
        background: #00000099;
        .Icono{
            font-size: 3rem;
            color: #fff;
            opacity: .8;
        }
        .Texto{
            font-size: .9rem;
            color: #ffffff99;
        }
    }
    .Fallo{
        position: absolute;
        margin: 0;
        padding: 0 1rem;
        text-align: center;
        font-size: .85rem;
        color: #ffffffb3;
    }
    .Cerrar{
        position: fixed;
        top: calc(.75rem + env(safe-area-inset-top, 0px));
        right: calc(.75rem + env(safe-area-inset-right, 0px));
        z-index: 2;
        // Área táctil amplia (44px) con visual discreto.
        display: flex;
        align-items: center;
        justify-content: center;
        min-width: 2.75rem;
        min-height: 2.75rem;
        padding: .3rem .6rem;
        border-radius: .4rem;
        border: 1px solid #ffffff1a;
        background: transparent;
        color: #ffffff66;
        font-size: .8rem;
        cursor: pointer;
        transition: background 150ms, color 150ms;
        &:hover{
            background: #ffffff14;
            color: #fff;
        }
        &:focus-visible{
            outline: 2px solid #fff;
            outline-offset: 2px;
        }
    }
}
</style>
