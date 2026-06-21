package dev.novastep.core.server;

import com.sun.net.httpserver.Filter;
import com.sun.net.httpserver.HttpContext;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;
import dev.novastep.core.downloader.DownloadManager;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.minecraft.InstallOrchestrator;
import dev.novastep.core.minecraft.MinecraftLauncher;
import dev.novastep.core.minecraft.SessionManager;
import dev.novastep.core.minecraft.manifest.ManifestClient;
import dev.novastep.core.modloader.ModLoaderOrchestrator;
import dev.novastep.core.server.handlers.ApiHandler;
import dev.novastep.core.server.handlers.CloseHandler;
import dev.novastep.core.server.handlers.DebugHandler;
import dev.novastep.core.server.handlers.EngineStateHandler;
import dev.novastep.core.server.handlers.InstallHandler;
import dev.novastep.core.server.handlers.LaunchHandler;
import dev.novastep.core.server.handlers.ModLoaderHandler;
import dev.novastep.core.server.handlers.ProgressHandler;
import dev.novastep.core.server.handlers.QuickPlayHandler;
import dev.novastep.core.server.handlers.RuntimeHandler;
import dev.novastep.core.server.handlers.SystemResourcesHandler;
import dev.novastep.core.server.handlers.TelemetryHandler;
import dev.novastep.core.server.handlers.VersionsHandler;
import dev.novastep.core.server.handlers.HealthHandler;
import dev.novastep.core.state.EngineStateManager;
import dev.novastep.core.websocket.EventBroadcaster;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.nio.file.Path;
import java.security.MessageDigest;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public class CoreHttpServer {

    private static final String LOG = "CoreHttpServer";

    private final HttpServer httpServer;
    private final ExecutorService executor;
    private final InstallOrchestrator orchestrator;
    private final ManifestClient manifestClient;
    private final MinecraftLauncher launcher;
    private final byte[] tokenBytes;
    private final Path instancesDirPath;
    private final EngineStateManager engineStateManager;

    public CoreHttpServer(int port, DownloadManager downloadManager,
                          EventBroadcaster broadcaster, String instancesDir,
                          String accessToken) throws IOException {
        this(port, downloadManager, broadcaster, instancesDir, accessToken, null);
    }

    public CoreHttpServer(int port, DownloadManager downloadManager,
                          EventBroadcaster broadcaster, String instancesDir,
                          String accessToken, Runnable shutdownCallback) throws IOException {
        this.tokenBytes = accessToken.getBytes(StandardCharsets.UTF_8);
        this.orchestrator = new InstallOrchestrator(downloadManager, broadcaster);
        this.manifestClient = new ManifestClient();
        ModLoaderOrchestrator modLoaderOrchestrator = new ModLoaderOrchestrator(downloadManager, broadcaster);
        this.launcher = new MinecraftLauncher(broadcaster, modLoaderOrchestrator);
        this.instancesDirPath = Path.of(instancesDir).toAbsolutePath();
        this.engineStateManager = new EngineStateManager(downloadManager, broadcaster);

        SessionManager.init(this.instancesDirPath);

        this.httpServer = HttpServer.create(new InetSocketAddress("localhost", port), 0);
        this.executor = Executors.newVirtualThreadPerTaskExecutor();
        this.httpServer.setExecutor(executor);

        Runnable safeShutdown = shutdownCallback != null ? shutdownCallback : () -> System.exit(0);
        registerRoutes(downloadManager, modLoaderOrchestrator, broadcaster, safeShutdown);
    }

    private void registerRoutes(DownloadManager downloadManager,
                                ModLoaderOrchestrator modLoaderOrchestrator,
                                EventBroadcaster broadcaster,
                                Runnable shutdownCallback) {
        secure(httpServer.createContext("/api", new ApiHandler()));
        secure(httpServer.createContext("/versions", new VersionsHandler(manifestClient)));
        secure(httpServer.createContext("/install", new InstallHandler(orchestrator, downloadManager, engineStateManager)));
        secure(httpServer.createContext("/progress", new ProgressHandler(downloadManager)));
        secure(httpServer.createContext("/novacore/worlds", new QuickPlayHandler(instancesDirPath)));
        secure(httpServer.createContext("/crashes/latest", new TelemetryHandler()));
        secure(httpServer.createContext("/sessions", new TelemetryHandler()));
        secure(httpServer.createContext("/system/resources", new SystemResourcesHandler()));
        secure(httpServer.createContext("/runtime", new RuntimeHandler(downloadManager, broadcaster, engineStateManager)));
        secure(httpServer.createContext("/launch", new LaunchHandler(launcher, engineStateManager)));
        secure(httpServer.createContext("/engine/state", new EngineStateHandler(engineStateManager)));
        secure(httpServer.createContext("/close", new CloseHandler(launcher, shutdownCallback)));
        secure(httpServer.createContext("/debug/download/client", new DebugHandler(downloadManager, "client")));
        secure(httpServer.createContext("/debug/download/libraries", new DebugHandler(downloadManager, "libraries")));
        secure(httpServer.createContext("/debug/download/assets", new DebugHandler(downloadManager, "assets")));
        secure(httpServer.createContext("/debug/download/natives", new DebugHandler(downloadManager, "natives")));
        secure(httpServer.createContext("/modloaders", new ModLoaderHandler(modLoaderOrchestrator, engineStateManager)));
        secure(httpServer.createContext("/health", new HealthHandler()));

        secure(httpServer.createContext("/", exchange -> {
            String path = exchange.getRequestURI().getPath();
            if (path.equals("/")) {
                exchange.getResponseHeaders().set("Location", "/api");
                exchange.sendResponseHeaders(302, -1);
                return;
            }
            HttpUtils.notFound(exchange, "Endpoint not found - see GET /api for available endpoints");
        }));

        CoreLogger.get().info(LOG, "Routes registered (all protected by access token)");
    }

    private HttpContext secure(HttpContext context) {
        context.getFilters().add(new AuthFilter(tokenBytes));
        return context;
    }

    private static final class AuthFilter extends Filter {
        private final byte[] tokenBytes;

        private AuthFilter(byte[] tokenBytes) {
            this.tokenBytes = tokenBytes;
        }

        @Override
        public String description() {
            return "NovaCore Access Token Auth";
        }

        @Override
        public void doFilter(HttpExchange exchange, Chain chain) throws IOException {
            if (HttpUtils.handleCors(exchange)) {
                return;
            }
            String header = exchange.getRequestHeaders().getFirst("X-Access-Token");
            if (header == null || header.isBlank()
                    || !MessageDigest.isEqual(tokenBytes, header.getBytes(StandardCharsets.UTF_8))) {
                exchange.sendResponseHeaders(404, -1);
                exchange.close();
                return;
            }
            chain.doFilter(exchange);
        }
    }

    public void start() {
        httpServer.start();
        CoreLogger.get().info(LOG, "HTTP server started on port " + httpServer.getAddress().getPort());
    }

    public void stop(int delaySeconds) {
        httpServer.stop(delaySeconds);
        executor.shutdownNow();
        CoreLogger.get().info(LOG, "HTTP server stopped");
    }

    public void stop() {
        stop(0);
    }

    public int getPort() {
        return httpServer.getAddress().getPort();
    }

    public MinecraftLauncher getLauncher() {
        return launcher;
    }
}
