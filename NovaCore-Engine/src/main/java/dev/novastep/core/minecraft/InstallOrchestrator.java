package dev.novastep.core.minecraft;

import dev.novastep.core.minecraft.common.ModuleStatus;
import dev.novastep.core.minecraft.instance.InstanceConfigStore;
import dev.novastep.core.minecraft.instance.InstanceTechnicalMetadataStore;
import dev.novastep.core.minecraft.instance.LegacyInstanceMetadataMigrator;


import dev.novastep.core.downloader.DownloadManager;
import dev.novastep.core.downloader.model.DownloadResult;
import dev.novastep.core.downloader.model.DownloadTask;
import dev.novastep.core.json.Json;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.minecraft.manifest.ManifestClient;
import dev.novastep.core.minecraft.manifest.VersionMerger;
import dev.novastep.core.minecraft.version.AssetIndexManifest;
import dev.novastep.core.minecraft.version.VersionInfo;
import dev.novastep.core.modloader.ModLoaderOrchestrator;
import dev.novastep.core.modloader.model.ModLoaderModels.InstalledLoader;
import dev.novastep.core.server.request.InstallRequest;
import dev.novastep.core.websocket.EventBroadcaster;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.*;
import java.util.concurrent.CompletableFuture;
import java.util.stream.Collectors;

public class InstallOrchestrator {

    private static final String LOG = "InstallOrchestrator";
    // JSON deserialization via Jackson (Json.java) — Gson removed
    private static final int MAX_MODULE_RETRIES = 2;

    private final DownloadManager downloadManager;
    private final EventBroadcaster broadcaster;
    private final ManifestClient manifestClient;
    private final RuntimeDownloader runtimeDownloader;
    private final ModLoaderOrchestrator modLoaderOrchestrator;

    public InstallOrchestrator(DownloadManager downloadManager, EventBroadcaster broadcaster) {
        this.downloadManager = downloadManager;
        this.broadcaster = broadcaster;
        this.manifestClient = new ManifestClient();
        this.runtimeDownloader = new RuntimeDownloader(downloadManager, broadcaster);
        this.modLoaderOrchestrator = new ModLoaderOrchestrator(downloadManager, broadcaster);
    }

    public String install(InstallRequest request) {
        String sessionId = downloadManager.createSession();

        CoreLogger.get().info(LOG, "Install: version=" + request.version
                + ", path=" + request.resolvedInstancePath()
                + (request.modloader != null ? ", modloader=" + request.modloader : ""));

        Thread.ofVirtual().name("install-" + sessionId).start(() -> {
            try {
                runInstall(sessionId, request);
            } catch (Throwable t) {
                String msg = t.getClass().getSimpleName() + ": " + t.getMessage();
                CoreLogger.get().error(LOG, "Installation failed for session " + sessionId + " (version="
                        + request.version + ", path=" + request.resolvedInstancePath() + ")", t);
                downloadManager.getSession(sessionId).ifPresent(s -> s.markFailed(msg));
                broadcaster.emitSessionFailed(sessionId, msg);
            }
        });

        return sessionId;
    }

