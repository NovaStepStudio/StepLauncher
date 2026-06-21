package dev.novastep.core.minecraft.instance;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

public final class DefaultInstanceConfig {

    private DefaultInstanceConfig() {
    }

    public static final int MIN_MEMORY_MB = 512;
    public static final int MAX_MEMORY_MB = 2048;
    public static final boolean HARDWARE_ACCEL = false;
    public static final String GC_PRESET = "none";
    public static final String GPU_PREFERENCE = "none";
    public static final String JAVA_PATH = "java";

    public static final int WINDOW_WIDTH = 854;
    public static final int WINDOW_HEIGHT = 480;
    public static final boolean WINDOW_FULLSCREEN = false;

    public static List<String> defaultJvmArgs() {
        return new ArrayList<>();
    }

    public static List<String> defaultPrependArgs() {
        return new ArrayList<>();
    }

    public static List<String> defaultExtraGameArgs() {
        return new ArrayList<>();
    }

    public static Map<String, String> defaultJvmProperties() {
        return new LinkedHashMap<>();
    }

    public static Map<String, Object> defaultCustomFields() {
        return new LinkedHashMap<>();
    }
}
