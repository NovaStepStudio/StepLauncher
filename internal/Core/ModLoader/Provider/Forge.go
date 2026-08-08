package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"StepLauncher/internal/Core/Cache"
	"StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Core/Downloader/Utils"
	"StepLauncher/internal/Core/ModLoader"
	"StepLauncher/internal/Core/ModLoader/Installer"
	globalutils "StepLauncher/internal/Core/Utils"
)

type AbstractForgeProvider struct {
	modloader.BaseProvider
	NameVal      string
	MetadataURL  string
	MavenBase    string
	GroupPath    string
	ArtifactName string
	CacheDir     string
	CacheManager *cache.Manager
	httpClient   *http.Client
}

func (p *AbstractForgeProvider) Name() string { return p.NameVal }

func (p *AbstractForgeProvider) GetVersions(mcVersion string) ([]modloader.LoaderVersion, error) {
	var metadata map[string][]string
	if err := fetchCachedJSON(p.CacheDir, p.MetadataURL, "forge-meta-"+p.NameVal, p.httpClient, &metadata, p.CacheManager); err != nil {
		return nil, fmt.Errorf("fetch %s metadata: %w", p.NameVal, err)
	}

	versions, ok := metadata[mcVersion]
	if !ok {
		return nil, nil
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i] > versions[j]
	})

	result := make([]modloader.LoaderVersion, 0, len(versions))
	for _, v := range versions {
		result = append(result, modloader.LoaderVersion{
			LoaderVersion:    v,
			MinecraftVersion: mcVersion,
			Stable:           !strings.Contains(strings.ToLower(v), "beta") && !strings.Contains(strings.ToLower(v), "alpha") && !strings.Contains(strings.ToLower(v), "pre"),
		})
	}
	return result, nil
}

func (p *AbstractForgeProvider) VersionJsonID(mcVersion, loaderVersion string) string {
	return fmt.Sprintf("%s-%s", p.NameVal, loaderVersion)
}

func (p *AbstractForgeProvider) ResolveDownload(mcVersion, loaderVersion, instancePath, librariesPath, modloaderCacheDir string) (*modloader.DownloadPlan, error) {
	installerJar := fmt.Sprintf("%s-%s-installer.jar", p.ArtifactName, loaderVersion)
	installerURL := fmt.Sprintf("%s/%s/%s/%s", strings.TrimRight(p.MavenBase, "/"), p.GroupPath, loaderVersion, installerJar)
	dest := filepath.Join(modloaderCacheDir, installerJar)

	if err := os.MkdirAll(modloaderCacheDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir modloader cache: %w", err)
	}

	return &modloader.DownloadPlan{
		Entries: []modloader.DownloadPlanEntry{
			{
				Name:        fmt.Sprintf("%s installer", p.NameVal),
				URL:         installerURL,
				Destination: dest,
				Category:    "modloader_installer",
			},
		},
		RequiresInstaller: true,
		InstallerDest:     dest,
		MinecraftJar:      filepath.Join(instancePath, "versions", mcVersion, mcVersion+".jar"),
	}, nil
}

func (p *AbstractForgeProvider) RunInstaller(sessionId string, plan *modloader.DownloadPlan, mcVersion, loaderVersion, instancePath, librariesPath, minecraftJar string, broadcast func([]byte)) error {
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

func (p *AbstractForgeProvider) BuildExecution(loader *modloader.InstalledLoader, versionsDir, librariesPath string) (*modloader.ExecutionPlan, error) {
	verPath := filepath.Join(versionsDir, loader.VersionJsonID, loader.VersionJsonID+".json")
	data, err := os.ReadFile(verPath)
	if err != nil {
		return nil, fmt.Errorf("read version json: %w", err)
	}

	var profile downloader.VersionJSON
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("parse version json: %w", err)
	}

	var classpath []string
	for _, lib := range profile.Libraries {
		if !downloader.MatchRules(lib.Rules) || downloader.IsNativeLibrary(lib) {
			continue
		}
		cp := p.libToClasspath(lib, librariesPath)
		if cp != "" {
			classpath = append(classpath, cp)
		}
	}

	plan := &modloader.ExecutionPlan{
		MainClass:           profile.MainClass,
		AdditionalClasspath: classpath,
	}

	if profile.Arguments != nil {
		for _, a := range profile.Arguments.JVM {
			if s, ok := a.(string); ok {
				if s == "--module-path" || s == "-p" {
					plan.UseModulePath = true
				}
				plan.AdditionalJVMArgs = append(plan.AdditionalJVMArgs, s)
			}
		}
		for _, a := range profile.Arguments.Game {
			if s, ok := a.(string); ok {
				plan.AdditionalGameArgs = append(plan.AdditionalGameArgs, s)
			}
		}
	}

	return plan, nil
}

func (p *AbstractForgeProvider) libToClasspath(lib downloader.Library, librariesPath string) string {
	if lib.Downloads != nil && lib.Downloads.Artifact != nil && lib.Downloads.Artifact.Path != "" {
		return filepath.Join(librariesPath, lib.Downloads.Artifact.Path)
	}
	if lib.Name != "" {
		parsed := globalutils.MavenPath(lib.Name)
		if parsed != "" {
			return filepath.Join(librariesPath, parsed)
		}
	}
	return ""
}

type ForgeProvider struct{ *AbstractForgeProvider }

func NewForgeProvider(cacheDir string, client *http.Client, cacheMgr *cache.Manager) *ForgeProvider {
	return &ForgeProvider{
		AbstractForgeProvider: &AbstractForgeProvider{
			NameVal:      "forge",
			MetadataURL:  "https://files.minecraftforge.net/net/minecraftforge/forge/maven-metadata.json",
			MavenBase:    "https://maven.minecraftforge.net",
			GroupPath:    "net/minecraftforge/forge",
			ArtifactName: "forge",
			CacheDir:     cacheDir,
			CacheManager: cacheMgr,
			httpClient:   client,
		},
	}
}

var _ modloader.ModLoaderProvider = (*ForgeProvider)(nil)
