package Config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	globalutils "StepLauncher/internal/Core/Utils"
)

type MinecraftConfig struct {
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

	// SeparateGameDir: true = <workDir>/game, false = workDir como gameDir
	// (nil equivale a true). En modo Minecraft se fuerza a false.
	SeparateGameDir *bool `json:"separateGameDir,omitempty"`
}

func (mc MinecraftConfig) SeparateGameDirValue() bool {
	if mc.SeparateGameDir == nil {
		return true
	}
	return *mc.SeparateGameDir
}

type LauncherConfig struct {
	MaxRAMGB              int     `json:"maxRamGB"`
	MaxMbps               float64 `json:"maxMbps"`
	ConcurrentDownloads   int     `json:"concurrentDownloads"`
	HideLauncherOnLaunch  bool    `json:"hideLauncherOnLaunch"`
	VerifyIntegrity       *bool   `json:"verifyIntegrity"`
	IntegritySector       string  `json:"integritySector"`
	CheckForUpdatesOnStart bool   `json:"checkForUpdatesOnStart"`
	LaunchAfterInstall    bool    `json:"launchAfterInstall"`
	VerifyBeforeLaunch    *bool   `json:"verifyBeforeLaunch"`
}

func (l LauncherConfig) VerifyEnabled() bool {
	if l.VerifyIntegrity == nil {
		return true
	}
	return *l.VerifyIntegrity
}

func (l LauncherConfig) VerifyBeforeLaunchEnabled() bool {
	if l.VerifyBeforeLaunch == nil {
		return true
	}
	return *l.VerifyBeforeLaunch
}

func boolPtr(b bool) *bool { return &b }
func floatPtr(v float64) *float64 { return &v }

func validIntegritySector(s string) bool {
	switch s {
	case "todo", "global", "instances":
		return true
	}
	return false
}

type IdleConfig struct {
	AutoCloseModals    bool `json:"autoCloseModals"`
	IdleMinutes        int  `json:"idleMinutes"`
	ConfigCheckEnabled bool `json:"configCheckEnabled"`
	ConfigCheckMinutes int  `json:"configCheckMinutes"`
}

type BackgroundConfig struct {
	Type            string   `json:"type"`
	ImagePath       string   `json:"imagePath"`
	VideoPath       string   `json:"videoPath"`
	DynamicImages   []string `json:"dynamicImages"`
	DynamicOrder    string   `json:"dynamicOrder"`
	DynamicInterval int      `json:"dynamicInterval"`
	ImageAuthor     string   `json:"imageAuthor,omitempty"`
	ImageModName    string   `json:"imageModName,omitempty"`
	ImageUrl        string   `json:"imageUrl,omitempty"`
}

type RichPresenceConfig struct {
	Enabled *bool `json:"enabled"`
}

func (r RichPresenceConfig) EnabledValue() bool {
	if r.Enabled == nil {
		return true
	}
	return *r.Enabled
}

type ThemeColors struct {
	Sidebar     string `json:"sidebar"`
	Modal       string `json:"modal"`
	Buttons     string `json:"buttons"`
	BorderModal string `json:"borderModal"`
	Border      string `json:"border"`
	Progress    string `json:"progress"`
	PlayButton    string `json:"playButton"`
	ButtonPrimary string `json:"buttonPrimary"`
	Error         string `json:"error"`
	Success       string `json:"success"`
	Tag           string `json:"tag"`
	Warning       string `json:"warning"`
}

type MusicConfig struct {
	Enabled      bool    `json:"enabled"`
	Position     string  `json:"position"`
	CoverStyle   string  `json:"coverStyle"`
	DiscRotation bool    `json:"discRotation"`
	Volume       float64 `json:"volume"`
}

type MusicPanelConfig struct {
	MusicFolder        string   `json:"musicFolder"`
	MusicFolders       []string `json:"musicFolders,omitempty"`
	CoverStyle         string   `json:"coverStyle"` // square | disc | huge
	ColorMode          string   `json:"colorMode"`  // vibrant | dominant | muted | least | random
	PageSize           int      `json:"pageSize"`
	ShowCovers         *bool    `json:"showCovers"`
	AllowAbsolute      *bool    `json:"allowAbsolute"`
	NowPlayingHuge     *bool    `json:"nowPlayingHuge,omitempty"`     // carátula enorme predomina
	NowPlayingCover    string   `json:"nowPlayingCover,omitempty"`    // square | disc | huge (específico Ahora suena)
	NowPlayingHugeSize *float64 `json:"nowPlayingHugeSize,omitempty"` // tamaño carátula predominante 15..30 rem
	AutoScan           string   `json:"autoScan,omitempty"`           // off | hourly | daily | onLaunch
	SMTCSource         string   `json:"smtcSource,omitempty"`         // background | library | auto
	CoverOpacity       *float64 `json:"coverOpacity,omitempty"`       // 0 transparente .. 1 opaco (fondo del panel Música ::after)
}

