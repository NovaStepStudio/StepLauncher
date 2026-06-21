package dev.novastep.core.downloader.model;

public final class DownloadResult {
    public final DownloadTask task;
    public final boolean success;
    public final String error;
    public final long bytesWritten;
    public final boolean skipped;
    public final boolean sha1Passed;

    private DownloadResult(DownloadTask task, boolean success, String error, long bytesWritten, boolean skipped,
            boolean sha1Passed) {
        this.task = task;
        this.success = success;
        this.error = error;
        this.bytesWritten = bytesWritten;
        this.skipped = skipped;
        this.sha1Passed = sha1Passed;
    }

    public static DownloadResult success(DownloadTask task, long bytes, boolean sha1Passed) {
        return new DownloadResult(task, true, null, bytes, false, sha1Passed);
    }

    public static DownloadResult skipped(DownloadTask task) {
        return new DownloadResult(task, true, null, task.expectedSize, true, true);
    }

    public static DownloadResult failure(DownloadTask task, String errorMessage) {
        return new DownloadResult(task, false, errorMessage, 0, false, false);
    }

    public boolean isFailed() {
        return !success;
    }

    @Override
    public String toString() {
        if (success && skipped)
            return String.format("DownloadResult{SKIPPED, file='%s'}", task.name);
        if (success)
            return String.format("DownloadResult{OK, file='%s', bytes=%d}", task.name, bytesWritten);
        return String.format("DownloadResult{FAILED, file='%s', error='%s'}", task.name, error);
    }
}
