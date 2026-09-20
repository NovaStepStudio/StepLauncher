package utils

import (
	"os"
	"path/filepath"
)

// protectedFiles es la lista de archivos que el launcher NUNCA sobrescribe:
// si el archivo ya existe, se deja intacto. Son archivos que pueden pertenecer
// a otros launchers o a la instalacion oficial de Minecraft (p. ej.
// launcher_profiles.json del launcher oficial).
var protectedFiles = map[string]bool{
	"launcher_profiles.json": true,
	"launcher_settings.json": true,
	"launcher.properties":    true,
	"options.txt":            true,
	"servers.dat":            true,
	"usercache.json":         true,
	"usernamecache.json":     true,
	"splashes.txt":           true,
}

// IsProtectedFile indica si un nombre de archivo esta en la lista de archivos
// que no deben sobrescribirse.
func IsProtectedFile(name string) bool {
	return protectedFiles[filepath.Base(name)]
}

// SafeWriteFile escribe el archivo solo si no existe (cuando esta protegido) o
// lo sobrescribe normalmente cuando no lo esta. Devuelve si se escribio.
func SafeWriteFile(path string, data []byte, perm os.FileMode) (bool, error) {
	if IsProtectedFile(path) {
		if _, err := os.Stat(path); err == nil {
			return false, nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return false, err
	}
	if err := os.WriteFile(path, data, perm); err != nil {
		return false, err
	}
	return true, nil
}