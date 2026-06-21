package dev.novastep.core.minecraft.instance;

import dev.novastep.core.json.Json;
import dev.novastep.core.log.CoreLogger;

import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

public final class LegacyInstanceMetadataMigrator {

    private static final String LOG = "LegacyInstanceMetadataMigrator";
    private static final String LEGACY_BIN = "novacore_engine.bin";
    private static final String DEFAULT_BRAND_FILE = "third_party_launcher.json";

    private static final class LegacyMetadata {
        public String instanceId;
        public String instancePath;
        public String createdAt;
        public String lastVerifiedAt;
        public String nextVerifyAt;
        public List<LegacyInstalledEntry> installedVersions;

        private static final class LegacyInstalledEntry {
            public String installedAt;
            public String mcVersion;
        }
    }

    private LegacyInstanceMetadataMigrator() {
    }

    public static Optional<InstanceTechnicalMetadataStore.TechnicalMetadata> migrateIfPresent(
            Path instancePath,
            String launcherBrandName) {

        Path target = InstanceTechnicalMetadataStore.file(instancePath);
        if (Files.exists(target)) {
            return InstanceTechnicalMetadataStore.read(instancePath);
        }

        List<Path> candidates = new ArrayList<>();
        candidates.add(instancePath.resolve(LEGACY_BIN));
        if (launcherBrandName != null && !launcherBrandName.isBlank()) {
            String branded = launcherBrandName.toLowerCase().replace(" ", "_").replace("-", "_") + ".json";
            candidates.add(instancePath.resolve(branded));
        }
        candidates.add(instancePath.resolve(DEFAULT_BRAND_FILE));

        for (Path candidate : candidates) {
            if (!Files.exists(candidate)) {
                continue;
            }
            Optional<InstanceTechnicalMetadataStore.TechnicalMetadata> migrated = tryMigrateFrom(instancePath, candidate);
            if (migrated.isPresent()) {
                return migrated;
            }
        }
        return Optional.empty();
    }

    private static Optional<InstanceTechnicalMetadataStore.TechnicalMetadata> tryMigrateFrom(
            Path instancePath,
            Path legacyFile) {
        try {
            LegacyMetadata legacy = Json.read(Files.readString(legacyFile, StandardCharsets.UTF_8), LegacyMetadata.class);
            if (legacy == null || legacy.instanceId == null || legacy.instanceId.isBlank()) {
                return Optional.empty();
            }

            InstanceTechnicalMetadataStore.TechnicalMetadata migrated = new InstanceTechnicalMetadataStore.TechnicalMetadata();
            migrated.instanceId = legacy.instanceId;
            migrated.instancePath = legacy.instancePath != null ? legacy.instancePath : instancePath.toAbsolutePath().toString();
            migrated.createdAt = legacy.createdAt;
            migrated.lastVerifiedAt = legacy.lastVerifiedAt;
            migrated.nextVerifyAt = legacy.nextVerifyAt;
            migrated.installedVersions = new ArrayList<>();

            if (legacy.installedVersions != null) {
                for (var entry : legacy.installedVersions) {
                    InstanceTechnicalMetadataStore.TechnicalMetadata.InstalledVersion version =
                            new InstanceTechnicalMetadataStore.TechnicalMetadata.InstalledVersion();
                    version.installedAt = entry.installedAt;
                    version.mcVersion = entry.mcVersion;
                    migrated.installedVersions.add(version);
                }
            }

            InstanceTechnicalMetadataStore.save(instancePath, migrated);
            CoreLogger.get().info(LOG, "Migrated legacy metadata " + legacyFile.getFileName() + " -> " + InstanceTechnicalMetadataStore.FILENAME);
            return Optional.of(migrated);
        } catch (Exception ex) {
            CoreLogger.get().warn(LOG, "Legacy metadata migration failed at " + legacyFile + ": " + ex.getMessage());
            return Optional.empty();
        }
    }
}
