package modloader

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Core/Downloader/Utils"
)

type Orchestrator struct {
	registry   *Registry
	workDir    string
	sharedDir  string
	cacheDir   string
	httpClient *http.Client
	broadcast  func([]byte)
	logFn      func(string, ...interface{})
	mu         sync.RWMutex
}

func NewOrchestrator(workDir, sharedDir, cacheDir string, client *http.Client, reg *Registry, broadcast func([]byte), logFn func(string, ...interface{})) *Orchestrator {
	return &Orchestrator{
		registry:   reg,
		workDir:    workDir,
		sharedDir:  sharedDir,
		cacheDir:   cacheDir,
		httpClient: client,
		broadcast:  broadcast,
		logFn:      logFn,
	}
}

func (o *Orchestrator) Registry() *Registry { return o.registry }

func (o *Orchestrator) ModloaderCacheDir() string {
	return filepath.Join(o.cacheDir, "modloader")
}

func (o *Orchestrator) LibrariesPath(instancePath string) string {
	if o.sharedDir != "" {
		return filepath.Join(o.sharedDir, "libraries")
	}
	return filepath.Join(instancePath, "libraries")
}

func (o *Orchestrator) GetVersions(loaderName, mcVersion string) ([]LoaderVersion, error) {
	p, err := o.registry.Get(loaderName)
	if err != nil {
		return nil, err
	}
	return p.GetVersions(mcVersion)
}

func (o *Orchestrator) ResolveVersion(loaderName, mcVersion, strategy string) (string, error) {
	versions, err := o.GetVersions(loaderName, mcVersion)
	if err != nil {
		return "", fmt.Errorf("get versions: %w", err)
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found for %s %s", loaderName, mcVersion)
	}
	switch strategy {
	case "recommended":
		for _, v := range versions {
			if v.Stable {
				return v.LoaderVersion, nil
			}
		}
		return versions[0].LoaderVersion, nil
	case "latest":
		return versions[0].LoaderVersion, nil
	default:
		return "", fmt.Errorf("unknown strategy: %s", strategy)
	}
}

func (o *Orchestrator) Install(sessionId, loaderName, loaderVersion, mcVersion, instancePath string) (*InstalledLoader, error) {
	p, err := o.registry.Get(loaderName)
	if err != nil {
		return nil, err
	}

	if o.broadcast != nil {
		o.broadcast(ResolvingEvent(sessionId, loaderName, loaderVersion, mcVersion))
	}

	librariesPath := o.LibrariesPath(instancePath)
	modloaderCache := o.ModloaderCacheDir()

	plan, err := p.ResolveDownload(mcVersion, loaderVersion, instancePath, librariesPath, modloaderCache)
	if err != nil {
		return nil, fmt.Errorf("resolve download: %w", err)
	}

	if err := o.downloadEntries(sessionId, plan.Entries, loaderName); err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}

	if plan.RequiresInstaller {
		minecraftJar := plan.MinecraftJar
		if mcJar := filepath.Join(instancePath, "versions", mcVersion, mcVersion+".jar"); minecraftJar == "" {
			minecraftJar = mcJar
		}
		if err := p.RunInstaller(sessionId, plan, mcVersion, loaderVersion, instancePath, librariesPath, minecraftJar, o.broadcast); err != nil {
			return nil, fmt.Errorf("run installer: %w", err)
		}
	}

	versionJsonID := p.VersionJsonID(mcVersion, loaderVersion)
	installed := NewInstalledLoader(loaderName, loaderVersion, mcVersion, versionJsonID, plan.InstallerDest)

	if err := o.saveState(instancePath, installed); err != nil {
		return nil, fmt.Errorf("save state: %w", err)
	}

	if o.broadcast != nil {
		o.broadcast(InstalledEvent(sessionId, loaderName, loaderVersion, mcVersion))
	}

	return installed, nil
}

func (o *Orchestrator) downloadEntries(sessionId string, entries []DownloadPlanEntry, loaderName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	for i, entry := range entries {
		if utils.FileExists(entry.Destination) {
			if entry.SHA1 != "" {
				if ok, _ := utils.VerifySHA1(entry.Destination, entry.SHA1); ok {
					continue
				}
				os.Remove(entry.Destination)
			} else {
				continue
			}
		}

		if o.broadcast != nil {
			o.broadcast(DownloadingEvent(sessionId, loaderName, entry.Name, i+1, len(entries)))
		}

		task := downloader.DownloadTask{
			URL:  entry.URL,
			Dest: entry.Destination,
			SHA1: entry.SHA1,
			Size: entry.Size,
		}

		if err := downloader.DownloadFile(ctx, task, o.httpClient, 3, nil, 60000, 3); err != nil {
			return fmt.Errorf("download %s: %w", entry.Name, err)
		}
	}
	return nil
}

func (o *Orchestrator) saveState(instancePath string, loader *InstalledLoader) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	state := InstallState{Loader: loader}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(LoaderStatePath(instancePath), data, 0644)
}

func (o *Orchestrator) LoadState(instancePath string) (*InstalledLoader, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	statePath := LoaderStatePath(instancePath)
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, err
	}

	var state InstallState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return state.Loader, nil
}

func (o *Orchestrator) RemoveState(instancePath string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	statePath := LoaderStatePath(instancePath)
	if err := os.Remove(statePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (o *Orchestrator) GetInstalledLoader(instancePath string) (*InstalledLoader, error) {
	return o.LoadState(instancePath)
}

func (o *Orchestrator) BuildExecution(instancePath, versionsDir, librariesPath string) (*ExecutionPlan, error) {
	loader, err := o.LoadState(instancePath)
	if err != nil {
		return nil, err
	}

	p, err := o.registry.Get(loader.LoaderType)
	if err != nil {
		return nil, err
	}

	return p.BuildExecution(loader, versionsDir, librariesPath)
}
