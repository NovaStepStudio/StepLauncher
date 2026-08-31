package game

import (
	engine "StepLauncher/internal/Handlers/Engine"
	"errors"
)

// GameService gestiona ejecucion de juegos e historial.
type GameService struct {
	engine *engine.Engine
}

func NewGameService(eng *engine.Engine) *GameService {
	return &GameService{engine: eng}
}

type HistoryStatsResult struct {
	Stats []engine.VersionStats `json:"stats"`
	Total int                   `json:"total"`
}

func (s *GameService) ClearHistory() int {
	if s.engine == nil {
		return 0
	}
	return s.engine.ClearHistory()
}

func (s *GameService) DeleteHistoryEntry(id string) (bool, error) {
	if s.engine == nil {
		return false, errors.New("engine no disponible")
	}
	return s.engine.DeleteHistoryEntry(id)
}

func (s *GameService) GetCrashHistory() []engine.CrashEntry {
	if s.engine == nil {
		return []engine.CrashEntry{}
	}
	return s.engine.GetCrashHistory()
}

func (s *GameService) GetGame(id string) *engine.GameResp {
	if s.engine == nil {
		return nil
	}
	return s.engine.GetGame(id)
}

func (s *GameService) GetGameLaunchInfo(id string) string {
	if s.engine == nil {
		return ""
	}
	return s.engine.GetGameLaunchInfo(id)
}

func (s *GameService) GetHistory() []engine.HistoryEntry {
	if s.engine == nil {
		return []engine.HistoryEntry{}
	}
	return s.engine.GetHistory()
}

func (s *GameService) GetHistoryByVersion(version string) []engine.HistoryEntry {
	if s.engine == nil {
		return []engine.HistoryEntry{}
	}
	return s.engine.GetHistoryByVersion(version)
}

func (s *GameService) GetHistoryStats() HistoryStatsResult {
	if s.engine == nil {
		return HistoryStatsResult{}
	}
	stats, total := s.engine.GetHistoryStats()
	return HistoryStatsResult{Stats: stats, Total: total}
}

func (s *GameService) GetInstanceStats(name string) engine.InstanceStats {
	if s.engine == nil {
		return engine.InstanceStats{}
	}
	return s.engine.GetInstanceStats(name)
}

func (s *GameService) GetMostPlayed(limit int) []engine.HistoryEntry {
	if s.engine == nil {
		return []engine.HistoryEntry{}
	}
	return s.engine.GetMostPlayed(limit)
}

func (s *GameService) GetRecentHistory(limit int) []engine.HistoryEntry {
	if s.engine == nil {
		return []engine.HistoryEntry{}
	}
	return s.engine.GetRecentHistory(limit)
}

func (s *GameService) LaunchInstance(name string, username, uuid, accessToken, xuid, clientID string) (*engine.InstanceLaunchResult, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.LaunchInstance(name, username, uuid, accessToken, xuid, clientID)
}

func (s *GameService) LaunchMinecraft(cfg engine.LaunchConfig) (*engine.GameResp, error) {
	if s.engine == nil {
		return nil, errors.New("engine no disponible")
	}
	return s.engine.LaunchMinecraft(cfg)
}

func (s *GameService) ListGames() []engine.GameResp {
	if s.engine == nil {
		return []engine.GameResp{}
	}
	return s.engine.ListGames()
}

func (s *GameService) StopAllGames() {
	if s.engine != nil {
		s.engine.StopAllGames()
	}
}

func (s *GameService) StopGame(id string) error {
	if s.engine == nil {
		return errors.New("engine no disponible")
	}
	return s.engine.StopGame(id)
}
