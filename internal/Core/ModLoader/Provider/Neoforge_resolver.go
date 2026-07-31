package provider

import (
	"strconv"
	"strings"
)

type NeoForgeVersionResolver struct{}

func (NeoForgeVersionResolver) FilterVersionsForMinecraft(allVersions []string, mcVersion string) []string {
	parts := strings.SplitN(mcVersion, ".", 3)
	if len(parts) < 2 {
		return nil
	}
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])

	var prefix string
	if major >= 1 && minor >= 26 {
		prefix = mcVersion
	} else if major == 1 {
		prefix = strconv.Itoa(minor)
	}

	if prefix == "" {
		return nil
	}

	var filtered []string
	for _, v := range allVersions {
		if strings.HasPrefix(v, prefix) || v == mcVersion {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func (NeoForgeVersionResolver) NeoForgeVersionToMcVersion(neoVersion string) string {
	parts := strings.SplitN(neoVersion, ".", 3)
	if len(parts) < 2 {
		return ""
	}
	first, _ := strconv.Atoi(parts[0])
	if first >= 26 {
		return strings.Join(parts[:2], ".")
	}
	return "1." + parts[0]
}
