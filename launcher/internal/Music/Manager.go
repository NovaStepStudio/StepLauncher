package music

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/simonhull/audiometa"
)

type Manager struct {
	mu           sync.RWMutex
	rootDir      string
	cacheDir     string
	metaDir      string
	coversDir    string
	progress     ScanProgress
	progMu       sync.Mutex
	metaMem      map[string]MusicMeta
	memMu        sync.RWMutex
	loaded       bool
	index        *MusicIndex
	scanMu       sync.Mutex
	coverMu      sync.Mutex
	scanCancel   context.CancelFunc
	scanCancelMu sync.Mutex
}

func NewManager(rootDir string) *Manager {
	rootDir = filepath.Clean(rootDir)
	cacheDir := filepath.Join(rootDir, "cache", "music")
	metaDir := filepath.Join(cacheDir, "meta")
	coversDir := filepath.Join(cacheDir, "covers")
	idx := NewMusicIndex(cacheDir)
	_ = idx.Load()
	// Migrar shards viejos si índice vacío
	if idx.Count() == 0 {
		_ = idx.MigrateFromShards(metaDir)
	}
	return &Manager{
		rootDir:   rootDir,
		cacheDir:  cacheDir,
		metaDir:   metaDir,
		coversDir: coversDir,
		metaMem:   make(map[string]MusicMeta),
		index:     idx,
	}
}

func (m *Manager) CacheDir() string { return m.cacheDir }

func (m *Manager) GetProgress() ScanProgress {
	m.progMu.Lock()
	defer m.progMu.Unlock()
	return m.progress
}

func (m *Manager) setProgress(p ScanProgress) {
	m.progMu.Lock()
	m.progress = p
	m.progMu.Unlock()
}

// CancelScan cancela el escaneo en curso si existe. Retorna true si había escaneo.
func (m *Manager) CancelScan() bool {
	m.scanCancelMu.Lock()
	defer m.scanCancelMu.Unlock()
	if m.scanCancel != nil {
		m.scanCancel()
		return true
	}
	return false
}

// IsScanning indica si hay escaneo activo según progress.
func (m *Manager) IsScanning() bool {
	m.progMu.Lock()
	defer m.progMu.Unlock()
	return m.progress.Scanning
}

// GetMetadata lee desde índice si está cacheado y válido, sino parsea al vuelo.
func cleanMetaTag(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\uFEFF' || r == '\uFFFD' {
			continue
		}
		if r < 0x20 || (r >= 0x7F && r < 0xA0) {
			continue
		}
		if r == '\u200B' || r == '\u200C' || r == '\u200D' || r == '\u2060' || r == '\u00AD' || r == '\u034F' {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

func (m *Manager) GetMetadata(trackPath string) (MusicMeta, bool) {
	trackPath = strings.TrimSpace(trackPath)
	if trackPath == "" {
		return MusicMeta{}, false
	}
	fi, err := os.Stat(trackPath)
	if err != nil {
		return MusicMeta{}, false
	}
	size := fi.Size()
	modTime := fi.ModTime().Unix()
	// Intentar índice primero (si no necesita update)
	if m.index != nil && !m.index.NeedsUpdate(trackPath, size, modTime) {
		if entry, ok := m.index.Get(trackPath); ok {
			meta := MusicMeta{
				Title:       cleanMetaTag(entry.Title),
				Artist:      cleanMetaTag(entry.Artist),
				Album:       cleanMetaTag(entry.Album),
				Year:        entry.Year,
				Duration:    entry.Duration,
				TrackNumber: entry.TrackNumber,
				Genre:       cleanMetaTag(entry.Genre),
				HasCover:    entry.HasCover,
				CoverID:     entry.CoverID,
				CoverMime:   entry.CoverMime,
				CachedAt:    time.Unix(entry.CachedAt, 0).Format(time.RFC3339),
				FileModTime: time.Unix(entry.ModTime, 0).Format(time.RFC3339),
				FileSize:    entry.Size,
			}
			if meta.Title == "" {
				meta.Title = cleanMetaTag(strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath)))
			}
			if meta.Artist == "" {
				meta.Artist = "Desconocido"
			}
			return meta, true
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	r := extractMetaFast(ctx, trackPath)
	hasCover := false
	coverMime := ""
	func() {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 1500*time.Millisecond)
		defer cancel2()
		f, err := audiometa.OpenContext(ctx2, trackPath)
		if err != nil {
			return
		}
		defer f.Close()
		if arts, err := f.ExtractArtworkContext(ctx2); err == nil && len(arts) > 0 && len(arts[0].Data) > 0 {
			hasCover = true
			coverMime = strings.TrimSpace(arts[0].MIMEType)
			if coverMime == "" {
				coverMime = "image/jpeg"
			}
		}
	}()
	meta := MusicMeta{
		Title:       cleanMetaTag(r.title),
		Artist:      cleanMetaTag(r.artist),
		Album:       cleanMetaTag(r.album),
		Year:        r.year,
		Duration:    r.duration,
		TrackNumber: r.trackNum,
		Genre:       cleanMetaTag(r.genre),
		HasCover:    hasCover,
		CoverMime:   coverMime,
		CoverID:     "",
		CachedAt:    time.Now().Format(time.RFC3339),
		FileModTime: fi.ModTime().Format(time.RFC3339),
		FileSize:    size,
	}
	if meta.Title == "" {
		meta.Title = cleanMetaTag(strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath)))
	}
	if meta.Artist == "" {
		meta.Artist = "Desconocido"
	}
	// Actualizar índice si hay coverID (se generará al guardar cover)
	if m.index != nil && hasCover {
		// No guardamos CoverID aún, se asignará al guardar cover
	}
	return meta, true
}

