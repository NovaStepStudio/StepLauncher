package music

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	metaShardSize  = 100
	coverShardSize = 10
)

func writeGobGZ(path string, v interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(gz)
	if err := enc.Encode(v); err != nil {
		_ = gz.Close()
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func readGobGZ(path string, out interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	data, err := io.ReadAll(gz)
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(bytes.NewReader(data))
	return dec.Decode(out)
}

func tryReadOldJSONGZ(path string, out interface{}) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return false
	}
	defer gz.Close()
	data, err := io.ReadAll(gz)
	if err != nil {
		return false
	}
	if err := json.Unmarshal(data, out); err == nil {
		return true
	}
	return false
}

func init() {
	gob.Register(map[string]MusicMeta{})
	gob.Register(map[string]CoverEntry{})
	gob.Register(map[string]CoverIndex{})
	gob.Register(MusicMeta{})
	gob.Register(CoverEntry{})
	gob.Register(CoverIndex{})
}

func (m *Manager) metaShardFiles() []string {
	entries, err := os.ReadDir(m.metaDir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasPrefix(n, "shard_") && (strings.HasSuffix(n, ".gob.gz") || strings.HasSuffix(n, ".json.gz")) {
			out = append(out, filepath.Join(m.metaDir, n))
		}
	}
	sort.Strings(out)
	return out
}

func (m *Manager) coverZipFiles() []string {
	entries, err := os.ReadDir(m.coversDir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasPrefix(n, "covers_") && strings.HasSuffix(n, ".zip") {
			out = append(out, filepath.Join(m.coversDir, n))
		}
	}
	sort.Strings(out)
	return out
}

func (m *Manager) coverIndexPath() string {
	return filepath.Join(m.coversDir, "index.gob.gz")
}

func (m *Manager) loadCoverIndex() map[string]CoverIndex {
	var idx map[string]CoverIndex
	if err := readGobGZ(m.coverIndexPath(), &idx); err == nil && idx != nil {
		return idx
	}
	if tryReadOldJSONGZ(m.coverIndexPath(), &idx) && idx != nil {
		return idx
	}
	return make(map[string]CoverIndex)
}

func (m *Manager) saveCoverIndex(idx map[string]CoverIndex) error {
	return writeGobGZ(m.coverIndexPath(), idx)
}

func (m *Manager) loadMetaShard(path string) map[string]MusicMeta {
	var mp map[string]MusicMeta
	if strings.HasSuffix(path, ".gob.gz") {
		if err := readGobGZ(path, &mp); err == nil && mp != nil {
			return mp
		}
	} else {
		if tryReadOldJSONGZ(path, &mp) && mp != nil {
			return mp
		}
		var mp2 map[string]MusicMeta
		if err := readGobGZ(path, &mp2); err == nil && mp2 != nil {
			return mp2
		}
	}
	if mp == nil {
		mp = map[string]MusicMeta{}
	}
	return mp
}

func (m *Manager) findMeta(trackPath string) (MusicMeta, string, bool) {
	trackPath = strings.TrimSpace(trackPath)
	for _, sf := range m.metaShardFiles() {
		mp := m.loadMetaShard(sf)
		if v, ok := mp[trackPath]; ok {
			return v, sf, true
		}
	}
	return MusicMeta{}, "", false
}

func coverHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func (m *Manager) findCover(trackPath string) (CoverEntry, string, bool) {
	trackPath = strings.TrimSpace(trackPath)
	meta, _, ok := m.findMeta(trackPath)
	if !ok || meta.CoverID == "" {
		return CoverEntry{}, "", false
	}
	hash := meta.CoverID
	idx := m.loadCoverIndex()
	cIdx, ok := idx[hash]
	if !ok {
		return CoverEntry{}, "", false
	}
	data, mime, err := m.readCoverFromZip(cIdx)
	if err != nil {
		return CoverEntry{}, "", false
	}
	return CoverEntry{TrackPath: trackPath, Mime: mime, Data: data, CachedAt: cIdx.CachedAt}, cIdx.ShardFile, true
}

