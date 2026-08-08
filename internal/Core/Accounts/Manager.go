package accounts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"StepLauncher/internal/Core/Auth"
)

type Manager struct {
	mu       sync.RWMutex
	filePath string
	data     AccountsFile
	logFn    func(format string, args ...interface{})
	emitFn   func(eventType string, data []byte)

	loginCancel context.CancelFunc
}

func NewManager(workDir string) *Manager {
	return &Manager{
		filePath: filepath.Join(workDir, "launcher_accounts.json"),
		data: AccountsFile{
			Accounts: make(map[string]*Account),
		},
	}
}

func (m *Manager) SetLogFn(fn func(format string, args ...interface{})) {
	m.mu.Lock()
	m.logFn = fn
	m.mu.Unlock()
}

func (m *Manager) SetEventFn(fn func(eventType string, data []byte)) {
	m.mu.Lock()
	m.emitFn = fn
	m.mu.Unlock()
}

func (m *Manager) logf(format string, args ...interface{}) {
	m.mu.RLock()
	fn := m.logFn
	m.mu.RUnlock()
	if fn != nil {
		fn("[Accounts] "+format, args...)
	}
}

func (m *Manager) emit(event string, payload map[string]any) {
	m.mu.RLock()
	fn := m.emitFn
	m.mu.RUnlock()
	if fn == nil {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	fn(event, data)
}

func (m *Manager) Load() error {
	m.mu.Lock()
	if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
		m.mu.Unlock()
		m.logf("No existe archivo de cuentas; arrancando vacio (%s)", m.filePath)
		return nil
	}
	raw, err := os.ReadFile(m.filePath)
	if err != nil {
		m.mu.Unlock()
		m.logf("WARN: no se pudo leer el archivo de cuentas: %v", err)
		return err
	}
	if err := json.Unmarshal(raw, &m.data); err != nil {
		var legacy struct {
			Accounts json.RawMessage `json:"accounts"`
		}
		migrated := false
		if lerr := json.Unmarshal(raw, &legacy); lerr == nil && len(legacy.Accounts) > 0 {
			var fresh AccountsFile
			if mlerr := json.Unmarshal(legacy.Accounts, &fresh); mlerr == nil {
				m.data = fresh
				migrated = true
			}
		}
		if !migrated {
			m.mu.Unlock()
			m.logf("WARN: archivo de cuentas corrupto, arrancando vacio: %v", err)
			return fmt.Errorf("parse accounts: %w", err)
		}
		m.logf("Archivo de cuentas en formato legacy; se migrara a la raiz")
	}
	m.sanitizeLoaded()
	m.mu.Unlock()
	if err := m.persist(); err != nil {
		m.logf("WARN: no se pudo normalizar el archivo de cuentas: %v", err)
	}
	m.logf("Cuentas cargadas: %d | seleccionada=%s | autoRefresh=%v",
		len(m.data.Accounts), m.data.SelectedAccount, m.data.AutoRefresh)
	return nil
}

func (m *Manager) sanitizeLoaded() {
	m.data.Accounts = m.sanitizeMap(m.data.Accounts)
	if _, ok := m.data.Accounts[m.data.SelectedAccount]; !ok {
		m.data.SelectedAccount = ""
	}
	if m.data.ClientToken == "" {
		m.data.ClientToken = newClientToken()
	}
}

func (m *Manager) sanitizeMap(in map[string]*Account) map[string]*Account {
	out := make(map[string]*Account, len(in))
	for id, a := range in {
		if a == nil {
			continue
		}
		a.ID = id
		a.Name = sanitizeName(a.Name)
		a.Type = NormalizeAccountType(string(a.Type))
		if a.Type == "" {
			continue
		}
		a.Username = strings.TrimSpace(a.Username)
		if a.Username == "" {
			continue
		}
		if a.Type == TypeOffline {
			a.UUID = OfflineUUID(a.Username)
			a.AccessToken = ""
		}
		a.AccessToken = normalizeAccessToken(a.AccessToken)
		if a.UUID == "" {
			a.UUID = OfflineUUID(a.Username)
		}
		if a.CreatedAt == "" {
			a.CreatedAt = time.Now().UTC().Format(time.RFC3339)
		}
		out[id] = a
	}
	return out
}

