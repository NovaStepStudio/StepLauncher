package dev.novastep.core.minecraft;

import dev.novastep.core.downloader.model.DownloadTask;
import dev.novastep.core.downloader.Sha1Verifier;
import dev.novastep.core.log.CoreLogger;

import java.io.IOException;
import java.nio.file.Files;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

public final class InstallVerifier {

    private static final String LOG = "InstallVerifier";

    private InstallVerifier() {}

    public record VerifyResult(List<DownloadTask> bad, int checked, int sha1Skipped) {
        public boolean ok() { return bad.isEmpty(); }
        public int passed() { return checked - bad.size(); }
        public Map<String, Long> badByCategory() {
            return bad.stream().collect(
                    Collectors.groupingBy(t -> t.category, Collectors.counting()));
        }
    }

    public static VerifyResult verify(String sessionId, List<DownloadTask> tasks) {
        List<DownloadTask> bad       = new ArrayList<>();
        int                sha1Skip  = 0;

        for (DownloadTask task : tasks) {
            if (!Files.exists(task.destination)) {
                CoreLogger.get().warn(LOG,
                        "[" + sessionId + "] MISSING [" + task.category + "] " + task.name
                        + " → " + task.destination);
                bad.add(task);
                continue;
            }

            if (task.sha1 == null || task.sha1.isBlank()) {
                sha1Skip++;
                continue;
            }

            try {
                boolean ok = Sha1Verifier.verifyFile(task.destination, task.sha1);
                if (!ok) {
                    CoreLogger.get().warn(LOG,
                            "[" + sessionId + "] CORRUPT [" + task.category + "] " + task.name
                            + " — SHA-1 mismatch, deleting.");
                    silentDelete(task);
                    bad.add(task);
                }
            } catch (IOException ex) {
                CoreLogger.get().warn(LOG,
                        "[" + sessionId + "] SHA-1 check failed for [" + task.category + "] "
                        + task.name + ": " + ex.getMessage() + " — marking for retry.");
                silentDelete(task);
                bad.add(task);
            }
        }

        return new VerifyResult(bad, tasks.size(), sha1Skip);
    }

    public static VerifyResult verifyCategories(
            String sessionId, List<DownloadTask> tasks, String... categories) {

        List<String> cats = List.of(categories);
        List<DownloadTask> filtered = tasks.stream()
                .filter(t -> cats.contains(t.category))
                .toList();
        return verify(sessionId, filtered);
    }

    private static void silentDelete(DownloadTask task) {
        try {
            Files.deleteIfExists(task.destination);
        } catch (IOException ex) {
            CoreLogger.get().warn(LOG, "Failed to delete bad file " + task.destination + ": " + ex.getMessage());
        }
    }
}
