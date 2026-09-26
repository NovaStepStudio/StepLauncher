package mods

import "encoding/json"

// ModContentEvent es el progreso de la descarga/instalación de un archivo de
// Modrinth (mod, shader, resourcepack o .mrpack suelto). Se emite con
// e.eventCb usando el campo Type como nombre del evento, igual que los
// eventos modloader_* del orquestador de modloaders.
type ModContentEvent struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId,omitempty"`
	Title     string `json:"title,omitempty"`
	File      string `json:"file,omitempty"`
	Dest      string `json:"dest,omitempty"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
	Progress  int    `json:"progress,omitempty"`
	Total     int    `json:"total,omitempty"`
}

func marshalContentEvent(e ModContentEvent) []byte {
	data, _ := json.Marshal(e)
	return data
}

// ResolvingContentEvent anuncia el inicio de la resolución del destino.
func ResolvingContentEvent(sessionID, title, dest string) []byte {
	return marshalContentEvent(ModContentEvent{
		Type: "modcontent_resolving", SessionID: sessionID,
		Title: title, Dest: dest,
	})
}

// DownloadingContentEvent informa del archivo en descarga (progreso 1/total).
func DownloadingContentEvent(sessionID, title, file string, progress, total int) []byte {
	return marshalContentEvent(ModContentEvent{
		Type: "modcontent_downloading", SessionID: sessionID,
		Title: title, File: file, Progress: progress, Total: total,
	})
}

// InstalledContentEvent cierra la sesión con la ruta relativa instalada.
func InstalledContentEvent(sessionID, title, dest string) []byte {
	return marshalContentEvent(ModContentEvent{
		Type: "modcontent_installed", SessionID: sessionID,
		Title: title, Dest: dest,
	})
}

// ErrorContentEvent cierra la sesión con el motivo del fallo.
func ErrorContentEvent(sessionID, title, err string) []byte {
	return marshalContentEvent(ModContentEvent{
		Type: "modcontent_error", SessionID: sessionID,
		Title: title, Error: err,
	})
}

// ModpackEvent es el progreso de la instalación de un modpack .mrpack:
// descarga del .mrpack, descarga de la versión de Minecraft, instalación del
// modloader, descarga de los archivos del manifiesto y copia de overrides.
type ModpackEvent struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId,omitempty"`
	Instance  string `json:"instance,omitempty"`
	Phase     string `json:"phase,omitempty"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
	Progress  int    `json:"progress,omitempty"`
	Total     int    `json:"total,omitempty"`
}

func marshalModpackEvent(e ModpackEvent) []byte {
	data, _ := json.Marshal(e)
	return data
}

// ModpackProgressEvent informa de una fase en curso del modpack.
func ModpackProgressEvent(sessionID, instance, phase, message string, progress, total int) []byte {
	return marshalModpackEvent(ModpackEvent{
		Type: "modpack_" + phase, SessionID: sessionID,
		Instance: instance, Phase: phase,
		Message: message, Progress: progress, Total: total,
	})
}

// ModpackDoneEvent cierra la instalación del modpack con la instancia lista.
// warning informa de algo no fatal (p. ej. el loader no se pudo instalar y
// hay que ponerlo a mano): viaja en Message para que la UI lo muestre.
func ModpackDoneEvent(sessionID, instance, warning string) []byte {
	return marshalModpackEvent(ModpackEvent{
		Type: "modpack_installed", SessionID: sessionID,
		Instance: instance, Phase: "installed", Message: warning,
	})
}

// ModpackErrorEvent cierra la instalación del modpack con el motivo.
func ModpackErrorEvent(sessionID, instance, err string) []byte {
	return marshalModpackEvent(ModpackEvent{
		Type: "modpack_error", SessionID: sessionID,
		Instance: instance, Phase: "error", Error: err,
	})
}
