package downloader

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	// DefaultTransport es la base de TODOS los clientes HTTP del launcher.
	// Reglas: siempre en directo (Proxy nil: no hereda el proxy del sistema
	// ni de variables de entorno) y solo HTTP/1.1 (la negociación HTTP/2
	// falla contra capas de inspección transparentes y produce el
	// "malformed HTTP status code"). El proxy manual de Ajustes > Red es
	// SOLO para el juego de Minecraft (flags -D de la JVM) y nunca se aplica
	// a peticiones del launcher.
	DefaultTransport = &http.Transport{
		Proxy: nil,
		// Desactivar HTTP/2 de forma explícita: ForceAttemptHTTP2 solo no
		// basta en transportes estándar; el mapa TLSNextProto no-nil evita
		// que net/http negocie h2 por ALPN.
		ForceAttemptHTTP2: false,
		TLSNextProto:      map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
		MaxIdleConns:      maxIdleConns,
		MaxConnsPerHost:   0,
		IdleConnTimeout:   idleConnTimeout,
		DisableCompression: false,
	}

	errStall = errors.New("download stalled: no data received within timeout")
)

func DefaultHTTPClient() *http.Client {
	return &http.Client{
		Transport: DefaultTransport.Clone(),
		Timeout:   90 * time.Second,
	}
}

// NewConfiguredHTTPClient construye el cliente HTTP del launcher: siempre en
// directo y sin HTTP/2 (ver DefaultTransport). El proxy manual de Ajustes >
// Red es SOLO para el juego de Minecraft (flags -D de la JVM en el lanzamiento)
// y nunca se aplica a peticiones del launcher: los parámetros proxy* se
// conservan por compatibilidad de firma pero se ignoran. Solo se aplica aquí
// el límite de velocidad.
func NewConfiguredHTTPClient(maxMbps float64, _ bool, _ string, _ int, _, _ string) (*http.Client, error) {
	transport := DefaultTransport.Clone()
	var roundTripper http.RoundTripper = transport
	if maxMbps > 0 {
		roundTripper = NewTransport(transport, maxMbps, 0)
	}
	return &http.Client{Transport: roundTripper, Timeout: 90 * time.Second}, nil
}

// isProxyProtocolError detecta el error típico de confundir un proxy SOCKS con uno HTTP
// (o viceversa), que se manifiesta como "malformed HTTP status code" con bytes binarios.
func isProxyProtocolError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "malformed HTTP status code") ||
		strings.Contains(msg, "malformed HTTP response") ||
		strings.Contains(msg, "HTTP/1.x transport connection broken") ||
		strings.Contains(msg, `"\x00`) ||
		strings.Contains(msg, "proxyconnect tcp")
}

// IsProxyProtocolError es la versión exportada para que otros paquetes puedan detectar
// si un error de fetch se debe a una mala configuración del proxy.
func IsProxyProtocolError(err error) bool {
	return isProxyProtocolError(err)
}

// WrapProxyError decora un error de red con una pista accionable sobre el proxy.
// Si no se conoce el proxy en uso (host vacío), no se inventa un "proxy :0":
// se indica que la petición iba directa o tras un proxy del sistema.
func WrapProxyError(err error, host string, port int) error {
	if err == nil {
		return nil
	}
	if isProxyProtocolError(err) {
		host = strings.TrimSpace(host)
		if host == "" {
			return fmt.Errorf("%w — la petición falló con respuesta no-HTTP (la red puede estar interceptada por un proxy del sistema, VPN o antivirus). Revisa tu conexión y Ajustes > Red > Proxy: si usas Clash/V2Ray, el puerto HTTP suele ser 7890, no el SOCKS 7891", err)
		}
		return fmt.Errorf("%w — el proxy %s:%d parece no ser HTTP (¿es SOCKS? Prueba con socks5://%s:%d o usa el puerto HTTP de tu proxy, ej. Clash usa 7890 para HTTP y 7891 para SOCKS; o desactívalo en Ajustes > Red)", err, host, port, host, port)
	}
	return err
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
