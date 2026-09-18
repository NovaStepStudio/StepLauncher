package music

import (
	"StepLauncher/internal/Config"
	"StepLauncher/internal/Handlers"
	music "StepLauncher/internal/Music"
	musichistory "StepLauncher/internal/Music/MusicHistory"
	nowplaying "StepLauncher/internal/Music/NowPlaying"
	playlists "StepLauncher/internal/Music/Playlist"
	"errors"
)

// MusicService gestiona biblioteca musical, scanner y playlists.
type MusicService struct {
	handler *Handlers.App
}

func NewMusicService(handler *Handlers.App) *MusicService {
	return &MusicService{handler: handler}
}

func (s *MusicService) AddMusicFolder(folder string) error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.AddMusicFolder(folder)
}

func (s *MusicService) AddMusicHistory(path, title, artist, coverUrl string) []musichistory.Entry {
	if s.handler == nil {
		return []musichistory.Entry{}
	}
	return s.handler.AddMusicHistory(path, title, artist, coverUrl)
}

func (s *MusicService) ClearCoverCache() int {
	if s.handler == nil {
		return 0
	}
	return s.handler.ClearCoverCache()
}

func (s *MusicService) ClearMusicHistory() error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.ClearMusicHistory()
}

func (s *MusicService) ClearNowPlayingQueue() error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.ClearNowPlayingQueue()
}

func (s *MusicService) CreatePlaylist(title string, trackPaths []string) (*playlists.Playlist, error) {
	if s.handler == nil {
		return nil, errors.New("handler no disponible")
	}
	return s.handler.CreatePlaylist(title, trackPaths)
}

func (s *MusicService) DeletePlaylist(id string) error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.DeletePlaylist(id)
}

func (s *MusicService) GetAllMusicMetadata() map[string]Handlers.MusicMetadataEntry {
	if s.handler == nil {
		return map[string]Handlers.MusicMetadataEntry{}
	}
	return s.handler.GetAllMusicMetadata()
}

func (s *MusicService) GetCachedCover(trackPath string) (string, error) {
	if s.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return s.handler.GetCachedCover(trackPath)
}

func (s *MusicService) GetCoverCacheInfo() Handlers.CoverCacheInfo {
	if s.handler == nil {
		return Handlers.CoverCacheInfo{}
	}
	return s.handler.GetCoverCacheInfo()
}

func (s *MusicService) GetCoverCacheStatus() Handlers.CoverCacheInfo {
	return s.GetCoverCacheInfo()
}

func (s *MusicService) GetMusicCoverBase64(trackPath, variant string) (string, error) {
	if s.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return s.handler.GetMusicCoverBase64(trackPath, variant)
}

func (s *MusicService) GetMusicFolders() []string {
	if s.handler == nil {
		return []string{}
	}
	return s.handler.GetMusicFolders()
}

func (s *MusicService) GetMusicHistory() []musichistory.Entry {
	if s.handler == nil {
		return []musichistory.Entry{}
	}
	return s.handler.GetMusicHistory()
}

func (s *MusicService) GetMusicMetadata(trackPath string) (Handlers.MusicMetadataEntry, error) {
	if s.handler == nil {
		return Handlers.MusicMetadataEntry{}, nil
	}
	entry, ok := s.handler.GetMusicMetadata(trackPath)
	if !ok {
		return Handlers.MusicMetadataEntry{}, nil
	}
	return entry, nil
}

func (s *MusicService) GetMusicPanelConfig() Config.MusicPanelConfig {
	if s.handler == nil {
		return Config.MusicPanelConfig{}
	}
	return s.handler.GetMusicPanelConfig()
}

func (s *MusicService) GetMusicScanProgress() Handlers.MusicScanProgress {
	if s.handler == nil {
		return Handlers.MusicScanProgress{}
	}
	return s.handler.GetMusicScanProgress()
}

func (s *MusicService) GetMusicTracksBatch(paths []string, coverVariant string) ([]music.MusicTrackDTO, error) {
	if s.handler == nil {
		return []music.MusicTrackDTO{}, nil
	}
	return s.handler.GetMusicTracksBatch(paths, coverVariant)
}

