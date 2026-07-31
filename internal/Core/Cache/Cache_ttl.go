package cache

import "time"

type CacheTTL struct {
	Manifest  time.Duration
	Assets    time.Duration
	Versions  time.Duration
	Modloader time.Duration
	Java      time.Duration
	Default   time.Duration
}

func DefaultTTLs() CacheTTL {
	return CacheTTL{
		Manifest:  24 * time.Hour,
		Assets:    48 * time.Hour,
		Versions:  24 * time.Hour,
		Modloader: 24 * time.Hour,
		Java:      48 * time.Hour,
		Default:   24 * time.Hour,
	}
}

func (m *Manager) ttlFor(category string) time.Duration {
	switch category {
	case "manifest":
		return m.ttls.Manifest
	case "assets":
		return m.ttls.Assets
	case "versions":
		return m.ttls.Versions
	case "modloader":
		return m.ttls.Modloader
	case "java":
		return m.ttls.Java
	default:
		return m.ttls.Default
	}
}
