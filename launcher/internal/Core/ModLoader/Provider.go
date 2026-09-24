package modloader

import "net/http"

type ModLoaderProvider interface {
	Name() string

	// SetHTTPClient reemplaza el cliente HTTP del provider para que los
	// cambios de red (proxy, límite de velocidad) apliquen sin reiniciar.
	SetHTTPClient(client *http.Client)

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
