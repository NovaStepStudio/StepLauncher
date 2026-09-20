package download

import (
	downloader "StepLauncher/internal/Core/Downloader"
	engine "StepLauncher/internal/Handlers/Engine"
	"errors"
)

// DownloadService gestiona descargas y manifests.
type DownloadService struct {
	engine *engine.Engine
}

func NewDownloadService(eng *engine.Engine) *DownloadService {
	return &DownloadService{engine: eng}
}

func (s *DownloadService) CancelDownload(id string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.CancelDownload(id)
}

func (s *DownloadService) FetchVersionManifest() (*downloader.Manifest, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.FetchVersionManifest()
}

func (s *DownloadService) GetDownload(id string) *engine.DownloadInfo {
	if s.engine == nil {
		return nil
	}
	return s.engine.GetDownload(id)
}

func (s *DownloadService) GetDownloadStatus(id string) (*engine.DownloadProgress, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.GetDownloadStatus(id)
}

func (s *DownloadService) GetVersions(versionType string) ([]engine.VersionInfo, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.GetVersions(versionType)
}

func (s *DownloadService) ListDownloadedVersions() []engine.InstalledVersion {
	if s.engine == nil {
		return []engine.InstalledVersion{}
	}
	return s.engine.ListDownloadedVersions()
}

func (s *DownloadService) ListDownloads() []*engine.DownloadInfo {
	if s.engine == nil {
		return []*engine.DownloadInfo{}
	}
	return s.engine.ListDownloads()
}

func (s *DownloadService) PauseDownload(id string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.PauseDownload(id)
}

func (s *DownloadService) RefreshManifests() (int, error) {
	if s.engine == nil {
		return 0, nil
	}
	return s.engine.RefreshManifests()
}

func (s *DownloadService) ResumeDownload(id string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.ResumeDownload(id)
}

func (s *DownloadService) StartDownload(version string, filter engine.DownloadFilter, maxRetries, maxConcurrency int, skipVerify bool, stallTimeoutMs, maxStallRetries int) *engine.DownloadInfo {
	if s.engine == nil {
		return nil
	}
	return s.engine.StartDownload(version, filter, maxRetries, maxConcurrency, skipVerify, stallTimeoutMs, maxStallRetries)
}

func (s *DownloadService) StartFullDownload(version string) *engine.DownloadInfo {
	if s.engine == nil {
		return nil
	}
	return s.engine.StartFullDownload(version)
}
