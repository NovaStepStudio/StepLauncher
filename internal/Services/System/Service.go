package system

import (
	"context"
	"errors"
	"os"
	"os/exec"

	"StepLauncher/internal/Handlers"
	engine "StepLauncher/internal/Handlers/Engine"
	engineconfig "StepLauncher/internal/Handlers/Engine/engineconfig"
	tray "StepLauncher/internal/Tray"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// SystemService gestiona lifecycle, directorios, updater y diagnostics.
type SystemService struct {
	ctx     context.Context
	app     *application.App
	handler *Handlers.App
	engine  *engine.Engine
	tray    *tray.Tray
}

func NewSystemService(wailsApp *application.App, handler *Handlers.App, eng *engine.Engine) *SystemService {
	return &SystemService{app: wailsApp, handler: handler, engine: eng}
}

var appIcon []byte

// SetAppIcon inyecta el icono del tray (embed desde main para evitar go:embed con ..).
func SetAppIcon(b []byte) { appIcon = b }

// runtimeBridge implementa Handlers.RuntimeBridge con la API de Wails v3.
type runtimeBridge struct {
	app *application.App
}

type DirectorySettings = engine.DirectoryInfo

type LauncherVersionInfo struct {
	CurrentVersion string `json:"currentVersion"`
	AppName        string `json:"appName"`
	AppAuthor      string `json:"appAuthor"`
	UpdaterState   string `json:"updaterState"`
	UpdaterVersion string `json:"updaterVersion"`
}

func defaultConfigDir() string {
	return engineconfig.LoadBootstrap().ResolveWorkDir()
}

func (s *SystemService) ApplyUpdate() error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.ApplyUpdate()
}

func (s *SystemService) BootstrapComplete(durationMs int, failedSteps int) {
	if s.engine == nil || s.engine.Logger() == nil {
		return
	}
	if failedSteps > 0 {
		s.engine.Logger().Warn("[Bootstrap] completado en %dms con %d pasos fallidos", durationMs, failedSteps)
	} else {
		s.engine.Logger().System("[Bootstrap] completado en %dms", durationMs)
	}
}

func (s *SystemService) BootstrapLog(level, message string) {
	if s.engine == nil || s.engine.Logger() == nil {
		return
	}
	switch level {
	case "error":
		s.engine.Logger().Error("[Bootstrap] %s", message)
	case "warn":
		s.engine.Logger().Warn("[Bootstrap] %s", message)
	case "success", "info":
		s.engine.Logger().Info("[Bootstrap] %s", message)
	case "debug":
		s.engine.Logger().Debug("[Bootstrap] %s", message)
	default:
		s.engine.Logger().Info("[Bootstrap] %s", message)
	}
}

func (s *SystemService) CheckForUpdates() {
	if s.handler != nil {
		s.handler.CheckForUpdates()
	}
}

func (s *SystemService) EngineConfig() engineconfig.Config {
	if s.engine == nil {
		return engineconfig.DefaultConfig()
	}
	return s.engine.Config()
}

func (s *SystemService) EngineInfo() engine.EngineInfo {
	if s.engine == nil {
		return engine.EngineInfo{}
	}
	return s.engine.EngineInfo()
}

func (s *SystemService) GetAppVersion() string {
	return s.GetLauncherVersionInfo().CurrentVersion
}

func (s *SystemService) GetDirectorySettings() DirectorySettings {
	if s.engine == nil {
		return DirectorySettings{}
	}
	return s.engine.DirectoryInfo()
}

func (s *SystemService) GetFirstLaunch() bool {
	if s.handler == nil {
		return false
	}
	return s.handler.GetFirstLaunch()
}

func (s *SystemService) GetLauncherVersionInfo() LauncherVersionInfo {
	cur := engineconfig.AppVersion
	state := "unconfigured"
	uver := ""
	if s.app != nil && s.app.Updater != nil {
		state = string(s.app.Updater.State())
		uver = s.app.Updater.CurrentVersion()
		if uver != "" {
			cur = uver
		}
	}
	if cur == "" {
		cur = "2.5.0"
	}
	return LauncherVersionInfo{
		CurrentVersion: cur,
		AppName:        engineconfig.AppName,
		AppAuthor:      engineconfig.AppAuthor,
		UpdaterState:   state,
		UpdaterVersion: uver,
	}
}

func (s *SystemService) LocalAssetsDir() string {
	return s.handler.LocalAssetsDir()
}

func (s *SystemService) NewsLoadChangelog(version string) {
	if s.handler != nil {
		s.handler.NewsLoadChangelog(version)
	}
}

func (s *SystemService) NewsLoadMarkdown(url string) {
	if s.handler != nil {
		s.handler.NewsLoadMarkdown(url)
	}
}

func (s *SystemService) NewsLoadRelease(version string) {
	if s.handler != nil {
		s.handler.NewsLoadRelease(version)
	}
}

