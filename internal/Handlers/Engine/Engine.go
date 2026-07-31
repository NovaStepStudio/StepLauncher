package engine
import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"time"

	"StepLauncher/internal/Core/Cache"
	"StepLauncher/internal/Core/Config"
	"StepLauncher/internal/Core/Downloader"
	lhistory "StepLauncher/internal/Core/Launcher/History"
	"StepLauncher/internal/Core/Launcher"
	linstance "StepLauncher/internal/Core/Launcher/Instance"
	"StepLauncher/internal/Core/Logger"
	"StepLauncher/internal/Core/ModLoader"
	"StepLauncher/internal/Core/ModLoader/Provider"
	"StepLauncher/internal/Core/Platform"
	lprofile "StepLauncher/internal/Core/Launcher/Profile"
)

type EventHandler func(eventType string, data []byte)

type Engine struct {
	config     *config.Manager
	log        *logger.Logger
	cache      *cache.Manager
	downloader *downloader.Manager
	launcher   *launcher.LaunchManager
	modloader  *modloader.Orchestrator
	history    *lhistory.Manager
	profiles   *lprofile.Manager
	instances  *linstance.InstanceManager
	sharedDl   *downloader.Manager

	eventCb EventHandler
}

type EngineInfo struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	Author          string `json:"author"`
	GoVersion       string `json:"goVersion"`
	OS              string `json:"os"`
	Arch            string `json:"arch"`
	NumCPU          int    `json:"numCpu"`
	NumGoroutine    int    `json:"numGoroutine"`
	LauncherName    string `json:"launcherName,omitempty"`
	LauncherVersion string `json:"launcherVersion,omitempty"`
}

func parseDuration(val string, fallback time.Duration) time.Duration {
	if val == "" {
		return fallback
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}
	return d
}

