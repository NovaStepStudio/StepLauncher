import { ref } from 'vue';
import type { LocalTrack } from './LocalStore';

export interface HistoryEntry {
    path: string;
    title: string;
    artist: string;
    coverUrl?: string;
    playedAt: string;
    playCount: number;
}

export const musicHistory = ref<HistoryEntry[]>([]);
export const historyLoading = ref(false);

export async function loadHistory(): Promise<void> {
    historyLoading.value = true;
    try {
        const { GetMusicHistory } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        const list = await GetMusicHistory();
        musicHistory.value = Array.isArray(list) ? (list as HistoryEntry[]) : [];
    } catch {
        // Fallback localStorage
        try {
            const raw = localStorage.getItem('stl_music_history');
            musicHistory.value = raw ? JSON.parse(raw) as HistoryEntry[] : [];
        } catch { musicHistory.value = []; }
    } finally {
        historyLoading.value = false;
    }
}

export async function addToHistory(track: LocalTrack): Promise<void> {
    const entry: HistoryEntry = {
        path: track.path,
        title: track.title,
        artist: track.artist,
        coverUrl: track.coverUrl,
        playedAt: new Date().toISOString(),
        playCount: 1,
    };
    // Optimista local
    const idx = musicHistory.value.findIndex((e) => e.path === track.path);
    if (idx >= 0) {
        const cur = musicHistory.value[idx]!;
        cur.playCount += 1;
        cur.playedAt = entry.playedAt;
        cur.title = track.title;
        cur.artist = track.artist;
        cur.coverUrl = track.coverUrl;
        musicHistory.value.splice(idx, 1);
        musicHistory.value.unshift(cur);
    } else {
        musicHistory.value.unshift(entry);
        if (musicHistory.value.length > 100) musicHistory.value = musicHistory.value.slice(0, 100);
    }
    try { localStorage.setItem('stl_music_history', JSON.stringify(musicHistory.value)); } catch (_e) {}
    try {
        const { AddMusicHistory } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        const list = await AddMusicHistory(track.path, track.title, track.artist, track.coverUrl || '');
        if (Array.isArray(list)) musicHistory.value = list as HistoryEntry[];
    } catch (_e) {}
}

export async function clearHistory(): Promise<void> {
    musicHistory.value = [];
    try { localStorage.removeItem('stl_music_history'); } catch (_e) {}
    try {
        const { ClearMusicHistory } = await import('@wailsjs/StepLauncher/internal/Services/Music/musicservice');
        await ClearMusicHistory();
    } catch (_e) {}
}