type Personalization struct {
	UIScale             int              `json:"uiScale"`
	Background          BackgroundConfig `json:"background"`
	BackgroundMusic     MusicConfig      `json:"backgroundMusic"`
	FontPrimary         string           `json:"fontPrimary"`
	FontSecondary       string           `json:"fontSecondary"`
	FontPrimaryColor    string           `json:"fontPrimaryColor"`
	FontSecondaryColor  string           `json:"fontSecondaryColor"`
	FontPrimarySize     float64          `json:"fontPrimarySize"`
	FontSecondarySize   float64          `json:"fontSecondarySize"`
	Colors              ThemeColors      `json:"colors"`
	RecentColors        []string         `json:"recentColors"`
	Animations          bool             `json:"animations"`
	Blur                bool             `json:"blur"`
	Shadows             bool             `json:"shadows"`
	TextShadow          bool             `json:"textShadow"`
	TextShadowIntensity float64          `json:"textShadowIntensity"`
}

func (m MusicPanelConfig) ShowCoversValue() bool {
	if m.ShowCovers == nil {
		return true
	}
	return *m.ShowCovers
}

func (m MusicPanelConfig) AllowAbsoluteValue() bool {
	if m.AllowAbsolute == nil {
		return true
	}
	return *m.AllowAbsolute
}

func (m MusicPanelConfig) CoverOpacityValue() float64 {
	if m.CoverOpacity == nil {
		return 1
	}
	v := *m.CoverOpacity
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (m MusicPanelConfig) NowPlayingHugeSizeValue() float64 {
	if m.NowPlayingHugeSize == nil {
		return 20
	}
	v := *m.NowPlayingHugeSize
	if v < 15 {
		return 15
	}
	if v > 30 {
		return 30
	}
	return v
}

const (
	FileAssets       = "launcher_assets.json"
	FileAccounts     = "launcher_accounts.json"
	FileHistory      = "launcher_history.json"
	FileProfiles     = "launcher_profiles.json"
	FileCrashHistory = "launcher_history_crashes.json"
	FilePlaylists    = "launcher_playlists.json"
	FileMusicHistory = "launcher_music_history.json"
)

const (
	ExtraKeyAssets       = "assets"
	ExtraKeyAccounts     = "accounts"
	ExtraKeyHistory      = "history"
	ExtraKeyProfiles     = "profiles"
	ExtraKeyCrashHistory = "crashHistory"
	ExtraKeyPlaylists    = "playlists"
	ExtraKeyMusicHistory = "musicHistory"
)

type ExtraData struct {
	Assets       string `json:"assets"`
	Accounts     string `json:"accounts"`
	History      string `json:"history"`
	Profiles     string `json:"profiles"`
	CrashHistory string `json:"crashHistory"`
	Playlists    string `json:"playlists"`
	MusicHistory string `json:"musicHistory"`
}

func (e *ExtraData) UnmarshalJSON(b []byte) error {
	type alias ExtraData
	var obj alias
	if err := json.Unmarshal(b, &obj); err == nil {
		*e = ExtraData(obj)
		return nil
	}
	var arr []string
	if err := json.Unmarshal(b, &arr); err != nil {
		return fmt.Errorf("extraData: formato invalido")
	}
	for _, n := range arr {
		switch n {
		case FileAssets:
			e.Assets = n
		case FileAccounts:
			e.Accounts = n
		case FileHistory:
			e.History = n
		case FileProfiles:
			e.Profiles = n
		case FileCrashHistory:
			e.CrashHistory = n
		case FilePlaylists:
			e.Playlists = n
		case FileMusicHistory:
			e.MusicHistory = n
		}
	}
	return nil
}

type Config struct {
	MinecraftConfig MinecraftConfig `json:"minecraftConfig"`
	Launcher        LauncherConfig  `json:"launcher"`
	Personalization Personalization `json:"personalization"`
	Idle            IdleConfig      `json:"idle"`
	RichPresence RichPresenceConfig `json:"richPresence"`
	MusicPanel   MusicPanelConfig   `json:"musicPanel"`
	ExtraData    ExtraData          `json:"extraData"`
	FirstLaunch    bool             `json:"firstLaunch"`
}

func Default() Config {
	return Config{
		MinecraftConfig: MinecraftConfig{
			HardwareEnabled:      false,
			HardwareAcceleration: false,
			GPUType:              "",
			GPUPreset:            "",
			JavaMode:             "official",
			AuthVerify:           true,
			WindowWidth:          854,
			WindowHeight:         480,
			SeparateGameDir:      boolPtr(true),
		},
		Launcher: LauncherConfig{
			MaxRAMGB:              2,
			MaxMbps:               0,
			ConcurrentDownloads:   4,
			HideLauncherOnLaunch:  true,
			VerifyIntegrity:       boolPtr(true),
			IntegritySector:       "todo",
			CheckForUpdatesOnStart: true,
			VerifyBeforeLaunch:    boolPtr(true),
		},		Idle: IdleConfig{
			AutoCloseModals:    false,
			IdleMinutes:        1,
			ConfigCheckEnabled: true,
			ConfigCheckMinutes: 3,
		},
		RichPresence: RichPresenceConfig{
			Enabled: boolPtr(true),
		},
		MusicPanel: MusicPanelConfig{
			CoverStyle: "square",
			ColorMode:  "vibrant",
			PageSize:   20,
			ShowCovers: boolPtr(true),
			AllowAbsolute: boolPtr(true),
			NowPlayingHuge: boolPtr(false),
			NowPlayingCover: "square",
			NowPlayingHugeSize: floatPtr(20),
			AutoScan: "off",
			SMTCSource: "auto",
			CoverOpacity: floatPtr(1),
		},
		ExtraData: ExtraData{
			Assets:       FileAssets,
			Accounts:     FileAccounts,
			History:      FileHistory,
			Profiles:     FileProfiles,
			CrashHistory: FileCrashHistory,
			Playlists:    FilePlaylists,
			MusicHistory: FileMusicHistory,
		},
		FirstLaunch: true,
		Personalization: Personalization{
			UIScale: 100,
			Background: BackgroundConfig{
				Type:            "none",
				DynamicOrder:    "sequential",
				DynamicInterval: 10,
			},
			BackgroundMusic: MusicConfig{
				Enabled:      false,
				Position:     "bottom-center",
				CoverStyle:   "disc",
				DiscRotation: true,
				Volume:       1,
			},
			FontPrimary:        "Lexend",
			FontSecondary:      "Inter",
			FontPrimaryColor:   "#ffffff",
			FontSecondaryColor: "#cfcfd6",
			FontPrimarySize:    1,
			FontSecondarySize:  1,
			Colors: ThemeColors{
				Sidebar:       "#0005",
				Modal:         "#111",
				Buttons:       "#111",
				BorderModal:   "#494949",
				Border:        "rgba(37, 37, 37, 0.3)",
				Progress:      "#5ed89a",
				PlayButton:    "#111",
				ButtonPrimary: "#111",
				Error:         "#ff6b6b",
				Success:       "#5ed89a",
				Tag:           "#a974ff",
				Warning:       "#ffb347",
			},
			Animations:          true,
			Blur:                true,
			Shadows:             true,
			TextShadow:          false,
			TextShadowIntensity: 1,
		},
	}
}

type Manager struct {
	mu         sync.RWMutex
	cfg        Config
	configPath string
	logFn      func(format string, args ...interface{})
}

func (m *Manager) SetLogFn(fn func(format string, args ...interface{})) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logFn = fn
}

