package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	globalutils "StepLauncher/internal/Core/Utils"
)

func MatchRules(rules []Rule) bool {
	if len(rules) == 0 {
		return true
	}
	hasAllow := false
	for _, r := range rules {
		if r.Action == "allow" {
			hasAllow = true
			break
		}
	}
	if !hasAllow {
		for _, r := range rules {
			if r.Action == "disallow" && r.OS != nil && r.OS.Name == globalutils.OsName() {
				return false
			}
		}
		return true
	}
	allow := false
	for _, r := range rules {
		switch r.Action {
		case "allow":
			pass := true
			if r.OS != nil && r.OS.Name != globalutils.OsName() {
				pass = false
			}
			if r.Features != nil {
				if r.Features.IsDemoUser != nil && *r.Features.IsDemoUser {
					pass = false
				}
				if r.Features.HasCustomResolution != nil && *r.Features.HasCustomResolution {
					pass = false
				}
				if r.Features.IsQuickPlay != nil && *r.Features.IsQuickPlay {
					pass = false
				}
			}
			if pass {
				allow = true
			}
		case "disallow":
			if r.OS != nil && r.OS.Name == globalutils.OsName() {
				return false
			}
		}
	}
	return allow
}

func IsNativeLibrary(lib Library) bool {
	os := globalutils.OsName()
	if lib.Natives != nil && lib.Natives[os] != "" {
		return true
	}
	if lib.Downloads == nil {
		return false
	}
	osPrefix := "natives-" + os
	if lib.Downloads.Classifiers != nil {
		for key := range lib.Downloads.Classifiers {
			if strings.HasPrefix(key, osPrefix) {
				return true
			}
		}
	}
	if lib.Downloads.Artifact != nil && lib.Downloads.Artifact.Path != "" {
		if strings.Contains(strings.ToLower(lib.Downloads.Artifact.Path), osPrefix) {
			return true
		}
	}
	if lib.Name != "" {
		name := strings.ToLower(lib.Name)
		if strings.Contains(name, ":"+osPrefix) {
			return true
		}
		if strings.Contains(name, ":natives") &&
			!strings.Contains(name, ":natives-linux") &&
			!strings.Contains(name, ":natives-osx") &&
			!strings.Contains(name, ":natives-windows") &&
			!strings.Contains(name, ":natives-macos") {
			return true
		}
	}
	return false
}

func IsNativeClassifierEntry(lib Library) bool {
	name := lib.Name
	if name == "" {
		return false
	}
	idx := strings.LastIndex(name, ":")
	if idx < 0 {
		return false
	}
	classifier := name[idx+1:]
	return strings.HasPrefix(classifier, "natives-")
}

func ResolveNativeArtifact(lib Library, os, arch string) *Artifact {
	exact := globalutils.NativeClassifierFor(os, arch)
	if exact == "" {
		return nil
	}

	if lib.Name != "" && strings.HasSuffix(strings.ToLower(lib.Name), ":"+exact) {
		if lib.Downloads != nil && lib.Downloads.Artifact != nil && lib.Downloads.Artifact.URL != "" {
			return lib.Downloads.Artifact
		}
		return nil
	}

	if lib.Natives != nil && lib.Downloads != nil && lib.Downloads.Classifiers != nil {
		if raw, ok := lib.Natives[os]; ok && raw != "" {
			archSuffix := "64"
			if arch == "x86" {
				archSuffix = "32"
			}
			if a, ok := lib.Downloads.Classifiers[strings.ReplaceAll(raw, "${arch}", archSuffix)]; ok && a.URL != "" {
				return &a
			}
		}
		if a, ok := lib.Downloads.Classifiers[exact]; ok && a.URL != "" {
			return &a
		}
	}

	return nil
}

