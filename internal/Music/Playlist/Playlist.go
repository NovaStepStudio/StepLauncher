package playlists

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Playlist representa una colección local de música.
// Puede tener rutas absolutas o relativas (relativas a RootDir o a MusicFolder).
type Playlist struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Favorite      bool     `json:"favorite"`
	Pinned        bool     `json:"pinned"`
	TrackPaths    []string `json:"trackPaths"`
	CustomColor   string   `json:"customColor,omitempty"`
	CustomCover   string   `json:"customCover,omitempty"` // ruta a imagen custom (relativa o absoluta)
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
	TotalDuration float64  `json:"totalDuration,omitempty"` // cacheado en Go, no recalcular en frontend
	TrackCount    int      `json:"trackCount,omitempty"`
	PreviewCovers []string `json:"previewCovers,omitempty"` // carátulas thumb (data URI) de las primeras 4 pistas, cacheadas para lectura instantánea
}

type Manager struct {
	mu          sync.RWMutex
	path        string // launcher_playlists.json (índice)
	dir         string // launcher/playlists/
	rootDir     string
	playlists   []Playlist
	getDuration func(path string) float64
	getCover    func(path string) string // thumb data URI de 128px, para previewCovers
}

func NewManager(rootDir string) *Manager {
	path := filepath.Join(rootDir, "launcher_playlists.json")
	dir := filepath.Join(rootDir, "launcher", "playlists")
	m := &Manager{path: path, dir: dir, rootDir: rootDir}
	m.load()
	return m
}

func (m *Manager) SetDurationResolver(fn func(path string) float64) {
	m.mu.Lock()
	m.getDuration = fn
	m.mu.Unlock()
}

func (m *Manager) SetCoverResolver(fn func(path string) string) {
	m.mu.Lock()
	m.getCover = fn
	m.mu.Unlock()
}

func (m *Manager) computeStats(trackPaths []string) (int, float64) {
	if m.getDuration == nil {
		return len(trackPaths), 0
	}
	var total float64
	for _, p := range trackPaths {
		total += m.getDuration(p)
	}
	return len(trackPaths), total
}

