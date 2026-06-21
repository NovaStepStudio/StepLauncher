package dev.novastep.core.minecraft;

import dev.novastep.core.json.Json;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.minecraft.manifest.VersionMerger;
import dev.novastep.core.minecraft.version.VersionInfo;
import dev.novastep.core.modloader.ModLoaderOrchestrator;
import dev.novastep.core.modloader.ModLoaderProvider;
import dev.novastep.core.modloader.ModLoaderRegistry;
import dev.novastep.core.modloader.model.ModLoaderModels.ExecutionPlan;
import dev.novastep.core.modloader.model.ModLoaderModels.InstalledLoader;
import dev.novastep.core.minecraft.instance.InstanceTechnicalMetadataStore;
import dev.novastep.core.server.request.LaunchRequest;
import dev.novastep.core.util.JavaResolver;
import dev.novastep.core.util.ProcessUtils;
import dev.novastep.core.websocket.EventBroadcaster;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Instant;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;

public class MinecraftLauncher {

    private static final String LOG = "MinecraftLauncher";
    // JSON deserialization handled via Jackson (Json.java) — Gson removed

    private final EventBroadcaster broadcaster;
    private final ModLoaderOrchestrator modLoaderOrchestrator;

    private final ConcurrentHashMap<String, ActiveInstance> activeInstances = new ConcurrentHashMap<>();
    private final AtomicInteger launchCounter = new AtomicInteger(0);

    public static class ActiveInstance {
        public final String launchId;
        public final String version;
        public final String username;
        public final String instancePath;
        public final long startedAt;
        public volatile long pid = -1;
        public volatile String status = "starting";
        public volatile int exitCode = -1;
        public volatile String logFile = null;

        private volatile Process process;

        public ActiveInstance(String launchId, String version, String username, String instancePath) {
            this.launchId = launchId;
            this.version = version;
            this.username = username;
            this.instancePath = instancePath;
            this.startedAt = Instant.now().toEpochMilli();
        }

        public Map<String, Object> toMap() {
            Map<String, Object> m = new LinkedHashMap<>();
            m.put("launchId", launchId);
            m.put("version", version);
            m.put("username", username);
            m.put("instancePath", instancePath);
            m.put("startedAt", startedAt);
            m.put("pid", pid);
            m.put("status", status);
            m.put("exitCode", exitCode);
            m.put("logFile", logFile);
            return m;
        }

        public boolean isAlive() {
            return process != null && process.isAlive();
        }

        void attachProcess(Process p) {
            this.process = p;
            this.pid = p.pid();
            this.status = "running";
        }

        boolean kill() {
            if (process != null && process.isAlive()) {
                ProcessUtils.killTree(pid);
                status = "stopping";
                return true;
            }
            return false;
        }
    }

    public MinecraftLauncher(EventBroadcaster broadcaster, ModLoaderOrchestrator modLoaderOrchestrator) {
        this(broadcaster, modLoaderOrchestrator, null);
    }

    public MinecraftLauncher(EventBroadcaster broadcaster,
            ModLoaderOrchestrator modLoaderOrchestrator,
            Path rootDir) {
        this.broadcaster = broadcaster;
        this.modLoaderOrchestrator = modLoaderOrchestrator;
    }

    public String launch(LaunchRequest req) {
        String launchId = "launch-" + System.currentTimeMillis() + "-" + launchCounter.incrementAndGet();
        CoreLogger.get().info(LOG, "Launch requested: " + launchId
                + " v=" + req.version + " user=" + req.resolvedUsername());

        ActiveInstance instance = new ActiveInstance(
                launchId, req.version,
                req.resolvedUsername(),
                req.resolvedInstancePath());
        activeInstances.put(launchId, instance);

        Thread.ofVirtual().name("launch-" + launchId).start(() -> {
            try {
                runLaunch(launchId, req, instance);
            } catch (Throwable t) {
                String msg = t.getClass().getSimpleName() + ": " + t.getMessage();
                CoreLogger.get().fatal(LOG,
                        "CRITICAL JVM ERROR during launch of " + launchId + " (version=" + req.version + ")", t);
                instance.status = "failed";
                broadcaster.emit("launch_failed", Map.of("launchId", launchId, "error", msg));
                activeInstances.remove(launchId);
                if (t instanceof Error)
                    throw (Error) t;
                if (t instanceof RuntimeException)
                    throw (RuntimeException) t;
                throw new RuntimeException(t);
            }
        });

        return launchId;
    }

