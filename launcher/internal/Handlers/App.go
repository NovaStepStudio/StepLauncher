package Handlers

import (
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
	musichistory "StepLauncher/internal/Music/MusicHistory"
	"StepLauncher/internal/Music/NowPlaying"
	news "StepLauncher/internal/Core/News"
	playlists "StepLauncher/internal/Music/Playlist"
	engine "StepLauncher/internal/Handlers/Engine"
	music "StepLauncher/internal/Music"
	RichPresence "StepLauncher/internal/RichPresence"
)

type App struct {
	engine       *engine.Engine
	config       *Config.Manager
	assets       *assets.Manager
	news         *news.Manager
	playlists    *playlists.Manager
	musicHistory *musichistory.Manager
	nowPlaying   *nowplaying.Manager
	musicCache   *music.Manager
	rp           *RichPresence.Manager
	runtime      RuntimeBridge

	firstLaunchPending bool
}

func NewApp(eng *engine.Engine, configPath string) *App {
	_, statErr := os.Stat(configPath)
	cfgMgr := Config.NewManager(configPath)
	if eng != nil && eng.Logger() != nil {
		cfgMgr.SetLogFn(func(f string, a ...interface{}) { eng.Logger().Info(f, a...) })
	}
	rootDir := filepath.Dir(configPath)
	if eng != nil && eng.ConfigManager() != nil {
		rootDir = eng.ConfigManager().RootDir()
	}
	a := &App{
		engine:             eng,
		config:             cfgMgr,
		rp:                 RichPresence.NewManager(),
		firstLaunchPending: os.IsNotExist(statErr),
	}
	a.rp.SetLogFn(func(f string, args ...interface{}) {
		a.logf("[RichPresence] "+f, args...)
	})
	a.news = news.NewManager(rootDir, func(f string, args ...interface{}) {
		a.logf(f, args...)
	})
	a.playlists = playlists.NewManager(rootDir)
	a.musicHistory = musichistory.NewManager(rootDir)
	a.nowPlaying = nowplaying.NewManager(rootDir)
	a.musicCache = music.NewManager(rootDir)
	// Wire resolvers para que playlists cachee TotalDuration/TrackCount y PreviewCovers (primeras 4) al crear/editar/importar
	// sin que la UI tenga que leer la playlist entera y recuperar las 4 carátulas cada vez
	if a.playlists != nil && a.musicCache != nil {
		a.playlists.SetDurationResolver(func(path string) float64 {
			if meta, ok := a.musicCache.GetMetadata(path); ok {
				return meta.Duration
			}
			return 0
		})
		a.playlists.SetCoverResolver(func(path string) string {
			thumb, err := a.musicCache.GetCoverBase64(path, "thumb")
			if err == nil && thumb != "" {
				return thumb
			}
			return ""
		})
		// Migrar playlists existentes con stats/preview vacíos en segundo plano
		go func() {
			_ = a.playlists.RefreshMissingStats()
		}()
	}
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
			{Config.ExtraKeyPlaylists, Config.FilePlaylists},
			{Config.ExtraKeyMusicHistory, Config.FileMusicHistory},
		} {
			if err := a.config.RegisterExtraFile(reg.key, reg.file); err != nil {
				a.logf("[Config] WARN: no se pudo registrar extraData %s: %v", reg.file, err)
			}
		}
	}
}

