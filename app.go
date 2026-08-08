package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"StepLauncher/internal/Config"
	launcherassets "StepLauncher/internal/Core/Assets"
	downloader "StepLauncher/internal/Core/Downloader"
	"StepLauncher/internal/Handlers"
	engine "StepLauncher/internal/Handlers/Engine"
	engineconfig "StepLauncher/internal/Handlers/Engine/engineconfig"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	handler *Handlers.App
	engine  *engine.Engine
}

func NewApp() *App {
	eng, err := engine.NewEngine()
	if err != nil {
		println("Engine init error:", err.Error())
		eng = nil
	}
	cfgPath := filepath.Join(defaultConfigDir(), "launcher_config.json")
	if eng != nil {
		cfgPath = filepath.Join(eng.ConfigManager().RootDir(), "launcher_config.json")
	}
	return &App{handler: Handlers.NewApp(eng, cfgPath), engine: eng}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.handler.SetContext(ctx)
	a.handler.SetEventCallback(func(eventType string, data []byte) {
		runtime.EventsEmit(ctx, eventType, string(data))
	})
	a.handler.Startup()
}

func (a *App) shutdown(ctx context.Context) {
	if a.handler != nil {
		a.handler.Shutdown()
	}
	if a.engine != nil {
		a.engine.Shutdown()
	}
}

func defaultConfigDir() string {
	if appData := os.Getenv("APPDATA"); appData != "" {
		return filepath.Join(appData, ".StepLauncher")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".StepLauncher")
}

func (a *App) GetConfig() Config.Config {
	return a.handler.GetConfig()
}

func (a *App) GetMinecraftConfig() Config.MinecraftConfig {
	return a.handler.GetMinecraftConfig()
}

func (a *App) UpdateMinecraftConfig(mc Config.MinecraftConfig) {
	a.handler.UpdateMinecraftConfig(mc)
}

func (a *App) SetAuthVerify(verify bool) {
	a.handler.SetAuthVerify(verify)
}

func (a *App) SetProxy(enabled bool, host string, port int, user, pass string) {
	a.handler.SetProxy(enabled, host, port, user, pass)
}

func (a *App) MaxRAMGB() int {
	return a.handler.MaxRAMGB()
}

func (a *App) SetMaxRAM(gb int) {
	a.handler.SetMaxRAM(gb)
}

func (a *App) SetMaxMbps(mbps float64) {
	a.handler.SetMaxMbps(mbps)
}

func (a *App) SetConcurrentDownloads(n int) {
	a.handler.SetConcurrentDownloads(n)
}

func (a *App) SetVerifyIntegrity(v bool) {
	a.handler.SetVerifyIntegrity(v)
}

func (a *App) GetRichPresenceConfig() Config.RichPresenceConfig {
	return a.handler.GetRichPresenceConfig()
}

func (a *App) SetRichPresenceEnabled(v bool) {
	a.handler.SetRichPresenceEnabled(v)
}

func (a *App) CheckForUpdates() {
	if a.handler != nil {
		a.handler.CheckForUpdates()
	}
}

func (a *App) NewsRefreshIndex() {
	if a.handler != nil {
		a.handler.NewsRefreshIndex()
	}
}

func (a *App) NewsLoadRelease(version string) {
	if a.handler != nil {
		a.handler.NewsLoadRelease(version)
	}
}

func (a *App) NewsLoadChangelog(version string) {
	if a.handler != nil {
		a.handler.NewsLoadChangelog(version)
	}
}

func (a *App) NewsLoadMarkdown(url string) {
	if a.handler != nil {
		a.handler.NewsLoadMarkdown(url)
	}
}

func (a *App) ApplyUpdate() error {
	if a.handler == nil {
		return errors.New("handler no disponible")
	}
	return a.handler.ApplyUpdate()
}

func (a *App) GetCheckForUpdatesOnStart() bool {
	if a.handler == nil {
		return false
	}
	return a.handler.GetCheckForUpdatesOnStart()
}

func (a *App) SetCheckForUpdatesOnStart(v bool) {
	if a.handler != nil {
		a.handler.SetCheckForUpdatesOnStart(v)
	}
}