    public boolean kill(String launchId) {
        ActiveInstance inst = activeInstances.get(launchId);
        if (inst != null && inst.kill()) {
            CoreLogger.get().info(LOG, "Process killed: " + launchId);
            return true;
        }
        return false;
    }

    public void killAll() {
        for (ActiveInstance inst : activeInstances.values()) {
            inst.kill();
        }
    }

    public boolean isRunning(String launchId) {
        ActiveInstance inst = activeInstances.get(launchId);
        return inst != null && inst.isAlive();
    }

    public Map<String, Object> getInstanceData(String launchId) {
        ActiveInstance inst = activeInstances.get(launchId);
        return inst != null ? inst.toMap() : null;
    }

    public List<Map<String, Object>> getAllInstances() {
        List<Map<String, Object>> result = new ArrayList<>();
        for (ActiveInstance inst : activeInstances.values()) {
            result.add(inst.toMap());
        }
        result.sort(Comparator.comparingLong(m -> (long) m.get("startedAt")));
        return result;
    }

    public Collection<ActiveInstance> getActiveLaunches() {
        return activeInstances.values();
    }

    public int getRunningCount() {
        int count = 0;
        for (ActiveInstance inst : activeInstances.values()) {
            if (inst.isAlive())
                count++;
        }
        return count;
    }

