package gamelog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultLimit = 1500
	maxLimit     = 10000
	minLimit     = 100
	defaultKeep  = 7
)

type GameLogConfig struct {
	LogDir       string
	LauncherName string
	Version      string
	Limit        int
	KeepDays     int
	MaxFiles     int
}

var mcLogPattern = regexp.MustCompile(`^\[(\d{2}:\d{2}:\d{2})\]\s+\[([^/]+)/(\w+)\](?:\s+\([^)]+\))?:\s+(.*)$`)

type ParsedLogLine struct {
	Time    string
	Thread  string
	Level   string
	Logger  string
	Message string
}

func ParseLine(line string) ParsedLogLine {
	if line == "" {
		return ParsedLogLine{Level: "RAW", Message: line}
	}
	m := mcLogPattern.FindStringSubmatch(line)
	if m != nil {
		return ParsedLogLine{
			Time:    m[1],
			Thread:  m[2],
			Level:   strings.ToUpper(m[3]),
			Logger:  m[2],
			Message: m[4],
		}
	}
	return ParsedLogLine{Level: "RAW", Message: line}
}

type ClasspathEntry struct {
	Path   string
	Exists bool
}

type NativeEntry struct {
	Path string
}

type PreLaunchInfo struct {
	Version, VanillaVersionID, MainClass, JavaExec string
	MinRAM, MaxRAM                                 int
	GCPreset, GPUPreference                        string
	HWAccelDisabled                                bool
	GameDir, AssetsDir, LibrariesDir, NativesDir   string
	AssetIndexID                                   string
	AssetIndexVirtual                              string
	LauncherName, LauncherVersion                  string
	ClasspathEntries                               []ClasspathEntry
	Natives                                        []NativeEntry
	JVMArgs, GameArgs                              []string
}

type GameLogManager struct {
	mu           sync.Mutex
	file         *os.File
	logDir       string
	date         string
	launcherName string
	version      string

	buffer  chan string
	limit   int
	dropped int
	closed  bool

	keepDays    int
	maxFiles    int
	broadcastFn func(stream, line string)
}

func (g *GameLogManager) SetBroadcastFn(fn func(stream, line string)) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.broadcastFn = fn
}

func NewGameLogManager(cfg GameLogConfig) (*GameLogManager, error) {
	limit := cfg.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit < minLimit {
		limit = minLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	keepDays := cfg.KeepDays
	maxFiles := cfg.MaxFiles

	l := &GameLogManager{
		logDir:       cfg.LogDir,
		launcherName: cfg.LauncherName,
		version:      cfg.Version,
		limit:        limit,
		keepDays:     keepDays,
		maxFiles:     maxFiles,
		buffer:       make(chan string, limit),
	}

	l.cleanupOldLogs()

	if err := l.rotate(); err != nil {
		return nil, err
	}

	return l, nil
}

func (g *GameLogManager) cleanupOldLogs() {
	if g.logDir == "" {
		return
	}

	entries, err := os.ReadDir(g.logDir)
	if err != nil {
		return
	}

	name := g.launcherName
	if name == "" {
		name = "StepLauncher"
	}

	prefix := name + "-" + g.version + "-"

	type logFile struct {
		path string
		mod  time.Time
	}

	var logs []logFile
	cutoff := time.Now().AddDate(0, 0, -g.keepDays)

	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), prefix) || !strings.HasSuffix(e.Name(), ".log") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		l := logFile{path: filepath.Join(g.logDir, e.Name()), mod: info.ModTime()}

		if g.keepDays > 0 && info.ModTime().Before(cutoff) {
			os.Remove(l.path)
			continue
		}
		logs = append(logs, l)
	}

	if g.maxFiles > 0 && len(logs) > g.maxFiles {
		sort.Slice(logs, func(i, j int) bool {
			return logs[i].mod.After(logs[j].mod)
		})
		for _, l := range logs[g.maxFiles:] {
			os.Remove(l.path)
		}
	}
}

