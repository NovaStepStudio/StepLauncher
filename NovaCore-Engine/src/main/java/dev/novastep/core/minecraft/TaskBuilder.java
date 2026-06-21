package dev.novastep.core.minecraft;

import dev.novastep.core.downloader.model.DownloadTask;
import dev.novastep.core.json.Json;
import dev.novastep.core.minecraft.version.AssetIndexManifest;
import dev.novastep.core.minecraft.version.VersionInfo;
import dev.novastep.core.server.request.InstallRequest;
import dev.novastep.core.websocket.EventBroadcaster;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;

public class TaskBuilder {

    private final boolean includeClient;
    private final boolean includeLibraries;
    private final boolean includeAssets;
    private final boolean includeNatives;
    private final EventBroadcaster broadcaster;
    private final Path librariesBase;
    private final Path assetsBase;
    private final Path instanceBase;

    public TaskBuilder(boolean includeClient, boolean includeLibraries,
                       boolean includeAssets, boolean includeNatives) {
        this(includeClient, includeLibraries, includeAssets, includeNatives, null, null, null, null);
    }

    public TaskBuilder(boolean includeClient, boolean includeLibraries,
                       boolean includeAssets, boolean includeNatives,
                       EventBroadcaster broadcaster) {
        this(includeClient, includeLibraries, includeAssets, includeNatives, broadcaster, null, null, null);
    }

    public TaskBuilder(boolean includeClient, boolean includeLibraries,
                       boolean includeAssets, boolean includeNatives,
                       EventBroadcaster broadcaster,
                       Path instanceBase, Path librariesBase, Path assetsBase) {
        this.includeClient = includeClient;
        this.includeLibraries = includeLibraries;
        this.includeAssets = includeAssets;
        this.includeNatives = includeNatives;
        this.broadcaster = broadcaster;
        this.instanceBase = instanceBase;
        this.librariesBase = librariesBase;
        this.assetsBase = assetsBase;
    }

    public static TaskBuilder fromRequest(InstallRequest req, EventBroadcaster broadcaster) {
        Path instance = Path.of(req.resolvedInstancePath()).toAbsolutePath();
        Path libraries = req.resolvedLibrariesPath().toAbsolutePath();
        Path assets = req.resolvedAssetsPath().toAbsolutePath();
        return new TaskBuilder(
                req.shouldDownloadClient(),
                req.shouldDownloadLibraries(),
                req.shouldDownloadAssets(),
                req.shouldDownloadNatives(),
                broadcaster, instance, libraries, assets
        );
    }

    public List<DownloadTask> build(String sessionId, VersionInfo versionInfo, AssetIndexManifest assetIndex) {
        Path instancePath = instanceBase;
        Path librariesPath = librariesBase != null ? librariesBase : instancePath.resolve("libraries");
        Path assetsPath = assetsBase != null ? assetsBase : instancePath.resolve("assets");
        return buildInternal(sessionId, versionInfo, assetIndex, instancePath, librariesPath, assetsPath);
    }

    public List<DownloadTask> build(String sessionId, VersionInfo versionInfo,
                                    AssetIndexManifest assetIndex, Path instancePath) {
        Path librariesPath = librariesBase != null ? librariesBase : instancePath.resolve("libraries");
        Path assetsPath = assetsBase != null ? assetsBase : instancePath.resolve("assets");
        return buildInternal(sessionId, versionInfo, assetIndex, instancePath, librariesPath, assetsPath);
    }

    public static void saveVersionJson(VersionInfo info, Path instancePath) throws IOException {
        Path jsonPath = instancePath.resolve("versions").resolve(info.id).resolve(info.id + ".json");
        Files.createDirectories(jsonPath.getParent());
        Files.writeString(jsonPath, Json.writePretty(info));
    }

    private List<DownloadTask> buildInternal(String sessionId, VersionInfo versionInfo,
                                             AssetIndexManifest assetIndex,
                                             Path instancePath, Path librariesPath, Path assetsPath) {
        List<DownloadTask> tasks = new ArrayList<>();
        if (includeClient) {
            addClientTask(tasks, sessionId, versionInfo, instancePath);
        }
        if (includeLibraries) {
            addLibraryTasks(tasks, sessionId, versionInfo, librariesPath);
        }
        if (includeNatives) {
            addNativeTasks(tasks, sessionId, versionInfo, librariesPath);
        }
        if (includeAssets && versionInfo.assetIndex != null) {
            addAssetIndexTask(tasks, sessionId, versionInfo, assetsPath);
            if (assetIndex != null) {
                addAssetTasks(tasks, sessionId, assetIndex, assetsPath);
            }
        }
        return tasks;
    }

    private void addClientTask(List<DownloadTask> tasks, String sessionId, VersionInfo info, Path instancePath) {
        if (info.downloads == null || info.downloads.client == null || !isValidArtifact(info.downloads.client)) {
            return;
        }
        VersionInfo.Artifact client = info.downloads.client;
        Path destination = instancePath.resolve("versions").resolve(info.id).resolve(info.id + ".jar");
        tasks.add(DownloadTask.client(sessionId, info.id, client.url, destination, client.size, client.sha1));
    }

