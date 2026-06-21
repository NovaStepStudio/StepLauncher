package dev.novastep.core.minecraft;

import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.minecraft.version.VersionInfo;

import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Enumeration;
import java.util.List;
import java.util.jar.JarEntry;
import java.util.jar.JarFile;

public final class NativeHandler {

    private static final String LOG = "NativeHandler";

    private NativeHandler() {
    }

    public static boolean isNativeLibrary(VersionInfo.Library library) {
        if (library == null || library.downloads == null) {
            return false;
        }
        String os = RuleEvaluator.currentOsName();
        if (library.natives != null && library.natives.containsKey(os)) {
            return true;
        }
        if (library.downloads.classifiers != null) {
            for (String key : library.downloads.classifiers.keySet()) {
                if (key.startsWith("natives-" + os)) {
                    return true;
                }
            }
        }
        if (library.downloads.artifact != null && library.downloads.artifact.path != null
                && library.downloads.artifact.path.toLowerCase().contains("natives-" + os)) {
            return true;
        }
        if (library.name != null) {
            String name = library.name.toLowerCase();
            if (name.contains(":natives-" + os)) {
                return true;
            }
            return name.contains(":natives")
                    && !name.contains(":natives-linux")
                    && !name.contains(":natives-osx")
                    && !name.contains(":natives-windows")
                    && !name.contains(":natives-macos");
        }
        return false;
    }

    public static NativeArtifactResult resolveNativeArtifact(VersionInfo.Library library, String os, String arch) {
        if (library == null || library.downloads == null) {
            return null;
        }

        if (library.natives != null && library.natives.containsKey(os)) {
            String template = library.natives.get(os);
            String archNum = "x86".equals(arch) ? "32" : "64";
            String classifierWithArch = template.replace("${arch}", archNum);
            String classifierWithoutArch = "natives-" + os;

            if (library.downloads.classifiers != null) {
                VersionInfo.Artifact artifact = library.downloads.classifiers.get(classifierWithArch);
                if (artifact != null) {
                    return new NativeArtifactResult(artifact, classifierWithArch);
                }
                artifact = library.downloads.classifiers.get(classifierWithoutArch);
                if (artifact != null) {
                    return new NativeArtifactResult(artifact, classifierWithoutArch);
                }
                for (var entry : library.downloads.classifiers.entrySet()) {
                    if (entry.getKey().startsWith("natives-" + os)) {
                        return new NativeArtifactResult(entry.getValue(), entry.getKey());
                    }
                }
            }
        }

        if (library.downloads.classifiers != null) {
            for (var entry : library.downloads.classifiers.entrySet()) {
                if (entry.getKey().startsWith("natives-" + os) && archMatches(entry.getKey(), arch)) {
                    return new NativeArtifactResult(entry.getValue(), entry.getKey());
                }
            }
            for (var entry : library.downloads.classifiers.entrySet()) {
                if (entry.getKey().startsWith("natives-" + os)) {
                    return new NativeArtifactResult(entry.getValue(), entry.getKey());
                }
            }
        }

        if (library.downloads.artifact != null && library.downloads.artifact.url != null) {
            String path = library.downloads.artifact.path;
            String name = library.name;
            boolean pathMatches = path != null && path.toLowerCase().contains("natives-" + os);
            boolean nameMatches = name != null && name.toLowerCase().contains(":natives-" + os);
            boolean genericNative = name != null && name.toLowerCase().contains(":natives")
                    && !name.toLowerCase().contains(":natives-linux")
                    && !name.toLowerCase().contains(":natives-osx")
                    && !name.toLowerCase().contains(":natives-windows")
                    && !name.toLowerCase().contains(":natives-macos");

            if ((pathMatches || nameMatches) && archMatches(path != null ? path : name, arch)) {
                return new NativeArtifactResult(library.downloads.artifact, null);
            }
            if (pathMatches || nameMatches || genericNative) {
                return new NativeArtifactResult(library.downloads.artifact, null);
            }
        }

        return null;
    }

