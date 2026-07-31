// Package Handlers expone las APIs del launcher al frontend (bindings Wails).
// Solo delegacion: config gestionada por internal/Config y motor por internal/Core.
package Handlers

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"StepLauncher/internal/Config"
	engine "StepLauncher/internal/Handlers/Engine"
)

type App struct {
	ctx    context.Context
	engine *engine.Engine
	config *Config.Manager
}

func NewApp(eng *engine.Engine, configPath string) *App {
	return &App{
		engine: eng,
		config: Config.NewManager(configPath),
	}
}

// Engine devuelve el motor NovaCore subyacente.
func (a *App) Engine() *engine.Engine {
	return a.engine
}

// SetEventCallback conecta el callback de eventos del motor (descargas,
// logs, juegos) para poder reenviarlos al frontend.
func (a *App) SetEventCallback(cb engine.EventHandler) {
	if a.engine != nil {
		a.engine.SetEventCallback(cb)
	}
}

// Startup aplica la configuracion guardada al motor al arrancar.
func (a *App) Startup() {
	if a.config == nil || a.engine == nil {
		return
	}
	cfg := a.config.Get()
	a.engine.SetMaxRAM(cfg.Launcher.MaxRAMGB)
	a.engine.SetMaxMbps(cfg.Launcher.MaxMbps)
	a.engine.SetConcurrentDownloads(cfg.Launcher.ConcurrentDownloads)
	a.applyMinecraft(cfg.MinecraftConfig)
}

func (a *App) GetConfig() Config.Config {
	if a.config == nil {
		return Config.Default()
	}
	return a.config.Get()
}

func (a *App) GetMinecraftConfig() Config.MinecraftConfig {
	return a.GetConfig().MinecraftConfig
}

func (a *App) UpdateMinecraftConfig(mc Config.MinecraftConfig) {
	if a.config == nil {
		return
	}
	a.config.UpdateMinecraft(mc)
	a.applyMinecraft(a.config.Get().MinecraftConfig)
}

func (a *App) SetAuthVerify(verify bool) {
	if a.config == nil {
		return
	}
	a.config.SetAuthVerify(verify)
	a.applyMinecraft(a.config.Get().MinecraftConfig)
}

func (a *App) SetProxy(enabled bool, host string, port int, user, pass string) {
	if a.config == nil {
		return
	}
	a.config.SetProxy(enabled, host, port, user, pass)
	a.applyMinecraft(a.config.Get().MinecraftConfig)
}

func (a *App) MaxRAMGB() int {
	if a.config == nil {
		return 2
	}
	return a.config.Get().Launcher.MaxRAMGB
}

func (a *App) SetMaxRAM(gb int) {
	if a.config == nil {
		return
	}
	a.config.SetMaxRAMGB(gb)
	if a.engine != nil {
		a.engine.SetMaxRAM(gb)
	}
}

func (a *App) SetMaxMbps(mbps float64) {
	if a.config == nil {
		return
	}
	a.config.SetMaxMbps(mbps)
	if a.engine != nil {
		a.engine.SetMaxMbps(mbps)
	}
}

func (a *App) SetConcurrentDownloads(n int) {
	if a.config == nil {
		return
	}
	a.config.SetConcurrentDownloads(n)
	if a.engine != nil {
		a.engine.SetConcurrentDownloads(n)
	}
}

func (a *App) GetUIScale() int {
	if a.config == nil {
		return 100
	}
	return a.config.Get().Personalization.UIScale
}

func (a *App) SetUIScale(percent int) {
	if a.config == nil {
		return
	}
	a.config.SetUIScale(percent)
}

func (a *App) UpdatePersonalization(p Config.Personalization) {
	if a.config == nil {
		return
	}
	a.config.UpdatePersonalization(p)
}

var imageExts = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".gif": true, ".bmp": true}
var videoExts = map[string]bool{".mp4": true, ".gif": true, ".webm": true}

// LocalAssetsDir devuelve el workdir del launcher.
func (a *App) LocalAssetsDir() string {
	if a.engine == nil {
		return ""
	}
	return a.engine.ConfigManager().RootDir()
}

// ReadLocalFile devuelve los bytes de un archivo del workdir (ej: fondos).
// El frontend crea blob URLs para mostrarlos en el webview sin depender de
// rutas /local/ que el dev server de Vite no sabe servir.
func (a *App) ReadLocalFile(rel string) ([]byte, error) {
	if a.engine == nil {
		return nil, fmt.Errorf("engine no disponible")
	}
	clean := filepath.Clean(strings.ReplaceAll(rel, "\\", "/"))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return nil, fmt.Errorf("ruta invalida")
	}
	path := filepath.Join(a.engine.ConfigManager().RootDir(), clean)
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("no es un archivo")
	}
	return os.ReadFile(path)
}

