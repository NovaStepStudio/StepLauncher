import { ref, watch } from 'vue';

// Paleta completa alineada con Ajustes > Música (colorMode):
//  - vibrant  -> colores más vivos de la carátula (alta saturación, luma media)
//  - dominant -> color principal más poblado (no promedio grisáceo)
//  - muted    -> tonos apagados y elegantes (baja saturación)
//  - least    -> sutil / menos saturado y claro (lightMuted)
//  - random   -> elección aleatoria entre variantes
// Se añaden lightVibrant / darkMuted para riqueza y generación de degradados más precisos.
export interface CoverPalette {
    vibrant: string;      // más vivo / saturado (Ajustes: Vibrante)
    muted: string;        // suave / desaturado (Ajustes: Suave)
    darkVibrant: string;  // vivo oscuro
    lightVibrant: string; // vivo claro
    darkMuted: string;    // suave oscuro
    lightMuted: string;   // sutil / claro (Ajustes: Sutil)
    dominant: string;     // color principal por población (Ajustes: Dominante)
    isDark: boolean;      // true si el dominante es oscuro (para contraste de texto)
}

// ───────────────────────── Utilidades de color ─────────────────────────

function clamp(v: number, lo: number, hi: number): number {
    return Math.max(lo, Math.min(hi, v));
}

function rgbToHex(r: number, g: number, b: number): string {
    return `#${[r, g, b].map((v) => clamp(Math.round(v), 0, 255).toString(16).padStart(2, '0')).join('')}`;
}

function hexToRgb(hex: string): [number, number, number] {
    const h = hex.replace('#', '').trim();
    const full = h.length === 3 ? h.split('').map((c) => c + c).join('') : h;
    return [
        parseInt(full.slice(0, 2), 16) || 0,
        parseInt(full.slice(2, 4), 16) || 0,
        parseInt(full.slice(4, 6), 16) || 0,
    ];
}

// Convierte RGB (0-255) a HSL (h 0-360, s 0-1, l 0-1)
function rgbToHsl(r: number, g: number, b: number): [number, number, number] {
    const rn = r / 255, gn = g / 255, bn = b / 255;
    const max = Math.max(rn, gn, bn), min = Math.min(rn, gn, bn);
    let h = 0, s = 0;
    const l = (max + min) / 2;
    const d = max - min;
    if (d !== 0) {
        s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
        switch (max) {
            case rn: h = (gn - bn) / d + (gn < bn ? 6 : 0); break;
            case gn: h = (bn - rn) / d + 2; break;
            default: h = (rn - gn) / d + 4; break;
        }
        h *= 60;
    }
    return [h, clamp(s, 0, 1), clamp(l, 0, 1)];
}

function hslToRgb(h: number, s: number, l: number): [number, number, number] {
    const hn = ((h % 360) + 360) % 360 / 360;
    const sn = clamp(s, 0, 1), ln = clamp(l, 0, 1);
    if (sn === 0) {
        const v = Math.round(ln * 255);
        return [v, v, v];
    }
    const q = ln < 0.5 ? ln * (1 + sn) : ln + sn - ln * sn;
    const p = 2 * ln - q;
    const hue2rgb = (pp: number, qq: number, t: number): number => {
        let tt = t;
        if (tt < 0) tt += 1;
        if (tt > 1) tt -= 1;
        if (tt < 1 / 6) return pp + (qq - pp) * 6 * tt;
        if (tt < 1 / 2) return qq;
        if (tt < 2 / 3) return pp + (qq - pp) * (2 / 3 - tt) * 6;
        return pp;
    };
    const r = Math.round(hue2rgb(p, q, hn + 1 / 3) * 255);
    const g = Math.round(hue2rgb(p, q, hn) * 255);
    const b = Math.round(hue2rgb(p, q, hn - 1 / 3) * 255);
    return [clamp(r, 0, 255), clamp(g, 0, 255), clamp(b, 0, 255)];
}

