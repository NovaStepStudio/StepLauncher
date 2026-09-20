import { ref, computed } from 'vue';

// ============================================================================
// Composable useReleases — Últimos releases del repo en GitHub.
// Expone el último release estable, el último prerelease y los datos del repo
// (estrellas, forks) para las páginas de Descarga y Acerca De.
// Si la API de GitHub falla (rate limit, caída), usa como fallback el worker
// propio con releases y previas ya filtrados, sin token:
//   /updates/steplauncher/releases + /updates/steplauncher/prereleases
// ============================================================================

const OWNER = 'NovaStepStudio';
const REPO = 'StepLauncher';
const WORKER_BASE = 'https://steplauncher.stepnicka012.workers.dev/updates/steplauncher';

export interface ReleaseAsset {
    name: string;
    size: number;
    downloads: number;
    url: string;
}

export interface GithubRelease {
    tag: string;
    name: string;
    body: string;
    publishedAt: string;
    prerelease: boolean;
    draft: boolean;
    url: string;
    assets: ReleaseAsset[];
}

export interface RepoInfo {
    stars: number;
    forks: number;
    url: string;
}

export type PlatformKey = 'windows' | 'macos' | 'linux' | 'other';

export type ReleaseChannel = 'stable' | 'beta' | 'alpha';

// Canal del release según su tag: si menciona alpha es alpha, si es previa
// pero no alpha (beta, rc, pre...) se agrupa como beta, si no es previa es estable.
export function channelOf(tag: string, prerelease: boolean): ReleaseChannel {
    if (!prerelease) return 'stable';
    if (tag.toLowerCase().includes('alpha')) return 'alpha';
    return 'beta';
}

const releases = ref<GithubRelease[]>([]);
const repo = ref<RepoInfo>({ stars: 0, forks: 0, url: `https://github.com/${OWNER}/${REPO}` });
const loaded = ref(false);
const loading = ref(false);
const error = ref('');

export function platformOf(fileName: string): PlatformKey {
    const n = fileName.toLowerCase();
    if (n.endsWith('.exe') || n.endsWith('.msi') || n.includes('windows') || n.includes('win64') || n.includes('win-') || n.includes('-win') || n.includes('win32')) return 'windows';
    if (n.endsWith('.dmg') || n.includes('macos') || n.includes('darwin') || n.includes('apple')) return 'macos';
    if (n.endsWith('.appimage') || n.endsWith('.deb') || n.endsWith('.rpm') || n.endsWith('.tar.gz') || n.includes('linux')) return 'linux';
    return 'other';
}

// Arquitectura del asset según el estándar nuevo (amd64/aarch64/x86_64) y el viejo (x64/ia32).
export function archOf(fileName: string): string {
    const n = fileName.toLowerCase();
    if (n.includes('aarch64') || n.includes('arm64')) return 'ARM64';
    if (n.includes('x86_64') || n.includes('amd64') || n.includes('x64') || n.includes('win64')) return 'x64';
    if (n.includes('ia32') || n.includes('i386')) return 'x86';
    return '';
}

// Tipo de paquete: instalador vs portable/binario suelto.
export function kindOf(fileName: string): string {
    const n = fileName.toLowerCase();
    if (n.includes('installer') || n.endsWith('.msi') || n.endsWith('.deb') || n.endsWith('.rpm')) return 'Instalador';
    if (n.endsWith('.zip') || n.endsWith('.appimage')) return 'Portable';
    return '';
}

export function assetsByPlatform(rel: GithubRelease | undefined): Record<PlatformKey, ReleaseAsset[]> {
    const groups: Record<PlatformKey, ReleaseAsset[]> = { windows: [], macos: [], linux: [], other: [] };
    if (!rel) return groups;
    for (const a of rel.assets) groups[platformOf(a.name)].push(a);
    return groups;
}

