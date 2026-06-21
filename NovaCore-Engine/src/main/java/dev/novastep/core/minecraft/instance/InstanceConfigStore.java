package dev.novastep.core.minecraft.instance;

import dev.novastep.core.json.Json;
import dev.novastep.core.log.CoreLogger;

import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardOpenOption;
import java.time.Instant;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

public final class InstanceConfigStore {

    private static final String LOG = "InstanceConfigStore";
    public static final String FILENAME = "instance.config.json";

    public static final class InstanceConfig {
        public String id;
        public InstanceMetadata instanceMetadata = new InstanceMetadata();
        public ConfigInstance configInstance = new ConfigInstance();
    }

    public static final class InstanceMetadata {
        public String createdAt;
        public String updatedAt;
        public Long totalPlayTimeMs = 0L;
        public FrontendMetadata frontend = new FrontendMetadata();
    }

    public static final class FrontendMetadata {
        public String name = "";
        public String description = "";
        public String icon = "";
        public String hero = "";
    }

    public static final class ConfigInstance {
        public Integer minMemoryMb;
        public Integer maxMemoryMb;
        public Boolean hardwareAccel;
        public String gcPreset;
        public String gpuPreference;
        public String javaPath;
        public WindowConfig window = new WindowConfig();
        public JvmConfig jvm = new JvmConfig();
        public List<String> extraGameArgs = new ArrayList<>();
        public Map<String, Object> customFields = new LinkedHashMap<>();
    }

    public static final class WindowConfig {
        public Boolean fullscreen = false;
        public Integer width = 854;
        public Integer height = 480;
    }

    public static final class JvmConfig {
        public Integer minMemoryMb;
        public Integer maxMemoryMb;
        public List<String> extraArgs = new ArrayList<>();
        public List<String> prependArgs = new ArrayList<>();
        public Map<String, String> jvmProperties = new LinkedHashMap<>();
    }

    private InstanceConfigStore() {
    }

    public static Path file(Path instancePath) {
        return instancePath.toAbsolutePath().resolve(FILENAME);
    }

    public static Optional<InstanceConfig> read(Path instancePath) {
        return readFromFile(file(instancePath));
    }

    public static Optional<InstanceConfig> readFromFile(Path file) {
        if (!Files.exists(file)) {
            return Optional.empty();
        }
        try {
            InstanceConfig config = Json.read(Files.readString(file, StandardCharsets.UTF_8), InstanceConfig.class);
            if (config != null) {
                if (config.instanceMetadata == null) {
                    config.instanceMetadata = new InstanceMetadata();
                }
                if (config.instanceMetadata.frontend == null) {
                    config.instanceMetadata.frontend = new FrontendMetadata();
                }
                if (config.configInstance == null) {
                    config.configInstance = new ConfigInstance();
                }
            }
            return Optional.ofNullable(config);
        } catch (Exception ex) {
            CoreLogger.get().error(LOG, "Failed to read or parse instance configuration file at " + file, ex);
            return Optional.empty();
        }
    }

    public static InstanceConfig readOrCreate(Path instancePath,
                                              InstanceTechnicalMetadataStore.TechnicalMetadata tech,
                                              String mcVersion) {
        Optional<InstanceConfig> existing = read(instancePath);
        if (existing.isPresent()) {
            return existing.get();
        }

        InstanceConfig config = new InstanceConfig();
        config.id = tech != null ? tech.instanceId : null;
        config.instanceMetadata.createdAt = Instant.now().toString();
        config.instanceMetadata.updatedAt = config.instanceMetadata.createdAt;
        config.instanceMetadata.totalPlayTimeMs = 0L;
        config.configInstance.minMemoryMb = DefaultInstanceConfig.MIN_MEMORY_MB;
        config.configInstance.maxMemoryMb = DefaultInstanceConfig.MAX_MEMORY_MB;
        config.configInstance.hardwareAccel = DefaultInstanceConfig.HARDWARE_ACCEL;
        config.configInstance.gcPreset = DefaultInstanceConfig.GC_PRESET;
        config.configInstance.gpuPreference = DefaultInstanceConfig.GPU_PREFERENCE;
        config.configInstance.javaPath = DefaultInstanceConfig.JAVA_PATH;
        config.configInstance.jvm.minMemoryMb = DefaultInstanceConfig.MIN_MEMORY_MB;
        config.configInstance.jvm.maxMemoryMb = DefaultInstanceConfig.MAX_MEMORY_MB;

        save(instancePath, config);
        return config;
    }

    public static void save(Path instancePath, InstanceConfig config) {
        Path file = file(instancePath);
        try {
            Files.createDirectories(instancePath);
            Files.writeString(file, Json.writePretty(config), StandardCharsets.UTF_8,
                    StandardOpenOption.CREATE, StandardOpenOption.TRUNCATE_EXISTING);
        } catch (Exception ex) {
            CoreLogger.get().error(LOG, "Failed to save instance configuration to " + file, ex);
        }
    }
}
