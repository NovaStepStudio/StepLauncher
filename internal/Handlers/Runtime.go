package Handlers

// FileFilter describe un filtro de extensión para los diálogos de archivo.
type FileFilter struct {
	DisplayName string
	Pattern     string
}

// RuntimeBridge abstrae las operaciones de runtime que el launcher necesita
// (diálogos nativos, abrir navegador y salir de la aplicación). Se inyecta
// desde el bootstrap para mantener este paquete desacoplado de la versión de
// Wails (v2/v3) y facilitar las pruebas.
type RuntimeBridge interface {
	// OpenFileDialog abre un diálogo nativo de selección de archivo.
	OpenFileDialog(title string, filters []FileFilter) (string, error)
	// OpenDirectoryDialog abre un diálogo nativo de selección de carpeta.
	OpenDirectoryDialog(title string) (string, error)
	// BrowserOpenURL abre una URL en el navegador por defecto del sistema.
	BrowserOpenURL(url string)
	// Quit solicita el cierre de la aplicación.
	Quit()
}