    private void runInstall(String sessionId, InstallRequest request) throws Exception {
        Path instancePath = Path.of(request.resolvedInstancePath());
        InstanceTechnicalMetadataStore.TechnicalMetadata techMeta = null;
        if (request.isInstance()) {
            techMeta = LegacyInstanceMetadataMigrator
                    .migrateIfPresent(instancePath, request.launcher != null ? request.launcher.name : null)
                    .orElseGet(() -> InstanceTechnicalMetadataStore.readOrCreate(instancePath, request.version));

            InstanceTechnicalMetadataStore.recordVerification(instancePath, techMeta);
            InstanceTechnicalMetadataStore.recordInstall(instancePath, techMeta, request.version);

            InstanceConfigStore.readOrCreate(instancePath, techMeta, request.version);

            CoreLogger.get().info(LOG, "[" + sessionId + "] Instance files ready: "
                    + InstanceTechnicalMetadataStore.FILENAME + ", " + InstanceConfigStore.FILENAME);
        }

        broadcaster.emit("install_step", Map.of(
                "sessionId", sessionId, "step", "resolving_version", "version", request.version));

        VersionInfo versionInfo;
        boolean isOnline;

        try {
            versionInfo = manifestClient.fetchVersionWithInheritance(request.version);
            isOnline = true;
            broadcaster.emitManifestResolved(sessionId, versionInfo.id);
            TaskBuilder.saveVersionJson(versionInfo, instancePath);
        } catch (IOException | InterruptedException networkEx) {
            CoreLogger.get().warn(LOG, "[" + sessionId + "] Network unavailable: "
                    + networkEx.getMessage() + " — checking local cache for " + request.version);

            versionInfo = loadLocalVersionInfo(request);
            isOnline = false;

            if (versionInfo == null) {
                throw new RuntimeException(
                        "No network and no local cache for version '" + request.version
                                + "'. Install at least once with internet access.");
            }

            broadcaster.emit("offline_mode", Map.of(
                    "sessionId", sessionId,
                    "version", request.version,
                    "reason", networkEx.getMessage()));
        }

        CoreLogger.get().info(LOG, "[" + sessionId + "] Version: " + versionInfo.id
                + " mainClass=" + versionInfo.mainClass
                + " [" + (isOnline ? "online" : "OFFLINE") + "]");

        AssetIndexManifest assetIndex = null;
        if (isOnline && request.shouldDownloadAssets() && versionInfo.assetIndex != null) {
            broadcaster.emit("install_step", Map.of(
                    "sessionId", sessionId, "step", "fetching_asset_index",
                    "indexId", versionInfo.assetIndex.id));
            assetIndex = manifestClient.fetchAssetIndex(versionInfo.assetIndex);
        }

        if (isOnline && request.shouldDownloadJvm() && versionInfo.javaVersion != null) {
            broadcaster.emit("install_step", Map.of(
                    "sessionId", sessionId, "step", "downloading_jvm",
                    "component", versionInfo.javaVersion.component,
                    "major", versionInfo.javaVersion.majorVersion));
            try {
                Path sharedPath = request.hasSharedPath()
                        ? Path.of(request.sharedPath).toAbsolutePath()
                        : null;
                String javaPath = runtimeDownloader.downloadRuntime(
                        sessionId, versionInfo.javaVersion.component, instancePath, sharedPath);
                CoreLogger.get().info(LOG, "[" + sessionId + "] JVM ready: " + javaPath);
            } catch (Exception ex) {
                CoreLogger.get()
                        .error(LOG, "[" + sessionId + "] JVM download failed for component="
                                + versionInfo.javaVersion.component + " major=" + versionInfo.javaVersion.majorVersion,
                                ex);
            }
        }

        broadcaster.emit("install_step", Map.of("sessionId", sessionId, "step", "building_task_list"));

        InstallRequest effectiveRequest = isOnline ? request : withNoDownloads(request);
        TaskBuilder builder = TaskBuilder.fromRequest(effectiveRequest, broadcaster);
        List<DownloadTask> tasks = builder.build(sessionId, versionInfo, assetIndex, instancePath);

        long totalBytes = tasks.stream().mapToLong(t -> t.expectedSize).sum();

        broadcaster.emit("tasks_ready", Map.of(
                "sessionId", sessionId,
                "totalTasks", tasks.size(),
                "totalBytes", totalBytes,
                "offline", !isOnline,
                "breakdown", countByCategory(tasks)));

        if (!tasks.isEmpty()) {
            broadcaster.emit("install_step", Map.of(
                    "sessionId", sessionId, "step", "downloading", "files", tasks.size()));

            ModuleTracker tracker = buildTracker(sessionId, effectiveRequest, tasks);

            tracker.modules().forEach(m -> tracker.transition(m, ModuleStatus.DOWNLOADING));

            CompletableFuture<List<DownloadResult>> future = downloadManager.submitAll(sessionId, tasks);
            List<DownloadResult> results = future.get();

            long succeeded = results.stream().filter(r -> r.success && !r.skipped).count();
            long skipped = results.stream().filter(r -> r.skipped).count();
            long failed = results.stream().filter(DownloadResult::isFailed).count();

            CoreLogger.get().info(LOG, "[" + sessionId + "] Download batch: "
                    + succeeded + " downloaded, " + skipped + " skipped, " + failed + " failed.");

            for (String module : tracker.modules()) {
                long moduleFailed = countFailedForCategory(results, moduleToCategory(module));
                tracker.transition(module,
                        moduleFailed > 0 ? ModuleStatus.FAILED : ModuleStatus.COMPLETED);
            }

            boolean allOk = runVerifyAndRetry(sessionId, tasks, tracker);

            if (!allOk) {
                String failMsg = "Install aborted: modules ["
                        + tracker.snapshot().entrySet().stream()
                                .filter(e -> "failed".equals(e.getValue()))
                                .map(Map.Entry::getKey)
                                .collect(Collectors.joining(", "))
                        + "] could not be verified after " + MAX_MODULE_RETRIES + " retries.";

                CoreLogger.get().error(LOG, "[" + sessionId + "] " + failMsg);
                broadcaster.emit("install_failed", Map.of(
                        "sessionId", sessionId,
                        "reason", failMsg,
                        "modules", tracker.snapshot()));
                return;
            }

            if (effectiveRequest.shouldDownloadNatives()) {
                extractNatives(sessionId, versionInfo, request, instancePath, effectiveRequest);
            }

            CoreLogger.get().info(LOG, "[" + sessionId + "] Vanilla install complete — modules: "
                    + tracker.snapshot());
        }

        if (isOnline && request.modloader != null && !request.modloader.isBlank()) {
            broadcaster.emit("install_step", Map.of(
                    "sessionId", sessionId, "step", "modloader", "loader", request.modloader));

            Path minecraftJar = instancePath
                    .resolve("versions").resolve(request.version)
                    .resolve(request.version + ".jar");

            modLoaderOrchestrator.install(
                    sessionId,
                    request.modloader,
                    request.modloaderVersion,
                    request.version,
                    instancePath,
                    effectiveRequest.resolvedLibrariesPath().toAbsolutePath(),
                    minecraftJar);

            try {
                repairModloaderLibraries(sessionId, instancePath, effectiveRequest);
            } catch (Exception ex) {
                CoreLogger.get().error(LOG,
                        "[" + sessionId + "] Modloader library repair failed for instance at " + instancePath, ex);
            }

            if (request.isInstance() && techMeta != null) {
                try {
                    var loaderState = modLoaderOrchestrator.loadState(instancePath);
                    if (loaderState.isPresent() && loaderState.get().versionJsonId != null) {
                        InstanceTechnicalMetadataStore.recordInstall(
                                instancePath, techMeta, loaderState.get().versionJsonId);
                    }
                } catch (Exception ex) {
                    CoreLogger.get().error(LOG,
                            "[" + sessionId + "] Failed to record modloader version in metadata", ex);
                }
            }
        }

        broadcaster.emit("install_completed", Map.of(
                "sessionId", sessionId,
                "version", request.version,
                "modloader", request.modloader != null ? request.modloader : "none"));

        CoreLogger.get().info(LOG, "[" + sessionId + "] Install finished: version=" + request.version);
    }

