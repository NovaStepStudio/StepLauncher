package dev.novastep.core.minecraft;
import dev.novastep.core.minecraft.common.ModuleStatus;

import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.websocket.EventBroadcaster;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

public final class ModuleTracker {

    private static final String LOG = "ModuleTracker";

    private final String sessionId;
    private final EventBroadcaster broadcaster;
    private final ConcurrentHashMap<String, ModuleStatus> states = new ConcurrentHashMap<>();

    public ModuleTracker(String sessionId, EventBroadcaster broadcaster) {
        this.sessionId   = sessionId;
        this.broadcaster = broadcaster;
    }

    public void register(String module) {
        states.put(module, ModuleStatus.PENDING);
        emitEvent(module, ModuleStatus.PENDING);
    }

    public void transition(String module, ModuleStatus newStatus) {
        ModuleStatus old = states.put(module, newStatus);
        if (old == newStatus) return;

        CoreLogger.get().debug(LOG,
                "[" + sessionId + "] " + module + ": "
                + (old != null ? old.name() : "?") + " → " + newStatus.name());
        emitEvent(module, newStatus);
    }

    public ModuleStatus getStatus(String module) {
        return states.getOrDefault(module, ModuleStatus.PENDING);
    }

    public boolean allVerified() {
        if (states.isEmpty()) return true;
        return states.values().stream().allMatch(s -> s == ModuleStatus.VERIFIED);
    }

    public boolean anyFailed() {
        return states.values().stream().anyMatch(s -> s == ModuleStatus.FAILED);
    }

    public Map<String, String> snapshot() {
        Map<String, String> snap = new LinkedHashMap<>();
        states.forEach((k, v) -> snap.put(k, v.name().toLowerCase()));
        return Collections.unmodifiableMap(snap);
    }

    public Set<String> modules() {
        return Collections.unmodifiableSet(states.keySet());
    }

    private void emitEvent(String module, ModuleStatus status) {
        try {
            broadcaster.emit("module_status", Map.of(
                    "sessionId", sessionId,
                    "module",    module,
                    "status",    status.name().toLowerCase()
            ));
        } catch (Exception ex) {
            CoreLogger.get().warn(LOG,
                    "[" + sessionId + "] Failed to emit module_status event: " + ex.getMessage());
        }
    }
}