func BuildTasks(cfg Config, ver *VersionJSON, version string, filter DownloadFilter) ([]DownloadTask, error) {
	var tasks []DownloadTask

	verDir := filepath.Join(cfg.WorkDir, "versions", version)
	if filter.InstanceVersionDir != "" {
		verDir = filter.InstanceVersionDir
	}
	libDir := filepath.Join(cfg.WorkDir, "libraries")

	os.MkdirAll(verDir, 0755)
	os.MkdirAll(libDir, 0755)
	os.MkdirAll(filepath.Join(cfg.WorkDir, "assets", "objects"), 0755)
	os.MkdirAll(filepath.Join(cfg.WorkDir, "assets", "indexes"), 0755)

	addClient(&tasks, ver, version, verDir, filter)
	addLibraryTasks(&tasks, ver, libDir, filter)
	addNativeTasks(&tasks, ver, libDir, filter)
	addAssetIndex(&tasks, ver, cfg, filter)
	addAssetObjects(&tasks, ver, cfg, filter)
	addJavaRuntime(&tasks, ver, cfg, filter)

	return tasks, nil
}

func addClient(tasks *[]DownloadTask, ver *VersionJSON, version, verDir string, filter DownloadFilter) {
	if !filter.Client || ver.Downloads.Client.URL == "" {
		return
	}
	*tasks = append(*tasks, DownloadTask{
		URL:     ver.Downloads.Client.URL,
		Dest:    filepath.Join(verDir, version+".jar"),
		SHA1:    ver.Downloads.Client.SHA1,
		Size:    ver.Downloads.Client.Size,
		Section: "client",
	})
}

func addLibraryTasks(tasks *[]DownloadTask, ver *VersionJSON, libDir string, filter DownloadFilter) {
	if !filter.Libraries || ver.Libraries == nil {
		return
	}
	for _, lib := range ver.Libraries {
		if !MatchRules(lib.Rules) {
			continue
		}
		if IsNativeClassifierEntry(lib) {
			continue
		}
		if lib.Downloads != nil && lib.Downloads.Artifact != nil {
			a := lib.Downloads.Artifact
			url := a.URL
			if url == "" && a.Path != "" {
				url = LibraryRepositoryBase(lib) + "/" + a.Path
			}
			if url != "" {
				*tasks = append(*tasks, DownloadTask{
					URL:     url,
					Dest:    filepath.Join(libDir, a.Path),
					SHA1:    a.SHA1,
					Size:    a.Size,
					Section: "libraries",
				})
			}
		} else if lib.Name != "" && !IsNativeLibrary(lib) && HasRepositoryFallback(lib) {
			path := globalutils.MavenPath(lib.Name)
			if path != "" {
				*tasks = append(*tasks, DownloadTask{
					URL:     LibraryRepositoryBase(lib) + "/" + path,
					Dest:    filepath.Join(libDir, path),
					Section: "libraries",
				})
			}
		}
	}
}

func addNativeTasks(tasks *[]DownloadTask, ver *VersionJSON, libDir string, filter DownloadFilter) {
	if !filter.Natives || ver.Libraries == nil {
		return
	}
	os := globalutils.OsName()
	arch := globalutils.OsArch()
	for _, lib := range ver.Libraries {
		if !MatchRules(lib.Rules) {
			continue
		}
		if !IsNativeLibrary(lib) {
			continue
		}
		artifact := ResolveNativeArtifact(lib, os, arch)
		if artifact == nil || artifact.URL == "" {
			continue
		}
		destPath := artifact.Path
		if destPath == "" {
			destPath = globalutils.MavenPath(lib.Name)
		}
		*tasks = append(*tasks, DownloadTask{
			URL:     artifact.URL,
			Dest:    filepath.Join(libDir, destPath),
			SHA1:    artifact.SHA1,
			Size:    artifact.Size,
			Section: "natives",
		})
	}
}

func addAssetIndex(tasks *[]DownloadTask, ver *VersionJSON, cfg Config, filter DownloadFilter) {
	if !filter.Assets || ver.AssetIndex.ID == "" {
		return
	}
	indexDest := filepath.Join(cfg.WorkDir, "assets", "indexes", ver.AssetIndex.ID+".json")
	*tasks = append(*tasks, DownloadTask{
		URL:     ver.AssetIndex.URL,
		Dest:    indexDest,
		SHA1:    ver.AssetIndex.SHA1,
		Size:    0,
		Section: "asset_index",
	})
}

