package provider

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"StepLauncher/internal/Core/Cache"
	"StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Core/ModLoader"
)

type NeoForgeProvider struct {
	*AbstractForgeProvider
	versionResolver NeoForgeVersionResolver
}

func NewNeoForgeProvider(cacheDir string, client *http.Client, cacheMgr *cache.Manager, javaResolver func(mcVersion, instancePath string) (string, error)) *NeoForgeProvider {
	return &NeoForgeProvider{
		AbstractForgeProvider: &AbstractForgeProvider{
			NameVal:      "neoforge",
			MetadataURL:  "https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml",
			MavenBase:    "https://maven.neoforged.net/releases",
			GroupPath:    "net/neoforged/neoforge",
			ArtifactName: "neoforge",
			CacheDir:     cacheDir,
			CacheManager: cacheMgr,
			httpClient:   client,
			JavaResolver: javaResolver,
		},
	}
}

func (p *NeoForgeProvider) GetVersions(mcVersion string) ([]modloader.LoaderVersion, error) {
	raw, err := p.fetchMetadata()
	if err != nil {
		return nil, err
	}

	type neoMetadata struct {
		XMLName    struct{} `xml:"metadata"`
		Versioning struct {
			Versions struct {
				Version []string `xml:"version"`
			} `xml:"versions"`
		} `xml:"versioning"`
	}
	var metadata neoMetadata
	if err := xml.Unmarshal(raw, &metadata); err != nil {
		return nil, fmt.Errorf("parse neoforge xml: %w", err)
	}

	allVersions := metadata.Versioning.Versions.Version

	filtered := p.versionResolver.FilterVersionsForMinecraft(allVersions, mcVersion)

	result := make([]modloader.LoaderVersion, 0, len(filtered))
	for _, v := range filtered {
		mcv := p.versionResolver.NeoForgeVersionToMcVersion(v)
		if mcv == "" {
			mcv = mcVersion
		}
		result = append(result, modloader.LoaderVersion{
			LoaderVersion:    v,
			MinecraftVersion: mcv,
			Stable:           !strings.Contains(strings.ToLower(v), "alpha") && !strings.Contains(strings.ToLower(v), "beta"),
		})
	}
	return result, nil
}

// fetchMetadata descarga el maven-metadata.xml con reintento directo si el
// proxy está mal configurado y fallback a caché expirado (igual que el resto
// de providers vía FetchJSON): el XML cambia poco y sin red no hay versiones.
func (p *NeoForgeProvider) fetchMetadata() ([]byte, error) {
	const cacheKey = "neoforge-metadata-xml"
	var cached string
	found, expired, _ := p.cachedXML(cacheKey, &cached)

	resp, err := p.httpClient.Get(p.MetadataURL)
	if err != nil {
		if downloader.IsProxyProtocolError(err) {
			if resp2, err2 := downloader.DefaultHTTPClient().Get(p.MetadataURL); err2 == nil {
				resp = resp2
				err = nil
			} else {
				err = err2
			}
		}
		if err != nil {
			if found && expired {
				return []byte(cached), nil
			}
			return nil, fmt.Errorf("fetch neoforge metadata: %w (revisa Ajustes > Red > Proxy)", downloader.WrapProxyError(err, "", 0))
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if found && expired && (resp.StatusCode == 429 || resp.StatusCode >= 500) {
			return []byte(cached), nil
		}
		return nil, fmt.Errorf("fetch neoforge metadata: HTTP %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read neoforge metadata: %w", err)
	}
	if p.CacheManager != nil {
		_ = p.CacheManager.Set("modloader", cacheKey, string(raw))
	}
	return raw, nil
}

// cachedXML lee el XML cacheado (fresco o expirado) como texto plano.
func (p *NeoForgeProvider) cachedXML(cacheKey string, out *string) (found, expired bool, err error) {
	if p.CacheManager == nil {
		return false, false, nil
	}
	return p.CacheManager.GetWithFallback("modloader", cacheKey, out)
}

func (p *NeoForgeProvider) RunInstaller(sessionId string, plan *modloader.DownloadPlan, mcVersion, loaderVersion, instancePath, librariesPath, minecraftJar string, broadcast func([]byte)) error {
	return p.AbstractForgeProvider.RunInstaller(sessionId, plan, mcVersion, loaderVersion, instancePath, librariesPath, minecraftJar, broadcast)
}

var _ modloader.ModLoaderProvider = (*NeoForgeProvider)(nil)
