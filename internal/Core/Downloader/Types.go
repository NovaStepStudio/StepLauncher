package downloader

import (
	"net/http"

	"StepLauncher/internal/Core/Cache"
)

type Config struct {
	WorkDir  string
	CacheDir string
	JavaRuntimeDir     string
	MaxRetries         int
	MaxConcurrency     int
	MaxRAM             int
	SkipVerify         bool
	LogFn              func(string, ...interface{})
	BroadcastFn        func([]byte)
	HTTPClient         *http.Client
	CacheManager       *cache.Manager
	MaxMbps            float64
	MinMbps            float64
	StallTimeout       int
	MaxStallReDownload int
}

type DownloadFilter struct {
	Version            string `json:"version"`
	Client             bool   `json:"client"`
	Libraries          bool   `json:"libraries"`
	Natives            bool   `json:"natives"`
	Assets             bool   `json:"assets"`
	Java               bool   `json:"java"`
	InstanceVersionDir string `json:"-"`
}

type DownloadState string

const (
	StatePending     DownloadState = "pending"
	StateDownloading DownloadState = "downloading"
	StatePaused      DownloadState = "paused"
	StateCancelled   DownloadState = "cancelled"
	StateCompleted   DownloadState = "completed"
	StateError       DownloadState = "error"
	StateVerifying   DownloadState = "verifying"
	StateReDownload  DownloadState = "redownloading"
)

type DownloadProgress struct {
	MBDownloaded      float64       `json:"mbDownloaded"`
	MBTotal           float64       `json:"mbTotal"`
	Percent           float64       `json:"percent"`
	State             DownloadState `json:"state"`
	CurrentSection    string        `json:"currentSection"`
	SectionsCompleted []string      `json:"sectionsCompleted"`
	SectionsTotal     int           `json:"sectionsTotal"`
	FilesDownloaded   int           `json:"filesDownloaded"`
	FilesTotal        int           `json:"filesTotal"`
	FilesExisting     int           `json:"filesExisting"`
	CurrentFile       string        `json:"currentFile"`
	CurrentURL        string        `json:"currentUrl"`
	CurrentDest       string        `json:"currentDest"`
	CurrentProgress   float64       `json:"currentProgress"`
	Log               []string      `json:"log"`

	Sections       []SectionProgress `json:"sections"`
	ActiveFiles    []FileProgress    `json:"activeFiles"`
	QueuedCount    int               `json:"queuedCount"`
	QueuedPreview  []string          `json:"queuedPreview"`
	SpeedMbps      float64           `json:"speedMbps"`
}

type FileProgress struct {
	Name       string  `json:"name"`
	Section    string  `json:"section"`
	Size       int64   `json:"size"`
	Downloaded int64   `json:"downloaded"`
	Percent    float64 `json:"percent"`
	State      string  `json:"state"`
}

type SectionProgress struct {
	Name         string  `json:"name"`
	TotalFiles   int     `json:"totalFiles"`
	DoneFiles    int     `json:"doneFiles"`
	MBTotal      float64 `json:"mbTotal"`
	MBDownloaded float64 `json:"mbDownloaded"`
}

const (
	FilePending     = "pending"
	FileDownloading = "downloading"
	FileDone        = "done"
	FileExisting    = "existing"
	FileError       = "error"
)

type ProgressUpdate struct {
	Section      string
	File         string
	Done         int
	Total        int
	Complete     bool
	Error        string
	Bytes        int64
	Existing     bool
	MBDownloaded float64
	MBTotal      float64
}

type DownloadTask struct {
	URL     string
	Dest    string
	SHA1    string
	Size    int64
	Section string
}

type ManifestVersion struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Type        string `json:"type"`
	ReleaseTime string `json:"releaseTime"`
}

type LatestInfo struct {
	Release  string `json:"release"`
	Snapshot string `json:"snapshot"`
}

type Manifest struct {
	Versions []ManifestVersion `json:"versions"`
	Latest   LatestInfo        `json:"latest"`
}

type VersionJSON struct {
	ID           string `json:"id"`
	InheritsFrom string `json:"inheritsFrom"`
	Downloads    struct {
		Client struct {
			URL  string `json:"url"`
			SHA1 string `json:"sha1"`
			Size int64  `json:"size"`
		} `json:"client"`
	} `json:"downloads"`
	Libraries          []Library  `json:"libraries"`
	JavaVersion        JavaVer    `json:"javaVersion"`
	AssetIndex         AssetIdx   `json:"assetIndex"`
	Logging            *Logging   `json:"logging"`
	Arguments          *Arguments `json:"arguments"`
	MinecraftArguments string     `json:"minecraftArguments"`
	MainClass          string     `json:"mainClass"`
	Type               string     `json:"type"`
}

type Library struct {
	Name      string            `json:"name"`
	Downloads *LibDownloads     `json:"downloads"`
	Rules     []Rule            `json:"rules"`
	Natives   map[string]string `json:"natives"`
	Extract   *Extract          `json:"extract"`
	URL       string            `json:"url"`
}

type LibDownloads struct {
	Artifact    *Artifact           `json:"artifact"`
	Classifiers map[string]Artifact `json:"classifiers"`
}

type Artifact struct {
	Path string `json:"path"`
	URL  string `json:"url"`
	SHA1 string `json:"sha1"`
	Size int64  `json:"size"`
}

type Rule struct {
	Action string `json:"action"`
	OS     *struct {
		Name string `json:"name"`
	} `json:"os"`
	Features *struct {
		IsDemoUser          *bool `json:"is_demo_user"`
		HasCustomResolution *bool `json:"has_custom_resolution"`
		IsQuickPlay         *bool `json:"is_quick_play"`
	} `json:"features"`
}

type Extract struct {
	Exclude []string `json:"exclude"`
}

type JavaVer struct {
	Component    string `json:"component"`
	MajorVersion int    `json:"majorVersion"`
}

type AssetIdx struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	SHA1      string `json:"sha1"`
	TotalSize int64  `json:"totalSize"`
}

type AssetIndexJSON struct {
	Objects        map[string]AssetObject `json:"objects"`
	MapToResources bool                   `json:"map_to_resources"`
	Virtual        bool                   `json:"virtual"`
}

type AssetObject struct {
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}

type Logging struct {
	Client *struct {
		Argument string `json:"argument"`
		File     struct {
			ID   string `json:"id"`
			Size int64  `json:"size"`
			SHA1 string `json:"sha1"`
			URL  string `json:"url"`
		} `json:"file"`
		Type string `json:"type"`
	} `json:"client"`
}

type Arguments struct {
	Game []interface{} `json:"game"`
	JVM  []interface{} `json:"jvm"`
}

type JavaProducts struct {
	WindowsX64 map[string][]JavaProduct `json:"windows-x64"`
	Linux      map[string][]JavaProduct `json:"linux"`
	MacOS      map[string][]JavaProduct `json:"mac-os"`
	MacARM64   map[string][]JavaProduct `json:"mac-os-arm64"`
}

type JavaProduct struct {
	Manifest struct {
		URL  string `json:"url"`
		SHA1 string `json:"sha1"`
	} `json:"manifest"`
}

type JavaManifest struct {
	Files map[string]JavaFile `json:"files"`
}

type JavaFile struct {
	Type       string `json:"type"`
	Executable bool   `json:"executable"`
	Downloads  *struct {
		Raw *struct {
			URL  string `json:"url"`
			SHA1 string `json:"sha1"`
			Size int64  `json:"size"`
		} `json:"raw"`
	} `json:"downloads"`
}
