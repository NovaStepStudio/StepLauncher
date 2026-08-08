//go:build windows

package RichPresence

import (
	"fmt"
	"net"

	"github.com/Microsoft/go-winio"
)

func dialDiscord() (net.Conn, error) {
	timeout := dialTimeout
	var lastErr error
	for i := 0; i < 10; i++ {
		addr := fmt.Sprintf(`\\.\pipe\discord-ipc-%d`, i)
		conn, err := winio.DialPipe(addr, &timeout)
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
