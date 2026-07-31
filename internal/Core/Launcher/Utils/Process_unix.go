//go:build !windows

package utils

import (
	"os/exec"
	"syscall"
)

func setDetachedAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
