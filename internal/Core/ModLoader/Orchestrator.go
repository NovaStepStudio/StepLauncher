package modloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Core/Downloader/Utils"
	"StepLauncher/internal/Core/ModLoader/Installer"
)

// El estado del modloader instalado vive en memoria (caché) y, si es necesario,
// se deriva del disco escaneando los version.json del destino: el usuario no
// quiere archivos de estado sueltos (carpeta loader-states/, loader-state.json).
type Orchestrator struct {
	registry     *Registry
	workDir      string
	cacheDir     string
	httpClient   *http.Client
	broadcast    func([]byte)
	logFn        func(string, ...interface{})
	mu           sync.RWMutex
	stateCache   map[string]*InstalledLoader
	stateRemoved map[string]bool
}

func NewOrchestrator(workDir, cacheDir string, client *http.Client, reg *Registry, broadcast func([]byte), logFn func(string, ...interface{})) *Orchestrator {
	// Limpieza de la persistencia antigua (carpeta loader-states/): el estado
	// ya no se guarda en archivos; se elimina lo que quedara de versiones
	// anteriores (best-effort, nunca condiciona el arranque).
	os.RemoveAll(filepath.Join(workDir, "loader-states"))
	return &Orchestrator{
		registry:     reg,
		workDir:      workDir,
		cacheDir:     cacheDir,
		httpClient:   client,
		broadcast:    broadcast,
		logFn:        logFn,
		stateCache:   make(map[string]*InstalledLoader),
		stateRemoved: make(map[string]bool),
	}
}

func (o *Orchestrator) Registry() *Registry { return o.registry }

func (o *Orchestrator) ModloaderCacheDir() string {
	return filepath.Join(o.cacheDir, "modloader")
}

func (o *Orchestrator) LibrariesPath(instancePath string) string {
	return filepath.Join(instancePath, "libraries")
}

