package dev.novastep.core.minecraft.diagnostic;

import java.nio.file.Path;
import java.nio.file.Paths;

public final class ContextAnalyzerRunner {

    public static void main(String[] args) throws Exception {
        if (args.length == 0) {
            System.out.println("Usage: ContextAnalyzerRunner <log-file-path>");
            System.out.println("Example: ContextAnalyzerRunner core/logs/logs-game/game-launch-1777350430645-1-2026-04-28_01-27-11.log");
            return;
        }

        Path logFile = Paths.get(args[0]);
        System.out.println("Analyzing log file: " + logFile);

        ContextAnalyzer.LaunchDiagnostic diag = ContextAnalyzer.analyzeLaunchLog(logFile);

        System.out.println("=== LAUNCH DIAGNOSTIC ===");
        System.out.println("Launch ID: " + diag.launchId());
        System.out.println("Version: " + diag.version() + " (vanilla: " + diag.vanillaVersion() + ")");
        System.out.println("Main Class: " + diag.mainClass());
        System.out.println("Classpath: " + diag.classpathEntries() + "/" + diag.declaredLibraries() +
                (diag.missingEntries() > 0 ? " (" + diag.missingEntries() + " missing)" : ""));
        System.out.println("Missing Entries: " + diag.missingEntries());

        if (!diag.errors().isEmpty()) {
            System.out.println("\n=== ERRORS FOUND ===");
            for (String error : diag.errors()) {
                System.out.println("- " + error);
            }
        }

        System.out.println("\n=== RECOMMENDATION ===");
        System.out.println(diag.recommendedAction());
    }
}