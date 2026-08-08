package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	engineconfig "StepLauncher/internal/Handlers/Engine/engineconfig"
)

const (
	updateRepo        = "NovaStepStudio/StepLauncher"
	updateAPILatest   = "https://api.github.com/repos/" + updateRepo + "/releases/latest"
	updaterAssetName  = "StepLauncher-Updater.exe"
	updaterTempDir    = "StepLauncher-Updater"
	updateMaxBodySize = 4 * 1024 * 1024
)

var updateHTTPClient = &http.Client{Timeout: 25 * time.Second}

type UpdateInfo struct {
	HasUpdate      bool   `json:"hasUpdate"`
	LatestVersion  string `json:"latestVersion"`
	CurrentVersion string `json:"currentVersion"`
	ReleaseURL     string `json:"releaseUrl"`
	ReleaseName    string `json:"releaseName"`
	ReleaseDate    string `json:"releaseDate"`
	Notes          string `json:"notes"`
	HasUpdater     bool   `json:"hasUpdater"`
	UpdaterURL     string `json:"updaterUrl"`
	Platform       string `json:"platform"`
	Error          string `json:"error"`
}

func (e *Engine) CheckForUpdates() {
	go func() {
		info := e.checkUpdate()

		e.updateMu.Lock()
		e.lastUpdate = info
		e.updateMu.Unlock()

		data, err := json.Marshal(info)
		if err != nil {
			return
		}
		if e.eventCb != nil {
			e.eventCb("update_check", data)
		}
	}()
}

func (e *Engine) LastUpdateInfo() *UpdateInfo {
	e.updateMu.Lock()
	defer e.updateMu.Unlock()
	return e.lastUpdate
}

func (e *Engine) checkUpdate() *UpdateInfo {
	info := &UpdateInfo{
		CurrentVersion: engineconfig.AppVersion,
		Platform:       runtime.GOOS,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, updateAPILatest, nil)
	if err != nil {
		info.Error = err.Error()
		return info
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "StepLauncher/"+engineconfig.AppVersion)

	resp, err := updateHTTPClient.Do(req)
	if err != nil {
		info.Error = "no se pudo conectar con GitHub: " + err.Error()
		return info
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		info.Error = fmt.Sprintf("GitHub respondió %s", resp.Status)
		return info
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, updateMaxBodySize))
	if err != nil {
		info.Error = "no se pudo leer la respuesta de GitHub: " + err.Error()
		return info
	}

	var rel struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		PublishedAt string `json:"published_at"`
		HTMLURL     string `json:"html_url"`
		Body        string `json:"body"`
		Assets      []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(body, &rel); err != nil {
		info.Error = "respuesta de GitHub inválida: " + err.Error()
		return info
	}

	info.LatestVersion = strings.TrimPrefix(strings.TrimSpace(rel.TagName), "v")
	info.ReleaseURL = rel.HTMLURL
	if info.ReleaseURL == "" {
		info.ReleaseURL = "https://github.com/" + updateRepo + "/releases/latest"
	}
	info.ReleaseName = rel.Name
	info.ReleaseDate = rel.PublishedAt
	info.Notes = strings.TrimSpace(rel.Body)

	if runtime.GOOS == "windows" {
		for _, a := range rel.Assets {
			if strings.EqualFold(a.Name, updaterAssetName) && a.BrowserDownloadURL != "" {
				info.HasUpdater = true
				info.UpdaterURL = a.BrowserDownloadURL
				break
			}
		}
	}

	if info.LatestVersion == "" {
		info.Error = "la release no trae versión"
		return info
	}
	if compareVersions(info.LatestVersion, info.CurrentVersion) > 0 {
		info.HasUpdate = true
	}
	return info
}

func compareVersions(a, b string) int {
	pa := strings.Split(strings.TrimPrefix(strings.TrimSpace(a), "v"), ".")
	pb := strings.Split(strings.TrimPrefix(strings.TrimSpace(b), "v"), ".")
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y string
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		if x == y {
			continue
		}
		xn, xerr := strconv.Atoi(x)
		yn, yerr := strconv.Atoi(y)
		if xerr == nil && yerr == nil {
			if xn < yn {
				return -1
			}
			return 1
		}
		if x == "" {
			return -1
		}
		if y == "" {
			return 1
		}
		if x < y {
			return -1
		}
		return 1
	}
	return 0
}

func (e *Engine) DownloadUpdater(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url del actualizador vacía")
	}

	dir := filepath.Join(os.TempDir(), updaterTempDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("crear carpeta temporal: %w", err)
	}
	dest := filepath.Join(dir, updaterAssetName)

	tmp, err := os.CreateTemp(dir, "updater-*.tmp")
	if err != nil {
		return "", fmt.Errorf("crear archivo temporal: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpPath)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "StepLauncher/"+engineconfig.AppVersion)

	resp, err := updateHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("descargar actualizador: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("descargar actualizador: GitHub respondió %s", resp.Status)
	}
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return "", fmt.Errorf("descargar actualizador: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	if err := os.Rename(tmpPath, dest); err != nil {
		os.Remove(dest)
		if err2 := os.Rename(tmpPath, dest); err2 != nil {
			return "", fmt.Errorf("mover actualizador: %w", err2)
		}
	}
	return dest, nil
}

func (e *Engine) LaunchUpdater(updaterPath string) error {
	if updaterPath == "" {
		return fmt.Errorf("ruta del actualizador vacía")
	}
	return launchUpdater(updaterPath)
}