// SetRuntimeBridge inyecta la implementación de runtime (diálogos, navegador,
// salir) que provee el bootstrap de la aplicación.
func (a *App) SetRuntimeBridge(b RuntimeBridge) {
	a.runtime = b
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
	a.engine.SetVerifyBeforeLaunch(cfg.Launcher.VerifyBeforeLaunchEnabled())
	a.applyMinecraft(cfg.MinecraftConfig)
	a.applyRichPresence(cfg)
	// Migración: purgar entradas huérfanas de gallery heredadas (p. ej. 22 entradas con solo 1 archivo existente).
	// Se ejecuta al arrancar para limpiar instalaciones ya afectadas sin esperar a un cambio manual.
	a.pruneOrphanGallery()
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
			if a.runtime != nil {
				a.runtime.BrowserOpenURL(info.ReleaseURL)
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
		if a.runtime != nil {
			a.runtime.Quit()
		}
		return nil
	}

	if a.runtime != nil {
		a.runtime.BrowserOpenURL(info.ReleaseURL)
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

func (a *App) SetLaunchAfterInstall(v bool) {
	if a.config == nil {
		return
	}
	a.config.SetLaunchAfterInstall(v)
}

func (a *App) GetConfig() Config.Config {
	if a.config == nil {
		return Config.Default()
	}
	return a.config.Get()
}

func (a *App) GetFirstLaunch() bool {
	if a.config == nil {
		return false
	}
	return a.config.Get().FirstLaunch || a.firstLaunchPending
}

func (a *App) SetFirstLaunchDone() {
	if a.config == nil {
		return
	}
	a.firstLaunchPending = false
	a.config.SetFirstLaunchDone()
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

func (a *App) SetVerifyBeforeLaunch(v bool) {
	if a.config == nil {
		return
	}
	a.config.SetVerifyBeforeLaunch(v)
	if a.engine != nil {
		a.engine.SetVerifyBeforeLaunch(v)
	}
}

func (a *App) StartIntegrityCheck(scope string) error {
	if a.config != nil {
		a.config.SetIntegritySector(scope)
	}
	if a.engine == nil {
		return fmt.Errorf("engine no disponible")
	}
	return a.engine.StartIntegrityCheck(scope)
}

func (a *App) CancelIntegrityCheck() {
	if a.engine != nil {
		a.engine.CancelIntegrityCheck()
	}
}

func (a *App) IntegrityStatus() engine.IntegrityProgress {
	if a.engine == nil {
		return engine.IntegrityProgress{State: engine.IntegrityStateIdle}
	}
	return a.engine.IntegrityStatus()
}

func (a *App) SetIntegritySector(s string) {
	if a.config == nil {
		return
	}
	a.config.SetIntegritySector(s)
}

func (a *App) GetIntegritySector() string {
	if a.config == nil {
		return "todo"
	}
	return a.config.Get().Launcher.IntegritySector
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
	_ = a.updatePersonalizationInternal(p)
}

// updatePersonalizationInternal aplica el cambio de personalización y purga entradas huérfanas
// de gallery en launcher_assets.json (evita acumulación infinita). Separa la lógica para que
// DownloadGalleryImageAsBackground pueda reutilizarla y propagar errores.
func (a *App) updatePersonalizationInternal(p Config.Personalization) error {
	if a.config == nil {
		return fmt.Errorf("config no disponible")
	}
	if err := a.config.UpdatePersonalization(p); err != nil {
		return err
	}
	a.pruneOrphanGallery()
	return nil
}

// pruneOrphanGallery elimina entradas huérfanas de launcher_assets.json.gallery que ya no están
// referenciadas en Personalization.Background (ImagePath, VideoPath, DynamicImages). Borra también
// el archivo físico si aún existe. Evita el bug reportado donde cada “Aplicar fondo” dejaba una
// entrada eterna aunque el archivo fuese borrado por cleanupBackgrounds.
func (a *App) pruneOrphanGallery() {
	if a.assets == nil || a.config == nil {
		return
	}
	ref := a.referencedBackgrounds()
	n, err := a.assets.PruneOrphanGallery(ref)
	if err != nil {
		a.logf("[Assets] WARN: no se pudo purgar gallery huérfana: %v", err)
		return
	}
	if n > 0 {
		a.logf("[Assets] Gallery huérfana purgada: %d entradas eliminadas", n)
	}
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
	return a.listScreenshots(root, searchDirs), nil
}

func (a *App) ListInstanceScreenshots(instanceName string) ([]ScreenshotInfo, error) {
	if a.engine == nil {
		return nil, fmt.Errorf("engine no disponible")
	}
	if err := validateInstanceName(instanceName); err != nil {
		return nil, err
	}
	root := a.engine.ConfigManager().RootDir()
	instDir := filepath.Join(root, a.engine.ConfigManager().Get().InstancesDir, instanceName)
	searchDirs := []string{
		filepath.Join(instDir, "game", "screenshots"),
		filepath.Join(instDir, "game"),
		filepath.Join(instDir, "screenshots"),
	}
	return a.listScreenshots(root, searchDirs), nil
}

func (a *App) listScreenshots(root string, searchDirs []string) []ScreenshotInfo {
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
	return out
}

var instanceAssetKinds = map[string]bool{"icon": true, "banner": true, "background": true}

// validateInstanceName replica las reglas del gestor de instancias
// (sin separadores de ruta ni "..") para validar entrada desde bindings.
func validateInstanceName(name string) error {
	if name == "" || strings.Contains(name, "..") || strings.ContainsAny(name, "/\\") || filepath.Base(name) != name {
		return fmt.Errorf("nombre de instancia invalido")
	}
	return nil
}

func (a *App) ImportInstanceAsset(name, kind, src string) (string, error) {
	if a.engine == nil {
		return "", fmt.Errorf("engine no disponible")
	}
	if err := validateInstanceName(name); err != nil {
		return "", err
	}
	if !instanceAssetKinds[kind] {
		return "", fmt.Errorf("tipo de asset no soportado: %s", kind)
	}
	ext := strings.ToLower(filepath.Ext(src))
	if !imageExts[ext] {
		return "", fmt.Errorf("formato de imagen no soportado: %s", ext)
	}
	root := a.engine.ConfigManager().RootDir()
	instDir := filepath.Join(root, a.engine.ConfigManager().Get().InstancesDir, name)
	if _, err := os.Stat(filepath.Join(instDir, "instance.metadata.json")); err != nil {
		return "", fmt.Errorf("la instancia %s no existe", name)
	}
	assetsDir := filepath.Join(instDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return "", err
	}
	dest := filepath.Join(assetsDir, kind+ext)
	in, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("no se pudo leer el archivo: %v", err)
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
	rel, err := filepath.Rel(root, dest)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

func (a *App) PickInstanceAssetFile() (string, error) {
	if a.runtime == nil {
		return "", fmt.Errorf("runtime no disponible")
	}
	return a.runtime.OpenFileDialog("Seleccionar imagen para la instancia", []FileFilter{
		{DisplayName: "Imagenes (*.png, *.jpg, *.jpeg, *.webp, *.gif, *.bmp)", Pattern: "*.png;*.jpg;*.jpeg;*.webp;*.gif;*.bmp"},
	})
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
	a.engine.SetVerifyBeforeLaunch(cfg.Launcher.VerifyBeforeLaunchEnabled())
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
	covers := a.coverCacheCount()
	if covers > 0 {
		if info.Categories == nil {
			info.Categories = map[string]int{}
		}
		info.Categories["covers"] = covers
		info.TotalEntries += covers
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
		if rel == "" {
			continue
		}
		clean := filepath.ToSlash(filepath.Clean(strings.ReplaceAll(rel, "\\", "/")))
		ref[clean] = true
		ref[filepath.Base(clean)] = true
		// compat: guardar también la clave original por si viene sin normalizar
		ref[rel] = true
	}
	for _, rel := range bg.DynamicImages {
		if rel == "" {
			continue
		}
		clean := filepath.ToSlash(filepath.Clean(strings.ReplaceAll(rel, "\\", "/")))
		ref[clean] = true
		ref[filepath.Base(clean)] = true
		ref[rel] = true
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
	// Incluir caché de carátulas en la limpieza total
	total += a.ClearCoverCache()
	// Purgar entradas huérfanas de gallery cuyo archivo fue borrado por la limpieza o ya era huérfana
	if a.assets != nil {
		if n, _ := a.assets.PruneOrphanGallery(ref); n > 0 {
			a.logf("[Assets] Gallery huérfana purgada tras limpiar caché: %d entradas", n)
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
		SeparateGameDir:      mc.SeparateGameDir,
	})
}

// GetSeparateGameDir indica si el gameDir es <workDir>/game (true) o el
// propio workDir (false). En modo Minecraft siempre es false.
func (a *App) GetSeparateGameDir() bool {
	if a.config == nil {
		return true
	}
	info := a.GetDirectoryInfoSafe()
	if info.Mode == "minecraft" {
		return false
	}
	return a.config.Get().MinecraftConfig.SeparateGameDirValue()
}

// SetSeparateGameDir persiste la opcion de gameDir separado.
func (a *App) SetSeparateGameDir(v bool) {
	if a.config == nil {
		return
	}
	mc := a.config.Get().MinecraftConfig
	mc.SeparateGameDir = &v
	a.config.UpdateMinecraft(mc)
	a.applyMinecraft(a.config.Get().MinecraftConfig)
}

func (a *App) GetDirectoryInfoSafe() engine.DirectoryInfo {
	if a.engine == nil {
		return engine.DirectoryInfo{}
	}
	return a.engine.DirectoryInfo()
}

func (a *App) GetMusicPanelConfig() Config.MusicPanelConfig {
	if a.config == nil {
		return Config.MusicPanelConfig{CoverStyle: "square", ColorMode: "vibrant", PageSize: 20}
	}
	return a.config.Get().MusicPanel
}

func (a *App) UpdateMusicPanelConfig(p Config.MusicPanelConfig) error {
	if a.config == nil {
		return fmt.Errorf("config no disponible")
	}
	return a.config.UpdateMusicPanel(p)
}

func (a *App) SetMusicFolder(folder string) error {
	if a.config == nil {
		return fmt.Errorf("config no disponible")
	}
	if err := a.config.SetMusicFolder(folder); err != nil {
		return err
	}
	if len(a.config.GetMusicFolders()) == 0 && a.musicCache != nil {
		_ = a.musicCache.Clear()
	}
	return nil
}

func (a *App) GetMusicFolders() []string {
	if a.config == nil {
		return []string{}
	}
	return a.config.GetMusicFolders()
}

func (a *App) SetMusicFolders(folders []string) error {
	if a.config == nil {
		return fmt.Errorf("config no disponible")
	}
	if err := a.config.SetMusicFolders(folders); err != nil {
		return err
	}
	// PROHIBIDO carpeta por defecto: si se vacían las carpetas, limpiar índice para no mostrar pistas fantasma
	if len(a.config.GetMusicFolders()) == 0 && a.musicCache != nil {
		_ = a.musicCache.Clear()
	}
	return nil
}

func (a *App) AddMusicFolder(folder string) error {
	if a.config == nil {
		return fmt.Errorf("config no disponible")
	}
	return a.config.AddMusicFolder(folder)
}

func (a *App) RemoveMusicFolder(folder string) error {
	if a.config == nil {
		return fmt.Errorf("config no disponible")
	}
	if err := a.config.RemoveMusicFolder(folder); err != nil {
		return err
	}
	// Si no quedan carpetas, limpiar índice — evita que aparezca carpeta fantasma en Ajustes > Música
	if len(a.config.GetMusicFolders()) == 0 && a.musicCache != nil {
		_ = a.musicCache.Clear()
	}
	return nil
}

func (a *App) PickMusicFolder() (string, error) {
	if a.runtime == nil {
		return "", fmt.Errorf("runtime no disponible")
	}
	return a.runtime.OpenDirectoryDialog("Seleccionar carpeta de música")
}

func (a *App) ScanAllMusicFolders() ([]string, error) {
	if a.config == nil {
		return nil, fmt.Errorf("config no disponible")
	}
	folders := a.config.GetMusicFolders()
	// Compatibilidad legacy: usar MusicFolder si MusicFolders vacío — NO es carpeta por defecto
	if len(folders) == 0 {
		if mf := a.config.Get().MusicPanel.MusicFolder; mf != "" {
			folders = []string{mf}
		}
	}
	// PROHIBIDO: nunca crear ni usar carpeta por defecto — error explícito si no hay carpetas
    if len(folders) == 0 {
        return nil, fmt.Errorf("no hay carpetas de música configuradas")
    }
    var resolved []string
    for _, f := range folders {
        resolved = append(resolved, filepath.Clean(f))
    }
    if a.musicCache == nil {
        return nil, fmt.Errorf("cache de música no disponible")
    }
    all, err := a.musicCache.ScanFolders(resolved)
	if err != nil {
		return nil, err
	}
	_ = a.autoImportPlaylistsFromFolders(resolved)
	return all, nil
}

func (a *App) autoImportPlaylistsFromFolders(folders []string) error {
	if a.playlists == nil {
		return nil
	}
	playlistExts := map[string]bool{".m3u": true, ".m3u8": true, ".pls": true}
	// Deduplicación por ruta absoluta normalizada y por contenido (título + tracks), sin usar hash
	seenPaths := map[string]bool{}
	for _, pl := range a.playlists.List() {
		for _, tp := range pl.TrackPaths {
			seenPaths[strings.ToLower(filepath.Clean(tp))] = true
		}
	}
	existingTitles := map[string]bool{}
	for _, pl := range a.playlists.List() {
		existingTitles[strings.ToLower(strings.TrimSpace(pl.Title))] = true
	}
    for _, folder := range folders {
        clean := filepath.Clean(folder)
        _ = filepath.WalkDir(clean, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if !playlistExts[ext] {
				return nil
			}
			normPath := strings.ToLower(filepath.Clean(path))
			if seenPaths[normPath] {
				return nil
			}
			title := strings.TrimSuffix(filepath.Base(path), ext)
			ltitle := strings.ToLower(strings.TrimSpace(title))
			if existingTitles[ltitle] {
				return nil
			}
			// Resolver tracks para comparar sin duplicar
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			tracks := parsePlaylistTracksForDedup(data, ext, filepath.Dir(path), a.GetMusicFolders())
			if len(tracks) == 0 {
				return nil
			}
			// Si ya existe playlist con mismo conjunto de tracks (mismo orden y cantidad), evitar duplicado
			for _, pl := range a.playlists.List() {
				if len(pl.TrackPaths) != len(tracks) {
					continue
				}
				match := true
				for i := range tracks {
					if strings.ToLower(filepath.Clean(pl.TrackPaths[i])) != strings.ToLower(filepath.Clean(tracks[i])) {
						match = false
						break
					}
				}
				if match {
					seenPaths[normPath] = true
					return nil
				}
			}
			if _, err := a.ImportPlaylistFile(path); err == nil {
				existingTitles[ltitle] = true
				seenPaths[normPath] = true
			}
			return nil
		})
	}
	return nil
}

func parsePlaylistTracksForDedup(data []byte, ext, base string, musicFolders []string) []string {
	lines := strings.Split(string(data), "\n")
	var tracks []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		if ext == ".pls" && strings.Contains(l, "=") {
			parts := strings.SplitN(l, "=", 2)
			if len(parts) == 2 && strings.HasPrefix(strings.TrimSpace(parts[0]), "File") {
				l = strings.TrimSpace(parts[1])
			}
		}
		clean := filepath.Clean(l)
		if !filepath.IsAbs(clean) {
			if base != "" {
				clean = filepath.Join(base, clean)
			}
		}
		clean = filepath.Clean(clean)
		tracks = append(tracks, clean)
	}
	return tracks
}

func (a *App) CancelMusicScan() bool {
	if a.musicCache == nil {
		return false
	}
	return a.musicCache.CancelScan()
}

func (a *App) ScanMusicFolder(folder string) ([]string, error) {
    if folder == "" {
        if a.config != nil {
            folder = a.config.Get().MusicPanel.MusicFolder
        }
        // PROHIBIDO: nunca usar carpeta por defecto — error si no hay carpeta explícita
        if folder == "" {
            return nil, fmt.Errorf("no hay carpeta de música configurada")
        }
    }
    folder = filepath.Clean(folder)
	if a.musicCache == nil {
		return nil, fmt.Errorf("cache de música no disponible")
	}
	return a.musicCache.ScanFolder(folder)
}

func (a *App) ReadAbsoluteFile(path string) ([]byte, error) {
	clean := filepath.Clean(path)
	if clean == "." || clean == "" {
		return nil, fmt.Errorf("ruta inválida")
	}
	// Permitir absolutas y relativas; para absolutas validar que existe y es archivo regular
	// y que está dentro de la carpeta de música configurada o es imagen/audio permitido
	info, err := os.Stat(clean)
	if err != nil {
		// Si es relativa, intentar resolver contra rootDir
		if !filepath.IsAbs(clean) && a.engine != nil {
			alt := filepath.Join(a.engine.ConfigManager().RootDir(), clean)
			if info2, err2 := os.Stat(alt); err2 == nil {
				clean = alt
				info = info2
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("no es un archivo")
	}
	if info.Size() > 20*1024*1024 {
		return nil, fmt.Errorf("archivo demasiado grande")
	}
	ext := strings.ToLower(filepath.Ext(clean))
	allowed := map[string]bool{".mp3": true, ".wav": true, ".ogg": true, ".m4a": true, ".flac": true, ".png": true, ".jpg": true, ".jpeg": true, ".webp": true}
	if !allowed[ext] {
		return nil, fmt.Errorf("formato no soportado: %s", ext)
	}
	return os.ReadFile(clean)
}

// --- Playlists ---

func (a *App) ListPlaylists() []playlists.Playlist {
	if a.playlists == nil {
		return []playlists.Playlist{}
	}
	return a.playlists.List()
}

func (a *App) GetPlaylist(id string) (*playlists.Playlist, error) {
	if a.playlists == nil {
		return nil, fmt.Errorf("playlists no disponible")
	}
	return a.playlists.Get(id)
}

func (a *App) CreatePlaylist(title string, trackPaths []string) (*playlists.Playlist, error) {
	if a.playlists == nil {
		return nil, fmt.Errorf("playlists no disponible")
	}
	return a.playlists.Create(title, trackPaths)
}

func (a *App) UpdatePlaylist(id string, p playlists.Playlist) (*playlists.Playlist, error) {
	if a.playlists == nil {
		return nil, fmt.Errorf("playlists no disponible")
	}
	return a.playlists.Update(id, p)
}

func (a *App) DeletePlaylist(id string) error {
	if a.playlists == nil {
		return fmt.Errorf("playlists no disponible")
	}
	return a.playlists.Delete(id)
}

func (a *App) ImportPlaylistFile(path string) (*playlists.Playlist, error) {
	if a.playlists == nil {
		return nil, fmt.Errorf("playlists no disponible")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".m3u" || ext == ".m3u8" || ext == ".pls" {
		lines := strings.Split(string(data), "\n")
		var tracks []string
		base := filepath.Dir(path)
		musicFolders := a.GetMusicFolders()
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" || strings.HasPrefix(l, "#") {
				continue
			}
			// Manejar entradas tipo File1=path en .pls
			if ext == ".pls" && strings.Contains(l, "=") {
				parts := strings.SplitN(l, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					if strings.HasPrefix(key, "File") {
						l = strings.TrimSpace(parts[1])
					}
				}
			}
			// Resolver desde base del m3u, con fallback a musicFolders y rootDir
			resolved := a.playlists.ResolveFromBase(l, base, musicFolders)
			tracks = append(tracks, resolved)
		}
		title := strings.TrimSuffix(filepath.Base(path), ext)
		return a.playlists.Create(title, tracks)
	}
	// Intentar JSON de playlist
	var pl playlists.Playlist
	if err := json.Unmarshal(data, &pl); err == nil && pl.Title != "" {
		return a.playlists.Create(pl.Title, pl.TrackPaths)
	}
	return nil, fmt.Errorf("formato de playlist no soportado")
}

func (a *App) PickPlaylistFile() (string, error) {
	if a.runtime == nil {
		return "", fmt.Errorf("runtime no disponible")
	}
	return a.runtime.OpenFileDialog("Importar playlist", []FileFilter{
		{DisplayName: "Playlist (*.m3u, *.m3u8, *.json)", Pattern: "*.m3u;*.m3u8;*.json"},
	})
}

// --- Historial de música (launcher_music_history.json) ---
func (a *App) GetMusicHistory() []musichistory.Entry {
	if a.musicHistory == nil {
		return []musichistory.Entry{}
	}
	return a.musicHistory.List()
}

func (a *App) AddMusicHistory(path, title, artist, coverUrl string) []musichistory.Entry {
	if a.musicHistory == nil {
		return []musichistory.Entry{}
	}
	return a.musicHistory.Add(path, title, artist, coverUrl)
}

func (a *App) ClearMusicHistory() error {
	if a.musicHistory == nil {
		return fmt.Errorf("historial no disponible")
	}
	return a.musicHistory.Clear()
}

func (a *App) GetNowPlayingQueue() nowplaying.Queue {
	if a.nowPlaying == nil {
		return nowplaying.Queue{Tracks: []string{}, CurrentIndex: -1}
	}
	return a.nowPlaying.Get()
}

func (a *App) SetNowPlayingQueue(tracks []string, idx int, curPath string) error {
	if a.nowPlaying == nil {
		return fmt.Errorf("cola no disponible")
	}
	return a.nowPlaying.Set(tracks, idx, curPath)
}

func (a *App) ClearNowPlayingQueue() error {
	if a.nowPlaying == nil {
		return fmt.Errorf("cola no disponible")
	}
	return a.nowPlaying.Clear()
}

func (a *App) Shutdown() {
	if a.rp != nil {
		a.rp.Close()
	}
}
