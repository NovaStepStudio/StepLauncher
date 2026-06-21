package dev.novastep.core.minecraft.diagnostic;

import dev.novastep.core.log.CoreLogger;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.*;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public final class ContextAnalyzer {

    private static final String LOG = "ContextAnalyzer";

    public record LaunchDiagnostic(
            String launchId,
            String version,
            String vanillaVersion,
            String mainClass,
            int classpathEntries,
            int missingEntries,
            int declaredLibraries,
            List<String> errors,
            String recommendedAction
    ) {}

    public static LaunchDiagnostic analyzeLaunchLog(Path logFile) throws IOException {
        List<String> lines = Files.readAllLines(logFile);
        return analyzeLines(lines, logFile.getFileName().toString());
    }

    private static LaunchDiagnostic analyzeLines(List<String> lines, String fileName) {
        String launchId = extractField(lines, "Launch ID", null);
        String version = extractField(lines, "Version", null);
        String vanillaVersion = extractVersionFromId(version);
        String mainClass = extractField(lines, "Main Class", null);
        int classpathCount = countPattern(lines, "\\[OK\\].*\\.jar");
        int missingCount = countPattern(lines, "\\[MISSING\\].*\\.jar");
        int declaredCount = Integer.parseInt(
                extractPattern(lines, "Libraries\\s*:\\s*(\\d+) declared", "0"));

        List<String> errors = extractErrors(lines);
        String recommendation = generateRecommendation(
                version, classpathCount, declaredCount, missingCount, errors, mainClass);

        return new LaunchDiagnostic(
                launchId,
                version,
                vanillaVersion,
                mainClass,
                classpathCount,
                missingCount,
                declaredCount,
                errors,
                recommendation
        );
    }

    private static String extractField(List<String> lines, String fieldName, String defaultValue) {
        for (String line : lines) {
            if (line.contains(fieldName)) {
                String[] parts = line.split(":", 2);
                if (parts.length == 2) {
                    return parts[1].trim();
                }
            }
        }
        return defaultValue;
    }

    private static String extractPattern(List<String> lines, String regex, String defaultValue) {
        Pattern p = Pattern.compile(regex);
        for (String line : lines) {
            Matcher m = p.matcher(line);
            if (m.find()) {
                return m.group(1);
            }
        }
        return defaultValue;
    }

    private static int countPattern(List<String> lines, String regex) {
        Pattern p = Pattern.compile(regex);
        int count = 0;
        for (String line : lines) {
            if (p.matcher(line).find()) count++;
        }
        return count;
    }

    private static List<String> extractErrors(List<String> lines) {
        List<String> errors = new ArrayList<>();
        for (String line : lines) {
            if (line.contains("[STDERR]") || line.contains("Error:") || line.contains("Exception")) {
                errors.add(line.trim());
            }
        }
        return errors;
    }

    private static String generateRecommendation(
            String version, int classpathCount, int declaredCount,
            int missingCount, List<String> errors, String mainClass) {

        StringBuilder rec = new StringBuilder();

        for (String error : errors) {
            if (error.contains("ClassNotFoundException") || error.contains("no se ha encontrado")) {
                rec.append("CRITICAL: Main class not found in classpath. ");
                if (mainClass != null && mainClass.contains("quilt")) {
                    rec.append("Quilt Loader libraries incomplete. ");
                }
                rec.append("Action: Verify all Quilt/Forge libraries are resolved and downloaded. ");
            }
        }

        if (classpathCount < declaredCount) {
            int diff = declaredCount - classpathCount;
            rec.append(String.format("WARNING: %d libraries missing from classpath (%d found, %d declared). ",
                    diff, classpathCount, declaredCount));
            rec.append("Action: Run InstallVerifier.verifyIntegrity() to detect and download missing libraries. ");
        }

        if (missingCount > 0) {
            rec.append(String.format("ERROR: %d classpath entries marked [MISSING]. ", missingCount));
            rec.append("Action: Inspect logs to identify which libraries failed to download. ");
        }

        if (rec.isEmpty()) {
            rec.append("Launch appears successful or errors are non-critical. ");
        }

        return rec.toString();
    }

    private static String extractVersionFromId(String versionId) {
        if (versionId == null) return null;
        if (versionId.contains("-")) {
            String[] parts = versionId.split("-");
            for (int i = parts.length - 1; i >= 0; i--) {
                if (parts[i].matches("\\d+\\.\\d+.*")) {
                    return parts[i];
                }
            }
        }
        return versionId;
    }

    public static void analyzeLaunchDirectory(Path logDir) throws IOException {
        if (!Files.isDirectory(logDir)) {
            CoreLogger.get().warn(LOG, "Not a directory: " + logDir);
            return;
        }

        List<Path> logFiles = Files.list(logDir)
                .filter(p -> p.getFileName().toString().startsWith("game-launch-"))
                .filter(p -> p.getFileName().toString().endsWith(".log"))
                .sorted(Comparator.reverseOrder())
                .toList();

        if (logFiles.isEmpty()) {
            CoreLogger.get().info(LOG, "No launch logs found in: " + logDir);
            return;
        }

        CoreLogger.get().info(LOG, "Analyzing " + logFiles.size() + " launch logs...");

        for (Path logFile : logFiles.stream().limit(5).toList()) {
            try {
                LaunchDiagnostic diag = analyzeLaunchLog(logFile);
                CoreLogger.get().info(LOG, "Log: " + logFile.getFileName());
                CoreLogger.get().info(LOG, "  Version: " + diag.version + " (vanilla: " + diag.vanillaVersion + ")");
                CoreLogger.get().info(LOG, "  Main Class: " + diag.mainClass);
                CoreLogger.get().info(LOG, "  Classpath: " + diag.classpathEntries + "/" + diag.declaredLibraries +
                        (diag.missingEntries > 0 ? " (" + diag.missingEntries + " missing)" : ""));
                
                if (!diag.errors.isEmpty()) {
                    CoreLogger.get().warn(LOG, "  Errors found:");
                    for (String error : diag.errors.stream().limit(3).toList()) {
                        CoreLogger.get().warn(LOG, "    - " + error);
                    }
                }
                
                CoreLogger.get().info(LOG, "  → " + diag.recommendedAction);
                CoreLogger.get().info(LOG, "");
            } catch (IOException e) {
                CoreLogger.get().error(LOG, "Failed to analyze " + logFile, e);
            }
        }
    }
}
