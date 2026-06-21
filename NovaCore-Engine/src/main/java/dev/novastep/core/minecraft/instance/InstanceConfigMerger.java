package dev.novastep.core.minecraft.instance;

import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.server.request.LaunchRequest;
import dev.novastep.core.util.SystemResources;

import java.util.ArrayList;
import java.util.LinkedHashMap;

public final class InstanceConfigMerger {

    private static final String LOG = "InstanceConfigMerger";

    private InstanceConfigMerger() {
    }

    public static LaunchRequest resolve(LaunchRequest req, InstanceConfigStore.InstanceConfig cfg) {
        if (req == null)
            req = new LaunchRequest();
        if (req.jvm == null)
            req.jvm = new LaunchRequest.JvmConfig();
        if (req.window == null)
            req.window = new LaunchRequest.WindowConfig();
        if (req.game == null)
            req.game = new LaunchRequest.GameCustomization();

        InstanceConfigStore.ConfigInstance ci = (cfg != null) ? cfg.configInstance
                : new InstanceConfigStore.ConfigInstance();

        resolveMemory(req, ci);
        resolveJvmConfig(req, ci);
        resolveHardwareAndGpu(req, ci);
        resolveWindow(req, ci);
        resolveGameCustomization(req, ci);
        validate(req);

        return req;
    }

    private static void resolveMemory(LaunchRequest req, InstanceConfigStore.ConfigInstance ci) {
        req.jvm.minMemoryMb = firstValidInt(
                req.jvm.minMemoryMb,
                (ci.jvm != null) ? ci.jvm.minMemoryMb : null,
                ci.minMemoryMb,
                DefaultInstanceConfig.MIN_MEMORY_MB);

        req.jvm.maxMemoryMb = firstValidInt(
                req.jvm.maxMemoryMb,
                (ci.jvm != null) ? ci.jvm.maxMemoryMb : null,
                ci.maxMemoryMb,
                DefaultInstanceConfig.MAX_MEMORY_MB);
    }

    private static void resolveJvmConfig(LaunchRequest req, InstanceConfigStore.ConfigInstance ci) {
        if (isMissing(req.jvm.extraArgs)) {
            req.jvm.extraArgs = (ci.jvm != null && !isMissing(ci.jvm.extraArgs))
                    ? new ArrayList<>(ci.jvm.extraArgs)
                    : DefaultInstanceConfig.defaultJvmArgs();
        }

        if (isMissing(req.jvm.prependArgs)) {
            req.jvm.prependArgs = (ci.jvm != null && !isMissing(ci.jvm.prependArgs))
                    ? new ArrayList<>(ci.jvm.prependArgs)
                    : DefaultInstanceConfig.defaultPrependArgs();
        }

        req.gcPreset = firstValidString(req.gcPreset, ci.gcPreset, DefaultInstanceConfig.GC_PRESET);
    }

    private static void resolveHardwareAndGpu(LaunchRequest req, InstanceConfigStore.ConfigInstance ci) {
        if (req.hardwareAcceleration == null) {
            req.hardwareAcceleration = (ci.hardwareAccel != null) ? ci.hardwareAccel
                    : DefaultInstanceConfig.HARDWARE_ACCEL;
        }

        req.gpuPreference = firstValidString(req.gpuPreference, ci.gpuPreference, DefaultInstanceConfig.GPU_PREFERENCE);
    }

    private static void resolveWindow(LaunchRequest req, InstanceConfigStore.ConfigInstance ci) {
        req.window.width = firstValidInt(
                req.window.width,
                (ci.window != null) ? ci.window.width : null,
                DefaultInstanceConfig.WINDOW_WIDTH);

        req.window.height = firstValidInt(
                req.window.height,
                (ci.window != null) ? ci.window.height : null,
                DefaultInstanceConfig.WINDOW_HEIGHT);

        if (req.window.fullscreen == null) {
            req.window.fullscreen = (ci.window != null && ci.window.fullscreen != null)
                    ? ci.window.fullscreen
                    : DefaultInstanceConfig.WINDOW_FULLSCREEN;
        }
    }

    private static void resolveGameCustomization(LaunchRequest req, InstanceConfigStore.ConfigInstance ci) {
        if (isMissing(req.game.extraGameArgs)) {
            req.game.extraGameArgs = (!isMissing(ci.extraGameArgs))
                    ? new ArrayList<>(ci.extraGameArgs)
                    : DefaultInstanceConfig.defaultExtraGameArgs();
        }

        if (isMissing(req.game.extraJvmProperties)) {
            req.game.extraJvmProperties = (ci.jvm != null && ci.jvm.jvmProperties != null
                    && !ci.jvm.jvmProperties.isEmpty())
                            ? new LinkedHashMap<>(ci.jvm.jvmProperties)
                            : DefaultInstanceConfig.defaultJvmProperties();
        }
    }

    private static void validate(LaunchRequest req) {
        if (req.jvm.minMemoryMb > req.jvm.maxMemoryMb) {
            CoreLogger.get().warn(LOG, "Invalid memory config: minMb (" + req.jvm.minMemoryMb + ") > maxMb ("
                    + req.jvm.maxMemoryMb + "). Swapping.");
            int temp = req.jvm.minMemoryMb;
            req.jvm.minMemoryMb = req.jvm.maxMemoryMb;
            req.jvm.maxMemoryMb = temp;
        }

        int[] safe = SystemResources.safeRam(req.jvm.minMemoryMb, req.jvm.maxMemoryMb);
        if (safe[0] != req.jvm.minMemoryMb || safe[1] != req.jvm.maxMemoryMb) {
            CoreLogger.get().info(LOG, "Adjusting memory to system safe bounds: " + req.jvm.minMemoryMb + "/"
                    + req.jvm.maxMemoryMb + " -> " + safe[0] + "/" + safe[1]);
            req.jvm.minMemoryMb = safe[0];
            req.jvm.maxMemoryMb = safe[1];
        }

        if (req.window.width <= 0)
            req.window.width = DefaultInstanceConfig.WINDOW_WIDTH;
        if (req.window.height <= 0)
            req.window.height = DefaultInstanceConfig.WINDOW_HEIGHT;
        if (req.jvm.extraArgs != null)
            req.jvm.extraArgs.removeIf(s -> s == null || s.isBlank());
        if (req.jvm.prependArgs != null)
            req.jvm.prependArgs.removeIf(s -> s == null || s.isBlank());
        if (req.game.extraGameArgs != null)
            req.game.extraGameArgs.removeIf(s -> s == null || s.isBlank());
    }

    private static String firstValidString(String... values) {
        for (String v : values) {
            if (v != null && !v.isBlank())
                return v;
        }
        return null;
    }

    private static Integer firstValidInt(Integer... values) {
        for (Integer v : values) {
            if (v != null && v > 0)
                return v;
        }
        return null;
    }

    private static boolean isMissing(Object obj) {
        if (obj == null)
            return true;
        if (obj instanceof java.util.Collection)
            return ((java.util.Collection<?>) obj).isEmpty();
        if (obj instanceof java.util.Map)
            return ((java.util.Map<?, ?>) obj).isEmpty();
        return false;
    }
}
