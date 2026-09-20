package assets

import (
	"encoding/json"
	"fmt"
	"io"
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
	CacheDir       = "cache"
	AudioSubDir    = "audio"
)

// Máximo de pistas de música de fondo permitidas y tamaño máximo de cada
// archivo de audio (15 MB).
const (
	MaxMusicTracks = 5
	MaxAudioBytes  = 15 * 1024 * 1024
)

type FontSlot struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type MusicSlot struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type GalleryBackground struct {
	Path    string `json:"path"`
	Author  string `json:"author"`
	ModName string `json:"modName"`
	Url     string `json:"url"`
	Title   string `json:"title,omitempty"`
}

type Assets struct {
	Fonts   []FontSlot          `json:"fonts"`
	Music   []MusicSlot         `json:"audio"`
	Gallery []GalleryBackground `json:"gallery,omitempty"`
}

func Default() Assets {
	return Assets{Fonts: []FontSlot{}, Music: []MusicSlot{}, Gallery: []GalleryBackground{}}
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

func AudioDirOf(rootDir string) string {
	return filepath.Join(rootDir, CacheDir, AudioSubDir)
}

var fontExts = map[string]bool{
	".ttf": true, ".otf": true, ".woff": true, ".woff2": true,
}

var audioExts = map[string]bool{
	".mp3": true, ".wav": true, ".ogg": true, ".m4a": true,
}

func IsFontExt(ext string) bool {
	return fontExts[strings.ToLower(ext)]
}

func IsAudioExt(ext string) bool {
	return audioExts[strings.ToLower(ext)]
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
func (m *Manager) AudioDir() string { return AudioDirOf(m.rootDir) }

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
	a.Music = normalizeMusic(a.Music)
	a.Gallery = normalizeGallery(a.Gallery)
}

// normalizeMusic conserva solo pistas válidas: con nombre, con ruta relativa
// dentro de cache/audio/ y sin duplicados.
func normalizeMusic(list []MusicSlot) []MusicSlot {
	var keep []MusicSlot
	seen := map[string]bool{}
	for _, s := range list {
		s.Name = strings.TrimSpace(s.Name)
		clean := filepath.ToSlash(filepath.Clean(strings.ReplaceAll(s.Path, "\\", "/")))
		if clean == "" || clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, "../") {
			continue
		}
		prefix := CacheDir + "/" + AudioSubDir + "/"
		if s.Name == "" || !strings.HasPrefix(clean, prefix) || seen[clean] {
			continue
		}
		seen[clean] = true
		s.Path = clean
		keep = append(keep, s)
	}
	return keep
}

func normalizeGallery(list []GalleryBackground) []GalleryBackground {
	var keep []GalleryBackground
	seen := map[string]bool{}
	for _, g := range list {
		g.Author = strings.TrimSpace(g.Author)
		g.ModName = strings.TrimSpace(g.ModName)
		g.Url = strings.TrimSpace(g.Url)
		clean := filepath.ToSlash(filepath.Clean(strings.ReplaceAll(g.Path, "\\", "/")))
		if clean == "" || clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, "../") {
			continue
		}
		// Permitir tanto cache/backgrounds como launcher/images
		if !(strings.HasPrefix(clean, CacheDir+"/backgrounds/") || strings.HasPrefix(clean, LauncherDir+"/"+ImagesSubDir+"/")) {
			// Permitir cache/backgrounds/* para compatibilidad con ImportBackground
			if !strings.HasPrefix(clean, "cache/backgrounds/") {
				continue
			}
		}
		if seen[clean] {
			continue
		}
		seen[clean] = true
		g.Path = clean
		keep = append(keep, g)
	}
	return keep
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

// ListMusic devuelve las pistas de música de fondo registradas en
// launcher_assets.json (ruta relativa al root del launcher).
func (m *Manager) ListMusic() ([]MusicSlot, error) {
	a, err := m.Load()
	if err != nil {
		return nil, err
	}
	return a.Music, nil
}

