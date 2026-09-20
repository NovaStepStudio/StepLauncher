package instance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Core/Downloader/Utils"
	"StepLauncher/internal/Core/Launcher"
)

// VerifyInstance comprueba la integridad de TODAS las versiones de la
// instancia de forma síncrona. La verificación asíncrona (StartInstanceVerify)
// bloquea la instancia durante el proceso; esta función solo queda disponible
// para consultas puntuales.
func (m *InstanceManager) VerifyInstance(name string) ([]VerifyResult, error) {
	if err := m.assertUsable(name); err != nil {
		return nil, err
	}
	meta, err := m.readMetadata(name)
	if err != nil {
		return nil, fmt.Errorf("instance %s not found", name)
	}

	dlCfg := m.verifyDownloaderConfig()
	var results []VerifyResult
	for _, version := range meta.Versions {
		results = append(results, *m.verifyVersion(name, dlCfg, version))
	}

	if len(results) == 0 {
		results = append(results, VerifyResult{Version: "", Valid: true, Issues: []VerifyIssue{}})
	}
	return results, nil
}

func (m *InstanceManager) VerifySingleVersion(name, version string) (*VerifyResult, error) {
	meta, err := m.readMetadata(name)
	if err != nil {
		return nil, fmt.Errorf("instance %s not found", name)
	}
	found := false
	for _, v := range meta.Versions {
		if v == version {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("version %s not found in instance %s", version, name)
	}
	dlCfg := m.verifyDownloaderConfig()
	return m.verifyVersion(name, dlCfg, version), nil
}

func (m *InstanceManager) verifyDownloaderConfig() downloader.Config {
	cacheDir := m.cacheDir
	if cacheDir == "" {
		cacheDir = filepath.Join(m.sharedDir, "cache")
	}
	return downloader.Config{
		WorkDir:    m.sharedDir,
		CacheDir:   cacheDir,
		HTTPClient: downloader.DefaultHTTPClient(),
	}
}

// verifyVersion verifica una sola versión de una instancia y devuelve el
// resultado (sin tocar locks ni metadata; se puede llamar desde goroutines).
func (m *InstanceManager) verifyVersion(name string, dlCfg downloader.Config, version string) *VerifyResult {
	result := &VerifyResult{Version: version, Valid: true}
	instPath, err := m.instancePath(name)
	if err != nil {
		result.Valid = false
		result.Issues = append(result.Issues, VerifyIssue{Type: "error", File: "", Message: fmt.Sprintf("instance path: %v", err)})
		return result
	}
	verDir := filepath.Join(instPath, "versions", version)

	verJSON := filepath.Join(verDir, version+".json")
	if _, err := os.Stat(verJSON); err != nil {
		result.Issues = append(result.Issues, VerifyIssue{Type: "missing", File: verJSON, Message: "version JSON not found"})
		result.Valid = false
		return result
	}

	var ver downloader.VersionJSON
	data, err := os.ReadFile(verJSON)
	if err != nil {
		result.Issues = append(result.Issues, VerifyIssue{Type: "error", File: verJSON, Message: fmt.Sprintf("cannot read: %v", err)})
		result.Valid = false
		return result
	}
	if err := json.Unmarshal(data, &ver); err != nil {
		result.Issues = append(result.Issues, VerifyIssue{Type: "error", File: verJSON, Message: fmt.Sprintf("cannot parse: %v", err)})
		result.Valid = false
		return result
	}
	downloader.NormalizeVersion(&ver)

	clientJar := filepath.Join(verDir, version+".jar")
	if _, err := os.Stat(clientJar); err != nil {
		result.Issues = append(result.Issues, VerifyIssue{Type: "missing", File: clientJar, Message: "client JAR not found"})
		result.Valid = false
	} else if ver.Downloads.Client.SHA1 != "" {
		if ok, err := utils.VerifySHA1(clientJar, ver.Downloads.Client.SHA1); err != nil || !ok {
			result.Issues = append(result.Issues, VerifyIssue{Type: "corrupt", File: clientJar, Message: "SHA-1 mismatch"})
			result.Valid = false
		}
	}

	filter := downloader.DownloadFilter{Version: version, Client: false, Libraries: true, Natives: true, Assets: true, Java: false}
	allTasks, err := downloader.BuildTasks(dlCfg, &ver, version, filter)
	if err != nil {
		result.Issues = append(result.Issues, VerifyIssue{Type: "error", Message: fmt.Sprintf("build tasks: %v", err)})
		result.Valid = false
		return result
	}

	for _, t := range allTasks {
		if t.SHA1 == "" {
			continue
		}
		if !utils.FileExists(t.Dest) {
			result.Issues = append(result.Issues, VerifyIssue{Type: "missing", File: t.Dest, Message: fmt.Sprintf("file not found (%s)", t.Section)})
			result.Valid = false
			continue
		}
		if ok, err := utils.VerifySHA1(t.Dest, t.SHA1); err != nil || !ok {
			result.Issues = append(result.Issues, VerifyIssue{Type: "corrupt", File: t.Dest, Message: fmt.Sprintf("SHA-1 mismatch (%s)", t.Section)})
			result.Valid = false
		}
	}

	if _, err := os.Stat(filepath.Join(verDir, "natives")); err != nil {
		result.Issues = append(result.Issues, VerifyIssue{Type: "missing", File: filepath.Join(verDir, "natives"), Message: "natives not extracted"})
	}

	return result
}

// StartInstanceVerify lanza una verificación de integridad ASÍNCRONA de esta
// instancia concreta. Mientras dure, la instancia NO se puede utilizar:
// lanzar, descargar, instalar modloaders, editar, borrar o clonar se rechaza.
// El progreso se consulta con InstanceVerifyStatus (patrón de polling, igual
// que la verificación global de integridad).
func (m *InstanceManager) StartInstanceVerify(name string) error {
	if err := sanitizeInstanceName(name); err != nil {
		return err
	}
	if _, err := m.readMetadata(name); err != nil {
		return fmt.Errorf("instancia no encontrada: %s", name)
	}

	m.mu.Lock()
	if m.verifying[name] {
		m.mu.Unlock()
		return fmt.Errorf("la instancia %s ya se está verificando", name)
	}
	if m.hasActiveDownload(name) {
		m.mu.Unlock()
		return fmt.Errorf("la instancia %s tiene una descarga en curso; espera a que termine", name)
	}
	m.mu.Unlock()

	if m.instanceRunning(name) {
		return fmt.Errorf("la instancia %s está en ejecución; ciérrala antes de verificar", name)
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.verifying[name] = true
	m.verifyCancel[name] = cancel
	m.verifyProgress[name] = &InstanceVerifyProgress{State: "verifying", Phase: "versions"}
	m.mu.Unlock()

	go m.runInstanceVerify(name, ctx)
	m.log("Verificación de integridad iniciada para la instancia %s", name)
	return nil
}

func (m *InstanceManager) CancelInstanceVerify(name string) {
	m.mu.Lock()
	cancel := m.verifyCancel[name]
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (m *InstanceManager) InstanceVerifyStatus(name string) *InstanceVerifyProgress {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p := m.verifyProgress[name]
	if p == nil {
		return &InstanceVerifyProgress{State: "idle"}
	}
	// Copia para que el llamador nunca vea mutaciones a mitad.
	cp := *p
	return &cp
}

func (m *InstanceManager) setVerifyProgress(name string, fn func(p *InstanceVerifyProgress)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p := m.verifyProgress[name]; p != nil {
		fn(p)
	}
}

func (m *InstanceManager) runInstanceVerify(name string, ctx context.Context) {
	defer func() {
		m.mu.Lock()
		m.verifying[name] = false
		delete(m.verifyCancel, name)
		m.mu.Unlock()
	}()

	versions := m.scanVersionsFromDisk(name)
	if len(versions) == 0 {
		m.setVerifyProgress(name, func(p *InstanceVerifyProgress) {
			p.State = "done"
			p.Phase = "done"
			p.Percent = 100
		})
		m.log("Verificación de %s: sin versiones que comprobar", name)
		return
	}

	dlCfg := m.verifyDownloaderConfig()
	total := len(versions)
	for i, version := range versions {
		select {
		case <-ctx.Done():
			m.setVerifyProgress(name, func(p *InstanceVerifyProgress) { p.State = "cancelled" })
			m.log("Verificación de %s cancelada", name)
			return
		default:
		}

		m.setVerifyProgress(name, func(p *InstanceVerifyProgress) {
			p.Version = version
			p.Phase = "verifying"
			p.Percent = i * 100 / total
		})

		res := m.verifyVersion(name, dlCfg, version)
		m.setVerifyProgress(name, func(p *InstanceVerifyProgress) {
			p.Found++
			if res != nil && !res.Valid {
				p.Issues += len(res.Issues)
			}
		})
	}

	m.setVerifyProgress(name, func(p *InstanceVerifyProgress) {
		p.State = "done"
		p.Phase = "done"
		p.Percent = 100
	})
	m.log("Verificación de %s completada: %d versiones, %d problemas", name, total, m.InstanceVerifyStatus(name).Issues)
}

func (m *InstanceManager) hasActiveDownload(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, dl := range m.downloads {
		if dl.InstanceName == name {
			return true
		}
	}
	return false
}

func (m *InstanceManager) instanceRunning(name string) bool {
	if m.launchManager == nil {
		return false
	}
	for _, g := range m.launchManager.List() {
		if g.InstanceName == name {
			s := g.GetStatus()
			if s == launcher.GameStarting || s == launcher.GameRunning {
				return true
			}
		}
	}
	return false
}