package modloader

type ModLoaderProvider interface {
	Name() string

	GetVersions(minecraftVersion string) ([]LoaderVersion, error)

	ResolveDownload(mcVersion, loaderVersion, instancePath, librariesPath, modloaderCacheDir string) (*DownloadPlan, error)

	RequiresInstallerRun() bool

	RunInstaller(sessionId string, plan *DownloadPlan, mcVersion, loaderVersion, instancePath, librariesPath, minecraftJar string, broadcast func([]byte)) error

	BuildExecution(loader *InstalledLoader, versionsDir, librariesPath string) (*ExecutionPlan, error)

	VersionJsonID(mcVersion, loaderVersion string) string
}

type BaseProvider struct{}

func (BaseProvider) RequiresInstallerRun() bool { return false }

func (BaseProvider) RunInstaller(sessionId string, plan *DownloadPlan, mcVersion, loaderVersion, instancePath, librariesPath, minecraftJar string, broadcast func([]byte)) error {
	return nil
}
