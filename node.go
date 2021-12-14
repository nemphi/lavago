package lavago

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Contains information about track position.
type PlayerUpdatedEvent struct {
	// Player for which this event fired.
	Player *Player `json:"-,omitempty"`
	// Track sent by Lavalink.
	Track *Track `json:"track,omitempty"`
	State struct {
		// Track's current position
		Position  int   `json:"position,omitempty"`
		Time      int64 `json:"time,omitempty"`
		Connected bool  `json:"connected,omitempty"`
	} `json:"state,omitempty"`
}

// Information about Lavalink statistics.
type StatsReceivedEvent struct {
	// Machine's CPU info.
	CPU string `json:"cpu,omitempty"`
	// Audio frames.
	Frames string `json:"frames,omitempty"`
	// General memory information about Lavalink.
	Memory string `json:"memory,omitempty"`
	// Connected players.
	Players int `json:"players,omitempty"`
	// Players that are currently playing.
	PlayingPlayers int `json:"playing_players,omitempty"`
	// Lavalink uptime.
	Uptime time.Time `json:"uptime,omitempty"`
}

// Information about the track that started.
type TrackStartedEvent struct {
	// Player for which this event fired.
	Player *Player `json:"-"`
	// Track sent by Lavalink.
	Track *Track `json:"track,omitempty"`
}

// Specifies the reason for why the track ended.
type TrackEndReason byte

const (
	// This means that the track itself emitted a terminator. This is usually caused by the track reaching the end,
	// however it will also be used when it ends due to an exception.
	FinishedReason TrackEndReason = 'F'
	// This means that the track failed to start, throwing an exception before providing any audio.
	LoadFailedReason TrackEndReason = 'L'
	// The track was stopped due to the player being stopped by either calling `Stop()` or `Play()`
	StoppedReason TrackEndReason = 'S'
	// The track stopped playing because a new track started playing. Note that with this reason, the old track will still
	// play until either its buffer runs out or audio from the new track is available.
	ReplacedReason TrackEndReason = 'R'
	// The track was stopped because the cleanup threshold for the audio player was reached. This triggers when the amount
	// of time passed since the last call to AudioPlayer#provide() has reached the threshold specified in player manager
	// configuration. This may also indicate either a leaked audio player which was discarded, but not stopped.
	CleanupReason TrackEndReason = 'C'
)

// Information about track that ended.
type TrackEndedEvent struct {
	// Player for which this event fired.
	Player *Player `json:"-"`
	// Track sent by Lavalink.
	Track *Track `json:"track,omitempty"`
	// Reason for track ending.
	Reason TrackEndReason `json:"reason,omitempty"`
}

// Information about track that threw an exception.
type TrackExceptionEvent struct {
	// Player for which this event fired.
	Player *Player `json:"-"`
	// Track sent by Lavalink.
	Track *Track `json:"track,omitempty"`
	// Reason for why track threw an exception.
	ErrorMessage string `json:"error_message,omitempty"`
}

// Information about track that got stuck.
type TrackStuckEvent struct {
	// Player for which this event fired.
	Player *Player `json:"-"`
	// Track sent by Lavalink.
	Track *Track `json:"track,omitempty"`
	// How long track was stuck for.
	Threshold time.Duration `json:"threshold,omitempty"`
}

// Discord's voice websocket event.
type WebSocketClosedEvent struct {
	// Guild's voice connection.
	GuildID string `json:"guild_id,omitempty"`
	// 4xxx codes are bad.
	Code int `json:"code,omitempty"`
	// Reason for closing websocket connection.
	Reason string `json:"reason,omitempty"`
	// ???
	ByRemote bool `json:"by_remote,omitempty"`
}

type voiceState struct {
	GuildID   string
	SessionID string
}

type Node struct {
	cfg         *Config
	socket      *Socket
	connected   bool
	players     *sync.Map // map[string(GuildID)]*Player
	voiceStates *sync.Map // map[string(GuildID)]voiceState

	ConnectVoice    func(guildID, channelID string, deaf bool) error
	PlayerUpdated   func(PlayerUpdatedEvent)
	StatsReceived   func(StatsReceivedEvent)
	TrackStarted    func(TrackStartedEvent)
	TrackEnded      func(TrackEndedEvent)
	TrackException  func(TrackExceptionEvent)
	TrackStuck      func(TrackStuckEvent)
	WebSocketClosed func(WebSocketClosedEvent)
}

