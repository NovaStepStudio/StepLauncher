//go:build !windows

package RichPresence

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func dialDiscord() (net.Conn, error) {
	var candidates []string
	base := ""
	if _, err := os.Stat("/run/user/1000/snap.discord"); err == nil {
		base = "/run/user/1000/snap.discord"
	} else if _, err := os.Stat("/run/user/1000/.flatpak/com.discordapp.Discord/xdg-run"); err == nil {
		base = "/run/user/1000/.flatpak/com.discordapp.Discord/xdg-run"
	} else {
		base = os.Getenv("XDG_RUNTIME_DIR")
		if base == "" {
			base = os.Getenv("TMPDIR")
		}
		if base == "" {
			base = "/tmp"
		}
	}
	for i := 0; i < 10; i++ {
		candidates = append(candidates, filepath.Join(base, fmt.Sprintf("discord-ipc-%d", i)))
	}
	var lastErr error
	for _, addr := range candidates {
		conn, err := net.DialTimeout("unix", addr, dialTimeout)
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
