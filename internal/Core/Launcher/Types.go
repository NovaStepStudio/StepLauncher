package launcher

import (
	"os/exec"
	"sync"
	"time"
)

type GameStatus string

const (
	GameStarting GameStatus = "starting"
	GameRunning  GameStatus = "running"
	GameExited   GameStatus = "exited"
	GameCrashed  GameStatus = "crashed"
	GameStopped  GameStatus = "stopped"
)

type GameInstance struct {
	mu            sync.RWMutex
	ID            string
	PID           int
	Version       string
	InstanceID    string
	PlayerName    string
	StartTime     time.Time
	Status        GameStatus
	ExitCode      int
	LogPath         string
	CrashLog        string
	CrashLogContent string
	CrashReason     string
	CrashCategory   string
	cmd           *exec.Cmd
	done          chan struct{}
	eventBuf      []GameEvent
	eventBufMu    sync.RWMutex
}

func (g *GameInstance) IsRunning() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Status == GameRunning || g.Status == GameStarting
}

func (g *GameInstance) GetStatus() GameStatus {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Status
}

func (g *GameInstance) setStatus(s GameStatus) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Status = s
}

func (g *GameInstance) Done() <-chan struct{} {
	return g.done
}

func (g *GameInstance) SetExitCode(code int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ExitCode = code
}

func (g *GameInstance) SetStatus(s GameStatus) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Status = s
}

func (g *GameInstance) GetExitCode() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.ExitCode
}

func (g *GameInstance) AppendEvent(evt GameEvent) {
	g.eventBufMu.Lock()
	defer g.eventBufMu.Unlock()
	const maxEvents = 50
	if len(g.eventBuf) >= maxEvents {
		g.eventBuf = g.eventBuf[1:]
	}
	g.eventBuf = append(g.eventBuf, evt)
}

func (g *GameInstance) GetEventBuf() []GameEvent {
	g.eventBufMu.RLock()
	defer g.eventBufMu.RUnlock()
	buf := make([]GameEvent, len(g.eventBuf))
	copy(buf, g.eventBuf)
	return buf
}
