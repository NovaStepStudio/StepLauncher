package dev.novastep.core.server.request;

import java.nio.file.Path;

public class ModLoaderRequest {
    public String loader;
    public String loaderType;
    public String loaderVersion;
    public String minecraftVersion;
    public String instancePath;
    public String sharedPath;
    public Integer maxThreads;
    public Boolean debug;

    public String resolvedInstancePath() {
        if (instancePath == null || instancePath.isBlank())
            return Path.of(System.getProperty("user.dir")).resolve("instances").resolve("default").toAbsolutePath()
                    .toString();
        return Path.of(instancePath).toAbsolutePath().toString();
    }

    public Path resolvedLibrariesPath() {
        if (sharedPath != null && !sharedPath.isBlank())
            return Path.of(sharedPath).toAbsolutePath().resolve("libraries");
        return Path.of(resolvedInstancePath()).resolve("libraries");
    }

    public Path resolvedMinecraftJar() {
        return Path.of(resolvedInstancePath()).resolve("bin").resolve("minecraft.jar");
    }

    public String validate() {
        if ((loader != null && !loader.isBlank()) && (loaderType == null || loaderType.isBlank())) {
            loaderType = loader;
        }
        if (loader == null || loader.isBlank()) {
            loader = loaderType;
        }
        if (loader == null || loader.isBlank())
            return "'loader' is required";
        if (loaderVersion == null || loaderVersion.isBlank())
            return "'loaderVersion' is required";
        if (minecraftVersion == null || minecraftVersion.isBlank())
            return "'minecraftVersion' is required";
        if (instancePath == null || instancePath.isBlank())
            return "'instancePath' is required";
        return null;
    }
}
