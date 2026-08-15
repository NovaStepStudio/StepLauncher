package engine

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	accounts "StepLauncher/internal/Core/Accounts"
	"StepLauncher/internal/Core/Cache"
	"StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Core/Launcher"
	helpers "StepLauncher/internal/Core/Launcher/Helpers"
	lhistory "StepLauncher/internal/Core/Launcher/History"
	linstance "StepLauncher/internal/Core/Launcher/Instance"
	lprofile "StepLauncher/internal/Core/Launcher/Profile"
	"StepLauncher/internal/Core/Logger"
	"StepLauncher/internal/Core/ModLoader"
	"StepLauncher/internal/Core/ModLoader/Provider"
	"StepLauncher/internal/Core/Platform"
	engineconfig "StepLauncher/internal/Handlers/Engine/engineconfig"
)

type EventHandler func(eventType string, data []byte)

type Engine struct {
	config     *engineconfig.Manager
	log        *logger.Logger
	cache      *cache.Manager
	downloader *downloader.Manager
	launcher   *launcher.LaunchManager
	modloader  *modloader.Orchestrator
	history    *lhistory.Manager
	crashHist  *lhistory.CrashManager
	profiles   *lprofile.Manager
	accounts   *accounts.Manager
	instances  *linstance.InstanceManager
	sharedDl   *downloader.Manager
	integrity  *integrityRunner

	eventCb EventHandler

	updateMu   sync.Mutex
	lastUpdate *UpdateInfo
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

	cfgMgr := engineconfig.NewManager()
	if err := cfgMgr.Load(); err != nil {
		return nil, err
	}
	e.config = cfgMgr

	cfg := cfgMgr.Get()

	launcherName := cfg.LauncherName
	if launcherName == "" {
		launcherName = engineconfig.AppName
	}
	launcherVersion := cfg.LauncherVersion
	if launcherVersion == "" {
		launcherVersion = engineconfig.AppVersion
	}

	log, err := logger.New(cfg.LogDir, engineconfig.AppName, launcherName, launcherVersion)
	if err != nil {
		return nil, err
	}
	e.log = log

	log.System("%s v%s", engineconfig.AppName, engineconfig.AppVersion)
	log.System("Author: %s", engineconfig.AppAuthor)
	log.System("GoVersion: %s", runtime.Version())
	log.System("Runtime:  CPU: %d | GOMAXPROCS: %d | RAM: %dMB | OS: %s | Arch: %s",
		runtime.NumCPU(), runtime.GOMAXPROCS(0), platform.TotalRAMMB(), runtime.GOOS, runtime.GOARCH)
	log.System("  WorkDir: %s", cfg.WorkDir)
	log.System("  LogDir: %s", cfg.LogDir)
	log.System("  CacheDir: %s", cfg.CacheDir)
	log.System("  Instances: %s (se crea al usar instancias)", filepath.Join(cfg.WorkDir, cfg.InstancesDir))
	log.System("  Shared: %s (se crea al usar instancias)", filepath.Join(cfg.WorkDir, cfg.SharedDir))
	log.System("  Config: cacheTTL manifest=%s | assets=%s | versions=%s | java=%s",
		cfg.CacheTTLManifest, cfg.CacheTTLAssets, cfg.CacheTTLVersions, cfg.CacheTTLJava)
	log.System("========================================")

	e.log.SetBroadcastFn(func(t logger.Type, msg string) {
		if e.eventCb == nil {
			return
		}
		data, _ := json.Marshal(map[string]string{
			"type": "engine_log", "level": string(t), "message": msg,
		})
		e.eventCb("engine_log", data)
	})

	broadcastFn := func(data []byte) {
		if e.eventCb == nil {
			return
		}
		var evt struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(data, &evt) == nil {
			e.eventCb(evt.Type, data)
		} else {
			e.eventCb("event", data)
		}
	}

	ttls := cache.DefaultTTLs()
	cfgTTL := cfgMgr.Get()
	if v := cfgTTL.CacheTTLManifest; v != "" {
		ttls.Manifest = parseDuration(v, ttls.Manifest)
	}
	if v := cfgTTL.CacheTTLAssets; v != "" {
		ttls.Assets = parseDuration(v, ttls.Assets)
	}
	if v := cfgTTL.CacheTTLVersions; v != "" {
		ttls.Versions = parseDuration(v, ttls.Versions)
	}
	if v := cfgTTL.CacheTTLModloader; v != "" {
		ttls.Modloader = parseDuration(v, ttls.Modloader)
	}
	if v := cfgTTL.CacheTTLJava; v != "" {
		ttls.Java = parseDuration(v, ttls.Java)
	}
	if v := cfgTTL.CacheTTLDefault; v != "" {
		ttls.Default = parseDuration(v, ttls.Default)
	}

	cacheMgr := cache.NewManager(filepath.Join(cfg.WorkDir, "cache"), ttls)
	e.cache = cacheMgr

	histMgr := lhistory.NewManager(cfg.WorkDir)
	if err := histMgr.Load(); err != nil {
		log.Warn("Failed to load history: %v", err)
	}
	if err := histMgr.EnsureFile(); err != nil {
		log.Warn("Failed to create history file: %v", err)
	}
	e.history = histMgr

	crashHistMgr := lhistory.NewCrashManager(cfg.WorkDir)
	if err := crashHistMgr.Load(); err != nil {
		log.Warn("Failed to load crash history: %v", err)
	}
	if err := crashHistMgr.EnsureFile(); err != nil {
		log.Warn("Failed to create crash history file: %v", err)
	}
	e.crashHist = crashHistMgr

	profMgr := lprofile.NewManager(cfg.WorkDir)
	if err := profMgr.Load(); err != nil {
		log.Warn("Failed to load profiles: %v", err)
	}
	if err := profMgr.EnsureFile(); err != nil {
		log.Warn("Failed to create profiles file: %v", err)
	}
	e.profiles = profMgr

	accMgr := accounts.NewManager(cfg.WorkDir)
	accMgr.SetLogFn(func(f string, a ...interface{}) { log.Info(f, a...) })
	accMgr.SetEventFn(func(et string, data []byte) {
		if e.eventCb == nil {
			return
		}
		e.eventCb(et, data)
	})
	if err := accMgr.Load(); err != nil {
		log.Warn("Failed to load accounts: %v", err)
	}
	if err := accMgr.Save(); err != nil {
		log.Warn("Failed to create accounts file: %v", err)
	}
	e.accounts = accMgr

	runtime.GOMAXPROCS(cfg.MaxCores)

	dlManager := downloader.NewManager(downloader.Config{
		WorkDir:      cfg.WorkDir,
		CacheDir:     filepath.Join(cfg.WorkDir, "cache"),
		CacheManager: cacheMgr,
		IDPrefix:     "ver-",
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
		SeparateGameDir: cfg.SeparateGameDirValue(),
		OnGameExitFn: func(gi *launcher.GameInstance, playTimeSeconds int) {
			entry := lhistory.Entry{
				Version:         gi.Version,
				InstanceID:      gi.ID,
				InstanceName:    gi.InstanceName,
				PlayerName:      gi.PlayerName,
				PlayTimeSeconds: playTimeSeconds,
				Timestamp:       time.Now().Unix(),
				ExitCode:        gi.ExitCode,
				CrashReason:     gi.CrashReason,
			}
			if err := histMgr.AddEntry(entry); err != nil {
				log.Warn("Failed to record play history: %v", err)
			}
			if gi.Status == launcher.GameCrashed {
				crash := lhistory.CrashEntry{
					Version:          gi.Version,
					ExitCode:         gi.ExitCode,
					CrashReason:      gi.CrashReason,
					CrashCategory:    gi.CrashCategory,
					LauncherLogPath:  relPathToWorkDir(cfg.WorkDir, e.log.GetLogPath()),
					MinecraftLogPath: relPathToWorkDir(cfg.WorkDir, gi.LogPath),
					JvmLogPath:       relPathToWorkDir(cfg.WorkDir, gi.CrashLog),
				}
				if err := crashHistMgr.AddEntry(crash); err != nil {
					log.Warn("Failed to record crash history: %v", err)
				} else {
					log.Info("Crash registrado en el historial: %s", crash.ID)
				}
			}
		},
		GameLogBroadcastFn: func(stream, line string) {
			if e.eventCb == nil {
				return
			}
			data, _ := json.Marshal(map[string]string{
				"type": "game_log", "level": stream, "message": line,
			})
			e.eventCb("game_log", data)
		},
		GameEventBroadcastFn: func(data []byte) {
			if e.eventCb == nil {
				return
			}
			var evt struct {
				Type string `json:"type"`
			}
			if json.Unmarshal(data, &evt) == nil {
				e.eventCb(evt.Type, data)
			} else {
				e.eventCb("game_event", data)
			}
		},
		GameEventReplayFn: func(data []byte) {
			if e.eventCb == nil {
				return
			}
			var evt struct {
				Type string `json:"type"`
			}
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
		CacheDir:     cfg.CacheDir,
		CacheManager: cacheMgr,
		IDPrefix:     "inst-",
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
	instMgr.SetCacheDir(cfg.CacheDir)
	instMgr.SetLaunchManager(launchMgr)
	instMgr.SetIdentity(cfg.LauncherName, cfg.LauncherVersion)
	instMgr.SetLogger(func(f string, a ...interface{}) { log.Info(f, a...) })
	e.instances = instMgr

	mlReg := modloader.NewRegistry()
	// Resolver de Java para los instaladores de Forge/NeoForge, por prioridad:
	// 1) el Java oficial que el launcher ya descargó para la versión base de MC
	//    (runtime/<component>/...) — es el Java exacto que Mojang eligió para
	//    esa versión y cumple el requisito del instalador correspondiente;
	// 2) el Java configurado en el launcher (JavaCustomPath) o el del sistema.
	//    El Java oficial de Mojang no se descarga en tiempo de instalación.
	runtimeDir := filepath.Join(cfg.WorkDir, "runtime")
	javaResolver := func(mcVersion, instancePath string) (string, error) {
		if javaPath, err := helpers.ResolveMinecraftJava(runtimeDir, instancePath, mcVersion); err == nil {
			return javaPath, nil
		}
		return helpers.ResolveJava("", "", false, cfg.JavaCustomPath)
	}
	mlReg.Register(provider.NewFabricProvider(cfg.CacheDir, httpClient, cacheMgr))
	mlReg.Register(provider.NewQuiltProvider(cfg.CacheDir, httpClient, cacheMgr))
	mlReg.Register(provider.NewLegacyFabricProvider(cfg.CacheDir, httpClient, cacheMgr))
	mlReg.Register(provider.NewForgeProvider(cfg.CacheDir, httpClient, cacheMgr, javaResolver))
	mlReg.Register(provider.NewNeoForgeProvider(cfg.CacheDir, httpClient, cacheMgr, javaResolver))

	mlOrch := modloader.NewOrchestrator(
		cfg.WorkDir,
		cfg.CacheDir,
		httpClient,
		mlReg,
		broadcastFn,
		func(f string, a ...interface{}) { log.Info(f, a...) },
	)
	e.modloader = mlOrch
	instMgr.SetModLoaderOrchestrator(mlOrch)

	e.integrity = &integrityRunner{}

	instMgr.SetOnVersionReady(func(name, version string) {
		go e.checkInstanceExistence(name, version)
	})

	// En modo Minecraft el gameDir es SIEMPRE el propio .minecraft.
	if cfgMgr.Bootstrap().Mode == engineconfig.ModeMinecraft {
		launchMgr.SetSeparateGameDir(false)
		instMgr.SetSeparateGameDir(false)
		log.System("Modo Minecraft: gameDir = %s (separado desactivado)", cfg.WorkDir)
	}

	return e, nil
}

func (e *Engine) SetEventCallback(cb EventHandler) {
	e.eventCb = cb
}

func (e *Engine) GetConfig() engineconfig.Config {
	return e.config.Get()
}

func (e *Engine) UpdateConfig(cfg engineconfig.Config) {
	e.config.UpdateConfig(cfg)
}

func (e *Engine) EngineInfo() EngineInfo {
	cfg := e.config.Get()
	return EngineInfo{
		Name:            engineconfig.AppName,
		Version:         engineconfig.AppVersion,
		Author:          engineconfig.AppAuthor,
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

func (e *Engine) CrashHistoryManager() *lhistory.CrashManager {
	return e.crashHist
}

func relPathToWorkDir(workDir, p string) string {
	if p == "" {
		return ""
	}
	r, err := filepath.Rel(workDir, p)
	if err != nil {
		return filepath.ToSlash(p)
	}
	return filepath.ToSlash(r)
}

func (e *Engine) ProfileManager() *lprofile.Manager {
	return e.profiles
}

func (e *Engine) AccountsManager() *accounts.Manager {
	return e.accounts
}

func (e *Engine) Shutdown() {
	if e.accounts != nil {
		e.accounts.CancelLogin()
	}
	e.StopAllGames()
	for _, mgr := range []*downloader.Manager{e.downloader, e.sharedDl} {
		if mgr != nil {
			mgr.CancelActive()
		}
	}
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
