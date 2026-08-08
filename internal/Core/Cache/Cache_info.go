package cache

import (
	"os"
	"path/filepath"
	"strings"
)

type Info struct {
	CacheDir     string            `json:"cacheDir"`
	TotalEntries int               `json:"totalEntries"`
	Categories   map[string]int    `json:"categories"`
	TTLs         map[string]string `json:"ttls"`
}

func (m *Manager) Info() Info {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info := Info{
		CacheDir:   m.cacheDir,
		Categories: make(map[string]int),
		TTLs: map[string]string{
			"manifest":  m.ttls.Manifest.String(),
			"assets":    m.ttls.Assets.String(),
			"versions":  m.ttls.Versions.String(),
			"modloader": m.ttls.Modloader.String(),
			"java":      m.ttls.Java.String(),
			"default":   m.ttls.Default.String(),
		},
	}

	for _, sub := range subdirs() {
		dir := filepath.Join(m.cacheDir, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		count := 0
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				count++
			}
		}
		if count > 0 {
			info.Categories[sub] = count
			info.TotalEntries += count
		}
	}

	return info
}
