package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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
		javaName = "java.exe"
	}
	javaPath := filepath.Join(javaDir, "bin", javaName)

	if _, err := os.Stat(javaPath); err != nil {
		return "", fmt.Errorf("official Java not found at %s", javaPath)
	}

	return javaPath, nil
}

func resolveCustomJava(javaExec string) (string, error) {
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
