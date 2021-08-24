package lavago

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/emirpasic/gods/lists"
	"github.com/emirpasic/gods/lists/arraylist"
)

// Describes the status of a `Player`
type PlayerState byte

const (
	// Player isn't conencted to a voice channel
	PlayerStateNone PlayerState = iota
	// Currently playing a track
	PlayerStatePlaying
	// Not playing anything
	PlayerStateStopped
	// Playing a track but paused
	PlayerStatePaused
)

// Arguments for Player.Play
type PlayArgs struct {
	//  Which track to play
	Track *Track
	// Whether to replace the track. Returns ReplacedReason when used
	NoReplace bool
	// Set the volume of the player when playing a Track
	Volume int
	// Whether to pause the player when Track is ready to play
	ShouldPause bool
	// Start time of Track
	StartTime time.Duration
	// End time of Track
	EndTime time.Duration
}

// Represents a `*discordgo.VoiceChannel` connection
type Player struct {
	// Last time player was updated.
	LastUpdate time.Time
	// Player's current state.
	State PlayerState
	// Default queue.
	Queue lists.List
	// Current track that is playing.
	Track *Track
	// Voice channel this player is connected to.
	GuildID string
	// Player's current volume.
	Volume int

	socket *Socket
	sync.RWMutex
}

// Creates a new player
func NewPlayer(socket *Socket, guildID string) *Player {
	return &Player{
		Queue:   arraylist.New(),
		GuildID: guildID,
		socket:  socket,
	}
}

func (p *Player) Close() error {
	p.Stop()
	data, err := json.Marshal(playerDestroyPayload{
		Op:      "destroy",
		GuildID: p.GuildID,
	})
	if err != nil {
		return err
	}
	p.Lock()
	err = p.socket.Send(data)
	p.Queue.Clear()
	p.Track = nil
	p.State = PlayerStateNone
	p.Unlock()
	return err
}

// Plays the specified track with provided arguments.
func (p *Player) Play(args PlayArgs) error {
	if args.Track == nil {
		return errors.New("can't play nil Track")
	}
	p.Lock()
	if args.ShouldPause {
		p.State = PlayerStatePaused
	} else {
		p.State = PlayerStatePlaying
	}
	p.Unlock()
	if args.Volume < 0 {
		return errors.New("can't play with volume < 0")
	}
	if args.Volume > 1000 {
		return errors.New("can't play with volume > 1000")
	}
	p.Lock()
	p.Volume = args.Volume
	p.Track = args.Track
	p.Unlock()
	data, err := json.Marshal(playerPlayPayload{
		Op:        "play",
		GuildID:   p.GuildID,
		Track:     args.Track.Track,
		NoReplace: args.NoReplace,
		StartTime: args.StartTime,
		EndTime:   args.EndTime,
		Volume:    args.Volume,
		Pause:     args.ShouldPause,
	})
	if err != nil {
		return err
	}
	return p.socket.Send(data)
}

// Plays the specified track.
func (p *Player) PlayTrack(track *Track) error {
	if track == nil {
		return errors.New("can't play nil Track")
	}
	p.Lock()
	p.State = PlayerStatePlaying
	p.Track = track
	p.Unlock()
	data, err := json.Marshal(playerPlayPayload{
		Op:      "play",
		GuildID: p.GuildID,
		Track:   track.Track,
		Volume:  100,
		Pause:   false,
	})
	if err != nil {
		return err
	}
	return p.socket.Send(data)
}

// Stops the current track if any is playing.
func (p *Player) Stop() error {
	p.Lock()
	p.State = PlayerStateStopped
	p.Unlock()
	data, err := json.Marshal(playerStopPayload{
		Op:      "stop",
		GuildID: p.GuildID,
	})
	if err != nil {
		return err
	}
	return p.socket.Send(data)
}

// Pauses the current track if any is playing.
func (p *Player) Pause() error {
	if p.State == PlayerStateNone {
		return errors.New("player's current state is set to None. Please make sure Player is connected to a voice channel")
	}
	p.Lock()
	if p.Track == nil {
		p.State = PlayerStateStopped
	} else {
		p.State = PlayerStatePaused
	}
	p.Unlock()
	data, err := json.Marshal(playerPausePayload{
		Op:      "pause",
		GuildID: p.GuildID,
		Pause:   true,
	})
	if err != nil {
		return err
	}
	return p.socket.Send(data)
}

// Resume the current track if any is playing.
func (p *Player) Resume() error {
	if p.State == PlayerStateNone {
		return errors.New("player's current state is set to None. Please make sure Player is connected to a voice channel")
	}
	p.Lock()
	if p.Track == nil {
		p.State = PlayerStateStopped
	} else {
		p.State = PlayerStatePlaying
	}
	p.Unlock()
	data, err := json.Marshal(playerPausePayload{
		Op:      "pause",
		GuildID: p.GuildID,
		Pause:   false,
	})
	if err != nil {
		return err
	}
	return p.socket.Send(data)
}

// Skips the current track after the specified delay.
func (p *Player) Skip(delay time.Duration) (skipped *Track, current *Track, err error) {
	if p.State == PlayerStateNone {
		return nil, nil, errors.New("player's current state is set to None. Please make sure Player is connected to a voice channel")
	}
	p.Lock()
	skipped = p.Track
	currentI, exists := p.Queue.Get(0)
	if !exists {
		p.Unlock()
		return skipped, nil, p.Stop()
	}
	p.Queue.Remove(0)
	p.Unlock()
	current = currentI.(*Track)
	if delay != 0 {
		time.Sleep(delay)
	}
	err = p.PlayTrack(current)
	return
}

// Seeks the current track to specified position.
func (p *Player) Seek(position time.Duration) error {
	if p.State == PlayerStateNone {
		return errors.New("player's current state is set to None. Please make sure Player is connected to a voice channel")
	}
	if position > p.Track.Info.Length {
		return fmt.Errorf("value must not be higer than %s", p.Track.Info.Length.String())
	}
	data, err := json.Marshal(playerSeekPayload{
		Op:       "seek",
		GuildID:  p.GuildID,
		Position: position,
	})
	if err != nil {
		return err
	}
	return p.socket.Send(data)
}

// Changes the current volume and updates p.Volume
func (p *Player) UpdateVolume(volume int) error {
	p.Lock()
	p.Volume = volume
	p.Unlock()
	data, err := json.Marshal(playerVolumePayload{
		Op:      "volume",
		GuildID: p.GuildID,
		Volume:  volume,
	})
	if err != nil {
		return err
	}
	return p.socket.Send(data)
}
