package dev.novastep.core.minecraft.instance;

import dev.novastep.core.json.Json;
import dev.novastep.core.log.CoreLogger;

import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardOpenOption;
import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

public final class InstanceTechnicalMetadataStore {

    private static final String LOG = "InstanceTechnicalMetadataStore";
    public static final String FILENAME = "instance.metadata.json";
    private static final int VERIFY_DAYS = 7;

    public static final class TechnicalMetadata {
        public String instanceId;
        public String instancePath;
        public String createdAt;
        public String lastVerifiedAt;
        public String nextVerifyAt;
        public List<InstalledVersion> installedVersions = new ArrayList<>();

        public static final class InstalledVersion {
            public String installedAt;
            public String mcVersion;
            public Boolean lastPlayed;
            public String lastPlayedAt;
        }
    }

    private InstanceTechnicalMetadataStore() {
    }

    public static Path file(Path instancePath) {
        return instancePath.toAbsolutePath().resolve(FILENAME);
    }

    public static Optional<TechnicalMetadata> read(Path instancePath) {
        Path file = file(instancePath);
        if (!Files.exists(file)) {
            return Optional.empty();
        }
        try {
            return Optional.ofNullable(Json.read(Files.readString(file, StandardCharsets.UTF_8), TechnicalMetadata.class));
        } catch (Exception ex) {
            CoreLogger.get().warn(LOG, "Failed to read " + file + ": " + ex.getMessage());
            return Optional.empty();
        }
    }

    public static TechnicalMetadata readOrCreate(Path instancePath, String mcVersion) {
        return read(instancePath).orElseGet(() -> create(instancePath, mcVersion));
    }

    public static TechnicalMetadata create(Path instancePath, String mcVersion) {
        TechnicalMetadata metadata = new TechnicalMetadata();
        metadata.instanceId = UUID.randomUUID().toString();
        metadata.instancePath = instancePath.toAbsolutePath().toString();
        metadata.createdAt = Instant.now().toString();
        metadata.lastVerifiedAt = metadata.createdAt;
        metadata.nextVerifyAt = Instant.now().plus(VERIFY_DAYS, ChronoUnit.DAYS).toString();
        save(instancePath, metadata);
        return metadata;
    }

    public static void recordInstall(Path instancePath, TechnicalMetadata metadata, String mcVersion) {
        if (metadata.installedVersions == null) {
            metadata.installedVersions = new ArrayList<>();
        }
        if (!metadata.installedVersions.isEmpty()) {
            TechnicalMetadata.InstalledVersion last = metadata.installedVersions.get(metadata.installedVersions.size() - 1);
            if (last != null && mcVersion != null && mcVersion.equals(last.mcVersion)) {
                save(instancePath, metadata);
                return;
            }
        }
        TechnicalMetadata.InstalledVersion entry = new TechnicalMetadata.InstalledVersion();
        entry.installedAt = Instant.now().toString();
        entry.mcVersion = mcVersion;
        entry.lastPlayed = false;
        metadata.installedVersions.add(entry);
        save(instancePath, metadata);
    }

    public static void recordVerification(Path instancePath, TechnicalMetadata metadata) {
        metadata.lastVerifiedAt = Instant.now().toString();
        metadata.nextVerifyAt = Instant.now().plus(VERIFY_DAYS, ChronoUnit.DAYS).toString();
        save(instancePath, metadata);
    }

    public static void markLastPlayed(Path instancePath, String mcVersion) {
        Optional<TechnicalMetadata> optional = read(instancePath);
        if (optional.isEmpty()) {
            return;
        }
        TechnicalMetadata metadata = optional.get();
        if (metadata.installedVersions == null || metadata.installedVersions.isEmpty()) {
            return;
        }
        String now = Instant.now().toString();
        for (var entry : metadata.installedVersions) {
            if (entry != null) {
                entry.lastPlayed = false;
            }
        }
        for (int i = metadata.installedVersions.size() - 1; i >= 0; i--) {
            var entry = metadata.installedVersions.get(i);
            if (entry != null && mcVersion != null && mcVersion.equals(entry.mcVersion)) {
                entry.lastPlayed = true;
                entry.lastPlayedAt = now;
                break;
            }
        }
        save(instancePath, metadata);
    }

    public static void save(Path instancePath, TechnicalMetadata metadata) {
        Path file = file(instancePath);
        try {
            Files.createDirectories(instancePath);
            Files.writeString(file, Json.writePretty(metadata), StandardCharsets.UTF_8,
                    StandardOpenOption.CREATE, StandardOpenOption.TRUNCATE_EXISTING);
        } catch (Exception ex) {
            CoreLogger.get().warn(LOG, "Failed to save " + file + ": " + ex.getMessage());
        }
    }
}
