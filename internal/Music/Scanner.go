package music

import (
	"bytes"
	"context"
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

var audioExts = map[string]bool{".mp3": true, ".flac": true, ".m4a": true, ".m4b": true, ".ogg": true, ".opus": true, ".wav": true, ".aiff": true, ".wma": true, ".aac": true}

type scanResult struct {
	path      string
	title     string
	artist    string
	album     string
	year      int
	duration  float64
	trackNum  int
	genre     string
	hasCover  bool
	coverMime string
	coverData []byte
}

func scanFolderPaths(folder string, onProgress func(current int, file string)) ([]string, error) {
	return scanFolderPathsCtx(context.Background(), folder, onProgress)
}

func scanFolderPathsCtx(ctx context.Context, folder string, onProgress func(current int, file string)) ([]string, error) {
	folder = filepath.Clean(folder)
	info, err := os.Stat(folder)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, err
	}
	var out []string
	err = filepath.WalkDir(folder, func(path string, d os.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !audioExts[ext] {
			return nil
		}
		out = append(out, path)
		if onProgress != nil {
			onProgress(len(out), path)
		}
		if len(out) >= 5000 {
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil && err != filepath.SkipAll && err != context.Canceled {
		return nil, err
	}
	if err == context.Canceled {
		return out, context.Canceled
	}
	sort.Strings(out)
	return out, nil
}

// ScanFolder incremental para una carpeta
func (m *Manager) ScanFolder(folder string) ([]string, error) {
	return m.ScanFolders([]string{folder})
}

// ScanFolders incremental con pool 3 workers, detección por size/modTime
// PROHIBIDO: nunca inyectar carpeta por defecto — solo escanea las carpetas explícitas del usuario.
func (m *Manager) ScanFolders(folders []string) ([]string, error) {
	if !m.scanMu.TryLock() {
		return nil, fmt.Errorf("escaneo ya en curso")
	}
	defer m.scanMu.Unlock()

	// Contexto cancelable para poder abortar desde CancelScan()
	scanCtx, cancel := context.WithCancel(context.Background())
	m.scanCancelMu.Lock()
	m.scanCancel = cancel
	m.scanCancelMu.Unlock()
	defer func() {
		m.scanCancelMu.Lock()
		m.scanCancel = nil
		m.scanCancelMu.Unlock()
		cancel()
	}()

	// 1. Discover — sin carpeta por defecto, solo las recibidas
	m.setProgress(ScanProgress{Scanning: true, Current: 0, Total: 0, Discovered: 0})
	seen := make(map[string]bool)
	var allPaths []string
	for _, f := range folders {
		select {
		case <-scanCtx.Done():
			m.setProgress(ScanProgress{Scanning: false, Discovered: len(allPaths), Total: len(allPaths), Current: 0})
			return nil, fmt.Errorf("escaneo cancelado")
		default:
		}
		f = filepath.Clean(f)
		m.setProgress(ScanProgress{Scanning: true, Folder: f, Discovered: len(allPaths)})
		list, err := scanFolderPathsCtx(scanCtx, f, nil)
		if err == context.Canceled {
			m.setProgress(ScanProgress{Scanning: false, Discovered: len(allPaths), Total: len(allPaths), Current: 0})
			return nil, fmt.Errorf("escaneo cancelado")
		}
		for _, p := range list {
			if !seen[p] {
				seen[p] = true
				allPaths = append(allPaths, p)
			}
		}
	}
	sort.Strings(allPaths)
	discovered := len(allPaths)
	existingMap := make(map[string]bool, discovered)
	for _, p := range allPaths {
		existingMap[p] = true
	}
	// 2. Compare contra índice
	var toProcess []string
	skipped := 0
	for _, p := range allPaths {
		select {
		case <-scanCtx.Done():
			m.setProgress(ScanProgress{Scanning: false, Discovered: discovered, Total: discovered, Current: 0, Skipped: skipped, Removed: 0})
			return nil, fmt.Errorf("escaneo cancelado")
		default:
		}
		fi, err := os.Stat(p)
		if err != nil {
			continue
		}
		size := fi.Size()
		modTime := fi.ModTime().Unix()
		if m.index != nil && !m.index.NeedsUpdate(p, size, modTime) {
			skipped++
			continue
		}
		toProcess = append(toProcess, p)
	}
	// Detectar eliminados
	removed := 0
	if m.index != nil {
		removed = m.index.RemoveMissing(existingMap)
	}
	m.setProgress(ScanProgress{Scanning: true, Discovered: discovered, Total: discovered, Current: 0, Skipped: skipped, Removed: removed})

	// 3. Procesar con pool 3 workers
	processed := 0
	added := 0
	updated := 0
	errorsCount := 0
	if len(toProcess) > 0 {
		// Canal de trabajos
		jobs := make(chan string, len(toProcess))
		for _, p := range toProcess {
			jobs <- p
		}
		close(jobs)
		var wg sync.WaitGroup
		var mu sync.Mutex
		workers := 3
		if len(toProcess) < workers {
			workers = len(toProcess)
		}
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-scanCtx.Done():
						return
					case path, ok := <-jobs:
						if !ok {
							return
						}
						select {
						case <-scanCtx.Done():
							return
						default:
						}
					fi, err := os.Stat(path)
					if err != nil {
						mu.Lock()
						errorsCount++
						mu.Unlock()
						continue
					}
					size := fi.Size()
					modTime := fi.ModTime().Unix()
					// Verificar si es nuevo o modificado
					isNew := true
					if m.index != nil {
						if _, ok := m.index.Get(path); ok {
							isNew = false
						}
					}
					// Extraer metadata rápida
					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					r := extractMetaFast(ctx, path)
					cancel()
					// Detectar cover sin decodificar imagen completa
					hasCover := false
					coverMime := ""
					coverData := []byte{}
					func() {
						ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
						defer cancel2()
						f, err := audiometa.OpenContext(ctx2, path)
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
							// Solo guardar cover si es nueva o cambió
							coverData = arts[0].Data
						}
					}()
					entry := &MusicIndexEntry{
						ID:          GenerateID(path, size, modTime),
						Path:        path,
						Size:        size,
						ModTime:     modTime,
						Title:       r.title,
						Artist:      r.artist,
						Album:       r.album,
						Genre:       r.genre,
						Year:        r.year,
						TrackNumber: r.trackNum,
						Duration:    r.duration,
						HasCover:    hasCover,
						CoverMime:   coverMime,
						CachedAt:    time.Now().Unix(),
					}
					if hasCover && len(coverData) > 0 {
						// Dedup por hash
						hash := coverHash(coverData)
						entry.CoverID = hash
						// Guardar cover en zip (con lock interno)
						_ = m.saveCover(path, CoverEntry{TrackPath: path, Mime: coverMime, Data: coverData, CachedAt: time.Now().Format(time.RFC3339)})
					} else if existing, ok := m.index.Get(path); ok {
						entry.CoverID = existing.CoverID
					}
					m.index.Set(entry)
					mu.Lock()
					processed++
					if isNew {
						added++
					} else {
						updated++
					}
					// Progreso cada 10
					if processed%10 == 0 || processed == len(toProcess) {
						m.setProgress(ScanProgress{Scanning: true, Discovered: discovered, Total: discovered, Current: processed, Processed: processed, Skipped: skipped, Added: added, Updated: updated, Removed: removed, Errors: errorsCount})
					}
					mu.Unlock()
					}
				}
			}()
		}
		wg.Wait()
		// Si fue cancelado, no guardar como completado — marcar cancelado
		select {
		case <-scanCtx.Done():
			m.setProgress(ScanProgress{Scanning: false, Discovered: discovered, Total: discovered, Current: processed, Processed: processed, Skipped: skipped, Added: added, Updated: updated, Removed: removed, Errors: errorsCount, Folder: ""})
			return nil, fmt.Errorf("escaneo cancelado")
		default:
		}
		if m.index != nil {
			_ = m.index.Save()
		}
	}
	// Cancelación también puede llegar entre compare y save
	select {
	case <-scanCtx.Done():
		m.setProgress(ScanProgress{Scanning: false, Discovered: discovered, Total: discovered, Current: processed, Processed: processed, Skipped: skipped, Added: added, Updated: updated, Removed: removed, Errors: errorsCount, Folder: ""})
		return nil, fmt.Errorf("escaneo cancelado")
	default:
	}
	m.setProgress(ScanProgress{Scanning: false, Discovered: discovered, Total: discovered, Current: discovered, Processed: processed, Skipped: skipped, Added: added, Updated: updated, Removed: removed, Errors: errorsCount, Folder: ""})
	// Retornar todos los paths ordenados (índice)
	if m.index != nil {
		list := m.index.List()
		out := make([]string, 0, len(list))
		for _, e := range list {
			// Filtrar por folders si se pidieron
			if len(folders) == 0 {
				out = append(out, e.Path)
			} else {
				for _, f := range folders {
					f = filepath.Clean(f)
					if strings.HasPrefix(e.Path, f+string(os.PathSeparator)) || e.Path == f {
						out = append(out, e.Path)
						break
					}
				}
			}
		}
		sort.Strings(out)
		return out, nil
	}
	sort.Strings(allPaths)
	return allPaths, nil
}