func (g *GameLogManager) rotate() error {
	today := time.Now().Format("2006-01-02")
	if today == g.date && g.file != nil {
		return nil
	}

	if g.file != nil {
		g.file.Close()
	}

	g.date = today

	name := g.launcherName
	if name == "" {
		name = "StepLauncher"
	}

	filename := fmt.Sprintf("%s-%s-%s.log", name, g.version, today)
	path := filepath.Join(g.logDir, filename)

	if err := os.MkdirAll(g.logDir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	g.file = f
	return nil
}

func (g *GameLogManager) WritePreLaunchInfo(info PreLaunchInfo) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.closed || g.file == nil {
		return
	}

	border := strings.Repeat("=", 70)
	dash := strings.Repeat("-", 70)

	g.writeLine(border)
	g.writeLine("  StepLauncher Game Log")
	g.writeLine("  Started    : " + time.Now().Format("2006-01-02 15:04:05"))
	g.writeLine(border)
	g.writeLine("")

	g.writeLine(dash)
	g.writeLine("  SYSTEM")
	g.writeLine(dash)
	g.writeLine(fmt.Sprintf("  OS      : %s / %s", runtime.GOOS, runtime.GOARCH))
	g.writeLine(fmt.Sprintf("  CPU     : %d cores", runtime.NumCPU()))
	g.writeLine(fmt.Sprintf("  RAM     : %d MB (%d min / %d max)", info.MaxRAM, info.MinRAM, info.MaxRAM))
	g.writeLine(fmt.Sprintf("  Go      : %s", runtime.Version()))
	g.writeLine("")

	g.writeLine(dash)
	g.writeLine("  MINECRAFT CONFIGURATION")
	g.writeLine(dash)
	g.writeLine(fmt.Sprintf("  Version        : %s", info.Version))
	g.writeLine(fmt.Sprintf("  Main Class     : %s", info.MainClass))
	g.writeLine(fmt.Sprintf("  Game Dir       : %s", info.GameDir))
	g.writeLine(fmt.Sprintf("  Assets Dir     : %s", info.AssetsDir))
	g.writeLine(fmt.Sprintf("  Libraries Dir  : %s", info.LibrariesDir))
	g.writeLine(fmt.Sprintf("  Natives Dir    : %s", info.NativesDir))
	g.writeLine(fmt.Sprintf("  Asset Index    : %s", info.AssetIndexID))
	g.writeLine(fmt.Sprintf("  Virtual Assets : %s", orStr(info.AssetIndexVirtual, "unset")))
	g.writeLine("")

	g.writeLine(dash)
	g.writeLine("  JVM CONFIGURATION")
	g.writeLine(dash)
	g.writeLine(fmt.Sprintf("  Java Exec  : %s", info.JavaExec))
	g.writeLine(fmt.Sprintf("  Min Memory : %d MB", info.MinRAM))
	g.writeLine(fmt.Sprintf("  Max Memory : %d MB", info.MaxRAM))
	if info.GCPreset != "" {
		g.writeLine(fmt.Sprintf("  GC Preset  : %s", info.GCPreset))
	}
	if info.GPUPreference != "" {
		g.writeLine(fmt.Sprintf("  GPU Pref   : %s", info.GPUPreference))
	}
	g.writeLine(fmt.Sprintf("  HW Accel   : %v", boolStr(!info.HWAccelDisabled)))
	g.writeLine("")

	g.writeLine(dash)
	g.writeLine("  CLASSPATH LIBRARIES")
	g.writeLine(dash)
	g.writeLine(fmt.Sprintf("  Total entries: %d", len(info.ClasspathEntries)))
	for _, e := range info.ClasspathEntries {
		mark := "OK"
		if !e.Exists {
			mark = "MISSING"
		}
		g.writeLine(fmt.Sprintf("    [%s] %s", mark, e.Path))
	}
	g.writeLine("")

	g.writeLine(dash)
	g.writeLine("  NATIVE LIBRARIES")
	g.writeLine(dash)
	if len(info.Natives) > 0 {
		for _, n := range info.Natives {
			g.writeLine(fmt.Sprintf("    %s", n.Path))
		}
	} else {
		g.writeLine("    (none)")
	}
	g.writeLine("")

	g.writeLine(dash)
	g.writeLine(fmt.Sprintf("  FULL LAUNCH COMMAND (%d args)", len(info.JVMArgs)+len(info.GameArgs)+2))
	g.writeLine(dash)
	allArgs := append([]string{info.JavaExec}, info.JVMArgs...)
	allArgs = append(allArgs, info.MainClass)
	allArgs = append(allArgs, info.GameArgs...)
	prevSensitive := false
	for i, a := range allArgs {
		if prevSensitive {
			g.writeLine(fmt.Sprintf("  [%3d] [REDACTED]", i))
			prevSensitive = false
			continue
		}
		if isSensitiveFlag(a) {
			prevSensitive = true
		}
		g.writeLine(fmt.Sprintf("  [%3d] %s", i, redactSensitive(a)))
	}
	g.writeLine("")

	g.writeLine(border)
	g.writeLine("  GAME OUTPUT FOLLOWS")
	g.writeLine(border)
	g.writeLine("")
	g.flush()
}