    private void addLibraryTasks(List<DownloadTask> tasks, String sessionId, VersionInfo info, Path librariesPath) {
        for (LibraryResolver.ResolvedLibrary library : LibraryResolver.resolveAll(info, librariesPath)) {
            // Library already on disk — no download needed (covers Forge legacy ASM, etc.)
            if (Files.exists(library.localPath()))
                continue;

            String downloadUrl = library.url();

            // No explicit URL (name-only library, e.g. Forge pre-1.12 ASM):
            // probe Maven repos to find a valid URL.
            if (downloadUrl == null || downloadUrl.isBlank()) {
                LibraryResolver.Resolution resolved = LibraryResolver.resolve(
                        info.libraries.stream()
                                .filter(l -> library.name().equals(l.name))
                                .findFirst()
                                .orElse(null));
                if (!resolved.found()) {
                    if (broadcaster != null)
                        broadcaster.emitDebug(sessionId, "Cannot resolve URL for library: " + library.name());
                    continue;
                }
                downloadUrl = ((LibraryResolver.Resolution.Found) resolved).url();
            }

            tasks.add(DownloadTask.library(
                    sessionId,
                    shortName(library.name()),
                    downloadUrl,
                    library.localPath(),
                    library.size(),
                    library.sha1()
            ));
        }
    }


    private void addNativeTasks(List<DownloadTask> tasks, String sessionId, VersionInfo info, Path librariesPath) {
        if (info.libraries == null) {
            return;
        }

        String os = currentOs();
        String arch = currentArch();
        int found = 0;
        int skipped = 0;

        for (VersionInfo.Library library : info.libraries) {
            if (!library.isAllowed() || !isNativeLib(library)) {
                continue;
            }

            NativeHandler.NativeArtifactResult result = NativeHandler.resolveNativeArtifact(library, os, arch);
            if (result == null || !isValidArtifact(result.artifact())) {
                skipped++;
                continue;
            }

            String relativePath = result.artifact().path != null
                    ? result.artifact().path
                    : mavenToPath(library.name, result.classifierKey());

            tasks.add(DownloadTask.nativeLib(
                    sessionId,
                    shortName(library.name) + "[" + os + "]",
                    result.artifact().url,
                    librariesPath.resolve(relativePath),
                    result.artifact().size,
                    result.artifact().sha1
            ));
            found++;
        }

        if (broadcaster != null) {
            broadcaster.emitDebug(sessionId,
                    "Natives: " + found + " tasks, " + skipped + " skipped (os=" + os + ", arch=" + arch + ")");
        }
    }

    public static boolean isNativeLib(VersionInfo.Library lib) {
        return NativeHandler.isNativeLibrary(lib);
    }

    public static boolean isValidArtifact(VersionInfo.Artifact artifact) {
        return artifact != null && artifact.url != null && !artifact.url.isBlank() && artifact.size != 0;
    }

    private void addAssetIndexTask(List<DownloadTask> tasks, String sessionId,
                                   VersionInfo info, Path assetsPath) {
        VersionInfo.AssetIndex assetIndex = info.assetIndex;
        tasks.add(DownloadTask.assetIndex(
                sessionId,
                assetIndex.id,
                assetIndex.url,
                assetsPath.resolve("indexes").resolve(assetIndex.id + ".json"),
                assetIndex.size,
                assetIndex.sha1
        ));
    }

    private void addAssetTasks(List<DownloadTask> tasks, String sessionId,
                               AssetIndexManifest assetIndex, Path assetsPath) {
        if (assetIndex.objects == null) {
            return;
        }
        for (var entry : assetIndex.objects.entrySet()) {
            AssetIndexManifest.Asset asset = entry.getValue();
            tasks.add(DownloadTask.asset(
                    sessionId,
                    entry.getKey(),
                    asset.downloadUrl(),
                    assetsPath.resolve("objects").resolve(asset.objectPath()),
                    asset.size,
                    asset.hash
            ));
        }
    }

    public static String currentOs() {
        return RuleEvaluator.currentOsName();
    }

    public static String currentArch() {
        return RuleEvaluator.currentArch();
    }

    public static String mavenToPath(String coord) {
        return mavenToPath(coord, null);
    }

    public static String mavenToPath(String coord, String classifier) {
        if (coord == null || coord.isBlank()) {
            return null;
        }
        if (classifier == null || classifier.isBlank()) {
            return MavenPathResolver.toPath(coord);
        }
        var parsed = MavenPathResolver.parse(coord);
        return MavenPathResolver.artifactPath(parsed.group, parsed.artifact, parsed.version, classifier, parsed.extension);
    }

    private static String shortName(String coordinate) {
        if (coordinate == null || coordinate.isBlank()) {
            return "library";
        }
        String[] parts = coordinate.split(":");
        return parts.length >= 2 ? parts[1] : coordinate;
    }
}
