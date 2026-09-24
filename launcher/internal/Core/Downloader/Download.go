package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/net/proxy"
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
		Transport: DefaultTransport.Clone(),
		Timeout:   90 * time.Second,
	}
}

func NewConfiguredHTTPClient(maxMbps float64, proxyEnabled bool, proxyHost string, proxyPort int, proxyUser, proxyPass string) (*http.Client, error) {
	transport := DefaultTransport.Clone()
	if proxyEnabled {
		host := strings.TrimSpace(proxyHost)
		if host == "" {
			return nil, fmt.Errorf("el host del proxy es obligatorio")
		}
		if proxyPort < 1 || proxyPort > 65535 {
			return nil, fmt.Errorf("el puerto del proxy es inválido")
		}

		// Permitir que el usuario escriba el esquema en el host (ej. "socks5://127.0.0.1" o "http://proxy.local").
		// Si no hay esquema, se asume HTTP. Si es SOCKS, se configura un DialContext vía x/net/proxy.
		lowerHost := strings.ToLower(host)
		isSocks := strings.HasPrefix(lowerHost, "socks5://") || strings.HasPrefix(lowerHost, "socks5h://") || strings.HasPrefix(lowerHost, "socks4://") || strings.HasPrefix(lowerHost, "socks://")
		if strings.Contains(host, "://") {
			if u, err := url.Parse(host); err == nil && u.Host != "" {
				scheme := strings.ToLower(u.Scheme)
				if scheme == "socks5" || scheme == "socks5h" || scheme == "socks" || scheme == "socks4" || scheme == "socks4a" {
					isSocks = true
					host = u.Host
					// Si la URL trae puerto, tiene prioridad sobre proxyPort
					if h, p, err := net.SplitHostPort(host); err == nil {
						host = h
						if pp, err := strconv.Atoi(p); err == nil && pp != 0 {
							proxyPort = pp
						}
					}
					// Credenciales en la URL tienen prioridad
					if u.User != nil {
						proxyUser = u.User.Username()
						if pw, ok := u.User.Password(); ok {
							proxyPass = pw
						}
					}
				} else if scheme == "http" || scheme == "https" {
					isSocks = false
					host = u.Host
					if h, p, err := net.SplitHostPort(host); err == nil {
						host = h
						if pp, err := strconv.Atoi(p); err == nil && pp != 0 {
							proxyPort = pp
						}
					}
					if u.User != nil {
						proxyUser = u.User.Username()
						if pw, ok := u.User.Password(); ok {
							proxyPass = pw
						}
					}
				}
			} else if isSocks {
				// Fallback: quitar prefijo manual si url.Parse falló
				for _, pref := range []string{"socks5h://", "socks5://", "socks4://", "socks://"} {
					if strings.HasPrefix(lowerHost, pref) {
						host = host[len(pref):]
						break
					}
				}
			}
		}

		// Host puede venir como "127.0.0.1:7891" sin esquema; separarlo
		if !isSocks {
			if h, p, err := net.SplitHostPort(host); err == nil {
				// Solo aplicar si el host original no era una IP con puerto y el usuario no dejó el puerto separado
				// Si proxyHost contenía puerto y proxyPort coincide con el default 8080, se respeta el puerto del host
				if proxyPort == 8080 || proxyPort == 0 {
					if pp, err := strconv.Atoi(p); err == nil {
						host = h
						proxyPort = pp
					}
				}
			}
		} else {
			// Para SOCKS, también limpiar posible ":puerto" en host
			if h, p, err := net.SplitHostPort(host); err == nil {
				host = h
				if pp, err := strconv.Atoi(p); err == nil && pp != 0 {
					proxyPort = pp
				}
			}
		}

		host = strings.TrimSpace(host)
		if host == "" {
			return nil, fmt.Errorf("el host del proxy es obligatorio")
		}

		if isSocks {
			// SOCKS5: el http.Transport no usa ProxyURL, sino un DialContext que hace el handshake SOCKS
			proxyAddr := net.JoinHostPort(host, strconv.Itoa(proxyPort))
			var auth *proxy.Auth
			if proxyUser != "" {
				auth = &proxy.Auth{User: proxyUser, Password: proxyPass}
			}
			socksDialer, err := proxy.SOCKS5("tcp", proxyAddr, auth, proxy.Direct)
			if err != nil {
				return nil, fmt.Errorf("proxy SOCKS5 %s: %w", proxyAddr, err)
			}
			// DialContext que respeta el context de la request
			transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				// Si el contexto ya está cancelado, retornar rápido
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
				}
				// proxy.Dialer no soporta context, se usa Dial con timeout implícito del Transport
				// y se comprueba el contexto antes y después
				conn, err := socksDialer.Dial(network, addr)
				if err != nil {
					// Envolver error para que el caller pueda detectar mala configuración
					if isProxyProtocolError(err) {
						return nil, fmt.Errorf("%w (verifica que %s:%d sea un proxy SOCKS5 válido)", err, host, proxyPort)
					}
					return nil, err
				}
				return conn, nil
			}
			transport.Proxy = nil
		} else {
			proxyURL := &url.URL{Scheme: "http", Host: net.JoinHostPort(host, strconv.Itoa(proxyPort))}
			if proxyUser != "" {
				proxyURL.User = url.UserPassword(proxyUser, proxyPass)
			}
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

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
