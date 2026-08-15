package launcher

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	authlib "StepLauncher/internal/Core/Auth"
	downloader "StepLauncher/internal/Core/Downloader"
	globalutils "StepLauncher/internal/Core/Utils"

	helpers "StepLauncher/internal/Core/Launcher/Helpers"
	gamelog "StepLauncher/internal/Core/Launcher/Log"
	utils "StepLauncher/internal/Core/Launcher/Utils"
)

type Launcher struct {
	cfg              LaunchConfig
	ver              *downloader.VersionJSON
	gameLog          *gamelog.GameLogManager
	gameLogBroadcast func(stream, line string)
	eventBroadcast   func([]byte)
	onGameExit       func(*GameInstance, int)
	gameInstance     *GameInstance
	resolvedAuthRoot     string
	prefetchedMeta       string
	resolvedInjectorPath string
}

func NewLauncher(cfg LaunchConfig) *Launcher {
	return &Launcher{cfg: cfg}
}

var crashPatterns = []struct {
	pattern  string
	category string
	reason   string
}{
	{"OutOfMemoryError", "oom", "out_of_memory"},
	{"# A fatal error has been detected by the Java Runtime Environment", "java_vm_crash", "jvm_fatal_error"},
	{"# Java VM: OpenJDK", "java_vm_crash", "jvm_fatal_error"},
	{"#  EXCEPTION_ACCESS_VIOLATION", "java_vm_crash", "access_violation"},
	{"#  EXCEPTION_STACK_OVERFLOW", "java_vm_crash", "stack_overflow"},
	{"# Internal Error", "java_vm_crash", "internal_error"},
	{"Exiting with error code", "game_error", "game_exit_error"},
	{"Process exited with code", "game_error", "process_exit_error"},
	{"Unhandled exception", "game_error", "unhandled_exception"},
	{"Could not create the Java Virtual Machine", "jvm_launch", "jvm_creation_failed"},
	{"Error: Could not find or load main class", "jvm_launch", "main_class_not_found"},
	{"Java heap space", "oom", "heap_oom"},
	{"Metaspace", "oom", "metaspace_oom"},
	{"GC overhead limit exceeded", "oom", "gc_overhead_limit"},
}

func (l *Launcher) adv() AdvancedConfig { return l.cfg.Adv() }

