package Handlers

import (
	"encoding/base64"
	"fmt"
	"strings"

	music "StepLauncher/internal/Music"
)

type MusicScanProgress = music.ScanProgress

// Tipos compatibles con bindings existentes (frontend)

type MusicMetadataEntry struct {
	Title       string  `json:"title"`
	Artist      string  `json:"artist"`
	Album       string  `json:"album,omitempty"`
	Year        int     `json:"year,omitempty"`
	Duration    float64 `json:"duration"`
	TrackNumber int     `json:"trackNumber,omitempty"`
	Genre       string  `json:"genre,omitempty"`
	HasCover    bool    `json:"hasCover"`
	CoverMime   string  `json:"coverMime,omitempty"`
	CachedAt    string  `json:"cachedAt"`
	FileModTime string  `json:"fileModTime,omitempty"`
	FileSize    int64   `json:"fileSize,omitempty"`
}

type CoverCacheInfo struct {
	Total   int   `json:"total"`
	Valid   int   `json:"valid"`
	Expired int   `json:"expired"`
	Size    int64 `json:"size"`
}

func toHandlerMeta(m music.MusicMeta) MusicMetadataEntry {
	return MusicMetadataEntry{
		Title:       m.Title,
		Artist:      m.Artist,
		Album:       m.Album,
		Year:        m.Year,
		Duration:    m.Duration,
		TrackNumber: m.TrackNumber,
		Genre:       m.Genre,
		HasCover:    m.HasCover,
		CoverMime:   m.CoverMime,
		CachedAt:    m.CachedAt,
		FileModTime: m.FileModTime,
		FileSize:    m.FileSize,
	}
}

func toMusicMeta(h MusicMetadataEntry) music.MusicMeta {
	return music.MusicMeta{
		Title:       h.Title,
		Artist:      h.Artist,
		Album:       h.Album,
		Year:        h.Year,
		Duration:    h.Duration,
		TrackNumber: h.TrackNumber,
		Genre:       h.Genre,
		HasCover:    h.HasCover,
		CoverMime:   h.CoverMime,
		CachedAt:    h.CachedAt,
		FileModTime: h.FileModTime,
		FileSize:    h.FileSize,
	}
}

// GetCachedCover devuelve data URI desde cache/music/covers
func (a *App) GetCachedCover(trackPath string) (string, error) {
	if a.musicCache == nil {
		return "", fmt.Errorf("cache de música no disponible")
	}
	trackPath = strings.TrimSpace(trackPath)
	if trackPath == "" {
		return "", fmt.Errorf("ruta vacía")
	}
	if uri, ok := a.musicCache.GetCoverDataURI(trackPath); ok {
		return uri, nil
	}
	return "", nil
}

// SaveCachedCover guarda carátula en shard comprimido 5-10 por archivo - eficiente binario
func (a *App) SaveCachedCover(trackPath, b64Data, mime string) error {
	if a.musicCache == nil {
		return fmt.Errorf("cache de música no disponible")
	}
	b64Data = strings.TrimSpace(b64Data)
	if b64Data == "" {
		return fmt.Errorf("datos incompletos")
	}
	data, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		data, err = base64.RawStdEncoding.DecodeString(b64Data)
		if err != nil {
			return fmt.Errorf("base64 inválido: %v", err)
		}
	}
	return a.musicCache.SaveCover(trackPath, data, mime)
}

