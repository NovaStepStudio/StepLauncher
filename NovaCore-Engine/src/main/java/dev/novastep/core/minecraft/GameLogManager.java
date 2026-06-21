package dev.novastep.core.minecraft;

import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.minecraft.version.VersionInfo;
import dev.novastep.core.server.request.LaunchRequest;

import java.io.*;
import java.nio.file.*;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class GameLogManager {
    private static final String LOG_PREFIX = "GameLogManager";
    private static final int DEFAULT_LIMIT = 1_500;
    private static final int MAX_LIMIT = 10_000;
    private static final int MIN_LIMIT = 100;

    private static final String REDACTED_PLACEHOLDER = "🔒🔒 (ˉ﹃ˉ) 🔒🔒";
    private static final DateTimeFormatter TIMESTAMP_FMT = DateTimeFormatter.ofPattern("yyyy-MM-dd_HH-mm-ss");
    private static final DateTimeFormatter LINE_FMT = DateTimeFormatter.ofPattern("HH:mm:ss.SSS");

    private static final Pattern MC_LOG_PATTERN = Pattern.compile(
            "^\\[(\\d{2}:\\d{2}:\\d{2})\\]\\s+\\[([^/]+)/(\\w+)\\](?:\\s+\\([^)]+\\))?:\\s+(.*)$");

    public static class ParsedLogLine {
        public final String time;
        public final String thread;
        public final String level;
        public final String logger;
        public final String message;

        ParsedLogLine(String time, String thread, String level, String logger, String message) {
            this.time = time;
            this.thread = thread;
            this.level = level;
            this.logger = logger;
            this.message = message;
        }
    }

    public static ParsedLogLine parseLine(String line) {
        if (line == null || line.isBlank())
            return new ParsedLogLine("", "", "RAW", "", line != null ? line : "");

        Matcher m = MC_LOG_PATTERN.matcher(line);
        if (m.matches()) {
            return new ParsedLogLine(
                    m.group(1),
                    m.group(2),
                    m.group(3).toUpperCase(),
                    m.group(2),
                    m.group(4));
        }
        return new ParsedLogLine("", "raw", "RAW", "raw", line);
    }

    private final Path logFile;
    private final PrintWriter writer;
    private final AtomicBoolean closed = new AtomicBoolean(false);

    private final int memoryLimit;
    private final AtomicInteger droppedLogs = new AtomicInteger(0);

    private final ArrayBlockingQueue<String> memoryBuffer;

    private GameLogManager(Path logFile, PrintWriter writer, int memoryLimit) {
        this.logFile = logFile;
        this.writer = writer;
        this.memoryLimit = Math.max(MIN_LIMIT, Math.min(MAX_LIMIT, memoryLimit));
        this.memoryBuffer = new ArrayBlockingQueue<>(this.memoryLimit);
    }

    public static GameLogManager open(Path rootDir, String launchId) throws IOException {
        return open(rootDir, launchId, DEFAULT_LIMIT);
    }

    public static GameLogManager open(Path rootDir, String launchId, int memoryLimit) throws IOException {
        Path logsDir = rootDir.resolve("logs-game");
        Files.createDirectories(logsDir);

        String timestamp = LocalDateTime.now().format(TIMESTAMP_FMT);
        String safeId = launchId.replaceAll("[^a-zA-Z0-9_\\-]", "_");
        Path logFile = logsDir.resolve("game-" + safeId + "-" + timestamp + ".log");

        PrintWriter writer = new PrintWriter(
                new BufferedWriter(new FileWriter(logFile.toFile(), true)), false);

        GameLogManager mgr = new GameLogManager(logFile, writer, memoryLimit);
        CoreLogger.get().info(LOG_PREFIX, "Game log session started: " + logFile
                + " (memLimit=" + mgr.memoryLimit + ")");
        return mgr;
    }

    public static GameLogManager openOrNull(Path rootDir, String launchId) {
        return openOrNull(rootDir, launchId, DEFAULT_LIMIT);
    }

    public static GameLogManager openOrNull(Path rootDir, String launchId, int memoryLimit) {
        try {
            return open(rootDir, launchId, memoryLimit);
        } catch (IOException ex) {
            CoreLogger.get().warn(LOG_PREFIX,
                    "Could not create game log for " + launchId + ": " + ex.getMessage());
            return new NullGameLogManager();
        }
    }

    public synchronized void writePreLaunchInfo(
            LaunchRequest req,
            VersionInfo effectiveInfo,
            String vanillaVersionId,
            String javaExec,
            String mainClass,
            List<String> fullCommand,
            ClasspathBuilder cpBuilder,
            ArgumentResolver argResolver) {

        if (closed.get())
            return;

        String border = "=".repeat(70);
        String dash = "-".repeat(70);

        writer.println(border);
        writer.println("  NovaCore-Engine — Game Log");
        writer.println("  Launch ID  : " + (req != null ? "will-be-assigned" : "unknown"));
        writer.println("  Started    : " + LocalDateTime.now());
        writer.println(border);
        writer.println();

        writer.println(dash);
        writer.println("  MINECRAFT CONFIGURATION");
        writer.println(dash);
        writer.println("  Version        : " + (req != null ? req.version : "?"));
        writer.println("  Vanilla Base   : " + vanillaVersionId);
        writer.println("  Main Class     : " + mainClass);
        writer.println("  Game Dir       : " + (req != null ? req.resolvedGameDir() : "?"));
        writer.println("  Assets Dir     : " + (req != null ? req.resolvedAssetsPath() : "?"));
        writer.println("  Libraries Dir  : " + (req != null ? req.resolvedLibrariesPath() : "?"));
        writer.println("  Asset Index    : " + (effectiveInfo != null && effectiveInfo.assetIndex != null
                ? effectiveInfo.assetIndex.id
                : "?"));
        writer.println();

        writer.println(dash);
        writer.println("  JVM CONFIGURATION");
        writer.println(dash);
        writer.println("  Java Exec   : " + javaExec);
        writer.println("  Min Memory  : " + (req != null ? req.resolvedMinMemory() : "?") + " MB");
        writer.println("  Max Memory  : " + (req != null ? req.resolvedMaxMemory() : "?") + " MB");
        writer.println("  GC Preset   : " + (req != null && req.gcPreset != null ? req.gcPreset : "auto"));
        writer.println("  GPU Pref    : " + (req != null && req.gpuPreference != null ? req.gpuPreference : "auto"));
        writer.println("  HW Accel    : " + (req != null ? req.disableHardwareAcceleration : "disabled"));
        if (req != null && req.isAuthlibEnabled()) {
            writer.println("  Authlib     : enabled — " + req.authlibInjector.serverUrl);
        }
        writer.println();

        writer.println(dash);
        writer.println("  CLASSPATH LIBRARIES");
        writer.println(dash);
        try {
            List<java.nio.file.Path> cp = cpBuilder.buildClasspathEntries();
            writer.println("  Total entries: " + cp.size());
            for (java.nio.file.Path p : cp) {
                boolean exists = Files.exists(p);
                writer.println("    [" + (exists ? "OK" : "MISSING") + "] " + p);
            }
        } catch (Exception e) {
            writer.println("  ERROR collecting classpath: " + e.getMessage());
        }
        writer.println();

        writer.println(dash);
        writer.println("  NATIVE LIBRARIES");
        writer.println(dash);
        try {
            java.nio.file.Path nativesDir = java.nio.file.Path.of(
                    req != null ? req.resolvedInstancePath() : ".")
                    .resolve("versions").resolve(vanillaVersionId).resolve("natives");
            if (Files.isDirectory(nativesDir)) {
                try (var stream = Files.list(nativesDir)) {
                    List<java.nio.file.Path> natives = stream.sorted().toList();
                    writer.println("  Dir     : " + nativesDir);
                    writer.println("  Total   : " + natives.size());
                    for (java.nio.file.Path n : natives) {
                        writer.println("    " + n.getFileName());
                    }
                }
            } else {
                writer.println("  Dir: " + nativesDir + " (not found)");
            }
        } catch (Exception e) {
            writer.println("  ERROR collecting natives: " + e.getMessage());
        }
        writer.println();

        writer.println(dash);
        writer.println("  FULL LAUNCH COMMAND (" + fullCommand.size() + " args)");
        writer.println(dash);
        for (int i = 0; i < fullCommand.size(); i++) {
            String arg = fullCommand.get(i);
            arg = redactSensitive(arg);
            writer.println("  [" + String.format("%3d", i) + "] " + arg);
        }
        writer.println();

        if (effectiveInfo != null) {
            writer.println(dash);
            writer.println("  VERSION MANIFEST");
            writer.println(dash);
            writer.println("  ID          : " + effectiveInfo.id);
            writer.println("  Type        : " + effectiveInfo.type);
            writer.println("  Main Class  : " + effectiveInfo.mainClass);
            writer.println(
                    "  Inherits    : " + (effectiveInfo.inheritsFrom != null ? effectiveInfo.inheritsFrom : "none"));
            if (effectiveInfo.libraries != null) {
                writer.println("  Libraries   : " + effectiveInfo.libraries.size() + " declared");
            }
            writer.println();
        }

        writer.println(border);
        writer.println("  GAME OUTPUT FOLLOWS");
        writer.println(border);
        writer.println();
        writer.flush();
    }

    public synchronized void log(String stream, String line) {
        if (closed.get())
            return;

        String ts = LocalDateTime.now().format(LINE_FMT);
        String fileLog = "[" + ts + "] [" + stream.toUpperCase() + "] " + line;

        writer.println(fileLog);
        writer.flush();

        if (!memoryBuffer.offer(fileLog)) {
            memoryBuffer.poll();
            memoryBuffer.offer(fileLog);
            droppedLogs.incrementAndGet();
        }
    }

    public List<String> getMemoryBuffer() {
        return new ArrayList<>(memoryBuffer);
    }

    public int getDroppedCount() {
        return droppedLogs.get();
    }

    public Path getLogFile() {
        return logFile;
    }

    public synchronized void close() {
        if (closed.compareAndSet(false, true)) {
            writeFooter();
            writer.flush();
            writer.close();
            CoreLogger.get().info(LOG_PREFIX, "Game log closed: " + logFile.getFileName()
                    + " (dropped=" + droppedLogs.get() + ")");
        }
    }

    private void writeFooter() {
        writer.println();
        String border = "-".repeat(70);
        writer.println(border);
        writer.println("  Log closed: " + LocalDateTime.now());
        if (droppedLogs.get() > 0) {
            writer.println("  WARNING: " + droppedLogs.get()
                    + " log line(s) were dropped from memory buffer (limit=" + memoryLimit + ")");
        }
        writer.println(border);
    }

    private static final class NullGameLogManager extends GameLogManager {
        private static final OutputStream DEV_NULL = new OutputStream() {
            @Override
            public void write(int b) {
            }

            @Override
            public void write(byte[] b, int off, int len) {
            }
        };

        NullGameLogManager() {
            super(Path.of("/dev/null"), new PrintWriter(DEV_NULL, false), DEFAULT_LIMIT);
        }

        @Override
        public synchronized void log(String stream, String line) {
        }

        @Override
        public synchronized void close() {
        }

        @Override
        public synchronized void writePreLaunchInfo(
                LaunchRequest req, VersionInfo effectiveInfo, String vanillaVersionId,
                String javaExec, String mainClass, List<String> fullCommand,
                ClasspathBuilder cpBuilder, ArgumentResolver argResolver) {
        }
    }

    private static final Pattern[] SENSITIVE_PATTERNS = {
            Pattern.compile("(--?accessToken[= ]+)[^ ]+", Pattern.CASE_INSENSITIVE),
            Pattern.compile("(--?session[= ]+)[^ ]+", Pattern.CASE_INSENSITIVE),
            Pattern.compile("(--?user(?:name|properties)?[= ]+)[^ ]+", Pattern.CASE_INSENSITIVE),
            Pattern.compile("(--?password[= ]+)[^ ]+", Pattern.CASE_INSENSITIVE),
            Pattern.compile("(--?token[= ]+)[^ ]+", Pattern.CASE_INSENSITIVE),
            Pattern.compile("(Authorization:?\\s*)\\S+", Pattern.CASE_INSENSITIVE),
            Pattern.compile("(X-Auth-Token:?\\s*)\\S+", Pattern.CASE_INSENSITIVE)
    };

    private static String redactSensitive(String line) {
        if (line == null)
            return "";
        String result = line;
        for (Pattern p : SENSITIVE_PATTERNS) {
            Matcher m = p.matcher(result);
            if (m.find()) {
                result = m.replaceAll("$1" + REDACTED_PLACEHOLDER);
            }
        }
        return result;
    }
}
