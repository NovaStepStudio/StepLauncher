package dev.novastep.core.downloader;

import dev.novastep.core.downloader.model.DownloadResult;
import dev.novastep.core.downloader.model.DownloadTask;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.util.MemoryOptimizer;
import dev.novastep.core.util.SystemResources;
import dev.novastep.core.websocket.EventBroadcaster;

import java.net.http.HttpClient;
import java.time.Duration;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

public class DownloadManager {

    private static final int HTTP_CONNECT_TIMEOUT = 15;
    private static final int HTTP_CONCURRENCY_MULTIPLIER = 2;
    private static final int MAX_RETAINED_SESSIONS = 64;
    private static final String LOG = "DownloadManager";

    private final int maxThreads;
    private final EventBroadcaster broadcaster;
    private final ExecutorService pool;
    private final ExecutorService httpExecutor;
    private final HttpClient http;
    private final java.util.concurrent.Semaphore httpSemaphore;
    private final DownloadPriority priority;
    private final AtomicInteger sessionCounter = new AtomicInteger(0);
    private final ConcurrentHashMap<String, DownloadSession> sessions = new ConcurrentHashMap<>();

    public DownloadManager(int maxThreads, EventBroadcaster broadcaster) {
        this.maxThreads = SystemResources.safeThreads(maxThreads);
        this.broadcaster = broadcaster;

        int httpConcurrency = Math.min(
                this.maxThreads * HTTP_CONCURRENCY_MULTIPLIER,
                SystemResources.MAX_DOWNLOAD_THREADS * HTTP_CONCURRENCY_MULTIPLIER);

        this.httpSemaphore = new java.util.concurrent.Semaphore(httpConcurrency, true);
        this.priority = new DownloadPriority(this.maxThreads);
        this.pool = new ThreadPoolExecutor(
                this.maxThreads,
                this.maxThreads,
                60L,
                TimeUnit.SECONDS,
                new LinkedBlockingQueue<>(),
                runnable -> {
                    Thread thread = new Thread(runnable, "mc-dl-" + sessionCounter.get());
                    thread.setDaemon(true);
                    return thread;
                }
        );
        this.httpExecutor = Executors.newFixedThreadPool(
                Math.max(4, this.maxThreads / 2),
                runnable -> {
                    Thread thread = new Thread(runnable, "mc-http-io");
                    thread.setDaemon(true);
                    return thread;
                }
        );
        this.http = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(HTTP_CONNECT_TIMEOUT))
                .followRedirects(HttpClient.Redirect.ALWAYS)
                .executor(httpExecutor)
                .build();