// SaveMusicMetadata guarda metadata ligera (audiometa ya lo hace, pero se mantiene para compatibilidad)
func (a *App) SaveMusicMetadata(trackPath, title, artist string, duration float64, coverData, mime string) error {
	if a.musicCache == nil {
		return fmt.Errorf("cache de música no disponible")
	}
	trackPath = strings.TrimSpace(trackPath)
	if trackPath == "" {
		return fmt.Errorf("ruta vacía")
	}
	if coverData != "" {
		data, err := base64.StdEncoding.DecodeString(coverData)
		if err != nil {
			data, err = base64.RawStdEncoding.DecodeString(coverData)
			if err != nil {
				return fmt.Errorf("cover base64 inválido")
			}
		}
		_ = a.musicCache.SaveCover(trackPath, data, mime)
	}
	// guardar metadata sin coverData
	hasCover := coverData != ""
	if mime == "" && hasCover {
		mime = "image/jpeg"
	}
	// intentar preservar album/year etc si ya existen
	var album string
	var year int
	var genre string
	var trackNum int
	if existing, ok := a.musicCache.GetMetadata(trackPath); ok {
		album = existing.Album
		year = existing.Year
		genre = existing.Genre
		trackNum = existing.TrackNumber
		if existing.HasCover {
			hasCover = true
		}
		if existing.CoverMime != "" {
			mime = existing.CoverMime
		}
	}
	return a.musicCache.SaveMetadata(trackPath, title, artist, album, year, duration, trackNum, genre, hasCover, mime)
}

func (a *App) GetMusicMetadata(trackPath string) (MusicMetadataEntry, bool) {
	if a.musicCache == nil {
		return MusicMetadataEntry{}, false
	}
	m, ok := a.musicCache.GetMetadata(trackPath)
	if !ok {
		return MusicMetadataEntry{}, false
	}
	return toHandlerMeta(m), true
}

func (a *App) GetAllMusicMetadata() map[string]MusicMetadataEntry {
	if a.musicCache == nil {
		return map[string]MusicMetadataEntry{}
	}
	all := a.musicCache.GetAllMetadata()
	out := make(map[string]MusicMetadataEntry, len(all))
	for k, v := range all {
		out[k] = toHandlerMeta(v)
	}
	return out
}

func (a *App) RefreshCachedCover(trackPath string) error {
	if a.musicCache == nil {
		return fmt.Errorf("cache de música no disponible")
	}
	return a.musicCache.RefreshCover(trackPath)
}

func (a *App) GetCoverCacheInfo() CoverCacheInfo {
	if a.musicCache == nil {
		return CoverCacheInfo{}
	}
	m := a.musicCache.GetCoverCacheInfo()
	return CoverCacheInfo{Total: m.Total, Valid: m.Valid, Expired: m.Expired, Size: m.Size}
}

func (a *App) GetCoverCacheStatus() CoverCacheInfo {
	return a.GetCoverCacheInfo()
}

func (a *App) ClearCoverCache() int {
	if a.musicCache == nil {
		return 0
	}
	return a.musicCache.Clear()
}

func (a *App) GetMusicLibraryStats() music.MusicLibraryStats {
	if a.musicCache == nil {
		return music.MusicLibraryStats{}
	}
	folders := a.GetMusicFolders()
	// Compatibilidad legacy — NO es carpeta por defecto
	if len(folders) == 0 && a.config != nil {
		if mf := a.config.Get().MusicPanel.MusicFolder; mf != "" {
			folders = []string{mf}
		}
	}
	// PROHIBIDO: nunca inyectar carpeta por defecto — si no hay carpetas, stats con 0 carpetas
	return a.musicCache.GetLibraryStats(folders)
}

func (a *App) RebuildMusicIndex() error {
	if a.musicCache == nil {
		return fmt.Errorf("cache de música no disponible")
	}
	folders := a.GetMusicFolders()
	// Compatibilidad legacy — NO es carpeta por defecto
	if len(folders) == 0 && a.config != nil {
		if mf := a.config.Get().MusicPanel.MusicFolder; mf != "" {
			folders = []string{mf}
		}
	}
	// PROHIBIDO: nunca usar carpeta por defecto
	return a.musicCache.RebuildIndex(folders)
}

func (a *App) ClearMusicCache() int {
	if a.musicCache == nil {
		return 0
	}
	return a.musicCache.Clear()
}

