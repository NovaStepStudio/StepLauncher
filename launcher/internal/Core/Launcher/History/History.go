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
	InstanceName    string `json:"instance_name,omitempty"`
	Version         string `json:"version"`
	PlayerName      string `json:"player_name"`
	PlayTimeSeconds int    `json:"play_time_seconds"`
	ExitCode        int    `json:"exit_code"`
	CrashReason     string `json:"crash_reason,omitempty"`
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

type VersionStat struct {
	Version      string `json:"version"`
	PlayCount    int    `json:"playCount"`
	TotalPlayed  int64  `json:"totalPlayed"`
	FirstPlayed  int64  `json:"firstPlayed"`
	LastPlayed   int64  `json:"lastPlayed"`
}

type InstanceStats struct {
	TotalPlayTime  int64         `json:"totalPlayTime"`
	TotalSessions  int           `json:"totalSessions"`
	FirstPlayed    int64         `json:"firstPlayed"`
	LastPlayed     int64         `json:"lastPlayed"`
	WeeklyPlayTime int64         `json:"weeklyPlayTime"`
	WeeklySessions int           `json:"weeklySessions"`
	WeeklyVersions []string      `json:"weeklyVersions"`
	Versions       []VersionStat `json:"versions"`
	Running        bool          `json:"running"`
}

// GetInstanceStats agrega el historial de una instancia: tiempo total jugado,
// sesiones, primera y última fecha, resumen de la última semana y desglose por
// versión. El campo Running lo rellena el engine consultando el LaunchManager.
func (m *Manager) GetInstanceStats(instanceName string) InstanceStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := InstanceStats{
		Versions: make([]VersionStat, 0),
	}
	if instanceName == "" {
		return stats
	}

	versionStats := make(map[string]*VersionStat)
	weekCutoff := time.Now().AddDate(0, 0, -7).Unix()
	weeklyVersions := make(map[string]bool)

	for _, e := range m.data.Entries {
		if e.InstanceName != instanceName {
			continue
		}
		stats.TotalSessions++
		stats.TotalPlayTime += int64(e.PlayTimeSeconds)

		if stats.FirstPlayed == 0 || e.Timestamp < stats.FirstPlayed {
			stats.FirstPlayed = e.Timestamp
		}
		if e.Timestamp > stats.LastPlayed {
			stats.LastPlayed = e.Timestamp
		}

		if e.Timestamp >= weekCutoff {
			stats.WeeklySessions++
			stats.WeeklyPlayTime += int64(e.PlayTimeSeconds)
			if e.Version != "" && !weeklyVersions[e.Version] {
				weeklyVersions[e.Version] = true
				stats.WeeklyVersions = append(stats.WeeklyVersions, e.Version)
			}
		}

		vs, ok := versionStats[e.Version]
		if !ok {
			vs = &VersionStat{Version: e.Version}
			versionStats[e.Version] = vs
		}
		vs.PlayCount++
		vs.TotalPlayed += int64(e.PlayTimeSeconds)
		if vs.FirstPlayed == 0 || e.Timestamp < vs.FirstPlayed {
			vs.FirstPlayed = e.Timestamp
		}
		if e.Timestamp > vs.LastPlayed {
			vs.LastPlayed = e.Timestamp
		}
	}

	for _, vs := range versionStats {
		stats.Versions = append(stats.Versions, *vs)
	}
	sort.Slice(stats.Versions, func(i, j int) bool {
		return stats.Versions[i].LastPlayed > stats.Versions[j].LastPlayed
	})
	sort.Slice(stats.WeeklyVersions, func(i, j int) bool {
		return stats.WeeklyVersions[i] < stats.WeeklyVersions[j]
	})

	return stats
}
