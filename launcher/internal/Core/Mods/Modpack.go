package mods

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MrpackIndex es el manifiesto modrinth.index.json de un modpack .mrpack:
// lista los archivos a descargar (path + downloads[0] + hashes) y las
// dependencias (versión de Minecraft y versión del loader).
type MrpackIndex struct {
	FormatVersion int               `json:"formatVersion"`
	Game          string            `json:"game"`
	VersionID     string            `json:"versionId"`
	Name          string            `json:"name"`
	Files         []MrpackFile      `json:"files"`
	Dependencies  map[string]string `json:"dependencies"`
}

// MrpackFile es un archivo del manifiesto con su destino relativo al gameDir.
type MrpackFile struct {
	Path      string            `json:"path"`
	Hashes    map[string]string `json:"hashes"`
	Env       map[string]string `json:"env"`
	Downloads []string          `json:"downloads"`
	FileSize  int64             `json:"fileSize"`
}

// SHA1 devuelve el hash sha1 del archivo (o "" si no lo trae).
func (f MrpackFile) SHA1() string {
	if f.Hashes == nil {
		return ""
	}
	return f.Hashes["sha1"]
}

// SHA512 devuelve el hash sha512 del archivo (o "" si no lo trae).
func (f MrpackFile) SHA512() string {
	if f.Hashes == nil {
		return ""
	}
	return f.Hashes["sha512"]
}

// ParseMrpackIndex lee y valida el manifiesto de un modpack ya extraído.
func ParseMrpackIndex(dir string) (*MrpackIndex, error) {
	data, err := os.ReadFile(filepath.Join(dir, "modrinth.index.json"))
	if err != nil {
		return nil, fmt.Errorf("leer modrinth.index.json: %w", err)
	}
	var index MrpackIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("parsear modrinth.index.json: %w", err)
	}
	if index.Game != "" && index.Game != "minecraft" {
		return nil, fmt.Errorf("el modpack no es de minecraft (game=%s)", index.Game)
	}
	if len(index.Files) == 0 {
		return nil, fmt.Errorf("el manifiesto no lista archivos para descargar")
	}
	return &index, nil
}

// MinecraftVersion devuelve la versión de Minecraft que exige el modpack.
func (m *MrpackIndex) MinecraftVersion() string {
	if m.Dependencies == nil {
		return ""
	}
	return strings.TrimSpace(m.Dependencies["minecraft"])
}

// RequiredLoader detecta el modloader que exige el modpack desde sus
// dependencias. Devuelve el id del loader del launcher y su versión. El
// manifiesto trae la versión en el formato que usa cada ecosistema: Forge la
// trae corta ("36.2.34") y aquí se normaliza a completa ("1.16.5-36.2.34"),
// que es lo que exige el Maven de Forge para el installer.
func (m *MrpackIndex) RequiredLoader() (loader, version string) {
	if m.Dependencies == nil {
		return "", ""
	}
	// Orden de prioridad: el primer loader presente en el manifiesto manda.
	for _, key := range []string{"fabric-loader", "quilt-loader", "forge", "neoforge"} {
		if v, ok := m.Dependencies[key]; ok && strings.TrimSpace(v) != "" {
			version = NormalizeLoaderVersion(key, strings.TrimSpace(v), m.MinecraftVersion())
			switch key {
			case "fabric-loader":
				return "fabric", version
			case "quilt-loader":
				return "quilt", version
			case "forge":
				return "forge", version
			case "neoforge":
				return "neoforge", version
			}
		}
	}
	return "", ""
}

// NormalizeLoaderVersion adapta la versión del manifiesto al formato que
// espera cada proveedor: Forge necesita la versión completa de Maven
// ("<mc>-<loader>", p. ej. "1.16.5-36.2.34"); el resto usa la corta tal cual.
func NormalizeLoaderVersion(manifestKey, version, mcVersion string) string {
	if manifestKey == "forge" && mcVersion != "" && !strings.HasPrefix(version, mcVersion+"-") {
		return mcVersion + "-" + version
	}
	return version
}

// ExtractMrpack descomprime un .mrpack (zip con otra extensión) a destDir.
// Rechaza rutas absolutas, ".." y enlaces para no escribir fuera del destino.
func ExtractMrpack(mrpackPath, destDir string) error {
	r, err := zip.OpenReader(mrpackPath)
	if err != nil {
		return fmt.Errorf("abrir modpack: %w", err)
	}
	defer r.Close()
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("crear destino: %w", err)
	}
	for _, f := range r.File {
		rel, err := sanitizeZipPath(f.Name)
		if err != nil {
			continue
		}
		target := filepath.Join(destDir, rel)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("crear carpeta: %w", err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("crear carpeta: %w", err)
		}
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("leer %s: %w", f.Name, err)
		}
		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return fmt.Errorf("crear %s: %w", rel, err)
		}
		_, copyErr := io.Copy(out, rc)
		closeErr := out.Close()
		rc.Close()
		if copyErr != nil {
			return fmt.Errorf("escribir %s: %w", rel, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("cerrar %s: %w", rel, closeErr)
		}
	}
	return nil
}

// SanitizeContentPath valida una ruta relativa del manifiesto (p. ej.
// "mods/foo.jar") y la normaliza para unirla al gameDir sin escapes.
func SanitizeContentPath(name string) (string, error) {
	return sanitizeZipPath(name)
}

// sanitizeZipPath normaliza una ruta dentro del zip y rechaza escapes. Toda
// entrada con ".." se descarta aunque al normalizar caiga dentro del destino:
// un manifiesto o zip legítimo nunca sube de nivel.
func sanitizeZipPath(name string) (string, error) {
	raw := strings.ReplaceAll(name, "\\", "/")
	for _, seg := range strings.Split(raw, "/") {
		if seg == ".." {
			return "", fmt.Errorf("ruta fuera del destino: %s", name)
		}
	}
	clean := filepath.ToSlash(filepath.Clean("/" + raw))
	rel := strings.TrimPrefix(clean, "/")
	if rel == "" || rel == "." || strings.HasPrefix(rel, "../") || strings.Contains(rel, "../") {
		return "", fmt.Errorf("ruta fuera del destino: %s", name)
	}
	return filepath.FromSlash(rel), nil
}

// CopyOverrides mueve el contenido de overrides/ (config, resourcepacks y
// demás archivos del modpack) al gameDir de destino, fusionando carpetas.
func CopyOverrides(extractedDir, gameDir string) (int, error) {
	src := filepath.Join(extractedDir, "overrides")
	if info, err := os.Stat(src); err != nil || !info.IsDir() {
		return 0, nil
	}
	moved := 0
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return nil
		}
		if _, err := sanitizeZipPath(filepath.ToSlash(rel)); err != nil {
			return nil
		}
		target := filepath.Join(gameDir, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("crear carpeta de %s: %w", rel, err)
		}
		if err := moveOrCopy(path, target); err != nil {
			return fmt.Errorf("mover %s: %w", rel, err)
		}
		moved++
		return nil
	})
	if err != nil {
		return moved, err
	}
	return moved, nil
}

// moveOrCopy intenta mover con rename y, si cruza volúmenes, copia y borra.
func moveOrCopy(src, dst string) error {
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
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Remove(src)
}