func (l *Launcher) Launch() (*GameInstance, error) {
	adv := l.adv()

	instance := &GameInstance{
		Version:      l.cfg.Version,
		InstanceID:   l.cfg.InstanceID,
		InstanceName: l.cfg.InstanceName,
		PlayerName:   l.cfg.Username,
		StartTime:    time.Now(),
		Status:       GameStarting,
		done:         make(chan struct{}),
	}
	l.gameInstance = instance

	BroadcastStarting(l.eventBroadcast, instance)

	if err := l.initPaths(); err != nil {
		return nil, err
	}

	if adv.AuthLibConfig.Enabled && adv.AuthLibConfig.PreVerifyServer {
		if err := l.preVerifyAuthServer(adv.AuthLibConfig); err != nil {
			return nil, fmt.Errorf("auth server pre-verify failed: %w", err)
		}
		l.log("Auth server verified: %s", adv.AuthLibConfig.AuthServerURL)
	}

	if adv.AuthLibConfig.Enabled {
		if err := l.prepareAuthInjector(adv.AuthLibConfig); err != nil {
			return nil, fmt.Errorf("authlib-injector: %w", err)
		}
	}

	if err := l.readVersionJSON(); err != nil {
		return nil, fmt.Errorf("read version: %w", err)
	}
	adv = l.adv()

	// Los builds tardíos de Forge para 1.16.5 (36.2.34+) parchean
	// MainWindow.class con los callbacks *CallbackI que solo existen en
	// LWJGL 3.3+, pero vanilla declara LWJGL 3.2.2: el juego muere al arrancar
	// con NoClassDefFoundError. Si el client jar de Forge usa esos callbacks,
	// todas las librerías org.lwjgl 3.2.x se suben a 3.3.1 (igual que hacen
	// otros launchers); las nuevas se descargan al lanzar si faltan.
	if forgeClientNeedsNewLWJGL(l.ver.Libraries, adv.LibrariesDir) {
		if n := overrideLWJGLVersion(l.ver.Libraries, lwjglCompatOverrideVersion); n > 0 {
			l.log("LWJGL 3.2.x → %s: %d librerías org.lwjgl reescritas (compatibilidad con el client de Forge)", lwjglCompatOverrideVersion, n)
		}
	}

	// Las versiones antiguas (pre-1.17) no declaran javaVersion.component en su
	// version.json; el launcher oficial usa entonces jre-legacy (Java 8), el
	// único runtime compatible con los modloaders de esas versiones.
	component := l.ver.JavaVersion.Component
	useOfficial := adv.UseOfficialJava
	if component == "" {
		component = "jre-legacy"
		if !useOfficial && adv.JavaExec == "" {
			// Auto-conmutación: si el Java del sistema es >= 17, los modloaders
			// antiguos (Forge 1.12.2, etc.) no arrancan. Se usa el Java 8
			// oficial automáticamente, igual que hace el launcher oficial.
			if sys, err := helpers.ResolveJava("", adv.RuntimeDir, false, ""); err == nil {
				if major := helpers.DetectJavaMajorVersion(sys); major >= 17 {
					useOfficial = true
					l.log("Java del sistema (%s, mayor %d) incompatible con %s; usando Java 8 oficial (jre-legacy)", sys, major, l.cfg.Version)
				}
			}
		}
	}

	if useOfficial {
		if err := l.ensureOfficialJava(); err != nil {
			l.log("WARN: Java oficial no disponible (%v); usando el Java del sistema", err)
			useOfficial = false
		}
	}

	javaPath, err := helpers.ResolveJava(
		component,
		adv.RuntimeDir,
		useOfficial,
		adv.JavaExec,
	)
	if err != nil {
		return nil, fmt.Errorf("resolve java: %w", err)
	}
	logName := l.cfg.LauncherName
	if l.cfg.InstanceName != "" {
		logName = l.cfg.InstanceName
	}
	gl, err := gamelog.NewGameLogManager(gamelog.GameLogConfig{
		LogDir:       l.cfg.LogDir,
		LauncherName: logName,
		Version:      l.cfg.Version,
		Limit:        adv.GameLogLines,
		KeepDays:     adv.LogKeepDays,
		MaxFiles:     adv.LogMaxFiles,
	})
	if err != nil {
		return nil, fmt.Errorf("create game log: %w", err)
	}
	l.gameLog = gl
	if l.gameLogBroadcast != nil {
		gl.SetBroadcastFn(l.gameLogBroadcast)
	}

	l.log("=== Launching Minecraft %s ===", l.cfg.Version)
	l.log("Launcher: %s v%s", l.cfg.LauncherName, l.cfg.LauncherVersion)
	l.log("Game dir: %s", adv.GameDir)
	l.log("Assets dir: %s", adv.AssetsDir)
	l.log("Libraries dir: %s", adv.LibrariesDir)
	l.log("Versions dir: %s", adv.VersionsDir)
	l.log("Java: %s", javaPath)

	if adv.ExecutionPlan != nil {
		l.log("Modloader: %s", adv.ExecutionPlan.MainClass)
	}

	baseVer := adv.BaseVersion
	if baseVer == "" {
		baseVer = l.cfg.Version
	}

	effectiveMain := l.ver.MainClass
	if adv.ExecutionPlan != nil && adv.ExecutionPlan.MainClass != "" {
		effectiveMain = adv.ExecutionPlan.MainClass
	}

	clientJarVer := baseVer
	if strings.Contains(effectiveMain, "bootstraplauncher") {
		clientJarVer = l.cfg.Version
		if err := ensureClientJarCopy(adv.VersionsDir, baseVer, l.cfg.Version); err != nil {
			l.log("WARN: no se pudo copiar el client jar para %s: %v", l.cfg.Version, err)
		}
	}

	nativesDir := adv.NativesDir
	if nativesDir == "" {
		nativesDir = helpers.NativesDir(adv.NativesBaseDir, baseVer)
	}

	classpath, cpEntries := helpers.BuildClasspath(
		l.ver.Libraries,
		adv.LibrariesDir,
		adv.VersionsDir,
		clientJarVer,
	)

	if adv.ExecutionPlan != nil && len(adv.ExecutionPlan.AdditionalClasspath) > 0 {
		sep := ";"
		if runtime.GOOS != "windows" {
			sep = ":"
		}
		for _, p := range adv.ExecutionPlan.AdditionalClasspath {
			exists := true
			if _, err := os.Stat(p); err != nil {
				exists = false
			}
			cpEntries = append(cpEntries, helpers.ClasspathEntry{Path: p, Exists: exists})
			classpath += sep + p
		}
	}

	if adv.ProfileVersion != "" {
		profileLibs := l.loadProfileLibraries(adv.ProfileVersion)
		if profileLibs != nil {
			l.ver.Libraries = append(l.ver.Libraries, profileLibs...)
		}
	}

	if err := l.downloadMissingLibraries(&cpEntries, clientJarVer); err != nil {
		l.gameLog.Close()
		return nil, fmt.Errorf("missing libraries: %w", err)
	}

	l.log("Classpath entries: %d", len(cpEntries))

	if !adv.SkipNativeExtract {
		var jvmArgs []interface{}
		if l.ver.Arguments != nil {
			jvmArgs = l.ver.Arguments.JVM
		}
		l.prepareEmit("natives", 0, 0, "", "Extrayendo archivos nativos…", false)
		extracted, err := helpers.ExtractNatives(l.ver.Libraries, adv.LibrariesDir, nativesDir, jvmArgs, func(cur, total int, name string) {
			switch {
			case total == 0:
				l.prepareEmit("natives", 0, 0, "", "", true)
			case cur == 0:
				l.prepareEmit("natives", 0, total, "", "Extrayendo archivos nativos…", false)
			case cur < total:
				l.prepareEmit("natives", cur, total, name, "Extrayendo archivos nativos…", false)
			default:
				l.prepareEmit("natives", total, total, "", "", true)
			}
		})
		if err != nil {
			l.log("WARN: native extraction failed: %v", err)
		} else {
			l.log("Extracted %d native file(s) to %s", extracted, nativesDir)
		}
	}

	l.ensureGameDir()

	launcherProps := ""
	if adv.ExecutionPlan != nil {
		launcherProps = l.ensureLauncherProperties(adv.GameDir)
	}

	vars := helpers.BuildVarsMap(
		helpers.VarConfig{
			Username:           l.cfg.Username,
			UUID:               l.cfg.UUID,
			AccessToken:        l.cfg.AccessToken,
			XUID:               l.cfg.XUID,
			ClientID:           l.cfg.ClientID,
			GameDir:            adv.GameDir,
			AssetsDir:          adv.AssetsDir,
			LibrariesDir:       adv.LibrariesDir,
			LauncherName:       l.cfg.LauncherName,
			LauncherVersion:    l.cfg.LauncherVersion,
			DemoUser:           adv.DemoUser,
			CustomResolution:   adv.CustomResolution,
			ResWidth:           adv.ResWidth,
			ResHeight:          adv.ResHeight,
			UserType:           adv.UserType,
			LauncherProperties: launcherProps,
		},
		l.ver.ID, l.ver.Type, l.advAssetIndexID(),
		classpath, nativesDir,
	)

	jvmArgs := l.buildJVMArgs(javaPath, vars, adv)
	gameArgs := l.buildGameArgs(vars, adv)
	mainClass := effectiveMain
	if adv.ExecutionPlan != nil {
		extraJVM := make([]string, len(adv.ExecutionPlan.AdditionalJVMArgs))
		for i, a := range adv.ExecutionPlan.AdditionalJVMArgs {
			extraJVM[i] = helpers.SubstituteVars(a, vars)
		}
		jvmArgs = append(extraJVM, jvmArgs...)
		for _, a := range adv.ExecutionPlan.AdditionalGameArgs {
			gameArgs = append(gameArgs, helpers.SubstituteVars(a, vars))
		}
	}

	assetIndexVirtual := "unset"
	if adv.AssetIndexVirtual != nil {
		if *adv.AssetIndexVirtual {
			assetIndexVirtual = "true"
		} else {
			assetIndexVirtual = "false"
		}
	}

	preInfo := gamelog.PreLaunchInfo{
		Version:           l.cfg.Version,
		MainClass:         mainClass,
		JavaExec:          javaPath,
		MinRAM:            adv.MinRAM,
		MaxRAM:            adv.MaxRAM,
		GCPreset:          adv.GCPreset,
		GPUPreference:     adv.GPUPreference,
		HWAccelDisabled:   l.isHWAccelDisabled(),
		GameDir:           adv.GameDir,
		AssetsDir:         adv.AssetsDir,
		LibrariesDir:      adv.LibrariesDir,
		NativesDir:        nativesDir,
		AssetIndexID:      l.ver.AssetIndex.ID,
		AssetIndexVirtual: assetIndexVirtual,
		LauncherName:      l.cfg.LauncherName,
		LauncherVersion:   l.cfg.LauncherVersion,
		JVMArgs:           jvmArgs,
		GameArgs:          gameArgs,
	}
	for _, e := range cpEntries {
		preInfo.ClasspathEntries = append(preInfo.ClasspathEntries, gamelog.ClasspathEntry{
			Path: e.Path, Exists: e.Exists,
		})
	}
	renderNatives := helpers.NativesSubDirs(nativesDir)
	for _, d := range renderNatives {
		preInfo.Natives = append(preInfo.Natives, gamelog.NativeEntry{Path: d})
	}
	preInfoCopy := preInfo
	instance.PreInfo = &preInfoCopy
	l.gameLog.WritePreLaunchInfo(preInfo)

	if adv.PreLaunchCommand != "" {
		l.log("Executing pre-launch command: %s", adv.PreLaunchCommand)
		if err := l.execHook(adv.PreLaunchCommand, adv.GameDir); err != nil {
			l.log("WARN: pre-launch command failed: %v", err)
		}
	}

	logPath := l.gameLog.GetLogPath()

	procEnv := make(map[string]string)
	for k, v := range helpers.GPUEnvVars(adv.GPUPreference) {
		procEnv[k] = v
	}
	for k, v := range adv.EnvironmentVars {
		procEnv[k] = v
	}
	if len(procEnv) > 0 {
		l.log("Setting %d environment variable(s) for game process", len(procEnv))
	}

	cmd, mcLogFile, err := utils.LaunchProcess(
		javaPath, mainClass, adv.GameDir, logPath,
		jvmArgs, gameArgs, procEnv,
	)
	if err != nil {
		l.log("ERROR: Launch failed: %v", err)
		l.gameLog.Close()
		return nil, err
	}

	instance.PID = cmd.Process.Pid
	instance.LogPath = logPath
	instance.cmd = cmd
	instance.Status = GameRunning

	l.log("Minecraft launched (PID: %d)", instance.PID)
	BroadcastStarted(l.eventBroadcast, instance)

	go l.waitForExit(instance, mcLogFile)

	return instance, nil
}