func (m *Manager) persist() error {
	if m.filePath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(m.filePath), 0755); err != nil {
		return err
	}
	m.mu.RLock()
	section := m.data
	m.mu.RUnlock()

	out, err := json.MarshalIndent(section, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.filePath, out, 0644)
}

func (m *Manager) Save() error {
	return m.persist()
}

func (m *Manager) ListInfo() []AccountInfo {
	m.mu.RLock()
	list := make([]AccountInfo, 0, len(m.data.Accounts))
	for _, a := range m.data.Accounts {
		list = append(list, a.Info())
	}
	m.mu.RUnlock()
	return list
}

func (m *Manager) List() map[string]AccountCredentials {
	m.mu.RLock()
	out := make(map[string]AccountCredentials, len(m.data.Accounts))
	for id, a := range m.data.Accounts {
		out[id] = AccountCredentials{Username: a.Username, UUID: a.UUID, UserType: UserTypeOf(a.Type)}
	}
	m.mu.RUnlock()
	return out
}

func (m *Manager) GetInfo(id string) (*AccountInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.data.Accounts[id]
	if !ok {
		return nil, fmt.Errorf("cuenta %q no encontrada", id)
	}
	info := a.Info()
	return &info, nil
}

func (m *Manager) Get(id string) (*Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.data.Accounts[id]
	if !ok {
		return nil, fmt.Errorf("cuenta %q no encontrada", id)
	}
	cp := *a
	return &cp, nil
}

func (m *Manager) Create(req CreateAccountReq) (*AccountInfo, error) {
	if errs := req.errors(); len(errs) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	typ := NormalizeAccountType(string(req.Type))
	id := fmt.Sprintf("acc-%d", time.Now().UnixNano())
	a, err := m.apply(req)
	if err != nil {
		return nil, err
	}
	a.ID = id
	a.Type = typ
	a.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	if a.Name == "" {
		a.Name = a.Username
	}
	m.mu.Lock()
	m.data.Accounts[id] = a
	if m.data.SelectedAccount == "" {
		m.data.SelectedAccount = id
	}
	m.mu.Unlock()
	if err := m.persist(); err != nil {
		m.logf("WARN: no se pudo persistir al crear la cuenta %q: %v", id, err)
		return nil, err
	}
	info := a.Info()
	m.logf("Cuenta creada: %s (%s) usuario=%s", id, typ, a.Username)
	return &info, nil
}

func (m *Manager) Update(id string, req CreateAccountReq) (*AccountInfo, error) {
	m.mu.RLock()
	existing, ok := m.data.Accounts[id]
	if !ok {
		m.mu.RUnlock()
		return nil, fmt.Errorf("cuenta %q no encontrada", id)
	}
	createdAt := existing.CreatedAt
	typ := existing.Type
	m.mu.RUnlock()

	if errs := req.errors(); len(errs) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	a, err := m.apply(req)
	if err != nil {
		return nil, err
	}
	a.ID = id
	a.Type = typ
	a.CreatedAt = createdAt

	m.mu.Lock()
	m.data.Accounts[id] = a
	m.mu.Unlock()
	if err := m.persist(); err != nil {
		m.logf("WARN: no se pudo persistir al actualizar la cuenta %q: %v", id, err)
		return nil, err
	}
	m.logf("Cuenta actualizada: %s (%s) usuario=%s", id, typ, a.Username)
	info := a.Info()
	return &info, nil
}

func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	if _, ok := m.data.Accounts[id]; !ok {
		m.mu.Unlock()
		return fmt.Errorf("cuenta %q no encontrada", id)
	}
	delete(m.data.Accounts, id)
	if m.data.SelectedAccount == id {
		m.data.SelectedAccount = ""
	}
	m.mu.Unlock()
	if err := m.persist(); err != nil {
		m.logf("WARN: no se pudo persistir al eliminar la cuenta %q: %v", id, err)
		return err
	}
	m.logf("Cuenta eliminada: %s", id)
	return nil
}

func (m *Manager) apply(req CreateAccountReq) (*Account, error) {
	username := strings.TrimSpace(req.Username)
	a := &Account{
		Type:          NormalizeAccountType(string(req.Type)),
		Name:          sanitizeName(req.Name),
		Username:      username,
		AccessToken:   normalizeAccessToken(req.AccessToken),
		AuthServerURL: strings.TrimSpace(req.AuthServerURL),
	}
	switch a.Type {
	case TypeOffline:
		a.UUID = OfflineUUID(username)
	case TypeAuthLib:
		a.UUID = strings.TrimSpace(req.UUID)
		if a.UUID == "" {
			a.UUID = OfflineUUID(username)
		}
	}
	if a.Name == "" {
		a.Name = username
	}
	return a, nil
}

