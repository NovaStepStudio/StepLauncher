package modloader

import (
	"path/filepath"
	"strings"
)

type MavenCoordinate struct {
	Group      string
	Artifact   string
	Version    string
	Classifier string
	Extension  string
}

func ParseMavenCoordinate(coord string) *MavenCoordinate {
	m := &MavenCoordinate{Extension: "jar"}

	if at := strings.LastIndexByte(coord, '@'); at >= 0 {
		m.Extension = coord[at+1:]
		coord = coord[:at]
	}

	parts := strings.Split(coord, ":")
	if len(parts) < 3 {
		return nil
	}
	m.Group = parts[0]
	m.Artifact = parts[1]
	m.Version = parts[2]

	if len(parts) >= 4 {
		m.Classifier = parts[3]
	}

	return m
}

func (m *MavenCoordinate) ToPath() string {
	groupPath := strings.ReplaceAll(m.Group, ".", "/")
	base := m.Artifact + "-" + m.Version
	if m.Classifier != "" {
		base += "-" + m.Classifier
	}
	return groupPath + "/" + m.Artifact + "/" + m.Version + "/" + base + "." + m.Extension
}

func (m *MavenCoordinate) ToLocalPath(baseDir string) string {
	return filepath.Join(baseDir, m.ToPath())
}

func (m *MavenCoordinate) ToRemoteURL(repoBase string) string {
	repo := strings.TrimRight(repoBase, "/")
	return repo + "/" + m.ToPath()
}
