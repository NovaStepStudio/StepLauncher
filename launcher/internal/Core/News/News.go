package News

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	engineconfig "StepLauncher/internal/Handlers/Engine/engineconfig"
)

const (
	BaseURL       = "https://wpnvconaefhdmvgvqbsv.supabase.co/storage/v1/object/public/news"
	IndexPath     = "Releases/index.json"
	newsMaxBody   = 512 * 1024
	changelogMax  = 4 * 1024 * 1024
	requestTimeout = 20 * time.Second
)

var httpClient = &http.Client{Timeout: 25 * time.Second}

type EventHandler func(eventType string, data []byte)

type ReleaseEntry struct {
	Version string `json:"version"`
	Path    string `json:"path"`
}

type Index struct {
	Latest  string          `json:"latest"`
	Content []ReleaseEntry `json:"content"`
}

type ReleaseInfo struct {
	Title     string `json:"title"`
	Type      string `json:"type"`
	Body      string `json:"body"`
	Date      string `json:"date"`
	Changelog string `json:"changelog"`
}

type Manager struct {
	mu       sync.Mutex
	rootDir  string
	cacheDir string
	eventCb  EventHandler
	logFn    func(format string, args ...interface{})
}

func NewManager(rootDir string, logFn func(format string, args ...interface{})) *Manager {
	return &Manager{
		rootDir:  rootDir,
		cacheDir: filepath.Join(rootDir, "cache", "news"),
		logFn:    logFn,
	}
}

func (m *Manager) SetEventCallback(cb EventHandler) {
	m.mu.Lock()
	m.eventCb = cb
	m.mu.Unlock()
}

func (m *Manager) logf(format string, args ...interface{}) {
	if m.logFn != nil {
		m.logFn("[News] "+format, args...)
	}
}

func (m *Manager) emit(eventType string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	m.mu.Lock()
	cb := m.eventCb
	m.mu.Unlock()
	if cb != nil {
		cb(eventType, data)
	}
}

func (m *Manager) fetch(url string, maxBody int64) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "StepLauncher/"+engineconfig.AppVersion)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("respuesta %s", resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxBody))
}

func (m *Manager) cacheFile(key string) string {
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(m.cacheDir, hex.EncodeToString(sum[:]))
}

func (m *Manager) writeCache(key string, data []byte) error {
	if err := os.MkdirAll(m.cacheDir, 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(m.cacheDir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, m.cacheFile(key)); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}

func (m *Manager) readCache(key string) ([]byte, error) {
	path := m.cacheFile(key)
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("no es un archivo")
	}
	return os.ReadFile(path)
}

func (m *Manager) indexURL() string {
	return BaseURL + "/" + IndexPath
}

func (m *Manager) releaseNewsURL(relPath string) string {
	clean := strings.ReplaceAll(relPath, "\\", "/")
	clean = strings.TrimPrefix(clean, "./")
	if strings.HasPrefix(clean, "/") {
		return BaseURL + clean
	}
	// El path de cada entrada es relativo al directorio del índice (…/news/Releases/)
	idx := m.indexURL()
	base := idx[:strings.LastIndex(idx, "/")+1]
	return base + clean
}

func (m *Manager) changelogCandidates(version, changelog string) []string {
	clean := strings.ReplaceAll(changelog, "\\", "/")
	clean = strings.TrimPrefix(clean, "./")
	if clean == "" {
		return nil
	}
	if strings.HasPrefix(clean, "http://") || strings.HasPrefix(clean, "https://") {
		return []string{clean}
	}
	var out []string
	if strings.HasPrefix(clean, "Changelogs/") {
		out = append(out, BaseURL+"/"+strings.TrimPrefix(clean, "Changelogs/"))
	}
	out = append(out, BaseURL+"/"+clean)
	out = append(out, BaseURL+"/Releases/StepLauncher-"+strings.TrimSpace(version)+"/StepLauncher-Release-"+strings.TrimSpace(version)+".md")
	return out
}


func (m *Manager) loadIndex() (*Index, error) {
	idxURL := m.indexURL()
	raw, err := m.readCache(idxURL)
	fromCache := err == nil
	if err != nil {
		raw, err = m.fetch(idxURL, newsMaxBody)
		if err != nil {
			return nil, err
		}
	}
	var idx Index
	if err := json.Unmarshal(raw, &idx); err != nil {
		return nil, fmt.Errorf("índice inválido: %w", err)
	}
	if !fromCache {
		if err := m.writeCache(idxURL, raw); err != nil {
			m.logf("WARN: no se pudo cachear el índice: %v", err)
		}
	}
	return &idx, nil
}

func (m *Manager) findEntry(version string) (*ReleaseEntry, bool) {
	idx, err := m.loadIndex()
	if err != nil {
		return nil, false
	}
	for i := range idx.Content {
		if idx.Content[i].Version == version {
			e := idx.Content[i]
			return &e, true
		}
	}
	return nil, false
}