// Luminancia relativa WCAG (0-1) para decidir isDark y contraste
function relativeLuminance(r: number, g: number, b: number): number {
    const toLin = (c: number): number => {
        const s = c / 255;
        return s <= 0.04045 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
    };
    const rl = toLin(r), gl = toLin(g), bl = toLin(b);
    return 0.2126 * rl + 0.7152 * gl + 0.0722 * bl;
}

// ───────────────────────── Caché LRU ─────────────────────────

const PALETTE_CACHE = new Map<string, CoverPalette>();
const CACHE_LIMIT = 48;

function cacheGet(key: string): CoverPalette | null {
    const v = PALETTE_CACHE.get(key);
    if (!v) return null;
    // Mover al final (LRU)
    PALETTE_CACHE.delete(key);
    PALETTE_CACHE.set(key, v);
    return v;
}

function cacheSet(key: string, pal: CoverPalette): void {
    if (PALETTE_CACHE.has(key)) PALETTE_CACHE.delete(key);
    PALETTE_CACHE.set(key, pal);
    if (PALETTE_CACHE.size > CACHE_LIMIT) {
        const first = PALETTE_CACHE.keys().next().value as string | undefined;
        if (first) PALETTE_CACHE.delete(first);
    }
}

export function clearPaletteCache(): void {
    PALETTE_CACHE.clear();
}

// ───────────────────────── Cuantización: Median Cut ─────────────────────────

type Rgb = [number, number, number];

interface Box {
    pixels: Rgb[];
    r1: number; r2: number;
    g1: number; g2: number;
    b1: number; b2: number;
}

function createBox(pixels: Rgb[]): Box {
    let r1 = 255, r2 = 0, g1 = 255, g2 = 0, b1 = 255, b2 = 0;
    for (let i = 0; i < pixels.length; i++) {
        const p = pixels[i]!;
        const r = p[0], g = p[1], b = p[2];
        if (r < r1) r1 = r; if (r > r2) r2 = r;
        if (g < g1) g1 = g; if (g > g2) g2 = g;
        if (b < b1) b1 = b; if (b > b2) b2 = b;
    }
    return { pixels, r1, r2, g1, g2, b1, b2 };
}

function longestChannel(box: Box): 0 | 1 | 2 {
    const rRange = box.r2 - box.r1;
    const gRange = box.g2 - box.g1;
    const bRange = box.b2 - box.b1;
    if (gRange >= rRange && gRange >= bRange) return 1;
    if (bRange >= rRange && bRange >= gRange) return 2;
    return 0;
}

function averageBox(box: Box): { r: number; g: number; b: number; count: number } {
    let rs = 0, gs = 0, bs = 0;
    const n = box.pixels.length;
    for (let i = 0; i < n; i++) {
        const p = box.pixels[i]!;
        rs += p[0]; gs += p[1]; bs += p[2];
    }
    if (n === 0) return { r: 0, g: 0, b: 0, count: 0 };
    return { r: Math.round(rs / n), g: Math.round(gs / n), b: Math.round(bs / n), count: n };
}

