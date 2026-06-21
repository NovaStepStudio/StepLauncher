package dev.novastep.core.modloader.resolver;

import java.util.ArrayList;
import java.util.List;

public final class NeoForgeVersionResolver {

    private static final int NEW_SCHEME_MAJOR = 26;

    private NeoForgeVersionResolver() {}

    public static String mcVersionToNeoForgePrefix(String mcVersion) {
        if (mcVersion == null || mcVersion.isBlank()) {
            throw new IllegalArgumentException("Minecraft version must not be blank");
        }
        String[] parts = mcVersion.split("\\.");
        if (parts.length < 2) {
            throw new IllegalArgumentException("Invalid Minecraft version: " + mcVersion);
        }
        int major = Integer.parseInt(parts[0]);
        if (major == 1) {
            if (parts.length == 2) return parts[1];
            return parts[1] + "." + parts[2];
        }
        if (major >= NEW_SCHEME_MAJOR) {
            return parts[0] + "." + (parts.length > 1 ? parts[1] : "0");
        }
        throw new IllegalArgumentException("Unrecognized Minecraft version scheme: " + mcVersion);
    }

    public static String neoForgeVersionToMcVersion(String neoForgeVersion) {
        if (neoForgeVersion == null || neoForgeVersion.isBlank()) {
            throw new IllegalArgumentException("NeoForge version must not be blank");
        }
        String[] parts = neoForgeVersion.split("\\.");
        if (parts.length < 2) {
            throw new IllegalArgumentException("Invalid NeoForge version: " + neoForgeVersion);
        }
        int major = Integer.parseInt(parts[0]);
        int minor = Integer.parseInt(parts[1]);
        if (major >= NEW_SCHEME_MAJOR) {
            return major + "." + minor;
        }
        if (parts.length == 2) return "1." + major;
        return "1." + major + "." + minor;
    }

    public static List<String> filterVersionsForMinecraft(List<String> allVersions, String mcVersion) {
        String prefix = mcVersionToNeoForgePrefix(mcVersion);
        List<String> result = new ArrayList<>();
        for (String v : allVersions) {
            if (v.startsWith(prefix + ".") || v.equals(prefix)) {
                result.add(v);
            }
        }
        return result;
    }
}
