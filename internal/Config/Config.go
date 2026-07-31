// Package Config gestiona el archivo config.json del launcher.
// El JSON se organiza en bloques claros: minecraftConfig, launcher y personalization.
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
	HardwareEnabled       bool   `json:"hardwareEnabled"`
	HardwareAcceleration  bool   `json:"hardwareAcceleration"`
	GPUType               string `json:"gpuType"`
	GPUPreset             string `json:"gpuPreset"`

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

	JavaArgs     string `json:"javaArgs"`
	GameArgs     string `json:"gameArgs"`

	OfflineMode  bool `json:"offlineMode"`
	CompatMode   bool `json:"compatMode"`
	DetailedLogs bool `json:"detailedLogs"`
}

type LauncherConfig struct {
	MaxRAMGB            int     `json:"maxRamGB"`
	MaxMbps             float64 `json:"maxMbps"`
	ConcurrentDownloads int     `json:"concurrentDownloads"`
}

type BackgroundConfig struct {
	Type            string   `json:"type"`          // none | image | video | dynamic
	ImagePath       string   `json:"imagePath"`
	VideoPath       string   `json:"videoPath"`
	DynamicImages   []string `json:"dynamicImages"`
	DynamicOrder    string   `json:"dynamicOrder"`    // sequential | random
	DynamicInterval int      `json:"dynamicInterval"` // segundos
}

type ThemeColors struct {
	Sidebar     string `json:"sidebar"`
	Modal       string `json:"modal"`
	Buttons     string `json:"buttons"`
	BorderModal string `json:"borderModal"`
	Border      string `json:"border"`
}

type Personalization struct {
	UIScale       int               `json:"uiScale"`
	Background    BackgroundConfig  `json:"background"`
	FontPrimary   string            `json:"fontPrimary"`
	FontSecondary string            `json:"fontSecondary"`
	Colors        ThemeColors       `json:"colors"`
	RecentColors  []string          `json:"recentColors"`
	Animations    bool              `json:"animations"`
	Blur          bool              `json:"blur"`
	Shadows       bool              `json:"shadows"`
}

type Config struct {
	MinecraftConfig MinecraftConfig `json:"minecraftConfig"`
	Launcher        LauncherConfig  `json:"launcher"`
	Personalization Personalization `json:"personalization"`
}

func Default() Config {
	return Config{
		MinecraftConfig: MinecraftConfig{
			HardwareEnabled:      true,
			HardwareAcceleration: true,
			GPUType:              "auto",
			GPUPreset:            "",
			JavaMode:  "auto",
			AuthVerify: true,
			WindowWidth:          854,
			WindowHeight:         480,
		},
		Launcher: LauncherConfig{
			MaxRAMGB:            2,
			MaxMbps:             0,
			ConcurrentDownloads: 4,
		},
		Personalization: Personalization{
			UIScale: 100,
			Background: BackgroundConfig{
				Type:            "none",
				DynamicOrder:    "sequential",
				DynamicInterval: 10,
			},
			FontPrimary:   "Lexend",
			FontSecondary: "Inter",
			Colors: ThemeColors{
				Sidebar:     "#0005",
				Modal:       "#111",
				Buttons:     "#111",
				BorderModal: "#494949",
				Border:      "rgba(37, 37, 37, 0.3)",
			},
			Animations: true,
			Blur:       true,
			Shadows:    true,
		},
	}
}

type Manager struct {
	mu         sync.RWMutex
	cfg        Config
	configPath string
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
		return
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return
	}
	if _, ok := raw["minecraftConfig"]; ok {
		var cfg Config
		if json.Unmarshal(data, &cfg) != nil {
			return
		}
		m.cfg = cfg
		m.sanitize()
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
}

func (m *Manager) sanitize() {
	c := &m.cfg
	if c.Launcher.MaxRAMGB < 1 {
		c.Launcher.MaxRAMGB = 2
	}
	if c.Launcher.ConcurrentDownloads < 1 {
		c.Launcher.ConcurrentDownloads = 4
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
	if bg.Type != "dynamic" {
		bg.DynamicImages = nil
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
	sanitizeColor(&c.Personalization.Colors.Sidebar, "#0005")
	sanitizeColor(&c.Personalization.Colors.Modal, "#111")
	sanitizeColor(&c.Personalization.Colors.Buttons, "#111")
	sanitizeColor(&c.Personalization.Colors.BorderModal, "#494949")
	sanitizeColor(&c.Personalization.Colors.Border, "rgba(37, 37, 37, 0.3)")
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
}

func sanitizeColor(v *string, fallback string) {
	clean := sanitizeColorString(*v)
	if clean == "" {
		*v = fallback
		return
	}
	*v = clean
}

// sanitizeColorString acepta #rrggbb, #rrggbbaa, rgb(r,g,b) o rgba(r,g,b,a).
// Devuelve el valor normalizado o "" si es invalido.
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
	defer m.mu.Unlock()
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
	return os.WriteFile(m.configPath, append(data, '\n'), 0644)
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
	return m.Save()
}

func (m *Manager) SetAuthVerify(verify bool) error {
	m.mu.Lock()
	m.cfg.MinecraftConfig.AuthVerify = verify
	m.mu.Unlock()
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
	return m.Save()
}

func (m *Manager) SetMaxRAMGB(gb int) error {
	m.mu.Lock()
	if gb < 1 {
		gb = 2
	}
	m.cfg.Launcher.MaxRAMGB = gb
	m.mu.Unlock()
	return m.Save()
}

func (m *Manager) SetMaxMbps(mbps float64) error {
	m.mu.Lock()
	m.cfg.Launcher.MaxMbps = mbps
	m.mu.Unlock()
	return m.Save()
}

func (m *Manager) SetConcurrentDownloads(n int) error {
	m.mu.Lock()
	if n < 1 {
		n = 1
	}
	m.cfg.Launcher.ConcurrentDownloads = n
	m.mu.Unlock()
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
	return m.Save()
}

// UpdatePersonalization guarda las preferencias de personalizacion
// preservando el UIScale y registrando los colores en el historial.
func (m *Manager) UpdatePersonalization(p Personalization) error {
	m.mu.Lock()
	p.UIScale = m.cfg.Personalization.UIScale
	hist := m.cfg.Personalization.RecentColors
	for _, col := range []string{
		p.Colors.Sidebar, p.Colors.Modal, p.Colors.Buttons, p.Colors.BorderModal, p.Colors.Border,
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
	if err := m.Save(); err != nil {
		return err
	}
	m.cleanupBackgrounds()
	return nil
}

// cleanupBackgrounds borra de los directorios de fondos los archivos que ya
// no estan referenciados por la configuracion actual (fondos reemplazados).
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

// Reset restaura la configuracion completa a los valores por defecto.
func (m *Manager) Reset() error {
	m.mu.Lock()
	m.cfg = Default()
	m.mu.Unlock()
	if err := m.Save(); err != nil {
		return err
	}
	m.cleanupBackgrounds()
	return nil
}
