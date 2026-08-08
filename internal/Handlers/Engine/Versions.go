package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"StepLauncher/internal/Core/Downloader"
)

type VersionInfo struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	ReleaseTime string `json:"time"`
}

var validTypes = map[string]bool{
	"release": true, "snapshot": true, "old_beta": true, "old_alpha": true,
}

type versionManifest struct {
	Versions []struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		URL         string `json:"url"`
		ReleaseTime string `json:"releaseTime"`
	} `json:"versions"`
}

type InstalledVersion struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (e *Engine) ListDownloadedVersions() []InstalledVersion {
	versionsDir := filepath.Join(e.config.Get().WorkDir, "versions")
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return []InstalledVersion{}
	}

	out := make([]InstalledVersion, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id := entry.Name()
		if id == "" {
			continue
		}
		var meta struct {
			Type string `json:"type"`
		}
		verPath := filepath.Join(versionsDir, id, id+".json")
		if data, err := os.ReadFile(verPath); err == nil {
			json.Unmarshal(data, &meta)
		} else {
			continue
		}
		if meta.Type == "" {
			meta.Type = "release"
		}
		out = append(out, InstalledVersion{ID: id, Type: meta.Type})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

func (e *Engine) GetVersions(versionType string) ([]VersionInfo, error) {
	if !validTypes[versionType] {
		return nil, fmt.Errorf("invalid version type '%s'. valid: release, snapshot, old_beta, old_alpha", versionType)
	}

	cacheKey := fmt.Sprintf("manifest-%s", versionType)
	var cached []VersionInfo
	found, _ := e.cache.Get("manifest", cacheKey, &cached)
	if found {
		return cached, nil
	}

	client := e.downloader.HTTPClient()
	resp, err := client.Get("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	var m versionManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}

	var result []VersionInfo
	for _, v := range m.Versions {
		if v.Type == versionType {
			result = append(result, VersionInfo{
				ID: v.ID, Type: v.Type, URL: v.URL, ReleaseTime: v.ReleaseTime,
			})
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no versions found for type '%s'", versionType)
	}

	e.cache.Set("manifest", cacheKey, result)
	return result, nil
}

func (e *Engine) RefreshManifests() (int, error) {
	before := e.cache.Info().Categories["manifest"]
	e.cache.DeleteCategory("manifest")
	m, err := e.FetchVersionManifest()
	if err != nil {
		return 0, err
	}
	e.log.Info("[Cache] Manifiestos refrescados: %d borrados, %d versiones disponibles", before, len(m.Versions))
	return len(m.Versions), nil
}

func (e *Engine) FetchVersionManifest() (*downloader.Manifest, error) {
	const cacheKey = "full"
	var m downloader.Manifest
	found, expired, err := e.cache.GetWithFallback("manifest", cacheKey, &m)
	if err == nil && found && !expired {
		return &m, nil
	}

	client := e.downloader.HTTPClient()
	resp, err := client.Get("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json")
	if err != nil {
		if found && expired {
			e.log.Info("[Cache] WARN: usando manifest expirado (red no disponible: %v)", err)
			return &m, nil
		}
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if found && expired && (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500) {
			e.log.Info("[Cache] WARN: usando manifest expirado (HTTP %d)", resp.StatusCode)
			return &m, nil
		}
		return nil, fmt.Errorf("fetch manifest: HTTP %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}

	if err := e.cache.Set("manifest", cacheKey, m); err != nil {
		e.log.Info("[Cache] WARN: no se pudo guardar el manifest: %v", err)
	}
	return &m, nil
}
