package engine

import (
	"StepLauncher/internal/Core/Downloader"
)

type DownloadFilter = downloader.DownloadFilter
type DownloadProgress = downloader.DownloadProgress
type DownloadState = downloader.DownloadState
type Download = downloader.Download

const (
	StatePending     = downloader.StatePending
	StateDownloading = downloader.StateDownloading
	StatePaused      = downloader.StatePaused
	StateCancelled   = downloader.StateCancelled
	StateCompleted   = downloader.StateCompleted
	StateError       = downloader.StateError
	StateVerifying   = downloader.StateVerifying
	StateReDownload  = downloader.StateReDownload
)

type DownloadInfo struct {
	ID      string        `json:"id"`
	Version string        `json:"version"`
	State   DownloadState `json:"state"`
	Error   string        `json:"error,omitempty"`
}

func (e *Engine) StartDownload(version string, filter DownloadFilter, maxRetries, maxConcurrency int, skipVerify bool, stallTimeoutMs, maxStallRetries int) *DownloadInfo {
	dl := e.downloader.Start(version, filter, maxRetries, maxConcurrency, skipVerify, stallTimeoutMs, maxStallRetries)
	return &DownloadInfo{ID: dl.ID, Version: dl.Version, State: dl.State, Error: dl.Error}
}

func (e *Engine) StartFullDownload(version string) *DownloadInfo {
	filter := downloader.DownloadFilter{
		Version:   version,
		Client:    true,
		Libraries: true,
		Natives:   true,
		Assets:    true,
		Java:      true,
	}
	skipVerify := !e.GetConfig().VerifyIntegrity
	return e.StartDownload(version, filter, 0, 0, skipVerify, 0, 0)
}

func (e *Engine) GetDownloadStatus(id string) (*DownloadProgress, error) {
	return e.downloader.Status(id)
}

func (e *Engine) PauseDownload(id string) error {
	return e.downloader.Pause(id)
}

func (e *Engine) ResumeDownload(id string) error {
	return e.downloader.Resume(id)
}

func (e *Engine) CancelDownload(id string) error {
	return e.downloader.Cancel(id)
}

func (e *Engine) ListDownloads() []*DownloadInfo {
	list := e.downloader.List()
	infos := make([]*DownloadInfo, 0, len(list))
	for _, dl := range list {
		infos = append(infos, &DownloadInfo{
			ID: dl.ID, Version: dl.Version,
			State: dl.State, Error: dl.Error,
		})
	}
	return infos
}

func (e *Engine) GetDownload(id string) *DownloadInfo {
	info := e.downloader.GetInfo(id)
	if info == nil {
		return nil
	}
	return &DownloadInfo{
		ID: info.ID, Version: info.Version,
		State: info.State, Error: info.Error,
	}
}
