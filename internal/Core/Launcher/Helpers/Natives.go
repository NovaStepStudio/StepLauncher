package helpers

import (
	"os"
	"path/filepath"
	"strings"

	"StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Core/Utils"
)

func nativesLoadSubDir(jvmArgs []interface{}) (loadDir string, allDirs []string) {
	seen := map[string]bool{}
	add := func(d string) {
		if d == "" || seen[d] {
			return
		}
		seen[d] = true
		allDirs = append(allDirs, d)
	}

	for _, arg := range jvmArgs {
		if s, ok := arg.(string); ok {
			if d := nativesSubDirFromString(s); d != "" {
				if strings.HasPrefix(s, "-Djava.library.path=") {
					loadDir = d
				}
				add(d)
			}
		}
		if m, ok := arg.(map[string]interface{}); ok {
			valRaw, ok := m["value"]
			if !ok {
				continue
			}
			switch val := valRaw.(type) {
			case string:
				d := nativesSubDirFromString(val)
				if d != "" {
					if strings.HasPrefix(val, "-Djava.library.path=") {
						loadDir = d
					}
					add(d)
				}
			case []interface{}:
				for _, item := range val {
					if s, ok := item.(string); ok {
						d := nativesSubDirFromString(s)
						if d != "" {
							if strings.HasPrefix(s, "-Djava.library.path=") {
								loadDir = d
							}
							add(d)
						}
					}
				}
			}
		}
	}
	if loadDir == "" && len(allDirs) > 0 {
		loadDir = allDirs[0]
	}
	return loadDir, allDirs
}

func nativesSubDirFromString(arg string) string {
	idx := strings.Index(arg, "${natives_directory}/")
	if idx < 0 {
		return ""
	}
	rest := arg[idx+len("${natives_directory}/"):]
	sub := rest
	if i := strings.IndexAny(rest, "\" "); i >= 0 {
		sub = rest[:i]
	}
	if i := strings.IndexByte(sub, '/'); i >= 0 {
		sub = sub[:i]
	}
	if sub == "" {
		return ""
	}
	return sub
}

func ExtractNatives(libraries []downloader.Library, librariesDir, nativesDir string, jvmArgs []interface{}, onProgress func(current, total int, name string)) (int, error) {
	osName := utils.OsName()

	var jarPaths []string
	for _, lib := range libraries {
		jarPath := resolveNativeJar(lib, librariesDir, osName)
		if jarPath == "" {
			continue
		}
		jarPaths = append(jarPaths, jarPath)
	}
	if len(jarPaths) == 0 {
		if onProgress != nil {
			onProgress(0, 0, "")
		}
		return 0, nil
	}

	if err := os.MkdirAll(nativesDir, 0755); err != nil {
		return 0, err
	}

	loadDir, workDirs := nativesLoadSubDir(jvmArgs)
	extractDir := nativesDir
	if loadDir != "" {
		extractDir = filepath.Join(nativesDir, loadDir)
		if err := os.MkdirAll(extractDir, 0755); err != nil {
			return 0, err
		}
		removeStaleNatives(nativesDir)
	}

	for _, d := range workDirs {
		dir := filepath.Join(nativesDir, d)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return 0, err
		}
		if d != loadDir {
			removeStaleNatives(dir)
		}
	}

	total := len(jarPaths)
	if onProgress != nil {
		onProgress(0, total, "")
	}
	extracted := 0
	for i, jarPath := range jarPaths {
		if onProgress != nil {
			onProgress(i+1, total, filepath.Base(jarPath))
		}
		n, err := utils.ExtractJarNatives(jarPath, extractDir)
		if err != nil {
			continue
		}
		extracted += n
	}
	if onProgress != nil {
		onProgress(total, total, "")
	}
	return extracted, nil
}

func removeStaleNatives(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if utils.IsNativeFile(e.Name()) {
			os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

func resolveNativeJar(lib downloader.Library, librariesDir, osName string) string {
	if suffix := utils.NativeClassifier(); suffix != "" && strings.HasSuffix(lib.Name, ":"+suffix) {
		if lib.Downloads != nil && lib.Downloads.Artifact != nil && lib.Downloads.Artifact.Path != "" {
			p := filepath.Join(librariesDir, lib.Downloads.Artifact.Path)
			if fileExists(p) {
				return p
			}
		}
		p := filepath.Join(librariesDir, utils.MavenPath(lib.Name))
		if fileExists(p) {
			return p
		}
		return ""
	}

	classifier := resolveClassifier(lib, osName)
	if classifier != "" && lib.Downloads != nil && lib.Downloads.Classifiers != nil {
		if art, ok := lib.Downloads.Classifiers[classifier]; ok && art.Path != "" {
			p := filepath.Join(librariesDir, art.Path)
			if fileExists(p) {
				return p
			}
		}
		if suffix := utils.NativeClassifier(); suffix != "" {
			if art, ok := lib.Downloads.Classifiers[suffix]; ok && art.Path != "" {
				p := filepath.Join(librariesDir, art.Path)
				if fileExists(p) {
					return p
				}
			}
		}
	}

	return ""
}

func resolveClassifier(lib downloader.Library, osName string) string {
	if lib.Natives == nil {
		return ""
	}
	raw, ok := lib.Natives[osName]
	if !ok || raw == "" {
		return ""
	}
	arch := utils.OsArch()
	archNum := "64"
	if arch == "x86" {
		archNum = "32"
	}
	return strings.ReplaceAll(raw, "${arch}", archNum)
}