func (l *Launcher) advAssetIndexID() string {
	adv := l.adv()
	if adv.AssetIndexID != "" {
		return adv.AssetIndexID
	}
	if l.ver != nil && l.ver.AssetIndex.ID != "" {
		return l.ver.AssetIndex.ID
	}
	return "index"
}

func (l *Launcher) buildJVMArgs(javaPath string, vars map[string]string, adv AdvancedConfig) []string {
	var args []string

if l.ver.Arguments != nil {
		args = append(args, helpers.BuildJVMArgs(l.ver.Arguments.JVM, vars)...)
	}

	hasLibPath := false
	for _, a := range args {
		if strings.HasPrefix(a, "-Djava.library.path") {
			hasLibPath = true
			break
		}
	}
	if !hasLibPath {
		args = append(args, "-Djava.library.path="+vars["natives_directory"])
	}

	maxMem := adv.MaxRAM
	minMem := helpers.MinRAM
	if maxMem > 0 && minMem > maxMem {
		minMem = maxMem
	}

	var override []string
	if minMem > 0 {
		override = append(override, fmt.Sprintf("-Xms%dM", minMem))
	}
	if maxMem > 0 {
		override = append(override, fmt.Sprintf("-Xmx%dM", maxMem))
	}
	if adv.MaxMetaspaceSize > 0 {
		override = append(override, fmt.Sprintf("-XX:MaxMetaspaceSize=%dM", adv.MaxMetaspaceSize))
	}
	if adv.StackSize > 0 {
		override = append(override, fmt.Sprintf("-Xss%dK", adv.StackSize))
	}

	if adv.DirectMemorySize > 0 {
		override = append(override, fmt.Sprintf("-XX:MaxDirectMemorySize=%dM", adv.DirectMemorySize))
	}
	if adv.ReservedCodeCache > 0 {
		override = append(override, fmt.Sprintf("-XX:ReservedCodeCacheSize=%dM", adv.ReservedCodeCache))
	}
	if adv.MetaspaceSize > 0 {
		override = append(override, fmt.Sprintf("-XX:MetaspaceSize=%dM", adv.MetaspaceSize))
	}

	gcFlags := helpers.GCFlags(adv.GCPreset)
	override = append(override, gcFlags...)

	if helpers.DetectJavaMajorVersion(javaPath) >= 17 {
		override = append(override, "--enable-native-access=ALL-UNNAMED")
	}

	if adv.JavaModulePath != "" {
		override = append(override, "--module-path", adv.JavaModulePath)
	}
	for _, mod := range adv.JavaAddModules {
		override = append(override, "--add-modules", mod)
	}
	for _, exp := range adv.JavaAddExports {
		override = append(override, "--add-exports", exp)
	}
	for _, opn := range adv.JavaAddOpens {
		override = append(override, "--add-opens", opn)
	}

	override = append(override, "-Dminecraft.launcher.brand="+l.cfg.LauncherName)
	override = append(override, "-Dminecraft.launcher.version="+l.cfg.LauncherVersion)

	if adv.WindowTitle != "" {
		override = append(override, "-Dminecraft.window.title="+adv.WindowTitle)
	}

	if adv.AuthLibConfig.Enabled && l.resolvedInjectorPath != "" {
		apiRoot := adv.AuthLibConfig.AuthServerURL
		if l.resolvedAuthRoot != "" {
			apiRoot = l.resolvedAuthRoot
		}
		javaagent := fmt.Sprintf("-javaagent:%s=%s", l.resolvedInjectorPath, apiRoot)
		override = append(override, javaagent)
	}
	if l.prefetchedMeta != "" {
		override = append(override, "-Dauthlibinjector.yggdrasil.prefetched="+l.prefetchedMeta)
	}

	for flag := range adv.JVMFlags {
		override = append(override, flag)
	}

	override = append(override, adv.JavaArgs...)

	cp := vars["classpath"]

	tplModulePath := false
	tplModulePathValue := ""
	var filtered []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--module-path" || a == "-p" {
			tplModulePath = true
			if i+1 < len(args) {
				nxt := args[i+1]
				if nxt != "${classpath}" && nxt != cp {
					tplModulePathValue = nxt
				}
				i++
			}
			continue
		}
		if a == "-cp" || a == "--class-path" || a == cp || a == "${classpath}" {
			continue
		}
		filtered = append(filtered, a)
	}

	useModulePath := adv.ExecutionPlan != nil && adv.ExecutionPlan.UseModulePath

	result := append(filtered, override...)
	switch {
	case tplModulePath && tplModulePathValue != "":
		result = append(result, "--module-path", tplModulePathValue)
		result = append(result, "-cp", cp)
	case tplModulePath:
		result = append(result, "--module-path", cp)
		result = append(result, "-cp", cp)
	case useModulePath:
		result = append(result, "--module-path", cp)
		result = append(result, "-cp", cp)
	default:
		result = append(result, "-cp", cp)
	}

	for i, a := range result {
		result[i] = helpers.SubstituteVars(a, vars)
	}

	return result
}