    private void runLaunch(String launchId, LaunchRequest req, ActiveInstance instance) throws Exception {
        Path instancePath = Path.of(req.resolvedInstancePath()).toAbsolutePath();

        broadcaster.emit("launch_preparing", Map.of("launchId", launchId, "version", req.version));

        String vanillaVersionId = resolveVanillaVersionId(req.version, instancePath);
        CoreLogger.get().info(LOG, "[" + launchId + "] Vanilla base: " + vanillaVersionId);

        VersionInfo effectiveInfo = loadLocalVersionInfo(req.version, instancePath);

        String loaderType = detectLoaderType(req.version, vanillaVersionId);
        ExecutionPlan executionPlan = null;

        if (loaderType != null) {
            CoreLogger.get().info(LOG, "[" + launchId + "] Modloader detectado: " + loaderType);
            Optional<ModLoaderProvider> provider = ModLoaderRegistry.get().find(loaderType);
            if (provider.isPresent()) {
                InstalledLoader syntheticLoader = new InstalledLoader();
                syntheticLoader.loaderType = loaderType;
                syntheticLoader.versionJsonId = req.version;
                syntheticLoader.minecraftVersion = vanillaVersionId;

                Optional<InstalledLoader> savedState = modLoaderOrchestrator.loadState(instancePath);
                if (savedState.isPresent() && loaderType.equals(savedState.get().loaderType)) {
                    syntheticLoader.versionJsonId = savedState.get().versionJsonId;
                }

                VersionInfo vanillaOnly = loadLocalVersionInfo(vanillaVersionId, instancePath);
                executionPlan = provider.get().buildExecution(
                        syntheticLoader, vanillaOnly, instancePath,
                        req.resolvedLibrariesPath().toAbsolutePath());
            }
        }

        MinecraftVerifier.VerificationResult verification = MinecraftVerifier.verify(req, effectiveInfo,
                vanillaVersionId);
        if (!verification.ok) {
            List<String> missingNames = verification.missing.stream()
                    .map(c -> c.category() + "/" + c.description()).toList();
            broadcaster.emit("launch_verification_failed", Map.of(
                    "launchId", launchId, "missing", missingNames,
                    "hint", "Re-run install to repair the missing components"));
            CoreLogger.get().warn(LOG, "[" + launchId + "] Pre-launch check FAILED — "
                    + verification.missing.size() + " component(s) missing: " + missingNames);
            throw new IllegalStateException("Pre-launch verification failed — missing: " + missingNames);
        }

        String javaExec = resolveJavaExecutable(req, instancePath,
                effectiveInfo != null ? effectiveInfo.javaVersion : null);
        CoreLogger.get().info(LOG, "[" + launchId + "] Java: " + javaExec);

        ArgumentResolver argResolver = ArgumentResolver.fromRequest(req, effectiveInfo, instancePath, vanillaVersionId);
        ClasspathBuilder cpBuilder = ClasspathBuilder.fromRequest(req, effectiveInfo, vanillaVersionId);

        if (executionPlan != null && !executionPlan.additionalClasspath.isEmpty()) {
            cpBuilder.appendModloaderEntries(executionPlan.additionalClasspath);
        }

        String mainClass = effectiveInfo.mainClass;
        if (executionPlan != null && executionPlan.mainClass != null) {
            mainClass = executionPlan.mainClass;
        }

        List<String> command = buildCommand(javaExec, mainClass, argResolver, cpBuilder, req);

        CoreLogger.get().info(LOG, "[" + launchId + "] Launching: mainClass=" + mainClass
                + " totalArgs=" + command.size());

        broadcaster.emit("launch_starting", Map.of(
                "launchId", launchId,
                "mainClass", mainClass,
                "version", req.version));

        Path workDir = instancePath.resolve("game");
        Files.createDirectories(workDir);

        ProcessBuilder pb = new ProcessBuilder(command)
                .directory(workDir.toFile())
                .redirectErrorStream(false);

        applyGpuEnvironment(pb, req.gpuPreference);

        Path logRoot = CoreLogger.get().getLogFile().getParent();
        GameLogManager gameLogs = GameLogManager.openOrNull(logRoot, launchId);

        if (gameLogs != null) {
            instance.logFile = gameLogs.getLogFile().toAbsolutePath().toString();
            gameLogs.writePreLaunchInfo(
                    req, effectiveInfo, vanillaVersionId,
                    javaExec, mainClass, command, cpBuilder, argResolver);
        }

        Process process = pb.start();

        instance.attachProcess(process);

        broadcaster.emit("launch_started", Map.of(
                "launchId", launchId,
                "pid", process.pid(),
                "logFile", instance.logFile != null ? instance.logFile : ""));
        CoreLogger.get().info(LOG, "[" + launchId + "] PID=" + process.pid());

        streamOutput(launchId, process, gameLogs, req);

        long startMs = System.currentTimeMillis();
        int exitCode = process.waitFor();
        long durationMs = System.currentTimeMillis() - startMs;

        try {
            InstanceTechnicalMetadataStore.markLastPlayed(instancePath, req.version);
        } catch (Exception ex) {
            CoreLogger.get().error(LOG,
                    "Failed to mark last played for instance at " + instancePath + " (version=" + req.version + ")",
                    ex);
        }

        instance.exitCode = exitCode;
        instance.status = exitCode == 0 ? "completed" : "failed";
        activeInstances.remove(launchId);

        if (gameLogs != null)
            gameLogs.close();

        CoreLogger.get().info(LOG, "[" + launchId + "] Exited: " + exitCode
                + " (runtime=" + durationMs + "ms)");

        broadcaster.emit("launch_exited", Map.of(
                "launchId", launchId,
                "exitCode", exitCode,
                "normal", exitCode == 0,
                "durationMs", durationMs));

        Map<String, Object> sessionData = new LinkedHashMap<>();
        sessionData.put("launchId", launchId);
        sessionData.put("instancePath", req.resolvedInstancePath());
        sessionData.put("version", req.version);
        sessionData.put("durationMs", durationMs);
        sessionData.put("startedAt", startMs);
        sessionData.put("exitCode", exitCode);
        SessionManager.recordSession(sessionData);

        if (exitCode != 0) {
            List<String> logsBuffer = gameLogs != null ? gameLogs.getMemoryBuffer() : Collections.emptyList();
            String reason = resolveCrashReason(exitCode);
            Map<String, Object> crashContext = CrashReporter.buildCrashContext(launchId, exitCode, reason, logsBuffer);

            broadcaster.emit("game_crash", crashContext);
            CoreLogger.get().warn(LOG, "[" + launchId + "] CRASH exitCode=" + exitCode);
        }

        try {
            Path crashDir = Path.of(req.resolvedGameDir()).resolve("crash-reports");
            if (Files.isDirectory(crashDir)) {
                Path latest = null;
                long latestMs = -1;
                try (var stream = Files.list(crashDir)) {
                    for (Path p : stream.toList()) {
                        if (!Files.isRegularFile(p))
                            continue;
                        String n = p.getFileName().toString().toLowerCase();
                        if (!n.endsWith(".txt"))
                            continue;
                        long ms = Files.getLastModifiedTime(p).toMillis();
                        if (ms > latestMs) {
                            latestMs = ms;
                            latest = p;
                        }
                    }
                }
                if (latest != null) {
                    CrashReporter.recordCrashReportFile(launchId, exitCode, latest);
                }
            }
        } catch (Exception ex) {
            CoreLogger.get().error(LOG,
                    "Failed to scan or record crash reports for launchId=" + launchId + " in " + req.resolvedGameDir(),
                    ex);
        }
    }

