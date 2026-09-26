package mods

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// iconos: los mods y texturas ya traen su imagen dentro del archivo:
//   - resourcepacks/shaders (.zip): pack.png en la raíz.
//   - mods Fabric (.jar): fabric.mod.json con el campo "icon" (ruta o mapa
//     de tamaños) apuntando a un archivo dentro del jar.
//   - mods Quilt (.jar): quilt.mod.json con metadata.icon de forma análoga.
//   - mods Forge/NeoForge (.jar): META-INF/mods.toml con logoFile="...".
// ExtractContentIcon saca esa imagen a la caché y devuelve su ruta relativa
// al workDir (lista para loadLocal). Si no hay icono, devuelve "" sin error
// para que la UI use su icono genérico sin ensuciar el log.
func ExtractContentIcon(gameDir, kind, name, cacheDir string) (string, error) {
	subdir, err := ContentSubdir(kind)
	if err != nil {
		return "", err
	}
	base, err := SanitizeFileName(name)
	if err != nil {
		return "", err
	}
	var path string
	for _, candidate := range []string{filepath.Join(gameDir, subdir, base), filepath.Join(gameDir, subdir, base+DisabledSuffix)} {
		if info, statErr := os.Stat(candidate); statErr == nil && !info.IsDir() {
			path = candidate
			break
		}
	}
	if path == "" {
		return "", nil
	}
	lower := strings.ToLower(base)
	var entryName string
	var data []byte
	if strings.HasSuffix(lower, ".jar") {
		data, entryName, err = modJarIcon(path)
	} else {
		data, entryName, err = zipRootIcon(path)
	}
	if err != nil || len(data) == 0 {
		// Sin icono embebido: caché por nombre de archivo (icono del proyecto
		// en Modrinth guardado al descargar). Cubre jars sin metadatos.
		return byNameIcon(cacheDir, base)
	}
	ext := strings.ToLower(filepath.Ext(entryName))
	if ext == "" {
		ext = ".png"
	}
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".bmp":
	default:
		return "", nil
	}
	sum := sha1.Sum([]byte(path))
	key := hex.EncodeToString(sum[:])
	outDir := filepath.Join(cacheDir, "content-icons")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}
	out := filepath.Join(outDir, key+ext)
	if info, statErr := os.Stat(out); statErr != nil || info.Size() != int64(len(data)) {
		if err := os.WriteFile(out, data, 0644); err != nil {
			return "", err
		}
	}
	return filepath.ToSlash(filepath.Join("cache", "content-icons", key+ext)), nil
}

// zipRootIcon busca pack.png en la raíz de un zip (texturas/shaders).
func zipRootIcon(path string) ([]byte, string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, "", err
	}
	defer r.Close()
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if filepath.ToSlash(f.Name) == "pack.png" {
			return readZipEntry(f, 2*1024*1024)
		}
	}
	return nil, "", fmt.Errorf("sin pack.png")
}

// modJarIcon resuelve el icono de un mod .jar según su modloader.
func modJarIcon(path string) ([]byte, string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, "", err
	}
	defer r.Close()
	byName := make(map[string]*zip.File, len(r.File))
	for _, f := range r.File {
		byName[filepath.ToSlash(f.Name)] = f
	}
	// Fabric: fabric.mod.json -> "icon" (string o mapa tamaño->ruta).
	if f := byName["fabric.mod.json"]; f != nil {
		if data, name, err := fabricIcon(r, byName, f); err == nil {
			return data, name, nil
		}
	}
	// Quilt: quilt.mod.json -> quilt_loader.metadata.icon (con fallback al
	// metadata de nivel superior por compatibilidad).
	if f := byName["quilt.mod.json"]; f != nil {
		if data, name, err := quiltIcon(r, byName, f); err == nil {
			return data, name, nil
		}
	}
	// Forge/NeoForge: neoforge.mods.toml (nombre nuevo) o mods.toml con
	// logoFile="...".
	for _, toml := range []string{"META-INF/neoforge.mods.toml", "META-INF/mods.toml"} {
		if f := byName[toml]; f != nil {
			if data, name, err := forgeIcon(r, byName, f); err == nil {
				return data, name, nil
			}
		}
	}
	return nil, "", fmt.Errorf("sin icono")
}

