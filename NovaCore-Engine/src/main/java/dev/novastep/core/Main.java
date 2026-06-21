package dev.novastep.core;

import dev.novastep.core.downloader.DownloadManager;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.server.CoreHttpServer;
import dev.novastep.core.websocket.EventBroadcaster;

import java.nio.file.Files;
import java.nio.file.Path;
import java.security.SecureRandom;
import java.util.concurrent.CountDownLatch;

public class Main {

    public static void main(String[] args) throws Exception {
        Thread.setDefaultUncaughtExceptionHandler((thread, ex) -> {
            System.err.println("[Core] UNCAUGHT in thread [" + thread.getName() + "]: " + ex);
            ex.printStackTrace(System.err);
            System.err.flush();
        });

        CountDownLatch shutdownLatch = new CountDownLatch(1);

        int httpPort = 7878;
        int wsPort = 7879;
        int maxThreads = 32;
        String instancesDir = null;
        String logDir = null;
        String launcherName = "novacore-engine";
        String logLevel = "INFO";
        String accessToken = null;
        boolean clearLogs = false;
        int logQueueLimit = 1_500;

        for (int i = 0; i < args.length - 1; i++) {
            switch (args[i]) {
                case "--port" -> httpPort = Integer.parseInt(args[i + 1]);
                case "--ws-port" -> wsPort = Integer.parseInt(args[i + 1]);
                case "--threads" -> maxThreads = Integer.parseInt(args[i + 1]);
                case "--instances-dir" -> instancesDir = args[i + 1];
                case "--log-dir" -> logDir = args[i + 1];
                case "--launcher-name" -> launcherName = args[i + 1];
                case "--log-level" -> logLevel = args[i + 1];
                case "--access-token" -> accessToken = args[i + 1];
                case "--clear-logs" -> clearLogs = "true".equalsIgnoreCase(args[i + 1]);
                case "--log-queue" -> logQueueLimit = Integer.parseInt(args[i + 1]);
            }
        }

        final String finalToken;
        if (accessToken == null || accessToken.isBlank()) {
            finalToken = generateSecureToken();
        } else if (accessToken.length() < 32) {
            System.err.println("[Core] FATAL: --access-token is too short (minimum 32 chars). Aborting.");
            System.err.flush();
            System.exit(1);
            return;
        } else {
            finalToken = accessToken;
        }

        System.out.println("TOKEN:" + finalToken);
        System.out.flush();

        Path instancesPath = instancesDir != null
                ? Path.of(instancesDir).toAbsolutePath()
                : Path.of(System.getProperty("user.dir")).resolve("instances").toAbsolutePath();

        Path logDirPath = logDir != null
                ? Path.of(logDir).toAbsolutePath()
                : instancesPath.getParent().resolve("logs").toAbsolutePath();

        CoreLogger.Level level;
        try {
            level = CoreLogger.Level.valueOf(logLevel.toUpperCase());
        } catch (IllegalArgumentException e) {
            level = CoreLogger.Level.INFO;
        }

        Files.createDirectories(logDirPath);
        CoreLogger.init(launcherName, logDirPath, level, clearLogs, logQueueLimit);
        CoreLogger log = CoreLogger.get();

        String engineVersion = CoreVersion.get();
        log.info("Core", "Initializing novacore-engine (version=" + engineVersion + ")");
        log.info("Core", "Resolved instances directory (path=" + instancesPath + ")");

        EventBroadcaster broadcaster = new EventBroadcaster(wsPort, finalToken);
        broadcaster.start();

        DownloadManager downloadManager = new DownloadManager(maxThreads, broadcaster);
        broadcaster.setDownloadManager(downloadManager);

        Runnable shutdownCallback = () -> {
            log.info("Core", "Shutting down via /close...");
            try { downloadManager.shutdown(); } catch (Exception ex) { log.warn("Core", "Failed to shutdown download manager: " + ex.getMessage()); }
            try { broadcaster.stop(1000); } catch (Exception ex) { log.warn("Core", "Failed to stop broadcaster: " + ex.getMessage()); }
            log.info("Core", "Stopped.");
            log.close();
            shutdownLatch.countDown();
        };

        CoreHttpServer httpServer = new CoreHttpServer(
                httpPort, downloadManager, broadcaster,
                instancesPath.toString(), finalToken, shutdownCallback);
        httpServer.start();

        log.info("Core", "HTTP server listening (url=http://localhost:" + httpPort + ")");
        log.info("Core", "WS broadcaster listening (url=ws://localhost:" + wsPort + ")");
        log.info("Core", "Worker threads configured (limit=" + maxThreads + ")");
        log.info("Core", "Engine version loaded (version=" + engineVersion + ")");
        log.info("Core", "NovaCore-Engine initialization completed successfully");

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            log.info("Core", "JVM shutdown hook triggered");
            try { httpServer.getLauncher().killAll(); } catch (Exception ex) { log.warn("Core", "Failed to kill instances: " + ex.getMessage()); }
            try { httpServer.stop(); } catch (Exception ex) { log.error("Core", "Error stopping HTTP server", ex); }
            try { downloadManager.shutdown(); } catch (Exception ex) { log.error("Core", "Error stopping DownloadManager", ex); }
            try { broadcaster.stop(2000); } catch (Exception ex) { log.warn("Core", "Failed to stop broadcaster: " + ex.getMessage()); }
            log.info("Core", "Stopped.");
            log.close();
        }, "shutdown-hook"));

        shutdownLatch.await();
    }

    private static String generateSecureToken() {
        SecureRandom random = new SecureRandom();
        byte[] bytes = new byte[32];
        random.nextBytes(bytes);
        StringBuilder sb = new StringBuilder(64);
        for (byte b : bytes)
            sb.append(String.format("%02x", b));
        return sb.toString();
    }
}
