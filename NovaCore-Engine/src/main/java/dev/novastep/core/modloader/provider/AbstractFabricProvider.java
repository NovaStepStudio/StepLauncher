package dev.novastep.core.modloader.provider;

import dev.novastep.core.json.Json;
import dev.novastep.core.json.JacksonCompatibilityAdapter.JsonNode;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.minecraft.version.VersionInfo;
import dev.novastep.core.modloader.ModLoaderProvider;
import dev.novastep.core.modloader.installer.MavenCoordinate;
import dev.novastep.core.modloader.model.ModLoaderModels.DownloadPlan;
import dev.novastep.core.modloader.model.ModLoaderModels.ExecutionPlan;
import dev.novastep.core.modloader.model.ModLoaderModels.InstalledLoader;
import dev.novastep.core.modloader.model.ModLoaderModels.LoaderVersion;
import dev.novastep.core.websocket.EventBroadcaster;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Duration;
import java.util.ArrayList;
import java.util.List;

abstract class AbstractFabricProvider implements ModLoaderProvider {

    private static final HttpClient HTTP = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(15))
            .followRedirects(HttpClient.Redirect.ALWAYS)
            .build();

    private static final String FABRIC_MAVEN_BASE = "https://maven.fabricmc.net/";

    protected abstract String versionsEndpoint(String mcVersion);
    protected abstract String profileEndpoint(String mcVersion, String loaderVersion);

    // ─── Loader versions ─────────────────────────────────────────────────────

    @Override
    public List<LoaderVersion> getVersions(String mcVersion) throws IOException, InterruptedException {
        JsonNode arr    = Json.readTree(get(versionsEndpoint(mcVersion)));
        List<LoaderVersion> result = new ArrayList<>();
        if (!arr.isArray()) return result;

        for (JsonNode el : arr) {
            JsonNode loader  = el.has("loader") ? el.get("loader") : el;
            if (!loader.has("version")) continue;
            String  version  = loader.get("version").asText();
            boolean stable   = loader.has("stable") ? loader.get("stable").asBoolean(true) : true;
            result.add(new LoaderVersion(version, mcVersion, stable));
        }
        return result;
    }

    // ─── Download plan ────────────────────────────────────────────────────────

    @Override
    public DownloadPlan resolveDownload(
            String mcVersion, String loaderVersion,
            Path instancePath, Path librariesPath) throws IOException, InterruptedException {

        String profileJson = get(profileEndpoint(mcVersion, loaderVersion));
        String versionId   = name() + "-" + mcVersion + "-" + loaderVersion;
        Path   versionDir  = instancePath.resolve("versions").resolve(versionId);
        Path   versionFile = versionDir.resolve(versionId + ".json");

        Files.createDirectories(versionDir);
        Files.writeString(versionFile, profileJson, StandardCharsets.UTF_8);

        JsonNode profile   = Json.readTree(profileJson);
        JsonNode libraries = profile.has("libraries") ? profile.get("libraries") : Json.emptyArrayNode();

        List<DownloadPlan.Entry> entries = new ArrayList<>();
        for (JsonNode lib : libraries) {
            resolveLibraryEntry(lib, librariesPath, entries);
        }

        return DownloadPlan.profileOnly(entries);
    }

    // ─── Execution plan ───────────────────────────────────────────────────────

    @Override
    public boolean requiresInstallerRun() { return false; }

    @Override
    public void runInstaller(String sessionId, InstalledLoader loader, Path instancePath,
                             Path librariesPath, Path minecraftJar, EventBroadcaster broadcaster) {}

    @Override
    public ExecutionPlan buildExecution(
            InstalledLoader loader, VersionInfo vanillaInfo,
            Path instancePath, Path librariesPath) {

        String versionId   = loader.versionJsonId;
        Path   versionFile = instancePath.resolve("versions").resolve(versionId).resolve(versionId + ".json");

        if (!Files.exists(versionFile)) {
            CoreLogger.get().warn(name(), "Version JSON not found: " + versionFile);
            return null;
        }

        try {
            JsonNode profile  = Json.readTree(versionFile);

            String mainClass  = profile.has("mainClass")
                    ? profile.get("mainClass").asText()
                    : vanillaInfo.mainClass;

            List<Path>   classpath = buildClasspath(profile, librariesPath, vanillaInfo, instancePath);
            List<String> jvmArgs  = new ArrayList<>();
            List<String> gameArgs = new ArrayList<>();

            if (profile.has("arguments")) {
                JsonNode args = profile.get("arguments");
                collectStringArgs(args.has("jvm")  ? args.get("jvm")  : Json.emptyArrayNode(), jvmArgs);
                collectStringArgs(args.has("game") ? args.get("game") : Json.emptyArrayNode(), gameArgs);
            }

            return ExecutionPlan.fromVersionJson(mainClass, classpath, jvmArgs, gameArgs);

        } catch (IOException ex) {
            CoreLogger.get().error(name(), "Failed to read version JSON: " + versionFile, ex);
            return null;
        }
    }

    // ─── Internal helpers ─────────────────────────────────────────────────────

    private void resolveLibraryEntry(
            JsonNode lib, Path librariesPath, List<DownloadPlan.Entry> entries) {

        // Case 1: explicit downloads.artifact block
        if (lib.has("downloads")) {
            JsonNode downloads = lib.get("downloads");
            if (!downloads.has("artifact")) return;
            JsonNode artifact = downloads.get("artifact");
            String url  = artifact.has("url")  ? artifact.get("url").asText()  : null;
            String path = artifact.has("path") ? artifact.get("path").asText() : null;
            if (url == null || url.isBlank() || path == null) return;
            long   size = artifact.has("size") ? artifact.get("size").asLong(-1L) : -1L;
            String sha1 = artifact.has("sha1") ? artifact.get("sha1").asText()   : null;
            String name = lib.has("name")      ? lib.get("name").asText()        : path;
            entries.add(DownloadPlan.Entry.library(name, url, librariesPath.resolve(path), size, sha1));
            return;
        }

        // Case 2: name + url (Fabric-style)
        if (lib.has("name") && lib.has("url")) {
            String coord = lib.get("name").asText();
            String base  = lib.get("url").asText();
            if (base.isBlank()) base = FABRIC_MAVEN_BASE;
            try {
                MavenCoordinate mc = MavenCoordinate.parse(coord);
                entries.add(DownloadPlan.Entry.library(coord, mc.toRemoteUrl(base), mc.toLocalPath(librariesPath), -1, null));
            } catch (IllegalArgumentException ex) {
                CoreLogger.get().warn(name(), "Failed to parse Maven coordinate: " + coord + " — " + ex.getMessage());
            }
            return;
        }

        // Case 3: name only — resolve from Fabric Maven
        if (lib.has("name")) {
            String coord = lib.get("name").asText();
            try {
                MavenCoordinate mc = MavenCoordinate.parse(coord);
                entries.add(DownloadPlan.Entry.library(coord, mc.toRemoteUrl(FABRIC_MAVEN_BASE), mc.toLocalPath(librariesPath), -1, null));
            } catch (IllegalArgumentException ex) {
                CoreLogger.get().warn(name(), "Failed to parse Maven coordinate: " + coord + " — " + ex.getMessage());
            }
        }
    }

    private List<Path> buildClasspath(
            JsonNode profile, Path librariesPath, VersionInfo vanillaInfo, Path instancePath) {

        List<Path> entries = new ArrayList<>();

        if (profile.has("libraries")) {
            for (JsonNode lib : profile.get("libraries")) {
                if (lib.has("downloads")) {
                    JsonNode downloads = lib.get("downloads");
                    if (!downloads.has("artifact")) continue;
                    JsonNode artifact = downloads.get("artifact");
                    if (!artifact.has("path")) continue;
                    Path jar = librariesPath.resolve(artifact.get("path").asText());
                    if (Files.exists(jar)) entries.add(jar);

                } else if (lib.has("name")) {
                    try {
                        Path jar = MavenCoordinate.parse(lib.get("name").asText()).toLocalPath(librariesPath);
                        if (Files.exists(jar)) entries.add(jar);
                    } catch (IllegalArgumentException ex) {
                        CoreLogger.get().warn(name(),
                                "Failed to parse library Maven coordinate: "
                                + lib.get("name").asText() + " — " + ex.getMessage());
                    }
                }
            }
        }

        if (vanillaInfo != null && vanillaInfo.id != null) {
            Path vanillaJar = instancePath.resolve("versions")
                    .resolve(vanillaInfo.id).resolve(vanillaInfo.id + ".jar");
            if (Files.exists(vanillaJar)) entries.add(vanillaJar);
        }

        return entries;
    }

    private void collectStringArgs(JsonNode arr, List<String> target) {
        if (!arr.isArray()) return;
        for (JsonNode el : arr) {
            if (el.isTextual()) target.add(el.asText());
        }
    }

    // ─── HTTP helper ──────────────────────────────────────────────────────────

    protected String get(String url) throws IOException, InterruptedException {
        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(url))
                .GET()
                .timeout(Duration.ofSeconds(30))
                .build();
        HttpResponse<String> res = HTTP.send(req, HttpResponse.BodyHandlers.ofString(StandardCharsets.UTF_8));
        if (res.statusCode() != 200)
            throw new IOException("HTTP " + res.statusCode() + " from: " + url);
        return res.body();
    }
}
