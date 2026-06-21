package dev.novastep.core.server.handlers;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import dev.novastep.core.downloader.DownloadManager;
import dev.novastep.core.json.JsonParserLite;
import dev.novastep.core.minecraft.RuntimeDownloader;
import dev.novastep.core.minecraft.manifest.ManifestClient;
import dev.novastep.core.minecraft.version.VersionInfo;
import dev.novastep.core.server.HttpUtils;
import dev.novastep.core.state.EngineStateManager;
import dev.novastep.core.websocket.EventBroadcaster;

import java.io.IOException;
import java.nio.file.Path;
import java.util.LinkedHashMap;
import java.util.Map;

public class RuntimeHandler implements HttpHandler {

    private final DownloadManager downloadManager;
    private final EventBroadcaster broadcaster;
    private final ManifestClient manifestClient;
    private final EngineStateManager engineStateManager;

    public RuntimeHandler(DownloadManager downloadManager,
                          EventBroadcaster broadcaster,
                          EngineStateManager engineStateManager) {
        this.downloadManager = downloadManager;
        this.broadcaster = broadcaster;
        this.engineStateManager = engineStateManager;
        this.manifestClient = new ManifestClient();
    }

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        if (HttpUtils.handleCors(exchange)) {
            return;
        }
        if (!HttpUtils.requireMethod(exchange, "POST")) {
            return;
        }
        if (engineStateManager.isSemiOff()) {
            HttpUtils.sendJson(exchange, 409, Map.of(
                    "error", "Engine is in semi-off mode",
                    "engine", engineStateManager.snapshot()
            ));
            return;
        }

        JsonParserLite.JsonValue requestValue;
        try {
            requestValue = JsonParserLite.parse(HttpUtils.readBody(exchange));
        } catch (Exception ex) {
            HttpUtils.badRequest(exchange, "Invalid JSON: " + ex.getMessage());
            return;
        }

        if (!requestValue.isObject()) {
            HttpUtils.badRequest(exchange, "Request must be a JSON object");
            return;
        }
        
        JsonParserLite.JsonObject request = requestValue.asObject();
        if (request.get("version").isNull() || request.get("instancePath").isNull()) {
            HttpUtils.badRequest(exchange, "'version' and 'instancePath' are required");
            return;
        }

        String version = request.getString("version").trim();
        String instancePath = request.getString("instancePath").trim();
        String sharedPath = !request.get("sharedPath").isNull() ? request.getString("sharedPath").trim() : null;
        boolean shared = sharedPath != null && !sharedPath.isBlank();
        String runtimeDir = shared
                ? Path.of(sharedPath).toAbsolutePath().resolve("java").toString()
                : Path.of(instancePath).toAbsolutePath().resolve("runtime").toString();

        Map<String, Object> response = new LinkedHashMap<>();
        response.put("status", "downloading");
        response.put("version", version);
        response.put("instancePath", instancePath);
        response.put("runtimeDir", runtimeDir + "/");
        response.put("shared", shared);
        response.put("message", "Runtime download started. Connect to WS for progress events.");
        HttpUtils.accepted(exchange, response);

        Path finalInstancePath = Path.of(instancePath).toAbsolutePath();
        Path finalSharedPath = shared ? Path.of(sharedPath).toAbsolutePath() : null;

        Thread.ofVirtual().name("runtime-dl-" + version).start(() -> {
            try {
                VersionInfo versionInfo = manifestClient.fetchVersionById(version);
                if (versionInfo.javaVersion == null) {
                    broadcaster.emit("runtime_error", Map.of("version", version, "error", "javaVersion missing"));
                    return;
                }

                RuntimeDownloader runtimeDownloader = new RuntimeDownloader(downloadManager, broadcaster);
                String javaPath = runtimeDownloader.downloadRuntime(
                        "runtime-" + version,
                        versionInfo.javaVersion.component,
                        finalInstancePath,
                        finalSharedPath
                );

                broadcaster.emit("runtime_ready", Map.of(
                        "version", version,
                        "component", versionInfo.javaVersion.component,
                        "javaPath", javaPath,
                        "shared", shared
                ));
            } catch (Exception ex) {
                broadcaster.emit("runtime_error", Map.of(
                        "version", version,
                        "error", ex.getMessage() != null ? ex.getMessage() : ex.getClass().getSimpleName()
                ));
            }
        });
    }
}
