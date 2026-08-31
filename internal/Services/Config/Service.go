package config

import (
	"StepLauncher/internal/Config"
	"StepLauncher/internal/Handlers"
	engine "StepLauncher/internal/Handlers/Engine"
	"errors"
)

// ConfigService gestiona configuracion general, cache y hardware.
type ConfigService struct {
	handler *Handlers.App
	engine  *engine.Engine
}

func NewConfigService(handler *Handlers.App, eng *engine.Engine) *ConfigService {
	return &ConfigService{handler: handler, engine: eng}
}

type RecommendedRAMResult struct {
	MinRam   int    `json:"minRam"`
	MaxRam   int    `json:"maxRam"`
	GcPreset string `json:"gcPreset"`
}

func (s *ConfigService) CancelIntegrityCheck() {
	if s.handler != nil {
		s.handler.CancelIntegrityCheck()
	}
}

func (s *ConfigService) ClearAllCache() int {
	return s.handler.ClearAllCache()
}

func (s *ConfigService) DeleteCacheCategory(category string) int {
	if s.engine == nil {
		return 0
	}
	return s.engine.DeleteCacheCategory(category)
}

func (s *ConfigService) DeleteCacheEntry(category, key string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.DeleteCacheEntry(category, key)
}

func (s *ConfigService) DetectJavaInstallations() []string {
	return s.handler.DetectJavaInstallations()
}

func (s *ConfigService) GetCacheInfo() engine.CacheInfo {
	if s.handler == nil {
		return engine.CacheInfo{}
	}
	return s.handler.GetCacheInfo()
}

func (s *ConfigService) GetCheckForUpdatesOnStart() bool {
	if s.handler == nil {
		return false
	}
	return s.handler.GetCheckForUpdatesOnStart()
}

func (s *ConfigService) GetConfig() Config.Config {
	return s.handler.GetConfig()
}

func (s *ConfigService) GetIntegritySector() string {
	if s.handler == nil {
		return "todo"
	}
	return s.handler.GetIntegritySector()
}

func (s *ConfigService) GetMinecraftConfig() Config.MinecraftConfig {
	return s.handler.GetMinecraftConfig()
}

func (s *ConfigService) GetSeparateGameDir() bool {
	if s.handler == nil {
		return true
	}
	return s.handler.GetSeparateGameDir()
}

func (s *ConfigService) IntegrityStatus() engine.IntegrityProgress {
	if s.handler == nil {
		return engine.IntegrityProgress{State: engine.IntegrityStateIdle}
	}
	return s.handler.IntegrityStatus()
}

func (s *ConfigService) MaxRAMGB() int {
	return s.handler.MaxRAMGB()
}

func (s *ConfigService) RecommendedRAM() RecommendedRAMResult {
	if s.engine == nil {
		return RecommendedRAMResult{}
	}
	minRAM, maxRAM, gcPreset := s.engine.RecommendedRAM()
	return RecommendedRAMResult{MinRam: minRAM, MaxRam: maxRAM, GcPreset: gcPreset}
}

func (s *ConfigService) RecommendedRAMGB() int {
	if s.engine == nil {
		return 0
	}
	return s.engine.RecommendedRAMGB()
}

func (s *ConfigService) RefreshCache(category, key string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.RefreshCache(category, key)
}

func (s *ConfigService) ResetConfig() error {
	return s.handler.ResetConfig()
}

func (s *ConfigService) SetAuthVerify(verify bool) {
	s.handler.SetAuthVerify(verify)
}

func (s *ConfigService) SetCheckForUpdatesOnStart(v bool) {
	if s.handler != nil {
		s.handler.SetCheckForUpdatesOnStart(v)
	}
}

func (s *ConfigService) SetConcurrentDownloads(n int) {
	s.handler.SetConcurrentDownloads(n)
}

func (s *ConfigService) SetIntegritySector(sector string) {
	if s.handler != nil {
		s.handler.SetIntegritySector(sector)
	}
}

func (s *ConfigService) SetLaunchAfterInstall(v bool) {
	if s.handler != nil {
		s.handler.SetLaunchAfterInstall(v)
	}
}

func (s *ConfigService) SetMaxMbps(mbps float64) {
	s.handler.SetMaxMbps(mbps)
}

func (s *ConfigService) SetMaxRAM(gb int) {
	s.handler.SetMaxRAM(gb)
}

func (s *ConfigService) SetProxy(enabled bool, host string, port int, user, pass string) {
	s.handler.SetProxy(enabled, host, port, user, pass)
}

func (s *ConfigService) SetSeparateGameDir(v bool) {
	if s.handler != nil {
		s.handler.SetSeparateGameDir(v)
	}
}

func (s *ConfigService) SetVerifyBeforeLaunch(v bool) {
	s.handler.SetVerifyBeforeLaunch(v)
}

func (s *ConfigService) SetVerifyIntegrity(v bool) {
	s.handler.SetVerifyIntegrity(v)
}

func (s *ConfigService) StartIntegrityCheck(scope string) error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.StartIntegrityCheck(scope)
}

func (s *ConfigService) TotalRAMGB() int {
	return s.handler.TotalRAMGB()
}

func (s *ConfigService) UpdateMinecraftConfig(mc Config.MinecraftConfig) {
	s.handler.UpdateMinecraftConfig(mc)
}
