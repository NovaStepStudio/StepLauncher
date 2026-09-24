package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"StepLauncher/internal/Core/Downloader"
)

type VersionInfo struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	ReleaseTime string `json:"time"`
}

var validTypes = map[string]bool{
	"release": true, "snapshot": true, "old_beta": true, "old_alpha": true,
}

type versionManifest struct {
	Versions []struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		URL         string `json:"url"`
		ReleaseTime string `json:"releaseTime"`
	} `json:"versions"`
}

type InstalledVersion struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (e *Engine) ListDownloadedVersions() []InstalledVersion {
	versionsDir := filepath.Join(e.config.Get().WorkDir, "versions")
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return []InstalledVersion{}
	}

	out := make([]InstalledVersion, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id := entry.Name()
		if id == "" {
			continue
		}
		var meta struct {
			Type string `json:"type"`
		}
		verPath := filepath.Join(versionsDir, id, id+".json")
		if data, err := os.ReadFile(verPath); err == nil {
			json.Unmarshal(data, &meta)
		} else {
			continue
		}
		if meta.Type == "" {
			meta.Type = "release"
		}
		out = append(out, InstalledVersion{ID: id, Type: meta.Type})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

// getManifestURL pide una URL con hasta 3 intentos y backoff para fallos
// transitorios. Los errores de protocolo de proxy fallan rápido: ya tienen
// su propia ruta (reintento directo + caché expirado) y reintentarlos solo
// alargaría el error.
func getManifestURL(client *http.Client, url string) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ {
		resp, err = client.Get(url)
		if err == nil {
			return resp, nil
		}
		if downloader.IsProxyProtocolError(err) {
			return nil, err
		}
		if i < 2 {
			time.Sleep(time.Duration(1<<i) * time.Second)
		}
	}
	return nil, err
}

