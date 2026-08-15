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
	d := &GameEventData{
		ID:               inst.ID,
		PID:              inst.PID,
		Version:          inst.Version,
		InstanceID:       inst.InstanceID,
		PlayerName:       inst.PlayerName,
		Status:           string(inst.GetStatus()),
		ExitCode:         inst.GetExitCode(),
		CrashLog:         inst.CrashLog,
		CrashLogText:     inst.CrashLogContent,
		GameOutputText:   inst.GameOutput,
		CrashReason:      inst.CrashReason,
		CrashCategory:    inst.CrashCategory,
		LauncherLogPath:  inst.LauncherLogPath,
		MinecraftLogPath: inst.LogPath,
		JvmLogPath:       inst.CrashLog,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
	}
	if inst.PreInfo != nil {
		d.LaunchInfo = gamelog.FormatPreLaunchInfo(*inst.PreInfo)
		d.JavaExec = inst.PreInfo.JavaExec
		d.MaxRAM = inst.PreInfo.MaxRAM
		d.VanillaVersion = inst.PreInfo.VanillaVersionID
	}
	if !inst.StartTime.IsZero() {
		d.UptimeMs = time.Since(inst.StartTime).Milliseconds()
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
