package Handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	goruntime "runtime"
	"sort"
	"strings"
	"time"

	"StepLauncher/internal/Config"
	assets "StepLauncher/internal/Core/Assets"
	news "StepLauncher/internal/Core/News"
	engine "StepLauncher/internal/Handlers/Engine"
	RichPresence "StepLauncher/internal/RichPresence"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx    context.Context
	engine *engine.Engine
	config *Config.Manager
	assets *assets.Manager
	news   *news.Manager
	rp     *RichPresence.Manager
}

func NewApp(eng *engine.Engine, configPath string) *App {
	cfgMgr := Config.NewManager(configPath)
	if eng != nil && eng.Logger() != nil {
		cfgMgr.SetLogFn(func(f string, a ...interface{}) { eng.Logger().Info(f, a...) })
	}
	rootDir := filepath.Dir(configPath)
	if eng != nil && eng.ConfigManager() != nil {
		rootDir = eng.ConfigManager().RootDir()
	}
	a := &App{
		engine: eng,
		config: cfgMgr,
		rp:     RichPresence.NewManager(),
	}
	a.rp.SetLogFn(func(f string, args ...interface{}) {
		a.logf("[RichPresence] "+f, args...)
	})
	a.news = news.NewManager(rootDir, func(f string, args ...interface{}) {
		a.logf(f, args...)
	})
	a.initAssets(rootDir)
	return a
}

func (a *App) logf(format string, args ...interface{}) {
	if a.engine == nil || a.engine.Logger() == nil {
		return
	}
	a.engine.Logger().Info(format, args...)
}

func (a *App) initAssets(rootDir string) {
	a.assets = assets.NewManager(rootDir)
	if err := a.assets.Ensure(); err != nil {
		a.logf("[Assets] WARN: no se pudo crear launcher_assets.json: %v", err)
		return
	}
	if a.config != nil {
		for _, reg := range []struct{ key, file string }{
			{Config.ExtraKeyAssets, Config.FileAssets},
			{Config.ExtraKeyAccounts, Config.FileAccounts},
			{Config.ExtraKeyHistory, Config.FileHistory},
			{Config.ExtraKeyProfiles, Config.FileProfiles},
			{Config.ExtraKeyCrashHistory, Config.FileCrashHistory},
		} {
			if err := a.config.RegisterExtraFile(reg.key, reg.file); err != nil {
				a.logf("[Config] WARN: no se pudo registrar extraData %s: %v", reg.file, err)
			}
		}
	}
}

func (a *App) Engine() *engine.Engine {
	return a.engine
}

func (a *App) SetEventCallback(cb engine.EventHandler) {
	if a.news != nil {
		a.news.SetEventCallback(news.EventHandler(cb))
	}
	if a.engine == nil {
		return
	}
	a.engine.SetEventCallback(func(eventType string, data []byte) {
		a.handleRichPresenceEvent(eventType, data)
		if cb != nil {
			cb(eventType, data)
		}
	})
}

func (a *App) Startup() {
	if a.config == nil || a.engine == nil {
		return
	}
	cfg := a.config.Get()
	a.engine.SetMaxRAM(cfg.Launcher.MaxRAMGB)
	a.engine.SetMaxMbps(cfg.Launcher.MaxMbps)
	a.engine.SetConcurrentDownloads(cfg.Launcher.ConcurrentDownloads)
	a.engine.SetVerifyIntegrity(cfg.Launcher.VerifyEnabled())
	a.applyMinecraft(cfg.MinecraftConfig)
	a.applyRichPresence(cfg)
}

func (a *App) applyRichPresence(cfg Config.Config) {
	if a.rp == nil {
		return
	}
	a.rp.SetEnabled(cfg.RichPresence.EnabledValue())
	if cfg.RichPresence.EnabledValue() {
		a.rp.SetActivity("StepLauncher", "Navegando por el menú", 0)
	}
}

