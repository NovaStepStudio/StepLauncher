package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"StepLauncher/internal/Core/Downloader"
	coremods "StepLauncher/internal/Core/Mods"
	linstance "StepLauncher/internal/Core/Launcher/Instance"
)

// ModContentRequest pide descargar un archivo de Modrinth (mod, shader,
// resourcepack o .mrpack suelto) al juego global o a una instancia.
// Destination es "global" o "instance"; Instance lleva el nombre cuando el
// destino es una instancia. HashSHA1/HashSHA512 verifican la descarga.
type ModContentRequest struct {
	ProjectType string `json:"projectType"`
	Title       string `json:"title"`
	FileURL     string `json:"fileUrl"`
	FileName    string `json:"fileName"`
	FileSize    int64  `json:"fileSize"`
	HashSHA1    string `json:"hashSha1,omitempty"`
	HashSHA512  string `json:"hashSha512,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	Destination string `json:"destination"`
	Instance    string `json:"instance,omitempty"`
}

// ModpackInstallRequest pide instalar un modpack .mrpack. Destination elige
// dónde: "global" (juego global), "instance" (una instancia existente,
// Instance) o "new" (crear una instancia nueva, InstanceName + IconURL).
type ModpackInstallRequest struct {
	Title        string `json:"title"`
	FileURL      string `json:"fileUrl"`
	FileName     string `json:"fileName"`
	FileSize     int64  `json:"fileSize"`
	HashSHA1     string `json:"hashSha1,omitempty"`
	HashSHA512   string `json:"hashSha512,omitempty"`
	Destination  string `json:"destination"`
	Instance     string `json:"instance,omitempty"`
	InstanceName string `json:"instanceName,omitempty"`
	IconURL      string `json:"iconUrl,omitempty"`
}

// ModContentResult devuelve el sessionId de la operación asíncrona. El
// progreso llega con eventos modcontent_* o modpack_* según el caso.
type ModContentResult struct {
	SessionID string `json:"sessionId"`
	Status    string `json:"status"`
}

var (
	modSessionsMu sync.Mutex
	modSessions   = map[string]context.CancelFunc{}
)

func registerModSession(id string, cancel context.CancelFunc) {
	modSessionsMu.Lock()
	modSessions[id] = cancel
	modSessionsMu.Unlock()
}

func cancelModSession(id string) bool {
	modSessionsMu.Lock()
	cancel, ok := modSessions[id]
	if ok {
		delete(modSessions, id)
	}
	modSessionsMu.Unlock()
	if ok {
		cancel()
	}
	return ok
}

func forgetModSession(id string) {
	modSessionsMu.Lock()
	delete(modSessions, id)
	modSessionsMu.Unlock()
}

func newModSession(prefix string) (string, context.Context, context.CancelFunc) {
	id := fmt.Sprintf("%s-%d", prefix, time.Now().UnixMilli())
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	registerModSession(id, cancel)
	return id, ctx, cancel
}

// emitModEvent envía un evento al frontend sin bloquear nunca el llamador.
// eventCb puede ser nil (motor sin UI): en ese caso se ignora.
func (e *Engine) emitModEvent(eventType string, data []byte) {
	if e.eventCb == nil {
		return
	}
	e.eventCb(eventType, data)
}

// GlobalGameDir devuelve el gameDir global efectivo: <workDir>/game si la
// opción de carpeta separada está activa, o el propio workDir si está
// deshabilitada (los mods/shaders/texturas viven entonces en el base). Se lee
// el valor efectivo del LaunchManager (igual que al lanzar) y no la config en
// crudo, porque en modo Minecraft el manager fuerza separado=false.
func (e *Engine) GlobalGameDir() string {
	workDir := e.config.Get().WorkDir
	separate := e.config.Get().SeparateGameDirValue()
	if e.launcher != nil {
		separate = e.launcher.GetSeparateGameDir()
	}
	return coremods.ResolveGameDir(workDir, separate)
}

// InstanceGameDir devuelve el gameDir efectivo de una instancia:
// <instancia>/game con la opción separada activa, o la propia carpeta de la
// instancia si está deshabilitada.
func (e *Engine) InstanceGameDir(name string) (string, error) {
	return e.instances.GameDirFor(name)
}

// ContentSubdirFor expone el mapeo tipo de proyecto -> subcarpeta del gameDir.
func (e *Engine) ContentSubdirFor(projectType string) (string, error) {
	return coremods.ContentSubdir(projectType)
}

// resolveContentGameDir traduce el destino del frontend ("global" o
// "instance" + nombre) al gameDir absoluto y su etiqueta legible.
func (e *Engine) resolveContentGameDir(destination, instance string) (gameDir, label string, err error) {
	switch strings.ToLower(strings.TrimSpace(destination)) {
	case "", "global":
		return e.GlobalGameDir(), "juego global", nil
	case "instance":
		if strings.TrimSpace(instance) == "" {
			return "", "", fmt.Errorf("falta la instancia de destino")
		}
		gameDir, err := e.instances.GameDirFor(instance)
		if err != nil {
			return "", "", err
		}
		return gameDir, "instancia " + instance, nil
	default:
		return "", "", fmt.Errorf("destino inválido: %s", destination)
	}
}

// InstallModContent descarga un archivo de Modrinth al destino elegido.
// No bloquea el binding: devuelve el sessionId y trabaja en goroutine con
// eventos modcontent_* (patrón de InstallModLoader).
func (e *Engine) InstallModContent(req ModContentRequest) (*ModContentResult, error) {
	subdir, err := coremods.ContentSubdir(req.ProjectType)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(req.FileURL, "https://") && !strings.HasPrefix(req.FileURL, "http://") {
		return nil, fmt.Errorf("URL de descarga inválida")
	}
	fileName, err := coremods.SanitizeFileName(req.FileName)
	if err != nil {
		return nil, err
	}
	gameDir, destLabel, err := e.resolveContentGameDir(req.Destination, req.Instance)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = fileName
	}

	sessionID, ctx, cancel := newModSession("mod")
	go func() {
		defer cancel()
		defer forgetModSession(sessionID)
		e.emitModEvent("modcontent_resolving", coremods.ResolvingContentEvent(sessionID, title, destLabel))
		if err := coremods.EnsureContentDirs(gameDir); err != nil {
			e.log.Error("[Mods] Contenido %s: %v", title, err)
			e.emitModEvent("modcontent_error", coremods.ErrorContentEvent(sessionID, title, err.Error()))
			return
		}
		dest := filepath.Join(gameDir, subdir, fileName)
		e.emitModEvent("modcontent_downloading", coremods.DownloadingContentEvent(sessionID, title, fileName, 1, 1))
		client := e.downloader.HTTPClient()
		if err := coremods.DownloadToFile(ctx, client, req.FileURL, dest, req.HashSHA1, req.HashSHA512, req.FileSize); err != nil {
			e.log.Error("[Mods] Contenido %s: %v", title, err)
			e.emitModEvent("modcontent_error", coremods.ErrorContentEvent(sessionID, title, err.Error()))
			return
		}
		// Icono del proyecto para los mods sin icono embebido (best-effort).
		coremods.CacheProjectIcon(e.config.Get().CacheDir, fileName, strings.TrimSpace(req.IconURL))
		e.log.Info("[Mods] Contenido instalado: %s -> %s", title, dest)
		e.emitModEvent("modcontent_installed", coremods.InstalledContentEvent(sessionID, title, filepath.Join(subdir, fileName)))
	}()

	return &ModContentResult{SessionID: sessionID, Status: "installing"}, nil
}

// CancelModContent cancela una descarga de contenido o modpack en curso.
func (e *Engine) CancelModContent(sessionID string) bool {
	return cancelModSession(sessionID)
}

// ListInstalledContent enumera el contenido instalado (mods, shaders y
// texturas, activos o deshabilitados) del juego global o de una instancia.
func (e *Engine) ListInstalledContent(destination, instance string) ([]coremods.InstalledFile, error) {
	gameDir, _, err := e.resolveContentGameDir(destination, instance)
	if err != nil {
		return nil, err
	}
	files, err := coremods.ListInstalled(gameDir)
	if err != nil {
		return nil, err
	}
	if files == nil {
		files = []coremods.InstalledFile{}
	}
	return files, nil
}

// SetInstalledContentEnabled activa o desactiva un archivo instalado.
func (e *Engine) SetInstalledContentEnabled(destination, instance, kind, name string, enabled bool) error {
	gameDir, _, err := e.resolveContentGameDir(destination, instance)
	if err != nil {
		return err
	}
	return coremods.SetInstalledEnabled(gameDir, kind, name, enabled)
}

// DeleteInstalledContent borra un archivo instalado del destino.
func (e *Engine) DeleteInstalledContent(destination, instance, kind, name string) error {
	gameDir, _, err := e.resolveContentGameDir(destination, instance)
	if err != nil {
		return err
	}
	return coremods.DeleteInstalledContent(gameDir, kind, name)
}

// MoveInstalledContent mueve un archivo instalado de un destino a otro
// (global <-> instancias), conservando si estaba habilitado.
func (e *Engine) MoveInstalledContent(srcDestination, srcInstance, kind, name, dstDestination, dstInstance string) error {
	srcDir, _, err := e.resolveContentGameDir(srcDestination, srcInstance)
	if err != nil {
		return err
	}
	dstDir, _, err := e.resolveContentGameDir(dstDestination, dstInstance)
	if err != nil {
		return err
	}
	return coremods.MoveInstalledContent(srcDir, dstDir, kind, name)
}

// RevealInstalledContent abre el explorador con el archivo instalado.
func (e *Engine) RevealInstalledContent(destination, instance, kind, name string) error {
	gameDir, _, err := e.resolveContentGameDir(destination, instance)
	if err != nil {
		return err
	}
	path, err := coremods.InstalledPath(gameDir, kind, name)
	if err != nil {
		return err
	}
	return coremods.RevealInExplorer(path)
}

// GetContentIcon devuelve la ruta relativa del icono de un contenido
// instalado (pack.png de texturas/shaders o icon del fabric.mod.json /
// quilt.mod.json / mods.toml de un mod). Vacío cuando no tiene: la UI usa su
// icono genérico. Nunca falla por falta de icono para no ensuciar el log.
func (e *Engine) GetContentIcon(destination, instance, kind, name string) string {
	gameDir, _, err := e.resolveContentGameDir(destination, instance)
	if err != nil {
		return ""
	}
	rel, err := coremods.ExtractContentIcon(gameDir, kind, name, e.config.Get().CacheDir)
	if err != nil {
		return ""
	}
	return rel
}
// juego global, instancia existente o instancia nueva (con el icono del pack).
// Todo ocurre en goroutine con eventos modpack_*; el binding devuelve el
// sessionId de una vez.
// InstallModpack descarga un .mrpack y lo instala en el destino elegido:
// juego global, instancia existente o instancia nueva (con el icono del pack).
// La instancia nueva queda oculta y bloqueada (provisioning) hasta estar al
// 100%: si algo falla o se cancela, se elimina sin dejar restos.
// Todo ocurre en goroutine con eventos modpack_*; el binding devuelve el
// sessionId de una vez.
func (e *Engine) InstallModpack(req ModpackInstallRequest) (*ModContentResult, error) {
	if !strings.HasPrefix(req.FileURL, "https://") && !strings.HasPrefix(req.FileURL, "http://") {
		return nil, fmt.Errorf("URL de descarga inválida")
	}
	fileName, err := coremods.SanitizeFileName(req.FileName)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(strings.ToLower(fileName), ".mrpack") {
		return nil, fmt.Errorf("el archivo no es un modpack (.mrpack)")
	}
	switch strings.ToLower(strings.TrimSpace(req.Destination)) {
	case "", "global", "instance", "new":
	default:
		return nil, fmt.Errorf("destino inválido: %s", req.Destination)
	}
	if strings.ToLower(strings.TrimSpace(req.Destination)) == "instance" && strings.TrimSpace(req.Instance) == "" {
		return nil, fmt.Errorf("falta la instancia de destino")
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = strings.TrimSuffix(fileName, ".mrpack")
	}

	sessionID, ctx, cancel := newModSession("mp")
	go func() {
		defer cancel()
		defer forgetModSession(sessionID)
		if err := e.runModpackInstall(ctx, sessionID, req, title, fileName); err != nil {
			e.log.Error("[Mods] Modpack %s: %v", title, err)
			e.emitModEvent("modpack_error", coremods.ModpackErrorEvent(sessionID, "", err.Error()))
		}
	}()

	return &ModContentResult{SessionID: sessionID, Status: "installing"}, nil
}

func (e *Engine) runModpackInstall(ctx context.Context, sessionID string, req ModpackInstallRequest, title, fileName string) error {
	dest := strings.ToLower(strings.TrimSpace(req.Destination))
	if dest == "" {
		dest = "new"
	}
	fail := func(instance string, err error) error {
		e.emitModEvent("modpack_error", coremods.ModpackErrorEvent(sessionID, instance, err.Error()))
		return err
	}
	emit := func(instance, phase, message string, progress, total int) {
		e.emitModEvent("modpack_"+phase, coremods.ModpackProgressEvent(sessionID, instance, phase, message, progress, total))
	}

	cacheDir := e.config.Get().CacheDir
	stageDir := filepath.Join(cacheDir, "modpacks", sessionID)
	if err := os.MkdirAll(stageDir, 0755); err != nil {
		return fail("", fmt.Errorf("preparar staging: %w", err))
	}
	defer os.RemoveAll(stageDir)

	// 1. Descargar el .mrpack.
	emit("", "resolving", "Descargando el modpack "+title+"…", 0, 1)
	mrpackPath := filepath.Join(stageDir, fileName)
	if err := coremods.DownloadToFile(ctx, e.downloader.HTTPClient(), req.FileURL, mrpackPath, req.HashSHA1, req.HashSHA512, req.FileSize); err != nil {
		return fail("", fmt.Errorf("descargar el modpack: %w", err))
	}

	// 2. Extraer y leer el manifiesto (origen de la verdad: qué MC, qué
	// loader, qué archivos y qué overrides trae el pack).
	emit("", "resolving", "Leyendo el manifiesto del modpack…", 0, 1)
	extractDir := filepath.Join(stageDir, "extracted")
	if err := coremods.ExtractMrpack(mrpackPath, extractDir); err != nil {
		return fail("", fmt.Errorf("extraer el modpack: %w", err))
	}
	index, err := coremods.ParseMrpackIndex(extractDir)
	if err != nil {
		return fail("", err)
	}
	mcVersion := index.MinecraftVersion()
	if mcVersion == "" {
		return fail("", fmt.Errorf("el manifiesto no indica la versión de Minecraft"))
	}
	loader, loaderVersion := index.RequiredLoader()

	// 3. Preparar el destino: instancia nueva, existente o juego global. En
	// todos los casos se garantiza la versión de Minecraft y se intenta el
	// modloader que exige el manifiesto. Si el loader falla (p. ej. versión
	// inexistente en Maven), NO es fatal: se crea igual con mods y overrides
	// y se avisa para instalarlo a mano desde la instancia.
	var instName, gameDir, loaderWarning string
	warnLoader := func(err error) {
		loaderWarning = fmt.Sprintf("Sin modloader %s %s (%v). El contenido está instalado; pon el loader desde Instancias > Descargar.", loader, loaderVersion, err)
		e.log.Error("[Mods] Modpack %s: %s", title, loaderWarning)
	}
	switch dest {
	case "new":
		instName = e.uniqueModpackInstanceName(strings.TrimSpace(req.InstanceName), title)
		iconURL := strings.TrimSpace(req.IconURL)
		if err := e.instances.BeginProvisioning(instName, title, iconURL, sessionID); err != nil {
			return fail("", err)
		}
		// A partir de aquí la instancia está reservada pero oculta: cualquier
		// fallo o cancelación la elimina para no dejar restos a medias.
		failNew := func(err error) error {
			e.instances.DeleteProvisioned(instName)
			return fail(instName, err)
		}
		emit(instName, "resolving", "Creando la instancia "+instName+"…", 0, 1)
		if _, _, err := e.instances.Create(linstance.CreateInstanceReq{Name: instName, Title: title}); err != nil {
			return failNew(fmt.Errorf("crear la instancia %s: %w", instName, err))
		}
		emit(instName, "downloading", "Descargando Minecraft "+mcVersion+"…", 0, 1)
		mcProg := func(pct int, downMB, totalMB float64) {
			emit(instName, "downloading", mcProgMsg(mcVersion, pct, downMB, totalMB), pct, 100)
		}
		if err := e.ensureInstanceVersion(ctx, instName, mcVersion, true, mcProg); err != nil {
			return failNew(err)
		}
		if loader != "" && loaderVersion != "" {
			emit(instName, "installing", fmt.Sprintf("Instalando %s %s…", loader, loaderVersion), 0, 1)
			if err := e.ensureInstanceLoader(ctx, instName, loader, loaderVersion, mcVersion, true); err != nil {
				warnLoader(err)
			} else {
				e.setDefaultLoaderVersion(instName)
			}
		}
		gameDir, err = e.instances.GameDirForSystem(instName)
		if err != nil {
			return failNew(err)
		}
		e.setInstanceIconFromURL(instName, iconURL, stageDir)
	case "instance":
		instName = strings.TrimSpace(req.Instance)
		gameDir, err = e.instances.GameDirFor(instName)
		if err != nil {
			return fail(instName, err)
		}
		emit(instName, "downloading", "Comprobando Minecraft "+mcVersion+" en "+instName+"…", 0, 1)
		mcProg := func(pct int, downMB, totalMB float64) {
			emit(instName, "downloading", mcProgMsg(mcVersion, pct, downMB, totalMB), pct, 100)
		}
		if err := e.ensureInstanceVersion(ctx, instName, mcVersion, false, mcProg); err != nil {
			return fail(instName, err)
		}
		if loader != "" && loaderVersion != "" {
			if st, _ := e.instances.InstalledModLoader(instName); st == nil {
				emit(instName, "installing", fmt.Sprintf("Instalando %s %s…", loader, loaderVersion), 0, 1)
				if err := e.ensureInstanceLoader(ctx, instName, loader, loaderVersion, mcVersion, false); err != nil {
					warnLoader(err)
				} else {
					e.setDefaultLoaderVersion(instName)
				}
			} else {
				e.log.Info("[Mods] Modpack %s: la instancia %s ya tiene %s, se reutiliza", title, instName, st.LoaderType)
			}
		}
	case "global":
		gameDir = e.GlobalGameDir()
		emit("", "downloading", "Comprobando Minecraft "+mcVersion+"…", 0, 1)
		mcProg := func(pct int, downMB, totalMB float64) {
			emit("", "downloading", mcProgMsg(mcVersion, pct, downMB, totalMB), pct, 100)
		}
		if err := e.ensureGlobalVersion(ctx, mcVersion, mcProg); err != nil {
			return fail("", err)
		}
		if loader != "" && loaderVersion != "" {
			emit("", "installing", fmt.Sprintf("Instalando %s %s…", loader, loaderVersion), 0, 1)
			if err := e.ensureGlobalLoader(ctx, sessionID, loader, loaderVersion, mcVersion); err != nil {
				warnLoader(err)
			}
		}
	}

	// 4. Crear los directorios de contenido y copiar overrides/ en el gameDir.
	// En instancia nueva cualquier fallo limpia lo creado (failNew).
	failHere := fail
	if dest == "new" {
		failHere = func(instance string, err error) error {
			e.instances.DeleteProvisioned(instName)
			return fail(instance, err)
		}
	}
	emit(instName, "installing", "Preparando las carpetas del destino…", 0, 1)
	if err := coremods.EnsureContentDirs(gameDir); err != nil {
		return failHere(instName, fmt.Errorf("crear carpetas: %w", err))
	}
	emit(instName, "installing", "Copiando la configuración del modpack…", 0, 1)
	if _, err := coremods.CopyOverrides(extractDir, gameDir); err != nil {
		return failHere(instName, fmt.Errorf("copiar overrides: %w", err))
	}

	// 5. Descargar los archivos del manifiesto (mods, resourcepacks...) a su
	// path relativo dentro del gameDir, verificando hashes.
	total := len(index.Files)
	done := 0
	for _, f := range index.Files {
		if ctx.Err() != nil {
			return failHere(instName, fmt.Errorf("instalación cancelada"))
		}
		if len(f.Downloads) == 0 {
			continue
		}
		rel, err := coremods.SanitizeContentPath(f.Path)
		if err != nil {
			e.log.Error("[Mods] Modpack %s: ruta ignorada %q: %v", title, f.Path, err)
			continue
		}
		done++
		emit(instName, "downloading", fmt.Sprintf("Descargando %s (%d/%d)…", filepath.Base(rel), done, total), done, total)
		dest := filepath.Join(gameDir, rel)
		if err := coremods.DownloadToFile(ctx, e.downloader.HTTPClient(), f.Downloads[0], dest, f.SHA1(), f.SHA512(), f.FileSize); err != nil {
			return failHere(instName, fmt.Errorf("descargar %s: %w", rel, err))
		}
	}

	os.Remove(mrpackPath)
	if dest == "new" {
		// Lista para jugar: recién ahora se hace visible.
		e.instances.EndProvisioning(instName)
	}
	e.log.Info("[Mods] Modpack instalado: %s -> %s (MC %s)%s", title, destLabel(instName), mcVersion, loaderWarningSuffix(loaderWarning))
	e.emitModEvent("modpack_installed", coremods.ModpackDoneEvent(sessionID, instName, loaderWarning))
	return nil
}

func loaderWarningSuffix(warning string) string {
	if warning == "" {
		return ""
	}
	return " | " + warning
}

func destLabel(instName string) string {
	if instName == "" {
		return "juego global"
	}
	return "instancia " + instName
}

// mcProgMsg resume el progreso 0-100 de la descarga de Minecraft para la UI.
func mcProgMsg(mcVersion string, pct int, downMB, totalMB float64) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	if totalMB > 0 {
		return fmt.Sprintf("Descargando Minecraft %s… %d%% (%.0f/%.0f MB)", mcVersion, pct, downMB, totalMB)
	}
	return fmt.Sprintf("Descargando Minecraft %s… %d%%", mcVersion, pct)
}

// ensureInstanceVersion descarga la versión base en la instancia si aún no
// está, y espera a que termine (la goroutine del manager ya actualiza el
// metadata al completarse). onProg informa el porcentaje 0-100 para la UI.
func (e *Engine) ensureInstanceVersion(ctx context.Context, instName, mcVersion string, system bool, onProg func(pct int, downMB, totalMB float64)) error {
	instPath := filepath.Join(e.config.Get().WorkDir, e.config.Get().InstancesDir, instName)
	if _, err := os.Stat(filepath.Join(instPath, "versions", mcVersion, mcVersion+".json")); err == nil {
		return nil
	}
	var dl *downloader.Download
	var err error
	if system {
		dl, err = e.instances.AddVersionSystem(instName, linstance.AddVersionReq{Version: mcVersion})
	} else {
		dl, err = e.instances.AddVersion(instName, linstance.AddVersionReq{Version: mcVersion})
	}
	if err != nil {
		return fmt.Errorf("descargar Minecraft %s: %w", mcVersion, err)
	}
	waitErr := make(chan error, 1)
	go func() {
		dl.Wait()
		info := e.sharedDl.GetInfo(dl.ID)
		if info == nil {
			waitErr <- fmt.Errorf("descarga de Minecraft %s sin estado", mcVersion)
			return
		}
		if info.State != downloader.StateCompleted {
			msg := info.Error
			if msg == "" {
				msg = string(info.State)
			}
			waitErr <- fmt.Errorf("descarga de Minecraft %s: %s", mcVersion, msg)
			return
		}
		waitErr <- nil
	}()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("instalación cancelada")
		case err := <-waitErr:
			return err
		case <-ticker.C:
			if onProg == nil {
				continue
			}
			if st, serr := e.sharedDl.Status(dl.ID); serr == nil && st != nil {
				onProg(int(st.Percent), st.MBDownloaded, st.MBTotal)
			}
		}
	}
}

// ensureInstanceLoader instala el modloader en la instancia y espera a que su
// estado refleje la versión instalada. Si la versión base no está descargada,
// el instalador lo indica y se traduce a un mensaje accionable.
func (e *Engine) ensureInstanceLoader(ctx context.Context, instName, loader, loaderVersion, mcVersion string, system bool) error {
	var sessionID string
	var err error
	if system {
		sessionID, err = e.instances.InstallModLoaderSystem(instName, loader, loaderVersion, mcVersion)
	} else {
		sessionID, err = e.instances.InstallModLoader(instName, loader, loaderVersion, mcVersion)
	}
	_ = sessionID
	if err != nil {
		if strings.Contains(err.Error(), "debe descargarse antes") {
			return fmt.Errorf("descarga primero la versión %s en la instancia %s y reintenta", mcVersion, instName)
		}
		return fmt.Errorf("instalar %s %s: %w", loader, loaderVersion, err)
	}
	return e.waitModpackLoader(ctx, instName, loaderVersion)
}

// setDefaultLoaderVersion deja la versión del loader recién instalado como
// versión activa de la instancia: así se puede jugar directamente sin pasar
// por Versiones a seleccionarla. Si no hay loader, no toca nada (la versión
// base ya quedó activa al descargarse).
func (e *Engine) setDefaultLoaderVersion(instName string) {
	st, err := e.instances.InstalledModLoader(instName)
	if err != nil || st == nil || strings.TrimSpace(st.VersionJsonID) == "" {
		return
	}
	if _, err := e.instances.UpdateConfigSystem(instName, &linstance.InstanceLaunchConfig{Version: st.VersionJsonID}); err != nil {
		e.log.Error("[Mods] Versión por defecto de %s: %v", instName, err)
		return
	}
	e.log.Info("[Mods] Versión por defecto de %s: %s", instName, st.VersionJsonID)
}

// ensureGlobalVersion descarga la versión completa en el juego global si falta.
func (e *Engine) ensureGlobalVersion(ctx context.Context, mcVersion string, onProg func(pct int, downMB, totalMB float64)) error {
	workDir := e.config.Get().WorkDir
	if _, err := os.Stat(filepath.Join(workDir, "versions", mcVersion, mcVersion+".json")); err == nil {
		return nil
	}
	info := e.StartFullDownload(mcVersion)
	if info == nil {
		return fmt.Errorf("no se pudo iniciar la descarga de Minecraft %s", mcVersion)
	}
	waitErr := make(chan error, 1)
	go func() {
		dl := e.downloader.Get(info.ID)
		if dl == nil {
			waitErr <- fmt.Errorf("descarga de Minecraft %s sin estado", mcVersion)
			return
		}
		dl.Wait()
		st := e.downloader.GetInfo(info.ID)
		if st == nil {
			waitErr <- fmt.Errorf("descarga de Minecraft %s sin estado", mcVersion)
			return
		}
		if st.State != downloader.StateCompleted {
			msg := st.Error
			if msg == "" {
				msg = string(st.State)
			}
			waitErr <- fmt.Errorf("descarga de Minecraft %s: %s", mcVersion, msg)
			return
		}
		waitErr <- nil
	}()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("instalación cancelada")
		case err := <-waitErr:
			return err
		case <-ticker.C:
			if onProg == nil {
				continue
			}
			if st, serr := e.downloader.Status(info.ID); serr == nil && st != nil {
				onProg(int(st.Percent), st.MBDownloaded, st.MBTotal)
			}
		}
	}
}

// ensureGlobalLoader instala el modloader en el juego global de forma
// síncrona dentro de la goroutine del modpack (los eventos modloader_*
// informan del progreso con el sessionId de la operación).
func (e *Engine) ensureGlobalLoader(ctx context.Context, sessionID, loader, loaderVersion, mcVersion string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("instalación cancelada")
	}
	done := make(chan error, 1)
	go func() {
		workDir := e.config.Get().WorkDir
		_, err := e.modloader.Install(sessionID, loader, loaderVersion, mcVersion, workDir)
		if err != nil {
			done <- err
			return
		}
		// Flujo clásico (WorkDir): el estado no se lee en ningún lado, así que
		// se limpia igual que en InstallModLoader.
		if err := e.modloader.RemoveState(workDir); err != nil {
			e.log.Error("Modloader state cleanup failed: %v", err)
		}
		done <- nil
	}()
	select {
	case <-ctx.Done():
		return fmt.Errorf("instalación cancelada")
	case err := <-done:
		if err != nil {
			return fmt.Errorf("instalar %s %s: %w", loader, loaderVersion, err)
		}
		return nil
	}
}

// setInstanceIconFromURL deja a la instancia nueva con el icono del modpack.
// Es best-effort: si el icono no se puede descargar, la instalación sigue.
func (e *Engine) setInstanceIconFromURL(instName, iconURL, stageDir string) {
	if iconURL == "" {
		return
	}
	final, err := coremods.DownloadIcon(e.downloader.HTTPClient(), iconURL, filepath.Join(stageDir, "icon"))
	if err != nil {
		e.log.Error("[Mods] Icono de %s: %v", instName, err)
		return
	}
	cfg := e.config.Get()
	instPath := filepath.Join(cfg.WorkDir, cfg.InstancesDir, instName)
	assetsDir := filepath.Join(instPath, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		e.log.Error("[Mods] Icono de %s: %v", instName, err)
		return
	}
	dest := filepath.Join(assetsDir, "icon"+filepath.Ext(final))
	if err := copyModFile(final, dest); err != nil {
		e.log.Error("[Mods] Icono de %s: %v", instName, err)
		return
	}
	rel, err := filepath.Rel(cfg.WorkDir, dest)
	if err != nil {
		return
	}
	if _, err := e.instances.UpdateMetadataSystem(instName, linstance.UpdateMetadataReq{Icon: filepath.ToSlash(rel)}); err != nil {
		e.log.Error("[Mods] Icono de %s: %v", instName, err)
		return
	}
	e.log.Info("[Mods] Icono aplicado a la instancia %s", instName)
}

func copyModFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// waitModpackLoader espera (con polling, sin locks en I/O) a que el estado
// del modloader de la instancia refleje la versión instalada.
func (e *Engine) waitModpackLoader(ctx context.Context, instName, loaderVersion string) error {
	deadline := time.Now().Add(20 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		if st, _ := e.instances.InstalledModLoader(instName); st != nil {
			if loaderVersion == "" || st.LoaderVersion == loaderVersion {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("instalación cancelada")
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("el modloader tardó demasiado en instalarse")
			}
		}
	}
}

// uniqueModpackInstanceName genera un nombre de instancia libre a partir del
// pedido por el usuario o del título del modpack.
func (e *Engine) uniqueModpackInstanceName(wanted, title string) string {
	base := strings.TrimSpace(wanted)
	if base == "" {
		base = strings.TrimSpace(title)
	}
	base = strings.ReplaceAll(base, "/", "-")
	base = strings.ReplaceAll(base, "\\", "-")
	base = strings.TrimSpace(base)
	if base == "" {
		base = "modpack"
	}
	if utf8.RuneCountInString(base) > 48 {
		runes := []rune(base)
		base = strings.TrimSpace(string(runes[:48]))
	}
	taken := map[string]bool{}
	for _, inst := range e.instances.List() {
		if inst != nil {
			taken[inst.Name] = true
		}
	}
	if !taken[base] {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !taken[candidate] {
			return candidate
		}
	}
}
