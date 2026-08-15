package instance

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"StepLauncher/internal/Core/Launcher"
)

func (m *InstanceManager) LaunchInstance(name string, auth launcher.LaunchConfig) (*InstanceLaunchResult, error) {
	meta, cfg, err := m.Get(name)
	if err != nil {
		return nil, fmt.Errorf("instance %s not found", name)
	}

	if cfg.Version == "" {
		return nil, fmt.Errorf("no version configured for instance %s", name)
	}

	if err := m.acquireLock(name, "launch", cfg.Version); err != nil {
		return nil, err
	}
	defer m.releaseLock(name)

	instPath, err := m.instancePath(name)
	if err != nil {
		return nil, err
	}
	verDir := filepath.Join(instPath, "versions")
	verPath := filepath.Join(verDir, cfg.Version, cfg.Version+".json")
	if _, err := os.Stat(verPath); err != nil {
		return nil, fmt.Errorf("version %s not downloaded in instance %s (run version/add first)", cfg.Version, name)
	}

	// Las librerías que el instalador del modloader deja en la instancia se
	// mueven a shared (donde se leen en el lanzamiento) y se eliminan.
	m.mergeInstanceLibraries(instPath)

	var adv launcher.AdvancedConfig
	if cfg.Advanced != nil {
		adv = *cfg.Advanced
	}

	adv.RuntimeDir = nonEmpty(adv.RuntimeDir, cfg.RuntimeDir, filepath.Join(filepath.Dir(m.instancesDir), "runtime"))
	m.mu.RLock()
	sep := m.separateGameDir
	m.mu.RUnlock()
	if sep {
		adv.GameDir = nonEmpty(adv.GameDir, cfg.GameDir, filepath.Join(instPath, "game"))
	} else {
		adv.GameDir = nonEmpty(adv.GameDir, cfg.GameDir, instPath)
	}
	adv.AssetsDir = nonEmpty(adv.AssetsDir, cfg.AssetsDir, filepath.Join(m.sharedDir, "assets"))
	adv.LibrariesDir = nonEmpty(adv.LibrariesDir, cfg.LibrariesDir, filepath.Join(m.sharedDir, "libraries"))
	adv.VersionsDir = nonEmpty(adv.VersionsDir, cfg.VersionsDir, verDir)
	adv.NativesBaseDir = nonEmpty(adv.NativesBaseDir, "", verDir)
	if adv.CleanupDelay <= 0 {
		adv.CleanupDelay = 5 * time.Minute
	}

	if cfg.Advanced == nil {
		adv.JavaExec = nonEmpty(adv.JavaExec, cfg.JavaExec)
		adv.WindowTitle = nonEmpty(adv.WindowTitle, cfg.WindowTitle)
		adv.UserType = nonEmpty(adv.UserType, cfg.UserType, "mojang")
		if cfg.GCPreset != nil {
			adv.GCPreset = *cfg.GCPreset
		}
		if cfg.GPUPreference != nil {
			adv.GPUPreference = *cfg.GPUPreference
		}
		adv.PreLaunchCommand = nonEmpty(adv.PreLaunchCommand, cfg.PreLaunchCommand)
		adv.PostLaunchCommand = nonEmpty(adv.PostLaunchCommand, cfg.PostLaunchCommand)
		adv.AssetIndexID = nonEmpty(adv.AssetIndexID, cfg.AssetIndexID)
		adv.ServerAddress = nonEmpty(adv.ServerAddress, cfg.ServerAddress)
		adv.CacheDir = nonEmpty(adv.CacheDir, cfg.CacheDir)

		if cfg.UseOfficialJava != nil {
			adv.UseOfficialJava = *cfg.UseOfficialJava
		}
		if cfg.Fullscreen != nil {
			adv.Fullscreen = *cfg.Fullscreen
		}
		if cfg.CustomResolution != nil {
			adv.CustomResolution = *cfg.CustomResolution
		}
		if cfg.ResWidth != nil {
			adv.ResWidth = *cfg.ResWidth
		}
		if cfg.ResHeight != nil {
			adv.ResHeight = *cfg.ResHeight
		}
		if cfg.MinRAM != nil {
			adv.MinRAM = *cfg.MinRAM
		}
		if cfg.MaxRAM != nil {
			adv.MaxRAM = *cfg.MaxRAM
		}
		if cfg.HardwareAcceleration != nil {
			adv.HardwareAcceleration = cfg.HardwareAcceleration
		}
		if cfg.DemoUser != nil {
			adv.DemoUser = *cfg.DemoUser
		}
		if cfg.DisableAssets != nil {
			adv.DisableAssets = *cfg.DisableAssets
		}
		if cfg.ServerPort != nil {
			adv.ServerPort = *cfg.ServerPort
		}
		if cfg.MaxMetaspaceSize != nil {
			adv.MaxMetaspaceSize = *cfg.MaxMetaspaceSize
		}
		if cfg.StackSize != nil {
			adv.StackSize = *cfg.StackSize
		}
		if cfg.FramerateLimit != nil {
			adv.FramerateLimit = *cfg.FramerateLimit
		}
		if cfg.LogKeepDays != nil {
			adv.LogKeepDays = *cfg.LogKeepDays
		}
		if cfg.LogMaxFiles != nil {
			adv.LogMaxFiles = *cfg.LogMaxFiles
		}
		if cfg.SkipLibraryCheck != nil {
			adv.SkipLibraryCheck = *cfg.SkipLibraryCheck
		}
		if cfg.SkipAssetCheck != nil {
			adv.SkipAssetCheck = *cfg.SkipAssetCheck
		}
		if cfg.SkipNativeExtract != nil {
			adv.SkipNativeExtract = *cfg.SkipNativeExtract
		}
		if cfg.SkipVersionDownload != nil {
			adv.SkipVersionDownload = *cfg.SkipVersionDownload
		}

		if cfg.ServerAddress != "" {
			adv.ServerAddress = cfg.ServerAddress
		}

		adv.AllowServerList = firstNonNil(adv.AllowServerList, cfg.AllowServerList)
		adv.AllowMultiplayer = firstNonNil(adv.AllowMultiplayer, cfg.AllowMultiplayer)
		adv.AllowChat = firstNonNil(adv.AllowChat, cfg.AllowChat)
		adv.AllowRealms = firstNonNil(adv.AllowRealms, cfg.AllowRealms)

		if len(cfg.EnvironmentVars) > 0 {
			adv.EnvironmentVars = cfg.EnvironmentVars
		}
		if len(cfg.JavaArgs) > 0 {
			adv.JavaArgs = cfg.JavaArgs
		}
		if len(cfg.GameArgs) > 0 {
			adv.GameArgs = cfg.GameArgs
		}
	}

	launchCfg := launcher.LaunchConfig{
		Version:         cfg.Version,
		Username:        auth.Username,
		UUID:            auth.UUID,
		AccessToken:     auth.AccessToken,
		XUID:            auth.XUID,
		ClientID:        auth.ClientID,
		LauncherName:    m.launcherName,
		LauncherVersion: m.launcherVersion,
		InstanceName:    name,
		LogDir:          filepath.Join(instPath, "logs"),
		Advanced:        &adv,
	}

	if m.mlOrchestrator != nil {
		instPath, _ := m.instancePath(name)
		if state, err := m.mlOrchestrator.LoadState(instPath); err == nil && state != nil {
			m.log("Detected modloader: %s %s for MC %s", state.LoaderType, state.LoaderVersion, state.MinecraftVersion)
			plan, err := m.mlOrchestrator.BuildExecution(instPath, verDir, filepath.Join(m.sharedDir, "libraries"))
			if err == nil && plan != nil {
				adv.ExecutionPlan = plan
				adv.ProfileVersion = state.VersionJsonID
			}
		}
	}

	if m.launchManager == nil {
		return nil, fmt.Errorf("launch manager not available")
	}

	started := time.Now()
	instance, err := m.launchManager.Launch(launchCfg)
	if err != nil {
		return nil, err
	}

	go func() {
		<-instance.Done()
		now := time.Now()
		secs := int(now.Sub(started).Seconds())
		meta.PlayTime += int64(secs)
		if meta.PlayTime < 0 {
			meta.PlayTime = 0
		}
		meta.LastPlayed = now.Format(time.RFC3339)
		if err := m.writeMetadata(name, meta); err != nil {
			m.log("WARN: failed to persist playtime for %s: %v", name, err)
		}
		m.log("Instance %s stopped | PID: %d | exit: %d | playtime: %ds", name, instance.PID, instance.ExitCode, secs)
	}()

	return &InstanceLaunchResult{
		ID: instance.ID, PID: instance.PID,
		Version: instance.Version, Status: string(instance.GetStatus()),
		LogPath: instance.LogPath,
	}, nil
}