// Median Cut determinístico: genera hasta maxColors colores representativos por población
function medianCut(pixels: Rgb[], maxColors: number): Array<{ r: number; g: number; b: number; count: number }> {
    if (!pixels.length) return [];
    if (pixels.length <= maxColors) {
        // Cada píxel ya es un color, promediar duplicados cercanos no aporta; devolver únicos aproximados
        const uniq = new Map<string, { r: number; g: number; b: number; count: number }>();
        for (const p of pixels) {
            const k = `${p[0]},${p[1]},${p[2]}`;
            const e = uniq.get(k);
            if (e) e.count++;
            else uniq.set(k, { r: p[0], g: p[1], b: p[2], count: 1 });
        }
        return Array.from(uniq.values()).slice(0, maxColors);
    }

    const boxes: Box[] = [createBox(pixels)];

    while (boxes.length < maxColors) {
        // Elegir caja con mayor volumen ponderado por población y rango
        let bestIdx = -1;
        let bestScore = -1;
        for (let i = 0; i < boxes.length; i++) {
            const b = boxes[i]!;
            if (b.pixels.length < 2) continue;
            const rR = b.r2 - b.r1, gR = b.g2 - b.g1, blR = b.b2 - b.b1;
            const range = Math.max(rR, gR, blR);
            if (range < 4) continue; // sin variación suficiente
            // Puntuación: población * (rango + volumen). Prioriza cajas grandes y pobladas
            const volume = (rR + 1) * (gR + 1) * (blR + 1);
            const score = b.pixels.length * (0.7 + 0.3 * Math.log(1 + volume) / 10) * (range + 1);
            // Equivalente ligero a count * range
            if (score > bestScore) { bestScore = score; bestIdx = i; }
        }
        if (bestIdx === -1) break;
        const box = boxes[bestIdx]!;
        const ch = longestChannel(box);
        // Ordenar por canal dominante y cortar por mediana
        box.pixels.sort((a, b) => a[ch]! - b[ch]!);
        const mid = Math.floor(box.pixels.length / 2);
        // Evitar cortes degenerados donde todos los valores son iguales: buscar borde de valor
        // Si el valor en mid es igual al de mid-1 y mid+1, ajustar mid al cambio de valor más cercano
        let leftEnd = mid;
        // Intentar equilibrar pero respetar cambio de valor si hay meseta grande
        const medianVal = box.pixels[mid]![ch]!;
        // Buscar el límite donde cambia el valor cerca de mid para no separar colores idénticos
        // Si hay meseta, mover leftEnd al inicio de la meseta o final según equilibrio
        // Simplificado: mantener mid; el impacto es mínimo con 96x96
        const leftPixels = box.pixels.slice(0, leftEnd);
        const rightPixels = box.pixels.slice(leftEnd);
        if (!leftPixels.length || !rightPixels.length) break;
        boxes.splice(bestIdx, 1);
        boxes.push(createBox(leftPixels), createBox(rightPixels));
        // Actualizar medianVal referencia para no variable sin uso (evita lint)
        void medianVal;
    }

    return boxes.map(averageBox).filter((c) => c.count > 0);
}

// ───────────────────────── Puntuación tipo Vibrant ─────────────────────────

interface ScoredColor {
    r: number; g: number; b: number; count: number;
    h: number; s: number; l: number;
    hex: string;
}

interface Target {
    sat: number;
    lum: number;
}

// Perfiles objetivo inspirados en Vibrant.js / Material: saturación y luminancia deseadas
const TARGETS: Record<string, Target> = {
    vibrant:      { sat: 0.95, lum: 0.52 },
    lightVibrant: { sat: 0.85, lum: 0.74 },
    darkVibrant:  { sat: 0.85, lum: 0.28 },
    muted:        { sat: 0.32, lum: 0.52 },
    lightMuted:   { sat: 0.24, lum: 0.76 },
    darkMuted:    { sat: 0.26, lum: 0.28 },
};

// Pesos: saturación 0.42, luminancia 0.34, población 0.24 (Vibrant)
const W_SAT = 0.42, W_LUM = 0.34, W_POP = 0.24;

function invertDiff(a: number, b: number): number {
    return 1 - Math.abs(a - b);
}

function scoreColor(c: ScoredColor, target: Target, popNorm: number): number {
    const satScore = invertDiff(c.s, target.sat);
    const lumScore = invertDiff(c.l, target.lum);
    return satScore * W_SAT + lumScore * W_LUM + popNorm * W_POP;
}

// Filtros de idoneidad por familia para evitar escoger grises para vibrantes y viceversa
function isSuitable(c: ScoredColor, kind: string): boolean {
    const { s, l } = c;
    switch (kind) {
        case 'vibrant':      return s >= 0.35 && l >= 0.28 && l <= 0.76;
        case 'lightVibrant': return s >= 0.30 && l >= 0.52 && l <= 0.93;
        case 'darkVibrant':  return s >= 0.30 && l >= 0.06 && l <= 0.50;
        case 'muted':        return s <= 0.55 && l >= 0.28 && l <= 0.72;
        case 'lightMuted':   return s <= 0.48 && l >= 0.52 && l <= 0.94;
        case 'darkMuted':    return s <= 0.48 && l >= 0.06 && l <= 0.50;
        default: return true;
    }
}

