//go:build !windows

package engine

import (
	"os/exec"
	"syscall"
)

func launchUpdater(updaterPath string) error {
	cmd := exec.Command(updaterPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return cmd.Start()
}