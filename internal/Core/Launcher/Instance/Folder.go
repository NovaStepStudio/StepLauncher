package instance

import (
	"StepLauncher/internal/Core/Launcher/Utils"
)

// openInExplorer abre una ruta en el explorador de archivos del sistema.
func openInExplorer(path string) error {
	return utils.OpenInExplorer(path)
}