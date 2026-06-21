package dev.novastep.core.minecraft;

import dev.novastep.core.minecraft.version.VersionInfo;
import dev.novastep.core.modloader.installer.MavenCoordinate;

import java.nio.file.Path;
import java.util.Objects;

public final class MavenPathResolver {

    private MavenPathResolver() {
    }

    public static MavenCoordinate parse(String coordinate) {
        return MavenCoordinate.parse(coordinate);
    }

    public static String toPath(String coordinate) {
        return parse(coordinate).toPath();
    }

    public static Path toLocalPath(Path librariesPath, String coordinate) {
        return parse(coordinate).toLocalPath(librariesPath);
    }

    public static String toRemoteUrl(String repositoryBase, String coordinate) {
        return parse(coordinate).toRemoteUrl(repositoryBase);
    }

    public static String normalizeRepositoryBase(String repositoryBase) {
        if (repositoryBase == null || repositoryBase.isBlank()) {
            return "";
        }
        return repositoryBase.endsWith("/") ? repositoryBase : repositoryBase + "/";
    }

    public static String artifactPath(String group, String artifact, String version, String classifier, String extension) {
        String groupPath = Objects.requireNonNull(group, "group").replace('.', '/');
        String artifactName = Objects.requireNonNull(artifact, "artifact");
        String resolvedVersion = Objects.requireNonNull(version, "version");
        String resolvedExtension = extension == null || extension.isBlank() ? "jar" : extension;
        StringBuilder fileName = new StringBuilder(artifactName)
                .append('-')
                .append(resolvedVersion);
        if (classifier != null && !classifier.isBlank()) {
            fileName.append('-').append(classifier);
        }
        fileName.append('.').append(resolvedExtension);
        return groupPath + '/' + artifactName + '/' + resolvedVersion + '/' + fileName;
    }

    public static String resolveLibraryPath(VersionInfo.Library library) {
        if (library == null) {
            return null;
        }
        if (library.downloads != null && library.downloads.artifact != null && library.downloads.artifact.path != null
                && !library.downloads.artifact.path.isBlank()) {
            return library.downloads.artifact.path;
        }
        if (library.name == null || library.name.isBlank()) {
            return null;
        }
        return toPath(library.name);
    }

    public static String coordinateKey(String coordinate) {
        if (coordinate == null || coordinate.isBlank()) {
            return "";
        }
        MavenCoordinate parsed = parse(coordinate);
        String classifier = parsed.classifier != null ? ":" + parsed.classifier : "";
        String extension = parsed.extension != null ? "@" + parsed.extension : "";
        return parsed.group + ":" + parsed.artifact + classifier + extension;
    }
}
