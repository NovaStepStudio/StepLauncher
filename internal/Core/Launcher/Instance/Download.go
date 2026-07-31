package instance

import (
	"fmt"
	"path/filepath"
	"time"

	"StepLauncher/internal/Core/Downloader"
)

func (m *InstanceManager) AddVersion(name string, req AddVersionReq) (*downloader.Download, error) {
	if req.Version == "" {
		return nil, fmt.Errorf("version is required")
	}
	if _, err := m.readMetadata(name); err != nil {
		return nil, fmt.Errorf("instance %s not found", name)
	}
	if m.sharedDlMgr == nil {
		return nil, fmt.Errorf("download manager not available")
	}

	filter := buildDownloadFilter(req.Version, req)

	instPath, err := m.instancePath(name)
	if err != nil {
		return nil, fmt.Errorf("instance path: %w", err)
	}
	filter.InstanceVersionDir = filepath.Join(instPath, "versions", req.Version)

	maxRetries := 3
	if req.MaxRetries != nil {
		maxRetries = *req.MaxRetries
	}
	maxConcurrency := 24
	if req.MaxConcurrency != nil {
		maxConcurrency = *req.MaxConcurrency
	}
	skipVerify := false
	if req.SkipVerify != nil {
		skipVerify = *req.SkipVerify
	}

	dl := m.sharedDlMgr.Start(req.Version, filter, maxRetries, maxConcurrency, skipVerify, 0, 0)

	m.mu.Lock()
	m.nextDLID++
	id := fmt.Sprintf("inst-dl-%d", m.nextDLID)
	m.downloads[id] = &instanceDownload{
		ID:           id,
		InstanceName: name,
		Version:      req.Version,
		DownloadID:   dl.ID,
		StartTime:    time.Now(),
	}
	m.mu.Unlock()

	go func() {
		dl.Wait()
		info := m.sharedDlMgr.GetInfo(dl.ID)

		m.mu.Lock()
		delete(m.downloads, id)
		m.mu.Unlock()

		if info != nil && info.State == downloader.StateCompleted {
			if err := m.addVersionToMetadata(name, req.Version); err != nil {
				m.log("ERROR: add version to metadata: %v", err)
			} else {
				m.log("Instance %s: version %s ready", name, req.Version)
			}
		} else {
			errStr := "unknown"
			if info != nil {
				errStr = info.Error
			}
			m.log("Download failed for %s: %s", req.Version, errStr)
		}
	}()

	return dl, nil
}
