package dev.novastep.core.server.handlers;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import dev.novastep.core.json.Json;
import dev.novastep.core.json.JacksonCompatibilityAdapter.JsonNode;
import dev.novastep.core.server.HttpUtils;
import dev.novastep.core.state.EngineStateManager;

import java.io.IOException;
import java.util.Map;

public final class EngineStateHandler implements HttpHandler {

    private final EngineStateManager engineStateManager;

    public EngineStateHandler(EngineStateManager engineStateManager) {
        this.engineStateManager = engineStateManager;
    }

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        if (HttpUtils.handleCors(exchange)) {
            return;
        }

        String method = exchange.getRequestMethod();
        if ("GET".equalsIgnoreCase(method)) {
            HttpUtils.ok(exchange, engineStateManager.snapshot());
            return;
        }

        if (!"POST".equalsIgnoreCase(method)) {
            HttpUtils.methodNotAllowed(exchange);
            return;
        }

        JsonNode body = Json.readTree(HttpUtils.readBody(exchange));
        String requestedState = body.hasNonNull("state") ? body.get("state").asText() : null;
        String reason = body.hasNonNull("reason") ? body.get("reason").asText() : null;

        if (requestedState == null || requestedState.isBlank()) {
            HttpUtils.badRequest(exchange, "Field 'state' is required");
            return;
        }

        EngineStateManager.State state = switch (requestedState.toLowerCase()) {
            case "active", "on" -> EngineStateManager.State.ACTIVE;
            case "semi-off", "semi_off", "idle", "sleep" -> EngineStateManager.State.SEMI_OFF;
            default -> null;
        };

        if (state == null) {
            HttpUtils.badRequest(exchange, "Unsupported state: " + requestedState);
            return;
        }

        HttpUtils.ok(exchange, Map.of(
                "updated", true,
                "engine", engineStateManager.setState(state, reason)
        ));
    }
}
