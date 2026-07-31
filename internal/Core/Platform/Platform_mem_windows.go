//go:build windows

package platform

import (
	"syscall"
	"unsafe"
)

func totalRAMMB() int64 {
	var mem struct {
		Length          uint32
		MemoryLoad      uint32
		TotalPhys       uint64
		AvailPhys       uint64
		TotalPageFile   uint64
		AvailPageFile   uint64
		TotalVirtual    uint64
		AvailVirtual    uint64
		ExtendedVirtual uint64
	}
	mem.Length = uint32(unsafe.Sizeof(mem))
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GlobalMemoryStatusEx")
	ret, _, _ := proc.Call(uintptr(unsafe.Pointer(&mem)))
	if ret == 0 {
		return 4096
	}
	return int64(mem.TotalPhys / 1024 / 1024)
}
