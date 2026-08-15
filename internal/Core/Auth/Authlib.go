package authlib

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const ALIHeader = "X-Authlib-Injector-API-Location"

type Profile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AuthResult struct {
	AccessToken       string    `json:"accessToken"`
	ClientToken       string    `json:"clientToken"`
	SelectedProfile   Profile   `json:"selectedProfile"`
	AvailableProfiles []Profile `json:"availableProfiles"`
}

type AuthError struct {
	Message    string
	RawError   string
	Cause      string
	StatusCode int
}

func (e *AuthError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("el servidor de autenticacion devolvio un error (%d)", e.StatusCode)
}

type Client struct {
	http    *http.Client
	baseURL string
	root    string
}

func New(baseURL string) (*Client, error) {
	base, err := NormalizeServerURL(baseURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		http:    &http.Client{Timeout: 20 * time.Second},
		baseURL: base,
		root:    strings.TrimSuffix(base, "/authserver"),
	}, nil
}

func NewResolved(ctx context.Context, baseURL string) (*Client, error) {
	root, err := ResolveServerURL(ctx, baseURL)
	if err != nil {
		return nil, err
	}
	return New(root)
}

func ResolveServerURL(ctx context.Context, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("la URL del servidor de autenticacion es obligatoria")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("URL de servidor de autenticacion invalida: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("el servidor de autenticacion debe ser una URL http(s)")
	}
	if u.Host == "" {
		return "", fmt.Errorf("el servidor de autenticacion necesita un dominio valido")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("no se pudo contactar con el servidor de autenticacion: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	final := resp.Request.URL
	if ali := resp.Header.Get(ALIHeader); ali != "" {
		if ref, perr := url.Parse(ali); perr == nil {
			if resolved := final.ResolveReference(ref); resolved.String() != final.String() {
				final = resolved
			}
		}
	}
	return strings.TrimRight(final.String(), "/"), nil
}

func FetchMetadata(ctx context.Context, apiRoot string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(apiRoot, "/"), nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener la metadata del auth server: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer la metadata del auth server: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("el auth server respondio %d al pedir la metadata", resp.StatusCode)
	}
	return body, nil
}

type AuthMeta struct {
	Meta *AuthMetaInfo `json:"meta,omitempty"`
}

type AuthMetaInfo struct {
	ServerName string `json:"serverName,omitempty"`
}

func (c *Client) AuthServerName(ctx context.Context) string {
	raw, err := FetchMetadata(ctx, c.root)
	if err != nil {
		return ""
	}
	var meta AuthMeta
	if err := json.Unmarshal(raw, &meta); err != nil || meta.Meta == nil {
		return ""
	}
	return strings.TrimSpace(meta.Meta.ServerName)
}

func NormalizeServerURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("la URL del servidor de autenticacion es obligatoria")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("URL de servidor de autenticacion invalida: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("el servidor de autenticacion debe ser una URL http(s)")
	}
	if u.Host == "" {
		return "", fmt.Errorf("el servidor de autenticacion necesita un dominio valido")
	}
	path := strings.TrimRight(u.Path, "/")
	if !strings.HasSuffix(path, "/authserver") {
		path = path + "/authserver"
	}
	u.Path = path
	return strings.TrimRight(u.String(), "/"), nil
}

