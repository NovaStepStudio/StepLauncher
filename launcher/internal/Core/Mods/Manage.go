package mods

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// EnsureContentDirs crea el gameDir y sus carpetas de contenido (mods,
// resourcepacks, shaderpacks) para que la instalación nunca falle por falta
// de directorios, venga de un mod suelto o de un modpack.
func EnsureContentDirs(gameDir string) error {
	for _, dir := range []string{gameDir, filepath.Join(gameDir, SubdirMods), filepath.Join(gameDir, SubdirResourcePacks), filepath.Join(gameDir, SubdirShaderPacks)} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("crear %s: %w", dir, err)
		}
	}
	return nil
}

// iconExtByType traduce el Content-Type detectado a extensión de imagen.
func iconExtByType(contentType string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])) {
	case "image/png":
		return ".png", true
	case "image/jpeg":
		return ".jpg", true
	case "image/webp":
		return ".webp", true
	case "image/gif":
		return ".gif", true
	case "image/bmp":
		return ".bmp", true
	default:
		return "", false
	}
}

// DownloadIcon descarga el icono de un proyecto de Modrinth a dest (sin
// extensión) y devuelve la ruta con la extensión detectada del contenido.
// Las URLs de Modrinth no siempre traen extensión, así que se detecta del
// propio contenido descargado.
func DownloadIcon(client *http.Client, url, dest string) (string, error) {
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return "", fmt.Errorf("URL de icono inválida")
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("descargar icono: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("descargar icono: HTTP %d", resp.StatusCode)
	}
	tmp := dest + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return "", fmt.Errorf("crear temporal: %w", err)
	}
	// Se leen como máximo 2 MB: un icono no necesita más.
	head := make([]byte, 512)
	n, _ := resp.Body.Read(head)
	contentType := http.DetectContentType(head[:n])
	if _, err := out.Write(head[:n]); err != nil {
		out.Close()
		os.Remove(tmp)
		return "", fmt.Errorf("escribir icono: %w", err)
	}
	const maxIconBytes = 2 * 1024 * 1024
	var copied int64
	buf := make([]byte, 32*1024)
	for copied < maxIconBytes {
		nr, er := resp.Body.Read(buf)
		if nr > 0 {
			if _, ew := out.Write(buf[:nr]); ew != nil {
				out.Close()
				os.Remove(tmp)
				return "", fmt.Errorf("escribir icono: %w", ew)
			}
			copied += int64(nr)
		}
		if er != nil {
			break
		}
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return "", fmt.Errorf("cerrar icono: %w", err)
	}
	ext, ok := iconExtByType(contentType)
	if !ok {
		os.Remove(tmp)
		return "", fmt.Errorf("el icono no es una imagen válida (%s)", contentType)
	}
	final := dest + ext
	os.Remove(final)
	if err := os.Rename(tmp, final); err != nil {
		os.Remove(tmp)
		return "", fmt.Errorf("mover icono: %w", err)
	}
	return final, nil
}

// DisabledSuffix marca un archivo como deshabilitado sin borrarlo: Minecraft
// ignora todo lo que no termine en .jar/.zip, así que renombrar basta para
// activar/desactivar mods, shaders y texturas.
const DisabledSuffix = ".disabled"

// InstalledFile describe un archivo de contenido instalado en un destino.
// Title es el nombre bonito declarado dentro del mod (vacío = usar Name).
type InstalledFile struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Size    int64  `json:"size"`
	Enabled bool   `json:"enabled"`
}

// contentKinds son los tipos gestionables con su subcarpeta del gameDir.
var contentKinds = []struct {
	kind   string
	subdir string
}{
	{"mod", SubdirMods},
	{"shader", SubdirShaderPacks},
	{"resourcepack", SubdirResourcePacks},
}

