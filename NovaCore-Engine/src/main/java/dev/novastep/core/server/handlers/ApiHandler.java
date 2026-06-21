package dev.novastep.core.server.handlers;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import dev.novastep.core.server.HttpUtils;

import java.io.IOException;
import java.util.LinkedHashMap;
import java.util.Map;

public class ApiHandler implements HttpHandler {

    private static final Map<String, Object> API_RESPONSE = buildResponse();

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        if (HttpUtils.handleCors(exchange)) return;
        if (!HttpUtils.requireMethod(exchange, "GET")) return;

        HttpUtils.ok(exchange, API_RESPONSE);
    }

    private static Map<String, Object> buildResponse() {
        Map<String, Object> debug = new LinkedHashMap<>();
        debug.put("client",    "/debug/download/client");
        debug.put("libraries", "/debug/download/libraries");
        debug.put("assets",    "/debug/download/assets");
        debug.put("natives",   "/debug/download/natives");

        Map<String, Object> endpoints = new LinkedHashMap<>();
        endpoints.put("install",       "/install");
        endpoints.put("launch",        "/launch");
        endpoints.put("launch_kill",   "/launch/kill/{launchId}");
        endpoints.put("runtime",       "/runtime/download");
        endpoints.put("system",        "/system/resources");
        endpoints.put("launch_status", "/launch/status/{launchId}");
        endpoints.put("progress",      "/progress?sessionId={id}");
        endpoints.put("sessions",      "/progress");
        endpoints.put("versions",      "/versions");
        endpoints.put("debug",         debug);

        Map<String, Object> resp = new LinkedHashMap<>();
        resp.put("name",      "novacore-engine");
        resp.put("vendor",    "NovaStepStudios");
        resp.put("version",   dev.novastep.core.CoreVersion.get());
        resp.put("java",      System.getProperty("java.version"));
        resp.put("os",        System.getProperty("os.name"));
        resp.put("endpoints", endpoints);
        return resp;
    }
}