func (l *Launcher) buildGameArgs(vars map[string]string, adv AdvancedConfig) []string {
	features := helpers.BuildFeaturesMap(adv.DemoUser, adv.CustomResolution)

	var raw []interface{}
	var mcArgs string
	if l.ver.MinecraftArguments != "" {
		mcArgs = l.ver.MinecraftArguments
	} else if l.ver.Arguments != nil {
		raw = l.ver.Arguments.Game
	}

	result := helpers.BuildGameArgs(raw, mcArgs, vars, features, adv.Fullscreen)

	if adv.ServerAddress != "" {
		result = append(result, "--server", adv.ServerAddress)
		port := adv.ServerPort
		if port <= 0 {
			port = 25565
		}
		result = append(result, "--port", fmt.Sprintf("%d", port))
	}

	if adv.QuickPlayPath != "" {
		result = append(result, "--quickPlayPath", adv.QuickPlayPath)
	}

	if adv.LogLevel != "" {
		result = append(result, "--log-level", adv.LogLevel)
	}

	if adv.MinecraftLogConfig != "" {
		result = append(result, "--log-config", adv.MinecraftLogConfig)
	}

	if adv.AllowServerList != nil {
		result = append(result, "--server-list-allowed", fmt.Sprintf("%t", *adv.AllowServerList))
	}
	if adv.AllowMultiplayer != nil {
		result = append(result, "--multiplayer-allowed", fmt.Sprintf("%t", *adv.AllowMultiplayer))
	}
	if adv.AllowChat != nil {
		result = append(result, "--chat-allowed", fmt.Sprintf("%t", *adv.AllowChat))
	}
	if adv.AllowRealms != nil {
		result = append(result, "--realm-allowed", fmt.Sprintf("%t", *adv.AllowRealms))
	}
	if adv.FramerateLimit > 0 {
		result = append(result, "--framerateLimit", fmt.Sprintf("%d", adv.FramerateLimit))
	}
	if adv.Renderer != "" {
		result = append(result, "--renderer", adv.Renderer)
	}

	result = append(result, adv.GameArgs...)

	for i, a := range result {
		result[i] = helpers.SubstituteVars(a, vars)
	}

	return result
}

func (l *Launcher) isHWAccelDisabled() bool {
	adv := l.adv()
	if adv.HardwareAcceleration != nil {
		return !*adv.HardwareAcceleration
	}
	return true
}

func (l *Launcher) initPaths() error {
	adv := l.adv()
	os.MkdirAll(adv.GameDir, 0755)
	os.MkdirAll(adv.AssetsDir, 0755)
	os.MkdirAll(adv.LibrariesDir, 0755)
	os.MkdirAll(adv.VersionsDir, 0755)
	return nil
}

func (l *Launcher) ensureOfficialJava() error {
	adv := l.adv()
	component := l.ver.JavaVersion.Component
	// Las versiones antiguas (pre-1.17) no declaran componente: se usa
	// jre-legacy (Java 8), el runtime oficial para esas versiones.
	if component == "" {
		component = "jre-legacy"
	}
	if _, err := helpers.ResolveJava(component, adv.RuntimeDir, true, ""); err == nil {
		return nil
	}
	if adv.RuntimeDir == "" {
		return fmt.Errorf("official java not found and no runtime dir configured")
	}

	l.log("Official Java component %q not found, downloading...", component)

	workDir := filepath.Dir(adv.RuntimeDir)
	cfg := downloader.Config{
		WorkDir:        workDir,
		CacheDir:       adv.CacheDir,
		JavaRuntimeDir: adv.RuntimeDir,
		HTTPClient:     downloader.DefaultHTTPClient(),
		LogFn:          l.log,
	}

	tasks, err := downloader.BuildJavaRuntimeTasks(l.ver, cfg)
	if err != nil {
		return fmt.Errorf("build java runtime tasks: %w", err)
	}

	ctx := context.Background()
	total := len(tasks)
	failed := 0
	for i, task := range tasks {
		l.log("Downloading [%d/%d] %s", i+1, total, filepath.Base(task.Dest))
		os.MkdirAll(filepath.Dir(task.Dest), 0755)
		if err := downloader.DownloadFile(ctx, task, cfg.HTTPClient, 3, nil, 60000, 3); err != nil {
			l.log("WARN [%d/%d] failed: %s: %v", i+1, total, filepath.Base(task.Dest), err)
			failed++
		}
	}
	if failed > 0 {
		return fmt.Errorf("failed to download %d/%d java runtime files", failed, total)
	}

	if _, err := helpers.ResolveJava(component, adv.RuntimeDir, true, ""); err != nil {
		return fmt.Errorf("official java still missing after download: %w", err)
	}
	l.log("Official Java component %q ready", component)
	return nil
}

