package engine

import (
	"fmt"
	"time"

	"StepLauncher/internal/Core/ModLoader"
)

type ModLoaderVersion = modloader.LoaderVersion
type InstalledLoader = modloader.InstalledLoader
type ExecutionPlan = modloader.ExecutionPlan

func (e *Engine) ListModLoaders() []string {
	return e.modloader.Registry().List()
}

func (e *Engine) GetModLoaderVersions(loader, mcVersion string) ([]ModLoaderVersion, error) {
	return e.modloader.GetVersions(loader, mcVersion)
}

func (e *Engine) ResolveModLoaderVersion(loader, mcVersion, strategy string) (string, error) {
	return e.modloader.ResolveVersion(loader, mcVersion, strategy)
}

type ModLoaderInstallResult struct {
	SessionID string `json:"sessionId"`
	Status    string `json:"status"`
}

func (e *Engine) InstallModLoader(loader, loaderVersion, mcVersion, instancePath string) (*ModLoaderInstallResult, error) {
	sessionID := fmt.Sprintf("ml-%d", time.Now().UnixMilli())

	targetDir := e.config.Get().WorkDir

	go func() {
		_, err := e.modloader.Install(sessionID, loader, loaderVersion, mcVersion, targetDir)
		if err != nil {
			e.log.Error("Modloader install failed: %v", err)
			return
		}
		// Flujo clásico (WorkDir): el estado del loader no se lee en ningún
		// lado — la versión queda visible en versions/ y el lanzamiento la usa
		// directamente — así que al terminar la instalación se elimina el
		// loader-state.json para no dejar archivos sueltos en .StepLauncher.
		e.log.Info("[Modloader] Instalacion finalizada: limpiando loader-state.json de %s %s", loader, loaderVersion)
		if err := e.modloader.RemoveState(targetDir); err != nil {
			e.log.Error("Modloader state cleanup failed: %v", err)
		}
	}()

	return &ModLoaderInstallResult{
		SessionID: sessionID,
		Status:    "installing",
	}, nil
}

func (e *Engine) GetInstalledModLoader(instancePath string) (*InstalledLoader, error) {
	return e.modloader.GetInstalledLoader(instancePath)
}

func (e *Engine) RemoveModLoaderState(instancePath string) error {
	return e.modloader.RemoveState(instancePath)
}

func (e *Engine) BuildModLoaderExecution(instancePath, versionsDir, librariesPath string) (*ExecutionPlan, error) {
	return e.modloader.BuildExecution(instancePath, versionsDir, librariesPath)
}
