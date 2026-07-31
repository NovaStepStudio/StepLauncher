package launcher

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"StepLauncher/internal/Core/Launcher/Helpers"
	"StepLauncher/internal/Core/Launcher/Utils"
)

type ManagerConfig struct {
	WorkDir              string
	LogDir               string
	LogFn                func(string, ...interface{})
	LauncherName         string
	LauncherVersion      string
	GameLogBroadcastFn   func(stream, line string)
	GameEventBroadcastFn func([]byte)
	GameEventReplayFn    func([]byte)
	OnGameExitFn         func(instance *GameInstance, playTimeSeconds int)
}

type LaunchManager struct {
	mu         sync.RWMutex
	games      map[string]*GameInstance
	nextID     uint64
	config     ManagerConfig
	eventBuf   []GameEvent
	eventBufMu sync.RWMutex
}

func NewManager(cfg ManagerConfig) *LaunchManager {
	m := &LaunchManager{
		games:    make(map[string]*GameInstance),
		config:   cfg,
		eventBuf: make([]GameEvent, 0, 100),
	}

	if cfg.GameEventBroadcastFn != nil {
		originalFn := cfg.GameEventBroadcastFn
		m.config.GameEventBroadcastFn = func(data []byte) {
			var evt GameEvent
			if json.Unmarshal(data, &evt) == nil {
				m.appendGameEvent(evt)
			}
			originalFn(data)
		}
	}

	return m
}

func (m *LaunchManager) appendGameEvent(evt GameEvent) {
	m.eventBufMu.Lock()
	defer m.eventBufMu.Unlock()
	const maxEvents = 100
	if len(m.eventBuf) >= maxEvents {
		m.eventBuf = m.eventBuf[1:]
	}
	m.eventBuf = append(m.eventBuf, evt)
}

func (m *LaunchManager) GetEventBuffer() []GameEvent {
	m.eventBufMu.RLock()
	defer m.eventBufMu.RUnlock()
	buf := make([]GameEvent, len(m.eventBuf))
	copy(buf, m.eventBuf)
	return buf
}

func (m *LaunchManager) ReplayEvents(fn func([]byte)) {
	if fn == nil {
		return
	}
	for _, evt := range m.GetEventBuffer() {
		data, _ := json.Marshal(evt)
		fn(data)
	}
}

func (m *LaunchManager) Launch(cfg LaunchConfig) (*GameInstance, error) {
	if cfg.LauncherName == "" {
		cfg.LauncherName = m.config.LauncherName
	}
	if cfg.LauncherVersion == "" {
		cfg.LauncherVersion = m.config.LauncherVersion
	}
	if cfg.LogDir == "" {
		cfg.LogDir = filepath.Join(m.config.LogDir, "game")
	}
	if cfg.LogFn == nil {
		cfg.LogFn = m.config.LogFn
	}

	adv := cfg.Adv()

	if adv.RuntimeDir == "" {
		adv.RuntimeDir = filepath.Join(m.config.WorkDir, "runtime")
	}
	if adv.GameDir == "" {
		adv.GameDir = filepath.Join(m.config.WorkDir, "game")
	}
	if adv.AssetsDir == "" {
		adv.AssetsDir = filepath.Join(m.config.WorkDir, "assets")
	}
	if adv.LibrariesDir == "" {
		adv.LibrariesDir = filepath.Join(m.config.WorkDir, "libraries")
	}
	if adv.VersionsDir == "" {
		adv.VersionsDir = filepath.Join(m.config.WorkDir, "versions")
	}
	if adv.NativesBaseDir == "" {
		adv.NativesBaseDir = filepath.Join(m.config.WorkDir, "versions")
	}
	if adv.CleanupDelay <= 0 {
		adv.CleanupDelay = 5 * time.Minute
	}
	if adv.MaxRAM <= 0 {
		adv.MaxRAM = 2048
	}
	if adv.MinRAM <= 0 {
		adv.MinRAM = 512
	}
	if adv.UserType == "" {
		adv.UserType = "mojang"
	}
	if adv.DownloadConcurrency <= 0 {
		adv.DownloadConcurrency = 4
	}
	if adv.MaxRetries <= 0 {
		adv.MaxRetries = 3
	}
	if adv.ConnectionTimeout <= 0 {
		adv.ConnectionTimeout = 30
	}
	if adv.GameLogLines <= 0 {
		adv.GameLogLines = 1500
	}
	if adv.GameLogLines < 100 {
		adv.GameLogLines = 100
	}
	if adv.GameLogLines > 10000 {
		adv.GameLogLines = 10000
	}
	cfg.Advanced = &adv

	launcher := &Launcher{
		cfg:              cfg,
		gameLogBroadcast: m.config.GameLogBroadcastFn,
		eventBroadcast:   m.config.GameEventBroadcastFn,
		onGameExit:       m.config.OnGameExitFn,
	}
	instance, err := launcher.Launch()
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.nextID++
	id := fmt.Sprintf("game-%d", m.nextID)
	instance.ID = id
	m.games[id] = instance
	m.mu.Unlock()

	if cfg.LogFn != nil {
		cfg.LogFn("[LaunchManager] Launched game %s (PID: %d, version: %s)", id, instance.PID, instance.Version)
	}

	go func() {
		<-instance.Done()
		m.mu.Lock()
		delete(m.games, id)
		m.mu.Unlock()
	}()

	return instance, nil
}

func (m *LaunchManager) Stop(id string) error {
	m.mu.RLock()
	instance, ok := m.games[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("game %s not found", id)
	}

	utils.KillTree(instance.PID)
	instance.setStatus(GameStopped)

	if m.config.LogFn != nil {
		m.config.LogFn("[LaunchManager] Stopped game %s (PID: %d)", id, instance.PID)
	}

	return nil
}

func (m *LaunchManager) StopAll() {
	m.mu.RLock()
	ids := make([]string, 0, len(m.games))
	for id := range m.games {
		ids = append(ids, id)
	}
	m.mu.RUnlock()

	for _, id := range ids {
		m.Stop(id)
	}
}

func (m *LaunchManager) Get(id string) *GameInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.games[id]
}

func (m *LaunchManager) List() []*GameInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*GameInstance, 0, len(m.games))
	for _, g := range m.games {
		list = append(list, g)
	}
	return list
}

func (m *LaunchManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.games, id)
}

func (m *LaunchManager) RecommendedRAM() (minRAM, maxRAM int, gcPreset string) {
	maxRAM = helpers.RecommendedMaxRAM(4096)
	minRAM = helpers.MinRAM
	if minRAM > maxRAM/2 {
		minRAM = maxRAM / 2
	}
	gcPreset = helpers.RecommendedGCPreset(maxRAM)
	return
}
