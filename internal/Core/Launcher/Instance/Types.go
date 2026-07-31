package instance

import (
	"time"

	"StepLauncher/internal/Core/Launcher"
)

type InstanceMetadata struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Banner      string   `json:"banner"`
	Background  string   `json:"background"`
	AccentColor string   `json:"accentColor"`
	Group       string   `json:"group"`
	Tags        []string `json:"tags"`
	Favorite    bool     `json:"favorite"`
	CreatedAt   string   `json:"createdAt"`
	LastPlayed  string   `json:"lastPlayed"`
	PlayTime    int64    `json:"playTime"`
	Versions    []string `json:"versions"`
	ConfigPath  string   `json:"configPath"`
}

type InstanceLaunchConfig struct {
	Version  string `json:"version,omitempty"`
	JavaExec string `json:"javaExec,omitempty"`
	MinRAM   *int   `json:"minRam,omitempty"`
	MaxRAM   *int   `json:"maxRam,omitempty"`

	UseOfficialJava      *bool  `json:"useOfficialJava,omitempty"`
	Fullscreen           *bool  `json:"fullscreen,omitempty"`
	HardwareAcceleration *bool  `json:"hardwareAcceleration,omitempty"`
	GCPreset             string `json:"gcPreset,omitempty"`
	GPUPreference        string `json:"gpuPreference,omitempty"`
	CustomResolution     *bool  `json:"customResolution,omitempty"`
	ResWidth             *int   `json:"resWidth,omitempty"`
	ResHeight            *int   `json:"resHeight,omitempty"`
	DemoUser             *bool  `json:"demoUser,omitempty"`
	WindowTitle          string `json:"windowTitle,omitempty"`
	UserType             string `json:"userType,omitempty"`

	Advanced *launcher.AdvancedConfig `json:"advanced,omitempty"`

	GameDir      string `json:"gameDir,omitempty"`
	AssetsDir    string `json:"assetsDir,omitempty"`
	LibrariesDir string `json:"librariesDir,omitempty"`
	NativesDir   string `json:"nativesDir,omitempty"`
	VersionsDir  string `json:"versionsDir,omitempty"`
	RuntimeDir   string `json:"runtimeDir,omitempty"`
	CacheDir     string `json:"cacheDir,omitempty"`

	AssetIndexID  string `json:"assetIndexId,omitempty"`
	DisableAssets *bool  `json:"disableAssets,omitempty"`

	ServerAddress string `json:"serverAddress,omitempty"`
	ServerPort    *int   `json:"serverPort,omitempty"`

	MaxMetaspaceSize *int `json:"maxMetaspaceSize,omitempty"`
	StackSize        *int `json:"stackSize,omitempty"`

	AllowServerList  *bool `json:"allowServerList,omitempty"`
	AllowMultiplayer *bool `json:"allowMultiplayer,omitempty"`
	AllowChat        *bool `json:"allowChat,omitempty"`
	AllowRealms      *bool `json:"allowRealms,omitempty"`
	FramerateLimit   *int  `json:"framerateLimit,omitempty"`

	LogKeepDays *int `json:"logKeepDays,omitempty"`
	LogMaxFiles *int `json:"logMaxFiles,omitempty"`

	PreLaunchCommand  string `json:"preLaunchCommand,omitempty"`
	PostLaunchCommand string `json:"postLaunchCommand,omitempty"`

	EnvironmentVars map[string]string `json:"environmentVars,omitempty"`
	JavaArgs        []string          `json:"javaArgs,omitempty"`
	GameArgs        []string          `json:"gameArgs,omitempty"`

	SkipLibraryCheck    *bool `json:"skipLibraryCheck,omitempty"`
	SkipAssetCheck      *bool `json:"skipAssetCheck,omitempty"`
	SkipNativeExtract   *bool `json:"skipNativeExtract,omitempty"`
	SkipVersionDownload *bool `json:"skipVersionDownload,omitempty"`
}

type InstanceLock struct {
	PID     int    `json:"pid"`
	Action  string `json:"action"`
	Version string `json:"version,omitempty"`
	Since   string `json:"since"`
}

type InstanceInfo struct {
	Name       string   `json:"name"`
	Title      string   `json:"title"`
	Versions   []string `json:"versions"`
	Favorite   bool     `json:"favorite"`
	Group      string   `json:"group"`
	LastPlayed string   `json:"lastPlayed"`
	PlayTime   int64    `json:"playTime"`
}

type CreateInstanceReq struct {
	Name         string                `json:"name"`
	Version      string                `json:"version,omitempty"`
	Title        string                `json:"title,omitempty"`
	Description  string                `json:"description,omitempty"`
	Icon         string                `json:"icon,omitempty"`
	Banner       string                `json:"banner,omitempty"`
	Background   string                `json:"background,omitempty"`
	AccentColor  string                `json:"accentColor,omitempty"`
	Group        string                `json:"group,omitempty"`
	Tags         []string              `json:"tags,omitempty"`
	Favorite     *bool                 `json:"favorite,omitempty"`
	LaunchConfig *InstanceLaunchConfig `json:"launchConfig,omitempty"`
}

type UpdateMetadataReq struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	Banner      string   `json:"banner,omitempty"`
	Background  string   `json:"background,omitempty"`
	AccentColor string   `json:"accentColor,omitempty"`
	Group       string   `json:"group,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Favorite    *bool    `json:"favorite,omitempty"`
}

type CloneInstanceReq struct {
	NewName      string `json:"newName"`
	CopyVersions bool   `json:"copyVersions"`
}

type AddVersionReq struct {
	Version        string `json:"version"`
	Client         *bool  `json:"client"`
	Libraries      *bool  `json:"libraries"`
	Natives        *bool  `json:"natives"`
	Assets         *bool  `json:"assets"`
	Java           *bool  `json:"java"`
	MaxRetries     *int   `json:"maxRetries"`
	MaxConcurrency *int   `json:"concurrency"`
	SkipVerify     *bool  `json:"skipVerify"`
}

type VerifyResult struct {
	Valid   bool          `json:"valid"`
	Version string        `json:"version"`
	Issues  []VerifyIssue `json:"issues"`
}

type VerifyIssue struct {
	Type    string `json:"type"`
	File    string `json:"file"`
	Message string `json:"message"`
}

type InstanceLaunchResult struct {
	ID      string `json:"id"`
	PID     int    `json:"pid"`
	Version string `json:"version"`
	Status  string `json:"status"`
	LogPath string `json:"logPath"`
}

type instanceDownload struct {
	ID           string
	InstanceName string
	Version      string
	DownloadID   string
	StartTime    time.Time
}