func (a *App) handleRichPresenceEvent(eventType string, data []byte) {
	if a.rp == nil {
		return
	}
	switch eventType {
	case "game_starting", "game_started", "game_exited", "game_crashed", "game_stopped":
	default:
		return
	}
	var evt struct {
		Type string `json:"type"`
		Data struct {
			Version string `json:"version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &evt); err != nil {
		return
	}
	version := evt.Data.Version
	switch eventType {
	case "game_starting":
		a.rp.SetActivity("StepLauncher "+version, "Lanzando Minecraft", 0)
	case "game_started":
		a.rp.SetActivity("StepLauncher "+version, "Jugando Minecraft", time.Now().UnixMilli())
	default:
		still := a.runningGameVersion()
		if still != "" {
			a.rp.SetActivity("StepLauncher "+still, "Jugando Minecraft", time.Now().UnixMilli())
		} else {
			a.rp.SetActivity("StepLauncher", "Navegando por el menú", 0)
		}
	}
}

func (a *App) runningGameVersion() string {
	if a.engine == nil {
		return ""
	}
	for _, g := range a.engine.ListGames() {
		if g.Status == engine.GameRunning || g.Status == engine.GameStarting {
			return g.Version
		}
	}
	return ""
}

func (a *App) GetRichPresenceConfig() Config.RichPresenceConfig {
	if a.config == nil {
		return Config.RichPresenceConfig{}
	}
	return a.config.Get().RichPresence
}

func (a *App) SetRichPresenceEnabled(v bool) {
	if a.config == nil || a.rp == nil {
		return
	}
	a.config.SetRichPresenceEnabled(v)
	a.rp.SetEnabled(v)
}

func (a *App) CheckForUpdates() {
	if a.engine != nil {
		a.engine.CheckForUpdates()
	}
}

func (a *App) NewsRefreshIndex() {
	if a.news != nil {
		a.news.RefreshIndex()
	}
}

func (a *App) NewsLoadRelease(version string) {
	if a.news != nil {
		a.news.LoadRelease(version)
	}
}

func (a *App) NewsLoadChangelog(version string) {
	if a.news != nil {
		a.news.LoadChangelog(version)
	}
}

func (a *App) NewsLoadMarkdown(url string) {
	if a.news != nil {
		a.news.LoadMarkdown(url)
	}
}

func (a *App) ApplyUpdate() error {
	if a.engine == nil {
		return fmt.Errorf("motor no disponible")
	}
	info := a.engine.LastUpdateInfo()
	if info == nil || !info.HasUpdate {
		return fmt.Errorf("no hay actualización disponible")
	}

	if goruntime.GOOS == "windows" {
		if info.UpdaterURL == "" {
			if a.ctx != nil {
				runtime.BrowserOpenURL(a.ctx, info.ReleaseURL)
			}
			return nil
		}
		path, err := a.engine.DownloadUpdater(info.UpdaterURL)
		if err != nil {
			return err
		}
		if err := a.engine.LaunchUpdater(path); err != nil {
			return err
		}
		if a.ctx != nil {
			runtime.Quit(a.ctx)
		}
		return nil
	}

	if a.ctx != nil {
		runtime.BrowserOpenURL(a.ctx, info.ReleaseURL)
	}
	return nil
}

func (a *App) GetCheckForUpdatesOnStart() bool {
	if a.config == nil {
		return false
	}
	return a.config.Get().Launcher.CheckForUpdatesOnStart
}

func (a *App) SetCheckForUpdatesOnStart(v bool) {
	if a.config == nil {
		return
	}
	a.config.SetCheckForUpdatesOnStart(v)
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

func (a *App) SetVerifyIntegrity(v bool) {
	if a.config == nil {
		return
	}
	a.config.SetVerifyIntegrity(v)
	if a.engine != nil {
		a.engine.SetVerifyIntegrity(v)
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

func (a *App) SetIdle(idle Config.IdleConfig) {
	if a.config == nil {
		return
	}
	a.config.UpdateIdle(idle)
}

func (a *App) SetHideLauncher(v bool) {
	if a.config == nil {
		return
	}
	a.config.SetHideLauncher(v)
}

func (a *App) RefreshManifests() (int, error) {
	if a.engine == nil {
		return 0, nil
	}
	return a.engine.RefreshManifests()
}

func (a *App) UpdatePersonalization(p Config.Personalization) {
	if a.config == nil {
		return
	}
	a.config.UpdatePersonalization(p)
}

var imageExts = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".gif": true, ".bmp": true}
var videoExts = map[string]bool{".mp4": true, ".gif": true, ".webm": true}

func (a *App) LocalAssetsDir() string {
	if a.engine == nil {
		return ""
	}
	return a.engine.ConfigManager().RootDir()
}

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

type ScreenshotInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
	Time string `json:"time,omitempty"`
}

var screenshotExts = map[string]bool{".png": true, ".jpg": true, ".jpeg": true}

func (a *App) ListScreenshots() ([]ScreenshotInfo, error) {
	if a.engine == nil {
		return nil, fmt.Errorf("engine no disponible")
	}
	root := a.engine.ConfigManager().RootDir()
	searchDirs := []string{
		filepath.Join(root, "game", "screenshots"),
		filepath.Join(root, "game"),
		filepath.Join(root, "screenshots"),
	}
	seen := map[string]bool{}
	var out []ScreenshotInfo
	for _, dir := range searchDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		var screenshots []os.DirEntry
		for _, e := range entries {
			if e.Type().IsRegular() && screenshotExts[strings.ToLower(filepath.Ext(e.Name()))] {
				screenshots = append(screenshots, e)
			}
		}
		sort.Slice(screenshots, func(i, j int) bool {
			ai, _ := screenshots[i].Info()
			aj, _ := screenshots[j].Info()
			return ai.ModTime().After(aj.ModTime())
		})
		for _, e := range screenshots {
			name := e.Name()
			if seen[name] {
				continue
			}
			seen[name] = true
			info, err := e.Info()
			if err != nil {
				continue
			}
			rel, err := filepath.Rel(root, filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			out = append(out, ScreenshotInfo{
				Name: name,
				Path: filepath.ToSlash(rel),
				Size: info.Size(),
				Time: info.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}
	return out, nil
}

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
	a.engine.SetVerifyIntegrity(cfg.Launcher.VerifyEnabled())
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
		a.logf("[Cache] GetCacheInfo: engine no disponible (nil)")
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

func (a *App) launcherBackgroundDirs() []string {
	root := a.engine.ConfigManager().RootDir()
	return []string{
		filepath.Join(root, "cache", "backgrounds"),
		filepath.Join(root, "backgrounds"),
	}
}

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

func (a *App) ClearAllCache() int {
	if a.engine == nil {
		a.logf("[Cache] ClearAllCache: engine no disponible (nil)")
		return 0
	}
	before := a.GetCacheInfo().TotalEntries
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
	after := a.GetCacheInfo().TotalEntries
	a.logf("[Cache] Limpieza completa: %d archivos eliminados (antes=%d despues=%d)", total, before, after)
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

func (a *App) Shutdown() {
	if a.rp != nil {
		a.rp.Close()
	}
}
