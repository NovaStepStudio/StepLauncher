import { ref, reactive, computed } from 'vue';
import { EventsOn } from '@wailsjs/runtime/runtime';

export interface NewsEntry {
    version: string;
    path: string;
}

export interface NewsIndex {
    ok: boolean;
    fromCache: boolean;
    error: string;
    latest: string;
    content: NewsEntry[];
}

export interface ReleaseDetail {
    version: string;
    title: string;
    type: string;
    body: string;
    date: string;
}

export interface ChangelogResult {
    version: string;
    url: string;
    markdown: string;
}

export interface ReaderEntry {
    /** Versión de release cuando el documento es el changelog de su card. */
    version: string;
    /** URL canónica del documento (vacía hasta conocerla si es un changelog). */
    url: string;
    label: string;
}

export const indexState = ref<NewsIndex>({ ok: false, fromCache: false, error: '', latest: '', content: [] });
export const indexLoading = ref(false);

export const details = reactive(new Map<string, ReleaseDetail>());
export const detailLoading = reactive(new Map<string, boolean>());
export const releaseErrors = reactive(new Map<string, string>());

export const changelogs = reactive(new Map<string, ChangelogResult>());
export const changelogLoading = reactive(new Map<string, boolean>());
export const changelogErrors = reactive(new Map<string, string>());

/** Todos los documentos vistos (changelogs + MD de auditoría abiertos por enlace), clave = URL. */
export const docs = reactive(new Map<string, string>());
export const docLoading = reactive(new Map<string, boolean>());
export const docErrors = reactive(new Map<string, string>());

export const readerStack = ref<ReaderEntry[]>([]);
export const readerIndex = ref(-1);

export const isReading = computed(() => readerIndex.value >= 0);

export const currentEntry = computed<ReaderEntry | null>(() => {
    const i = readerIndex.value;
    if (i < 0 || i >= readerStack.value.length) return null;
    return readerStack.value[i]!;
});

export const currentDocUrl = computed<string>(() => {
    const e = currentEntry.value;
    if (!e) return '';
    if (e.url) return e.url;
    const cl = e.version ? changelogs.get(e.version) : null;
    return cl?.url ?? '';
});

export const currentMarkdown = computed<string>(() => {
    const url = currentDocUrl.value;
    return url ? (docs.get(url) ?? '') : '';
});

function parseEvent(raw: unknown): any {
    if (typeof raw === 'string') {
        try {
            return JSON.parse(raw);
        } catch {
            return null;
        }
    }
    return raw;
}

function goNews() {
    return (window as any)?.go?.main?.App;
}

let bound = false;

export function bindNewsEvents() {
    if (bound) return;
    bound = true;
    try {
        EventsOn('news_index', (raw: unknown) => {
            const p = parseEvent(raw) as NewsIndex;
            if (p) {
                indexState.value = {
                    ok: !!p.ok,
                    fromCache: !!p.fromCache,
                    error: p.error ?? '',
                    latest: p.latest ?? '',
                    content: Array.isArray(p.content) ? p.content : [],
                };
            } else {
                indexState.value = {
                    ok: false,
                    fromCache: false,
                    error: 'Respuesta inválida del servidor.',
                    latest: '',
                    content: [],
                };
            }
            indexLoading.value = false;
        });
        EventsOn('news_release', (raw: unknown) => {
            const p = parseEvent(raw) as ReleaseDetail & { ok?: boolean; error?: string; newsPath?: string };
            if (!p || !p.version) return;
            if (!p.ok) {
                detailLoading.set(p.version, false);
                releaseErrors.set(p.version, p.error ?? 'No se pudo cargar la noticia.');
                return;
            }
            details.set(p.version, {
                version: p.version,
                title: p.title ?? '',
                type: p.type ?? '',
                body: p.body ?? '',
                date: p.date ?? '',
            });
            detailLoading.set(p.version, false);
            releaseErrors.delete(p.version);
        });
        EventsOn('news_changelog', (raw: unknown) => {
            const p = parseEvent(raw) as ChangelogResult & { ok?: boolean; error?: string };
            if (!p || !p.version) return;
            if (!p.ok) {
                changelogLoading.set(p.version, false);
                changelogErrors.set(p.version, p.error ?? 'No se pudo cargar el changelog.');
                return;
            }
            changelogLoading.set(p.version, false);
            changelogErrors.delete(p.version);
            const entry: ChangelogResult = { version: p.version, url: p.url ?? '', markdown: p.markdown ?? '' };
            changelogs.set(p.version, entry);
            if (entry.url && entry.markdown) {
                docs.set(entry.url, entry.markdown);
                docLoading.set(entry.url, false);
                docErrors.delete(entry.url);
            }
        });
        EventsOn('news_markdown', (raw: unknown) => {
            const p = parseEvent(raw) as { ok?: boolean; error?: string; url?: string; markdown?: string };
            if (!p || !p.url) return;
            if (!p.ok) {
                docLoading.set(p.url, false);
                docErrors.set(p.url, p.error ?? 'No se pudo cargar el documento.');
                return;
            }
            docLoading.set(p.url, false);
            docErrors.delete(p.url);
            if (p.markdown) {
                docs.set(p.url, p.markdown);
            }
        });
    } catch { }
}