func readZipEntry(f *zip.File, max int64) ([]byte, string, error) {
	if f.UncompressedSize64 > uint64(max) {
		return nil, "", fmt.Errorf("icono demasiado grande")
	}
	rc, err := f.Open()
	if err != nil {
		return nil, "", err
	}
	defer rc.Close()
	data, err := io.ReadAll(io.LimitReader(rc, max))
	if err != nil {
		return nil, "", err
	}
	if len(data) == 0 {
		return nil, "", fmt.Errorf("icono vacío")
	}
	return data, f.Name, nil
}

// iconPathOf normaliza el campo icon de fabric/quilt (string o mapa).
func iconPathOf(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	// El mapa es tamaño->ruta: se prefiere una resolución mediana.
	for _, size := range []string{"128", "64", "256", "32", "512", "16"} {
		if p, ok := m[size]; ok && p != "" {
			return p
		}
	}
	for _, p := range m {
		if p != "" {
			return p
		}
	}
	return ""
}

func fabricIcon(r *zip.ReadCloser, byName map[string]*zip.File, meta *zip.File) ([]byte, string, error) {
	data, _, err := readZipEntry(meta, 512*1024)
	if err != nil {
		return nil, "", err
	}
	var doc struct {
		Icon json.RawMessage `json:"icon"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, "", err
	}
	return iconFromDoc(r, byName, doc.Icon)
}

func quiltIcon(r *zip.ReadCloser, byName map[string]*zip.File, meta *zip.File) ([]byte, string, error) {
	data, _, err := readZipEntry(meta, 512*1024)
	if err != nil {
		return nil, "", err
	}
	var doc struct {
		Metadata struct {
			Icon json.RawMessage `json:"icon"`
			Name string          `json:"name"`
		} `json:"metadata"`
		QuiltLoader struct {
			Metadata struct {
				Icon json.RawMessage `json:"icon"`
				Name string          `json:"name"`
			} `json:"metadata"`
		} `json:"quilt_loader"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, "", err
	}
	if len(doc.QuiltLoader.Metadata.Icon) > 0 {
		return iconFromDoc(r, byName, doc.QuiltLoader.Metadata.Icon)
	}
	if len(doc.Metadata.Icon) > 0 {
		return iconFromDoc(r, byName, doc.Metadata.Icon)
	}
	return nil, "", fmt.Errorf("sin icon")
}

func iconFromDoc(r *zip.ReadCloser, byName map[string]*zip.File, raw json.RawMessage) ([]byte, string, error) {
	_ = r
	iconPath := iconPathOf(raw)
	if iconPath == "" {
		return nil, "", fmt.Errorf("sin icon")
	}
	f := byName[filepath.ToSlash(filepath.Clean(iconPath))]
	if f == nil || f.FileInfo().IsDir() {
		return nil, "", fmt.Errorf("icono no encontrado")
	}
	return readZipEntry(f, 2*1024*1024)
}

// forgeIcon lee logoFile del mods.toml (formato clave="valor", con # comentarios).
func forgeIcon(r *zip.ReadCloser, byName map[string]*zip.File, meta *zip.File) ([]byte, string, error) {
	data, _, err := readZipEntry(meta, 512*1024)
	if err != nil {
		return nil, "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.Index(line, "=")
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		if !strings.EqualFold(key, "logoFile") {
			continue
		}
		val := strings.Trim(strings.TrimSpace(line[eq+1:]), `"'`)
		if val == "" {
			continue
		}
		f := byName[filepath.ToSlash(filepath.Clean(val))]
		if f == nil || f.FileInfo().IsDir() {
			continue
		}
		return readZipEntry(f, 2*1024*1024)
	}
	return nil, "", fmt.Errorf("sin logoFile")
}

