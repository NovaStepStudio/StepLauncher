package modloader

import "time"

type LoaderVersion struct {
	LoaderVersion    string `json:"loaderVersion"`
	MinecraftVersion string `json:"minecraftVersion"`
	Stable           bool   `json:"stable"`
}

type InstalledLoader struct {
	LoaderType       string `json:"loaderType"`
	LoaderVersion    string `json:"loaderVersion"`
	MinecraftVersion string `json:"minecraftVersion"`
	VersionJsonID    string `json:"versionJsonId"`
	InstallerJarPath string `json:"installerJarPath,omitempty"`
	InstalledAt      int64  `json:"installedAt"`
}

type DownloadPlanEntry struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Destination string `json:"destination"`
	Size        int64  `json:"size"`
	SHA1        string `json:"sha1"`
	Category    string `json:"category"`
}

type DownloadPlan struct {
	Entries           []DownloadPlanEntry
	RequiresInstaller bool
	InstallerDest     string
	MinecraftJar      string
}

type ExecutionPlan struct {
	MainClass           string   `json:"mainClass"`
	AdditionalClasspath []string `json:"additionalClasspath"`
	AdditionalJVMArgs   []string `json:"additionalJvmArgs"`
	AdditionalGameArgs  []string `json:"additionalGameArgs"`
	UseModulePath       bool     `json:"useModulePath"`
}

type InstallState struct {
	Loader *InstalledLoader `json:"loader"`
}

func LoaderStatePath(instancePath string) string {
	return instancePath + "/loader-state.json"
}

func (v LoaderVersion) String() string {
	return v.LoaderVersion
}

func NewInstalledLoader(loaderType, loaderVersion, mcVersion, versionJsonID, installerJarPath string) *InstalledLoader {
	return &InstalledLoader{
		LoaderType:       loaderType,
		LoaderVersion:    loaderVersion,
		MinecraftVersion: mcVersion,
		VersionJsonID:    versionJsonID,
		InstallerJarPath: installerJarPath,
		InstalledAt:      time.Now().UnixMilli(),
	}
}