func (l *Launcher) prepareEmit(phase string, current, total int, label, message string, finished bool) {
	if l.eventBroadcast == nil {
		return
	}
	BroadcastPrepare(l.eventBroadcast, &GamePrepareData{
		Version:  l.cfg.Version,
		Phase:    phase,
		Current:  current,
		Total:    total,
		Label:    label,
		Message:  message,
		Finished: finished,
	})
}

func (l *Launcher) downloadMissingLibraries(cpEntries *[]helpers.ClasspathEntry, clientJarVer string) error {
	adv := l.adv()
	if adv.DisableLibraries {
		l.log("Library download disabled by config, skipping")
		return nil
	}

	type missingEntry struct {
		path string
		lib  downloader.Library
	}
	var missing []missingEntry
	for _, e := range *cpEntries {
		if !e.Exists {
			missing = append(missing, missingEntry{path: e.Path})
		}
	}
	// Los jars nativos también se comprueban: el override de LWJGL por
	// compatibilidad (3.2.x → 3.3.1) introduce jars que el instalador de la
	// versión no bajó (descargó los 3.2.2); sin ellos la extracción de
	// natives no encuentra nada y el juego muere con UnsatisfiedLinkError.
	for _, lib := range l.ver.Libraries {
		if !downloader.MatchRules(lib.Rules) || !downloader.IsNativeLibrary(lib) {
			continue
		}
		dest, _, _, _ := helpers.ResolveNativeJarDownload(lib, adv.LibrariesDir, globalutils.OsName())
		if dest == "" {
			continue
		}
		if _, err := os.Stat(dest); err == nil {
			continue
		}
		missing = append(missing, missingEntry{path: dest})
	}
	if len(missing) == 0 {
		return nil
	}

	libByPath := make(map[string]downloader.Library)
	for _, lib := range l.ver.Libraries {
		if !downloader.MatchRules(lib.Rules) || downloader.IsNativeLibrary(lib) {
			continue
		}
		dest, _, _, _ := helpers.ResolveLibraryDownload(lib, adv.LibrariesDir)
		if dest != "" {
			libByPath[dest] = lib
		}
	}
	// Nativos: mismo mapa para poder resolver la URL en el bucle de descarga.
	for _, lib := range l.ver.Libraries {
		if !downloader.MatchRules(lib.Rules) || !downloader.IsNativeLibrary(lib) {
			continue
		}
		dest, _, _, _ := helpers.ResolveNativeJarDownload(lib, adv.LibrariesDir, globalutils.OsName())
		if dest != "" {
			libByPath[dest] = lib
		}
	}
	// El client jar debe llevar el id de la versión lanzada (clientJarVer)
	// para que el ignoreList del bootstraplauncher lo excluya del procesado
	// de módulos; si la copia temprana falló porque faltaba el jar base, se
	// descarga directamente al destino correcto.
	clientJar := filepath.Join(adv.VersionsDir, clientJarVer, clientJarVer+".jar")

	for i := range missing {
		p := &missing[i]
		if lib, ok := libByPath[p.path]; ok {
			p.lib = lib
		}
	}

	l.log("Found %d missing libraries, downloading...", len(missing))
	l.prepareEmit("libraries", 0, len(missing), "", "", false)

	ctx := context.Background()
	total := len(missing)
	failed := 0

	for i, m := range missing {
		l.prepareEmit("libraries", i+1, total, filepath.Base(m.path), "", false)
		if m.path == clientJar && l.ver.Downloads.Client.URL != "" {
			label := "client jar: " + filepath.Base(m.path)
			l.log("Downloading [%d/%d] %s", i+1, total, label)
			os.MkdirAll(filepath.Dir(m.path), 0755)
			task := downloader.DownloadTask{
				URL:  l.ver.Downloads.Client.URL,
				Dest: m.path,
				SHA1: l.ver.Downloads.Client.SHA1,
				Size: l.ver.Downloads.Client.Size,
			}
			if err := downloader.DownloadFile(ctx, task, http.DefaultClient, 3, nil, 60000, 3); err != nil {
				l.log("WARN [%d/%d] failed: %s: %v", i+1, total, label, err)
				failed++
			} else {
				l.log("  âœ“ [%d/%d] %s", i+1, total, label)
			}
			continue
		}

		if m.lib.Name == "" {
			l.log("WARN [%d/%d] cannot resolve download for %s", i+1, total, m.path)
			continue
		}

		_, url, sha1, size := helpers.ResolveLibraryDownload(m.lib, adv.LibrariesDir)
		if url == "" && downloader.IsNativeLibrary(m.lib) {
			_, url, sha1, size = helpers.ResolveNativeJarDownload(m.lib, adv.LibrariesDir, globalutils.OsName())
		}
		if url == "" {
			l.log("WARN [%d/%d] no URL for %s", i+1, total, m.lib.Name)
			failed++
			continue
		}

		sizeStr := ""
		if size > 0 {
			sizeStr = fmt.Sprintf(" (%.1f MB)", float64(size)/1024/1024)
		}
		l.log("Downloading [%d/%d] %s%s", i+1, total, m.lib.Name, sizeStr)
		os.MkdirAll(filepath.Dir(m.path), 0755)
		task := downloader.DownloadTask{URL: url, Dest: m.path, SHA1: sha1, Size: size}
		if err := downloader.DownloadFile(ctx, task, http.DefaultClient, 3, nil, 60000, 3); err != nil {
			// Los version.json antiguos de Forge apuntan al jar "plain"
			// (forge-X.jar) que ya no existe en maven; el jar real se llama
			// forge-X-universal.jar. Se reintenta con ese sufijo.
			if fbURL := universalForgeURL(m.lib, url); fbURL != "" {
				l.log("Retrying [%d/%d] %s with -universal.jar", i+1, total, m.lib.Name)
				fbTask := downloader.DownloadTask{URL: fbURL, Dest: m.path}
				if err2 := downloader.DownloadFile(ctx, fbTask, http.DefaultClient, 3, nil, 60000, 3); err2 == nil {
					l.log("  âœ“ [%d/%d] %s", i+1, total, m.lib.Name)
					continue
				}
			}
			l.log("WARN [%d/%d] failed: %s: %v", i+1, total, m.lib.Name, err)
			failed++
		} else {
			l.log("  âœ“ [%d/%d] %s", i+1, total, m.lib.Name)
		}
	}

	*cpEntries = helpers.RecheckClasspathEntries(*cpEntries)

	if failed > 0 {
		l.log("WARN: %d/%d libraries could not be downloaded, continuing anyway", failed, total)
	}
	l.prepareEmit("libraries", total, total, "", "", true)
	if failed == total {
		return fmt.Errorf("all %d libraries failed to download", failed)
	}
	return nil
}

