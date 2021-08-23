package lavago

// Lavalink's REST response.
type SearchResult struct {
	Status    SearchStatus    `json:"loadType,omitempty"`
	Playlist  SearchPlaylist  `json:"playlistInfo,omitempty"`
	Exception SearchException `json:"exception,omitempty"`
	Tracks    []*Track        `json:"tracks,omitempty"`
}

// Search status when searching for songs via Lavalink.
type SearchStatus string

const (
	// Returned when a single track is loaded.
	TrackLoadedSearchStatus SearchStatus = "TRACK_LOADED"
	// Returned when a playlist is loaded.
	PlaylistLoadedSearchStatus SearchStatus = "PLAYLIST_LOADED"
	// Returned when a search result is made (i.e ytsearch: some song).
	SearchResultSearchStatus SearchStatus = "SEARCH_RESULT"
	// Returned if no matches/sources could be found for a given identifier.
	NoMatchesSearchStatus SearchStatus = "NO_MATCHES"
	// Returned if Lavaplayer failed to load something for some reason.
	LoadFailedSearchStatus SearchStatus = "LOAD_FAILED"
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
