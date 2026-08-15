package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"StepLauncher/internal/Core/Downloader"
)

const (
	IntegrityScopeTodo      = "todo"
	IntegrityScopeGlobal    = "global"
	IntegrityScopeInstances = "instances"
)

type IntegrityState string

const (
	IntegrityStateIdle      IntegrityState = "idle"
	IntegrityStateRunning   IntegrityState = "running"
	IntegrityStateCompleted IntegrityState = "completed"
	IntegrityStateCancelled IntegrityState = "cancelled"
	IntegrityStateError     IntegrityState = "error"
)

const (
	integrityPhaseIndexing = "indexing"
	integrityPhaseExists   = "existence"
	integrityPhaseRetry    = "retry"
	integrityPhaseVerify   = "verify"
	integrityPhaseDone     = "done"
)

const (
	integrityAttemptsPass1 = 3
	integrityAttemptsPass2 = 5
	integrityStallTimeout  = 60000
	integrityMaxStall      = 3
	integrityMaxSkipped    = 200
)

type IntegritySkipped struct {
	File   string `json:"file"`
	Reason string `json:"reason"`
}

type IntegrityProgress struct {
	State          IntegrityState     `json:"state"`
	Phase          string             `json:"phase"`
	Scope          string             `json:"scope"`
	CurrentVersion string             `json:"currentVersion"`
	CurrentFile    string             `json:"currentFile"`
	TasksTotal     int                `json:"tasksTotal"`
	TasksDone      int                `json:"tasksDone"`
	FilesMissing   int                `json:"filesMissing"`
	FilesRestored  int                `json:"filesRestored"`
	FilesCorrupt   int                `json:"filesCorrupt"`
	FilesSkipped   int                `json:"filesSkipped"`
	VersionsScanned int               `json:"versionsScanned"`
	Percent        int                `json:"percent"`
	StartedAt      string             `json:"startedAt,omitempty"`
	FinishedAt     string             `json:"finishedAt,omitempty"`
	Skipped        []IntegritySkipped `json:"skipped,omitempty"`
}

type integrityTask struct {
	Task    downloader.DownloadTask
	Version string
	Origin  string
}

type integrityRunner struct {
	mu     sync.Mutex
	active bool
	cancel context.CancelFunc
	prog   IntegrityProgress

	phaseBase  int
	phaseSpan  int
	phaseDone  int
	phaseTotal int
}

func (e *Engine) StartIntegrityCheck(scope string) error {
	switch scope {
	case IntegrityScopeTodo, IntegrityScopeGlobal, IntegrityScopeInstances:
	default:
		return fmt.Errorf("sector de integridad invalido: %s", scope)
	}
	ir := e.integrity
	if ir == nil {
		return fmt.Errorf("integridad no disponible")
	}
	ir.mu.Lock()
	if ir.active {
		ir.mu.Unlock()
		return fmt.Errorf("ya hay una verificacion de integridad en curso")
	}
	ir.active = true
	ir.prog = IntegrityProgress{
		State:     IntegrityStateRunning,
		Phase:     integrityPhaseIndexing,
		Scope:     scope,
		StartedAt: time.Now().Format(time.RFC3339),
	}
	ir.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	ir.mu.Lock()
	ir.cancel = cancel
	ir.mu.Unlock()

	e.log.Info("[Integrity] Verificacion iniciada (sector=%s)", scope)
	go e.runIntegrity(ctx, scope)
	return nil
}

func (e *Engine) CancelIntegrityCheck() {
	ir := e.integrity
	if ir == nil {
		return
	}
	ir.mu.Lock()
	cancel := ir.cancel
	active := ir.active
	ir.mu.Unlock()
	if cancel != nil && active {
		e.log.Warn("[Integrity] Verificacion cancelada por el usuario")
		cancel()
	}
}

func (e *Engine) IntegrityStatus() IntegrityProgress {
	ir := e.integrity
	if ir == nil {
		return IntegrityProgress{State: IntegrityStateIdle}
	}
	ir.mu.Lock()
	defer ir.mu.Unlock()
	return ir.prog
}