func (m *Manager) loadReleaseInfo(entry ReleaseEntry) (*ReleaseInfo, bool, bool) {
	url := m.releaseNewsURL(entry.Path)
	raw, err := m.readCache(url)
	fromCache := err == nil
	if err != nil {
		raw, err = m.fetch(url, newsMaxBody)
		if err != nil {
			return nil, false, false
		}
	}
	var info ReleaseInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return nil, false, false
	}
	if !fromCache {
		if err := m.writeCache(url, raw); err != nil {
			m.logf("WARN: no se pudo cachear la noticia %s: %v", entry.Version, err)
		}
	}
	return &info, fromCache, true
}

func (m *Manager) RefreshIndex() {
	go func() {
		payload := map[string]interface{}{
			"ok":        false,
			"fromCache": false,
			"error":     "",
			"latest":    "",
			"content":   []ReleaseEntry{},
		}
		idx, err := m.loadIndex()
		if err != nil {
			payload["error"] = "no se pudo obtener el índice de noticias: " + err.Error()
			m.logf("index: %v", err)
			m.emit("news_index", payload)
			return
		}
		payload["ok"] = true
		payload["latest"] = idx.Latest
		payload["content"] = idx.Content
		m.emit("news_index", payload)
	}()
}

func (m *Manager) LoadRelease(version string) {
	go func() {
		entry, ok := m.findEntry(version)
		if !ok {
			m.emit("news_release", map[string]interface{}{
				"ok": false, "version": version, "error": "la release " + version + " no existe en el índice",
			})
			return
		}
		info, fromCache, ok := m.loadReleaseInfo(*entry)
		if !ok {
			m.emit("news_release", map[string]interface{}{
				"ok": false, "version": version, "error": "no se pudo cargar la noticia de " + version,
			})
			return
		}
		m.emit("news_release", map[string]interface{}{
			"ok":        true,
			"fromCache": fromCache,
			"version":   version,
			"newsPath":  entry.Path,
			"title":     info.Title,
			"type":      info.Type,
			"body":      info.Body,
			"date":      info.Date,
			"changelog": info.Changelog,
		})
	}()
}

func (m *Manager) LoadChangelog(version string) {
	go func() {
		entry, ok := m.findEntry(version)
		if !ok {
			m.emit("news_changelog", map[string]interface{}{
				"ok": false, "version": version, "error": "la release " + version + " no existe en el índice",
			})
			return
		}
		info, _, ok := m.loadReleaseInfo(*entry)
		if !ok {
			m.emit("news_changelog", map[string]interface{}{
				"ok": false, "version": version, "error": "no se pudo cargar la noticia de " + version,
			})
			return
		}

		var markdown []byte
		var chosen string
		fromCache := false
		for _, cand := range m.changelogCandidates(entry.Version, info.Changelog) {
			if raw, err := m.readCache(cand); err == nil {
				markdown, chosen, fromCache = raw, cand, true
				break
			}
			if raw, err := m.fetch(cand, changelogMax); err == nil {
				markdown, chosen, fromCache = raw, cand, false
				if err := m.writeCache(cand, raw); err != nil {
					m.logf("WARN: no se pudo cachear el changelog %s: %v", version, err)
				}
				break
			}
			m.logf("changelog %s: el candidato %s no respondió", version, cand)
		}
		if markdown == nil {
			m.emit("news_changelog", map[string]interface{}{
				"ok": false, "version": version, "error": "no se pudo descargar el changelog de " + version,
			})
			return
		}
		m.emit("news_changelog", map[string]interface{}{
			"ok":        true,
			"fromCache": fromCache,
			"version":   version,
			"url":       chosen,
			"markdown":  string(markdown),
		})
	}()
}

// LoadMarkdown carga un MD arbitrario del mismo bucket (p. ej. un enlace
// interno a un registro de Error/Change de la auditoría).
func (m *Manager) LoadMarkdown(url string) {
	go func() {
		if !strings.HasPrefix(url, BaseURL) {
			m.emit("news_markdown", map[string]interface{}{
				"ok": false, "url": url, "error": "url no permitida",
			})
			return
		}
		raw, err := m.readCache(url)
		fromCache := err == nil
		if err != nil {
			raw, err = m.fetch(url, changelogMax)
			if err != nil {
				m.emit("news_markdown", map[string]interface{}{
					"ok": false, "url": url, "error": "no se pudo descargar el documento: " + err.Error(),
				})
				return
			}
			if err := m.writeCache(url, raw); err != nil {
				m.logf("WARN: no se pudo cachear %s: %v", url, err)
			}
		}
		m.emit("news_markdown", map[string]interface{}{
			"ok":        true,
			"fromCache": fromCache,
			"url":       url,
			"markdown":  string(raw),
		})
	}()
}