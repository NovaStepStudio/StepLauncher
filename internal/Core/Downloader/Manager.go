package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	globalutils "StepLauncher/internal/Core/Utils"
)

const terminalStateTTL = 30 * time.Second

var (
	ErrDownloadNotFound = fmt.Errorf("download not found")
	ErrNotDownloading   = fmt.Errorf("download is not in progress")
	ErrNotPaused        = fmt.Errorf("download is not paused")
	ErrAlreadyExists    = fmt.Errorf("download ID already exists")
)

type Manager struct {
	cfg       Config
	downloads map[string]*Download
	mu        sync.RWMutex
	nextID    uint64
	queue     *Queue
}

type Download struct {
	ID       string
	Version  string
	Filter   DownloadFilter
	State    DownloadState
	Progress DownloadProgress
	Error    string

	maxRetries      int
	maxConcurrency  int
	skipVerify      bool
	stallTimeout    int
	maxStallRetries int

	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
	paused    bool
	pauseCond *sync.Cond
	startTime time.Time

	tasks        []DownloadTask
	done         int
	existing     int
	failed       int
	mbDownloaded float64
	mbTotal      float64
	logLines     []string
	onProgress   func(*DownloadProgress)
	doneCh       chan struct{}

	fileStates   map[string]string
	fileBytes    map[string]int64
	fileSizes    map[string]int64
	fileSections map[string]string
	fileOrder    []string
	sectionOrder []string
	sectionStats map[string]*SectionProgress
	liveBytes    int64

	speedMbps     float64
	lastSpeedMB   float64
	lastSpeedTime time.Time

	lastEmittedState DownloadState
	lastEmitTime     time.Time

	nativeJars []string
}

func NewManager(cfg Config) *Manager {
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.MaxConcurrency <= 0 {
		cfg.MaxConcurrency = 24
	}
	if cfg.CacheDir == "" {
		cfg.CacheDir = filepath.Join(cfg.WorkDir, "cache")
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = DefaultHTTPClient()
	}
	if cfg.MaxMbps > 0 || cfg.MinMbps > 0 {
		cfg.HTTPClient = &http.Client{
			Transport: NewTransport(cfg.HTTPClient.Transport, cfg.MaxMbps, cfg.MinMbps),
			Timeout:   cfg.HTTPClient.Timeout,
		}
	}
	if cfg.MaxRAM == 0 {
		cfg.MaxRAM = 2048
	}
	if cfg.StallTimeout <= 0 {
		cfg.StallTimeout = 60000
	}
	if cfg.MaxStallReDownload <= 0 {
		cfg.MaxStallReDownload = 3
	}

	return &Manager{
		cfg:       cfg,
		downloads: make(map[string]*Download),
		queue:     NewQueue(cfg.MaxConcurrency),
	}
}

func (m *Manager) Start(version string, filter DownloadFilter, maxRetries int, maxConcurrency int, skipVerify bool, stallTimeout int, maxStallRetries int) *Download {
	m.mu.Lock()
	m.nextID++
	id := fmt.Sprintf("dl-%d", m.nextID)
	m.mu.Unlock()

	if maxRetries <= 0 {
		maxRetries = m.cfg.MaxRetries
	}
	if maxConcurrency <= 0 {
		maxConcurrency = m.cfg.MaxConcurrency
	}
	if stallTimeout <= 0 {
		stallTimeout = m.cfg.StallTimeout
	}
	if maxStallRetries <= 0 {
		maxStallRetries = m.cfg.MaxStallReDownload
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)

	if !filter.Client && !filter.Libraries && !filter.Natives && !filter.Assets && !filter.Java {
		filter.Client = true
		filter.Libraries = true
		filter.Natives = true
		filter.Assets = true
		filter.Java = true
	}

	dl := &Download{
		ID:              id,
		Version:         version,
		Filter:          filter,
		State:           StatePending,
		maxRetries:      maxRetries,
		maxConcurrency:  maxConcurrency,
		skipVerify:      skipVerify,
		stallTimeout:    stallTimeout,
		maxStallRetries: maxStallRetries,
		ctx:             ctx,
		cancel:          cancel,
		startTime:       time.Now(),
		lastSpeedTime:   time.Now(),
		doneCh:          make(chan struct{}),
		Progress: DownloadProgress{
			State:             StatePending,
			SectionsTotal:     countSections(filter),
			Sections:          []SectionProgress{},
			SectionsCompleted: []string{},
			ActiveFiles:       []FileProgress{},
			QueuedPreview:     []string{},
		},
	}
	dl.pauseCond = sync.NewCond(&dl.mu)

	m.mu.Lock()
	m.downloads[id] = dl
	m.mu.Unlock()

	BroadcastState(m.cfg.BroadcastFn, id, StatePending)
	m.emitProgress(dl)

	m.queue.Add(func() {
		m.runDownload(dl)
	})

	return dl
}

