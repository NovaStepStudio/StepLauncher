package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

const (
	AppName    = "StepLauncher"
	AppVersion = "2.3.0"
	AppAuthor  = "NovaStepStudio"
)

type Config struct {
	MaxCores  int    `json:"maxCores,omitempty"`
	MaxRAMMB  int    `json:"maxRam,omitempty"`
	CacheDir  string `json:"cacheDir,omitempty"`
	LogDir    string `json:"logDir,omitempty"`
	WorkDir   string `json:"workDir,omitempty"`

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

	HardwareEnabled     bool    `json:"hardwareEnabled"`
	HardwareAcceleration bool   `json:"hardwareAcceleration"`
	GPUType             string  `json:"gpuType"`
	GPUPreset           string  `json:"gpuPreset"`

	JavaMode        string `json:"javaMode"`
	JavaCustomPath  string `json:"javaCustomPath"`

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

	OfflineMode   bool `json:"offlineMode"`
	CompatMode    bool `json:"compatMode"`
	DetailedLogs  bool `json:"detailedLogs"`

	ConcurrentDownloads int `json:"concurrentDownloads"`
}

type Manager struct {
	cfg        Config
	configPath string
}

func DefaultConfig() Config {
	cores := runtime.NumCPU()
	if cores > 1 {
		cores--
	}
	return Config{
		MaxCores:  cores,
		MaxRAMMB:  2048,

		InstancesDir: "instances",
		SharedDir:    "shared",

		HardwareEnabled:      true,
		HardwareAcceleration: true,
		GPUType:             "",
		GPUPreset:           "",

		JavaMode: "auto",

		AuthVerify: true,

		WindowWidth:  854,
		WindowHeight: 480,
		Fullscreen:   false,

		ConcurrentDownloads: 4,
	}
}

func NewManager() *Manager {
	return &Manager{cfg: DefaultConfig()}
}

func defaultStepLauncherDir() string {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, ".StepLauncher")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".StepLauncher")
}

func (m *Manager) Load() error {
	m.cfg = DefaultConfig()

	if m.cfg.WorkDir == "" {
		m.cfg.WorkDir = defaultStepLauncherDir()
	}
	if m.cfg.CacheDir == "" {
		m.cfg.CacheDir = filepath.Join(m.cfg.WorkDir, "cache")
	}
	if m.cfg.LogDir == "" {
		m.cfg.LogDir = filepath.Join(m.cfg.WorkDir, "logs")
	}

	return m.ensureDirs()
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
	m.cfg = fileCfg
	m.configPath = path

	if m.cfg.WorkDir == "" {
		m.cfg.WorkDir = defaultStepLauncherDir()
	}
	if m.cfg.CacheDir == "" {
		m.cfg.CacheDir = filepath.Join(m.cfg.WorkDir, "cache")
	}
	if m.cfg.LogDir == "" {
		m.cfg.LogDir = filepath.Join(m.cfg.WorkDir, "logs")
	}

	return m.ensureDirs()
}

func (m *Manager) ensureDirs() error {
	dirs := []string{m.cfg.LogDir, m.cfg.WorkDir, m.cfg.CacheDir}
	for _, d := range dirs {
		if d != "" {
			if err := os.MkdirAll(d, 0755); err != nil {
				return err
			}
		}
	}
	instancesDir := filepath.Join(m.cfg.WorkDir, m.cfg.InstancesDir)
	sharedDir := filepath.Join(m.cfg.WorkDir, m.cfg.SharedDir)
	for _, d := range []string{instancesDir, sharedDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	for _, sub := range []string{"libraries", "assets", "runtime", "cache"} {
		if err := os.MkdirAll(filepath.Join(sharedDir, sub), 0755); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Join(sharedDir, "assets", "indexes"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(sharedDir, "assets", "objects"), 0755); err != nil {
		return err
	}
	return nil
}

func (m *Manager) Get() Config       { return m.cfg }
func (m *Manager) RootDir() string   { return m.cfg.WorkDir }

func (m *Manager) UpdateConfig(cfg Config) {
	m.cfg = cfg
	if m.configPath != "" {
		data, _ := json.MarshalIndent(cfg, "", "  ")
		os.WriteFile(m.configPath, data, 0644)
	}
}
