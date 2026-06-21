package dev.novastep.core.server.handlers;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import dev.novastep.core.minecraft.manifest.ManifestClient;
import dev.novastep.core.minecraft.version.VersionManifest;
import dev.novastep.core.server.HttpUtils;

import java.io.IOException;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

public class VersionsHandler implements HttpHandler {

    private final ManifestClient manifestClient;

    public VersionsHandler(ManifestClient manifestClient) {
        this.manifestClient = manifestClient;
    }

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        if (HttpUtils.handleCors(exchange)) return;
        if (!HttpUtils.requireMethod(exchange, "GET")) return;

        String typeFilter = HttpUtils.queryParam(exchange, "type"); 

        VersionManifest manifest;
        try {
            manifest = manifestClient.fetchManifest();
        } catch (Exception e) {
            dev.novastep.core.log.CoreLogger.get().error("VersionsHandler", "Critical failure while fetching version manifest from Mojang/Remote API", e);
            HttpUtils.serverError(exchange, "Failed to fetch version manifest: " + e.getMessage());
            return;
        }

        List<Map<String, Object>> versions = new ArrayList<>();
        for (VersionManifest.VersionEntry entry : manifest.versions) {
            if (typeFilter != null && !typeFilter.isBlank() && !entry.type.equals(typeFilter)) {
                continue;
            }
            Map<String, Object> v = new LinkedHashMap<>();
            v.put("id", entry.id);
            v.put("type", entry.type);
            v.put("releaseTime", entry.releaseTime);
            v.put("url", entry.url);
            versions.add(v);
        }

        Map<String, Object> response = new LinkedHashMap<>();
        response.put("latest", Map.of(
            "release", manifest.latest.release,
            "snapshot", manifest.latest.snapshot
        ));
        response.put("count", versions.size());
        if (typeFilter != null) response.put("filter", typeFilter);
        response.put("versions", versions);

        HttpUtils.ok(exchange, response);
    }
}