func (m *Manager) List() []*Download {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*Download, 0, len(m.downloads))
	for _, dl := range m.downloads {
		list = append(list, dl)
	}
	return list
}

func (m *Manager) Get(id string) *Download {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.downloads[id]
}

func (m *Manager) HTTPClient() *http.Client {
	return m.cfg.HTTPClient
}

func (m *Manager) Pause(id string) error {
	dl := m.Get(id)
	if dl == nil {
		return fmt.Errorf("%w: %s", ErrDownloadNotFound, id)
	}

	dl.mu.Lock()
	if dl.State != StateDownloading {
		dl.mu.Unlock()
		return fmt.Errorf("%w (state: %s)", ErrNotDownloading, dl.State)
	}
	dl.paused = true
	dl.State = StatePaused
	dl.Progress.State = StatePaused
	dl.mu.Unlock()

	BroadcastState(m.cfg.BroadcastFn, id, StatePaused)
	m.log(dl, "Download paused")
	return nil
}

func (m *Manager) Resume(id string) error {
	dl := m.Get(id)
	if dl == nil {
		return fmt.Errorf("%w: %s", ErrDownloadNotFound, id)
	}

	dl.mu.Lock()
	if dl.State != StatePaused {
		dl.mu.Unlock()
		return fmt.Errorf("%w (state: %s)", ErrNotPaused, dl.State)
	}
	dl.paused = false
	dl.State = StateDownloading
	dl.Progress.State = StateDownloading
	dl.pauseCond.Broadcast()
	dl.mu.Unlock()

	BroadcastState(m.cfg.BroadcastFn, id, StateDownloading)
	m.log(dl, "Download resumed")
	return nil
}

func (m *Manager) CancelActive() {
	m.mu.RLock()
	ids := make([]string, 0, len(m.downloads))
	for id, dl := range m.downloads {
		dl.mu.Lock()
		st := dl.State
		dl.mu.Unlock()
		switch st {
		case StatePending, StateDownloading, StatePaused, StateVerifying, StateReDownload:
			ids = append(ids, id)
		}
	}
	m.mu.RUnlock()
	for _, id := range ids {
		_ = m.Cancel(id)
	}
}

func (m *Manager) Cancel(id string) error {
	dl := m.Get(id)
	if dl == nil {
		return fmt.Errorf("%w: %s", ErrDownloadNotFound, id)
	}
	dl.mu.Lock()
	wasPaused := dl.paused
	dl.mu.Unlock()

	if wasPaused {
		dl.mu.Lock()
		dl.paused = false
		dl.pauseCond.Broadcast()
		dl.mu.Unlock()
	}

	dl.cancel()
	time.Sleep(100 * time.Millisecond)

	dl.mu.Lock()
	dl.State = StateCancelled
	dl.Progress.State = StateCancelled
	dl.liveBytes = 0
	dl.mu.Unlock()

	BroadcastState(m.cfg.BroadcastFn, id, StateCancelled)
	m.log(dl, "Download cancelled")

	time.AfterFunc(terminalStateTTL, func() {
		m.mu.Lock()
		delete(m.downloads, id)
		m.mu.Unlock()
	})
	return nil
}

func (m *Manager) Status(id string) (*DownloadProgress, error) {
	dl := m.Get(id)
	if dl == nil {
		return nil, fmt.Errorf("%w: %s", ErrDownloadNotFound, id)
	}
	dl.mu.Lock()
	defer dl.mu.Unlock()
	p := dl.Progress
	return &p, nil
}

func (dl *Download) SetProgressCallback(cb func(*DownloadProgress)) {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	dl.onProgress = cb
}

func (dl *Download) OnProgress(cb func(*DownloadProgress)) func() {
	dl.SetProgressCallback(cb)
	return func() {
		dl.mu.Lock()
		defer dl.mu.Unlock()
		dl.onProgress = nil
	}
}

func (dl *Download) Wait() {
	<-dl.doneCh
}

type DownloadInfo struct {
	ID      string
	Version string
	State   DownloadState
	Error   string
}

