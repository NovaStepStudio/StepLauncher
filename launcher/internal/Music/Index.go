package music

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// MusicIndexEntry representa una entrada indexada ultra-ligera.
// ID es hash estable de Path+Size+ModTime para invalidación.
type MusicIndexEntry struct {
	ID          string  `json:"id"`
	Path        string  `json:"path"`
	Size        int64   `json:"size"`
	ModTime     int64   `json:"modTime"`
	Title       string  `json:"title"`
	Artist      string  `json:"artist"`
	Album       string  `json:"album"`
	Genre       string  `json:"genre"`
	Year        int     `json:"year"`
	TrackNumber int     `json:"trackNumber"`
	DiscNumber  int     `json:"discNumber"`
	Duration    float64 `json:"duration"`
	HasCover    bool    `json:"hasCover"`
	CoverID     string  `json:"coverId,omitempty"`
	CoverMime   string  `json:"coverMime,omitempty"`
	CachedAt    int64   `json:"cachedAt"`
}

// MusicLibraryStats para debugging y Settings.
type MusicLibraryStats struct {
	TotalTracks       int   `json:"totalTracks"`
	TotalFolders      int   `json:"totalFolders"`
	CachedTracks      int   `json:"cachedTracks"`
	CachedCovers      int   `json:"cachedCovers"`
	MissingMetadata   int   `json:"missingMetadata"`
	MissingCovers     int   `json:"missingCovers"`
	LastScanUnix      int64 `json:"lastScanUnix"`
	DatabaseSizeBytes int64 `json:"databaseSizeBytes"`
}

// MusicIndex es la autoridad persistente de metadata.
// Usa gob+gzip single file `index_v2.gob.gz` versionado, migrable desde shards.
type MusicIndex struct {
	mu      sync.RWMutex
	saveMu  sync.Mutex
	path    string
	entries map[string]*MusicIndexEntry // key: Path
	version int
}

const indexVersion = 2

func NewMusicIndex(cacheDir string) *MusicIndex {
	path := filepath.Join(cacheDir, "index_v2.gob.gz")
	return &MusicIndex{
		path:    path,
		entries: make(map[string]*MusicIndexEntry),
		version: indexVersion,
	}
}

// indexFilePayload para persistencia versionada.
type indexFilePayload struct {
	Version int                          `json:"version"`
	Entries map[string]*MusicIndexEntry `json:"entries"`
	SavedAt int64                        `json:"savedAt"`
}

func (idx *MusicIndex) Load() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	// Intentar cargar v2
	var payload indexFilePayload
	if err := readGobGZ(idx.path, &payload); err == nil && payload.Version == indexVersion && payload.Entries != nil {
		idx.entries = payload.Entries
		return nil
	}
	// Fallback: intentar JSON gz viejo o migrar shards
	if tryReadOldJSONGZ(idx.path, &payload) && payload.Entries != nil {
		idx.entries = payload.Entries
		_ = idx.saveLocked()
		return nil
	}
	// Migración desde shards viejos (shard_*.gob.gz en metaDir)
	// El Manager llamará MigrateFromShards si es necesario.
	idx.entries = make(map[string]*MusicIndexEntry)
	return nil
}

func (idx *MusicIndex) saveLocked() error {
	idx.saveMu.Lock()
	defer idx.saveMu.Unlock()
	payload := indexFilePayload{
		Version: indexVersion,
		Entries: idx.entries,
		SavedAt: time.Now().Unix(),
	}
	return writeGobGZ(idx.path, payload)
}

func (idx *MusicIndex) Save() error {
	idx.saveMu.Lock()
	defer idx.saveMu.Unlock()
	idx.mu.RLock()
	payload := indexFilePayload{
		Version: indexVersion,
		Entries: idx.entries,
		SavedAt: time.Now().Unix(),
	}
	idx.mu.RUnlock()
	// Copia para evitar lock prolongado en I/O
	// writeGobGZ hace tmp+Rename atómico
	return writeGobGZ(idx.path, payload)
}

// Get retorna copia de la entrada si existe.
func (idx *MusicIndex) Get(path string) (*MusicIndexEntry, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	e, ok := idx.entries[path]
	if !ok {
		return nil, false
	}
	cp := *e
	return &cp, true
}

// Set inserta o actualiza.
func (idx *MusicIndex) Set(entry *MusicIndexEntry) {
	if entry == nil || entry.Path == "" {
		return
	}
	idx.mu.Lock()
	idx.entries[entry.Path] = entry
	idx.mu.Unlock()
}

// Delete elimina una entrada.
func (idx *MusicIndex) Delete(path string) {
	idx.mu.Lock()
	delete(idx.entries, path)
	idx.mu.Unlock()
}

// List retorna todas las entradas ordenadas por Path.
func (idx *MusicIndex) List() []*MusicIndexEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	out := make([]*MusicIndexEntry, 0, len(idx.entries))
	for _, e := range idx.entries {
		cp := *e
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// Count retorna tamaño.
func (idx *MusicIndex) Count() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.entries)
}

