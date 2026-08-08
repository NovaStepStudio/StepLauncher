package helpers

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Core/Utils"
)

type ClasspathEntry struct {
	Path   string
	Exists bool
}

func BuildClasspath(libraries []downloader.Library, librariesDir, versionsDir, version string) (string, []ClasspathEntry) {
	entries := make([]ClasspathEntry, 0, len(libraries)+1)

	clientJar := filepath.Join(versionsDir, version, version+".jar")
	exists := fileExists(clientJar)
	entries = append(entries, ClasspathEntry{Path: clientJar, Exists: exists})

	for _, lib := range libraries {
		if !downloader.MatchRules(lib.Rules) {
			continue
		}

		if downloader.IsNativeClassifierEntry(lib) {
			continue
		}

if downloader.IsNativeLibrary(lib) {
			if lib.Natives == nil {
				continue
			}
			if lib.Downloads == nil || lib.Downloads.Artifact == nil || lib.Downloads.Artifact.Path == "" {
				continue
			}
		}

		var libPath string
		if lib.Downloads != nil && lib.Downloads.Artifact != nil && lib.Downloads.Artifact.Path != "" {
			libPath = lib.Downloads.Artifact.Path
		} else if lib.Name != "" {
			libPath = utils.MavenPath(lib.Name)
		}

		if libPath == "" {
			continue
		}

		fullPath := filepath.Join(librariesDir, libPath)
		entries = append(entries, ClasspathEntry{Path: fullPath, Exists: fileExists(fullPath)})
	}

	sep := ";"
	if runtime.GOOS != "windows" {
		sep = ":"
	}
	var parts []string
	for _, e := range entries {
		parts = append(parts, e.Path)
	}
	return strings.Join(parts, sep), entries
}

func NativesDir(nativesBaseDir, version string) string {
	return filepath.Join(nativesBaseDir, version, "natives")
}

func NativesSubDirs(nativesDir string) []string {
	var existing []string
	entries, err := os.ReadDir(nativesDir)
	if err != nil {
		return []string{nativesDir}
	}
	for _, e := range entries {
		if e.IsDir() {
			existing = append(existing, filepath.Join(nativesDir, e.Name()))
		}
	}
	if len(existing) == 0 {
		existing = append(existing, nativesDir)
	}
	return existing
}

func RecheckClasspathEntries(entries []ClasspathEntry) []ClasspathEntry {
	updated := make([]ClasspathEntry, len(entries))
	for i, e := range entries {
		updated[i] = ClasspathEntry{Path: e.Path, Exists: fileExists(e.Path)}
	}
	return updated
}

func ResolveLibraryDownload(lib downloader.Library, librariesDir string) (dest, url, sha1 string, size int64) {
	if lib.Downloads != nil && lib.Downloads.Artifact != nil {
		a := lib.Downloads.Artifact
		artifactURL := a.URL
		if artifactURL == "" && a.Path != "" {
			artifactURL = downloader.LibraryRepositoryBase(lib) + "/" + a.Path
		}
		if artifactURL != "" {
			return filepath.Join(librariesDir, a.Path), artifactURL, a.SHA1, a.Size
		}
	}
	if lib.Name != "" && !downloader.IsNativeLibrary(lib) {
		p := utils.MavenPath(lib.Name)
		base := downloader.LibraryRepositoryBase(lib)
		if base == "" {
			return "", "", "", 0
		}
		return filepath.Join(librariesDir, p), strings.TrimRight(base, "/") + "/" + p, "", 0
	}
	return "", "", "", 0
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