func (e *Engine) runIntegrity(ctx context.Context, scope string) {
	ir := e.integrity
	defer func() {
		ir.mu.Lock()
		ir.active = false
		ir.cancel = nil
		if ir.prog.State == IntegrityStateRunning {
			if ctx.Err() != nil {
				ir.prog.State = IntegrityStateCancelled
			} else {
				ir.prog.State = IntegrityStateCompleted
			}
		}
		ir.prog.Phase = integrityPhaseDone
		ir.prog.FinishedAt = time.Now().Format(time.RFC3339)
		ir.prog.Percent = 100
		final := ir.prog
		ir.mu.Unlock()
		e.log.Info("[Integrity] Verificacion terminada: estado=%s | %d archivos (%d pendientes, %d corruptos, %d descartados, %d reparados) en %d versiones (%s)",
			final.State, final.TasksTotal, final.FilesMissing, final.FilesCorrupt, final.FilesSkipped, final.FilesRestored,
			final.VersionsScanned, final.Scope)
	}()

	cfg := e.config.Get()
	globalDir := cfg.WorkDir
	instancesDir := filepath.Join(cfg.WorkDir, cfg.InstancesDir)
	client := e.downloader.HTTPClient()

	tasks, versions, err := e.indexIntegrity(ctx, scope, globalDir, instancesDir, client)
	if err != nil {
		ir.mu.Lock()
		ir.prog.State = IntegrityStateError
		ir.mu.Unlock()
		e.log.Error("[Integrity] ERROR: %v", err)
		return
	}
	if len(tasks) == 0 {
		e.log.Info("[Integrity] No hay JSON de versiones que verificar (sector=%s)", scope)
	}

	ir.mu.Lock()
	ir.prog.TasksTotal = len(tasks)
	ir.prog.VersionsScanned = versions
	ir.mu.Unlock()

	ir.setPhase(integrityPhaseExists, len(tasks), 10, 45)
	pending := e.integrityPassExistence(ctx, tasks, client, ir)

	retryCount := 0
	for _, t := range pending {
		if !downloader.FileExists(t.Task.Dest) {
			retryCount++
		}
	}
	ir.setPhase(integrityPhaseRetry, retryCount, 55, 20)
	e.integrityPassRetry(ctx, pending, client, ir)

	if ctx.Err() != nil {
		return
	}

	ir.setPhase(integrityPhaseVerify, len(tasks), 75, 25)
	e.integrityPassVerify(ctx, tasks, client, ir)
}