func (m *Manager) GetInfo(id string) *DownloadInfo {
	dl := m.Get(id)
	if dl == nil {
		return nil
	}
	dl.mu.Lock()
	defer dl.mu.Unlock()
	return &DownloadInfo{
		ID:      dl.ID,
		Version: dl.Version,
		State:   dl.State,
		Error:   dl.Error,
	}
}

func countSections(f DownloadFilter) int {
	n := 0
	if f.Client {
		n++
	}
	if f.Libraries {
		n++
	}
	if f.Natives {
		n++
	}
	if f.Assets {
		n++
	}
	if f.Java {
		n++
	}
	return n
}

func (m *Manager) runDownload(dl *Download) {
	defer close(dl.doneCh)

	log := func(f string, a ...interface{}) {
		m.log(dl, f, a...)
	}

	dl.mu.Lock()
	if dl.ctx.Err() != nil {
		dl.State = StateCancelled
		dl.Progress.State = StateCancelled
		dl.mu.Unlock()
		return
	}
	dl.State = StateDownloading
	dl.Progress.State = StateDownloading
	dl.Progress.SectionsTotal = countSections(dl.Filter)
	dl.mu.Unlock()
	BroadcastState(m.cfg.BroadcastFn, dl.ID, StateDownloading)

	log("Starting download of Minecraft %s", dl.Version)
	m.emitProgress(dl)

	var manifest Manifest
	if err := FetchJSON(m.cfg, "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json", "manifest", &manifest); err != nil {
		m.setError(dl, fmt.Errorf("manifest: %w", err))
		return
	}

	var verURL string
	for _, v := range manifest.Versions {
		if v.ID == dl.Version {
			verURL = v.URL
			break
		}
	}
	if verURL == "" {
		m.setError(dl, fmt.Errorf("version %s not found", dl.Version))
		return
	}

	var ver VersionJSON
	if err := FetchJSON(m.cfg, verURL, "version/"+dl.Version, &ver); err != nil {
		m.setError(dl, fmt.Errorf("version json: %w", err))
		return
	}
	log("Fetched version JSON for %s", dl.Version)

	verDir := filepath.Join(m.cfg.WorkDir, "versions", dl.Version)
	if dl.Filter.InstanceVersionDir != "" {
		verDir = dl.Filter.InstanceVersionDir
	}
	verPath := filepath.Join(verDir, dl.Version+".json")
	os.MkdirAll(filepath.Dir(verPath), 0755)
	if data, err := json.Marshal(&ver); err == nil {
		os.WriteFile(verPath, data, 0644)
	} else {
		log("WARN: saving version json: %v", err)
	}

	allTasks, err := BuildTasks(m.cfg, &ver, dl.Version, dl.Filter)
	if err != nil {
		m.setError(dl, fmt.Errorf("build tasks: %w", err))
		return
	}

	var totalBytes int64
	for _, t := range allTasks {
		if t.Size > 0 {
			totalBytes += t.Size
		}
	}

	dl.mu.Lock()
	dl.tasks = allTasks
	dl.mbTotal = float64(totalBytes) / 1024 / 1024
	dl.Progress.MBTotal = dl.mbTotal
	dl.Progress.FilesTotal = len(allTasks)
	dl.Progress.FilesDownloaded = 0
	dl.Progress.FilesExisting = 0
	dl.Progress.SectionsCompleted = []string{}
	dl.Progress.SectionsTotal = countSections(dl.Filter)
	dl.fileStates = make(map[string]string, len(allTasks))
	dl.fileBytes = make(map[string]int64, len(allTasks))
	dl.fileSizes = make(map[string]int64, len(allTasks))
	dl.fileSections = make(map[string]string, len(allTasks))
	dl.fileOrder = make([]string, 0, len(allTasks))
	dl.sectionOrder = nil
	dl.sectionStats = make(map[string]*SectionProgress)
	dl.liveBytes = 0
	for _, t := range allTasks {
		dl.fileOrder = append(dl.fileOrder, t.Dest)
		dl.fileStates[t.Dest] = FilePending
		dl.fileSizes[t.Dest] = t.Size
		dl.fileSections[t.Dest] = t.Section
		if _, ok := dl.sectionStats[t.Section]; !ok {
			dl.sectionStats[t.Section] = &SectionProgress{Name: t.Section}
			dl.sectionOrder = append(dl.sectionOrder, t.Section)
		}
		s := dl.sectionStats[t.Section]
		s.TotalFiles++
		if t.Size > 0 {
			s.MBTotal += float64(t.Size) / 1024 / 1024
		}
	}
	dl.mu.Unlock()

	if len(allTasks) == 0 {
		m.setCompleted(dl)
		return
	}

	m.emitProgress(dl)

	m.downloadAll(dl, allTasks)

	if !dl.skipVerify && dl.State == StateDownloading {
		m.verifyAll(dl, allTasks)
	}

	if dl.State == StateDownloading && len(dl.nativeJars) > 0 {
		dl.mu.Lock()
		dl.State = StateVerifying
		dl.Progress.State = StateVerifying
		dl.Progress.CurrentSection = "extracting_natives"
		dl.Progress.CurrentFile = ""
		dl.mu.Unlock()
		BroadcastState(m.cfg.BroadcastFn, dl.ID, StateVerifying)
		m.emitProgress(dl)
		log("Extracting %d native libraries...", len(dl.nativeJars))

		nativesDir := filepath.Join(verDir, "natives")
		extracted, extrErr := globalutils.ExtractNatives(dl.nativeJars, nativesDir)
		if extrErr != nil {
			log("Extract natives: %v", extrErr)
		}
		if extracted > 0 {
			log("Extracted %d native files to %s", extracted, nativesDir)
		}
	}

	dl.mu.Lock()
	shouldComplete := dl.State == StateDownloading || dl.State == StateVerifying
	hasFailed := dl.failed
	hasError := dl.State == StateError
	dl.mu.Unlock()

	if !shouldComplete {
		if hasError {
			return
		}
		return
	}
	if hasFailed > 0 {
		m.setError(dl, fmt.Errorf("%d files failed to download", hasFailed))
	} else {
		m.setCompleted(dl)
	}
}