    private static String resolveCrashReason(int code) {
        return switch (code) {
            case 0 -> "normal_exit";
            case 1 -> "generic_error";
            case -1 -> "killed_or_oom";
            case 134 -> "sigsegv_or_abort";
            case 139 -> "sigsegv";
            case 143 -> "sigterm";
            case 130 -> "sigint";
            default -> exitCodeLabel(code);
        };
    }

    private static String exitCodeLabel(int c) {
        if (c > 128)
            return "signal_" + (c - 128);
        return "exit_" + c;
    }

    private void streamOutput(String launchId, Process process, GameLogManager gameLogs, LaunchRequest req) {
        Thread.ofVirtual().name("stdout-" + launchId).start(() -> {
            try (BufferedReader reader = new BufferedReader(
                    new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    if (gameLogs != null)
                        gameLogs.log("stdout", line);
                    GameLogManager.ParsedLogLine parsed = GameLogManager.parseLine(line);
                    dispatchGameLog(launchId, line, "stdout", parsed);
                }
            } catch (IOException ex) {
                CoreLogger.get().warn(LOG, "[" + launchId + "] stdout stream closed: " + ex.getMessage());
                if (gameLogs != null)
                    gameLogs.log("system", "stdout stream error: " + ex.getMessage());
            }
        });

        Thread.ofVirtual().name("stderr-" + launchId).start(() -> {
            try (BufferedReader reader = new BufferedReader(
                    new InputStreamReader(process.getErrorStream(), StandardCharsets.UTF_8))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    if (gameLogs != null)
                        gameLogs.log("stderr", line);
                    GameLogManager.ParsedLogLine parsed = GameLogManager.parseLine(line);
                    dispatchGameLog(launchId, line, "stderr", parsed);
                }
            } catch (IOException ex) {
                CoreLogger.get().warn(LOG, "[" + launchId + "] stderr stream closed: " + ex.getMessage());
                if (gameLogs != null)
                    gameLogs.log("system", "stderr stream error: " + ex.getMessage());
            }
        });
    }

    private void dispatchGameLog(String launchId, String rawLine,
            String stream, GameLogManager.ParsedLogLine parsed) {
        broadcaster.emit("game_log", Map.of(
                "launchId", launchId,
                "line", rawLine,
                "stream", stream,
                "level", parsed.level,
                "logger", parsed.logger,
                "message", parsed.message));

        switch (parsed.level) {
            case "WARN" -> broadcaster.emit("game_log_warn",
                    Map.of("launchId", launchId, "line", rawLine, "logger", parsed.logger));
            case "ERROR" -> broadcaster.emit("game_log_error",
                    Map.of("launchId", launchId, "line", rawLine, "logger", parsed.logger));
            case "FATAL" -> broadcaster.emit("game_log_fatal",
                    Map.of("launchId", launchId, "line", rawLine, "logger", parsed.logger));
        }

        if ("stdout".equals(stream)) {
            broadcaster.emit("game_stdout", Map.of("launchId", launchId, "line", rawLine));
        } else {
            broadcaster.emit("game_stderr", Map.of("launchId", launchId, "line", rawLine));
        }
    }

    private static int detectJavaMajorVersion(String javaExec) {
        try {
            Process p = new ProcessBuilder(javaExec, "-version")
                    .redirectErrorStream(true)
                    .start();
            String output = new String(p.getInputStream().readAllBytes(), StandardCharsets.UTF_8).trim();
            p.waitFor();
            for (String line : output.split("\n")) {
                line = line.trim();
                java.util.regex.Matcher m = java.util.regex.Pattern
                        .compile("version \"(\\d+)(?:\\.(\\d+))?")
                        .matcher(line);
                if (m.find()) {
                    int first = Integer.parseInt(m.group(1));
                    if (first == 1 && m.group(2) != null)
                        return Integer.parseInt(m.group(2));
                    return first;
                }
            }
        } catch (Exception ex) {
            CoreLogger.get().error(LOG, "Failed to detect Java major version for executable: " + javaExec, ex);
        }
        return 8;
    }

    private List<String> buildCommand(String javaExec, String mainClass,
            ArgumentResolver argResolver,
            ClasspathBuilder cpBuilder,
            LaunchRequest req) {
        List<String> cmd = new ArrayList<>();
        cmd.add(javaExec);

        if (req.jvm != null && req.jvm.prependArgs != null)
            cmd.addAll(req.jvm.prependArgs);

        int minMem = req.resolvedMinMemory();
        int maxMem = req.resolvedMaxMemory();
        if (minMem > 0)
            cmd.add("-Xms" + minMem + "m");
        if (maxMem > 0)
            cmd.add("-Xmx" + maxMem + "m");

        cmd.addAll(buildGcArgs(req.gcPreset));

        if (req.isHardwareAccelerationDisabled() && !Boolean.FALSE.equals(req.hardwareAcceleration)) {
            cmd.add("-Dsun.java2d.d3d=false");
            cmd.add("-Dsun.java2d.opengl=false");
            cmd.add("-Dsun.java2d.noddraw=true");
            cmd.add("-Dsun.java2d.ddoffscreen=false");
        }

        if (detectJavaMajorVersion(javaExec) >= 17)
            cmd.add("--enable-native-access=ALL-UNNAMED");

        if (req.isAuthlibEnabled()) {
            cmd.add("-javaagent:" + req.authlibInjector.jarPath
                    + "=" + req.authlibInjector.serverUrl);
        }

        cmd.add("-Dminecraft.telemetry.disabled=true");

        if (req.game != null && req.game.extraJvmProperties != null) {
            for (var entry : req.game.extraJvmProperties.entrySet()) {
                cmd.add("-D" + entry.getKey() + "=" + entry.getValue());
            }
        }

        cmd.addAll(argResolver.buildJvmArgs(cpBuilder));

        if (req.jvm != null && req.jvm.extraArgs != null)
            cmd.addAll(req.jvm.extraArgs);

        cmd.add(mainClass);
        cmd.addAll(argResolver.buildGameArgs());

        if (req.game != null && req.game.extraGameArgs != null)
            cmd.addAll(req.game.extraGameArgs);

        if (req.game != null && req.game.serverHost != null && !req.game.serverHost.isBlank()) {
            cmd.add("--server");
            cmd.add(req.game.serverHost);
            if (req.game.serverPort != null) {
                cmd.add("--port");
                cmd.add(String.valueOf(req.game.serverPort));
            }
        }

        return cmd;
    }

    private static List<String> buildGcArgs(String gcPreset) {
        if (gcPreset == null || gcPreset.isBlank())
            return List.of();

        return switch (gcPreset.toLowerCase().trim()) {
            case "auto", "none", "disabled", "off" -> List.of();

            case "g1gc_basic" -> List.of("-XX:+UseG1GC", "-XX:MaxGCPauseMillis=50");

            case "g1gc_optimized" -> List.of(
                    "-XX:+UseG1GC",
                    "-XX:+UnlockExperimentalVMOptions",
                    "-XX:MaxGCPauseMillis=50",
                    "-XX:G1HeapRegionSize=16m",
                    "-XX:G1NewSizePercent=20",
                    "-XX:G1MaxNewSizePercent=60",
                    "-XX:G1ReservePercent=20",
                    "-XX:G1MixedGCLiveThresholdPercent=85",
                    "-XX:G1OldCSetRegionThresholdPercent=5",
                    "-XX:+ParallelRefProcEnabled");

            case "zgc" -> List.of("-XX:+UseZGC", "-XX:+ZGenerational");

            case "shenandoah" -> List.of("-XX:+UseShenandoahGC", "-XX:ShenandoahGCMode=iu");

            default -> {
                CoreLogger.get().warn(LOG, "Unknown gcPreset '" + gcPreset
                        + "' — no GC flags injected. Valid: auto, none, g1gc_basic, g1gc_optimized, zgc, shenandoah");
                yield List.of();
            }
        };
    }

    private static void applyGpuEnvironment(ProcessBuilder pb, String gpuPreference) {
        if (gpuPreference == null || gpuPreference.isBlank() || "auto".equalsIgnoreCase(gpuPreference))
            return;
        if ("none".equalsIgnoreCase(gpuPreference) || "disabled".equalsIgnoreCase(gpuPreference))
            return;

        String os = System.getProperty("os.name", "").toLowerCase();
        boolean isLinux = !os.contains("win") && !os.contains("mac");
        if (!isLinux)
            return;

        Map<String, String> env = pb.environment();
        switch (gpuPreference.toLowerCase()) {
            case "dgpu" -> {
                env.put("__NV_PRIME_RENDER_OFFLOAD", "1");
                env.put("__VK_LAYER_NV_optimus", "NVIDIA_only");
                env.put("__GLX_VENDOR_LIBRARY_NAME", "nvidia");
                env.put("DRI_PRIME", "1");
            }
            case "igpu" -> {
                env.put("DRI_PRIME", "0");
                env.remove("__NV_PRIME_RENDER_OFFLOAD");
                env.remove("__VK_LAYER_NV_optimus");
                env.remove("__GLX_VENDOR_LIBRARY_NAME");
            }
        }
    }

    private String resolveVanillaVersionId(String versionId, Path instancePath) throws IOException {
        Path versionFile = instancePath.resolve("versions").resolve(versionId).resolve(versionId + ".json");
        if (!Files.exists(versionFile))
            return versionId;
        VersionInfo raw = Json.read(Files.readString(versionFile, StandardCharsets.UTF_8), VersionInfo.class);
        if (raw.inheritsFrom != null && !raw.inheritsFrom.isBlank())
            return resolveVanillaVersionId(raw.inheritsFrom, instancePath);
        return versionId;
    }

    private VersionInfo loadLocalVersionInfo(String versionId, Path instancePath) throws IOException {
        Path versionFile = instancePath.resolve("versions").resolve(versionId).resolve(versionId + ".json");
        if (!Files.exists(versionFile))
            throw new IOException("Version JSON not found: " + versionFile + " — run install first.");
        VersionInfo info = Json.read(Files.readString(versionFile, StandardCharsets.UTF_8), VersionInfo.class);
        if (info.inheritsFrom != null && !info.inheritsFrom.isBlank()) {
            VersionInfo parent = loadLocalVersionInfo(info.inheritsFrom, instancePath);
            info = VersionMerger.merge(parent, info);
        }
        return info;
    }

    private static String detectLoaderType(String versionId, String vanillaVersionId) {
        if (versionId.equals(vanillaVersionId))
            return null;
        String lower = versionId.toLowerCase();
        if (lower.startsWith("legacyfabric-"))
            return "legacyfabric";
        if (lower.startsWith("fabric-"))
            return "fabric";
        if (lower.startsWith("quilt-"))
            return "quilt";
        if (lower.startsWith("neoforge-"))
            return "neoforge";
        if (lower.startsWith("forge-") || lower.contains("-forge-"))
            return "forge";
        if (lower.contains("optifine") || lower.contains("hd_u_"))
            return "optifine";
        return null;
    }

    private String resolveJavaExecutable(LaunchRequest req, Path instancePath, VersionInfo.JavaVersion javaVersion) {
        if (req.javaPath != null && !req.javaPath.isBlank() && !req.javaPath.equals("java"))
            return req.javaPath;
        Path sharedPath = (req.sharedPath != null && !req.sharedPath.isBlank())
                ? Path.of(req.sharedPath).toAbsolutePath()
                : null;
        if (javaVersion != null && javaVersion.majorVersion > 0) {
            String fromRuntime = RuntimeDownloader.findExistingRuntime(instancePath, sharedPath, javaVersion.majorVersion);
            if (fromRuntime != null && Files.isRegularFile(Path.of(fromRuntime)))
                return fromRuntime;
        }
        String fromRuntime = RuntimeDownloader.findExistingRuntime(instancePath, sharedPath);
        if (fromRuntime != null && Files.isRegularFile(Path.of(fromRuntime)))
            return fromRuntime;
        return JavaResolver.resolve(instancePath, sharedPath);
    }
}
