package dev.novastep.core.minecraft;

import dev.novastep.core.minecraft.version.VersionInfo;
import dev.novastep.core.server.request.LaunchRequest;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.LinkedHashSet;
import java.util.List;

public class ClasspathBuilder {

    private final VersionInfo versionInfo;
    private final Path instancePath;
    private final Path librariesPath;
    private final Path assetsPath;
    private final String vanillaVersionId;
    private Path resolvedNativesDir;
    private final List<Path> modloaderEntries = new ArrayList<>();

    public ClasspathBuilder(VersionInfo versionInfo, Path instancePath,
                            Path librariesPath, Path assetsPath, String vanillaVersionId) {
        this.versionInfo = versionInfo;
        this.instancePath = instancePath;
        this.librariesPath = librariesPath;
        this.assetsPath = assetsPath;
        this.vanillaVersionId = vanillaVersionId != null ? vanillaVersionId : versionInfo.id;
    }

    public ClasspathBuilder(VersionInfo versionInfo, Path instancePath) {
        this(versionInfo, instancePath, instancePath.resolve("libraries"), instancePath.resolve("assets"), versionInfo.id);
    }

    public ClasspathBuilder(VersionInfo versionInfo, Path instancePath,
                            Path librariesPath, Path assetsPath) {
        this(versionInfo, instancePath, librariesPath, assetsPath, versionInfo.id);
    }

    public static ClasspathBuilder fromRequest(LaunchRequest req, VersionInfo versionInfo) {
        Path instance = Path.of(req.resolvedInstancePath()).toAbsolutePath();
        Path libraries = req.resolvedLibrariesPath().toAbsolutePath();
        Path assets = req.resolvedAssetsPath().toAbsolutePath();
        return new ClasspathBuilder(versionInfo, instance, libraries, assets, versionInfo.id);
    }

    public static ClasspathBuilder fromRequest(LaunchRequest req, VersionInfo versionInfo, String vanillaVersionId) {
        Path instance = Path.of(req.resolvedInstancePath()).toAbsolutePath();
        Path libraries = req.resolvedLibrariesPath().toAbsolutePath();
        Path assets = req.resolvedAssetsPath().toAbsolutePath();
        return new ClasspathBuilder(versionInfo, instance, libraries, assets, vanillaVersionId);
    }

    public void appendModloaderEntries(List<Path> entries) {
        if (entries == null) {
            return;
        }
        for (Path entry : entries) {
            if (entry != null && !modloaderEntries.contains(entry)) {
                modloaderEntries.add(entry);
            }
        }
    }

    public String getVanillaVersionId() {
        return vanillaVersionId;
    }

    public String getLibrariesPath() {
        return librariesPath.toAbsolutePath().toString();
    }

    public List<Path> buildClasspathEntries() {
        LinkedHashSet<Path> ordered = new LinkedHashSet<>();

        for (Path modloaderEntry : modloaderEntries) {
            if (modloaderEntry != null && Files.exists(modloaderEntry)) {
                ordered.add(modloaderEntry.toAbsolutePath().normalize());
            }
        }

        for (LibraryResolver.ResolvedLibrary library : LibraryResolver.resolveAll(versionInfo, librariesPath)) {
            if (Files.exists(library.localPath())) {
                ordered.add(library.localPath().toAbsolutePath().normalize());
            }
        }

        Path clientJar = instancePath.resolve("versions")
                .resolve(vanillaVersionId)
                .resolve(vanillaVersionId + ".jar");
        if (Files.exists(clientJar)) {
            ordered.add(clientJar.toAbsolutePath().normalize());
        }

        return new ArrayList<>(ordered);
    }

    public String buildClasspathString() {
        StringBuilder classpath = new StringBuilder();
        for (Path entry : buildClasspathEntries()) {
            if (!classpath.isEmpty()) {
                classpath.append(File.pathSeparator);
            }
            classpath.append(entry.toAbsolutePath());
        }
        return classpath.toString();
    }

    public String getAssetsDir() {
        return assetsPath.toAbsolutePath().toString();
    }

    public String getNativesDir() {
        if (resolvedNativesDir == null) {
            Path defaultNatives = instancePath.resolve("versions").resolve(vanillaVersionId).resolve("natives");
            List<String> jvmArgs = extractJvmArguments(versionInfo);
            for (String arg : jvmArgs) {
                if (arg.startsWith("-Djava.library.path=")) {
                    String resolved = arg.substring("-Djava.library.path=".length())
                            .replace("${natives_directory}", defaultNatives.toString());
                    resolvedNativesDir = Path.of(resolved).toAbsolutePath().normalize();
                    break;
                }
            }
            if (resolvedNativesDir == null) {
                resolvedNativesDir = defaultNatives.toAbsolutePath().normalize();
            }
        }
        return resolvedNativesDir.toString();
    }

    public Path extractNatives() throws IOException {
        return NativeHandler.extractNatives(versionInfo, librariesPath, Path.of(getNativesDir()));
    }

    private List<String> extractJvmArguments(VersionInfo info) {
        List<String> result = new ArrayList<>();
        if (info.arguments == null || info.arguments.jvm == null) {
            return result;
        }
        for (Object entry : info.arguments.jvm) {
            if (entry instanceof String value) {
                result.add(value);
            }
        }
        return result;
    }
}