func (o *Orchestrator) GetVersions(loaderName, mcVersion string) ([]LoaderVersion, error) {
	p, err := o.registry.Get(loaderName)
	if err != nil {
		return nil, err
	}
	versions, err := p.GetVersions(mcVersion)
	if err != nil {
		return nil, err
	}
	// Garantía única de orden: todos los providers devuelven la lista de mayor
	// a menor versión (numérica, no lexicográfica) para que la UI y el
	// algoritmo de recomendación tomen siempre la más reciente/estable.
	SortLoaderVersionsDesc(versions)
	return versions, nil
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
		// La lista llega ordenada de mayor a menor: la primera estable es la
		// versión más reciente y estable; si ninguna lo es, la más reciente.
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
	// El instalador oficial puede escribir el version.json con un id propio
	// (p. ej. "26.2-forge-65.1.0" en lugar del derivado "forge-26.2-65.1.0"):
	// usar el id real de disco para que BuildExecution/ProfileVersion apunten
	// al archivo que de verdad existe.
	if realID := installer.LookupInstalledJsonID(instancePath, loaderName, loaderVersion); realID != "" {
		versionJsonID = realID
	}
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

// stateKey normaliza el path del destino de instalación para usarlo como
// clave de la caché en memoria (sin distinguir mayúsculas en Windows).
func (o *Orchestrator) stateKey(instancePath string) string {
	abs, err := filepath.Abs(instancePath)
	if err != nil {
		abs = instancePath
	}
	return strings.ToLower(abs)
}

var errNoLoaderState = errors.New("no modloader installed")

func (o *Orchestrator) saveState(instancePath string, loader *InstalledLoader) error {
	key := o.stateKey(instancePath)
	o.mu.Lock()
	o.stateCache[key] = loader
	delete(o.stateRemoved, key)
	o.mu.Unlock()
	// Migración: elimina el loader-state.json legacy que quedaba en la raíz
	// del destino de instalación (best-effort, nunca condiciona la escritura).
	os.Remove(filepath.Join(instancePath, "loader-state.json"))
	return nil
}

func (o *Orchestrator) LoadState(instancePath string) (*InstalledLoader, error) {
	key := o.stateKey(instancePath)

	o.mu.RLock()
	if o.stateRemoved[key] {
		o.mu.RUnlock()
		return nil, errNoLoaderState
	}
	if loader := o.stateCache[key]; loader != nil {
		o.mu.RUnlock()
		return loader, nil
	}
	o.mu.RUnlock()

	// Sin caché (p. ej. tras reiniciar el launcher): el estado se reconstruye
	// desde el disco escaneando los version.json del destino.
	loader, err := o.deriveFromDisk(instancePath)
	if err != nil {
		return nil, err
	}
	o.mu.Lock()
	o.stateCache[key] = loader
	delete(o.stateRemoved, key)
	o.mu.Unlock()
	return loader, nil
}

func (o *Orchestrator) RemoveState(instancePath string) error {
	key := o.stateKey(instancePath)
	o.mu.Lock()
	delete(o.stateCache, key)
	o.stateRemoved[key] = true
	o.mu.Unlock()
	// Limpie el archivo legacy por si quedó de una instalación anterior.
	legacyPath := filepath.Join(instancePath, "loader-state.json")
	if err := os.Remove(legacyPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// deriveFromDisk reconstruye el estado del modloader instalado escaneando los
// version.json del destino (versions/*.json con inheritsFrom): el id y el
// inheritsFrom bastan para clasificar el loader y su versión.
func (o *Orchestrator) deriveFromDisk(instancePath string) (*InstalledLoader, error) {
	versionsDir := filepath.Join(instancePath, "versions")
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return nil, errNoLoaderState
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		jsonPath := filepath.Join(versionsDir, id, id+".json")
		raw, err := os.ReadFile(jsonPath)
		if err != nil {
			continue
		}
		var vj struct {
			ID           string `json:"id"`
			InheritsFrom string `json:"inheritsFrom"`
		}
		if err := json.Unmarshal(raw, &vj); err != nil {
			continue
		}
		if vj.InheritsFrom == "" {
			continue
		}
		jsonID := vj.ID
		if jsonID == "" {
			jsonID = id
		}
		loaderType, ok := DetectLoaderFromID(jsonID)
		if !ok {
			continue
		}
		loaderVersion := extractLoaderVersion(jsonID, vj.InheritsFrom, loaderType)
		if o.logFn != nil {
			o.logFn("Modloader %s %s detectado en disco para MC %s", loaderType, loaderVersion, vj.InheritsFrom)
		}
		return NewInstalledLoader(loaderType, loaderVersion, vj.InheritsFrom, id, ""), nil
	}
	return nil, errNoLoaderState
}

// DetectLoaderFromID clasifica el tipo de modloader a partir del id del
// version.json (o del nombre de carpeta) del perfil instalado. El orden es
// importante: neoforge contiene "forge" y fabric-loader empieza por "fabric-".
func DetectLoaderFromID(id string) (string, bool) {
	switch {
	case strings.HasPrefix(id, "fabric-loader-"), strings.HasPrefix(id, "fabric-"):
		return "fabric", true
	case strings.HasPrefix(id, "legacyfabric-loader-"), strings.HasPrefix(id, "legacyfabric-"):
		return "legacyfabric", true
	case strings.HasPrefix(id, "quilt-loader-"), strings.HasPrefix(id, "quilt-"):
		return "quilt", true
	case strings.HasPrefix(id, "neoforge-"), strings.Contains(id, "-neoforge-"):
		return "neoforge", true
	case strings.HasPrefix(id, "forge-"), strings.Contains(id, "-forge-"):
		return "forge", true
	}
	return "", false
}

// extractLoaderVersion quita del id el prefijo "<mc>-<loader>-" para obtener la
// versión del modloader (p. ej. "1.12.2-forge-14.23.5.2864" -> "14.23.5.2864").
func extractLoaderVersion(id, mcVersion, loaderType string) string {
	v := id
	if mcVersion != "" && strings.HasPrefix(v, mcVersion+"-") {
		v = strings.TrimPrefix(v, mcVersion+"-")
	}
	switch loaderType {
	case "fabric", "quilt":
		v = strings.TrimPrefix(v, loaderType+"-loader-")
		v = strings.TrimPrefix(v, loaderType+"-")
	case "legacyfabric":
		v = strings.TrimPrefix(v, "legacyfabric-loader-")
		v = strings.TrimPrefix(v, "legacyfabric-")
	default:
		v = strings.TrimPrefix(v, loaderType+"-")
	}
	return v
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
