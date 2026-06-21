package dev.novastep.core.downloader;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.Semaphore;

public final class DownloadPriority {

    public static final int UNLIMITED = Integer.MAX_VALUE;

    private final Map<String, Semaphore> slots = new ConcurrentHashMap<>();

    public DownloadPriority(int maxThreads) {
        int t = Math.max(4, maxThreads);

        register("client",      cap(2,            t));
        register("library",     cap(t,            t));
        register("native",      cap(t / 2,        t));
        register("asset",       cap(t * 2,        t * 4));
        register("asset_index", cap(2,            t));
    }

    public void acquire(String category) throws InterruptedException {
        Semaphore s = slots.get(category);
        if (s != null) s.acquire();
    }

    public void release(String category) {
        Semaphore s = slots.get(category);
        if (s != null) s.release();
    }

    public int availablePermits(String category) {
        Semaphore s = slots.get(category);
        return s == null ? UNLIMITED : s.availablePermits();
    }

    private void register(String category, int permits) {
        slots.put(category, new Semaphore(permits, true));
    }

    private static int cap(int desired, int max) {
        return Math.max(1, Math.min(desired, max));
    }
}
