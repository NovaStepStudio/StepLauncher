package dev.novastep.core.minecraft;

import com.fasterxml.jackson.core.type.TypeReference;
import dev.novastep.core.json.Json;
import dev.novastep.core.log.CoreLogger;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.atomic.AtomicBoolean;

public class SessionManager {

    private static final String LOG = "SessionManager";

    private static Path sessionsFile;
    private static final AtomicBoolean privacyEnabled = new AtomicBoolean(false);
    private static List<Map<String, Object>> sessions = new ArrayList<>();

    public static void init(Path rootDir) {
        sessionsFile = rootDir.resolve("sessions.json");
        loadSessions();
    }

    public static void setPrivacyEnabled(boolean enabled) {
        privacyEnabled.set(enabled);
        if (!enabled) {
            CoreLogger.get().info("Privacy", "Session tracking disabled by configuration");
            sessions.clear();
            saveSessions();
        } else {
            CoreLogger.get().info("Privacy", "Session tracking enabled");
        }
    }

    public static boolean isPrivacyEnabled() {
        return privacyEnabled.get();
    }

    private static void loadSessions() {
        if (sessionsFile == null || !Files.exists(sessionsFile)) {
            return;
        }
        try {
            sessions = Json.read(Files.readString(sessionsFile, StandardCharsets.UTF_8),
                    new TypeReference<List<Map<String, Object>>>() {
                    });
            if (sessions == null) {
                sessions = new ArrayList<>();
            }
        } catch (IOException ex) {
            CoreLogger.get().warn(LOG, "Could not load sessions: " + ex.getMessage());
        }
    }

    private static void saveSessions() {
        if (sessionsFile == null) {
            return;
        }
        try {
            Files.writeString(sessionsFile, Json.writePretty(sessions), StandardCharsets.UTF_8);
        } catch (IOException ex) {
            CoreLogger.get().warn(LOG, "Could not save sessions: " + ex.getMessage());
        }
    }

    public static void recordSession(Map<String, Object> sessionData) {
        if (!privacyEnabled.get()) {
            return;
        }
        sessions.add(sessionData);
        if (sessions.size() > 100) {
            sessions.remove(0);
        }
        saveSessions();
    }

    public static List<Map<String, Object>> getSessions() {
        return privacyEnabled.get() ? new ArrayList<>(sessions) : new ArrayList<>();
    }
}
