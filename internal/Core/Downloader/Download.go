package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

const (
	bufferSize       = 256 * 1024
	progressInterval = 300 * time.Millisecond
	retryBaseDelay   = 1000 * time.Millisecond

	maxIdleConns    = 100
	idleConnTimeout = 90 * time.Second
)

var (
	DefaultTransport = &http.Transport{
		MaxIdleConns:       maxIdleConns,
		MaxConnsPerHost:    0,
		IdleConnTimeout:    idleConnTimeout,
		DisableCompression: false,
	}

	errStall = errors.New("download stalled: no data received within timeout")
)

func DefaultHTTPClient() *http.Client {
	return &http.Client{
		Transport: DefaultTransport,
		Timeout:   90 * time.Second,
	}
}

func DownloadFile(ctx context.Context, task DownloadTask, client *http.Client, maxRetries int, onProgress func(int64, int64), stallTimeoutMs int, maxStallRetries int) error {
	dir := filepath.Dir(task.Dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	tmpDest := task.Dest + ".tmp"
	stallCount := 0

	for i := 0; i <= maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := tryDownload(ctx, task.URL, tmpDest, task.Dest, client, onProgress, stallTimeoutMs)
		if err == nil {
			return nil
		}

		os.Remove(tmpDest)

		if errors.Is(err, errStall) {
			stallCount++
			if stallCount >= maxStallRetries {
				return fmt.Errorf("download failed after %d stall retries: %w", stallCount, errStall)
			}
			if i < maxRetries {
				continue
			}
			return fmt.Errorf("download failed after %d retries (last: stall): %w", maxRetries, err)
		}

		if i < maxRetries {
			delay := retryBaseDelay * (1 << i)
			jitter := time.Duration(rand.Int63n(int64(delay / 2)))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay + jitter):
			}
		} else {
			return fmt.Errorf("download failed after %d retries: %w", maxRetries, err)
		}
	}
	return nil
}

func tryDownload(ctx context.Context, url, tmpDest, finalDest string, client *http.Client, onProgress func(int64, int64), stallTimeoutMs int) error {
	stallTimeout := time.Duration(stallTimeoutMs) * time.Millisecond

	attemptCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	req, err := http.NewRequestWithContext(attemptCtx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(tmpDest)
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}

	var lastRead atomic.Int64
	lastRead.Store(time.Now().UnixNano())

	go func() {
		ticker := time.NewTicker(stallTimeout)
		defer ticker.Stop()
		for {
			select {
			case <-attemptCtx.Done():
				return
			case <-ticker.C:
				if time.Duration(time.Now().UnixNano()-lastRead.Load()) >= stallTimeout {
					cancel()
					return
				}
			}
		}
	}()

	var written int64
	lastUpdate := time.Now()
	contentLen := resp.ContentLength
	buf := make([]byte, bufferSize)

	for {
		select {
		case <-attemptCtx.Done():
			out.Close()
			os.Remove(tmpDest)
			if errors.Is(attemptCtx.Err(), context.Canceled) {
				return errStall
			}
			return attemptCtx.Err()
		default:
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			lastRead.Store(time.Now().UnixNano())

			wn, werr := out.Write(buf[:n])
			if werr != nil {
				out.Close()
				return fmt.Errorf("write: %w", werr)
			}
			written += int64(wn)

			if onProgress != nil && contentLen > 0 {
				now := time.Now()
				if now.Sub(lastUpdate) >= progressInterval {
					lastUpdate = now
					onProgress(written, contentLen)
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			out.Close()
			if errors.Is(readErr, context.Canceled) {
				return errStall
			}
			return fmt.Errorf("read: %w", readErr)
		}
	}

	out.Close()

	if onProgress != nil && contentLen > 0 {
		onProgress(written, contentLen)
	}

	if err := os.Rename(tmpDest, finalDest); err != nil {
		return fmt.Errorf("rename: %w", err)
	}

	return nil
}
