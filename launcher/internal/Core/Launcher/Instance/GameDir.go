package instance

import (
	"fmt"
	"os"
	"path/filepath"
)

// HasInstance indica si la instancia existe (tiene su metadata en disco).
func (m *InstanceManager) HasInstance(name string) bool {
	if err := sanitizeInstanceName(name); err != nil {
		return false
	}
	metaPath := filepath.Join(m.instancesDir, name, metadataFile)
	info, err := os.Stat(metaPath)
	return err == nil && !info.IsDir()
}

// GameDirFor resuelve el gameDir de una instancia respetando la opción de
// carpeta "game" separada: <instancia>/game si está activa, o la propia
// carpeta de la instancia si está deshabilitada. Falla si la instancia no
// existe, está en verificación o sigue en creación.
func (m *InstanceManager) GameDirFor(name string) (string, error) {
	return m.gameDirFor(name, false)
}

// GameDirForSystem igual pero para el flujo interno del motor con una
// instancia en creación (ya existe en disco, aún oculta).
func (m *InstanceManager) GameDirForSystem(name string) (string, error) {
	return m.gameDirFor(name, true)
}

func (m *InstanceManager) gameDirFor(name string, system bool) (string, error) {
	if !system {
		if err := m.assertUsable(name); err != nil {
			return "", err
		}
	}
	instPath, err := m.instancePath(name)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(instPath, metadataFile)); err != nil {
		return "", fmt.Errorf("la instancia %s no existe", name)
	}
	m.mu.RLock()
	sep := m.separateGameDir
	m.mu.RUnlock()
	if sep {
		return filepath.Join(instPath, "game"), nil
	}
	return instPath, nil
}