func (m *Manager) cacheTracksBatch(paths []string) {
	// Desactivado: ya no cacheamos nada en disco (pedido: solo leer y retornar)
	// Se mantiene por compatibilidad pero no hace trabajo pesado
	m.setProgress(ScanProgress{Scanning: false, Current: len(paths), Total: len(paths)})
}

// cleanTag elimina BOM, reemplazo U+FFFD y caracteres de control/zero-width que aparecen como "☐" en la UI.
// Carga el título tal cual debe verse, sin basura de tags ID3 mal codificados.
func cleanTag(s string) string {
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
		// zero-width y soft hyphen que se ven como caja vacía
		if r == '\u200B' || r == '\u200C' || r == '\u200D' || r == '\u2060' || r == '\u00AD' || r == '\u034F' {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

func extractMetaFast(ctx context.Context, path string) scanResult {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	title := cleanTag(base)
	if title == "" {
		title = base
	}
	artist := "Desconocido"
	album := ""
	year := 0
	duration := 0.0
	trackNum := 0
	genre := ""
	// Sin covers: lectura ultra ligera, evita 3GB de RAM por thumbnails
	c2, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	f, err := audiometa.OpenContext(c2, path)
	if err == nil {
		defer f.Close()
		if t := cleanTag(f.Tags.Title); t != "" {
			title = t
		}
		if a := cleanTag(f.Tags.Artist); a != "" {
			artist = a
		} else if len(f.Tags.Artists) > 0 {
			if a2 := cleanTag(f.Tags.Artists[0]); a2 != "" {
				artist = a2
			}
		}
		album = cleanTag(f.Tags.Album)
		year = f.Tags.Year
		if len(f.Tags.Genres) > 0 {
			genre = cleanTag(f.Tags.Genres[0])
		}
		trackNum = f.Tags.TrackNumber
		if f.Audio.Duration > 0 {
			duration = f.Audio.Duration.Seconds()
		}
	}
	return scanResult{path: path, title: title, artist: artist, album: album, year: year, duration: duration, trackNum: trackNum, genre: genre, hasCover: false, coverMime: "", coverData: nil}
}

func thumbCover(data []byte, mime string) ([]byte, string, error) {
	// decodificar con verificación rápida de tamaño para evitar bitmaps gigantes
	// si la imagen es 4000x4000, decodificarla consume ~64MB; usar Box es más barato que Lanczos
	cfg, _, errCfg := image.DecodeConfig(bytes.NewReader(data))
	if errCfg == nil {
		// si es absurdamente grande (>3000), evitar decodificar del todo si excede límite
		if cfg.Width > 4000 || cfg.Height > 4000 {
			return nil, mime, fmt.Errorf("imagen demasiado grande %dx%d", cfg.Width, cfg.Height)
		}
		if cfg.Width > 1200 || cfg.Height > 1200 {
			// para imágenes grandes, decodificar y hacer thumbnail rápido con Box
			img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
			if err != nil {
				if img2, _, err2 := image.Decode(bytes.NewReader(data)); err2 == nil {
					img = img2
				} else {
					return nil, mime, err
				}
			}
			img = imaging.Thumbnail(img, 512, 512, imaging.Box)
			var buf bytes.Buffer
			if strings.Contains(strings.ToLower(mime), "png") && hasTransparency(img) {
				if err := png.Encode(&buf, img); err != nil {
					return nil, mime, err
				}
				return buf.Bytes(), "image/png", nil
			}
			opts := jpeg.Options{Quality: 75}
			if err := jpeg.Encode(&buf, img, &opts); err != nil {
				return nil, mime, err
			}
			return buf.Bytes(), "image/jpeg", nil
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
	// limitar a 512x512 manteniendo aspecto - Box es 3-4x más rápido que Lanczos y suficiente para thumbnails
	if img.Bounds().Dx() > 512 || img.Bounds().Dy() > 512 {
		img = imaging.Thumbnail(img, 512, 512, imaging.Box)
	}
	var buf bytes.Buffer
	if strings.Contains(strings.ToLower(mime), "png") && hasTransparency(img) {
		if err := png.Encode(&buf, img); err != nil {
			return nil, mime, err
		}
		return buf.Bytes(), "image/png", nil
	}
	opts := jpeg.Options{Quality: 75}
	if err := jpeg.Encode(&buf, img, &opts); err != nil {
		return nil, mime, err
	}
	// liberar referencia para GC rápido
	img = nil
	return buf.Bytes(), "image/jpeg", nil
}

func hasTransparency(img image.Image) bool {
	switch v := img.(type) {
	case *image.NRGBA:
		for i := 3; i < len(v.Pix); i += 4 {
			if v.Pix[i] != 255 {
				return true
			}
		}
	case *image.RGBA:
		for i := 3; i < len(v.Pix); i += 4 {
			if v.Pix[i] != 255 {
				return true
			}
		}
	}
	return false
}
