package account

import (
	"StepLauncher/internal/Handlers"
	engine "StepLauncher/internal/Handlers/Engine"
	"errors"
)

// AccountService gestiona cuentas, perfiles y Rich Presence.
type AccountService struct {
	handler *Handlers.App
	engine  *engine.Engine
}

func NewAccountService(handler *Handlers.App, eng *engine.Engine) *AccountService {
	return &AccountService{handler: handler, engine: eng}
}

func (s *AccountService) CancelAuthlibLogin() {
	if s.engine != nil {
		s.engine.CancelLogin()
	}
}

func (s *AccountService) CreateAccount(req engine.CreateAccountReq) (*engine.AccountInfo, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.CreateAccount(req)
}

func (s *AccountService) CreateProfile(p *engine.Profile) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.CreateProfile(p)
}

func (s *AccountService) DeleteAccount(id string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.DeleteAccount(id)
}

func (s *AccountService) DeleteProfile(name string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.DeleteProfile(name)
}

func (s *AccountService) GetAccount(id string) (*engine.AccountInfo, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.GetAccount(id)
}

func (s *AccountService) GetAccountAssets(id string) {
	if s.engine != nil {
		s.engine.GetAccountAssets(id)
	}
}

func (s *AccountService) GetAccountsAutoRefresh() bool {
	if s.engine == nil {
		return false
	}
	return s.engine.GetAccountsAutoRefresh()
}

func (s *AccountService) GetProfile(name string) (*engine.Profile, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.GetProfile(name)
}

func (s *AccountService) GetRichPresenceConfig() bool {
	if s.handler == nil {
		return true
	}
	return s.handler.GetRichPresenceConfig()
}

func (s *AccountService) GetSelectedAccount() string {
	if s.engine == nil {
		return ""
	}
	return s.engine.GetSelectedAccount()
}

func (s *AccountService) GetSelectedProfile() string {
	if s.engine == nil {
		return ""
	}
	return s.engine.GetSelectedProfile()
}

func (s *AccountService) GetSelectedVersion() string {
	if s.engine == nil {
		return ""
	}
	return s.engine.GetSelectedVersion()
}

func (s *AccountService) ListAccounts() []engine.AccountInfo {
	if s.engine == nil {
		return []engine.AccountInfo{}
	}
	return s.engine.ListAccounts()
}

func (s *AccountService) ListProfiles() map[string]*engine.Profile {
	if s.engine == nil {
		return map[string]*engine.Profile{}
	}
	return s.engine.ListProfiles()
}

func (s *AccountService) LoginAuthlib(req engine.AuthlibLoginReq) {
	if s.engine != nil {
		s.engine.LoginAuthlib(req)
	}
}

func (s *AccountService) RefreshAccount(id string) {
	if s.engine != nil {
		s.engine.RefreshAccount(id)
	}
}

func (s *AccountService) RefreshAllAccounts() int {
	if s.engine == nil {
		return 0
	}
	return s.engine.RefreshAllAccounts()
}

func (s *AccountService) ResolveAccountCredentials(id string) (*engine.AccountCredentials, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.ResolveAccountCredentials(id)
}

func (s *AccountService) SetAccountsAutoRefresh(v bool) {
	if s.engine != nil {
		s.engine.SetAccountsAutoRefresh(v)
	}
}

func (s *AccountService) SetRichPresenceEnabled(v bool) {
	s.handler.SetRichPresenceEnabled(v)
}

func (s *AccountService) SetSelectedAccount(id string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.SetSelectedAccount(id)
}

func (s *AccountService) SetSelectedProfile(name string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.SetSelectedProfile(name)
}

func (s *AccountService) SetSelectedVersion(version string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.SetSelectedVersion(version)
}

func (s *AccountService) UpdateAccount(id string, req engine.CreateAccountReq) (*engine.AccountInfo, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.UpdateAccount(id, req)
}

func (s *AccountService) UpdateProfile(name string, p *engine.Profile) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.UpdateProfile(name, p)
}
