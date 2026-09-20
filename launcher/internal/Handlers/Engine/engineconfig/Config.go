package engineconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	AppName    = "StepLauncher"
	AppVersion = "2.5.0"
	AppAuthor  = "NovaStepStudio"
)

type Config struct {
	MaxCores int    `json:"maxCores,omitempty"`
	MaxRAMMB int    `json:"maxRam,omitempty"`
	CacheDir string `json:"cacheDir,omitempty"`
	LogDir   string `json:"logDir,omitempty"`
	WorkDir  string `json:"workDir,omitempty"`

	LauncherName    string `json:"launcherName,omitempty"`
	LauncherVersion string `json:"launcherVersion,omitempty"`

	InstancesDir string `json:"instancesDir,omitempty"`
	SharedDir    string `json:"sharedDir,omitempty"`

	MaxMbps float64 `json:"maxMbps,omitempty"`
	MinMbps float64 `json:"minMbps,omitempty"`

	CacheTTLManifest  string `json:"cacheTTLManifest,omitempty"`
	CacheTTLAssets    string `json:"cacheTTLAssets,omitempty"`
	CacheTTLVersions  string `json:"cacheTTLVersions,omitempty"`
	CacheTTLModloader string `json:"cacheTTLModloader,omitempty"`
	CacheTTLJava      string `json:"cacheTTLJava,omitempty"`
	CacheTTLDefault   string `json:"cacheTTLDefault,omitempty"`

	HardwareEnabled      bool   `json:"hardwareEnabled"`
	HardwareAcceleration bool   `json:"hardwareAcceleration"`
	GPUType              string `json:"gpuType"`
	GPUPreset            string `json:"gpuPreset"`

	JavaMode       string `json:"javaMode"`
	JavaCustomPath string `json:"javaCustomPath"`

	ProxyEnabled bool   `json:"proxyEnabled"`
	ProxyHost    string `json:"proxyHost"`
	ProxyPort    int    `json:"proxyPort"`
	ProxyUser    string `json:"proxyUser"`
	ProxyPass    string `json:"proxyPass"`

	AuthVerify bool `json:"authVerify"`

	WindowWidth  int  `json:"windowWidth"`
	WindowHeight int  `json:"windowHeight"`
	Fullscreen   bool `json:"fullscreen"`

	JavaArgs string `json:"javaArgs"`
	GameArgs string `json:"gameArgs"`

	OfflineMode  bool `json:"offlineMode"`
	CompatMode   bool `json:"compatMode"`
	DetailedLogs bool `json:"detailedLogs"`

	ConcurrentDownloads int `json:"concurrentDownloads"`

	VerifyIntegrity bool `json:"verifyIntegrity"`

	VerifyBeforeLaunch bool `json:"verifyBeforeLaunch"`

	// SeparateGameDir indica si el gameDir es <workDir>/game (true) o el
	// propio workDir (false). nil equivale a true. En modo Minecraft se
	// fuerza a false para usar .minecraft directamente como gameDir.
	SeparateGameDir *bool `json:"separateGameDir,omitempty"`
}

func (c Config) SeparateGameDirValue() bool {
	if c.SeparateGameDir == nil {
		return true
	}
	return *c.SeparateGameDir
}

type Manager struct {
	mu         sync.RWMutex
	cfg        Config
	configPath string
	bootstrap  Bootstrap
}

func DefaultConfig() Config {
	cores := runtime.NumCPU()
	if cores > 1 {
		cores--
	}
	return Config{
		MaxCores: cores,
		MaxRAMMB: 2048,

		InstancesDir: "instances",
		SharedDir:    "shared",

		HardwareEnabled:      true,
		HardwareAcceleration: true,
		GPUType:              "",
		GPUPreset:            "",

		JavaMode: "auto",

		AuthVerify: true,

		WindowWidth:  854,
		WindowHeight: 480,
		Fullscreen:   false,

		ConcurrentDownloads: 4,

		VerifyIntegrity: true,

		VerifyBeforeLaunch: true,
	}
}

func NewManager() *Manager {
	return &Manager{cfg: DefaultConfig(), bootstrap: LoadBootstrap()}
}

func (m *Manager) Load() error {
	cfg := DefaultConfig()
	bootstrap := LoadBootstrap()

	if cfg.WorkDir == "" {
		cfg.WorkDir = bootstrap.ResolveWorkDir()
	}
	if cfg.CacheDir == "" {
		cfg.CacheDir = filepath.Join(cfg.WorkDir, "cache")
	}
	if cfg.LogDir == "" {
		cfg.LogDir = filepath.Join(cfg.WorkDir, "logs")
	}

	if err := ensureDirs(cfg); err != nil {
		return err
	}
	m.mu.Lock()
	m.cfg = cfg
	m.bootstrap = bootstrap
	m.mu.Unlock()
	return nil
}

func (m *Manager) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return err
	}
	bootstrap := LoadBootstrap()

	if fileCfg.WorkDir == "" {
		fileCfg.WorkDir = bootstrap.ResolveWorkDir()
	}
	if fileCfg.CacheDir == "" {
		fileCfg.CacheDir = filepath.Join(fileCfg.WorkDir, "cache")
	}
	if fileCfg.LogDir == "" {
		fileCfg.LogDir = filepath.Join(fileCfg.WorkDir, "logs")
	}

	if err := ensureDirs(fileCfg); err != nil {
		return err
	}
	m.mu.Lock()
	m.cfg = fileCfg
	m.configPath = path
	m.bootstrap = bootstrap
	m.mu.Unlock()
	return nil
}

func ensureDirs(cfg Config) error {
	dirs := []string{cfg.LogDir, cfg.WorkDir, cfg.CacheDir}
	for _, d := range dirs {
		if d != "" {
			if err := os.MkdirAll(d, 0755); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

func (m *Manager) RootDir() string { return m.Get().WorkDir }

// Bootstrap devuelve la preferencia de directorio cargada.
func (m *Manager) Bootstrap() Bootstrap {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bootstrap
}

// SetBootstrap actualiza la preferencia de directorio en memoria y la persiste.
func (m *Manager) SetBootstrap(b Bootstrap) error {
	if err := SaveBootstrap(b); err != nil {
		return err
	}
	m.mu.Lock()
	m.bootstrap = b
	m.mu.Unlock()
	return nil
}

func (m *Manager) UpdateConfig(cfg Config) {
	m.mu.Lock()
	m.cfg = cfg
	configPath := m.configPath
	m.mu.Unlock()
	if configPath != "" {
		data, _ := json.MarshalIndent(cfg, "", "  ")
		_ = os.WriteFile(configPath, data, 0644)
	}
}
