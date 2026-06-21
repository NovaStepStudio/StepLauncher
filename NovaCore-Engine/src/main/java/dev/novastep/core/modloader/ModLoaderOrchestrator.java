package dev.novastep.core.modloader;

import dev.novastep.core.downloader.DownloadManager;
import dev.novastep.core.downloader.model.DownloadResult;
import dev.novastep.core.downloader.model.DownloadTask;
import dev.novastep.core.json.Json;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.modloader.model.ModLoaderModels.DownloadPlan;
import dev.novastep.core.modloader.model.ModLoaderModels.InstalledLoader;
import dev.novastep.core.modloader.model.ModLoaderModels.LoaderVersion;
import dev.novastep.core.websocket.EventBroadcaster;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;

public final class ModLoaderOrchestrator {

    private static final String LOG = "ModLoaderOrchestrator";
    private static final String STATE_DIR = ".novacore";
    private static final String STATE_FILE = "loader-state.json";

    private final DownloadManager downloadManager;
    private final EventBroadcaster broadcaster;

    public ModLoaderOrchestrator(DownloadManager downloadManager, EventBroadcaster broadcaster) {
        this.downloadManager = downloadManager;
        this.broadcaster = broadcaster;
    }

    public void install(String sessionId,
                        String loaderName,
                        String loaderVersion,
                        String mcVersion,
                        Path instancePath,
                        Path librariesPath,
                        Path minecraftJar) throws Exception {

        ModLoaderProvider provider = ModLoaderRegistry.get()
                .find(loaderName)
                .orElseThrow(() -> new IllegalArgumentException("Unknown modloader: " + loaderName));

        String displayVersion = loaderVersion == null || loaderVersion.isBlank() ? "latest" : loaderVersion;
        CoreLogger.get().info(LOG, "[" + sessionId + "] Installing " + loaderName + " " + displayVersion + " for MC " + mcVersion);

        Map<String, Object> resolvingEvent = new HashMap<>();
        resolvingEvent.put("sessionId", sessionId);
        resolvingEvent.put("loader", loaderName);
        resolvingEvent.put("loaderVersion", displayVersion);
        resolvingEvent.put("mcVersion", mcVersion);
        broadcaster.emit("modloader_resolving", resolvingEvent);

        String resolvedVersion = resolveVersion(provider, loaderVersion, mcVersion, sessionId);
        DownloadPlan plan = provider.resolveDownload(mcVersion, resolvedVersion, instancePath, librariesPath);
        downloadAll(sessionId, plan, loaderName);

        String versionJsonId = buildVersionJsonId(loaderName, mcVersion, resolvedVersion);
        String installerPath = provider.requiresInstallerRun() && plan.installerDestination() != null
                ? plan.installerDestination().toAbsolutePath().toString()
                : null;

        InstalledLoader state = new InstalledLoader(loaderName, resolvedVersion, mcVersion, versionJsonId, installerPath);
        if (provider.requiresInstallerRun() && plan.installerDestination() != null) {
            provider.runInstaller(sessionId, state, instancePath, librariesPath, minecraftJar, broadcaster);
        }

        persistState(instancePath, state);
        broadcaster.emit("modloader_installed", Map.of(
                "sessionId", sessionId,
                "loader", loaderName,
                "loaderVersion", resolvedVersion,
                "versionJsonId", versionJsonId));
    }

    public List<LoaderVersion> getVersions(String loaderName, String mcVersion)
            throws IOException, InterruptedException {
        ModLoaderProvider provider = ModLoaderRegistry.get()
                .find(loaderName)
                .orElseThrow(() -> new IllegalArgumentException("Unknown modloader: " + loaderName));
        return provider.getVersions(mcVersion);
    }

    public Optional<InstalledLoader> loadState(Path instancePath) {
        Path stateFile = instancePath.resolve(STATE_DIR).resolve(STATE_FILE);
        if (!Files.exists(stateFile)) {
            return Optional.empty();
        }
        try {
            return Optional.of(Json.read(Files.readString(stateFile, StandardCharsets.UTF_8), InstalledLoader.class));
        } catch (IOException ex) {
            CoreLogger.get().error(LOG, "Failed to read modloader state file for instance at " + instancePath, ex);
            return Optional.empty();
        }
    }

    public void removeState(Path instancePath) throws IOException {
        Files.deleteIfExists(instancePath.resolve(STATE_DIR).resolve(STATE_FILE));
    }

    private String resolveVersion(ModLoaderProvider provider, String requested, String mcVersion, String sessionId)
            throws IOException, InterruptedException {
        if (requested == null || requested.isBlank() || "latest".equalsIgnoreCase(requested)) {
            List<LoaderVersion> versions = provider.getVersions(mcVersion);
            LoaderVersion latest = versions.stream()
                    .filter(version -> version.stable)
                    .findFirst()
                    .orElse(versions.isEmpty() ? null : versions.get(0));
            if (latest == null) {
                throw new IOException("No versions found for " + provider.name() + " on MC " + mcVersion);
            }
            CoreLogger.get().info(LOG, "[" + sessionId + "] Resolved latest loader version: " + latest.loaderVersion);
            return latest.loaderVersion;
        }
        return requested;
    }

    private void downloadAll(String sessionId, DownloadPlan plan, String loaderName) throws Exception {
        if (plan.entries().isEmpty()) {
            return;
        }

        List<DownloadTask> tasks = new ArrayList<>();
        for (DownloadPlan.Entry entry : plan.entries()) {
            if (!Files.exists(entry.destination)) {
                tasks.add(new DownloadTask(sessionId, entry.category, entry.name, entry.url, entry.destination, entry.size, entry.sha1));
            }
        }

        if (tasks.isEmpty()) {
            return;
        }

        downloadManager.registerSessionIfAbsent(sessionId);
        broadcaster.emit("modloader_downloading", Map.of("sessionId", sessionId, "loader", loaderName, "files", tasks.size()));
        CompletableFuture<List<DownloadResult>> future = downloadManager.submitAll(sessionId, tasks);
        List<DownloadResult> results = future.get();
        long failed = results.stream().filter(DownloadResult::isFailed).count();
        if (failed > 0) {
            throw new IOException("Failed to download " + failed + " modloader file(s).");
        }
    }

    private String buildVersionJsonId(String loaderName, String mcVersion, String loaderVersion) {
        if ("fabric".equals(loaderName) || "quilt".equals(loaderName) || "legacyfabric".equals(loaderName)) {
            return loaderName + "-" + mcVersion + "-" + loaderVersion;
        }
        if ("forge".equals(loaderName)) {
            return loaderVersion.startsWith(mcVersion + "-") ? "forge-" + loaderVersion : "forge-" + mcVersion + "-" + loaderVersion;
        }
        if ("neoforge".equals(loaderName)) {
            return "neoforge-" + loaderVersion;
        }
        return loaderName + "-" + mcVersion + "-" + loaderVersion;
    }

    private void persistState(Path instancePath, InstalledLoader state) throws IOException {
        Path stateDir = instancePath.resolve(STATE_DIR);
        Path stateFile = stateDir.resolve(STATE_FILE);
        Files.createDirectories(stateDir);
        Files.writeString(stateFile, Json.writePretty(state), StandardCharsets.UTF_8);
    }
}
