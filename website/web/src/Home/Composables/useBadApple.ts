import { onMounted, onUnmounted, ref, watch } from 'vue';
import { useRoute, useRouter, type LocationQuery } from 'vue-router';

// ============================================================================
// Composable useBadApple — Easter egg Bad Apple en el HOME.
// Se abre de tres formas: con ?apple (o ?badapple) en la URL, con el
// código Konami (↑↑↓↓←→←→BA) o con 5 clics seguidos en el logo del Hero.
// Vive montado solo en Home/Index para no interferir con el resto de la web.
// ============================================================================

// Secuencia Konami en formato KeyboardEvent.key (minúsculas).
const KONAMI = [
    'arrowup',
    'arrowup',
    'arrowdown',
    'arrowdown',
    'arrowleft',
    'arrowright',
    'arrowleft',
    'arrowright',
    'b',
    'a',
];

// Clics en el logo necesarios y ventana de tiempo para contarlos.
const CLICS_NECESARIOS = 5;
const VENTANA_CLICS_MS = 3000;

// Archivo del huevo: se precarga para que la apertura sea instantánea.
const URL_VIDEO = '/bad_apple.mp4';

// Precarga diferida en caché HTTP: cuando el navegador está ocioso y sin
// apuro. Se salta con ahorro de datos o red 2G para no gastar MB de más.
function prefetch(): void {
    try {
        const red = (navigator as Navigator & { connection?: { saveData?: boolean; effectiveType?: string } }).connection;
        if (red?.saveData) return;
        if (red?.effectiveType === 'slow-2g' || red?.effectiveType === '2g') return;
    } catch {
        // Sin Network Information: se precarga igual.
    }
    const pedir = () => {
        void fetch(URL_VIDEO, { credentials: 'same-origin' }).catch(() => undefined);
    };
    if (typeof window.requestIdleCallback === 'function') {
        window.requestIdleCallback(pedir, { timeout: 8000 });
    } else {
        window.setTimeout(pedir, 4000);
    }
}

const abierto = ref(false);

function queryPideApple(query: LocationQuery): boolean {
    return 'apple' in query || 'badapple' in query;
}

export function useBadApple() {
    const route = useRoute();
    const router = useRouter();
    let progresoKonami = 0;
    let clics = 0;
    let ultimoClic = 0;

    function abrir(): void {
        if (abierto.value) return;
        abierto.value = true;
    }

    function cerrar(): void {
        if (!abierto.value) return;
        abierto.value = false;
        // Limpia el trigger de la URL sin recargar ni ensuciar el historial.
        if (queryPideApple(route.query)) {
            const query = { ...route.query };
            delete query.apple;
            delete query.badapple;
            void router.replace({ query });
        }
    }

    // Cuenta clics en el logo: 5 seguidos (ventana de 3 s) abren el huevo.
    function clicsLogo(): void {
        const ahora = Date.now();
        if (ahora - ultimoClic > VENTANA_CLICS_MS) {
            clics = 0;
        }
        ultimoClic = ahora;
        clics += 1;
        if (clics >= CLICS_NECESARIOS) {
            clics = 0;
            abrir();
        }
    }

    function alPulsarTecla(evento: KeyboardEvent): void {
        if (abierto.value) return;
        const objetivo = evento.target as HTMLElement | null;
        // No robar teclas mientras se escribe en un campo o hay un diálogo.
        if (objetivo && (objetivo.tagName === 'INPUT' || objetivo.tagName === 'TEXTAREA' || objetivo.isContentEditable)) {
            progresoKonami = 0;
            return;
        }
        const tecla = evento.key.toLowerCase();
        if (tecla === KONAMI[progresoKonami]) {
            progresoKonami += 1;
            if (progresoKonami === KONAMI.length) {
                progresoKonami = 0;
                abrir();
            }
        } else {
            // Permite reenganchar si la tecla coincide con el inicio.
            progresoKonami = tecla === KONAMI[0] ? 1 : 0;
        }
    }

    // Mientras el overlay está abierto se congela el scroll de la página.
    const vigilarScroll = watch(abierto, (vale) => {
        document.body.style.overflow = vale ? 'hidden' : '';
    });

    // ?apple abre directo (también vale ?badapple); reacciona si la query
    // llega después (p. ej. navegación interna al HOME con la query puesta).
    const vigilarQuery = watch(
        () => route.query,
        (query) => {
            if (queryPideApple(query)) {
                abrir();
            }
        },
    );

    onMounted(() => {
        window.addEventListener('keydown', alPulsarTecla);
        prefetch();
        if (queryPideApple(route.query)) {
            abrir();
        }
    });

    onUnmounted(() => {
        window.removeEventListener('keydown', alPulsarTecla);
        vigilarScroll();
        vigilarQuery();
        document.body.style.overflow = '';
    });

    return { abierto, abrir, cerrar, clicsLogo };
}
