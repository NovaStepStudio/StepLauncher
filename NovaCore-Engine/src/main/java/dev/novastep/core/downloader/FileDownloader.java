package dev.novastep.core.downloader;
import dev.novastep.core.downloader.model.DownloadResult;
import dev.novastep.core.downloader.model.DownloadTask;

import java.io.BufferedInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.ByteBuffer;
import java.nio.channels.FileChannel;
import java.nio.file.Files;
import java.nio.file.StandardOpenOption;
import java.security.MessageDigest;
import java.time.Duration;

import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.websocket.EventBroadcaster;

public class FileDownloader {

    private static final int    BUFFER_SIZE     = 64 * 1024;
    private static final long   REPORT_INTERVAL = 128 * 1024;
    private static final int    READ_TIMEOUT    = 60;
    private static final int    MAX_RETRIES     = 3;
    private static final long   RETRY_DELAY_MS  = 1_500;
    private static final String LOG             = "FileDownloader";

    private static final String ENGINE_VERSION  = dev.novastep.core.CoreVersion.get();

    private final HttpClient       http;
    private final EventBroadcaster broadcaster;
    private final DownloadSession  session;

    private final java.util.concurrent.Semaphore httpSemaphore;

    private final DownloadPriority priority;

    public FileDownloader(HttpClient http, EventBroadcaster broadcaster, DownloadSession session) {
        this(http, broadcaster, session, null, null);
    }

    public FileDownloader(HttpClient http, EventBroadcaster broadcaster,
                          DownloadSession session,
                          java.util.concurrent.Semaphore httpSemaphore) {
        this(http, broadcaster, session, httpSemaphore, null);
    }

    public FileDownloader(HttpClient http, EventBroadcaster broadcaster,
                          DownloadSession session,
                          java.util.concurrent.Semaphore httpSemaphore,
                          DownloadPriority priority) {
        this.http          = http;
        this.broadcaster   = broadcaster;
        this.session       = session;
        this.httpSemaphore = httpSemaphore;
        this.priority      = priority;
    }

    public DownloadResult download(DownloadTask task) {
        DownloadControl ctrl = session.getControl();

        if (ctrl.isCancelled())
            return DownloadResult.failure(task, "Cancelled");

        if (Files.exists(task.destination)) {
            if (task.sha1 != null) {
                try {
                    if (Sha1Verifier.verifyFile(task.destination, task.sha1)) {
                        broadcaster.emitDownloadComplete(task.sessionId, task.category, task.name, task.expectedSize, true);
                        return DownloadResult.skipped(task);
                    }
                } catch (IOException ex) {
                    CoreLogger.get().warn(LOG, "Pre-download SHA-1 check failed: " + ex.getMessage());
                }
            } else {
                broadcaster.emitDownloadComplete(task.sessionId, task.category, task.name, task.expectedSize, true);
                return DownloadResult.skipped(task);
            }
        }

        try { Files.createDirectories(task.destination.getParent()); }
        catch (IOException e) { return DownloadResult.failure(task, "mkdir: " + e.getMessage()); }

        if (priority != null) {
            try {
                priority.acquire(task.category);
            } catch (InterruptedException ie) {
                Thread.currentThread().interrupt();
                return DownloadResult.failure(task, "Interrupted waiting for category slot");
            }
        }

        try {
            broadcaster.emitDownloadStart(task.sessionId, task.category, task.name, task.expectedSize);

            Exception lastError = null;
            for (int attempt = 1; attempt <= MAX_RETRIES; attempt++) {
                try {
                    if (!ctrl.checkPoint())
                        return DownloadResult.failure(task, "Cancelled");
                } catch (InterruptedException ie) {
                    Thread.currentThread().interrupt();
                    return DownloadResult.failure(task, "Interrupted");
                }

                try {
                    return attemptDownload(task, ctrl);
                } catch (CancelledException e) {
                    return DownloadResult.failure(task, "Cancelled");
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    return DownloadResult.failure(task, "Interrupted");
                } catch (Exception e) {
                    lastError = e;
                    CoreLogger.get().warn(LOG, "Attempt " + attempt + "/" + MAX_RETRIES + " failed for "
                            + task.name + ": " + e.getMessage());
                    if (attempt < MAX_RETRIES) {
                        try { Thread.sleep(RETRY_DELAY_MS * attempt); }
                        catch (InterruptedException ie) {
                            Thread.currentThread().interrupt();
                            return DownloadResult.failure(task, "Interrupted");
                        }
                        try { 
                            Files.deleteIfExists(task.destination); 
                        } catch (IOException ex) {
                            CoreLogger.get().warn(LOG, "Failed to delete corrupt file " + task.destination + ": " + ex.getMessage());
                        }
                    }
                }
            }

            String msg = lastError != null
                    ? lastError.getClass().getSimpleName() + ": " + lastError.getMessage()
                    : "Unknown";
            CoreLogger.get().error(LOG, "Download failed: " + task.name + " → " + msg);
            broadcaster.emitDownloadError(task.sessionId, task.category, task.name, msg);
            return DownloadResult.failure(task, msg);

        } finally {
            if (priority != null) priority.release(task.category);
        }
    }

