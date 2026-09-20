import { ref } from 'vue';

// ============================================================================
// Composable useVersion — Versión viva del proyecto sin hardcodear.
// Lee el badge de versión del README en GitHub y extrae el texto:
//   [![Versión](https://img.shields.io/badge/Versión-2.5.0--beta-31b3ff?...)]
// En shields.io el guion se escapa como doble guion ("--"), por eso
// "2.5.0--beta" se normaliza a "2.5.0-beta" y se muestra como "v2.5.0-beta".
// Si el README no responde, usa la API de releases como respaldo, y si
// GitHub tampoco responde, el worker propio (sin token).
// ============================================================================

const README_URL = 'https://raw.githubusercontent.com/NovaStepStudio/StepLauncher/main/README.md';
const RELEASES_URL = 'https://api.github.com/repos/NovaStepStudio/StepLauncher/releases?per_page=5';
const WORKER_RELEASES_URL = 'https://steplauncher.stepnicka012.workers.dev/updates/steplauncher/releases';

const version = ref('');
const loading = ref(false);
const loaded = ref(false);

function normalizarBadge(raw: string): string {
    // raw llega como "2.5.0--beta" → "2.5.0-beta".
    const limpio = raw.replace(/--/g, '-').trim();
    if (!limpio) return '';
    return limpio.startsWith('v') ? limpio : `v${limpio}`;
}

function extraerDelReadme(texto: string): string {
    // Busca "badge/Versión-<version>-<color>" en el markdown del README.
    const match = texto.match(/badge\/Versi[oó]n-([A-Za-z0-9._\-]+?)-[0-9a-fA-F]{4,}/);
    if (!match) return '';
    return normalizarBadge(match[1] ?? '');
}

async function cargarDesdeReadme(): Promise<string> {
    const res = await fetch(README_URL);
    if (!res.ok) throw new Error(`README respondió ${res.status}`);
    const texto = await res.text();
    return extraerDelReadme(texto);
}

function normalizarTag(tag: string): string {
    // "StepLauncher-v2.5.0-beta.2" → "v2.5.0-beta.2", "2.3.1" → "v2.3.1".
    const limpio = tag.replace(/^StepLauncher-/i, '').trim();
    if (!limpio) return '';
    return limpio.startsWith('v') ? limpio : `v${limpio}`;
}

async function tagDesde(url: string): Promise<string> {
    const res = await fetch(url);
    if (!res.ok) throw new Error(`${url} respondió ${res.status}`);
    const raw = (await res.json()) as Array<{ tag_name: string; draft: boolean }>;
    return normalizarTag(raw.find((r) => !r.draft)?.tag_name ?? '');
}

async function cargarDesdeApi(): Promise<string> {
    try {
        const v = await tagDesde(RELEASES_URL);
        if (v) return v;
    } catch {
        // GitHub caído o con rate limit: se prueba el worker.
    }
    return tagDesde(WORKER_RELEASES_URL);
}

export function useVersion() {
    async function load() {
        if (loaded.value || loading.value) return;
        loading.value = true;
        try {
            let v = '';
            try {
                v = await cargarDesdeReadme();
            } catch {
                v = '';
            }
            if (!v) {
                try {
                    v = await cargarDesdeApi();
                } catch {
                    v = '';
                }
            }
            if (v) {
                version.value = v;
                loaded.value = true;
            }
        } finally {
            loading.value = false;
        }
    }

    return { version, loading, load };
}
