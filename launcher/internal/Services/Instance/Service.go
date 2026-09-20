package instance

import (
	"StepLauncher/internal/Handlers"
	engine "StepLauncher/internal/Handlers/Engine"
	"errors"
)

// InstanceService gestiona CRUD de instancias y verificacion.
type InstanceService struct {
	handler *Handlers.App
	engine  *engine.Engine
}

func NewInstanceService(handler *Handlers.App, eng *engine.Engine) *InstanceService {
	return &InstanceService{handler: handler, engine: eng}
}

type CreateInstanceResult struct {
	Metadata   *engine.InstanceMetadata `json:"metadata"`
	DownloadId string                   `json:"downloadId"`
}

type GetInstanceResult struct {
	Metadata *engine.InstanceMetadata     `json:"metadata"`
	Config   *engine.InstanceLaunchConfig `json:"config"`
}

type AddInstanceVersionResult struct {
	DownloadId string `json:"downloadId"`
	Version    string `json:"version"`
}

type InstanceDownloadStatusResult struct {
	Id      string `json:"id"`
	Version string `json:"version"`
	State   string `json:"state"`
}

type ScreenshotInfo = Handlers.ScreenshotInfo

func (s *InstanceService) AddInstanceVersion(name string, req engine.AddVersionReq) (*AddInstanceVersionResult, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	downloadID, version, err := s.engine.AddInstanceVersion(name, req)
	if err != nil {
		return nil, err
	}
	return &AddInstanceVersionResult{DownloadId: downloadID, Version: version}, nil
}

func (s *InstanceService) CancelInstanceDownload(dlID string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.CancelInstanceDownload(dlID)
}

func (s *InstanceService) CancelInstanceVerify(name string) {
	if s.engine != nil {
		s.engine.CancelInstanceVerify(name)
	}
}

func (s *InstanceService) CloneInstance(name, newName string, copyVersions bool) (*engine.InstanceMetadata, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.CloneInstance(name, newName, copyVersions)
}

func (s *InstanceService) CreateInstance(req engine.CreateInstanceReq) (*CreateInstanceResult, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	metadata, downloadID, err := s.engine.CreateInstance(req)
	if err != nil {
		return nil, err
	}
	return &CreateInstanceResult{Metadata: metadata, DownloadId: downloadID}, nil
}

func (s *InstanceService) CreateInstanceBackup(name string) (string, error) {
	if s.engine == nil {
		return "", errors.New("engine no disponible")
	}
	return s.engine.CreateInstanceBackup(name)
}

func (s *InstanceService) DeleteInstance(name string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.DeleteInstance(name)
}

func (s *InstanceService) GetInstance(name string) (*GetInstanceResult, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	metadata, cfg, err := s.engine.GetInstance(name)
	if err != nil {
		return nil, err
	}
	return &GetInstanceResult{Metadata: metadata, Config: cfg}, nil
}

func (s *InstanceService) GetInstanceDownloadStatus(dlID string) InstanceDownloadStatusResult {
	if s.engine == nil {
		return InstanceDownloadStatusResult{}
	}
	id, version, state, _ := s.engine.GetInstanceDownloadStatus(dlID)
	return InstanceDownloadStatusResult{Id: id, Version: version, State: state}
}

func (s *InstanceService) ImportInstanceAsset(name, kind, src string) (string, error) {
	if s.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return s.handler.ImportInstanceAsset(name, kind, src)
}

func (s *InstanceService) InstanceVerifyStatus(name string) engine.InstanceVerifyProgress {
	if s.engine == nil {
		return engine.InstanceVerifyProgress{State: "idle"}
	}
	p := s.engine.InstanceVerifyStatus(name)
	if p == nil {
		return engine.InstanceVerifyProgress{State: "idle"}
	}
	return *p
}

func (s *InstanceService) ListInstanceScreenshots(instanceName string) ([]ScreenshotInfo, error) {
	if s.handler == nil {
		return nil, errors.New("handler no disponible")
	}
	return s.handler.ListInstanceScreenshots(instanceName)
}

func (s *InstanceService) ListInstanceVersions(name string) ([]string, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.ListInstanceVersions(name)
}

func (s *InstanceService) ListInstances() []*engine.InstanceInfo {
	if s.engine == nil {
		return []*engine.InstanceInfo{}
	}
	return s.engine.ListInstances()
}

func (s *InstanceService) OpenInstanceFolder(name string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.OpenInstanceFolder(name)
}

func (s *InstanceService) PickInstanceAssetFile() (string, error) {
	if s.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return s.handler.PickInstanceAssetFile()
}

func (s *InstanceService) RemoveInstanceVersion(name, version string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.RemoveInstanceVersion(name, version)
}

func (s *InstanceService) StartInstanceVerify(name string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.StartInstanceVerify(name)
}

func (s *InstanceService) UpdateInstanceConfig(name string, cfg *engine.InstanceLaunchConfig) (*engine.InstanceLaunchConfig, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.UpdateInstanceConfig(name, cfg)
}

func (s *InstanceService) UpdateInstanceMetadata(name string, req engine.UpdateMetadataReq) (*engine.InstanceMetadata, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.UpdateInstanceMetadata(name, req)
}

func (s *InstanceService) VerifyInstance(name string) ([]engine.VerifyResult, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.VerifyInstance(name)
}

func (s *InstanceService) VerifyInstanceVersion(name, version string) (*engine.VerifyResult, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.VerifyInstanceVersion(name, version)
}
