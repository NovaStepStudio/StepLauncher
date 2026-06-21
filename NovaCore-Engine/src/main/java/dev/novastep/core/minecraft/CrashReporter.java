package dev.novastep.core.minecraft;

import dev.novastep.core.log.CoreLogger;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.LinkedHashMap;

public class CrashReporter {
    
    private static final String LOG = "CrashReporter";
    private static final int MAX_CONTEXT_LINES = 25;
    
    private static Map<String, Object> latestCrash = null;

    public static Map<String, Object> buildCrashContext(String instanceId, int exitCode, String reason, List<String> buffer) {
        List<String> contextLines;
        if (buffer == null || buffer.isEmpty()) {
            contextLines = Collections.emptyList();
        } else {
            int size = buffer.size();
            int from = Math.max(0, size - MAX_CONTEXT_LINES);
            contextLines = buffer.subList(from, size);
        }

        Map<String, Object> crashReport = new LinkedHashMap<>();
        crashReport.put("instanceId", instanceId);
        crashReport.put("exitCode", exitCode);
        crashReport.put("source", determineSource(reason, contextLines));
        crashReport.put("reason", reason);
        crashReport.put("context", contextLines);
        crashReport.put("timestamp", System.currentTimeMillis());

        latestCrash = crashReport;
        CoreLogger.get().debug(LOG, "Contexto previo adjuntado (lines=" + contextLines.size() + ")");
        return crashReport;
    }

    public static void recordCrashReportFile(String instanceId, int exitCode, Path crashReportFile) {
        if (crashReportFile == null)
            return;
        List<String> lines = new ArrayList<>();
        try {
            List<String> all = Files.readAllLines(crashReportFile, StandardCharsets.UTF_8);
            int limit = Math.min(all.size(), MAX_CONTEXT_LINES);
            lines = all.subList(0, limit);
        } catch (IOException e) {
            lines = List.of("Could not read crash report: " + e.getMessage(), String.valueOf(crashReportFile));
        }

        Map<String, Object> crashReport = new LinkedHashMap<>();
        crashReport.put("instanceId", instanceId);
        crashReport.put("exitCode", exitCode);
        crashReport.put("source", "crash_report_file");
        crashReport.put("reason", "minecraft_crash_report");
        crashReport.put("context", lines);
        crashReport.put("timestamp", System.currentTimeMillis());
        crashReport.put("file", crashReportFile.toAbsolutePath().toString());
        latestCrash = crashReport;
    }

    private static String determineSource(String reason, List<String> contextLines) {
        if ("sigsegv".equals(reason) || "killed_or_oom".equals(reason)) return "jvm";
        
        for (String line : contextLines) {
            String lower = line.toLowerCase();
            if (lower.contains("mixin") || lower.contains("modloader") || lower.contains("spongepowered")) {
                return "modloader";
            }
            if (lower.contains("java.lang.outofmemoryerror")) {
                return "jvm_memory";
            }
            if (lower.contains("org.lwjgl")) {
                return "graphics";
            }
        }
        return "game";
    }

    public static Map<String, Object> getLatestCrash() {
        return latestCrash;
    }
}
