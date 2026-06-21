package dev.novastep.core.server.handlers;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import dev.novastep.core.downloader.DownloadManager;
import dev.novastep.core.minecraft.InstallOrchestrator;
import dev.novastep.core.server.HttpUtils;
import dev.novastep.core.server.request.InstallRequest;
import dev.novastep.core.state.EngineStateManager;

import java.io.IOException;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

public class InstallHandler implements HttpHandler {

    private final InstallOrchestrator orchestrator;
    private final DownloadManager downloadManager;
    private final EngineStateManager engineStateManager;

    public InstallHandler(InstallOrchestrator orchestrator,
                          DownloadManager downloadManager,
                          EngineStateManager engineStateManager) {
        this.orchestrator = orchestrator;
        this.downloadManager = downloadManager;
        this.engineStateManager = engineStateManager;
    }

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        if (HttpUtils.handleCors(exchange)) {
            return;
        }

        String path = exchange.getRequestURI().getPath();
        String method = exchange.getRequestMethod();

        if ("POST".equalsIgnoreCase(method) && path.equals("/install")) {
            handleInstall(exchange);
            return;
        }
        if ("POST".equalsIgnoreCase(method) && path.startsWith("/install/pause/")) {
            handleControl(exchange, "pause", path.substring("/install/pause/".length()));
            return;
        }
        if ("POST".equalsIgnoreCase(method) && path.startsWith("/install/resume/")) {
            handleControl(exchange, "resume", path.substring("/install/resume/".length()));
            return;
        }
        if ("POST".equalsIgnoreCase(method) && path.startsWith("/install/cancel/")) {
            handleControl(exchange, "cancel", path.substring("/install/cancel/".length()));
            return;
        }
        if ("GET".equalsIgnoreCase(method) && path.equals("/install/recovery")) {
            handleRecovery(exchange);
            return;
        }

        HttpUtils.notFound(exchange, "Unknown install endpoint: " + path);
    }

    private void handleInstall(HttpExchange exchange) throws IOException {
        if (engineStateManager.isSemiOff()) {
            HttpUtils.sendJson(exchange, 409, Map.of(
                    "error", "Engine is in semi-off mode",
                    "engine", engineStateManager.snapshot()
            ));
            return;
        }

        String body = HttpUtils.readBody(exchange);
        if (body == null || body.isBlank()) {
            HttpUtils.badRequest(exchange, "Request body is empty");
            return;
        }

        InstallRequest request;
        try {
            request = HttpUtils.readJson(body, InstallRequest.class);
        } catch (Exception ex) {
            HttpUtils.badRequest(exchange, "Invalid JSON: " + ex.getMessage());
            return;
        }

        if (request == null) {
            HttpUtils.badRequest(exchange, "Request body cannot be null");
            return;
        }

        String validationError = request.validate();
        if (validationError != null) {
            HttpUtils.badRequest(exchange, validationError);
            return;
        }

        String sessionId;
        try {
            sessionId = orchestrator.install(request);
        } catch (Exception ex) {
            HttpUtils.serverError(exchange, "Failed to start install: " + ex.getMessage());
            return;
        }

        Map<String, Object> response = new LinkedHashMap<>();
        response.put("sessionId", sessionId);
        response.put("version", request.version);
        response.put("instancePath", request.resolvedInstancePath());
        response.put("status", "started");
        response.put("progress", "/progress?sessionId=" + sessionId);
        response.put("pause", "/install/pause/" + sessionId);
        response.put("resume", "/install/resume/" + sessionId);
        response.put("cancel", "/install/cancel/" + sessionId);
        response.put("websocket", "Connect to WS for realtime events");
        HttpUtils.accepted(exchange, response);
    }

    private void handleControl(HttpExchange exchange, String action, String sessionId) throws IOException {
        if (sessionId == null || sessionId.isBlank()) {
            HttpUtils.badRequest(exchange, "sessionId required");
            return;
        }

        boolean ok = switch (action) {
            case "pause" -> downloadManager.pauseSession(sessionId);
            case "resume" -> downloadManager.resumeSession(sessionId);
            case "cancel" -> downloadManager.cancelSession(sessionId);
            default -> false;
        };

        if (!ok) {
            HttpUtils.notFound(exchange, "Session not found or incompatible state: " + sessionId + " [" + action + "]");
            return;
        }

        HttpUtils.ok(exchange, Map.of("sessionId", sessionId, "action", action, "status", "ok"));
    }

    private void handleRecovery(HttpExchange exchange) throws IOException {
        List<Map<String, Object>> snapshots = downloadManager.getRecoverySnapshots();
        HttpUtils.ok(exchange, Map.of("count", snapshots.size(), "snapshots", snapshots));
    }
}
