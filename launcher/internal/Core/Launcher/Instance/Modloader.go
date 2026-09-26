package instance

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	modloader "StepLauncher/internal/Core/ModLoader"
)

// InstallModLoader instala un modloader dentro del directorio de la instancia
// (targetDir = <instancesDir>/<name>). Requiere que la versión de Minecraft ya
// esté descargada (directorio versions/<mcVersion>). Devuelve el sessionId de
// la instalación; el progreso se emite con los eventos modloader_* y el
// sessionId correspondiente.
func (m *InstanceManager) InstallModLoader(name, loader, loaderVersion, mcVersion string) (string, error) {
	return m.installModLoader(name, loader, loaderVersion, mcVersion, false)
}

// InstallModLoaderSystem instala saltando el bloqueo de provisioning. Solo la
// usa el flujo interno del motor para la instancia en creación.
func (m *InstanceManager) InstallModLoaderSystem(name, loader, loaderVersion, mcVersion string) (string, error) {
	return m.installModLoader(name, loader, loaderVersion, mcVersion, true)
}

func (m *InstanceManager) installModLoader(name, loader, loaderVersion, mcVersion string, system bool) (string, error) {
	if !system {
		if err := m.assertUsable(name); err != nil {
			return "", err
		}
	}
	if m.mlOrchestrator == nil {
		return "", fmt.Errorf("modloader engine not available")
	}
	instPath, err := m.instancePath(name)
	if err != nil {
		return "", err
	}
	if mcVersion == "" || strings.Contains(mcVersion, "..") || strings.ContainsAny(mcVersion, "/\\") {
		return "", fmt.Errorf("invalid minecraft version")
	}
	verDir := filepath.Join(instPath, "versions", mcVersion)
	if _, err := os.Stat(verDir); err != nil {
		return "", fmt.Errorf("la versión %s debe descargarse antes de instalar el modloader", mcVersion)
	}

	sessionID := fmt.Sprintf("ml-inst-%d", time.Now().UnixMilli())
	go func() {
		if _, err := m.mlOrchestrator.Install(sessionID, loader, loaderVersion, mcVersion, instPath); err != nil {
			m.log("Modloader install failed for instance %s: %v", name, err)
			return
		}
		m.log("Modloader install finished for instance %s: %s %s", name, loader, loaderVersion)
		// Las librerías del instalador quedan en <instancia>/libraries: se
		// mueven a shared (donde se leen en el lanzamiento) y se eliminan.
		m.mergeInstanceLibraries(instPath)
		// La instalación del loader es parte de una instalación COMPLETA de la
		// instancia: el instalador oficial añade su propia versión en
		// versions/ (p. ej. "26.2-forge-65.1.0") y solo ahora, terminada, se
		// puede leer la carpeta con seguridad para registrar el modloader en
		// el metadata (la descarga base estaba en escritura continua).
		if err := m.syncVersionsFromDisk(name); err != nil {
			m.log("ERROR: sync versions from disk after modloader install: %v", err)
		} else {
			m.log("Instance %s: versions synced from disk after modloader install", name)
		}
	}()

	return sessionID, nil
}

// InstalledModLoader devuelve el estado del modloader instalado en la instancia
// (nil, nil si no hay ninguno: "sin loader" es un estado normal, no un error,
// para que el binding no registre fallos espurios en el log).
func (m *InstanceManager) InstalledModLoader(name string) (*modloader.InstalledLoader, error) {
	if m.mlOrchestrator == nil {
		return nil, fmt.Errorf("modloader engine not available")
	}
	instPath, err := m.instancePath(name)
	if err != nil {
		return nil, err
	}
	st, err := m.mlOrchestrator.LoadState(instPath)
	if err != nil {
		if errors.Is(err, modloader.ErrNoLoaderState) {
			return nil, nil
		}
		return nil, err
	}
	return st, nil
}

// RemoveModLoaderState borra el estado del modloader de la instancia sin tocar
// los archivos descargados.
func (m *InstanceManager) RemoveModLoaderState(name string) error {
	if m.mlOrchestrator == nil {
		return nil
	}
	instPath, err := m.instancePath(name)
	if err != nil {
		return err
	}
	return m.mlOrchestrator.RemoveState(instPath)
}