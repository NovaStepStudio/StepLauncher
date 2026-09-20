//go:build darwin

package platform

import "golang.org/x/sys/unix"

func totalRAMMB() int64 {
	val, err := unix.SysctlUint64("hw.memsize")
	if err != nil {
		return 4096
	}
	return int64(val / 1024 / 1024)
}