func (m *Manager) downloadAll(dl *Download, taskList []DownloadTask) {
	numWorkers := dl.maxConcurrency * 3
	if numWorkers > len(taskList) {
		numWorkers = len(taskList)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}
	const maxWorkers = 64
	if numWorkers > maxWorkers {
		numWorkers = maxWorkers
	}

	jobs := make(chan DownloadTask, len(taskList))
	var wg sync.WaitGroup
	var nativeMu sync.Mutex

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobs {
				if !m.checkState(dl) {
					return
				}
				m.processTask(dl, task, &nativeMu)
			}
		}()
	}

	for _, task := range taskList {
		if !m.checkState(dl) {
			close(jobs)
			return
		}
		jobs <- task
	}
	close(jobs)

	wg.Wait()
}

func (m *Manager) processTask(dl *Download, task DownloadTask, nativeMu *sync.Mutex) {
	dl.mu.Lock()
	dl.Progress.CurrentSection = task.Section
	dl.Progress.CurrentFile = filepath.Base(task.Dest)
	dl.Progress.CurrentURL = task.URL
	dl.Progress.CurrentDest = task.Dest
	dl.fileStates[task.Dest] = FileDownloading
	dl.fileBytes[task.Dest] = 0
	dl.mu.Unlock()
	m.emitProgress(dl)

	if FileExists(task.Dest) {
		dl.mu.Lock()
		dl.done++
		dl.existing++
		dl.Progress.FilesDownloaded = dl.done
		dl.Progress.FilesExisting = dl.existing
		if task.Size > 0 {
			dl.mbDownloaded += float64(task.Size) / 1024 / 1024
			dl.Progress.MBDownloaded = dl.mbDownloaded
		}
		dl.Progress.Percent = CalcPercent(dl.mbDownloaded, dl.mbTotal, dl.done, len(dl.tasks))
		dl.Progress.CurrentProgress = 100
		dl.fileStates[task.Dest] = FileExisting
		dl.fileBytes[task.Dest] = task.Size
		if s := dl.sectionStats[task.Section]; s != nil {
			s.DoneFiles++
			if task.Size > 0 {
				s.MBDownloaded += float64(task.Size) / 1024 / 1024
			}
		}
		dl.mu.Unlock()
		m.emitProgress(dl)

		if task.Section == "natives" {
			nativeMu.Lock()
			dl.nativeJars = append(dl.nativeJars, task.Dest)
			nativeMu.Unlock()
		}
		return
	}

	err := DownloadFile(dl.ctx, task, m.cfg.HTTPClient, dl.maxRetries, func(done, total int64) {
		dl.mu.Lock()
		prev := dl.fileBytes[task.Dest]
		dl.fileBytes[task.Dest] = done
		dl.liveBytes += done - prev
		dl.mu.Unlock()
		m.emitProgress(dl)
	}, dl.stallTimeout, dl.maxStallRetries)
	if err != nil {
		dl.mu.Lock()
		dl.failed++
		dl.fileStates[task.Dest] = FileError
		dl.liveBytes -= dl.fileBytes[task.Dest]
		dl.mu.Unlock()
		m.log(dl, "FAIL: %s | %s -> %s (%d B)", task.URL, err.Error(), task.Dest, task.Size)
		return
	}

	if task.Section == "natives" {
		nativeMu.Lock()
		dl.nativeJars = append(dl.nativeJars, task.Dest)
		nativeMu.Unlock()
	}

	dl.mu.Lock()
	dl.done++
	dl.Progress.FilesDownloaded = dl.done
	dl.liveBytes -= dl.fileBytes[task.Dest]
	dl.fileBytes[task.Dest] = task.Size
	if task.Size > 0 {
		dl.mbDownloaded += float64(task.Size) / 1024 / 1024
		dl.Progress.MBDownloaded = dl.mbDownloaded
	}
	dl.Progress.Percent = CalcPercent(dl.mbDownloaded, dl.mbTotal, dl.done, len(dl.tasks))
	dl.Progress.CurrentProgress = 100
	dl.fileStates[task.Dest] = FileDone
	if s := dl.sectionStats[task.Section]; s != nil {
		s.DoneFiles++
		if task.Size > 0 {
			s.MBDownloaded += float64(task.Size) / 1024 / 1024
		}
	}
	dl.mu.Unlock()
	m.emitProgress(dl)
}

