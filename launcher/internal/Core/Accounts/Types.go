package accounts

type AccountType string

const (
	TypeOffline AccountType = "offline"
	TypeAuthLib AccountType = "authlib"
)

type Account struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Type             AccountType       `json:"type"`
	Username         string            `json:"username"`
	Email            string            `json:"email,omitempty"`
	UUID             string            `json:"uuid"`
	AccessToken      string            `json:"accessToken"`
	AuthServerURL    string            `json:"authServerUrl,omitempty"`
	ServerName       string            `json:"serverName,omitempty"`
	SessionValid     bool              `json:"sessionValid,omitempty"`
	CreatedAt        string            `json:"createdAt"`
	LastUsed         string            `json:"lastUsed,omitempty"`
	CustomProperties map[string]string `json:"customProperties,omitempty"`
}

type AccountsFile struct {
	Accounts        map[string]*Account `json:"accounts"`
	SelectedAccount string              `json:"selectedAccount,omitempty"`
	ClientToken     string              `json:"clientToken,omitempty"`
	AutoRefresh     bool                `json:"autoRefresh,omitempty"`
}

type AccountCredentials struct {
	Username    string `json:"username"`
	UUID        string `json:"uuid"`
	AccessToken string `json:"accessToken"`
	UserType    string `json:"userType,omitempty"`
}

type AccountInfo struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Type          AccountType       `json:"type"`
	Username      string            `json:"username"`
	UUID          string            `json:"uuid"`
	AuthServerURL string            `json:"authServerUrl,omitempty"`
	ServerName    string            `json:"serverName,omitempty"`
	HasToken      bool              `json:"hasToken"`
	SessionValid  bool              `json:"sessionValid"`
	CreatedAt     string            `json:"createdAt"`
	LastUsed      string            `json:"lastUsed,omitempty"`
	Custom        map[string]string `json:"customProperties,omitempty"`
}

type AccountAssets struct {
	Uuid          string `json:"uuid"`
	Name          string `json:"name"`
	SkinURL       string `json:"skinUrl,omitempty"`
	CapeURL       string `json:"capeUrl,omitempty"`
	SkinDataURL   string `json:"skinDataUrl,omitempty"`
	CapeDataURL   string `json:"capeDataUrl,omitempty"`
	AvatarDataURL string `json:"avatarDataUrl,omitempty"`
	Slim          bool   `json:"slim"`
}

type CreateAccountReq struct {
	Type          AccountType `json:"type"`
	Name          string      `json:"name,omitempty"`
	Username      string      `json:"username"`
	AccessToken   string      `json:"accessToken,omitempty"`
	AuthServerURL string      `json:"authServerUrl,omitempty"`
	UUID          string      `json:"uuid,omitempty"`
}

type AuthlibLoginReq struct {
	AuthServerURL string `json:"authServerUrl"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}