function findBest(
    colors: ScoredColor[],
    target: Target,
    kind: string,
    maxPop: number,
    excludeHex?: Set<string>,
): ScoredColor | null {
    let best: ScoredColor | null = null;
    let bestScore = -1;
    // Primer intento: solo candidatos idóneos
    const tryPass = (filter: boolean): void => {
        for (const c of colors) {
            if (excludeHex && excludeHex.has(c.hex)) continue;
            if (filter && !isSuitable(c, kind)) continue;
            const popNorm = maxPop > 0 ? c.count / maxPop : 0;
            const sc = scoreColor(c, target, popNorm);
            // Penalizar colores con muy baja saturación para vibrantes si rozan gris
            let penal = 0;
            if ((kind === 'vibrant' || kind === 'lightVibrant' || kind === 'darkVibrant') && c.s < 0.18) penal = 0.35;
            const final = sc - penal;
            if (final > bestScore) { bestScore = final; best = c; }
        }
    };
    tryPass(true);
    if (best) return best;
    // Sin candidato idóneo, probar sin filtro (imagen gris / mono)
    tryPass(false);
    return best;
}

function findDominant(colors: ScoredColor[]): ScoredColor | null {
    if (!colors.length) return null;
    const sorted = [...colors].sort((a, b) => b.count - a.count);
    const max = sorted[0]!.count;
    const first = sorted[0]!;
    const isExtreme = (c: ScoredColor): boolean => c.s < 0.06 && (c.l > 0.94 || c.l < 0.06);
    if (!isExtreme(first)) return first;
    // Si el más poblado es blanco/negro casi puro, buscar alternativa con >28% de población y no extremo
    for (let i = 1; i < sorted.length; i++) {
        const c = sorted[i]!;
        if (!isExtreme(c) && c.count >= max * 0.28) return c;
    }
    // Si todo es extremo (cover blanca/negra), quedarse con el primero
    return first;
}

// Genera un color fallback a partir del tono dominante ajustando sat/lum al objetivo
function fallbackFromDominant(dom: ScoredColor | null, target: Target): string {
    if (!dom) return '#5ed89a';
    const isGray = dom.s < 0.08;
    const h = dom.h;
    const s = isGray ? clamp(target.sat * 0.08, 0, 0.06) : target.sat;
    const l = target.lum;
    const rgb = hslToRgb(h, s, l);
    return rgbToHex(rgb[0], rgb[1], rgb[2]);
}

// ───────────────────────── Carga de imagen ─────────────────────────

function isDataOrBlob(src: string): boolean {
    return src.startsWith('data:') || src.startsWith('blob:');
}

function needsCors(src: string): boolean {
    return src.startsWith('http://') || src.startsWith('https://') || src.startsWith('//');
}

function loadImage(src: string, timeoutMs = 7000): Promise<HTMLImageElement> {
    return new Promise((resolve, reject) => {
        const img = new Image();
        if (needsCors(src)) img.crossOrigin = 'anonymous';
        // Para wails://, asset:// o file:// no se usa CORS
        let timer: number | undefined;
        let settled = false;
        const done = (fn: () => void) => {
            if (settled) return;
            settled = true;
            if (timer !== undefined) clearTimeout(timer);
            fn();
        };
        timer = window.setTimeout(() => {
            done(() => reject(new Error('timeout al cargar carátula')));
        }, timeoutMs) as unknown as number;

        img.onload = () => {
            // Intentar decode() si está disponible para asegurar píxeles listos
            const finish = () => done(() => resolve(img));
            try {
                const anyImg = img as unknown as { decode?: () => Promise<void> };
                if (anyImg.decode) {
                    anyImg.decode().then(finish).catch(finish);
                } else {
                    finish();
                }
            } catch {
                done(() => resolve(img));
            }
        };
        img.onerror = () => done(() => reject(new Error('no se pudo cargar la imagen')));
        img.src = src;
        // Si ya está en caché y completa, onload puede no disparar de nuevo en algunos WebView
        if (img.complete && img.naturalWidth > 0) {
            // Dar un tick para que onload tenga chance, si no forzar
            window.setTimeout(() => {
                if (!settled) {
                    try {
                        const anyImg = img as unknown as { decode?: () => Promise<void> };
                        if (anyImg.decode) anyImg.decode().then(() => done(() => resolve(img))).catch(() => done(() => resolve(img)));
                        else done(() => resolve(img));
                    } catch {
                        done(() => resolve(img));
                    }
                }
            }, 20);
        }
    });
}

