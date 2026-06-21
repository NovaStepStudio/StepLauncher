package dev.novastep.core.server.handlers;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import dev.novastep.core.server.HttpUtils;
import dev.novastep.core.util.SystemResources;

import java.io.IOException;

public class SystemResourcesHandler implements HttpHandler {

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        if (HttpUtils.handleCors(exchange)) return;
        if (!HttpUtils.requireMethod(exchange, "GET")) return;
        HttpUtils.ok(exchange, SystemResources.snapshot());
    }
}