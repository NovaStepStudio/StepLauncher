package engine

import (
	"sort"

	lhistory "StepLauncher/internal/Core/Launcher/History"
)

type HistoryEntry = lhistory.Entry

type CrashEntry = lhistory.CrashEntry

type InstanceStats = lhistory.InstanceStats

type VersionStat = lhistory.VersionStat

func (e *Engine) GetHistory() []HistoryEntry {
	return e.history.GetEntries()
}

func (e *Engine) GetCrashHistory() []CrashEntry {
	return e.crashHist.GetEntries()
}

func (e *Engine) GetHistoryByVersion(version string) []HistoryEntry {
	return e.history.GetEntriesByVersion(version)
}

func (e *Engine) GetMostPlayed(limit int) []HistoryEntry {
	return e.history.GetMostPlayed(limit)
}

func (e *Engine) GetRecentHistory(limit int) []HistoryEntry {
	return e.history.GetRecent(limit)
}

func (e *Engine) DeleteHistoryEntry(id string) (bool, error) {
	return e.history.DeleteEntry(id)
}

func (e *Engine) GetInstanceStats(name string) InstanceStats {
	stats := e.history.GetInstanceStats(name)
	if e.launcher != nil {
		for _, g := range e.launcher.List() {
			if g.InstanceName == name && g.IsRunning() {
				stats.Running = true
				break
			}
		}
	}
	return stats
}

func (e *Engine) ClearHistory() int {
	return e.history.Clear()
}

type VersionStats struct {
	Version     string `json:"version"`
	PlayCount   int    `json:"playCount"`
	LastPlayed  int64  `json:"lastPlayed"`
	TotalPlayed int    `json:"totalPlayed"`
}

func (e *Engine) GetHistoryStats() ([]VersionStats, int) {
	entries := e.history.GetEntries()
	stats := make(map[string]*VersionStats)
	for _, entry := range entries {
		s, ok := stats[entry.Version]
		if !ok {
			s = &VersionStats{Version: entry.Version}
			stats[entry.Version] = s
		}
		s.PlayCount++
		s.TotalPlayed += entry.PlayTimeSeconds
		if entry.Timestamp > s.LastPlayed {
			s.LastPlayed = entry.Timestamp
		}
	}

	result := make([]VersionStats, 0, len(stats))
	for _, s := range stats {
		result = append(result, *s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastPlayed > result[j].LastPlayed
	})

	return result, len(entries)
}
