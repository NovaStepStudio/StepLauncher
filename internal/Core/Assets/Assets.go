package assets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	AssetsFileName = "launcher_assets.json"
	LauncherDir    = "launcher"
	FontsSubDir    = "fonts"
	ImagesSubDir   = "images"
)

type FontSlot struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type Assets struct {
	Fonts []FontSlot `json:"fonts"`
}

func Default() Assets {
	return Assets{Fonts: []FontSlot{}}
}

func LauncherDirOf(rootDir string) string {
	return filepath.Join(rootDir, LauncherDir)
}

func FontsDirOf(rootDir string) string {
	return filepath.Join(rootDir, LauncherDir, FontsSubDir)
}

func ImagesDirOf(rootDir string) string {
	return filepath.Join(rootDir, LauncherDir, ImagesSubDir)
}

var fontExts = map[string]bool{
	".ttf": true, ".otf": true, ".woff": true, ".woff2": true,
}

func IsFontExt(ext string) bool {
	return fontExts[strings.ToLower(ext)]
}

type Manager struct {
	rootDir string
	path    string
}

func NewManager(rootDir string) *Manager {
	return &Manager{
		rootDir: rootDir,
		path:    filepath.Join(rootDir, AssetsFileName),
	}
}

func (m *Manager) RootDir() string  { return m.rootDir }
func (m *Manager) Path() string     { return m.path }
func (m *Manager) FontsDir() string { return FontsDirOf(m.rootDir) }

func (m *Manager) Ensure() error {
	if _, err := os.Stat(m.path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
		return err
	}
	return m.Save(Default())
}

func (m *Manager) Load() (Assets, error) {
	data, err := os.ReadFile(m.path)
	if err != nil {
		return Default(), err
	}
	a, migrated := parseAssets(data)
	if migrated {
		m.Save(a)
	}
	return a, nil
}

type legacyFonts struct {
	Primary   FontSlot `json:"primary"`
	Secundary FontSlot `json:"secundary"`
}

func parseAssets(data []byte) (Assets, bool) {
	var a Assets
	if err := json.Unmarshal(data, &a); err == nil {
		return a, false
	}
	var legacy struct {
		Fonts *legacyFonts `json:"fonts"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil || legacy.Fonts == nil {
		return Default(), false
	}
	var out []FontSlot
	for _, s := range []FontSlot{legacy.Fonts.Primary, legacy.Fonts.Secundary} {
		if strings.TrimSpace(s.Path) != "" {
			out = append(out, s)
		}
	}
	return Assets{Fonts: out}, true
}

func (m *Manager) Save(a Assets) error {
	m.normalize(&a)
	if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, append(data, '\n'), 0644)
}

func (m *Manager) normalize(a *Assets) {
	var keep []FontSlot
	seenPaths := map[string]bool{}
	for _, s := range a.Fonts {
		s.Type = strings.TrimSpace(strings.ToLower(s.Type))
		s.Name = strings.TrimSpace(s.Name)
		clean := filepath.ToSlash(filepath.Clean(strings.ReplaceAll(s.Path, "\\", "/")))
		if clean == "" || clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, "../") {
			continue
		}
		if !strings.HasPrefix(clean, LauncherDir+"/") || seenPaths[clean] {
			continue
		}
		seenPaths[clean] = true
		s.Path = clean
		keep = append(keep, s)
	}
	a.Fonts = keep
}

func (m *Manager) ListFontFiles() ([]string, error) {
	entries, err := os.ReadDir(m.FontsDir())
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if fontExts[strings.ToLower(filepath.Ext(e.Name()))] {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

func removeWithRetry(path string) error {
	var lastErr error
	for i := 0; i < 4; i++ {
		err := os.Remove(path)
		if err == nil {
			return nil
		}
		if os.IsNotExist(err) {
			return nil
		}
		lastErr = err
		time.Sleep(150 * time.Millisecond)
	}
	return lastErr
}

func (m *Manager) DeleteFontFile(name string) error {
	if name == "" || filepath.Base(name) != name {
		return fmt.Errorf("nombre de archivo invalido")
	}
	if !fontExts[strings.ToLower(filepath.Ext(name))] {
		return fmt.Errorf("extension de fuente no soportada")
	}
	path := filepath.Join(m.FontsDir(), name)
	if err := removeWithRetry(path); err != nil {
		return fmt.Errorf("no se pudo eliminar %s (el archivo puede estar en uso): %v", name, err)
	}

	a, err := m.Load()
	if err != nil {
		return err
	}
	ref := LauncherDir + "/" + FontsSubDir + "/" + name
	var keep []FontSlot
	changed := false
	for _, s := range a.Fonts {
		if s.Path == ref {
			changed = true
			continue
		}
		keep = append(keep, s)
	}
	if changed {
		a.Fonts = keep
		return m.Save(a)
	}
	return nil
}
