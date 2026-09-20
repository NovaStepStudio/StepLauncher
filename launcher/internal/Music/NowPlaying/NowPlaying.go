package nowplaying

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Queue representa la cola temporal de reproducción.
type Queue struct {
	Tracks       []string `json:"tracks"`
	CurrentIndex int      `json:"currentIndex"`
	CurrentPath  string   `json:"currentPath"`
}

type Manager struct {
	mu   sync.RWMutex
	path string
	q    Queue
}

func NewManager(rootDir string) *Manager {
	p := filepath.Join(rootDir, "launcher_music_nowplaying.json")
	m := &Manager{path: p, q: Queue{Tracks: []string{}, CurrentIndex: -1}}
	m.load()
	return m
}

func (m *Manager) load() {
	data, err := os.ReadFile(m.path)
	if err != nil {
		return
	}
	var q Queue
	if err := json.Unmarshal(data, &q); err == nil {
		m.q = q
	}
}

func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.q, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
		return err
	}
	tmp := m.path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, m.path)
}

func (m *Manager) Get() Queue {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cp := Queue{Tracks: append([]string{}, m.q.Tracks...), CurrentIndex: m.q.CurrentIndex, CurrentPath: m.q.CurrentPath}
	return cp
}

func (m *Manager) Set(tracks []string, idx int, curPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.q.Tracks = append([]string{}, tracks...)
	m.q.CurrentIndex = idx
	m.q.CurrentPath = curPath
	return m.save()
}

func (m *Manager) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.q = Queue{Tracks: []string{}, CurrentIndex: -1}
	return m.save()
}
