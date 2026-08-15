package engineconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

const (
	AppName    = "StepLauncher"
	AppVersion = "2.3.1"
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
	}
}

func NewManager() *Manager {
	return &Manager{cfg: DefaultConfig(), bootstrap: LoadBootstrap()}
}

func (m *Manager) Load() error {
	m.cfg = DefaultConfig()
	m.bootstrap = LoadBootstrap()

	if m.cfg.WorkDir == "" {
		m.cfg.WorkDir = m.bootstrap.ResolveWorkDir()
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
	m.bootstrap = LoadBootstrap()

	if m.cfg.WorkDir == "" {
		m.cfg.WorkDir = m.bootstrap.ResolveWorkDir()
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
	return nil
}

func (m *Manager) Get() Config     { return m.cfg }
func (m *Manager) RootDir() string { return m.cfg.WorkDir }

// Bootstrap devuelve la preferencia de directorio cargada.
func (m *Manager) Bootstrap() Bootstrap { return m.bootstrap }

// SetBootstrap actualiza la preferencia de directorio en memoria y la persiste.
func (m *Manager) SetBootstrap(b Bootstrap) error {
	if err := SaveBootstrap(b); err != nil {
		return err
	}
	m.bootstrap = b
	return nil
}

func (m *Manager) UpdateConfig(cfg Config) {
	m.cfg = cfg
	if m.configPath != "" {
		data, _ := json.MarshalIndent(cfg, "", "  ")
		os.WriteFile(m.configPath, data, 0644)
	}
}