func (e *Engine) GetVersions(versionType string) ([]VersionInfo, error) {
	if !validTypes[versionType] {
		return nil, fmt.Errorf("invalid version type '%s'. valid: release, snapshot, old_beta, old_alpha", versionType)
	}

	cacheKey := fmt.Sprintf("manifest-%s", versionType)
	var cached []VersionInfo
	found, _ := e.cache.Get("manifest", cacheKey, &cached)
	if found {
		return cached, nil
	}
	// También intentar fallback a caché si la red falla (igual que FetchVersionManifest).
	// Vale cualquier copia guardada, fresca o expirada: sin red se sirve lo que haya.
	const fullCacheKey = "full"
	var fullCached downloader.Manifest
	foundFull, _, _ := e.cache.GetWithFallback("manifest", fullCacheKey, &fullCached)

	client := e.downloader.HTTPClient()
	resp, err := getManifestURL(client, "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json")
	if err != nil {
		// Detectar mala configuración del proxy (SOCKS vs HTTP) y dar pista accionable
		if downloader.IsProxyProtocolError(err) {
			cfg := e.config.Get()
			if cfg.ProxyEnabled {
				err = downloader.WrapProxyError(err, cfg.ProxyHost, cfg.ProxyPort)
			} else {
				err = fmt.Errorf("%w — parece que un proxy del sistema interceptó la conexión. Desactiva el proxy en Ajustes > Red o verifica su puerto", err)
			}
			// Fallback a caché si existe (fresca o expirada): sin red se sirve
			// lo guardado en vez de romper la lista de versiones.
			if foundFull {
				e.log.Info("[Cache] WARN: usando manifest en caché tras error de proxy: %v", err)
				return filterManifestByType(&fullCached, versionType, cacheKey)
			}
			// Último intento: reintentar sin proxy (útil si el proxy está mal configurado pero hay internet directo)
			if cfg.ProxyEnabled {
				e.log.Info("[Network] Reintentando manifest sin proxy tras error de protocolo: %v", err)
				direct := downloader.DefaultHTTPClient()
				if resp2, err2 := direct.Get("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"); err2 == nil {
					resp = resp2
					err = nil
				} else {
					if foundFull {
						e.log.Info("[Cache] WARN: usando manifest en caché tras fallo directo: %v", err2)
						return filterManifestByType(&fullCached, versionType, cacheKey)
					}
				}
			}
			if err != nil {
				return nil, fmt.Errorf("failed to fetch manifest: %w", err)
			}
		} else {
			if foundFull {
				e.log.Info("[Cache] WARN: usando manifest en caché (red no disponible: %v)", err)
				return filterManifestByType(&fullCached, versionType, cacheKey)
			}
			return nil, fmt.Errorf("failed to fetch manifest: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if foundFull {
			e.log.Info("[Cache] WARN: usando manifest en caché (HTTP %d)", resp.StatusCode)
			return filterManifestByType(&fullCached, versionType, cacheKey)
		}
		return nil, fmt.Errorf("failed to fetch manifest: HTTP %d", resp.StatusCode)
	}

	var m versionManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		if foundFull {
			e.log.Info("[Cache] WARN: usando manifest en caché (respuesta inválida: %v)", err)
			return filterManifestByType(&fullCached, versionType, cacheKey)
		}
		// Si el body es binario por proxy mal configurado, el error de decode también debe sugerir proxy
		if strings.Contains(err.Error(), "invalid character") && downloader.IsProxyProtocolError(fmt.Errorf("%v", err)) {
			// no es fiable, solo envolver genérico
		}
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}

	var result []VersionInfo
	for _, v := range m.Versions {
		if v.Type == versionType {
			result = append(result, VersionInfo{
				ID: v.ID, Type: v.Type, URL: v.URL, ReleaseTime: v.ReleaseTime,
			})
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no versions found for type '%s'", versionType)
	}

	e.cache.Set("manifest", cacheKey, result)
	return result, nil
}

// filterManifestByType convierte un Manifest completo cacheado al filtro por tipo solicitado
func filterManifestByType(m *downloader.Manifest, versionType, cacheKey string) ([]VersionInfo, error) {
	var result []VersionInfo
	for _, v := range m.Versions {
		if v.Type == versionType {
			result = append(result, VersionInfo{
				ID: v.ID, Type: v.Type, URL: v.URL, ReleaseTime: v.ReleaseTime,
			})
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no versions found for type '%s' (caché expirado vacío)", versionType)
	}
	return result, nil
}

func (e *Engine) RefreshManifests() (int, error) {
	// No destructivo: primero se intenta descargar y solo con datos nuevos
	// se invalidan los derivados por tipo. Si la red falla, FetchVersionManifest
	// sirve la copia guardada y aquí se retorna tal cual, sin borrar nada.
	m, err := e.FetchVersionManifest()
	if err != nil {
		return 0, err
	}
	for _, t := range []string{"release", "snapshot", "old_beta", "old_alpha"} {
		_ = e.cache.Delete("manifest", "manifest-"+t)
	}
	e.log.Info("[Cache] Manifiestos refrescados: %d versiones disponibles", len(m.Versions))
	return len(m.Versions), nil
}

func (e *Engine) FetchVersionManifest() (*downloader.Manifest, error) {
	const cacheKey = "full"
	var m downloader.Manifest
	found, expired, _ := e.cache.GetWithFallback("manifest", cacheKey, &m)
	if found && !expired {
		return &m, nil
	}

	client := e.downloader.HTTPClient()
	resp, err := getManifestURL(client, "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json")
	if err != nil {
		// Sin red se sirve la copia guardada (fresca o expirada) en vez de
		// romper la lista de versiones. Solo falla si no hay nada guardado.
		if downloader.IsProxyProtocolError(err) {
			cfg := e.config.Get()
			if cfg.ProxyEnabled {
				err = downloader.WrapProxyError(err, cfg.ProxyHost, cfg.ProxyPort)
			} else {
				err = fmt.Errorf("%w — un proxy del sistema puede estar interceptando. Revisa Ajustes > Red", err)
			}
			if found {
				e.log.Info("[Cache] WARN: usando manifest en caché tras error de proxy: %v", err)
				return &m, nil
			}
			// Reintento directo sin proxy
			if e.config.Get().ProxyEnabled {
				e.log.Info("[Network] Reintentando manifest sin proxy tras error de protocolo")
				direct := downloader.DefaultHTTPClient()
				if resp2, err2 := direct.Get("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"); err2 == nil {
					resp = resp2
					err = nil
				} else {
					if found {
						e.log.Info("[Cache] WARN: usando manifest en caché tras fallo directo: %v", err2)
						return &m, nil
					}
					return nil, fmt.Errorf("fetch manifest: %w", err)
				}
			} else {
				return nil, fmt.Errorf("fetch manifest: %w", err)
			}
		} else {
			if found {
				e.log.Info("[Cache] WARN: usando manifest en caché (red no disponible: %v)", err)
				return &m, nil
			}
			return nil, fmt.Errorf("fetch manifest: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if found {
			e.log.Info("[Cache] WARN: usando manifest en caché (HTTP %d)", resp.StatusCode)
			return &m, nil
		}
		return nil, fmt.Errorf("fetch manifest: HTTP %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		if found {
			e.log.Info("[Cache] WARN: usando manifest en caché (respuesta inválida: %v)", err)
			return &m, nil
		}
		return nil, fmt.Errorf("decode manifest: %w", err)
	}

	if err := e.cache.Set("manifest", cacheKey, m); err != nil {
		e.log.Info("[Cache] WARN: no se pudo guardar el manifest: %v", err)
	}
	return &m, nil
}
