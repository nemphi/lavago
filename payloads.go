package lavago

import "time"

const (
	trackStartEvent      = "TrackStartEvent"
	trackEndEvent        = "TrackEndEvent"
	trackExceptionEvent  = "TrackExceptionEvent"
	trackStuckEvent      = "TrackStuckEvent"
	webSocketClosedEvent = "WebSocketClosedEvent"
)

type recvDataEventPayload struct {
	Op          string `json:"op,omitempty"`
	GuildID     string `json:"guildId,omitempty"`
	Type        string `json:"type,omitempty"`
	Track       string `json:"track,omitempty"`
	Reason      string `json:"reason,omitempty"`
	Error       string `json:"error,omitempty"`
	ThresholdMs int    `json:"thresholdMs,omitempty"`
	Code        int    `json:"code,omitempty"`
	ByRemote    bool   `json:"byRemote"`
}

type basePayload struct {
	Op      string `json:"op,omitempty"`
	GuildID string `json:"guildId,omitempty"`
}

type resumePayload struct {
	Op      string `json:"op,omitempty"`
	Key     string `json:"key,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

type serverUpdatePayload struct {
	Op        string             `json:"op,omitempty"`
	GuildID   string             `json:"guildId,omitempty"`
	SessionID string             `json:"sessionId,omitempty"`
	Event     voiceServerPayload `json:"event,omitempty"`
}

type voiceServerPayload struct {
	Token    string `json:"token,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

type playerPlayPayload struct {
	Op        string        `json:"op,omitempty"`
	GuildID   string        `json:"guildId,omitempty"`
	Track     string        `json:"track,omitempty"`
	NoReplace bool          `json:"noReplace,omitempty"`
	StartTime time.Duration `json:"startTime,omitempty"`
	EndTime   time.Duration `json:"endTime,omitempty"`
	Volume    int           `json:"volume,omitempty"`
	Pause     bool          `json:"pause"`
}

type playerStopPayload struct {
	Op      string `json:"op,omitempty"`
	GuildID string `json:"guildId,omitempty"`
}

type playerPausePayload struct {
	Op      string `json:"op,omitempty"`
	GuildID string `json:"guildId,omitempty"`
	Pause   bool   `json:"pause"`
}

type playerSeekPayload struct {
	Op       string        `json:"op,omitempty"`
	GuildID  string        `json:"guildId,omitempty"`
	Position time.Duration `json:"position,omitempty"`
}

type playerVolumePayload struct {
	Op      string `json:"op,omitempty"`
	GuildID string `json:"guildId,omitempty"`
	Volume  int    `json:"position,omitempty"`
}

type playerDestroyPayload struct {
	Op      string `json:"op,omitempty"`
	GuildID string `json:"guildId,omitempty"`
}