func (g *GameLogManager) WriteGameExit(exitCode int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.closed || g.file == nil {
		return
	}

	g.writeLine("")
	border := strings.Repeat("=", 50)
	g.writeLine(border)
	if exitCode == 0 {
		g.writeLine("  GAME EXIT: clean (0)")
	} else {
		g.writeLine(fmt.Sprintf("  GAME CRASH: exit code %d", exitCode))
	}
	g.writeLine("  " + time.Now().Format("2006-01-02 15:04:05"))
	g.writeLine(border)
	g.flush()
}

func (g *GameLogManager) Log(stream, line string) {
	g.mu.Lock()
	if g.closed || g.file == nil {
		g.mu.Unlock()
		return
	}

	ts := time.Now().Format("15:04:05.000")
	entry := fmt.Sprintf("[%s] [%s] %s", ts, strings.ToUpper(stream), line)

	fmt.Fprintln(g.file, entry)

	select {
	case g.buffer <- entry:
	default:
		select {
		case <-g.buffer:
			g.dropped++
		default:
		}
		select {
		case g.buffer <- entry:
		default:
			g.dropped++
		}
	}

	broadcastFn := g.broadcastFn
	g.mu.Unlock()

	if broadcastFn != nil {
		broadcastFn(stream, line)
	}
}

func (g *GameLogManager) GetMemoryBuffer() []string {
	g.mu.Lock()
	defer g.mu.Unlock()

	var result []string
	for {
		select {
		case line := <-g.buffer:
			result = append(result, line)
		default:
			for _, line := range result {
				select {
				case g.buffer <- line:
				default:
				}
			}
			return result
		}
	}
}

func (g *GameLogManager) GetDroppedCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.dropped
}

func (g *GameLogManager) GetLogPath() string {
	name := g.launcherName
	if name == "" {
		name = "StepLauncher"
	}
	return filepath.Join(g.logDir,
		fmt.Sprintf("%s-%s-%s.log", name, g.version, g.date))
}

func (g *GameLogManager) Close() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.closed {
		return
	}
	g.closed = true

	if g.file != nil {
		g.writeLine("")
		g.writeLine(strings.Repeat("-", 70))
		if g.dropped > 0 {
			g.writeLine(fmt.Sprintf("  WARNING: %d log line(s) dropped (limit=%d)", g.dropped, g.limit))
		}
		g.writeLine(fmt.Sprintf("  Log closed: %s", time.Now().Format("2006-01-02 15:04:05")))
		g.writeLine(strings.Repeat("-", 70))
		g.flush()
		g.file.Close()
		g.file = nil
	}
}

func (g *GameLogManager) writeLine(line string) {
	if g.file != nil {
		fmt.Fprintln(g.file, line)
	}
}

func (g *GameLogManager) flush() {
	if g.file != nil {
		g.file.Sync()
	}
}

var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(--?accessToken[= ]+)[^ ]+`),
	regexp.MustCompile(`(--?session[= ]+)[^ ]+`),
	regexp.MustCompile(`(--?user(?:name|properties)?[= ]+)[^ ]+`),
	regexp.MustCompile(`(--?password[= ]+)[^ ]+`),
	regexp.MustCompile(`(--?token[= ]+)[^ ]+`),
	regexp.MustCompile(`(Authorization:?\s*)\S+`),
}

var sensitiveFlags = []string{
	"--accessToken", "-accessToken",
	"--session", "-session",
	"--password", "-password",
	"--token", "-token",
}

func isSensitiveFlag(arg string) bool {
	for _, f := range sensitiveFlags {
		if arg == f {
			return true
		}
	}
	return false
}

func redactSensitive(line string) string {
	if line == "" {
		return ""
	}
	result := line
	for _, p := range sensitivePatterns {
		result = p.ReplaceAllString(result, "${1}[REDACTED]")
	}
	return result
}

func HasCleanShutdownMarker(logPath string) bool {
	f, err := os.Open(logPath)
	if err != nil {
		return false
	}
	defer f.Close()

	const maxScan = 128 * 1024
	info, err := f.Stat()
	if err != nil {
		return false
	}
	size := info.Size()
	offset := int64(0)
	if size > maxScan {
		offset = size - maxScan
	}
	buf := make([]byte, size-offset)
	if _, err := f.ReadAt(buf, offset); err != nil {
		return false
	}
	return strings.Contains(string(buf), "Stopping!")
}

func orStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func boolStr(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}
