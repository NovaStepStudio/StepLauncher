package dev.novastep.core.server.handlers;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import dev.novastep.core.minecraft.MinecraftLauncher;
import dev.novastep.core.minecraft.instance.LaunchInstanceConfigResolver;
import dev.novastep.core.server.HttpUtils;
import dev.novastep.core.server.request.LaunchRequest;
import dev.novastep.core.state.EngineStateManager;

import java.io.IOException;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

public class LaunchHandler implements HttpHandler {

    private final MinecraftLauncher launcher;
    private final EngineStateManager engineStateManager;

    public LaunchHandler(MinecraftLauncher launcher, EngineStateManager engineStateManager) {
        this.launcher = launcher;
        this.engineStateManager = engineStateManager;
    }

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        if (HttpUtils.handleCors(exchange)) {
            return;
        }

        String path = exchange.getRequestURI().getPath();
        String method = exchange.getRequestMethod();

        if ("POST".equalsIgnoreCase(method) && path.equals("/launch")) {
            handleLaunch(exchange);
            return;
        }
        if ("POST".equalsIgnoreCase(method) && path.startsWith("/launch/kill/")) {
            handleKill(exchange, path.substring("/launch/kill/".length()));
            return;
        }
        if ("GET".equalsIgnoreCase(method) && path.startsWith("/launch/status/")) {
            handleStatus(exchange, path.substring("/launch/status/".length()));
            return;
        }
        if ("GET".equalsIgnoreCase(method) && path.equals("/launch/instances")) {
            handleListInstances(exchange);
            return;
        }
        if ("GET".equalsIgnoreCase(method) && path.startsWith("/launch/instances/")) {
            handleGetInstance(exchange, path.substring("/launch/instances/".length()));
            return;
        }

        HttpUtils.notFound(exchange, "Unknown launch endpoint: " + path);
    }

    private void handleLaunch(HttpExchange exchange) throws IOException {
        if (engineStateManager.isSemiOff()) {
            HttpUtils.sendJson(exchange, 409, Map.of(
                    "error", "Engine is in semi-off mode",
                    "engine", engineStateManager.snapshot()
            ));
            return;
        }

        String body = HttpUtils.readBody(exchange);
        if (body == null || body.isBlank()) {
            HttpUtils.badRequest(exchange, "Empty body");
            return;
        }

        LaunchRequest request;
        try {
            request = HttpUtils.readJson(body, LaunchRequest.class);
        } catch (Exception ex) {
            HttpUtils.badRequest(exchange, "Invalid JSON: " + ex.getMessage());
            return;
        }

        if (request == null) {
            HttpUtils.badRequest(exchange, "Null body");
            return;
        }

        String validationError = request.validate();
        if (validationError != null) {
            HttpUtils.badRequest(exchange, validationError);
            return;
        }

        try {
            LaunchInstanceConfigResolver.applyFromInstancePath(request);
        } catch (Exception ignored) {
        }

        String launchId = launcher.launch(request);
        Map<String, Object> response = new LinkedHashMap<>();
        response.put("launchId", launchId);
        response.put("status", "launching");
        response.put("version", request.version);
        response.put("username", request.resolvedUsername());
        response.put("instancePath", request.resolvedInstancePath());
        response.put("runningInstances", launcher.getRunningCount());
        response.put("authlibInjector", request.isAuthlibEnabled()
                ? Map.of("enabled", true, "server", request.authlibInjector.serverUrl)
                : Map.of("enabled", false));
        response.put("message", "Minecraft launching. Connect to WS for realtime logs.");
        response.put("kill", "/launch/kill/" + launchId);
        response.put("statusPath", "/launch/status/" + launchId);
        response.put("details", "/launch/instances/" + launchId);
        HttpUtils.accepted(exchange, response);
    }

    private void handleKill(HttpExchange exchange, String launchId) throws IOException {
        if (!HttpUtils.requireMethod(exchange, "POST")) {
            return;
        }
        if (launcher.kill(launchId)) {
            HttpUtils.ok(exchange, Map.of("launchId", launchId, "status", "killed"));
        } else {
            HttpUtils.notFound(exchange, "Process not found or already finished: " + launchId);
        }
    }

    private void handleStatus(HttpExchange exchange, String launchId) throws IOException {
        if (!HttpUtils.requireMethod(exchange, "GET")) {
            return;
        }
        boolean running = launcher.isRunning(launchId);
        Map<String, Object> data = launcher.getInstanceData(launchId);
        Map<String, Object> response = new LinkedHashMap<>();
        response.put("launchId", launchId);
        response.put("running", running);
        response.put("status", running ? "running" : "stopped");
        if (data != null) {
            response.put("details", data);
        }
        HttpUtils.ok(exchange, response);
    }

    private void handleListInstances(HttpExchange exchange) throws IOException {
        if (!HttpUtils.requireMethod(exchange, "GET")) {
            return;
        }
        List<Map<String, Object>> instances = launcher.getAllInstances();
        HttpUtils.ok(exchange, Map.of(
                "count", instances.size(),
                "running", launcher.getRunningCount(),
                "instances", instances
        ));
    }

    private void handleGetInstance(HttpExchange exchange, String launchId) throws IOException {
        if (!HttpUtils.requireMethod(exchange, "GET")) {
            return;
        }
        Map<String, Object> data = launcher.getInstanceData(launchId);
        if (data == null) {
            HttpUtils.notFound(exchange, "Instance not found: " + launchId);
            return;
        }
        HttpUtils.ok(exchange, data);
    }
}
