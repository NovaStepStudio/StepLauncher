package dev.novastep.core.server.handlers;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import dev.novastep.core.CoreVersion;
import dev.novastep.core.server.HttpUtils;

import java.io.IOException;
import java.util.Map;

public class HealthHandler implements HttpHandler {

    private static final long START_TIME = System.currentTimeMillis();

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        if (HttpUtils.handleCors(exchange)) return;
        if (!HttpUtils.requireMethod(exchange, "GET")) return;

        long uptime = System.currentTimeMillis() - START_TIME;

        HttpUtils.ok(exchange, Map.of(
                "status", "ok",
                "version", CoreVersion.get(),
                "uptime", uptime
        ));
    }
}