func (m *Manager) computePreviewCovers(trackPaths []string) []string {
	if m.getCover == nil {
		return nil
	}
	limit := 4
	if len(trackPaths) < limit {
		limit = len(trackPaths)
	}
	var out []string
	for i := 0; i < limit; i++ {
		p := strings.TrimSpace(trackPaths[i])
		if p == "" {
			continue
		}
		// Solo guardar si hay carátula (thumb 128), si no, no incluir (frontend mostrará placeholder)
		if thumb := m.getCover(p); thumb != "" {
			out = append(out, thumb)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (m *Manager) load() {
	// Intentar cargar desde directorio individual primero
	if _, err := os.Stat(m.dir); err == nil {
		entries, err := os.ReadDir(m.dir)
		if err == nil {
			var list []Playlist
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				if !strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
					continue
				}
				data, err := os.ReadFile(filepath.Join(m.dir, e.Name()))
				if err != nil {
					continue
				}
				var pl Playlist
				if err := json.Unmarshal(data, &pl); err != nil {
					continue
				}
				if strings.TrimSpace(pl.Title) == "" {
					pl.Title = "Sin título"
				}
				if pl.ID == "" {
					pl.ID = strings.TrimSuffix(e.Name(), ".json")
				}
				list = append(list, pl)
			}
			// Si encontramos playlists en dir, usarlas y reconstruir índice
			if len(list) > 0 {
				// Sanitizar IDs
				for i := range list {
					if list[i].ID == "" {
						list[i].ID = newID()
					}
				}
				m.playlists = list
				_ = m.saveIndex()
				return
			}
		}
	}
	// Si no hay dir o está vacío, intentar cargar índice o formato antiguo
	data, err := os.ReadFile(m.path)
	if err != nil {
		m.playlists = []Playlist{}
		return
	}
	// Intentar como índice nuevo: []string (IDs)
	var ids []string
	if err := json.Unmarshal(data, &ids); err == nil {
		// Verificar que todos sean strings sin campo title (índice)
		// Si es array de strings vacío o con IDs, intentar cargar cada archivo
		if len(ids) == 0 {
			m.playlists = []Playlist{}
			return
		}
		// Si el primer elemento parece ID (sin espacios, hex), tratar como índice
		// Pero también podría ser lista vacía de playlists antigua serializada como []Playlist con 0 elementos -> también es []string vacío, ya manejado
		// Para distinguir, verificar si ids contiene solo strings que parecen IDs (hex 16 chars)
		isIndex := true
		for _, id := range ids {
			if len(id) != 16 {
				// Si alguno no parece ID hex, probablemente no es índice sino error de parsing
				// Pero si es array de strings con títulos? No, títulos suelen ser más largos con espacios
				// Así que si hay strings no hex, no es índice
				// Intentaremos fallback a []Playlist
				if strings.Contains(id, " ") || len(id) > 20 {
					isIndex = false
					break
				}
			}
		}
		if isIndex {
			var list []Playlist
			for _, id := range ids {
				p := filepath.Join(m.dir, id+".json")
				d, err := os.ReadFile(p)
				if err != nil {
					continue
				}
				var pl Playlist
				if err := json.Unmarshal(d, &pl); err != nil {
					continue
				}
				list = append(list, pl)
			}
			// Si índice apunta a archivos inexistentes pero hay datos, fallback
			if len(list) > 0 || len(ids) == 0 {
				m.playlists = list
				return
			}
		}
	}
	// Intentar como formato antiguo: []Playlist
	var list []Playlist
	if err := json.Unmarshal(data, &list); err != nil {
		m.playlists = []Playlist{}
		return
	}
	// Sanitizar y migrar a nuevo formato
	for i := range list {
		if strings.TrimSpace(list[i].Title) == "" {
			list[i].Title = "Sin título"
		}
		if list[i].ID == "" {
			list[i].ID = newID()
		}
	}
	m.playlists = list
	// Migrar: guardar cada playlist individual y reescribir índice
	_ = os.MkdirAll(m.dir, 0755)
	for _, pl := range m.playlists {
		_ = m.savePlaylistFile(pl)
	}
	_ = m.saveIndex()
}

func (m *Manager) savePlaylistFile(pl Playlist) error {
	if err := os.MkdirAll(m.dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(pl, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(m.dir, pl.ID+".json.tmp")
	final := filepath.Join(m.dir, pl.ID+".json")
	if err := os.WriteFile(tmp, append(data, '\n'), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, final)
}

func (m *Manager) saveIndex() error {
	m.mu.RLock()
	ids := make([]string, len(m.playlists))
	for i, pl := range m.playlists {
		ids[i] = pl.ID
	}
	m.mu.RUnlock()
	data, err := json.MarshalIndent(ids, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
		return err
	}
	tmp := m.path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, m.path)
}

func (m *Manager) List() []Playlist {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Playlist, len(m.playlists))
	copy(out, m.playlists)
	return out
}

func (m *Manager) Get(id string) (*Playlist, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.playlists {
		if p.ID == id {
			cp := p
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("playlist no encontrada")
}

func (m *Manager) Create(title string, trackPaths []string) (*Playlist, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("título requerido")
	}
	norm := normalizePaths(trackPaths)
	// Deduplicación sin hash: título exacto (case-insensitive) + mismo conjunto de pistas
	m.mu.RLock()
	for _, pl := range m.playlists {
		if strings.EqualFold(strings.TrimSpace(pl.Title), title) {
			if len(pl.TrackPaths) == len(norm) {
				same := true
				for i := range norm {
					if strings.ToLower(filepath.Clean(pl.TrackPaths[i])) != strings.ToLower(filepath.Clean(norm[i])) {
						same = false
						break
					}
				}
				if same {
					cp := pl
					m.mu.RUnlock()
					return &cp, nil
				}
			}
		}
	}
	m.mu.RUnlock()
	now := time.Now().Format(time.RFC3339)
	trackCount, totalDur := m.computeStats(norm)
	preview := m.computePreviewCovers(norm)
	pl := Playlist{
		ID:            newID(),
		Title:         title,
		TrackPaths:    norm,
		CreatedAt:     now,
		UpdatedAt:     now,
		TrackCount:    trackCount,
		TotalDuration: totalDur,
		PreviewCovers: preview,
	}
	m.mu.Lock()
	m.playlists = append(m.playlists, pl)
	m.mu.Unlock()
	if err := m.savePlaylistFile(pl); err != nil {
		return nil, err
	}
	if err := m.saveIndex(); err != nil {
		return nil, err
	}
	return &pl, nil
}

func (m *Manager) Update(id string, upd Playlist) (*Playlist, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, p := range m.playlists {
		if p.ID == id {
			upd.ID = id
			upd.CreatedAt = p.CreatedAt
			upd.UpdatedAt = time.Now().Format(time.RFC3339)
			if strings.TrimSpace(upd.Title) == "" {
				upd.Title = p.Title
			}
			upd.TrackPaths = normalizePaths(upd.TrackPaths)
			upd.TrackCount, upd.TotalDuration = m.computeStats(upd.TrackPaths)
			// Preservar PreviewCovers si no cambió el orden de las primeras 4, sino recalcular
			needPreview := true
			if len(p.TrackPaths) >= 4 && len(upd.TrackPaths) >= 4 {
				same := true
				for k := 0; k < 4; k++ {
					if p.TrackPaths[k] != upd.TrackPaths[k] {
						same = false
						break
					}
				}
				if same && len(p.PreviewCovers) > 0 {
					upd.PreviewCovers = p.PreviewCovers
					needPreview = false
				}
			}
			if needPreview {
				upd.PreviewCovers = m.computePreviewCovers(upd.TrackPaths)
			}
			m.playlists[i] = upd
			// Guardar archivo individual
			data, _ := json.MarshalIndent(upd, "", "  ")
			os.MkdirAll(m.dir, 0755)
			tmp := filepath.Join(m.dir, id+".json.tmp")
			final := filepath.Join(m.dir, id+".json")
			os.WriteFile(tmp, append(data, '\n'), 0644)
			os.Rename(tmp, final)
			// Guardar índice (IDs no cambian, pero reescribir por si acaso)
			ids := make([]string, len(m.playlists))
			for j, pp := range m.playlists {
				ids[j] = pp.ID
			}
			idxData, _ := json.MarshalIndent(ids, "", "  ")
			os.MkdirAll(filepath.Dir(m.path), 0755)
			tmpIdx := m.path + ".tmp"
			os.WriteFile(tmpIdx, append(idxData, '\n'), 0644)
			os.Rename(tmpIdx, m.path)
			cp := upd
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("playlist no encontrada")
}

func (m *Manager) RefreshMissingStats() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	changed := false
	for i, pl := range m.playlists {
		needsStats := pl.TrackCount == 0 || pl.TotalDuration == 0
		needsPreview := len(pl.PreviewCovers) == 0 && len(pl.TrackPaths) > 0
		if needsStats {
			tc, td := m.computeStats(pl.TrackPaths)
			if tc != pl.TrackCount || td != pl.TotalDuration {
				m.playlists[i].TrackCount = tc
				m.playlists[i].TotalDuration = td
				changed = true
			}
		}
		if needsPreview {
			if preview := m.computePreviewCovers(pl.TrackPaths); len(preview) > 0 {
				m.playlists[i].PreviewCovers = preview
				changed = true
			}
		}
		if needsStats || needsPreview {
			if changed {
				_ = m.savePlaylistFile(m.playlists[i])
			}
		}
	}
	if changed {
		_ = m.saveIndex()
	}
	return nil
}

func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, p := range m.playlists {
		if p.ID == id {
			m.playlists = append(m.playlists[:i], m.playlists[i+1:]...)
			// Borrar archivo individual
			_ = os.Remove(filepath.Join(m.dir, id+".json"))
			// Guardar índice
			ids := make([]string, len(m.playlists))
			for j, pp := range m.playlists {
				ids[j] = pp.ID
			}
			idxData, _ := json.MarshalIndent(ids, "", "  ")
			os.MkdirAll(filepath.Dir(m.path), 0755)
			tmpIdx := m.path + ".tmp"
			os.WriteFile(tmpIdx, append(idxData, '\n'), 0644)
			os.Rename(tmpIdx, m.path)
			return nil
		}
	}
	return fmt.Errorf("playlist no encontrada")
}

func normalizePaths(in []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, p := range in {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		clean := filepath.Clean(p)
		if clean == "." {
			continue
		}
		if !seen[clean] {
			seen[clean] = true
			out = append(out, clean)
		}
	}
	return out
}

// ResolvePath resuelve una ruta de track a absoluta para lectura.
// Si es absoluta, la valida; si es relativa, la une a rootDir o musicFolder.
func (m *Manager) ResolvePath(p string, musicFolder string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	if musicFolder != "" {
		cand := filepath.Join(musicFolder, p)
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
	}
	return filepath.Join(m.rootDir, p)
}

// ResolveFromBase resuelve una ruta de track que puede ser relativa a la
// playlist base (donde está el .m3u) o absoluta. Intenta primero relativa a base,
// luego a musicFolders, luego a rootDir.
func (m *Manager) ResolveFromBase(p string, baseDir string, musicFolders []string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	// 1) Relativa a la carpeta del .m3u
	if baseDir != "" {
		cand := filepath.Join(baseDir, p)
		if _, err := os.Stat(cand); err == nil {
			return filepath.Clean(cand)
		}
		// Aunque no exista, devolver la ruta resuelta para intentar cargar luego
		// Si la ruta contiene .. o es relativa con subcarpetas, igualmente la resolvemos
		// como baseDir + p (aunque no exista en FS, puede ser que se importe y luego el archivo no exista, pero lo guardamos)
		return filepath.Clean(cand)
	}
	// 2) Intentar en cada musicFolder configurada
	for _, mf := range musicFolders {
		if mf == "" {
			continue
		}
		cand := filepath.Join(mf, p)
		if _, err := os.Stat(cand); err == nil {
			return filepath.Clean(cand)
		}
	}
	return filepath.Join(m.rootDir, p)
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