func NewEngine(opts ...Option) (*Engine, error) {
	e := &Engine{}

	for _, opt := range opts {
		opt(e)
	}

	cfgMgr := config.NewManager()
	if err := cfgMgr.Load(); err != nil {
		return nil, err
	}
	e.config = cfgMgr

	cfg := cfgMgr.Get()

	launcherName := cfg.LauncherName
	if launcherName == "" {
		launcherName = config.AppName
	}
	launcherVersion := cfg.LauncherVersion
	if launcherVersion == "" {
		launcherVersion = config.AppVersion
	}

	log, err := logger.New(cfg.LogDir, config.AppName, launcherName, launcherVersion)
	if err != nil {
		return nil, err
	}
	e.log = log

	log.System("%s v%s", config.AppName, config.AppVersion)
	log.System("Author: %s", config.AppAuthor)
	log.System("GoVersion: %s", runtime.Version())
	log.System("Runtime:  CPU: %d | GOMAXPROCS: %d | RAM: %dMB | OS: %s | Arch: %s",
		runtime.NumCPU(), runtime.GOMAXPROCS(0), platform.TotalRAMMB(), runtime.GOOS, runtime.GOARCH)
	log.System("  WorkDir: %s", cfg.WorkDir)
	log.System("  LogDir: %s", cfg.LogDir)
	log.System("  CacheDir: %s", cfg.CacheDir)
	log.System("========================================")

	e.log.SetBroadcastFn(func(t logger.Type, msg string) {
		if e.eventCb == nil { return }
		data, _ := json.Marshal(map[string]string{
			"type": "engine_log", "level": string(t), "message": msg,
		})
		e.eventCb("engine_log", data)
	})

	broadcastFn := func(data []byte) {
		if e.eventCb == nil { return }
		var evt struct { Type string `json:"type"` }
		if json.Unmarshal(data, &evt) == nil {
			e.eventCb(evt.Type, data)
		} else {
			e.eventCb("event", data)
		}
	}

	ttls := cache.DefaultTTLs()
	cfgTTL := cfgMgr.Get()
	if v := cfgTTL.CacheTTLManifest; v != "" { ttls.Manifest = parseDuration(v, ttls.Manifest) }
	if v := cfgTTL.CacheTTLAssets; v != "" { ttls.Assets = parseDuration(v, ttls.Assets) }
	if v := cfgTTL.CacheTTLVersions; v != "" { ttls.Versions = parseDuration(v, ttls.Versions) }
	if v := cfgTTL.CacheTTLModloader; v != "" { ttls.Modloader = parseDuration(v, ttls.Modloader) }
	if v := cfgTTL.CacheTTLJava; v != "" { ttls.Java = parseDuration(v, ttls.Java) }
	if v := cfgTTL.CacheTTLDefault; v != "" { ttls.Default = parseDuration(v, ttls.Default) }

	cacheMgr := cache.NewManager(filepath.Join(cfg.WorkDir, "cache"), ttls)
	e.cache = cacheMgr

	histMgr := lhistory.NewManager(cfg.WorkDir)
	if err := histMgr.Load(); err != nil {
		log.Warn("Failed to load history: %v", err)
	}
	e.history = histMgr

	profMgr := lprofile.NewManager(cfg.WorkDir)
	if err := profMgr.Load(); err != nil {
		log.Warn("Failed to load profiles: %v", err)
	}
	e.profiles = profMgr

	runtime.GOMAXPROCS(cfg.MaxCores)

	dlManager := downloader.NewManager(downloader.Config{
		WorkDir:      cfg.WorkDir,
		CacheDir:     filepath.Join(cfg.WorkDir, "cache"),
		CacheManager: cacheMgr,
		MaxRAM:       cfg.MaxRAMMB,
		LogFn:        func(f string, a ...interface{}) { log.Info(f, a...) },
		BroadcastFn:  broadcastFn,
		MaxMbps:      cfg.MaxMbps,
		MinMbps:      cfg.MinMbps,
	})
	e.downloader = dlManager
	httpClient := dlManager.HTTPClient()

	launchMgr := launcher.NewManager(launcher.ManagerConfig{
		WorkDir:         cfg.WorkDir,
		LogDir:          cfg.LogDir,
		LogFn:           func(f string, a ...interface{}) { log.Info(f, a...) },
		LauncherName:    launcherName,
		LauncherVersion: launcherVersion,
		OnGameExitFn: func(gi *launcher.GameInstance, playTimeSeconds int) {
			entry := lhistory.Entry{
				Version:         gi.Version,
				InstanceID:      gi.ID,
				PlayerName:      gi.PlayerName,
				PlayTimeSeconds: playTimeSeconds,
				Timestamp:       time.Now().Unix(),
				ExitCode:        gi.ExitCode,
				CrashReason:     gi.CrashReason,
			}
			if err := histMgr.AddEntry(entry); err != nil {
				log.Warn("Failed to record play history: %v", err)
			}
		},
		GameLogBroadcastFn: func(stream, line string) {
			if e.eventCb == nil { return }
			data, _ := json.Marshal(map[string]string{
				"type": "game_log", "level": stream, "message": line,
			})
			e.eventCb("game_log", data)
		},
		GameEventBroadcastFn: func(data []byte) {
			if e.eventCb == nil { return }
			var evt struct { Type string `json:"type"` }
			if json.Unmarshal(data, &evt) == nil {
				e.eventCb(evt.Type, data)
			} else {
				e.eventCb("game_event", data)
			}
		},
		GameEventReplayFn: func(data []byte) {
			if e.eventCb == nil { return }
			var evt struct { Type string `json:"type"` }
			if json.Unmarshal(data, &evt) == nil {
				e.eventCb(evt.Type, data)
			} else {
				e.eventCb("game_event_replay", data)
			}
		},
	})
	e.launcher = launchMgr

	sharedDlMgr := downloader.NewManager(downloader.Config{
		WorkDir:      filepath.Join(cfg.WorkDir, cfg.SharedDir),
		CacheDir:     filepath.Join(cfg.WorkDir, cfg.SharedDir, "cache"),
		CacheManager: cacheMgr,
		MaxRAM:       cfg.MaxRAMMB,
		LogFn:        func(f string, a ...interface{}) { log.Info(f, a...) },
		BroadcastFn:  broadcastFn,
		HTTPClient:   httpClient,
		MaxMbps:      cfg.MaxMbps,
		MinMbps:      cfg.MinMbps,
	})
	e.sharedDl = sharedDlMgr

	instancesDir := filepath.Join(cfg.WorkDir, cfg.InstancesDir)
	sharedDir := filepath.Join(cfg.WorkDir, cfg.SharedDir)
	instMgr := linstance.NewManager(instancesDir, sharedDir)
	instMgr.SetSharedDownloadManager(sharedDlMgr)
	instMgr.SetLaunchManager(launchMgr)
	instMgr.SetIdentity(cfg.LauncherName, cfg.LauncherVersion)
	instMgr.SetLogger(func(f string, a ...interface{}) { log.Info(f, a...) })
	e.instances = instMgr

	mlReg := modloader.NewRegistry()
	mlReg.Register(provider.NewFabricProvider(cfg.CacheDir, httpClient, cacheMgr))
	mlReg.Register(provider.NewQuiltProvider(cfg.CacheDir, httpClient, cacheMgr))
	mlReg.Register(provider.NewLegacyFabricProvider(cfg.CacheDir, httpClient, cacheMgr))
	mlReg.Register(provider.NewForgeProvider(cfg.CacheDir, httpClient, cacheMgr))
	mlReg.Register(provider.NewNeoForgeProvider(cfg.CacheDir, httpClient, cacheMgr))

	mlOrch := modloader.NewOrchestrator(
		cfg.WorkDir,
		filepath.Join(cfg.WorkDir, cfg.SharedDir),
		cfg.CacheDir,
		httpClient,
		mlReg,
		broadcastFn,
		func(f string, a ...interface{}) { log.Info(f, a...) },
	)
	e.modloader = mlOrch
	instMgr.SetModLoaderOrchestrator(mlOrch)

	return e, nil
}