func (m *Manager) verifyAll(dl *Download, tasks []DownloadTask) {
	dl.mu.Lock()
	dl.State = StateVerifying
	dl.Progress.State = StateVerifying
	dl.mu.Unlock()
	BroadcastState(m.cfg.BroadcastFn, dl.ID, StateVerifying)
	m.log(dl, "Verifying %d files...", len(tasks))

	failed := VerifyBatch(tasks, dl.maxConcurrency)

	if len(failed) == 0 {
		return
	}

	m.log(dl, "Verify: %s", FormatFailed(failed))

	maxAttempts := 3
	for attempt := 1; attempt <= maxAttempts && len(failed) > 0; attempt++ {
		dl.mu.Lock()
		dl.State = StateReDownload
		dl.Progress.State = StateReDownload
		dl.mu.Unlock()
		BroadcastState(m.cfg.BroadcastFn, dl.ID, StateReDownload)
		m.log(dl, "Re-downloading %d failed files (attempt %d/%d)", len(failed), attempt, maxAttempts)

		m.downloadAll(dl, failed)

		dl.mu.Lock()
		dl.State = StateVerifying
		dl.Progress.State = StateVerifying
		dl.mu.Unlock()
		BroadcastState(m.cfg.BroadcastFn, dl.ID, StateVerifying)

		failed = VerifyBatch(failed, dl.maxConcurrency)
	}

	if len(failed) > 0 {
		m.setError(dl, fmt.Errorf("verification failed for %d files", len(failed)))
	}
}

func (m *Manager) checkState(dl *Download) bool {
	select {
	case <-dl.ctx.Done():
		return false
	default:
	}

	dl.mu.Lock()
	for dl.paused {
		dl.pauseCond.Wait()
	}
	dl.mu.Unlock()

	select {
	case <-dl.ctx.Done():
		return false
	default:
	}
	return true
}

func (m *Manager) setError(dl *Download, err error) {
	dl.mu.Lock()
	dl.State = StateError
	dl.Error = err.Error()
	dl.Progress.State = StateError
	dl.mu.Unlock()
	BroadcastError(m.cfg.BroadcastFn, dl.ID, err.Error())
	m.emitProgress(dl)
	m.log(dl, "ERROR: %v", err)

	time.AfterFunc(terminalStateTTL, func() {
		m.mu.Lock()
		delete(m.downloads, dl.ID)
		m.mu.Unlock()
	})
}

