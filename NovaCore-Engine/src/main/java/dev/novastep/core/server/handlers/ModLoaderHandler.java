package dev.novastep.core.server.handlers;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.modloader.ModLoaderOrchestrator;
import dev.novastep.core.modloader.ModLoaderRegistry;
import dev.novastep.core.modloader.model.ModLoaderModels.InstalledLoader;
import dev.novastep.core.modloader.model.ModLoaderModels.LoaderVersion;
import dev.novastep.core.server.HttpUtils;
import dev.novastep.core.server.request.ModLoaderRequest;
import dev.novastep.core.state.EngineStateManager;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Path;
import java.util.List;
import java.util.Map;
import java.util.Optional;

public final class ModLoaderHandler implements HttpHandler {

    private static final String LOG = "ModLoaderHandler";

    private final ModLoaderOrchestrator orchestrator;
    private final EngineStateManager engineStateManager;

    public ModLoaderHandler(ModLoaderOrchestrator orchestrator, EngineStateManager engineStateManager) {
        this.orchestrator = orchestrator;
        this.engineStateManager = engineStateManager;
    }

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        String path = exchange.getRequestURI().getPath();
        String method = exchange.getRequestMethod();

        try {
            if ("GET".equals(method) && path.equals("/modloaders")) {
                HttpUtils.send(exchange, 200, Map.of("loaders", ModLoaderRegistry.get().names()));
                return;
            }
            if ("GET".equals(method) && path.startsWith("/modloaders/versions/")) {
                handleGetVersions(exchange, path);
                return;
            }
            if ("POST".equals(method) && path.equals("/modloaders/install")) {
                handleInstall(exchange);
                return;
            }
            if ("GET".equals(method) && path.startsWith("/modloaders/state/")) {
                handleGetState(exchange, path);
                return;
            }
            if ("DELETE".equals(method) && path.startsWith("/modloaders/state/")) {
                handleDeleteState(exchange, path);
                return;
            }
            HttpUtils.send(exchange, 404, Map.of("error", "Not found: " + path));
        } catch (Exception ex) {
            CoreLogger.get().error(LOG, "Unhandled error", ex);
            HttpUtils.send(exchange, 500, Map.of("error", ex.getMessage()));
        }
    }

    private void handleGetVersions(HttpExchange exchange, String path) throws IOException {
        String[] parts = path.split("/");
        if (parts.length < 5) {
            HttpUtils.send(exchange, 400, Map.of("error", "Expected: /modloaders/versions/{loader}/{mcVersion}"));
            return;
        }
        String loaderName = parts[3];
        String mcVersion = parts[4];

        try {
            List<LoaderVersion> versions = orchestrator.getVersions(loaderName, mcVersion);
            if (versions == null || versions.isEmpty()) {
                HttpUtils.send(exchange, 404, Map.of(
                        "code", "MODLOADER_VERSION_NOT_FOUND",
                        "error", "No version " + mcVersion + " for modloader " + loaderName,
                        "loader", loaderName,
                        "minecraftVersion", mcVersion,
                        "supported", false
                ));
                return;
            }
            HttpUtils.send(exchange, 200, Map.of("versions", versions));
        } catch (IllegalArgumentException ex) {
            HttpUtils.send(exchange, 400, Map.of("error", ex.getMessage()));
        } catch (IOException | InterruptedException ex) {
            HttpUtils.send(exchange, 502, Map.of("error", "Upstream error: " + ex.getMessage()));
        }
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
        ModLoaderRequest request;
        try {
            request = HttpUtils.readJson(body, ModLoaderRequest.class);
        } catch (Exception ex) {
            HttpUtils.send(exchange, 400, Map.of("error", "Invalid JSON: " + ex.getMessage()));
            return;
        }
        if (request == null) {
            HttpUtils.send(exchange, 400, Map.of("error", "Request body is empty"));
            return;
        }

        String validationError = request.validate();
        if (validationError != null) {
            HttpUtils.send(exchange, 400, Map.of("error", validationError));
            return;
        }

        String sessionId = Long.toHexString(System.currentTimeMillis());
        Thread.ofVirtual().name("modloader-install-" + sessionId).start(() -> {
            try {
                orchestrator.install(
                        sessionId,
                        request.loader,
                        request.loaderVersion,
                        request.minecraftVersion,
                        Path.of(request.resolvedInstancePath()),
                        request.resolvedLibrariesPath(),
                        request.resolvedMinecraftJar());
            } catch (Exception ex) {
                CoreLogger.get().error(LOG, "[" + sessionId + "] Install failed", ex);
            }
        });

        HttpUtils.send(exchange, 202, Map.of(
                "sessionId", sessionId,
                "loader", request.loader,
                "mcVersion", request.minecraftVersion,
                "status", "started"
        ));
    }

    private void handleGetState(HttpExchange exchange, String path) throws IOException {
        String instancePath = extractInstancePath(path, "/modloaders/state/");
        Optional<InstalledLoader> state = orchestrator.loadState(Path.of(instancePath));
        if (state.isPresent()) {
            HttpUtils.send(exchange, 200, state.get());
        } else {
            HttpUtils.send(exchange, 404, Map.of("error", "No modloader installed in: " + instancePath));
        }
    }

    private void handleDeleteState(HttpExchange exchange, String path) throws IOException {
        String instancePath = extractInstancePath(path, "/modloaders/state/");
        try {
            orchestrator.removeState(Path.of(instancePath));
            HttpUtils.send(exchange, 200, Map.of("removed", true));
        } catch (IOException ex) {
            HttpUtils.send(exchange, 500, Map.of("error", ex.getMessage()));
        }
    }

    private String extractInstancePath(String path, String prefix) {
        return java.net.URLDecoder.decode(path.substring(prefix.length()), StandardCharsets.UTF_8);
    }
}
