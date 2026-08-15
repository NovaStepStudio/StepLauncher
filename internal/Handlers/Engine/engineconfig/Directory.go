package engineconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// DirMode define los modos de directorio de trabajo del launcher.
type DirMode string

const (
	ModeNormal    DirMode = "normal"
	ModeMinecraft DirMode = "minecraft"
	ModePortable  DirMode = "portable"
	ModeCustom    DirMode = "custom"
)

// Bootstrap persiste la preferencia de directorio FUERA del WorkDir para que
// pueda leerse antes de resolver la ruta de trabajo (y sobreviva a borrados o
// cambios de carpeta). Solo se copia launcher_config.json al cambiar.
type Bootstrap struct {
	Mode       DirMode `json:"mode"`
	CustomPath string  `json:"customPath,omitempty"`
	Configured bool    `json:"configured"`
}

// bootstrapDir devuelve la carpeta fija (fuera del WorkDir) donde se guarda la
// preferencia de directorio, segun el SO.
func bootstrapDir() string {
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "StepLauncher")
		}
	case "darwin":
		if home, _ := os.UserHomeDir(); home != "" {
			return filepath.Join(home, "Library", "Application Support", "StepLauncher")
		}
	default:
		if home, _ := os.UserHomeDir(); home != "" {
			return filepath.Join(home, ".config", "StepLauncher")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "StepLauncher")
}

// BootstrapPath es la ruta completa del archivo de preferencia de directorio.
func BootstrapPath() string {
	return filepath.Join(bootstrapDir(), "directory.json")
}

// LoadBootstrap lee la preferencia de directorio (defaults si no existe).
func LoadBootstrap() Bootstrap {
	b := Bootstrap{Mode: ModeNormal}
	data, err := os.ReadFile(BootstrapPath())
	if err != nil {
		return b
	}
	if json.Unmarshal(data, &b) != nil {
		return Bootstrap{Mode: ModeNormal}
	}
	if b.Mode != ModeNormal && b.Mode != ModeMinecraft && b.Mode != ModePortable && b.Mode != ModeCustom {
		b.Mode = ModeNormal
	}
	return b
}

// SaveBootstrap persiste la preferencia de directorio.
func SaveBootstrap(b Bootstrap) error {
	if err := os.MkdirAll(bootstrapDir(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(BootstrapPath(), data, 0644)
}

// defaultStepLauncherDir es el WorkDir del modo Normal.
func defaultStepLauncherDir() string {
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, ".StepLauncher")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".StepLauncher")
}

// DefaultMinecraftDir devuelve la carpeta del launcher oficial de Minecraft
// segun el SO (la misma que usa la instalacion oficial).
func DefaultMinecraftDir() string {
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, ".minecraft")
		}
	case "darwin":
		if home, _ := os.UserHomeDir(); home != "" {
			return filepath.Join(home, "Library", "Application Support", "minecraft")
		}
	default:
		if home, _ := os.UserHomeDir(); home != "" {
			return filepath.Join(home, ".minecraft")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".minecraft")
}

// DetectMinecraftDir devuelve la ruta de .minecraft si existe en este SO.
func DetectMinecraftDir() string {
	dir := DefaultMinecraftDir()
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}
	return ""
}

// DefaultNormalDir es el WorkDir del modo Normal.
func DefaultNormalDir() string {
	return defaultStepLauncherDir()
}

// PortableDir es el WorkDir del modo Portable (junto al ejecutable).
func PortableDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), ".StepLauncher")
	}
	return defaultStepLauncherDir()
}

// ResolveWorkDir calcula el WorkDir segun el modo del bootstrap.
func (b Bootstrap) ResolveWorkDir() string {
	switch b.Mode {
	case ModeMinecraft:
		return DefaultMinecraftDir()
	case ModePortable:
		return PortableDir()
	case ModeCustom:
		if b.CustomPath != "" {
			return b.CustomPath
		}
	}
	return defaultStepLauncherDir()
}