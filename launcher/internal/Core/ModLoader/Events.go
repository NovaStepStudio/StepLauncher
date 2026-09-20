package modloader

import "encoding/json"

type ModLoaderEvent struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId,omitempty"`
	Loader    string `json:"loader,omitempty"`
	Version   string `json:"version,omitempty"`
	MCVersion string `json:"mcVersion,omitempty"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
	Progress  int    `json:"progress,omitempty"`
	Total     int    `json:"total,omitempty"`
}

func marshalEvent(e ModLoaderEvent) []byte {
	data, _ := json.Marshal(e)
	return data
}

func ResolvingEvent(sessionId, loader, version, mcVersion string) []byte {
	return marshalEvent(ModLoaderEvent{
		Type: "modloader_resolving", SessionID: sessionId,
		Loader: loader, Version: version, MCVersion: mcVersion,
	})
}

func DownloadingEvent(sessionId, loader, name string, progress, total int) []byte {
	return marshalEvent(ModLoaderEvent{
		Type: "modloader_downloading", SessionID: sessionId,
		Loader: loader, Message: name, Progress: progress, Total: total,
	})
}

func InstallingEvent(sessionId, loader, message string) []byte {
	return marshalEvent(ModLoaderEvent{
		Type: "modloader_installing", SessionID: sessionId,
		Loader: loader, Message: message,
	})
}

func InstalledEvent(sessionId, loader, version, mcVersion string) []byte {
	return marshalEvent(ModLoaderEvent{
		Type: "modloader_installed", SessionID: sessionId,
		Loader: loader, Version: version, MCVersion: mcVersion,
	})
}

func ErrorEvent(sessionId, loader, err string) []byte {
	return marshalEvent(ModLoaderEvent{
		Type: "modloader_error", SessionID: sessionId,
		Loader: loader, Error: err,
	})
}
