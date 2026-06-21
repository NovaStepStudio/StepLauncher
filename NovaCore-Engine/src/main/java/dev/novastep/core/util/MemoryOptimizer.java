package dev.novastep.core.util;

import dev.novastep.core.downloader.DownloadSession;

import java.util.ArrayList;
import java.util.Collection;
import java.util.Comparator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public final class MemoryOptimizer {

    private MemoryOptimizer() {
    }

    public static <K, V> Map<K, V> newSynchronizedLruCache(int maxEntries) {
        return java.util.Collections.synchronizedMap(new LinkedHashMap<>(16, 0.75f, true) {
            @Override
            protected boolean removeEldestEntry(Map.Entry<K, V> eldest) {
                return size() > maxEntries;
            }
        });
    }

    public static void trimCompletedSessions(ConcurrentHashMap<String, DownloadSession> sessions, int maxRetained) {
        if (sessions.size() <= maxRetained) {
            return;
        }

        List<DownloadSession> completed = new ArrayList<>();
        for (DownloadSession session : sessions.values()) {
            DownloadSession.Status status = session.getStatus();
            if (status == DownloadSession.Status.COMPLETED
                    || status == DownloadSession.Status.CANCELLED
                    || status == DownloadSession.Status.FAILED) {
                completed.add(session);
            }
        }

        completed.sort(Comparator.comparingLong(session ->
                ((Number) session.toSnapshot().getOrDefault("createdAt", 0L)).longValue()));

        int removable = Math.max(0, sessions.size() - maxRetained);
        for (int i = 0; i < removable && i < completed.size(); i++) {
            sessions.remove(completed.get(i).getSessionId());
        }
    }

    public static List<String> sessionIds(Collection<DownloadSession> sessions, DownloadSession.Status status) {
        List<String> result = new ArrayList<>();
        for (DownloadSession session : sessions) {
            if (session.getStatus() == status) {
                result.add(session.getSessionId());
            }
        }
        return result;
    }
}
