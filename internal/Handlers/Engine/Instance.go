package engine

import (
	"StepLauncher/internal/Core/Launcher"
	linstance "StepLauncher/internal/Core/Launcher/Instance"
)

type InstanceMetadata = linstance.InstanceMetadata
type InstanceLaunchConfig = linstance.InstanceLaunchConfig
type InstanceInfo = linstance.InstanceInfo
type CreateInstanceReq = linstance.CreateInstanceReq
type UpdateMetadataReq = linstance.UpdateMetadataReq
type CloneInstanceReq = linstance.CloneInstanceReq
type AddVersionReq = linstance.AddVersionReq
type VerifyResult = linstance.VerifyResult
type VerifyIssue = linstance.VerifyIssue
type InstanceLaunchResult = linstance.InstanceLaunchResult

func (e *Engine) CreateInstance(req CreateInstanceReq) (*InstanceMetadata, string, error) {
	return e.instances.Create(req)
}

func (e *Engine) ListInstances() []*InstanceInfo {
	return e.instances.List()
}

func (e *Engine) GetInstance(name string) (*InstanceMetadata, *InstanceLaunchConfig, error) {
	return e.instances.Get(name)
}

func (e *Engine) DeleteInstance(name string) error {
	return e.instances.Delete(name)
}

func (e *Engine) OpenInstanceFolder(name string) error {
	return e.instances.OpenFolder(name)
}

func (e *Engine) UpdateInstanceMetadata(name string, req UpdateMetadataReq) (*InstanceMetadata, error) {
	return e.instances.UpdateMetadata(name, req)
}

func (e *Engine) UpdateInstanceConfig(name string, cfg *InstanceLaunchConfig) (*InstanceLaunchConfig, error) {
	return e.instances.UpdateConfig(name, cfg)
}

func (e *Engine) AddInstanceVersion(name string, req AddVersionReq) (string, string, error) {
	dl, err := e.instances.AddVersion(name, req)
	if err != nil {
		return "", "", err
	}
	return dl.ID, dl.Version, nil
}

func (e *Engine) ListInstanceVersions(name string) ([]string, error) {
	return e.instances.Versions(name)
}

func (e *Engine) RemoveInstanceVersion(name, version string) error {
	return e.instances.RemoveVersion(name, version)
}

func (e *Engine) VerifyInstance(name string) ([]VerifyResult, error) {
	return e.instances.VerifyInstance(name)
}

func (e *Engine) VerifyInstanceVersion(name, version string) (*VerifyResult, error) {
	return e.instances.VerifySingleVersion(name, version)
}

func (e *Engine) CloneInstance(name, newName string, copyVersions bool) (*InstanceMetadata, error) {
	return e.instances.Clone(name, newName, copyVersions)
}

func (e *Engine) LaunchInstance(name string, username, uuid, accessToken, xuid, clientID string) (*InstanceLaunchResult, error) {
	cfg := launcher.LaunchConfig{
		Username: username, UUID: uuid, AccessToken: accessToken,
		XUID: xuid, ClientID: clientID,
	}
	if err := e.fillAccountCredentials(&cfg, ""); err != nil {
		return nil, err
	}
	res, err := e.instances.LaunchInstance(name, cfg)
	if err != nil {
		return nil, err
	}
	if inst := e.launcher.Get(res.ID); inst != nil {
		inst.LauncherLogPath = e.log.GetLogPath()
	}
	return res, nil
}

func (e *Engine) GetInstanceDownloadStatus(dlID string) (string, string, string, string) {
	info, err := e.instances.DownloadStatus(dlID)
	if err != nil {
		return "", "", "", err.Error()
	}
	return info.ID, info.Version, string(info.State), info.Error
}

func (e *Engine) CancelInstanceDownload(dlID string) error {
	return e.instances.CancelDownload(dlID)
}

func (e *Engine) InstallInstanceModLoader(name, loader, loaderVersion, mcVersion string) (*ModLoaderInstallResult, error) {
	sessionID, err := e.instances.InstallModLoader(name, loader, loaderVersion, mcVersion)
	if err != nil {
		return nil, err
	}
	return &ModLoaderInstallResult{SessionID: sessionID, Status: "installing"}, nil
}

func (e *Engine) GetInstalledInstanceModLoader(name string) (*InstalledLoader, error) {
	return e.instances.InstalledModLoader(name)
}

func (e *Engine) RemoveInstanceModLoaderState(name string) error {
	return e.instances.RemoveModLoaderState(name)
}
