package provider

import (
	"strconv"
	"strings"
)

type NeoForgeVersionResolver struct{}

func (NeoForgeVersionResolver) FilterVersionsForMinecraft(allVersions []string, mcVersion string) []string {
	var filtered []string
	for _, v := range allVersions {
		if NeoForgeVersionToMcVersion(v) == mcVersion {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func (NeoForgeVersionResolver) NeoForgeVersionToMcVersion(neoVersion string) string {
	return NeoForgeVersionToMcVersion(neoVersion)
}

func NeoForgeVersionToMcVersion(neoVersion string) string {
	// Series 26.x (formato X.Y.Z.BUILD): la MC ya NO lleva el prefijo "1.".
	// Verificado contra install_profile.json reales: 26.2.0.57 -> "26.2"
	// (Z "0" se omite) y 26.1.2.75 -> "26.1.2" (Z se incluye).
	parts := strings.SplitN(neoVersion, ".", 4)
	if len(parts) < 2 {
		return ""
	}
	if _, err := strconv.Atoi(parts[0]); err != nil {
		return ""
	}
	if len(parts) == 4 {
		mc := parts[0] + "." + parts[1]
		if parts[2] != "0" {
			mc += "." + parts[2]
		}
		return mc
	}
	// Series antiguas (X.Y.BUILD): 21.1.234 -> "1.21.1", 21.11.44 -> "1.21.11".
	mc := "1." + parts[0]
	if parts[1] != "0" {
		mc += "." + parts[1]
	}
	return mc
}