func (a *App) GetUIScale() int {
	return a.handler.GetUIScale()
}

func (a *App) SetUIScale(percent int) {
	a.handler.SetUIScale(percent)
}

func (a *App) SetIdle(idle Config.IdleConfig) {
	a.handler.SetIdle(idle)
}

func (a *App) SetHideLauncher(v bool) {
	a.handler.SetHideLauncher(v)
}

func (a *App) UpdatePersonalization(p Config.Personalization) {
	a.handler.UpdatePersonalization(p)
}

func (a *App) LocalAssetsDir() string {
	return a.handler.LocalAssetsDir()
}

func (a *App) ReadLocalFile(rel string) ([]byte, error) {
	return a.handler.ReadLocalFile(rel)
}

type ScreenshotInfo = Handlers.ScreenshotInfo

func (a *App) ListScreenshots() ([]ScreenshotInfo, error) {
	if a.handler == nil {
		return nil, errors.New("handler no disponible")
	}
	return a.handler.ListScreenshots()
}

func (a *App) ImportBackground(src, kind string) (string, error) {
	return a.handler.ImportBackground(src, kind)
}

func (a *App) PickBackgroundFile(kind string) (string, error) {
	return a.handler.PickBackgroundFile(kind)
}

func (a *App) ResetConfig() error {
	return a.handler.ResetConfig()
}

func (a *App) TotalRAMGB() int {
	return a.handler.TotalRAMGB()
}

func (a *App) DetectJavaInstallations() []string {
	return a.handler.DetectJavaInstallations()
}

func (a *App) GetCacheInfo() engine.CacheInfo {
	if a.handler == nil {
		return engine.CacheInfo{}
	}
	return a.handler.GetCacheInfo()
}

func (a *App) ClearAllCache() int {
	return a.handler.ClearAllCache()
}

func (a *App) GetLauncherAssets() launcherassets.Assets {
	if a.handler == nil {
		return launcherassets.Default()
	}
	return a.handler.GetLauncherAssets()
}

func (a *App) SaveLauncherAssets(asset launcherassets.Assets) {
	if a.handler != nil {
		a.handler.SaveLauncherAssets(asset)
	}
}

func (a *App) ListFontFiles() []string {
	if a.handler == nil {
		return []string{}
	}
	return a.handler.ListFontFiles()
}

func (a *App) PickFontFile() (string, error) {
	if a.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return a.handler.PickFontFile()
}

func (a *App) ImportFont(src string) (string, error) {
	if a.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return a.handler.ImportFont(src)
}

func (a *App) DeleteFontFile(name string) error {
	if a.handler == nil {
		return errors.New("handler no disponible")
	}
	return a.handler.DeleteFontFile(name)
}

func (a *App) EngineInfo() engine.EngineInfo {
	if a.engine == nil {
		return engine.EngineInfo{}
	}
	return a.engine.EngineInfo()
}

func (a *App) EngineConfig() engineconfig.Config {
	if a.engine == nil {
		return engineconfig.DefaultConfig()
	}
	return a.engine.Config()
}

func (a *App) UpdateEngineConfig(cfg engineconfig.Config) {
	if a.engine == nil {
		return
	}
	a.engine.UpdateConfig(cfg)
}

func (a *App) RecommendedRAM() RecommendedRAMResult {
	if a.engine == nil {
		return RecommendedRAMResult{}
	}
	minRAM, maxRAM, gcPreset := a.engine.RecommendedRAM()
	return RecommendedRAMResult{MinRam: minRAM, MaxRam: maxRAM, GcPreset: gcPreset}
}

func (a *App) RecommendedRAMGB() int {
	if a.engine == nil {
		return 0
	}
	return a.engine.RecommendedRAMGB()
}

func (a *App) DeleteCacheCategory(category string) int {
	if a.engine == nil {
		return 0
	}
	return a.engine.DeleteCacheCategory(category)
}

func (a *App) DeleteCacheEntry(category, key string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.DeleteCacheEntry(category, key)
}

func (a *App) RefreshCache(category, key string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.RefreshCache(category, key)
}