// ───────────────────────── Extracción principal ─────────────────────────

export async function extractPalette(src: string): Promise<CoverPalette | null> {
    if (!src) return null;
    if (typeof document === 'undefined') return null;

    const cached = cacheGet(src);
    if (cached) return cached;

    try {
        const img = await loadImage(src);

        // Muestreo de alta calidad: 96x96 con recorte centrado (cover) para no deformar
        // 96 aporta ~9216 píxeles, suficiente para median-cut preciso sin penalizar rendimiento
        const size = 96;
        const canvas = document.createElement('canvas');
        canvas.width = size;
        canvas.height = size;
        const ctx = canvas.getContext('2d', { willReadFrequently: true } as unknown as CanvasRenderingContext2DSettings);
        if (!ctx) return null;
        // Suavizado de alta calidad para que el downscale conserve mezcla perceptual de colores
        (ctx as unknown as { imageSmoothingEnabled: boolean }).imageSmoothingEnabled = true;
        try { (ctx as unknown as { imageSmoothingQuality: string }).imageSmoothingQuality = 'high'; } catch {}

        const iw = img.naturalWidth || size;
        const ih = img.naturalHeight || size;
        let sx = 0, sy = 0, sw = iw, sh = ih;
        // Recorte centrado tipo "cover" para preservar proporción y priorizar centro (donde suele estar el motivo)
        if (iw > 0 && ih > 0) {
            const scale = Math.max(size / iw, size / ih);
            sw = size / scale;
            sh = size / scale;
            sx = (iw - sw) / 2;
            sy = (ih - sh) / 2;
        }
        ctx.clearRect(0, 0, size, size);
        try {
            ctx.drawImage(img, sx, sy, sw, sh, 0, 0, size, size);
        } catch {
            // Fallback simple stretch si el recorte falla (imagen corrupta / dimensiones raras)
            ctx.drawImage(img, 0, 0, size, size);
        }

        let data: Uint8ClampedArray;
        try {
            data = ctx.getImageData(0, 0, size, size).data;
        } catch (e) {
            // Canvas contaminado por CORS (tainted) — no se puede leer
            try { canvas.width = 0; canvas.height = 0; } catch {}
            return null;
        }

        // Recolectar píxeles válidos: descartar semitransparentes y píxeles con alfa muy bajo
        // Umbral 24 (más permisivo que 128) para carátulas con bordes suaves/antialias
        const pixels: Rgb[] = [];
        for (let i = 0; i < data.length; i += 4) {
            const a = data[i + 3]!;
            if (a < 24) continue;
            const r = data[i]!, g = data[i + 1]!, b = data[i + 2]!;
            // No descartar blancos/negros aquí; el filtrado inteligente se hace después en scoring
            // Solo ignorar píxeles totalmente transparentes ya filtrados
            pixels.push([r, g, b]);
        }

        // Liberar canvas pronto
        try { canvas.width = 0; canvas.height = 0; } catch {}

        if (pixels.length < 8) return null;

        // Si hay demasiados píxeles (>8000), muestrear decimando para acelerar median-cut sin perder precisión
        let sampled: Rgb[] = pixels;
        if (pixels.length > 8000) {
            const step = Math.ceil(pixels.length / 7000);
            sampled = [];
            for (let i = 0; i < pixels.length; i += step) sampled.push(pixels[i]!);
        }

        // Cuantización: extraer hasta 12 clusters representativos
        const quantizedRaw = medianCut(sampled, 12);
        if (!quantizedRaw.length) return null;

        // Enriquecer con HSL y hex para scoring
        const maxPop = Math.max(...quantizedRaw.map((c) => c.count));
        const scored: ScoredColor[] = quantizedRaw.map((c) => {
            const hsl = rgbToHsl(c.r, c.g, c.b);
            return {
                r: c.r, g: c.g, b: c.b, count: c.count,
                h: hsl[0], s: hsl[1], l: hsl[2],
                hex: rgbToHex(c.r, c.g, c.b),
            };
        });

        // Detectar si la imagen es esencialmente gris (baja saturación global)
        const maxSat = Math.max(...scored.map((c) => c.s));

        // Dominante: el cluster más poblado con corrección de extremos blancos/negros
        const dominantScored = findDominant(scored);
        const dominantHex = dominantScored ? dominantScored.hex : scored[0]!.hex;
        const domRgb: [number, number, number] = dominantScored ? [dominantScored.r, dominantScored.g, dominantScored.b] : hexToRgb(dominantHex);
        const dominantForFallback = dominantScored ?? scored[0]!;

        // Selección de cada variante con scoring Vibrant
        // Para evitar duplicados exactos cuando hay paleta rica, se intenta excluir el dominante
        // si existe alternativa con puntuación cercana (>=75% de la mejor)
        const pickWithDedup = (kind: string, target: Target): string => {
            const bestAll = findBest(scored, target, kind, maxPop);
            if (!bestAll) return fallbackFromDominant(dominantForFallback, target);
            // Si es gris total, no forzar saturación: devolver el mejor tal cual
            if (maxSat < 0.12) return bestAll.hex;
            // Intentar alternativa distinta al dominante si el mejor es el dominante
            if (bestAll.hex.toLowerCase() === dominantHex.toLowerCase() && scored.length > 1) {
                const exclude = new Set<string>([dominantHex.toLowerCase()]);
                const alt = findBest(scored, target, kind, maxPop, exclude);
                if (alt) {
                    const popBest = bestAll.count / maxPop;
                    const popAlt = alt.count / maxPop;
                    const sBest = scoreColor(bestAll, target, popBest);
                    const sAlt = scoreColor(alt, target, popAlt);
                    // Si la alternativa es razonablemente buena (>=75% de la mejor), preferir variedad
                    if (sAlt >= sBest * 0.75) return alt.hex;
                }
            }
            // Si la puntuación es muy baja (<0.35) y no es gris, generar fallback ajustado al tono dominante
            const popNorm = bestAll.count / maxPop;
            const bestScore = scoreColor(bestAll, target, popNorm);
            if (bestScore < 0.34 && maxSat >= 0.12) {
                // Solo hacer fallback si el color elegido está muy lejos del objetivo
                const satDist = Math.abs(bestAll.s - target.sat);
                const lumDist = Math.abs(bestAll.l - target.lum);
                if (satDist > 0.5 || lumDist > 0.45) {
                    return fallbackFromDominant(dominantForFallback, target);
                }
            }
            return bestAll.hex;
        };

        const vibrantHex = pickWithDedup('vibrant', TARGETS.vibrant!);
        const lightVibrantHex = pickWithDedup('lightVibrant', TARGETS.lightVibrant!);
        const darkVibrantHex = pickWithDedup('darkVibrant', TARGETS.darkVibrant!);
        const mutedHex = pickWithDedup('muted', TARGETS.muted!);
        const lightMutedHex = pickWithDedup('lightMuted', TARGETS.lightMuted!);
        const darkMutedHex = pickWithDedup('darkMuted', TARGETS.darkMuted!);

        // Asegurar que vibrant/muted no sean idénticos a dominant por generación fallback en grises:
        // En grises, vibrant y muted tenderán a ser grises muy similares; es correcto y preciso.

        const lumDom = relativeLuminance(domRgb[0], domRgb[1], domRgb[2]);
        const palette: CoverPalette = {
            vibrant: vibrantHex,
            muted: mutedHex,
            darkVibrant: darkVibrantHex,
            lightVibrant: lightVibrantHex,
            darkMuted: darkMutedHex,
            lightMuted: lightMutedHex,
            dominant: dominantHex,
            isDark: lumDom < 0.5,
        };

        cacheSet(src, palette);
        return palette;
    } catch {
        return null;
    }
}

