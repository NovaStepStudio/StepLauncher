import { onMounted, onUnmounted } from 'vue';

// ---------- Capas de Escape (ESC por pila) ----------
// Cada overlay del launcher (panel, modal, diálogo o confirmación) registra
// una capa al montarse. Al pulsar ESC se cierra SIEMPRE la capa activa
// superior: primero por prioridad (a mayor prioridad, antes se cierra) y, en
// caso de empate, la registrada más recientemente (la que está visualmente
// encima, ya que se montó después).
// Así, con un diálogo abierto sobre un panel, ESC cierra primero el diálogo
// y solo al pulsarlo de nuevo se cierra el panel de debajo.

export interface UseOverlayEscapeOptions {
    // Comprueba si la capa está visible/activa (por defecto, siempre activa).
    // Debe reflejar la visibilidad REAL del overlay (incluyendo la del panel
    // que lo contiene), para que una capa oculta nunca robe el ESC.
    isActive?: () => boolean;
    // Nivel de apilado: 0 = paneles y modales de nivel superior,
    // 1 = sub-vistas dentro de un panel, 2 = diálogos (Host), 3 = Confirm.
    priority?: number;
}

interface EscapeLayer {
    id: number;
    handler: () => void;
    isActive: () => boolean;
    priority: number;
}

const layers: EscapeLayer[] = [];
let listening = false;
let nextId = 1;

function onWindowKeydown(e: KeyboardEvent): void {
    if (e.key !== 'Escape') return;
    let best: EscapeLayer | null = null;
    for (const layer of layers) {
        if (!layer.isActive()) continue;
        if (
            !best ||
            layer.priority > best.priority ||
            (layer.priority === best.priority && layer.id > best.id)
        ) {
            best = layer;
        }
    }
    if (!best) return;
    e.preventDefault();
    best.handler();
}

export function useOverlayEscape(handler: () => void, options: UseOverlayEscapeOptions = {}): void {
    const layer: EscapeLayer = {
        id: nextId++,
        handler,
        isActive: options.isActive ?? (() => true),
        priority: options.priority ?? 0,
    };

    onMounted(() => {
        layers.push(layer);
        if (!listening) {
            window.addEventListener('keydown', onWindowKeydown);
            listening = true;
        }
    });

    onUnmounted(() => {
        const index = layers.indexOf(layer);
        if (index >= 0) layers.splice(index, 1);
        if (listening && layers.length === 0) {
            window.removeEventListener('keydown', onWindowKeydown);
            listening = false;
        }
    });
}