func (a *App) StartDownload(version string, filter engine.DownloadFilter, maxRetries, maxConcurrency int, skipVerify bool, stallTimeoutMs, maxStallRetries int) *engine.DownloadInfo {
	if a.engine == nil {
		return nil
	}
	return a.engine.StartDownload(version, filter, maxRetries, maxConcurrency, skipVerify, stallTimeoutMs, maxStallRetries)
}

func (a *App) StartFullDownload(version string) *engine.DownloadInfo {
	if a.engine == nil {
		return nil
	}
	return a.engine.StartFullDownload(version)
}

func (a *App) GetDownloadStatus(id string) (*engine.DownloadProgress, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.GetDownloadStatus(id)
}

func (a *App) PauseDownload(id string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.PauseDownload(id)
}

func (a *App) ResumeDownload(id string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.ResumeDownload(id)
}

func (a *App) CancelDownload(id string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.CancelDownload(id)
}

func (a *App) ListDownloads() []*engine.DownloadInfo {
	if a.engine == nil {
		return []*engine.DownloadInfo{}
	}
	return a.engine.ListDownloads()
}

func (a *App) GetDownload(id string) *engine.DownloadInfo {
	if a.engine == nil {
		return nil
	}
	return a.engine.GetDownload(id)
}

func (a *App) GetVersions(versionType string) ([]engine.VersionInfo, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.GetVersions(versionType)
}

func (a *App) ListDownloadedVersions() []engine.InstalledVersion {
	if a.engine == nil {
		return []engine.InstalledVersion{}
	}
	return a.engine.ListDownloadedVersions()
}

func (a *App) FetchVersionManifest() (*downloader.Manifest, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.FetchVersionManifest()
}

func (a *App) RefreshManifests() (int, error) {
	if a.engine == nil {
		return 0, nil
	}
	return a.engine.RefreshManifests()
}

func (a *App) GetHistory() []engine.HistoryEntry {
	if a.engine == nil {
		return []engine.HistoryEntry{}
	}
	return a.engine.GetHistory()
}

func (a *App) GetCrashHistory() []engine.CrashEntry {
	if a.engine == nil {
		return []engine.CrashEntry{}
	}
	return a.engine.GetCrashHistory()
}

func (a *App) GetHistoryByVersion(version string) []engine.HistoryEntry {
	if a.engine == nil {
		return []engine.HistoryEntry{}
	}
	return a.engine.GetHistoryByVersion(version)
}

func (a *App) GetMostPlayed(limit int) []engine.HistoryEntry {
	if a.engine == nil {
		return []engine.HistoryEntry{}
	}
	return a.engine.GetMostPlayed(limit)
}

func (a *App) GetRecentHistory(limit int) []engine.HistoryEntry {
	if a.engine == nil {
		return []engine.HistoryEntry{}
	}
	return a.engine.GetRecentHistory(limit)
}

func (a *App) DeleteHistoryEntry(id string) (bool, error) {
	if a.engine == nil {
		return false, errors.New("engine no disponible")
	}
	return a.engine.DeleteHistoryEntry(id)
}

func (a *App) ClearHistory() int {
	if a.engine == nil {
		return 0
	}
	return a.engine.ClearHistory()
}

func (a *App) GetHistoryStats() HistoryStatsResult {
	if a.engine == nil {
		return HistoryStatsResult{}
	}
	stats, total := a.engine.GetHistoryStats()
	return HistoryStatsResult{Stats: stats, Total: total}
}

func (a *App) CreateInstance(req engine.CreateInstanceReq) (*CreateInstanceResult, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	metadata, downloadID, err := a.engine.CreateInstance(req)
	if err != nil {
		return nil, err
	}
	return &CreateInstanceResult{Metadata: metadata, DownloadId: downloadID}, nil
}

func (a *App) ListInstances() []*engine.InstanceInfo {
	if a.engine == nil {
		return []*engine.InstanceInfo{}
	}
	return a.engine.ListInstances()
}

func (a *App) GetInstance(name string) (*GetInstanceResult, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	metadata, cfg, err := a.engine.GetInstance(name)
	if err != nil {
		return nil, err
	}
	return &GetInstanceResult{Metadata: metadata, Config: cfg}, nil
}

