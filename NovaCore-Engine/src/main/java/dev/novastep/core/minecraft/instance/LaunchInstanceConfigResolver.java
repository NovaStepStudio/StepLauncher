package dev.novastep.core.minecraft.instance;

import dev.novastep.core.server.request.LaunchRequest;
import dev.novastep.core.log.CoreLogger;

import java.nio.file.Path;

public final class LaunchInstanceConfigResolver {

    private static final String LOG = "LaunchInstanceConfigResolver";

    private LaunchInstanceConfigResolver() {
    }

    public static void applyFromInstancePath(LaunchRequest req) {
        if (req == null)
            return;
        if (req.instancePath == null || req.instancePath.isBlank())
            return;

        Path instancePath = Path.of(req.resolvedInstancePath()).toAbsolutePath();
        InstanceConfigStore.InstanceConfig cfg = null;

        if (req.configPath != null && !req.configPath.isBlank()) {
            try {
                cfg = InstanceConfigStore.readFromFile(Path.of(req.configPath).toAbsolutePath()).orElse(null);
            } catch (Exception e) {
                CoreLogger.get().error(LOG, "Failed to read explicit config file at " + req.configPath, e);
            }
        }

        if (cfg == null) {
            try {
                cfg = InstanceConfigStore.read(instancePath).orElse(null);
            } catch (Exception e) {
                CoreLogger.get().error(LOG, "Failed to read instance configuration for path=" + instancePath, e);
            }
        }

        InstanceConfigMerger.resolve(req, cfg);
    }
}