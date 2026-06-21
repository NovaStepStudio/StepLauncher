package dev.novastep.core.server.handlers;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.minecraft.world.WorldModels.WorldFlags;
import dev.novastep.core.minecraft.world.WorldModels.WorldMetadata;
import dev.novastep.core.server.HttpUtils;
import dev.novastep.core.util.MemoryOptimizer;
import dev.novastep.core.util.NbtReader;

import java.io.IOException;
import java.io.InputStream;
import java.nio.channels.FileChannel;
import java.nio.channels.FileLock;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardOpenOption;
import java.util.ArrayList;
import java.util.Base64;
import java.util.Comparator;
import java.util.List;
import java.util.Map;

import static dev.novastep.core.util.NbtReader.getAsBoolean;
import static dev.novastep.core.util.NbtReader.getAsInt;
import static dev.novastep.core.util.NbtReader.getAsLong;
import static dev.novastep.core.util.NbtReader.getAsString;
import static dev.novastep.core.util.NbtReader.getNested;

public class QuickPlayHandler implements HttpHandler {

    private static final String LOG = "QuickPlayHandler";
    private static final int MAX_CACHE_ENTRIES = 256;

    private final Path instancesDir;
    private static final Map<String, CachedWorld> CACHE = MemoryOptimizer.newSynchronizedLruCache(MAX_CACHE_ENTRIES);

    private record CachedWorld(WorldMetadata metadata, long lastModified) {
    }

