package lavago

// Lavalink's REST response.
type SearchResult struct {
	Status    SearchStatus    `json:"loadType,omitempty"`
	Playlist  SearchPlaylist  `json:"playlistInfo,omitempty"`
	Exception SearchException `json:"exception,omitempty"`
	Tracks    []*Track        `json:"tracks,omitempty"`
}

// Search status when searching for songs via Lavalink.
type SearchStatus byte

const (
	// Returned when a single track is loaded.
	TrackLoadedSearchStatus SearchStatus = 84
	// Returned when a playlist is loaded.
	PlaylistLoadedSearchStatus SearchStatus = 80
	// Returned when a search result is made (i.e ytsearch: some song).
	SearchResultSearchStatus SearchStatus = 83
	// Returned if no matches/sources could be found for a given identifier.
	NoMatchesSearchStatus SearchStatus = 78
	// Returned if Lavaplayer failed to load something for some reason.
	LoadFailedSearchStatus SearchStatus = 76
)

// Only available if SearchStatus was PlaylistLoaded.
type SearchPlaylist struct {
	Name          string `json:"name,omitempty"`
	SelectedTrack int    `json:"selectedTrack,omitempty"`
}

// If SearchStatus was LoadFailed then Exception is returned.
type SearchException struct {
	Message  string `json:"message,omitempty"`
	Severity string `json:"severity,omitempty"`
}

type SearchType byte

const (
	YouTube SearchType = iota
	YouTubeMusic
	SoundCloud
	Direct
)
