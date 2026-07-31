package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Type string

const (
	INFO   Type = "INFO"
	WARN   Type = "WARN"
	ERROR  Type = "ERROR"
	DEBUG  Type = "DEBUG"
	FATAL  Type = "FATAL"
	SYSTEM Type = "SYSTEM"
)

type Logger struct {
	mu              sync.Mutex
	file            *os.File
	module          string
	logDir          string
	date            string
	launcherName    string
	launcherVersion string
	keepDays        int
	maxFiles        int
	broadcastFn     func(Type, string)
}

func New(logDir, module, launcherName, launcherVersion string) (*Logger, error) {
	l := &Logger{
		logDir:          logDir,
		module:          module,
		launcherName:    launcherName,
		launcherVersion: launcherVersion,
		date:            time.Now().Format("2006-01-02"),
	}

	if err := l.rotate(); err != nil {
		return nil, err
	}

	l.cleanupOldLogs()

	return l, nil
}

func (l *Logger) SetLauncher(name, version string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.launcherName = name
	l.launcherVersion = version
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
	l.date = ""
	l.rotateLocked()
}

func (l *Logger) SetRetention(keepDays, maxFiles int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.keepDays = keepDays
	l.maxFiles = maxFiles
}

func (l *Logger) cleanupOldLogs() {
	if l.logDir == "" {
		return
	}

	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		return
	}

	name := l.launcherName
	if name == "" {
		name = "StepLauncher"
	}
	ver := l.launcherVersion
	if ver == "" {
		ver = "0.1.0"
	}

	prefix := name + "-" + ver + "-"

	type logFile struct {
		path string
		mod  time.Time
	}

	var logFiles []logFile
	cutoff := time.Now().AddDate(0, 0, -l.keepDays)

	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), prefix) || !strings.HasSuffix(e.Name(), ".log") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		lf := logFile{path: filepath.Join(l.logDir, e.Name()), mod: info.ModTime()}

		if l.keepDays > 0 && info.ModTime().Before(cutoff) {
			os.Remove(lf.path)
			continue
		}
		logFiles = append(logFiles, lf)
	}

	if l.maxFiles > 0 && len(logFiles) > l.maxFiles {
		sort.Slice(logFiles, func(i, j int) bool {
			return logFiles[i].mod.After(logFiles[j].mod)
		})
		for _, lf := range logFiles[l.maxFiles:] {
			os.Remove(lf.path)
		}
	}
}

func (l *Logger) rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.rotateLocked()
}

func (l *Logger) rotateLocked() error {
	today := time.Now().Format("2006-01-02")
	if today == l.date && l.file != nil {
		return nil
	}

	if l.file != nil {
		l.file.Close()
	}

	l.date = today

	l.cleanupOldLogs()

	name := l.launcherName
	if name == "" {
		name = "StepLauncher"
	}
	ver := l.launcherVersion
	if ver == "" {
		ver = "0.1.0"
	}

	filename := fmt.Sprintf("%s-%s-%s.log", name, ver, today)
	path := filepath.Join(l.logDir, filename)

	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	l.file = f
	return nil
}

func (l *Logger) Log(t Type, format string, args ...interface{}) {
	l.mu.Lock()

	l.rotateLocked()

	msg := fmt.Sprintf(format, args...)
	ts := time.Now().Format("2006-01-02 15:04:05")
	plain := fmt.Sprintf("[ %s ] [ %s ] [ %s ] - %s\n", ts, string(t), l.module, msg)

	if l.file != nil {
		fmt.Fprint(l.file, plain)
	}

	if t == FATAL {
		l.mu.Unlock()
		os.Exit(1)
	}

	broadcastFn := l.broadcastFn
	l.mu.Unlock()

	if broadcastFn != nil {
		broadcastFn(t, msg)
	}
}

func (l *Logger) Info(f string, a ...interface{})   { l.Log(INFO, f, a...) }
func (l *Logger) Warn(f string, a ...interface{})   { l.Log(WARN, f, a...) }
func (l *Logger) Error(f string, a ...interface{})  { l.Log(ERROR, f, a...) }
func (l *Logger) Debug(f string, a ...interface{})  { l.Log(DEBUG, f, a...) }
func (l *Logger) Fatal(f string, a ...interface{})  { l.Log(FATAL, f, a...) }
func (l *Logger) System(f string, a ...interface{}) { l.Log(SYSTEM, f, a...) }

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
