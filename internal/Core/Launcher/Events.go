package launcher

import (
	"encoding/json"
	"time"
)

type GameEventType string

const (
	EvGameStarting GameEventType = "game_starting"
	EvGameStarted  GameEventType = "game_started"
	EvGameExited   GameEventType = "game_exited"
	EvGameCrashed  GameEventType = "game_crashed"
	EvGameStopped  GameEventType = "game_stopped"
)

type GameEventData struct {
	ID            string `json:"id"`
	PID           int    `json:"pid"`
	Version       string `json:"version"`
	InstanceID    string `json:"instanceId,omitempty"`
	PlayerName    string `json:"playerName,omitempty"`
	Status        string `json:"status"`
	ExitCode      int    `json:"exitCode,omitempty"`
	CrashLog      string `json:"crashLog,omitempty"`
	CrashReason   string `json:"crashReason,omitempty"`
	CrashCategory string `json:"crashCategory,omitempty"`
	UptimeMs      int64  `json:"uptimeMs,omitempty"`
	Timestamp     string `json:"timestamp"`
}

type GameEvent struct {
	Type GameEventType  `json:"type"`
	Data *GameEventData `json:"data"`
}

func NewGameEventData(inst *GameInstance) *GameEventData {
	d := &GameEventData{
		ID:            inst.ID,
		PID:           inst.PID,
		Version:       inst.Version,
		InstanceID:    inst.InstanceID,
		PlayerName:    inst.PlayerName,
		Status:        string(inst.GetStatus()),
		ExitCode:      inst.GetExitCode(),
		CrashLog:      inst.CrashLog,
		CrashReason:   inst.CrashReason,
		CrashCategory: inst.CrashCategory,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
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
