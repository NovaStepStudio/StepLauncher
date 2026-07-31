package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type entryMeta struct {
	CreatedAt time.Time       `json:"created_at"`
	ExpiresAt time.Time       `json:"expires_at"`
	Data      json.RawMessage `json:"data"`
}

type Manager struct {
	mu       sync.RWMutex
	cacheDir string
	ttls     CacheTTL
}

func NewManager(cacheDir string, ttls CacheTTL) *Manager {
	m := &Manager{
		cacheDir: cacheDir,
		ttls:     ttls,
	}
	go m.cleanupLoop()
	return m
}

func subdirs() []string {
	return []string{"versions", "assets", "java", "forge", "neoforge", "fabric", "quilt", "legacyfabric", "assets/indexes", "assets/manifests"}
}

func (m *Manager) ensureDirs() {
	for _, sub := range subdirs() {
		os.MkdirAll(filepath.Join(m.cacheDir, sub), 0755)
	}
}

func (m *Manager) keyToPath(category, key string) string {
	safe := strings.ReplaceAll(key, "/", "_")
	safe = strings.ReplaceAll(safe, ":", "_")
	return filepath.Join(m.cacheDir, category, safe+".json")
}

func (m *Manager) Set(category, key string, data interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureDirs()

	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}

	meta := entryMeta{
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(m.ttlFor(category)),
		Data:      raw,
	}

	metaRaw, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("cache meta marshal: %w", err)
	}

	path := m.keyToPath(category, key)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("cache mkdir: %w", err)
	}
	if err := os.WriteFile(path, metaRaw, 0644); err != nil {
		return fmt.Errorf("cache write: %w", err)
	}
	return nil
}

func (m *Manager) Get(category, key string, out interface{}) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	path := m.keyToPath(category, key)
	metaRaw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("cache read: %w", err)
	}

	var meta entryMeta
	if err := json.Unmarshal(metaRaw, &meta); err != nil {
		return false, fmt.Errorf("cache meta unmarshal: %w", err)
	}

	if time.Now().After(meta.ExpiresAt) {
		return false, nil
	}

	if err := json.Unmarshal(meta.Data, out); err != nil {
		return false, fmt.Errorf("cache data unmarshal: %w", err)
	}
	return true, nil
}

func (m *Manager) GetWithFallback(category, key string, out interface{}) (bool, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	path := m.keyToPath(category, key)
	metaRaw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, fmt.Errorf("cache read: %w", err)
	}

	var meta entryMeta
	if err := json.Unmarshal(metaRaw, &meta); err != nil {
		return false, false, fmt.Errorf("cache meta unmarshal: %w", err)
	}

	expired := time.Now().After(meta.ExpiresAt)

	if err := json.Unmarshal(meta.Data, out); err != nil {
		return false, false, fmt.Errorf("cache data unmarshal: %w", err)
	}

	return true, expired, nil
}

func (m *Manager) Delete(category, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.keyToPath(category, key)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cache delete: %w", err)
	}
	return nil
}

func (m *Manager) Clear() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, sub := range subdirs() {
		dir := filepath.Join(m.cacheDir, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				os.Remove(filepath.Join(dir, e.Name()))
				count++
			}
		}
	}
	return count
}

func (m *Manager) DeleteCategory(category string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	dir := filepath.Join(m.cacheDir, category)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			os.Remove(filepath.Join(dir, e.Name()))
			count++
		}
	}
	return count
}

func (m *Manager) Refresh(category, key string, fetcher func() (interface{}, error)) error {
	data, err := fetcher()
	if err != nil {
		return err
	}
	return m.Set(category, key, data)
}

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		m.cleanup()
	}
}

func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for _, sub := range subdirs() {
		dir := filepath.Join(m.cacheDir, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			path := filepath.Join(dir, e.Name())
			metaRaw, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var meta entryMeta
			if err := json.Unmarshal(metaRaw, &meta); err != nil {
				continue
			}
			if now.After(meta.ExpiresAt) {
				os.Remove(path)
			}
		}
	}
}