func (m *Manager) GetAllMetadata() map[string]MusicMeta {
	if m.index == nil {
		return make(map[string]MusicMeta)
	}
	list := m.index.List()
	out := make(map[string]MusicMeta, len(list))
	for _, e := range list {
		out[e.Path] = MusicMeta{
			Title:       e.Title,
			Artist:      e.Artist,
			Album:       e.Album,
			Year:        e.Year,
			Duration:    e.Duration,
			TrackNumber: e.TrackNumber,
			Genre:       e.Genre,
			HasCover:    e.HasCover,
			CoverID:     e.CoverID,
			CoverMime:   e.CoverMime,
			CachedAt:    time.Unix(e.CachedAt, 0).Format(time.RFC3339),
			FileModTime: time.Unix(e.ModTime, 0).Format(time.RFC3339),
			FileSize:    e.Size,
		}
	}
	return out
}

func (m *Manager) SaveMetadata(trackPath, title, artist, album string, year int, duration float64, trackNumber int, genre string, hasCover bool, coverMime string) error {
	if m.index == nil {
		return nil
	}
	trackPath = strings.TrimSpace(trackPath)
	if trackPath == "" {
		return fmt.Errorf("ruta vacía")
	}
	fi, err := os.Stat(trackPath)
	var size int64
	var modTime int64
	if err == nil {
		size = fi.Size()
		modTime = fi.ModTime().Unix()
	} else {
		size = 0
		modTime = time.Now().Unix()
	}
	entry := &MusicIndexEntry{
		ID:          GenerateID(trackPath, size, modTime),
		Path:        trackPath,
		Size:        size,
		ModTime:     modTime,
		Title:       title,
		Artist:      artist,
		Album:       album,
		Genre:       genre,
		Year:        year,
		TrackNumber: trackNumber,
		Duration:    duration,
		HasCover:    hasCover,
		CoverMime:   coverMime,
		CachedAt:    time.Now().Unix(),
	}
	if existing, ok := m.index.Get(trackPath); ok {
		entry.CoverID = existing.CoverID
		if entry.CoverID == "" {
			entry.HasCover = hasCover
		}
	}
	m.index.Set(entry)
	return m.index.Save()
}