func (a *Account) Info() AccountInfo {
	return AccountInfo{
		ID:            a.ID,
		Name:          a.Name,
		Type:          a.Type,
		Username:      a.Username,
		UUID:          a.UUID,
		AuthServerURL: a.AuthServerURL,
		ServerName:    a.ServerName,
		HasToken:      a.AccessToken != "",
		SessionValid:  a.SessionValid,
		CreatedAt:     a.CreatedAt,
		LastUsed:      a.LastUsed,
		Custom:        a.CustomProperties,
	}
}

func (m *Manager) Selected() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sel := m.data.SelectedAccount
	if _, ok := m.data.Accounts[sel]; !ok {
		sel = ""
	}
	return sel
}

func (m *Manager) SetSelected(id string) error {
	m.mu.Lock()
	if id != "" {
		if _, ok := m.data.Accounts[id]; !ok {
			m.mu.Unlock()
			return fmt.Errorf("cuenta %q no encontrada", id)
		}
	}
	m.data.SelectedAccount = id
	m.mu.Unlock()
	if err := m.persist(); err != nil {
		m.logf("WARN: no se pudo persistir la seleccion de cuenta: %v", err)
		return err
	}
	m.logf("Cuenta seleccionada: %q", id)
	return nil
}

func (m *Manager) TouchLastUsed(id string) {
	m.mu.Lock()
	if a, ok := m.data.Accounts[id]; ok {
		a.LastUsed = time.Now().UTC().Format(time.RFC3339)
	}
	m.mu.Unlock()
	m.persist()
}

func (m *Manager) ResolveCredentials(id string) (*AccountCredentials, string, error) {
	m.mu.RLock()
	target := id
	if target == "" {
		target = m.data.SelectedAccount
	}
	if _, ok := m.data.Accounts[target]; !ok {
		var first string
		for k := range m.data.Accounts {
			first = k
			break
		}
		target = first
	}
	if target == "" {
		m.mu.RUnlock()
		return nil, "", fmt.Errorf("no hay cuentas configuradas")
	}
	a := m.data.Accounts[target]
	creds := &AccountCredentials{
		Username:    a.Username,
		UUID:        a.UUID,
		AccessToken: a.AccessToken,
		UserType:    UserTypeOf(a.Type),
	}
	m.mu.RUnlock()
	return creds, target, nil
}