func (m *Manager) readCoverFromZip(cIdx CoverIndex) ([]byte, string, error) {
	zr, err := zip.OpenReader(cIdx.ShardFile)
	if err != nil {
		return nil, "", err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.Name == cIdx.FileName {
			rc, err := f.Open()
			if err != nil {
				return nil, "", err
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, "", err
			}
			return data, cIdx.Mime, nil
		}
	}
	return nil, "", fmt.Errorf("cover no encontrado en zip")
}

func (m *Manager) readCoverFromZipByHash(hash string) ([]byte, string, error) {
	idx := m.loadCoverIndex()
	cIdx, ok := idx[hash]
	if !ok {
		return nil, "", fmt.Errorf("hash no encontrado")
	}
	return m.readCoverFromZip(cIdx)
}

func (m *Manager) saveCover(trackPath string, entry CoverEntry) error {
	m.coverMu.Lock()
	defer m.coverMu.Unlock()
	hash := coverHash(entry.Data)
	idx := m.loadCoverIndex()
	if _, ok := idx[hash]; ok {
		return nil
	}
	var targetZip string
	for _, zf := range m.coverZipFiles() {
		if countFilesInZip(zf) < coverShardSize {
			targetZip = zf
			break
		}
	}
	if targetZip == "" {
		if err := os.MkdirAll(m.coversDir, 0755); err != nil {
			return err
		}
		nextIdx := len(m.coverZipFiles())
		targetZip = filepath.Join(m.coversDir, fmt.Sprintf("covers_%03d.zip", nextIdx))
		for {
			if _, err := os.Stat(targetZip); os.IsNotExist(err) {
				break
			}
			nextIdx++
			targetZip = filepath.Join(m.coversDir, fmt.Sprintf("covers_%03d.zip", nextIdx))
		}
		f, err := os.Create(targetZip)
		if err != nil {
			return err
		}
		zw := zip.NewWriter(f)
		zw.Close()
		f.Close()
	}
	if err := addFileToZip(targetZip, hash, entry.Data, entry.Mime); err != nil {
		return err
	}
	idx = m.loadCoverIndex()
	ext := ".jpg"
	if strings.Contains(entry.Mime, "png") {
		ext = ".png"
	}
	idx[hash] = CoverIndex{
		Hash:      hash,
		ShardFile: targetZip,
		FileName:  hash + ext,
		Mime:      entry.Mime,
		Size:      int64(len(entry.Data)),
		CachedAt:  entry.CachedAt,
	}
	return m.saveCoverIndex(idx)
}

func (m *Manager) saveMeta(trackPath string, meta MusicMeta) error {
	trackPath = strings.TrimSpace(trackPath)
	if _, sf, ok := m.findMeta(trackPath); ok {
		mp := m.loadMetaShard(sf)
		mp[trackPath] = meta
		return writeGobGZ(sf, mp)
	}
	for _, sf := range m.metaShardFiles() {
		mp := m.loadMetaShard(sf)
		if len(mp) < metaShardSize {
			mp[trackPath] = meta
			return writeGobGZ(sf, mp)
		}
	}
	if err := os.MkdirAll(m.metaDir, 0755); err != nil {
		return err
	}
	idx := len(m.metaShardFiles())
	newPath := filepath.Join(m.metaDir, fmt.Sprintf("shard_%03d.gob.gz", idx))
	for {
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			break
		}
		idx++
		newPath = filepath.Join(m.metaDir, fmt.Sprintf("shard_%03d.gob.gz", idx))
	}
	mp := map[string]MusicMeta{trackPath: meta}
	return writeGobGZ(newPath, mp)
}

func countFilesInZip(path string) int {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return 0
	}
	defer zr.Close()
	return len(zr.File)
}

