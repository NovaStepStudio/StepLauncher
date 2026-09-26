package mods

import (
	engine "StepLauncher/internal/Handlers/Engine"
	coremods "StepLauncher/internal/Core/Mods"
	"errors"
)

// ModsService instala contenido de Modrinth (mods, shaders, texturas y
// modpacks .mrpack) en el juego global o en instancias.
type ModsService struct {
	engine *engine.Engine
}

func NewModsService(eng *engine.Engine) *ModsService {
	return &ModsService{engine: eng}
}

func (s *ModsService) GlobalGameDir() string {
	if s.engine == nil {
		return ""
	}
	return s.engine.GlobalGameDir()
}

func (s *ModsService) InstanceGameDir(name string) (string, error) {
	if s.engine == nil {
		return "", errors.New("engine no disponible")
	}
	return s.engine.InstanceGameDir(name)
}

func (s *ModsService) ContentSubdirFor(projectType string) (string, error) {
	if s.engine == nil {
		return "", errors.New("engine no disponible")
	}
	return s.engine.ContentSubdirFor(projectType)
}

func (s *ModsService) InstallModContent(req engine.ModContentRequest) (*engine.ModContentResult, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.InstallModContent(req)
}

func (s *ModsService) InstallModpack(req engine.ModpackInstallRequest) (*engine.ModContentResult, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.InstallModpack(req)
}

func (s *ModsService) CancelModContent(sessionID string) bool {
	if s.engine == nil {
		return false
	}
	return s.engine.CancelModContent(sessionID)
}

func (s *ModsService) ListInstalledContent(destination, instance string) ([]coremods.InstalledFile, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.ListInstalledContent(destination, instance)
}

func (s *ModsService) SetInstalledContentEnabled(destination, instance, kind, name string, enabled bool) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.SetInstalledContentEnabled(destination, instance, kind, name, enabled)
}

func (s *ModsService) DeleteInstalledContent(destination, instance, kind, name string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.DeleteInstalledContent(destination, instance, kind, name)
}

// MoveInstalledContent traslada un archivo instalado entre destinos.
func (s *ModsService) MoveInstalledContent(srcDestination, srcInstance, kind, name, dstDestination, dstInstance string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.MoveInstalledContent(srcDestination, srcInstance, kind, name, dstDestination, dstInstance)
}

// RevealInstalledContent abre el explorador con el archivo instalado.
func (s *ModsService) RevealInstalledContent(destination, instance, kind, name string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.RevealInstalledContent(destination, instance, kind, name)
}

// GetContentIcon devuelve la ruta relativa del icono de un contenido
// instalado (vacío si no tiene). La UI la resuelve con loadLocal.
func (s *ModsService) GetContentIcon(destination, instance, kind, name string) string {
	if s.engine == nil {
		return ""
	}
	return s.engine.GetContentIcon(destination, instance, kind, name)
}
