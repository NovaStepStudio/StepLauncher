package launcher

type LaunchConfig struct {
	Version         string
	Username        string
	InstanceID      string
	InstanceName    string
	UUID            string
	AccessToken     string
	XUID            string
	ClientID        string
	LauncherName    string
	LauncherVersion string
	LogDir          string
	LogFn           func(string, ...interface{})
	Advanced        *AdvancedConfig
	Profile         string
}

func (c LaunchConfig) Adv() AdvancedConfig {
	if c.Advanced != nil {
		return *c.Advanced
	}
	return DefaultAdvancedConfig()
}
