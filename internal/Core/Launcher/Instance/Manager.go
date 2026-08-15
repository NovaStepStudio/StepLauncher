package instance

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"unicode/utf8"

	"StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Core/Launcher"
	"StepLauncher/internal/Core/ModLoader"
)

const (
	metadataFile = "instance.metadata.json"
	configFile   = "instance.config.json"
	lockFile     = "instance.lock"
)

type InstanceManager struct {
	mu           sync.RWMutex
	instancesDir string
	sharedDir    string
	cacheDir     string

	downloads   map[string]*instanceDownload
	nextDLID    uint64
	sharedDlMgr *downloader.Manager

	launchManager  *launcher.LaunchManager
	mlOrchestrator *modloader.Orchestrator

	launcherName    string
	launcherVersion string
	logFn           func(string, ...interface{})

	separateGameDir bool

	onVersionReady func(name, version string)
}

func NewManager(instancesDir, sharedDir string) *InstanceManager {
	return &InstanceManager{
		instancesDir:    instancesDir,
		sharedDir:       sharedDir,
		separateGameDir: true,
		downloads:       make(map[string]*instanceDownload),
	}
}

// SetSeparateGameDir controla si el gameDir de las instancias es
// <instancia>/game (true) o la propia carpeta de la instancia (false).
func (m *InstanceManager) SetSeparateGameDir(v bool) {
	m.mu.Lock()
	m.separateGameDir = v
	m.mu.Unlock()
}

func (m *InstanceManager) SetSharedDownloadManager(dlMgr *downloader.Manager) {
	m.sharedDlMgr = dlMgr
}

// SetCacheDir fija el directorio de cache global del launcher (p. ej. <WorkDir>/cache).
// El cache nunca se comparte entre instancias ni se guarda en shared/.
func (m *InstanceManager) SetCacheDir(dir string) {
	m.cacheDir = dir
}

func (m *InstanceManager) SetLaunchManager(lm *launcher.LaunchManager) {
	m.launchManager = lm
}

func (m *InstanceManager) SetModLoaderOrchestrator(o *modloader.Orchestrator) {
	m.mlOrchestrator = o
}

func (m *InstanceManager) SetIdentity(name, version string) {
	m.launcherName = name
	m.launcherVersion = version
}

func (m *InstanceManager) SetLogger(fn func(string, ...interface{})) {
	m.logFn = fn
}

// SetOnVersionReady registra un callback que se invoca al terminar exitosamente
// la descarga de una version en una instancia (sin bloquear al llamador: va en goroutine).
func (m *InstanceManager) SetOnVersionReady(fn func(name, version string)) {
	m.onVersionReady = fn
}

func (m *InstanceManager) fireVersionReady(name, version string) {
	if m.onVersionReady != nil {
		m.onVersionReady(name, version)
	}
}

// mergeInstanceLibraries traslada las librerías que el instalador del
// modloader dejó en <instancia>/libraries a shared/libraries (donde el
// lanzamiento las lee) y elimina la carpeta de la instancia, que quedaría
// como peso muerto. Los archivos ya presentes en shared con el mismo tamaño
// se omiten; si algún archivo no se puede mover, no se borra nada.
func (m *InstanceManager) mergeInstanceLibraries(instPath string) {
	src := filepath.Join(instPath, "libraries")
	info, err := os.Stat(src)
	if err != nil || !info.IsDir() {
		return
	}
	dst := filepath.Join(m.sharedDir, "libraries")
	moved, skipped, failed := 0, 0, 0

	walkErr := filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(src, path)
		if relErr != nil {
			failed++
			return nil
		}
		target := filepath.Join(dst, rel)
		if tInfo, tErr := os.Stat(target); tErr == nil && tInfo.Size() == fi.Size() {
			skipped++
			return nil
		}
		if mkErr := os.MkdirAll(filepath.Dir(target), 0755); mkErr != nil {
			failed++
			return nil
		}
		// El destino existe con distinto tamaño: se reemplaza con el de la
		// instancia (la copia del instalador es la autoritativa).
		if _, tErr := os.Stat(target); tErr == nil {
			if rmErr := os.Remove(target); rmErr != nil {
				failed++
				return nil
			}
		}
		if mvErr := os.Rename(path, target); mvErr != nil {
			if cpErr := copyFile(path, target); cpErr != nil {
				failed++
				return nil
			}
			os.Remove(path)
		}
		moved++
		return nil
	})
	if walkErr != nil {
		m.log("WARN: merge instance libraries: %v", walkErr)
		return
	}

	if failed > 0 {
		m.log("WARN: merge instance libraries: %d movidos, %d omitidos, %d fallidos (carpeta conservada)", moved, skipped, failed)
		return
	}
	if err := os.RemoveAll(src); err != nil {
		m.log("WARN: no se pudo eliminar la carpeta libraries de la instancia: %v", err)
		return
	}
	if moved > 0 {
		m.log("Librerías de la instancia movidas a shared: %d movidas, %d ya existentes", moved, skipped)
	}
}