func NewNode(cfg *Config) (*Node, error) {
	n := &Node{
		cfg:         cfg,
		socket:      NewSocket(cfg),
		players:     &sync.Map{},
		voiceStates: &sync.Map{},
	}
	n.socket.DataReceived = n.socketDataReceived
	n.socket.ErrorReceived = n.socketOnError
	n.socket.OnOpen = n.socketOnOpen
	return n, nil
}

func (n *Node) Connect(userID, shardCount string) error {
	headers := http.Header{}
	headers.Add("User-Id", userID)
	headers.Add("Num-Shards", shardCount)
	headers.Add("Authorization", n.cfg.Authorization)
	headers.Add("Client-Name", "Lavago")
	if n.cfg.EnableResume {
		headers.Add("Resume-Key", n.cfg.ResumeKey)
	}
	if n.cfg.UserAgent != "" {
		headers.Add("User-Agent", n.cfg.UserAgent)
	}
	err := n.socket.Connect(headers)
	if err != nil {
		return err
	}
	n.connected = true
	return nil
}

func (n *Node) Close() error {
	if !n.connected {
		return errors.New("can't close non-connected node")
	}
	n.connected = false
	n.players = nil
	n.voiceStates = nil
	return n.socket.Close()
}

func (n *Node) Join(guildID, voiceChannelID string) (*Player, error) {
	if !n.connected {
		return nil, errors.New("can't join on non-connected node")
	}
	if voiceChannelID == "" {
		return nil, errors.New("can't join (empty string) voice channel")
	}
	playerI, exists := n.players.Load(guildID)
	if exists {
		return playerI.(*Player), nil
	}

	if n.ConnectVoice != nil {
		err := n.ConnectVoice(guildID, voiceChannelID, n.cfg.SelfDeaf)
		if err != nil {
			return nil, err
		}
	}

	p := NewPlayer(n.socket, guildID)
	n.players.Store(guildID, p)
	return p, nil
}

func (n *Node) Leave(guildID string) error {
	if !n.connected {
		return errors.New("can't leave on non-connected node")
	}
	playerI, exists := n.players.Load(guildID)
	if !exists {
		return nil
	}
	p := playerI.(*Player)
	err := p.Close()
	n.players.Delete(guildID)
	return err
}

func (n *Node) HasPlayer(guildID string) bool {
	_, exists := n.players.Load(guildID)
	return exists
}

func (n *Node) GetPlayer(guildID string) *Player {
	p, exists := n.players.Load(guildID)
	if !exists {
		return nil
	}
	return p.(*Player)
}

