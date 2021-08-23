package lavago

import (
	"fmt"
	"time"
)

// Config for a `Node`
type Config struct {
	// Authorization is the password for the server.
	Authorization string
	// Max buffer size for receiving websocket message.
	BufferSize int
	// Toggle Lavalink's resume capability.
	EnableResume bool
	// Server's IP/Hostname.
	Hostname string
	// Log serverity for logging everything.
	LogSeverity int
	// Port to connect to.
	Port int
	// Use Secure Socket Layer (SSL) security protocol when connecting to Lavalink.
	SSL bool
	// Applies User-Agent header to all requests.
	UserAgent string
	// How many reconnect attempts are allowed.
	ReconnectAttempts int
	// Reconnection delay for retrying websocket connection.
	ReconnectDelay time.Duration
	// ResumeKey utilized to identify the client with the node
	ResumeKey string
	// Timeout duration for the resume request
	ResumeTimeout time.Duration
	// Whether to enable self deaf for bot.
	SelfDeaf bool
}

func NewConfig() *Config {
	return &Config{
		Authorization:     "youshallnotpass",
		BufferSize:        512,
		EnableResume:      true,
		Hostname:          "127.0.0.1",
		LogSeverity:       5,
		Port:              2333,
		SSL:               false,
		ReconnectAttempts: 10,
		ReconnectDelay:    10 * time.Second,
		ResumeKey:         "Lavago",
		ResumeTimeout:     30 * time.Second,
		SelfDeaf:          true,
	}
}

func (cfg *Config) socketEndpoint() string {
	if cfg.SSL {
		return fmt.Sprintf("wss://%s:%v", cfg.Hostname, cfg.Port)
	}
	return fmt.Sprintf("ws://%s:%v", cfg.Hostname, cfg.Port)
}

func (cfg *Config) httpEndpoint() string {
	if cfg.SSL {
		return fmt.Sprintf("https://%s:%v", cfg.Hostname, cfg.Port)
	}
	return fmt.Sprintf("http://%s:%v", cfg.Hostname, cfg.Port)
}