func (m *InstanceManager) log(format string, args ...interface{}) {
	if m.logFn != nil {
		m.logFn("[InstanceManager] "+format, args...)
	}
}

func (m *InstanceManager) Create(req CreateInstanceReq) (*InstanceMetadata, string, error) {
	if err := sanitizeInstanceName(req.Name); err != nil {
		return nil, "", fmt.Errorf("invalid instance name: %w", err)
	}
	if utf8.RuneCountInString(req.Title) > 64 {
		return nil, "", fmt.Errorf("el título no puede superar los 64 caracteres")
	}
	if utf8.RuneCountInString(req.Description) > 512 {
		return nil, "", fmt.Errorf("la descripción no puede superar los 512 caracteres")
	}

	instPath, err := m.instancePath(req.Name)
	if err != nil {
		return nil, "", err
	}
	if _, err := os.Stat(instPath); err == nil {
		return nil, "", fmt.Errorf("instance %s already exists", req.Name)
	}

	versionsDir := filepath.Join(instPath, "versions")
	if err := os.MkdirAll(versionsDir, 0755); err != nil {
		return nil, "", fmt.Errorf("create instance dir: %w", err)
	}

	now := time.Now().Format(time.RFC3339)
	meta := &InstanceMetadata{
		ID:          generateID(),
		Name:        req.Name,
		Title:       req.Title,
		Description: req.Description,
		Icon:        req.Icon,
		Banner:      req.Banner,
		Group:       req.Group,
		Tags:        req.Tags,
		Favorite:    false,
		Pinned:      false,
		CreatedAt:   now,
		Versions:    []string{},
		ConfigPath:  configFile,
	}
	if req.Favorite != nil {
		meta.Favorite = *req.Favorite
	}
	if req.Pinned != nil {
		meta.Pinned = *req.Pinned
	}
	if meta.Title == "" {
		meta.Title = req.Name
	}

	if err := m.writeMetadata(req.Name, meta); err != nil {
		os.RemoveAll(instPath)
		return nil, "", err
	}

	dlVersion := req.Version
	if dlVersion == "" && req.LaunchConfig != nil {
		dlVersion = req.LaunchConfig.Version
	}

	if req.LaunchConfig != nil {
		if dlVersion != "" {
			req.LaunchConfig.Version = dlVersion
		}
		if err := m.writeConfig(req.Name, req.LaunchConfig); err != nil {
			os.RemoveAll(instPath)
			return nil, "", err
		}
	} else {
		cfg := &InstanceLaunchConfig{}
		if dlVersion != "" {
			cfg.Version = dlVersion
		}
		if err := m.writeConfig(req.Name, cfg); err != nil {
			os.RemoveAll(instPath)
			return nil, "", err
		}
	}

	m.log("Instance created: %s", req.Name)

	var downloadID string
	if dlVersion != "" {
		m.log("Starting auto-download for version %s in instance %s", dlVersion, req.Name)
		downloadID = m.downloadForInstance(req.Name, dlVersion)
	}

	return meta, downloadID, nil
}

func (m *InstanceManager) List() []*InstanceInfo {
	entries, err := os.ReadDir(m.instancesDir)
	if err != nil {
		return nil
	}
	var result []*InstanceInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		meta, err := m.readMetadata(e.Name())
		if err != nil {
			continue
		}
		result = append(result, &InstanceInfo{
			Name: meta.Name, Title: meta.Title, Versions: meta.Versions,
			Favorite: meta.Favorite, Pinned: meta.Pinned, Group: meta.Group,
			LastPlayed: meta.LastPlayed, PlayTime: meta.PlayTime,
		})
	}
	return result
}

