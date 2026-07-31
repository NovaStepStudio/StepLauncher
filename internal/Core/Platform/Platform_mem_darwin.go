//go:build darwin

package platform

import "syscall"

func totalRAMMB() int64 {
	val, err := syscall.SysctlUint64("hw.memsize")
	if err != nil {
		return 4096
	}
	return int64(val / 1024 / 1024)
}