func addFileToZip(zipPath, hash string, data []byte, mime string) error {
	var existingData = make(map[string][]byte)
	if _, err := os.Stat(zipPath); err == nil {
		zr, err := zip.OpenReader(zipPath)
		if err == nil {
			for _, f := range zr.File {
				rc, err := f.Open()
				if err != nil {
					continue
				}
				b, err := io.ReadAll(rc)
				rc.Close()
				if err != nil {
					continue
				}
				existingData[f.Name] = b
			}
			zr.Close()
		}
	}
	ext := ".jpg"
	if strings.Contains(mime, "png") {
		ext = ".png"
	}
	fileName := hash + ext
	existingData[fileName] = data
	tmp := zipPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(f)
	for name, b := range existingData {
		w, err := zw.Create(name)
		if err != nil {
			zw.Close()
			f.Close()
			return err
		}
		w.Write(b)
	}
	zw.Close()
	f.Close()
	return os.Rename(tmp, zipPath)
}

func (m *Manager) saveMetasBatch(metas map[string]MusicMeta) error {
	if len(metas) == 0 {
		return nil
	}
	m.memMu.Lock()
	for k, v := range metas {
		if v.CachedAt == "" {
			v.CachedAt = time.Now().Format(time.RFC3339)
		}
		if v.FileModTime == "" {
			if fi, err := os.Stat(k); err == nil {
				v.FileModTime = fi.ModTime().Format(time.RFC3339)
				v.FileSize = fi.Size()
			}
		}
		if v.Title == "" {
			v.Title = strings.TrimSuffix(filepath.Base(k), filepath.Ext(k))
		}
		if v.Artist == "" {
			v.Artist = "Desconocido"
		}
		m.metaMem[k] = v
		metas[k] = v
	}
	all := make(map[string]MusicMeta, len(m.metaMem))
	for k, v := range m.metaMem {
		all[k] = v
	}
	m.memMu.Unlock()
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, sf := range m.metaShardFiles() {
		_ = os.Remove(sf)
	}
	keys := make([]string, 0, len(all))
	for k := range all {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var cur map[string]MusicMeta
	var idx int
	for i, k := range keys {
		if cur == nil {
			cur = make(map[string]MusicMeta, metaShardSize)
		}
		cur[k] = all[k]
		if len(cur) >= metaShardSize || i == len(keys)-1 {
			path := filepath.Join(m.metaDir, fmt.Sprintf("shard_%03d.gob.gz", idx))
			_ = writeGobGZ(path, cur)
			idx++
			cur = nil
		}
	}
	return nil
}

func (m *Manager) saveCoversBatch(covers map[string]CoverEntry) error {
	if len(covers) == 0 {
		return nil
	}
	dedup := make(map[string]CoverEntry)
	trackToHash := make(map[string]string)
	for trackPath, entry := range covers {
		h := coverHash(entry.Data)
		if _, ok := dedup[h]; !ok {
			dedup[h] = entry
		}
		trackToHash[trackPath] = h
	}
	m.memMu.Lock()
	for trackPath, hash := range trackToHash {
		if meta, ok := m.metaMem[trackPath]; ok {
			meta.CoverID = hash
			meta.HasCover = true
			if e, ok := dedup[hash]; ok {
				meta.CoverMime = e.Mime
			}
			m.metaMem[trackPath] = meta
		} else {
			m.metaMem[trackPath] = MusicMeta{
				Title:     strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath)),
				Artist:    "Desconocido",
				HasCover:  true,
				CoverID:   hash,
				CoverMime: dedup[hash].Mime,
				CachedAt:  time.Now().Format(time.RFC3339),
			}
		}
	}
	m.memMu.Unlock()
	for hash, entry := range dedup {
		idx := m.loadCoverIndex()
		if _, ok := idx[hash]; ok {
			continue
		}
		m.mu.Lock()
		var targetZip string
		for _, zf := range m.coverZipFiles() {
			if countFilesInZip(zf) < coverShardSize {
				targetZip = zf
				break
			}
		}
		if targetZip == "" {
			os.MkdirAll(m.coversDir, 0755)
			nextIdx := len(m.coverZipFiles())
			targetZip = filepath.Join(m.coversDir, fmt.Sprintf("covers_%03d.zip", nextIdx))
			for {
				if _, err := os.Stat(targetZip); os.IsNotExist(err) {
					break
				}
				nextIdx++
				targetZip = filepath.Join(m.coversDir, fmt.Sprintf("covers_%03d.zip", nextIdx))
			}
			f, _ := os.Create(targetZip)
			if f != nil {
				zw := zip.NewWriter(f)
				zw.Close()
				f.Close()
			}
		}
		addFileToZip(targetZip, hash, entry.Data, entry.Mime)
		idx = m.loadCoverIndex()
		ext := ".jpg"
		if strings.Contains(entry.Mime, "png") {
			ext = ".png"
		}
		idx[hash] = CoverIndex{
			Hash:      hash,
			ShardFile: targetZip,
			FileName:  hash + ext,
			Mime:      entry.Mime,
			Size:      int64(len(entry.Data)),
			CachedAt:  entry.CachedAt,
		}
		m.saveCoverIndex(idx)
		m.mu.Unlock()
	}
	metasToSave := make(map[string]MusicMeta)
	m.memMu.RLock()
	for k, v := range m.metaMem {
		metasToSave[k] = v
	}
	m.memMu.RUnlock()
	return m.saveMetasBatch(metasToSave)
}

