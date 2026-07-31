package downloader

import "encoding/json"

type EventType string

const (
	EventProgress EventType = "download_progress"
	EventState    EventType = "download_state"
	EventLog      EventType = "download_log"
	EventError    EventType = "download_error"
)

type Event struct {
	Type EventType   `json:"type"`
	ID   string      `json:"id"`
	Data interface{} `json:"data"`
}

func BroadcastProgress(fn func([]byte), id string, p *DownloadProgress) {
	if fn == nil {
		return
	}
	data, _ := json.Marshal(Event{Type: EventProgress, ID: id, Data: p})
	fn(data)
}

func BroadcastState(fn func([]byte), id string, state DownloadState) {
	if fn == nil {
		return
	}
	data, _ := json.Marshal(Event{Type: EventState, ID: id, Data: map[string]string{"state": string(state)}})
	fn(data)
}

func BroadcastLog(fn func([]byte), id string, msg string) {
	if fn == nil {
		return
	}
	data, _ := json.Marshal(Event{Type: EventLog, ID: id, Data: map[string]string{"message": msg}})
	fn(data)
}

func BroadcastError(fn func([]byte), id string, errMsg string) {
	if fn == nil {
		return
	}
	data, _ := json.Marshal(Event{Type: EventError, ID: id, Data: map[string]string{"error": errMsg}})
	fn(data)
}
