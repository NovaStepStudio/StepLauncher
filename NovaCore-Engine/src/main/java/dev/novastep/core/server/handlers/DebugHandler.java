package dev.novastep.core.server.handlers;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import dev.novastep.core.downloader.DownloadManager;
import dev.novastep.core.downloader.DownloadSession;
import dev.novastep.core.server.HttpUtils;

import java.io.IOException;
import java.util.*;

public class DebugHandler implements HttpHandler {

    private static final Map<String, String> CATEGORY_MAP = Map.of(
        "client", "client",
        "libraries", "library",
        "assets", "asset",
        "natives", "native"
    );

    private final DownloadManager downloadManager;
    private final String          urlSegment;  

    public DebugHandler(DownloadManager downloadManager, String urlSegment) {
        this.downloadManager = downloadManager;
        this.urlSegment      = urlSegment;
    }

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        if (HttpUtils.handleCors(exchange)) return;
        if (!HttpUtils.requireMethod(exchange, "GET")) return;

        String internalCategory = CATEGORY_MAP.get(urlSegment);
        if (internalCategory == null) {
            HttpUtils.notFound(exchange, "Unknown debug category: " + urlSegment);
            return;
        }

        String sessionId = HttpUtils.queryParam(exchange, "sessionId");
        Optional<DownloadSession> sessionOpt = resolveSession(sessionId);

        if (sessionOpt.isEmpty()) {
            HttpUtils.notFound(exchange,
                sessionId != null
                    ? "Session not found: " + sessionId
                    : "No active sessions found. Start an install first."
            );
            return;
        }

        DownloadSession session = sessionOpt.get();
        List<Map<String, Object>> files = session.getFilesByCategory(internalCategory);

        long done = files.stream().filter(f -> "done".equals(f.get("status"))).count();
        long skipped = files.stream().filter(f -> "skipped".equals(f.get("status"))).count();
        long failed = files.stream().filter(f -> "failed".equals(f.get("status"))).count();
        long pending = files.stream().filter(f -> "pending".equals(f.get("status"))).count();
        long dloading = files.stream().filter(f -> "downloading".equals(f.get("status"))).count();

        Map<String, Object> response = new LinkedHashMap<>();
        response.put("sessionId",   session.getSessionId());
        response.put("category",    urlSegment);
        response.put("total",       files.size());
        response.put("summary", Map.of(
            "done", done,
            "skipped", skipped,
            "failed", failed,
            "pending", pending,
            "downloading", dloading
        ));
        response.put("files", files);

        HttpUtils.ok(exchange, response);
    }

    private Optional<DownloadSession> resolveSession(String sessionId) {
        if (sessionId != null && !sessionId.isBlank()) {
            return downloadManager.getSession(sessionId);
        }
        
        Collection<DownloadSession> all = downloadManager.getAllSessions();

        return all.stream()
            .filter(DownloadSession::isRunning)
            .findFirst()
            .or(() -> all.stream()
                .max(Comparator.comparingLong(s ->
                    (long) s.toSnapshot().getOrDefault("createdAt", 0L)
                ))
            );
    }
}