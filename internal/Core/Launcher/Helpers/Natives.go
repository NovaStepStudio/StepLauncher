package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Core/Utils"
)

func ExtractNatives(libraries []downloader.Library, librariesDir, nativesDir string) (int, error) {
	if err := os.MkdirAll(nativesDir, 0755); err != nil {
		return 0, fmt.Errorf("mkdir natives: %w", err)
	}

	osName := utils.OsName()
	extracted := 0

	for _, lib := range libraries {
		jarPath := resolveNativeJar(lib, librariesDir, osName)
		if jarPath == "" {
			continue
		}
		n, err := utils.ExtractJarNatives(jarPath, nativesDir)
		if err != nil {
			continue
		}
		extracted += n
	}

	return extracted, nil
}

func resolveNativeJar(lib downloader.Library, librariesDir, osName string) string {
	classifier := resolveClassifier(lib, osName)

	if classifier != "" && lib.Downloads != nil && lib.Downloads.Classifiers != nil {
		if art, ok := lib.Downloads.Classifiers[classifier]; ok && art.Path != "" {
			p := filepath.Join(librariesDir, art.Path)
			if fileExists(p) {
				return p
			}
		}
		for key, art := range lib.Downloads.Classifiers {
			if strings.HasPrefix(key, "natives-"+osName) && art.Path != "" {
				p := filepath.Join(librariesDir, art.Path)
				if fileExists(p) {
					return p
				}
			}
		}
	}

	if lib.Name != "" && strings.Contains(strings.ToLower(lib.Name), ":natives-"+osName) {
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




