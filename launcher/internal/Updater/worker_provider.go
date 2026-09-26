package worker

import (
	"net/http"
	"strings"
	"time"

	"StepLauncher/internal/Core/Downloader"

	"github.com/wailsapp/wails/v3/pkg/updater"
	ghprovider "github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

// Este paquete configura el actualizador de Wails apuntando DIRECTO a
// GitHub Releases del repositorio, sin APIs externas intermedias.
// La verificación usa el sidecar CHECKSUMS-SHA256.txt que publica el
// workflow de release junto a los binarios.

// NewGithubProvider crea el provider oficial de GitHub Releases contra
// el repo indicado ("owner/repo"). En Windows selecciona el instalador
// NSIS (*-installer.exe) de la arquitectura en curso; en el resto de
// plataformas delega al matcher por defecto (binario portable).
func NewGithubProvider(repo string) (*ghprovider.Provider, error) {
	return ghprovider.New(ghprovider.Config{
		Repository:    repo,
		ChecksumAsset: "CHECKSUMS-SHA256.txt",
		AssetMatcher:  MatchReleaseAsset,
		HTTPClient:    DirectHTTPClient(),
	})
}

// DirectHTTPClient devuelve el cliente directo del launcher (sin proxy
// del sistema ni HTTP/2): el proxy de Ajustes > Red es SOLO para el
// juego de Minecraft y nunca debe afectar al actualizador.
func DirectHTTPClient() *http.Client {
	return &http.Client{Transport: downloader.DefaultTransport.Clone(), Timeout: 20 * time.Second}
}

// MatchReleaseAsset elige el asset de la release para la plataforma en curso.
// En Windows devuelve el instalador NSIS de la arquitectura pedida, porque
// la actualización se aplica ejecutando el instalador con el launcher
// cerrado (no con swap del binario en caliente).
func MatchReleaseAsset(req updater.CheckRequest, assets []ghprovider.ReleaseAsset) int {
	if strings.EqualFold(req.Platform, "windows") {
		for i, a := range assets {
			if IsWindowsInstaller(a.Name, req.Arch) {
				return i
			}
		}
		return -1
	}
	return ghprovider.DefaultAssetMatcher(req, assets)
}

// IsWindowsInstaller indica si name es el instalador NSIS de Windows para
// la arquitectura indicada (p. ej. steplauncher-v2.5.0-windows-amd64-installer.exe).
func IsWindowsInstaller(name, arch string) bool {
	lower := strings.ToLower(name)
	if !strings.HasSuffix(lower, ".exe") || !strings.Contains(lower, "installer") {
		return false
	}
	if !strings.Contains(lower, "windows") && !strings.Contains(lower, "win") {
		return false
	}
	return archMatches(lower, arch)
}

func archMatches(name, arch string) bool {
	switch strings.ToLower(arch) {
	case "amd64", "x86_64", "x64":
		return strings.Contains(name, "amd64") || strings.Contains(name, "x86_64") || strings.Contains(name, "x64")
	case "arm64", "aarch64":
		return strings.Contains(name, "arm64") || strings.Contains(name, "aarch64")
	default:
		return arch == "" || strings.Contains(name, strings.ToLower(arch))
	}
}
