package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var nativeExts = map[string]bool{
	".dll":    true,
	".so":     true,
	".dylib":  true,
	".jnilib": true,
}

func IsNativeFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return nativeExts[ext]
}

func ExtractNatives(jarPaths []string, nativesDir string) (int, error) {
	if err := os.RemoveAll(nativesDir); err != nil {
		return 0, fmt.Errorf("clean natives dir: %w", err)
	}
	if err := os.MkdirAll(nativesDir, 0755); err != nil {
		return 0, err
	}

	extracted := 0
	var errs []string
	for _, jarPath := range jarPaths {
		n, err := ExtractJarNatives(jarPath, nativesDir)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", filepath.Base(jarPath), err))
			continue
		}
		extracted += n
	}
	if len(errs) > 0 {
		return extracted, fmt.Errorf("extraction errors: %s", strings.Join(errs, "; "))
	}
	return extracted, nil
}

func ExtractJarNatives(jarPath, extractDir string) (int, error) {
	r, err := zip.OpenReader(jarPath)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	extracted := 0
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := f.Name
		if strings.HasPrefix(name, "META-INF") {
			continue
		}
		if !IsNativeFile(name) {
			continue
		}

		fileName := name
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			fileName = name[idx+1:]
		}

		dest := filepath.Join(extractDir, fileName)
		if _, err := os.Stat(dest); err == nil {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			continue
		}

		out, err := os.Create(dest)
		if err != nil {
			rc.Close()
			continue
		}

		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()

		if err != nil {
			os.Remove(dest)
			continue
		}

		extracted++
	}
	return extracted, nil
}