// indexIntegrity recorre TODOS los JSON de versiones descargadas (globales e
// instancias, segun el sector) y construye la lista de archivos esperados,
// respetando el tipo de cada version (estructura de carpetas global/instancia).
func (e *Engine) indexIntegrity(ctx context.Context, scope, globalDir, instancesDir string, client *http.Client) ([]integrityTask, int, error) {
	ir := e.integrity
	var out []integrityTask
	seen := make(map[string]bool)
	versions := 0
	addTasks := func(list []integrityTask) {
		for _, t := range list {
			if seen[t.Task.Dest] {
				continue
			}
			seen[t.Task.Dest] = true
			out = append(out, t)
		}
	}

	if scope != IntegrityScopeInstances {
		verRoot := filepath.Join(globalDir, "versions")
		entries, err := os.ReadDir(verRoot)
		if err != nil {
			e.log.Info("[Integrity] Sector global: sin carpeta %s (%v)", verRoot, err)
		} else {
			for _, entry := range entries {
				if ctx.Err() != nil {
					return nil, versions, nil
				}
				if !entry.IsDir() {
					continue
				}
				id := entry.Name()
				ver, err := e.loadVersionJSON(filepath.Join(verRoot, id, id+".json"))
				if err != nil {
					e.log.Warn("[Integrity] version global %s: %v", id, err)
					continue
				}
				filter := downloader.DownloadFilter{
					Version:   id,
					Client:    true,
					Libraries: true,
					Natives:   true,
					Assets:    true,
					Java:      true,
				}
				dlCfg := downloader.Config{
					WorkDir:      globalDir,
					CacheDir:     filepath.Join(globalDir, "cache"),
					CacheManager: e.cache,
					HTTPClient:   client,
				}
				tasks, err := downloader.BuildTasks(dlCfg, ver, id, filter)
				if err != nil {
					e.log.Warn("[Integrity] version global %s: build tasks: %v", id, err)
					continue
				}
				ir.mu.Lock()
				ir.prog.CurrentVersion = id
				ir.prog.VersionsScanned++
				ir.mu.Unlock()
				wrapped := make([]integrityTask, 0, len(tasks))
				for _, t := range tasks {
					wrapped = append(wrapped, integrityTask{Task: t, Version: id, Origin: "global"})
				}
				addTasks(wrapped)
				versions++
				e.log.Info("[Integrity] JSON global %s: %d archivos esperados", id, len(tasks))
			}
		}
	}

	instVersions := 0
	if scope != IntegrityScopeGlobal {
		instEntries, err := os.ReadDir(instancesDir)
		if err != nil {
			e.log.Info("[Integrity] Sector instancias: sin carpeta %s", instancesDir)
		} else {
			for _, instEntry := range instEntries {
				if ctx.Err() != nil {
					return nil, versions, nil
				}
				if !instEntry.IsDir() {
					continue
				}
				instName := instEntry.Name()
				instVersRoot := filepath.Join(instancesDir, instName, "versions")
				verEntries, verErr := os.ReadDir(instVersRoot)
				if verErr != nil {
					continue
				}
				for _, verEntry := range verEntries {
					if ctx.Err() != nil {
						return nil, versions, nil
					}
					if !verEntry.IsDir() {
						continue
					}
					id := verEntry.Name()
					ver, err := e.loadVersionJSON(filepath.Join(instVersRoot, id, id+".json"))
					if err != nil {
						e.log.Warn("[Integrity] version de instancia %s (%s): %v", instName, id, err)
						continue
					}
					instVerDir := filepath.Join(instVersRoot, id)
					filter := downloader.DownloadFilter{
						Version:            id,
						Client:             true,
						Libraries:          true,
						Natives:            true,
						Assets:             true,
						Java:               true,
						InstanceVersionDir: instVerDir,
					}
					dlCfg := downloader.Config{
						WorkDir:      filepath.Join(globalDir, "shared"),
						CacheDir:     filepath.Join(globalDir, "cache"),
						CacheManager: e.cache,
						HTTPClient:   client,
					}
					tasks, err := downloader.BuildTasks(dlCfg, ver, id, filter)
					if err != nil {
						e.log.Warn("[Integrity] version de instancia %s (%s): build tasks: %v", instName, id, err)
						continue
					}
					ir.mu.Lock()
					ir.prog.CurrentVersion = instName + "/" + id
					ir.mu.Unlock()
					wrapped := make([]integrityTask, 0, len(tasks))
					for _, t := range tasks {
						wrapped = append(wrapped, integrityTask{Task: t, Version: id, Origin: "instance:" + instName})
					}
					addTasks(wrapped)
					versions++
					instVersions++
					e.log.Info("[Integrity] JSON de instancia %s (%s): %d archivos esperados", instName, id, len(tasks))
				}
			}
		}
	}

	if scope != IntegrityScopeInstances && versions == instVersions {
		e.log.Info("[Integrity] Sector global: sin versiones descargadas")
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Origin == out[j].Origin {
			return out[i].Task.Dest < out[j].Task.Dest
		}
		return out[i].Origin < out[j].Origin
	})
	e.log.Info("[Integrity] Indice construido: %d versiones (%d globales, %d de instancias), %d archivos unicos",
		versions, versions-instVersions, instVersions, len(out))
	return out, versions, nil
}

func (e *Engine) loadVersionJSON(path string) (*downloader.VersionJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ver downloader.VersionJSON
	if err := json.Unmarshal(data, &ver); err != nil {
		return nil, err
	}
	return &ver, nil
}

