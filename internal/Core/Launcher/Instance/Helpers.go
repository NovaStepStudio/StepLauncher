package instance

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"StepLauncher/internal/Core/Downloader"
)

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("inst-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func sanitizeInstanceName(name string) error {
	if name == "" {
		return fmt.Errorf("instance name cannot be empty")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("instance name cannot contain '..'")
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("instance name cannot contain path separators")
	}
	if filepath.Base(name) != name {
		return fmt.Errorf("invalid instance name")
	}
	return nil
}

func (m *InstanceManager) instancePath(name string) (string, error) {
	if err := sanitizeInstanceName(name); err != nil {
		return "", err
	}
	return filepath.Join(m.instancesDir, name), nil
}

func (m *InstanceManager) metadataPath(name string) (string, error) {
	if err := sanitizeInstanceName(name); err != nil {
		return "", err
	}
	return filepath.Join(m.instancesDir, name, metadataFile), nil
}

func (m *InstanceManager) configPath(name string) (string, error) {
	if err := sanitizeInstanceName(name); err != nil {
		return "", err
	}
	return filepath.Join(m.instancesDir, name, configFile), nil
}

func (m *InstanceManager) lockPath(name string) (string, error) {
	if err := sanitizeInstanceName(name); err != nil {
		return "", err
	}
	return filepath.Join(m.instancesDir, name, lockFile), nil
}

func (m *InstanceManager) readMetadata(name string) (*InstanceMetadata, error) {
	path, err := m.metadataPath(name)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read metadata: %w", err)
	}
	var meta InstanceMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parse metadata: %w", err)
	}
	return &meta, nil
}

func (m *InstanceManager) writeMetadata(name string, meta *InstanceMetadata) error {
	path, err := m.metadataPath(name)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename metadata: %w", err)
	}
	return nil
}

func (m *InstanceManager) readConfig(name string) (*InstanceLaunchConfig, error) {
	path, err := m.configPath(name)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg InstanceLaunchConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func (m *InstanceManager) writeConfig(name string, cfg *InstanceLaunchConfig) error {
	path, err := m.configPath(name)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename config: %w", err)
	}
	return nil
}

func (m *InstanceManager) acquireLock(name, action, version string) error {
	lockPath, err := m.lockPath(name)
	if err != nil {
		return err
	}
	lock := InstanceLock{
		PID:     os.Getpid(),
		Action:  action,
		Version: version,
		Since:   time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(lock)
	if err != nil {
		return fmt.Errorf("marshal lock: %w", err)
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			existingData, _ := os.ReadFile(lockPath)
			var existingLock InstanceLock
			if json.Unmarshal(existingData, &existingLock) == nil {
				return fmt.Errorf("instance %s is locked: %s (PID %d since %s)", name, existingLock.Action, existingLock.PID, existingLock.Since)
			}
			return fmt.Errorf("instance %s is already locked", name)
		}
		return fmt.Errorf("create lock: %w", err)
	}
	f.Write(data)
	return f.Close()
}

func (m *InstanceManager) releaseLock(name string) {
	lockPath, err := m.lockPath(name)
	if err != nil {
		return
	}
	os.Remove(lockPath)
}

func (m *InstanceManager) addVersionToMetadata(name, version string) error {
	meta, err := m.readMetadata(name)
	if err != nil {
		return err
	}
	for _, v := range meta.Versions {
		if v == version {
			return nil
		}
	}
	meta.Versions = append(meta.Versions, version)
	if err := m.writeMetadata(name, meta); err != nil {
		return err
	}
	cfg, err := m.readConfig(name)
	if err != nil {
		return err
	}
	if cfg.Version != version {
		cfg.Version = version
		if err := m.writeConfig(name, cfg); err != nil {
			return err
		}
	}
	return nil
}

// syncVersionsFromDisk relee el directorio versions/ de la instancia y
// registra en el metadata todas las versiones presentes que todavía no
// constan. Se invoca al terminar una instalación COMPLETA (descarga de la
// versión base + instalación del modloader), cuando la carpeta ya no está en
// plena escritura: el instalador oficial de Forge/NeoForge crea su propio
// directorio de versión (p. ej. "26.2-forge-65.1.0") que la descarga inicial
// no conoce y que sin este escaneo nunca quedaría registrado en el JSON.
// Se registra TODO lo que haya dentro de versions/: la carpeta es la fuente
// de verdad y no se condiciona a que exista <version>.json (algunos
// instaladores dejan el directorio sin su json durante un instante, o con
// nombres que no coinciden exactamente con el contenido del manifest).
func (m *InstanceManager) syncVersionsFromDisk(name string) error {
	instPath, err := m.instancePath(name)
	if err != nil {
		return err
	}
	versionsDir := filepath.Join(instPath, "versions")
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return fmt.Errorf("read versions dir: %w", err)
	}

	meta, err := m.readMetadata(name)
	if err != nil {
		return err
	}
	seen := make(map[string]bool, len(meta.Versions))
	for _, v := range meta.Versions {
		seen[v] = true
	}
	changed := false
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		ver := e.Name()
		if seen[ver] {
			continue
		}
		seen[ver] = true
		meta.Versions = append(meta.Versions, ver)
		changed = true
	}
	if !changed {
		return nil
	}
	return m.writeMetadata(name, meta)
}

func buildDownloadFilter(version string, req AddVersionReq) downloader.DownloadFilter {
	filter := downloader.DownloadFilter{Version: version}
	if req.Client != nil {
		filter.Client = *req.Client
	}
	if req.Libraries != nil {
		filter.Libraries = *req.Libraries
	}
	if req.Natives != nil {
		filter.Natives = *req.Natives
	}
	if req.Assets != nil {
		filter.Assets = *req.Assets
	}
	if req.Java != nil {
		filter.Java = *req.Java
	}
	if !filter.Client && !filter.Libraries && !filter.Natives && !filter.Assets && !filter.Java {
		filter.Client, filter.Libraries, filter.Natives, filter.Assets, filter.Java = true, true, true, true, true
	}
	return filter
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func nonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func firstNonNil[T any](vals ...*T) *T {
	for _, v := range vals {
		if v != nil {
			return v
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}
