package dev.novastep.core.modloader.installer;

import java.nio.file.Path;

public final class MavenCoordinate {

    public final String group;
    public final String artifact;
    public final String version;
    public final String classifier;
    public final String extension;

    private MavenCoordinate(String group, String artifact, String version,
                             String classifier, String extension) {
        this.group      = group;
        this.artifact   = artifact;
        this.version    = version;
        this.classifier = classifier;
        this.extension  = extension != null ? extension : "jar";
    }

    public static MavenCoordinate parse(String coordinate) {
        String raw = coordinate != null ? coordinate.trim() : "";
        if (raw.startsWith("[") && raw.endsWith("]")) {
            raw = raw.substring(1, raw.length() - 1).trim();
        }

        String ext = "jar";
        if (raw.contains("@")) {
            int atIdx = raw.lastIndexOf('@');
            ext = raw.substring(atIdx + 1).trim();
            raw = raw.substring(0, atIdx).trim();
        }

        String[] parts = raw.split(":");
        if (parts.length < 3) {
            throw new IllegalArgumentException("Invalid Maven coordinate (need at least group:artifact:version): " + coordinate);
        }

        String classifier = parts.length >= 4 && !parts[3].isBlank() ? parts[3] : null;
        return new MavenCoordinate(parts[0], parts[1], parts[2], classifier, ext);
    }

    public String toPath() {
        String groupPath = group.replace('.', '/');
        StringBuilder name = new StringBuilder(artifact)
                .append('-').append(version);
        if (classifier != null && !classifier.isBlank()) {
            name.append('-').append(classifier);
        }
        name.append('.').append(extension);
        return groupPath + '/' + artifact + '/' + version + '/' + name;
    }

    public Path toLocalPath(Path baseDir) {
        return baseDir.resolve(toPath());
    }

    public String toRemoteUrl(String repoBase) {
        if (repoBase == null || repoBase.isBlank()) return toPath();
        String base = repoBase.endsWith("/") ? repoBase : repoBase + "/";
        return base + toPath();
    }

    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder(group).append(':').append(artifact).append(':').append(version);
        if (classifier != null && !classifier.isBlank()) {
            sb.append(':').append(classifier);
        }
        if (!"jar".equals(extension)) {
            sb.append('@').append(extension);
        }
        return sb.toString();
    }
}