    private boolean runVerifyAndRetry(String sessionId, List<DownloadTask> tasks,
            ModuleTracker tracker) throws Exception {

        List<DownloadTask> verifyList = new ArrayList<>(tasks);

        for (int round = 0; round <= MAX_MODULE_RETRIES; round++) {

            tracker.modules().forEach(m -> tracker.transition(m, ModuleStatus.VERIFYING));

            broadcaster.emit("install_step", Map.of(
                    "sessionId", sessionId,
                    "step", "verifying",
                    "round", round,
                    "files", verifyList.size()));

            InstallVerifier.VerifyResult result = InstallVerifier.verify(sessionId, verifyList);

            CoreLogger.get().info(LOG, "[" + sessionId + "] Verify round " + round
                    + ": " + result.passed() + "/" + result.checked() + " passed"
                    + (result.sha1Skipped() > 0 ? " (" + result.sha1Skipped() + " no-sha1)" : "")
                    + (result.ok() ? " ✓ ALL OK" : " — " + result.bad().size() + " bad"));

            if (result.ok()) {
                tracker.modules().forEach(m -> tracker.transition(m, ModuleStatus.VERIFIED));
                broadcaster.emit("install_step", Map.of(
                        "sessionId", sessionId, "step", "verified",
                        "modules", tracker.snapshot()));
                return true;
            }

            for (String module : tracker.modules()) {
                String cat = moduleToCategory(module);
                boolean hasBad = result.bad().stream().anyMatch(t -> cat.equals(t.category));
                tracker.transition(module, hasBad ? ModuleStatus.FAILED : ModuleStatus.VERIFIED);
            }

            if (round == MAX_MODULE_RETRIES)
                break;

            List<DownloadTask> badTasks = result.bad();
            Map<String, Long> badMap = result.badByCategory();

            CoreLogger.get().warn(LOG, "[" + sessionId + "] Bad files after verify: "
                    + badMap + " — starting retry round " + (round + 1));

            broadcaster.emit("install_step", Map.of(
                    "sessionId", sessionId,
                    "step", "retrying",
                    "round", round + 1,
                    "maxRetries", MAX_MODULE_RETRIES,
                    "badFiles", badTasks.size(),
                    "byModule", badMap));

            tracker.modules().stream()
                    .filter(m -> tracker.getStatus(m) == ModuleStatus.FAILED)
                    .forEach(m -> tracker.transition(m, ModuleStatus.RETRYING));

            String retrySessionId = sessionId + "-retry-" + (round + 1);
            downloadManager.registerSessionIfAbsent(retrySessionId);

            List<DownloadTask> retryTasks = badTasks.stream()
                    .map(t -> new DownloadTask(retrySessionId, t.category, t.name,
                            t.url, t.destination, t.expectedSize, t.sha1))
                    .toList();

            List<DownloadResult> retryResults = downloadManager.submitAll(retrySessionId, retryTasks).get();

            long retryFailed = retryResults.stream().filter(DownloadResult::isFailed).count();
            CoreLogger.get().info(LOG, "[" + sessionId + "] Retry round " + (round + 1)
                    + ": " + (retryTasks.size() - retryFailed) + " ok, " + retryFailed + " still failed.");
            verifyList = new ArrayList<>(tasks);
        }

        return false;
    }

