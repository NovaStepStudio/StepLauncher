package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"StepLauncher/internal/Core/Cache"
	"StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Core/ModLoader"
	globalutils "StepLauncher/internal/Core/Utils"
)



type fabricVersion struct {
	Loader struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	} `json:"loader"`
}

type AbstractFabricProvider struct {
	modloader.BaseProvider
	NameVal     string
	MetaBase    string
	MavenBase   string
	CacheDir    string
	CacheManager *cache.Manager
	HttpClient  *http.Client
}

func (p *AbstractFabricProvider) Name() string { return p.NameVal }

func (p *AbstractFabricProvider) GetVersions(mcVersion string) ([]modloader.LoaderVersion, error) {
	url := fmt.Sprintf("%s/versions/loader/%s", p.MetaBase, mcVersion)
	cacheKey := fmt.Sprintf("fabric-versions-%s-%s", p.NameVal, mcVersion)

	var raw []fabricVersion
	if err := fetchCachedJSON(p.CacheDir, url, cacheKey, p.HttpClient, &raw, p.CacheManager); err != nil {
		return nil, fmt.Errorf("fetch %s versions: %w", p.NameVal, err)
	}

	versions := make([]modloader.LoaderVersion, 0, len(raw))
	for _, v := range raw {
		versions = append(versions, modloader.LoaderVersion{
			LoaderVersion:    v.Loader.Version,
			MinecraftVersion: mcVersion,
			Stable:           v.Loader.Stable,
		})
	}
	return versions, nil
}

func (p *AbstractFabricProvider) VersionJsonID(mcVersion, loaderVersion string) string {
	return fmt.Sprintf("%s-%s-%s", p.NameVal, mcVersion, loaderVersion)
}

func (p *AbstractFabricProvider) ResolveDownload(mcVersion, loaderVersion, instancePath, librariesPath, _ string) (*modloader.DownloadPlan, error) {
	profileURL := fmt.Sprintf("%s/versions/loader/%s/%s/profile/json", p.MetaBase, mcVersion, loaderVersion)
	versionID := p.VersionJsonID(mcVersion, loaderVersion)
	cacheKey := fmt.Sprintf("fabric-profile-%s", versionID)

	var profile downloader.VersionJSON
	if err := fetchCachedJSON(p.CacheDir, profileURL, cacheKey, p.HttpClient, &profile, p.CacheManager); err != nil {
		return nil, fmt.Errorf("fetch profile: %w", err)
	}

	verDir := filepath.Join(instancePath, "versions", versionID)
	verPath := filepath.Join(verDir, versionID+".json")
	if err := os.MkdirAll(verDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir version: %w", err)
	}
	data, _ := json.MarshalIndent(profile, "", "  ")
	if err := os.WriteFile(verPath, data, 0644); err != nil {
		return nil, fmt.Errorf("save version json: %w", err)
	}

	var entries []modloader.DownloadPlanEntry
	for _, lib := range profile.Libraries {
		if !downloader.MatchRules(lib.Rules) {
			continue
		}
		entry := p.libToDownloadEntry(lib, librariesPath)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	return &modloader.DownloadPlan{Entries: entries}, nil
}

func (p *AbstractFabricProvider) BuildExecution(loader *modloader.InstalledLoader, versionsDir, librariesPath string) (*modloader.ExecutionPlan, error) {
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

func (p *AbstractFabricProvider) libToDownloadEntry(lib downloader.Library, librariesPath string) *modloader.DownloadPlanEntry {
	if downloader.IsNativeLibrary(lib) {
		return nil
	}
	path, url, sha1, size := resolveLib(lib, p.MavenBase)
	if path == "" || url == "" {
		return nil
	}
	return &modloader.DownloadPlanEntry{
		Name:        lib.Name,
		URL:         url,
		Destination: filepath.Join(librariesPath, path),
		Size:        size,
		SHA1:        sha1,
		Category:    "modloader_library",
	}
}

func (p *AbstractFabricProvider) libToClasspath(lib downloader.Library, librariesPath string) string {
	path, _, _, _ := resolveLib(lib, p.MavenBase)
	if path == "" {
		return ""
	}
	return filepath.Join(librariesPath, path)
}

func resolveLib(lib downloader.Library, mavenBase string) (path, url, sha1 string, size int64) {
	if lib.Downloads != nil && lib.Downloads.Artifact != nil && lib.Downloads.Artifact.URL != "" {
		a := lib.Downloads.Artifact
		return a.Path, a.URL, a.SHA1, a.Size
	}
	if lib.Name != "" {
		parsed := globalutils.MavenPath(lib.Name)
		base := mavenBase
		if lib.URL != "" {
			base = lib.URL
		}
		return parsed, strings.TrimRight(base, "/") + "/" + parsed, "", 0
	}
	return "", "", "", 0
}

type FabricProvider struct{ *AbstractFabricProvider }

func NewFabricProvider(cacheDir string, client *http.Client, cacheMgr *cache.Manager) *FabricProvider {
	return &FabricProvider{
		AbstractFabricProvider: &AbstractFabricProvider{
			NameVal:      "fabric",
			MetaBase:     "https://meta.fabricmc.net/v2",
			MavenBase:    "https://maven.fabricmc.net/",
			CacheDir:     cacheDir,
			CacheManager: cacheMgr,
			HttpClient:   client,
		},
	}
}

func fetchCachedJSON(cacheDir, url, cacheKey string, client *http.Client, out interface{}, mgr *cache.Manager) error {
	if mgr != nil {
		cfg := downloader.Config{
			CacheManager: mgr,
			HTTPClient:   client,
		}
		return downloader.FetchJSON(cfg, url, cacheKey, out)
	}
	cfg := downloader.Config{
		CacheDir:   filepath.Join(cacheDir, "modloader-cache"),
		HTTPClient: client,
	}
	return downloader.FetchJSON(cfg, url, cacheKey, out)
}
