package provider

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"StepLauncher/internal/Core/Cache"
	"StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Core/Downloader/Utils"
	"StepLauncher/internal/Core/ModLoader"
	"StepLauncher/internal/Core/ModLoader/Installer"
)

type NeoForgeProvider struct {
	*AbstractForgeProvider
	versionResolver NeoForgeVersionResolver
}

func NewNeoForgeProvider(cacheDir string, client *http.Client, cacheMgr *cache.Manager) *NeoForgeProvider {
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

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i] > filtered[j]
	})

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
	versionID := p.VersionJsonID(mcVersion, loaderVersion)

	if broadcast != nil {
		broadcast(modloader.InstallingEvent(sessionId, p.NameVal, "Extracting installer..."))
	}

	profileLibs, err := installer.ExecuteInstaller(plan.InstallerDest, versionID, instancePath, librariesPath)
	if err != nil {
		return fmt.Errorf("execute installer: %w", err)
	}

	if len(profileLibs) > 0 && broadcast != nil {
		broadcast(modloader.InstallingEvent(sessionId, p.NameVal, fmt.Sprintf("Downloading %d profile libraries...", len(profileLibs))))
	}

	ctx := context.Background()
	for i, lib := range profileLibs {
		if utils.FileExists(lib.Dest) {
			if lib.SHA1 != "" {
				if ok, _ := utils.VerifySHA1(lib.Dest, lib.SHA1); ok {
					continue
				}
				os.Remove(lib.Dest)
			} else {
				continue
			}
		}
		if broadcast != nil {
			broadcast(modloader.InstallingEvent(sessionId, p.NameVal, fmt.Sprintf("Library %d/%d: %s", i+1, len(profileLibs), lib.Name)))
		}
		task := downloader.DownloadTask{
			URL:  lib.URL,
			Dest: lib.Dest,
			SHA1: lib.SHA1,
			Size: lib.Size,
		}
		if err := downloader.DownloadFile(ctx, task, p.httpClient, 3, nil, 60000, 3); err != nil {
			return fmt.Errorf("download profile lib %s: %w", lib.Name, err)
		}
	}

	return nil
}

var _ modloader.ModLoaderProvider = (*NeoForgeProvider)(nil)
