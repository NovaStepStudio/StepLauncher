import { ref } from 'vue';

export interface Playlist {
    id: string;
    title: string;
    favorite: boolean;
    pinned: boolean;
    trackPaths: string[];
    customColor?: string;
    customCover?: string;
    createdAt: string;
    updatedAt: string;
    totalDuration?: number;
    trackCount?: number;
    previewCovers?: string[]; // thumbs 128 de las primeras 4 pistas, cacheadas en Go para lectura instantánea
}

export const playlists = ref<Playlist[]>([]);
export const loadingPlaylists = ref(false);

// Carga playlists desde backend (launcher_playlists.json)
export async function loadPlaylists(): Promise<void> {
    loadingPlaylists.value = true;
    try {
        const { ListPlaylists } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        const list = await ListPlaylists();
        playlists.value = Array.isArray(list) ? (list as Playlist[]) : [];
    } catch {
        playlists.value = [];
    } finally {
        loadingPlaylists.value = false;
    }
}

export async function createPlaylist(title: string, trackPaths: string[]): Promise<string | null> {
    try {
        const { CreatePlaylist } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        const res = await CreatePlaylist(title, trackPaths);
        if (res) await loadPlaylists();
        return null;
    } catch (e: any) {
        return e?.message ?? 'No se pudo crear la lista de reproducción';
    }
}

export async function updatePlaylist(id: string, data: Partial<Playlist>): Promise<string | null> {
    try {
        const { GetPlaylist, UpdatePlaylist } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        const cur = await GetPlaylist(id);
        if (!cur) return 'Lista de reproducción no encontrada';
        const merged = { ...cur, ...data } as Playlist;
        await UpdatePlaylist(id, merged as any);
        await loadPlaylists();
        return null;
    } catch (e: any) {
        return e?.message ?? 'No se pudo actualizar';
    }
}

export async function deletePlaylist(id: string): Promise<string | null> {
    try {
        const { DeletePlaylist } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        await DeletePlaylist(id);
        await loadPlaylists();
        return null;
    } catch (e: any) {
        return e?.message ?? 'No se pudo eliminar';
    }
}

export async function importPlaylist(): Promise<string | null> {
    try {
        const { PickPlaylistFile, ImportPlaylistFile } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        const path = await PickPlaylistFile();
        if (!path) return null;
        await ImportPlaylistFile(path);
        await loadPlaylists();
        return null;
    } catch (e: any) {
        return e?.message ?? 'No se pudo importar';
    }
}
