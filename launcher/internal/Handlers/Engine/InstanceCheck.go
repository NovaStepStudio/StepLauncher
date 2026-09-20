package engine

import (
	"path/filepath"

	"StepLauncher/internal/Core/Downloader"
)

// checkInstanceExistence ejecuta en segundo plano la verificacion silenciosa de
// existencia de la version recien descargada en una instancia: solo comprueba que
// existan TODOS los archivos que declaran los JSON de la version (nada de SHA1 ni
// tamano) y descarga en background los que falten (el gestor marcara como
// "existing" los presentes y solo bajara los ausentes). Los archivos que no se
// pueden descargar quedan registrados en el log del motor (backend) y en el log
// de archivos (LogFn del gestor compartido).
func (e *Engine) checkInstanceExistence(name, version string) {
	if e.instances == nil || e.sharedDl == nil || e.log == nil {
		return
	}

	cfg := e.config.Get()
	instVerDir := filepath.Join(cfg.WorkDir, cfg.InstancesDir, name, "versions", version)

	filter := downloader.DownloadFilter{
		Version:            version,
		Client:             true,
		Libraries:          true,
		Natives:            true,
		Assets:             true,
		Java:               true,
		InstanceVersionDir: instVerDir,
	}

	dl := e.sharedDl.Start(version, filter, 0, 0, true, 0, 0)
	dl.Wait()

	info := e.sharedDl.GetInfo(dl.ID)
	if info == nil {
		e.log.Warn("[InstanceCheck] Instancia %s, version %s: no se pudo consultar el estado de la verificacion", name, version)
		return
	}

	status, _ := e.sharedDl.Status(dl.ID)
	total, existing, downloaded := 0, 0, 0
	if status != nil {
		total = status.FilesTotal
		existing = status.FilesExisting
		downloaded = status.FilesDownloaded - status.FilesExisting
	}

	if info.State == downloader.StateCompleted {
		e.log.Info("[InstanceCheck] Instancia %s, version %s: verificacion OK - %d archivos esperados, %d existentes, %d descargados en segundo plano",
			name, version, total, existing, downloaded)
		return
	}

	failed := 0
	if status != nil && status.FilesTotal > 0 {
		failed = status.FilesTotal - status.FilesDownloaded
		if failed < 0 {
			failed = 0
		}
	}
	errMsg := info.Error
	if errMsg == "" {
		errMsg = "descarga incompleta"
	}
	e.log.Warn("[InstanceCheck] Instancia %s, version %s: %d archivo(s) no se pudieron descargar (%s) - revisar log de descargas",
		name, version, failed, errMsg)
}