    public static List<Path> resolveNativeJars(VersionInfo.Library library, Path librariesPath, String os, String arch) {
        List<Path> result = new ArrayList<>();
        NativeArtifactResult artifactResult = resolveNativeArtifact(library, os, arch);
        if (artifactResult != null) {
            String relativePath = artifactResult.artifact().path != null
                    ? artifactResult.artifact().path
                    : MavenPathResolver.toPath(library.name);
            Path candidate = librariesPath.resolve(relativePath);
            if (Files.exists(candidate)) {
                result.add(candidate);
            }
        }

        if (!result.isEmpty()) {
            return result;
        }

        if (library.downloads != null && library.downloads.classifiers != null) {
            for (var entry : library.downloads.classifiers.entrySet()) {
                VersionInfo.Artifact artifact = entry.getValue();
                if (artifact != null && artifact.path != null && entry.getKey().startsWith("natives-" + os)) {
                    Path candidate = librariesPath.resolve(artifact.path);
                    if (Files.exists(candidate)) {
                        result.add(candidate);
                    }
                }
            }
        }

        return result;
    }

    public static Path extractNatives(VersionInfo versionInfo, Path librariesPath, Path nativesDir) throws IOException {
        Files.createDirectories(nativesDir);
        if (versionInfo == null || versionInfo.libraries == null) {
            return nativesDir;
        }

        String os = RuleEvaluator.currentOsName();
        String arch = RuleEvaluator.currentArch();
        int extracted = 0;

        for (VersionInfo.Library library : versionInfo.libraries) {
            if (!library.isAllowed() || !isNativeLibrary(library)) {
                continue;
            }
            for (Path nativeJar : resolveNativeJars(library, librariesPath, os, arch)) {
                extracted += extractFromJar(nativeJar, nativesDir);
            }
        }

        CoreLogger.get().info(LOG, "Extracted " + extracted + " native file(s) into " + nativesDir);
        return nativesDir;
    }

    private static int extractFromJar(Path jarPath, Path nativesDir) {
        int extracted = 0;
        try (JarFile jarFile = new JarFile(jarPath.toFile())) {
            Enumeration<JarEntry> entries = jarFile.entries();
            while (entries.hasMoreElements()) {
                JarEntry entry = entries.nextElement();
                String name = entry.getName();
                if (entry.isDirectory() || name.startsWith("META-INF") || !isNativeFile(name)) {
                    continue;
                }
                String fileName = name.contains("/") ? name.substring(name.lastIndexOf('/') + 1) : name;
                Path destination = nativesDir.resolve(fileName);
                if (Files.exists(destination)) {
                    continue;
                }
                try (InputStream inputStream = jarFile.getInputStream(entry)) {
                    Files.copy(inputStream, destination);
                    destination.toFile().setExecutable(true);
                    extracted++;
                }
            }
        } catch (IOException ex) {
            CoreLogger.get().warn(LOG, "Failed to extract natives from " + jarPath + ": " + ex.getMessage());
        }
        return extracted;
    }

    private static boolean archMatches(String name, String arch) {
        if (name == null) {
            return true;
        }
        String lower = name.toLowerCase();
        if ("x86_64".equals(arch)) {
            if (lower.contains("arm64") || lower.contains("aarch64")) {
                return false;
            }
            return !lower.contains("-x86-") && !lower.endsWith("-x86.jar");
        }
        if ("arm64".equals(arch)) {
            return lower.contains("arm64") || lower.contains("aarch64");
        }
        if ("x86".equals(arch)) {
            return !lower.contains("arm64") && !lower.contains("x86_64");
        }
        return true;
    }

    private static boolean isNativeFile(String fileName) {
        String lower = fileName.toLowerCase();
        return lower.endsWith(".dll") || lower.endsWith(".so") || lower.endsWith(".dylib") || lower.endsWith(".jnilib");
    }

    public record NativeArtifactResult(VersionInfo.Artifact artifact, String classifierKey) {
    }
}
