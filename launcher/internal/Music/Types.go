package music

// MusicMeta representa metadata ligera cacheada - CoverID apunta a cover deduplicada
type MusicMeta struct {
	Title       string  `json:"title"`
	Artist      string  `json:"artist"`
	Album       string  `json:"album"`
	Year        int     `json:"year"`
	Duration    float64 `json:"duration"` // segundos
	TrackNumber int     `json:"trackNumber"`
	Genre       string  `json:"genre"`
	HasCover    bool    `json:"hasCover"`
	CoverMime   string  `json:"coverMime,omitempty"`
	CoverID     string  `json:"coverId,omitempty"` // hash SHA256 de la carátula deduplicada
	CachedAt    string  `json:"cachedAt"`
	FileModTime string  `json:"fileModTime,omitempty"`
	FileSize    int64   `json:"fileSize,omitempty"`
}

// CoverEntry guarda carátula en shard comprimido - Data es bytes crudos (thumbnail)
type CoverEntry struct {
	TrackPath   string `json:"trackPath"`
	Mime        string `json:"mime"`
	Data        []byte `json:"data"`
	CachedAt    string `json:"cachedAt"`
	FileModTime string `json:"fileModTime,omitempty"`
}

// CoverIndex apunta a archivo separado dentro de zip (5-10 por zip)
type CoverIndex struct {
	Hash      string `json:"hash"`
	ShardFile string `json:"shardFile"` // ej: shard_000.zip
	FileName  string `json:"fileName"`  // ej: a1b2c3d4.jpg
	Mime      string `json:"mime"`
	Size      int64  `json:"size"`
	CachedAt  string `json:"cachedAt"`
}

// ScanProgress progreso exacto sin redondeo - extendido para incremental
type ScanProgress struct {
	Scanning    bool   `json:"scanning"`
	Current     int    `json:"current"`
	Total       int    `json:"total"`
	Folder      string `json:"folder"`
	CurrentFile string `json:"currentFile"`
	Discovered  int    `json:"discovered"`
	Processed   int    `json:"processed"`
	Skipped     int    `json:"skipped"`
	Added       int    `json:"added"`
	Updated     int    `json:"updated"`
	Removed     int    `json:"removed"`
	Errors      int    `json:"errors"`
}

// CoverCacheInfo estadísticas
type CoverCacheInfo struct {
	Total   int   `json:"total"`
	Valid   int   `json:"valid"`
	Expired int   `json:"expired"`
	Size    int64 `json:"size"`
}

// MusicTrackDTO para paginación con carátulas en base64
type MusicTrackDTO struct {
	Path       string  `json:"path"`
	FileName   string  `json:"fileName"`
	Title      string  `json:"title"`
	Artist     string  `json:"artist"`
	Album      string  `json:"album"`
	Year       int     `json:"year"`
	Duration   float64 `json:"duration"`
	Genre      string  `json:"genre"`
	HasCover   bool    `json:"hasCover"`
	CoverMime  string  `json:"coverMime,omitempty"`
	CoverThumb string  `json:"coverThumb,omitempty"` // data:image/...;base64, 128x128 comprimida
	CoverRaw   string  `json:"coverRaw,omitempty"`   // data:image/...;base64, original
}

type MusicPageDTO struct {
	Tracks []MusicTrackDTO `json:"tracks"`
	Total  int             `json:"total"`
	Offset int             `json:"offset"`
	Limit  int             `json:"limit"`
}
