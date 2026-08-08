package utils

import (
	"fmt"
	"strings"
)

func MavenPath(name string) string {
	parts := strings.Split(name, ":")
	if len(parts) < 3 {
		return ""
	}
	group := strings.ReplaceAll(parts[0], ".", "/")
	artifact := parts[1]
	version := parts[2]
	classifier := ""
	if len(parts) >= 4 {
		classifier = "-" + parts[3]
	}
	return fmt.Sprintf("%s/%s/%s/%s-%s%s.jar", group, artifact, version, artifact, version, classifier)
}
