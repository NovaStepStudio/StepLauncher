package dev.novastep.core.server.request;

import java.nio.file.Path;
import dev.novastep.core.util.SystemResources;

public class InstallRequest {

    public String  version;
    public String  instancePath;
    public String  sharedPath;
    public DownloadOptions download;
    public Boolean isInstance;
    public Boolean verifySHA1;
    public Integer maxThreads;
    public Boolean debug;
    public String  modloader;
    public String  modloaderVersion;
    public LauncherBranding launcher;
    
    public static class LauncherBranding {
        public String name;
        public String version;
    }

    public static class DownloadOptions {
        public Boolean client;
        public Boolean libraries;
        public Boolean assets;
        public Boolean natives;
        public Boolean jvm;
    }

    public String resolvedInstancePath() {
        if (instancePath == null || instancePath.isBlank())
            return Path.of(System.getProperty("user.dir")).resolve("instances").resolve("default").toAbsolutePath().toString();
        return Path.of(instancePath).toAbsolutePath().toString();
    }

    public Path resolvedLibrariesPath() {
        if (sharedPath != null && !sharedPath.isBlank())
            return Path.of(sharedPath).toAbsolutePath().resolve("libraries");
        return Path.of(resolvedInstancePath()).resolve("libraries");
    }

    public Path resolvedAssetsPath() {
        if (sharedPath != null && !sharedPath.isBlank())
            return Path.of(sharedPath).toAbsolutePath().resolve("assets");
        return Path.of(resolvedInstancePath()).resolve("assets");
    }

    public Path resolvedRuntimePath() {
        if (sharedPath != null && !sharedPath.isBlank())
            return Path.of(sharedPath).toAbsolutePath().resolve("runtime");
        return Path.of(resolvedInstancePath()).resolve("runtime");
    }

    public boolean hasSharedPath() { return sharedPath != null && !sharedPath.isBlank(); }
    public boolean shouldDownloadClient() { return download == null || download.client == null || download.client; }
    public boolean shouldDownloadLibraries(){ return download == null || download.libraries == null || download.libraries; }
    public boolean shouldDownloadAssets() { return download == null || download.assets == null || download.assets; }
    public boolean shouldDownloadNatives() { return download == null || download.natives == null || download.natives; }
    public boolean shouldDownloadJvm() { return download != null && Boolean.TRUE.equals(download.jvm); }
    public boolean shouldVerifySHA1() { return verifySHA1 == null || verifySHA1; }
    public boolean isDebug() { return debug != null && debug; }
    public boolean isInstance() { return isInstance != null && isInstance; }
    public int resolvedMaxThreads() { return SystemResources.safeThreads(maxThreads != null ? maxThreads : 0); }

    public String validate() {
        if (version == null || version.isBlank()) return "Field 'version' is required";
        if (instancePath == null || instancePath.isBlank()) return "Field 'instancePath' is required";
        return null;
    }
}