func (s *MusicService) GetMusicTracksPaged(offset, limit int, coverVariant, query string) (music.MusicPageDTO, error) {
	if s.handler == nil {
		return music.MusicPageDTO{Tracks: []music.MusicTrackDTO{}, Total: 0, Offset: offset, Limit: limit}, nil
	}
	return s.handler.GetMusicTracksPaged(offset, limit, coverVariant, query)
}

func (s *MusicService) GetNowPlayingQueue() nowplaying.Queue {
	if s.handler == nil {
		return nowplaying.Queue{Tracks: []string{}, CurrentIndex: -1}
	}
	return s.handler.GetNowPlayingQueue()
}

func (s *MusicService) GetPlaylist(id string) (*playlists.Playlist, error) {
	if s.handler == nil {
		return nil, errors.New("handler no disponible")
	}
	return s.handler.GetPlaylist(id)
}

func (s *MusicService) ImportPlaylistFile(path string) (*playlists.Playlist, error) {
	if s.handler == nil {
		return nil, errors.New("handler no disponible")
	}
	return s.handler.ImportPlaylistFile(path)
}

func (s *MusicService) ListPlaylists() []playlists.Playlist {
	if s.handler == nil {
		return []playlists.Playlist{}
	}
	return s.handler.ListPlaylists()
}

func (s *MusicService) PickMusicFolder() (string, error) {
	if s.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return s.handler.PickMusicFolder()
}

func (s *MusicService) PickPlaylistFile() (string, error) {
	if s.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return s.handler.PickPlaylistFile()
}

func (s *MusicService) PickCustomCoverFile() (string, error) {
	if s.handler == nil {
		return "", errors.New("handler no disponible")
	}
	return s.handler.PickCustomCoverFile()
}

func (s *MusicService) RefreshCachedCover(trackPath string) error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.RefreshCachedCover(trackPath)
}

func (s *MusicService) RemoveMusicFolder(folder string) error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.RemoveMusicFolder(folder)
}

func (s *MusicService) SaveCachedCover(trackPath, b64Data, mime string) error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.SaveCachedCover(trackPath, b64Data, mime)
}

func (s *MusicService) SaveMusicMetadata(trackPath, title, artist string, duration float64, coverData, mime string) error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.SaveMusicMetadata(trackPath, title, artist, duration, coverData, mime)
}

func (s *MusicService) ScanAllMusicFolders() ([]string, error) {
	if s.handler == nil {
		return nil, errors.New("handler no disponible")
	}
	return s.handler.ScanAllMusicFolders()
}

func (s *MusicService) ScanMusicFolder(folder string) ([]string, error) {
	if s.handler == nil {
		return nil, errors.New("handler no disponible")
	}
	return s.handler.ScanMusicFolder(folder)
}

func (s *MusicService) CancelMusicScan() bool {
	if s.handler == nil {
		return false
	}
	return s.handler.CancelMusicScan()
}

func (s *MusicService) SetMusicFolder(folder string) error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.SetMusicFolder(folder)
}

func (s *MusicService) SetMusicFolders(folders []string) error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.SetMusicFolders(folders)
}

func (s *MusicService) SetNowPlayingQueue(tracks []string, idx int, curPath string) error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.SetNowPlayingQueue(tracks, idx, curPath)
}

func (s *MusicService) UpdateMusicPanelConfig(p Config.MusicPanelConfig) error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.UpdateMusicPanelConfig(p)
}

func (s *MusicService) UpdatePlaylist(id string, p playlists.Playlist) (*playlists.Playlist, error) {
	if s.handler == nil {
		return nil, errors.New("handler no disponible")
	}
	return s.handler.UpdatePlaylist(id, p)
}

func (s *MusicService) GetMusicLibraryStats() music.MusicLibraryStats {
	if s.handler == nil {
		return music.MusicLibraryStats{}
	}
	return s.handler.GetMusicLibraryStats()
}

func (s *MusicService) RebuildMusicIndex() error {
	if s.handler == nil {
		return errors.New("handler no disponible")
	}
	return s.handler.RebuildMusicIndex()
}

func (s *MusicService) ClearMusicCache() int {
	if s.handler == nil {
		return 0
	}
	return s.handler.ClearMusicCache()
}

func (s *MusicService) InitializeMusicLibrary() ([]string, error) {
	if s.handler == nil {
		return nil, errors.New("handler no disponible")
	}
	return s.handler.InitializeMusicLibrary()
}