// AddMusic copia el archivo de audio `src` (MP3/WAV/OGG/M4A, <= 15 MB) a
// cache/audio/, lo registra en launcher_assets.json como pista de música de
// fondo y devuelve la ruta relativa. Como máximo caben MaxMusicTracks pistas.
func (m *Manager) AddMusic(name, src string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("el nombre del audio no puede estar vacio")
	}
	if !strings.ContainsRune(name, '.') || !IsAudioExt(filepath.Ext(name)) {
		return "", fmt.Errorf("extension de audio no soportada (usa mp3, wav, ogg o m4a)")
	}
	st, err := os.Stat(src)
	if err != nil {
		return "", fmt.Errorf("no se pudo leer el archivo: %v", err)
	}
	if st.Size() > MaxAudioBytes {
		return "", fmt.Errorf("el audio pesa mas de 15 MB")
	}

	a, err := m.Load()
	if err != nil {
		return "", err
	}
	if len(a.Music) >= MaxMusicTracks {
		return "", fmt.Errorf("maximo de %d audios de fondo", MaxMusicTracks)
	}

	ext := strings.ToLower(filepath.Ext(name))
	// TrimSuffix es case-sensitive; el recorte se hace sobre la longitud de la
	// extensión ya normalizada para no duplicarla ("Song.MP3" -> "Song.MP3.mp3").
	base := name[:len(name)-len(ext)]
	fileName := sanitizeFileName(base) + ext
	dest := filepath.Join(m.AudioDir(), fileName)
	for i := 2; fileExists(dest); i++ {
		fileName = sanitizeFileName(base) + fmt.Sprintf(" (%d)", i) + ext
		dest = filepath.Join(m.AudioDir(), fileName)
	}
	if err := os.MkdirAll(m.AudioDir(), 0755); err != nil {
		return "", err
	}
	if err := copyFile(src, dest); err != nil {
		return "", fmt.Errorf("no se pudo copiar el audio: %v", err)
	}

	displayName := sanitizeFileName(base)
	for i := 2; musicNameExists(a.Music, displayName); i++ {
		displayName = sanitizeFileName(base) + fmt.Sprintf(" (%d)", i)
	}
	a.Music = append(a.Music, MusicSlot{
		Name: displayName,
		Path: filepath.ToSlash(filepath.Join(CacheDir, AudioSubDir, fileName)),
	})
	if err := m.Save(a); err != nil {
		_ = removeWithRetry(dest)
		return "", err
	}
	return a.Music[len(a.Music)-1].Path, nil
}

// RemoveMusic elimina la pista `name` del launcher_assets.json y su archivo
// de cache/audio/ (si ninguna otra pista lo comparte).
func (m *Manager) RemoveMusic(name string) error {
	a, err := m.Load()
	if err != nil {
		return err
	}
	var removed *MusicSlot
	var keep []MusicSlot
	for _, s := range a.Music {
		if s.Name == name {
			removed = &s
			continue
		}
		keep = append(keep, s)
	}
	if removed == nil {
		return nil
	}
	a.Music = keep
	if err := m.Save(a); err != nil {
		return err
	}
	shared := false
	for _, s := range keep {
		if s.Path == removed.Path {
			shared = true
			break
		}
	}
	if !shared {
		fileName := filepath.Base(removed.Path)
		_ = removeWithRetry(filepath.Join(m.AudioDir(), fileName))
	}
	return nil
}

