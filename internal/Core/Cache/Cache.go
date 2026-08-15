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

// subdirs son las categorías de cache que almacenan JSON con metadatos.
// No se crean al arrancar: Set() crea únicamente la carpeta de la categoría
// que realmente se escribe, cuando se escribe.
func subdirs() []string {
	return []string{"default", "manifest", "versions", "assets", "java"}
}

// obsoleteDirs son carpetas que versiones antiguas del launcher creaban al
// arrancar pero que nunca se usan como almacenamiento (no guardan nada). Se
// eliminan del disco si quedaron vacías.
func obsoleteDirs() []string {
	return []string{"forge", "neoforge", "fabric", "quilt", "legacyfabric", "assets/indexes", "assets/manifests"}
}

// artifactDirs son carpetas de cache que no almacenan JSON con metadatos
// (instaladores de modloaders y sus logs) y se limpian por antigüedad.
func artifactDirs() []string {
	return []string{"modloader", "modloader-logs"}
}

func (m *Manager) keyToPath(category, key string) string {
	safe := strings.ReplaceAll(key, "/", "_")
	safe = strings.ReplaceAll(safe, ":", "_")
	return filepath.Join(m.cacheDir, category, safe+".json")
}

func (m *Manager) Set(category, key string, data interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

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
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, metaRaw, 0644); err != nil {
		return fmt.Errorf("cache write: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cache rename: %w", err)
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
	for _, sub := range artifactDirs() {
		dir := filepath.Join(m.cacheDir, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
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
		if e.IsDir() {
			continue
		}
		if isArtifactDir(category) || strings.HasSuffix(e.Name(), ".json") {
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
	m.cleanup()
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

	for _, sub := range artifactDirs() {
		dir := filepath.Join(m.cacheDir, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		ttl := m.ttlFor(sub)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			if now.Sub(info.ModTime()) > ttl {
				os.Remove(filepath.Join(dir, e.Name()))
			}
		}
	}

	// Carpetas obsoletas que quedaron de versiones antiguas: se eliminan solo
	// si están vacías (nunca se escribió nada en ellas).
	for _, sub := range obsoleteDirs() {
		dir := filepath.Join(m.cacheDir, sub)
		if isEmptyDir(dir) {
			os.Remove(dir)
		}
	}
}

func isEmptyDir(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return len(entries) == 0
}

func isArtifactDir(category string) bool {
	for _, a := range artifactDirs() {
		if category == a {
			return true
		}
	}
	return false
}
