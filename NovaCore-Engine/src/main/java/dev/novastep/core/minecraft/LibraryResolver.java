package dev.novastep.core.minecraft;

import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.minecraft.version.VersionInfo;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.file.Path;
import java.time.Duration;
import java.util.ArrayList;
import java.util.List;

public final class LibraryResolver {

    private static final String LOG = "LibraryResolver";

    private static final String[] MAVEN_REPOS = {
            "https://libraries.minecraft.net/",
            "https://maven.minecraftforge.net/",
            "https://repo1.maven.org/maven2/",
            "https://maven.fabricmc.net/",
            "https://maven.neoforged.net/releases/",
            "https://maven.quiltmc.org/repository/release/",
            "https://jitpack.io/"
    };

    private static final HttpClient HTTP = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(5))
            .followRedirects(HttpClient.Redirect.ALWAYS)
            .build();

    private LibraryResolver() {
    }

    public sealed interface Resolution permits Resolution.Found, Resolution.NotFound {

        record Found(String url, String path, String repo) implements Resolution {
            @Override
            public boolean found() {
                return true;
            }
        }

        record NotFound(String name) implements Resolution {
            @Override
            public boolean found() {
                return false;
            }
        }

        boolean found();
    }

    public static Resolution resolve(VersionInfo.Library library) {
        if (library == null)
            return new Resolution.NotFound("<null>");

        if (hasExplicitDownload(library)) {
            String path = MavenPathResolver.resolveLibraryPath(library);
            if (path == null)
                return new Resolution.NotFound(libName(library));
            return new Resolution.Found(library.downloads.artifact.url, path, "explicit");
        }

        if (library.name == null || library.name.isBlank())
            return new Resolution.NotFound("<unnamed>");

        try {
            String mavenPath = MavenPathResolver.toPath(library.name);
            return probeRepositories(library.name, mavenPath);
        } catch (IllegalArgumentException ex) {
            CoreLogger.get().warn(LOG, "Invalid Maven coordinate: '" + library.name + "'");
            return new Resolution.NotFound(library.name);
        }
    }

    public static List<ResolvedLibrary> resolveAll(VersionInfo info, Path librariesPath) {
        List<ResolvedLibrary> result = new ArrayList<>();
        if (info == null || info.libraries == null)
            return result;

        for (VersionInfo.Library library : info.libraries) {
            if (!library.isAllowed())
                continue;
            if (NativeHandler.isNativeLibrary(library))
                continue;

            String resolvedPath = derivePath(library);
            if (resolvedPath == null) {
                CoreLogger.get().warn(LOG, "Cannot derive local path for library: " + libName(library));
                continue;
            }

            Path localPath = librariesPath.resolve(resolvedPath);

            String url = null;
            String sha1 = null;
            long size = -1L;

            if (hasExplicitDownload(library)) {
                url = library.downloads.artifact.url;
                sha1 = library.downloads.artifact.sha1;
                size = library.downloads.artifact.size;
            }

            result.add(new ResolvedLibrary(libName(library), url, resolvedPath, localPath, sha1, size));
        }

        return result;
    }

    private static String derivePath(VersionInfo.Library library) {
        if (library.downloads != null
                && library.downloads.artifact != null
                && library.downloads.artifact.path != null
                && !library.downloads.artifact.path.isBlank()) {
            return library.downloads.artifact.path;
        }
        if (library.name != null && !library.name.isBlank()) {
            try {
                return MavenPathResolver.toPath(library.name);
            } catch (IllegalArgumentException ignored) {
            }
        }
        return null;
    }

    private static boolean hasExplicitDownload(VersionInfo.Library library) {
        return library.downloads != null
                && library.downloads.artifact != null
                && library.downloads.artifact.url != null
                && !library.downloads.artifact.url.isBlank();
    }

    private static String libName(VersionInfo.Library library) {
        return library.name != null ? library.name : "<unnamed>";
    }

    private static Resolution probeRepositories(String name, String mavenPath) {
        for (String repoBase : MAVEN_REPOS) {
            String url = repoBase + mavenPath;
            try {
                HttpRequest request = HttpRequest.newBuilder()
                        .uri(URI.create(url))
                        .method("HEAD", HttpRequest.BodyPublishers.noBody())
                        .timeout(Duration.ofSeconds(3))
                        .build();
                HttpResponse<Void> response = HTTP.send(request, HttpResponse.BodyHandlers.discarding());
                if (response.statusCode() >= 200 && response.statusCode() < 300)
                    return new Resolution.Found(url, mavenPath, repoBase);
            } catch (Exception ex) {
                CoreLogger.get().debug(LOG, "HEAD failed for " + url + ": " + ex.getMessage());
            }
        }
        return new Resolution.NotFound(name);
    }


    public record ResolvedLibrary(
            String name,
            String url,
            String mavenPath,
            Path localPath,
            String sha1,
            long size) {
    }
}
