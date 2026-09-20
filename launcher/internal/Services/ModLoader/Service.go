package modloader

import (
	engine "StepLauncher/internal/Handlers/Engine"
	"errors"
)

// ModLoaderService gestiona instalacion y resolucion de mod loaders.
type ModLoaderService struct {
	engine *engine.Engine
}

func NewModLoaderService(eng *engine.Engine) *ModLoaderService {
	return &ModLoaderService{engine: eng}
}

func (s *ModLoaderService) BuildModLoaderExecution(instancePath, versionsDir, librariesPath string) (*engine.ExecutionPlan, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.BuildModLoaderExecution(instancePath, versionsDir, librariesPath)
}

func (s *ModLoaderService) GetInstalledInstanceModLoader(name string) (*engine.InstalledLoader, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.GetInstalledInstanceModLoader(name)
}

func (s *ModLoaderService) GetInstalledModLoader(instancePath string) (*engine.InstalledLoader, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.GetInstalledModLoader(instancePath)
}

func (s *ModLoaderService) GetModLoaderVersions(loader, mcVersion string) ([]engine.ModLoaderVersion, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.GetModLoaderVersions(loader, mcVersion)
}

func (s *ModLoaderService) InstallInstanceModLoader(name, loader, loaderVersion, mcVersion string) (*engine.ModLoaderInstallResult, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.InstallInstanceModLoader(name, loader, loaderVersion, mcVersion)
}

func (s *ModLoaderService) InstallModLoader(loader, loaderVersion, mcVersion, instancePath string) (*engine.ModLoaderInstallResult, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.InstallModLoader(loader, loaderVersion, mcVersion, instancePath)
}

func (s *ModLoaderService) ListModLoaders() []string {
	if s.engine == nil {
		return []string{}
	}
	return s.engine.ListModLoaders()
}

func (s *ModLoaderService) RemoveInstanceModLoaderState(name string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.RemoveInstanceModLoaderState(name)
}

func (s *ModLoaderService) RemoveModLoaderState(instancePath string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.RemoveModLoaderState(instancePath)
}

func (s *ModLoaderService) ResolveModLoaderVersion(loader, mcVersion, strategy string) (string, error) {
	if s.engine == nil {
		return "", errors.New("engine no disponible")
	}
	return s.engine.ResolveModLoaderVersion(loader, mcVersion, strategy)
}
