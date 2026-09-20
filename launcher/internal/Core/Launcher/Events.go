package launcher

import (
	"encoding/json"
	"time"

	gamelog "StepLauncher/internal/Core/Launcher/Log"
)

type GameEventType string

const (
	EvGameStarting GameEventType = "game_starting"
	EvGamePrepare  GameEventType = "game_prepare"
	EvGameStarted  GameEventType = "game_started"
	EvGameExited   GameEventType = "game_exited"
	EvGameCrashed  GameEventType = "game_crashed"
	EvGameStopped  GameEventType = "game_stopped"
)

type GamePrepareData struct {
	Version  string `json:"version"`
	Phase    string `json:"phase"`
	Current  int    `json:"current"`
	Total    int    `json:"total"`
	Label    string `json:"label,omitempty"`
	Message  string `json:"message,omitempty"`
	Finished bool   `json:"finished,omitempty"`
}

type GameEventData struct {
	ID               string `json:"id"`
	PID              int    `json:"pid"`
	Version          string `json:"version"`
	InstanceID       string `json:"instanceId,omitempty"`
	PlayerName       string `json:"playerName,omitempty"`
	Status           string `json:"status"`
	ExitCode         int    `json:"exitCode,omitempty"`
	CrashLog         string `json:"crashLog,omitempty"`
	CrashLogText     string `json:"crashLogText,omitempty"`
	GameOutputText   string `json:"gameOutputText,omitempty"`
	CrashReason      string `json:"crashReason,omitempty"`
	CrashCategory    string `json:"crashCategory,omitempty"`
	LauncherLogPath  string `json:"launcherLogPath,omitempty"`
	MinecraftLogPath string `json:"minecraftLogPath,omitempty"`
	JvmLogPath       string `json:"jvmLogPath,omitempty"`
	LaunchInfo       string `json:"launchInfo,omitempty"`
	UptimeMs         int64  `json:"uptimeMs,omitempty"`
	Timestamp        string `json:"timestamp"`
	JavaExec         string `json:"javaExec,omitempty"`
	MaxRAM           int    `json:"maxRam,omitempty"`
	VanillaVersion   string `json:"vanillaVersion,omitempty"`
}

type GameEvent struct {
	Type GameEventType  `json:"type"`
	Data *GameEventData `json:"data"`
}

func NewGameEventData(inst *GameInstance) *GameEventData {
	snapshot := inst.Snapshot()
	d := &GameEventData{
		ID:               snapshot.ID,
		PID:              snapshot.PID,
		Version:          snapshot.Version,
		InstanceID:       snapshot.InstanceID,
		PlayerName:       snapshot.PlayerName,
		Status:           string(snapshot.Status),
		ExitCode:         snapshot.ExitCode,
		CrashLog:         snapshot.CrashLog,
		CrashLogText:     snapshot.CrashLogContent,
		GameOutputText:   snapshot.GameOutput,
		CrashReason:      snapshot.CrashReason,
		CrashCategory:    snapshot.CrashCategory,
		LauncherLogPath:  snapshot.LauncherLogPath,
		MinecraftLogPath: snapshot.LogPath,
		JvmLogPath:       snapshot.CrashLog,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
	}
	if snapshot.PreInfo != nil {
		d.LaunchInfo = gamelog.FormatPreLaunchInfo(*snapshot.PreInfo)
		d.JavaExec = snapshot.PreInfo.JavaExec
		d.MaxRAM = snapshot.PreInfo.MaxRAM
		d.VanillaVersion = snapshot.PreInfo.VanillaVersionID
	}
	if !snapshot.StartTime.IsZero() {
		d.UptimeMs = time.Since(snapshot.StartTime).Milliseconds()
	}
	return d
}

func broadcastEvent(fn func([]byte), evt *GameEvent) {
	if fn == nil {
		return
	}
	data, _ := json.Marshal(evt)
	fn(data)
}

func BroadcastPrepare(fn func([]byte), data *GamePrepareData) {
	if fn == nil {
		return
	}
	raw, _ := json.Marshal(map[string]interface{}{
		"type": string(EvGamePrepare),
		"data": data,
	})
	fn(raw)
}

func BroadcastStarting(fn func([]byte), inst *GameInstance) {
	broadcastEvent(fn, &GameEvent{
		Type: EvGameStarting,
		Data: NewGameEventData(inst),
	})
}

func BroadcastStarted(fn func([]byte), inst *GameInstance) {
	broadcastEvent(fn, &GameEvent{
		Type: EvGameStarted,
		Data: NewGameEventData(inst),
	})
}

func BroadcastExited(fn func([]byte), inst *GameInstance) {
	broadcastEvent(fn, &GameEvent{
		Type: EvGameExited,
		Data: NewGameEventData(inst),
	})
}

func BroadcastCrashed(fn func([]byte), inst *GameInstance) {
	broadcastEvent(fn, &GameEvent{
		Type: EvGameCrashed,
		Data: NewGameEventData(inst),
	})
}

func BroadcastStopped(fn func([]byte), inst *GameInstance) {
	broadcastEvent(fn, &GameEvent{
		Type: EvGameStopped,
		Data: NewGameEventData(inst),
	})
}
