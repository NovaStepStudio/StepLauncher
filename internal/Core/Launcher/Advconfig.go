package launcher

import (
	"time"

	"StepLauncher/internal/Core/ModLoader"
)

type AuthLibConfig struct {
	Enabled          bool   `json:"enabled"`
	InjectorPath     string `json:"injectorPath,omitempty"`
	AuthServerURL    string `json:"authServerUrl,omitempty"`
	PreVerifyServer  bool   `json:"preVerifyServer"`
	PreVerifyTimeout int    `json:"preVerifyTimeout,omitempty"`
	SkipServerCheck  bool   `json:"skipServerCheck"`
	Username         string `json:"username,omitempty"`
	ServerToken      string `json:"serverToken,omitempty"`
}

type AdvancedConfig struct {
	AuthLibConfig AuthLibConfig `json:"authlib,omitempty"`

	JavaExec             string `json:"javaExec,omitempty"`
	UseOfficialJava      bool   `json:"useOfficialJava"`
	UseSystemJava        bool   `json:"useSystemJava"`
	MinRAM               int    `json:"minRam"`
	MaxRAM               int    `json:"maxRam"`
	GCPreset             string `json:"gcPreset,omitempty"`
	GPUPreference        string `json:"gpuPreference,omitempty"`
	HardwareAcceleration *bool  `json:"hardwareAcceleration,omitempty"`
	Fullscreen           bool   `json:"fullscreen"`
	CustomResolution     bool   `json:"customResolution"`
	ResWidth             int    `json:"resWidth"`
	ResHeight            int    `json:"resHeight"`
	WindowTitle          string `json:"windowTitle,omitempty"`
	DemoUser             bool   `json:"demoUser"`
	UserType             string `json:"userType,omitempty"`
	LogLevel             string `json:"logLevel,omitempty"`

	RuntimeDir     string `json:"runtimeDir,omitempty"`
	GameDir        string `json:"gameDir,omitempty"`
	AssetsDir      string `json:"assetsDir,omitempty"`
	LibrariesDir   string `json:"librariesDir,omitempty"`
	VersionsDir    string `json:"versionsDir,omitempty"`
	NativesBaseDir string `json:"nativesBaseDir,omitempty"`
	NativesDir     string `json:"nativesDir,omitempty"`
	CacheDir       string `json:"cacheDir,omitempty"`
	WorkingDir     string `json:"workingDir,omitempty"`

	AssetIndexID      string `json:"assetIndexId,omitempty"`
	AssetIndexType    string `json:"assetIndexType,omitempty"`
	AssetIndexURL     string `json:"assetIndexUrl,omitempty"`
	AssetIndexVirtual *bool  `json:"assetIndexVirtual,omitempty"`
	DisableAssets     bool   `json:"disableAssets"`
	ForceReindex      bool   `json:"forceReindex"`

	ServerAddress string `json:"serverAddress,omitempty"`
	ServerPort    int    `json:"serverPort"`

	QuickPlayPath      string `json:"quickPlayPath,omitempty"`
	MinecraftLogConfig string `json:"minecraftLogConfig,omitempty"`
	MaxMetaspaceSize   int    `json:"maxMetaspaceSize,omitempty"`
	StackSize          int    `json:"stackSize,omitempty"`

	JavaArgs []string `json:"javaArgs,omitempty"`
	GameArgs []string `json:"gameArgs,omitempty"`

	DisableLibraries    bool `json:"disableLibraries"`
	SkipLibraryCheck    bool `json:"skipLibraryCheck"`
	SkipAssetCheck      bool `json:"skipAssetCheck"`
	SkipNativeExtract   bool `json:"skipNativeExtract"`
	SkipVersionDownload bool `json:"skipVersionDownload"`

	ExecutionPlan  *modloader.ExecutionPlan `json:"-"`
	ProfileVersion string                   `json:"profileVersion,omitempty"`
	BaseVersion    string                   `json:"baseVersion,omitempty"`

	EnvironmentVars map[string]string `json:"environmentVars,omitempty"`
	JVMFlags        map[string]bool   `json:"jvmFlags,omitempty"`
	CleanupDelay    time.Duration     `json:"-"`

	DownloadConcurrency int    `json:"downloadConcurrency,omitempty"`
	MaxRetries          int    `json:"maxRetries,omitempty"`
	ConnectionTimeout   int    `json:"connectionTimeout,omitempty"`
	ProxyHost           string `json:"proxyHost,omitempty"`
	ProxyPort           int    `json:"proxyPort,omitempty"`
	ProxyUser           string `json:"proxyUser,omitempty"`
	ProxyPass           string `json:"proxyPass,omitempty"`
	LibraryCustomRepo   string `json:"libraryCustomRepo,omitempty"`
	AssetCustomURL      string `json:"assetCustomUrl,omitempty"`
	ForceRedownload     bool   `json:"forceRedownload"`

	JavaModulePath string   `json:"javaModulePath,omitempty"`
	JavaAddModules []string `json:"javaAddModules,omitempty"`
	JavaAddExports []string `json:"javaAddExports,omitempty"`
	JavaAddOpens   []string `json:"javaAddOpens,omitempty"`

	DirectMemorySize  int `json:"directMemorySize,omitempty"`
	ReservedCodeCache int `json:"reservedCodeCache,omitempty"`
	MetaspaceSize     int `json:"metaspaceSize,omitempty"`

	AllowServerList  *bool  `json:"allowServerList,omitempty"`
	AllowMultiplayer *bool  `json:"allowMultiplayer,omitempty"`
	AllowChat        *bool  `json:"allowChat,omitempty"`
	AllowRealms      *bool  `json:"allowRealms,omitempty"`
	FramerateLimit   int    `json:"framerateLimit,omitempty"`
	Renderer         string `json:"renderer,omitempty"`

	PreLaunchCommand  string `json:"preLaunchCommand,omitempty"`
	PostLaunchCommand string `json:"postLaunchCommand,omitempty"`

	LogKeepDays  int `json:"logKeepDays,omitempty"`
	LogMaxFiles  int `json:"logMaxFiles,omitempty"`
	GameLogLines int `json:"gameLogLines,omitempty"`

	LibraryExcludePatterns []string `json:"libraryExcludePatterns,omitempty"`
}

func DefaultAdvancedConfig() AdvancedConfig {
	return AdvancedConfig{
		MinRAM:              512,
		MaxRAM:              2048,
		UserType:            "mojang",
		CleanupDelay:        5 * time.Minute,
		DownloadConcurrency: 4,
		MaxRetries:          3,
		ConnectionTimeout:   30,
		GameLogLines:        1500,
	}
}