export async function refreshNews() {
    bindNewsEvents();
    indexLoading.value = true;
    try {
        await goNews()?.NewsRefreshIndex?.();
    } catch {
        indexLoading.value = false;
        indexState.value = { ...indexState.value, ok: false, error: 'No se pudo contactar el servidor de noticias.' };
    }
}

/** Precarga el news.json de cada release del índice en paralelo (Go ya corre en goroutines). */
export function preloadDetails() {
    bindNewsEvents();
    for (const entry of indexState.value.content) {
        const v = entry.version;
        if (details.has(v) || detailLoading.get(v) || releaseErrors.has(v)) continue;
        detailLoading.set(v, true);
        try {
            goNews()?.NewsLoadRelease?.(v);
        } catch {
            detailLoading.set(v, false);
            releaseErrors.set(v, 'No se pudo cargar la noticia.');
        }
    }
}

/** Recarga el detalle de una release que falló (retry por card). */
export function reloadDetail(version: string) {
    if (!version) return;
    detailLoading.set(version, true);
    releaseErrors.delete(version);
    try {
        goNews()?.NewsLoadRelease?.(version);
    } catch {
        detailLoading.set(version, false);
        releaseErrors.set(version, 'No se pudo cargar la noticia.');
    }
}

export function loadChangelog(version: string) {
    if (changelogs.has(version) || changelogLoading.get(version)) return;
    changelogLoading.set(version, true);
    changelogErrors.delete(version);
    try {
        goNews()?.NewsLoadChangelog?.(version);
    } catch {
        changelogLoading.set(version, false);
        changelogErrors.set(version, 'No se pudo solicitar el changelog.');
    }
}

export function loadMarkdown(url: string) {
    if (!url || docs.has(url) || docLoading.get(url)) return;
    docLoading.set(url, true);
    docErrors.delete(url);
    try {
        goNews()?.NewsLoadMarkdown?.(url);
    } catch {
        docLoading.set(url, false);
        docErrors.set(url, 'No se pudo solicitar el documento.');
    }
}

function detailTitle(version: string): string {
    const d = details.get(version);
    return d?.title ? d.title : `StepLauncher ${version}`;
}

/** Abre en el lector el changelog de una release (raíz de navegación para esa card). */
export function openReaderVersion(version: string) {
    readerStack.value = [{ version, url: changelogs.get(version)?.url ?? '', label: detailTitle(version) }];
    readerIndex.value = 0;
    loadChangelog(version);
}

/** Abre en el lector un documento arbitrario (enlace interno a un MD de auditoría). */
export function openReaderUrl(url: string, label: string) {
    const next = { version: '', url, label };
    const stack = readerStack.value.slice(0, readerIndex.value + 1);
    stack.push(next);
    readerStack.value = stack;
    readerIndex.value = stack.length - 1;
    loadMarkdown(url);
}

export function readerBack() {
    if (readerIndex.value > 0) {
        readerIndex.value--;
        ensureReaderDoc();
        return;
    }
    closeReader();
}

export function readerForward() {
    if (readerIndex.value < readerStack.value.length - 1) {
        readerIndex.value++;
        ensureReaderDoc();
    }
}

export function readerTo(index: number) {
    if (index >= 0 && index < readerStack.value.length) {
        readerIndex.value = index;
        ensureReaderDoc();
    }
}

export function closeReader() {
    readerStack.value = [];
    readerIndex.value = -1;
}

function ensureCurrentDoc() {
    const url = currentDocUrl.value;
    const e = currentEntry.value;
    if (!e) return;
    if (url) {
        loadMarkdown(url);
        return;
    }
    if (e.version) {
        loadChangelog(e.version);
    }
}

export function ensureReaderDoc() {
    ensureCurrentDoc();
}