func (e *Engine) SetEventCallback(cb EventHandler) {
	e.eventCb = cb
}

func (e *Engine) GetConfig() config.Config {
	return e.config.Get()
}

func (e *Engine) UpdateConfig(cfg config.Config) {
	e.config.UpdateConfig(cfg)
}

func (e *Engine) EngineInfo() EngineInfo {
	cfg := e.config.Get()
	return EngineInfo{
		Name:            config.AppName,
		Version:         config.AppVersion,
		Author:          config.AppAuthor,
		GoVersion:       runtime.Version(),
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		NumCPU:          runtime.NumCPU(),
		NumGoroutine:    runtime.NumGoroutine(),
		LauncherName:    cfg.LauncherName,
		LauncherVersion: cfg.LauncherVersion,
	}
}

func (e *Engine) Logger() *logger.Logger {
	return e.log
}

func (e *Engine) DownloadManager() *downloader.Manager {
	return e.downloader
}

func (e *Engine) LaunchManager() *launcher.LaunchManager {
	return e.launcher
}

func (e *Engine) CacheManager() *cache.Manager {
	return e.cache
}

func (e *Engine) ModLoader() *modloader.Orchestrator {
	return e.modloader
}

func (e *Engine) HistoryManager() *lhistory.Manager {
	return e.history
}

func (e *Engine) ProfileManager() *lprofile.Manager {
	return e.profiles
}

func (e *Engine) InstanceManager() *linstance.InstanceManager {
	return e.instances
}

type Option func(*Engine)

func WithEventCallback(cb EventHandler) Option {
	return func(e *Engine) {
		e.eventCb = cb
	}
}
