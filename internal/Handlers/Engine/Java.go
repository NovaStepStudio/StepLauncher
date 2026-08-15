package engine

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	engineconfig "StepLauncher/internal/Handlers/Engine/engineconfig"
)

func (e *Engine) DetectJavaInstallations() []string {
	seen := map[string]bool{}
	var results []string

	add := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		real, err := filepath.EvalSymlinks(path)
		if err != nil {
			real = path
		}
		key := strings.ToLower(real)
		if seen[key] {
			return
		}
		if filepath.Base(filepath.Dir(real)) == "javapath" || filepath.Base(filepath.Dir(real)) != "bin" {
			return
		}
		v, err := javaVersion(real)
		if err != nil {
			return
		}
		seen[key] = true
		results = append(results, real+" ("+v+")")
	}

	if jh := os.Getenv("JAVA_HOME"); jh != "" {
		add(filepath.Join(jh, "bin", javaExe()))
	}

	pathDirs := filepath.SplitList(os.Getenv("PATH"))
	for _, dir := range pathDirs {
		add(filepath.Join(dir, javaExe()))
	}

	if runtime.GOOS == "windows" {
		for _, p := range scanWindowsJava() {
			add(p)
		}
	} else {
		for _, p := range scanUnixJava() {
			add(p)
		}
	}

	return results
}

func javaExe() string {
	if runtime.GOOS == "windows" {
		return "javaw.exe"
	}
	return "java"
}

func javaVersion(path string) (string, error) {
	out, err := exec.Command(path, "-version").CombinedOutput()
	if err != nil {
		return "", err
	}
	line := strings.SplitN(string(out), "\n", 2)[0]
	line = strings.TrimSpace(line)
	if idx := strings.LastIndex(line, "\""); idx >= 0 {
		start := strings.LastIndex(line[:idx], "\"")
		if start >= 0 {
			return line[start+1 : idx], nil
		}
	}
	return line, nil
}

func scanWindowsJava() []string {
	var found []string
	dirs := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "Java"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Java"),
		filepath.Join(os.Getenv("ProgramFiles"), "Eclipse Adoptium"),
		filepath.Join(os.Getenv("ProgramFiles"), "Microsoft"),
		filepath.Join(os.Getenv("LocalAppData"), "Programs", "Eclipse Adoptium"),
	}
	for _, base := range dirs {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			bin := filepath.Join(base, entry.Name(), "bin", "javaw.exe")
			if _, err := os.Stat(bin); err == nil {
				found = append(found, bin)
			}
		}
	}
	return found
}

func scanUnixJava() []string {
	var found []string
	dirs := []string{"/usr/lib/jvm", "/usr/java", "/Library/Java/JavaVirtualMachines"}
	for _, base := range dirs {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			var bin string
			if runtime.GOOS == "darwin" {
				bin = filepath.Join(base, entry.Name(), "Contents", "Home", "bin", "java")
			} else {
				bin = filepath.Join(base, entry.Name(), "bin", "java")
			}
			if _, err := os.Stat(bin); err == nil {
				found = append(found, bin)
			}
		}
	}
	return found
}

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

	JavaArgs            string `json:"javaArgs"`
	GameArgs            string `json:"gameArgs"`
	OfflineMode         bool   `json:"offlineMode"`
	CompatMode          bool   `json:"compatMode"`
	DetailedLogs        bool   `json:"detailedLogs"`
	ConcurrentDownloads int    `json:"concurrentDownloads"`
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

func minecraftFromConfig(cfg engineconfig.Config) MinecraftConfig {
	return MinecraftConfig{
		HardwareEnabled:      cfg.HardwareEnabled,
		HardwareAcceleration: cfg.HardwareAcceleration,
		GPUType:              cfg.GPUType,
		GPUPreset:            cfg.GPUPreset,
		JavaMode:             cfg.JavaMode,
		JavaCustomPath:       cfg.JavaCustomPath,
		ProxyEnabled:         cfg.ProxyEnabled,
		ProxyHost:            cfg.ProxyHost,
		ProxyPort:            cfg.ProxyPort,
		ProxyUser:            cfg.ProxyUser,
		ProxyPass:            cfg.ProxyPass,
		AuthVerify:           cfg.AuthVerify,
		WindowWidth:          cfg.WindowWidth,
		WindowHeight:         cfg.WindowHeight,
		Fullscreen:           cfg.Fullscreen,
		JavaArgs:             cfg.JavaArgs,
		GameArgs:             cfg.GameArgs,
OfflineMode:         cfg.OfflineMode,
		CompatMode:          cfg.CompatMode,
		DetailedLogs:        cfg.DetailedLogs,
		ConcurrentDownloads: cfg.ConcurrentDownloads,
		SeparateGameDir:     cfg.SeparateGameDir,
	}
}

func configFromMinecraft(mc MinecraftConfig, base engineconfig.Config) engineconfig.Config {
	base.HardwareEnabled = mc.HardwareEnabled
	base.HardwareAcceleration = mc.HardwareAcceleration
	base.GPUType = mc.GPUType
	base.GPUPreset = mc.GPUPreset
	base.JavaMode = mc.JavaMode
	base.JavaCustomPath = mc.JavaCustomPath
	base.ProxyEnabled = mc.ProxyEnabled
	base.ProxyHost = mc.ProxyHost
	base.ProxyPort = mc.ProxyPort
	base.ProxyUser = mc.ProxyUser
	base.ProxyPass = mc.ProxyPass
	base.AuthVerify = mc.AuthVerify
	base.WindowWidth = mc.WindowWidth
	base.WindowHeight = mc.WindowHeight
	base.Fullscreen = mc.Fullscreen
	base.JavaArgs = mc.JavaArgs
	base.GameArgs = mc.GameArgs
	base.OfflineMode = mc.OfflineMode
	base.CompatMode = mc.CompatMode
	base.DetailedLogs = mc.DetailedLogs
	if mc.ConcurrentDownloads > 0 {
		base.ConcurrentDownloads = mc.ConcurrentDownloads
	}
	base.SeparateGameDir = mc.SeparateGameDir
	return base
}

func (e *Engine) GetMinecraftConfig() MinecraftConfig {
	return minecraftFromConfig(e.config.Get())
}

func (e *Engine) UpdateMinecraftConfig(mc MinecraftConfig) {
	// En modo Minecraft el gameDir SIEMPRE es el propio .minecraft.
	if e.config.Bootstrap().Mode == engineconfig.ModeMinecraft {
		falseVal := false
		mc.SeparateGameDir = &falseVal
	}
	cfg := configFromMinecraft(mc, e.config.Get())
	e.config.UpdateConfig(cfg)
	e.applySeparateGameDir(mc.SeparateGameDirValue())
}

func (e *Engine) applySeparateGameDir(separate bool) {
	if e.launcher != nil {
		e.launcher.SetSeparateGameDir(separate)
	}
	if e.instances != nil {
		e.instances.SetSeparateGameDir(separate)
	}
}
