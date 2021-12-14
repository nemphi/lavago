package lavago

import "time"

// Track information.
type Track struct {
	// Track's encoded hash.
	Track string    `json:"track,omitempty"`
	Info  TrackInfo `json:"info,omitempty"`
}

type TrackInfo struct {
	// Audio / Video track Id.
	Identifier string `json:"identifier,omitempty"`
	// Track's author.
	Author string `json:"author,omitempty"`
	// Track's title.
	Title string `json:"title,omitempty"`
	// Whether the track is seekable.
	CanSeek bool `json:"isSeekable,omitempty"`
	// Track's length.
	Length time.Duration `json:"length,omitempty"`
	//  Whether the track is a stream.
	IsStream bool `json:"is_stream,omitempty"`
	// Track's current position.
	Position time.Duration `json:"position,omitempty"`
	// Track's url.
	URL string `json:"uri,omitempty"`
	// Name of the source
	SourceName string `json:"sourceName,omitempty"`
}

func (t *Track) updatePosition(pos time.Duration) {
	t.Info.Position = pos
}
