package dev.novastep.core.minecraft;

import dev.novastep.core.downloader.DownloadManager;
import dev.novastep.core.downloader.model.DownloadTask;
import dev.novastep.core.json.Json;
import dev.novastep.core.json.JacksonCompatibilityAdapter.JsonNode;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.websocket.EventBroadcaster;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;

public class RuntimeDownloader {

    private static final String JAVA_ALL_URL = "https://launchermeta.mojang.com/v1/products/java-runtime/2ec0cc96c44e5a76b9c8b7c39df7210883d12871/all.json";
    private static final int TIMEOUT_SEC = 30;
    private static final String LOG = "RuntimeDownloader";

    private final HttpClient http;
    private final DownloadManager downloadManager;
    private final EventBroadcaster broadcaster;

    public RuntimeDownloader(DownloadManager downloadManager, EventBroadcaster broadcaster) {
        this.downloadManager = downloadManager;
        this.broadcaster = broadcaster;
        this.http = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(TIMEOUT_SEC))
                .followRedirects(HttpClient.Redirect.ALWAYS)
                .build();
    }

    public String downloadRuntime(String sessionId, String component,
                                  Path instancePath, Path sharedPath)
            throws IOException, InterruptedException, ExecutionException {

        String platform = detectPlatform();
        JsonNode allJson = fetchJson(JAVA_ALL_URL);
        JsonNode platformNode = allJson.path(platform);
        if (platformNode.isMissingNode()) {
            throw new IOException("No Java runtime available for platform: " + platform);
        }

        JsonNode runtimes = platformNode.path(component);
        if (!runtimes.isArray() || runtimes.isEmpty()) {
            throw new IOException("Component '" + component + "' not available for " + platform);
        }

        JsonNode runtimeMeta = runtimes.get(runtimes.size() - 1);
        String javaVersionName = runtimeMeta.path("version").path("name").asText();
        String manifestUrl = runtimeMeta.path("manifest").path("url").asText();
        Path javaRoot = resolveRuntimeRoot(javaVersionName, instancePath, sharedPath);
        boolean shared = sharedPath != null && javaRoot.startsWith(sharedPath);

        String execPath = getJavaExecutable(javaRoot);
        if (Files.exists(Path.of(execPath))) {
            return execPath;
        }

        JsonNode manifest = fetchJson(manifestUrl);
        JsonNode files = manifest.path("files");
        if (!files.isObject()) {
            throw new IOException("Invalid runtime manifest: missing files map");
        }

        List<DownloadTask> tasks = new ArrayList<>();
        files.fields().forEachRemaining(entry -> {
            String relativePath = entry.getKey();
            JsonNode fileNode = entry.getValue();
            String type = fileNode.path("type").asText();
            try {
                if ("directory".equals(type)) {
                    Files.createDirectories(javaRoot.resolve(relativePath));
                    return;
                }
            } catch (IOException ex) {
                throw new RuntimeException(ex);
            }
            if (!"file".equals(type)) {
                return;
            }
            JsonNode raw = fileNode.path("downloads").path("raw");
            tasks.add(new DownloadTask(
                    sessionId,
                    "runtime",
                    relativePath,
                    raw.path("url").asText(),
                    javaRoot.resolve(relativePath),
                    raw.path("size").asLong(-1L),
                    raw.path("sha1").asText(null)
            ));
        });

        broadcaster.emit("runtime_download_start", Map.of(
                "session", sessionId,
                "component", component,
                "javaVersion", javaVersionName,
                "totalFiles", tasks.size(),
                "shared", shared
        ));

        String runtimeSessionId = downloadManager.createSession();
        downloadManager.submitAll(runtimeSessionId, tasks).get();
        markExecutables(javaRoot);

        broadcaster.emit("runtime_download_complete", Map.of(
                "session", sessionId,
                "javaVersion", javaVersionName,
                "javaPath", execPath,
                "shared", shared
        ));
        return execPath;
    }

    public String downloadRuntime(String sessionId, String component, Path instancePath)
            throws IOException, InterruptedException, ExecutionException {
        return downloadRuntime(sessionId, component, instancePath, null);
    }

    public static String getJavaExecutable(Path javaRoot) {
        boolean windows = System.getProperty("os.name", "").toLowerCase().contains("win");
        return javaRoot.toAbsolutePath().resolve("bin").resolve(windows ? "java.exe" : "java").toString();
    }

    public static String findExistingRuntime(Path instancePath, Path sharedPath) {
        return findExistingRuntime(instancePath, sharedPath, 0);
    }

    public static String findExistingRuntime(Path instancePath, Path sharedPath, int majorVersion) {
        Path[] roots = sharedPath != null
                ? new Path[]{sharedPath.resolve("runtime"), instancePath.resolve("runtime")}
                : new Path[]{instancePath.resolve("runtime")};

        for (Path root : roots) {
            if (!Files.isDirectory(root)) continue;
            try (var stream = Files.list(root)) {
                var found = stream.filter(Files::isDirectory)
                        .filter(path -> path.getFileName().toString().startsWith("java-"))
                        .filter(path -> majorVersion <= 0 || matchesMajorVersion(path, majorVersion))
                        .map(RuntimeDownloader::getJavaExecutable)
                        .filter(exec -> Files.exists(Path.of(exec)))
                        .findFirst();
                if (found.isPresent()) return found.get();
            } catch (Exception ex) {
                CoreLogger.get().error(LOG, "Failed to scan runtimes in " + root, ex);
            }
        }
        return null;
    }

    private static boolean matchesMajorVersion(Path dir, int majorVersion) {
        String name = dir.getFileName().toString().substring(5);
        int ver = 0;
        int i = 0;
        while (i < name.length() && Character.isDigit(name.charAt(i))) {
            ver = ver * 10 + (name.charAt(i) - '0');
            i++;
        }
        return ver == majorVersion;
    }

    private static Path resolveRuntimeRoot(String javaVersionName, Path instancePath, Path sharedPath) {
        if (sharedPath != null) {
            return sharedPath.toAbsolutePath().resolve("runtime").resolve("java-" + javaVersionName);
        }
        return instancePath.toAbsolutePath().resolve("runtime").resolve("java-" + javaVersionName);
    }

    private static void markExecutables(Path javaRoot) {
        try {
            Path bin = javaRoot.resolve("bin");
            if (Files.isDirectory(bin)) {
                try (var stream = Files.list(bin)) {
                    stream.forEach(file -> file.toFile().setExecutable(true));
                }
            }
        } catch (Exception ex) {
            CoreLogger.get().error(LOG, "Failed to mark runtime files executable in " + javaRoot, ex);
        }
    }

    private static String detectPlatform() {
        String os = System.getProperty("os.name", "").toLowerCase();
        String arch = System.getProperty("os.arch", "").toLowerCase();
        if (os.contains("win")) {
            return arch.equals("amd64") || arch.equals("x86_64") ? "windows-x64" : "windows-x86";
        }
        if (os.contains("mac")) {
            return arch.equals("aarch64") || arch.equals("arm64") ? "mac-os-arm64" : "mac-os";
        }
        if (arch.equals("aarch64") || arch.equals("arm64")) {
            return "linux-arm64";
        }
        if (arch.equals("arm") || arch.equals("armv7l")) {
            return "linux-arm32";
        }
        if (arch.equals("i386") || arch.equals("i686")) {
            return "linux-i386";
        }
        return "linux";
    }

    private JsonNode fetchJson(String url) throws IOException, InterruptedException {
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(url))
                .timeout(Duration.ofSeconds(TIMEOUT_SEC))
                .header("User-Agent", "novacore-engine/" + dev.novastep.core.CoreVersion.get() + " (NovaStepStudios)")
                .GET()
                .build();
        HttpResponse<String> response = http.send(request, HttpResponse.BodyHandlers.ofString());
        if (response.statusCode() != 200) {
            throw new IOException("HTTP " + response.statusCode() + " fetching: " + url);
        }
        return Json.readTree(response.body());
    }
}
