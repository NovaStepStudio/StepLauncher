package dev.novastep.core.minecraft.version;

import java.util.Map;

public class AssetIndexManifest {
    private static final String ASSETS_CDN = "https://resources.download.minecraft.net";
    public boolean virtual;
    public boolean map_to_resources;
    public Map<String, Asset> objects;

    public static class Asset {
        public String hash;
        public long size;

        public String prefix() {
            return hash.substring(0, 2);
        }

        public String objectPath() {
            return prefix() + "/" + hash;
        }

        public String downloadUrl() {
            return ASSETS_CDN + "/" + prefix() + "/" + hash;
        }
    }

    public int totalCount() {
        return objects == null ? 0 : objects.size();
    }

    public long totalBytes() {
        if (objects == null)
            return 0L;
        long total = 0L;
        for (Asset a : objects.values())
            total += a.size;
        return total;
    }
}
