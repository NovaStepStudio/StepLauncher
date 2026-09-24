package engine

import (
	"StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Core/Platform"
	engineconfig "StepLauncher/internal/Handlers/Engine/engineconfig"
)

func (e *Engine) Config() engineconfig.Config {
	return e.config.Get()
}

func (e *Engine) ConfigManager() *engineconfig.Manager {
	return e.config
}

func (e *Engine) TotalRAMGB() int {
	totalMB := platform.TotalRAMMB()
	if totalMB <= 0 {
		return 4
	}
	gb := int(totalMB / 1024)
	if gb < 1 {
		return 1
	}
	return gb
}

func (e *Engine) RecommendedRAMGB() int {
	total := e.TotalRAMGB()
	switch {
	case total <= 2:
		if total-1 < 1 {
			return 1
		}
		return total - 1
	case total <= 4:
		return 2
	case total <= 8:
		return 4
	case total <= 16:
		return 6
	default:
		return 8
	}
}

func (e *Engine) MaxRAMGB() int {
	cfg := e.config.Get()
	return cfg.MaxRAMMB / 1024
}

func (e *Engine) SetMaxRAM(gb int) {
	cfg := e.config.Get()
	cfg.MaxRAMMB = gb * 1024
	e.config.UpdateConfig(cfg)
}

func (e *Engine) SetConcurrentDownloads(n int) {
	cfg := e.config.Get()
	cfg.ConcurrentDownloads = n
	e.config.UpdateConfig(cfg)
	if e.downloader != nil {
		e.downloader.SetMaxConcurrency(n)
	}
	if e.sharedDl != nil {
		e.sharedDl.SetMaxConcurrency(n)
	}
}

func (e *Engine) SetMaxMbps(mbps float64) {
	cfg := e.config.Get()
	cfg.MaxMbps = mbps
	e.config.UpdateConfig(cfg)
	e.applyNetworkConfig(cfg)
}

func (e *Engine) applyNetworkConfig(cfg engineconfig.Config) {
	client, err := downloader.NewConfiguredHTTPClient(
		cfg.MaxMbps,
		cfg.ProxyEnabled,
		cfg.ProxyHost,
		cfg.ProxyPort,
		cfg.ProxyUser,
		cfg.ProxyPass,
	)
	if err != nil {
		if e.log != nil {
			e.log.Warn("No se pudo aplicar la configuración de red: %v", err)
		}
		return
	}
	if e.downloader != nil {
		e.downloader.SetHTTPClient(client)
	}
	if e.sharedDl != nil {
		e.sharedDl.SetHTTPClient(client)
	}
	// Los providers de modloaders guardan su propio cliente: propagarlo para
	// que el proxy/límite nuevo aplique a los metadatos sin reiniciar.
	// (nil durante la inicialización: el orquestador aún no existe.)
	if e.modloader != nil {
		e.modloader.SetHTTPClient(client)
	}
}

func (e *Engine) SetVerifyIntegrity(v bool) {
	cfg := e.config.Get()
	cfg.VerifyIntegrity = v
	e.config.UpdateConfig(cfg)
}

func (e *Engine) SetVerifyBeforeLaunch(v bool) {
	cfg := e.config.Get()
	cfg.VerifyBeforeLaunch = v
	e.config.UpdateConfig(cfg)
}
