package ui

import (
	"github.com/MattiaPun/SubTUI/v2/internal/api"
	tea "github.com/charmbracelet/bubbletea"
)

func fetchLyricsCmd(songID, artist, title string) tea.Cmd {
	return func() tea.Msg {
		result, err := api.FetchLyrics(songID, artist, title)
		if err != nil {
			return LyricsErrorMsg{Err: err}
		}
		return LyricsLoadedMsg{Result: result, SongID: songID}
	}
}