// ImportBackground copia un archivo al directorio cache/backgrounds del
// workdir y devuelve la ruta relativa con "/" (ej: "cache/backgrounds/1234.png").
// El frontend la lee via ReadLocalFile para crear blob URLs.
func (a *App) ImportBackground(src, kind string) (string, error) {
	if a.engine == nil {
		return "", fmt.Errorf("engine no disponible")
	}
	ext := strings.ToLower(filepath.Ext(src))
	if kind == "image" && !imageExts[ext] {
		return "", fmt.Errorf("formato de imagen no soportado: %s", ext)
	}
	if kind == "video" && !videoExts[ext] {
		return "", fmt.Errorf("el fondo animado debe ser MP4, GIF o WEBM")
	}

	info, err := os.Stat(src)
	if err != nil {
		return "", fmt.Errorf("no se pudo leer el archivo: %v", err)
	}
	if kind == "video" && info.Size() > 20*1024*1024 {
		return "", fmt.Errorf("el fondo animado no debe pesar mas de 20MB")
	}
	if err := checkResolution(src, kind); err != nil {
		return "", err
	}

	destDir := filepath.Join(a.engine.ConfigManager().RootDir(), "cache", "backgrounds")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	dest := filepath.Join(destDir, fmt.Sprintf("%d%s", time.Now().UnixNano(), ext))

	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return "", err
	}

	return "cache/backgrounds/" + filepath.Base(dest), nil
}

// ResetConfig restaura toda la configuracion a los valores por defecto
// y la reaplica al motor.
func (a *App) ResetConfig() error {
	if a.config == nil {
		return nil
	}
	if err := a.config.Reset(); err != nil {
		return err
	}
	if a.engine == nil {
		return nil
	}
	cfg := a.config.Get()
	a.engine.SetMaxRAM(cfg.Launcher.MaxRAMGB)
	a.engine.SetMaxMbps(cfg.Launcher.MaxMbps)
	a.engine.SetConcurrentDownloads(cfg.Launcher.ConcurrentDownloads)
	a.applyMinecraft(cfg.MinecraftConfig)
	return nil
}

func (a *App) TotalRAMGB() int {
	if a.engine == nil {
		return 8
	}
	return a.engine.TotalRAMGB()
}

func (a *App) DetectJavaInstallations() []string {
	if a.engine == nil {
		return []string{}
	}
	return a.engine.DetectJavaInstallations()
}

func (a *App) GetCacheInfo() engine.CacheInfo {
	if a.engine == nil {
		return engine.CacheInfo{}
	}
	info := a.engine.GetCacheInfo()
	launcher := a.launcherCacheCount()
	if launcher > 0 {
		if info.Categories == nil {
			info.Categories = map[string]int{}
		}
		info.Categories["launcher"] = launcher
		info.TotalEntries += launcher
	}
	return info
}

// launcherBackgroundDirs devuelve los directorios de fondos del launcher:
// el nuevo dentro de cache y el legado por compatibilidad.
func (a *App) launcherBackgroundDirs() []string {
	root := a.engine.ConfigManager().RootDir()
	return []string{
		filepath.Join(root, "cache", "backgrounds"),
		filepath.Join(root, "backgrounds"),
	}
}

// referencedBackgrounds devuelve los nombres de archivo de fondo que estan
// siendo usados por la configuracion actual (no deben borrarse al limpiar).
func (a *App) referencedBackgrounds() map[string]bool {
	ref := map[string]bool{}
	if a.config == nil {
		return ref
	}
	bg := a.config.Get().Personalization.Background
	for _, rel := range []string{bg.ImagePath, bg.VideoPath} {
		if rel != "" {
			ref[filepath.Base(rel)] = true
		}
	}
	for _, rel := range bg.DynamicImages {
		if rel != "" {
			ref[filepath.Base(rel)] = true
		}
	}
	return ref
}

// launcherCacheCount cuenta los archivos de la cache del launcher
// (fondos no usados por la configuracion actual).
func (a *App) launcherCacheCount() int {
	if a.engine == nil {
		return 0
	}
	ref := a.referencedBackgrounds()
	count := 0
	for _, dir := range a.launcherBackgroundDirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && !ref[e.Name()] {
				count++
			}
		}
	}
	return count
}

// ClearAllCache limpia la cache de NovaCore (manifests) y la del launcher
// (fondos importados que no esten en uso). Devuelve la cantidad total de
// archivos eliminados.
func (a *App) ClearAllCache() int {
	if a.engine == nil {
		return 0
	}
	total := a.engine.ClearAllCache()
	ref := a.referencedBackgrounds()
	for _, dir := range a.launcherBackgroundDirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || ref[e.Name()] {
				continue
			}
			if os.Remove(filepath.Join(dir, e.Name())) == nil {
				total++
			}
		}
	}
	return total
}

func (a *App) applyMinecraft(mc Config.MinecraftConfig) {
	if a.engine == nil {
		return
	}
	cur := a.engine.GetConfig()
	a.engine.UpdateMinecraftConfig(engine.MinecraftConfig{
		HardwareEnabled:      mc.HardwareEnabled,
		HardwareAcceleration: mc.HardwareAcceleration,
		GPUType:              mc.GPUType,
		GPUPreset:            mc.GPUPreset,
		JavaMode:             mc.JavaMode,
		JavaCustomPath:       mc.JavaCustomPath,
		ProxyEnabled:         mc.ProxyEnabled,
		ProxyHost:            mc.ProxyHost,
		ProxyPort:            mc.ProxyPort,
		ProxyUser:            mc.ProxyUser,
		ProxyPass:            mc.ProxyPass,
		AuthVerify:           mc.AuthVerify,
		WindowWidth:          mc.WindowWidth,
		WindowHeight:         mc.WindowHeight,
		Fullscreen:           mc.Fullscreen,
		JavaArgs:             mc.JavaArgs,
		GameArgs:             mc.GameArgs,
		OfflineMode:          mc.OfflineMode,
		CompatMode:           mc.CompatMode,
		DetailedLogs:         mc.DetailedLogs,
		ConcurrentDownloads:  cur.ConcurrentDownloads,
	})
}