// ContentDisplayName devuelve el nombre bonito de un mod (name de
// fabric.mod.json, quilt_loader.metadata.name o displayName del mods.toml).
// Vacío si el archivo no trae nombre: la UI muestra el filename.
func ContentDisplayName(gameDir, kind, name string) string {
	subdir, err := ContentSubdir(kind)
	if err != nil {
		return ""
	}
	base, err := SanitizeFileName(name)
	if err != nil {
		return ""
	}
	for _, candidate := range []string{filepath.Join(gameDir, subdir, base), filepath.Join(gameDir, subdir, base+DisabledSuffix)} {
		if title := jarDisplayName(candidate); title != "" {
			return title
		}
	}
	return ""
}

// jarDisplayName lee el nombre declarado dentro de un .jar ("" si no hay).
func jarDisplayName(path string) string {
	if !strings.HasSuffix(strings.ToLower(path), ".jar") {
		return ""
	}
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		return ""
	}
	r, err := zip.OpenReader(path)
	if err != nil {
		return ""
	}
	defer r.Close()
	byName := make(map[string]*zip.File, len(r.File))
	for _, f := range r.File {
		byName[filepath.ToSlash(f.Name)] = f
	}
	if f := byName["fabric.mod.json"]; f != nil {
		if data, _, err := readZipEntry(f, 512*1024); err == nil {
			var doc struct {
				Name string `json:"name"`
			}
			if json.Unmarshal(data, &doc) == nil && strings.TrimSpace(doc.Name) != "" {
				return strings.TrimSpace(doc.Name)
			}
		}
	}
	if f := byName["quilt.mod.json"]; f != nil {
		if data, _, err := readZipEntry(f, 512*1024); err == nil {
			var doc struct {
				QuiltLoader struct {
					Metadata struct {
						Name string `json:"name"`
					} `json:"metadata"`
				} `json:"quilt_loader"`
			}
			if json.Unmarshal(data, &doc) == nil && strings.TrimSpace(doc.QuiltLoader.Metadata.Name) != "" {
				return strings.TrimSpace(doc.QuiltLoader.Metadata.Name)
			}
		}
	}
	for _, toml := range []string{"META-INF/neoforge.mods.toml", "META-INF/mods.toml"} {
		if f := byName[toml]; f != nil {
			if data, _, err := readZipEntry(f, 512*1024); err == nil {
				if title := tomlDisplayName(string(data)); title != "" {
					return title
				}
			}
		}
	}
	return ""
}

// tomlDisplayName saca el primer displayName="..." no comentado del toml.
func tomlDisplayName(data string) string {
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.Index(line, "=")
		if eq < 0 {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(line[:eq]), "displayName") {
			continue
		}
		val := strings.Trim(strings.TrimSpace(line[eq+1:]), `"'`)
		if val != "" {
			return val
		}
	}
	return ""
}

// projectIconKey identifica el icono de proyecto guardado por nombre de archivo.
func projectIconKey(fileName string) string {
	sum := sha1.Sum([]byte(strings.ToLower(fileName)))
	return hex.EncodeToString(sum[:])
}

// CacheProjectIcon guarda el icono del proyecto (Modrinth) para un archivo:
// cubre los mods sin icono embebido. Best-effort, nunca falla la instalación.
func CacheProjectIcon(cacheDir, fileName, iconURL string) {
	if strings.TrimSpace(iconURL) == "" {
		return
	}
	if _, err := SanitizeFileName(fileName); err != nil {
		return
	}
	outDir := filepath.Join(cacheDir, "content-icons", "by-name")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return
	}
	dest := filepath.Join(outDir, projectIconKey(fileName))
	final, err := DownloadIcon(nil, iconURL, dest)
	if err != nil || final == "" {
		return
	}
	// DownloadIcon añade la extensión detectada; se deja tal cual.
	_ = final
}

// byNameIcon devuelve la ruta relativa del icono de proyecto guardado para un
// archivo ("" si no hay).
func byNameIcon(cacheDir, fileName string) (string, error) {
	outDir := filepath.Join(cacheDir, "content-icons", "by-name")
	key := projectIconKey(fileName)
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return "", nil
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), key) {
			continue
		}
		switch strings.ToLower(filepath.Ext(e.Name())) {
		case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".bmp":
			return filepath.ToSlash(filepath.Join("cache", "content-icons", "by-name", e.Name())), nil
		}
	}
	return "", nil
}
