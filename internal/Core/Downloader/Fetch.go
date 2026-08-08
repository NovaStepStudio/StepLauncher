package downloader

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxFetchSize = 50 * 1024 * 1024

type cachedJSON struct {
	CreatedAt time.Time       `json:"created_at"`
	ExpiresAt time.Time       `json:"expires_at"`
	Data      json.RawMessage `json:"data"`
}

func defaultTTL() time.Duration {
	return 24 * time.Hour
}

func cachePath(baseDir, cacheKey string) string {
	safe := strings.ReplaceAll(cacheKey, "/", "_")
	safe = strings.ReplaceAll(safe, ":", "_")
	return filepath.Join(baseDir, safe+".json")
}

func cacheCategory(key string) (category, subkey string) {
	if key == "manifest" || strings.HasPrefix(key, "manifest/") {
		return "manifest", key
	}
	if strings.HasPrefix(key, "version/") {
		return "versions", strings.TrimPrefix(key, "version/")
	}
	if strings.HasPrefix(key, "assets/") {
		return "assets", strings.TrimPrefix(key, "assets/")
	}
	if key == "java-products" {
		return "java", "products"
	}
	if strings.HasPrefix(key, "java/") {
		return "java", strings.TrimPrefix(key, "java/")
	}
	return "default", key
}

func FetchJSON(cfg Config, url, cacheKey string, out interface{}) error {
	if cfg.CacheManager != nil {
		return fetchJSONWithCache(cfg, url, cacheKey, out)
	}
	return fetchJSONFlat(cfg, url, cacheKey, out)
}

func fetchJSONWithCache(cfg Config, url, cacheKey string, out interface{}) error {
	category, subkey := cacheCategory(cacheKey)

	found, expired, err := cfg.CacheManager.GetWithFallback(category, subkey, out)
	if err == nil && found && !expired {
		return nil
	}

	client := cfg.HTTPClient
	if client == nil {
		return fmt.Errorf("HTTPClient is nil")
	}

	resp, err := client.Get(url)
	if err != nil {
		if found && expired {
			if cfg.LogFn != nil {
				cfg.LogFn("[Cache] WARN: using expired cache for %s/%s (remote unavailable: %v)", category, subkey, err)
			}
			return nil
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if found && expired && (resp.StatusCode == 429 || resp.StatusCode >= 500) {
			if cfg.LogFn != nil {
				cfg.LogFn("[Cache] WARN: using expired cache for %s/%s (HTTP %d)", category, subkey, resp.StatusCode)
			}
			return nil
		}
		return fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, url)
	}

	limited := io.LimitReader(resp.Body, maxFetchSize)
	data, err := io.ReadAll(limited)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, out); err != nil {
		return err
	}

	if saveErr := cfg.CacheManager.Set(category, subkey, data); saveErr != nil && cfg.LogFn != nil {
		cfg.LogFn("[Cache] WARN: failed to save cache %s/%s: %v", category, subkey, saveErr)
	}

	return nil
}

func fetchJSONFlat(cfg Config, url, cacheKey string, out interface{}) error {
	cacheDir := cfg.CacheDir
	if cacheDir == "" {
		cacheDir = filepath.Join(cfg.WorkDir, "cache")
	}
	path := cachePath(cacheDir, cacheKey)
	if data, expired, err := readCached(path); err == nil {
		if !expired {
			if err := json.Unmarshal(data, out); err == nil {
				return nil
			}
		}
	}
	client := cfg.HTTPClient
	if client == nil {
		return fmt.Errorf("HTTPClient is nil")
	}
	resp, err := client.Get(url)
	if err != nil {
		if data, _, cachedErr := readCached(path); cachedErr == nil {
			if jsonErr := json.Unmarshal(data, out); jsonErr == nil {
				return nil
			}
		}
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			if data, _, cachedErr := readCached(path); cachedErr == nil {
				if jsonErr := json.Unmarshal(data, out); jsonErr == nil {
					return nil
				}
			}
		}
		return fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, url)
	}
	limited := io.LimitReader(resp.Body, maxFetchSize)
	data, err := io.ReadAll(limited)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return err
	}
	meta := cachedJSON{
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(defaultTTL()),
		Data:      data,
	}
	metaRaw, _ := json.Marshal(meta)
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	os.WriteFile(path, metaRaw, 0644)
	return nil
}

func readCached(path string) ([]byte, bool, error) {
	metaRaw, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}
	var meta cachedJSON
	if err := json.Unmarshal(metaRaw, &meta); err != nil {
		return nil, false, err
	}
	expired := time.Now().After(meta.ExpiresAt)
	return meta.Data, expired, nil
}
