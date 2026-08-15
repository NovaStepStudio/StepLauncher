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
	Spec       int               `json:"spec"`
	Version    string            `json:"version"`
	Minecraft  string            `json:"minecraft"`
	Libraries  []ProfileLibrary  `json:"libraries"`
	Processors []json.RawMessage `json:"processors"`
	// VersionInfo: los instaladores legacy de Forge (1.8.9 y anteriores) no
	// traen un version.json aparte en el jar; el contenido del version.json va
	// embebido como objeto en este campo de install_profile.json.
	VersionInfo json.RawMessage `json:"versionInfo"`
	// Install: sección del formato clásico de instalador (era 1.7.x-1.8.x).
	// filePath es el archivo contenido en la raíz del jar que el instalador
	// oficial extrae como librería del perfil (p. ej. el universal jar) y path
	// su coordenada maven de destino (VersionInfo.extractFile/getLibraryPath).
	Install *InstallSection `json:"install"`
}

// InstallSection replica la sección "install" de install_profile.json de los
// instaladores clásicos de Forge (1.7.x-1.8.x).
type InstallSection struct {
	ProfileName string `json:"profileName"`
	Target      string `json:"target"`
	Path        string `json:"path"`
	FilePath    string `json:"filePath"`
	Minecraft   string `json:"minecraft"`
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

// ExecuteInstaller extrae del jar del instalador el version.json (lo escribe en
// versions/<id>/), las librerías empaquetadas en maven/ y devuelve: las
// librerías del install_profile.json descargables por URL, el id real del
// version.json (vacío si no lo tiene) y si el perfil declara procesadores
// (binarypatcher/PROCESS_MINECRAFT_JAR) que requieren ejecutar el instalador
// oficial con Java para generar los jars parcheados.
func ExecuteInstaller(installerJar, versionID, instancePath, librariesPath string) ([]LibraryEntry, string, bool, error) {
	r, err := zip.OpenReader(installerJar)
	if err != nil {
		return nil, "", false, fmt.Errorf("open installer jar: %w", err)
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
				return nil, "", false, fmt.Errorf("read version.json: %w", err)
			}
		case name == "install_profile.json":
			profileData, err = readZipFile(f)
			if err != nil {
				return nil, "", false, fmt.Errorf("read install_profile.json: %w", err)
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
		// Instaladores legacy (Forge 1.8.9 y anteriores): no hay version.json;
		// el version.json va embebido como "versionInfo" en install_profile.json.
		var profile InstallProfile
		if err := json.Unmarshal(profileData, &profile); err == nil && len(profile.VersionInfo) > 0 {
			versionJsonData = profile.VersionInfo
		}
	}

	if versionJsonData == nil {
		return nil, "", false, fmt.Errorf("version.json not found in installer jar")
	}

	var versionJson map[string]interface{}
	if err := json.Unmarshal(versionJsonData, &versionJson); err != nil {
		return nil, "", false, fmt.Errorf("parse version.json: %w", err)
	}

	jsonID, _ := versionJson["id"].(string)
	if jsonID == "" {
		jsonID = versionID
	}

	verDir := filepath.Join(instancePath, "versions", jsonID)
	if err := os.MkdirAll(verDir, 0755); err != nil {
		return nil, "", false, fmt.Errorf("mkdir version dir: %w", err)
	}

	jsonBytes, _ := json.MarshalIndent(versionJson, "", "  ")
	if err := os.WriteFile(filepath.Join(verDir, jsonID+".json"), jsonBytes, 0644); err != nil {
		return nil, "", false, fmt.Errorf("save version.json: %w", err)
	}

	for _, bl := range bundledLibs {
		for _, f := range r.File {
			if f.Name == bl.SourcePath {
				dest := filepath.Join(librariesPath, bl.DestPath)
				if err := extractZipEntryTo(f, dest); err != nil {
					return nil, "", false, err
				}
				break
			}
		}
	}

	hasProcessors := false
	var profileLibs []LibraryEntry
	if profileData != nil {
		var profile InstallProfile
		if err := json.Unmarshal(profileData, &profile); err == nil {
			hasProcessors = len(profile.Processors) > 0

			// Protocolo legacy (era 1.7.x-1.8.x): el instalador oficial extrae
			// el archivo contenido (install.filePath, p. ej.
			// forge-1.8.9-11.15.1.2318-1.8.9-universal.jar) y lo deja en
			// libraries/ bajo la coordenada maven install.path
			// (VersionInfo.extractFile -> getLibraryPath). Sin esto, la librería
			// principal del perfil (net.minecraftforge:forge:...) nunca existe en
			// disco y la versión no arranca.
			if sec := profile.Install; sec != nil && sec.FilePath != "" && sec.Path != "" {
				for _, f := range r.File {
					name := f.Name
					if name == sec.FilePath || strings.TrimPrefix(name, "/") == sec.FilePath {
						dest := filepath.Join(librariesPath, utils.MavenPath(sec.Path))
						if err := extractZipEntryTo(f, dest); err != nil {
							return nil, "", false, err
						}
						break
					}
				}
			}

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

	return profileLibs, jsonID, hasProcessors, nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

// extractZipEntryTo extrae una entrada del zip del instalador a dest.
func extractZipEntryTo(f *zip.File, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("mkdir lib dir: %w", err)
	}
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open %s: %w", f.Name, err)
	}
	out, err := os.Create(dest)
	if err != nil {
		rc.Close()
		return fmt.Errorf("create %s: %w", dest, err)
	}
	_, err = io.Copy(out, rc)
	rc.Close()
	out.Close()
	if err != nil {
		return fmt.Errorf("extract %s: %w", f.Name, err)
	}
	return nil
}