// universalForgeURL devuelve la URL del jar -universal de Forge cuando la URL
// original apunta al jar "plain" (forge-X.jar), que ya no existe en maven pero
// que los version.json antiguos de Forge referencian (p. ej. Forge 1.12.2).
func universalForgeURL(lib downloader.Library, url string) string {
	if !strings.Contains(lib.Name, ":forge:") {
		return ""
	}
	if filepath.Ext(url) != ".jar" {
		return ""
	}
	base := strings.TrimSuffix(url, ".jar")
	if strings.HasSuffix(base, "-universal") {
		return ""
	}
	return base + "-universal.jar"
}

func (l *Launcher) readVersionJSON() error {
	ver, err := l.loadVersion(l.cfg.Version)
	if err != nil {
		return err
	}
	l.ver = ver
	adv := l.adv()
	if adv.BaseVersion == "" {
		verPath := filepath.Join(adv.VersionsDir, l.cfg.Version, l.cfg.Version+".json")
		if raw, err := os.ReadFile(verPath); err == nil {
			var partial struct {
				InheritsFrom string `json:"inheritsFrom"`
				Jar          string `json:"jar"`
			}
			switch {
			case json.Unmarshal(raw, &partial) == nil && partial.InheritsFrom != "":
				adv.BaseVersion = partial.InheritsFrom
				l.cfg.Advanced = &adv
			case partial.Jar != "":
				adv.BaseVersion = partial.Jar
				l.cfg.Advanced = &adv
			}
		}
		if adv.BaseVersion == "" {
			adv.BaseVersion = l.cfg.Version
			l.cfg.Advanced = &adv
		}
	}

	return nil
}

func (l *Launcher) loadProfileLibraries(profileVersion string) []downloader.Library {
	merged, err := l.loadVersion(profileVersion)
	if err != nil {
		return nil
	}
	existing := make(map[string]bool)
	for _, lib := range l.ver.Libraries {
		if lib.Name != "" {
			existing[lib.Name] = true
		}
	}
	var newLibs []downloader.Library
	for _, lib := range merged.Libraries {
		if !existing[lib.Name] {
			newLibs = append(newLibs, lib)
		}
	}
	l.log("Merged %d new libraries from profile %s", len(newLibs), profileVersion)
	return newLibs
}

func (l *Launcher) loadVersion(version string) (*downloader.VersionJSON, error) {
	adv := l.adv()
	verPath := filepath.Join(adv.VersionsDir, version, version+".json")
	data, err := os.ReadFile(verPath)
	if err != nil {
		return nil, fmt.Errorf("version JSON not found at %s", verPath)
	}
	var ver downloader.VersionJSON
	if err := json.Unmarshal(data, &ver); err != nil {
		return nil, fmt.Errorf("parse version JSON for %s: %w", version, err)
	}

	if ver.InheritsFrom != "" {
		parent, err := l.loadVersion(ver.InheritsFrom)
		if err != nil {
			return nil, fmt.Errorf("resolve parent version %s: %w", ver.InheritsFrom, err)
		}
		merged := mergeVersions(parent, &ver)
		return merged, nil
	}

	return &ver, nil
}

func mergeVersions(parent, child *downloader.VersionJSON) *downloader.VersionJSON {
	merged := *parent

	merged.ID = child.ID
	merged.InheritsFrom = ""

	merged.Libraries = mergeLibraries(parent.Libraries, child.Libraries)

	if child.MainClass != "" {
		merged.MainClass = child.MainClass
	}

	if child.Type != "" {
		merged.Type = child.Type
	}

	if child.Downloads.Client.URL != "" {
		merged.Downloads = child.Downloads
	}

	if child.AssetIndex.ID != "" {
		merged.AssetIndex = child.AssetIndex
	} else if parent.AssetIndex.ID != "" {
		merged.AssetIndex = parent.AssetIndex
	}

	if child.JavaVersion.Component != "" {
		merged.JavaVersion = child.JavaVersion
	} else if parent.JavaVersion.Component != "" {
		merged.JavaVersion = parent.JavaVersion
	}

	if child.Logging != nil {
		merged.Logging = child.Logging
	} else if parent.Logging != nil {
		merged.Logging = parent.Logging
	}

	if child.Arguments != nil || parent.Arguments != nil {
		var mergedArgs downloader.Arguments
		if parent.Arguments != nil {
			mergedArgs.JVM = append([]interface{}{}, parent.Arguments.JVM...)
			mergedArgs.Game = append([]interface{}{}, parent.Arguments.Game...)
		}
		if child.Arguments != nil {
			mergedArgs.JVM = append(mergedArgs.JVM, child.Arguments.JVM...)
			mergedArgs.Game = append(mergedArgs.Game, child.Arguments.Game...)
		}
		merged.Arguments = &mergedArgs
	}

	if child.MinecraftArguments != "" {
		merged.MinecraftArguments = child.MinecraftArguments
	} else if parent.MinecraftArguments != "" {
		merged.MinecraftArguments = parent.MinecraftArguments
	}

	return &merged
}

