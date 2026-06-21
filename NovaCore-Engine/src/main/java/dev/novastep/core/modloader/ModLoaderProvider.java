package dev.novastep.core.modloader;

import dev.novastep.core.minecraft.version.VersionInfo;
import dev.novastep.core.modloader.model.ModLoaderModels.DownloadPlan;
import dev.novastep.core.modloader.model.ModLoaderModels.ExecutionPlan;
import dev.novastep.core.modloader.model.ModLoaderModels.InstalledLoader;
import dev.novastep.core.modloader.model.ModLoaderModels.LoaderVersion;
import dev.novastep.core.websocket.EventBroadcaster;

import java.io.IOException;
import java.nio.file.Path;
import java.util.List;

public interface ModLoaderProvider {

    String name();

    List<LoaderVersion> getVersions(String minecraftVersion) throws IOException, InterruptedException;

    DownloadPlan resolveDownload(
            String minecraftVersion,
            String loaderVersion,
            Path instancePath,
            Path librariesPath
    ) throws IOException, InterruptedException;

    boolean requiresInstallerRun();

    void runInstaller(
            String sessionId,
            InstalledLoader loader,
            Path instancePath,
            Path librariesPath,
            Path minecraftJar,
            EventBroadcaster broadcaster
    ) throws Exception;

    ExecutionPlan buildExecution(
            InstalledLoader loader,
            VersionInfo vanillaInfo,
            Path instancePath,
            Path librariesPath
    );
}
