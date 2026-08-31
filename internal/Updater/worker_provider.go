package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	wailsupdater "github.com/wailsapp/wails/v3/pkg/updater"
	"golang.org/x/mod/semver"
)

// WorkerProvider implementa wailsupdater.Provider contra el Worker de Cloudflare
// que proxy-cachea releases de GitHub. Usa las dos URLs de la API del launcher:
//
//   - Releases:    https://steplauncher.stepnicka012.workers.dev/updates/steplauncher/releases
//   - Prereleases: https://steplauncher.stepnicka012.workers.dev/updates/steplauncher/prereleases
//
// La respuesta es un JSON array con objetos tipo GitHub Release (tag_name, name, body, assets).
type WorkerProvider struct {
	ReleasesURL    string
	PrereleasesURL string
	Channel        string
	HTTPClient     *http.Client
}

type workerRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	PublishedAt string `json:"published_at"`
	HTMLURL     string `json:"html_url"`
	Body        string `json:"body"`
	Prerelease  bool   `json:"prerelease"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
		ContentType        string `json:"content_type"`
	} `json:"assets"`
}

func NewWorker(cfg WorkerProvider) (*WorkerProvider, error) {
	if cfg.ReleasesURL == "" {
		cfg.ReleasesURL = "https://steplauncher.stepnicka012.workers.dev/updates/steplauncher/releases"
	}
	if cfg.PrereleasesURL == "" {
		cfg.PrereleasesURL = "https://steplauncher.stepnicka012.workers.dev/updates/steplauncher/prereleases"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &WorkerProvider{
		ReleasesURL:    cfg.ReleasesURL,
		PrereleasesURL: cfg.PrereleasesURL,
		Channel:        cfg.Channel,
		HTTPClient:     cfg.HTTPClient,
	}, nil
}

func (p *WorkerProvider) Name() string { return "worker" }

func (p *WorkerProvider) Check(ctx context.Context, req wailsupdater.CheckRequest) (*wailsupdater.Release, error) {
	if req.CurrentVersion == "" {
		return nil, fmt.Errorf("worker: CurrentVersion requerido")
	}
	url := p.ReleasesURL
	// Si el canal pide prerelease, usa el endpoint de prereleases
	ch := strings.ToLower(strings.TrimSpace(p.Channel))
	if ch == "prerelease" || ch == "pre" || ch == "beta" || ch == "alpha" || ch == "rc" {
		url = p.PrereleasesURL
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "StepLauncher/"+req.CurrentVersion)

	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "no such host") || strings.Contains(msg, "dial tcp") || strings.Contains(msg, "network is unreachable") || strings.Contains(msg, "no internet") {
			// Sin internet no es un error de actualización, solo no hay conexión
			return nil, nil
		}
		return nil, fmt.Errorf("worker: fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("worker: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil {
		return nil, err
	}
	var releases []workerRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		// Algunos workers devuelven objeto único si pidio latest; soporta ambos
		var single workerRelease
		if err2 := json.Unmarshal(body, &single); err2 == nil && single.TagName != "" {
			releases = []workerRelease{single}
		} else {
			return nil, fmt.Errorf("worker: decode: %w", err)
		}
	}
	if len(releases) == 0 {
		return nil, nil
	}

	cur := normalizeVersion(req.CurrentVersion)
	// Buscar la release más nueva > current
	var best *workerRelease
	var bestVer string
	for i := range releases {
		r := &releases[i]
		ver := normalizeVersion(strings.TrimSpace(strings.TrimPrefix(r.TagName, "v")))
		if ver == "" {
			continue
		}
		if !semver.IsValid("v" + ver) {
			// fallback a compare manual si no es semver válido
			if compareVersions(ver, strings.TrimPrefix(cur, "v")) <= 0 {
				continue
			}
		} else {
			if semver.Compare("v"+ver, cur) <= 0 {
				continue
			}
		}
		if best == nil {
			best = r
			bestVer = ver
			continue
		}
		// compara entre best y r, queda el mayor
		if semver.IsValid("v"+ver) && semver.IsValid("v"+bestVer) {
			if semver.Compare("v"+ver, "v"+bestVer) > 0 {
				best = r
				bestVer = ver
			}
		} else {
			if compareVersions(ver, bestVer) > 0 {
				best = r
				bestVer = ver
			}
		}
	}
	if best == nil {
		return nil, nil // up-to-date
	}

	// Selección de asset por plataforma
	idx := pickWorkerAsset(req.Platform, req.Arch, best.Assets)
	if idx == -1 {
		return nil, fmt.Errorf("worker: release %s no tiene asset para %s/%s", bestVer, req.Platform, req.Arch)
	}
	asset := best.Assets[idx]
	// Parse time
	var publishedAt time.Time
	if best.PublishedAt != "" {
		publishedAt, _ = time.Parse(time.RFC3339, best.PublishedAt)
	}

	rel := &wailsupdater.Release{
		Version:     bestVer,
		Name:        best.Name,
		Notes:       strings.TrimSpace(best.Body),
		PublishedAt: publishedAt,
		Artifact: wailsupdater.Artifact{
			Filename: asset.Name,
			Filetype: filetypeFromName(asset.Name),
			Size:     asset.Size,
			Platform: req.Platform,
			Arch:     req.Arch,
		},
		Metadata: map[string]any{
			"worker.download.url": asset.BrowserDownloadURL,
			"worker.html_url":     best.HTMLURL,
			"worker.tag":          best.TagName,
		},
	}
	return rel, nil
}

func (p *WorkerProvider) Download(ctx context.Context, rel *wailsupdater.Release, dst io.Writer, onProgress func(written, total int64)) error {
	if rel == nil || rel.Metadata == nil {
		return fmt.Errorf("worker: release sin metadata")
	}
	urlStr, ok := rel.Metadata["worker.download.url"].(string)
	if !ok || urlStr == "" {
		return fmt.Errorf("worker: release sin URL de descarga")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "StepLauncher")
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("worker: download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("worker: download HTTP %d", resp.StatusCode)
	}
	total := resp.ContentLength
	if total <= 0 {
		total = rel.Artifact.Size
	}
	written := int64(0)
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return werr
			}
			written += int64(n)
			if onProgress != nil {
				onProgress(written, total)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}
	return nil
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	// semver requiere v prefix; lo añadiremos en Compare, aquí solo limpia
	if v == "" {
		return ""
	}
	// Si viene con sufijo tipo "2.3.1-beta", semver lo maneja
	return "v" + v
}

func filetypeFromName(name string) string {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, ".zip") {
		return "zip"
	}
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return "tar.gz"
	}
	if strings.HasSuffix(lower, ".exe") {
		return "exe"
	}
	return ""
}

// pickWorkerAsset replica la lógica del github provider con fallback Steps.
// Primero intenta match por GOOS+GOARCH, luego fallback genérico steplauncher.
func pickWorkerAsset(platform, arch string, assets []struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	ContentType        string `json:"content_type"`
}) int {
	if len(assets) == 0 {
		return -1
	}
	plat := strings.ToLower(platform)
	arc := strings.ToLower(arch)

	// Helper para saber si un asset corresponde a platform/arch
	matches := func(name string) bool {
		lower := strings.ToLower(name)
		// excluye installers
		if lower == "installer.exe" || strings.Contains(lower, "-installer.") || strings.Contains(lower, "_installer.") {
			return false
		}
		hasPlat := plat == "" || platformMatches(lower, plat)
		hasArch := arc == "" || archMatches(lower, arc)
		return hasPlat && hasArch
	}

	// 1) Intenta match estricto
	for i, a := range assets {
		if matches(a.Name) {
			return i
		}
	}
	// 2) Si solo hay un asset y no es installer, tómale
	if len(assets) == 1 {
		lower := strings.ToLower(assets[0].Name)
		if lower != "installer.exe" && !strings.Contains(lower, "-installer.") && !strings.Contains(lower, "_installer.") {
			return 0
		}
	}
	// 3) Fallback genérico steplauncher
	for i, a := range assets {
		lower := strings.ToLower(a.Name)
		if lower == "installer.exe" || strings.Contains(lower, "-installer.") || strings.Contains(lower, "_installer.") {
			continue
		}
		if strings.Contains(lower, "steplauncher") {
			return i
		}
	}
	return -1
}

func platformMatches(name, plat string) bool {
	switch plat {
	case "windows", "win":
		return strings.Contains(name, "windows") || strings.Contains(name, "win") || strings.Contains(name, ".exe")
	case "darwin", "macos", "mac":
		return strings.Contains(name, "darwin") || strings.Contains(name, "macos") || strings.Contains(name, "mac")
	case "linux":
		return strings.Contains(name, "linux")
	default:
		return strings.Contains(name, plat)
	}
}

func archMatches(name, arch string) bool {
	switch arch {
	case "amd64", "x86_64", "x64":
		return strings.Contains(name, "amd64") || strings.Contains(name, "x86_64") || strings.Contains(name, "x64") || strings.Contains(name, "x86-64")
	case "arm64", "aarch64":
		return strings.Contains(name, "arm64") || strings.Contains(name, "aarch64")
	case "386", "i386", "x86", "ia32":
		return strings.Contains(name, "386") || strings.Contains(name, "i386") || strings.Contains(name, "x86") || strings.Contains(name, "ia32")
	default:
		return arch == "" || strings.Contains(name, arch)
	}
}

func compareVersions(a, b string) int {
	pa := strings.Split(strings.TrimPrefix(a, "v"), ".")
	pb := strings.Split(strings.TrimPrefix(b, "v"), ".")
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
		// intenta numérico
		var xn, yn int
		_, xerr := fmt.Sscan(x, &xn)
		_, yerr := fmt.Sscan(y, &yn)
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
