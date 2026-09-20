package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const CrashFileName = "launcher_history_crashes.json"

type CrashEntry struct {
	ID            string `json:"id"`
	Timestamp     int64  `json:"timestamp"`
	Version       string `json:"version"`
	ExitCode      int    `json:"exit_code"`
	CrashReason   string `json:"crash_reason,omitempty"`
	CrashCategory string `json:"crash_category,omitempty"`
	LauncherLogPath  string `json:"launcherLogPath,omitempty"`
	MinecraftLogPath string `json:"minecraftLogPath,omitempty"`
	JvmLogPath       string `json:"jvmLogPath,omitempty"`
}

type CrashFileData struct {
	Entries []CrashEntry `json:"entries"`
}

type CrashManager struct {
	mu       sync.RWMutex
	filePath string
	data     CrashFileData
}

func NewCrashManager(workDir string) *CrashManager {
	return &CrashManager{
		filePath: filepath.Join(workDir, CrashFileName),
		data: CrashFileData{
			Entries: make([]CrashEntry, 0),
		},
	}
}

func (m *CrashManager) Load() error {
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
		return fmt.Errorf("failed to parse crash history file: %w", err)
	}

	return nil
}

func (m *CrashManager) save() error {
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

func (m *CrashManager) EnsureFile() error {
	m.mu.Lock()
	if _, err := os.Stat(m.filePath); err == nil {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()
	return m.save()
}

func (m *CrashManager) AddEntry(entry CrashEntry) error {
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

func (m *CrashManager) GetEntries() []CrashEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]CrashEntry, len(m.data.Entries))
	copy(res, m.data.Entries)
	return res
}

func (m *CrashManager) Clear() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := len(m.data.Entries)
	m.data.Entries = make([]CrashEntry, 0)
	if err := m.save(); err != nil {
		return 0
	}
	return count
}