// integrityPassExistence: fase 1. Verifica la existencia de cada archivo; los
// que faltan se descargan. Un archivo que falla tras 3 intentos se cancela,
// queda registrado en el backend y se guarda en memoria para la fase 2.
func (e *Engine) integrityPassExistence(ctx context.Context, tasks []integrityTask, client *http.Client, ir *integrityRunner) []integrityTask {
	if len(tasks) == 0 {
		return nil
	}
	e.log.Info("[Integrity] Fase 1: existen %d archivos? descargando faltantes (%d intentos por archivo)", len(tasks), integrityAttemptsPass1)
	var pending []integrityTask
	var pendingMu sync.Mutex

	runIntegrityPool(ctx, tasks, e.integrityWorkers(), func(t integrityTask) {
		if downloader.FileExists(t.Task.Dest) {
			ir.tickProgress(1, t)
			return
		}
		ir.updateProgress(func(p *IntegrityProgress) {
			p.FilesMissing++
			p.CurrentFile = filepath.Base(t.Task.Dest)
		})
		err := e.integrityDownload(ctx, t, client, integrityAttemptsPass1)
		if err != nil {
			e.log.Warn("[Integrity] FASE 1: %s no descargable tras %d intentos: %v", t.Task.Dest, integrityAttemptsPass1, err)
			pendingMu.Lock()
			pending = append(pending, t)
			pendingMu.Unlock()
			ir.tickProgress(1, t)
			return
		}
		ir.updateProgress(func(p *IntegrityProgress) {
			p.FilesRestored++
		})
		ir.tickProgress(1, t)
		e.log.Info("[Integrity] FASE 1: descargado %s", t.Task.Dest)
	})
	return pending
}

// integrityPassRetry: fase 2. Vuelve a verificar la existencia de los archivos
// pendientes y los re-descarga; si un archivo falla 5 veces seguidas se saltea
// y queda registrado en el backend.
func (e *Engine) integrityPassRetry(ctx context.Context, pending []integrityTask, client *http.Client, ir *integrityRunner) {
	if len(pending) == 0 {
		e.log.Info("[Integrity] Fase 2: sin archivos pendientes")
		return
	}
	e.log.Info("[Integrity] Fase 2: reintentando %d archivos (%d intentos por archivo)", len(pending), integrityAttemptsPass2)
	runIntegrityPool(ctx, pending, e.integrityWorkers(), func(t integrityTask) {
		if !downloader.FileExists(t.Task.Dest) {
			err := e.integrityDownload(ctx, t, client, integrityAttemptsPass2)
			if err != nil {
				ir.updateProgress(func(p *IntegrityProgress) {
					p.FilesSkipped++
					if len(p.Skipped) < integrityMaxSkipped {
						p.Skipped = append(p.Skipped, IntegritySkipped{File: t.Task.Dest, Reason: err.Error()})
					}
				})
				e.log.Error("[Integrity] FASE 2: %s descartado tras %d intentos: %v", t.Task.Dest, integrityAttemptsPass2, err)
				ir.tickProgress(1, t)
				return
			}
			ir.updateProgress(func(p *IntegrityProgress) {
				p.FilesRestored++
			})
			e.log.Info("[Integrity] FASE 2: recuperado %s", t.Task.Dest)
		}
		ir.tickProgress(1, t)
	})
}

