package downloader

import (
	"strings"

	globalutils "StepLauncher/internal/Core/Utils"
)

// NormalizeVersion completa los campos de metadatos que algunos version.json
// de terceros (p. ej. BatMod) omiten: los artifacts y classifiers de
// downloads se declaran solo con url/sha1/size, sin "path". El resto del
// launcher asume que path existe para localizar, verificar y extraer las
// librerías (y los jars nativos), así que se deriva de la coordenada maven.
// Debe llamarse tras cargar o fusionar un VersionJSON, antes de construir
// tareas de descarga, classpath o extracción de natives.
func NormalizeVersion(ver *VersionJSON) {
	if ver == nil {
		return
	}
	for i := range ver.Libraries {
		normalizeLibrary(&ver.Libraries[i])
	}
}

func normalizeLibrary(lib *Library) {
	if lib == nil || lib.Name == "" || lib.Downloads == nil {
		return
	}

	if lib.Downloads.Artifact != nil && lib.Downloads.Artifact.Path == "" {
		if p := globalutils.MavenPath(lib.Name); p != "" {
			lib.Downloads.Artifact.Path = p
		}
	}

	if len(lib.Downloads.Classifiers) == 0 {
		return
	}
	for classifier, art := range lib.Downloads.Classifiers {
		if art.Path != "" {
			continue
		}
		// MavenPath soporta coordenada de 4 partes (group:artifact:version:classifier),
		// que produce <artifact>-<version>-<classifier>.jar en la misma ruta.
		if p := globalutils.MavenPath(lib.Name + ":" + classifier); p != "" {
			art.Path = p
			lib.Downloads.Classifiers[classifier] = art
		}
	}
}

// NativeClassifierKey devuelve la key del classifier nativo de la librería
// para el SO/arquitectura actuales (resolviendo ${arch}), o "" si la librería
// no declara natives para ese SO.
func NativeClassifierKey(lib Library, os, arch string) string {
	if lib.Natives == nil {
		return ""
	}
	raw, ok := lib.Natives[os]
	if !ok || raw == "" {
		return ""
	}
	archNum := "64"
	if arch == "x86" {
		archNum = "32"
	}
	return strings.ReplaceAll(raw, "${arch}", archNum)
}
