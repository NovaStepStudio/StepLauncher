package engine
import (
	"encoding/json"
	"fmt"

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

func (e *Engine) FetchVersionManifest() (*downloader.Manifest, error) {
	var m downloader.Manifest
	client := e.downloader.HTTPClient()
	resp, err := client.Get("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json")
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	return &m, nil
}