func addAssetObjects(tasks *[]DownloadTask, ver *VersionJSON, cfg Config, filter DownloadFilter) {
	if !filter.Assets || ver.AssetIndex.ID == "" {
		return
	}
	assetObjDir := filepath.Join(cfg.WorkDir, "assets", "objects")
	var assetIdx AssetIndexJSON
	if err := FetchJSON(cfg, ver.AssetIndex.URL, "assets/"+ver.AssetIndex.ID, &assetIdx); err != nil {
		return
	}
	for _, obj := range assetIdx.Objects {
		if len(obj.Hash) < 2 {
			continue
		}
		dest := filepath.Join(assetObjDir, obj.Hash[:2], obj.Hash)
		url := fmt.Sprintf("https://resources.download.minecraft.net/%s/%s", obj.Hash[:2], obj.Hash)
		*tasks = append(*tasks, DownloadTask{
			URL:     url,
			Dest:    dest,
			SHA1:    obj.Hash,
			Size:    obj.Size,
			Section: "assets",
		})
	}
}

func addJavaRuntime(tasks *[]DownloadTask, ver *VersionJSON, cfg Config, filter DownloadFilter) {
	if !filter.Java || ver.JavaVersion.Component == "" {
		return
	}
	javaTasks, err := BuildJavaRuntimeTasks(ver, cfg)
	if err != nil {
		return
	}
	*tasks = append(*tasks, javaTasks...)
}

func BuildJavaRuntimeTasks(ver *VersionJSON, cfg Config) ([]DownloadTask, error) {
	if ver.JavaVersion.Component == "" {
		return nil, fmt.Errorf("version has no javaVersion.component")
	}
	var all JavaProducts
	if err := FetchJSON(cfg,
		"https://launchermeta.mojang.com/v1/products/java-runtime/2ec0cc96c44e5a76b9c8b7c39df7210883d12871/all.json",
		"java-products", &all); err != nil {
		return nil, fmt.Errorf("fetch java products: %w", err)
	}
	key := globalutils.OsKey()
	var products map[string][]JavaProduct
	switch key {
	case "windows-x64":
		products = all.WindowsX64
	case "linux":
		products = all.Linux
	default:
		if key == "mac-os-arm64" {
			products = all.MacARM64
		} else {
			products = all.MacOS
		}
	}
	list, ok := products[ver.JavaVersion.Component]
	if !ok || len(list) == 0 {
		return nil, fmt.Errorf("no java product for component %s", ver.JavaVersion.Component)
	}
	var jm JavaManifest
	if err := FetchJSON(cfg, list[0].Manifest.URL, "java/"+ver.JavaVersion.Component+"/"+key, &jm); err != nil {
		return nil, fmt.Errorf("fetch java manifest: %w", err)
	}
	runtimeBase := filepath.Join(cfg.WorkDir, "runtime")
	if cfg.JavaRuntimeDir != "" {
		runtimeBase = cfg.JavaRuntimeDir
	}
	base := filepath.Join(runtimeBase, ver.JavaVersion.Component, key)
	var tasks []DownloadTask
	for relPath, f := range jm.Files {
		fullPath := filepath.Join(base, relPath)
		if f.Type == "directory" {
			os.MkdirAll(fullPath, 0755)
			continue
		}
		if f.Downloads == nil || f.Downloads.Raw == nil {
			continue
		}
		dl := f.Downloads.Raw
		tasks = append(tasks, DownloadTask{
			URL:     dl.URL,
			Dest:    fullPath,
			SHA1:    dl.SHA1,
			Size:    dl.Size,
			Section: "java",
		})
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("no java runtime files for component %s", ver.JavaVersion.Component)
	}
	return tasks, nil
}

func SubstituteVars(template string, vars map[string]string) string {
	result := template
	for k, v := range vars {
		result = strings.ReplaceAll(result, "${"+k+"}", v)
	}
	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return result
}

func BuildVarsMap(workDir, version string) map[string]string {
	verDir := filepath.Join(workDir, "versions", version)
	nativesDir := filepath.Join(verDir, "natives")
	assetsDir := filepath.Join(workDir, "assets")
	libDir := filepath.Join(workDir, "libraries")
	return map[string]string{
		"natives_directory": nativesDir,
		"game_directory":    workDir,
		"version_name":      version,
		"assets_root":       assetsDir,
		"library_directory": libDir,
		"classpath":         "",
		"version_type":      "release",
		"user_properties":   "{}",
		"profile_name":      version,
	}
}
