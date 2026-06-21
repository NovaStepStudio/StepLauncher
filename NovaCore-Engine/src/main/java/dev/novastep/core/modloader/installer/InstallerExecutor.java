package dev.novastep.core.modloader.installer;

import dev.novastep.core.json.Json;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.modloader.model.InstallProfile;
import dev.novastep.core.websocket.EventBroadcaster;

import java.io.*;
import java.net.URI;
import java.net.http.*;
import java.nio.charset.StandardCharsets;
import java.nio.file.*;
import java.time.Duration;
import java.util.*;
import java.util.jar.*;
import java.util.zip.ZipEntry;

public final class InstallerExecutor {

    private static final String LOG = "InstallerExecutor";
    // JSON handled via Jackson (Json.java) — Gson removed

    private static final HttpClient HTTP = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(20))
            .followRedirects(HttpClient.Redirect.ALWAYS)
            .build();

    private static final List<String> FALLBACK_REPOS = List.of(
            "https://maven.minecraftforge.net/",
            "https://maven.neoforged.net/releases/",
            "https://repo1.maven.org/maven2/",
            "https://libraries.minecraft.net/");

    private InstallerExecutor() {
    }

    public static void execute(
            String sessionId,
            Path installerJar,
            Path librariesPath,
            Path minecraftJar,
            Path instancePath,
            String javaExecutable,
            String mavenRepoBase,
            EventBroadcaster broadcaster) throws Exception {

        CoreLogger.get().info(LOG, "[" + sessionId + "] Starting installer pipeline: " + installerJar);

        Path tempDir = Files.createTempDirectory("novacore-installer-");
        try {
            createLauncherProfiles(tempDir);

            InstallProfile profile = extractAndPrepare(installerJar, librariesPath, tempDir, sessionId);
            if (profile == null) {
                throw new IllegalStateException("install_profile.json not found in installer: " + installerJar);
            }

            if (profile.libraries != null && !profile.libraries.isEmpty()) {
                downloadProfileLibraries(sessionId, profile, librariesPath);
            }

            runProcessors(sessionId, profile, installerJar, librariesPath, minecraftJar,
                    instancePath, tempDir, javaExecutable, mavenRepoBase, broadcaster);

        } finally {
            deleteDirectory(tempDir);
        }

        CoreLogger.get().info(LOG, "[" + sessionId + "] Installer pipeline complete.");
    }

    private static void createLauncherProfiles(Path dir) throws IOException {
        Map<String, Object> profile = new LinkedHashMap<>();
        profile.put("name", "NovaCore");
        profile.put("type", "custom");
        profile.put("created", "1970-01-01T00:00:00.000Z");
        profile.put("lastUsed", "1970-01-01T00:00:00.000Z");
        profile.put("lastVersionId", "latest-release");

        Map<String, Object> root = new LinkedHashMap<>();
        root.put("selectedProfile", "NovaCore");
        root.put("profiles", Map.of("NovaCore", profile));
        root.put("clientToken", UUID.randomUUID().toString().replace("-", ""));
        root.put("authenticationDatabase", Map.of());
        root.put("settings",
                Map.of("enableSnapshots", false, "enableAdvanced", false, "profileSorting", "ByLastPlayed"));
        root.put("version", 3);

        Files.writeString(dir.resolve("launcher_profiles.json"), Json.write(root), StandardCharsets.UTF_8);
    }

    private static InstallProfile extractAndPrepare(
            Path installerJar, Path librariesPath, Path tempDir, String sessionId) throws IOException {

        InstallProfile profile = null;

        try (JarFile jar = new JarFile(installerJar.toFile())) {
            Enumeration<? extends ZipEntry> entries = jar.entries();
            while (entries.hasMoreElements()) {
                ZipEntry entry = entries.nextElement();
                String name = entry.getName();

                if ("install_profile.json".equals(name)) {
                    try (InputStream in = jar.getInputStream(entry)) {
                        profile = Json.read(new String(in.readAllBytes(), StandardCharsets.UTF_8), InstallProfile.class);
                    }
                } else if (name.startsWith("maven/") && !entry.isDirectory()) {
                    String relative = name.substring("maven/".length());
                    Path dest = librariesPath.resolve(relative);
                    if (!Files.exists(dest)) {
                        Files.createDirectories(dest.getParent());
                        try (InputStream in = jar.getInputStream(entry)) {
                            Files.copy(in, dest, StandardCopyOption.REPLACE_EXISTING);
                        }
                        CoreLogger.get().debug(LOG, "[" + sessionId + "] Extracted bundled lib: " + relative);
                    }
                } else if (!entry.isDirectory()) {
                    Path dest = tempDir.resolve(name);
                    Files.createDirectories(dest.getParent());
                    try (InputStream in = jar.getInputStream(entry)) {
                        Files.copy(in, dest, StandardCopyOption.REPLACE_EXISTING);
                    }
                }
            }
        }

        return profile;
    }

    private static void downloadProfileLibraries(
            String sessionId, InstallProfile profile, Path librariesPath) {

        for (InstallProfile.Library lib : profile.libraries) {
            if (lib.downloads == null || lib.downloads.artifact == null)
                continue;
            InstallProfile.Library.LibDownload.Artifact art = lib.downloads.artifact;
            if (art.url == null || art.url.isBlank())
                continue;
            if (art.path == null || art.path.isBlank())
                continue;

            Path dest = librariesPath.resolve(art.path);
            if (Files.exists(dest))
                continue;

            try {
                Files.createDirectories(dest.getParent());
                HttpRequest req = HttpRequest.newBuilder()
                        .uri(URI.create(art.url))
                        .GET()
                        .timeout(Duration.ofSeconds(60))
                        .build();
                HttpResponse<InputStream> res = HTTP.send(req, HttpResponse.BodyHandlers.ofInputStream());
                if (res.statusCode() == 200) {
                    Files.copy(res.body(), dest, StandardCopyOption.REPLACE_EXISTING);
                    CoreLogger.get().debug(LOG, "[" + sessionId + "] Downloaded profile lib: " + art.path);
                } else {
                    CoreLogger.get().warn(LOG, "[" + sessionId + "] HTTP " + res.statusCode()
                            + " for profile lib: " + art.url);
                }
                res.body().close();
            } catch (Exception e) {
                CoreLogger.get().error(LOG, "[" + sessionId + "] Critical failure downloading profile library: "
                        + art.path + " from " + art.url, e);
            }
        }
    }

    private static void runProcessors(
            String sessionId,
            InstallProfile profile,
            Path installerJar,
            Path librariesPath,
            Path minecraftJar,
            Path instancePath,
            Path tempDir,
            String javaExecutable,
            String mavenRepoBase,
            EventBroadcaster broadcaster) throws Exception {

        if (profile.processors == null || profile.processors.isEmpty())
            return;

        Map<String, String> vars = buildVariables(profile, installerJar, librariesPath, minecraftJar, instancePath,
                tempDir);

        List<InstallProfile.Processor> clientProcessors = profile.processors.stream()
                .filter(InstallProfile.Processor::isClientSide)
                .toList();

        int total = clientProcessors.size();
        int index = 0;

        for (InstallProfile.Processor processor : clientProcessors) {
            index++;
            CoreLogger.get().info(LOG, "[" + sessionId + "] Processor " + index + "/" + total + ": " + processor.jar);
            broadcaster.emit("modloader_processor", Map.of(
                    "sessionId", sessionId,
                    "step", index,
                    "total", total,
                    "jar", processor.jar));

            if (allOutputsExist(processor, vars, librariesPath)) {
                CoreLogger.get().debug(LOG, "[" + sessionId + "] Outputs already exist, skipping: " + processor.jar);
                continue;
            }

            runSingleProcessor(sessionId, processor, librariesPath, vars, javaExecutable, mavenRepoBase, broadcaster);
        }
    }

    private static boolean allOutputsExist(
            InstallProfile.Processor processor,
            Map<String, String> vars,
            Path librariesPath) {

        if (processor.outputs == null || processor.outputs.isEmpty())
            return false;

        for (String rawKey : processor.outputs.keySet()) {
            String key = rawKey.trim();
            Path target;

            if (key.startsWith("[") && key.endsWith("]")) {
                try {
                    target = MavenCoordinate.parse(key).toLocalPath(librariesPath);
                } catch (IllegalArgumentException e) {
                    return false;
                }
            } else {
                String resolved = substituteVars(key, vars);
                if (resolved.equals(key))
                    return false;
                target = Path.of(resolved);
            }

            if (!Files.exists(target))
                return false;
        }
        return true;
    }

    private static void runSingleProcessor(
            String sessionId,
            InstallProfile.Processor processor,
            Path librariesPath,
            Map<String, String> vars,
            String javaExecutable,
            String mavenRepoBase,
            EventBroadcaster broadcaster) throws Exception {

        MavenCoordinate coord = MavenCoordinate.parse(processor.jar);
        Path processorJar = coord.toLocalPath(librariesPath);

        if (!Files.exists(processorJar)) {
            processorJar = downloadJar(sessionId, coord, librariesPath, mavenRepoBase);
        }

        String mainClass = readMainClass(processorJar);
        if (mainClass == null) {
            throw new IllegalStateException("No Main-Class manifest attribute in: " + processorJar);
        }

        List<String> classpath = new ArrayList<>();
        classpath.add(processorJar.toAbsolutePath().toString());

        if (processor.classpath != null) {
            for (String dep : processor.classpath) {
                MavenCoordinate depCoord = MavenCoordinate.parse(dep);
                Path depJar = depCoord.toLocalPath(librariesPath);
                if (!Files.exists(depJar)) {
                    try {
                        depJar = downloadJar(sessionId, depCoord, librariesPath, mavenRepoBase);
                    } catch (IOException e) {
                        CoreLogger.get().error(LOG,
                                "[" + sessionId + "] Failed to download required processor dependency: " + depJar, e);
                        continue;
                    }
                }
                classpath.add(depJar.toAbsolutePath().toString());
            }
        }

        List<String> args = new ArrayList<>();
        if (processor.args != null) {
            for (String raw : processor.args) {
                args.add(resolveArg(raw, vars, librariesPath));
            }
        }

        List<String> command = new ArrayList<>();
        command.add(javaExecutable);
        command.add("-cp");
        command.add(String.join(File.pathSeparator, classpath));
        command.add(mainClass);
        command.addAll(args);

        CoreLogger.get().debug(LOG, "[" + sessionId + "] Executing: " + mainClass);

        Process process = new ProcessBuilder(command)
                .redirectErrorStream(true)
                .start();

        try (BufferedReader reader = new BufferedReader(
                new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
            String line;
            while ((line = reader.readLine()) != null) {
                CoreLogger.get().debug(LOG, "[" + sessionId + "][proc] " + line);

                if (line.contains("%") || line.contains("/") || line.length() < 120) {
                    broadcaster.emit("modloader_processor_log", Map.of(
                            "sessionId", sessionId,
                            "line", line.trim(),
                            "processor", mainClass));
                }
            }
        }

        int exitCode = process.waitFor();
        if (exitCode != 0) {
            throw new RuntimeException("Processor exited with code " + exitCode + ": " + processor.jar);
        }
    }

    @SuppressWarnings("unused")
    private static Path downloadJar(
            String sessionId, MavenCoordinate coord, Path librariesPath, String primaryRepo)
            throws IOException, InterruptedException {

        Path dest = coord.toLocalPath(librariesPath);
        if (Files.exists(dest))
            return dest;

        Files.createDirectories(dest.getParent());

        List<String> repos = new ArrayList<>();
        if (primaryRepo != null && !primaryRepo.isBlank())
            repos.add(primaryRepo);
        repos.addAll(FALLBACK_REPOS);

        for (String repo : repos) {
            String base = repo.endsWith("/") ? repo : repo + "/";
            String url = base + coord.toPath();
            try {
                CoreLogger.get().info(LOG, "[" + sessionId + "] Fetching: " + url);
                HttpRequest req = HttpRequest.newBuilder()
                        .uri(URI.create(url))
                        .GET()
                        .timeout(Duration.ofSeconds(90))
                        .build();
                HttpResponse<InputStream> res = HTTP.send(req, HttpResponse.BodyHandlers.ofInputStream());
                if (res.statusCode() != 200)
                    continue;

                Files.copy(res.body(), dest, StandardCopyOption.REPLACE_EXISTING);
                res.body().close();

                try (JarFile jar = new JarFile(dest.toFile())) {
                    CoreLogger.get().info(LOG, "[" + sessionId + "] Downloaded: " + dest.getFileName());
                    return dest;
                } catch (IOException e) {
                    Files.deleteIfExists(dest);
                }
            } catch (Exception e) {
                CoreLogger.get().error(LOG,
                        "[" + sessionId + "] Failed to download JAR from repo=" + repo + " for coordinate=" + coord, e);
            }
        }

        throw new IOException("Could not download: " + coord);
    }

    private static String readMainClass(Path jar) throws IOException {
        try (JarFile jf = new JarFile(jar.toFile())) {
            Manifest mf = jf.getManifest();
            if (mf == null)
                return null;
            return mf.getMainAttributes().getValue("Main-Class");
        }
    }

    private static Map<String, String> buildVariables(
            InstallProfile profile,
            Path installerJar,
            Path librariesPath,
            Path minecraftJar,
            Path instancePath,
            Path tempDir) {

        Map<String, String> vars = new LinkedHashMap<>();
        vars.put("{MINECRAFT_JAR}", minecraftJar.toAbsolutePath().toString());
        vars.put("{INSTALLER}", installerJar.toAbsolutePath().toString());
        vars.put("{ROOT}", instancePath.toAbsolutePath().toString());
        vars.put("{LIBRARY_DIR}", librariesPath.toAbsolutePath().toString());
        vars.put("{TEMP}", tempDir.toAbsolutePath().toString());
        vars.put("{SIDE}", "client");

        if (profile.data != null) {
            for (Map.Entry<String, InstallProfile.DataVal> e : profile.data.entrySet()) {
                if (e.getValue() == null)
                    continue;
                String rawValue = e.getValue().client;
                if (rawValue == null || rawValue.isBlank())
                    continue;

                String key = "{" + e.getKey() + "}";
                if (rawValue.startsWith("[") && rawValue.endsWith("]")) {
                    try {
                        vars.put(key, MavenCoordinate.parse(rawValue)
                                .toLocalPath(librariesPath).toAbsolutePath().toString());
                    } catch (IllegalArgumentException ex) {
                        vars.put(key, rawValue);
                    }
                } else if (rawValue.startsWith("/")) {
                    String relative = rawValue.substring(1).replace("/", java.io.File.separator);
                    vars.put(key, tempDir.toAbsolutePath().resolve(relative).toString());
                } else {
                    vars.put(key, rawValue);
                }
            }
        }

        return vars;
    }

    private static String substituteVars(String token, Map<String, String> vars) {
        String result = token;
        for (Map.Entry<String, String> e : vars.entrySet()) {
            result = result.replace(e.getKey(), e.getValue());
        }
        return result;
    }

    private static String resolveArg(String arg, Map<String, String> vars, Path librariesPath) {
        String trimmed = arg.trim();
        if (trimmed.startsWith("[") && trimmed.endsWith("]")) {
            try {
                return MavenCoordinate.parse(trimmed)
                        .toLocalPath(librariesPath).toAbsolutePath().toString();
            } catch (IllegalArgumentException e) {
                return arg;
            }
        }
        return substituteVars(arg, vars);
    }

    private static void deleteDirectory(Path dir) {
        if (!Files.exists(dir))
            return;
        try (var stream = Files.walk(dir)) {
            stream.sorted(Comparator.reverseOrder())
                    .forEach(p -> {
                        try {
                            Files.delete(p);
                        } catch (IOException e) {
                            CoreLogger.get().warn(LOG, "Could not delete file/folder: " + p, e);
                        }
                    });
        } catch (IOException ex) {
            CoreLogger.get().error(LOG, "Failed to walk directory for deletion: " + dir, ex);
        }
    }
}