func (s *SystemService) NewsRefreshIndex() {
	if s.handler != nil {
		s.handler.NewsRefreshIndex()
	}
}

func (s *SystemService) OpenPath(path string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.OpenPath(path)
}

func (s *SystemService) PickDirectory() (string, error) {
	if s.app == nil {
		return "", errors.New("aplicación no disponible")
	}
	return s.app.Dialog.OpenFile().SetTitle("Seleccionar carpeta de trabajo del launcher").
		CanChooseDirectories(true).CanChooseFiles(false).PromptForSingleSelection()
}

func (s *SystemService) ReadAbsoluteFile(path string) ([]byte, error) {
	if s.handler == nil {
		return nil, errors.New("handler no disponible")
	}
	return s.handler.ReadAbsoluteFile(path)
}

func (s *SystemService) ReadLocalFile(rel string) ([]byte, error) {
	return s.handler.ReadLocalFile(rel)
}

func (s *SystemService) RestartApp() error {
	if s.app == nil {
		return errors.New("aplicación no disponible")
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe)
	if err := cmd.Start(); err != nil {
		return err
	}
	s.app.Quit()
	return nil
}

func (s *SystemService) ServiceShutdown() error {
	if s.tray != nil {
		s.tray.Shutdown()
	}
	if s.handler != nil {
		s.handler.Shutdown()
	}
	if s.engine != nil {
		s.engine.Shutdown()
	}
	return nil
}

func (s *SystemService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.ctx = ctx
	s.tray = tray.New(s.app, s.engine)
	s.handler.SetRuntimeBridge(&runtimeBridge{app: s.app})
	s.handler.SetEventCallback(func(eventType string, data []byte) {
		s.app.Event.Emit(eventType, string(data))
		s.tray.OnEngineEvent(eventType)
	})
	s.handler.Startup()
	s.tray.Setup(appIcon)
	return nil
}

func (s *SystemService) SetDirectoryMode(mode string, customPath string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.SetDirectory(engineconfig.DirMode(mode), customPath)
}

func (s *SystemService) SetFirstLaunchDone() {
	if s.handler != nil {
		s.handler.SetFirstLaunchDone()
	}
}

func (s *SystemService) UpdateEngineConfig(cfg engineconfig.Config) {
	if s.engine == nil {
		return
	}
	s.engine.UpdateConfig(cfg)
}

func (s *SystemService) WailsCheckForUpdate() error {
	if s.app == nil || s.app.Updater == nil {
		return errors.New("updater no disponible")
	}
	go func() {
		// Check emite eventos automáticamente; no hace falta emitir manualmente.
		_, _ = s.app.Updater.Check(context.Background())
	}()
	return nil
}

func (s *SystemService) WailsGetCurrentVersion() string {
	if s.app == nil || s.app.Updater == nil {
		return ""
	}
	return s.app.Updater.CurrentVersion()
}

func (s *SystemService) WailsGetState() string {
	if s.app == nil || s.app.Updater == nil {
		return "unconfigured"
	}
	return string(s.app.Updater.State())
}

func (s *SystemService) WailsInstallUpdate() error {
	if s.app == nil || s.app.Updater == nil {
		return errors.New("updater no disponible")
	}
	go func() {
		if err := s.app.Updater.DownloadAndInstall(context.Background()); err != nil {
			// El updater ya emite wails:updater:error; log adicional por si acaso.
			println("[wails-updater] DownloadAndInstall error:", err.Error())
		}
	}()
	return nil
}

func (s *SystemService) WailsRestart() error {
	if s.app == nil || s.app.Updater == nil {
		return errors.New("updater no disponible")
	}
	return s.app.Updater.Restart(context.Background())
}

func (s *SystemService) WailsSkipVersion(version string) {
	if s.app != nil && s.app.Updater != nil {
		s.app.Updater.SkipVersion(version)
	}
}

func (b *runtimeBridge) OpenFileDialog(title string, filters []Handlers.FileFilter) (string, error) {
	if b.app == nil {
		return "", errors.New("aplicación no disponible")
	}
	dlg := b.app.Dialog.OpenFile().SetTitle(title)
	for _, f := range filters {
		dlg.AddFilter(f.DisplayName, f.Pattern)
	}
	return dlg.PromptForSingleSelection()
}

func (b *runtimeBridge) OpenDirectoryDialog(title string) (string, error) {
	if b.app == nil {
		return "", errors.New("aplicación no disponible")
	}
	return b.app.Dialog.OpenFile().SetTitle(title).CanChooseDirectories(true).CanChooseFiles(false).PromptForSingleSelection()
}

func (b *runtimeBridge) BrowserOpenURL(url string) {
	if b.app != nil {
		_ = b.app.Browser.OpenURL(url)
	}
}

func (b *runtimeBridge) Quit() {
	if b.app != nil {
		b.app.Quit()
	}
}