func (m *Manager) SaveCover(trackPath string, data []byte, mime string) error {
	if m.index == nil {
		return nil
	}
	trackPath = strings.TrimSpace(trackPath)
	if trackPath == "" || len(data) == 0 {
		return fmt.Errorf("datos incompletos")
	}
	hash := coverHash(data)
	// Guardar cover en shards (reusa lógica Cache.go)
	entry := CoverEntry{TrackPath: trackPath, Mime: mime, Data: data, CachedAt: time.Now().Format(time.RFC3339)}
	if err := m.saveCover(trackPath, entry); err != nil {
		return err
	}
	// Actualizar índice con CoverID
	fi, _ := os.Stat(trackPath)
	var size int64
	var modTime int64
	if fi != nil {
		size = fi.Size()
		modTime = fi.ModTime().Unix()
	}
	if existing, ok := m.index.Get(trackPath); ok {
		existing.HasCover = true
		existing.CoverID = hash
		existing.CoverMime = mime
		existing.CachedAt = time.Now().Unix()
		m.index.Set(existing)
	} else {
		// Crear entrada mínima si no existe
		meta, _ := m.GetMetadata(trackPath)
		e := &MusicIndexEntry{
			ID:        GenerateID(trackPath, size, modTime),
			Path:      trackPath,
			Size:      size,
			ModTime:   modTime,
			Title:     meta.Title,
			Artist:    meta.Artist,
			Album:     meta.Album,
			Genre:     meta.Genre,
			Year:      meta.Year,
			Duration:  meta.Duration,
			HasCover:  true,
			CoverID:   hash,
			CoverMime: mime,
			CachedAt:  time.Now().Unix(),
		}
		m.index.Set(e)
	}
	return m.index.Save()
}

func (m *Manager) GetCoverDataURI(trackPath string) (string, bool) {
	// Retorna thumb 128 comprimida para listas/cola (10KB vs 500KB raw) — base64 lista para usar
	thumb, err := m.GetCoverBase64(trackPath, "thumb")
	if err != nil {
		return "", false
	}
	return thumb, true
}

func (m *Manager) RefreshCover(trackPath string) error { return nil }

func (m *Manager) GetCoverCacheStatus() CoverCacheInfo { return CoverCacheInfo{} }

func (m *Manager) GetCoverCacheInfo() CoverCacheInfo { return CoverCacheInfo{} }

func (m *Manager) Clear() int {
	if m.index == nil {
		return 0
	}
	count := m.index.Count()
	m.index.mu.Lock()
	m.index.entries = make(map[string]*MusicIndexEntry)
	m.index.mu.Unlock()
	_ = m.index.Save()
	_ = os.RemoveAll(m.coversDir)
	_ = os.MkdirAll(m.coversDir, 0755)
	return count
}

func (m *Manager) GetLibraryStats(folders []string) MusicLibraryStats {
	if m.index == nil {
		return MusicLibraryStats{}
	}
	// PROHIBIDO: sin carpetas explícitas => estadísticas vacías, nunca mostrar carpeta por defecto
	if len(folders) == 0 {
		return MusicLibraryStats{}
	}
	return m.index.Stats(len(folders))
}

func (m *Manager) RebuildIndex(folders []string) error {
	if m.index == nil {
		return fmt.Errorf("index no disponible")
	}
	m.index.mu.Lock()
	m.index.entries = make(map[string]*MusicIndexEntry)
	m.index.mu.Unlock()
	_ = m.index.Save()
	_, err := m.ScanFolders(folders)
	return err
}