// Composable reactivo: observa un coverUrl y expone paleta con caché y tokens anti-carrera
export function useCoverPalette(coverUrl: () => string | undefined) {
    const palette = ref<CoverPalette | null>(null);
    const loading = ref(false);

    let token = 0;
    async function refresh() {
        const src = coverUrl();
        if (!src) {
            // No limpiar paleta al cambiar de pista: mantiene el degradado anterior
            // y evita el flash transparente hasta que la nueva carátula se resuelva
            return;
        }
        // Atajo por caché síncrono antes de marcar loading
        const hit = cacheGet(src);
        if (hit) {
            palette.value = hit;
            loading.value = false;
            return;
        }
        const myToken = ++token;
        loading.value = true;
        const p = await extractPalette(src);
        if (myToken !== token) return;
        if (p) palette.value = p;
        loading.value = false;
    }

    watch(coverUrl, refresh, { immediate: true });

    return { palette, loading, refresh, extractPalette };
}

// Helpers para aplicar al panel: genera variables CSS y degradados a partir de la paleta
export function paletteToCss(p: CoverPalette | null, alpha = 0.18) {
    if (!p) return {};
    const toRgba = (hex: string, a: number) => {
        const [r, g, b] = hexToRgb(hex);
        return `rgba(${r}, ${g}, ${b}, ${a})`;
    };
    // Overlay más rico: vibrant protagonista + dominante como apoyo, con transparencia adaptada
    const overlay = `linear-gradient(135deg, ${toRgba(p.vibrant, alpha)} 0%, ${toRgba(p.dominant, alpha * 0.72)} 48%, ${toRgba(p.muted, alpha * 0.38)} 78%, transparent 100%)`;
    return {
        '--cover-vibrant': p.vibrant,
        '--cover-vibrant-light': p.lightVibrant,
        '--cover-vibrant-dark': p.darkVibrant,
        '--cover-muted': p.muted,
        '--cover-muted-dark': p.darkMuted,
        '--cover-dominant': p.dominant,
        '--cover-dark': p.darkVibrant,
        '--cover-light': p.lightMuted,
        '--cover-overlay': overlay,
        // Alias semánticos para compatibilidad con SCSS existentes
        '--cover-accent': p.vibrant,
        '--cover-subtle': p.lightMuted,
    } as Record<string, string>;
}