func (m *Manager) logf(format string, args ...interface{}) {
	m.mu.RLock()
	fn := m.logFn
	m.mu.RUnlock()
	if fn != nil {
		fn("[Config] "+format, args...)
	}
}

func NewManager(path string) *Manager {
	m := &Manager{cfg: Default(), configPath: path}
	m.load()
	if _, err := os.Stat(path); err != nil {
		m.Save()
	}
	return m
}

func (m *Manager) load() {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		m.logf("Configuracion cargada: no existe archivo, usando valores por defecto (%s)", m.configPath)
		return
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		m.logf("WARN: configuracion corrupta (%s), usando valores por defecto: %v", m.configPath, err)
		return
	}
	if _, ok := raw["minecraftConfig"]; ok {
		var cfg Config
		if json.Unmarshal(data, &cfg) != nil {
			m.logf("WARN: configuracion invalida (%s), usando valores por defecto", m.configPath)
			return
		}
		m.cfg = cfg
		m.sanitize()
		m.Save()
		m.logf("Configuracion cargada: %s | uiScale=%d%% | background=%s | downloads=%d | proxy=%v | idle=%v/%dmin check=%v/%dmin",
			m.configPath, m.cfg.Personalization.UIScale, m.cfg.Personalization.Background.Type,
			m.cfg.Launcher.ConcurrentDownloads, m.cfg.MinecraftConfig.ProxyEnabled,
			m.cfg.Idle.AutoCloseModals, m.cfg.Idle.IdleMinutes,
			m.cfg.Idle.ConfigCheckEnabled, m.cfg.Idle.ConfigCheckMinutes)
		return
	}
	m.migrateLegacy(data)
}