// GetCoverBase64 retorna dos variantes: thumb 128 (10KB, para listas/cola) y raw (original, para NowPlaying)
// Usa cache zip si existe (instantáneo), si no extrae del archivo
func (m *Manager) GetCoverBase64(trackPath, variant string) (string, error) {
	trackPath = strings.TrimSpace(trackPath)
	if trackPath == "" {
		return "", fmt.Errorf("ruta vacía")
	}
	variant = strings.ToLower(strings.TrimSpace(variant))
	if variant == "" {
		variant = "thumb"
	}
	// 1) Intentar cache zip via índice (sin I/O de archivo, mucho más rápido para Ahora Suena)
	if m.index != nil {
		if entry, ok := m.index.Get(trackPath); ok && entry.HasCover && entry.CoverID != "" {
			if data, _, err := m.readCoverFromZipByHash(entry.CoverID); err == nil && len(data) > 0 {
				mime := entry.CoverMime
				if mime == "" {
					mime = "image/jpeg"
				}
				if variant == "raw" || variant == "original" {
					b64 := base64.StdEncoding.EncodeToString(data)
					return fmt.Sprintf("data:%s;base64,%s", mime, b64), nil
				}
				thumb, thumbMime, err := makeThumb128(data, mime)
				if err == nil {
					b64 := base64.StdEncoding.EncodeToString(thumb)
					thumb = nil
					return fmt.Sprintf("data:%s;base64,%s", thumbMime, b64), nil
				}
				b64 := base64.StdEncoding.EncodeToString(data)
				return fmt.Sprintf("data:%s;base64,%s", mime, b64), nil
			}
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	f, err := audiometa.OpenContext(ctx, trackPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	arts, err := f.ExtractArtworkContext(ctx)
	if err != nil || len(arts) == 0 {
		return "", fmt.Errorf("sin carátula")
	}
	art := arts[0]
	if len(art.Data) == 0 {
		return "", fmt.Errorf("sin datos de carátula")
	}
	mime := strings.TrimSpace(art.MIMEType)
	if mime == "" {
		mime = "image/jpeg"
	}
	if variant == "raw" || variant == "original" {
		b64 := base64.StdEncoding.EncodeToString(art.Data)
		return fmt.Sprintf("data:%s;base64,%s", mime, b64), nil
	}
	// thumb 128 comprimida para listas (evita 500KB raw por cover)
	thumb, thumbMime, err := makeThumb128(art.Data, mime)
	if err != nil {
		// fallback a raw si falla thumb
		b64 := base64.StdEncoding.EncodeToString(art.Data)
		return fmt.Sprintf("data:%s;base64,%s", mime, b64), nil
	}
	b64 := base64.StdEncoding.EncodeToString(thumb)
	// liberar buffer thumb para GC
	thumb = nil
	return fmt.Sprintf("data:%s;base64,%s", thumbMime, b64), nil
}

// BuildTrackDTO construye DTO ligero — sin decodificar imagen para thumb (evita 36MB por cover)
func (m *Manager) BuildTrackDTO(trackPath string, withThumb, withRaw bool) (MusicTrackDTO, error) {
	trackPath = strings.TrimSpace(trackPath)
	if trackPath == "" {
		return MusicTrackDTO{}, fmt.Errorf("ruta vacía")
	}
	fi, err := os.Stat(trackPath)
	if err != nil {
		return MusicTrackDTO{}, err
	}
	fileName := filepath.Base(trackPath)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	r := extractMetaFast(ctx, trackPath)
	dto := MusicTrackDTO{
		Path:     trackPath,
		FileName: fileName,
		Title:    cleanMetaTag(r.title),
		Artist:   cleanMetaTag(r.artist),
		Album:    cleanMetaTag(r.album),
		Year:     r.year,
		Duration: r.duration,
		Genre:    cleanMetaTag(r.genre),
		HasCover: false,
	}
	if dto.Title == "" {
		dto.Title = cleanMetaTag(strings.TrimSuffix(fileName, filepath.Ext(fileName)))
	}
	if dto.Artist == "" {
		dto.Artist = "Desconocido"
	}
	if withThumb || withRaw {
		f, err := audiometa.OpenContext(ctx, trackPath)
		if err == nil {
			defer f.Close()
			if arts, err := f.ExtractArtworkContext(ctx); err == nil && len(arts) > 0 && len(arts[0].Data) > 0 {
				art := arts[0]
				mime := strings.TrimSpace(art.MIMEType)
				if mime == "" {
					mime = "image/jpeg"
				}
				dto.HasCover = true
				dto.CoverMime = mime
				if withThumb {
					thumb, thumbMime, err := makeThumb128(art.Data, mime)
					if err == nil {
						b64t := base64.StdEncoding.EncodeToString(thumb)
						dto.CoverThumb = fmt.Sprintf("data:%s;base64,%s", thumbMime, b64t)
					} else {
						b64 := base64.StdEncoding.EncodeToString(art.Data)
						dto.CoverThumb = fmt.Sprintf("data:%s;base64,%s", mime, b64)
					}
				}
				if withRaw {
					b64r := base64.StdEncoding.EncodeToString(art.Data)
					dto.CoverRaw = fmt.Sprintf("data:%s;base64,%s", mime, b64r)
				}
			}
		}
	}
	_ = fi
	return dto, nil
}

// GetTracksPage pagina via índice (sin WalkDir por página). Go es autoridad.
// PROHIBIDO: sin carpetas explícitas no hay pistas — nunca usar carpeta por defecto.
func (m *Manager) GetTracksPage(folders []string, offset, limit int, query string, withThumb, withRaw bool) (MusicPageDTO, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 5000 {
		limit = 5000
	}
	if offset < 0 {
		offset = 0
	}
	// Normalizar folders para filtrar índice por carpeta
	normalizedFolders := make([]string, 0, len(folders))
	for _, f := range folders {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if !filepath.IsAbs(f) {
			f = filepath.Join(m.rootDir, f)
		}
		normalizedFolders = append(normalizedFolders, filepath.Clean(f))
	}
	// Sin carpetas configuradas => vacío inmediato, no listar índice completo
	if len(normalizedFolders) == 0 {
		return MusicPageDTO{Tracks: []MusicTrackDTO{}, Total: 0, Offset: offset, Limit: limit}, nil
	}
	// Si índice vacío, intentar cargar desde disco o fallback a WalkDir ligero
	var allEntries []*MusicIndexEntry
	if m.index != nil && m.index.Count() > 0 {
		if strings.TrimSpace(query) != "" {
			allEntries = m.index.Search(query)
		} else {
			allEntries = m.index.List()
		}
		// Filtrar por folders si se especifican
		if len(normalizedFolders) > 0 {
			filtered := make([]*MusicIndexEntry, 0, len(allEntries))
			for _, e := range allEntries {
				for _, f := range normalizedFolders {
					if strings.HasPrefix(e.Path, f+string(os.PathSeparator)) || e.Path == f {
						filtered = append(filtered, e)
						break
					}
				}
			}
			allEntries = filtered
		}
	} else {
		// Fallback: WalkDir ligero si índice vacío (primera vez)
		var allPaths []string
		seen := make(map[string]bool)
		for _, f := range normalizedFolders {
			list, err := scanFolderPaths(f, nil)
			if err != nil {
				continue
			}
			for _, p := range list {
				if !seen[p] {
					seen[p] = true
					allPaths = append(allPaths, p)
				}
			}
		}
		sort.Strings(allPaths)
		// Filtrar por query sobre path
		filteredPaths := allPaths
		if q := strings.TrimSpace(query); q != "" {
			q = strings.ToLower(q)
			tmp := make([]string, 0, len(allPaths))
			for _, p := range allPaths {
				base := strings.ToLower(filepath.Base(p))
				if strings.Contains(strings.ToLower(p), q) || strings.Contains(base, q) {
					tmp = append(tmp, p)
				}
			}
			filteredPaths = tmp
		}
		total := len(filteredPaths)
		if offset >= total {
			return MusicPageDTO{Tracks: []MusicTrackDTO{}, Total: total, Offset: offset, Limit: limit}, nil
		}
		end := offset + limit
		if end > total {
			end = total
		}
		pagePaths := filteredPaths[offset:end]
		tracks := make([]MusicTrackDTO, 0, len(pagePaths))
		for _, p := range pagePaths {
			dto, err := m.BuildTrackDTO(p, withThumb, withRaw)
			if err != nil {
				dto = MusicTrackDTO{
					Path:     p,
					FileName: filepath.Base(p),
					Title:    strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)),
					Artist:   "Desconocido",
				}
			}
			tracks = append(tracks, dto)
		}
		return MusicPageDTO{Tracks: tracks, Total: total, Offset: offset, Limit: limit}, nil
	}
	// Ya tenemos allEntries desde índice (filtradas por query y folder)
	// Orden ya es por Path (List/Search lo garantiza)
	total := len(allEntries)
	if offset >= total {
		return MusicPageDTO{Tracks: []MusicTrackDTO{}, Total: total, Offset: offset, Limit: limit}, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	pageEntries := allEntries[offset:end]
	tracks := make([]MusicTrackDTO, 0, len(pageEntries))
	for _, e := range pageEntries {
		dto := MusicTrackDTO{
			Path:     e.Path,
			FileName: filepath.Base(e.Path),
			Title:    cleanMetaTag(e.Title),
			Artist:   cleanMetaTag(e.Artist),
			Album:    cleanMetaTag(e.Album),
			Year:     e.Year,
			Duration: e.Duration,
			Genre:    cleanMetaTag(e.Genre),
			HasCover: e.HasCover,
		}
		if dto.Title == "" {
			dto.Title = cleanMetaTag(strings.TrimSuffix(dto.FileName, filepath.Ext(dto.FileName)))
		}
		if dto.Artist == "" {
			dto.Artist = "Desconocido"
		}
		// CoverID en lugar de data URI para lista ligera
		if e.HasCover && e.CoverID != "" {
			dto.CoverMime = e.CoverMime
			// No incluir thumb/raw en lista por defecto (lazy)
			// Si conThumb/withRaw true, generar bajo demanda via GetCoverBase64 con CoverID
			if withThumb || withRaw {
				// Intentar obtener thumb desde cache via CoverID
				if uri, err := m.getCoverByID(e.CoverID, e.CoverMime, withRaw); err == nil {
					if withThumb {
						dto.CoverThumb = uri
					}
					if withRaw {
						dto.CoverRaw = uri
					}
				} else if withThumb || withRaw {
					// Fallback: extraer y cachear si no hay thumb
					if uri, err := m.GetCoverBase64(e.Path, "thumb"); err == nil {
						if withThumb {
							dto.CoverThumb = uri
						}
						if withRaw {
							dto.CoverRaw = uri
						}
					}
				}
			}
		}
		tracks = append(tracks, dto)
	}
	return MusicPageDTO{Tracks: tracks, Total: total, Offset: offset, Limit: limit}, nil
}

// getCoverByID intenta obtener cover por CoverID desde zip shards.
func (m *Manager) getCoverByID(coverID, mime string, raw bool) (string, error) {
	if coverID == "" {
		return "", fmt.Errorf("sin CoverID")
	}
	data, _, err := m.readCoverFromZipByHash(coverID)
	if err != nil {
		return "", err
	}
	if raw {
		b64 := base64.StdEncoding.EncodeToString(data)
		if mime == "" {
			mime = "image/jpeg"
		}
		return fmt.Sprintf("data:%s;base64,%s", mime, b64), nil
	}
	thumb, thumbMime, err := makeThumb128(data, mime)
	if err != nil {
		b64 := base64.StdEncoding.EncodeToString(data)
		return fmt.Sprintf("data:%s;base64,%s", mime, b64), nil
	}
	b64 := base64.StdEncoding.EncodeToString(thumb)
	return fmt.Sprintf("data:%s;base64,%s", thumbMime, b64), nil
}

// makeThumb128 genera thumbnail 128x128 comprimido, preserva transparencia PNG
func makeThumb128(data []byte, mime string) ([]byte, string, error) {
	if len(data) == 0 {
		return nil, mime, fmt.Errorf("sin datos")
	}
	// Verificación rápida de tamaño
	if cfg, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
		if cfg.Width > 5000 || cfg.Height > 5000 {
			return nil, mime, fmt.Errorf("imagen demasiado grande %dx%d", cfg.Width, cfg.Height)
		}
	}
	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		if img2, _, err2 := image.Decode(bytes.NewReader(data)); err2 == nil {
			img = img2
		} else {
			return nil, mime, err
		}
	}
	// Thumbnail 128x128 con Box (rápido, suficiente para lista/cola)
	img = imaging.Thumbnail(img, 128, 128, imaging.Box)
	var buf bytes.Buffer
	if strings.Contains(strings.ToLower(mime), "png") && hasTransparency(img) {
		if err := png.Encode(&buf, img); err != nil {
			return nil, mime, err
		}
		return buf.Bytes(), "image/png", nil
	}
	opts := jpeg.Options{Quality: 70}
	if err := jpeg.Encode(&buf, img, &opts); err != nil {
		return nil, mime, err
	}
	return buf.Bytes(), "image/jpeg", nil
}
