package dev.novastep.core.server;

import com.fasterxml.jackson.core.type.TypeReference;
import com.sun.net.httpserver.HttpExchange;
import dev.novastep.core.json.Json;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.util.Map;

public final class HttpUtils {

    private HttpUtils() {
    }

    public static boolean handleCors(HttpExchange exchange) throws IOException {
        exchange.getResponseHeaders().set("Access-Control-Allow-Origin", "*");
        exchange.getResponseHeaders().set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS");
        exchange.getResponseHeaders().set("Access-Control-Allow-Headers", "Content-Type, X-Access-Token, Authorization");
        exchange.getResponseHeaders().set("Access-Control-Max-Age", "86400");

        if ("OPTIONS".equalsIgnoreCase(exchange.getRequestMethod())) {
            exchange.sendResponseHeaders(204, -1);
            exchange.close();
            return true;
        }
        return false;
    }

    public static void sendJson(HttpExchange exchange, int status, Object body) throws IOException {
        byte[] bytes = Json.write(body).getBytes(StandardCharsets.UTF_8);
        exchange.getResponseHeaders().set("Content-Type", "application/json; charset=UTF-8");
        exchange.sendResponseHeaders(status, bytes.length);
        try (OutputStream outputStream = exchange.getResponseBody()) {
            outputStream.write(bytes);
        }
    }

    public static void send(HttpExchange exchange, int statusCode, Object responseBody) {
        try {
            byte[] bytes = Json.write(responseBody).getBytes(StandardCharsets.UTF_8);
            exchange.getResponseHeaders().set("Content-Type", "application/json; charset=UTF-8");
            exchange.sendResponseHeaders(statusCode, bytes.length);
            try (OutputStream outputStream = exchange.getResponseBody()) {
                outputStream.write(bytes);
            }
        } catch (Exception ex) {
            dev.novastep.core.log.CoreLogger.get().error("HttpUtils", "Failed to send HTTP response with status " + statusCode, ex);
        }
    }

    public static <T> T readJson(String body, Class<T> type) throws IOException {
        return Json.read(body, type);
    }

    public static <T> T readJson(String body, TypeReference<T> type) throws IOException {
        return Json.read(body, type);
    }

    public static void ok(HttpExchange exchange, Object body) throws IOException {
        sendJson(exchange, 200, body);
    }

    public static void accepted(HttpExchange exchange, Object body) throws IOException {
        sendJson(exchange, 202, body);
    }

    public static void badRequest(HttpExchange exchange, String error) throws IOException {
        sendJson(exchange, 400, Map.of("error", error, "status", 400));
    }

    public static void notFound(HttpExchange exchange, String error) throws IOException {
        sendJson(exchange, 404, Map.of("error", error, "status", 404));
    }

    public static void serverError(HttpExchange exchange, String error) throws IOException {
        sendJson(exchange, 500, Map.of("error", error, "status", 500));
    }

    public static void unauthorized(HttpExchange exchange) throws IOException {
        sendJson(exchange, 401, Map.of("error", "Unauthorized", "status", 401));
    }

    public static void methodNotAllowed(HttpExchange exchange) throws IOException {
        sendJson(exchange, 405, Map.of("error", "Method not allowed: " + exchange.getRequestMethod(), "status", 405));
    }

    public static String readBody(HttpExchange exchange) throws IOException {
        try (InputStream inputStream = exchange.getRequestBody()) {
            return new String(inputStream.readAllBytes(), StandardCharsets.UTF_8);
        }
    }

    public static String queryParam(HttpExchange exchange, String key) {
        String query = exchange.getRequestURI().getQuery();
        if (query == null || query.isBlank()) {
            return null;
        }
        for (String pair : query.split("&")) {
            int eq = pair.indexOf('=');
            if (eq < 0) {
                continue;
            }
            if (pair.substring(0, eq).equals(key)) {
                return pair.substring(eq + 1);
            }
        }
        return null;
    }

    public static boolean requireMethod(HttpExchange exchange, String method) throws IOException {
        if (!method.equalsIgnoreCase(exchange.getRequestMethod())) {
            methodNotAllowed(exchange);
            return false;
        }
        return true;
    }
}
