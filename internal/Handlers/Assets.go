package Handlers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	assets "StepLauncher/internal/Core/Assets"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetLauncherAssets() assets.Assets {
	if a.assets == nil {
		return assets.Default()
	}
	val, err := a.assets.Load()
	if err != nil {
		a.logf("[Assets] WARN: no se pudo leer %s: %v", assets.AssetsFileName, err)
		return assets.Default()
	}
	return val
}

func (a *App) SaveLauncherAssets(asset assets.Assets) {
	if a.assets == nil {
		return
	}
	if err := a.assets.Save(asset); err != nil {
		a.logf("[Assets] WARN: no se pudo guardar %s: %v", assets.AssetsFileName, err)
		return
	}
	var primary, secundary string
	for _, s := range asset.Fonts {
		switch s.Type {
		case "primary":
			primary = s.Name + "/" + s.Path
		case "secundary":
			secundary = s.Name + "/" + s.Path
		}
	}
	a.logf("[Assets] Assets guardados: primary=%s secundary=%s (%d entradas)", primary, secundary, len(asset.Fonts))
}

func (a *App) ListFontFiles() []string {
	if a.assets == nil {
		return []string{}
	}
	files, err := a.assets.ListFontFiles()
	if err != nil {
		a.logf("[Assets] WARN: no se pudo listar launcher/fonts: %v", err)
		return []string{}
	}
	return files
}

func (a *App) DeleteFontFile(name string) error {
	if a.assets == nil {
		return fmt.Errorf("assets no disponible")
	}
	ref := assets.LauncherDir + "/" + assets.FontsSubDir + "/" + name
	var wasPrimary, wasSecundary string
	if cur, err := a.assets.Load(); err == nil {
		for _, s := range cur.Fonts {
			if s.Path != ref {
				continue
			}
			switch s.Type {
			case "primary":
				wasPrimary = s.Name
			case "secundary":
				wasSecundary = s.Name
			}
		}
	}
	if err := a.assets.DeleteFontFile(name); err != nil {
		return err
	}
	a.logf("[Assets] Fuente eliminada: %s", name)
	if a.config != nil && (wasPrimary != "" || wasSecundary != "") {
		a.config.ResetFontIfMatches(wasPrimary, wasSecundary)
	}
	return nil
}

func (a *App) PickFontFile() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("contexto no disponible")
	}
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Seleccionar tipografía",
		Filters: []runtime.FileFilter{
			{DisplayName: "Tipografías (*.ttf, *.otf, *.woff, *.woff2)", Pattern: "*.ttf;*.otf;*.woff;*.woff2"},
		},
	})
}

func (a *App) ImportFont(src string) (string, error) {
	if a.assets == nil {
		return "", fmt.Errorf("assets no disponible")
	}
	ext := strings.ToLower(filepath.Ext(src))
	if !assets.IsFontExt(ext) {
		return "", fmt.Errorf("formato de tipografia no soportado: %s (usa .ttf, .otf, .woff o .woff2)", ext)
	}
	info, err := os.Stat(src)
	if err != nil {
		return "", fmt.Errorf("no se pudo leer el archivo: %v", err)
	}
	if info.Size() > 15*1024*1024 {
		return "", fmt.Errorf("la tipografia no debe pesar mas de 15MB")
	}

	destDir := a.assets.FontsDir()
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	dest := filepath.Join(destDir, filepath.Base(src))

	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return "", err
	}

	a.logf("[Assets] Tipografia importada: %s -> %s", src, dest)
	return "launcher/fonts/" + filepath.Base(dest), nil
}
