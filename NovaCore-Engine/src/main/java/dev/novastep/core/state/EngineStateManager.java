package dev.novastep.core.state;

import dev.novastep.core.downloader.DownloadManager;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.websocket.EventBroadcaster;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

public final class EngineStateManager {

    private static final String LOG = "EngineStateManager";

    public enum State {
        ACTIVE,
        IDLE,
        SEMI_OFF
    }

    private final DownloadManager  downloadManager;
    private final EventBroadcaster broadcaster;
    private final Object           monitor = new Object();

    private volatile State  state          = State.ACTIVE;
    private volatile long   changedAt      = System.currentTimeMillis();
    private List<String>    pausedBySemiOff = List.of();

    public EngineStateManager(DownloadManager downloadManager, EventBroadcaster broadcaster) {
        this.downloadManager = downloadManager;
        this.broadcaster     = broadcaster;
    }

    // ─── Queries ──────────────────────────────────────────────────────────────

    public State   currentState() { return state; }
    public boolean isActive()     { return state == State.ACTIVE;   }
    public boolean isIdle()       { return state == State.IDLE;     }
    public boolean isSemiOff()    { return state == State.SEMI_OFF; }

    public Map<String, Object> snapshot() {
        Map<String, Object> snap = new LinkedHashMap<>();
        snap.put("state",            state.name().toLowerCase());
        snap.put("changedAt",        changedAt);
        snap.put("pausedSessions",   new ArrayList<>(pausedBySemiOff));
        snap.put("essentialServices", List.of("http", "websocket", "state-manager"));
        return snap;
    }

    // ─── Transitions ──────────────────────────────────────────────────────────

    public Map<String, Object> setState(State newState, String reason) {
        synchronized (monitor) {
            if (newState == state)
                return snapshot();

            applyTransition(state, newState);

            state     = newState;
            changedAt = System.currentTimeMillis();

            CoreLogger.get().info(LOG, "Engine state → " + state.name().toLowerCase()
                    + (reason != null && !reason.isBlank() ? " (" + reason + ")" : ""));

            broadcaster.emit("engine_state", snapshot());
            return snapshot();
        }
    }

    // ─── Internal ─────────────────────────────────────────────────────────────

    private void applyTransition(State from, State to) {
        // Leaving SEMI_OFF → resume paused sessions first
        if (from == State.SEMI_OFF) {
            downloadManager.resumeSessions(pausedBySemiOff);
            pausedBySemiOff = List.of();
        }

        switch (to) {
            case ACTIVE -> {
                // Full operation: nothing extra to do — resumeSessions above is sufficient.
            }
            case IDLE -> {
                // Throttle the download pool so the game process gets more CPU.
                // Sessions remain active but new downloads are rate-limited.
                downloadManager.throttle(1);
            }
            case SEMI_OFF -> {
                // Pause all running sessions and minimize the pool.
                pausedBySemiOff = downloadManager.pauseRunningSessions();
                downloadManager.throttle(0);
            }
        }

        // Restore full pool capacity when leaving IDLE
        if (from == State.IDLE && to != State.SEMI_OFF) {
            downloadManager.throttle(-1); // -1 = restore to configured max
        }
    }
}
