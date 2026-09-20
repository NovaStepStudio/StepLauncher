import { ref, computed } from 'vue';

// --- Tipos base para la biblioteca local (solo layout, sin lógica aún) ---

export interface LocalTrack {
    id: string;
    fileName: string;
    title: string;
    artist: string;
    album?: string;
    duration: number; // segundos
    path: string; // ruta local relativa o absoluta (futura)
    coverUrl?: string;
    addedAt: string; // ISO
}

// Estado preparado para la futura carga local.
// Por ahora solo estructura reactiva, sin lectura real de archivos.
export const localTracks = ref<LocalTrack[]>([]);
export const selectedTrackId = ref<string | null>(null);
export const isLoadingLibrary = ref(false);
export const searchQuery = ref('');

export const filteredTracks = computed(() => {
    const q = searchQuery.value.trim().toLowerCase();
    if (!q) return localTracks.value;
    return localTracks.value.filter((t) =>
        t.title.toLowerCase().includes(q) ||
        t.artist.toLowerCase().includes(q) ||
        t.fileName.toLowerCase().includes(q)
    );
});

export const selectedTrack = computed(() =>
    localTracks.value.find((t) => t.id === selectedTrackId.value) ?? null
);

// --- Placeholders para la futura lógica de carga local ---
// TODO: implementar con backend Go (ReadMusicFile / diálogo nativo) cuando se habilite.

export async function loadLocalLibrary(): Promise<void> {
    // Placeholder: aquí se cargará la lista desde disco / backend.
    // De momento no hace nada para mantener el layout desacoplado.
    isLoadingLibrary.value = false;
}

export async function addLocalTracks(): Promise<string | null> {
    // TODO: abrir diálogo nativo, validar y copiar a cache/audio.
    return 'Función aún no implementada — solo layout.';
}

export async function removeLocalTrack(_id: string): Promise<string | null> {
    // TODO: eliminar archivo y refrescar lista.
    return null;
}

export function selectTrack(id: string): void {
    selectedTrackId.value = id;
}

export function clearSelection(): void {
    selectedTrackId.value = null;
}
