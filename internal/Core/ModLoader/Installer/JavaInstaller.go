package installer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// RunInstallerJar ejecuta el instalador oficial de un modloader (Forge/NeoForge)
// en modo headless contra el directorio objetivo usando --installClient. Es la
// única vía que genera los jars parcheados (binarypatcher/PROCESS_MINECRAFT_JAR)
// y deja las librerías del perfil en su lugar; el propio instalador descarga
// todo lo que necesita. El stdout/stderr se reenvía línea a línea a logFn para
// poder mostrarlo en el flujo de instalación del launcher. Al terminar, los
// archivos de log que el instalador deja en el directorio de trabajo se mueven
// a logsCacheDir (carpeta cache del launcher) en lugar de borrarlos.
func RunInstallerJar(javaPath, installerJar, installDir, logsCacheDir string, logFn func(string)) error {
	if err := ensureLauncherProfiles(installDir); err != nil {
		return fmt.Errorf("ensure launcher profiles: %w", err)
	}

	cmd := exec.Command(javaPath, "-jar", installerJar, "--installClient", installDir)
	cmd.Dir = installDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start installer: %w", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		scan := func(r *bufio.Scanner) {
			for r.Scan() {
				if line := strings.TrimRight(r.Text(), "\r"); line != "" && logFn != nil {
					logFn(line)
				}
			}
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); scan(bufio.NewScanner(stdout)) }()
		go func() { defer wg.Done(); scan(bufio.NewScanner(stderr)) }()
		wg.Wait()
	}()

	err = cmd.Wait()
	<-done
	moveInstallerLogs(installDir, logsCacheDir, installerJar)
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("el instalador terminó con código de salida %d (consulta el log de instalación)", ee.ExitCode())
		}
		return fmt.Errorf("ejecutar instalador: %w", err)
	}
	return nil
}

// moveInstallerLogs mueve los archivos de log que los instaladores oficiales
// de Forge/NeoForge dejan en el directorio de trabajo al terminar
// (SimpleInstaller.getLog: f.getName() + ".log" con el nombre del jar, o
// "installer.log" si no se pudo determinar) a la carpeta cache del launcher
// (logsCacheDir) para conservarlos sin ensuciar la instancia. La ejecución usa
// cmd.Dir = installDir, así que los restos caen en la raíz de la instancia (o
// en el WorkDir del flujo sin instancia). Si logsCacheDir está vacío, los
// elimina como antes. Best-effort: nunca condiciona el resultado de la
// instalación.
func moveInstallerLogs(installDir, logsCacheDir, installerJar string) {
	if installDir == "" {
		return
	}
	jarName := filepath.Base(installerJar)
	names := []string{
		"installer.log",
		jarName + ".log",
		strings.TrimSuffix(jarName, ".jar") + ".jarlog",
	}
	for _, n := range names {
		src := filepath.Join(installDir, n)
		if logsCacheDir == "" {
			removeWithRetry(src)
			continue
		}
		if err := os.MkdirAll(logsCacheDir, 0755); err != nil {
			removeWithRetry(src)
			continue
		}
		moveWithRetry(src, filepath.Join(logsCacheDir, n))
	}
}

// moveWithRetry mueve un archivo tolerando el retardo de Windows en liberar el
// handle del proceso recién salido (best-effort). Si rename falla (p. ej. por
// estar en otro volumen), copia y borra el origen.
func moveWithRetry(src, dst string) {
	for attempt := 0; attempt < 5; attempt++ {
		if _, err := os.Stat(src); err != nil {
			return
		}
		if err := os.Rename(src, dst); err == nil {
			return
		}
		if attempt == 4 {
			if err := copyFile(src, dst); err == nil {
				os.Remove(src)
			}
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// copyFile copia el contenido de src a dst (fallback de moveWithRetry).
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// removeWithRetry intenta borrar un archivo tolerando el retardo de Windows
// en liberar el handle del proceso recién salido (best-effort).
func removeWithRetry(path string) {
	for attempt := 0; attempt < 5; attempt++ {
		err := os.Remove(path)
		if err == nil || os.IsNotExist(err) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// ensureLauncherProfiles garantiza que exista launcher_profiles.json en el
// directorio objetivo: ClientInstall de Forge/NeoForge valida su presencia
// antes de arrancar los procesadores.
func ensureLauncherProfiles(installDir string) error {
	path := filepath.Join(installDir, "launcher_profiles.json")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte("{}"), 0644)
}

// LookupInstalledJsonID busca en versions/ el id real escrito por el instalador
// (el instalador usa su propio id, p. ej. "26.2-forge-65.1.0" en lugar del
// "forge-26.2-65.1.0" derivado por el provider) y devuelve el primero cuyo "id"
// contenga loaderName y loaderVersion. Vacío si no encuentra ninguno.
func LookupInstalledJsonID(instancePath, loaderName, loaderVersion string) string {
	versionsDir := filepath.Join(instancePath, "versions")
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		verDir := filepath.Join(versionsDir, e.Name())
		data, err := os.ReadFile(filepath.Join(verDir, e.Name()+".json"))
		if err != nil {
			continue
		}
		var m struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(data, &m) != nil || m.ID == "" {
			continue
		}
		if strings.Contains(m.ID, loaderName) && strings.Contains(m.ID, loaderVersion) {
			return m.ID
		}
	}
	return ""
}
