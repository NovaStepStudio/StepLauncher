package Config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
}

type LauncherConfig struct {
	MaxRAMGB             int     `json:"maxRamGB"`
	MaxMbps              float64 `json:"maxMbps"`
	ConcurrentDownloads  int     `json:"concurrentDownloads"`
	HideLauncherOnLaunch bool    `json:"hideLauncherOnLaunch"`
	VerifyIntegrity *bool `json:"verifyIntegrity"`
	CheckForUpdatesOnStart bool `json:"checkForUpdatesOnStart"`
}

func (l LauncherConfig) VerifyEnabled() bool {
	if l.VerifyIntegrity == nil {
		return true
	}
	return *l.VerifyIntegrity
}

func boolPtr(b bool) *bool { return &b }

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

type Personalization struct {
	UIScale             int              `json:"uiScale"`
	Background          BackgroundConfig `json:"background"`
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

const (
	FileAssets       = "launcher_assets.json"
	FileAccounts     = "launcher_accounts.json"
	FileHistory      = "launcher_history.json"
	FileProfiles     = "launcher_profiles.json"
	FileCrashHistory = "launcher_history_crashes.json"
)

const (
	ExtraKeyAssets       = "assets"
	ExtraKeyAccounts     = "accounts"
	ExtraKeyHistory      = "history"
	ExtraKeyProfiles     = "profiles"
	ExtraKeyCrashHistory = "crashHistory"
)

type ExtraData struct {
	Assets       string `json:"assets"`
	Accounts     string `json:"accounts"`
	History      string `json:"history"`
	Profiles     string `json:"profiles"`
	CrashHistory string `json:"crashHistory"`
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
	ExtraData    ExtraData          `json:"extraData"`
}

func Default() Config {
	return Config{
		MinecraftConfig: MinecraftConfig{
			HardwareEnabled:      true,
			HardwareAcceleration: true,
			GPUType:              "auto",
			GPUPreset:            "",
			JavaMode:             "auto",
			AuthVerify:           true,
			WindowWidth:          854,
			WindowHeight:         480,
		},
		Launcher: LauncherConfig{
			MaxRAMGB:              2,
			MaxMbps:               0,
			ConcurrentDownloads:   4,
			HideLauncherOnLaunch:  true,
			VerifyIntegrity:       boolPtr(true),
			CheckForUpdatesOnStart: false,
		},		Idle: IdleConfig{
			AutoCloseModals:    true,
			IdleMinutes:        1,
			ConfigCheckEnabled: true,
			ConfigCheckMinutes: 3,
		},
		RichPresence: RichPresenceConfig{
			Enabled: boolPtr(true),
		},
		ExtraData: ExtraData{
			Assets:       FileAssets,
			Accounts:     FileAccounts,
			History:      FileHistory,
			Profiles:     FileProfiles,
			CrashHistory: FileCrashHistory,
		},
		Personalization: Personalization{
			UIScale: 100,
			Background: BackgroundConfig{
				Type:            "none",
				DynamicOrder:    "sequential",
				DynamicInterval: 10,
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
	if err := os.WriteFile(m.configPath, append(data, '\n'), 0644); err != nil {
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

func (m *Manager) SetCheckForUpdatesOnStart(v bool) error {
	m.mu.Lock()
	m.cfg.Launcher.CheckForUpdatesOnStart = v
	m.mu.Unlock()
	m.logf("CheckForUpdatesOnStart -> %v", v)
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
