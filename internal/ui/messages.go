package ui

import "github.com/MattiaPun/SubTUI/v2/internal/api"

// FetchLyricsMsg is sent internally to trigger a lyrics fetch for the current song.
type FetchLyricsMsg struct {
	SongID string
	Artist string
	Title  string
}

// LyricsLoadedMsg is sent when lyrics have been successfully fetched.
type LyricsLoadedMsg struct {
	Result api.LyricsResult
	SongID string // the ID of the song these lyrics belong to
}

// LyricsErrorMsg is sent when the lyrics fetch fails.
type LyricsErrorMsg struct {
	Err    error
	SongID string
}
