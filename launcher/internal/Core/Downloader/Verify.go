package downloader

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"sync"
)

func VerifySHA1(path, expected string) (bool, error) {
	if expected == "" {
		return true, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}
	return hex.EncodeToString(h.Sum(nil)) == expected, nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func VerifyBatch(tasks []DownloadTask, maxWorkers int) []DownloadTask {
	return VerifyBatchWithProgress(tasks, maxWorkers, nil)
}

// VerifyBatchWithProgress verifica los archivos en paralelo y, al avanzar cada
// archivo (verificado, ausente o sin sha1), notifica el progreso con
// onProgress(done, total). Sirve para que la UI muestre "Verificando archivos
// X/Y" en vivo en lugar de un porcentaje congelado durante la verificación.
func VerifyBatchWithProgress(tasks []DownloadTask, maxWorkers int, onProgress func(done, total int)) []DownloadTask {
	total := len(tasks)
	var done int64
	var mu sync.Mutex
	sem := make(chan struct{}, maxWorkers)
	var failed []DownloadTask
	var wg sync.WaitGroup

	next := func() {
		mu.Lock()
		done++
		d := done
		mu.Unlock()
		if onProgress != nil {
			onProgress(int(d), total)
		}
	}

	for i := range tasks {
		t := tasks[i]
		if t.SHA1 == "" {
			next()
			continue
		}
		if !FileExists(t.Dest) {
			mu.Lock()
			failed = append(failed, t)
			mu.Unlock()
			next()
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(task DownloadTask) {
			defer wg.Done()
			defer func() { <-sem }()
			ok, err := VerifySHA1(task.Dest, task.SHA1)
			if err != nil || !ok {
				os.Remove(task.Dest)
				mu.Lock()
				failed = append(failed, task)
				mu.Unlock()
			}
			next()
		}(t)
	}
	wg.Wait()
	return failed
}