func (m *InstanceManager) Get(name string) (*InstanceMetadata, *InstanceLaunchConfig, error) {
	meta, err := m.readMetadata(name)
	if err != nil {
		return nil, nil, err
	}
	cfg, err := m.readConfig(name)
	if err != nil {
		return nil, nil, fmt.Errorf("config not found for instance %s: %w", name, err)
	}
	return meta, cfg, nil
}

// OpenFolder abre la carpeta raíz de la instancia en el explorador de archivos.
func (m *InstanceManager) OpenFolder(name string) error {
	if err := sanitizeInstanceName(name); err != nil {
		return err
	}
	instPath, err := m.instancePath(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(instPath); err != nil {
		return fmt.Errorf("instancia no encontrada: %s", name)
	}
	return openInExplorer(instPath)
}

func (m *InstanceManager) Delete(name string) error {
	if err := sanitizeInstanceName(name); err != nil {
		return err
	}
	instPath, err := m.instancePath(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(instPath); err != nil {
		return nil
	}
	if err := m.acquireLock(name, "delete", ""); err != nil {
		return err
	}
	defer m.releaseLock(name)
	return os.RemoveAll(instPath)
}

func (m *InstanceManager) UpdateMetadata(name string, req UpdateMetadataReq) (*InstanceMetadata, error) {
	meta, err := m.readMetadata(name)
	if err != nil {
		return nil, err
	}
	if utf8.RuneCountInString(req.Title) > 64 {
		return nil, fmt.Errorf("el título no puede superar los 64 caracteres")
	}
	if utf8.RuneCountInString(req.Description) > 512 {
		return nil, fmt.Errorf("la descripción no puede superar los 512 caracteres")
	}
	if req.Title != "" {
		meta.Title = req.Title
	}
	if req.Description != "" {
		meta.Description = req.Description
	}
	if req.Icon != "" {
		meta.Icon = req.Icon
	}
	if req.Banner != "" {
		meta.Banner = req.Banner
	}
	if req.Group != "" {
		meta.Group = req.Group
	}
	if req.Tags != nil {
		meta.Tags = req.Tags
	}
	if req.Favorite != nil {
		meta.Favorite = *req.Favorite
	}
	if req.Pinned != nil {
		meta.Pinned = *req.Pinned
	}
	if err := m.writeMetadata(name, meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func (m *InstanceManager) UpdateConfig(name string, cfg *InstanceLaunchConfig) (*InstanceLaunchConfig, error) {
	existing, err := m.readConfig(name)
	if err != nil {
		return nil, fmt.Errorf("instance %s not found", name)
	}

	if cfg.Advanced != nil {
		existing.Advanced = cfg.Advanced
	}

	if cfg.Version != "" {
		existing.Version = cfg.Version
	}
	if cfg.JavaExec != "" {
		existing.JavaExec = cfg.JavaExec
	}
	if cfg.MinRAM != nil {
		existing.MinRAM = cfg.MinRAM
	}
	if cfg.MaxRAM != nil {
		existing.MaxRAM = cfg.MaxRAM
	}
	if cfg.UseOfficialJava != nil {
		existing.UseOfficialJava = cfg.UseOfficialJava
	}
	if cfg.Fullscreen != nil {
		existing.Fullscreen = cfg.Fullscreen
	}
	if cfg.HardwareAcceleration != nil {
		existing.HardwareAcceleration = cfg.HardwareAcceleration
	}
	if cfg.GCPreset != nil {
		existing.GCPreset = cfg.GCPreset
	}
	if cfg.GPUPreference != nil {
		existing.GPUPreference = cfg.GPUPreference
	}
	if cfg.CustomResolution != nil {
		existing.CustomResolution = cfg.CustomResolution
	}
	if cfg.ResWidth != nil {
		existing.ResWidth = cfg.ResWidth
	}
	if cfg.ResHeight != nil {
		existing.ResHeight = cfg.ResHeight
	}

	if err := m.writeConfig(name, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (m *InstanceManager) Versions(name string) ([]string, error) {
	meta, err := m.readMetadata(name)
	if err != nil {
		return nil, err
	}
	return meta.Versions, nil
}

func (m *InstanceManager) RemoveVersion(name, version string) error {
	meta, err := m.readMetadata(name)
	if err != nil {
		return err
	}
	found := false
	for i, v := range meta.Versions {
		if v == version {
			meta.Versions = append(meta.Versions[:i], meta.Versions[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("version %s not found in instance %s", version, name)
	}
	verDir, err := m.instancePath(name)
	if err != nil {
		return err
	}
	verDir = filepath.Join(verDir, "versions", version)
	os.RemoveAll(verDir)
	return m.writeMetadata(name, meta)
}

func (m *InstanceManager) Clone(name, newName string, copyVersions bool) (*InstanceMetadata, error) {
	meta, cfg, err := m.Get(name)
	if err != nil {
		return nil, fmt.Errorf("source instance not found: %w", err)
	}
	createReq := CreateInstanceReq{
		Name: newName, Title: meta.Title + " (copy)", Description: meta.Description,
		Icon: meta.Icon, Banner: meta.Banner, Group: meta.Group, Tags: meta.Tags,
		LaunchConfig: cfg,
	}
	newMeta, _, err := m.Create(createReq)
	if err != nil {
		return nil, err
	}
	if copyVersions {
		srcPath, err := m.instancePath(name)
		if err != nil {
			return nil, err
		}
		dstPath, err := m.instancePath(newName)
		if err != nil {
			return nil, err
		}
		src := filepath.Join(srcPath, "versions")
		dst := filepath.Join(dstPath, "versions")
		for _, ver := range meta.Versions {
			if err := copyDir(filepath.Join(src, ver), filepath.Join(dst, ver)); err != nil {
				m.log("WARN: failed to copy version %s: %v", ver, err)
				continue
			}
			newMeta.Versions = append(newMeta.Versions, ver)
		}
		m.writeMetadata(newName, newMeta)
	}
	return newMeta, nil
}

func (m *InstanceManager) DownloadStatus(dlID string) (*downloader.DownloadInfo, error) {
	if m.sharedDlMgr == nil {
		return nil, fmt.Errorf("download manager not available")
	}
	info := m.sharedDlMgr.GetInfo(dlID)
	if info == nil {
		return nil, fmt.Errorf("download %s not found", dlID)
	}
	return info, nil
}

func (m *InstanceManager) CancelDownload(dlID string) error {
	if m.sharedDlMgr == nil {
		return fmt.Errorf("download manager not available")
	}
	return m.sharedDlMgr.Cancel(dlID)
}

func (m *InstanceManager) downloadForInstance(instName, version string) string {
	if m.sharedDlMgr == nil {
		m.log("WARN: download manager not set, cannot auto-download %s", version)
		return ""
	}

	instPath, err := m.instancePath(instName)
	if err != nil {
		m.log("ERROR: instance path: %v", err)
		return ""
	}
	instVerDir := filepath.Join(instPath, "versions", version)

	filter := downloader.DownloadFilter{
		Version:            version,
		Client:             true,
		Libraries:          true,
		Natives:            true,
		Assets:             true,
		Java:               true,
		InstanceVersionDir: instVerDir,
	}

	dl := m.sharedDlMgr.Start(version, filter, 0, 0, false, 0, 0)
	dlID := dl.ID

	m.mu.Lock()
	m.nextDLID++
	id := fmt.Sprintf("inst-dl-%d", m.nextDLID)
	m.downloads[id] = &instanceDownload{
		ID:           id,
		InstanceName: instName,
		Version:      version,
		DownloadID:   dlID,
		StartTime:    time.Now(),
	}
	m.mu.Unlock()

	go func() {
		dl.Wait()
		info := m.sharedDlMgr.GetInfo(dlID)

		m.mu.Lock()
		delete(m.downloads, id)
		m.mu.Unlock()

		if info == nil || info.State != downloader.StateCompleted {
			errStr := "unknown"
			if info != nil {
				errStr = info.Error
			}
			m.log("Auto-download failed for %s: %s", version, errStr)
			return
		}

		if err := m.addVersionToMetadata(instName, version); err != nil {
			m.log("ERROR: add version to metadata: %v", err)
		} else {
			m.log("Instance %s: version %s ready", instName, version)
			m.fireVersionReady(instName, version)
		}
	}()

	return dlID
}
