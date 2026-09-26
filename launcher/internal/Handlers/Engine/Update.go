package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	engineconfig "StepLauncher/internal/Handlers/Engine/engineconfig"

	"StepLauncher/internal/Core/Downloader"
)

const (
	updateRepo = "NovaStepStudio/StepLauncher"
	// GitHub directo: la API pública de releases no necesita token y
	// devuelve el mismo JSON que consumía el Worker (tag_name, assets...).
	updateGitHubReleases = "https://api.github.com/repos/" + updateRepo + "/releases"
	// Carpeta temporal donde se descarga el instalador en Windows.
	updaterTempDir = "StepLauncher-Updater"
	// Nombre de respaldo si la URL del instalador no trae nombre usable.
	updaterFallbackName = "steplauncher-installer.exe"
	updateMaxBodySize    = 4 * 1024 * 1024
)

// Cliente directo del launcher (sin proxy del sistema ni HTTP/2).
var updateHTTPClient = &http.Client{Transport: downloader.DefaultTransport.Clone(), Timeout: 25 * time.Second}

type githubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	PublishedAt string `json:"published_at"`
	HTMLURL     string `json:"html_url"`
	Body        string `json:"body"`
	Prerelease  bool   `json:"prerelease"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

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

	// GitHub directo: /releases devuelve las publicadas (incluye
	// prereleases, que se filtran abajo para seguir el canal estable).
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, updateGitHubReleases, nil)
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

	// GitHub devuelve un array de releases; busca la más nueva > current.
	var releases []githubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		// Fallback: si algún día se consulta /latest (objeto único), soportarlo
		var single githubRelease
		if err2 := json.Unmarshal(body, &single); err2 == nil && single.TagName != "" {
			releases = []githubRelease{single}
		} else {
			info.Error = "respuesta de GitHub inválida: " + err.Error()
			return info
		}
	}

	if len(releases) == 0 {
		info.Error = "no hay releases disponibles"
		return info
	}

	// Encuentra la release estable más nueva que sea > currentVersion.
	// Las prereleases se ignoran (canal estable).
	var best *githubRelease
	bestVer := ""
	for i := range releases {
		r := &releases[i]
		if r.Prerelease {
			continue
		}
		ver := strings.TrimPrefix(strings.TrimSpace(r.TagName), "v")
		if ver == "" {
			continue
		}
		if compareVersions(ver, info.CurrentVersion) <= 0 {
			continue
		}
		if best == nil || compareVersions(ver, bestVer) > 0 {
			best = r
			bestVer = ver
		}
	}
	if best == nil {
		// No hay versión más nueva -> up-to-date, usa la current como latest para mostrar
		info.LatestVersion = info.CurrentVersion
		return info
	}

	rel := best
	info.LatestVersion = strings.TrimPrefix(strings.TrimSpace(rel.TagName), "v")
	info.ReleaseURL = rel.HTMLURL
	if info.ReleaseURL == "" {
		info.ReleaseURL = "https://github.com/" + updateRepo + "/releases/latest"
	}
	info.ReleaseName = rel.Name
	info.ReleaseDate = rel.PublishedAt
	info.Notes = strings.TrimSpace(rel.Body)

	// En Windows la actualización se aplica con el instalador NSIS de la
	// release (*-installer.exe de la arquitectura en curso): se descarga,
	// se ejecuta y se cierra el launcher para que instale limpio.
	// En Linux/macOS no hay instalador/actualizador: el diálogo invita a
	// instalar manualmente desde la página de la release.
	if runtime.GOOS == "windows" {
		for _, a := range rel.Assets {
			if isWindowsInstallerAsset(a.Name, runtime.GOARCH) && a.BrowserDownloadURL != "" {
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

// isWindowsInstallerAsset indica si name es el instalador NSIS de Windows
// para la arquitectura indicada (p. ej. steplauncher-v2.5.0-windows-amd64-installer.exe).
func isWindowsInstallerAsset(name, arch string) bool {
	lower := strings.ToLower(name)
	if !strings.HasSuffix(lower, ".exe") || !strings.Contains(lower, "installer") {
		return false
	}
	if !strings.Contains(lower, "windows") && !strings.Contains(lower, "win") {
		return false
	}
	switch strings.ToLower(arch) {
	case "amd64", "x86_64", "x64":
		return strings.Contains(lower, "amd64") || strings.Contains(lower, "x86_64") || strings.Contains(lower, "x64")
	case "arm64", "aarch64":
		return strings.Contains(lower, "arm64") || strings.Contains(lower, "aarch64")
	default:
		return arch == "" || strings.Contains(lower, strings.ToLower(arch))
	}
}

func (e *Engine) DownloadUpdater(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url del actualizador vacía")
	}

	dir := filepath.Join(os.TempDir(), updaterTempDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("crear carpeta temporal: %w", err)
	}
	// Guarda el instalador con su nombre real de la release para no chocar
	// con descargas de otras versiones.
	destName := installerFileName(url)
	dest := filepath.Join(dir, destName)

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
		return "", fmt.Errorf("descargar instalador: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("descargar instalador: GitHub respondió %s", resp.Status)
	}
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return "", fmt.Errorf("descargar instalador: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	if err := os.Rename(tmpPath, dest); err != nil {
		os.Remove(dest)
		if err2 := os.Rename(tmpPath, dest); err2 != nil {
			return "", fmt.Errorf("mover instalador: %w", err2)
		}
	}
	return dest, nil
}

// installerFileName extrae un nombre de archivo seguro desde la URL del
// asset de GitHub (p. ej. steplauncher-v2.5.0-windows-amd64-installer.exe).
func installerFileName(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed == nil {
		return updaterFallbackName
	}
	base := path.Base(strings.TrimSpace(parsed.Path))
	if base == "" || base == "." || base == "/" || strings.ContainsAny(base, `/\:`) {
		return updaterFallbackName
	}
	return base
}

func (e *Engine) LaunchUpdater(updaterPath string) error {
	if updaterPath == "" {
		return fmt.Errorf("ruta del actualizador vacía")
	}
	return launchUpdater(updaterPath)
}