func (n *Node) Search(stype SearchType, query string) (*SearchResult, error) {
	if query == "" {
		return nil, errors.New("can't search with empty query string")
	}
	urlPath := ""
	switch stype {
	case SoundCloud:
		urlPath = "/loadtracks?identifier=scsearch:" + url.QueryEscape(query)
	case YouTubeMusic:
		urlPath = "/loadtracks?identifier=ytmsearch:" + url.QueryEscape(query)
	case YouTube:
		urlPath = "/loadtracks?identifier=ytsearch:" + url.QueryEscape(query)
	case Direct:
		urlPath = "/loadtracks?identifier=" + url.QueryEscape(query)
	default:
		urlPath = "/loadtracks?identifier=" + url.QueryEscape(query)
	}
	req, err := http.NewRequest("GET", n.cfg.httpEndpoint()+urlPath, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", n.cfg.Authorization)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	sr := &SearchResult{}
	err = json.NewDecoder(res.Body).Decode(sr)
	if err != nil {
		return nil, err
	}
	return sr, nil
}

func (n *Node) socketOnOpen() {
	n.connected = true
	if n.cfg.EnableResume {
		data, err := json.Marshal(resumePayload{
			Op:      "configureResuming",
			Key:     n.cfg.ResumeKey,
			Timeout: int(n.cfg.ResumeTimeout.Seconds()),
		})
		if err != nil {
			panic("Could not marshal resume")
		}
		err = n.socket.Send(data)
		if err != nil {
			panic("Could not set resume")
		}
	}
}

func (n *Node) socketOnError(err error) {
	// TODO: make better
	fmt.Println("ERR: " + err.Error())
}

func (n *Node) socketDataReceived(data []byte) {
	if len(data) == 0 {
		// TODO: Handle
		// n.logger.LogError(...)
		panic("*Node.DataReceived: len(data) = 0")
	}
	bp := &basePayload{}
	err := json.Unmarshal(data, bp)
	if err != nil {
		panic("*Node.DataReceived: json.Unmarshal => " + err.Error())
	}
	switch bp.Op {
	case "stats":
		if n.StatsReceived == nil {
			break
		}
		sr := StatsReceivedEvent{}
		err = json.Unmarshal(data, &sr)
		if err != nil {
			panic("*Node.DataReceived: json.Unmarshal 'stats' => " + err.Error())
		}
		if n.StatsReceived == nil {
			break
		}
		n.StatsReceived(sr)
	case "playerUpdate":
		pu := PlayerUpdatedEvent{}
		err = json.Unmarshal(data, &pu)
		if err != nil {
			panic("*Node.DataReceived: json.Unmarshal 'playerUpdated' => " + err.Error())
		}
		p := n.GetPlayer(bp.GuildID)
		if p == nil {
			break
		}
		if p.Track != nil {
			p.Track.updatePosition(pu.State.Position)
		}
		p.LastUpdate = time.Unix(pu.State.Time, 0)
		if n.PlayerUpdated == nil {
			break
		}
		pu.Player = p
		n.PlayerUpdated(pu)
	case "event":
		rp := recvDataEventPayload{}
		err = json.Unmarshal(data, &rp)
		if err != nil {
			panic("*Node.DataReceived: json.Unmarshal 'event' => " + err.Error())
		}
		switch rp.Type {
		case trackStartEvent:
			p := n.GetPlayer(bp.GuildID)
			if p == nil {
				break
			}
			p.Lock()
			p.State = PlayerStatePlaying
			p.Unlock()
			if n.TrackStarted == nil {
				break
			}
			n.TrackStarted(TrackStartedEvent{Player: p, Track: p.Track})
		case trackEndEvent:
			p := n.GetPlayer(bp.GuildID)
			if p == nil {
				break
			}
			p.Lock()
			p.State = PlayerStateStopped
			p.Unlock()
			if n.TrackEnded == nil {
				break
			}
			n.TrackEnded(TrackEndedEvent{Player: p, Track: p.Track, Reason: TrackEndReason(rp.Reason[0])})
		case trackExceptionEvent:
			p := n.GetPlayer(bp.GuildID)
			if p == nil {
				break
			}
			p.Lock()
			p.State = PlayerStateStopped
			p.Unlock()
			if n.TrackException == nil {
				break
			}
			n.TrackException(TrackExceptionEvent{Player: p, Track: p.Track, ErrorMessage: rp.Error})
		case trackStuckEvent:
			p := n.GetPlayer(bp.GuildID)
			if p == nil {
				break
			}
			p.Lock()
			p.State = PlayerStateStopped
			p.Unlock()
			if n.TrackStuck == nil {
				break
			}
			dur, err := time.ParseDuration(fmt.Sprintf("%vms", rp.ThresholdMs))
			if err != nil {
				panic("*Node.DataReceived: time.ParseDuration 'event' => " + err.Error())
			}
			n.TrackStuck(TrackStuckEvent{Player: p, Track: p.Track, Threshold: dur})
		case webSocketClosedEvent:
			if n.WebSocketClosed == nil {
				break
			}
			n.WebSocketClosed(WebSocketClosedEvent{
				GuildID:  rp.GuildID,
				Reason:   rp.Reason,
				Code:     rp.Code,
				ByRemote: rp.ByRemote,
			})
		}
	default:
		panic("*Node.DataReceived: switch.default")
	}
}

func (n *Node) OnVoiceStateUpdate(shardUserID, triggerUserID, guildID, sessionID string) {
	if shardUserID != triggerUserID {
		return
	}
	n.voiceStates.Store(guildID, voiceState{GuildID: guildID, SessionID: sessionID})
}

func (n *Node) OnVoiceServerUpdate(guildID, endpoint, token string) {
	vsI, exists := n.voiceStates.Load(guildID)
	if !exists {
		return
	}
	vs := vsI.(voiceState)
	sp := &serverUpdatePayload{
		Op:        "voiceUpdate",
		GuildID:   vs.GuildID,
		SessionID: vs.SessionID,
		Event: voiceServerPayload{
			Endpoint: endpoint,
			Token:    token,
		},
	}
	data, err := json.Marshal(sp)
	if err != nil {
		panic("*Node.OnVoiceServerUpdate json.Marshal")
	}
	err = n.socket.Send(data)
	if err != nil {
		fmt.Println("*Node.OnVoiceServerUpdate ERR socked.Send: " + err.Error())
	}
}
