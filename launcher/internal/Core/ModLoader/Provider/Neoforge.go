package provider

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"StepLauncher/internal/Core/Cache"
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
	resp, err := p.httpClient.Get(p.MetadataURL)
	if err != nil {
		return nil, fmt.Errorf("fetch neoforge metadata: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read neoforge metadata: %w", err)
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

func (p *NeoForgeProvider) RunInstaller(sessionId string, plan *modloader.DownloadPlan, mcVersion, loaderVersion, instancePath, librariesPath, minecraftJar string, broadcast func([]byte)) error {
	return p.AbstractForgeProvider.RunInstaller(sessionId, plan, mcVersion, loaderVersion, instancePath, librariesPath, minecraftJar, broadcast)
}

var _ modloader.ModLoaderProvider = (*NeoForgeProvider)(nil)
