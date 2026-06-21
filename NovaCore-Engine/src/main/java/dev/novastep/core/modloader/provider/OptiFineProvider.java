package dev.novastep.core.modloader.provider;

import dev.novastep.core.json.Json;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.minecraft.version.VersionInfo;
import dev.novastep.core.modloader.ModLoaderProvider;
import dev.novastep.core.modloader.model.ModLoaderModels.DownloadPlan;
import dev.novastep.core.modloader.model.ModLoaderModels.ExecutionPlan;
import dev.novastep.core.modloader.model.ModLoaderModels.InstalledLoader;
import dev.novastep.core.modloader.model.ModLoaderModels.LoaderVersion;
import dev.novastep.core.util.JavaResolver;
import dev.novastep.core.websocket.EventBroadcaster;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardCopyOption;
import java.time.Duration;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public final class OptiFineProvider implements ModLoaderProvider {

    private static final String LOG = "OptiFineProvider";
    private static final String DOWNLOADS_URL = "https://optifine.net/downloads";
    private static final String ADDOWNLOAD = "https://optifine.net/addownload?f=";

    private static final Pattern FILE_PATTERN = Pattern
            .compile("OptiFine_([\\d.]+)_HD_U_([A-Z]\\d+(?:_pre\\d+)?)\\.jar");
    private static final Pattern HREF_PATTERN = Pattern
            .compile("href=['\"]([^'\"]*OptiFine_[^'\"]+\\.jar[^'\"]*)['\"\\s]");


    private static final HttpClient HTTP = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(20))
            .followRedirects(HttpClient.Redirect.ALWAYS)
            .build();

    @Override
    public String name() {
        return "optifine";
    }

    @Override
    public List<LoaderVersion> getVersions(String mcVersion) throws IOException, InterruptedException {
        String html = get(DOWNLOADS_URL);
        return parseVersions(html, mcVersion);
    }

    private List<LoaderVersion> parseVersions(String html, String mcVersion) {
        List<LoaderVersion> result = new ArrayList<>();
        Matcher href = HREF_PATTERN.matcher(html);
        while (href.find()) {
            Matcher fm = FILE_PATTERN.matcher(href.group(1));
            if (!fm.find())
                continue;
            if (!fm.group(1).equals(mcVersion))
                continue;
            String build = fm.group(2);
            result.add(new LoaderVersion("HD_U_" + build, mcVersion, !build.contains("_pre")));
        }
        return result;
    }

    @Override
    public DownloadPlan resolveDownload(
            String mcVersion, String loaderVersion,
            Path instancePath, Path librariesPath) throws IOException, InterruptedException {

        String buildCode = loaderVersion.replace("HD_U_", "");
        String fileName = "OptiFine_" + mcVersion + "_HD_U_" + buildCode + ".jar";
        Path dest = instancePath.resolve("installers").resolve(fileName);

        Files.createDirectories(dest.getParent());

        List<DownloadPlan.Entry> entries = new ArrayList<>();
        entries.add(DownloadPlan.Entry.installer(fileName, ADDOWNLOAD + fileName, dest));
        return DownloadPlan.withInstaller(entries, dest);
    }

    @Override
    public boolean requiresInstallerRun() {
        return true;
    }

    @Override
    public void runInstaller(
            String sessionId,
            InstalledLoader loader,
            Path instancePath,
            Path librariesPath,
            Path minecraftJar,
            EventBroadcaster broadcaster) throws Exception {

        Path installerJar = Path.of(loader.installerJarPath);
        String javaExec = JavaResolver.resolve(instancePath);

        CoreLogger.get().info(LOG, "[" + sessionId + "] Running OptiFine installer: "
                + installerJar.getFileName());
        broadcaster.emit("modloader_install_start", Map.of(
                "sessionId", sessionId,
                "loader", "optifine",
                "version", loader.loaderVersion));

        Path patchedJar = patchOptiFineInstaller(installerJar);

        Path tempDir = Files.createTempDirectory("novacore-optifine-");
        try {
            createLauncherProfiles(tempDir);

            if (Files.exists(minecraftJar)) {
                Path versionDir = tempDir.resolve("versions").resolve(loader.minecraftVersion);
                Files.createDirectories(versionDir);
                Files.copy(minecraftJar,
                        versionDir.resolve(loader.minecraftVersion + ".jar"),
                        StandardCopyOption.REPLACE_EXISTING);
            }

            CoreLogger.get().info(LOG, "[" + sessionId + "] Running patched installer with --mcdir");

            Process process = new ProcessBuilder(
                    javaExec, "-Djava.awt.headless=true",
                    "-jar", patchedJar.toAbsolutePath().toString(),
                    "--mcdir", tempDir.toAbsolutePath().toString())
                    .redirectErrorStream(true)
                    .start();

            try (BufferedReader reader = new BufferedReader(
                    new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    CoreLogger.get().debug(LOG, "[" + sessionId + "][optifine] " + line);
                }
            }

            int exitCode = process.waitFor();
            if (exitCode != 0) {
                throw new RuntimeException("OptiFine installer exited with code " + exitCode);
            }

            integrateResults(sessionId, tempDir, instancePath, librariesPath, loader.minecraftVersion);

        } finally {
            deleteDirectory(tempDir);
            Files.deleteIfExists(patchedJar);
        }

        broadcaster.emit("modloader_install_done", Map.of(
                "sessionId", sessionId,
                "loader", "optifine",
                "versionId", loader.versionJsonId));

        CoreLogger.get().info(LOG, "[" + sessionId + "] OptiFine installed successfully.");
    }

    private Path patchOptiFineInstaller(Path installerJar) throws Exception {
        Path workDir = Files.createTempDirectory("novacore-optifine-patch-");
        Path srcDir = workDir.resolve("src");
        Path binDir = workDir.resolve("bin");
        Files.createDirectories(srcDir);
        Files.createDirectories(binDir);

        Path extractedClass = workDir.resolve("optifine").resolve("Installer.class");
        Files.createDirectories(extractedClass.getParent());

        try (java.util.jar.JarFile jar = new java.util.jar.JarFile(installerJar.toFile())) {
            java.util.jar.JarEntry entry = jar.getJarEntry("optifine/Installer.class");
            if (entry == null)
                throw new IOException("optifine/Installer.class not found in OptiFine JAR");
            try (java.io.InputStream in = jar.getInputStream(entry)) {
                Files.copy(in, extractedClass, StandardCopyOption.REPLACE_EXISTING);
            }
        }

        org.benf.cfr.reader.api.CfrDriver driver = new org.benf.cfr.reader.api.CfrDriver.Builder().withOptions(Map.of(
                "outputdir", srcDir.toAbsolutePath().toString(),
                "silent", "true")).build();
        driver.analyse(java.util.Collections.singletonList(extractedClass.toAbsolutePath().toString()));

        Path installerJava = findFile(srcDir, "Installer.java");
        if (installerJava == null) {
            deleteDirectory(workDir);
            throw new IOException("CFR did not generate Installer.java");
        }

        String source = Files.readString(installerJava, StandardCharsets.UTF_8);

        if (!source.contains("package optifine")) {
            source = "package optifine;\n\n" + source;
        }

        Pattern getWdPattern = Pattern.compile("File\\s+(\\w+)\\s*=\\s*Utils\\.getWorkingDirectory\\s*\\(\\s*\\)\\s*;");
        Matcher matcher = getWdPattern.matcher(source);
        if (!matcher.find()) {
            deleteDirectory(workDir);
            throw new IOException("Could not find Utils.getWorkingDirectory() in Installer.java");
        }

        String varName = matcher.group(1);
        String replacement = "File " + varName + " = null;\n" +
                "        for (int _i = 0; _i < args.length; _i++) {\n" +
                "            if (args[_i].equals(\"--mcdir\") && _i + 1 < args.length) {\n" +
                "                " + varName + " = new File(args[++_i]);\n" +
                "            }\n" +
                "        }\n" +
                "        if (" + varName + " == null) {\n" +
                "            " + varName + " = Utils.getWorkingDirectory();\n" +
                "        }";

        source = matcher.replaceFirst(Matcher.quoteReplacement(replacement));
        Files.writeString(installerJava, source, StandardCharsets.UTF_8);

        javax.tools.JavaCompiler compiler = javax.tools.ToolProvider.getSystemJavaCompiler();
        if (compiler == null) {
            deleteDirectory(workDir);
            throw new IOException("javac not available — JDK required for OptiFine patching");
        }

        int result = compiler.run(null, null, null,
                "-source", "8", "-target", "8",
                "-cp", installerJar.toAbsolutePath().toString(),
                "-d", binDir.toAbsolutePath().toString(),
                installerJava.toAbsolutePath().toString());

        if (result != 0) {
            result = compiler.run(null, null, null,
                    "-cp", installerJar.toAbsolutePath().toString(),
                    "-d", binDir.toAbsolutePath().toString(),
                    installerJava.toAbsolutePath().toString());
        }

        if (result != 0) {
            deleteDirectory(workDir);
            throw new IOException("Failed to compile patched Installer.java (javac exit " + result + ")");
        }

        Path compiledClass = findFile(binDir, "Installer.class");
        if (compiledClass == null) {
            deleteDirectory(workDir);
            throw new IOException("Compiled Installer.class not found after compilation");
        }

        Path patchedJar = installerJar.resolveSibling(
                installerJar.getFileName().toString().replace(".jar", "_patched.jar"));

        try (java.util.zip.ZipInputStream zin = new java.util.zip.ZipInputStream(Files.newInputStream(installerJar));
                java.util.zip.ZipOutputStream zout = new java.util.zip.ZipOutputStream(
                        Files.newOutputStream(patchedJar))) {

            java.util.zip.ZipEntry entry;
            boolean replacedClass = false;

            while ((entry = zin.getNextEntry()) != null) {
                String name = entry.getName();
                zout.putNextEntry(new java.util.zip.ZipEntry(name));

                if (name.equals("optifine/Installer.class")) {
                    Files.copy(compiledClass, zout);
                    replacedClass = true;
                } else if (name.equals("META-INF/MANIFEST.MF")) {
                    String manifest = new String(zin.readAllBytes(), StandardCharsets.UTF_8);
                    manifest = manifest.replaceAll("(?m)^Main-Class:.*$", "Main-Class: optifine.Installer");
                    zout.write(manifest.getBytes(StandardCharsets.UTF_8));
                } else {
                    zin.transferTo(zout);
                }
                zout.closeEntry();
                zin.closeEntry();
            }

            if (!replacedClass) {
                zout.putNextEntry(new java.util.zip.ZipEntry("optifine/Installer.class"));
                Files.copy(compiledClass, zout);
                zout.closeEntry();
            }
        }

        deleteDirectory(workDir);
        return patchedJar;
    }

    private static Path findFile(Path root, String filename) throws IOException {
        if (!Files.isDirectory(root))
            return null;
        try (var stream = Files.walk(root)) {
            return stream.filter(p -> p.getFileName().toString().equals(filename)).findFirst().orElse(null);
        }
    }

    private void integrateResults(
            String sessionId,
            Path tempDir,
            Path instancePath,
            Path librariesPath,
            String mcVersion) throws IOException {

        Path tempVersions = tempDir.resolve("versions");
        if (Files.isDirectory(tempVersions)) {
            try (var stream = Files.list(tempVersions)) {
                stream.filter(Files::isDirectory)
                        .filter(p -> !p.getFileName().toString().equals(mcVersion))
                        .forEach(vDir -> {
                            try {
                                Path dest = instancePath.resolve("versions").resolve(vDir.getFileName());
                                copyTree(vDir, dest);
                                CoreLogger.get().info(LOG, "[" + sessionId + "] Integrated version: "
                                        + vDir.getFileName());
                            } catch (IOException e) {
                                CoreLogger.get().warn(LOG, "[" + sessionId + "] Failed copying version "
                                        + vDir.getFileName() + ": " + e.getMessage());
                            }
                        });
            }
        }

        Path tempLibs = tempDir.resolve("libraries");
        if (Files.isDirectory(tempLibs)) {
            copyTree(tempLibs, librariesPath);
            CoreLogger.get().info(LOG, "[" + sessionId + "] Integrated OptiFine libraries.");
        }
    }

    private static void copyTree(Path src, Path dest) throws IOException {
        Files.createDirectories(dest);
        try (var stream = Files.walk(src)) {
            for (Path p : (Iterable<Path>) stream::iterator) {
                Path target = dest.resolve(src.relativize(p));
                if (Files.isDirectory(p)) {
                    Files.createDirectories(target);
                } else {
                    Files.createDirectories(target.getParent());
                    Files.copy(p, target, StandardCopyOption.REPLACE_EXISTING);
                }
            }
        }
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
        root.put("settings", Map.of("enableSnapshots", false));
        root.put("version", 3);

        Files.writeString(dir.resolve("launcher_profiles.json"), Json.write(root), StandardCharsets.UTF_8);
    }

    private static void deleteDirectory(Path dir) {
        if (!Files.exists(dir))
            return;
        try {
            Files.walk(dir).sorted(Comparator.reverseOrder()).map(Path::toFile).forEach(java.io.File::delete);
        } catch (IOException ignored) {
        }
    }

    @Override
    public ExecutionPlan buildExecution(
            InstalledLoader loader, VersionInfo vanillaInfo,
            Path instancePath, Path librariesPath) {
        return null;
    }

    private String get(String url) throws IOException, InterruptedException {
        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(url))
                .header("User-Agent", "Mozilla/5.0 NovaCore/" + dev.novastep.core.CoreVersion.get())
                .GET()
                .timeout(Duration.ofSeconds(30))
                .build();
        HttpResponse<String> res = HTTP.send(req, HttpResponse.BodyHandlers.ofString(StandardCharsets.UTF_8));
        if (res.statusCode() != 200) {
            throw new IOException("HTTP " + res.statusCode() + " from: " + url);
        }
        return res.body();
    }
}
