package dev.novastep.core.minecraft.world;

public final class WorldModels {
    private WorldModels() {
    }

    public record WorldMetadata(
            String folderName,
            String levelName,
            long lastPlayed,
            int gameType,
            boolean hardcore,
            String versionName,
            int dataVersion,
            long dayTime,
            long seed,
            String path,
            String instanceId,
            String iconBase64,
            WorldFlags flags) {
    }

    public record WorldFlags(
            boolean locked,
            boolean corrupted,
            boolean modded,
            boolean valid) {
    }
}
