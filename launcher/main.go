package main

import (
	"embed"
	"log"
	"path/filepath"

	"StepLauncher/internal/Handlers"
	engine "StepLauncher/internal/Handlers/Engine"
	engineconfig "StepLauncher/internal/Handlers/Engine/engineconfig"
	worker "StepLauncher/internal/Updater"

	accsvc "StepLauncher/internal/Services/Account"
	appsvc "StepLauncher/internal/Services/Appearance"
	cfgsvc "StepLauncher/internal/Services/Config"
	dlsvc "StepLauncher/internal/Services/Download"
	gamesvc "StepLauncher/internal/Services/Game"
	instsvc "StepLauncher/internal/Services/Instance"
	modsvc "StepLauncher/internal/Services/ModLoader"
	modscontent "StepLauncher/internal/Services/Mods"
	musicsvc "StepLauncher/internal/Services/Music"
	syssvc "StepLauncher/internal/Services/System"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/updater"
)

// Wails embebe los archivos del frontend compilado dentro del binario.
// Cualquier archivo de frontend/dist se incrusta y queda disponible para el
// frontend. Ver https://pkg.go.dev/embed para más información.

// Repositorio de releases contra el que comprueba actualizaciones,
// directo a GitHub sin APIs externas intermedias.
const updateRepo = "NovaStepStudio/StepLauncher"

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/apptray.png
var trayIcon []byte

func main() {
	app := application.New(application.Options{
		Name:        "StepLauncher",
		Description: "Launcher de Minecraft de StepNick",
		Flags: map[string]any{
			"version": engineconfig.AppVersion,
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		// La música de fondo se reproduce con Howler en el frontend; WebView2
		// bloquea por defecto el audio que arranca sin gesto del usuario
		// (reintentos automáticos, cambio de pista), así que se desactiva esa
		// política.
		Windows: application.WindowsOptions{
			AdditionalBrowserArgs: []string{"--autoplay-policy=no-user-gesture-required"},
		},
	})

	// Inicializar Engine y Handler compartidos (misma lógica que App.NewApp).
	eng, handler := createEngineAndHandler()

	// Inyectar icono del tray al SystemService (evita go:embed con .. en el servicio).
	syssvc.SetAppIcon(trayIcon)

	// Registro de servicios por dominio (nueva API).
	registerDomainServices(app, eng, handler)

	// Inicializa el updater Wails3 con provider GitHub en modo headless.
	// La UI personalizada (Update.vue) escucha los eventos wails:updater:*
	// y muestra el prompt "¿Quieres actualizar? Actualizar Ahora / Más tarde".
	// Sin ventana builtin: el frontend controla todo el flujo.
	if err := initWailsUpdater(app); err != nil {
		log.Printf("[updater] no se pudo inicializar: %v", err)
	}

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            "StepLauncher",
		Width:            1024,
		Height:           600,
		MinWidth:         950,
		MinHeight:        600,
		BackgroundColour: application.NewRGB(0, 0, 0),
		URL:              "/",
	})

	// Run bloquea hasta que la aplicación se cierra. Si ocurre un error,
	// se registra y se sale.
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func defaultConfigDir() string {
	return engineconfig.LoadBootstrap().ResolveWorkDir()
}

// createEngineAndHandler replica la inicialización que hacía App.NewApp.
func createEngineAndHandler() (*engine.Engine, *Handlers.App) {
	eng, err := engine.NewEngine()
	if err != nil {
		log.Printf("Engine init error: %v", err)
		eng = nil
	}
	cfgPath := filepath.Join(defaultConfigDir(), "launcher_config.json")
	if eng != nil {
		cfgPath = filepath.Join(eng.ConfigManager().RootDir(), "launcher_config.json")
	}
	handler := Handlers.NewApp(eng, cfgPath)
	return eng, handler
}

// registerDomainServices crea y registra los 10 servicios por dominio.
func registerDomainServices(wailsApp *application.App, eng *engine.Engine, handler *Handlers.App) {
	systemService := syssvc.NewSystemService(wailsApp, handler, eng)
	configService := cfgsvc.NewConfigService(handler, eng)
	instanceService := instsvc.NewInstanceService(handler, eng)
	gameService := gamesvc.NewGameService(eng)
	downloadService := dlsvc.NewDownloadService(eng)
	accountService := accsvc.NewAccountService(handler, eng)
	modLoaderService := modsvc.NewModLoaderService(eng)
	modsService := modscontent.NewModsService(eng)
	musicService := musicsvc.NewMusicService(handler)
	appearanceService := appsvc.NewAppearanceService(handler, eng)

	wailsApp.RegisterService(application.NewService(systemService))
	wailsApp.RegisterService(application.NewService(configService))
	wailsApp.RegisterService(application.NewService(instanceService))
	wailsApp.RegisterService(application.NewService(gameService))
	wailsApp.RegisterService(application.NewService(downloadService))
	wailsApp.RegisterService(application.NewService(accountService))
	wailsApp.RegisterService(application.NewService(modLoaderService))
	wailsApp.RegisterService(application.NewService(modsService))
	wailsApp.RegisterService(application.NewService(musicService))
	wailsApp.RegisterService(application.NewService(appearanceService))
}

// initWailsUpdater configura app.Updater con el provider oficial de
// GitHub Releases en headless, directo contra el repositorio.
// La UI personalizada (Update.vue) escucha los eventos wails:updater:*
// y muestra el prompt "¿Quieres actualizar? Actualizar Ahora / Más tarde".
// Sin ventana builtin: el frontend controla todo el flujo.
// La instalación en sí la aplica el flujo legacy (Engine/Update.go):
// en Windows descarga el *-installer.exe y cierra el launcher.
func initWailsUpdater(app *application.App) error {
	provider, err := worker.NewGithubProvider(updateRepo)
	if err != nil {
		return err
	}
	return app.Updater.Init(updater.Config{
		CurrentVersion: engineconfig.AppVersion,
		Providers:      []updater.Provider{provider},
		Window:         updater.WindowNone,
	})
}