func (m *Manager) ResolveForLaunch(id string) (*AccountCredentials, error) {
	creds, target, err := m.ResolveCredentials(id)
	if err != nil {
		return nil, err
	}
	m.mu.RLock()
	a := m.data.Accounts[target]
	auto := m.data.AutoRefresh
	m.mu.RUnlock()
	if a == nil || a.Type != TypeAuthLib {
		return creds, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	client, err := authlib.NewResolved(ctx, a.AuthServerURL)
	if err != nil {
		m.logf("WARN: URL de auth invalida en %q; lanzando con token guardado", a.Username)
		return creds, nil
	}

	vErr := client.Validate(ctx, a.AccessToken, m.ClientToken())
	if vErr == nil {
		m.markSession(target, true)
		return creds, nil
	}
	if !isAuthError(vErr) {
		m.logf("WARN: no se pudo validar la sesion de %q (red); lanzando con token guardado", a.Username)
		return creds, nil
	}

	if !auto {
		m.markSession(target, false)
		return nil, fmt.Errorf("la sesion de %q expiro. Abre Cuentas e inicia sesion de nuevo", a.Username)
	}
	res, rErr := client.Refresh(ctx, a.AccessToken, m.ClientToken(), &authlib.Profile{ID: a.UUID, Name: a.Username})
	if rErr != nil {
		m.markSession(target, false)
		return nil, fmt.Errorf("la sesion de %q expiro y no se pudo renovar: %v", a.Username, rErr)
	}
	m.mu.Lock()
	if cur := m.data.Accounts[target]; cur != nil {
		cur.AccessToken = normalizeAccessToken(res.AccessToken)
		cur.SessionValid = true
		if res.SelectedProfile.ID != "" {
			cur.UUID = res.SelectedProfile.ID
		}
		if res.SelectedProfile.Name != "" {
			cur.Username = res.SelectedProfile.Name
		}
	}
	m.mu.Unlock()
	m.persist()
	creds.AccessToken = normalizeAccessToken(res.AccessToken)
	m.clearSkinCache(target)
	m.logf("Sesion renovada automaticamente para %q", target)
	return creds, nil
}

func (m *Manager) markSession(id string, valid bool) {
	m.mu.Lock()
	if a, ok := m.data.Accounts[id]; ok {
		a.SessionValid = valid
	}
	m.mu.Unlock()
}

func (m *Manager) fillServerName(id, name string) {
	if name == "" {
		return
	}
	m.mu.Lock()
	if cur := m.data.Accounts[id]; cur != nil && cur.ServerName == "" {
		cur.ServerName = name
	}
	m.mu.Unlock()
	m.persist()
}

func isAuthError(err error) bool {
	var ae *authlib.AuthError
	return errors.As(err, &ae)
}

func (m *Manager) ClientToken() string {
	m.mu.Lock()
	ct := m.data.ClientToken
	needPersist := false
	if ct == "" {
		ct = newClientToken()
		m.data.ClientToken = ct
		needPersist = true
	}
	m.mu.Unlock()
	if needPersist {
		if err := m.persist(); err != nil {
			m.logf("WARN: no se pudo persistir el clientToken: %v", err)
		}
	}
	return ct
}

func (m *Manager) LoginAuthLib(req AuthlibLoginReq) (*AccountInfo, error) {
	if errs := req.errorsLogin(); len(errs) > 0 {
		err := fmt.Errorf("%s", strings.Join(errs, "; "))
		m.emit("account_login", map[string]any{"ok": false, "error": err.Error()})
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	m.mu.Lock()
	m.loginCancel = cancel
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		m.loginCancel = nil
		m.mu.Unlock()
		cancel()
	}()

	client, err := authlib.NewResolved(ctx, req.AuthServerURL)
	if err != nil {
		m.emit("account_login", map[string]any{"ok": false, "error": err.Error()})
		return nil, err
	}

	login := strings.TrimSpace(req.Username)
	res, err := client.Authenticate(ctx, login, req.Password, m.ClientToken())
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			m.emit("account_login", map[string]any{"ok": false, "error": err.Error()})
		}
		return nil, err
	}

	playerName := res.SelectedProfile.Name
	if playerName == "" && len(res.AvailableProfiles) > 0 {
		playerName = res.AvailableProfiles[0].Name
	}
	if playerName == "" {
		playerName = sanitizeName(login)
	}
	uuid := strings.ToLower(strings.TrimSpace(res.SelectedProfile.ID))
	acc := &Account{
		Type:          TypeAuthLib,
		Name:          sanitizeName(login),
		Username:      playerName,
		Email:         login,
		UUID:          uuid,
		AccessToken:   normalizeAccessToken(res.AccessToken),
		AuthServerURL: req.AuthServerURL,
		SessionValid:  true,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
	}
	if sn := client.AuthServerName(ctx); sn != "" {
		acc.ServerName = sn
	}

	id := fmt.Sprintf("acc-%d", time.Now().UnixNano())
	updated := false
	m.mu.Lock()
	for k, existing := range m.data.Accounts {
		if existing == nil || existing.Type != TypeAuthLib {
			continue
		}
		if strings.EqualFold(existing.Email, login) &&
			strings.EqualFold(strings.TrimRight(existing.AuthServerURL, "/"), strings.TrimRight(req.AuthServerURL, "/")) {
			id = k
			acc.ID = k
			acc.CreatedAt = existing.CreatedAt
			updated = true
			break
		}
	}
	acc.ID = id
	m.data.Accounts[id] = acc
	if m.data.SelectedAccount == "" {
		m.data.SelectedAccount = id
	}
	m.mu.Unlock()

if err := m.persist(); err != nil {
		m.logf("WARN: no se pudo persistir la sesion de %q: %v", playerName, err)
	}
	m.clearSkinCache(id)
	info := acc.Info()
	m.emit("account_login", map[string]any{"ok": true, "account": info, "updated": updated})
	m.logf("Login AuthLib exitoso: %s (%s) en %s", playerName, id, req.AuthServerURL)
	return &info, nil
}

