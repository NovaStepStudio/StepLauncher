//go:build windows

package utils

import (
	"os/exec"
	"syscall"
)

const (
	detachedProc     = 0x00000008
	createNewProcGrp = 0x00000200
)

func setDetachedAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: createNewProcGrp | detachedProc,
	}
}