// RegisterGalleryBackground registra una imagen de galería descargada en launcher_assets.json.
// Guarda ruta relativa (cache/backgrounds/xxx.png), autor y nombre del mod.
func (m *Manager) RegisterGalleryBackground(relPath, author, modName, url, title string) error {
	relPath = filepath.ToSlash(filepath.Clean(strings.ReplaceAll(relPath, "\\", "/")))
	if relPath == "" || filepath.IsAbs(relPath) || strings.HasPrefix(relPath, "../") {
		return fmt.Errorf("ruta invalida")
	}
	a, err := m.Load()
	if err != nil {
		a = Default()
	}
	urlNorm := strings.TrimSpace(url)
	authorNorm := strings.TrimSpace(author)
	modNorm := strings.TrimSpace(modName)
	titleNorm := strings.TrimSpace(title)

	// Deduplicación por URL: si la misma URL ya existe con otro path, eliminar la entrada antigua
	// y su archivo (evita acumulación de duplicados como en el reporte del usuario donde la misma
	// URL de Modrinth aparecía con varios timestamps distintos).
	if urlNorm != "" {
		var filtered []GalleryBackground
		for _, g := range a.Gallery {
			if strings.TrimSpace(g.Url) == urlNorm && g.Path != relPath {
				_ = removeWithRetry(filepath.Join(m.rootDir, filepath.FromSlash(g.Path)))
				continue
			}
			filtered = append(filtered, g)
		}
		a.Gallery = filtered
	}
	// Evitar duplicados por path (actualizar metadata)
	for i, g := range a.Gallery {
		if g.Path == relPath {
			a.Gallery[i].Author = authorNorm
			a.Gallery[i].ModName = modNorm
			a.Gallery[i].Url = urlNorm
			a.Gallery[i].Title = titleNorm
			return m.Save(a)
		}
	}
	a.Gallery = append(a.Gallery, GalleryBackground{
		Path:    relPath,
		Author:  authorNorm,
		ModName: modNorm,
		Url:     urlNorm,
		Title:   titleNorm,
	})
	return m.Save(a)
}

// PruneOrphanGallery elimina del launcher_assets.json todas las entradas de gallery no referenciadas
// en `referenced` (map con claves por path relativo completo y/o basename). Borra también el archivo
// físico huérfano si existe. Es la contraparte en assets de Config.cleanupBackgrounds: tras cambiar
// o quitar el fondo, las entradas que ya no apuntan a ImagePath/VideoPath/DynamicImages se purgan.
// Retorna cuántas entradas se eliminaron.
func (m *Manager) PruneOrphanGallery(referenced map[string]bool) (int, error) {
	if m == nil {
		return 0, nil
	}
	a, err := m.Load()
	if err != nil {
		return 0, err
	}
	if len(a.Gallery) == 0 {
		return 0, nil
	}
	var keep []GalleryBackground
	removed := 0
	for _, g := range a.Gallery {
		clean := filepath.ToSlash(filepath.Clean(strings.ReplaceAll(g.Path, "\\", "/")))
		base := filepath.Base(clean)
		if referenced[clean] || referenced[base] || referenced[g.Path] {
			keep = append(keep, g)
			continue
		}
		removed++
		full := filepath.Join(m.rootDir, filepath.FromSlash(clean))
		if _, err := os.Stat(full); err == nil {
			_ = removeWithRetry(full)
		}
	}
	if removed > 0 {
		a.Gallery = keep
		if err := m.Save(a); err != nil {
			return 0, err
		}
	}
	return removed, nil
}

// RemoveGalleryBackground elimina una entrada concreta de gallery por su path relativo y borra su archivo.
func (m *Manager) RemoveGalleryBackground(relPath string) error {
	relPath = filepath.ToSlash(filepath.Clean(strings.ReplaceAll(relPath, "\\", "/")))
	if relPath == "" {
		return fmt.Errorf("ruta vacia")
	}
	a, err := m.Load()
	if err != nil {
		return err
	}
	var keep []GalleryBackground
	var removed *GalleryBackground
	for _, g := range a.Gallery {
		if g.Path == relPath {
			c := g
			removed = &c
			continue
		}
		keep = append(keep, g)
	}
	if removed == nil {
		return nil
	}
	a.Gallery = keep
	if err := m.Save(a); err != nil {
		return err
	}
	_ = removeWithRetry(filepath.Join(m.rootDir, filepath.FromSlash(removed.Path)))
	return nil
}

func sanitizeFileName(name string) string {
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
		return "audio"
	}
	return out
}

func musicNameExists(list []MusicSlot, name string) bool {
	for _, s := range list {
		if s.Name == name {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	_, cErr := io.Copy(out, in)
	closeErr := out.Close()
	if cErr != nil {
		_ = removeWithRetry(dest)
		return cErr
	}
	return closeErr
}
