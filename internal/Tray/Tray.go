package tray

import (
	"sort"

	"StepLauncher/internal/Core/Launcher"
	engine "StepLauncher/internal/Handlers/Engine"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Eventos que el tray emite al frontend para alternar la música, lanzar
// versiones o instancias, abrir vistas y sincronizar la cuenta seleccionada;
// los escuchan Common/Stores/Music.ts, Launcher/Store.ts, Instances/Store.ts,
// Accounts/Store.ts y App.vue.
const (
	musicToggleEvent     = "music_tray_toggle"
	launchVersionEvent   = "tray_launch_version"
	launchInstanceEvent  = "tray_launch_instance"
	openSettingsEvent    = "tray_open_settings"
	openInstancesEvent   = "tray_open_instances"
	openDownloadsEvent   = "tray_open_downloads"
	accountSelectedEvent = "tray_account_selected"
)

// Tray gestiona el icono del área de notificaciones: su menú con las últimas
// sesiones e instancias jugadas, la cuenta activa y las acciones rápidas.
type Tray struct {
	app         *application.App
	engine      *engine.Engine
	sysTray     *application.SystemTray
	menu        *application.Menu
	musicItem   *application.MenuItem
	musicPaused bool
}

// New crea el gestor del tray sin mostrarlo; Setup lo registra en el sistema.
func New(app *application.App, engine *engine.Engine) *Tray {
	return &Tray{app: app, engine: engine}
}

// Setup crea el icono del área de notificaciones con su menú de acciones
// rápidas. Se llama desde ServiceStartup, cuando el historial ya está cargado,
// para que "Últimas Sesiones" e "Instancias" salgan con datos.
//
// Nota: SystemTray.New() ya difiere Run() al arranque de la app
// (runOrDeferToAppRun); llamar aquí a Run() (o en ApplicationStarted)
// ejecutaría run() dos veces y crearía un segundo icono en la bandeja.
func (t *Tray) Setup(icon []byte) {
	t.sysTray = t.app.SystemTray.New()
	t.sysTray.SetTooltip("StepLauncher")
	t.sysTray.SetIcon(icon)

	t.menu = t.buildMenu()
	t.sysTray.SetMenu(t.menu)

	// Clic izquierdo: acción principal (mostrar/ocultar la ventana). Clic
	// derecho: abrir el menú. Doble clic: mostrar y enfocar la ventana.
	t.sysTray.OnClick(func() {
		t.toggleMainWindow()
	})
	t.sysTray.OnRightClick(func() {
		t.sysTray.OpenMenu()
	})
	t.sysTray.OnDoubleClick(func() {
		t.showMainWindow()
	})
}

// Shutdown destruye el icono del tray al cerrar la aplicación para liberar
// los recursos del sistema. Se llama desde ServiceShutdown.
func (t *Tray) Shutdown() {
	if t.sysTray == nil {
		return
	}
	t.sysTray.Destroy()
}

// OnEngineEvent se invoca por cada evento del engine: cuando arranca o termina
// una partida se refresca el menú del tray para reflejar las últimas sesiones.
func (t *Tray) OnEngineEvent(eventType string) {
	switch eventType {
	case string(launcher.EvGameStarting), string(launcher.EvGameStarted),
		string(launcher.EvGameExited), string(launcher.EvGameCrashed), string(launcher.EvGameStopped):
		t.refresh()
	}
}

// buildMenu construye el menú del tray: últimas versiones jugadas (máx. 5),
// últimas instancias jugadas (máx. 5), cuenta activa (máx. 5) y las acciones
// rápidas (música, ventana, ajustes, instancias, descargas, GitHub y salir).
func (t *Tray) buildMenu() *application.Menu {
	t.menu = application.NewMenu()

	// Últimas sesiones: submenú con las versiones jugadas más recientes;
	// cada una lanza esa versión al hacer clic.
	versionsMenu := t.menu.AddSubmenu("Últimas Sesiones")
	versions := t.recentVersions(5)
	if len(versions) == 0 {
		versionsMenu.Add("Sin sesiones aún").SetEnabled(false)
	} else {
		for _, v := range versions {
			versionsMenu.Add("Jugar " + v).OnClick(func(*application.Context) {
				t.launchVersion(v)
			})
		}
	}

	// Últimas instancias: submenú con las instancias jugadas más recientes;
	// cada una lanza esa instancia al hacer clic.
	instancesMenu := t.menu.AddSubmenu("Últimas Instancias")
	instances := t.recentInstances(5)
	if len(instances) == 0 {
		instancesMenu.Add("Sin instancias aún").SetEnabled(false)
	} else {
		for _, name := range instances {
			instancesMenu.Add("Jugar " + name).OnClick(func(*application.Context) {
				t.launchInstance(name)
			})
		}
	}

	// Cuenta: submenú para elegir la cuenta activa (máx. 5) con radio items;
	// la marcada es la cuenta seleccionada actualmente.
	accountsMenu := t.menu.AddSubmenu("Cuenta")
	accounts := t.recentAccounts(5)
	if len(accounts) == 0 {
		accountsMenu.Add("Sin cuentas aún").SetEnabled(false)
	} else {
		selectedID := t.selectedAccountID()
		for _, acc := range accounts {
			label := acc.Username
			if label == "" {
				label = acc.Name
			}
			accountsMenu.AddRadio(label, acc.ID == selectedID).OnClick(func(*application.Context) {
				t.selectAccount(acc.ID)
			})
		}
	}

	t.menu.AddSeparator()

	// Música: botón que alterna la reproducción y cambia su etiqueta según el
	// estado; el menú se actualiza con Update() tras cada cambio.
	musicLabel := "Pausar Música"
	if t.musicPaused {
		musicLabel = "Reproducir Música"
	}
	t.musicItem = t.menu.Add(musicLabel)
	t.musicItem.OnClick(func(*application.Context) {
		t.musicPaused = !t.musicPaused
		t.app.Event.Emit(musicToggleEvent, "")
		t.setMusicLabel(t.musicPaused)
		t.menu.Update()
	})

	t.menu.Add("Mostrar / Ocultar Launcher").OnClick(func(*application.Context) {
		t.toggleMainWindow()
	})

	t.menu.AddSeparator()

	// Las vistas que abre el tray (Ajustes, Instancias, Descargas) muestran la
	// ventana primero: si está minimizada u oculta, se restaura y enfoca para
	// que el usuario vea la vista abierta.
	t.menu.Add("Configuración").OnClick(func(*application.Context) {
		t.showMainWindow()
		t.app.Event.Emit(openSettingsEvent, "")
	})

	t.menu.Add("Abrir Instancias").OnClick(func(*application.Context) {
		t.showMainWindow()
		t.app.Event.Emit(openInstancesEvent, "")
	})

	t.menu.Add("Abrir Descargas").OnClick(func(*application.Context) {
		t.showMainWindow()
		t.app.Event.Emit(openDownloadsEvent, "")
	})

	t.menu.AddSeparator()

	t.menu.Add("GitHub").OnClick(func(*application.Context) {
		_ = t.app.Browser.OpenURL("https://github.com/NovaStepStudio/StepLauncher")
	})

	t.menu.Add("Salir").OnClick(func(*application.Context) {
		t.app.Quit()
	})

	return t.menu
}

// launchVersion emite el evento que pide al frontend lanzar una versión
// concreta (reutiliza el flujo de lanzamiento de Launcher/Store.ts).
func (t *Tray) launchVersion(version string) {
	t.app.Event.Emit(launchVersionEvent, version)
}

// launchInstance emite el evento que pide al frontend lanzar una instancia
// concreta (reutiliza el flujo de lanzamiento de Instances/Store.ts).
func (t *Tray) launchInstance(name string) {
	t.app.Event.Emit(launchInstanceEvent, name)
}

// selectAccount marca la cuenta elegida en el engine, sincroniza al frontend
// y reconstruye el menú para reflejar el nuevo radio marcado.
func (t *Tray) selectAccount(id string) {
	if t.engine != nil {
		_ = t.engine.SetSelectedAccount(id)
	}
	t.app.Event.Emit(accountSelectedEvent, id)
	t.refresh()
}

// setMusicLabel guarda el estado de pausa y ajusta la etiqueta del item de
// música según corresponda.
func (t *Tray) setMusicLabel(paused bool) {
	t.musicPaused = paused
	if t.musicItem == nil {
		return
	}
	if paused {
		t.musicItem.SetLabel("Reproducir Música")
	} else {
		t.musicItem.SetLabel("Pausar Música")
	}
}

// mainWindow devuelve la ventana principal del launcher, o nil si no existe.
func (t *Tray) mainWindow() application.Window {
	win, ok := t.app.Window.Get("main")
	if !ok {
		return nil
	}
	return win
}

// toggleMainWindow muestra la ventana si está oculta y la oculta si está
// visible; si está minimizada la restaura en lugar de ocultarla.
func (t *Tray) toggleMainWindow() {
	win := t.mainWindow()
	if win == nil {
		return
	}
	if win.IsMinimised() {
		win.Restore()
		win.Show()
		win.Focus()
		return
	}
	if win.IsVisible() {
		win.Hide()
	} else {
		win.Show()
		win.Focus()
	}
}

// showMainWindow muestra y enfoca la ventana principal; si está minimizada la
// restaura primero para que vuelva a verse.
func (t *Tray) showMainWindow() {
	win := t.mainWindow()
	if win == nil {
		return
	}
	if win.IsMinimised() {
		win.Restore()
	}
	win.Show()
	win.Focus()
}

// recentVersions devuelve las versiones jugadas más recientes (únicas, en
// orden de última sesión) hasta `limit`.
func (t *Tray) recentVersions(limit int) []string {
	if t.engine == nil || limit <= 0 {
		return nil
	}
	seen := make(map[string]bool)
	out := make([]string, 0, limit)
	for _, e := range recentHistoryEntries(t.engine.GetHistory()) {
		if e.Version == "" || seen[e.Version] {
			continue
		}
		seen[e.Version] = true
		out = append(out, e.Version)
		if len(out) >= limit {
			break
		}
	}
	return out
}

// recentInstances devuelve las instancias jugadas más recientes (únicas, en
// orden de última sesión) hasta `limit`.
func (t *Tray) recentInstances(limit int) []string {
	if t.engine == nil || limit <= 0 {
		return nil
	}
	seen := make(map[string]bool)
	out := make([]string, 0, limit)
	for _, e := range recentHistoryEntries(t.engine.GetHistory()) {
		if e.InstanceName == "" || seen[e.InstanceName] {
			continue
		}
		seen[e.InstanceName] = true
		out = append(out, e.InstanceName)
		if len(out) >= limit {
			break
		}
	}
	return out
}

// recentAccounts devuelve las cuentas disponibles hasta `limit`.
func (t *Tray) recentAccounts(limit int) []engine.AccountInfo {
	if t.engine == nil || limit <= 0 {
		return nil
	}
	list := t.engine.ListAccounts()
	if len(list) > limit {
		list = list[:limit]
	}
	return list
}

// selectedAccountID devuelve el id de la cuenta activa, o "" si no hay.
func (t *Tray) selectedAccountID() string {
	if t.engine == nil {
		return ""
	}
	return t.engine.GetSelectedAccount()
}

// recentHistoryEntries ordena el historial de más reciente a más antiguo.
func recentHistoryEntries(entries []engine.HistoryEntry) []engine.HistoryEntry {
	sorted := make([]engine.HistoryEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp > sorted[j].Timestamp
	})
	return sorted
}

// refresh reconstruye el menú del tray: los datos del historial y de las
// cuentas cambian con el tiempo.
func (t *Tray) refresh() {
	if t.sysTray == nil {
		return
	}
	t.menu = t.buildMenu()
	t.sysTray.SetMenu(t.menu)
}