// ListInstalled enumera el contenido instalado en un gameDir, por tipo. Los
// temporales (.tmp) se ignoran; los terminados en .disabled cuentan como
// deshabilitados (se muestra su nombre real).
func ListInstalled(gameDir string) ([]InstalledFile, error) {
	var out []InstalledFile
	for _, ck := range contentKinds {
		dir := filepath.Join(gameDir, ck.subdir)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || strings.HasSuffix(e.Name(), ".tmp") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			name := e.Name()
			enabled := true
			if strings.HasSuffix(name, DisabledSuffix) {
				enabled = false
				name = strings.TrimSuffix(name, DisabledSuffix)
			}
			out = append(out, InstalledFile{Kind: ck.kind, Name: name, Title: ContentDisplayName(gameDir, ck.kind, name), Size: info.Size(), Enabled: enabled})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

// resolveInstalled valida tipo y nombre, y devuelve las rutas actual y
// objetivo para activar/desactivar o borrar.
func resolveInstalled(gameDir, kind, name string) (subdir, current, target string, enabled bool, err error) {
	subdir, err = ContentSubdir(kind)
	if err != nil {
		return "", "", "", false, err
	}
	base, err := SanitizeFileName(name)
	if err != nil {
		return "", "", "", false, err
	}
	plain := filepath.Join(gameDir, subdir, base)
	disabled := plain + DisabledSuffix
	if info, statErr := os.Stat(disabled); statErr == nil && !info.IsDir() {
		return subdir, disabled, plain, false, nil
	}
	if info, statErr := os.Stat(plain); statErr == nil && !info.IsDir() {
		return subdir, plain, disabled, true, nil
	}
	return "", "", "", false, fmt.Errorf("no se encontró %s en %s", base, subdir)
}

// SetInstalledEnabled activa o desactiva un archivo instalado renombrándolo
// con o sin el sufijo .disabled.
func SetInstalledEnabled(gameDir, kind, name string, enabled bool) error {
	_, current, target, isEnabled, err := resolveInstalled(gameDir, kind, name)
	if err != nil {
		return err
	}
	if isEnabled == enabled {
		return nil
	}
	// current es el archivo tal como está; target, como debe quedar.
	if err := os.Rename(current, target); err != nil {
		return fmt.Errorf("cambiar estado de %s: %w", name, err)
	}
	return nil
}

// DeleteInstalledContent borra un archivo instalado (activo o deshabilitado).
func DeleteInstalledContent(gameDir, kind, name string) error {
	_, current, _, _, err := resolveInstalled(gameDir, kind, name)
	if err != nil {
		return err
	}
	if err := os.Remove(current); err != nil {
		return fmt.Errorf("borrar %s: %w", name, err)
	}
	return nil
}

// MoveInstalledContent mueve un archivo instalado (conservando su estado
// activo/deshabilitado) de un gameDir a otro: global <-> instancias.
func MoveInstalledContent(srcDir, dstDir, kind, name string) error {
	_, current, _, enabled, err := resolveInstalled(srcDir, kind, name)
	if err != nil {
		return err
	}
	if err := EnsureContentDirs(dstDir); err != nil {
		return err
	}
	subdir, err := ContentSubdir(kind)
	if err != nil {
		return err
	}
	base, err := SanitizeFileName(name)
	if err != nil {
		return err
	}
	target := filepath.Join(dstDir, subdir, base)
	if !enabled {
		target += DisabledSuffix
	}
	if sameFile(current, target) {
		return fmt.Errorf("ya está en ese destino")
	}
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("%s ya existe en el destino", base)
	}
	if err := moveInstalledFile(current, target); err != nil {
		return fmt.Errorf("mover %s: %w", base, err)
	}
	return nil
}

func sameFile(a, b string) bool {
	absA, errA := filepath.Abs(a)
	absB, errB := filepath.Abs(b)
	return errA == nil && errB == nil && absA == absB
}

// moveInstalledFile mueve con rename y, si cruza volúmenes, copia y borra.
func moveInstalledFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		os.Remove(dst)
		return copyErr
	}
	if closeErr != nil {
		os.Remove(dst)
		return closeErr
	}
	return os.Remove(src)
}

// RevealInExplorer abre el explorador del sistema con el archivo seleccionado
// (o su carpeta si la plataforma no soporta selección).
func RevealInExplorer(path string) error {
	if path == "" {
		return fmt.Errorf("ruta vacía")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", "/select,"+path)
	case "darwin":
		cmd = exec.Command("open", "-R", path)
	default:
		cmd = exec.Command("xdg-open", filepath.Dir(path))
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("abrir ubicación: %w", err)
	}
	return nil
}

// InstalledPath devuelve la ruta absoluta de un archivo instalado.
func InstalledPath(gameDir, kind, name string) (string, error) {
	_, current, _, _, err := resolveInstalled(gameDir, kind, name)
	if err != nil {
		return "", err
	}
	return current, nil
}