func (a *App) InitializeMusicLibrary() ([]string, error) {
	if a.musicCache == nil {
		return nil, fmt.Errorf("cache de música no disponible")
	}
	folders := a.GetMusicFolders()
	// Compatibilidad legacy — NO es carpeta por defecto
	if len(folders) == 0 && a.config != nil {
		if mf := a.config.Get().MusicPanel.MusicFolder; mf != "" {
			folders = []string{mf}
		}
	}
	// PROHIBIDO: nunca crear carpeta por defecto — error si no hay carpetas explícitas
	if len(folders) == 0 {
		return nil, fmt.Errorf("no hay carpetas configuradas")
	}
	return a.musicCache.ScanFolders(folders)
}

func (a *App) coverCacheCount() int {
	return a.GetCoverCacheInfo().Total
}

func (a *App) coverCacheSizeBytes() int64 {
	return a.GetCoverCacheInfo().Size
}

// GetMusicScanProgress expone progreso exacto para "Importando carpeta X / Y"
func (a *App) GetMusicScanProgress() music.ScanProgress {
	if a.musicCache == nil {
		return music.ScanProgress{}
	}
	return a.musicCache.GetProgress()
}

// GetMusicTracksPaged paginación con dos carátulas: thumb 128 (lista/cola) y raw (detalle)
// PROHIBIDO: nunca inyectar carpeta por defecto — solo usa carpetas explícitas del usuario
func (a *App) GetMusicTracksPaged(offset, limit int, coverVariant, query string) (music.MusicPageDTO, error) {
	if a.musicCache == nil {
		return music.MusicPageDTO{Tracks: []music.MusicTrackDTO{}, Total: 0, Offset: offset, Limit: limit}, nil
	}
	folders := a.GetMusicFolders()
	// Compatibilidad legacy — NO es carpeta por defecto
	if len(folders) == 0 && a.config != nil {
		if mf := a.config.Get().MusicPanel.MusicFolder; mf != "" {
			folders = []string{mf}
		}
	}
	withThumb := false
	withRaw := false
	cv := strings.ToLower(strings.TrimSpace(coverVariant))
	switch cv {
	case "thumb", "thumb128", "128", "small", "compressed":
		withThumb = true
	case "raw", "original", "hd":
		withRaw = true
	case "both", "all", "thumb+raw":
		withThumb = true
		withRaw = true
	case "none", "":
		// sin covers
	default:
		withThumb = true
	}
	return a.musicCache.GetTracksPage(folders, offset, limit, query, withThumb, withRaw)
}

// GetMusicTracksBatch retorna DTOs para paths específicos con cover variant
func (a *App) GetMusicTracksBatch(paths []string, coverVariant string) ([]music.MusicTrackDTO, error) {
	if a.musicCache == nil {
		return []music.MusicTrackDTO{}, nil
	}
	withThumb := false
	withRaw := false
	cv := strings.ToLower(strings.TrimSpace(coverVariant))
	switch cv {
	case "thumb", "thumb128", "128", "small":
		withThumb = true
	case "raw", "original":
		withRaw = true
	case "both", "all":
		withThumb = true
		withRaw = true
	case "none", "":
	default:
		withThumb = true
	}
	out := make([]music.MusicTrackDTO, 0, len(paths))
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		dto, err := a.musicCache.BuildTrackDTO(p, withThumb, withRaw)
		if err != nil {
			dto = music.MusicTrackDTO{Path: p, FileName: strings.TrimSpace(p), Title: p, Artist: "Desconocido"}
		}
		out = append(out, dto)
	}
	return out, nil
}

// GetMusicCoverBase64 retorna carátula en base64 data URI: variant "raw" o "thumb"
func (a *App) GetMusicCoverBase64(trackPath, variant string) (string, error) {
	if a.musicCache == nil {
		return "", fmt.Errorf("cache no disponible")
	}
	return a.musicCache.GetCoverBase64(trackPath, variant)
}
