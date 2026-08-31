import { reactive, readonly, ref, computed } from 'vue';
import type { DeepReadonly, Ref } from 'vue';

// ============================================================================
// useModrinth — Cliente reutilizable de la API de Modrinth (Labrinth v2).
//
// Referencia: https://docs.modrinth.com/api — base https://api.modrinth.com/v2
//
//   GET /search          Búsqueda con facets (AND por array exterior, OR por
//                        elementos del mismo array), index (relevance,
//                        downloads, follows, newest, updated), offset y limit.
//   GET /project/{id}    Detalle de un proyecto (descripción, galería, etc.).
//   GET /project/{id}/version?include_changelog=false
//                        Versiones de un proyecto con sus archivos.
//   GET /version/{id}    Una versión concreta (changelog completo).
//   GET /tag/game_version / GET /tag/loader   Listados de versiones y loaders.
//
// Diseño:
//   - Funciones de bajo nivel (fetch*): HTTP + caché TTL + retries; sin estado,
//     reutilizables desde cualquier módulo.
//   - createModrinthClient(): instancia con estado reactivo (búsqueda,
//     proyecto, changelog, tags) con cancelación de peticiones y debounce.
//   - useModrinth(): instancia singleton compartida por todo el dominio.
//     Cualquier otro dominio puede crear sus propios clientes si lo necesita.
// ============================================================================

// ---------- Tipos de la API ----------

export type ModrinthProjectType = 'mod' | 'modpack' | 'resourcepack' | 'shader';
export type SearchIndex = 'relevance' | 'downloads' | 'follows' | 'newest' | 'updated';
export type VersionType = 'release' | 'beta' | 'alpha';

export interface SearchHit {
    project_id: string;
    project_type: string;
    title: string;
    description: string;
    author: string;
    categories: readonly string[];
    versions: readonly string[];
    downloads: number;
    follows: number;
    icon_url: string | null;
    date_created: string;
    date_modified: string;
    latest_version: string;
    license: string;
    environment: readonly string[];
    slug: string | null;
    color: number | null;
    featured_gallery: string | null;
}

export interface SearchResponse {
    hits: SearchHit[];
    offset: number;
    limit: number;
    total_hits: number;
}

export interface GalleryImage {
    url: string;
    raw_url: string;
    featured: boolean;
    title: string | null;
    description: string | null;
    created: string;
    ordering?: number;
}

export interface ModrinthProject {
    id: string;
    slug: string | null;
    title: string;
    description: string;
    body: string | null;
    body_url: string | null;
    project_type: ModrinthProjectType;
    downloads: number;
    follows: number;
    icon_url: string | null;
    gallery: readonly GalleryImage[];
    featured_gallery: string | null;
    categories: readonly string[];
    loaders: readonly string[];
    game_versions: readonly string[];
    environment: readonly string[];
    license: { id: string; name: string; url: string | null };
    published: string;
    updated: string;
    status: string;
    team: string;
    color: number | null;
    source_url: string | null;
}

export interface ProjectFile {
    url: string;
    filename: string;
    size: number;
    hashes: Record<string, string>;
    primary: boolean;
    file_type: string | null;
}

export interface ProjectDependency {
    version_id: string | null;
    project_id: string | null;
    file_name: string | null;
    dependency_type: 'required' | 'optional' | 'incompatible' | 'embedded';
}

export interface ModrinthVersion {
    id: string;
    project_id: string;
    featured: boolean;
    name: string;
    version_number: string;
    changelog: string | null;
    date_published: string;
    downloads: number;
    version_type: VersionType;
    status: string;
    loaders: readonly string[];
    game_versions: readonly string[];
    files: readonly ProjectFile[];
    dependencies: readonly ProjectDependency[];
    environment: string | null;
}

export interface GameVersion {
    version: string;
    version_type: 'release' | 'snapshot' | 'old_beta' | 'old_alpha';
    date: string;
    major: boolean;
}

export interface LoaderTag {
    name: string;
    icon: string;
    supported_project_types: string[];
}

// ---------- HTTP base ----------

const API_BASE = 'https://api.modrinth.com/v2';
const USER_AGENT = 'StepLauncher/2.5.0';

// --- IndexedDB cache for persistence across sessions ---
const DB_NAME = 'StepLauncher_ModrinthCache';
const DB_VERSION = 1;
const STORE_NAME = 'cache';