func (m *Manager) migrateLegacy(data []byte) {
	var legacy struct {
		MaxRAMMB             int     `json:"maxRam"`
		MaxMbps              float64 `json:"maxMbps"`
		ConcurrentDownloads  int     `json:"concurrentDownloads"`
		UIScale              int     `json:"uiScale"`
		HardwareEnabled      bool    `json:"hardwareEnabled"`
		HardwareAcceleration bool    `json:"hardwareAcceleration"`
		GPUType              string  `json:"gpuType"`
		GPUPreset            string  `json:"gpuPreset"`
		JavaMode             string  `json:"javaMode"`
		JavaCustomPath       string  `json:"javaCustomPath"`
		ProxyEnabled         bool    `json:"proxyEnabled"`
		ProxyHost            string  `json:"proxyHost"`
		ProxyPort            int     `json:"proxyPort"`
		ProxyUser            string  `json:"proxyUser"`
		ProxyPass            string  `json:"proxyPass"`
		AuthVerify           bool    `json:"authVerify"`
		WindowWidth          int     `json:"windowWidth"`
		WindowHeight         int     `json:"windowHeight"`
		Fullscreen           bool    `json:"fullscreen"`
		JavaArgs             string  `json:"javaArgs"`
		GameArgs             string  `json:"gameArgs"`
		OfflineMode          bool    `json:"offlineMode"`
		CompatMode           bool    `json:"compatMode"`
		DetailedLogs         bool    `json:"detailedLogs"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return
	}
	cfg := Default()
	cfg.MinecraftConfig.HardwareEnabled = legacy.HardwareEnabled
	cfg.MinecraftConfig.HardwareAcceleration = legacy.HardwareAcceleration
	cfg.MinecraftConfig.GPUType = legacy.GPUType
	cfg.MinecraftConfig.GPUPreset = legacy.GPUPreset
	cfg.MinecraftConfig.JavaMode = legacy.JavaMode
	cfg.MinecraftConfig.JavaCustomPath = legacy.JavaCustomPath
	cfg.MinecraftConfig.ProxyEnabled = legacy.ProxyEnabled
	cfg.MinecraftConfig.ProxyHost = legacy.ProxyHost
	cfg.MinecraftConfig.ProxyPort = legacy.ProxyPort
	cfg.MinecraftConfig.ProxyUser = legacy.ProxyUser
	cfg.MinecraftConfig.ProxyPass = legacy.ProxyPass
	cfg.MinecraftConfig.AuthVerify = legacy.AuthVerify
	cfg.MinecraftConfig.WindowWidth = legacy.WindowWidth
	cfg.MinecraftConfig.WindowHeight = legacy.WindowHeight
	cfg.MinecraftConfig.Fullscreen = legacy.Fullscreen
	cfg.MinecraftConfig.JavaArgs = legacy.JavaArgs
	cfg.MinecraftConfig.GameArgs = legacy.GameArgs
	cfg.MinecraftConfig.OfflineMode = legacy.OfflineMode
	cfg.MinecraftConfig.CompatMode = legacy.CompatMode
	cfg.MinecraftConfig.DetailedLogs = legacy.DetailedLogs
	if legacy.MaxRAMMB > 0 {
		cfg.Launcher.MaxRAMGB = legacy.MaxRAMMB / 1024
	}
	cfg.Launcher.MaxMbps = legacy.MaxMbps
	cfg.Launcher.ConcurrentDownloads = legacy.ConcurrentDownloads
	if legacy.UIScale >= 50 && legacy.UIScale <= 200 {
		cfg.Personalization.UIScale = legacy.UIScale
	}
	m.cfg = cfg
	m.sanitize()
	m.logf("Configuracion legada migrada al formato nuevo (uiScale=%d%%, downloads=%d)", cfg.Personalization.UIScale, cfg.Launcher.ConcurrentDownloads)
}

func (m *Manager) sanitize() {
	c := &m.cfg
	if c.Launcher.MaxRAMGB < 1 {
		c.Launcher.MaxRAMGB = 2
	}
	if c.Launcher.ConcurrentDownloads < 1 {
		c.Launcher.ConcurrentDownloads = 4
	}
	if c.Launcher.VerifyIntegrity == nil {
		c.Launcher.VerifyIntegrity = boolPtr(true)
	}
	if c.Launcher.VerifyBeforeLaunch == nil {
		c.Launcher.VerifyBeforeLaunch = boolPtr(true)
	}
	if !validIntegritySector(c.Launcher.IntegritySector) {
		c.Launcher.IntegritySector = "todo"
	}
	if c.Idle.IdleMinutes < 1 || c.Idle.IdleMinutes > 30 {
		c.Idle.IdleMinutes = 1
	}
	if c.Idle.ConfigCheckMinutes < 1 || c.Idle.ConfigCheckMinutes > 30 {
		c.Idle.ConfigCheckMinutes = 3
	}
	if c.RichPresence.Enabled == nil {
		c.RichPresence.Enabled = boolPtr(true)
	}
	if c.Personalization.UIScale < 50 || c.Personalization.UIScale > 200 {
		c.Personalization.UIScale = 100
	}
	bg := &c.Personalization.Background
	switch bg.Type {
	case "image", "video", "dynamic":
	default:
		bg.Type = "none"
	}
	if len(bg.DynamicImages) > 10 {
		bg.DynamicImages = bg.DynamicImages[:10]
	}
	if bg.DynamicOrder != "sequential" && bg.DynamicOrder != "random" {
		bg.DynamicOrder = "sequential"
	}
	if bg.DynamicInterval < 3 {
		bg.DynamicInterval = 10
	}
	if bg.DynamicInterval > 300 {
		bg.DynamicInterval = 300
	}
	music := &c.Personalization.BackgroundMusic
	// Única posición soportada: abajo en el centro.
	music.Position = "bottom-center"
	if music.CoverStyle != "square" && music.CoverStyle != "background" {
		music.CoverStyle = "disc"
	}
	if music.Volume < 0 || music.Volume > 1 {
		music.Volume = 1
	}
	mp := &c.MusicPanel
	if mp.CoverStyle != "disc" && mp.CoverStyle != "huge" {
		mp.CoverStyle = "square"
	}
	if mp.NowPlayingCover != "" && mp.NowPlayingCover != "square" && mp.NowPlayingCover != "disc" && mp.NowPlayingCover != "huge" {
		mp.NowPlayingCover = "square"
	}
	if mp.ColorMode != "dominant" && mp.ColorMode != "muted" && mp.ColorMode != "least" && mp.ColorMode != "random" {
		mp.ColorMode = "vibrant"
	}
	if mp.PageSize < 10 || mp.PageSize > 100 {
		mp.PageSize = 20
	}
	if mp.ShowCovers == nil {
		mp.ShowCovers = boolPtr(true)
	}
	if mp.AllowAbsolute == nil {
		mp.AllowAbsolute = boolPtr(true)
	}
	if mp.NowPlayingHuge == nil {
		mp.NowPlayingHuge = boolPtr(false)
	}
	if mp.NowPlayingHugeSize == nil {
		mp.NowPlayingHugeSize = floatPtr(20)
	} else {
		v := *mp.NowPlayingHugeSize
		if v < 15 {
			v = 15
		}
		if v > 30 {
			v = 30
		}
		mp.NowPlayingHugeSize = &v
	}
	if mp.AutoScan != "hourly" && mp.AutoScan != "daily" && mp.AutoScan != "onLaunch" {
		mp.AutoScan = "off"
	}
	if mp.SMTCSource != "background" && mp.SMTCSource != "library" {
		mp.SMTCSource = "auto"
	}
	// Opacidad del fondo del panel Música (::after con --panel-cover-color)
	if mp.CoverOpacity == nil {
		mp.CoverOpacity = floatPtr(1)
	} else {
		v := *mp.CoverOpacity
		if v < 0 {
			v = 0
		}
		if v > 1 {
			v = 1
		}
		mp.CoverOpacity = &v
	}
	// Migración y normalización multi-carpeta
	// PROHIBIDO: nunca crear carpeta por defecto — solo migra MusicFolder existente si hay valor explícito
	if len(mp.MusicFolders) == 0 && strings.TrimSpace(mp.MusicFolder) != "" {
		mp.MusicFolders = []string{filepath.Clean(strings.TrimSpace(mp.MusicFolder))}
	}
	// Normalizar MusicFolders: limpiar, deduplicar, quitar vacíos — sin inyectar carpeta por defecto
	seen := map[string]bool{}
	var cleaned []string
	for _, f := range mp.MusicFolders {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		c := filepath.Clean(f)
		if c == "." {
			continue
		}
		if seen[c] {
			continue
		}
		seen[c] = true
		cleaned = append(cleaned, c)
	}
	mp.MusicFolders = cleaned
	// Mantener MusicFolder sincronizado con el primero para compatibilidad
	// PROHIBIDO: si no hay carpetas, queda vacío — nunca asignar carpeta por defecto
	if len(mp.MusicFolders) > 0 {
		mp.MusicFolder = mp.MusicFolders[0]
	} else {
		mp.MusicFolder = ""
	}
	if c.Personalization.FontPrimary == "" {
		c.Personalization.FontPrimary = "Lexend"
	}
	if c.Personalization.FontSecondary == "" {
		c.Personalization.FontSecondary = "Inter"
	}
	sanitizeColor(&c.Personalization.FontPrimaryColor, "#ffffff")
	sanitizeColor(&c.Personalization.FontSecondaryColor, "#cfcfd6")
	if c.Personalization.FontPrimarySize < 0.5 || c.Personalization.FontPrimarySize > 2 {
		c.Personalization.FontPrimarySize = 1
	}
	if c.Personalization.FontSecondarySize < 0.5 || c.Personalization.FontSecondarySize > 2 {
		c.Personalization.FontSecondarySize = 1
	}
	if c.Personalization.TextShadowIntensity < 0.5 || c.Personalization.TextShadowIntensity > 2 {
		c.Personalization.TextShadowIntensity = 1
	}
	sanitizeColor(&c.Personalization.Colors.Sidebar, "#0005")
	sanitizeColor(&c.Personalization.Colors.Modal, "#111")
	sanitizeColor(&c.Personalization.Colors.Buttons, "#111")
	sanitizeColor(&c.Personalization.Colors.BorderModal, "#494949")
	sanitizeColor(&c.Personalization.Colors.Border, "rgba(37, 37, 37, 0.3)")
	sanitizeColor(&c.Personalization.Colors.Progress, "#5ed89a")
	sanitizeColor(&c.Personalization.Colors.PlayButton, "#111")
	sanitizeColor(&c.Personalization.Colors.ButtonPrimary, "#111")
	sanitizeColor(&c.Personalization.Colors.Error, "#ff6b6b")
	sanitizeColor(&c.Personalization.Colors.Success, "#5ed89a")
	sanitizeColor(&c.Personalization.Colors.Tag, "#a974ff")
	sanitizeColor(&c.Personalization.Colors.Warning, "#ffb347")
	keep := c.Personalization.RecentColors[:0]
	for _, col := range c.Personalization.RecentColors {
		if sanitizeColorString(col) == "" {
			continue
		}
		keep = append(keep, col)
	}
	if len(keep) > 12 {
		keep = keep[len(keep)-12:]
	}
	c.Personalization.RecentColors = keep
	if c.MinecraftConfig.JavaMode == "" {
		c.MinecraftConfig.JavaMode = "auto"
	}
	if c.MinecraftConfig.WindowWidth <= 0 {
		c.MinecraftConfig.WindowWidth = 854
	}
	if c.MinecraftConfig.WindowHeight <= 0 {
		c.MinecraftConfig.WindowHeight = 480
	}
	extra := &c.ExtraData
	normalizeExtraFile(&extra.Assets, FileAssets)
	normalizeExtraFile(&extra.Accounts, FileAccounts)
	normalizeExtraFile(&extra.History, FileHistory)
	normalizeExtraFile(&extra.Profiles, FileProfiles)
	normalizeExtraFile(&extra.CrashHistory, FileCrashHistory)
	normalizeExtraFile(&extra.Playlists, FilePlaylists)
	normalizeExtraFile(&extra.MusicHistory, FileMusicHistory)
}

func normalizeExtraFile(v *string, def string) {
	name := strings.TrimSpace(*v)
	if name == "" || !strings.HasPrefix(name, "launcher_") || !strings.HasSuffix(name, ".json") {
		*v = def
	}
}

func sanitizeColor(v *string, fallback string) {
	clean := sanitizeColorString(*v)
	if clean == "" {
		*v = fallback
		return
	}
	*v = clean
}

func sanitizeColorString(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if v[0] == '#' {
		if len(v) == 5 || len(v) == 7 || len(v) == 9 {
			for _, r := range v[1:] {
				if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F') {
					return ""
				}
			}
			return v
		}
		return ""
	}
	lower := strings.ToLower(v)
	open := strings.IndexByte(lower, '(')
	if open < 3 || !strings.HasSuffix(lower, ")") {
		return ""
	}
	name := lower[:open]
	body := lower[open+1 : len(lower)-1]
	parts := strings.Split(body, ",")
	if len(parts) != 4 && len(parts) != 3 {
		return ""
	}
	rgb := make([]int, 3)
	for i := 0; i < 3; i++ {
		n, err := strconv.Atoi(strings.TrimSpace(parts[i]))
		if err != nil || n < 0 || n > 255 {
			return ""
		}
		rgb[i] = n
	}
	switch name {
	case "rgb":
		if len(parts) != 3 {
			return ""
		}
		return fmt.Sprintf("rgb(%d,%d,%d)", rgb[0], rgb[1], rgb[2])
	case "rgba":
		if len(parts) != 4 {
			return ""
		}
		a, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		if err != nil || a < 0 || a > 1 {
			return ""
		}
		return fmt.Sprintf("rgba(%d,%d,%d,%.2f)", rgb[0], rgb[1], rgb[2], a)
	}
	return ""
}

func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

func (m *Manager) Path() string {
	return m.configPath
}

func (m *Manager) Save() error {
	m.mu.Lock()
	saved := false
	defer func() {
		m.mu.Unlock()
		if saved {
			m.logf("Configuracion guardada: %s", m.configPath)
		}
	}()
	if m.configPath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m.cfg, "", "  ")
	if err != nil {
		return err
	}
	if _, err := globalutils.SafeWriteFile(m.configPath, append(data, '\n'), 0644); err != nil {
		return err
	}
	saved = true
	return nil
}

func (m *Manager) UpdateMinecraft(mc MinecraftConfig) error {
	m.mu.Lock()
	cur := m.cfg.MinecraftConfig
	mc.ProxyEnabled = cur.ProxyEnabled
	mc.ProxyHost = cur.ProxyHost
	mc.ProxyPort = cur.ProxyPort
	mc.ProxyUser = cur.ProxyUser
	mc.ProxyPass = cur.ProxyPass
	mc.AuthVerify = cur.AuthVerify
	m.cfg.MinecraftConfig = mc
	m.sanitize()
	m.mu.Unlock()
	m.logf("Configuracion Minecraft actualizada: javaMode=%s javaCustom=%s window=%dx%d fullscreen=%v offline=%v compat=%v",
		mc.JavaMode, mc.JavaCustomPath, mc.WindowWidth, mc.WindowHeight, mc.Fullscreen, mc.OfflineMode, mc.CompatMode)
	return m.Save()
}

func (m *Manager) SetAuthVerify(verify bool) error {
	m.mu.Lock()
	m.cfg.MinecraftConfig.AuthVerify = verify
	m.mu.Unlock()
	m.logf("AuthVerify -> %v", verify)
	return m.Save()
}

func (m *Manager) SetProxy(enabled bool, host string, port int, user, pass string) error {
	m.mu.Lock()
	m.cfg.MinecraftConfig.ProxyEnabled = enabled
	m.cfg.MinecraftConfig.ProxyHost = host
	m.cfg.MinecraftConfig.ProxyPort = port
	m.cfg.MinecraftConfig.ProxyUser = user
	m.cfg.MinecraftConfig.ProxyPass = pass
	m.mu.Unlock()
	m.logf("Proxy -> enabled=%v host=%s:%d user=%s", enabled, host, port, user)
	return m.Save()
}

func (m *Manager) SetMaxRAMGB(gb int) error {
	m.mu.Lock()
	if gb < 1 {
		gb = 2
	}
	m.cfg.Launcher.MaxRAMGB = gb
	m.mu.Unlock()
	m.logf("MaxRAM -> %dGB", gb)
	return m.Save()
}

func (m *Manager) SetMaxMbps(mbps float64) error {
	m.mu.Lock()
	m.cfg.Launcher.MaxMbps = mbps
	m.mu.Unlock()
	m.logf("MaxMbps -> %.1f", mbps)
	return m.Save()
}

func (m *Manager) SetConcurrentDownloads(n int) error {
	m.mu.Lock()
	if n < 1 {
		n = 1
	}
	m.cfg.Launcher.ConcurrentDownloads = n
	m.mu.Unlock()
	m.logf("ConcurrentDownloads -> %d", n)
	return m.Save()
}

func (m *Manager) SetHideLauncher(v bool) error {
	m.mu.Lock()
	m.cfg.Launcher.HideLauncherOnLaunch = v
	m.mu.Unlock()
	m.logf("HideLauncherOnLaunch -> %v", v)
	return m.Save()
}

func (m *Manager) SetVerifyIntegrity(v bool) error {
	val := v
	m.mu.Lock()
	m.cfg.Launcher.VerifyIntegrity = &val
	m.mu.Unlock()
	m.logf("VerifyIntegrity -> %v", val)
	return m.Save()
}

func (m *Manager) SetVerifyBeforeLaunch(v bool) error {
	val := v
	m.mu.Lock()
	m.cfg.Launcher.VerifyBeforeLaunch = &val
	m.mu.Unlock()
	m.logf("VerifyBeforeLaunch -> %v", val)
	return m.Save()
}

func (m *Manager) SetIntegritySector(s string) error {
	if !validIntegritySector(s) {
		s = "todo"
	}
	m.mu.Lock()
	m.cfg.Launcher.IntegritySector = s
	m.mu.Unlock()
	m.logf("IntegritySector -> %s", s)
	return m.Save()
}

func (m *Manager) SetCheckForUpdatesOnStart(v bool) error {
	m.mu.Lock()
	m.cfg.Launcher.CheckForUpdatesOnStart = v
	m.mu.Unlock()
	m.logf("CheckForUpdatesOnStart -> %v", v)
	return m.Save()
}

func (m *Manager) SetLaunchAfterInstall(v bool) error {
	m.mu.Lock()
	m.cfg.Launcher.LaunchAfterInstall = v
	m.mu.Unlock()
	m.logf("LaunchAfterInstall -> %v", v)
	return m.Save()
}

func (m *Manager) SetRichPresenceEnabled(v bool) error {
	val := v
	m.mu.Lock()
	m.cfg.RichPresence.Enabled = &val
	m.mu.Unlock()
	m.logf("RichPresence -> %v", val)
	return m.Save()
}

func (m *Manager) SetUIScale(percent int) error {
	m.mu.Lock()
	if percent < 50 {
		percent = 50
	}
	if percent > 200 {
		percent = 200
	}
	m.cfg.Personalization.UIScale = percent
	m.mu.Unlock()
	m.logf("UIScale -> %d%%", percent)
	return m.Save()
}

func (m *Manager) UpdateIdle(idle IdleConfig) error {
	m.mu.Lock()
	m.cfg.Idle = idle
	m.sanitize()
	m.mu.Unlock()
	m.logf("Idle/check actualizado: autoClose=%v/%dmin check=%v/%dmin",
		idle.AutoCloseModals, idle.IdleMinutes, idle.ConfigCheckEnabled, idle.ConfigCheckMinutes)
	return m.Save()
}

func (m *Manager) GetMusicPanel() MusicPanelConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg.MusicPanel
}

func (m *Manager) UpdateMusicPanel(p MusicPanelConfig) error {
	m.mu.Lock()
	m.cfg.MusicPanel = p
	m.sanitize()
	m.mu.Unlock()
	m.logf("MusicPanel actualizado: folder=%s cover=%s color=%s pageSize=%d", p.MusicFolder, p.CoverStyle, p.ColorMode, p.PageSize)
	return m.Save()
}

func (m *Manager) SetMusicFolder(folder string) error {
	m.mu.Lock()
	folder = strings.TrimSpace(folder)
	if folder == "" {
		m.cfg.MusicPanel.MusicFolders = []string{}
		m.cfg.MusicPanel.MusicFolder = ""
	} else {
		// Compatibilidad: reemplazar con lista de una sola entrada
		clean := filepath.Clean(folder)
		m.cfg.MusicPanel.MusicFolders = []string{clean}
		m.cfg.MusicPanel.MusicFolder = clean
	}
	m.sanitize()
	m.mu.Unlock()
	m.logf("MusicFolder -> %s", folder)
	return m.Save()
}

func (m *Manager) GetMusicFolders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, len(m.cfg.MusicPanel.MusicFolders))
	copy(out, m.cfg.MusicPanel.MusicFolders)
	return out
}

func (m *Manager) SetMusicFolders(folders []string) error {
	m.mu.Lock()
	// Normalizar
	seen := map[string]bool{}
	var cleaned []string
	for _, f := range folders {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		c := filepath.Clean(f)
		if c == "." {
			continue
		}
		if seen[c] {
			continue
		}
		seen[c] = true
		cleaned = append(cleaned, c)
	}
	m.cfg.MusicPanel.MusicFolders = cleaned
	if len(cleaned) > 0 {
		m.cfg.MusicPanel.MusicFolder = cleaned[0]
	} else {
		m.cfg.MusicPanel.MusicFolder = ""
	}
	m.sanitize()
	m.mu.Unlock()
	m.logf("MusicFolders -> %v", cleaned)
	return m.Save()
}

func (m *Manager) AddMusicFolder(folder string) error {
	folder = strings.TrimSpace(folder)
	if folder == "" {
		return fmt.Errorf("carpeta vacía")
	}
	clean := filepath.Clean(folder)
	m.mu.Lock()
	for _, f := range m.cfg.MusicPanel.MusicFolders {
		if f == clean {
			m.mu.Unlock()
			return nil
		}
	}
	m.cfg.MusicPanel.MusicFolders = append(m.cfg.MusicPanel.MusicFolders, clean)
	if len(m.cfg.MusicPanel.MusicFolders) == 1 {
		m.cfg.MusicPanel.MusicFolder = clean
	}
	m.sanitize()
	m.mu.Unlock()
	m.logf("MusicFolder añadida -> %s", clean)
	return m.Save()
}

func (m *Manager) RemoveMusicFolder(folder string) error {
	folder = strings.TrimSpace(folder)
	if folder == "" {
		return fmt.Errorf("carpeta vacía")
	}
	clean := filepath.Clean(folder)
	m.mu.Lock()
	var out []string
	for _, f := range m.cfg.MusicPanel.MusicFolders {
		if f != clean {
			out = append(out, f)
		}
	}
	m.cfg.MusicPanel.MusicFolders = out
	if len(out) > 0 {
		m.cfg.MusicPanel.MusicFolder = out[0]
	} else {
		m.cfg.MusicPanel.MusicFolder = ""
	}
	m.sanitize()
	m.mu.Unlock()
	m.logf("MusicFolder eliminada -> %s", clean)
	return m.Save()
}

func (m *Manager) SetFirstLaunchDone() error {
	m.mu.Lock()
	m.cfg.FirstLaunch = false
	m.mu.Unlock()
	m.logf("Primer inicio completado (onboarding cerrado)")
	return m.Save()
}

func (m *Manager) RegisterExtraFile(key, name string) error {
	m.mu.Lock()
	if name == "" || !strings.HasPrefix(name, "launcher_") || !strings.HasSuffix(name, ".json") {
		m.mu.Unlock()
		return nil
	}
	switch key {
	case ExtraKeyAssets:
		m.cfg.ExtraData.Assets = name
	case ExtraKeyAccounts:
		m.cfg.ExtraData.Accounts = name
	case ExtraKeyHistory:
		m.cfg.ExtraData.History = name
	case ExtraKeyProfiles:
		m.cfg.ExtraData.Profiles = name
	case ExtraKeyCrashHistory:
		m.cfg.ExtraData.CrashHistory = name
	case ExtraKeyPlaylists:
		m.cfg.ExtraData.Playlists = name
	case ExtraKeyMusicHistory:
		m.cfg.ExtraData.MusicHistory = name
	default:
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()
	m.logf("ExtraData registrado: %s=%s", key, name)
	return m.Save()
}

func (m *Manager) ResetFontIfMatches(primary, secundary string) error {
	m.mu.Lock()
	changed := false
	if primary != "" && m.cfg.Personalization.FontPrimary == primary {
		m.cfg.Personalization.FontPrimary = "Lexend"
		changed = true
	}
	if secundary != "" && m.cfg.Personalization.FontSecondary == secundary {
		m.cfg.Personalization.FontSecondary = "Inter"
		changed = true
	}
	m.mu.Unlock()
	if !changed {
		return nil
	}
	m.logf("Tipografia de config restaurada a la de defecto (fuente eliminada)")
	return m.Save()
}

func (m *Manager) UpdatePersonalization(p Personalization) error {
	m.mu.Lock()
	p.UIScale = m.cfg.Personalization.UIScale
	hist := m.cfg.Personalization.RecentColors
	for _, col := range []string{
		p.Colors.Sidebar, p.Colors.Modal, p.Colors.Buttons, p.Colors.BorderModal, p.Colors.Border, p.Colors.Progress,
		p.Colors.PlayButton, p.Colors.ButtonPrimary,
		p.Colors.Error, p.Colors.Success, p.Colors.Tag, p.Colors.Warning,
	} {
		clean := sanitizeColorString(col)
		if clean == "" {
			continue
		}
		dup := false
		for _, h := range hist {
			if strings.EqualFold(h, clean) {
				dup = true
				break
			}
		}
		if !dup {
			hist = append(hist, clean)
		}
	}
	if len(hist) > 12 {
		hist = hist[len(hist)-12:]
	}
	p.RecentColors = hist
	m.cfg.Personalization = p
	m.sanitize()
	m.mu.Unlock()
	m.logf("Personalizacion actualizada: background=%s (%d dinamicos) font=%s/%s anim=%v blur=%v shadows=%v textshadow=%v/%v colores=%s",
		p.Background.Type, len(p.Background.DynamicImages), p.FontPrimary, p.FontSecondary,
		p.Animations, p.Blur, p.Shadows, p.TextShadow, p.TextShadowIntensity, p.Colors.Buttons)
	if err := m.Save(); err != nil {
		return err
	}
	m.cleanupBackgrounds()
	return nil
}

func (m *Manager) cleanupBackgrounds() {
	base := filepath.Dir(m.configPath)
	dirs := []string{
		filepath.Join(base, "cache", "backgrounds"),
		filepath.Join(base, "backgrounds"),
	}
	bg := m.cfg.Personalization.Background
	referenced := map[string]bool{}
	for _, rel := range []string{bg.ImagePath, bg.VideoPath} {
		if rel != "" {
			referenced[filepath.Base(rel)] = true
		}
	}
	for _, rel := range bg.DynamicImages {
		if rel != "" {
			referenced[filepath.Base(rel)] = true
		}
	}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || referenced[e.Name()] {
				continue
			}
			os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

func (m *Manager) Reset() error {
	m.mu.Lock()
	m.cfg = Default()
	m.mu.Unlock()
	m.logf("Configuracion restablecida a los valores por defecto")
	if err := m.Save(); err != nil {
		return err
	}
	m.cleanupBackgrounds()
	return nil
}
