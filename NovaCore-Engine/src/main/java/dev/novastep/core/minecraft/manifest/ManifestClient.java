package dev.novastep.core.minecraft.manifest;

import dev.novastep.core.json.Json;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.minecraft.version.AssetIndexManifest;
import dev.novastep.core.minecraft.version.VersionInfo;
import dev.novastep.core.minecraft.version.VersionManifest;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Duration;

public class ManifestClient {

    private static final String LOG = "ManifestClient";
    private static final String MANIFEST_URL = "https://launchermeta.mojang.com/mc/game/version_manifest_v2.json";

    private static final HttpClient HTTP = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(15))
            .followRedirects(HttpClient.Redirect.ALWAYS)
            .build();

    public VersionManifest fetchManifest() throws IOException, InterruptedException {
        return Json.read(get(MANIFEST_URL), VersionManifest.class);
    }

    public VersionInfo fetchVersionById(String versionId) throws IOException, InterruptedException {
        VersionManifest manifest = fetchManifest();
        String url = manifest.versions.stream()
                .filter(version -> version.id.equals(versionId))
                .findFirst()
                .map(version -> version.url)
                .orElseThrow(() -> new IOException("Version not found in Mojang manifest: " + versionId));
        return fetchVersionFromUrl(url);
    }

    public VersionInfo fetchVersionWithInheritance(String versionId) throws IOException, InterruptedException {
        return fetchVersionWithInheritance(versionId, null);
    }

    public VersionInfo fetchVersionWithInheritance(String versionId, Path localBasePath)
            throws IOException, InterruptedException {
        VersionInfo version = fetchVersionByIdFlexible(versionId, localBasePath);
        return resolveInheritance(version, localBasePath);
    }

    public VersionInfo resolveInheritance(VersionInfo version) throws IOException, InterruptedException {
        return resolveInheritance(version, null);
    }

    public VersionInfo resolveInheritance(VersionInfo version, Path localBasePath)
            throws IOException, InterruptedException {
        if (version.inheritsFrom == null || version.inheritsFrom.isBlank()) {
            return version;
        }

        CoreLogger.get().info(LOG, "Resolving inheritance: " + version.id + " -> " + version.inheritsFrom);
        VersionInfo parent = fetchVersionByIdFlexible(version.inheritsFrom, localBasePath);
        parent = resolveInheritance(parent, localBasePath);
        return VersionMerger.merge(parent, version);
    }

    public AssetIndexManifest fetchAssetIndex(VersionInfo.AssetIndex assetIndex)
            throws IOException, InterruptedException {
        return Json.read(get(assetIndex.url), AssetIndexManifest.class);
    }

    private VersionInfo fetchVersionByIdFlexible(String versionId, Path localBasePath)
            throws IOException, InterruptedException {
        try {
            VersionManifest manifest = fetchManifest();
            String url = manifest.versions.stream()
                    .filter(version -> version.id.equals(versionId))
                    .findFirst()
                    .map(version -> version.url)
                    .orElse(null);
            if (url != null) {
                return fetchVersionFromUrl(url);
            }
        } catch (IOException networkEx) {
            CoreLogger.get().warn(LOG, "Network error while resolving version " + versionId + ": " + networkEx.getMessage());
        }

        if (localBasePath != null) {
            Path localJson = localBasePath.resolve("versions").resolve(versionId).resolve(versionId + ".json");
            if (Files.exists(localJson)) {
                try {
                    return Json.read(Files.readString(localJson, StandardCharsets.UTF_8), VersionInfo.class);
                } catch (Exception ex) {
                    throw new IOException("Corrupted local version JSON for '" + versionId + "': " + localJson, ex);
                }
            }
        }

        throw new IOException("Version '" + versionId + "' not found in Mojang or local storage.");
    }

    private VersionInfo fetchVersionFromUrl(String url) throws IOException, InterruptedException {
        return Json.read(get(url), VersionInfo.class);
    }

    private String get(String url) throws IOException, InterruptedException {
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(url))
                .GET()
                .timeout(Duration.ofSeconds(30))
                .build();
        HttpResponse<String> response = HTTP.send(request, HttpResponse.BodyHandlers.ofString(StandardCharsets.UTF_8));
        if (response.statusCode() != 200) {
            throw new IOException("HTTP " + response.statusCode() + " fetching: " + url);
        }
        return response.body();
    }
}
