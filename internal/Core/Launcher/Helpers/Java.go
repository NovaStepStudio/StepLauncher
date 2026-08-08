package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"StepLauncher/internal/Core/Utils"
)

func ResolveJava(component, runtimeDir string, useOfficialJava bool, javaExec string) (string, error) {
	if useOfficialJava {
		return resolveOfficialJava(component, runtimeDir)
	}
	return resolveCustomJava(javaExec)
}

func resolveOfficialJava(component, runtimeDir string) (string, error) {
	if component == "" {
		return "", fmt.Errorf("version has no javaVersion.component")
	}

	osKey := utils.OsKey()
	javaDir := filepath.Join(runtimeDir, component, osKey)
	javaName := "java"
	if runtime.GOOS == "windows" {
		javaName = "javaw.exe"
	}
	javaPath := filepath.Join(javaDir, "bin", javaName)

if _, err := os.Stat(javaPath); err != nil {
		alt := filepath.Join(javaDir, "bin", "java.exe")
		if runtime.GOOS == "windows" {
			if _, err2 := os.Stat(alt); err2 == nil {
				return alt, nil
			}
		}
		return "", fmt.Errorf("official Java not found at %s", javaPath)
	}

	return javaPath, nil
}

func resolveCustomJava(javaExec string) (string, error) {
	if runtime.GOOS == "windows" {
		if javaExec == "" || javaExec == "java" || javaExec == "java.exe" {
			if p, err := exec.LookPath("javaw"); err == nil {
				return p, nil
			}
			if p, err := exec.LookPath("java"); err == nil {
				return p, nil
			}
			return "", fmt.Errorf("Java not found (se busco javaw/java en el sistema)")
		}

		dir := javaExec
		base := ""
		if filepath.Ext(javaExec) != "" {
			dir = filepath.Dir(javaExec)
			base = filepath.Base(javaExec)
		}
		if base == "" || strings.EqualFold(base, "java") || strings.EqualFold(base, "java.exe") {
			javawPath := filepath.Join(dir, "javaw.exe")
			if _, err := os.Stat(javawPath); err == nil {
				return javawPath, nil
			}
		}

		if _, err := os.Stat(javaExec); err == nil {
			return javaExec, nil
		}
		if p, err := exec.LookPath(javaExec); err == nil {
			return p, nil
		}
		return "", fmt.Errorf("Java not found: %s", javaExec)
	}

	if javaExec == "" {
		javaExec = "java"
	}
	path, err := exec.LookPath(javaExec)
	if err != nil {
		return "", fmt.Errorf("Java not found: %s", javaExec)
	}
	return path, nil
}

func GetJavaVersion(javaPath string) (string, error) {
	cmd := exec.Command(javaPath, "-version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("get Java version: %w", err)
	}
	return string(out), nil
}

func DetectJavaMajorVersion(javaPath string) int {
	out, err := exec.Command(javaPath, "-version").CombinedOutput()
	if err != nil {
		return 8
	}
	output := string(out)
	var major int
	if _, err := fmt.Sscanf(output, `java version "1.%d`, &major); err == nil {
		return major
	}
	if _, err := fmt.Sscanf(output, `java version "%d`, &major); err == nil {
		return major
	}
	if _, err := fmt.Sscanf(output, `openjdk version "%d`, &major); err == nil {
		return major
	}
	return 8
}