func (c *Client) Authenticate(ctx context.Context, username, password, clientToken string) (*AuthResult, error) {
	body := map[string]any{
		"agent":       map[string]any{"name": "Minecraft", "version": 1},
		"username":    username,
		"password":    password,
		"clientToken": clientToken,
		"requestUser": true,
	}
	var out AuthResult
	if err := c.do(ctx, "authenticate", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Validate(ctx context.Context, accessToken, clientToken string) error {
	body := map[string]any{
		"accessToken": accessToken,
		"clientToken": clientToken,
	}
	req, err := c.newRequest(ctx, "validate", body)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("no se pudo conectar con el servidor de autenticacion: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	msg := "la sesion de la cuenta no es valida"
	if resp.StatusCode == http.StatusNotFound {
		msg = "el endpoint de validacion no existe en este servidor de auth"
	}
	return &AuthError{Message: msg, StatusCode: resp.StatusCode}
}

func (c *Client) Refresh(ctx context.Context, accessToken, clientToken string, prof *Profile) (*AuthResult, error) {
	body := map[string]any{
		"accessToken": accessToken,
		"clientToken": clientToken,
		"requestUser": true,
	}
	if prof != nil {
		body["selectedProfile"] = prof
	}
	var out AuthResult
	if err := c.do(ctx, "refresh", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Invalidate(ctx context.Context, accessToken, clientToken string) error {
	body := map[string]any{
		"accessToken": accessToken,
		"clientToken": clientToken,
	}
	req, err := c.newRequest(ctx, "invalidate", body)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("no se pudo conectar con el servidor de autenticacion: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return nil
}

func (c *Client) do(ctx context.Context, endpoint string, body map[string]any, out *AuthResult) error {
	req, err := c.newRequest(ctx, endpoint, body)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("no se pudo conectar con el servidor de autenticacion: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("no se pudo leer la respuesta del servidor: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeYggError(raw, resp.StatusCode)
	}
	if len(raw) == 0 {
		return fmt.Errorf("el servidor de autenticacion respondio vacio")
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("respuesta del servidor invalida: %w", err)
	}
	return nil
}

func (c *Client) newRequest(ctx context.Context, endpoint string, body map[string]any) (*http.Request, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	u := c.baseURL + "/" + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

type yggError struct {
	Error        string `json:"error"`
	ErrorMessage string `json:"errorMessage"`
	Cause        string `json:"cause"`
}

func decodeYggError(body []byte, status int) error {
	var likely yggError
	if json.Unmarshal(body, &likely) == nil && likely.ErrorMessage != "" {
		return &AuthError{
			Message:    likely.ErrorMessage,
			RawError:   likely.Error,
			Cause:      likely.Cause,
			StatusCode: status,
		}
	}
	return &AuthError{
		Message:    readableStatus(status),
		StatusCode: status,
	}
}

func readableStatus(status int) string {
	switch status {
	case 400:
		return "peticion invalida (400)"
	case 401:
		return "credenciales incorrectas (401)"
	case 403:
		return "esta prohibido el login (403)"
	case 429:
		return "demasiadas peticiones al servidor de auth (429)"
	default:
		return "el servidor de autenticacion devolvio un error"
	}
}

type PlayerProfile struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Properties []ProfileProperty `json:"properties"`
}

type ProfileProperty struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Signature string `json:"signature,omitempty"`
}

type TexturesPayload struct {
	Timestamp   int64         `json:"timestamp"`
	ProfileID   string        `json:"profileId"`
	ProfileName string        `json:"profileName"`
	Textures    TexturesEntry `json:"textures"`
}

type TexturesEntry struct {
	SKIN SkinTexture `json:"SKIN,omitempty"`
	CAPE SkinTexture `json:"CAPE,omitempty"`
}

type SkinTexture struct {
	URL      string       `json:"url"`
	Metadata *TextureMeta `json:"metadata,omitempty"`
}

type TextureMeta struct {
	Model string `json:"model"`
}

func (c *Client) Profile(ctx context.Context, uuid string) (*PlayerProfile, error) {
	if strings.TrimSpace(uuid) == "" {
		return nil, fmt.Errorf("el UUID del jugador es obligatorio")
	}
	u := fmt.Sprintf("%s/sessionserver/session/minecraft/profile/%s?unsigned=false", c.root, uuid)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener el perfil del jugador: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el perfil del jugador: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusNoContent {
		return nil, fmt.Errorf("el servidor de autenticacion no tiene perfil para el UUID %s", uuid)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, decodeYggError(raw, resp.StatusCode)
	}
	var out PlayerProfile
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("perfil del jugador invalido: %w", err)
	}
	return &out, nil
}

func (p *PlayerProfile) Textures() (*TexturesPayload, error) {
	for _, prop := range p.Properties {
		if prop.Name != "textures" || prop.Value == "" {
			continue
		}
		raw, err := base64.StdEncoding.DecodeString(prop.Value)
		if err != nil {
			return nil, fmt.Errorf("propiedad textures del perfil invalida: %w", err)
		}
		var payload TexturesPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, fmt.Errorf("payload de texturas invalido: %w", err)
		}
		return &payload, nil
	}
	return nil, fmt.Errorf("el perfil del jugador no tiene texturas")
}

func fetchImageBytes(ctx context.Context, imageURL string) ([]byte, string, error) {
	if strings.TrimSpace(imageURL) == "" {
		return nil, "", fmt.Errorf("la URL de la imagen esta vacia")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "StepLauncher/2.3.1")
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("no se pudo descargar la imagen: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("la descarga de la imagen respondio %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil {
		return nil, "", fmt.Errorf("no se pudo leer la imagen: %w", err)
	}
	ctype := resp.Header.Get("Content-Type")
	if ctype == "" || !strings.HasPrefix(ctype, "image/") {
		ctype = "image/png"
	}
	return raw, ctype, nil
}

func FetchImageDataURL(ctx context.Context, imageURL string) (string, error) {
	raw, ctype, err := fetchImageBytes(ctx, imageURL)
	if err != nil {
		return "", err
	}
	return "data:" + ctype + ";base64," + base64.StdEncoding.EncodeToString(raw), nil
}

func FetchSkinWithHead(ctx context.Context, imageURL string) (skinDataURL, headDataURL string, err error) {
	raw, ctype, err := fetchImageBytes(ctx, imageURL)
	if err != nil {
		return "", "", err
	}
	skinDataURL = "data:" + ctype + ";base64," + base64.StdEncoding.EncodeToString(raw)
	headPNG, herr := extractHeadPNG(raw)
	if herr != nil {
		return skinDataURL, "", nil
	}
	return skinDataURL, "data:image/png;base64," + base64.StdEncoding.EncodeToString(headPNG), nil
}

func extractHeadPNG(skin []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(skin))
	if err != nil {
		return nil, fmt.Errorf("la skin no es una imagen valida: %w", err)
	}
	b := img.Bounds()
	if b.Dx() < 16 || b.Dy() < 16 {
		return nil, fmt.Errorf("la skin tiene dimensiones demasiado pequenas (%dx%d)", b.Dx(), b.Dy())
	}
	face := cropSkin(img, image.Rect(8, 8, 16, 16))
	out := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	drawHead(out, face, false)
	if b.Dx() >= 48 && b.Dy() >= 16 {
		hat := cropSkin(img, image.Rect(40, 8, 48, 16))
		drawHead(out, hat, true)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func cropSkin(img image.Image, r image.Rectangle) *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			c := color.NRGBAModel.Convert(img.At(r.Min.X+x, r.Min.Y+y)).(color.NRGBA)
			out.SetNRGBA(x, y, c)
		}
	}
	return out
}

func drawHead(dst *image.NRGBA, sprite *image.NRGBA, overlay bool) {
	const scale = 8
	for y := 0; y < sprite.Bounds().Dy(); y++ {
		for x := 0; x < sprite.Bounds().Dx(); x++ {
			c := sprite.NRGBAAt(x, y)
			if overlay && c.A == 0 {
				continue
			}
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					dst.SetNRGBA(x*scale+dx, y*scale+dy, c)
				}
			}
		}
	}
}
