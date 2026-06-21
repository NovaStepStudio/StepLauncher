package dev.novastep.core.server.handlers;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import dev.novastep.core.minecraft.CrashReporter;
import dev.novastep.core.minecraft.SessionManager;
import dev.novastep.core.server.HttpUtils;

import java.io.IOException;

public class TelemetryHandler implements HttpHandler {
    @Override
    public void handle(HttpExchange exchange) throws IOException {
        String path = exchange.getRequestURI().getPath();
        String method = exchange.getRequestMethod();

        if (!"GET".equalsIgnoreCase(method)) {
            HttpUtils.methodNotAllowed(exchange);
            return;
        }

        if (path.equals("/crashes/latest")) {
            var crash = CrashReporter.getLatestCrash();
            if (crash == null) {
                HttpUtils.notFound(exchange, "No crashes reported yet");
                return;
            }
            HttpUtils.ok(exchange, crash);
        } else if (path.equals("/sessions")) {
            HttpUtils.ok(exchange, SessionManager.getSessions());
        } else {
            HttpUtils.notFound(exchange, "Endpoint not found");
        }
    }
}