// integrityPassVerify: fase 3. Verifica TODOS los archivos por SHA1 y tamano;
// los corruptos se borran y se re-descargan; si vuelven a fallar (5 intentos)
// se saltean y quedan registrados en el backend.
func (e *Engine) integrityPassVerify(ctx context.Context, tasks []integrityTask, client *http.Client, ir *integrityRunner) {
	if len(tasks) == 0 {
		return
	}
	e.log.Info("[Integrity] Fase 3: verificando SHA1 y tamano de %d archivos", len(tasks))
	runIntegrityPool(ctx, tasks, e.integrityWorkers(), func(t integrityTask) {
		reason := integrityVerifyFile(t.Task)
		if reason == "" {
			ir.tickProgress(1, t)
			return
		}
		ir.updateProgress(func(p *IntegrityProgress) {
			p.FilesCorrupt++
			p.CurrentFile = filepath.Base(t.Task.Dest)
		})
		e.log.Warn("[Integrity] FASE 3: corrupto (%s): %s", reason, t.Task.Dest)
		os.Remove(t.Task.Dest)
		err := e.integrityDownload(ctx, t, client, integrityAttemptsPass2)
		if err != nil {
			ir.updateProgress(func(p *IntegrityProgress) {
				p.FilesSkipped++
				if len(p.Skipped) < integrityMaxSkipped {
					p.Skipped = append(p.Skipped, IntegritySkipped{File: t.Task.Dest, Reason: reason + ": " + err.Error()})
				}
			})
			e.log.Error("[Integrity] FASE 3: %s no recuperable: %v", t.Task.Dest, err)
		} else {
			ir.updateProgress(func(p *IntegrityProgress) {
				p.FilesRestored++
			})
			e.log.Info("[Integrity] FASE 3: re-descargado %s", t.Task.Dest)
		}
		ir.tickProgress(1, t)
	})
}

func integrityVerifyFile(t downloader.DownloadTask) string {
	if !downloader.FileExists(t.Dest) {
		return "no existe"
	}
	if t.SHA1 != "" {
		ok, err := downloader.VerifySHA1(t.Dest, t.SHA1)
		if err != nil || !ok {
			return "SHA1 no coincide"
		}
	}
	if t.Size > 0 {
		info, err := os.Stat(t.Dest)
		if err != nil || info.Size() != t.Size {
			return "tamano incorrecto"
		}
	}
	return ""
}

func (e *Engine) integrityDownload(ctx context.Context, t integrityTask, client *http.Client, attempts int) error {
	return downloader.DownloadFile(ctx, t.Task, client, attempts-1, nil, integrityStallTimeout, integrityMaxStall)
}

func (e *Engine) integrityWorkers() int {
	n := e.config.Get().ConcurrentDownloads * 3
	if n < 1 {
		n = 1
	}
	if n > 64 {
		n = 64
	}
	return n
}

func runIntegrityPool(ctx context.Context, tasks []integrityTask, workers int, fn func(integrityTask)) {
	if len(tasks) == 0 || workers < 1 {
		return
	}
	jobs := make(chan integrityTask, len(tasks))
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}
				fn(t)
			}
		}()
	}
sendLoop:
	for _, t := range tasks {
		select {
		case <-ctx.Done():
			break sendLoop
		default:
		}
		select {
		case jobs <- t:
		case <-ctx.Done():
			break sendLoop
		}
	}
	close(jobs)
	wg.Wait()
}

func (ir *integrityRunner) setPhase(phase string, phaseTotal, base, span int) {
	ir.mu.Lock()
	ir.prog.Phase = phase
	ir.phaseBase = base
	ir.phaseSpan = span
	ir.phaseTotal = phaseTotal
	ir.phaseDone = 0
	ir.mu.Unlock()
}

func (ir *integrityRunner) tickProgress(done int, t integrityTask) {
	ir.mu.Lock()
	ir.phaseDone += done
	if ir.prog.TasksDone < ir.prog.TasksTotal {
		ir.prog.TasksDone += done
	}
	if ir.prog.TasksTotal > 0 && ir.phaseTotal > 0 {
		frac := ir.phaseDone * ir.phaseSpan / ir.phaseTotal
		ir.prog.Percent = ir.phaseBase + frac
		if ir.prog.Percent > 100 {
			ir.prog.Percent = 100
		}
	}
	ir.prog.CurrentVersion = t.Version
	ir.prog.CurrentFile = ""
	ir.mu.Unlock()
}

func (ir *integrityRunner) updateProgress(fn func(*IntegrityProgress)) {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	fn(&ir.prog)
}