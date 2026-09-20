//go:build linux

package platform

import (
	"os"
	"strconv"
	"strings"
)

func totalRAMMB() int64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 4096
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, _ := strconv.ParseInt(fields[1], 10, 64)
				return kb / 1024
			}
		}
	}
	return 4096
}
