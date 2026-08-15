package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	engineconfig "StepLauncher/internal/Handlers/Engine/engineconfig"
)

// DirectoryInfo describe el estado actual del directorio de trabajo para la UI.
type DirectoryInfo struct {
	Mode            engineconfig.DirMode `json:"mode"`
	CustomPath      string               `json:"customPath"`
	Configured      bool                 `json:"configured"`
	WorkDir         string               `json:"workDir"`
	NormalDir       string               `json:"normalDir"`
	MinecraftDir    string               `json:"minecraftDir"`
	MinecraftExists bool                 `json:"minecraftExists"`
	PortableDir     string               `json:"portableDir"`
}

func (e *Engine) DirectoryInfo() DirectoryInfo {
	b := e.config.Bootstrap()
	normal := engineconfig.DefaultNormalDir()
	portable := engineconfig.PortableDir()
	return DirectoryInfo{
		Mode:            b.Mode,
		CustomPath:      b.CustomPath,
		Configured:      b.Configured,
		WorkDir:         e.config.Get().WorkDir,
		NormalDir:       normal,
		MinecraftDir:    engineconfig.DefaultMinecraftDir(),
		MinecraftExists: engineconfig.DetectMinecraftDir() != "",
		PortableDir:     portable,
	}
}

// SetDirectory cambia el modo/ubicacion del directorio de trabajo. Solo se
// copia launcher_config.json (si el destino no tiene uno); el resto de datos
// no se migran. Requiere reiniciar el launcher para aplicarse.
func (e *Engine) SetDirectory(mode engineconfig.DirMode, customPath string) error {
	if mode != engineconfig.ModeNormal && mode != engineconfig.ModeMinecraft &&
		mode != engineconfig.ModePortable && mode != engineconfig.ModeCustom {
		return fmt.Errorf("modo de directorio invalido: %s", mode)
	}
	if mode == engineconfig.ModeCustom && strings.TrimSpace(customPath) == "" {
		return fmt.Errorf("la ruta personalizada no puede estar vacia")
	}
	if e.hasActiveSessions() {
		return fmt.Errorf("no se puede cambiar el directorio con juegos o descargas activos")
	}

	b := e.config.Bootstrap()
	oldWorkDir := b.ResolveWorkDir()
	b.Mode = mode
	b.CustomPath = customPath
	b.Configured = true
	newWorkDir := b.ResolveWorkDir()

	if newWorkDir != oldWorkDir {
		if err := copyLauncherConfigIfMissing(oldWorkDir, newWorkDir); err != nil {
			return err
		}
	}
	if err := e.config.SetBootstrap(b); err != nil {
		return err
	}

	e.applyDirectoryMode()
	if e.log != nil {
		e.log.System("Directorio de trabajo cambiado: %s -> %s (modo %s)", oldWorkDir, newWorkDir, mode)
	}
	return nil
}

func (e *Engine) hasActiveSessions() bool {
	for _, g := range e.ListGames() {
		if g.Status == GameRunning || g.Status == GameStarting {
			return true
		}
	}
	for _, d := range e.ListDownloads() {
		switch d.State {
		case "pending", "downloading", "verifying", "redownloading":
			return true
		}
	}
	return false
}

func (e *Engine) applyDirectoryMode() {
	if e.config.Bootstrap().Mode == engineconfig.ModeMinecraft {
		e.applySeparateGameDir(false)
		return
	}
	e.applySeparateGameDir(e.config.Get().SeparateGameDirValue())
}

// copyLauncherConfigIfMissing copia launcher_config.json de un WorkDir a otro
// SOLO si el destino no tiene uno (nunca sobrescribe archivos existentes).
func copyLauncherConfigIfMissing(srcDir, dstDir string) error {
	src := filepath.Join(srcDir, "launcher_config.json")
	dst := filepath.Join(dstDir, "launcher_config.json")
	if _, err := os.Stat(dst); err == nil {
		return nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return nil
	}
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}