func (m *Manager) CancelLogin() {
	m.mu.Lock()
	cancel := m.loginCancel
	m.loginCancel = nil
	m.mu.Unlock()
	if cancel != nil {
		cancel()
		m.emit("account_login", map[string]any{"ok": false, "error": "inicio de sesion cancelado"})
		m.logf("Login AuthLib cancelado por el usuario")
	}
}

func (m *Manager) RefreshAuthLib(id string) error {
	m.mu.RLock()
	a, ok := m.data.Accounts[id]
	m.mu.RUnlock()
	if !ok {
		err := fmt.Errorf("cuenta %q no encontrada", id)
		m.emit("account_refresh", map[string]any{"id": id, "ok": false, "error": err.Error()})
		return err
	}
	if a.Type != TypeAuthLib {
		err := fmt.Errorf("la cuenta %q no es AuthLib", id)
		m.emit("account_refresh", map[string]any{"id": id, "ok": false, "error": err.Error()})
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	ct := m.ClientToken()

	client, err := authlib.NewResolved(ctx, a.AuthServerURL)
	if err != nil {
		m.emit("account_refresh", map[string]any{"id": id, "ok": false, "error": err.Error()})
		return err
	}

	serverName := ""
	if a.ServerName == "" && ctx.Err() == nil {
		serverName = client.AuthServerName(ctx)
	}

	vErr := client.Validate(ctx, a.AccessToken, ct)
	if vErr == nil {
		m.markSession(id, true)
		m.persist()
		m.fillServerName(id, serverName)
		m.emit("account_refresh", map[string]any{"id": id, "ok": true, "renewed": false})
		m.clearSkinCache(id)
		go m.AccountAssets(id)
		m.logf("Sesion validada para %q", a.Username)
		return nil
	}
	if !isAuthError(vErr) {
		m.emit("account_refresh", map[string]any{"id": id, "ok": true, "renewed": false, "warning": vErr.Error()})
		return nil
	}

	res, rErr := client.Refresh(ctx, a.AccessToken, ct, &authlib.Profile{ID: a.UUID, Name: a.Username})
	if rErr != nil {
		m.markSession(id, false)
		m.persist()
		m.emit("account_refresh", map[string]any{"id": id, "ok": false, "error": rErr.Error()})
		m.logf("WARN: no se pudo renovar la sesion de %q: %v", a.Username, rErr)
		return rErr
	}
	m.mu.Lock()
	if cur := m.data.Accounts[id]; cur != nil {
		cur.AccessToken = normalizeAccessToken(res.AccessToken)
		cur.SessionValid = true
		if res.SelectedProfile.ID != "" {
			cur.UUID = res.SelectedProfile.ID
		}
		if res.SelectedProfile.Name != "" {
			cur.Username = res.SelectedProfile.Name
		}
	}
	m.mu.Unlock()
	m.persist()
	m.fillServerName(id, serverName)
	m.emit("account_refresh", map[string]any{"id": id, "ok": true, "renewed": true})
	m.clearSkinCache(id)
	go m.AccountAssets(id)
	m.logf("Sesion renovada para %q", a.Username)
	return nil
}

func (m *Manager) RefreshAll() {
	m.mu.RLock()
	auto := m.data.AutoRefresh
	ids := make([]string, 0)
	for id, a := range m.data.Accounts {
		if a != nil && a.Type == TypeAuthLib {
			ids = append(ids, id)
		}
	}
	m.mu.RUnlock()

	if !auto {
		m.emit("account_refresh_all", map[string]any{"started": 0, "ok": 0, "failed": 0, "skipped": true})
		m.logf("AutoRefresh desactivado; no se renuevan sesiones")
		return
	}
	ok := 0
	failed := 0
	for _, id := range ids {
		if err := m.RefreshAuthLib(id); err != nil {
			failed++
		} else {
			ok++
		}
	}
	m.emit("account_refresh_all", map[string]any{"started": len(ids), "ok": ok, "failed": failed, "skipped": false})
	m.logf("AutoRefresh terminado: %d ok, %d fallos de %d", ok, failed, len(ids))
}

func (m *Manager) SetAutoRefresh(v bool) {
	m.mu.Lock()
	m.data.AutoRefresh = v
	m.mu.Unlock()
	m.persist()
	m.logf("AutoRefresh de sesiones: %v", v)
}

func (m *Manager) GetAutoRefresh() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data.AutoRefresh
}

