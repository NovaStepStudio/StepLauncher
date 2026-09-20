package instance

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// backupSkip indica qué rutas (relativas a la carpeta de la instancia) NO
// entran en el backup: logs (ruido), el archivo de lock y cualquier .lock.
func backupSkip(rel string) bool {
	if rel == "logs" || strings.HasPrefix(rel, "logs/") {
		return true
	}
	if rel == lockFile || strings.HasSuffix(rel, ".lock") {
		return true
	}
	return false
}

// backupDir es <instances>/backups, hermana de las instancias.
func (m *InstanceManager) backupDir() string {
	return filepath.Join(m.instancesDir, "backups")
}

// CreateBackup comprime la carpeta completa de la instancia (sin logs ni
// locks) en <instances>/backups/<name>.zip. Escritura atómica: se genera en
// un .tmp y se renombra al final, así el zip nunca queda a medias.
func (m *InstanceManager) CreateBackup(name string) (string, error) {
	if err := sanitizeInstanceName(name); err != nil {
		return "", err
	}
	if err := m.assertUsable(name); err != nil {
		return "", err
	}
	instPath, err := m.instancePath(name)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(instPath); err != nil {
		return "", fmt.Errorf("instancia no encontrada: %s", name)
	}

	m.mu.Lock()
	if m.backingUp[name] {
		m.mu.Unlock()
		return "", fmt.Errorf("ya hay un backup en curso para %s", name)
	}
	m.backingUp[name] = true
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		delete(m.backingUp, name)
		m.mu.Unlock()
	}()

	dir := m.backupDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	dest := filepath.Join(dir, name+".zip")
	tmp := dest + ".tmp"

	out, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	zw := zip.NewWriter(out)

	walkErr := filepath.Walk(instPath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(instPath, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if backupSkip(rel) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if fi.IsDir() {
			if rel == "." {
				return nil
			}
			_, hErr := zw.Create(rel + "/")
			return hErr
		}
		header := &zip.FileHeader{Name: rel, Method: zip.Deflate}
		header.SetMode(fi.Mode())
		w, hErr := zw.CreateHeader(header)
		if hErr != nil {
			return hErr
		}
		f, oErr := os.Open(path)
		if oErr != nil {
			return oErr
		}
		_, cErr := io.Copy(w, f)
		f.Close()
		return cErr
	})
	if walkErr != nil {
		zw.Close()
		out.Close()
		os.Remove(tmp)
		return "", walkErr
	}
	if err := zw.Close(); err != nil {
		out.Close()
		os.Remove(tmp)
		return "", err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return "", err
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return "", err
	}

	m.log("Backup de %s creado: %s", name, dest)
	return dest, nil
}

// maybeAutoBackup lanza un backup automático si la instancia lo tiene
// configurado y le toca según su planificación:
//
//   - "session": siempre al cerrarse una sesión de juego de la instancia.
//   - "hours" / "days": solo si pasaron BackupInterval horas/días desde la
//     última copia (LastBackupAt).
//
// Solo debe llamarse con la instancia cerrada (el disparador es el fin de la
// sesión de juego); el zip se genera en goroutine para no bloquear al llamador
// y el LastBackupAt se persiste releyendo la config actual (nunca mutando una
// copia compartida).
func (m *InstanceManager) maybeAutoBackup(name string, cfg *InstanceLaunchConfig) {
	if cfg == nil {
		return
	}
	schedule := strings.TrimSpace(cfg.BackupSchedule)
	if schedule == "" || schedule == "off" {
		return
	}

	due := false
	switch schedule {
	case "session":
		due = true
	case "hours", "days":
		if cfg.BackupInterval > 0 {
			if cfg.LastBackupAt == "" {
				due = true
			} else if t, err := time.Parse(time.RFC3339, cfg.LastBackupAt); err == nil {
				unit := time.Hour
				if schedule == "days" {
					unit = 24 * time.Hour
				}
				due = time.Since(t) >= time.Duration(cfg.BackupInterval)*unit
			}
		}
	}

	if !due {
		return
	}

	go func() {
		dest, err := m.CreateBackup(name)
		if err != nil {
			m.log("WARN: backup automático de %s falló: %v", name, err)
			return
		}
		now := time.Now().Format(time.RFC3339)
		lock := m.persistenceLock(name)
		lock.Lock()
		defer lock.Unlock()
		if cur, rErr := m.readConfig(name); rErr == nil {
			cur.LastBackupAt = now
			if wErr := m.writeConfig(name, cur); wErr != nil {
				m.log("WARN: no se pudo persistir lastBackupAt de %s: %v", name, wErr)
			}
		}
		m.log("Backup automático de %s: %s", name, dest)
	}()
}