func (a *App) DeleteInstance(name string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.DeleteInstance(name)
}

func (a *App) UpdateInstanceMetadata(name string, req engine.UpdateMetadataReq) (*engine.InstanceMetadata, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.UpdateInstanceMetadata(name, req)
}

func (a *App) UpdateInstanceConfig(name string, cfg *engine.InstanceLaunchConfig) (*engine.InstanceLaunchConfig, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.UpdateInstanceConfig(name, cfg)
}

func (a *App) AddInstanceVersion(name string, req engine.AddVersionReq) (*AddInstanceVersionResult, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	downloadID, version, err := a.engine.AddInstanceVersion(name, req)
	if err != nil {
		return nil, err
	}
	return &AddInstanceVersionResult{DownloadId: downloadID, Version: version}, nil
}

func (a *App) ListInstanceVersions(name string) ([]string, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.ListInstanceVersions(name)
}

func (a *App) RemoveInstanceVersion(name, version string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.RemoveInstanceVersion(name, version)
}

func (a *App) VerifyInstance(name string) ([]engine.VerifyResult, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.VerifyInstance(name)
}

func (a *App) VerifyInstanceVersion(name, version string) (*engine.VerifyResult, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.VerifyInstanceVersion(name, version)
}

func (a *App) CloneInstance(name, newName string, copyVersions bool) (*engine.InstanceMetadata, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.CloneInstance(name, newName, copyVersions)
}

func (a *App) LaunchInstance(name string, username, uuid, accessToken, xuid, clientID string) (*engine.InstanceLaunchResult, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.LaunchInstance(name, username, uuid, accessToken, xuid, clientID)
}

func (a *App) GetInstanceDownloadStatus(dlID string) InstanceDownloadStatusResult {
	if a.engine == nil {
		return InstanceDownloadStatusResult{}
	}
	id, version, state, _ := a.engine.GetInstanceDownloadStatus(dlID)
	return InstanceDownloadStatusResult{Id: id, Version: version, State: state}
}

func (a *App) CancelInstanceDownload(dlID string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.CancelInstanceDownload(dlID)
}

func (a *App) LaunchMinecraft(cfg engine.LaunchConfig) (*engine.GameResp, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.LaunchMinecraft(cfg)
}

func (a *App) ListGames() []engine.GameResp {
	if a.engine == nil {
		return []engine.GameResp{}
	}
	return a.engine.ListGames()
}

func (a *App) GetGame(id string) *engine.GameResp {
	if a.engine == nil {
		return nil
	}
	return a.engine.GetGame(id)
}

func (a *App) StopGame(id string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.StopGame(id)
}

func (a *App) StopAllGames() {
	if a.engine != nil {
		a.engine.StopAllGames()
	}
}

func (a *App) ListModLoaders() []string {
	if a.engine == nil {
		return []string{}
	}
	return a.engine.ListModLoaders()
}

func (a *App) GetModLoaderVersions(loader, mcVersion string) ([]engine.ModLoaderVersion, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.GetModLoaderVersions(loader, mcVersion)
}

func (a *App) ResolveModLoaderVersion(loader, mcVersion, strategy string) (string, error) {
	if a.engine == nil {
		return "", errors.New("engine no disponible")
	}
	return a.engine.ResolveModLoaderVersion(loader, mcVersion, strategy)
}

func (a *App) InstallModLoader(loader, loaderVersion, mcVersion, instancePath string) (*engine.ModLoaderInstallResult, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.InstallModLoader(loader, loaderVersion, mcVersion, instancePath)
}

func (a *App) GetInstalledModLoader(instancePath string) (*engine.InstalledLoader, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.GetInstalledModLoader(instancePath)
}

func (a *App) RemoveModLoaderState(instancePath string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.RemoveModLoaderState(instancePath)
}

func (a *App) BuildModLoaderExecution(instancePath, versionsDir, librariesPath string) (*engine.ExecutionPlan, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.BuildModLoaderExecution(instancePath, versionsDir, librariesPath)
}