func (m *Manager) setCompleted(dl *Download) {
	dl.mu.Lock()
	elapsed := time.Since(dl.startTime)
	dl.State = StateCompleted
	dl.Progress.State = StateCompleted
	dl.Progress.Percent = 100
	dl.Progress.CurrentProgress = 100
	dl.liveBytes = 0
	mbAvg := float64(0)
	if elapsed.Seconds() > 0 && dl.mbDownloaded > 0 {
		mbAvg = dl.mbDownloaded / elapsed.Seconds()
	}
	dl.mu.Unlock()
	BroadcastState(m.cfg.BroadcastFn, dl.ID, StateCompleted)
	m.emitProgress(dl)
	m.log(dl, "Complete: %d files | %d existing | %d failed | %.1f MB | %.1f MB/s | %s",
		dl.done, dl.existing, dl.failed, dl.mbDownloaded, mbAvg, elapsed.Round(time.Second))

	time.AfterFunc(terminalStateTTL, func() {
		m.mu.Lock()
		delete(m.downloads, dl.ID)
		m.mu.Unlock()
	})
}

func (m *Manager) log(dl *Download, f string, a ...interface{}) {
	msg := fmt.Sprintf(f, a...)
	if m.cfg.LogFn != nil {
		m.cfg.LogFn("[Downloader] %s", msg)
	}
	dl.mu.Lock()
	dl.logLines = append(dl.logLines, msg)
	if len(dl.logLines) > 100 {
		dl.logLines = dl.logLines[len(dl.logLines)-100:]
	}
	dl.Progress.Log = dl.logLines
	dl.mu.Unlock()
	BroadcastLog(m.cfg.BroadcastFn, dl.ID, msg)
}

func (m *Manager) emitProgress(dl *Download) {
	dl.mu.Lock()

	totalMB := dl.mbDownloaded + float64(dl.liveBytes)/1024/1024
	dl.Progress.MBDownloaded = totalMB

	now := time.Now()
	dt := now.Sub(dl.lastSpeedTime).Seconds()
	if dt >= 0.4 {
		if dt <= 2.5 && dl.Progress.State == StateDownloading {
			inst := (totalMB - dl.lastSpeedMB) / dt
			if inst < 0 {
				inst = 0
			}
			if dl.speedMbps <= 0 {
				dl.speedMbps = inst
			} else {
				dl.speedMbps = 0.7*dl.speedMbps + 0.3*inst
			}
		} else {
			dl.speedMbps = 0
		}
		dl.lastSpeedMB = totalMB
		dl.lastSpeedTime = now
	}
	dl.Progress.SpeedMbps = dl.speedMbps

	sections := make([]SectionProgress, 0, len(dl.sectionOrder))
	for _, name := range dl.sectionOrder {
		if s := dl.sectionStats[name]; s != nil {
			c := *s
			sections = append(sections, c)
		}
	}
	dl.Progress.Sections = sections

	completed := make([]string, 0, len(sections))
	for _, s := range sections {
		if s.TotalFiles > 0 && s.DoneFiles >= s.TotalFiles {
			completed = append(completed, s.Name)
		}
	}
	dl.Progress.SectionsCompleted = completed

	active := make([]FileProgress, 0, 8)
	queued := 0
	preview := make([]string, 0, 5)
	for _, dest := range dl.fileOrder {
		st := dl.fileStates[dest]
		switch st {
		case FileDownloading:
			if len(active) >= 8 {
				continue
			}
			size := dl.fileSizes[dest]
			bytes := dl.fileBytes[dest]
			percent := 0.0
			if size > 0 {
				percent = float64(bytes) / float64(size) * 100
			}
			active = append(active, FileProgress{
				Name: filepath.Base(dest), Section: dl.fileSections[dest],
				Size: size, Downloaded: bytes, Percent: percent, State: st,
			})
		case FilePending:
			queued++
			if len(preview) < 5 {
				preview = append(preview, filepath.Base(dest))
			}
		}
	}
	dl.Progress.ActiveFiles = active
	dl.Progress.QueuedCount = queued
	dl.Progress.QueuedPreview = preview

	stateChanged := dl.Progress.State != dl.lastEmittedState
	timeElapsed := now.Sub(dl.lastEmitTime) >= 200*time.Millisecond
	shouldEmit := stateChanged || timeElapsed || dl.Progress.Percent >= 100

	if !shouldEmit {
		dl.mu.Unlock()
		return
	}

	p := dl.Progress
	cb := dl.onProgress
	dl.lastEmittedState = p.State
	dl.lastEmitTime = now
	dl.mu.Unlock()

	if cb != nil {
		cb(&p)
	}
	BroadcastProgress(m.cfg.BroadcastFn, dl.ID, &p)
}
