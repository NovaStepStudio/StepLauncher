package musichistory

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Entry representa una pista reproducida - CoverID apunta a cache/music/covers/*.zip
type Entry struct {
	Path      string `json:"path"`
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	CoverUrl  string `json:"coverUrl,omitempty"` // legado: data URI, migrado a CoverID
	CoverID   string `json:"coverId,omitempty"`  // hash de la carátula en cache comprimido
	PlayedAt  string `json:"playedAt"`
	PlayCount int    `json:"playCount"`
}

type Manager struct {
	mu   sync.RWMutex
	path string
	list []Entry
}

func NewManager(rootDir string) *Manager {
	p := filepath.Join(rootDir, "launcher_music_history.json")
	m := &Manager{path: p}
	m.load()
	return m
}

func coverIDFromDataURI(uri string) string {
	if !strings.HasPrefix(uri, "data:") {
		return ""
	}
	comma := strings.Index(uri, ",")
	if comma < 0 {
		return ""
	}
	b64 := uri[comma+1:]
	if b64 == "" {
		return ""
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		data, err = base64.RawStdEncoding.DecodeString(b64)
		if err != nil {
			return ""
		}
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func (m *Manager) load() {
	data, err := os.ReadFile(m.path)
	if err != nil {
		m.list = []Entry{}
		return
	}
	var list []Entry
	if err := json.Unmarshal(data, &list); err != nil {
		m.list = []Entry{}
		return
	}
	// migrar CoverUrl data URI -> CoverID y limpiar para no guardar 500 covers duplicados
	changed := false
	for i, e := range list {
		if e.CoverUrl != "" && strings.HasPrefix(e.CoverUrl, "data:") && e.CoverID == "" {
			if id := coverIDFromDataURI(e.CoverUrl); id != "" {
				list[i].CoverID = id
				list[i].CoverUrl = ""
				changed = true
			} else if len(e.CoverUrl) > 1024 {
				list[i].CoverUrl = ""
				changed = true
			}
		}
		// si CoverUrl es muy grande (>2KB), limpiarlo
		if len(e.CoverUrl) > 2048 {
			list[i].CoverUrl = ""
			changed = true
		}
	}
	m.list = list
	if changed {
		_ = m.save()
	}
}

func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.list, "", "  ")
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

// Add registra una reproducción - guarda CoverID (hash) no data URI gigante
func (m *Manager) Add(path, title, artist, coverUrl string) []Entry {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().Format(time.RFC3339)
	coverID := ""
	cleanCoverUrl := coverUrl
	if strings.HasPrefix(coverUrl, "data:") {
		coverID = coverIDFromDataURI(coverUrl)
		cleanCoverUrl = ""
	} else if len(coverUrl) > 2048 {
		cleanCoverUrl = ""
	}
	for i, e := range m.list {
		if e.Path == path {
			e.PlayCount++
			e.PlayedAt = now
			if title != "" {
				e.Title = title
			}
			if artist != "" {
				e.Artist = artist
			}
			if coverID != "" {
				e.CoverID = coverID
				e.CoverUrl = ""
			} else if cleanCoverUrl != "" {
				e.CoverUrl = cleanCoverUrl
			}
			m.list = append(m.list[:i], m.list[i+1:]...)
			m.list = append([]Entry{e}, m.list...)
			if len(m.list) > 100 {
				m.list = m.list[:100]
			}
			_ = m.save()
			out := make([]Entry, len(m.list))
			copy(out, m.list)
			return out
		}
	}
	e := Entry{Path: path, Title: title, Artist: artist, CoverUrl: cleanCoverUrl, CoverID: coverID, PlayedAt: now, PlayCount: 1}
	m.list = append([]Entry{e}, m.list...)
	if len(m.list) > 100 {
		m.list = m.list[:100]
	}
	_ = m.save()
	out := make([]Entry, len(m.list))
	copy(out, m.list)
	return out
}

func (m *Manager) List() []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Entry, len(m.list))
	copy(out, m.list)
	return out
}

func (m *Manager) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.list = []Entry{}
	return m.save()
}

func (m *Manager) SortedByPlays(limit int) []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cp := make([]Entry, len(m.list))
	copy(cp, m.list)
	sort.Slice(cp, func(i, j int) bool { return cp[i].PlayCount > cp[j].PlayCount })
	if limit > 0 && len(cp) > limit {
		cp = cp[:limit]
	}
	return cp
}
