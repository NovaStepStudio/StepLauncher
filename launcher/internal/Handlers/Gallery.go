package Handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DownloadGalleryImageAsBackground descarga una imagen de la galería de Modrinth
// (preferentemente en alta resolución .png), la guarda en cache/backgrounds/,
// la registra en launcher_assets.json con su autor y nombre del mod, y la
// establece como fondo del launcher (personalization.background.type = "image").
//
// Devuelve la ruta relativa (cache/backgrounds/...png) para que el frontend
// pueda confirmar y refrescar. También guarda attribution para el label.
func (a *App) DownloadGalleryImageAsBackground(urlStr, author, modName, title string) (string, error) {
	if a.engine == nil || a.config == nil || a.assets == nil {
		return "", fmt.Errorf("engine no disponible")
	}
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		return "", fmt.Errorf("URL vacía")
	}
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return "", fmt.Errorf("URL no válida")
	}

	// Deduplicación por URL: si la misma URL ya está registrada y el archivo aún existe,
	// reutilizarla sin descargar de nuevo (evita los duplicados reportados con mismo CDN URL
	// pero distinto timestamp como 35b1b4eb... y 52b5e9... repetidos).
	urlNorm := strings.TrimSpace(urlStr)
	if a.assets != nil && urlNorm != "" {
		if curAssets, err := a.assets.Load(); err == nil {
			for _, g := range curAssets.Gallery {
				if strings.TrimSpace(g.Url) == urlNorm {
					full := filepath.Join(a.engine.ConfigManager().RootDir(), filepath.FromSlash(g.Path))
					if _, err := os.Stat(full); err == nil {
						_ = a.assets.RegisterGalleryBackground(g.Path, author, modName, urlStr, title)
						cfg := a.config.Get()
						cfg.Personalization.Background.Type = "image"
						cfg.Personalization.Background.ImagePath = g.Path
						cfg.Personalization.Background.ImageAuthor = strings.TrimSpace(author)
						cfg.Personalization.Background.ImageModName = strings.TrimSpace(modName)
						cfg.Personalization.Background.ImageUrl = urlNorm
						cfg.Personalization.Background.VideoPath = ""
						if err := a.updatePersonalizationInternal(cfg.Personalization); err != nil {
							return "", fmt.Errorf("no se pudo guardar personalización: %v", err)
						}
						a.logf("[Gallery] Fondo reutilizado (deduplicado por URL): %s -> %s (autor=%s mod=%s)", urlStr, g.Path, author, modName)
						return g.Path, nil
					}
					break // archivo no existe -> se descargará de nuevo y Register reemplazará la entrada vieja
				}
			}
		}
	}

	// Determinar extensión
	ext := strings.ToLower(filepath.Ext(strings.Split(urlStr, "?")[0]))
	if ext == "" {
		ext = ".png"
	}
	if !imageExts[ext] {
		ext = ".png"
	}

	baseName := gallerySanitize(strings.TrimSpace(modName))
	if baseName == "" {
		baseName = "gallery"
	}
	if strings.TrimSpace(author) != "" {
		baseName = baseName + "_" + gallerySanitize(author)
	}
	// Limitar longitud
	if len(baseName) > 60 {
		baseName = baseName[:60]
	}

	destDir := filepath.Join(a.engine.ConfigManager().RootDir(), "cache", "backgrounds")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("no se pudo crear directorio de fondos: %v", err)
	}

	fileName := fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixNano(), ext)
	destPath := filepath.Join(destDir, fileName)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("URL inválida: %v", err)
	}
	req.Header.Set("User-Agent", "StepLauncher/2.5.0")
	req.Header.Set("Accept", "image/*")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("no se pudo descargar la imagen: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Modrinth respondió %d", resp.StatusCode)
	}

	// Si el servidor devuelve content-type con extensión distinta, ajustar?
	// Mantener la extensión original para respetar el formato descargado.

	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("no se pudo crear archivo local: %v", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		_ = os.Remove(destPath)
		return "", fmt.Errorf("error al guardar imagen: %v", err)
	}
	// Verificar que es imagen válida (opcional, ya se valida resolución si es imagen)
	// No bloqueamos si falla, solo log.

	rel := filepath.ToSlash(filepath.Join("cache", "backgrounds", fileName))

	// Registrar en launcher_assets.json (deduplica por URL dentro del método)
	_ = a.assets.RegisterGalleryBackground(rel, author, modName, urlStr, title)

	// Actualizar personalización como fondo de imagen
	cfg := a.config.Get()
	cfg.Personalization.Background.Type = "image"
	cfg.Personalization.Background.ImagePath = rel
	cfg.Personalization.Background.ImageAuthor = strings.TrimSpace(author)
	cfg.Personalization.Background.ImageModName = strings.TrimSpace(modName)
	cfg.Personalization.Background.ImageUrl = strings.TrimSpace(urlStr)
	cfg.Personalization.Background.VideoPath = ""
	// Mantener dynamicImages vacío? No tocar otros campos.

	if err := a.updatePersonalizationInternal(cfg.Personalization); err != nil {
		// Rollback: si no se pudo guardar la personalización, eliminar el archivo y la entrada huérfana
		_ = os.Remove(destPath)
		_ = a.assets.RemoveGalleryBackground(rel)
		return "", fmt.Errorf("no se pudo guardar personalización: %v", err)
	}

	a.logf("[Gallery] Fondo colocado desde galería: %s -> %s (autor=%s mod=%s)", urlStr, rel, author, modName)
	return rel, nil
}

func gallerySanitize(name string) string {
	name = strings.TrimSpace(name)
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '-', r == '_', r == ' ', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "gallery"
	}
	return out
}
