package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func LaunchProcess(javaPath, mainClass, workDir, logPath string, jvmArgs, gameArgs []string, extraEnv map[string]string) (*exec.Cmd, *os.File, error) {
	fullArgs := append(jvmArgs, mainClass)
	fullArgs = append(fullArgs, gameArgs...)

	cmd := exec.Command(javaPath, fullArgs...)

	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), envMapToSlice(extraEnv)...)
	}

	logDir := filepath.Dir(logPath)
	os.MkdirAll(logDir, 0755)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot open log file %s: %w", logPath, err)
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Dir = workDir

	setDetachedAttr(cmd)

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return nil, nil, fmt.Errorf("cannot start process: %w", err)
	}

	return cmd, logFile, nil
}

func envMapToSlice(m map[string]string) []string {
	s := make([]string, 0, len(m))
	for k, v := range m {
		s = append(s, k+"="+v)
	}
	return s
}

func FindJVMCrashLog(gameDir string) string {
	entries, err := os.ReadDir(gameDir)
	if err != nil {
		return ""
	}
	var latest string
	var latestMod int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if !strings.HasPrefix(name, "hs_err_pid") && !strings.HasSuffix(name, ".mdmp") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		mod := info.ModTime().UnixMilli()
		if mod > latestMod {
			latestMod = mod
			latest = filepath.Join(gameDir, e.Name())
		}
	}
	return latest
}

func KillTree(pid int) {
	if pid <= 0 {
		return
	}
	if runtime.GOOS == "windows" {
		exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", pid)).Run()
		return
	}

	exec.Command("kill", "--", fmt.Sprintf("-%d", pid)).Run()
	exec.Command("pkill", "-9", "-P", fmt.Sprintf("%d", pid)).Run()
	exec.Command("kill", "-9", fmt.Sprintf("%d", pid)).Run()
}

func FindLatestCrashReport(gameDir string) string {
	crashDir := filepath.Join(gameDir, "crash-reports")
	entries, err := os.ReadDir(crashDir)
	if err != nil {
		return ""
	}
	var latest string
	var latestMod int64
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".txt") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		mod := info.ModTime().UnixMilli()
		if mod > latestMod {
			latestMod = mod
			latest = filepath.Join(crashDir, e.Name())
		}
	}
	return latest
}

func JoinArgsForLog(args []string) string {
	if len(args) > 25 {
		shown := make([]string, 25)
		copy(shown, args[:25])
		return fmt.Sprintf("%s ... [%d total]", strings.Join(shown, " "), len(args))
	}
	return strings.Join(args, " ")
}
