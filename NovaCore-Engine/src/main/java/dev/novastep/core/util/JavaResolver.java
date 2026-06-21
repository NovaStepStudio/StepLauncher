package dev.novastep.core.util;

import dev.novastep.core.log.CoreLogger;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Optional;

public final class JavaResolver {

    private static final String LOG = "JavaResolver";

    private JavaResolver() {}

    public static String resolve(Path instancePath) {
        String os   = System.getProperty("os.name", "").toLowerCase();
        String exec = os.contains("win") ? "java.exe" : "java";

        Path runtimeBase = instancePath.resolve("runtime");
        if (Files.isDirectory(runtimeBase)) {
            Optional<Path> found = walkForExecutable(runtimeBase, exec, 6);
            if (found.isPresent()) return found.get().toAbsolutePath().toString();
        }

        String javaHome = System.getProperty("java.home");
        if (javaHome != null && !javaHome.isBlank()) {
            Path bin = Path.of(javaHome, "bin", exec);
            if (Files.exists(bin) && Files.isExecutable(bin)) {
                CoreLogger.get().debug(LOG, "Java from java.home: " + bin);
                return bin.toAbsolutePath().toString();
            }
        }

        return "java";
    }

    public static String resolve(Path instancePath, Path sharedPath) {
        if (sharedPath != null && Files.isDirectory(sharedPath)) {
            String os   = System.getProperty("os.name", "").toLowerCase();
            String exec = os.contains("win") ? "java.exe" : "java";
            Path sharedJava = sharedPath.resolve("runtime");
            if (Files.isDirectory(sharedJava)) {
                Optional<Path> found = walkForExecutable(sharedJava, exec, 6);
                if (found.isPresent()) return found.get().toAbsolutePath().toString();
            }
        }
        return resolve(instancePath);
    }

    private static Optional<Path> walkForExecutable(Path root, String execName, int depth) {
        try (var stream = Files.walk(root, depth)) {
            return stream
                    .filter(p -> p.getFileName().toString().equals(execName))
                    .filter(Files::isExecutable)
                    .findFirst();
        } catch (IOException e) {
            CoreLogger.get().error(LOG, "Failed to walk runtime directory for executable=" + execName + " in " + root, e);
            return Optional.empty();
        }
    }
}
