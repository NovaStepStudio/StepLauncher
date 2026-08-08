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
	sem := make(chan struct{}, maxWorkers)
	var mu sync.Mutex
	var failed []DownloadTask
	var wg sync.WaitGroup
	for i := range tasks {
		t := tasks[i]
		if t.SHA1 == "" {
			continue
		}
		if !FileExists(t.Dest) {
			mu.Lock()
			failed = append(failed, t)
			mu.Unlock()
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
		}(t)
	}
	wg.Wait()
	return failed
}
