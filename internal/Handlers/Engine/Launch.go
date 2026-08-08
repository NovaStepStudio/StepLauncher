package engine

import (
	"strings"

	"StepLauncher/internal/Core/Launcher"
	"StepLauncher/internal/Core/Launcher/Profile"
)

type LaunchConfig = launcher.LaunchConfig
type AdvancedConfig = launcher.AdvancedConfig
type GameInstance = launcher.GameInstance
type GameStatus = launcher.GameStatus

const (
	GameStarting = launcher.GameStarting
	GameRunning  = launcher.GameRunning
	GameExited   = launcher.GameExited
	GameCrashed  = launcher.GameCrashed
	GameStopped  = launcher.GameStopped
)

type GameResp struct {
	ID            string     `json:"id"`
	PID           int        `json:"pid"`
	Version       string     `json:"version"`
	Status        GameStatus `json:"status"`
	StartTime     string     `json:"startTime,omitempty"`
	ExitCode      int        `json:"exitCode"`
	LogPath       string     `json:"logPath,omitempty"`
	CrashLog      string     `json:"crashLog,omitempty"`
	CrashReason   string     `json:"crashReason,omitempty"`
	CrashCategory string     `json:"crashCategory,omitempty"`
}

func gameToResp(g *launcher.GameInstance) GameResp {
	r := GameResp{
		ID: g.ID, PID: g.PID, Version: g.Version,
		Status: g.GetStatus(), ExitCode: g.ExitCode,
		LogPath: g.LogPath, CrashLog: g.CrashLog,
		CrashReason: g.CrashReason, CrashCategory: g.CrashCategory,
	}
	if !g.StartTime.IsZero() {
		r.StartTime = g.StartTime.Format("2006-01-02 15:04:05")
	}
	return r
}

func (e *Engine) LaunchMinecraft(cfg LaunchConfig) (*GameResp, error) {
	cfg = e.buildBaseLaunchConfig(cfg)
	if cfg.Profile != "" && e.profiles != nil {
		p, err := e.profiles.Get(cfg.Profile)
		if err == nil {
			cfg = mergeProfileIntoConfig(cfg, p)
		}
	}
	if err := e.fillAccountCredentials(&cfg, ""); err != nil {
		return nil, err
	}

	inst, err := e.launcher.Launch(cfg)
	if err != nil {
		return nil, err
	}
	resp := gameToResp(inst)
	return &resp, nil
}

func (e *Engine) buildBaseLaunchConfig(cfg launcher.LaunchConfig) launcher.LaunchConfig {
	ec := e.config.Get()
	adv := launcher.DefaultAdvancedConfig()

	maxRAM := ec.MaxRAMMB
	if maxRAM <= 0 {
		maxRAM = 2048
	}
	adv.MaxRAM = maxRAM
	minRAM := maxRAM / 2
	if minRAM < 512 {
		minRAM = 512
	}
	adv.MinRAM = minRAM

	switch ec.JavaMode {
	case "official":
		adv.UseOfficialJava = true
	case "system":
		adv.UseSystemJava = true
	case "custom":
		adv.JavaExec = ec.JavaCustomPath
	}

	if s := strings.TrimSpace(ec.JavaArgs); s != "" {
		adv.JavaArgs = splitArgs(s)
	}
	if s := strings.TrimSpace(ec.GameArgs); s != "" {
		adv.GameArgs = splitArgs(s)
	}

	if ec.WindowWidth > 0 && ec.WindowHeight > 0 {
		adv.ResWidth = ec.WindowWidth
		adv.ResHeight = ec.WindowHeight
		adv.CustomResolution = true
	}
	adv.Fullscreen = ec.Fullscreen

	hwAccel := ec.HardwareAcceleration
	if !ec.HardwareEnabled {
		hwAccel = false
	}
	if ec.CompatMode {
		hwAccel = false
	}
	if hwAccel {
		enabled := true
		adv.HardwareAcceleration = &enabled
	} else {
		adv.HardwareAcceleration = &hwAccel
	}
	adv.GPUPreference = ""
	if hwAccel {
		switch ec.GPUType {
		case "dedicated":
			adv.GPUPreference = "dgpu"
		case "integrated":
			adv.GPUPreference = "igpu"
		}
	}

	cfg.Advanced = &adv
	return cfg
}

func splitArgs(s string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	for _, r := range s {
		switch {
		case r == '"':
			inQuote = !inQuote
		case (r == ' ' || r == '\t' || r == '\n' || r == '\r') && !inQuote:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

func (e *Engine) ListGames() []GameResp {
	instances := e.launcher.List()
	resp := make([]GameResp, 0, len(instances))
	for _, g := range instances {
		resp = append(resp, gameToResp(g))
	}
	return resp
}

func (e *Engine) GetGame(id string) *GameResp {
	g := e.launcher.Get(id)
	if g == nil {
		return nil
	}
	resp := gameToResp(g)
	return &resp
}

func (e *Engine) StopGame(id string) error {
	return e.launcher.Stop(id)
}

func (e *Engine) StopAllGames() {
	e.launcher.StopAll()
}

func (e *Engine) RecommendedRAM() (minRAM, maxRAM int, gcPreset string) {
	return e.launcher.RecommendedRAM()
}

func mergeProfileIntoConfig(cfg launcher.LaunchConfig, p *profile.Profile) launcher.LaunchConfig {
	if p.Version != "" {
		cfg.Version = p.Version
	}
	adv := cfg.Adv()
	if p.GameDir != "" {
		adv.GameDir = p.GameDir
	}
	if p.JavaExec != "" {
		adv.JavaExec = p.JavaExec
	}
	if p.JavaArgs != "" {
		adv.JavaArgs = append(adv.JavaArgs, splitArgs(p.JavaArgs)...)
	}
	if p.ResWidth > 0 {
		adv.ResWidth = p.ResWidth
		adv.CustomResolution = true
	}
	if p.ResHeight > 0 {
		adv.ResHeight = p.ResHeight
		adv.CustomResolution = true
	}
	adv.Fullscreen = p.Fullscreen
	cfg.Advanced = &adv
	return cfg
}
