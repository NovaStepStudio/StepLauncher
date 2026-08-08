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
	parts := strings.SplitN(neoVersion, ".", 3)
	if len(parts) < 2 {
		return ""
	}
	if _, err := strconv.Atoi(parts[0]); err != nil {
		return ""
	}
	mc := "1." + parts[0]
	if parts[1] != "0" {
		mc += "." + parts[1]
	}
	return mc
}
