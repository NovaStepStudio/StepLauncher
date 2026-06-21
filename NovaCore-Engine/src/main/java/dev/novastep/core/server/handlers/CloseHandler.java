package dev.novastep.core.server.handlers;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.minecraft.MinecraftLauncher;
import dev.novastep.core.server.HttpUtils;

import java.io.IOException;
import java.util.Map;

public class CloseHandler implements HttpHandler {

    private static final String LOG = "CloseHandler";

    private final MinecraftLauncher launcher;
    private final Runnable          shutdownCallback;

    public CloseHandler(MinecraftLauncher launcher, Runnable shutdownCallback) {
        this.launcher         = launcher;
        this.shutdownCallback = shutdownCallback;
    }

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        if (HttpUtils.handleCors(exchange)) return;
        if (!HttpUtils.requireMethod(exchange, "POST")) return;

        int running = launcher.getRunningCount();
        CoreLogger.get().info(LOG, "Close requested — running instances: " + running);

        if (running > 0) {
            CoreLogger.get().info(LOG, "Killing " + running + " active Minecraft process(es) before shutdown");
            launcher.killAll();
        }

        HttpUtils.ok(exchange, Map.of(
                "status",           "closing",
                "killedInstances",  running,
                "message",          "NovaCore-Engine is shutting down cleanly"
        ));

        Thread.ofVirtual().name("graceful-shutdown").start(() -> {
            try { Thread.sleep(300); } catch (InterruptedException ex) { 
                CoreLogger.get().warn(LOG, "Shutdown delay interrupted: " + ex.getMessage());
                Thread.currentThread().interrupt();
            }
            CoreLogger.get().info(LOG, "Executing shutdown callback");
            shutdownCallback.run();
        });
    }
}