// Obtiene el color según el modo seleccionado en Ajustes → Música
// Mapea directamente las 5 opciones del panel: Vibrante / Dominante / Suave / Sutil / Aleatorio
export function getPaletteColor(p: CoverPalette | null, mode: string): string {
    if (!p) return '#5ed89a';
    switch (mode) {
        case 'dominant': return p.dominant;
        case 'muted': return p.muted;
        case 'least': return p.lightMuted; // Sutil: menos saturado y más claro
        case 'darkVibrant': return p.darkVibrant;
        case 'lightVibrant': return p.lightVibrant;
        case 'darkMuted': return p.darkMuted;
        case 'lightMuted': return p.lightMuted;
        case 'random': {
            // Pool ampliado para mayor variedad manteniendo coherencia de la carátula
            const opts = [p.vibrant, p.dominant, p.muted, p.lightMuted, p.darkVibrant, p.lightVibrant, p.darkMuted]
                .filter(Boolean) as string[];
            // Eliminar duplicados exactos para no sesgar el random
            const uniq = Array.from(new Set(opts.map((c) => c.toLowerCase()))).map((lc) => opts.find((o) => o.toLowerCase() === lc)!);
            return uniq[Math.floor(Math.random() * uniq.length)]!;
        }
        case 'vibrant':
        default: return p.vibrant;
    }
}