func (m *Manager) AccountAssets(id string) *AccountAssets {
	m.mu.RLock()
	a, ok := m.data.Accounts[id]
	m.mu.RUnlock()
	if !ok {
		m.emit("account_assets", map[string]any{"id": id, "ok": false, "error": fmt.Sprintf("cuenta %q no encontrada", id)})
		return nil
	}
	if a.Type != TypeAuthLib {
		m.emit("account_assets", map[string]any{"id": id, "ok": false, "error": "la cuenta no es AuthLib"})
		return nil
	}

	if skinData, headData, ok := m.readSkinCache(id); ok {
		assets := &AccountAssets{Uuid: a.UUID, Name: a.Username, SkinDataURL: skinData, AvatarDataURL: headData}
		m.emit("account_assets", map[string]any{"id": id, "ok": true, "assets": assets, "cached": true})
		m.logf("Assets de skin servidos desde cache para %q", a.Username)
		return assets
	}

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	client, err := authlib.NewResolved(ctx, a.AuthServerURL)
	if err != nil {
		m.emit("account_assets", map[string]any{"id": id, "ok": false, "error": err.Error()})
		return nil
	}
	prof, err := client.Profile(ctx, a.UUID)
	if err != nil {
		m.emit("account_assets", map[string]any{"id": id, "ok": false, "error": err.Error()})
		return nil
	}
	payload, err := prof.Textures()
	if err != nil {
		m.emit("account_assets", map[string]any{"id": id, "ok": true, "assets": &AccountAssets{Uuid: a.UUID, Name: prof.Name}})
		return nil
	}

	assets := &AccountAssets{Uuid: a.UUID, Name: prof.Name}
	if payload.Textures.SKIN.URL != "" {
		assets.SkinURL = payload.Textures.SKIN.URL
		if skinData, headData, derr := authlib.FetchSkinWithHead(ctx, payload.Textures.SKIN.URL); derr == nil {
			assets.SkinDataURL = skinData
			assets.AvatarDataURL = headData
		}
		if payload.Textures.SKIN.Metadata != nil {
			assets.Slim = payload.Textures.SKIN.Metadata.Model == "slim"
		}
	}
	if payload.Textures.CAPE.URL != "" {
		assets.CapeURL = payload.Textures.CAPE.URL
		if data, derr := authlib.FetchImageDataURL(ctx, payload.Textures.CAPE.URL); derr == nil {
			assets.CapeDataURL = data
		}
	}
	m.writeSkinCache(id, assets.SkinDataURL, assets.AvatarDataURL)
	m.emit("account_assets", map[string]any{"id": id, "ok": true, "assets": assets})
	m.logf("Assets de skin obtenidos para %q", a.Username)
	return assets
}

func (m *Manager) skinCacheDir() string {
	return filepath.Join(filepath.Dir(m.filePath), "cache", "avatars")
}

func (m *Manager) readSkinCache(id string) (skinData, headData string, ok bool) {
	if id == "" {
		return "", "", false
	}
	raw, err := os.ReadFile(filepath.Join(m.skinCacheDir(), id+".json"))
	if err != nil {
		return "", "", false
	}
	var e skinCacheEntry
	if err := json.Unmarshal(raw, &e); err != nil {
		return "", "", false
	}
	if e.SkinDataURL == "" || e.AvatarDataURL == "" {
		return "", "", false
	}
	return e.SkinDataURL, e.AvatarDataURL, true
}

func (m *Manager) writeSkinCache(id, skinData, headData string) error {
	if id == "" || skinData == "" || headData == "" {
		return nil
	}
	dir := m.skinCacheDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(skinCacheEntry{SkinDataURL: skinData, AvatarDataURL: headData}, "", "  ")
	if err != nil {
		return err
	}
	p := filepath.Join(dir, id+".json")
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, raw, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func (m *Manager) clearSkinCache(id string) {
	if id == "" {
		return
	}
	dir := m.skinCacheDir()
	for _, name := range []string{id + ".json", id + ".json.tmp"} {
		if err := os.Remove(filepath.Join(dir, name)); err == nil {
			m.logf("Cache de skin reestablecido para %q", id)
		}
	}
}

type skinCacheEntry struct {
	SkinDataURL   string `json:"skinDataUrl,omitempty"`
	AvatarDataURL string `json:"avatarDataUrl,omitempty"`
}
