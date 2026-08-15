package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenInExplorer abre una ruta (archivo o carpeta) en el explorador de
// archivos del sistema según la plataforma actual.
func OpenInExplorer(path string) error {
	if path == "" {
		return fmt.Errorf("ruta vacía")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("abrir ruta: %w", err)
	}
	return nil
}