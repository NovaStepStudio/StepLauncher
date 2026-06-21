package dev.novastep.core.minecraft.version;

import java.util.List;

public class VersionManifest {

    public Latest latest;
    public List<VersionEntry> versions;

    public static class Latest {
        public String release;
        public String snapshot;
    }

    public static class VersionEntry {
        public String id;
        public String type;
        public String url;
        public String time;
        public String releaseTime;
        public String sha1;
        public int complianceLevel;
        
        @Override
        public String toString() {
            return String.format("VersionEntry{id='%s', type='%s'}", id, type);
        }
    }

    public VersionEntry findById(String id) {
        if (versions == null) return null;
        for (VersionEntry v : versions) {
            if (id.equals(v.id)) return v;
        }
        return null;
    }

    public boolean contains(String id) {
        return findById(id) != null;
    }
}