func (m *Manager) removeMeta(trackPath string) {
	trackPath = strings.TrimSpace(trackPath)
	for _, sf := range m.metaShardFiles() {
		mp := m.loadMetaShard(sf)
		if _, ok := mp[trackPath]; ok {
			delete(mp, trackPath)
			if len(mp) == 0 {
				_ = os.Remove(sf)
			} else {
				_ = writeGobGZ(sf, mp)
			}
			return
		}
	}
}

func (m *Manager) removeCover(trackPath string) {
	trackPath = strings.TrimSpace(trackPath)
	m.memMu.RLock()
	meta, ok := m.metaMem[trackPath]
	m.memMu.RUnlock()
	if !ok || meta.CoverID == "" {
		return
	}
	hash := meta.CoverID
	m.memMu.RLock()
	cnt := 0
	for _, mm := range m.metaMem {
		if mm.CoverID == hash {
			cnt++
		}
	}
	m.memMu.RUnlock()
	if cnt > 1 {
		m.memMu.Lock()
		if mm, ok := m.metaMem[trackPath]; ok {
			mm.HasCover = false
			mm.CoverID = ""
			mm.CoverMime = ""
			m.metaMem[trackPath] = mm
		}
		m.memMu.Unlock()
		m.mu.Lock()
		_ = m.saveMeta(trackPath, m.metaMem[trackPath])
		m.mu.Unlock()
		return
	}
	idx := m.loadCoverIndex()
	cIdx, ok := idx[hash]
	if !ok {
		return
	}
	removeFileFromZip(cIdx.ShardFile, cIdx.FileName)
	delete(idx, hash)
	m.saveCoverIndex(idx)
	if countFilesInZip(cIdx.ShardFile) == 0 {
		_ = os.Remove(cIdx.ShardFile)
	}
	m.memMu.Lock()
	if mm, ok := m.metaMem[trackPath]; ok {
		mm.HasCover = false
		mm.CoverID = ""
		mm.CoverMime = ""
		m.metaMem[trackPath] = mm
		m.memMu.Unlock()
		m.mu.Lock()
		_ = m.saveMeta(trackPath, mm)
		m.mu.Unlock()
	} else {
		m.memMu.Unlock()
	}
}

func removeFileFromZip(zipPath, fileName string) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return
	}
	defer zr.Close()
	tmp := zipPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return
	}
	zw := zip.NewWriter(f)
	for _, zf := range zr.File {
		if zf.Name == fileName {
			continue
		}
		rc, _ := zf.Open()
		b, _ := io.ReadAll(rc)
		rc.Close()
		w, _ := zw.Create(zf.Name)
		w.Write(b)
	}
	zw.Close()
	f.Close()
	zr.Close()
	os.Rename(tmp, zipPath)
}