    private void extractNatives(String sessionId, VersionInfo versionInfo,
            InstallRequest request, Path instancePath,
            InstallRequest effectiveRequest) {
        broadcaster.emit("install_step",
                Map.of("sessionId", sessionId, "step", "extracting_natives"));
        try {
            String vanillaVersionId = resolveVanillaVersionId(request.version, instancePath);

            ClasspathBuilder cpBuilder = new ClasspathBuilder(
                    versionInfo, instancePath,
                    effectiveRequest.resolvedLibrariesPath().toAbsolutePath(),
                    effectiveRequest.resolvedAssetsPath().toAbsolutePath(),
                    vanillaVersionId);

            Path nativesDir = cpBuilder.extractNatives();
            CoreLogger.get().info(LOG, "[" + sessionId + "] Natives extracted: " + nativesDir
                    + " (vanilla base: " + vanillaVersionId + ")");

        } catch (Exception ex) {
            CoreLogger.get().error(LOG, "[" + sessionId + "] Natives extraction failed for version=" + request.version
                    + " in " + instancePath, ex);
        }
    }

    private void repairModloaderLibraries(String sessionId, Path instancePath,
            InstallRequest request) {
        Optional<InstalledLoader> stateOpt = modLoaderOrchestrator.loadState(instancePath);
        if (stateOpt.isEmpty())
            return;

        String versionJsonId = stateOpt.get().versionJsonId;
        if (versionJsonId == null || versionJsonId.isBlank())
            return;

        Path versionFile = instancePath.resolve("versions")
                .resolve(versionJsonId).resolve(versionJsonId + ".json");
        if (!Files.exists(versionFile))
            return;

        VersionInfo loaderInfo;
        try {
            String raw = Files.readString(versionFile, StandardCharsets.UTF_8);
            loaderInfo = Json.read(raw, VersionInfo.class);
            if (loaderInfo.inheritsFrom != null && !loaderInfo.inheritsFrom.isBlank()) {
                Path parentFile = instancePath.resolve("versions")
                        .resolve(loaderInfo.inheritsFrom).resolve(loaderInfo.inheritsFrom + ".json");
                if (Files.exists(parentFile)) {
                    String parentRaw = Files.readString(parentFile, StandardCharsets.UTF_8);
                    VersionInfo parent = Json.read(parentRaw, VersionInfo.class);
                    loaderInfo = VersionMerger.merge(parent, loaderInfo);
                }
            }
        } catch (Exception e) {
            CoreLogger.get().warn(LOG,
                    "[" + sessionId + "] Modloader version JSON read: " + e.getMessage());
            return;
        }

        if (loaderInfo.libraries == null)
            return;

        Path librariesPath = request.resolvedLibrariesPath().toAbsolutePath();
        List<DownloadTask> tasks = new ArrayList<>();

        for (VersionInfo.Library lib : loaderInfo.libraries) {
            if (!lib.isAllowed() || TaskBuilder.isNativeLib(lib))
                continue;
            if (lib.downloads == null || lib.downloads.artifact == null)
                continue;
            VersionInfo.Artifact art = lib.downloads.artifact;
            if (art.path == null || art.url == null || art.url.isBlank())
                continue;
            Path dest = librariesPath.resolve(art.path);
            if (Files.exists(dest))
                continue;
            tasks.add(DownloadTask.library(sessionId, art.path, art.url, dest, art.size, art.sha1));
        }

        if (tasks.isEmpty()) {
            CoreLogger.get().info(LOG, "[" + sessionId + "] Modloader runtime libs: all present.");
            return;
        }

        CoreLogger.get().info(LOG, "[" + sessionId + "] Modloader runtime libs missing: "
                + tasks.size() + " — downloading...");
        try {
            List<DownloadResult> results = downloadManager.submitAll(sessionId, tasks).get();
            long ok = results.stream().filter(r -> r.success && !r.skipped).count();
            long skip = results.stream().filter(r -> r.skipped).count();
            long fail = results.stream().filter(DownloadResult::isFailed).count();
            CoreLogger.get().info(LOG, "[" + sessionId + "] Modloader runtime libs: "
                    + ok + " downloaded, " + skip + " skipped, " + fail + " failed.");
        } catch (Exception e) {
            CoreLogger.get().error(LOG,
                    "[" + sessionId + "] Failed to complete modloader runtime library download/repair", e);
        }
    }