        CoreLogger.get().info(LOG, "DownloadManager initialized: threads=" + this.maxThreads + " httpConcurrency=" + httpConcurrency);
    }

    public String createSession() {
        String sessionId = "session-" + System.currentTimeMillis() + "-" + sessionCounter.incrementAndGet();
        sessions.put(sessionId, new DownloadSession(sessionId));
        trimSessions();
        return sessionId;
    }

    public Optional<DownloadSession> getSession(String sessionId) {
        return Optional.ofNullable(sessions.get(sessionId));
    }

    public String registerSessionIfAbsent(String sessionId) {
        trimSessions();
        return sessions.computeIfAbsent(sessionId, DownloadSession::new).getSessionId();
    }

    public Collection<DownloadSession> getAllSessions() {
        return Collections.unmodifiableCollection(sessions.values());
    }

    public boolean pauseSession(String sessionId) {
        return getSession(sessionId).map(session -> {
            boolean ok = session.pause();
            if (ok) {
                broadcaster.emit("session_paused", Map.of("sessionId", sessionId));
            }
            return ok;
        }).orElse(false);
    }

    public boolean resumeSession(String sessionId) {
        return getSession(sessionId).map(session -> {
            boolean ok = session.resume();
            if (ok) {
                broadcaster.emit("session_resumed", Map.of("sessionId", sessionId));
            }
            return ok;
        }).orElse(false);
    }

    public boolean cancelSession(String sessionId) {
        return getSession(sessionId).map(session -> {
            boolean ok = session.cancel();
            if (ok) {
                broadcaster.emit("session_cancelled", Map.of("sessionId", sessionId));
            }
            return ok;
        }).orElse(false);
    }

    public List<String> pauseRunningSessions() {
        List<String> paused = new ArrayList<>();
        for (DownloadSession session : sessions.values()) {
            if (session.getStatus() == DownloadSession.Status.RUNNING && pauseSession(session.getSessionId())) {
                paused.add(session.getSessionId());
            }
        }
        return paused;
    }

    public void resumeSessions(List<String> sessionIds) {
        if (sessionIds == null) {
            return;
        }
        for (String sessionId : sessionIds) {
            resumeSession(sessionId);
        }
    }

    public void throttle(int slots) {
        int current  = httpSemaphore.availablePermits();
        int maxSlots = Math.min(maxThreads * HTTP_CONCURRENCY_MULTIPLIER,
                SystemResources.MAX_DOWNLOAD_THREADS * HTTP_CONCURRENCY_MULTIPLIER);
        int target   = slots < 0 ? maxSlots : Math.min(slots, maxSlots);

        int delta = target - current;
        if (delta > 0) {
            httpSemaphore.release(delta);
        } else if (delta < 0) {
            // Drain permits without blocking — only remove available slots
            httpSemaphore.tryAcquire(Math.min(-delta, httpSemaphore.availablePermits()));
        }
        CoreLogger.get().debug(LOG, "Download throttle: slots=" + target + " (was=" + current + ")");
    }


    public CompletableFuture<List<DownloadResult>> submitAll(String sessionId, List<DownloadTask> tasks) {
        DownloadSession session = sessions.get(sessionId);
        if (session == null) {
            return CompletableFuture.failedFuture(new IllegalArgumentException("Unknown session: " + sessionId));
        }

        for (DownloadTask task : tasks) {
            session.registerTask(task);
        }
        session.markRunning();

        broadcaster.emit("session_started", Map.of(
                "sessionId", sessionId,
                "totalFiles", tasks.size(),
                "totalBytes", tasks.stream().mapToLong(task -> task.expectedSize).sum()));

        FileDownloader downloader = new FileDownloader(http, broadcaster, session, httpSemaphore, priority);
        List<CompletableFuture<DownloadResult>> futures = new ArrayList<>();
        for (DownloadTask task : tasks) {
            futures.add(CompletableFuture.supplyAsync(() -> executeTask(session, task, downloader), pool));
        }

        return CompletableFuture.allOf(futures.toArray(new CompletableFuture[0]))
                .thenApply(ignored -> {
                    List<DownloadResult> results = futures.stream().map(CompletableFuture::join).toList();
                    if (!session.isCancelled()) {
                        long failedCount = results.stream().filter(DownloadResult::isFailed).count();
                        if (failedCount > 0) {
                            session.markFailed(failedCount + " file(s) failed");
                            broadcaster.emitSessionFailed(sessionId, failedCount + " of " + tasks.size() + " files failed");
                        } else {
                            session.markCompleted();
                            Map<String, Object> snapshot = session.toSnapshot();
                            broadcaster.emitSessionCompleted(sessionId, tasks.size(),
                                    (long) snapshot.get("downloadedBytes"));
                        }
                    }
                    trimSessions();
                    return results;
                });
    }

    private DownloadResult executeTask(DownloadSession session, DownloadTask task, FileDownloader downloader) {
        DownloadResult result = downloader.download(task);
        session.applyResult(result);
        Map<String, Object> snapshot = session.toSnapshot();
        broadcaster.emitSessionProgress(
                task.sessionId,
                (int) snapshot.get("completedFiles"),
                (int) snapshot.get("skippedFiles"),
                (int) snapshot.get("totalFiles"),
                (int) snapshot.get("overallPercent"),
                (long) snapshot.get("downloadedBytes"),
                (long) snapshot.get("totalBytes")
        );
        return result;
    }

    public List<Map<String, Object>> getRecoverySnapshots() {
        List<Map<String, Object>> list = new ArrayList<>();
        for (DownloadSession session : sessions.values()) {
            if (session.getStatus() != DownloadSession.Status.COMPLETED) {
                list.add(session.getLastPersistedSnapshot());
            }
        }
        return list;
    }

    public void shutdown() {
        shutdownExecutor(pool, "download pool");
        shutdownExecutor(httpExecutor, "http executor");
    }

    private void shutdownExecutor(ExecutorService executor, String name) {
        executor.shutdown();
        try {
            if (!executor.awaitTermination(10, TimeUnit.SECONDS)) {
                executor.shutdownNow();
            }
        } catch (InterruptedException ex) {
            CoreLogger.get().error(LOG, "DownloadManager " + name + " shutdown interrupted", ex);
            executor.shutdownNow();
            Thread.currentThread().interrupt();
        }
    }

    private void trimSessions() {
        MemoryOptimizer.trimCompletedSessions(sessions, MAX_RETAINED_SESSIONS);
    }
}
