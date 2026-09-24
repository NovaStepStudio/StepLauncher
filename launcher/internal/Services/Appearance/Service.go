package appearance

import (
	"StepLauncher/internal/Config"
	"StepLauncher/internal/Handlers"
	engine "StepLauncher/internal/Handlers/Engine"
	"errors"
	launcherassets "StepLauncher/internal/Core/Assets"
)

// AppearanceService gestiona personalizacion visual y assets.
type AppearanceService struct {
	handler *Handlers.App
	engine  *engine.Engine
}

func NewAppearanceService(handler *Handlers.App, eng *engine.Engine) *AppearanceService {
	return &AppearanceService{handler: handler, engine: eng}
}

type ScreenshotInfo = Handlers.ScreenshotInfo

func (s *AppearanceService) AddMusic(name, src string) (string, error) {
	if s.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return s.handler.AddMusic(name, src)
}

func (s *AppearanceService) DeleteFontFile(name string) error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.DeleteFontFile(name)
}

func (s *AppearanceService) DownloadGalleryImageAsBackground(url, author, modName, title string) (string, error) {
	if s.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return s.handler.DownloadGalleryImageAsBackground(url, author, modName, title)
}

func (s *AppearanceService) GetLauncherAssets() launcherassets.Assets {
	if s.handler == nil {
		return launcherassets.Default()
	}
	return s.handler.GetLauncherAssets()
}

func (s *AppearanceService) GetUIScale() int {
	return s.handler.GetUIScale()
}

func (s *AppearanceService) ImportBackground(src, kind string) (string, error) {
	return s.handler.ImportBackground(src, kind)
}

func (s *AppearanceService) ImportFont(src string) (string, error) {
	if s.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return s.handler.ImportFont(src)
}

func (s *AppearanceService) ListFontFiles() []string {
	if s.handler == nil {
		return []string{}
	}
	return s.handler.ListFontFiles()
}

func (s *AppearanceService) ListMusic() []launcherassets.MusicSlot {
	if s.handler == nil {
		return []launcherassets.MusicSlot{}
	}
	return s.handler.ListMusic()
}

func (s *AppearanceService) ListScreenshots() ([]ScreenshotInfo, error) {
	if s.handler == nil {
		return nil, errors.New("handler no disponible")
	}
	return s.handler.ListScreenshots()
}

func (s *AppearanceService) SetScreenshotAsBackground(relPath string) (string, error) {
	if s.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return s.handler.SetScreenshotAsBackground(relPath)
}

func (s *AppearanceService) PickBackgroundFile(kind string) (string, error) {
	return s.handler.PickBackgroundFile(kind)
}

func (s *AppearanceService) PickFontFile() (string, error) {
	if s.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return s.handler.PickFontFile()
}

func (s *AppearanceService) PickMusicFile() (string, error) {
	if s.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return s.handler.PickMusicFile()
}

func (s *AppearanceService) ReadMusicFile(src string) ([]byte, error) {
	if s.handler == nil {
		return nil, errors.New("handler no disponible")
	}
	return s.handler.ReadMusicFile(src)
}

func (s *AppearanceService) RemoveMusic(name string) error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.RemoveMusic(name)
}

func (s *AppearanceService) SaveLauncherAssets(asset launcherassets.Assets) {
	if s.handler != nil {
		s.handler.SaveLauncherAssets(asset)
	}
}

func (s *AppearanceService) SetHideLauncher(v bool) {
	s.handler.SetHideLauncher(v)
}

func (s *AppearanceService) SetIdle(idle Config.IdleConfig) {
	s.handler.SetIdle(idle)
}

func (s *AppearanceService) SetUIScale(percent int) {
	s.handler.SetUIScale(percent)
}

func (s *AppearanceService) UpdatePersonalization(p Config.Personalization) {
	s.handler.UpdatePersonalization(p)
}