func mergeLibraries(parent, child []downloader.Library) []downloader.Library {
	type entry struct {
		lib  downloader.Library
		from string
	}
	keyOf := func(lib downloader.Library) string {
		parts := strings.Split(lib.Name, ":")
		if len(parts) >= 4 {
			return parts[0] + ":" + parts[1] + ":" + parts[3]
		}
		if len(parts) >= 2 {
			return parts[0] + ":" + parts[1]
		}
		return lib.Name
	}
	idx := make(map[string]int, len(parent)+len(child))
	out := make([]entry, 0, len(parent)+len(child))
	add := func(lib downloader.Library, from string) {
		k := keyOf(lib)
		if i, ok := idx[k]; ok {
			if from == "child" && out[i].from == "parent" {
				out[i] = entry{lib: lib, from: from}
			}
			return
		}
		idx[k] = len(out)
		out = append(out, entry{lib: lib, from: from})
	}
	for _, lib := range parent {
		add(lib, "parent")
	}
	for _, lib := range child {
		add(lib, "child")
	}
	merged := make([]downloader.Library, 0, len(out))
	for _, e := range out {
		merged = append(merged, e.lib)
	}
	return merged
}

func (l *Launcher) ensureGameDir() {
	adv := l.adv()
	os.MkdirAll(adv.GameDir, 0755)
}

func (l *Launcher) ensureLauncherProperties(gameDir string) string {
	if gameDir == "" {
		return ""
	}
	if err := os.MkdirAll(gameDir, 0755); err != nil {
		return ""
	}
	path := filepath.Join(gameDir, "launcher.properties")
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	content := fmt.Sprintf("fml.client.secret=%s\n", hex.EncodeToString(buf))
	written, err := globalutils.SafeWriteFile(path, []byte(content), 0644)
	if err != nil {
		l.log("WARN: no se pudo crear %s: %v", path, err)
		return ""
	}
	if !written {
		l.log("launcher.properties ya existe, no se sobrescribe: %s", path)
	}
	return path
}

func (l *Launcher) scanLogForCrashPatterns(logPath string) (category, reason string) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return "", ""
	}
	content := string(data)
	for _, cp := range crashPatterns {
		if strings.Contains(content, cp.pattern) {
			return cp.category, cp.reason
		}
	}
	return "", ""
}

