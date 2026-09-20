package authlib

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	InjectorLatestAPI = "https://authlib-injector.yushi.moe/artifact/latest.json"
	InjectorBMCLAPI   = "https://bmclapi2.bangbang93.com/mirrors/authlib-injector/artifact/latest.json"
	InjectorFileName  = "authlib-injector.jar"
)

type injectorArtifact struct {
	BuildNumber int    `json:"build_number"`
	Version     string `json:"version"`
	DownloadURL string `json:"download_url"`
	Checksums   struct {
		SHA256 string `json:"sha256"`
	} `json:"checksums"`
}

func EnsureInjector(ctx context.Context, jarPath string, force bool) (string, error) {
	if !force {
		if st, err := os.Stat(jarPath); err == nil && st.Size() > 0 {
			return jarPath, nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(jarPath), 0755); err != nil {
		return "", fmt.Errorf("no se pudo crear el directorio del injector: %w", err)
	}

	art, err := fetchInjectorArtifact(ctx, InjectorLatestAPI)
	if err != nil {
		art, err = fetchInjectorArtifact(ctx, InjectorBMCLAPI)
		if err != nil {
			return "", fmt.Errorf("no se pudo obtener la version de authlib-injector: %w", err)
		}
	}
	if art.DownloadURL == "" {
		return "", fmt.Errorf("la API de authlib-injector no devolvio download_url")
	}

	tmp := jarPath + ".part"
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, art.DownloadURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("no se pudo descargar authlib-injector: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("la descarga de authlib-injector respondio %d", resp.StatusCode)
	}

	f, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	_, err = io.Copy(io.MultiWriter(f, hash), resp.Body)
	cerr := f.Close()
	if err != nil {
		os.Remove(tmp)
		return "", fmt.Errorf("no se pudo guardar authlib-injector: %w", err)
	}
	if cerr != nil {
		os.Remove(tmp)
		return "", cerr
	}
	if art.Checksums.SHA256 != "" {
		got := hex.EncodeToString(hash.Sum(nil))
		if !strings.EqualFold(got, art.Checksums.SHA256) {
			os.Remove(tmp)
			return "", fmt.Errorf("checksum SHA-256 invalido en authlib-injector")
		}
	}
	if err := os.Rename(tmp, jarPath); err != nil {
		os.Remove(tmp)
		return "", err
	}
	return jarPath, nil
}

func fetchInjectorArtifact(ctx context.Context, api string) (*injectorArtifact, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("la API de authlib-injector respondio %d", resp.StatusCode)
	}
	var art injectorArtifact
	if err := json.NewDecoder(resp.Body).Decode(&art); err != nil {
		return nil, err
	}
	return &art, nil
}