export function formatSize(bytes: number): string {
    if (!bytes || bytes <= 0) return '—';
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function formatDate(iso: string): string {
    if (!iso) return '—';
    try {
        return new Date(iso).toLocaleDateString('es-AR', { day: 'numeric', month: 'short', year: 'numeric' });
    } catch {
        return '—';
    }
}

interface RawAsset {
    name: string;
    size: number;
    download_count: number;
    browser_download_url: string;
}

interface RawRelease {
    tag_name: string;
    name: string;
    body: string;
    published_at: string;
    prerelease: boolean;
    draft: boolean;
    html_url: string;
    assets: RawAsset[];
}

async function fetchJson(url: string): Promise<RawRelease[]> {
    const res = await fetch(url);
    if (!res.ok) throw new Error(`${url} respondió ${res.status}`);
    return (await res.json()) as RawRelease[];
}

function mapearReleases(raw: RawRelease[]): GithubRelease[] {
    return raw.map((r) => ({
        tag: r.tag_name,
        name: r.name || r.tag_name,
        body: r.body || '',
        publishedAt: r.published_at,
        prerelease: r.prerelease,
        draft: r.draft,
        url: r.html_url,
        assets: (r.assets || []).map((a) => ({
            name: a.name,
            size: a.size,
            downloads: a.download_count,
            url: a.browser_download_url,
        })),
    }));
}

// Fallback sin token: junta estables + previas del worker, saca duplicados
// por tag y ordena de más nuevo a más viejo (igual que GitHub).
// Tolera que uno de los dos endpoints falle mientras el otro responda.
async function cargarDesdeWorker(): Promise<GithubRelease[]> {
    const resultados = await Promise.allSettled([
        fetchJson(`${WORKER_BASE}/releases`),
        fetchJson(`${WORKER_BASE}/prereleases`),
    ]);
    const crudo = resultados.flatMap((r) => (r.status === 'fulfilled' ? r.value : []));
    if (crudo.length === 0) throw new Error('El worker no devolvió releases');
    const vistos = new Set<string>();
    return mapearReleases(
        crudo
            .filter((r) => {
                if (vistos.has(r.tag_name)) return false;
                vistos.add(r.tag_name);
                return true;
            })
            .sort((a, b) => Date.parse(b.published_at || '') - Date.parse(a.published_at || '')),
    );
}

async function loadReleases() {
    try {
        releases.value = mapearReleases(await fetchJson(`https://api.github.com/repos/${OWNER}/${REPO}/releases?per_page=10`));
    } catch {
        releases.value = await cargarDesdeWorker();
    }
}

async function loadRepo() {
    const res = await fetch(`https://api.github.com/repos/${OWNER}/${REPO}`);
    if (!res.ok) throw new Error(`GitHub respondió ${res.status}`);
    const raw = (await res.json()) as { stargazers_count: number; forks_count: number; html_url: string };
    repo.value = { stars: raw.stargazers_count, forks: raw.forks_count, url: raw.html_url };
}

export function useReleases() {
    async function load() {
        if (loaded.value || loading.value) return;
        loading.value = true;
        error.value = '';
        try {
            // Los releases (GitHub o worker) deciden si hay error o no.
            await loadReleases();
            loaded.value = true;
        } catch {
            error.value = 'No se pudo contactar a GitHub. Probá de nuevo en un rato.';
        } finally {
            loading.value = false;
        }
        // Estrellas/forks: mejor esfuerzo, no bloquean las descargas.
        try {
            await loadRepo();
        } catch {
            // Se conservan los valores previos.
        }
    }

    const stable = computed(() => releases.value.find((r) => !r.prerelease && !r.draft));
    const prerelease = computed(() => releases.value.find((r) => r.prerelease && !r.draft));
    const main = computed(() => stable.value ?? prerelease.value);

    // Todas las previas por canal (beta y alpha), sin borradores.
    const betas = computed(() =>
        releases.value.filter((r) => !r.draft && channelOf(r.tag, r.prerelease) === 'beta'),
    );
    const alphas = computed(() =>
        releases.value.filter((r) => !r.draft && channelOf(r.tag, r.prerelease) === 'alpha'),
    );

    return { releases, repo, loaded, loading, error, load, stable, prerelease, main, betas, alphas, assetsByPlatform };
}
