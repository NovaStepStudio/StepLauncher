package engine

import (
	"errors"

	"StepLauncher/internal/Core/Accounts"
	"StepLauncher/internal/Core/Launcher"
)

var errAccountsUnavailable = errors.New("cuentas no disponibles")

type Account = accounts.Account
type AccountInfo = accounts.AccountInfo
type AccountCredentials = accounts.AccountCredentials
type CreateAccountReq = accounts.CreateAccountReq
type AuthlibLoginReq = accounts.AuthlibLoginReq

func (e *Engine) ListAccounts() []AccountInfo {
	if e.accounts == nil {
		return []AccountInfo{}
	}
	return e.accounts.ListInfo()
}

func (e *Engine) GetAccount(id string) (*AccountInfo, error) {
	if e.accounts == nil {
		return nil, errAccountsUnavailable
	}
	return e.accounts.GetInfo(id)
}

func (e *Engine) CreateAccount(req CreateAccountReq) (*AccountInfo, error) {
	if e.accounts == nil {
		return nil, errAccountsUnavailable
	}
	return e.accounts.Create(req)
}

func (e *Engine) UpdateAccount(id string, req CreateAccountReq) (*AccountInfo, error) {
	if e.accounts == nil {
		return nil, errAccountsUnavailable
	}
	return e.accounts.Update(id, req)
}

func (e *Engine) DeleteAccount(id string) error {
	if e.accounts == nil {
		return errAccountsUnavailable
	}
	return e.accounts.Delete(id)
}

func (e *Engine) GetSelectedAccount() string {
	if e.accounts == nil {
		return ""
	}
	return e.accounts.Selected()
}

func (e *Engine) SetSelectedAccount(id string) error {
	if e.accounts == nil {
		return errAccountsUnavailable
	}
	return e.accounts.SetSelected(id)
}

func (e *Engine) ResolveAccountCredentials(id string) (*AccountCredentials, error) {
	if e.accounts == nil {
		return nil, errAccountsUnavailable
	}
	creds, _, err := e.accounts.ResolveCredentials(id)
	if err != nil {
		return nil, err
	}
	return creds, nil
}

func (e *Engine) LoginAuthlib(req AuthlibLoginReq) {
	if e.accounts == nil {
		return
	}
	go e.accounts.LoginAuthLib(req)
}

func (e *Engine) RefreshAccount(id string) {
	if e.accounts == nil {
		return
	}
	go e.accounts.RefreshAuthLib(id)
}

func (e *Engine) CancelLogin() {
	if e.accounts == nil {
		return
	}
	e.accounts.CancelLogin()
}

func (e *Engine) RefreshAllAccounts() int {
	if e.accounts == nil {
		return 0
	}
	count := 0
	for _, a := range e.accounts.ListInfo() {
		if a.Type == accounts.TypeAuthLib {
			count++
		}
	}
	go e.accounts.RefreshAll()
	return count
}

func (e *Engine) SetAccountsAutoRefresh(v bool) {
	if e.accounts == nil {
		return
	}
	e.accounts.SetAutoRefresh(v)
}

func (e *Engine) GetAccountsAutoRefresh() bool {
	if e.accounts == nil {
		return false
	}
	return e.accounts.GetAutoRefresh()
}

func (e *Engine) GetAccountAssets(id string) {
	if e.accounts == nil {
		return
	}
	go e.accounts.AccountAssets(id)
}

func (e *Engine) fillAccountCredentials(cfg *launcher.LaunchConfig, accountID string) error {
	if e.accounts == nil {
		return nil
	}
	creds, id, err := e.accounts.ResolveCredentials(accountID)
	if err != nil {
		return err
	}
	var accInfo *accounts.AccountInfo
	if info, err := e.accounts.GetInfo(id); err == nil {
		accInfo = info
	}

	if accInfo != nil && accInfo.Type == accounts.TypeAuthLib {
		creds, err = e.accounts.ResolveForLaunch(id)
		if err != nil {
			return err
		}
		cfg.Username = creds.Username
		cfg.UUID = creds.UUID
		cfg.AccessToken = creds.AccessToken
		adv := cfg.Adv()
		if !adv.AuthLibConfig.Enabled {
			adv.AuthLibConfig.Enabled = true
		}
		if adv.AuthLibConfig.AuthServerURL == "" {
			adv.AuthLibConfig.AuthServerURL = accInfo.AuthServerURL
		}
		if adv.UserType == "" {
			adv.UserType = creds.UserType
		}
		cfg.Advanced = &adv
		e.accounts.TouchLastUsed(id)
		return nil
	}

	if cfg.Username != "" {
		return nil
	}
	creds, err = e.accounts.ResolveForLaunch(id)
	if err != nil {
		return err
	}
	cfg.Username = creds.Username
	cfg.UUID = creds.UUID
	cfg.AccessToken = creds.AccessToken
	if cfg.Adv().UserType == "" {
		adv := cfg.Adv()
		adv.UserType = creds.UserType
		cfg.Advanced = &adv
	}
	e.accounts.TouchLastUsed(id)
	return nil
}
