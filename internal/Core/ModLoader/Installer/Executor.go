package installer

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"StepLauncher/internal/Core/Utils"
)

type InstallProfile struct {
	Spec      int              `json:"spec"`
	Version   string           `json:"version"`
	Minecraft string           `json:"minecraft"`
	Libraries []ProfileLibrary `json:"libraries"`
}

type ProfileLibrary struct {
	Name      string `json:"name"`
	Downloads *struct {
		Artifact *struct {
			Path string `json:"path"`
			URL  string `json:"url"`
			SHA1 string `json:"sha1"`
			Size int64  `json:"size"`
		} `json:"artifact"`
	} `json:"downloads"`
	URL string `json:"url"`
}

type LibraryEntry struct {
	Name string
	URL  string
	Dest string
	SHA1 string
	Size int64
}

type BundledLib struct {
	SourcePath string
	DestPath   string
}

func ExecuteInstaller(installerJar, versionID, instancePath, librariesPath string) ([]LibraryEntry, error) {
	r, err := zip.OpenReader(installerJar)
	if err != nil {
		return nil, fmt.Errorf("open installer jar: %w", err)
	}
	defer r.Close()

	var versionJsonData []byte
	var profileData []byte
	var bundledLibs []BundledLib

	for _, f := range r.File {
		name := f.Name
		switch {
		case name == "version.json":
			versionJsonData, err = readZipFile(f)
			if err != nil {
				return nil, fmt.Errorf("read version.json: %w", err)
			}
		case name == "install_profile.json":
			profileData, err = readZipFile(f)
			if err != nil {
				return nil, fmt.Errorf("read install_profile.json: %w", err)
			}
		case strings.HasPrefix(name, "maven/"):
			if !f.FileInfo().IsDir() {
				relPath := strings.TrimPrefix(name, "maven/")
				bundledLibs = append(bundledLibs, BundledLib{
					SourcePath: name,
					DestPath:   relPath,
				})
			}
		}
	}

	if versionJsonData == nil {
		return nil, fmt.Errorf("version.json not found in installer jar")
	}

	var versionJson map[string]interface{}
	if err := json.Unmarshal(versionJsonData, &versionJson); err != nil {
		return nil, fmt.Errorf("parse version.json: %w", err)
	}

	jsonID, _ := versionJson["id"].(string)
	if jsonID == "" {
		jsonID = versionID
	}

	verDir := filepath.Join(instancePath, "versions", jsonID)
	if err := os.MkdirAll(verDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir version dir: %w", err)
	}

	jsonBytes, _ := json.MarshalIndent(versionJson, "", "  ")
	if err := os.WriteFile(filepath.Join(verDir, jsonID+".json"), jsonBytes, 0644); err != nil {
		return nil, fmt.Errorf("save version.json: %w", err)
	}

	for _, bl := range bundledLibs {
		for _, f := range r.File {
			if f.Name == bl.SourcePath {
				dest := filepath.Join(librariesPath, bl.DestPath)
				if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
					return nil, fmt.Errorf("mkdir lib dir: %w", err)
				}
				rc, err := f.Open()
				if err != nil {
					return nil, fmt.Errorf("open %s: %w", f.Name, err)
				}
				out, err := os.Create(dest)
				if err != nil {
					rc.Close()
					return nil, fmt.Errorf("create %s: %w", dest, err)
				}
				_, err = io.Copy(out, rc)
				rc.Close()
				out.Close()
				if err != nil {
					return nil, fmt.Errorf("extract %s: %w", f.Name, err)
				}
				break
			}
		}
	}

	var profileLibs []LibraryEntry
	if profileData != nil {
		var profile InstallProfile
		if err := json.Unmarshal(profileData, &profile); err == nil {
			for _, pl := range profile.Libraries {
				if pl.Downloads != nil && pl.Downloads.Artifact != nil && pl.Downloads.Artifact.URL != "" {
					a := pl.Downloads.Artifact
					profileLibs = append(profileLibs, LibraryEntry{
						Name: pl.Name,
						URL:  a.URL,
						Dest: filepath.Join(librariesPath, a.Path),
						SHA1: a.SHA1,
						Size: a.Size,
					})
				} else if pl.URL != "" && pl.Name != "" {
					path := utils.MavenPath(pl.Name)
					profileLibs = append(profileLibs, LibraryEntry{
						Name: pl.Name,
						URL:  strings.TrimRight(pl.URL, "/") + "/" + path,
						Dest: filepath.Join(librariesPath, path),
					})
				}
			}
		}
	}

	return profileLibs, nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}