let dbPromise: Promise<IDBDatabase> | null = null;

function getDB(): Promise<IDBDatabase> {
    if (!dbPromise) {
        dbPromise = new Promise((resolve, reject) => {
            const request = indexedDB.open(DB_NAME, DB_VERSION);
            request.onupgradeneeded = () => {
                const db = request.result;
                if (!db.objectStoreNames.contains(STORE_NAME)) {
                    db.createObjectStore(STORE_NAME);
                }
            };
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }
    return dbPromise;
}

interface CacheEntry {
    expires: number;
    data: unknown;
}

const memoryCache = new Map<string, CacheEntry>();

async function cacheGet<T>(key: string): Promise<T | undefined> {
    // Check memory first (fastest)
    const memEntry = memoryCache.get(key);
    if (memEntry) {
        if (Date.now() > memEntry.expires) {
            memoryCache.delete(key);
        } else {
            return memEntry.data as T;
        }
    }
    // Check IndexedDB
    try {
        const db = await getDB();
        const entry = await new Promise<CacheEntry | undefined>((resolve, reject) => {
            const tx = db.transaction(STORE_NAME, 'readonly');
            const store = tx.objectStore(STORE_NAME);
            const req = store.get(key);
            req.onsuccess = () => resolve(req.result);
            req.onerror = () => reject(req.error);
        });
        if (entry && Date.now() <= entry.expires) {
            memoryCache.set(key, entry);
            return entry.data as T;
        }
        if (entry) {
            await cacheDelete(key);
        }
    } catch {
        // Ignore IndexedDB errors, fall back to memory only
    }
    return undefined;
}

async function cacheSet<T>(key: string, data: T, ttlMs: number): Promise<void> {
    const entry: CacheEntry = { expires: Date.now() + ttlMs, data };
    memoryCache.set(key, entry);
    if (memoryCache.size > 500) {
        // Clean oldest entries
        const keys = [...memoryCache.keys()];
        for (let i = 0; i < 100; i++) memoryCache.delete(keys[i]!);
    }
    try {
        const db = await getDB();
        await new Promise<void>((resolve, reject) => {
            const tx = db.transaction(STORE_NAME, 'readwrite');
            const store = tx.objectStore(STORE_NAME);
            const req = store.put(entry, key);
            req.onsuccess = () => resolve();
            req.onerror = () => reject(req.error);
        });
    } catch {
        // Ignore IndexedDB errors
    }
}

async function cacheDelete(key: string): Promise<void> {
    memoryCache.delete(key);
    try {
        const db = await getDB();
        await new Promise<void>((resolve, reject) => {
            const tx = db.transaction(STORE_NAME, 'readwrite');
            const store = tx.objectStore(STORE_NAME);
            const req = store.delete(key);
            req.onsuccess = () => resolve();
            req.onerror = () => reject(req.error);
        });
    } catch {
        // Ignore
    }
}

async function cacheClear(): Promise<void> {
    memoryCache.clear();
    try {
        const db = await getDB();
        await new Promise<void>((resolve, reject) => {
            const tx = db.transaction(STORE_NAME, 'readwrite');
            const store = tx.objectStore(STORE_NAME);
            const req = store.clear();
            req.onsuccess = () => resolve();
            req.onerror = () => reject(req.error);
        });
    } catch {
        // Ignore
    }
}

function delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

async function fetchJSON<T>(path: string, opts: { signal?: AbortSignal; retries?: number; cacheTtlMs?: number } = {}): Promise<T> {
    const cacheKey = `fetch:${path}`;
    if (!opts.signal && opts.cacheTtlMs) {
        const cached = await cacheGet<T>(cacheKey);
        if (cached) return cached;
    }

    const retries = opts.retries ?? 1;
    let lastErr: unknown = new Error('Error de red');
    for (let attempt = 0; attempt <= retries; attempt++) {
        try {
            const res = await fetch(`${API_BASE}${path}`, {
                headers: { 'User-Agent': USER_AGENT, Accept: 'application/json' },
                signal: opts.signal,
            });
            if (res.status === 429 || res.status >= 500) {
                lastErr = new Error(`Modrinth respondió ${res.status}. Inténtalo de nuevo.`);
                if (attempt < retries) await delay(600 * (attempt + 1));
                continue;
            }
            if (!res.ok) throw new Error(`Modrinth API ${res.status}`);
            const data = (await res.json()) as T;
            if (!opts.signal && opts.cacheTtlMs) {
                await cacheSet(cacheKey, data, opts.cacheTtlMs);
            }
            return data;
        } catch (err) {
            if ((err as Error).name === 'AbortError') throw err;
            lastErr = err;
            if (attempt < retries) await delay(400 * (attempt + 1));
        }
    }
    throw lastErr;
}

// ---------- Search History (localStorage) ----------
const SEARCH_HISTORY_KEY = 'modrinth_search_history';
const MAX_HISTORY = 20;

export function getSearchHistory(): string[] {
    try {
        const raw = localStorage.getItem(SEARCH_HISTORY_KEY);
        return raw ? JSON.parse(raw) : [];
    } catch {
        return [];
    }
}

export function addSearchHistory(query: string): void {
    const trimmed = query.trim();
    if (!trimmed) return;
    const history = getSearchHistory().filter((q) => q.toLowerCase() !== trimmed.toLowerCase());
    history.unshift(trimmed);
    if (history.length > MAX_HISTORY) history.pop();
    try {
        localStorage.setItem(SEARCH_HISTORY_KEY, JSON.stringify(history));
    } catch {
        // Ignore quota errors
    }
}

export function clearSearchHistory(): void {
    try {
        localStorage.removeItem(SEARCH_HISTORY_KEY);
    } catch {
        // Ignore
    }
}

// ---------- Funciones de bajo nivel (sin estado, reutilizables) ----------

// Valores de `environment` compatibles con el CLIENTE del juego. Se excluyen
// los mods de servidor puro (server_only, dedicated_server_only, ...).
export const CLIENT_ENVIRONMENTS = [
    'client_only',
    'client_and_server',
    'client_only_server_optional',
    'singleplayer_only',
    'client_or_server',
    'client_or_server_prefers_both',
] as const;

// Construye las facets de búsqueda para el panel: tipos de proyecto permitidos,
// entorno de cliente y los filtros opcionales (loader y versión de Minecraft).
export function buildSearchFacets(opts: {
    types?: ModrinthProjectType[] | ModrinthProjectType | 'all';
    loader?: string;
    mcVersion?: string;
}): string[][] {
    const wanted = Array.isArray(opts.types) ? opts.types : opts.types === 'all' ? ['mod', 'modpack', 'shader', 'resourcepack'] : opts.types ? [opts.types] : ['mod', 'modpack', 'shader', 'resourcepack'];
    const facets: string[][] = [wanted.map((t) => `project_type:${t}`)];
    facets.push(CLIENT_ENVIRONMENTS.map((env) => `environment:${env}`));
    if (opts.loader) facets.push([`categories:${opts.loader}`]);
    if (opts.mcVersion) facets.push([`versions:${opts.mcVersion}`]);
    return facets;
}

export interface SearchProjectsOptions {
    query?: string;
    facets?: string[][];
    index?: SearchIndex;
    offset?: number;
    limit?: number;
    signal?: AbortSignal;
}

export async function fetchSearchProjects(opts: SearchProjectsOptions = {}): Promise<SearchResponse> {
    const qs = new URLSearchParams();
    if (opts.query) qs.set('query', opts.query);
    if (opts.facets?.length) qs.set('facets', JSON.stringify(opts.facets));
    qs.set('index', opts.index ?? 'relevance');
    qs.set('offset', String(opts.offset ?? 0));
    qs.set('limit', String(opts.limit ?? 20));
    // Cache search results for 5 minutes (only for first page, no query)
    const cacheTtlMs = opts.offset === 0 && !opts.query ? 5 * 60 * 1000 : undefined;
    return fetchJSON<SearchResponse>(`/search?${qs.toString()}`, { signal: opts.signal, cacheTtlMs });
}

export async function fetchProject(slugOrId: string, opts: { signal?: AbortSignal } = {}): Promise<ModrinthProject> {
    const cacheKey = `project:${slugOrId}`;
    const cached = await cacheGet<ModrinthProject>(cacheKey);
    if (cached && !opts.signal) return cached;

    const project = await fetchJSON<ModrinthProject>(`/project/${encodeURIComponent(slugOrId)}`, { ...opts, cacheTtlMs: 30 * 60 * 1000 });
    // Algunos proyectos no traen el cuerpo embebido: lo resuelve body_url.
    if (!project.body && project.body_url) {
        try {
            const res = await fetch(project.body_url, { signal: opts.signal });
            if (res.ok) project.body = await res.text();
        } catch {
            project.body = null;
        }
    }
    if (!opts.signal) await cacheSet(cacheKey, project, 30 * 60 * 1000);
    return project;
}

export async function fetchProjectVersions(
    slugOrId: string,
    opts: { includeChangelog?: boolean; signal?: AbortSignal } = {}
): Promise<ModrinthVersion[]> {
    const cacheKey = `versions:${slugOrId}:${opts.includeChangelog}`;
    const cached = await cacheGet<ModrinthVersion[]>(cacheKey);
    if (cached && !opts.signal) return cached;

    const include = opts.includeChangelog ?? false;
    const versions = await fetchJSON<ModrinthVersion[]>(
        `/project/${encodeURIComponent(slugOrId)}/version?include_changelog=${include}`,
        { ...opts, cacheTtlMs: 10 * 60 * 1000 }
    );
    if (!opts.signal) await cacheSet(cacheKey, versions, 10 * 60 * 1000);
    return versions;
}

export async function fetchVersion(id: string, opts: { signal?: AbortSignal } = {}): Promise<ModrinthVersion> {
    const cacheKey = `version:${id}`;
    const cached = await cacheGet<ModrinthVersion>(cacheKey);
    if (cached && !opts.signal) return cached;

    const version = await fetchJSON<ModrinthVersion>(`/version/${encodeURIComponent(id)}`, { ...opts, cacheTtlMs: 30 * 60 * 1000 });
    if (!opts.signal) await cacheSet(cacheKey, version, 30 * 60 * 1000);
    return version;
}

// Versiones de Minecraft release y snapshot, ordenadas de la más reciente a la
// más antigua. Se cachea 1 hora (cambian pocas veces al día).
export async function fetchGameVersions(opts: { signal?: AbortSignal } = {}): Promise<GameVersion[]> {
    const key = 'tag:game_version';
    const cached = await cacheGet<GameVersion[]>(key);
    if (cached) return cached;
    const all = await fetchJSON<GameVersion[]>('/tag/game_version', { ...opts, cacheTtlMs: 60 * 60 * 1000 });
    const filtered = all
        .filter((v) => v.version_type === 'release' || v.version_type === 'snapshot')
        .sort((a, b) => b.date.localeCompare(a.date));
    await cacheSet(key, filtered, 60 * 60 * 1000);
    return filtered;
}

export async function fetchLoaders(opts: { signal?: AbortSignal } = {}): Promise<LoaderTag[]> {
    const key = 'tag:loader';
    const cached = await cacheGet<LoaderTag[]>(key);
    if (cached) return cached;
    const loaders = await fetchJSON<LoaderTag[]>('/tag/loader', { ...opts, cacheTtlMs: 60 * 60 * 1000 });
    await cacheSet(key, loaders, 60 * 60 * 1000);
    return loaders;
}

// ---------- Helpers de formato ----------

export function formatCount(n: number | null | undefined): string {
    if (n == null || !Number.isFinite(n)) return '0';
    if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1).replace(/\.0$/, '')}M`;
    if (n >= 1_000) return `${Math.round(n / 1_000)}K`;
    return String(n);
}

export function formatBytes(bytes: number): string {
    if (!Number.isFinite(bytes) || bytes <= 0) return '0 MB';
    if (bytes >= 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
    if (bytes >= 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    return `${Math.max(1, Math.round(bytes / 1024))} KB`;
}

export function formatDate(iso: string): string {
    if (!iso) return '';
    try {
        return new Intl.DateTimeFormat('es-ES', { day: 'numeric', month: 'short', year: 'numeric' }).format(new Date(iso));
    } catch {
        return iso;
    }
}

// ---------- Cliente con estado (composable) ----------

export interface SearchState {
    hits: SearchHit[];
    totalHits: number;
    offset: number;
    page: number;
    totalPages: number;
    loading: boolean;
    loadingMore: boolean;
    error: string;
}

export interface ProjectState {
    project: ModrinthProject | null;
    versions: ModrinthVersion[];
    loading: boolean;
    error: string;
}

export interface ChangelogState {
    versionId: string;
    versionNumber: string;
    changelog: string;
    loading: boolean;
    error: string;
}

export interface TagsState {
    gameVersions: GameVersion[];
    loaders: LoaderTag[];
    loading: boolean;
    loaded: boolean;
    error: string;
}

export interface SearchCallOptions {
    query?: string;
    facets?: string[][];
    index?: SearchIndex;
    limit?: number;
}

export interface ModrinthClient {
    readonly searchState: DeepReadonly<SearchState>;
    readonly projectState: DeepReadonly<ProjectState>;
    readonly changelogState: DeepReadonly<ChangelogState>;
    readonly tagsState: DeepReadonly<TagsState>;
    readonly searchHistory: Ref<string[]>;
    search(opts: SearchCallOptions, debounceMs?: number): Promise<void>;
    goToPage(page: number): Promise<void>;
    loadMore(): Promise<void>;
    loadProject(slugOrId: string): Promise<void>;
    clearProject(): void;
    loadChangelog(versionId: string): Promise<void>;
    clearChangelog(): void;
    loadTags(): Promise<void>;
    abortSearch(): void;
    clearCache(): Promise<void>;
}

export function createModrinthClient(): ModrinthClient {
    const searchState = reactive<SearchState>({ hits: [], totalHits: 0, offset: 0, page: 1, totalPages: 1, loading: false, loadingMore: false, error: '' });
    const projectState = reactive<ProjectState>({ project: null, versions: [], loading: false, error: '' });
    const changelogState = reactive<ChangelogState>({ versionId: '', versionNumber: '', changelog: '', loading: false, error: '' });
    const tagsState = reactive<TagsState>({ gameVersions: [], loaders: [], loading: false, loaded: false, error: '' });
    const searchHistory = ref<string[]>(getSearchHistory());

    let searchTimer: ReturnType<typeof setTimeout> | null = null;
    let searchAbort: AbortController | null = null;
    let lastSearch: SearchCallOptions = {};
    let prefetchAbort: AbortController | null = null;

    function runSearch(opts: SearchCallOptions, offset: number, page: number, append = false): Promise<void> {
        searchAbort?.abort();
        const abort = new AbortController();
        searchAbort = abort;
        if (append) {
            searchState.loadingMore = true;
        } else {
            searchState.loading = true;
            // NO limpiamos hits aquí - mantenemos los resultados anteriores visibles durante debounce
        }
        searchState.error = '';
        return fetchSearchProjects({
            query: opts.query,
            facets: opts.facets,
            index: opts.index,
            offset,
            limit: opts.limit,
            signal: abort.signal,
        })
            .then((res) => {
                if (append) {
                    searchState.hits = [...searchState.hits, ...res.hits];
                } else {
                    searchState.hits = res.hits;
                }
                searchState.offset = res.offset;
                searchState.totalHits = res.total_hits;
                searchState.page = page;
                searchState.totalPages = Math.max(1, Math.ceil(res.total_hits / (opts.limit ?? 20)));
            })
            .catch((err: unknown) => {
                if ((err as Error).name === 'AbortError') return;
                searchState.error = (err as Error)?.message ?? 'Error al buscar en Modrinth';
            })
            .finally(() => {
                if (abort.signal.aborted) return;
                searchState.loading = false;
                searchState.loadingMore = false;
            });
    }

    async function search(opts: SearchCallOptions, debounceMs = 0): Promise<void> {
        lastSearch = opts;
        if (searchTimer !== null) {
            clearTimeout(searchTimer);
            searchTimer = null;
        }
        if (debounceMs > 0) {
            // Durante el debounce: MANTENEMOS resultados actuales, solo mostramos loading sutil
            // NO limpiamos searchState.hits - el usuario ve los resultados anteriores mientras escribe
            searchState.error = '';
            searchState.loading = true;
            searchTimer = setTimeout(() => {
                searchTimer = null;
                void runSearch(opts, 0, 1, false);
            }, debounceMs);
            return;
        }
        await runSearch(opts, 0, 1, false);
    }

    async function goToPage(page: number): Promise<void> {
        if (searchState.loading || searchState.loadingMore) return;
        const target = Math.min(Math.max(1, page), searchState.totalPages);
        if (target === searchState.page) return;
        const perPage = lastSearch.limit ?? 20;
        await runSearch(lastSearch, (target - 1) * perPage, target, false);
    }

    async function loadMore(): Promise<void> {
        if (searchState.loading || searchState.loadingMore) return;
        if (searchState.page >= searchState.totalPages) return;
        const nextPage = searchState.page + 1;
        const perPage = lastSearch.limit ?? 20;
        await runSearch(lastSearch, (nextPage - 1) * perPage, nextPage, true);
    }

    function abortSearch(): void {
        if (searchTimer !== null) {
            clearTimeout(searchTimer);
            searchTimer = null;
        }
        searchAbort?.abort();
        searchState.loading = false;
        searchState.loadingMore = false;
    }

    // Prefetch next page in background
    async function prefetchNextPage(): Promise<void> {
        if (searchState.page >= searchState.totalPages) return;
        if (prefetchAbort) prefetchAbort.abort();
        const abort = new AbortController();
        prefetchAbort = abort;
        const nextPage = searchState.page + 1;
        const perPage = lastSearch.limit ?? 20;
        try {
            await fetchSearchProjects({
                query: lastSearch.query,
                facets: lastSearch.facets,
                index: lastSearch.index,
                offset: (nextPage - 1) * perPage,
                limit: perPage,
                signal: abort.signal,
            });
        } catch {
            // Ignore prefetch errors
        }
    }

    async function loadProject(slugOrId: string): Promise<void> {
        projectState.loading = true;
        projectState.error = '';
        const abort = new AbortController();
        try {
            const [project, versions] = await Promise.all([
                fetchProject(slugOrId, { signal: abort.signal }),
                fetchProjectVersions(slugOrId, { includeChangelog: false, signal: abort.signal }),
            ]);
            projectState.project = project;
            projectState.versions = versions.sort((a, b) => b.date_published.localeCompare(a.date_published));
        } catch (err) {
            if ((err as Error).name === 'AbortError') return;
            projectState.error = (err as Error)?.message ?? 'No se pudo cargar el proyecto';
        } finally {
            if (!abort.signal.aborted) projectState.loading = false;
        }
    }

    function clearProject(): void {
        projectState.project = null;
        projectState.versions = [];
        projectState.error = '';
        projectState.loading = false;
        clearChangelog();
    }

    async function loadChangelog(versionId: string): Promise<void> {
        if (changelogState.loading) return;
        changelogState.loading = true;
        changelogState.error = '';
        try {
            const v = await fetchVersion(versionId);
            changelogState.versionId = v.id;
            changelogState.versionNumber = v.version_number;
            changelogState.changelog = v.changelog ?? '';
        } catch (err) {
            changelogState.error = (err as Error)?.message ?? 'No se pudo cargar el changelog';
        } finally {
            changelogState.loading = false;
        }
    }

    function clearChangelog(): void {
        changelogState.versionId = '';
        changelogState.versionNumber = '';
        changelogState.changelog = '';
        changelogState.error = '';
        changelogState.loading = false;
    }

    async function loadTags(): Promise<void> {
        if (tagsState.loading || tagsState.loaded) return;
        tagsState.loading = true;
        tagsState.error = '';
        try {
            const [gameVersions, loaders] = await Promise.all([fetchGameVersions(), fetchLoaders()]);
            tagsState.gameVersions = gameVersions;
            tagsState.loaders = loaders;
            tagsState.loaded = true;
        } catch (err) {
            tagsState.error = (err as Error)?.message ?? 'No se pudieron cargar los filtros';
        } finally {
            tagsState.loading = false;
        }
    }

    async function clearCache(): Promise<void> {
        await cacheClear();
    }

    return {
        searchState: readonly(searchState),
        projectState: readonly(projectState),
        changelogState: readonly(changelogState),
        tagsState: readonly(tagsState),
        // Historial mutable a propósito: los componentes lo mutan vía splice para feedback instantáneo
        searchHistory: searchHistory as Ref<string[]>,
        search,
        goToPage,
        loadMore,
        loadProject,
        clearProject,
        loadChangelog,
        clearChangelog,
        loadTags,
        abortSearch,
        clearCache,
    };
}

// Instancia singleton del dominio: todos los componentes de Mods comparten
// la misma búsqueda, proyecto y changelog en curso.
let shared: ModrinthClient | null = null;

export function useModrinth(): ModrinthClient {
    if (!shared) shared = createModrinthClient();
    return shared;
}