    private DownloadResult attemptDownload(DownloadTask task, DownloadControl ctrl)
            throws Exception {

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(task.url))
                .timeout(Duration.ofSeconds(READ_TIMEOUT))
                .header("User-Agent", "novacore-engine/" + ENGINE_VERSION + " (NovaStepStudios)")
                .GET().build();

        if (httpSemaphore != null) {
            try {
                httpSemaphore.acquire();
            } catch (InterruptedException ie) {
                Thread.currentThread().interrupt();
                return DownloadResult.failure(task, "Interrupted waiting for HTTP slot");
            }
        }

        try {
            HttpResponse<InputStream> response = http.send(request, HttpResponse.BodyHandlers.ofInputStream());
            if (response.statusCode() != 200)
                throw new IOException("HTTP " + response.statusCode() + " → " + task.url);

            MessageDigest digest  = task.sha1 != null ? Sha1Verifier.newDigest() : null;
            long written          = 0L;
            long lastReported     = 0L;
            byte[] buf            = new byte[BUFFER_SIZE];

            try (InputStream in = new BufferedInputStream(response.body(), BUFFER_SIZE);
                 FileChannel out = FileChannel.open(task.destination,
                         StandardOpenOption.CREATE, StandardOpenOption.WRITE,
                         StandardOpenOption.TRUNCATE_EXISTING)) {

                int read;
                while ((read = in.read(buf)) != -1) {

                    if (!ctrl.checkPoint()) {
                        try { 
                            Files.deleteIfExists(task.destination); 
                        } catch (IOException ex) {
                            CoreLogger.get().warn(LOG, "Failed to delete cancelled file " + task.destination + ": " + ex.getMessage());
                        }
                        throw new CancelledException();
                    }

                    ByteBuffer bb = ByteBuffer.wrap(buf, 0, read);
                    while (bb.hasRemaining()) out.write(bb);
                    if (digest != null) digest.update(buf, 0, read);
                    written += read;
                    session.addDownloadedBytes(task, read);

                    if (written - lastReported >= REPORT_INTERVAL) {
                        broadcaster.emitDownloadProgress(task.sessionId, task.category, task.name, written, task.expectedSize);
                        lastReported = written;
                    }
                }
            }

            boolean sha1Ok = true;
            if (digest != null && task.sha1 != null) {
                String computed = Sha1Verifier.finalize(digest);
                sha1Ok = Sha1Verifier.matches(computed, task.sha1);
                broadcaster.emitSha1Check(task.sessionId, task.name, sha1Ok, task.sha1, computed);
                if (!sha1Ok) {
                    try { 
                        Files.deleteIfExists(task.destination); 
                    } catch (IOException ex) {
                        CoreLogger.get().warn(LOG, "Failed to delete mismatched file " + task.destination + ": " + ex.getMessage());
                    }
                    throw new IOException("SHA-1 mismatch: " + task.name
                            + " expected=" + task.sha1 + " got=" + computed);
                }
            }

            broadcaster.emitDownloadComplete(task.sessionId, task.category, task.name, written, false);
            return DownloadResult.success(task, written, sha1Ok);

        } finally {
            if (httpSemaphore != null) httpSemaphore.release();
        }
    }

    private static final class CancelledException extends RuntimeException {
        CancelledException() { super(null, null, true, false); }
    }
}
