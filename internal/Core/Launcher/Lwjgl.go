package launcher

import (
	"archive/zip"
	"bytes"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	downloader "StepLauncher/internal/Core/Downloader"
)

// lwjglCompatOverrideVersion es la versión mínima de LWJGL 3.x que incluye
// los callbacks *CallbackI (GLFWFramebufferSizeCallbackI, GLFWWindowSizeCallbackI,
// etc.) que los builds tardíos de Forge para 1.16.5 (36.2.34+) usan en su
// client jar parcheado. Vanilla 1.16.5 declara LWJGL 3.2.2, que carece de esas
// interfaces: el juego muere al arrancar con NoClassDefFoundError. Es la misma
// versión de override que usan otros launchers (Prism, ATLauncher).
const lwjglCompatOverrideVersion = "3.3.1"

// forgeClientNeedsNewLWJGL detecta si algún client jar de Forge de las
// librerías de la versión parchea net/minecraft/client/MainWindow.class con
// los callbacks *CallbackI de LWJGL 3.3+.
func forgeClientNeedsNewLWJGL(libraries []downloader.Library, librariesDir string) bool {
	for _, lib := range libraries {
		if downloader.IsNativeLibrary(lib) || lib.Downloads == nil || lib.Downloads.Artifact == nil {
			continue
		}
		p := lib.Downloads.Artifact.Path
		low := strings.ToLower(p)
		if !strings.Contains(low, "/net/minecraftforge/forge/") || !strings.HasSuffix(low, "-client.jar") {
			continue
		}
		if mainWindowUsesCallbackI(filepath.Join(librariesDir, p)) {
			return true
		}
	}
	return false
}

// mainWindowUsesCallbackI abre el jar y busca la referencia al callback en el
// bytecode de MainWindow.class (el nombre aparece como string UTF-8 del pool
// de constantes, así que basta con buscarlo en los bytes).
func mainWindowUsesCallbackI(jarPath string) bool {
	rc, err := zip.OpenReader(jarPath)
	if err != nil {
		return false
	}
	defer rc.Close()
	for _, f := range rc.File {
		if f.Name != "net/minecraft/client/MainWindow.class" {
			continue
		}
		cr, err := f.Open()
		if err != nil {
			return false
		}
		data, err := io.ReadAll(io.LimitReader(cr, 1<<20))
		cr.Close()
		if err != nil {
			return false
		}
		return bytes.Contains(data, []byte("GLFWFramebufferSizeCallbackI"))
	}
	return false
}

// overrideLWJGLVersion reescribe las librerías org.lwjgl con versión 3.2.x a
// la versión indicada: nombre maven, ruta/URL del artefacto y de los
// clasificadores (natives), y limpia sha1/tamaño porque la versión nueva no
// comparte hashes con la antigua. Devuelve cuántas librerías se reescribieron.
func overrideLWJGLVersion(libraries []downloader.Library, to string) int {
	changed := 0
	for i := range libraries {
		lib := &libraries[i]
		if !strings.HasPrefix(lib.Name, "org.lwjgl:") {
			continue
		}
		old := lwjglVersionOf(lib.Name)
		if old == "" || !lwjglTooOld(old) {
			continue
		}
		lib.Name = strings.ReplaceAll(lib.Name, old, to)
		if lib.Downloads != nil {
			if a := lib.Downloads.Artifact; a != nil {
				a.Path = strings.ReplaceAll(a.Path, old, to)
				a.URL = strings.ReplaceAll(a.URL, old, to)
				a.SHA1 = ""
				a.Size = 0
			}
			for k, a := range lib.Downloads.Classifiers {
				a.Path = strings.ReplaceAll(a.Path, old, to)
				a.URL = strings.ReplaceAll(a.URL, old, to)
				a.SHA1 = ""
				a.Size = 0
				lib.Downloads.Classifiers[k] = a
			}
		}
		changed++
	}
	return changed
}

// lwjglVersionOf extrae la versión de un nombre maven (group:artifact:version
// o group:artifact:version:classifier).
func lwjglVersionOf(name string) string {
	parts := strings.Split(name, ":")
	if len(parts) < 3 || len(parts) > 4 {
		return ""
	}
	return parts[2]
}

// lwjglTooOld: versión 3.x con minor < 3 (3.2.x y anteriores).
func lwjglTooOld(ver string) bool {
	parts := strings.Split(ver, ".")
	if len(parts) < 2 {
		return false
	}
	maj, err1 := strconv.Atoi(parts[0])
	min, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return false
	}
	return maj == 3 && min < 3
}