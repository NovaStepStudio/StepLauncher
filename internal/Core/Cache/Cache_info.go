package cache

import (
	"os"
	"path/filepath"
	"strings"
)

type Info struct {
	CacheDir     string            `json:"cacheDir"`
	TotalEntries int               `json:"totalEntries"`
	TotalBytes   int64             `json:"totalBytes"`
	Categories   map[string]int    `json:"categories"`
	Sizes        map[string]int64  `json:"sizes"`
	TTLs         map[string]string `json:"ttls"`
}

func (m *Manager) Info() Info {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info := Info{
		CacheDir:   m.cacheDir,
		Categories: make(map[string]int),
		Sizes:      make(map[string]int64),
		TTLs: map[string]string{
			"manifest":  m.ttls.Manifest.String(),
			"assets":    m.ttls.Assets.String(),
			"versions":  m.ttls.Versions.String(),
			"modloader": m.ttls.Modloader.String(),
			"java":      m.ttls.Java.String(),
			"default":   m.ttls.Default.String(),
		},
	}

	seen := map[string]bool{}
	for _, sub := range append(subdirs(), artifactDirs()...) {
		if seen[sub] {
			continue
		}
		seen[sub] = true

		dir := filepath.Join(m.cacheDir, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		count := 0
		var size int64
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if !isArtifactDir(sub) && !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			count++
			size += info.Size()
		}
		if count > 0 {
			info.Categories[sub] = count
			info.Sizes[sub] = size
			info.TotalEntries += count
			info.TotalBytes += size
		}
	}

	return info
}