func (a *App) ListProfiles() map[string]*engine.Profile {
	if a.engine == nil {
		return map[string]*engine.Profile{}
	}
	return a.engine.ListProfiles()
}

func (a *App) GetProfile(name string) (*engine.Profile, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.GetProfile(name)
}

func (a *App) CreateProfile(p *engine.Profile) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.CreateProfile(p)
}

func (a *App) UpdateProfile(name string, p *engine.Profile) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.UpdateProfile(name, p)
}

func (a *App) DeleteProfile(name string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.DeleteProfile(name)
}

func (a *App) GetSelectedProfile() string {
	if a.engine == nil {
		return ""
	}
	return a.engine.GetSelectedProfile()
}

func (a *App) SetSelectedProfile(name string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.SetSelectedProfile(name)
}

func (a *App) GetSelectedVersion() string {
	if a.engine == nil {
		return ""
	}
	return a.engine.GetSelectedVersion()
}

func (a *App) SetSelectedVersion(version string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.SetSelectedVersion(version)
}

func (a *App) ListAccounts() []engine.AccountInfo {
	if a.engine == nil {
		return []engine.AccountInfo{}
	}
	return a.engine.ListAccounts()
}

func (a *App) GetAccount(id string) (*engine.AccountInfo, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.GetAccount(id)
}

func (a *App) CreateAccount(req engine.CreateAccountReq) (*engine.AccountInfo, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.CreateAccount(req)
}

func (a *App) UpdateAccount(id string, req engine.CreateAccountReq) (*engine.AccountInfo, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.UpdateAccount(id, req)
}

func (a *App) DeleteAccount(id string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.DeleteAccount(id)
}

func (a *App) GetSelectedAccount() string {
	if a.engine == nil {
		return ""
	}
	return a.engine.GetSelectedAccount()
}

func (a *App) SetSelectedAccount(id string) error {
	if a.engine == nil {
		return errors.New("engine no disponible")
	}
	return a.engine.SetSelectedAccount(id)
}

func (a *App) ResolveAccountCredentials(id string) (*engine.AccountCredentials, error) {
	if a.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return a.engine.ResolveAccountCredentials(id)
}

func (a *App) LoginAuthlib(req engine.AuthlibLoginReq) {
	if a.engine != nil {
		a.engine.LoginAuthlib(req)
	}
}

func (a *App) CancelAuthlibLogin() {
	if a.engine != nil {
		a.engine.CancelLogin()
	}
}

func (a *App) RefreshAccount(id string) {
	if a.engine != nil {
		a.engine.RefreshAccount(id)
	}
}

func (a *App) RefreshAllAccounts() int {
	if a.engine == nil {
		return 0
	}
	return a.engine.RefreshAllAccounts()
}

func (a *App) SetAccountsAutoRefresh(v bool) {
	if a.engine != nil {
		a.engine.SetAccountsAutoRefresh(v)
	}
}

func (a *App) GetAccountsAutoRefresh() bool {
	if a.engine == nil {
		return false
	}
	return a.engine.GetAccountsAutoRefresh()
}

func (a *App) GetAccountAssets(id string) {
	if a.engine != nil {
		a.engine.GetAccountAssets(id)
	}
}

type CreateInstanceResult struct {
	Metadata   *engine.InstanceMetadata `json:"metadata"`
	DownloadId string                   `json:"downloadId"`
}

type GetInstanceResult struct {
	Metadata *engine.InstanceMetadata     `json:"metadata"`
	Config   *engine.InstanceLaunchConfig `json:"config"`
}

type AddInstanceVersionResult struct {
	DownloadId string `json:"downloadId"`
	Version    string `json:"version"`
}

type HistoryStatsResult struct {
	Stats []engine.VersionStats `json:"stats"`
	Total int                   `json:"total"`
}

type InstanceDownloadStatusResult struct {
	Id      string `json:"id"`
	Version string `json:"version"`
	State   string `json:"state"`
}

type RecommendedRAMResult struct {
	MinRam   int    `json:"minRam"`
	MaxRam   int    `json:"maxRam"`
	GcPreset string `json:"gcPreset"`
}
