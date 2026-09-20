//go:build windows

package engine

import (
	"os/exec"
	"syscall"
)

const (
	detachedProc     = 0x00000008
	createNewProcGrp = 0x00000200
)

func launchUpdater(updaterPath string) error {
	cmd := exec.Command(updaterPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: createNewProcGrp | detachedProc,
	}
	return cmd.Start()
}