    public QuickPlayHandler(Path instancesDir) {
        this.instancesDir = instancesDir;
    }

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        if (!"GET".equalsIgnoreCase(exchange.getRequestMethod())) {
            HttpUtils.methodNotAllowed(exchange);
            return;
        }
        handleGetWorlds(exchange);
    }

    private void handleGetWorlds(HttpExchange exchange) throws IOException {
        try {
            List<WorldMetadata> allWorlds = new ArrayList<>();
            if (!Files.exists(instancesDir)) {
                HttpUtils.ok(exchange, Map.of("worlds", List.of()));
                return;
            }

            try (DirectoryStream<Path> stream = Files.newDirectoryStream(instancesDir)) {
                for (Path instancePath : stream) {
                    if (!Files.isDirectory(instancePath)) {
                        continue;
                    }
                    Path savesDir = resolveSavesDir(instancePath);
                    if (savesDir != null) {
                        scanSavesFolder(savesDir, allWorlds, instancePath.getFileName().toString());
                    }
                }
            }

            allWorlds.sort(Comparator.comparingLong(WorldMetadata::lastPlayed).reversed());
            HttpUtils.ok(exchange, Map.of("worlds", allWorlds));
        } catch (Exception ex) {
            CoreLogger.get().error(LOG, "World discovery failed", ex);
            HttpUtils.serverError(exchange, ex.getMessage());
        }
    }

    private Path resolveSavesDir(Path instanceDir) {
        Path[] candidates = {
                instanceDir.resolve("game").resolve("saves"),
                instanceDir.resolve("saves"),
                instanceDir.resolve(".minecraft").resolve("saves")
        };
        for (Path candidate : candidates) {
            if (Files.isDirectory(candidate)) {
                return candidate;
            }
        }
        return null;
    }

    private void scanSavesFolder(Path savesDir, List<WorldMetadata> worlds, String instanceId) {
        try (DirectoryStream<Path> stream = Files.newDirectoryStream(savesDir)) {
            for (Path worldDir : stream) {
                if (!Files.isDirectory(worldDir)) {
                    continue;
                }
                WorldMetadata metadata = processWorldFolder(worldDir, instanceId);
                if (metadata != null) {
                    worlds.add(metadata);
                }
            }
        } catch (IOException ex) {
            CoreLogger.get().error(LOG, "Failed reading saves folder " + savesDir, ex);
        }
    }

    private WorldMetadata processWorldFolder(Path worldDir, String instanceId) throws IOException {
        Path levelDat = worldDir.resolve("level.dat");
        Path levelDatOld = worldDir.resolve("level.dat_old");
        if (!Files.exists(levelDat) && !Files.exists(levelDatOld)) {
            return null;
        }

        Path target = Files.exists(levelDat) ? levelDat : levelDatOld;
        long lastModified = Files.getLastModifiedTime(target).toMillis();
        String cacheKey = target.toAbsolutePath().toString();

        CachedWorld cached = CACHE.get(cacheKey);
        if (cached != null && cached.lastModified == lastModified) {
            return cached.metadata;
        }

        boolean locked = isLocked(worldDir);
        Map<String, Object> nbt = null;
        boolean corrupted = false;

        try {
            if (Files.exists(levelDat)) {
                try (InputStream inputStream = Files.newInputStream(levelDat)) {
                    nbt = new NbtReader(inputStream).parse();
                }
            }
            if (nbt == null && Files.exists(levelDatOld)) {
                try (InputStream inputStream = Files.newInputStream(levelDatOld)) {
                    nbt = new NbtReader(inputStream).parse();
                }
            }
        } catch (Exception ex) {
            corrupted = true;
            CoreLogger.get().error(LOG, "Corrupted world: " + worldDir.getFileName(), ex);
        }

        if (nbt == null) {
            return createFallback(worldDir, instanceId, locked, corrupted);
        }

        Map<String, Object> data = getNested(nbt, "Data");
        if (data == null) {
            data = nbt;
        }

        String levelName = getAsString(data.get("LevelName"), worldDir.getFileName().toString());
        long lastPlayed = getAsLong(data.get("LastPlayed"), lastModified);
        int gameType = getAsInt(data.get("GameType"), 0);
        boolean hardcore = getAsBoolean(data.get("hardcore"), false);
        long dayTime = getAsLong(data.get("Time"), 0);

        Long seedValue = getNested(data, "WorldGenSettings.seed");
        long seed = seedValue != null ? seedValue : getAsLong(data.get("RandomSeed"), 0);

        int dataVersion = getAsInt(data.get("DataVersion"), 0);
        String versionName = getAsString(getNested(data, "Version.Name"), inferVersion(dataVersion));
        boolean modded = getAsBoolean(data.get("WasModded"), false);

        String icon = null;
        Path iconPath = worldDir.resolve("icon.png");
        if (Files.exists(iconPath)) {
            try {
                icon = "data:image/png;base64," + Base64.getEncoder().encodeToString(Files.readAllBytes(iconPath));
            } catch (IOException ignored) {
            }
        }

        WorldMetadata metadata = new WorldMetadata(
                worldDir.getFileName().toString(),
                levelName,
                lastPlayed,
                gameType,
                hardcore,
                versionName,
                dataVersion,
                dayTime,
                seed,
                worldDir.toAbsolutePath().toString(),
                instanceId,
                icon,
                new WorldFlags(locked, corrupted, modded, true)
        );

        CACHE.put(cacheKey, new CachedWorld(metadata, lastModified));
        return metadata;
    }

    private boolean isLocked(Path worldDir) {
        Path lock = worldDir.resolve("session.lock");
        if (!Files.exists(lock)) {
            return false;
        }
        try (FileChannel channel = FileChannel.open(lock, StandardOpenOption.WRITE)) {
            FileLock lockHandle = channel.tryLock();
            if (lockHandle == null) {
                return true;
            }
            lockHandle.release();
            return false;
        } catch (Exception ex) {
            return true;
        }
    }

    private WorldMetadata createFallback(Path dir, String instanceId, boolean locked, boolean corrupted) {
        return new WorldMetadata(
                dir.getFileName().toString(),
                dir.getFileName().toString() + " (Inaccessible)",
                0, 0, false, "Unknown", 0, 0, 0,
                dir.toAbsolutePath().toString(),
                instanceId,
                null,
                new WorldFlags(locked, corrupted, false, false)
        );
    }

    private String inferVersion(int dataVersion) {
        if (dataVersion >= 3953) return "1.21.x";
        if (dataVersion >= 3463) return "1.20.x";
        if (dataVersion >= 3105) return "1.19.x";
        if (dataVersion >= 2860) return "1.18.x";
        if (dataVersion >= 2724) return "1.17.x";
        if (dataVersion >= 2566) return "1.16.x";
        if (dataVersion >= 2230) return "1.15.x";
        if (dataVersion >= 1976) return "1.14.x";
        if (dataVersion >= 1519) return "1.13.x";
        if (dataVersion >= 1343) return "1.12.x";
        return "Legacy";
    }
}