    private ModuleTracker buildTracker(String sessionId, InstallRequest request,
            List<DownloadTask> tasks) {
        ModuleTracker tracker = new ModuleTracker(sessionId, broadcaster);
        Set<String> cats = tasks.stream().map(t -> t.category).collect(Collectors.toSet());

        if (request.shouldDownloadClient() && cats.contains("client"))
            tracker.register("client");
        if (request.shouldDownloadLibraries() && cats.contains("library"))
            tracker.register("libraries");
        if (request.shouldDownloadAssets() && (cats.contains("asset") || cats.contains("asset_index")))
            tracker.register("assets");
        if (request.shouldDownloadNatives() && cats.contains("native"))
            tracker.register("natives");

        return tracker;
    }

    private static String moduleToCategory(String module) {
        return switch (module) {
            case "client" -> "client";
            case "libraries" -> "library";
            case "assets" -> "asset";
            case "natives" -> "native";
            default -> module;
        };
    }

    private static long countFailedForCategory(List<DownloadResult> results, String category) {
        return results.stream()
                .filter(r -> !r.success && category.equals(r.task.category))
                .count();
    }

    private String resolveVanillaVersionId(String versionId, Path instancePath) {
        Path versionFile = instancePath.resolve("versions")
                .resolve(versionId).resolve(versionId + ".json");
        if (!Files.exists(versionFile))
            return versionId;
        try {
            VersionInfo raw = Json.read(
                    Files.readString(versionFile, StandardCharsets.UTF_8), VersionInfo.class);
            if (raw.inheritsFrom != null && !raw.inheritsFrom.isBlank())
                return resolveVanillaVersionId(raw.inheritsFrom, instancePath);
        } catch (Exception e) {
            CoreLogger.get().warn(LOG, "Could not read " + versionFile + ": " + e.getMessage());
        }
        return versionId;
    }

    private VersionInfo loadLocalVersionInfo(InstallRequest request) {
        String versionId = request.version;
        Path[] candidates = request.hasSharedPath()
                ? new Path[] {
                        Path.of(request.resolvedInstancePath())
                                .resolve("versions").resolve(versionId).resolve(versionId + ".json"),
                        Path.of(request.sharedPath).toAbsolutePath()
                                .resolve("versions").resolve(versionId).resolve(versionId + ".json")
                }
                : new Path[] {
                        Path.of(request.resolvedInstancePath())
                                .resolve("versions").resolve(versionId).resolve(versionId + ".json")
                };

        for (Path candidate : candidates) {
            if (Files.exists(candidate)) {
                try {
                    VersionInfo info = Json.read(Files.readString(candidate), VersionInfo.class);
                    CoreLogger.get().info(LOG, "Using local cache: " + candidate);
                    return info;
                } catch (Exception e) {
                    CoreLogger.get().warn(LOG,
                            "Cannot parse local cache " + candidate + ": " + e.getMessage());
                }
            }
        }
        return null;
    }

    private static InstallRequest withNoDownloads(InstallRequest original) {
        InstallRequest r = new InstallRequest();
        r.version = original.version;
        r.instancePath = original.instancePath;
        r.sharedPath = original.sharedPath;
        r.verifySHA1 = original.verifySHA1;
        r.maxThreads = original.maxThreads;
        r.debug = original.debug;
        r.modloader = original.modloader;
        r.modloaderVersion = original.modloaderVersion;

        InstallRequest.DownloadOptions dl = new InstallRequest.DownloadOptions();
        dl.client = dl.libraries = dl.assets = dl.natives = dl.jvm = false;
        r.download = dl;
        return r;
    }

    private static Map<String, Long> countByCategory(List<DownloadTask> tasks) {
        Map<String, Long> counts = new LinkedHashMap<>();
        for (String cat : List.of("client", "libraries", "assets", "natives", "asset_index"))
            counts.put(cat, 0L);
        for (DownloadTask t : tasks) {
            String key = switch (t.category) {
                case "library" -> "libraries";
                case "asset_index" -> "asset_index";
                default -> t.category;
            };
            counts.merge(key, 1L, Long::sum);
        }
        return counts;
    }
}
