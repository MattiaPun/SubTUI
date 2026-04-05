package ui

import (
	"github.com/MattiaPun/SubTUI/v2/internal/api"
	tea "github.com/charmbracelet/bubbletea"
)

func fetchLyricsCmd(songID, artist, title string) tea.Cmd {
	return func() tea.Msg {
		result, err := api.FetchLyrics(songID, artist, title)
		if err != nil {
			return LyricsErrorMsg{Err: err, SongID: songID}
		}
		return LyricsLoadedMsg{Result: result, SongID: songID}
	}
}

func fetchMediaViewLyricsCmd(m model) tea.Cmd {
	if len(m.queue) == 0 || m.queueIndex < 0 || m.queueIndex >= len(m.queue) {
		return nil
	}

	currentSong := m.queue[m.queueIndex]
	cmds := []tea.Cmd{fetchLyricsCmd(currentSong.ID, currentSong.Artist, currentSong.Title)}

	if m.queueIndex+1 < len(m.queue) {
		nextSong := m.queue[m.queueIndex+1]
		cmds = append(cmds, fetchLyricsCmd(nextSong.ID, nextSong.Artist, nextSong.Title))
	}

	return tea.Batch(cmds...)
}
