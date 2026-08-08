package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Entry struct {
	ID              string `json:"id"`
	Timestamp       int64  `json:"timestamp"`
	InstanceID      string `json:"instance_id,omitempty"`
	Version         string `json:"version"`
	PlayerName      string `json:"player_name"`
	PlayTimeSeconds int    `json:"play_time_seconds"`
	ExitCode        int    `json:"exit_code"`
	CrashReason     string `json:"crash_reason,omitempty"`
	ModLoader       string `json:"modloader,omitempty"`
	MaxRAM          int    `json:"max_ram,omitempty"`
}

type FileData struct {
	Entries []Entry `json:"entries"`
}

type Manager struct {
	mu       sync.RWMutex
	filePath string
	data     FileData
}

func NewManager(workDir string) *Manager {
	return &Manager{
		filePath: filepath.Join(workDir, "launcher_history.json"),
		data: FileData{
			Entries: make([]Entry, 0),
		},
	}
}

func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &m.data); err != nil {
		return fmt.Errorf("failed to parse history file: %w", err)
	}

	return nil
}

func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.data, "", "  ")
	if err != nil {
		return err
	}
	tmpPath := m.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, m.filePath)
}

func (m *Manager) EnsureFile() error {
	m.mu.Lock()
	if _, err := os.Stat(m.filePath); err == nil {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()
	return m.save()
}

func (m *Manager) AddEntry(entry Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("%d-%d", time.Now().UnixNano(), len(m.data.Entries))
	}
	if entry.Timestamp == 0 {
		entry.Timestamp = time.Now().Unix()
	}

	m.data.Entries = append(m.data.Entries, entry)
	return m.save()
}

func (m *Manager) GetEntries() []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]Entry, len(m.data.Entries))
	copy(res, m.data.Entries)
	return res
}

func (m *Manager) GetEntriesByVersion(version string) []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []Entry
	for _, e := range m.data.Entries {
		if e.Version == version {
			res = append(res, e)
		}
	}
	return res
}

func (m *Manager) GetMostPlayed(limit int) []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sorted := make([]Entry, len(m.data.Entries))
	copy(sorted, m.data.Entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PlayTimeSeconds > sorted[j].PlayTimeSeconds
	})
	if limit > len(sorted) {
		limit = len(sorted)
	}
	return sorted[:limit]
}

func (m *Manager) GetRecent(limit int) []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sorted := make([]Entry, len(m.data.Entries))
	copy(sorted, m.data.Entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp > sorted[j].Timestamp
	})
	if limit > len(sorted) {
		limit = len(sorted)
	}
	return sorted[:limit]
}

func (m *Manager) DeleteEntry(id string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	found := false
	var updated []Entry
	for _, e := range m.data.Entries {
		if e.ID == id {
			found = true
			continue
		}
		updated = append(updated, e)
	}

	if !found {
		return false, nil
	}

	m.data.Entries = updated
	err := m.save()
	return true, err
}

func (m *Manager) Clear() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := len(m.data.Entries)
	m.data.Entries = make([]Entry, 0)
	if err := m.save(); err != nil {
		return 0
	}
	return count
}
