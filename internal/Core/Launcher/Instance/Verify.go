package instance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Core/Downloader/Utils"
)

func (m *InstanceManager) VerifyInstance(name string) ([]VerifyResult, error) {
	meta, err := m.readMetadata(name)
	if err != nil {
		return nil, fmt.Errorf("instance %s not found", name)
	}

	var results []VerifyResult
	instPath, err := m.instancePath(name)
	if err != nil {
		return nil, err
	}
	instVerDir := filepath.Join(instPath, "versions")

	dlCfg := downloader.Config{
		WorkDir:    m.sharedDir,
		CacheDir:   filepath.Join(m.sharedDir, "cache"),
		HTTPClient: downloader.DefaultHTTPClient(),
	}

	for _, version := range meta.Versions {
		result := VerifyResult{Version: version, Valid: true}
		verDir := filepath.Join(instVerDir, version)

		verJSON := filepath.Join(verDir, version+".json")
		if _, err := os.Stat(verJSON); err != nil {
			result.Issues = append(result.Issues, VerifyIssue{Type: "missing", File: verJSON, Message: "version JSON not found"})
			result.Valid = false
			results = append(results, result)
			continue
		}

		var ver downloader.VersionJSON
		data, err := os.ReadFile(verJSON)
		if err != nil {
			result.Issues = append(result.Issues, VerifyIssue{Type: "error", File: verJSON, Message: fmt.Sprintf("cannot read: %v", err)})
			result.Valid = false
			results = append(results, result)
			continue
		}
		if err := json.Unmarshal(data, &ver); err != nil {
			result.Issues = append(result.Issues, VerifyIssue{Type: "error", File: verJSON, Message: fmt.Sprintf("cannot parse: %v", err)})
			result.Valid = false
			results = append(results, result)
			continue
		}

		clientJar := filepath.Join(verDir, version+".jar")
		if _, err := os.Stat(clientJar); err != nil {
			result.Issues = append(result.Issues, VerifyIssue{Type: "missing", File: clientJar, Message: "client JAR not found"})
			result.Valid = false
		} else if ver.Downloads.Client.SHA1 != "" {
			if ok, err := utils.VerifySHA1(clientJar, ver.Downloads.Client.SHA1); err != nil || !ok {
				result.Issues = append(result.Issues, VerifyIssue{Type: "corrupt", File: clientJar, Message: "SHA-1 mismatch"})
				result.Valid = false
			}
		}

		filter := downloader.DownloadFilter{Version: version, Client: false, Libraries: true, Natives: true, Assets: true, Java: false}
		allTasks, err := downloader.BuildTasks(dlCfg, &ver, version, filter)
		if err != nil {
			result.Issues = append(result.Issues, VerifyIssue{Type: "error", Message: fmt.Sprintf("build tasks: %v", err)})
			result.Valid = false
			results = append(results, result)
			continue
		}

		for _, t := range allTasks {
			if t.SHA1 == "" {
				continue
			}
			if !utils.FileExists(t.Dest) {
				result.Issues = append(result.Issues, VerifyIssue{Type: "missing", File: t.Dest, Message: fmt.Sprintf("file not found (%s)", t.Section)})
				result.Valid = false
				continue
			}
			if ok, err := utils.VerifySHA1(t.Dest, t.SHA1); err != nil || !ok {
				result.Issues = append(result.Issues, VerifyIssue{Type: "corrupt", File: t.Dest, Message: fmt.Sprintf("SHA-1 mismatch (%s)", t.Section)})
				result.Valid = false
			}
		}

		if _, err := os.Stat(filepath.Join(verDir, "natives")); err != nil {
			result.Issues = append(result.Issues, VerifyIssue{Type: "missing", File: filepath.Join(verDir, "natives"), Message: "natives not extracted"})
		}

		results = append(results, result)
	}

	if len(results) == 0 {
		results = append(results, VerifyResult{Version: "", Valid: true, Issues: []VerifyIssue{}})
	}
	return results, nil
}

func (m *InstanceManager) VerifySingleVersion(name, version string) (*VerifyResult, error) {
	meta, err := m.readMetadata(name)
	if err != nil {
		return nil, fmt.Errorf("instance %s not found", name)
	}
	found := false
	for _, v := range meta.Versions {
		if v == version {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("version %s not found in instance %s", version, name)
	}
	results, err := m.VerifyInstance(name)
	if err != nil {
		return nil, err
	}
	for _, r := range results {
		if r.Version == version {
			return &r, nil
		}
	}
	return &VerifyResult{Version: version, Valid: true}, nil
}
