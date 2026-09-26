import { ref } from 'vue';
import type { ModrinthProjectType } from '@/Common/Composables/useModrinth';

// ============================================================================
// Dominio Mods — Estado de UI del navegador de contenido de Modrinth.
// La lógica de datos (búsqueda, proyecto, changelog, tags) vive en
// Common/Composables/useModrinth.ts (reutilizable); aquí solo el estado del panel.
// ============================================================================

export type ModTypeKey = 'all' | ModrinthProjectType;

export interface ModTypeTab {
    type: ModTypeKey;
    label: string;
    description: string;
}

export const MOD_TYPE_TABS: ModTypeTab[] = [
    { type: 'all', label: 'Todos', description: 'Mods, modpacks, shaders y texturas' },
    { type: 'mod', label: 'Mods', description: 'Mejoras y optimizaciones del juego' },
    { type: 'modpack', label: 'Modpacks', description: 'Colecciones completas de mods' },
    { type: 'shader', label: 'Shaders', description: 'Iluminación y gráficos avanzados' },
    { type: 'resourcepack', label: 'Texturas', description: 'Resource packs visuales' },
];

// Modloaders soportados por el launcher para el filtro lateral.
export interface LoaderOption {
    id: string;
    label: string;
}

export const LOADER_OPTIONS: LoaderOption[] = [
    { id: 'vanilla', label: 'Vanilla' },
    { id: 'fabric', label: 'Fabric' },
    { id: 'forge', label: 'Forge' },
    { id: 'quilt', label: 'QuiltMC' },
    { id: 'neoforge', label: 'NeoForge' },
    { id: 'legacy-fabric', label: 'Legacy Fabric' },
];

// Cómo ordenar los resultados (índices de la API de Modrinth).
// '' = sin índice: Modrinth devuelve lo que encuentre (orden por defecto).
export interface SortOption {
    id: 'any' | 'relevance' | 'downloads' | 'follows';
    label: string;
}

export const SORT_OPTIONS: SortOption[] = [
    { id: 'any', label: 'Cualquiera' },
    { id: 'relevance', label: 'Relevancia' },
    { id: 'downloads', label: 'Más descargas' },
    { id: 'follows', label: 'Más seguidores' },
];

// Cantidades de elementos a cargar por página.
export const LIMIT_OPTIONS = [5, 10, 15, 20, 50, 100] as const;

// Tipos de versión de un proyecto (canales de publicación).
export const VERSION_TYPES: Array<{ id: 'release' | 'beta' | 'alpha'; label: string }> = [
    { id: 'release', label: 'Release' },
    { id: 'beta', label: 'Beta' },
    { id: 'alpha', label: 'Alpha' },
];

// ---------- Facets server-side para Modrinth ----------

/**
 * Construye las facets de búsqueda de Modrinth usando `project_type` en el
 * servidor. Ejemplo: `facets=[["project_type:resourcepack"]]` para texturas.
 * Para `type === 'all'` no se añade filtro de project_type (el servidor
 * devuelve todos los tipos). No se filtra por `environment` para no limitar
 * resultados — el entorno se muestra como etiqueta en cada card.
 * Nunca se filtra por `hit.project_type` en cliente.
 */
export function buildModsFacets(opts: {
    type: ModTypeKey;
    loader?: string;
    mcVersion?: string;
}): string[][] {
    const facets: string[][] = [];
    if (opts.type !== 'all') {
        facets.push([`project_type:${opts.type}`]);
    }
    if (opts.loader) facets.push([`categories:${opts.loader}`]);
    if (opts.mcVersion) facets.push([`versions:${opts.mcVersion}`]);
    return facets;
}

/**
 * Etiqueta legible del entorno de un proyecto para mostrar como tag.
 * Devuelve null si es solo cliente (caso más común) para no saturar la UI.
 * Ejemplos: "Para servidor" (server_only), "Cliente y servidor".
 */
export function environmentLabel(envs: readonly string[] | undefined | null): string | null {
    if (!envs || !envs.length) return null;
    const joined = envs.join(' ').toLowerCase();
    const hasServer = joined.includes('server');
    const hasClient = joined.includes('client');
    if (hasServer && hasClient) return 'Cliente y servidor';
    if (hasServer) return 'Para servidor';
    // Solo cliente es el caso por defecto en Modrinth, no necesita tag
    return null;
}

// ---------- Estado reactivo del panel ----------

export const activeTab = ref<ModTypeKey>('all');
export const viewMode = ref<'grid' | 'list'>('grid');

// Instancia preferida para el diálogo de descarga (la pone quien abre Mods
// desde el detalle de una instancia con "Añadir"). El diálogo la usa para
// preseleccionar el destino y la conserva durante la sesión.
export const pendingInstance = ref<string | null>(null);

export function preferInstance(name: string | null): void {
    pendingInstance.value = name && name.trim() ? name.trim() : null;
}

export function selectTab(tab: ModTypeKey): void {
    activeTab.value = tab;
}

export function setViewMode(mode: 'grid' | 'list'): void {
    viewMode.value = mode;
}

export function typeLabel(type: string): string {
    return MOD_TYPE_TABS.find((t) => t.type === type)?.label ?? type;
}