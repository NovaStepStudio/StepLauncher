package mods

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Subdirectorios de contenido dentro del gameDir. Son los que crea Minecraft
// con modloaders: mods para los .jar, resourcepacks para texturas y
// shaderpacks para shaders.
const (
	SubdirMods           = "mods"
	SubdirResourcePacks  = "resourcepacks"
	SubdirShaderPacks    = "shaderpacks"
	SubdirModpacksStaged = "modpacks"
)

// ContentSubdir traduce el project_type de Modrinth a su subcarpeta del
// gameDir. Los modpacks no se instalan como archivo suelto (se expanden), pero
// se les asigna su carpeta de staging por si hiciera falta.
func ContentSubdir(projectType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(projectType)) {
	case "mod":
		return SubdirMods, nil
	case "resourcepack":
		return SubdirResourcePacks, nil
	case "shader":
		return SubdirShaderPacks, nil
	case "modpack":
		return SubdirModpacksStaged, nil
	default:
		return "", fmt.Errorf("tipo de contenido no soportado: %s", projectType)
	}
}

// ResolveGameDir calcula el gameDir global: <workDir>/game si la opción de
// carpeta "game" separada está activa, o el propio workDir si está
// deshabilitada (los mods/shaders/texturas se buscan entonces en las carpetas
// del directorio base).
func ResolveGameDir(workDir string, separate bool) string {
	if separate {
		return filepath.Join(workDir, "game")
	}
	return workDir
}

// SanitizeFileName valida que el nombre sea un archivo plano (sin rutas ni
// escapes) y devuelve su base. Rechaza separadores y ".." en vez de
// recortarlos: un filename remoto nunca debe traer ruta.
func SanitizeFileName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || trimmed == "." {
		return "", fmt.Errorf("nombre de archivo inválido")
	}
	if strings.ContainsAny(trimmed, "/\\") || strings.Contains(trimmed, "..") {
		return "", fmt.Errorf("nombre de archivo inválido: %s", name)
	}
	base := filepath.Base(trimmed)
	if base == "" || base == "." || base == string(filepath.Separator) {
		return "", fmt.Errorf("nombre de archivo inválido")
	}
	return base, nil
}

// DownloadToFile descarga una URL a dest con escritura atómica (temporal +
// rename) y verificación opcional de hash (sha1 o sha512 en hex). Si el
// destino ya existe con el mismo tamaño y hash válido, se omite la descarga.
func DownloadToFile(ctx context.Context, client *http.Client, url, dest, sha1, sha512 string, size int64) error {
	if client == nil {
		client = &http.Client{Timeout: 90 * time.Second}
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("crear directorio: %w", err)
	}
	if info, err := os.Stat(dest); err == nil && !info.IsDir() {
		if size > 0 && info.Size() == size {
			if sha1 == "" && sha512 == "" {
				return nil
			}
			if ok, _ := VerifyFile(dest, sha1, sha512); ok {
				return nil
			}
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("crear petición: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("descargar %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("descargar %s: HTTP %d", url, resp.StatusCode)
	}
	tmp := dest + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("crear temporal: %w", err)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(tmp)
		return fmt.Errorf("escribir %s: %w", dest, err)
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("cerrar %s: %w", dest, err)
	}
	if sha1 != "" || sha512 != "" {
		if ok, err := VerifyFile(tmp, sha1, sha512); err != nil || !ok {
			os.Remove(tmp)
			if err != nil {
				return fmt.Errorf("verificar %s: %w", dest, err)
			}
			return fmt.Errorf("verificar %s: el hash no coincide", dest)
		}
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("mover %s: %w", dest, err)
	}
	return nil
}