// NeedsUpdate verifica si un archivo necesita re-parsear.
// Retorna true si es nuevo, size/modTime cambió, o no existe en índice.
func (idx *MusicIndex) NeedsUpdate(path string, size int64, modTime int64) bool {
	idx.mu.RLock()
	e, ok := idx.entries[path]
	idx.mu.RUnlock()
	if !ok {
		return true
	}
	return e.Size != size || e.ModTime != modTime
}

// RemoveMissing elimina entradas cuyos archivos ya no existen en las carpetas dadas.
// Retorna cuántos eliminó.
func (idx *MusicIndex) RemoveMissing(existingPaths map[string]bool) int {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	removed := 0
	for path := range idx.entries {
		if !existingPaths[path] {
			delete(idx.entries, path)
			removed++
		}
	}
	return removed
}

// Search filtra por query sobre title/artist/album/path (case-insensitive).
// Si query vacío, retorna todas ordenadas por Path.
func (idx *MusicIndex) Search(query string) []*MusicIndexEntry {
	query = strings.TrimSpace(strings.ToLower(query))
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	if query == "" {
		out := make([]*MusicIndexEntry, 0, len(idx.entries))
		for _, e := range idx.entries {
			cp := *e
			out = append(out, &cp)
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
		return out
	}
	out := make([]*MusicIndexEntry, 0)
	for _, e := range idx.entries {
		if strings.Contains(strings.ToLower(e.Title), query) ||
			strings.Contains(strings.ToLower(e.Artist), query) ||
			strings.Contains(strings.ToLower(e.Album), query) ||
			strings.Contains(strings.ToLower(e.Path), query) {
			cp := *e
			out = append(out, &cp)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// Stats calcula estadísticas.
func (idx *MusicIndex) Stats(totalFolders int) MusicLibraryStats {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	stats := MusicLibraryStats{
		TotalTracks:  len(idx.entries),
		TotalFolders: totalFolders,
	}
	for _, e := range idx.entries {
		if e.Title == "" && e.Artist == "Desconocido" {
			stats.MissingMetadata++
		}
		if e.HasCover {
			stats.CachedCovers++
		} else {
			stats.MissingCovers++
		}
	}
	stats.CachedTracks = stats.TotalTracks
	if fi, err := os.Stat(idx.path); err == nil {
		stats.DatabaseSizeBytes = fi.Size()
	}
	// LastScanUnix se actualiza desde Manager al guardar
	if len(idx.entries) > 0 {
		var max int64
		for _, e := range idx.entries {
			if e.CachedAt > max {
				max = e.CachedAt
			}
		}
		stats.LastScanUnix = max
	}
	return stats
}

// GenerateID crea ID estable para una pista.
func GenerateID(path string, size int64, modTime int64) string {
	h := sha256.Sum256([]byte(path + "|" + string(rune(size)) + "|" + string(rune(modTime))))
	return hex.EncodeToString(h[:8]) // 16 chars hex
}

// MigrateFromShards importa shards viejos shard_*.gob.gz si el índice está vacío.
func (idx *MusicIndex) MigrateFromShards(metaDir string) int {
	idx.mu.Lock()
	if len(idx.entries) > 0 {
		idx.mu.Unlock()
		return 0
	}
	idx.mu.Unlock()
	// Buscar shards en metaDir
	entries, err := os.ReadDir(metaDir)
	if err != nil {
		return 0
	}
	imported := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if !strings.HasPrefix(n, "shard_") || !strings.HasSuffix(n, ".gob.gz") {
			continue
		}
		path := filepath.Join(metaDir, n)
		var mp map[string]MusicMeta
		if err := readGobGZ(path, &mp); err != nil {
			continue
		}
		for trackPath, meta := range mp {
			fi, err := os.Stat(trackPath)
			var size int64
			var modTime int64
			if err == nil {
				size = fi.Size()
				modTime = fi.ModTime().Unix()
			}
			entry := &MusicIndexEntry{
				ID:          GenerateID(trackPath, size, modTime),
				Path:        trackPath,
				Size:        size,
				ModTime:     modTime,
				Title:       meta.Title,
				Artist:      meta.Artist,
				Album:       meta.Album,
				Genre:       meta.Genre,
				Year:        meta.Year,
				TrackNumber: meta.TrackNumber,
				Duration:    meta.Duration,
				HasCover:    meta.HasCover,
				CoverID:     meta.CoverID,
				CoverMime:   meta.CoverMime,
				CachedAt:    time.Now().Unix(),
			}
			if entry.Title == "" {
				entry.Title = strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
			}
			if entry.Artist == "" {
				entry.Artist = "Desconocido"
			}
			idx.Set(entry)
			imported++
		}
	}
	if imported > 0 {
		_ = idx.Save()
	}
	return imported
}