func (l *Launcher) waitForExit(instance *GameInstance, mcLogFile *os.File) {
	adv := l.adv()

	defer func() {
		if mcLogFile != nil {
			mcLogFile.Close()
		}
		l.gameLog.Close()
	}()

	err := instance.cmd.Wait()
	instance.mu.Lock()

	stoppedByUser := instance.Status == GameStopped

	cleanMarker := gamelog.HasCleanShutdownMarker(instance.LogPath)
	exitCode := 0

	if err != nil {
		if exitErr, ok := err.(interface{ ExitCode() int }); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	if exitCode == 0 || cleanMarker || stoppedByUser {
		if exitCode != 0 && cleanMarker {
			l.log("Game exited with code %d but clean shutdown detected via log marker", exitCode)
		}
		instance.ExitCode = exitCode
		instance.Status = GameExited
		if stoppedByUser {
			instance.Status = GameStopped
			l.log("Game stopped by user (exit code %d)", exitCode)
		}
		instance.CrashReason = helpers.CrashReasonLabel(0)
		instance.CrashCategory = helpers.CrashCategory(0, "")
		l.log("Game exited cleanly")
		writeExit := exitCode
		if stoppedByUser {
			writeExit = 0
		}
		l.gameLog.WriteGameExit(writeExit)
		instance.mu.Unlock()
		if stoppedByUser {
			BroadcastStopped(l.eventBroadcast, instance)
		} else {
			BroadcastExited(l.eventBroadcast, instance)
		}
	} else {
		instance.ExitCode = exitCode
		instance.Status = GameCrashed
		l.log("Game crashed with exit code %d", exitCode)

		crashCategory, crashReason := l.scanLogForCrashPatterns(instance.LogPath)
		if crashCategory == "" {
			crashReason = helpers.CrashReasonLabel(exitCode)
			crashCategory = helpers.CrashCategory(exitCode, crashReason)
		}
		instance.CrashReason = crashReason
		instance.CrashCategory = crashCategory

		l.recordCrash(instance)
		l.gameLog.WriteGameExit(exitCode)
		instance.mu.Unlock()
		BroadcastCrashed(l.eventBroadcast, instance)
	}

	close(instance.done)

	if l.onGameExit != nil {
		playTime := int(time.Since(instance.StartTime).Seconds())
		l.onGameExit(instance, playTime)
	}

	if instance.ExitCode == 0 {
		l.scheduleCleanup()
	}

	if adv.PostLaunchCommand != "" {
		l.log("Executing post-launch command: %s", adv.PostLaunchCommand)
		if err := l.execHook(adv.PostLaunchCommand, adv.GameDir); err != nil {
			l.log("WARN: post-launch command failed: %v", err)
		}
	}
}

func (l *Launcher) recordCrash(instance *GameInstance) {
	gameDir := l.adv().GameDir
	buffer := l.gameLog.GetMemoryBuffer()
	contextLines := buffer
	if len(contextLines) > 25 {
		contextLines = contextLines[len(contextLines)-25:]
	}

	crashFile := utils.FindLatestCrashReport(gameDir)
	if crashFile == "" {
		crashFile = utils.FindJVMCrashLog(gameDir)
	}
	if crashFile != "" {
		instance.CrashLog = crashFile
		l.log("Crash report: %s", crashFile)
	}

instance.CrashLogContent = readCrashLogText(crashFile, contextLines)
	instance.GameOutput = gamelog.ReadGameOutput(instance.LogPath)

	if instance.CrashReason == "" {
		instance.CrashReason = helpers.CrashReasonLabel(instance.ExitCode)
	}
	if instance.CrashCategory == "" {
		instance.CrashCategory = helpers.CrashCategory(instance.ExitCode, instance.CrashReason)
	}

	l.log("Crash category: %s, reason: %s", instance.CrashCategory, instance.CrashReason)
	l.log("Crash context (%d lines from log buffer)", len(contextLines))
	for _, line := range contextLines {
		l.log("  %s", line)
	}

	if l.cfg.LogFn != nil {
		l.cfg.LogFn("[Crash] Game %s crashed: %s/%s (exit %d)", instance.ID, instance.CrashCategory, instance.CrashReason, instance.ExitCode)
	}
}

func readCrashLogText(crashFile string, contextLines []string) string {
	const maxBytes = 32 * 1024
	const maxLines = 400

	if crashFile != "" {
		data, err := os.ReadFile(crashFile)
		if err == nil && len(data) > 0 {
			if len(data) > maxBytes {
				data = data[len(data)-maxBytes:]
			}
			text := string(data)
			lines := strings.Split(text, "\n")
			if len(lines) > maxLines {
				lines = lines[len(lines)-maxLines:]
			}
			if joined := strings.Join(lines, "\n"); strings.TrimSpace(joined) != "" {
				return joined
			}
		}
	}

	if len(contextLines) == 0 {
		return ""
	}
	lines := contextLines
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return strings.Join(lines, "\n")
}

func (l *Launcher) scheduleCleanup() {
	adv := l.adv()
	delay := adv.CleanupDelay
	if delay <= 0 {
		delay = 5 * time.Minute
	}

	nativesDir := adv.NativesDir
	if nativesDir == "" {
		baseVer := adv.BaseVersion
		if baseVer == "" {
			baseVer = l.cfg.Version
		}
		nativesDir = helpers.NativesDir(adv.NativesBaseDir, baseVer)
	}

	logFn := func(f string, a ...interface{}) {
		if l.gameLog != nil {
			l.gameLog.Log("launcher", fmt.Sprintf(f, a...))
		}
		if l.cfg.LogFn != nil {
			l.cfg.LogFn("[GameLauncher] "+f, a...)
		}
	}

	time.AfterFunc(delay, func() {
		if err := os.RemoveAll(nativesDir); err != nil {
			logFn("Cleanup: failed to remove %s: %v", nativesDir, err)
		} else {
			logFn("Cleanup: removed natives %s", nativesDir)
		}
		logFn("Engine idle - waiting for launch requests")
	})
}

func (l *Launcher) execHook(command, gameDir string) error {
	cmd := exec.Command("cmd", "/C", command)
	if runtime.GOOS != "windows" {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Dir = gameDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (l *Launcher) log(format string, args ...interface{}) {
	if l.gameLog != nil {
		l.gameLog.Log("launcher", fmt.Sprintf(format, args...))
	}
	if l.cfg.LogFn != nil {
		l.cfg.LogFn("[GameLauncher] "+format, args...)
	}
}

func (l *Launcher) prepareAuthInjector(cfg AuthLibConfig) error {
	if cfg.InjectorPath != "" {
		if st, err := os.Stat(cfg.InjectorPath); err != nil || st.Size() == 0 {
			return fmt.Errorf("el jar de authlib-injector configurado no existe: %s", cfg.InjectorPath)
		}
		l.resolvedInjectorPath = cfg.InjectorPath
	} else {
		adv := l.adv()
		dir := adv.CacheDir
		if dir == "" {
			dir = filepath.Join(filepath.Dir(adv.VersionsDir), "authlib-injector")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		jarPath, err := authlib.EnsureInjector(ctx, filepath.Join(dir, authlib.InjectorFileName), false)
		if err != nil {
			return err
		}
		l.resolvedInjectorPath = jarPath
		l.log("Authlib-injector listo: %s", jarPath)
	}

	if err := l.prefetchAuthMeta(cfg); err != nil {
		l.log("WARN: no se pudo precargar la metadata del auth server: %v", err)
	}
	return nil
}

func (l *Launcher) prefetchAuthMeta(cfg AuthLibConfig) error {
	if cfg.AuthServerURL == "" {
		return fmt.Errorf("auth server URL is empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	root, err := authlib.ResolveServerURL(ctx, cfg.AuthServerURL)
	if err != nil {
		return err
	}
	meta, err := authlib.FetchMetadata(ctx, root)
	if err != nil {
		return err
	}
	l.resolvedAuthRoot = root
	l.prefetchedMeta = base64.StdEncoding.EncodeToString(meta)
	l.log("Auth server metadata precargada: %s (%d bytes)", root, len(meta))
	return nil
}

func (l *Launcher) preVerifyAuthServer(cfg AuthLibConfig) error {
	if cfg.AuthServerURL == "" {
		return fmt.Errorf("auth server URL is empty")
	}
	timeout := cfg.PreVerifyTimeout
	if timeout <= 0 {
		timeout = 10
	}
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Get(cfg.AuthServerURL)
	if err != nil {
		return fmt.Errorf("auth server unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth server returned status %d", resp.StatusCode)
	}
	return nil
}

func ensureClientJarCopy(versionsDir, baseVer, launchVer string) error {
	if launchVer == "" || launchVer == baseVer {
		return nil
	}
	src := filepath.Join(versionsDir, baseVer, baseVer+".jar")
	dst := filepath.Join(versionsDir, launchVer, launchVer+".jar")
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("client jar base no encontrado: %w", err)
	}
	if dstInfo, err := os.Stat(dst); err == nil && dstInfo.Size() == srcInfo.Size() {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
