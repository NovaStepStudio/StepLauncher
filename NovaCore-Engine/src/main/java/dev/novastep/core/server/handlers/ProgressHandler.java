package dev.novastep.core.server.handlers;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import dev.novastep.core.downloader.DownloadManager;
import dev.novastep.core.downloader.DownloadSession;
import dev.novastep.core.server.HttpUtils;

import java.io.IOException;
import java.util.*;

public class ProgressHandler implements HttpHandler {

    private final DownloadManager downloadManager;

    public ProgressHandler(DownloadManager downloadManager) {
        this.downloadManager = downloadManager;
    }

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        if (HttpUtils.handleCors(exchange))
            return;
        if (!HttpUtils.requireMethod(exchange, "GET"))
            return;

        String path = exchange.getRequestURI().getPath();

        if (path.equals("/progress/global")) {
            handleGlobal(exchange);
            return;
        }

        handleSessionProgress(exchange);
    }

    private void handleSessionProgress(HttpExchange exchange) throws IOException {
        String sessionId = HttpUtils.queryParam(exchange, "sessionId");

        if (sessionId == null || sessionId.isBlank()) {
            List<Map<String, Object>> snapshots = new ArrayList<>();
            for (DownloadSession session : downloadManager.getAllSessions()) {
                snapshots.add(session.toSnapshot());
            }
            HttpUtils.ok(exchange, Map.of(
                    "count", snapshots.size(),
                    "sessions", snapshots));
            return;
        }

        Optional<DownloadSession> sessionOpt = downloadManager.getSession(sessionId);
        if (sessionOpt.isEmpty()) {
            HttpUtils.notFound(exchange, "Session not found: " + sessionId);
            return;
        }

        HttpUtils.ok(exchange, sessionOpt.get().toSnapshot());
    }

    private void handleGlobal(HttpExchange exchange) throws IOException {
        Collection<DownloadSession> sessions = downloadManager.getAllSessions();

        long totalBytes = 0;
        long downloadedBytes = 0;
        int activeStages = 0;

        List<Map<String, Object>> parts = new ArrayList<>();

        for (DownloadSession s : sessions) {
            Map<String, Object> snap = s.toSnapshot();
            long sTotal = (long) snap.getOrDefault("totalBytes", 0L);
            long sDone = (long) snap.getOrDefault("downloadedBytes", 0L);
            totalBytes += sTotal;
            downloadedBytes += sDone;
            activeStages++;

            double weight = 0.33;
            if (!s.getFilesByCategory("jvm").isEmpty())
                weight = 0.4;
            else if (!s.getFilesByCategory("assets").isEmpty())
                weight = 0.2;

            int pct = sTotal > 0 ? (int) ((sDone * 100L) / sTotal) : 0;

            Map<String, Object> part = new LinkedHashMap<>();
            part.put("sessionId", s.getSessionId());
            part.put("percent", pct);
            part.put("weight", weight);
            parts.add(part);
        }

        int percent = calculateGlobalProgress(parts);

        Map<String, Object> resp = new LinkedHashMap<>();
        resp.put("percent", percent);
        resp.put("activeStages", activeStages);
        resp.put("totalBytes", totalBytes);
        resp.put("downloadedBytes", downloadedBytes);
        resp.put("message", sessions.isEmpty() ? "No active sessions" : "Installation in progress");

        dev.novastep.core.log.CoreLogger.get().debug("Installer",
                "Progreso global recalculado (percent=" + percent + ", activeStages=" + activeStages + ")");

        HttpUtils.ok(exchange, resp);
    }

    public static int calculateGlobalProgress(List<Map<String, Object>> parts) {
        if (parts.isEmpty())
            return 0;

        double totalWeightedPercent = 0;
        double totalWeight = 0;

        for (Map<String, Object> part : parts) {
            double weight = (double) part.get("weight");
            int percent = (int) part.get("percent");

            totalWeightedPercent += (percent * weight);
            totalWeight += weight;
        }

        if (totalWeight == 0)
            return 0;
        return (int) (totalWeightedPercent / totalWeight);
    }
}