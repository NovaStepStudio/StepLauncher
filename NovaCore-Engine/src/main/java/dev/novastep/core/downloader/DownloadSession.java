package dev.novastep.core.downloader;
import dev.novastep.core.downloader.model.DownloadResult;
import dev.novastep.core.downloader.model.DownloadTask;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicLong;

public class DownloadSession {

    public enum Status {
        PENDING, RUNNING, PAUSED, CANCELLED, COMPLETED, FAILED
    }

    private final String sessionId;
    private final long   createdAt;
    private volatile Status status      = Status.PENDING;
    private volatile String errorDetail = null;

    private final AtomicInteger totalFiles     = new AtomicInteger(0);
    private final AtomicInteger completedFiles = new AtomicInteger(0);
    private final AtomicInteger failedFiles    = new AtomicInteger(0);
    private final AtomicInteger skippedFiles   = new AtomicInteger(0);

    private final AtomicLong totalBytes      = new AtomicLong(0);
    private final AtomicLong downloadedBytes = new AtomicLong(0);

    private final ConcurrentHashMap<String, FileState> fileStates = new ConcurrentHashMap<>();

    private final DownloadControl control = new DownloadControl();

    private volatile Map<String, Object> lastPersistedSnapshot = null;

    public DownloadSession(String sessionId) {
        this.sessionId = sessionId;
        this.createdAt = System.currentTimeMillis();
    }

    public void markRunning() {
        this.status = Status.RUNNING;
    }

    public void markCompleted() {
        this.status = Status.COMPLETED;
    }

    public void markFailed(String detail) {
        this.status      = Status.FAILED;
        this.errorDetail = detail;
    }

    public synchronized boolean pause() {
        if (status != Status.RUNNING) return false;
        status = Status.PAUSED;
        control.pause();
        return true;
    }

    public synchronized boolean resume() {
        if (status != Status.PAUSED) return false;
        status = Status.RUNNING;
        control.resume();
        return true;
    }

    public synchronized boolean cancel() {
        if (status == Status.COMPLETED || status == Status.CANCELLED) return false;
        status = Status.CANCELLED;
        errorDetail = "Cancelled by user";
        control.cancel();
        return true;
    }

    public DownloadControl getControl() { return control; }

    public void registerTask(DownloadTask task) {
        totalFiles.incrementAndGet();
        if (task.expectedSize > 0)
            totalBytes.addAndGet(task.expectedSize);
        fileStates.put(fileKey(task), new FileState(
                task.category, task.name, task.url,
                task.destination.toString(), task.expectedSize));
    }

    public void addDownloadedBytes(DownloadTask task, long chunkBytes) {
        if (chunkBytes <= 0) return;
        downloadedBytes.addAndGet(chunkBytes);
        FileState state = fileStates.get(fileKey(task));
        if (state != null) {
            state.downloaded += chunkBytes;
            state.status = "downloading";
            if (task.expectedSize > 0) {
                state.percent = (int) Math.min(99, state.downloaded * 100L / task.expectedSize);
            }
        }
    }

    public void applyResult(DownloadResult result) {
        FileState state = fileStates.get(fileKey(result.task));
        if (result.success) {
            if (result.skipped) {
                skippedFiles.incrementAndGet();
                downloadedBytes.addAndGet(result.task.expectedSize);
                if (state != null) {
                    state.status   = "skipped";
                    state.percent  = 100;
                    state.downloaded = result.task.expectedSize;
                }
            } else {
                completedFiles.incrementAndGet();
                if (state != null) {
                    state.status   = "done";
                    state.percent  = 100;
                    state.downloaded = result.bytesWritten;
                }
            }
        } else {
            failedFiles.incrementAndGet();
            if (state != null) {
                state.status = "failed";
                state.error  = result.error;
            }
        }
        lastPersistedSnapshot = toSnapshot();
    }

    public Map<String, Object> toSnapshot() {
        int total    = totalFiles.get();
        int done     = completedFiles.get();
        int failed   = failedFiles.get();
        int skipped  = skippedFiles.get();
        int pending  = Math.max(0, total - done - failed - skipped);
        long tBytes  = totalBytes.get();
        long dBytes  = downloadedBytes.get();
        int percent  = tBytes > 0 ? (int) Math.min(100, dBytes * 100L / tBytes) : 0;

        Map<String, Object> snap = new LinkedHashMap<>();
        snap.put("sessionId",      sessionId);
        snap.put("status",         status.name().toLowerCase());
        snap.put("createdAt",      createdAt);
        snap.put("totalFiles",     total);
        snap.put("completedFiles", done);
        snap.put("skippedFiles",   skipped);
        snap.put("failedFiles",    failed);
        snap.put("pendingFiles",   pending);
        snap.put("totalBytes",     tBytes);
        snap.put("downloadedBytes", dBytes);
        snap.put("overallPercent", percent);
        if (errorDetail != null)
            snap.put("error", errorDetail);
        return snap;
    }

    public Map<String, Object> getLastPersistedSnapshot() {
        return lastPersistedSnapshot != null ? lastPersistedSnapshot : toSnapshot();
    }

    public List<Map<String, Object>> getFilesByCategory(String category) {
        List<Map<String, Object>> result = new ArrayList<>();
        for (FileState state : fileStates.values()) {
            if (category.equals(state.category))
                result.add(state.toMap());
        }
        result.sort(Comparator.comparing(m -> (String) m.get("status")));
        return result;
    }

    public String  getSessionId()   { return sessionId; }
    public Status  getStatus()      { return status; }
    public boolean isRunning()      { return status == Status.RUNNING; }
    public boolean isPaused()       { return status == Status.PAUSED; }
    public boolean isCancelled()    { return status == Status.CANCELLED; }
    public int     getTotalFiles()  { return totalFiles.get(); }
    public int     getFailedFiles() { return failedFiles.get(); }

    static class FileState {
        final String category, name, url, destination;
        final long   totalSize;
        volatile long   downloaded = 0;
        volatile int    percent    = 0;
        volatile String status     = "pending";
        volatile String error      = null;

        FileState(String category, String name, String url, String destination, long totalSize) {
            this.category    = category;
            this.name        = name;
            this.url         = url;
            this.destination = destination;
            this.totalSize   = totalSize;
        }

        Map<String, Object> toMap() {
            Map<String, Object> m = new LinkedHashMap<>();
            m.put("file",        name);
            m.put("progress",    percent);
            m.put("size",        totalSize);
            m.put("downloaded",  downloaded);
            m.put("destination", destination);
            m.put("status",      status);
            m.put("url",         url);
            if (error != null) m.put("error", error);
            return m;
        }
    }

    private static String fileKey(DownloadTask task) {
        return task.category + ":" + task.name;
    }
}