func (m *Manager) cleanLegacy() {
	root := m.rootDir
	legacy := []string{
		filepath.Join(root, "cache", "music_metadata.json"),
		filepath.Join(root, "cache", "covers"),
	}
	for _, p := range legacy {
		if p == filepath.Join(root, "cache", "covers") {
			if _, err := os.Stat(p); err == nil {
				entries, _ := os.ReadDir(p)
				hasLegacy := false
				for _, e := range entries {
					n := e.Name()
					if strings.HasSuffix(n, ".cover") || strings.HasSuffix(n, ".json") && !strings.HasPrefix(n, "covers_") {
						hasLegacy = true
						break
					}
					if strings.HasPrefix(n, "shard_") && !strings.HasPrefix(n, "covers_") {
						hasLegacy = true
						break
					}
				}
				if hasLegacy {
					_ = os.RemoveAll(p)
				}
			}
		} else {
			_ = os.Remove(p)
		}
	}
	_ = os.Remove(filepath.Join(root, "cache", "music_metadata.json"))
	// migrar viejos gob de covers a zip
	for _, sf := range m.coverShardFilesOld() {
		mp := m.loadCoverShardOld(sf)
		for _, e := range mp {
			h := coverHash(e.Data)
			idx := m.loadCoverIndex()
			if _, ok := idx[h]; ok {
				continue
			}
			m.saveCover(e.TrackPath, e)
		}
		_ = os.Remove(sf)
	}
	for _, sf := range m.metaShardFiles() {
		if strings.HasSuffix(sf, ".json.gz") {
			mp := m.loadMetaShard(sf)
			newPath := strings.TrimSuffix(sf, ".json.gz") + ".gob.gz"
			_ = writeGobGZ(newPath, mp)
			_ = os.Remove(sf)
		}
	}
}

func (m *Manager) coverShardFilesOld() []string {
	entries, err := os.ReadDir(m.coversDir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasPrefix(n, "covers_") && (strings.HasSuffix(n, ".gob.gz") || strings.HasSuffix(n, ".json.gz")) {
			out = append(out, filepath.Join(m.coversDir, n))
		}
	}
	sort.Strings(out)
	return out
}

func (m *Manager) loadCoverShardOld(path string) map[string]CoverEntry {
	var mp map[string]CoverEntry
	if strings.HasSuffix(path, ".gob.gz") {
		readGobGZ(path, &mp)
	} else {
		tryReadOldJSONGZ(path, &mp)
	}
	if mp == nil {
		mp = map[string]CoverEntry{}
	}
	return mp
}

func (m *Manager) getAllMetaRaw() map[string]MusicMeta {
	out := map[string]MusicMeta{}
	now := time.Now()
	for _, sf := range m.metaShardFiles() {
		mp := m.loadMetaShard(sf)
		for k, v := range mp {
			if t, err := time.Parse(time.RFC3339, v.CachedAt); err == nil {
				if now.Sub(t) > 7*24*time.Hour {
					continue
				}
			}
			if v.FileModTime != "" {
				if fi, err := os.Stat(k); err == nil {
					if fi.ModTime().Format(time.RFC3339) != v.FileModTime || fi.Size() != v.FileSize {
						continue
					}
				}
			}
			out[k] = v
		}
	}
	return out
}

func (m *Manager) getAllMetaRawUnlocked() map[string]MusicMeta {
	out := map[string]MusicMeta{}
	now := time.Now()
	for _, sf := range m.metaShardFiles() {
		mp := m.loadMetaShard(sf)
		for k, v := range mp {
			if t, err := time.Parse(time.RFC3339, v.CachedAt); err == nil {
				if now.Sub(t) > 7*24*time.Hour {
					continue
				}
			}
			out[k] = v
		}
	}
	return out
}
