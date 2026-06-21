package dev.novastep.core;

import dev.novastep.core.log.CoreLogger;

import java.io.InputStream;
import java.util.Properties;
import java.util.jar.Attributes;
import java.util.jar.Manifest;

public final class CoreVersion {

    private static final String LOG             = "CoreVersion";
    private static final String FALLBACK        = "dev";
    private static volatile String cachedVersion = null;

    private CoreVersion() {}

    public static String get() {
        String v = cachedVersion;
        if (v != null) return v;

        synchronized (CoreVersion.class) {
            v = cachedVersion;
            if (v != null) return v;
            cachedVersion = v = resolve();
        }
        return v;
    }

    private static String resolve() {
        try {
            InputStream is = CoreVersion.class.getResourceAsStream("/META-INF/MANIFEST.MF");
            if (is != null) {
                try (is) {
                    Manifest manifest = new Manifest(is);
                    Attributes attrs  = manifest.getMainAttributes();
                    String v = attrs.getValue("Implementation-Version");
                    if (v != null && !v.isBlank()) {
                        CoreLogger.get().debug(LOG, "Version from MANIFEST.MF: " + v);
                        return v.trim();
                    }
                }
            }
        } catch (Exception ex) {
            CoreLogger.get().debug(LOG, "MANIFEST.MF read failed: " + ex.getMessage());
        }

        try {
            InputStream is = CoreVersion.class.getResourceAsStream("/META-INF/novacore.properties");
            if (is != null) {
                try (is) {
                    Properties props = new Properties();
                    props.load(is);
                    String v = props.getProperty("version");
                    if (v != null && !v.isBlank()) {
                        CoreLogger.get().debug(LOG, "Version from novacore.properties: " + v);
                        return v.trim();
                    }
                }
            }
        } catch (Exception ex) {
            CoreLogger.get().debug(LOG, "novacore.properties read failed: " + ex.getMessage());
        }

        CoreLogger.get().debug(LOG, "Version not found in resources — using fallback: " + FALLBACK);
        return FALLBACK;
    }
}
